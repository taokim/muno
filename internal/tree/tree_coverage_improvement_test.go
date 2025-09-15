package tree

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
)

func TestManager_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &git.MockGit{}
	
	CreateTestConfig(t, tmpDir, "")
	mgr, err := NewManager(tmpDir, mockGit)
	require.NoError(t, err)
	
	t.Run("GetState returns state", func(t *testing.T) {
		state := mgr.GetState()
		assert.NotNil(t, state)
		assert.Equal(t, "/", state.CurrentPath)
		assert.NotZero(t, state.LastUpdated)
	})
	
	t.Run("GetNode with various paths", func(t *testing.T) {
		// Root exists
		root := mgr.GetNode("/")
		assert.NotNil(t, root)
		assert.Equal(t, "root", root.Name)
		
		// Non-existent returns nil
		nonExistent := mgr.GetNode("/non-existent/path")
		assert.Nil(t, nonExistent)
	})
	
	t.Run("ComputeFilesystemPath edge cases", func(t *testing.T) {
		// Root path
		fsPath := mgr.ComputeFilesystemPath("/")
		assert.Equal(t, filepath.Join(tmpDir, config.GetDefaultReposDir()), fsPath)
		
		// Deep nested path
		fsPath = mgr.ComputeFilesystemPath("/a/b/c/d")
		expected := filepath.Join(tmpDir, config.GetDefaultReposDir(), "a", "b", "c", "d")
		assert.Equal(t, expected, fsPath)
	})
	
	t.Run("UseNode with auto-clone", func(t *testing.T) {
		// Add a lazy repo
		err := mgr.AddRepo("/", "lazy-repo", "https://github.com/test/lazy.git", true)
		require.NoError(t, err)
		
		// Mock clone function
		cloneCalled := false
		mockGit.CloneFunc = func(url, path string) error {
			cloneCalled = true
			// Create .git directory to simulate successful clone
			gitDir := filepath.Join(path, ".git")
			return os.MkdirAll(gitDir, 0755)
		}
		
		// Navigate to lazy repo should trigger clone
		err = mgr.UseNode("/lazy-repo")
		require.NoError(t, err)
		assert.True(t, cloneCalled)
		assert.Equal(t, "/lazy-repo", mgr.GetCurrentPath())
	})
	
	t.Run("AddRepo with clone failure rollback", func(t *testing.T) {
		// Mock clone to fail
		mockGit.CloneFunc = func(url, path string) error {
			return assert.AnError
		}
		
		// Try to add non-lazy repo (should clone immediately)
		err := mgr.AddRepo("/", "fail-repo", "https://github.com/test/fail.git", false)
		assert.Error(t, err)
		
		// Verify rollback - repo should not exist
		failNode := mgr.GetNode("/fail-repo")
		assert.Nil(t, failNode)
	})
	
	t.Run("RemoveNode navigates to parent if current removed", func(t *testing.T) {
		// Add and navigate to a repo
		mockGit.CloneFunc = nil // Reset to success
		err := mgr.AddRepo("/", "temp-repo", "https://github.com/test/temp.git", true)
		require.NoError(t, err)
		
		err = mgr.UseNode("/temp-repo")
		require.NoError(t, err)
		assert.Equal(t, "/temp-repo", mgr.GetCurrentPath())
		
		// Remove current node
		err = mgr.RemoveNode("/temp-repo")
		require.NoError(t, err)
		
		// Should navigate back to root
		assert.Equal(t, "/", mgr.GetCurrentPath())
	})
	
	t.Run("RemoveNode with nested path", func(t *testing.T) {
		// Add nested structure
		err := mgr.AddRepo("/", "parent", "https://github.com/test/parent.git", true)
		require.NoError(t, err)
		err = mgr.AddRepo("/parent", "child", "https://github.com/test/child.git", true)
		require.NoError(t, err)
		
		// Navigate to nested
		err = mgr.UseNode("/parent/child")
		require.NoError(t, err)
		
		// Remove parent (which removes child too)
		err = mgr.RemoveNode("/parent")
		require.NoError(t, err)
		
		// Should be back at root
		assert.Equal(t, "/", mgr.GetCurrentPath())
		
		// Both should be gone
		assert.Nil(t, mgr.GetNode("/parent"))
		assert.Nil(t, mgr.GetNode("/parent/child"))
	})
	
	t.Run("CloneLazyRepos recursive", func(t *testing.T) {
		// Reset manager
		mgr, err = NewManager(tmpDir, mockGit)
		require.NoError(t, err)
		
		// Add nested lazy repos
		err = mgr.AddRepo("/", "level1", "https://github.com/test/l1.git", true)
		require.NoError(t, err)
		err = mgr.AddRepo("/level1", "level2", "https://github.com/test/l2.git", true)
		require.NoError(t, err)
		
		cloneCount := 0
		mockGit.CloneFunc = func(url, path string) error {
			cloneCount++
			gitDir := filepath.Join(path, ".git")
			return os.MkdirAll(gitDir, 0755)
		}
		
		// Clone recursively from root
		err = mgr.CloneLazyRepos("/", true)
		require.NoError(t, err)
		assert.Equal(t, 2, cloneCount) // Both level1 and level2
	})
	
	t.Run("CloneLazyRepos non-recursive", func(t *testing.T) {
		// Reset manager
		mgr, err = NewManager(tmpDir, mockGit)
		require.NoError(t, err)
		
		// Add nested lazy repos
		err = mgr.AddRepo("/", "top", "https://github.com/test/top.git", true)
		require.NoError(t, err)
		err = mgr.AddRepo("/top", "nested", "https://github.com/test/nested.git", true)
		require.NoError(t, err)
		
		cloneCount := 0
		mockGit.CloneFunc = func(url, path string) error {
			cloneCount++
			gitDir := filepath.Join(path, ".git")
			return os.MkdirAll(gitDir, 0755)
		}
		
		// Clone non-recursively from root
		err = mgr.CloneLazyRepos("/", false)
		require.NoError(t, err)
		assert.Equal(t, 1, cloneCount) // Only top level
	})
}

func TestManager_StateFileCorruption(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &git.MockGit{}
	
	CreateTestConfig(t, tmpDir, "")
	mgr, err := NewManager(tmpDir, mockGit)
	require.NoError(t, err)
	
	// Add some repos
	err = mgr.AddRepo("/", "repo1", "https://github.com/test/repo1.git", true)
	require.NoError(t, err)
	
	// Corrupt the state file
	statePath := filepath.Join(tmpDir, ".muno-tree.json")
	err = os.WriteFile(statePath, []byte("invalid json"), 0644)
	require.NoError(t, err)
	
	// Try to load corrupted state
	// loadState is now private - we can't test it directly
	// Instead, create a new manager which will try to load the corrupted state
	_, err = NewManager(tmpDir, mockGit)
	// Should still succeed as loadState errors are handled gracefully
	assert.NoError(t, err)
	
	// Manager should still work with default state
	mgr.state = &TreeState{
		Nodes: map[string]*TreeNode{
			"/": {
				Name:     "root",
				Type:     NodeTypeRoot,
				Children: []string{},
			},
		},
		CurrentPath: "/",
		LastUpdated: time.Now(),
	}
	
	// Should be able to continue operations
	err = mgr.AddRepo("/", "repo2", "https://github.com/test/repo2.git", true)
	require.NoError(t, err)
}

func TestStatelessManager_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a minimal config file
	configPath := filepath.Join(tmpDir, "muno.yaml")
	configContent := `workspace:
  name: test-workspace
  repos_dir: repos
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	
	mockGit := &git.MockGit{}
	mgr, err := NewStatelessManager(tmpDir, mockGit)
	require.NoError(t, err)
	assert.NotNil(t, mgr)
	
	t.Run("GetCurrentPath", func(t *testing.T) {
		path := mgr.GetCurrentPath()
		assert.Equal(t, "/", path)
	})
	
	t.Run("UseNode with relative path", func(t *testing.T) {
		// Add a repo first
		err := mgr.AddRepo("/", "test", "https://github.com/test/test.git", true)
		require.NoError(t, err)
		
		// Use relative path
		mgr.currentPath = "/"
		err = mgr.UseNode("test")
		require.NoError(t, err)
		assert.Equal(t, "/test", mgr.currentPath)
		
		// Use .. to go back
		err = mgr.UseNode("..")
		require.NoError(t, err)
		assert.Equal(t, "/", mgr.currentPath)
	})
	
	t.Run("GetNodeByPath with empty parts", func(t *testing.T) {
		node, err := mgr.GetNodeByPath("///")
		require.NoError(t, err)
		assert.Nil(t, node) // Root node
	})
	
	t.Run("RemoveNode invalid paths", func(t *testing.T) {
		// Try to remove root
		err := mgr.RemoveNode("/")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot remove root")
		
		// Remove with empty parts
		err = mgr.RemoveNode("///")
		assert.NoError(t, err) // Doesn't error, just doesn't find anything
	})
	
	t.Run("DisplayChildren at various paths", func(t *testing.T) {
		// At root with children
		mgr.currentPath = "/"
		output := mgr.DisplayChildren()
		assert.Contains(t, output, "Children:")
		
		// Navigate to leaf
		mgr.currentPath = "/test"
		output = mgr.DisplayChildren()
		assert.Contains(t, output, "No children")
	})
	
	t.Run("DisplayStatus with various states", func(t *testing.T) {
		output := mgr.DisplayStatus()
		assert.Contains(t, output, "Tree Status")
		assert.Contains(t, output, "Current Path:")
		assert.Contains(t, output, "Workspace:")
		assert.Contains(t, output, "Repositories:")
	})
	
	t.Run("ComputeFilesystemPath with various inputs", func(t *testing.T) {
		// Empty string
		fsPath := mgr.ComputeFilesystemPath("")
		assert.Equal(t, filepath.Join(tmpDir, config.GetDefaultReposDir()), fsPath)
		
		// Multiple levels
		fsPath = mgr.ComputeFilesystemPath("/a/b/c")
		expected := filepath.Join(tmpDir, config.GetDefaultReposDir(), "a", "b", "c")
		assert.Equal(t, expected, fsPath)
		
		// Relative path (though not typical usage)
		fsPath = mgr.ComputeFilesystemPath("relative")
		assert.Equal(t, filepath.Join(tmpDir, config.GetDefaultReposDir(), "relative"), fsPath)
	})
	
	t.Run("CloneLazyRepos with config nodes", func(t *testing.T) {
		// Add a node with config reference
		mgr.config.Nodes = append(mgr.config.Nodes, config.NodeDefinition{
			Name:   "config-node",
			ConfigRef: "sub/muno.yaml",
		})
		
		// Should handle config nodes without URL
		err := mgr.CloneLazyRepos("/", false)
		assert.NoError(t, err)
	})
}

func TestTreeNode_Methods(t *testing.T) {
	node := &TreeNode{
		Name:     "test-node",
		Type:     NodeTypeRepo,
		URL:      "https://github.com/test/repo.git",
		Lazy:     true,
		State:    RepoStateMissing,
		Children: []string{"child1", "child2"},
	}
	
	t.Run("Node fields", func(t *testing.T) {
		assert.Equal(t, "test-node", node.Name)
		assert.Equal(t, NodeTypeRepo, node.Type)
		assert.Equal(t, "https://github.com/test/repo.git", node.URL)
		assert.True(t, node.Lazy)
		assert.Equal(t, RepoStateMissing, node.State)
		assert.Len(t, node.Children, 2)
	})
}

func TestConfigResolver(t *testing.T) {
	tmpDir := t.TempDir()
	resolver := NewConfigResolver(tmpDir)
	
	t.Run("Resolver creation", func(t *testing.T) {
		assert.NotNil(t, resolver)
		assert.NotNil(t, resolver.cache)
		assert.Equal(t, tmpDir, resolver.root)
	})
	
	// resolveConfigPath is private - test it through public methods
	
	t.Run("LoadNodeConfig with missing file", func(t *testing.T) {
		node := &config.NodeDefinition{
			Name:   "test",
			ConfigRef: "missing.yaml",
		}
		cfg, err := resolver.LoadNodeConfig("/non-existent", node)
		assert.Error(t, err) // Should error on missing file
		assert.Nil(t, cfg)
	})
	
	t.Run("LoadNodeConfig with valid config", func(t *testing.T) {
		// Create a config file
		configDir := filepath.Join(tmpDir, config.GetDefaultReposDir(), "test")
		os.MkdirAll(configDir, 0755)
		
		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test-sub",
				ReposDir: "nodes",
			},
			Nodes: []config.NodeDefinition{
				{
					Name: "sub-repo",
					URL:  "https://github.com/test/sub.git",
				},
			},
		}
		
		configPath := filepath.Join(configDir, "muno.yaml")
		err := cfg.Save(configPath)
		require.NoError(t, err)
		
		// Load it via BuildDistributedTree
		nodes, err := resolver.BuildDistributedTree(cfg, configDir)
		require.NoError(t, err)
		assert.NotNil(t, nodes)
		// Check root node exists
		rootNode, ok := nodes[""]
		assert.True(t, ok)
		assert.Equal(t, "test-sub", rootNode.Name)
	})
}

func TestRepoStates(t *testing.T) {
	t.Run("RepoState constants", func(t *testing.T) {
		assert.Equal(t, RepoState("missing"), RepoStateMissing)
		assert.Equal(t, RepoState("cloned"), RepoStateCloned)
		assert.Equal(t, RepoState("modified"), RepoStateModified)
	})
	
	t.Run("NodeType constants", func(t *testing.T) {
		assert.Equal(t, NodeType("root"), NodeTypeRoot)
		assert.Equal(t, NodeType("repo"), NodeTypeRepo)
		// NodeTypeGroup doesn't exist, removed
	})
}

func TestManager_DisplayMethods(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &git.MockGit{}
	
	CreateTestConfig(t, tmpDir, "")
	mgr, err := NewManager(tmpDir, mockGit)
	require.NoError(t, err)
	
	// Add some repos for display
	mgr.AddRepo("/", "repo1", "https://github.com/test/repo1.git", false)
	mgr.AddRepo("/", "repo2", "https://github.com/test/repo2.git", true)
	mgr.AddRepo("/repo1", "nested", "https://github.com/test/nested.git", false)
	
	t.Run("DisplayTree", func(t *testing.T) {
		output := mgr.DisplayTree()
		assert.Contains(t, output, "repo1")
		assert.Contains(t, output, "repo2")
		assert.Contains(t, output, "nested")
	})
	
	t.Run("DisplayTreeWithDepth", func(t *testing.T) {
		// Depth 0 - only workspace name
		output := mgr.DisplayTreeWithDepth(0)
		assert.Contains(t, output, "Workspace")
		assert.NotContains(t, output, "repo1")
		
		// Depth 1 - root and immediate children
		output = mgr.DisplayTreeWithDepth(1)
		assert.Contains(t, output, "repo1")
		assert.Contains(t, output, "repo2")
		assert.NotContains(t, output, "nested")
		
		// Depth 2 - should include nested
		output = mgr.DisplayTreeWithDepth(2)
		assert.Contains(t, output, "repo1")
		assert.Contains(t, output, "repo2")
		assert.Contains(t, output, "nested")
		
		// Depth -1 - unlimited
		output = mgr.DisplayTreeWithDepth(-1)
		assert.Contains(t, output, "repo1")
		assert.Contains(t, output, "repo2")
		assert.Contains(t, output, "nested")
	})
	
	t.Run("DisplayStatus", func(t *testing.T) {
		output := mgr.DisplayStatus()
		assert.Contains(t, output, "Tree Status")
		assert.Contains(t, output, "Current Path:")
		assert.Contains(t, output, "Repositories: 3")
	})
	
	t.Run("DisplayPath", func(t *testing.T) {
		output := mgr.DisplayPath()
		assert.Equal(t, "/", output)
		
		// Navigate and check again
		mgr.UseNode("/repo1")
		output = mgr.DisplayPath()
		assert.Equal(t, "/repo1", output)
	})
}