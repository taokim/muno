package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("test-project")
	
	assert.Equal(t, "test-project", cfg.Workspace.Name)
	assert.Equal(t, "origin", cfg.Workspace.Manifest.RemoteName)
	assert.Equal(t, "main", cfg.Workspace.Manifest.DefaultRevision)
	assert.Len(t, cfg.Workspace.Manifest.Projects, 4)
	assert.Len(t, cfg.Agents, 3)
	
	// Check default agents
	backend, exists := cfg.Agents["backend-agent"]
	assert.True(t, exists)
	assert.Equal(t, "claude-sonnet-4", backend.Model)
	assert.True(t, backend.AutoStart)
	
	// Check project-agent mapping
	backendProject := cfg.Workspace.Manifest.Projects[0]
	assert.Equal(t, "backend", backendProject.Name)
	assert.Equal(t, "backend-agent", backendProject.Agent)
}

func TestConfigSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "repo-claude.yaml")
	
	// Create and save config
	original := DefaultConfig("test-project")
	original.Workspace.Manifest.RemoteFetch = "https://github.com/test/"
	
	err := original.Save(configPath)
	require.NoError(t, err)
	
	// Load config
	loaded, err := Load(configPath)
	require.NoError(t, err)
	
	// Compare
	assert.Equal(t, original.Workspace.Name, loaded.Workspace.Name)
	assert.Equal(t, original.Workspace.Manifest.RemoteFetch, loaded.Workspace.Manifest.RemoteFetch)
	assert.Len(t, loaded.Workspace.Manifest.Projects, len(original.Workspace.Manifest.Projects))
	assert.Len(t, loaded.Agents, len(original.Agents))
}

func TestGetProjectForAgent(t *testing.T) {
	cfg := DefaultConfig("test")
	
	// Test existing agent
	project, err := cfg.GetProjectForAgent("backend-agent")
	require.NoError(t, err)
	assert.Equal(t, "backend", project.Name)
	
	// Test non-existent agent
	_, err = cfg.GetProjectForAgent("non-existent")
	assert.Error(t, err)
}

func TestConfigValidation(t *testing.T) {
	// Test loading invalid YAML
	tmpFile := filepath.Join(t.TempDir(), "invalid.yaml")
	err := os.WriteFile(tmpFile, []byte("invalid: yaml: content:"), 0644)
	require.NoError(t, err)
	
	_, err = Load(tmpFile)
	assert.Error(t, err)
	
	// Test loading non-existent file
	_, err = Load("/non/existent/path.yaml")
	assert.Error(t, err)
}