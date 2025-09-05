package config

import (
	"os"
	"path/filepath"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigTree(t *testing.T) {
	cfg := DefaultConfigTree("test-workspace")
	
	assert.Equal(t, "test-workspace", cfg.Workspace.Name)
	assert.Equal(t, "repos", cfg.Workspace.ReposDir)
	assert.Empty(t, cfg.Nodes)
}

func TestConfigTreeLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "muno.yaml")
	
	// Create config
	cfg := &ConfigTree{
		Workspace: WorkspaceTree{
			Name:     "test",
			ReposDir: "nodes",
		},
		Nodes: []NodeDefinition{
			{
				Name: "repo1",
				URL:  "https://github.com/test/repo1.git",
				Lazy: true,
			},
			{
				Name:   "meta1",
				Config: "meta1/muno.yaml",
			},
		},
	}
	
	// Save config
	err := cfg.Save(configPath)
	require.NoError(t, err)
	
	// Load config
	loaded, err := LoadTree(configPath)
	require.NoError(t, err)
	assert.Equal(t, "test", loaded.Workspace.Name)
	assert.Len(t, loaded.Nodes, 2)
	assert.Equal(t, "repo1", loaded.Nodes[0].Name)
	assert.Equal(t, "meta1", loaded.Nodes[1].Name)
}

func TestConfigTreeValidate(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *ConfigTree
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			cfg: &ConfigTree{
				Workspace: WorkspaceTree{
					Name:     "valid",
					ReposDir: "nodes",
				},
				Nodes: []NodeDefinition{
					{Name: "repo1", URL: "https://github.com/test/repo1.git"},
				},
			},
			wantError: false,
		},
		{
			name: "empty workspace name",
			cfg: &ConfigTree{
				Workspace: WorkspaceTree{
					Name:     "",
					ReposDir: "nodes",
				},
			},
			wantError: true,
			errorMsg:  "workspace name",
		},
		{
			name: "empty repos dir",
			cfg: &ConfigTree{
				Workspace: WorkspaceTree{
					Name:     "test",
					ReposDir: "",
				},
			},
			wantError: false, // Empty repos dir is valid, defaults to "repos"
		},
		{
			name: "node with both URL and Config",
			cfg: &ConfigTree{
				Workspace: WorkspaceTree{
					Name:     "test",
					ReposDir: "nodes",
				},
				Nodes: []NodeDefinition{
					{
						Name:   "invalid",
						URL:    "https://github.com/test/repo.git",
						Config: "config.yaml",
					},
				},
			},
			wantError: true,
			errorMsg:  "cannot have both",
		},
		{
			name: "node with neither URL nor Config",
			cfg: &ConfigTree{
				Workspace: WorkspaceTree{
					Name:     "test",
					ReposDir: "nodes",
				},
				Nodes: []NodeDefinition{
					{Name: "invalid"},
				},
			},
			wantError: true,
			errorMsg:  "must have either",
		},
		{
			name: "node without name",
			cfg: &ConfigTree{
				Workspace: WorkspaceTree{
					Name:     "test",
					ReposDir: "nodes",
				},
				Nodes: []NodeDefinition{
					{URL: "https://github.com/test/repo.git"},
				},
			},
			wantError: true,
			errorMsg:  "node name",
		},
		{
			name: "duplicate node names",
			cfg: &ConfigTree{
				Workspace: WorkspaceTree{
					Name:     "test",
					ReposDir: "nodes",
				},
				Nodes: []NodeDefinition{
					{Name: "dup", URL: "https://github.com/test/repo1.git"},
					{Name: "dup", URL: "https://github.com/test/repo2.git"},
				},
			},
			wantError: false, // Duplicate names are not checked in Validate
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetReposDir(t *testing.T) {
	cfg := &ConfigTree{
		Workspace: WorkspaceTree{
			ReposDir: "custom-repos",
		},
	}
	assert.Equal(t, "custom-repos", cfg.GetReposDir())
	
	cfg.Workspace.ReposDir = ""
	assert.Equal(t, "repos", cfg.GetReposDir())
}

func TestLoadTreeWithInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Test non-existent file
	_, err := LoadTree(filepath.Join(tmpDir, "nonexistent.yaml"))
	assert.Error(t, err)
	
	// Test invalid YAML
	invalidPath := filepath.Join(tmpDir, "invalid.yaml")
	err = os.WriteFile(invalidPath, []byte("invalid: yaml: content:"), 0644)
	require.NoError(t, err)
	
	_, err = LoadTree(invalidPath)
	assert.Error(t, err)
}