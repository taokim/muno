package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/adapters"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
)

// TestWorkspace provides a test workspace with helper methods
type TestWorkspace struct {
	t         *testing.T
	Root      string
	NodesDir  string
	ConfigPath string
}

// CreateTestWorkspace creates a temporary workspace for testing
func CreateTestWorkspace(t *testing.T) *TestWorkspace {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	require.NoError(t, os.MkdirAll(workspace, 0755))
	
	nodesDir := filepath.Join(workspace, ".nodes")
	require.NoError(t, os.MkdirAll(nodesDir, 0755))
	
	return &TestWorkspace{
		t:          t,
		Root:       workspace,
		NodesDir:   nodesDir,
		ConfigPath: filepath.Join(workspace, "muno.yaml"),
	}
}

// AddRepository adds a repository directory to the workspace
func (tw *TestWorkspace) AddRepository(path string) string {
	fullPath := filepath.Join(tw.NodesDir, path)
	require.NoError(tw.t, os.MkdirAll(fullPath, 0755))
	return fullPath
}

// AddRepositoryWithConfig adds a repository with a muno.yaml file
func (tw *TestWorkspace) AddRepositoryWithConfig(path string, cfg *config.ConfigTree) string {
	fullPath := tw.AddRepository(path)
	configPath := filepath.Join(fullPath, "muno.yaml")
	require.NoError(tw.t, cfg.Save(configPath))
	return fullPath
}

// CreateConfig creates a muno.yaml in the workspace root
func (tw *TestWorkspace) CreateConfig(cfg *config.ConfigTree) {
	require.NoError(tw.t, cfg.Save(tw.ConfigPath))
}

// CreateConfigReference creates a config reference file
func (tw *TestWorkspace) CreateConfigReference(path string, cfg *config.ConfigTree) string {
	fullPath := filepath.Join(tw.Root, path)
	dir := filepath.Dir(fullPath)
	require.NoError(tw.t, os.MkdirAll(dir, 0755))
	require.NoError(tw.t, cfg.Save(fullPath))
	return fullPath
}

// CreateTestManagerOptions creates manager options for testing
func CreateTestManagerOptions(workspace string) ManagerOptions {
	// Create actual providers using adapters
	configAdapter := adapters.NewConfigAdapter()
	gitProvider := adapters.NewGitProvider()
	fsAdapter := adapters.NewFileSystemAdapter()
	treeAdapter := adapters.NewTreeAdapter() // No arguments needed
	uiAdapter := adapters.NewUIAdapter()
	
	return ManagerOptions{
		ConfigProvider: configAdapter,
		GitProvider:    gitProvider,
		FSProvider:     fsAdapter,
		TreeProvider:   treeAdapter,
		UIProvider:     uiAdapter,
	}
}

// CreateTestManager creates a manager for testing with minimal setup
func CreateTestManager(t *testing.T, workspace string) *Manager {
	opts := CreateTestManagerOptions(workspace)
	m, err := NewManager(opts)
	require.NoError(t, err)
	m.workspace = workspace
	m.initialized = true
	
	// Set a basic config if not already set
	if m.config == nil {
		m.config = &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: ".nodes",
			},
		}
	}
	
	return m
}

// CreateTestManagerWithConfig creates a manager with specific config
func CreateTestManagerWithConfig(t *testing.T, workspace string, cfg *config.ConfigTree) *Manager {
	m := CreateTestManager(t, workspace)
	m.config = cfg
	// Tree provider is already set up from CreateTestManager
	return m
}

// AddNodeToTree adds a node to the manager's tree provider
func AddNodeToTree(m *Manager, path string, node interfaces.NodeInfo) {
	if m.treeProvider != nil {
		// Extract parent path from the full path
		parentPath := filepath.Dir(path)
		if parentPath == "/" || parentPath == "." {
			parentPath = "/"
		}
		m.treeProvider.AddNode(parentPath, node)
	}
}

// CreateSimpleNode creates a simple tree node for testing
func CreateSimpleNode(name string, repository string) interfaces.NodeInfo {
	return interfaces.NodeInfo{
		Name:       name,
		Path:       "/" + name,
		Repository: repository,
		IsCloned:   true,
	}
}

// CreateConfigNode creates a config reference node for testing
func CreateConfigNode(name string, configFile string) interfaces.NodeInfo {
	return interfaces.NodeInfo{
		Name:       name,
		Path:       "/" + name,
		ConfigFile: configFile,
		IsConfig:   true,
	}
}

// CreateLazyNode creates a lazy repository node for testing
func CreateLazyNode(name string, repository string) interfaces.NodeInfo {
	return interfaces.NodeInfo{
		Name:       name,
		Path:       "/" + name,
		Repository: repository,
		IsLazy:     true,
		IsCloned:   false,
	}
}

// AssertPathResolution checks that a path resolves correctly
func AssertPathResolution(t *testing.T, m *Manager, target string, expected string, ensure bool) {
	result, err := m.ResolvePath(target, ensure)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

// AssertPathResolutionError checks that a path resolution returns an error
func AssertPathResolutionError(t *testing.T, m *Manager, target string, ensure bool, errorContains string) {
	result, err := m.ResolvePath(target, ensure)
	require.Error(t, err)
	require.Empty(t, result)
	if errorContains != "" {
		require.Contains(t, err.Error(), errorContains)
	}
}