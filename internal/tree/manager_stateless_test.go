package tree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
)

// TestManagerStatelessCore tests core stateless manager functionality
func TestManagerStatelessCore(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create basic config
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-ws",
			ReposDir: "repositories", 
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Fetch: "eager"},
			{Name: "repo2", URL: "https://github.com/test/repo2.git", Fetch: "lazy"},
		},
	}
	
	configPath := filepath.Join(tmpDir, "muno.yaml")
	require.NoError(t, cfg.Save(configPath))
	
	// Create repo directories
	reposDir := filepath.Join(tmpDir, "repositories")
	require.NoError(t, os.MkdirAll(filepath.Join(reposDir, "repo1", ".git"), 0755))
	require.NoError(t, os.MkdirAll(reposDir, 0755))
	
	mockGit := &git.MockGit{
		StatusFunc: func(path string) (string, error) {
			return "clean", nil
		},
		CloneFunc: func(url, path string) error {
			return os.MkdirAll(filepath.Join(path, ".git"), 0755)
		},
	}
	
	mgr, err := NewManager(tmpDir, mockGit)
	require.NoError(t, err)
	
	// Test stateless behavior - current path always root
	assert.Equal(t, "/", mgr.GetCurrentPath())
	assert.Equal(t, "/", mgr.currentPath)
	
	// Test filesystem path computation
	tests := []struct {
		logical  string
		expected string
	}{
		{"/", filepath.Join(tmpDir, "repositories")},
		{"/repo1", filepath.Join(tmpDir, "repositories", "repo1")},
		{"/repo1/sub", filepath.Join(tmpDir, "repositories", "repo1", "sub")},
		{"/repo2", filepath.Join(tmpDir, "repositories", "repo2")},
		{"", filepath.Join(tmpDir, "repositories")}, // empty = root
	}
	
	for _, tt := range tests {
		actual := mgr.ComputeFilesystemPath(tt.logical)
		assert.Equal(t, tt.expected, actual, "Path: %s", tt.logical)
	}
	
	// Test GetNodeByPath
	node, err := mgr.GetNodeByPath("/repo1")
	assert.NoError(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, "repo1", node.Name)
	
	node, err = mgr.GetNodeByPath("/")
	assert.NoError(t, err)
	assert.Nil(t, node) // root returns nil
	
	node, err = mgr.GetNodeByPath("/nonexistent")
	assert.Error(t, err)
	
	// Test GetNode
	treeNode := mgr.GetNode("/repo1")
	assert.NotNil(t, treeNode)
	assert.Equal(t, "repo1", treeNode.Name)
	assert.Equal(t, RepoStateCloned, treeNode.State)
	
	treeNode = mgr.GetNode("/repo2")
	assert.NotNil(t, treeNode)
	assert.Equal(t, RepoStateMissing, treeNode.State)
	
	// Test GetState
	state := mgr.GetState()
	assert.NotNil(t, state)
	assert.Equal(t, "/", state.CurrentPath)
	assert.Contains(t, state.Nodes, "/repo1")
	assert.Contains(t, state.Nodes, "/repo2")
	
	// Test ListChildren
	children, err := mgr.ListChildren("/")
	assert.NoError(t, err)
	assert.Len(t, children, 2)
	
	// Test AddRepo
	err = mgr.AddRepo("/", "new", "https://github.com/test/new.git", true)
	assert.NoError(t, err)
	assert.Len(t, mgr.config.Nodes, 3)
	
	// Test RemoveNode
	err = mgr.RemoveNode("/new")
	assert.NoError(t, err)
	assert.Len(t, mgr.config.Nodes, 2)
}

// TestManagerDisplayStateless tests display methods in stateless mode
func TestManagerDisplayStateless(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name: "display-test",
		},
		Nodes: []config.NodeDefinition{
			{Name: "test", URL: "https://github.com/test/test.git"},
		},
	}
	
	configPath := filepath.Join(tmpDir, "muno.yaml")
	require.NoError(t, cfg.Save(configPath))
	
	mgr, err := NewManager(tmpDir, nil)
	require.NoError(t, err)
	
	// Test display methods
	output := mgr.DisplayTree()
	assert.Contains(t, output, "display-test")
	assert.Contains(t, output, "test")
	
	output = mgr.DisplayTreeWithDepth(1)
	assert.Contains(t, output, "display-test")
	
	output = mgr.DisplayStatus()
	assert.Contains(t, output, "Tree Status")
	assert.Contains(t, output, "Current Path:")
	
	output = mgr.DisplayPath()
	assert.Contains(t, output, "/")
	
	output = mgr.DisplayChildren()
	assert.Contains(t, output, "test")
}

// TestManagerStatelessEdgeCases tests edge cases in stateless mode  
func TestManagerStatelessEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Test with no config file - in stateless mode this may fail
	mgr, err := NewManager(tmpDir, nil)
	// Either it succeeds with a default config or fails gracefully
	if err == nil {
		assert.NotNil(t, mgr)
		assert.NotNil(t, mgr.config)
	} else {
		// It's okay to fail if no config exists
		assert.Contains(t, err.Error(), "config")
	}
	
	// Test with empty workspace path
	_, err = NewManager("", nil)
	assert.Error(t, err)
	
	// Test CloneLazyRepos
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{Name: "test"},
		Nodes: []config.NodeDefinition{
			{Name: "lazy", URL: "https://github.com/test/lazy.git", Fetch: "lazy"},
		},
	}
	
	workDir := filepath.Join(tmpDir, "work")
	require.NoError(t, os.MkdirAll(workDir, 0755))
	require.NoError(t, cfg.Save(filepath.Join(workDir, "muno.yaml")))
	
	cloneCalled := false
	mockGit := &git.MockGit{
		CloneFunc: func(url, path string) error {
			cloneCalled = true
			return os.MkdirAll(filepath.Join(path, ".git"), 0755)
		},
	}
	
	mgr2, err := NewManager(workDir, mockGit)
	require.NoError(t, err)
	
	err = mgr2.CloneLazyRepos("/", false)
	assert.NoError(t, err)
	assert.True(t, cloneCalled)
}

// TestGetNodeStateless provides additional GetNode coverage
func TestGetNodeStateless(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: config.GetDefaultNodesDir(),
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git"},
			{Name: "config-ref", File: "sub/muno.yaml"},
		},
	}
	
	configPath := filepath.Join(tmpDir, "muno.yaml")
	require.NoError(t, cfg.Save(configPath))
	
	// Create repos directory
	// Get the actual nodes directory from config
	nodesDir := config.GetDefaultNodesDir()
	reposDir := filepath.Join(tmpDir, nodesDir)
	require.NoError(t, os.MkdirAll(reposDir, 0755))
	
	// Create repo1 with .git
	require.NoError(t, os.MkdirAll(filepath.Join(reposDir, "repo1", ".git"), 0755))
	
	// Create config-ref directory
	require.NoError(t, os.MkdirAll(filepath.Join(reposDir, "config-ref"), 0755))
	
	mgr, err := NewManager(tmpDir, nil)
	require.NoError(t, err)
	
	// Test getting regular repo
	node := mgr.GetNode("/repo1")
	assert.NotNil(t, node)
	assert.Equal(t, "repo1", node.Name)
	assert.Equal(t, NodeTypeRepo, node.Type)
	assert.Equal(t, RepoStateCloned, node.State)
	
	// Test getting config reference
	node = mgr.GetNode("/config-ref")
	assert.NotNil(t, node)
	assert.Equal(t, "config-ref", node.Name)
	// Config refs have their own type now
	assert.Equal(t, NodeTypeFile, node.Type)
	
	// Test non-existent node
	node = mgr.GetNode("/nonexistent")
	assert.Nil(t, node)
	
	// Test root node
	node = mgr.GetNode("/")
	assert.NotNil(t, node)
	assert.Equal(t, NodeTypeRoot, node.Type)
}

// TestComputeFilesystemPathStateless provides additional path computation coverage
func TestComputeFilesystemPathStateless(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: "custom",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo", URL: "https://github.com/test/repo.git"},
		},
	}
	
	configPath := filepath.Join(tmpDir, "muno.yaml")
	require.NoError(t, cfg.Save(configPath))
	
	mgr, err := NewManager(tmpDir, nil)
	require.NoError(t, err)
	
	// Test various path computations
	tests := []struct {
		input    string
		expected string
	}{
		{"/", filepath.Join(tmpDir, "custom")},
		{"", filepath.Join(tmpDir, "custom")},
		{"/repo", filepath.Join(tmpDir, "custom", "repo")},
		{"/repo/nested", filepath.Join(tmpDir, "custom", "repo", "nested")},
		{"/repo/nested/deep", filepath.Join(tmpDir, "custom", "repo", "nested", "deep")},
	}
	
	for _, tt := range tests {
		actual := mgr.ComputeFilesystemPath(tt.input)
		assert.Equal(t, tt.expected, actual, "Input: %s", tt.input)
	}
}

// TestListChildrenStateless provides additional ListChildren coverage
func TestListChildrenStateless(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{Name: "test"},
		Nodes: []config.NodeDefinition{
			{Name: "a", URL: "https://github.com/test/a.git"},
			{Name: "b", URL: "https://github.com/test/b.git"},
			{Name: "c", File: "c/muno.yaml"},
		},
	}
	
	configPath := filepath.Join(tmpDir, "muno.yaml")
	require.NoError(t, cfg.Save(configPath))
	
	mgr, err := NewManager(tmpDir, nil)
	require.NoError(t, err)
	
	// List root children
	children, err := mgr.ListChildren("/")
	assert.NoError(t, err)
	assert.Len(t, children, 3)
	
	// Check names
	names := make(map[string]bool)
	for _, child := range children {
		names[child.Name] = true
	}
	assert.True(t, names["a"])
	assert.True(t, names["b"])
	assert.True(t, names["c"])
	
	// List non-existent path - this returns an error now
	children, err = mgr.ListChildren("/nonexistent")
	assert.Error(t, err)
	assert.Empty(t, children)
}