package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/adapters"
	"github.com/taokim/muno/internal/config"
)

func TestDebug_TreeLoading(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create config with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "team", URL: "https://example.com/team"},
		},
	}
	
	// Create nested structure for team repo
	teamRepoConfig := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "team",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "service", URL: "https://example.com/service"},
		},
	}
	tw.AddRepositoryWithConfig("team", teamRepoConfig)
	tw.AddRepository("team/.nodes/service")
	
	// Create manager with config
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	
	// Check if muno.yaml exists at team
	teamMunoPath := filepath.Join(tw.NodesDir, "team", "muno.yaml")
	t.Logf("Checking for muno.yaml at: %s", teamMunoPath)
	_, err := os.Stat(teamMunoPath)
	require.NoError(t, err, "team muno.yaml should exist")
	
	// Check if tree has the nodes
	rootNode, err := m.treeProvider.GetNode("/")
	require.NoError(t, err)
	t.Logf("Root node: %+v", rootNode)
	
	teamNode, err := m.treeProvider.GetNode("/team")
	require.NoError(t, err)
	t.Logf("Team node: %+v", teamNode)
	
	serviceNode, err := m.treeProvider.GetNode("/team/service")
	if err != nil {
		t.Logf("Service node NOT FOUND: %v", err)
		
		// List all nodes in tree
		t.Logf("All nodes in tree:")
		if stubAdapter, ok := m.treeProvider.(*adapters.TreeAdapterStub); ok {
			for path := range stubAdapter.DebugNodes() {
				t.Logf("  - %s", path)
			}
		}
	} else {
		t.Logf("Service node: %+v", serviceNode)
	}
}
