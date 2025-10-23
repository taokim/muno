package navigator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/constants"
)

func TestFilesystemNavigator(t *testing.T) {
	t.Run("NewFilesystemNavigator", func(t *testing.T) {
		workspace := t.TempDir()
		
		// Test with empty workspace
		_, err := NewFilesystemNavigator("", nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workspace path cannot be empty")
		
		// Test with non-existent workspace
		_, err = NewFilesystemNavigator("/non/existent/path", nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workspace does not exist")
		
		// Test with valid workspace
		nav, err := NewFilesystemNavigator(workspace, nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, nav)
		assert.Equal(t, workspace, nav.workspace)
		assert.Equal(t, "/", nav.currentPath)
		
		// Test with provided config
		cfg := config.DefaultConfigTree("test")
		nav, err = NewFilesystemNavigator(workspace, cfg, nil)
		require.NoError(t, err)
		assert.Equal(t, cfg, nav.config)
		
		// Test with nil git command (acceptable)
		nav, err = NewFilesystemNavigator(workspace, cfg, nil)
		require.NoError(t, err)
		assert.Nil(t, nav.gitCmd)
	})

	t.Run("GetCurrentPath", func(t *testing.T) {
		workspace := t.TempDir()
		nav, err := NewFilesystemNavigator(workspace, nil, nil)
		require.NoError(t, err)
		
		path, err := nav.GetCurrentPath()
		assert.NoError(t, err)
		assert.Equal(t, "/", path)
	})

	t.Run("Navigate", func(t *testing.T) {
		workspace := t.TempDir()
		cfg := config.DefaultConfigTree("test")
		
		// Add test nodes to config
		cfg.Nodes = []config.NodeDefinition{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", Fetch: "eager"},
			{Name: "lazy-repo", URL: "https://github.com/test/lazy.git", Fetch: "lazy"},
		}
		
		// Save config
		configPath := filepath.Join(workspace, "muno.yaml")
		require.NoError(t, cfg.Save(configPath))
		
		// Create test-repo directory
		testRepoPath := filepath.Join(workspace, "test-repo")
		require.NoError(t, os.MkdirAll(testRepoPath, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(testRepoPath, ".git"), 0755))
		
		nav, err := NewFilesystemNavigator(workspace, cfg, nil)
		require.NoError(t, err)
		
		// Navigate to existing repo
		err = nav.Navigate("/test-repo")
		assert.NoError(t, err)
		
		path, _ := nav.GetCurrentPath()
		assert.Equal(t, "/test-repo", path)
		
		// Navigate to root
		err = nav.Navigate("/")
		assert.NoError(t, err)
		
		path, _ = nav.GetCurrentPath()
		assert.Equal(t, "/", path)
		
		// Navigate to non-existent node
		err = nav.Navigate("/non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "node not found")
		
		// Test relative navigation
		nav.currentPath = "/test-repo"
		err = nav.Navigate("..")
		assert.NoError(t, err)
		path, _ = nav.GetCurrentPath()
		assert.Equal(t, "/", path)
	})

	t.Run("GetNode", func(t *testing.T) {
		workspace := t.TempDir()
		cfg := config.DefaultConfigTree("test")
		cfg.Nodes = []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git"},
		}
		
		configPath := filepath.Join(workspace, "muno.yaml")
		require.NoError(t, cfg.Save(configPath))
		
		nav, err := NewFilesystemNavigator(workspace, cfg, nil)
		require.NoError(t, err)
		
		// Get root node
		node, err := nav.GetNode("/")
		assert.NoError(t, err)
		assert.NotNil(t, node)
		assert.Equal(t, "/", node.Path)
		assert.Equal(t, "root", node.Name)
		assert.Equal(t, NodeTypeRoot, node.Type)
		assert.Contains(t, node.Children, "repo1")
		
		// Get child node
		node, err = nav.GetNode("/repo1")
		assert.NoError(t, err)
		assert.NotNil(t, node)
		assert.Equal(t, "/repo1", node.Path)
		assert.Equal(t, "repo1", node.Name)
		assert.Equal(t, NodeTypeRepo, node.Type)
		assert.Equal(t, "https://github.com/test/repo1.git", node.URL)
		
		// Get non-existent node
		node, err = nav.GetNode("/non-existent")
		assert.NoError(t, err)
		assert.Nil(t, node)
	})

	t.Run("ListChildren", func(t *testing.T) {
		workspace := t.TempDir()
		cfg := config.DefaultConfigTree("test")
		cfg.Nodes = []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git"},
			{Name: "repo2", URL: "https://github.com/test/repo2.git"},
		}
		
		configPath := filepath.Join(workspace, "muno.yaml")
		require.NoError(t, cfg.Save(configPath))
		
		nav, err := NewFilesystemNavigator(workspace, cfg, nil)
		require.NoError(t, err)
		
		// List root children
		children, err := nav.ListChildren("/")
		assert.NoError(t, err)
		assert.Len(t, children, 2)
		
		var names []string
		for _, child := range children {
			names = append(names, child.Name)
		}
		assert.Contains(t, names, "repo1")
		assert.Contains(t, names, "repo2")
		
		// List non-existent node
		children, err = nav.ListChildren("/non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "node not found")
	})

	t.Run("GetTree", func(t *testing.T) {
		workspace := t.TempDir()
		cfg := config.DefaultConfigTree("test")
		cfg.Nodes = []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git"},
			{Name: "repo2", URL: "https://github.com/test/repo2.git"},
		}
		
		configPath := filepath.Join(workspace, "muno.yaml")
		require.NoError(t, cfg.Save(configPath))
		
		// Create repo directories with nested structure
		reposDir := cfg.Workspace.ReposDir
		if reposDir == "" {
			reposDir = constants.DefaultReposDir
		}
		repo1Path := filepath.Join(workspace, reposDir, "repo1")
		require.NoError(t, os.MkdirAll(filepath.Join(repo1Path, ".git"), 0755))
		
		// Create a nested muno.yaml in repo1
		nestedCfg := config.DefaultConfigTree("repo1")
		nestedCfg.Nodes = []config.NodeDefinition{
			{Name: "nested1", URL: "https://github.com/test/nested1.git"},
		}
		require.NoError(t, nestedCfg.Save(filepath.Join(repo1Path, "muno.yaml")))
		
		nav, err := NewFilesystemNavigator(workspace, cfg, nil)
		require.NoError(t, err)
		
		// Get tree with depth 1
		tree, err := nav.GetTree("/", 1)
		assert.NoError(t, err)
		assert.NotNil(t, tree)
		assert.Equal(t, "/", tree.Root.Path)
		assert.Len(t, tree.Root.Children, 2)
		assert.Equal(t, 1, tree.Depth)
		
		// Get tree with depth 2
		tree, err = nav.GetTree("/", 2)
		assert.NoError(t, err)
		assert.NotNil(t, tree)
		
		// Verify repo1 is included (nested config would require actual navigation/loading)
		assert.Contains(t, tree.Nodes, "/repo1")
		repo1Node := tree.Nodes["/repo1"]
		assert.NotNil(t, repo1Node)
		// Note: nested1 won't be loaded automatically without navigating to repo1 first
		
		// Get tree for non-existent node
		tree, err = nav.GetTree("/non-existent", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "node not found")
	})

	t.Run("GetNodeStatus", func(t *testing.T) {
		workspace := t.TempDir()
		cfg := config.DefaultConfigTree("test")
		cfg.Nodes = []config.NodeDefinition{
			{Name: "cloned-repo", URL: "https://github.com/test/cloned.git", Fetch: "eager"},
			{Name: "lazy-repo", URL: "https://github.com/test/lazy.git", Fetch: "lazy"},
		}
		
		configPath := filepath.Join(workspace, "muno.yaml")
		require.NoError(t, cfg.Save(configPath))
		
		// Create cloned repo with .git in the repos directory
		reposDir := cfg.Workspace.ReposDir
		if reposDir == "" {
			reposDir = constants.DefaultReposDir
		}
		clonedPath := filepath.Join(workspace, reposDir, "cloned-repo")
		require.NoError(t, os.MkdirAll(filepath.Join(clonedPath, ".git"), 0755))
		
		nav, err := NewFilesystemNavigator(workspace, cfg, nil)
		require.NoError(t, err)
		
		// Get status of cloned repo
		status, err := nav.GetNodeStatus("/cloned-repo")
		assert.NoError(t, err)
		assert.NotNil(t, status)
		assert.True(t, status.Exists)
		assert.True(t, status.Cloned)
		assert.False(t, status.Lazy)
		assert.Equal(t, RepoStateCloned, status.State)
		// Without git mock, branch and remote URL won't be set
		assert.Empty(t, status.Branch)
		assert.Empty(t, status.RemoteURL)
		assert.False(t, status.Modified)
		
		// Get status of lazy repo (not cloned)
		status, err = nav.GetNodeStatus("/lazy-repo")
		assert.NoError(t, err)
		assert.NotNil(t, status)
		assert.False(t, status.Exists)
		assert.False(t, status.Cloned)
		assert.True(t, status.Lazy)
		assert.Equal(t, RepoStateMissing, status.State)
		
		// Get status of non-existent node
		status, err = nav.GetNodeStatus("/non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "node not found")
	})

	t.Run("RefreshStatus", func(t *testing.T) {
		workspace := t.TempDir()
		nav, err := NewFilesystemNavigator(workspace, nil, nil)
		require.NoError(t, err)
		
		// RefreshStatus is a no-op for filesystem navigator
		err = nav.RefreshStatus("/")
		assert.NoError(t, err)
	})

	t.Run("IsLazy", func(t *testing.T) {
		workspace := t.TempDir()
		cfg := config.DefaultConfigTree("test")
		cfg.Nodes = []config.NodeDefinition{
			{Name: "eager-repo", URL: "https://github.com/test/eager.git", Fetch: "eager"},
			{Name: "lazy-repo", URL: "https://github.com/test/lazy.git", Fetch: "lazy"},
		}
		
		configPath := filepath.Join(workspace, "muno.yaml")
		require.NoError(t, cfg.Save(configPath))
		
		nav, err := NewFilesystemNavigator(workspace, cfg, nil)
		require.NoError(t, err)
		
		// Check eager repo
		isLazy, err := nav.IsLazy("/eager-repo")
		assert.NoError(t, err)
		assert.False(t, isLazy)
		
		// Check lazy repo
		isLazy, err = nav.IsLazy("/lazy-repo")
		assert.NoError(t, err)
		assert.True(t, isLazy)
		
		// Check non-existent node
		isLazy, err = nav.IsLazy("/non-existent")
		assert.NoError(t, err)
		assert.False(t, isLazy)
	})

	t.Run("TriggerLazyLoad", func(t *testing.T) {
		workspace := t.TempDir()
		cfg := config.DefaultConfigTree("test")
		cfg.Nodes = []config.NodeDefinition{
			{Name: "lazy-repo", URL: "https://github.com/test/lazy.git", Fetch: "lazy"},
		}
		
		configPath := filepath.Join(workspace, "muno.yaml")
		require.NoError(t, cfg.Save(configPath))
		
		nav, err := NewFilesystemNavigator(workspace, cfg, nil)
		require.NoError(t, err)
		
		// Trigger lazy load (will fail without git mock but that's expected)
		err = nav.TriggerLazyLoad("/lazy-repo")
		// Without git, this will error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "git command not configured")
		
		// Try to trigger on non-lazy repo
		cfg.Nodes = append(cfg.Nodes, config.NodeDefinition{
			Name:  "eager-repo",
			URL:   "https://github.com/test/eager.git",
			Fetch: "eager",
		})
		require.NoError(t, cfg.Save(configPath))
		
		nav, err = NewFilesystemNavigator(workspace, cfg, nil)
		require.NoError(t, err)
		
		err = nav.TriggerLazyLoad("/eager-repo")
		assert.Error(t, err)
		// Will fail with git command not configured since gitCmd is nil
		
		// Try on non-existent node
		err = nav.TriggerLazyLoad("/non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "node not found")
	})

	t.Run("CurrentPathPersistence", func(t *testing.T) {
		workspace := t.TempDir()
		cfg := config.DefaultConfigTree("test")
		cfg.Nodes = []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git"},
		}
		
		configPath := filepath.Join(workspace, "muno.yaml")
		require.NoError(t, cfg.Save(configPath))
		
		// Create repo directory
		reposDir := cfg.Workspace.ReposDir
		if reposDir == "" {
			reposDir = constants.DefaultReposDir
		}
		repo1Path := filepath.Join(workspace, reposDir, "repo1")
		require.NoError(t, os.MkdirAll(filepath.Join(repo1Path, ".git"), 0755))
		
		// Create .muno directory
		munoDir := filepath.Join(workspace, ".muno")
		require.NoError(t, os.MkdirAll(munoDir, 0755))
		
		// First navigator instance
		nav1, err := NewFilesystemNavigator(workspace, cfg, nil)
		require.NoError(t, err)
		
		// Navigate to repo1
		err = nav1.Navigate("/repo1")
		assert.NoError(t, err)
		
		// Create new navigator instance - should load saved path
		nav2, err := NewFilesystemNavigator(workspace, cfg, nil)
		require.NoError(t, err)
		
		path, _ := nav2.GetCurrentPath()
		assert.Equal(t, "/repo1", path)
	})
}

// Helper tests for private methods
func TestFilesystemNavigatorHelpers(t *testing.T) {
	t.Run("normalizePath", func(t *testing.T) {
		workspace := t.TempDir()
		nav, _ := NewFilesystemNavigator(workspace, nil, nil)
		
		// Test various path formats
		tests := []struct {
			input    string
			current  string
			expected string
		}{
			{"/", "/", "/"},
			{"", "/", "/"},
			{".", "/", "/"},
			{"..", "/", "/"},
			{"/repo", "/", "/repo"},
			{"repo", "/", "/repo"},
			{"../", "/repo", "/"},
			{"../../", "/repo/sub", "/"},
			{"./sub", "/repo", "/repo/sub"},
			{"sub", "/repo", "/repo/sub"},
		}
		
		for _, tt := range tests {
			nav.currentPath = tt.current
			result := nav.normalizePath(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %s, Current: %s", tt.input, tt.current)
		}
	})

	t.Run("computeFilesystemPath", func(t *testing.T) {
		workspace := t.TempDir()
		nav, _ := NewFilesystemNavigator(workspace, nil, nil)
		
		// The filesystem navigator uses a repos directory for root,
		// but top-level repos go directly in workspace
		// Without config, findNodeDefinition returns nil, so nested paths use repos dir
		// Get the actual default nodes directory from config
		nodesDir := config.GetDefaultNodesDir()
		tests := []struct {
			nodePath string
			expected string
		}{
			{"/", filepath.Join(workspace, nodesDir)},
			{"/repo", filepath.Join(workspace, nodesDir, "repo")},  // Top-level repo goes in repos directory
			{"/repo/sub", filepath.Join(workspace, nodesDir, "repo", "sub")}, // Nested path also uses repos dir
		}
		
		for _, tt := range tests {
			result := nav.computeFilesystemPath(tt.nodePath)
			assert.Equal(t, tt.expected, result, "Node path: %s", tt.nodePath)
		}
	})

	t.Run("pathExists", func(t *testing.T) {
		workspace := t.TempDir()
		nav, _ := NewFilesystemNavigator(workspace, nil, nil)
		
		// Test existing path
		assert.True(t, nav.pathExists(workspace))
		
		// Test non-existing path
		assert.False(t, nav.pathExists(filepath.Join(workspace, "non-existent")))
		
		// Create a file and test
		testFile := filepath.Join(workspace, "test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)
		assert.True(t, nav.pathExists(testFile))
	})
}