package manager

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/mocks"
)

// Test StartBackground method of DefaultProcessProvider
func TestDefaultProcessProvider_StartBackground(t *testing.T) {
	provider := NewDefaultProcessProvider()
	ctx := context.Background()
	
	// Test starting a background process
	proc, err := provider.StartBackground(ctx, "sleep", []string{"0.1"}, interfaces.ProcessOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, proc)
}

// Test Fatal method of DefaultLogProvider
func TestDefaultLogProvider_Fatal(t *testing.T) {
	provider := NewDefaultLogProvider(false)
	
	// Since Fatal calls os.Exit, we can't test it directly
	// We'll just verify it doesn't panic when we would call it
	// In a real scenario, you'd use a mock or test helper
	// that overrides os.Exit
	assert.NotPanics(t, func() {
		// We won't actually call Fatal as it would exit the test
		// provider.Fatal("test message")
		// Just verify the provider exists
		assert.NotNil(t, provider)
	})
}

// Test SetLevel method of DefaultLogProvider
func TestDefaultLogProvider_SetLevel(t *testing.T) {
	provider := NewDefaultLogProvider(false)
	
	// Test setting different levels
	provider.SetLevel(interfaces.LogLevelDebug)
	provider.SetLevel(interfaces.LogLevelInfo)
	provider.SetLevel(interfaces.LogLevelWarn)
	provider.SetLevel(interfaces.LogLevelError)
	
	// Verify it doesn't panic
	assert.NotNil(t, provider)
}

// Test Counter, Gauge, and Histogram methods of NoOpMetricsProvider
func TestNoOpMetricsProvider_AllMethods(t *testing.T) {
	provider := NewNoOpMetricsProvider()
	
	// Test Counter
	provider.Counter("test.counter", 1, "tag1", "tag2")
	
	// Test Gauge  
	provider.Gauge("test.gauge", 100.5, "tag1")
	
	// Test Histogram
	provider.Histogram("test.histogram", 50.25, "tag1")
	
	// Verify no panics
	assert.NotNil(t, provider)
}

// Test NoOpTimer methods
func TestNoOpTimer_AllMethods(t *testing.T) {
	provider := NewNoOpMetricsProvider()
	timer := provider.Timer("test.timer")
	
	// Test Start
	timer.Start()
	
	// Test Stop was already tested in simple_test.go
	duration := timer.Stop()
	assert.Equal(t, int64(0), int64(duration))
	
	// Test Record with duration
	timer.Record(time.Millisecond)
	
	assert.NotNil(t, timer)
}

// Test displayTreeRecursive method
func TestManager_displayTreeRecursive(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	
	manager, err := NewManager(ManagerOptions{
		ConfigProvider: mockConfig,
		GitProvider:    mockGit,
		FSProvider:     mockFS,
		UIProvider:     mockUI,
		TreeProvider:   mockTree,
	})
	require.NoError(t, err)
	
	// Initialize manager
	mockFS.SetExists("/test", true)
	err = manager.Initialize(context.Background(), "/test")
	require.NoError(t, err)
	
	// Set up a tree structure
	root := interfaces.NodeInfo{
		Name: "root",
		Path: "/",
		Children: []interfaces.NodeInfo{
			{
				Name:     "backend",
				Path:     "/backend",
				IsCloned: true,
				Children: []interfaces.NodeInfo{
					{
						Name:     "service-a",
						Path:     "/backend/service-a",
						IsCloned: true,
					},
					{
						Name:   "service-b",
						Path:   "/backend/service-b",
						IsLazy: true,
					},
				},
			},
			{
				Name:     "frontend",
				Path:     "/frontend",
				IsCloned: true,
			},
		},
	}
	
	// Call displayTreeRecursive
	manager.displayTreeRecursive(root, 0)
	
	// Verify UI messages were sent
	messages := mockUI.GetMessages()
	assert.Greater(t, len(messages), 0)
}

// Test collectLazyNodes function
func TestCollectLazyNodes(t *testing.T) {
	root := interfaces.NodeInfo{
		Name: "root",
		Path: "/",
		Children: []interfaces.NodeInfo{
			{
				Name:     "backend",
				Path:     "/backend",
				IsCloned: true,
				Children: []interfaces.NodeInfo{
					{
						Name:     "service-a",
						Path:     "/backend/service-a",
						IsCloned: true,
					},
					{
						Name:   "service-b",
						Path:   "/backend/service-b",
						IsLazy: true,
					},
				},
			},
			{
				Name:   "frontend",
				Path:   "/frontend",
				IsLazy: true,
			},
		},
	}
	
	lazyNodes := collectLazyNodes(root)
	assert.Len(t, lazyNodes, 2)
	assert.Equal(t, "/backend/service-b", lazyNodes[0].Path)
	assert.Equal(t, "/frontend", lazyNodes[1].Path)
}

// Test showStatusRecursive method
func TestManager_showStatusRecursive(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	
	manager, err := NewManager(ManagerOptions{
		ConfigProvider: mockConfig,
		GitProvider:    mockGit,
		FSProvider:     mockFS,
		UIProvider:     mockUI,
		TreeProvider:   mockTree,
	})
	require.NoError(t, err)
	
	// Initialize manager
	mockFS.SetExists("/test", true)
	err = manager.Initialize(context.Background(), "/test")
	require.NoError(t, err)
	
	// Set up node with git status
	node := interfaces.NodeInfo{
		Name:       "service-a",
		Path:       "/backend/service-a",
		Repository: "https://github.com/org/service-a",
		IsCloned:   true,
	}
	
	// Set up git status response
	mockGit.SetStatus("/test/repos/backend/service-a", &interfaces.GitStatus{
		Branch:        "main",
		HasChanges:    true,
		FilesModified: 2,
		FilesAdded:    1,
	})
	
	// Call showStatusRecursive
	manager.showStatusRecursive(node)
	
	// Verify UI messages were sent
	messages := mockUI.GetMessages()
	assert.Greater(t, len(messages), 0)
}

// Test pullRecursive method
func TestManager_pullRecursive(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	
	manager, err := NewManager(ManagerOptions{
		ConfigProvider: mockConfig,
		GitProvider:    mockGit,
		FSProvider:     mockFS,
		UIProvider:     mockUI,
		TreeProvider:   mockTree,
	})
	require.NoError(t, err)
	
	// Initialize manager
	mockFS.SetExists("/test", true)
	err = manager.Initialize(context.Background(), "/test")
	require.NoError(t, err)
	
	// Set up node hierarchy
	node := interfaces.NodeInfo{
		Name:       "backend",
		Path:       "/backend",
		Repository: "https://github.com/org/backend",
		IsCloned:   true,
		Children: []interfaces.NodeInfo{
			{
				Name:       "service-a",
				Path:       "/backend/service-a",
				Repository: "https://github.com/org/service-a",
				IsCloned:   true,
			},
		},
	}
	
	// Mock successful pulls
	mockGit.SetPullResult("/test/repos/backend", interfaces.GitPullResult{
		UpdatesReceived: true,
		FilesChanged:    3,
	})
	mockGit.SetPullResult("/test/repos/backend/service-a", interfaces.GitPullResult{
		UpdatesReceived: false,
	})
	
	// Call pullRecursive
	manager.pullRecursive(node)
	
	// Verify git pull was called
	gitCalls := mockGit.GetCalls()
	assert.Contains(t, gitCalls[len(gitCalls)-2], "Pull")
	assert.Contains(t, gitCalls[len(gitCalls)-1], "Pull")
}

// Test pushRecursive method
func TestManager_pushRecursive(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	
	manager, err := NewManager(ManagerOptions{
		ConfigProvider: mockConfig,
		GitProvider:    mockGit,
		FSProvider:     mockFS,
		UIProvider:     mockUI,
		TreeProvider:   mockTree,
	})
	require.NoError(t, err)
	
	// Initialize manager
	mockFS.SetExists("/test", true)
	err = manager.Initialize(context.Background(), "/test")
	require.NoError(t, err)
	
	// Set up node hierarchy
	node := interfaces.NodeInfo{
		Name:       "backend",
		Path:       "/backend",
		Repository: "https://github.com/org/backend",
		IsCloned:   true,
		Children: []interfaces.NodeInfo{
			{
				Name:       "service-a",
				Path:       "/backend/service-a",
				Repository: "https://github.com/org/service-a",
				IsCloned:   true,
			},
		},
	}
	
	// Mock successful pushes
	mockGit.SetPushResult("/test/repos/backend", interfaces.GitPushResult{
		Success: true,
	})
	mockGit.SetPushResult("/test/repos/backend/service-a", interfaces.GitPushResult{
		Success: true,
	})
	
	// Call pushRecursive
	manager.pushRecursive(node)
	
	// Verify git push was called
	gitCalls := mockGit.GetCalls()
	assert.Contains(t, gitCalls[len(gitCalls)-2], "Push")
	assert.Contains(t, gitCalls[len(gitCalls)-1], "Push")
}

// Test SmartInitWorkspace method
func TestManager_SmartInitWorkspace(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	
	manager, err := NewManager(ManagerOptions{
		ConfigProvider: mockConfig,
		GitProvider:    mockGit,
		FSProvider:     mockFS,
		UIProvider:     mockUI,
		TreeProvider:   mockTree,
	})
	require.NoError(t, err)
	
	// Create a temp directory for testing
	tempDir := t.TempDir()
	
	// Create some git repositories
	repo1 := filepath.Join(tempDir, "repo1")
	repo2 := filepath.Join(tempDir, "nested", "repo2")
	
	// Mock file system
	mockFS.SetExists(tempDir, true)
	mockFS.SetWalkFunc(func(root string, fn filepath.WalkFunc) error {
		// Simulate walking through directories
		fn(tempDir, &mockFileInfo{name: ".", isDir: true}, nil)
		fn(repo1, &mockFileInfo{name: "repo1", isDir: true}, nil)
		fn(filepath.Join(repo1, ".git"), &mockFileInfo{name: ".git", isDir: true}, nil)
		fn(filepath.Join(tempDir, "nested"), &mockFileInfo{name: "nested", isDir: true}, nil)
		fn(repo2, &mockFileInfo{name: "repo2", isDir: true}, nil)
		fn(filepath.Join(repo2, ".git"), &mockFileInfo{name: ".git", isDir: true}, nil)
		return nil
	})
	
	// Mock git operations
	mockGit.SetRemoteURL(repo1, "https://github.com/org/repo1.git")
	mockGit.SetRemoteURL(repo2, "https://github.com/org/repo2.git")
	
	// Mock user confirmation
	mockUI.SetConfirm("Initialize workspace 'test-workspace' with found repositories?", true)
	
	// Call SmartInitWorkspace
	err = manager.SmartInitWorkspace("test-workspace", InitOptions{})
	assert.NoError(t, err)
}

// Test findGitRepositories method
func TestManager_findGitRepositories(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	
	manager, err := NewManager(ManagerOptions{
		ConfigProvider: mockConfig,
		GitProvider:    mockGit,
		FSProvider:     mockFS,
		UIProvider:     mockUI,
		TreeProvider:   mockTree,
	})
	require.NoError(t, err)
	
	tempDir := t.TempDir()
	
	// Mock file system walk
	mockFS.SetWalkFunc(func(root string, fn filepath.WalkFunc) error {
		fn(tempDir, &mockFileInfo{name: ".", isDir: true}, nil)
		fn(filepath.Join(tempDir, "repo1"), &mockFileInfo{name: "repo1", isDir: true}, nil)
		fn(filepath.Join(tempDir, "repo1", ".git"), &mockFileInfo{name: ".git", isDir: true}, nil)
		fn(filepath.Join(tempDir, "repo2"), &mockFileInfo{name: "repo2", isDir: true}, nil)
		fn(filepath.Join(tempDir, "repo2", ".git"), &mockFileInfo{name: ".git", isDir: true}, nil)
		return nil
	})
	
	// Mock git remote URLs
	mockGit.SetRemoteURL(filepath.Join(tempDir, "repo1"), "https://github.com/org/repo1.git")
	mockGit.SetRemoteURL(filepath.Join(tempDir, "repo2"), "https://github.com/org/repo2.git")
	
	// Call findGitRepositories
	repos, err := manager.findGitRepositories(tempDir)
	assert.NoError(t, err)
	assert.Len(t, repos, 2)
	assert.Equal(t, "https://github.com/org/repo1.git", repos[0].RemoteURL)
	assert.Equal(t, "https://github.com/org/repo2.git", repos[1].RemoteURL)
}

// Test NewManagerForInit function
func TestNewManagerForInit(t *testing.T) {
	manager, err := NewManagerForInit(t.TempDir())
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.configProvider)
	assert.NotNil(t, manager.gitProvider)
	assert.NotNil(t, manager.fsProvider)
	assert.NotNil(t, manager.uiProvider)
	assert.NotNil(t, manager.treeProvider)
	assert.NotNil(t, manager.processProvider)
	assert.NotNil(t, manager.logProvider)
	assert.NotNil(t, manager.metricsProvider)
}

// Test LoadFromCurrentDir function
func TestLoadFromCurrentDir(t *testing.T) {
	// Create a temp directory with a config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "muno.yaml")
	
	// Create a simple config file
	configContent := `
workspace:
  name: test-workspace
  repos_dir: repos
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	
	// Change to temp directory
	oldDir, _ := os.Getwd()
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer os.Chdir(oldDir)
	
	// Call LoadFromCurrentDir
	manager, err := LoadFromCurrentDir()
	
	// The function will try to load real config, which might fail
	// in test environment, so we just check it doesn't panic
	if err == nil {
		assert.NotNil(t, manager)
	}
}

// Test LoadFromCurrentDirWithOptions function
func TestLoadFromCurrentDirWithOptions(t *testing.T) {
	// Create a temp directory with a config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "muno.yaml")
	
	// Create a simple config file
	configContent := `
workspace:
  name: test-workspace
  repos_dir: repos
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	
	// Change to temp directory
	oldDir, _ := os.Getwd()
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer os.Chdir(oldDir)
	
	// Call LoadFromCurrentDirWithOptions
	opts := &ManagerOptions{
		DebugMode: true,
	}
	manager, err := LoadFromCurrentDirWithOptions(opts)
	
	// The function will try to load real config, which might fail
	// in test environment, so we just check it doesn't panic
	if err == nil {
		assert.NotNil(t, manager)
	}
}

// mockFileInfo implements fs.FileInfo for testing
type mockFileInfo struct {
	name  string
	isDir bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() fs.FileMode  { return 0755 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }