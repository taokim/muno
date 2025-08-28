package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestV3WorkspaceHierarchy tests creating and navigating a workspace hierarchy
func TestV3WorkspaceHierarchy(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create root workspace config
	rootConfig := &ConfigV3{
		Version: 3,
		Workspace: WorkspaceConfig{
			Name:          "platform",
			IsolationMode: true,
			BasePath:      "workspaces",
		},
		Defaults: DefaultDefaults(),
		Repositories: map[string]RepositoryV3{
			// Meta-repos (will be detected as workspaces)
			"backend-meta": {
				URL:    "https://github.com/acme/backend-meta.git",
				Branch: "main",
			},
			"frontend-repo": {
				URL:    "https://github.com/acme/frontend-repo.git",
				Branch: "main",
			},
			// Regular repos
			"shared-libs": {
				URL:    "https://github.com/acme/shared-libs.git",
				Branch: "main",
				Groups: []string{"common"},
			},
		},
		Scopes: map[string]ScopeV3{
			"platform": {
				Type:        "persistent",
				Repos:       []string{"shared-libs"},
				Description: "Platform-wide development",
			},
			"backend": {
				Type: "persistent",
				WorkspaceScopes: map[string][]string{
					"backend-meta": {"api", "services"},
				},
				Description: "Backend development with sub-scopes",
			},
		},
	}
	
	// Save root config
	rootPath := filepath.Join(tempDir, "platform")
	require.NoError(t, os.MkdirAll(rootPath, 0755))
	
	configPath := filepath.Join(rootPath, "repo-claude.yaml")
	require.NoError(t, rootConfig.SaveV3(configPath))
	
	// Load and verify
	loaded, err := LoadV3(configPath)
	require.NoError(t, err)
	
	assert.Equal(t, "platform", loaded.Workspace.Name)
	assert.Len(t, loaded.Repositories, 3)
	assert.Len(t, loaded.Scopes, 2)
	
	// Check lazy loading detection
	backendMeta := loaded.Repositories["backend-meta"]
	assert.False(t, backendMeta.IsLazy("backend-meta", loaded.Defaults), "Meta-repos should be eager")
	
	sharedLibs := loaded.Repositories["shared-libs"]
	assert.True(t, sharedLibs.IsLazy("shared-libs", loaded.Defaults), "Regular repos should be lazy")
}

// TestV3RecursiveScopes tests scopes that reference sub-workspace scopes
func TestV3RecursiveScopes(t *testing.T) {
	cfg := &ConfigV3{
		Version: 3,
		Workspace: WorkspaceConfig{
			Name: "root",
		},
		Defaults: DefaultDefaults(),
		Repositories: map[string]RepositoryV3{
			"backend-repo": {
				URL: "https://github.com/test/backend-repo.git",
			},
			"frontend-repo": {
				URL: "https://github.com/test/frontend-repo.git",
			},
		},
		Scopes: map[string]ScopeV3{
			"full-stack": {
				Type: "persistent",
				WorkspaceScopes: map[string][]string{
					"backend-repo":  {"api", "services"},
					"frontend-repo": {"web", "mobile"},
				},
				Description: "Full-stack development across workspaces",
			},
			"backend-only": {
				Type: "persistent",
				WorkspaceScopes: map[string][]string{
					"backend-repo": {"api"},
				},
				Repos: []string{}, // Can be empty when using workspace scopes
			},
		},
	}
	
	// Validate config
	err := cfg.ValidateV3()
	require.NoError(t, err)
	
	// Check scope references
	fullStack := cfg.Scopes["full-stack"]
	assert.Len(t, fullStack.WorkspaceScopes, 2)
	assert.Contains(t, fullStack.WorkspaceScopes["backend-repo"], "api")
	assert.Contains(t, fullStack.WorkspaceScopes["frontend-repo"], "web")
}

// TestV3SmartLoading tests the smart loading strategy
func TestV3SmartLoading(t *testing.T) {
	tests := []struct {
		name         string
		repoName     string
		repoConfig   RepositoryV3
		defaults     Defaults
		expectedLazy bool
		description  string
	}{
		{
			name:     "meta-repo eager by default",
			repoName: "payments-meta",
			repoConfig: RepositoryV3{
				URL: "https://github.com/acme/payments-meta.git",
			},
			defaults:     DefaultDefaults(),
			expectedLazy: false,
			description:  "Meta-repos should be eager-loaded for structure discovery",
		},
		{
			name:     "monorepo eager by pattern",
			repoName: "backend-monorepo",
			repoConfig: RepositoryV3{
				URL: "https://github.com/acme/backend-monorepo.git",
			},
			defaults:     DefaultDefaults(),
			expectedLazy: false,
			description:  "Monorepos should be eager-loaded",
		},
		{
			name:     "service lazy by default",
			repoName: "payment-service",
			repoConfig: RepositoryV3{
				URL: "https://github.com/acme/payment-service.git",
			},
			defaults:     DefaultDefaults(),
			expectedLazy: true,
			description:  "Regular services should be lazy-loaded",
		},
		{
			name:     "explicit lazy override",
			repoName: "critical-meta",
			repoConfig: RepositoryV3{
				URL:  "https://github.com/acme/critical-meta.git",
				Lazy: boolPtr(true), // Explicit override
			},
			defaults:     DefaultDefaults(),
			expectedLazy: true,
			description:  "Explicit lazy flag should override pattern detection",
		},
		{
			name:     "explicit eager override",
			repoName: "important-service",
			repoConfig: RepositoryV3{
				URL:  "https://github.com/acme/important-service.git",
				Lazy: boolPtr(false), // Explicit override
			},
			defaults:     DefaultDefaults(),
			expectedLazy: false,
			description:  "Explicit eager flag should override default lazy loading",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isLazy := tt.repoConfig.IsLazy(tt.repoName, tt.defaults)
			assert.Equal(t, tt.expectedLazy, isLazy, tt.description)
		})
	}
}

// TestV3PerformanceOptimization tests that v3 config optimizes for performance
func TestV3PerformanceOptimization(t *testing.T) {
	// Create a large workspace config
	cfg := &ConfigV3{
		Version: 3,
		Workspace: WorkspaceConfig{
			Name: "enterprise",
		},
		Defaults: DefaultDefaults(),
		Repositories: make(map[string]RepositoryV3),
		Scopes:       make(map[string]ScopeV3),
	}
	
	// Add many repositories
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("service-%d", i)
		cfg.Repositories[name] = RepositoryV3{
			URL:    fmt.Sprintf("https://github.com/acme/%s.git", name),
			Branch: "main",
		}
	}
	
	// Add meta-repos
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("domain-%d-meta", i)
		cfg.Repositories[name] = RepositoryV3{
			URL:    fmt.Sprintf("https://github.com/acme/%s.git", name),
			Branch: "main",
		}
	}
	
	// Count eager vs lazy repos
	eagerCount := 0
	lazyCount := 0
	
	for name, repo := range cfg.Repositories {
		if repo.IsLazy(name, cfg.Defaults) {
			lazyCount++
		} else {
			eagerCount++
		}
	}
	
	// Meta-repos (10) should be eager, services (100) should be lazy
	assert.Equal(t, 10, eagerCount, "Only meta-repos should be eager")
	assert.Equal(t, 100, lazyCount, "Regular services should be lazy")
	
	// This ensures initial load is fast (only 10 repos) instead of all 110
	assert.Less(t, float64(eagerCount)/float64(eagerCount+lazyCount), 0.15, 
		"Less than 15% of repos should be eager-loaded for performance")
}