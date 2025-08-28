package tree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, tmpDir, manager.rootPath)
	assert.Equal(t, filepath.Join(tmpDir, "repos"), manager.reposPath)
}

func TestInitialize(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	// Test initialization without root repo
	err = manager.Initialize("test-project", "")
	require.NoError(t, err)
	
	// Verify repos directory was created
	assert.DirExists(t, filepath.Join(tmpDir, "repos"))
	
	// Verify state file was created
	assert.FileExists(t, filepath.Join(tmpDir, ".repo-claude-tree.json"))
	
	// Verify root node
	assert.NotNil(t, manager.rootNode)
	assert.Equal(t, "root", manager.rootNode.ID)
	assert.Equal(t, "test-project", manager.rootNode.Name)
	assert.Equal(t, "/", manager.rootNode.Path)
}

func TestAddRepo(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	err = manager.Initialize("test-project", "")
	require.NoError(t, err)
	
	t.Run("AddRegularRepo", func(t *testing.T) {
		repo, err := manager.AddRepo("https://github.com/test/repo.git", AddOptions{
			Name: "test-repo",
			Lazy: false,
		})
		
		// Will fail to clone with test URL
		assert.Error(t, err) // Clone will fail with test URL
		assert.Nil(t, repo) // Returns nil on error
		
		// Repo is NOT added when clone fails (by design)
		assert.Len(t, manager.currentNode.Repos, 0)
	})
	
	t.Run("AddLazyRepo", func(t *testing.T) {
		// Reset
		manager.currentNode.Repos = []RepoConfig{}
		
		repo, err := manager.AddRepo("https://github.com/test/lazy-repo.git", AddOptions{
			Name: "lazy-repo",
			Lazy: true,
		})
		
		require.NoError(t, err)
		assert.NotNil(t, repo)
		assert.Equal(t, "lazy-repo", repo.Name)
		assert.True(t, repo.Lazy)
		assert.Equal(t, string(RepoStateMissing), repo.State)
		
		// Verify child node was created
		assert.NotNil(t, manager.currentNode.Children)
		assert.Contains(t, manager.currentNode.Children, "lazy-repo")
	})
	
	t.Run("AddDuplicateRepo", func(t *testing.T) {
		repo, err := manager.AddRepo("https://github.com/test/lazy-repo.git", AddOptions{
			Name: "lazy-repo",
			Lazy: true,
		})
		
		assert.Error(t, err)
		assert.Nil(t, repo)
		assert.Contains(t, err.Error(), "already exists")
	})
	
	t.Run("ExtractNameFromURL", func(t *testing.T) {
		repo, err := manager.AddRepo("https://github.com/test/auto-name.git", AddOptions{
			Lazy: true,
		})
		
		require.NoError(t, err)
		assert.Equal(t, "auto-name", repo.Name)
	})
}

func TestUseNode(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	err = manager.Initialize("test-project", "")
	require.NoError(t, err)
	
	// Add a lazy repo
	_, err = manager.AddRepo("https://github.com/test/child.git", AddOptions{
		Name: "child",
		Lazy: true,
	})
	require.NoError(t, err)
	
	t.Run("UseRoot", func(t *testing.T) {
		node, err := manager.UseNode("/", false)
		require.NoError(t, err)
		assert.Equal(t, "root", node.ID)
		assert.Equal(t, manager.rootNode, manager.currentNode)
	})
	
	t.Run("UseNonExistent", func(t *testing.T) {
		node, err := manager.UseNode("nonexistent", false)
		assert.Error(t, err)
		assert.Nil(t, node)
	})
	
	t.Run("UseSpecialPaths", func(t *testing.T) {
		// Test empty path
		node, err := manager.UseNode("", false)
		require.NoError(t, err)
		assert.Equal(t, manager.rootNode, node)
		
		// Test dot
		node, err = manager.UseNode(".", false)
		require.NoError(t, err)
		assert.Equal(t, manager.rootNode, node)
		
		// Test tilde
		node, err = manager.UseNode("~", false)
		require.NoError(t, err)
		assert.Equal(t, manager.rootNode, node)
	})
}

func TestResolveTarget(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	err = manager.Initialize("test-project", "")
	require.NoError(t, err)
	
	t.Run("ExplicitPath", func(t *testing.T) {
		resolution, err := manager.ResolveTarget("/")
		require.NoError(t, err)
		assert.Equal(t, SourceExplicit, resolution.Source)
		assert.Equal(t, manager.rootNode, resolution.Node)
	})
	
	t.Run("StoredCurrent", func(t *testing.T) {
		// When outside workspace, should use stored current
		oldCwd, _ := os.Getwd()
		os.Chdir("/tmp")
		defer os.Chdir(oldCwd)
		
		resolution, err := manager.ResolveTarget("")
		require.NoError(t, err)
		assert.Equal(t, SourceStored, resolution.Source)
		assert.Equal(t, manager.currentNode, resolution.Node)
	})
	
	t.Run("DefaultRoot", func(t *testing.T) {
		manager.currentNode = nil
		resolution, err := manager.ResolveTarget("")
		require.NoError(t, err)
		assert.Equal(t, SourceRoot, resolution.Source)
		assert.Equal(t, manager.rootNode, resolution.Node)
	})
}

func TestRemoveRepo(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	err = manager.Initialize("test-project", "")
	require.NoError(t, err)
	
	// Add a lazy repo
	_, err = manager.AddRepo("https://github.com/test/remove-me.git", AddOptions{
		Name: "remove-me",
		Lazy: true,
	})
	require.NoError(t, err)
	
	t.Run("RemoveExisting", func(t *testing.T) {
		err := manager.RemoveRepo("remove-me")
		require.NoError(t, err)
		
		// Verify removed from repos list
		assert.Len(t, manager.currentNode.Repos, 0)
		
		// Verify child node removed
		assert.NotContains(t, manager.currentNode.Children, "remove-me")
	})
	
	t.Run("RemoveNonExistent", func(t *testing.T) {
		err := manager.RemoveRepo("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestStateManagement(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	err = manager.Initialize("test-project", "")
	require.NoError(t, err)
	
	// Add some repos
	_, err = manager.AddRepo("https://github.com/test/repo1.git", AddOptions{
		Name: "repo1",
		Lazy: true,
	})
	require.NoError(t, err)
	
	t.Run("SaveAndLoadState", func(t *testing.T) {
		// Create new manager instance
		manager2, err := NewManager(tmpDir)
		require.NoError(t, err)
		
		// Load tree
		err = manager2.LoadTree()
		require.NoError(t, err)
		
		// Verify tree structure restored
		assert.NotNil(t, manager2.rootNode)
		assert.Equal(t, "test-project", manager2.rootNode.Name)
		assert.Len(t, manager2.rootNode.Repos, 1)
		assert.Equal(t, "repo1", manager2.rootNode.Repos[0].Name)
		
		// Verify children restored
		assert.Contains(t, manager2.rootNode.Children, "repo1")
	})
	
	t.Run("LoadWithoutState", func(t *testing.T) {
		tmpDir2 := t.TempDir()
		manager3, err := NewManager(tmpDir2)
		require.NoError(t, err)
		
		// Should not error even without state file
		err = manager3.LoadTree()
		require.NoError(t, err)
		
		// Should have default root
		assert.NotNil(t, manager3.rootNode)
	})
}

func TestTreeNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	err = manager.Initialize("test-project", "")
	require.NoError(t, err)
	
	// Create a deeper structure
	_, err = manager.AddRepo("https://github.com/test/level1.git", AddOptions{
		Name: "level1",
		Lazy: true,
	})
	require.NoError(t, err)
	
	// Navigate to level1 - this won't work since it's lazy and not cloned
	// Create the directory first for testing
	level1Dir := filepath.Join(tmpDir, "repos", "level1")
	err = os.MkdirAll(level1Dir, 0755)
	require.NoError(t, err)
	
	_, err = manager.UseNode("level1", false)
	require.NoError(t, err)
	
	// Add child to level1
	_, err = manager.AddRepo("https://github.com/test/level2.git", AddOptions{
		Name: "level2",
		Lazy: true,
	})
	require.NoError(t, err)
	
	t.Run("NavigateRelative", func(t *testing.T) {
		// Go to level2
		node, err := manager.UseNode("./level2", false)
		require.NoError(t, err)
		assert.Equal(t, "level2", node.Name)
		
		// Go up
		node, err = manager.UseNode("..", false)
		require.NoError(t, err)
		assert.Equal(t, "level1", node.Name)
		
		// Go up again
		node, err = manager.UseNode("..", false)
		require.NoError(t, err)
		assert.Equal(t, "root", node.ID)
	})
	
	t.Run("NavigateAbsolute", func(t *testing.T) {
		node, err := manager.UseNode("/level1/level2", false)
		require.NoError(t, err)
		assert.Equal(t, "level2", node.Name)
	})
	
	t.Run("NavigatePrevious", func(t *testing.T) {
		// Set up previous
		_, err = manager.UseNode("/level1", false)
		require.NoError(t, err)
		_, err = manager.UseNode("/", false)
		require.NoError(t, err)
		
		// Use dash to go to previous
		node, err := manager.UseNode("-", false)
		require.NoError(t, err)
		assert.Equal(t, "level1", node.Name)
	})
}

func TestCloneLazy(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	err = manager.Initialize("test-project", "")
	require.NoError(t, err)
	
	// Add multiple lazy repos
	_, err = manager.AddRepo("https://github.com/test/lazy1.git", AddOptions{
		Name: "lazy1",
		Lazy: true,
	})
	require.NoError(t, err)
	
	_, err = manager.AddRepo("https://github.com/test/lazy2.git", AddOptions{
		Name: "lazy2",
		Lazy: true,
	})
	require.NoError(t, err)
	
	t.Run("CloneLazyNonRecursive", func(t *testing.T) {
		// This will try to clone but fail with test URLs
		err := manager.CloneLazy(false)
		// Error expected due to invalid URLs
		assert.Error(t, err)
	})
	
	t.Run("CloneLazyRecursive", func(t *testing.T) {
		// Create directory for lazy1 first
		lazy1Dir := filepath.Join(tmpDir, "repos", "lazy1")
		err = os.MkdirAll(lazy1Dir, 0755)
		require.NoError(t, err)
		
		// Navigate to lazy1
		_, err = manager.UseNode("lazy1", false)
		require.NoError(t, err)
		
		// Add a child
		_, err = manager.AddRepo("https://github.com/test/lazy3.git", AddOptions{
			Name: "lazy3",
			Lazy: true,
		})
		require.NoError(t, err)
		
		// Try recursive clone from root
		_, err = manager.UseNode("/", false)
		require.NoError(t, err)
		
		err = manager.CloneLazy(true)
		// Error expected due to invalid URLs
		assert.Error(t, err)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("ExtractRepoName", func(t *testing.T) {
		tests := []struct {
			url      string
			expected string
		}{
			{"https://github.com/user/repo.git", "repo"},
			{"https://github.com/user/repo", "repo"},
			{"git@github.com:user/repo.git", "repo"},
			{"user/repo", "repo"},
			{"repo", "repo"},
		}
		
		for _, tt := range tests {
			result := extractRepoName(tt.url)
			assert.Equal(t, tt.expected, result, "URL: %s", tt.url)
		}
	})
}

func TestCWDMapping(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	err = manager.Initialize("test-project", "")
	require.NoError(t, err)
	
	// Create repos directory structure for testing
	reposDir := filepath.Join(tmpDir, "repos")
	testDir := filepath.Join(reposDir, "test")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)
	
	t.Run("MapCWDInWorkspace", func(t *testing.T) {
		// Change to repos directory
		oldCwd, _ := os.Getwd()
		defer os.Chdir(oldCwd)
		
		err := os.Chdir(reposDir)
		require.NoError(t, err)
		
		node := manager.mapCWDToNode(reposDir)
		assert.NotNil(t, node)
		assert.Equal(t, manager.rootNode, node)
	})
	
	t.Run("MapCWDOutsideWorkspace", func(t *testing.T) {
		oldCwd, _ := os.Getwd()
		defer os.Chdir(oldCwd)
		
		err := os.Chdir("/tmp")
		require.NoError(t, err)
		
		node := manager.mapCWDToNode("/tmp")
		assert.Nil(t, node)
	})
}

func TestGetters(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	err = manager.Initialize("test-project", "")
	require.NoError(t, err)
	
	t.Run("GetCurrentNode", func(t *testing.T) {
		node := manager.GetCurrentNode()
		assert.NotNil(t, node)
		assert.Equal(t, manager.currentNode, node)
	})
	
	t.Run("GetRootNode", func(t *testing.T) {
		node := manager.GetRootNode()
		assert.NotNil(t, node)
		assert.Equal(t, manager.rootNode, node)
	})
}

func TestNilMapFix(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	err = manager.Initialize("test-project", "")
	require.NoError(t, err)
	
	// Ensure Children is nil initially
	manager.currentNode.Children = nil
	
	// Add repo should initialize Children map
	_, err = manager.AddRepo("https://github.com/test/repo.git", AddOptions{
		Name: "test",
		Lazy: true,
	})
	require.NoError(t, err)
	
	// Should not panic and Children should be initialized
	assert.NotNil(t, manager.currentNode.Children)
	assert.Contains(t, manager.currentNode.Children, "test")
}

func TestStateFileCorruption(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	err = manager.Initialize("test-project", "")
	require.NoError(t, err)
	
	// Corrupt the state file
	statePath := filepath.Join(tmpDir, ".repo-claude-tree.json")
	err = os.WriteFile(statePath, []byte("invalid json"), 0644)
	require.NoError(t, err)
	
	// Try to load - should handle gracefully
	manager2, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	err = manager2.LoadTree()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}