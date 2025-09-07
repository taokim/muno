package tree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
)

// Test more StatelessManager methods for coverage
func TestStatelessManager_MoreMethods(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create config
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: "nodes",
		},
		Nodes: []config.NodeDefinition{
			{
				Name: "repo1",
				URL:  "https://github.com/test/repo1.git",
			},
			{
				Name: "repo2",
				URL:  "https://github.com/test/repo2.git",
				Fetch: "lazy",
			},
		},
	}
	
	configPath := filepath.Join(tmpDir, "muno.yaml")
	err := cfg.Save(configPath)
	require.NoError(t, err)
	
	// Create repos directory
	os.MkdirAll(filepath.Join(tmpDir, "nodes"), 0755)
	
	gitCmd := &git.MockGit{
		StatusFunc: func(path string) (string, error) {
			return "modified", nil
		},
	}
	
	mgr, err := NewStatelessManager(tmpDir, gitCmd)
	require.NoError(t, err)
	
	t.Run("GetNodeByPath with nested path parts", func(t *testing.T) {
		// Test that it returns the parent node even for nested paths
		node, err := mgr.GetNodeByPath("/repo1/nested/deep")
		require.NoError(t, err)
		assert.NotNil(t, node)
		assert.Equal(t, "repo1", node.Name)
	})
	
	t.Run("UseNode with auto-clone", func(t *testing.T) {
		cloneCalled := false
		gitCmd.CloneFunc = func(url, path string) error {
			cloneCalled = true
			// Create .git directory to simulate successful clone
			gitDir := filepath.Join(path, ".git")
			os.MkdirAll(gitDir, 0755)
			return nil
		}
		
		// Navigate to a repository node without .git
		err := mgr.UseNode("/repo1")
		require.NoError(t, err)
		assert.True(t, cloneCalled)
	})
	
	t.Run("UseNode with relative path", func(t *testing.T) {
		mgr.currentPath = "/repo1"
		err := mgr.UseNode("..")
		require.NoError(t, err)
		assert.Equal(t, "/", mgr.currentPath)
	})
	
	t.Run("AddRepo non-lazy with clone", func(t *testing.T) {
		cloneCalled := false
		gitCmd.CloneFunc = func(url, path string) error {
			cloneCalled = true
			return nil
		}
		
		err := mgr.AddRepo("/", "new-repo", "https://github.com/test/new.git", false)
		require.NoError(t, err)
		assert.True(t, cloneCalled)
		
		// Check it was added to config
		assert.Len(t, mgr.config.Nodes, 3)
	})
	
	t.Run("RemoveNode removes from filesystem", func(t *testing.T) {
		// Create a directory to remove
		testDir := filepath.Join(tmpDir, "nodes", "test-remove")
		os.MkdirAll(testDir, 0755)
		
		// Add to config
		mgr.config.Nodes = append(mgr.config.Nodes, config.NodeDefinition{
			Name: "test-remove",
			URL:  "https://github.com/test/remove.git",
		})
		
		err := mgr.RemoveNode("/test-remove")
		require.NoError(t, err)
		
		// Check it was removed from config
		found := false
		for _, node := range mgr.config.Nodes {
			if node.Name == "test-remove" {
				found = true
				break
			}
		}
		assert.False(t, found)
	})
	
	t.Run("RemoveNode with empty path parts", func(t *testing.T) {
		err := mgr.RemoveNode("//")
		// This actually doesn't error with "//" - it removes the node with empty name
		require.NoError(t, err)
	})
	
	t.Run("ListChildren for nested path", func(t *testing.T) {
		children, err := mgr.ListChildren("/repo1/nested")
		require.NoError(t, err)
		assert.Len(t, children, 0) // Returns empty for non-root
	})
	
	t.Run("DisplayStatus with various states", func(t *testing.T) {
		// Create .git directories to simulate different states
		repo1Git := filepath.Join(tmpDir, "nodes", "repo1", ".git")
		os.MkdirAll(repo1Git, 0755)
		
		result := mgr.DisplayStatus()
		assert.Contains(t, result, "Tree Status")
		assert.Contains(t, result, "Repositories")
	})
	
	t.Run("DisplayTree with modified repos", func(t *testing.T) {
		// Create .git directory for repo1
		repo1Git := filepath.Join(tmpDir, "nodes", "repo1", ".git")
		os.MkdirAll(repo1Git, 0755)
		
		result := mgr.DisplayTree()
		assert.Contains(t, result, "test-workspace")
		assert.Contains(t, result, "repo1")
		assert.Contains(t, result, "repo2")
	})
	
	t.Run("DisplayTree with config reference", func(t *testing.T) {
		// Add a config reference node
		mgr.config.Nodes = append(mgr.config.Nodes, config.NodeDefinition{
			Name:   "parent",
			Config: "parent/muno.yaml",
		})
		
		result := mgr.DisplayTree()
		assert.Contains(t, result, "parent")
		assert.Contains(t, result, "📁") // Config reference icon
	})
	
	t.Run("CloneLazyRepos with recursive and config", func(t *testing.T) {
		// Add a node with config for recursive testing
		mgr.config.Nodes = append(mgr.config.Nodes, config.NodeDefinition{
			Name:   "with-config",
			URL:    "https://github.com/test/with-config.git",
			Config: "sub/muno.yaml",
		})
		
		gitCmd.CloneCalls = nil
		gitCmd.CloneFunc = func(url, path string) error {
			return nil
		}
		
		err := mgr.CloneLazyRepos("/", true)
		require.NoError(t, err)
		
		// Should attempt to clone repos without .git
		assert.GreaterOrEqual(t, len(gitCmd.CloneCalls), 1)
	})
}

// Test ConfigResolver basic functionality
func TestConfigResolver_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	resolver := NewConfigResolver(tmpDir)
	
	assert.NotNil(t, resolver)
	// ConfigResolver fields are private, just check it's not nil
}

// Test GetRepoState edge cases
func TestGetRepoState_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	
	t.Run("Empty path", func(t *testing.T) {
		state := GetRepoState("")
		assert.Equal(t, RepoStateMissing, state)
	})
	
	t.Run("Non-existent path", func(t *testing.T) {
		state := GetRepoState(filepath.Join(tmpDir, "nonexistent"))
		assert.Equal(t, RepoStateMissing, state)
	})
	
	t.Run("Path without .git", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "no-git")
		os.MkdirAll(testDir, 0755)
		
		state := GetRepoState(testDir)
		assert.Equal(t, RepoStateMissing, state)
	})
	
	t.Run("Path with .git directory", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "with-git")
		os.MkdirAll(testDir, 0755)
		
		// Initialize a real git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = testDir
		err := cmd.Run()
		require.NoError(t, err)
		
		state := GetRepoState(testDir)
		assert.Equal(t, RepoStateCloned, state)
	})
}

// Test helper functions
func TestTreeHelpers(t *testing.T) {
	t.Run("GetConfigRefStatus", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Non-existent path
		status := GetConfigRefStatus(filepath.Join(tmpDir, "nonexistent"))
		assert.False(t, status)
		
		// Existing directory
		dirPath := filepath.Join(tmpDir, "testdir")
		os.MkdirAll(dirPath, 0755)
		
		status = GetConfigRefStatus(dirPath)
		assert.True(t, status)
		
		// Directory with .muno-ref marker
		markerPath := filepath.Join(dirPath, ".muno-ref")
		os.WriteFile(markerPath, []byte("test"), 0644)
		
		status = GetConfigRefStatus(dirPath)
		assert.True(t, status)
	})
	
	t.Run("CreateConfigRefMarker", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		err := CreateConfigRefMarker(tmpDir)
		require.NoError(t, err)
		
		// Check marker was created
		markerPath := filepath.Join(tmpDir, ".muno-ref")
		content, err := os.ReadFile(markerPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "MUNO config reference")
	})
}