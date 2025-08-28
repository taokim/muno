package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaRepoDetection(t *testing.T) {
	defaults := DefaultDefaults()
	
	tests := []struct {
		name     string
		repoName string
		url      string
		isMeta   bool
	}{
		// Meta-repos (should be eager)
		{"backend-repo suffix", "backend-repo", "https://github.com/acme/backend-repo.git", true},
		{"frontend-monorepo suffix", "frontend-monorepo", "https://github.com/acme/frontend-monorepo.git", true},
		{"payments-rc suffix", "payments-rc", "https://github.com/acme/payments-rc.git", true},
		{"platform-meta suffix", "platform-meta", "https://github.com/acme/platform-meta.git", true},
		
		// Regular repos (should be lazy)
		{"payment-service", "payment-service", "https://github.com/acme/payment-service.git", false},
		{"fraud-detection-api", "fraud-detection-api", "https://github.com/acme/fraud-detection-api.git", false},
		{"web-app", "web-app", "https://github.com/acme/web-app.git", false},
		{"utils", "utils", "https://github.com/acme/utils.git", false},
		
		// Edge cases
		{"repo-in-middle", "some-repo-service", "https://github.com/acme/some-repo-service.git", false},
		{"meta-in-middle", "some-meta-service", "https://github.com/acme/some-meta-service.git", false},
		{"UPPERCASE-REPO", "BACKEND-REPO", "https://github.com/acme/BACKEND-REPO.git", true},
		{"MixedCase-Meta", "Platform-Meta", "https://github.com/acme/Platform-Meta.git", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMetaRepo(tt.repoName, tt.url, defaults)
			assert.Equal(t, tt.isMeta, result)
		})
	}
}

func TestRepositoryIsLazy(t *testing.T) {
	defaults := DefaultDefaults()
	
	tests := []struct {
		name         string
		repo         RepositoryV3
		repoName     string
		expectedLazy bool
	}{
		{
			name:         "meta-repo is eager",
			repo:         RepositoryV3{URL: "https://github.com/acme/backend-repo.git"},
			repoName:     "backend-repo",
			expectedLazy: false,
		},
		{
			name:         "regular repo is lazy",
			repo:         RepositoryV3{URL: "https://github.com/acme/payment-service.git"},
			repoName:     "payment-service",
			expectedLazy: true,
		},
		{
			name: "explicit lazy true",
			repo: RepositoryV3{
				URL:  "https://github.com/acme/backend-repo.git",
				Lazy: boolPtr(true),
			},
			repoName:     "backend-repo",
			expectedLazy: true, // Explicit overrides pattern
		},
		{
			name: "explicit lazy false",
			repo: RepositoryV3{
				URL:  "https://github.com/acme/payment-service.git",
				Lazy: boolPtr(false),
			},
			repoName:     "payment-service",
			expectedLazy: false, // Explicit overrides default
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.repo.IsLazy(tt.repoName, defaults)
			assert.Equal(t, tt.expectedLazy, result)
		})
	}
}

func TestCustomPatterns(t *testing.T) {
	customDefaults := Defaults{
		Lazy:         true,
		EagerPattern: `(?i)(-(repo|monorepo|rc|meta|workspace|platform)$)`,
		LazyPattern:  `(?i)(-(service|api|lib)$)`,
	}
	
	tests := []struct {
		name         string
		repoName     string
		expectedLazy bool
	}{
		// Custom eager patterns
		{"backend-workspace", "backend-workspace", false},
		{"payment-platform", "payment-platform", false},
		
		// Custom lazy patterns
		{"payment-service", "payment-service", true},
		{"fraud-api", "fraud-api", true},
		{"common-lib", "common-lib", true},
		
		// Default behavior for unmatched
		{"random-name", "random-name", true}, // Falls back to default (lazy=true)
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := RepositoryV3{URL: "https://github.com/acme/" + tt.repoName + ".git"}
			result := repo.IsLazy(tt.repoName, customDefaults)
			assert.Equal(t, tt.expectedLazy, result)
		})
	}
}

func TestConfigV3Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *ConfigV3
		wantErr string
	}{
		{
			name: "valid config",
			config: &ConfigV3{
				Version: 3,
				Workspace: WorkspaceConfig{
					Name: "test",
				},
				Defaults: DefaultDefaults(),
				Repositories: map[string]RepositoryV3{
					"repo1": {URL: "https://github.com/test/repo1.git"},
				},
				Scopes: map[string]ScopeV3{
					"default": {Type: "persistent"},
				},
			},
			wantErr: "",
		},
		{
			name: "missing workspace name",
			config: &ConfigV3{
				Version:   3,
				Workspace: WorkspaceConfig{},
				Repositories: map[string]RepositoryV3{
					"repo1": {URL: "https://github.com/test/repo1.git"},
				},
				Scopes: map[string]ScopeV3{
					"default": {Type: "persistent"},
				},
			},
			wantErr: "workspace name is required",
		},
		{
			name: "no repositories",
			config: &ConfigV3{
				Version: 3,
				Workspace: WorkspaceConfig{
					Name: "test",
				},
				Repositories: map[string]RepositoryV3{},
				Scopes: map[string]ScopeV3{
					"default": {Type: "persistent"},
				},
			},
			wantErr: "at least one repository must be defined",
		},
		{
			name: "invalid regex pattern",
			config: &ConfigV3{
				Version: 3,
				Workspace: WorkspaceConfig{
					Name: "test",
				},
				Defaults: Defaults{
					Lazy:         true,
					EagerPattern: "[(invalid regex",
				},
				Repositories: map[string]RepositoryV3{
					"repo1": {URL: "https://github.com/test/repo1.git"},
				},
				Scopes: map[string]ScopeV3{
					"default": {Type: "persistent"},
				},
			},
			wantErr: "invalid eager_pattern regex",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateV3()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}


func TestConfigV3YAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "repo-claude.yaml")
	
	yamlContent := `
version: 3
workspace:
  name: test-platform
  isolation_mode: true
  base_path: workspaces

defaults:
  lazy: true
  eager_pattern: '(?i)(-(repo|monorepo|rc|meta)$)'

repositories:
  # Meta-repo (eager)
  backend-repo:
    url: https://github.com/acme/backend-repo.git
    branch: main
    
  # Regular repo (lazy)
  payment-service:
    url: https://github.com/acme/payment-service.git
    branch: main
    groups: ["backend", "services"]
    
  # Explicit lazy override
  special-config:
    url: https://github.com/acme/special-config.git
    lazy: false

scopes:
  backend:
    type: persistent
    repos: ["payment-service"]
    description: Backend development
    model: claude-3-5-sonnet-20241022
    
documentation:
  path: docs
  sync_to_git: true
`
	
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)
	
	// Load and validate
	cfg, err := LoadV3(configPath)
	require.NoError(t, err)
	
	assert.Equal(t, 3, cfg.Version)
	assert.Equal(t, "test-platform", cfg.Workspace.Name)
	assert.Len(t, cfg.Repositories, 3)
	
	// Check lazy detection
	backendRepo := cfg.Repositories["backend-repo"]
	assert.False(t, backendRepo.IsLazy("backend-repo", cfg.Defaults))
	
	paymentService := cfg.Repositories["payment-service"]
	assert.True(t, paymentService.IsLazy("payment-service", cfg.Defaults))
	
	specialConfig := cfg.Repositories["special-config"]
	assert.False(t, specialConfig.IsLazy("special-config", cfg.Defaults)) // Explicit false override
}

// TestConfigV3SaveLoad tests saving and loading v3 config
func TestConfigV3SaveLoad(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "repo-claude.yaml")
	
	// Create a config
	cfg := &ConfigV3{
		Version: 3,
		Workspace: WorkspaceConfig{
			Name:          "test-project",
			IsolationMode: true,
			BasePath:      "workspaces",
		},
		Defaults: DefaultDefaults(),
		Repositories: map[string]RepositoryV3{
			"backend-repo": {
				URL:    "https://github.com/test/backend-repo.git",
				Branch: "main",
			},
			"payment-service": {
				URL:    "https://github.com/test/payment-service.git",
				Branch: "develop",
				Groups: []string{"backend", "core"},
			},
		},
		Scopes: map[string]ScopeV3{
			"backend": {
				Type:        "persistent",
				Repos:       []string{"payment-service"},
				Description: "Backend services",
				Model:       "claude-3-5-sonnet",
			},
		},
		Documentation: DocumentationConfig{
			Path:       "docs",
			SyncToGit:  true,
		},
	}
	
	// Save config
	err := cfg.SaveV3(configPath)
	require.NoError(t, err)
	
	// Load config back
	loaded, err := LoadV3(configPath)
	require.NoError(t, err)
	
	// Verify loaded config matches original
	assert.Equal(t, cfg.Version, loaded.Version)
	assert.Equal(t, cfg.Workspace.Name, loaded.Workspace.Name)
	assert.Equal(t, cfg.Workspace.IsolationMode, loaded.Workspace.IsolationMode)
	assert.Equal(t, cfg.Workspace.BasePath, loaded.Workspace.BasePath)
	
	assert.Len(t, loaded.Repositories, 2)
	assert.Contains(t, loaded.Repositories, "backend-repo")
	assert.Contains(t, loaded.Repositories, "payment-service")
	
	assert.Equal(t, "develop", loaded.Repositories["payment-service"].Branch)
	assert.Equal(t, []string{"backend", "core"}, loaded.Repositories["payment-service"].Groups)
	
	assert.Len(t, loaded.Scopes, 1)
	assert.Equal(t, "Backend services", loaded.Scopes["backend"].Description)
}

// TestRecursiveWorkspaceDetection tests detecting nested workspaces
func TestRecursiveWorkspaceDetection(t *testing.T) {
	cfg := &ConfigV3{
		Version: 3,
		Workspace: WorkspaceConfig{
			Name: "root",
		},
		Defaults: DefaultDefaults(),
		Repositories: map[string]RepositoryV3{
			"backend-meta": {
				URL:    "https://github.com/test/backend-meta.git",
				Branch: "main",
			},
			"frontend-repo": {
				URL:    "https://github.com/test/frontend-repo.git",
				Branch: "main",
			},
			"payment-service": {
				URL:    "https://github.com/test/payment-service.git",
				Branch: "main",
			},
		},
	}
	
	// Test meta-repo detection (should be eager-loaded)
	backendMeta := cfg.Repositories["backend-meta"]
	assert.False(t, backendMeta.IsLazy("backend-meta", cfg.Defaults))
	
	frontendRepo := cfg.Repositories["frontend-repo"]
	assert.False(t, frontendRepo.IsLazy("frontend-repo", cfg.Defaults))
	
	// Test regular service detection (should be lazy-loaded)
	paymentService := cfg.Repositories["payment-service"]
	assert.True(t, paymentService.IsLazy("payment-service", cfg.Defaults))
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}