package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test ConfigTree methods
func TestConfigTree_Methods(t *testing.T) {
	t.Run("GetReposDir", func(t *testing.T) {
		cfg := &ConfigTree{
			Workspace: WorkspaceTree{
				ReposDir: "custom-repos",
			},
		}
		assert.Equal(t, "custom-repos", cfg.GetReposDir())
		
		// Test with empty ReposDir
		cfg.Workspace.ReposDir = ""
		assert.Equal(t, "repos", cfg.GetReposDir())
	})
	

}

// Test DefaultConfigTree is already in tree_test.go

// Test GetDefaultReposDir
func TestGetDefaultReposDir(t *testing.T) {
	assert.Equal(t, "repos", GetDefaultReposDir())
}

// Test LoadTree
func TestLoadTree(t *testing.T) {
	t.Run("ValidYAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "muno.yaml")
		
		content := `
workspace:
  name: test-workspace
  repos_dir: custom-repos
nodes:
  - name: backend
    url: https://github.com/test/backend.git
    fetch: lazy
  - name: frontend
    url: https://github.com/test/frontend.git
    fetch: eager
  - name: config-node
    config_ref: ../shared/muno.yaml
`
		err := os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)
		
		cfg, err := LoadTree(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "test-workspace", cfg.Workspace.Name)
		assert.Equal(t, "custom-repos", cfg.Workspace.ReposDir)
		assert.Len(t, cfg.Nodes, 3)
		
		// Check nodes
		assert.Equal(t, "backend", cfg.Nodes[0].Name)
		assert.Equal(t, "https://github.com/test/backend.git", cfg.Nodes[0].URL)
		assert.Equal(t, "lazy", cfg.Nodes[0].Fetch)
		
		assert.Equal(t, "frontend", cfg.Nodes[1].Name)
		assert.Equal(t, "eager", cfg.Nodes[1].Fetch)
		
		assert.Equal(t, "config-node", cfg.Nodes[2].Name)
		assert.Equal(t, "../shared/muno.yaml", cfg.Nodes[2].ConfigRef)
	})
	
	t.Run("InvalidYAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "muno.yaml")
		
		content := `
workspace:
  name: [invalid
`
		err := os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)
		
		cfg, err := LoadTree(configPath)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
	
	t.Run("FileNotFound", func(t *testing.T) {
		cfg, err := LoadTree("/nonexistent/file.yaml")
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
	
	t.Run("EmptyFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "muno.yaml")
		
		err := os.WriteFile(configPath, []byte(""), 0644)
		require.NoError(t, err)
		
		cfg, err := LoadTree(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		// Empty file gets default workspace name
		assert.Equal(t, "muno-workspace", cfg.Workspace.Name)
	})
}

// Test Save
func TestConfigTree_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "muno.yaml")
		
		cfg := &ConfigTree{
			Workspace: WorkspaceTree{
				Name:     "test-workspace",
				ReposDir: "custom-repos",
			},
			Nodes: []NodeDefinition{
				{
					Name:  "backend",
					URL:   "https://github.com/test/backend.git",
					Fetch: "lazy",
				},
				{
					Name:   "config-ref",
					ConfigRef: "../shared/muno.yaml",
				},
			},
		}
		
		err := cfg.Save(configPath)
		assert.NoError(t, err)
		
		// Verify the file was created
		_, err = os.Stat(configPath)
		assert.NoError(t, err)
		
		// Load it back and verify
		loaded, err := LoadTree(configPath)
		assert.NoError(t, err)
		assert.Equal(t, cfg.Workspace.Name, loaded.Workspace.Name)
		assert.Equal(t, cfg.Workspace.ReposDir, loaded.Workspace.ReposDir)
		assert.Len(t, loaded.Nodes, 2)
	})
	
	t.Run("InvalidPath", func(t *testing.T) {
		cfg := &ConfigTree{
			Workspace: WorkspaceTree{
				Name: "test",
			},
		}
		
		err := cfg.Save("/nonexistent/dir/muno.yaml")
		assert.Error(t, err)
	})
}

// Test Validate
func TestConfigTree_Validate(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		cfg := &ConfigTree{
			Workspace: WorkspaceTree{
				Name: "test-workspace",
			},
			Nodes: []NodeDefinition{
				{
					Name: "backend",
					URL:  "https://github.com/test/backend.git",
				},
				{
					Name:   "config-ref",
					ConfigRef: "../shared/muno.yaml",
				},
			},
		}
		
		err := cfg.Validate()
		assert.NoError(t, err)
	})
	
	t.Run("InvalidNode_BothURLAndConfig", func(t *testing.T) {
		cfg := &ConfigTree{
			Workspace: WorkspaceTree{
				Name: "test-workspace",
			},
			Nodes: []NodeDefinition{
				{
					Name:   "invalid",
					URL:    "https://github.com/test/repo.git",
					ConfigRef: "../shared/muno.yaml",
				},
			},
		}
		
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot have both URL and config")
	})
	
	t.Run("InvalidNode_NeitherURLNorConfig", func(t *testing.T) {
		cfg := &ConfigTree{
			Workspace: WorkspaceTree{
				Name: "test-workspace",
			},
			Nodes: []NodeDefinition{
				{
					Name: "invalid",
				},
			},
		}
		
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must have either URL or config")
	})
	
	t.Run("InvalidNode_NoName", func(t *testing.T) {
		cfg := &ConfigTree{
			Workspace: WorkspaceTree{
				Name: "test-workspace",
			},
			Nodes: []NodeDefinition{
				{
					URL: "https://github.com/test/repo.git",
				},
			},
		}
		
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})
	
	t.Run("InvalidFetchMode", func(t *testing.T) {
		// Note: Currently the config doesn't validate fetch mode
		// This test is for future implementation
		cfg := &ConfigTree{
			Workspace: WorkspaceTree{
				Name: "test-workspace",
			},
			Nodes: []NodeDefinition{
				{
					Name:  "backend",
					URL:   "https://github.com/test/backend.git",
					Fetch: "invalid", // This is currently allowed
				},
			},
		}
		
		err := cfg.Validate()
		// Currently doesn't validate fetch mode, so no error
		assert.NoError(t, err)
	})
}

// Test GetConfigFileNames
func TestGetConfigFileNames(t *testing.T) {
	names := GetConfigFileNames()
	assert.Contains(t, names, "muno.yaml")
	assert.Contains(t, names, "muno.yml")
	assert.Contains(t, names, ".muno.yaml")
	assert.Contains(t, names, ".muno.yml")
}

