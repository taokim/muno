package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/mocks"
)

// TestManager_StartAgent tests the StartAgent function
func TestManager_StartAgent(t *testing.T) {
	tests := []struct {
		name           string
		agentName      string
		path           string
		agentArgs      []string
		initialized    bool
		getCurrentErr  error
		getNodeErr     error
		executeErr     error
		exitCode       int
		wantErr        bool
		errMsg         string
		expectedCmd    string
	}{
		{
			name:        "successful start with claude",
			agentName:   "claude",
			path:        "/test/repo",
			initialized: true,
			wantErr:     false,
			expectedCmd: "cd /workspace/nodes/test/nodes/repo && claude",
		},
		{
			name:        "successful start with gemini",
			agentName:   "gemini",
			path:        "/test/repo",
			initialized: true,
			wantErr:     false,
			expectedCmd: "cd /workspace/nodes/test/nodes/repo && gemini",
		},
		{
			name:        "default to claude when agent not specified",
			agentName:   "",
			path:        "/test/repo",
			initialized: true,
			wantErr:     false,
			expectedCmd: "cd /workspace/nodes/test/nodes/repo && claude",
		},
		{
			name:        "successful with agent args",
			agentName:   "gemini",
			path:        "/test/repo",
			agentArgs:   []string{"--model", "pro", "--temperature", "0.7"},
			initialized: true,
			wantErr:     false,
			expectedCmd: "cd /workspace/nodes/test/nodes/repo && gemini --model pro --temperature 0.7",
		},
		{
			name:          "successful without path uses current",
			agentName:     "claude",
			path:          "",
			initialized:   true,
			getCurrentErr: nil,
			wantErr:       false,
		},
		{
			name:        "fails when not initialized",
			agentName:   "claude",
			path:        "/test/repo",
			initialized: false,
			wantErr:     true,
			errMsg:      "manager not initialized",
		},
		{
			name:          "fails when getting current node fails",
			agentName:     "claude",
			path:          "",
			initialized:   true,
			getCurrentErr: fmt.Errorf("current node error"),
			wantErr:       true,
			errMsg:        "getting current node",
		},
		{
			name:        "fails when getting node fails",
			agentName:   "gemini",
			path:        "/test/repo",
			initialized: true,
			getNodeErr:  fmt.Errorf("node not found"),
			wantErr:     true,
			errMsg:      "getting node",
		},
		{
			name:        "fails when execute shell fails",
			agentName:   "cursor",
			path:        "/test/repo",
			initialized: true,
			executeErr:  fmt.Errorf("command failed"),
			wantErr:     true,
			errMsg:      "starting cursor",
		},
		{
			name:        "fails when exit code is non-zero",
			agentName:   "copilot",
			path:        "/test/repo",
			initialized: true,
			exitCode:    1,
			wantErr:     true,
			errMsg:      "copilot exited with code 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockTree := mocks.NewMockTreeProvider()
			mockProcess := &MockProcessProviderWithCommand{
				executeErr:  tt.executeErr,
				exitCode:    tt.exitCode,
				expectedCmd: tt.expectedCmd,
			}
			mockUI := mocks.NewMockUIProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockConfig := mocks.NewMockConfigProvider()
			mockGit := mocks.NewMockGitProvider()

			// Setup tree mock behavior
			if tt.getCurrentErr != nil {
				mockTree.SetError("GetCurrent", "", tt.getCurrentErr)
			} else {
				mockTree.SetCurrent(interfaces.NodeInfo{
					Path: "/current/node",
					Name: "current",
				})
			}

			// Only set up GetNode if we're not expecting GetCurrent to fail
			if tt.getCurrentErr == nil {
				nodePath := tt.path
				if nodePath == "" {
					nodePath = "/current/node"  // Use the current node path when no path specified
				}
				if tt.getNodeErr != nil {
					mockTree.SetError("GetNode", nodePath, tt.getNodeErr)
				} else {
					mockTree.SetNode(nodePath, interfaces.NodeInfo{
						Path: nodePath,
						Name: "test-node",
					})
				}
			}

			m := &Manager{
				initialized:     tt.initialized,
				workspace:       "/workspace",
				treeProvider:    mockTree,
				processProvider: mockProcess,
				uiProvider:      mockUI,
				fsProvider:      mockFS,
				configProvider:  mockConfig,
				gitProvider:     mockGit,
				logProvider:     NewDefaultLogProvider(false),
				metricsProvider: NewNoOpMetricsProvider(),
			}

			err := m.StartAgent(tt.agentName, tt.path, tt.agentArgs, false)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// Verify the command was built correctly if expectedCmd is set
				if tt.expectedCmd != "" && mockProcess.lastCommand != "" {
					assert.Contains(t, mockProcess.lastCommand, tt.expectedCmd)
				}
			}
		})
	}
}

// MockProcessProviderWithCommand is a mock that can track and validate ExecuteShell commands
type MockProcessProviderWithCommand struct {
	DefaultProcessProvider
	executeErr  error
	exitCode    int
	expectedCmd string
	lastCommand string
}

func (m *MockProcessProviderWithCommand) ExecuteShell(ctx context.Context, command string, opts interfaces.ProcessOptions) (*interfaces.ProcessResult, error) {
	m.lastCommand = command
	if m.executeErr != nil {
		return nil, m.executeErr
	}
	return &interfaces.ProcessResult{
		ExitCode: m.exitCode,
		Stdout:   "",
		Stderr:   "error output",
	}, nil
}

// TestManager_StartClaude tests the StartClaude function
func TestManager_StartClaude(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		initialized    bool
		getCurrentErr  error
		getNodeErr     error
		executeErr     error
		exitCode       int
		wantErr        bool
		errMsg         string
	}{
		{
			name:        "successful start with path",
			path:        "/test/repo",
			initialized: true,
			wantErr:     false,
		},
		{
			name:          "successful start without path uses current",
			path:          "",
			initialized:   true,
			getCurrentErr: nil,
			wantErr:       false,
		},
		{
			name:        "fails when not initialized",
			path:        "/test/repo",
			initialized: false,
			wantErr:     true,
			errMsg:      "manager not initialized",
		},
		{
			name:          "fails when getting current node fails",
			path:          "",
			initialized:   true,
			getCurrentErr: fmt.Errorf("current node error"),
			wantErr:       true,
			errMsg:        "getting current node",
		},
		{
			name:        "fails when getting node fails",
			path:        "/test/repo",
			initialized: true,
			getNodeErr:  fmt.Errorf("node not found"),
			wantErr:     true,
			errMsg:      "getting node",
		},
		{
			name:        "fails when execute shell fails",
			path:        "/test/repo",
			initialized: true,
			executeErr:  fmt.Errorf("command failed"),
			wantErr:     true,
			errMsg:      "starting Claude",
		},
		{
			name:        "fails when exit code is non-zero",
			path:        "/test/repo",
			initialized: true,
			exitCode:    1,
			wantErr:     true,
			errMsg:      "Claude exited with code 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockTree := mocks.NewMockTreeProvider()
			mockProcess := &MockProcessProviderWithResult{
				executeErr: tt.executeErr,
				exitCode:   tt.exitCode,
			}
			mockUI := mocks.NewMockUIProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockConfig := mocks.NewMockConfigProvider()
			mockGit := mocks.NewMockGitProvider()

			// Setup tree mock behavior
			if tt.getCurrentErr != nil {
				mockTree.SetError("GetCurrent", "", tt.getCurrentErr)
			} else {
				mockTree.SetCurrent(interfaces.NodeInfo{
					Path: "/current/node",
					Name: "current",
				})
			}

			// Only set up GetNode if we're not expecting GetCurrent to fail
			if tt.getCurrentErr == nil {
				nodePath := tt.path
				if nodePath == "" {
					nodePath = "/current/node"  // Use the current node path when no path specified
				}
				if tt.getNodeErr != nil {
					mockTree.SetError("GetNode", nodePath, tt.getNodeErr)
				} else {
					mockTree.SetNode(nodePath, interfaces.NodeInfo{
						Path: nodePath,
						Name: "test-node",
					})
				}
			}

			m := &Manager{
				initialized:     tt.initialized,
				workspace:       "/workspace",
				treeProvider:    mockTree,
				processProvider: mockProcess,
				uiProvider:      mockUI,
				fsProvider:      mockFS,
				configProvider:  mockConfig,
				gitProvider:     mockGit,
				logProvider:     NewDefaultLogProvider(false),
				metricsProvider: NewNoOpMetricsProvider(),
			}

			err := m.StartClaude(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// MockProcessProviderWithResult is a mock that can control ExecuteShell results
type MockProcessProviderWithResult struct {
	DefaultProcessProvider
	executeErr error
	exitCode   int
}

func (m *MockProcessProviderWithResult) ExecuteShell(ctx context.Context, command string, opts interfaces.ProcessOptions) (*interfaces.ProcessResult, error) {
	if m.executeErr != nil {
		return nil, m.executeErr
	}
	return &interfaces.ProcessResult{
		ExitCode: m.exitCode,
		Stdout:   "",
		Stderr:   "error output",
	}, nil
}

// TestManager_Close tests the Close function
func TestManager_Close_Extended(t *testing.T) {
	tests := []struct {
		name           string
		hasPlugins     bool
		pluginError    error
		metricsError   error
		wantErr        bool
	}{
		{
			name:    "successful close without plugins",
			wantErr: false,
		},
		{
			name:         "close with plugin unload error",
			hasPlugins:   true,
			pluginError:  fmt.Errorf("plugin error"),
			wantErr:      false, // Close doesn't return plugin errors
		},
		{
			name:         "close with metrics flush error",
			metricsError: fmt.Errorf("metrics error"),
			wantErr:      false, // Close doesn't return metrics errors
		},
		{
			name:         "close with both errors",
			hasPlugins:   true,
			pluginError:  fmt.Errorf("plugin error"),
			metricsError: fmt.Errorf("metrics error"),
			wantErr:      false,
		},
		{
			name:    "close with nil providers",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricsProvider := &MockMetricsProviderWithError{
				flushError: tt.metricsError,
			}

			m := &Manager{
				logProvider:     NewDefaultLogProvider(false),
				metricsProvider: metricsProvider,
			}
			
			if tt.hasPlugins {
				m.pluginManager = &MockPluginManager{
					plugins:     []interfaces.PluginMetadata{{Name: "test-plugin"}},
					unloadError: tt.pluginError,
				}
			}

			err := m.Close()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Test with nil providers
	t.Run("nil providers", func(t *testing.T) {
		m := &Manager{
			logProvider:     nil,
			metricsProvider: nil,
			pluginManager:   nil,
		}
		err := m.Close()
		assert.NoError(t, err)
	})
}

// MockPluginManager for testing plugin operations
type MockPluginManager struct {
	plugins     []interfaces.PluginMetadata
	unloadError error
}

func (m *MockPluginManager) DiscoverPlugins(ctx context.Context) ([]interfaces.PluginMetadata, error) {
	return m.plugins, nil
}

func (m *MockPluginManager) LoadPlugin(ctx context.Context, name string) error {
	return nil
}

func (m *MockPluginManager) UnloadPlugin(ctx context.Context, name string) error {
	return m.unloadError
}

func (m *MockPluginManager) ExecuteCommand(ctx context.Context, command string, args []string) (interfaces.Result, error) {
	return interfaces.Result{Success: true}, nil
}

func (m *MockPluginManager) ListPlugins() []interfaces.PluginMetadata {
	return m.plugins
}

func (m *MockPluginManager) GetPlugin(name string) (interfaces.Plugin, error) {
	for _, p := range m.plugins {
		if p.Name == name {
			return nil, nil // Return nil plugin for testing
		}
	}
	return nil, fmt.Errorf("plugin not found")
}

func (m *MockPluginManager) ReloadPlugin(ctx context.Context, name string) error {
	return nil
}

func (m *MockPluginManager) IsLoaded(name string) bool {
	for _, p := range m.plugins {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (m *MockPluginManager) GetCommand(name string) (*interfaces.CommandDefinition, interfaces.Plugin, error) {
	return nil, nil, fmt.Errorf("command not found")
}

func (m *MockPluginManager) ListCommands() []interfaces.CommandDefinition {
	return []interfaces.CommandDefinition{}
}

func (m *MockPluginManager) InstallPlugin(ctx context.Context, source string) error {
	return nil
}

func (m *MockPluginManager) UpdatePlugin(ctx context.Context, name string) error {
	return nil
}

func (m *MockPluginManager) RemovePlugin(ctx context.Context, name string) error {
	return nil
}

func (m *MockPluginManager) GetPluginConfig(name string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (m *MockPluginManager) SetPluginConfig(name string, config map[string]interface{}) error {
	return nil
}

func (m *MockPluginManager) HealthCheck(ctx context.Context) map[string]error {
	return map[string]error{}
}

// MockMetricsProviderWithError for testing metrics operations
type MockMetricsProviderWithError struct {
	NoOpMetricsProvider
	flushError error
}

func (m *MockMetricsProviderWithError) Flush() error {
	return m.flushError
}

// TestManager_ClearCurrent tests the ClearCurrent function
func TestManager_ClearCurrent(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
		setPathErr  error
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "successful clear",
			initialized: true,
			wantErr:     false,
		},
		{
			name:        "fails when not initialized",
			initialized: false,
			wantErr:     true,
			errMsg:      "manager not initialized",
		},
		{
			name:        "fails when SetPath fails",
			initialized: true,
			setPathErr:  fmt.Errorf("set path error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTree := mocks.NewMockTreeProvider()
			if tt.setPathErr != nil {
				mockTree.SetError("SetPath", "/", tt.setPathErr)
			}

			m := &Manager{
				initialized:  tt.initialized,
				treeProvider: mockTree,
			}

			err := m.ClearCurrent()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestManager_InitializeWithConfig tests the InitializeWithConfig function
func TestManager_InitializeWithConfig_Extended(t *testing.T) {
	tests := []struct {
		name          string
		workspace     string
		cfg           *config.ConfigTree
		loadTreeError error
		saveError     error
		wantErr       bool
		errMsg        string
	}{
		{
			name:      "successful initialization",
			workspace: "/test/workspace",
			cfg: &config.ConfigTree{
				Workspace: config.WorkspaceTree{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name:      "fails when Load tree fails",
			workspace: "/test/workspace",
			cfg:       &config.ConfigTree{},
			loadTreeError: fmt.Errorf("load tree error"),
			wantErr:       true,
			errMsg:        "failed to load tree",
		},
		{
			name:      "fails when Save config fails",
			workspace: "/test/workspace",
			cfg:       &config.ConfigTree{},
			saveError: fmt.Errorf("save error"),
			wantErr:   true,
			errMsg:    "failed to save config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTree := mocks.NewMockTreeProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockConfig := mocks.NewMockConfigProvider()
			mockGit := mocks.NewMockGitProvider()
			mockUI := mocks.NewMockUIProvider()

			if tt.loadTreeError != nil {
				mockTree.SetError("Load", "", tt.loadTreeError)
			}
			if tt.saveError != nil {
				mockConfig.SetError("save", tt.saveError)
			}

			// Setup filesystem mock
			mockFS.SetExists(tt.workspace, true)

			m := &Manager{
				treeProvider:    mockTree,
				fsProvider:      mockFS,
				configProvider:  mockConfig,
				gitProvider:     mockGit,
				uiProvider:      mockUI,
				logProvider:     NewDefaultLogProvider(false),
				metricsProvider: NewNoOpMetricsProvider(),
				opts: ManagerOptions{
					AutoLoadConfig: false,
				},
			}

			ctx := context.Background()
			err := m.InitializeWithConfig(ctx, tt.workspace, tt.cfg)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.cfg, m.config)
			}
		})
	}
}

// TestManager_SmartInitWorkspace tests the SmartInitWorkspace function
func TestManager_SmartInitWorkspace_Extended(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		options     InitOptions
		findRepos   []GitRepoInfo
		findError   error
		mkdirError  error
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "successful initialization with no repos",
			projectName: "test-project",
			options:     InitOptions{},
			findRepos:   []GitRepoInfo{},
			wantErr:     false,
		},
		{
			name:        "successful initialization with repos",
			projectName: "test-project",
			options:     InitOptions{NonInteractive: true},
			findRepos: []GitRepoInfo{
				{
					Path:      "repos/existing",
					RemoteURL: "https://github.com/test/repo.git",
					Branch:    "main",
				},
			},
			wantErr: false,
		},
		{
			name:        "fails when scanning repositories fails without force",
			projectName: "test-project",
			options:     InitOptions{Force: false},
			findError:   fmt.Errorf("scan error"),
			wantErr:     true,
			errMsg:      "scanning repositories",
		},
		{
			name:        "continues when scanning fails with force",
			projectName: "test-project",
			options:     InitOptions{Force: true},
			findError:   fmt.Errorf("scan error"),
			wantErr:     false,
		},
		{
			name:        "fails when creating repos directory fails",
			projectName: "test-project",
			options:     InitOptions{},
			mkdirError:  fmt.Errorf("mkdir error"),
			wantErr:     true,
			errMsg:      "creating repos directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			mockTree := mocks.NewMockTreeProvider()
			mockFS := &MockFSProviderWithWalk{
				MockFileSystemProvider: *mocks.NewMockFileSystemProvider(),
				repos:                  tt.findRepos,
				walkError:              tt.findError,
			}
			mockConfig := mocks.NewMockConfigProvider()
			mockGit := &MockGitProviderWithRemote{
				MockGitProvider: *mocks.NewMockGitProvider(),
			}
			mockUI := mocks.NewMockUIProvider()

			// Setup filesystem mock
			if tt.mkdirError != nil {
				mockFS.mkdirError = tt.mkdirError
			}
			mockFS.SetExists(filepath.Join(tmpDir, "CLAUDE.md"), false)

			// Setup UI mock for confirmations
			mockUI.SetConfirmResponse(true)

			m := &Manager{
				workspace:       tmpDir,
				initialized:     false,
				treeProvider:    mockTree,
				fsProvider:      mockFS,
				configProvider:  mockConfig,
				gitProvider:     mockGit,
				uiProvider:      mockUI,
				logProvider:     NewDefaultLogProvider(false),
				metricsProvider: NewNoOpMetricsProvider(),
			}

			err := m.SmartInitWorkspace(tt.projectName, tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, m.initialized)
				assert.NotNil(t, m.config)
				assert.Equal(t, tt.projectName, m.config.Workspace.Name)
			}
		})
	}
}

// MockFSProviderWithWalk extends the mock filesystem provider with Walk support
type MockFSProviderWithWalk struct {
	mocks.MockFileSystemProvider
	repos      []GitRepoInfo
	walkError  error
	mkdirError error
}

func (m *MockFSProviderWithWalk) MkdirAll(path string, perm os.FileMode) error {
	if m.mkdirError != nil {
		return m.mkdirError
	}
	return m.MockFileSystemProvider.MkdirAll(path, perm)
}

func (m *MockFSProviderWithWalk) Walk(root string, fn filepath.WalkFunc) error {
	if m.walkError != nil {
		return m.walkError
	}
	
	// Simulate finding repos
	for _, repo := range m.repos {
		gitPath := filepath.Join(repo.Path, ".git")
		info := &mockFileInfoExt{
			name:  ".git",
			isDir: true,
		}
		if err := fn(gitPath, info, nil); err != nil {
			// filepath.SkipDir is not an error, it's a control flow signal
			if err == filepath.SkipDir {
				continue
			}
			return err
		}
	}
	
	return nil
}

type mockFileInfoExt struct {
	name  string
	isDir bool
}

func (m *mockFileInfoExt) Name() string       { return m.name }
func (m *mockFileInfoExt) Size() int64        { return 0 }
func (m *mockFileInfoExt) Mode() os.FileMode  { return 0755 }
func (m *mockFileInfoExt) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfoExt) IsDir() bool        { return m.isDir }
func (m *mockFileInfoExt) Sys() interface{}   { return nil }

// MockGitProviderWithRemote extends the mock git provider with remote URL support
type MockGitProviderWithRemote struct {
	mocks.MockGitProvider
}

func (m *MockGitProviderWithRemote) GetRemoteURL(path string) (string, error) {
	return "https://github.com/test/repo.git", nil
}

func (m *MockGitProviderWithRemote) Branch(path string) (string, error) {
	return "main", nil
}