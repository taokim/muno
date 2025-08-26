package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("test-project")
	
	if cfg.Version != 2 {
		t.Errorf("Expected version 2, got %d", cfg.Version)
	}
	
	if cfg.Workspace.Name != "test-project" {
		t.Errorf("Expected workspace name 'test-project', got %s", cfg.Workspace.Name)
	}
	
	if !cfg.Workspace.IsolationMode {
		t.Error("Expected isolation mode to be true")
	}
	
	if cfg.Workspace.BasePath != "workspaces" {
		t.Errorf("Expected base path 'workspaces', got %s", cfg.Workspace.BasePath)
	}
	
	// Check repositories
	if len(cfg.Repositories) == 0 {
		t.Error("Expected default repositories")
	}
	
	// Check scopes
	if len(cfg.Scopes) == 0 {
		t.Error("Expected default scopes")
	}
	
	// Check specific e-commerce scopes
	wms, exists := cfg.Scopes["wms"]
	if !exists {
		t.Error("Expected 'wms' scope to exist")
	}
	
	if wms.Type != "persistent" {
		t.Errorf("Expected wms scope type 'persistent', got %s", wms.Type)
	}
	
	if len(wms.Repos) != 5 {
		t.Errorf("Expected 5 repos in wms scope, got %d", len(wms.Repos))
	}
	
	// Check OMS scope
	oms, exists := cfg.Scopes["oms"]
	if !exists {
		t.Error("Expected 'oms' scope to exist")
	}
	
	if oms.Type != "persistent" {
		t.Errorf("Expected oms scope type 'persistent', got %s", oms.Type)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid config",
			config: &Config{
				Version: 2,
				Workspace: WorkspaceConfig{
					Name:     "test",
					BasePath: "workspaces",
				},
				Repositories: map[string]Repository{
					"repo1": {URL: "https://github.com/test/repo1.git", DefaultBranch: "main"},
				},
				Scopes: map[string]Scope{
					"test": {Type: "persistent", Repos: []string{"repo1"}},
				},
			},
			expectErr: false,
		},
		{
			name: "Missing workspace name",
			config: &Config{
				Version: 2,
				Workspace: WorkspaceConfig{
					BasePath: "workspaces",
				},
				Repositories: map[string]Repository{
					"repo1": {URL: "https://github.com/test/repo1.git"},
				},
				Scopes: map[string]Scope{
					"test": {Type: "persistent", Repos: []string{"repo1"}},
				},
			},
			expectErr: true,
			errMsg:    "workspace name is required",
		},
		{
			name: "No repositories",
			config: &Config{
				Version: 2,
				Workspace: WorkspaceConfig{
					Name:     "test",
					BasePath: "workspaces",
				},
				Repositories: map[string]Repository{},
				Scopes: map[string]Scope{
					"test": {Type: "persistent"},
				},
			},
			expectErr: true,
			errMsg:    "at least one repository must be defined",
		},
		{
			name: "No scopes",
			config: &Config{
				Version: 2,
				Workspace: WorkspaceConfig{
					Name:     "test",
					BasePath: "workspaces",
				},
				Repositories: map[string]Repository{
					"repo1": {URL: "https://github.com/test/repo1.git"},
				},
				Scopes: map[string]Scope{},
			},
			expectErr: true,
			errMsg:    "at least one scope must be defined",
		},
		{
			name: "Invalid scope type",
			config: &Config{
				Version: 2,
				Workspace: WorkspaceConfig{
					Name:     "test",
					BasePath: "workspaces",
				},
				Repositories: map[string]Repository{
					"repo1": {URL: "https://github.com/test/repo1.git"},
				},
				Scopes: map[string]Scope{
					"test": {Type: "invalid", Repos: []string{"repo1"}},
				},
			},
			expectErr: true,
			errMsg:    "invalid type",
		},
		{
			name: "Scope references undefined repo",
			config: &Config{
				Version: 2,
				Workspace: WorkspaceConfig{
					Name:     "test",
					BasePath: "workspaces",
				},
				Repositories: map[string]Repository{
					"repo1": {URL: "https://github.com/test/repo1.git"},
				},
				Scopes: map[string]Scope{
					"test": {Type: "persistent", Repos: []string{"repo2"}},
				},
			},
			expectErr: true,
			errMsg:    "references undefined repository",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			
			if tt.expectErr && err == nil {
				t.Error("Expected validation error but got none")
			}
			
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
			
			if tt.expectErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "repo-claude.yaml")
	
	// Create a config
	originalCfg := &Config{
		Version: 2,
		Workspace: WorkspaceConfig{
			Name:          "test-save",
			IsolationMode: true,
			BasePath:      "custom-workspaces",
		},
		Repositories: map[string]Repository{
			"repo1": {
				URL:           "https://github.com/test/repo1.git",
				DefaultBranch: "develop",
				Groups:        []string{"backend", "services"},
			},
			"repo2": {
				URL:           "https://github.com/test/repo2.git",
				DefaultBranch: "main",
				Groups:        []string{"frontend"},
			},
		},
		Scopes: map[string]Scope{
			"backend": {
				Type:        "persistent",
				Repos:       []string{"repo1"},
				Description: "Backend development",
				Model:       "claude-3-sonnet",
				AutoStart:   true,
			},
			"fullstack": {
				Type:        "ephemeral",
				Repos:       []string{"repo1", "repo2"},
				Description: "Full stack development",
				Model:       "claude-3-sonnet",
			},
		},
		Documentation: DocumentationConfig{
			Path:      "documentation",
			SyncToGit: true,
			RemoteURL: "git@github.com:test/docs.git",
		},
	}
	
	// Save the config
	err := originalCfg.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}
	
	// Load the config
	loadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Compare configs
	if loadedCfg.Version != originalCfg.Version {
		t.Errorf("Version mismatch: expected %d, got %d", originalCfg.Version, loadedCfg.Version)
	}
	
	if loadedCfg.Workspace.Name != originalCfg.Workspace.Name {
		t.Errorf("Workspace name mismatch: expected %s, got %s", 
			originalCfg.Workspace.Name, loadedCfg.Workspace.Name)
	}
	
	if loadedCfg.Workspace.BasePath != originalCfg.Workspace.BasePath {
		t.Errorf("Base path mismatch: expected %s, got %s",
			originalCfg.Workspace.BasePath, loadedCfg.Workspace.BasePath)
	}
	
	// Check repositories
	if len(loadedCfg.Repositories) != len(originalCfg.Repositories) {
		t.Errorf("Repository count mismatch: expected %d, got %d",
			len(originalCfg.Repositories), len(loadedCfg.Repositories))
	}
	
	repo1, exists := loadedCfg.Repositories["repo1"]
	if !exists {
		t.Error("repo1 not found in loaded config")
	} else {
		if repo1.URL != originalCfg.Repositories["repo1"].URL {
			t.Error("repo1 URL mismatch")
		}
		if repo1.DefaultBranch != originalCfg.Repositories["repo1"].DefaultBranch {
			t.Error("repo1 branch mismatch")
		}
	}
	
	// Check scopes
	if len(loadedCfg.Scopes) != len(originalCfg.Scopes) {
		t.Errorf("Scope count mismatch: expected %d, got %d",
			len(originalCfg.Scopes), len(loadedCfg.Scopes))
	}
	
	backend, exists := loadedCfg.Scopes["backend"]
	if !exists {
		t.Error("backend scope not found in loaded config")
	} else {
		if backend.Type != originalCfg.Scopes["backend"].Type {
			t.Error("backend scope type mismatch")
		}
		if backend.Description != originalCfg.Scopes["backend"].Description {
			t.Error("backend scope description mismatch")
		}
	}
	
	// Check documentation config
	if loadedCfg.Documentation.Path != originalCfg.Documentation.Path {
		t.Error("Documentation path mismatch")
	}
	if loadedCfg.Documentation.SyncToGit != originalCfg.Documentation.SyncToGit {
		t.Error("Documentation sync to git mismatch")
	}
}

func TestLoadWithDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "repo-claude.yaml")
	
	// Create minimal config without version and base path
	minimalCfg := map[string]interface{}{
		"workspace": map[string]interface{}{
			"name": "minimal",
		},
		"repositories": map[string]interface{}{
			"repo1": map[string]interface{}{
				"url":            "https://github.com/test/repo1.git",
				"default_branch": "main",
			},
		},
		"scopes": map[string]interface{}{
			"test": map[string]interface{}{
				"type":  "persistent",
				"repos": []string{"repo1"},
			},
		},
	}
	
	// Write as YAML
	data, err := yaml.Marshal(minimalCfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	
	// Load config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Check defaults were applied
	if cfg.Version != 2 {
		t.Errorf("Expected default version 2, got %d", cfg.Version)
	}
	
	if cfg.Workspace.BasePath != "workspaces" {
		t.Errorf("Expected default base path 'workspaces', got %s", cfg.Workspace.BasePath)
	}
	
	if !cfg.Workspace.IsolationMode {
		t.Error("Expected isolation mode to be true")
	}
	
	if cfg.Documentation.Path != "docs" {
		t.Errorf("Expected default docs path 'docs', got %s", cfg.Documentation.Path)
	}
}

func TestGetRepository(t *testing.T) {
	cfg := &Config{
		Repositories: map[string]Repository{
			"repo1": {URL: "https://github.com/test/repo1.git", DefaultBranch: "main"},
			"repo2": {URL: "https://github.com/test/repo2.git", DefaultBranch: "develop"},
		},
	}
	
	// Test existing repository
	repo, err := cfg.GetRepository("repo1")
	if err != nil {
		t.Errorf("Failed to get existing repository: %v", err)
	}
	if repo.URL != "https://github.com/test/repo1.git" {
		t.Error("Repository URL mismatch")
	}
	
	// Test non-existent repository
	_, err = cfg.GetRepository("repo3")
	if err == nil {
		t.Error("Expected error for non-existent repository")
	}
}

func TestGetScope(t *testing.T) {
	cfg := &Config{
		Scopes: map[string]Scope{
			"backend": {Type: "persistent", Description: "Backend scope"},
			"frontend": {Type: "ephemeral", Description: "Frontend scope"},
		},
	}
	
	// Test existing scope
	scope, err := cfg.GetScope("backend")
	if err != nil {
		t.Errorf("Failed to get existing scope: %v", err)
	}
	if scope.Type != "persistent" {
		t.Error("Scope type mismatch")
	}
	
	// Test non-existent scope
	_, err = cfg.GetScope("mobile")
	if err == nil {
		t.Error("Expected error for non-existent scope")
	}
}

func TestGetScopesForRepo(t *testing.T) {
	cfg := &Config{
		Scopes: map[string]Scope{
			"backend": {Repos: []string{"auth", "api", "db"}},
			"frontend": {Repos: []string{"web", "mobile"}},
			"fullstack": {Repos: []string{"api", "web"}},
		},
	}
	
	// Test repo in multiple scopes
	scopes := cfg.GetScopesForRepo("api")
	if len(scopes) != 2 {
		t.Errorf("Expected 2 scopes for 'api', got %d", len(scopes))
	}
	
	// Verify correct scopes
	scopeMap := make(map[string]bool)
	for _, s := range scopes {
		scopeMap[s] = true
	}
	if !scopeMap["backend"] || !scopeMap["fullstack"] {
		t.Error("Incorrect scopes returned for 'api'")
	}
	
	// Test repo in single scope
	scopes = cfg.GetScopesForRepo("mobile")
	if len(scopes) != 1 {
		t.Errorf("Expected 1 scope for 'mobile', got %d", len(scopes))
	}
	if scopes[0] != "frontend" {
		t.Errorf("Expected 'frontend' scope for 'mobile', got %s", scopes[0])
	}
	
	// Test repo not in any scope
	scopes = cfg.GetScopesForRepo("nonexistent")
	if len(scopes) != 0 {
		t.Errorf("Expected 0 scopes for 'nonexistent', got %d", len(scopes))
	}
}