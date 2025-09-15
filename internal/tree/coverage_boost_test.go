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

// TestUncoveredDisplayFunctions tests the uncovered display functions
func TestUncoveredDisplayFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &git.MockGit{}
	
	// Create config
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: "nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git"},
			{Name: "repo2", URL: "https://github.com/test/repo2.git"},
		},
	}
	
	configPath := filepath.Join(tmpDir, "muno.yaml")
	err := cfg.Save(configPath)
	require.NoError(t, err)
	
	t.Run("DisplayChildren", func(t *testing.T) {
		mgr, err := NewManager(tmpDir, mockGit)
		require.NoError(t, err)
		
		// Test DisplayChildren
		result := mgr.DisplayChildren()
		assert.Contains(t, result, "repo1")
		assert.Contains(t, result, "repo2")
	})
	
	t.Run("StatelessManager_DisplayChildren", func(t *testing.T) {
		mgr, err := NewStatelessManager(tmpDir, mockGit)
		require.NoError(t, err)
		
		result := mgr.DisplayChildren()
		assert.Contains(t, result, "Children")
		})
}

// TestLoadConfigReference tests the loadConfigReference function
func TestLoadConfigReference(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &git.MockGit{}
	
	// Create main config
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "main-workspace",
			ReposDir: "nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "config-ref", ConfigRef: "sub.yaml"},
		},
	}
	
	configPath := filepath.Join(tmpDir, "muno.yaml")
	err := cfg.Save(configPath)
	require.NoError(t, err)
	
	// Create sub config
	subDir := filepath.Join(tmpDir, "nodes", "config-ref")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)
	
	subCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name: "sub-workspace",
		},
		Nodes: []config.NodeDefinition{
			{Name: "sub-repo", URL: "https://github.com/test/sub.git"},
		},
	}
	
	subConfigPath := filepath.Join(subDir, "sub.yaml")
	err = subCfg.Save(subConfigPath)
	require.NoError(t, err)
	
	// Test with manager that loads config references
	mgr, err := NewManager(tmpDir, mockGit)
	require.NoError(t, err)
	
	// Navigate to trigger config loading
	err = mgr.UseNode("/config-ref")
	assert.NoError(t, err)
}

// TestSaveStateEdgeCases tests edge cases in saveState
func TestSaveStateEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &git.MockGit{}
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: "nodes",
		},
	}
	
	configPath := filepath.Join(tmpDir, "muno.yaml")
	err := cfg.Save(configPath)
	require.NoError(t, err)
	
	mgr, err := NewManager(tmpDir, mockGit)
	require.NoError(t, err)
	
	// Make state file read-only to trigger save error
	stateFile := filepath.Join(tmpDir, ".muno-state.json")
	err = os.WriteFile(stateFile, []byte("{}"), 0444)
	require.NoError(t, err)
	
	// Try to save state (should handle error gracefully)
	mgr.saveState()
	
	// Clean up
	os.Chmod(stateFile, 0644)
}

// TestCloneLazyReposEdgeCases tests edge cases in CloneLazyRepos
func TestCloneLazyReposEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &git.MockGit{
		CloneFunc: func(url, path string) error {
			// Simulate successful clone
			return os.MkdirAll(filepath.Join(path, ".git"), 0755)
		},
	}
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: "nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "lazy1", URL: "https://github.com/test/lazy1.git", Fetch: config.FetchLazy},
			{Name: "eager", URL: "https://github.com/test/eager.git"},
			{Name: "config-node", ConfigRef: "sub.yaml"},
		},
	}
	
	configPath := filepath.Join(tmpDir, "muno.yaml")
	err := cfg.Save(configPath)
	require.NoError(t, err)
	
	mgr, err := NewManager(tmpDir, mockGit)
	require.NoError(t, err)
	
	// Clone at specific path - CloneLazyRepos returns error, not list
	err = mgr.CloneLazyRepos("/lazy1", false)
	assert.NoError(t, err)
	
	// Clone with config node - should succeed even if it's a config node
	err = mgr.CloneLazyRepos("/config-node", false)
	assert.NoError(t, err)
}

// TestTreeNodeMethodsComplete tests all TreeNode methods thoroughly
func TestTreeNodeMethodsComplete(t *testing.T) {
	t.Run("NeedsClone comprehensive", func(t *testing.T) {
		// Node needs clone: is a repo and state is missing
		node := &TreeNode{
			Type:  NodeTypeRepo,
			URL:   "https://github.com/test/repo.git",
			Lazy:  true,
			State: RepoStateMissing,
		}
		assert.True(t, node.NeedsClone())
		
		// Already cloned
		node.State = RepoStateCloned
		assert.False(t, node.NeedsClone())
		
		// Root node type (not repo)
		node.Type = NodeTypeRoot
		node.State = RepoStateMissing
		assert.False(t, node.NeedsClone())
		
		// Config node type
		node = &TreeNode{Type: NodeTypeConfig, State: RepoStateMissing}
		assert.False(t, node.NeedsClone())
	})
	
	t.Run("HasLazyRepos comprehensive", func(t *testing.T) {
		// Node itself is lazy repo with missing state
		node := &TreeNode{
			Type:  NodeTypeRepo,
			URL:   "repo.git",
			Lazy:  true,
			State: RepoStateMissing,
		}
		assert.True(t, node.HasLazyRepos())
		
		// Node has lazy children - but we can't test this since Children is []string
		// Just ensure the method works
		node = &TreeNode{
			Children: []string{"child1", "child2"},
		}
		// This will return false since we can't check children's lazy status
		assert.False(t, node.HasLazyRepos())
	})
}

// TestResolverFunctionsFixed tests resolver functions with correct expectations
func TestResolverFunctionsFixed(t *testing.T) {
	t.Run("IsMetaRepo fixed", func(t *testing.T) {
		// The function checks for suffix patterns
		assert.True(t, IsMetaRepo("my-monorepo"))  // Ends with -monorepo
		assert.True(t, IsMetaRepo("team-workspace")) // Ends with -workspace
		assert.False(t, IsMetaRepo("root-repo"))    // Doesn't end with -root-repo
		
		// These don't match patterns
		assert.False(t, IsMetaRepo("regular-repo"))
		assert.False(t, IsMetaRepo(""))
	})
	
	t.Run("GetEffectiveLazy", func(t *testing.T) {
		// Test with explicit lazy
		node := &config.NodeDefinition{
			Name:  "test",
			Fetch: config.FetchLazy,
		}
		assert.True(t, GetEffectiveLazy(node))
		
		// Test with explicit eager
		node.Fetch = config.FetchEager
		assert.False(t, GetEffectiveLazy(node))
		
		// Test default (auto mode - depends on name patterns)
		node.Fetch = ""
		node.Name = "regular-repo"
		assert.True(t, GetEffectiveLazy(node))
		
		// Test meta-repo pattern with auto mode
		node.Name = "test-monorepo"
		assert.False(t, GetEffectiveLazy(node))
	})
	
	t.Run("AutoDiscoverConfig", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// No config file
		configFile, found := AutoDiscoverConfig(tmpDir)
		assert.False(t, found)
		assert.Empty(t, configFile)
		
		// Create muno.yaml
		munoPath := filepath.Join(tmpDir, "muno.yaml")
		err := os.WriteFile(munoPath, []byte("test"), 0644)
		require.NoError(t, err)
		
		configFile, found = AutoDiscoverConfig(tmpDir)
		assert.True(t, found)
		assert.Equal(t, munoPath, configFile)
		
		// Test priority: muno.yaml > .muno.yaml
		hiddenPath := filepath.Join(tmpDir, ".muno.yaml")
		err = os.WriteFile(hiddenPath, []byte("test"), 0644)
		require.NoError(t, err)
		
		configFile, found = AutoDiscoverConfig(tmpDir)
		assert.True(t, found)
		assert.Equal(t, munoPath, configFile) // Prefers non-hidden
		
		// Remove muno.yaml, should find .muno.yaml
		os.Remove(munoPath)
		configFile, found = AutoDiscoverConfig(tmpDir)
		assert.True(t, found)
		assert.Equal(t, hiddenPath, configFile)
	})
}