package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorkspaceHierarchy tests creating and navigating a workspace hierarchy
func TestWorkspaceHierarchy(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create root workspace config
	rootConfig := &Config{
		Version: 3,
		Workspace: WorkspaceConfig{
			Name:     "platform",
			RootPath: "repos",
		},
		Defaults: DefaultDefaults(),
		Repositories: map[string]Repository{
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
		// Scopes removed
	}
	
	// Save root config
	rootPath := filepath.Join(tempDir, "platform")
	require.NoError(t, os.MkdirAll(rootPath, 0755))
	
	configPath := filepath.Join(rootPath, "repo-claude.yaml")
	require.NoError(t, rootConfig.Save(configPath))
	
	// Load and verify
	loaded, err := Load(configPath)
	require.NoError(t, err)
	
	assert.Equal(t, "platform", loaded.Workspace.Name)
	assert.Len(t, loaded.Repositories, 3)
	// Scopes removed in tree-based architecture
	// assert.Len(t, loaded.Scopes, 2)
	
	// Check lazy loading detection
	backendMeta := loaded.Repositories["backend-meta"]
	assert.False(t, backendMeta.IsLazy("backend-meta", loaded.Defaults), "Meta-repos should be eager")
	
	sharedLibs := loaded.Repositories["shared-libs"]
	assert.True(t, sharedLibs.IsLazy("shared-libs", loaded.Defaults), "Regular repos should be lazy")
}

// TestRecursiveScopes tests scopes that reference sub-workspace scopes
func TestRecursiveScopes(t *testing.T) {
	cfg := &Config{
		Version: 3,
		Workspace: WorkspaceConfig{
			Name: "root",
		},
		Defaults: DefaultDefaults(),
		Repositories: map[string]Repository{
			"backend-repo": {
				URL: "https://github.com/test/backend-repo.git",
			},
			"frontend-repo": {
				URL: "https://github.com/test/frontend-repo.git",
			},
		},
		// Scopes removed
	}
	
	// Validate config
	err := cfg.Validate()
	require.NoError(t, err)
	
	// Scopes removed in tree-based architecture
	// Check scope references
	// fullStack := cfg.Scopes["full-stack"]
	// assert.Len(t, fullStack.WorkspaceScopes, 2)
	// assert.Contains(t, fullStack.WorkspaceScopes["backend-repo"], "api")
	// assert.Contains(t, fullStack.WorkspaceScopes["frontend-repo"], "web")
}

// TestSmartLoading tests the smart loading strategy
func TestSmartLoading(t *testing.T) {
	tests := []struct {
		name         string
		repoName     string
		repoConfig   Repository
		defaults     Defaults
		expectedLazy bool
		description  string
	}{
		{
			name:     "meta-repo eager by default",
			repoName: "payments-meta",
			repoConfig: Repository{
				URL: "https://github.com/acme/payments-meta.git",
			},
			defaults:     DefaultDefaults(),
			expectedLazy: false,
			description:  "Meta-repos should be eager-loaded for structure discovery",
		},
		{
			name:     "monorepo eager by pattern",
			repoName: "backend-monorepo",
			repoConfig: Repository{
				URL: "https://github.com/acme/backend-monorepo.git",
			},
			defaults:     DefaultDefaults(),
			expectedLazy: false,
			description:  "Monorepos should be eager-loaded",
		},
		{
			name:     "service lazy by default",
			repoName: "payment-service",
			repoConfig: Repository{
				URL: "https://github.com/acme/payment-service.git",
			},
			defaults:     DefaultDefaults(),
			expectedLazy: true,
			description:  "Regular services should be lazy-loaded",
		},
		{
			name:     "explicit lazy override",
			repoName: "critical-meta",
			repoConfig: Repository{
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
			repoConfig: Repository{
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

// TestPerformanceOptimization tests that config optimizes for performance
func TestPerformanceOptimization(t *testing.T) {
	// Create a large workspace config
	cfg := &Config{
		Version: 3,
		Workspace: WorkspaceConfig{
			Name: "enterprise",
		},
		Defaults: DefaultDefaults(),
		Repositories: make(map[string]Repository),
		// Scopes removed
	}
	
	// Add many repositories
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("service-%d", i)
		cfg.Repositories[name] = Repository{
			URL:    fmt.Sprintf("https://github.com/acme/%s.git", name),
			Branch: "main",
		}
	}
	
	// Add meta-repos
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("domain-%d-meta", i)
		cfg.Repositories[name] = Repository{
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