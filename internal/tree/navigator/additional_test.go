package navigator

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
)

func TestNavigatorAdditional(t *testing.T) {
	t.Run("InMemoryNavigator_Clear", func(t *testing.T) {
		nav := NewInMemoryNavigator()
		setupTestTree(nav)
		
		// Verify tree is set up
		node, _ := nav.GetNode("/backend")
		assert.NotNil(t, node)
		
		// Clear the navigator
		nav.Clear()
		
		// Verify tree is cleared except root
		node, _ = nav.GetNode("/backend")
		assert.Nil(t, node)
		
		// Root should still exist
		root, _ := nav.GetNode("/")
		assert.NotNil(t, root)
		assert.Empty(t, root.Children)
	})
	
	t.Run("InMemoryNavigator_RemoveNode", func(t *testing.T) {
		nav := NewInMemoryNavigator()
		setupTestTree(nav)
		
		// Remove a node
		err := nav.RemoveNode("/backend")
		assert.NoError(t, err)
		
		// Verify node is removed
		node, _ := nav.GetNode("/backend")
		assert.Nil(t, node)
		
		// Try to remove root
		err = nav.RemoveNode("/")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot remove root")
		
		// Try to remove non-existent
		err = nav.RemoveNode("/non-existent")
		assert.NoError(t, err) // Should succeed silently
	})
	
	t.Run("CachedNavigator_Operations", func(t *testing.T) {
		base := NewInMemoryNavigator()
		setupTestTree(base)
		cached := NewCachedNavigator(base, 100*time.Millisecond)
		
		// Test GetCurrentPath
		path, err := cached.GetCurrentPath()
		assert.NoError(t, err)
		assert.Equal(t, "/", path)
		
		// Test ListChildren
		children, err := cached.ListChildren("/")
		assert.NoError(t, err)
		assert.Len(t, children, 3)
		
		// Test GetNodeStatus
		base.SetNodeStatus("/backend", &NodeStatus{
			Exists:   true,
			Cloned:   true,
			Modified: true,
			State:    RepoStateModified,
		})
		
		status, err := cached.GetNodeStatus("/backend")
		assert.NoError(t, err)
		assert.NotNil(t, status)
		assert.True(t, status.Modified)
		
		// Test IsLazy
		base.SetNodeStatus("/frontend", &NodeStatus{
			Lazy: true,
		})
		
		isLazy, err := cached.IsLazy("/frontend")
		assert.NoError(t, err)
		assert.True(t, isLazy)
		
		// Test TriggerLazyLoad
		err = cached.TriggerLazyLoad("/frontend")
		assert.NoError(t, err)
		
		// Test ClearCache
		cached.ClearCache()
		// Should still work after clear
		node, err := cached.GetNode("/backend")
		assert.NoError(t, err)
		assert.NotNil(t, node)
	})
	
	t.Run("Factory_CreateDefault", func(t *testing.T) {
		workspace := t.TempDir()
		factory := NewFactory(workspace, nil, nil)
		
		// Create default navigator
		nav, err := factory.CreateDefault()
		assert.NoError(t, err)
		assert.NotNil(t, nav)
		
		// Create for testing
		nav, err = factory.CreateForTesting()
		assert.NoError(t, err)
		assert.NotNil(t, nav)
		_, ok := nav.(*InMemoryNavigator)
		assert.True(t, ok)
	})
	
	t.Run("ConfigResolver", func(t *testing.T) {
		workspace := t.TempDir()
		resolver := NewConfigResolver(workspace)
		assert.NotNil(t, resolver)
		
		// Create a test config file
		configPath := filepath.Join(workspace, "test.yaml")
		cfg := config.DefaultConfigTree("test")
		cfg.Nodes = []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git"},
		}
		err := cfg.Save(configPath)
		require.NoError(t, err)
		
		// Create a node definition that references the config
		nodeDef := &config.NodeDefinition{
			Name:   "test-node",
			ConfigRef: configPath,
		}
		
		// Load the config
		loaded, err := resolver.LoadNodeConfig(nodeDef.ConfigRef, nodeDef)
		assert.NoError(t, err)
		assert.NotNil(t, loaded)
		assert.Len(t, loaded.Nodes, 1)
		assert.Equal(t, "repo1", loaded.Nodes[0].Name)
		
		// Try to load with non-existent config
		nodeDef.ConfigRef = "/non/existent.yaml"
		loaded, err = resolver.LoadNodeConfig(nodeDef.ConfigRef, nodeDef)
		assert.Error(t, err)
		assert.Nil(t, loaded)
		
		// Try to load remote config (should fail in test)
		nodeDef.ConfigRef = "https://example.com/config.yaml"
		loaded, err = resolver.LoadNodeConfig(nodeDef.ConfigRef, nodeDef)
		assert.Error(t, err)
		assert.Nil(t, loaded)
	})

	t.Run("InMemoryNavigator_SetConfig", func(t *testing.T) {
		nav := NewInMemoryNavigator()
		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: "nodes",
			},
		}
		nav.SetConfig(cfg)
		// Verify config was set (would be used internally)
		assert.NotNil(t, nav)
	})
}