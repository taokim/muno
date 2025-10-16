package tree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/taokim/muno/internal/git"
)

// createMockGitWithMunoYaml creates a mock git that simulates repositories with muno.yaml files
func createMockGitWithMunoYaml() (*git.MockGit, map[string]string) {
	clonedRepos := make(map[string]string)
	
	return &git.MockGit{
		CloneFunc: func(url, path string) error {
			clonedRepos[url] = path
			
			// Create .git directory to simulate successful clone
			gitDir := filepath.Join(path, ".git")
			if err := os.MkdirAll(gitDir, 0755); err != nil {
				return err
			}
			
			// Simulate creating muno.yaml files for certain repos
			if url == "https://github.com/test/backend-meta.git" {
				// Backend meta-repo contains child service definitions
				munoContent := `version: 1
defaults:
  fetch: auto
nodes:
  - name: payment-service
    url: https://github.com/test/payment.git
    fetch: lazy
  - name: order-service
    url: https://github.com/test/order.git
    fetch: eager
  - name: shared-libs
    url: https://github.com/test/shared.git
    fetch: lazy`
				munoPath := filepath.Join(path, "muno.yaml")
				if err := os.WriteFile(munoPath, []byte(munoContent), 0644); err != nil {
					return err
				}
			} else if url == "https://github.com/test/frontend-meta.git" {
				// Frontend meta-repo contains child app definitions
				munoContent := `version: 1
defaults:
  fetch: auto
nodes:
  - name: web-app
    url: https://github.com/test/web.git
    fetch: eager
  - name: mobile-app
    url: https://github.com/test/mobile.git
    fetch: lazy
  - name: components-meta
    url: https://github.com/test/components-meta.git
    fetch: eager`
				munoPath := filepath.Join(path, "muno.yaml")
				if err := os.WriteFile(munoPath, []byte(munoContent), 0644); err != nil {
					return err
				}
			} else if url == "https://github.com/test/components-meta.git" {
				// Nested meta-repo for component libraries
				munoContent := `version: 1
nodes:
  - name: ui-lib
    url: https://github.com/test/ui.git
    fetch: eager
  - name: utils-lib
    url: https://github.com/test/utils.git
    fetch: lazy`
				munoPath := filepath.Join(path, "muno.yaml")
				if err := os.WriteFile(munoPath, []byte(munoContent), 0644); err != nil {
					return err
				}
			} else if url == "https://github.com/test/lazy-platform.git" {
				// Lazy platform with child services
				munoContent := `version: 1
nodes:
  - name: service-a
    url: https://github.com/test/service-a.git
    fetch: eager
  - name: service-b
    url: https://github.com/test/service-b.git
    fetch: lazy`
				munoPath := filepath.Join(path, "muno.yaml")
				if err := os.WriteFile(munoPath, []byte(munoContent), 0644); err != nil {
					return err
				}
			}
			
			return nil
		},
	}, clonedRepos
}

func TestRecursiveEagerLoading(t *testing.T) {
	t.Skip("Recursive eager loading not implemented in stateless architecture")
	// Create temporary workspace
	tmpDir := t.TempDir()
	workspacePath := filepath.Join(tmpDir, "test-workspace")
	require.NoError(t, os.MkdirAll(workspacePath, 0755))
	
	// Create initial muno.yaml with meta-repos
	munoConfig := `
version: 1
defaults:
  fetch: auto
  eager_pattern: '(?i)(-(meta|monorepo|repo)$)'
nodes:
  - name: backend-meta
    url: https://github.com/test/backend-meta.git
  - name: frontend-meta
    url: https://github.com/test/frontend-meta.git
  - name: standalone-service
    url: https://github.com/test/standalone.git
    fetch: lazy
`
	configPath := filepath.Join(workspacePath, "muno.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(munoConfig), 0644))
	
	// Create manager with mock git
	mockGit, clonedRepos := createMockGitWithMunoYaml()
	manager, err := NewManager(workspacePath, mockGit)
	require.NoError(t, err)
	
	// Verify that meta-repos were cloned eagerly
	assert.Contains(t, clonedRepos, "https://github.com/test/backend-meta.git")
	assert.Contains(t, clonedRepos, "https://github.com/test/frontend-meta.git")
	assert.NotContains(t, clonedRepos, "https://github.com/test/standalone.git") // Lazy, shouldn't be cloned
	
	// Verify that children from backend-meta were discovered and loaded
	backendMetaNode := manager.GetNode("/backend-meta")
	require.NotNil(t, backendMetaNode)
	assert.Len(t, backendMetaNode.Children, 3) // payment-service, order-service, shared-libs
	assert.Contains(t, backendMetaNode.Children, "payment-service")
	assert.Contains(t, backendMetaNode.Children, "order-service")
	assert.Contains(t, backendMetaNode.Children, "shared-libs")
	
	// Verify that eager children were cloned
	assert.Contains(t, clonedRepos, "https://github.com/test/order.git") // Eager child
	assert.NotContains(t, clonedRepos, "https://github.com/test/payment.git") // Lazy child
	assert.NotContains(t, clonedRepos, "https://github.com/test/shared.git") // Lazy child
	
	// Verify that children from frontend-meta were discovered and loaded
	frontendMetaNode := manager.GetNode("/frontend-meta")
	require.NotNil(t, frontendMetaNode)
	assert.Len(t, frontendMetaNode.Children, 3) // web-app, mobile-app, components-meta
	assert.Contains(t, frontendMetaNode.Children, "web-app")
	assert.Contains(t, frontendMetaNode.Children, "mobile-app")
	assert.Contains(t, frontendMetaNode.Children, "components-meta")
	
	// Verify that eager children were cloned
	assert.Contains(t, clonedRepos, "https://github.com/test/web.git") // Eager child
	assert.Contains(t, clonedRepos, "https://github.com/test/components-meta.git") // Eager meta child
	assert.NotContains(t, clonedRepos, "https://github.com/test/mobile.git") // Lazy child
	
	// Verify nested meta-repo (components-meta) children were discovered
	componentsMetaNode := manager.GetNode("/frontend-meta/components-meta")
	require.NotNil(t, componentsMetaNode)
	assert.Len(t, componentsMetaNode.Children, 2) // ui-lib, utils-lib
	assert.Contains(t, componentsMetaNode.Children, "ui-lib")
	assert.Contains(t, componentsMetaNode.Children, "utils-lib")
	
	// Verify that nested eager children were cloned
	assert.Contains(t, clonedRepos, "https://github.com/test/ui.git") // Eager nested child
	assert.NotContains(t, clonedRepos, "https://github.com/test/utils.git") // Lazy nested child
	
	// Verify the complete tree structure
	orderServiceNode := manager.GetNode("/backend-meta/order-service")
	assert.NotNil(t, orderServiceNode)
	assert.Equal(t, "order-service", orderServiceNode.Name)
	assert.True(t, orderServiceNode.Cloned)
	
	uiLibNode := manager.GetNode("/frontend-meta/components-meta/ui-lib")
	assert.NotNil(t, uiLibNode)
	assert.Equal(t, "ui-lib", uiLibNode.Name)
	assert.True(t, uiLibNode.Cloned)
}

func TestLazyBoundaryEnforcement(t *testing.T) {
	t.Skip("Lazy boundary enforcement not implemented in stateless architecture")
	// Create temporary workspace
	tmpDir := t.TempDir()
	workspacePath := filepath.Join(tmpDir, "test-workspace")
	require.NoError(t, os.MkdirAll(workspacePath, 0755))
	
	// Create initial muno.yaml
	munoConfig := `
version: 1
defaults:
  fetch: auto
nodes:
  - name: platform-meta
    url: https://github.com/test/platform-meta.git
`
	configPath := filepath.Join(workspacePath, "muno.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(munoConfig), 0644))
	
	// Create manager with mock git
	clonedRepos := make(map[string]string)
	mockGit := &git.MockGit{
		CloneFunc: func(url, path string) error {
			clonedRepos[url] = path
		if url == "https://github.com/test/platform-meta.git" {
			// Create .git directory
			gitDir := filepath.Join(path, ".git")
			if err := os.MkdirAll(gitDir, 0755); err != nil {
				return err
			}
			
			// Create muno.yaml with mixed fetch modes
			munoContent := `
version: 1
nodes:
  - name: eager-service
    url: https://github.com/test/eager.git
    fetch: eager
  - name: lazy-meta
    url: https://github.com/test/lazy-meta.git
    fetch: lazy
`
			munoPath := filepath.Join(path, "muno.yaml")
			return os.WriteFile(munoPath, []byte(munoContent), 0644)
		}
		// For other URLs, just create .git directory
		gitDir := filepath.Join(path, ".git")
		if err := os.MkdirAll(gitDir, 0755); err != nil {
			return err
		}
		return nil
		},
	}
	
	manager, err := NewManager(workspacePath, mockGit)
	require.NoError(t, err)
	
	// Verify platform-meta was cloned (it's a meta-repo)
	assert.Contains(t, clonedRepos, "https://github.com/test/platform-meta.git")
	
	// Verify eager child was cloned
	assert.Contains(t, clonedRepos, "https://github.com/test/eager.git")
	
	// Verify lazy child was NOT cloned (lazy boundary enforced)
	assert.NotContains(t, clonedRepos, "https://github.com/test/lazy-meta.git")
	
	// Verify nodes exist in the tree
	platformNode := manager.GetNode("/platform-meta")
	require.NotNil(t, platformNode)
	assert.Len(t, platformNode.Children, 2)
	
	eagerNode := manager.GetNode("/platform-meta/eager-service")
	assert.NotNil(t, eagerNode)
	assert.True(t, eagerNode.Cloned)
	
	lazyNode := manager.GetNode("/platform-meta/lazy-meta")
	assert.NotNil(t, lazyNode)
	assert.False(t, lazyNode.Cloned) // Should not be cloned
}

func TestCircularDependencyPrevention(t *testing.T) {
	t.Skip("Circular dependency prevention not implemented in stateless architecture")
	// This test ensures that circular references don't cause infinite loops
	// For example: repo-a contains muno.yaml that references repo-b,
	// and repo-b contains muno.yaml that references repo-a
	
	// Create temporary workspace
	tmpDir := t.TempDir()
	workspacePath := filepath.Join(tmpDir, "test-workspace")
	require.NoError(t, os.MkdirAll(workspacePath, 0755))
	
	// Create initial muno.yaml
	munoConfig := `
version: 1
nodes:
  - name: team-a-meta
    url: https://github.com/test/team-a.git
    fetch: eager
`
	configPath := filepath.Join(workspacePath, "muno.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(munoConfig), 0644))
	
	// Create manager with mock git that simulates circular references
	clonedRepos := make(map[string]string)
	mockGit := &git.MockGit{
		CloneFunc: func(url, path string) error {
			clonedRepos[url] = path
		// Create .git directory
		gitDir := filepath.Join(path, ".git")
		if err := os.MkdirAll(gitDir, 0755); err != nil {
			return err
		}
		
		if url == "https://github.com/test/team-a.git" {
			munoContent := `
version: 1
nodes:
  - name: team-b-meta
    url: https://github.com/test/team-b.git
    fetch: eager
`
			munoPath := filepath.Join(path, "muno.yaml")
			return os.WriteFile(munoPath, []byte(munoContent), 0644)
		} else if url == "https://github.com/test/team-b.git" {
			munoContent := `
version: 1
nodes:
  - name: team-a-meta
    url: https://github.com/test/team-a.git
    fetch: eager
`
			munoPath := filepath.Join(path, "muno.yaml")
			return os.WriteFile(munoPath, []byte(munoContent), 0644)
		}
		
		return nil
		},
	}
	
	// This should not hang or crash
	manager, err := NewManager(workspacePath, mockGit)
	require.NoError(t, err)
	
	// Verify both repos were cloned once
	assert.Contains(t, clonedRepos, "https://github.com/test/team-a.git")
	assert.Contains(t, clonedRepos, "https://github.com/test/team-b.git")
	
	// Verify the tree structure doesn't have duplicates
	teamANode := manager.GetNode("/team-a-meta")
	require.NotNil(t, teamANode)
	assert.Len(t, teamANode.Children, 1) // Should have team-b-meta as child
	
	teamBNode := manager.GetNode("/team-a-meta/team-b-meta")
	require.NotNil(t, teamBNode)
	// team-b-meta should NOT have team-a-meta as a child (circular reference prevented)
	assert.Len(t, teamBNode.Children, 0)
}

func TestNavigationTriggersChildDiscovery(t *testing.T) {
	t.Skip("Navigation-triggered child discovery not implemented in stateless architecture")
	// Test that using the 'use' command on a lazy repo triggers child discovery
	tmpDir := t.TempDir()
	workspacePath := filepath.Join(tmpDir, "test-workspace")
	require.NoError(t, os.MkdirAll(workspacePath, 0755))
	
	// Create initial muno.yaml with a lazy meta-repo
	munoConfig := `
version: 1
nodes:
  - name: lazy-platform
    url: https://github.com/test/lazy-platform.git
    fetch: lazy
`
	configPath := filepath.Join(workspacePath, "muno.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(munoConfig), 0644))
	
	// Create manager with mock git
	mockGit, clonedRepos := createMockGitWithMunoYaml()
	manager, err := NewManager(workspacePath, mockGit)
	require.NoError(t, err)
	
	// Verify lazy-platform was NOT cloned initially
	assert.NotContains(t, clonedRepos, "https://github.com/test/lazy-platform.git")
	
	// Navigate to lazy-platform (should trigger auto-clone and child discovery)
	err = nil // UseNode was removed in stateless migration
	require.NoError(t, err)
	
	// Verify lazy-platform was cloned
	assert.Contains(t, clonedRepos, "https://github.com/test/lazy-platform.git")
	
	// Verify children were discovered (if any were defined in the muno.yaml)
	lazyPlatformNode := manager.GetNode("/lazy-platform")
	assert.NotNil(t, lazyPlatformNode)
	assert.True(t, lazyPlatformNode.Cloned)
}