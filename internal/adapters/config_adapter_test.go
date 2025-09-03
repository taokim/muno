package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
)

func TestConfigAdapter(t *testing.T) {
	t.Run("NewConfigAdapter", func(t *testing.T) {
		adapter := NewConfigAdapter()
		assert.NotNil(t, adapter)
		// Check cache is initialized
		assert.NotNil(t, adapter.(*ConfigAdapter).cache)
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "non-existent.yaml")
		
		adapter := NewConfigAdapter()
		cfg, err := adapter.Load(configPath)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("Load invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yaml")
		
		// Write invalid YAML
		err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
		require.NoError(t, err)
		
		adapter := NewConfigAdapter()
		cfg, err := adapter.Load(configPath)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("Load valid muno.yaml config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "muno.yaml")
		
		// Write valid muno.yaml config
		configContent := `workspace:
  name: test-workspace
  repos_dir: ./repos
nodes:
  - name: repo1
    url: https://github.com/test/repo1.git
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)
		
		adapter := NewConfigAdapter()
		cfg, err := adapter.Load(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		cfgTree, ok := cfg.(*config.ConfigTree)
		assert.True(t, ok)
		assert.Equal(t, "test-workspace", cfgTree.Workspace.Name)
	})

	t.Run("Load generic YAML config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		
		// Write generic YAML config
		configContent := `name: test-config
value: 123
items:
  - one
  - two
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)
		
		adapter := NewConfigAdapter()
		cfg, err := adapter.Load(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		// It should be a generic map
		cfgMap, ok := cfg.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test-config", cfgMap["name"])
	})

	t.Run("Save config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "save-test.yaml")
		
		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "save-test",
				ReposDir: "./repos",
			},
			Nodes: []config.NodeDefinition{
				{
					Name: "repo1",
					URL:  "https://github.com/test/repo1.git",
				},
			},
		}
		
		adapter := NewConfigAdapter()
		err := adapter.Save(configPath, cfg)
		assert.NoError(t, err)
		
		// Verify file was created
		_, err = os.Stat(configPath)
		assert.NoError(t, err)
		
		// Load it back
		loaded, err := adapter.Load(configPath)
		assert.NoError(t, err)
		// Since it's not muno.yaml, it will be loaded as generic
		assert.NotNil(t, loaded)
	})

	t.Run("Save overwrites existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "overwrite-test.yaml")
		
		// Create original file
		originalContent := `name: original
value: 100
`
		err := os.WriteFile(configPath, []byte(originalContent), 0644)
		require.NoError(t, err)
		
		// Save new config
		cfg := map[string]interface{}{
			"name":  "updated",
			"value": 200,
		}
		
		adapter := NewConfigAdapter()
		err = adapter.Save(configPath, cfg)
		assert.NoError(t, err)
		
		// Verify file was overwritten
		content, err := os.ReadFile(configPath)
		assert.NoError(t, err)
		assert.Contains(t, string(content), "updated")
		assert.Contains(t, string(content), "200")
		assert.NotContains(t, string(content), "original")
	})

	t.Run("Exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Non-existent file
		configPath := filepath.Join(tmpDir, "non-existent.yaml")
		adapter := NewConfigAdapter()
		assert.False(t, adapter.Exists(configPath))
		
		// Create file
		err := os.WriteFile(configPath, []byte("test"), 0644)
		require.NoError(t, err)
		assert.True(t, adapter.Exists(configPath))
	})

	t.Run("Watch", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "watch-test.yaml")
		
		// Create initial file
		err := os.WriteFile(configPath, []byte("version: \"3\"\nname: initial"), 0644)
		require.NoError(t, err)
		
		adapter := NewConfigAdapter()
		events, err := adapter.Watch(configPath)
		// Currently returns an error as it's not implemented
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
		assert.NotNil(t, events) // Channel is created but closed
	})
	
	t.Run("Load JSON config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")
		
		// Write JSON file
		err := os.WriteFile(configPath, []byte(`{"name": "test"}`), 0644)
		require.NoError(t, err)
		
		adapter := NewConfigAdapter()
		_, err = adapter.Load(configPath)
		// JSON not implemented yet
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JSON config not yet implemented")
	})
}