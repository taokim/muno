package e2e

import (
	// "fmt"
	// "os"
	// "path/filepath"
	"testing"

	// "github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
	// "github.com/taokim/muno/internal/config"
	// "github.com/taokim/muno/internal/manager"
)

// TestCompleteWorkflow is commented out as it tests the old v2 config structure
// which is no longer used in the tree-based architecture
// TODO: Rewrite this test for the tree-based architecture
func TestCompleteWorkflow(t *testing.T) {
	t.Skip("This test needs to be rewritten for the tree-based architecture without v2 config structure")
}

/* Original test commented out - needs rewrite for tree-based architecture
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
		assert.FileExists(t, filepath.Join(projectPath, "muno.yaml"))
		assert.DirExists(t, filepath.Join(projectPath, "workspaces"))
		assert.DirExists(t, filepath.Join(projectPath, "docs"))
		assert.FileExists(t, filepath.Join(projectPath, "CLAUDE.md"))
	})
	
	// Step 2: Load and modify the config
	t.Run("ConfigureWorkspace", func(t *testing.T) {
		configPath := filepath.Join(projectPath, "muno.yaml")
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
			"auth-service": {
				URL:    "https://github.com/test/auth-service.git",
				Branch: "main",
				Groups: []string{"backend", "services"},
			},
			"web-app": {
				URL:    "https://github.com/test/web-app.git",
				Branch: "main",
				Groups: []string{"frontend"},
			},
		}
		
		// Add scopes  
		cfg.Scopes = map[string]config.Scope{
			"backend": {
				Type:        "persistent",
				Repos:       []string{"backend-meta", "payment-service", "auth-service"},
				Description: "Backend development",
			},
			"frontend": {
				Type:        "persistent",
				Repos:       []string{"frontend-repo", "web-app"},
				Description: "Frontend development",
			},
			"hotfix": {
				Type:        "ephemeral",
				Repos:       []string{"payment-service"},
				Description: "Emergency hotfix",
			},
		}
		
		err = cfg.Save(configPath)
		require.NoError(t, err)
	})
	
	// Step 3: Load workspace and verify configuration
	t.Run("LoadWorkspace", func(t *testing.T) {
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
		
		// Test tree structure
		err = mgr.ShowTree("/", 0)
		assert.NoError(t, err)
		
		// Test navigation
		err = mgr.UseNode("/", false)
		assert.NoError(t, err)
		
		// Test listing
		err = mgr.ListNodes(false)
		assert.NoError(t, err)
	})
}

func TestMetaRepoHeuristics(t *testing.T) {
	// Test that the system correctly identifies meta-repos
	testCases := []struct {
		name     string
		repoName string
		url      string
		expected bool
	}{
		{
			name:     "backend-repo suffix",
			repoName: "backend-repo",
			url:      "https://github.com/org/backend-repo.git",
			expected: true,
		},
		{
			name:     "frontend-monorepo suffix",
			repoName: "frontend-monorepo",
			url:      "https://github.com/org/frontend-monorepo.git",
			expected: true,
		},
		{
			name:     "payments-muno suffix",
			repoName: "payments-rc",
			url:      "https://github.com/org/payments-rc.git",
			expected: true,
		},
		{
			name:     "regular service",
			repoName: "payment-service",
			url:      "https://github.com/org/payment-service.git",
			expected: false,
		},
		{
			name:     "regular api",
			repoName: "auth-api",
			url:      "https://github.com/org/auth-api.git",
			expected: false,
		},
	}
	
	defaults := config.DefaultDefaults()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := config.IsMetaRepo(tc.repoName, tc.url, defaults)
			assert.Equal(t, tc.expected, result, 
				fmt.Sprintf("Expected %s to be meta-repo: %v", tc.repoName, tc.expected))
		})
	}
}
*/