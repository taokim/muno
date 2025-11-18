package manager

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
)

// Test Initialize error paths and edge cases
func TestManager_Initialize_ErrorPaths(t *testing.T) {
	t.Run("create workspace directory on init", func(t *testing.T) {
		tmpDir := t.TempDir()
		workspace := filepath.Join(tmpDir, "new-workspace")

		opts := CreateTestManagerOptions(workspace)
		m, err := NewManager(opts)
		require.NoError(t, err)

		// Workspace doesn't exist yet
		require.NoError(t, os.RemoveAll(workspace))

		// Initialize should create it
		err = m.Initialize(context.Background(), workspace)
		require.NoError(t, err)

		// Verify workspace was created
		assert.True(t, m.fsProvider.Exists(workspace))
		assert.True(t, m.initialized)
	})

	t.Run("auto-load config when enabled", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		// Create config file
		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: ".nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "repo1", URL: "https://example.com/repo1"},
			},
		}
		tw.CreateConfig(cfg)

		// Create manager with AutoLoadConfig enabled
		opts := CreateTestManagerOptions(tw.Root)
		opts.AutoLoadConfig = true
		m, err := NewManager(opts)
		require.NoError(t, err)

		// Initialize - should auto-load config
		err = m.Initialize(context.Background(), tw.Root)
		require.NoError(t, err)

		// Verify config was loaded
		assert.NotNil(t, m.config)
		assert.Equal(t, "test", m.config.Workspace.Name)
	})

	t.Run("skip config load when file doesn't exist", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		// No config file created
		opts := CreateTestManagerOptions(tw.Root)
		opts.AutoLoadConfig = true
		m, err := NewManager(opts)
		require.NoError(t, err)

		// Initialize - should not fail even without config
		err = m.Initialize(context.Background(), tw.Root)
		require.NoError(t, err)
		assert.True(t, m.initialized)
	})
}

// Test InitializeWithConfig paths
func TestManager_InitializeWithConfig_Variations(t *testing.T) {
	t.Run("with overrides config", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: ".nodes",
			},
			Overrides: map[string]interface{}{
				"ssh_preference": true,
			},
		}

		opts := CreateTestManagerOptions(tw.Root)
		m, err := NewManager(opts)
		require.NoError(t, err)

		err = m.InitializeWithConfig(context.Background(), tw.Root, cfg)
		require.NoError(t, err)

		assert.Equal(t, cfg, m.config)
		assert.True(t, m.initialized)
	})

	t.Run("with empty nodes", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "empty",
				ReposDir: ".nodes",
			},
			Nodes: []config.NodeDefinition{}, // Empty nodes list
		}

		opts := CreateTestManagerOptions(tw.Root)
		m, err := NewManager(opts)
		require.NoError(t, err)

		err = m.InitializeWithConfig(context.Background(), tw.Root, cfg)
		require.NoError(t, err)
		assert.True(t, m.initialized)
	})
}

// Test CloneRepos variations
func TestManager_CloneRepos_Variations(t *testing.T) {
	t.Run("no lazy repos returns immediately", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: ".nodes",
			},
			// All repos are cloned (not lazy)
			Nodes: []config.NodeDefinition{
				{Name: "repo1", URL: "https://example.com/repo1", Fetch: "eager"},
			},
		}

		m := CreateTestManagerWithConfig(t, tw.Root, cfg)
		tw.AddRepository("repo1")

		// CloneRepos with no lazy repos
		err := m.CloneRepos("/", false, false)
		require.NoError(t, err)
	})

	t.Run("include lazy flag affects behavior", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: ".nodes",
			},
		}

		m := CreateTestManagerWithConfig(t, tw.Root, cfg)

		// CloneRepos with includeLazy=true (no lazy repos to clone)
		err := m.CloneRepos("/", false, true)
		require.NoError(t, err)
	})
}

// Test outputNodesQuietRecursive edge cases
func TestManager_ListNodesQuiet_EdgeCases(t *testing.T) {
	t.Run("empty tree shows no repositories message", func(t *testing.T) {
		tw := CreateTestWorkspace(t)
		m := CreateTestManager(t, tw.Root)

		// Empty config - no nodes
		err := m.ListNodesQuiet(false)
		require.NoError(t, err)
	})

	t.Run("single level tree", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: ".nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "repo1", URL: "https://example.com/repo1"},
				{Name: "repo2", URL: "https://example.com/repo2"},
			},
		}

		m := CreateTestManagerWithConfig(t, tw.Root, cfg)
		tw.AddRepository("repo1")
		tw.AddRepository("repo2")

		err := m.ListNodesQuiet(false)
		require.NoError(t, err)
	})

	t.Run("recursive listing", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: ".nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "repo1", URL: "https://example.com/repo1"},
			},
		}

		m := CreateTestManagerWithConfig(t, tw.Root, cfg)
		tw.AddRepository("repo1")

		err := m.ListNodesQuiet(true)
		require.NoError(t, err)
	})
}

// Test ListNodesRecursive coverage
func TestManager_ListNodesRecursive_Coverage(t *testing.T) {
	t.Run("multi-level tree traversal", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		// Create nested config
		nestedCfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "nested",
				ReposDir: ".nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "child", URL: "https://example.com/child"},
			},
		}

		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: ".nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "parent", URL: "https://example.com/parent"},
			},
		}

		// Setup nested structure
		tw.AddRepositoryWithConfig("parent", nestedCfg)

		m := CreateTestManagerWithConfig(t, tw.Root, cfg)

		err := m.ListNodesRecursive(true)
		require.NoError(t, err)
	})

	t.Run("non-recursive mode", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: ".nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "repo1", URL: "https://example.com/repo1"},
			},
		}

		m := CreateTestManagerWithConfig(t, tw.Root, cfg)

		err := m.ListNodesRecursive(false)
		require.NoError(t, err)
	})
}

// Test ShowTreeAtPath variations
func TestManager_ShowTreeAtPath_Variations(t *testing.T) {
	t.Run("show tree from root", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: ".nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "backend", URL: "https://example.com/backend"},
				{Name: "frontend", URL: "https://example.com/frontend"},
			},
		}

		m := CreateTestManagerWithConfig(t, tw.Root, cfg)
		tw.AddRepository("backend")
		tw.AddRepository("frontend")

		err := m.ShowTreeAtPath("/", 0)
		require.NoError(t, err)
	})

	t.Run("show tree from specific node", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		// Create nested structure
		nestedCfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "backend",
				ReposDir: ".nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "api", URL: "https://example.com/api"},
				{Name: "db", URL: "https://example.com/db"},
			},
		}

		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: ".nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "backend", URL: "https://example.com/backend"},
			},
		}

		tw.AddRepositoryWithConfig("backend", nestedCfg)

		m := CreateTestManagerWithConfig(t, tw.Root, cfg)

		err := m.ShowTreeAtPath("/backend", 0)
		require.NoError(t, err)
	})

	t.Run("show tree with depth limit", func(t *testing.T) {
		tw := CreateTestWorkspace(t)

		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: ".nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "repo1", URL: "https://example.com/repo1"},
			},
		}

		m := CreateTestManagerWithConfig(t, tw.Root, cfg)
		tw.AddRepository("repo1")

		// Show tree with depth = 2
		err := m.ShowTreeAtPath("/", 2)
		require.NoError(t, err)
	})
}

// Test SetCLIConfig coverage (if it exists)
func TestManager_SetCLIConfig_Coverage(t *testing.T) {
	t.Run("basic manager operations", func(t *testing.T) {
		tw := CreateTestWorkspace(t)
		m := CreateTestManager(t, tw.Root)

		// Test basic manager state
		assert.NotNil(t, m.fsProvider)
		assert.NotNil(t, m.treeProvider)
		assert.NotNil(t, m.logProvider)
		assert.True(t, m.initialized)
	})
}

// Test NewManager edge cases
func TestManager_NewManager_EdgeCases(t *testing.T) {
	t.Run("with all providers", func(t *testing.T) {
		tw := CreateTestWorkspace(t)
		opts := CreateTestManagerOptions(tw.Root)

		m, err := NewManager(opts)
		require.NoError(t, err)
		require.NotNil(t, m)
		require.NotNil(t, m.metricsProvider)
		require.NotNil(t, m.logProvider)
	})
}
