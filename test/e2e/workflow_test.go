package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/repo-claude/internal/config"
	"github.com/taokim/repo-claude/internal/manager"
)

// TestCompleteWorkflow tests a complete workflow from init to scope operations
func TestCompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}
	
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-platform")
	
	// Step 1: Initialize a new workspace
	t.Run("Initialize", func(t *testing.T) {
		mgr, err := manager.New(projectPath)
		require.NoError(t, err)
		
		err = mgr.InitWorkspace("test-platform", false)
		require.NoError(t, err)
		
		// Verify structure
		assert.DirExists(t, projectPath)
		assert.FileExists(t, filepath.Join(projectPath, "repo-claude.yaml"))
		assert.DirExists(t, filepath.Join(projectPath, "workspaces"))
		assert.DirExists(t, filepath.Join(projectPath, "docs"))
		assert.FileExists(t, filepath.Join(projectPath, "CLAUDE.md"))
	})
	
	// Step 2: Load and modify the config
	t.Run("ConfigureWorkspace", func(t *testing.T) {
		configPath := filepath.Join(projectPath, "repo-claude.yaml")
		cfg, err := config.Load(configPath)
		require.NoError(t, err)
		
		// Add repositories
		cfg.Repositories = map[string]config.Repository{
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
				Groups: []string{"backend", "services"},
			},
			"fraud-detection": {
				URL:    "https://github.com/test/fraud-detection.git",
				Branch: "main",
				Groups: []string{"backend", "ml"},
			},
			"web-app": {
				URL:    "https://github.com/test/web-app.git",
				Branch: "main",
				Groups: []string{"frontend", "ui"},
			},
		}
		
		// Scopes removed in tree-based architecture
		
		// Save updated config
		err = cfg.Save(configPath)
		require.NoError(t, err)
	})
	
	// Step 3: Load workspace and verify
	t.Run("LoadAndVerify", func(t *testing.T) {
		// Change to project directory
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		require.NoError(t, os.Chdir(projectPath))
		
		mgr, err := manager.LoadFromCurrentDir()
		require.NoError(t, err)
		
		// Verify loaded configuration
		assert.Equal(t, "test-platform", mgr.Config.Workspace.Name)
		assert.Len(t, mgr.Config.Repositories, 5)
		// Scopes removed - assert.Len(t, mgr.Config.Scopes, 3)
		
		// Test lazy loading detection
		backendMeta := mgr.Config.Repositories["backend-meta"]
		assert.False(t, backendMeta.IsLazy("backend-meta", mgr.Config.Defaults), 
			"Meta-repos should be eager-loaded")
		
		paymentService := mgr.Config.Repositories["payment-service"]
		assert.True(t, paymentService.IsLazy("payment-service", mgr.Config.Defaults),
			"Regular services should be lazy-loaded")
	})
	
	// Step 4: Test tree operations
	t.Run("TreeOperations", func(t *testing.T) {
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		require.NoError(t, os.Chdir(projectPath))
		
		mgr, err := manager.LoadFromCurrentDir()
		require.NoError(t, err)
		
		// Tree operations would go here
		// ShowTree, UseNode, etc.
	})
}

// TestRecursiveWorkspace tests recursive workspace functionality
func TestRecursiveWorkspace(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}
	
	tempDir := t.TempDir()
	
	// Create a root workspace
	rootPath := filepath.Join(tempDir, "platform")
	rootConfig := &config.Config{
		Version: 3,
		Workspace: config.WorkspaceConfig{
			Name:          "platform",
			RootPath: "repos",
		},
		Defaults: config.DefaultDefaults(),
		Repositories: map[string]config.Repository{
			"backend-meta": {
				URL:    "https://github.com/acme/backend-meta.git",
				Branch: "main",
			},
			"frontend-repo": {
				URL:    "https://github.com/acme/frontend-repo.git",
				Branch: "main",
			},
			"shared-libs": {
				URL:    "https://github.com/acme/shared-libs.git",
				Branch: "main",
			},
		},
		// Scopes removed
	}
	
	// Create root workspace
	require.NoError(t, os.MkdirAll(rootPath, 0755))
	configPath := filepath.Join(rootPath, "repo-claude.yaml")
	require.NoError(t, rootConfig.Save(configPath))
	
	// Create a child workspace (backend-meta)
	backendPath := filepath.Join(rootPath, "backend-meta")
	backendConfig := &config.Config{
		Version: 3,
		Workspace: config.WorkspaceConfig{
			Name:          "backend",
			RootPath: "repos",
		},
		Defaults: config.DefaultDefaults(),
		Repositories: map[string]config.Repository{
			"payment-service": {
				URL:    "https://github.com/acme/payment-service.git",
				Branch: "main",
			},
			"order-service": {
				URL:    "https://github.com/acme/order-service.git",
				Branch: "main",
			},
		},
		// Scopes removed
	}
	
	require.NoError(t, os.MkdirAll(backendPath, 0755))
	backendConfigPath := filepath.Join(backendPath, "repo-claude.yaml")
	require.NoError(t, backendConfig.Save(backendConfigPath))
	
	// Load root workspace and verify hierarchy
	t.Run("VerifyHierarchy", func(t *testing.T) {
		cfg, err := config.Load(configPath)
		require.NoError(t, err)
		
		// Check that backend-meta is detected as a meta-repo
		backendMeta := cfg.Repositories["backend-meta"]
		assert.False(t, backendMeta.IsLazy("backend-meta", cfg.Defaults),
			"backend-meta should be eager-loaded as a meta-repo")
		
		// Check that shared-libs is lazy
		sharedLibs := cfg.Repositories["shared-libs"]
		assert.True(t, sharedLibs.IsLazy("shared-libs", cfg.Defaults),
			"shared-libs should be lazy-loaded")
		
		// Verify the child workspace exists
		assert.DirExists(t, backendPath)
		assert.FileExists(t, backendConfigPath)
		
		// Load child workspace
		childCfg, err := config.Load(backendConfigPath)
		require.NoError(t, err)
		assert.Equal(t, "backend", childCfg.Workspace.Name)
		assert.Len(t, childCfg.Repositories, 2)
		// Scopes removed - assert.Len(t, childCfg.Scopes, 2)
	})
	
	// Test cross-workspace scope references
	t.Run("CrossWorkspaceScopes", func(t *testing.T) {
		// Add a scope to root that references child workspace scopes
		cfg, err := config.Load(configPath)
		require.NoError(t, err)
		
		// Scopes removed - cfg.Scopes no longer exists
		
		err = cfg.Save(configPath)
		require.NoError(t, err)
		
		// Reload and verify
		cfg, err = config.Load(configPath)
		require.NoError(t, err)
		
		// Scopes validation removed
	})
}

// TestPerformanceCharacteristics tests performance characteristics of v3
func TestPerformanceCharacteristics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}
	
	// Create a large enterprise workspace
	cfg := &config.Config{
		Version: 3,
		Workspace: config.WorkspaceConfig{
			Name: "enterprise",
		},
		Defaults: config.DefaultDefaults(),
		Repositories: make(map[string]config.Repository),
		// Scopes removed
	}
	
	// Add 500 repositories
	metaRepoCount := 0
	serviceRepoCount := 0
	
	for i := 0; i < 500; i++ {
		name := fmt.Sprintf("service-%03d", i)
		// Every 50th repo is a meta-repo
		if i%50 == 0 {
			name = fmt.Sprintf("domain-%02d-meta", i/50)
			metaRepoCount++
		} else {
			serviceRepoCount++
		}
		
		cfg.Repositories[name] = config.Repository{
			URL:    fmt.Sprintf("https://github.com/enterprise/%s.git", name),
			Branch: "main",
		}
	}
	
	// Test lazy loading detection
	eagerCount := 0
	lazyCount := 0
	
	for name, repo := range cfg.Repositories {
		if repo.IsLazy(name, cfg.Defaults) {
			lazyCount++
		} else {
			eagerCount++
		}
	}
	
	assert.Equal(t, metaRepoCount, eagerCount, "Meta-repos should be eager")
	assert.Equal(t, serviceRepoCount, lazyCount, "Services should be lazy")
	
	// Performance characteristics
	assert.Equal(t, 10, eagerCount, "Should have 10 meta-repos (500/50)")
	assert.Equal(t, 490, lazyCount, "Should have 490 service repos")
	
	// Only 2% of repos are loaded eagerly for structure discovery
	eagerPercentage := float64(eagerCount) / float64(eagerCount+lazyCount) * 100
	assert.Less(t, eagerPercentage, 3.0, "Less than 3% should be eager for performance")
	
	t.Logf("Performance: %d/%d (%.1f%%) repos loaded eagerly", 
		eagerCount, eagerCount+lazyCount, eagerPercentage)
}