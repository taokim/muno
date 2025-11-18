package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
)

func TestTreeAdapterStub_LoadNestedConfigs(t *testing.T) {
	// Create temporary test structure
	tmpDir := t.TempDir()

	// Create root config
	rootConfig := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "team", URL: "https://example.com/team"},
		},
	}

	// Create nested directory and config
	teamDir := filepath.Join(tmpDir, ".nodes", "team")
	os.MkdirAll(teamDir, 0755)

	teamConfigContent := `workspace:
  name: team
  repos_dir: .nodes
nodes:
  - name: service
    url: https://example.com/service
`
	err := os.WriteFile(filepath.Join(teamDir, "muno.yaml"), []byte(teamConfigContent), 0644)
	require.NoError(t, err)

	// Create tree adapter with workspace context
	adapter := NewTreeAdapter().(*TreeAdapterStub)
	fsProvider := NewFileSystemAdapter()
	adapter.SetWorkspaceContext(tmpDir, fsProvider)

	// Load the tree
	err = adapter.Load(rootConfig)
	require.NoError(t, err)

	// Verify nodes were loaded
	rootNode, err := adapter.GetNode("/")
	require.NoError(t, err)
	assert.Equal(t, "test", rootNode.Name)

	teamNode, err := adapter.GetNode("/team")
	require.NoError(t, err)
	assert.Equal(t, "team", teamNode.Name)

	// Verify nested node was loaded
	serviceNode, err := adapter.GetNode("/team/service")
	require.NoError(t, err, "Nested service node should be loaded")
	assert.Equal(t, "service", serviceNode.Name)
	assert.Equal(t, "/team/service", serviceNode.Path)
}
