package manager

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/repo-claude/internal/config"
	"github.com/taokim/repo-claude/internal/scope"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	
	m, err := New(tmpDir)
	require.NoError(t, err)
	assert.NotNil(t, m)
	assert.NotEmpty(t, m.ProjectPath)
	assert.NotNil(t, m.CmdExecutor)
}

func TestLoadFromCurrentDir(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Create a valid config
		cfg := config.DefaultConfig("test-project")
		configPath := filepath.Join(tmpDir, "repo-claude.yaml")
		err := cfg.Save(configPath)
		require.NoError(t, err)
		
		// Create workspaces directory
		err = os.MkdirAll(filepath.Join(tmpDir, "workspaces"), 0755)
		require.NoError(t, err)
		
		// Create docs directory
		err = os.MkdirAll(filepath.Join(tmpDir, "docs"), 0755)
		require.NoError(t, err)
		
		// Change to temp directory
		oldCwd, _ := os.Getwd()
		defer os.Chdir(oldCwd)
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		
		// Load from current directory
		m, err := LoadFromCurrentDir()
		require.NoError(t, err)
		assert.NotNil(t, m)
		assert.NotNil(t, m.Config)
		assert.NotNil(t, m.ScopeManager)
		assert.NotNil(t, m.DocsManager)
		assert.Equal(t, tmpDir, m.ProjectPath)
	})
	
	t.Run("NoConfig", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Change to temp directory without config
		oldCwd, _ := os.Getwd()
		defer os.Chdir(oldCwd)
		err := os.Chdir(tmpDir)
		require.NoError(t, err)
		
		// Should fail
		_, err = LoadFromCurrentDir()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no repo-claude.yaml")
	})
}

func TestInitWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "test-init"
	projectPath := filepath.Join(tmpDir, projectName)
	
	m := &Manager{
		ProjectPath: projectPath,
		CmdExecutor: &MockCommandExecutor{},
	}
	
	err := m.InitWorkspace(projectName, false)
	require.NoError(t, err)
	
	// Check created files and directories
	assert.DirExists(t, projectPath)
	assert.FileExists(t, filepath.Join(projectPath, "repo-claude.yaml"))
	assert.FileExists(t, filepath.Join(projectPath, "CLAUDE.md"))
	assert.DirExists(t, filepath.Join(projectPath, "workspaces"))
	assert.DirExists(t, filepath.Join(projectPath, "docs"))
	
	// Check config was saved properly
	assert.NotNil(t, m.Config)
	assert.Equal(t, projectName, m.Config.Workspace.Name)
	assert.NotNil(t, m.ScopeManager)
	assert.NotNil(t, m.DocsManager)
}

func TestListScopes(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a manager with config
	m := &Manager{
		ProjectPath: tmpDir,
		Config: &config.Config{
			Workspace: config.WorkspaceConfig{
				Name:     "test",
				BasePath: "workspaces",
			},
			Repositories: map[string]config.Repository{
				"repo1": {URL: "https://example.com/repo1.git"},
				"repo2": {URL: "https://example.com/repo2.git"},
			},
			Scopes: map[string]config.Scope{
				"scope1": {
					Type:        "persistent",
					Repos:       []string{"repo1"},
					Description: "Test scope 1",
				},
				"scope2": {
					Type:        "ephemeral",
					Repos:       []string{"repo2"},
					Description: "Test scope 2",
				},
			},
		},
		CmdExecutor: &MockCommandExecutor{},
	}
	
	// Create scope manager
	scopeMgr, err := scope.NewManager(m.Config, tmpDir)
	require.NoError(t, err)
	m.ScopeManager = scopeMgr
	
	// List scopes should work without error
	err = m.ListScopes()
	assert.NoError(t, err)
}

func TestStartScope(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create workspaces directory
	workspacesDir := filepath.Join(tmpDir, "workspaces")
	err := os.MkdirAll(workspacesDir, 0755)
	require.NoError(t, err)
	
	// Create a manager with config
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Name:     "test",
			BasePath: "workspaces",
		},
		Repositories: map[string]config.Repository{
			"repo1": {URL: "https://example.com/repo1.git"},
		},
		Scopes: map[string]config.Scope{
			"test-scope": {
				Type:        "persistent",
				Repos:       []string{"repo1"},
				Description: "Test scope",
				Model:       "claude-3-sonnet",
			},
		},
	}
	
	m := &Manager{
		ProjectPath: tmpDir,
		Config:      cfg,
		CmdExecutor: &MockCommandExecutor{},
	}
	
	// Create scope manager
	scopeMgr, err := scope.NewManager(cfg, tmpDir)
	require.NoError(t, err)
	m.ScopeManager = scopeMgr
	
	// Create the scope directory
	scopePath := filepath.Join(workspacesDir, "test-scope")
	err = os.MkdirAll(scopePath, 0755)
	require.NoError(t, err)
	
	// Start scope
	err = m.StartScope("test-scope", false)
	// Will fail because claude command doesn't exist, but that's expected
	// Just verify no panic and proper error handling
	assert.Error(t, err)
}

func TestPullScope(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Name:     "test",
			BasePath: "workspaces",
		},
		Repositories: map[string]config.Repository{
			"repo1": {URL: "https://example.com/repo1.git"},
		},
		Scopes: map[string]config.Scope{
			"test-scope": {
				Type:  "persistent",
				Repos: []string{"repo1"},
			},
		},
	}
	
	m := &Manager{
		ProjectPath: tmpDir,
		Config:      cfg,
		CmdExecutor: &MockCommandExecutor{},
	}
	
	// Create scope manager
	scopeMgr, err := scope.NewManager(cfg, tmpDir)
	require.NoError(t, err)
	m.ScopeManager = scopeMgr
	
	// Pull will fail because scope doesn't exist, but should handle gracefully
	err = m.PullScope("test-scope", true)
	assert.Error(t, err)
}

func TestStatusScope(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create workspaces directory
	workspacesDir := filepath.Join(tmpDir, "workspaces")
	err := os.MkdirAll(workspacesDir, 0755)
	require.NoError(t, err)
	
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Name:     "test",
			BasePath: "workspaces",
		},
		Repositories: map[string]config.Repository{
			"repo1": {URL: "https://example.com/repo1.git"},
		},
		Scopes: map[string]config.Scope{
			"test-scope": {
				Type:  "persistent",
				Repos: []string{"repo1"},
			},
		},
	}
	
	m := &Manager{
		ProjectPath: tmpDir,
		Config:      cfg,
		CmdExecutor: &MockCommandExecutor{},
	}
	
	// Create scope manager
	scopeMgr, err := scope.NewManager(cfg, tmpDir)
	require.NoError(t, err)
	m.ScopeManager = scopeMgr
	
	// Status will fail because scope doesn't exist, but should handle gracefully
	err = m.StatusScope("test-scope")
	assert.Error(t, err)
}

func TestCreateRootCLAUDE(t *testing.T) {
	tmpDir := t.TempDir()
	
	m := &Manager{
		ProjectPath: tmpDir,
		Config: &config.Config{
			Workspace: config.WorkspaceConfig{
				BasePath: "workspaces",
			},
		},
	}
	
	err := m.createRootCLAUDE()
	require.NoError(t, err)
	
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	assert.FileExists(t, claudePath)
	
	content, err := os.ReadFile(claudePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "CLAUDE.md - Root Level")
	assert.Contains(t, string(content), "repo-claude v2")
	assert.Contains(t, string(content), "scope-based isolation")
}

// MockCommandExecutor for testing
type MockCommandExecutor struct{}

func (m *MockCommandExecutor) Command(name string, args ...string) Cmd {
	return &MockCmd{name: name, args: args}
}

type MockCmd struct {
	name string
	args []string
	dir  string
	env  []string
}

func (m *MockCmd) Output() ([]byte, error) {
	return []byte("mock output"), nil
}

func (m *MockCmd) Run() error {
	return nil
}

func (m *MockCmd) Start() error {
	return nil
}

func (m *MockCmd) Wait() error {
	return nil
}

func (m *MockCmd) StdoutPipe() (io.ReadCloser, error) {
	return nil, nil
}

func (m *MockCmd) StderrPipe() (io.ReadCloser, error) {
	return nil, nil
}

func (m *MockCmd) SetDir(dir string) {
	m.dir = dir
}

func (m *MockCmd) SetEnv(env []string) {
	m.env = env
}

func (m *MockCmd) Process() *os.Process {
	return nil
}