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
	assert.Len(t, cfg.Workspace.Manifest.Projects, 6)
	assert.Len(t, cfg.Scopes, 5)
	
	// Check default scopes
	backend, exists := cfg.Scopes["backend"]
	assert.True(t, exists)
	assert.Equal(t, "claude-sonnet-4", backend.Model)
	assert.True(t, backend.AutoStart)
	assert.Equal(t, "Backend services development", backend.Description)
	assert.Equal(t, []string{"auth-service", "order-service", "payment-service"}, backend.Repos)
	
	// Check project groups
	authProject := cfg.Workspace.Manifest.Projects[0]
	assert.Equal(t, "auth-service", authProject.Name)
	assert.Equal(t, "backend,services", authProject.Groups)
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
	assert.Len(t, loaded.Scopes, len(original.Scopes))
}

func TestGetProjectsByScope(t *testing.T) {
	cfg := DefaultConfig("test")
	
	// Test backend scope
	backendScope := cfg.Scopes["backend"]
	assert.NotNil(t, backendScope)
	assert.Contains(t, backendScope.Repos, "auth-service")
	assert.Contains(t, backendScope.Repos, "order-service")
	assert.Contains(t, backendScope.Repos, "payment-service")
	
	// Test frontend scope
	frontendScope := cfg.Scopes["frontend"]
	assert.NotNil(t, frontendScope)
	assert.Contains(t, frontendScope.Repos, "web-app")
	assert.Contains(t, frontendScope.Repos, "mobile-app")
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