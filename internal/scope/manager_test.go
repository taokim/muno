package scope

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/taokim/repo-claude/internal/config"
)

func TestManagerCreation(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := &config.Config{
		Version: 2,
		Workspace: config.WorkspaceConfig{
			Name:     "test",
			BasePath: "workspaces",
		},
		Repositories: map[string]config.Repository{
			"test-repo": {
				URL:           "https://github.com/test/repo.git",
				DefaultBranch: "main",
			},
		},
		Scopes: map[string]config.Scope{
			"test-scope": {
				Type:        "persistent",
				Repos:       []string{"test-repo"},
				Description: "Test scope",
				Model:       "claude-3-sonnet",
			},
		},
	}
	
	mgr, err := NewManager(cfg, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	if mgr.projectPath != tmpDir {
		t.Errorf("Expected project path %s, got %s", tmpDir, mgr.projectPath)
	}
	
	expectedWorkspace := filepath.Join(tmpDir, "workspaces")
	if mgr.workspacePath != expectedWorkspace {
		t.Errorf("Expected workspace path %s, got %s", expectedWorkspace, mgr.workspacePath)
	}
	
	// Check that workspace directory was created
	if _, err := os.Stat(expectedWorkspace); os.IsNotExist(err) {
		t.Error("Workspace directory was not created")
	}
}

func TestScopeCreate(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	tests := []struct {
		name        string
		scopeName   string
		opts        CreateOptions
		expectError bool
	}{
		{
			name:      "Create persistent scope",
			scopeName: "test-persistent",
			opts: CreateOptions{
				Type:        TypePersistent,
				Repos:       []string{"test-repo"},
				Description: "Test persistent scope",
				CloneRepos:  false,
			},
			expectError: false,
		},
		{
			name:      "Create ephemeral scope",
			scopeName: "test-ephemeral",
			opts: CreateOptions{
				Type:        TypeEphemeral,
				Repos:       []string{"test-repo"},
				Description: "Test ephemeral scope",
				CloneRepos:  false,
			},
			expectError: false,
		},
		{
			name:      "Create duplicate scope",
			scopeName: "test-persistent",
			opts: CreateOptions{
				Type: TypePersistent,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.Create(tt.scopeName, tt.opts)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if !tt.expectError {
				// Verify scope was created
				scopePath := filepath.Join(mgr.workspacePath, tt.scopeName)
				if _, err := os.Stat(scopePath); os.IsNotExist(err) {
					t.Error("Scope directory was not created")
				}
				
				// Verify metadata file exists
				metaPath := filepath.Join(scopePath, MetaFileName)
				if _, err := os.Stat(metaPath); os.IsNotExist(err) {
					t.Error("Metadata file was not created")
				}
				
				// Verify shared memory file exists
				sharedMemPath := filepath.Join(scopePath, SharedMemoryFileName)
				if _, err := os.Stat(sharedMemPath); os.IsNotExist(err) {
					t.Error("Shared memory file was not created")
				}
			}
		})
	}
}

func TestScopeList(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create some scopes
	mgr.Create("scope1", CreateOptions{
		Type:        TypePersistent,
		Description: "First scope",
	})
	mgr.Create("scope2", CreateOptions{
		Type:        TypeEphemeral,
		Description: "Second scope",
	})
	
	// List scopes
	scopes, err := mgr.List()
	if err != nil {
		t.Fatalf("Failed to list scopes: %v", err)
	}
	
	if len(scopes) != 2 {
		t.Errorf("Expected 2 scopes, got %d", len(scopes))
	}
	
	// Verify scope details
	scopeNames := make(map[string]bool)
	for _, scope := range scopes {
		scopeNames[scope.Name] = true
		
		if scope.Name == "scope1" && scope.Type != TypePersistent {
			t.Errorf("Expected scope1 to be persistent, got %s", scope.Type)
		}
		if scope.Name == "scope2" && scope.Type != TypeEphemeral {
			t.Errorf("Expected scope2 to be ephemeral, got %s", scope.Type)
		}
	}
	
	if !scopeNames["scope1"] || !scopeNames["scope2"] {
		t.Error("Not all scopes were listed")
	}
}

func TestScopeGet(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create a scope
	scopeName := "test-scope"
	mgr.Create(scopeName, CreateOptions{
		Type:        TypePersistent,
		Description: "Test scope",
		Repos:       []string{"test-repo"},
	})
	
	// Get the scope
	scope, err := mgr.Get(scopeName)
	if err != nil {
		t.Fatalf("Failed to get scope: %v", err)
	}
	
	if scope.meta.Name != scopeName {
		t.Errorf("Expected scope name %s, got %s", scopeName, scope.meta.Name)
	}
	
	if scope.meta.Type != TypePersistent {
		t.Errorf("Expected persistent type, got %s", scope.meta.Type)
	}
	
	// Try to get non-existent scope
	_, err = mgr.Get("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent scope")
	}
}

func TestScopeDelete(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create a scope
	scopeName := "test-delete"
	mgr.Create(scopeName, CreateOptions{
		Type: TypePersistent,
	})
	
	// Verify it exists
	scopePath := filepath.Join(mgr.workspacePath, scopeName)
	if _, err := os.Stat(scopePath); os.IsNotExist(err) {
		t.Fatal("Scope was not created")
	}
	
	// Delete the scope
	err := mgr.Delete(scopeName)
	if err != nil {
		t.Fatalf("Failed to delete scope: %v", err)
	}
	
	// Verify it's gone
	if _, err := os.Stat(scopePath); !os.IsNotExist(err) {
		t.Error("Scope directory still exists after deletion")
	}
	
	// Try to delete non-existent scope
	err = mgr.Delete("non-existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent scope")
	}
}



func TestListWithRepos(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	cfg.Scopes["multi-repo"] = config.Scope{
		Type:        "persistent",
		Repos:       []string{"repo1", "repo2", "repo3"},
		Description: "Multi-repo scope",
	}
	
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create the scope
	mgr.CreateFromConfig("multi-repo")
	
	// List with repos
	details, err := mgr.ListWithRepos()
	if err != nil {
		t.Fatalf("Failed to list with repos: %v", err)
	}
	
	// Find our scope
	var found *ScopeDetail
	for _, detail := range details {
		if detail.Name == "multi-repo" {
			found = &detail
			break
		}
	}
	
	if found == nil {
		t.Fatal("Scope not found in list")
	}
	
	// Check details
	if found.Number != 1 {
		t.Errorf("Expected number 1, got %d", found.Number)
	}
	if found.RepoCount != 3 {
		t.Errorf("Expected 3 repos, got %d", found.RepoCount)
	}
	if len(found.Repos) != 3 {
		t.Errorf("Expected 3 repo names, got %d", len(found.Repos))
	}
}

func TestGetByNumberOrName(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create some scopes
	mgr.Create("first", CreateOptions{Type: TypePersistent})
	mgr.Create("second", CreateOptions{Type: TypePersistent})
	
	tests := []struct {
		name       string
		identifier string
		expectName string
		expectErr  bool
	}{
		{"Get by name", "first", "first", false},
		{"Get by number 1", "1", "first", false},
		{"Get by number 2", "2", "second", false},
		{"Invalid number", "3", "", true},
		{"Invalid name", "nonexistent", "", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope, err := mgr.GetByNumberOrName(tt.identifier)
			
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if !tt.expectErr && scope.meta.Name != tt.expectName {
				t.Errorf("Expected scope %s, got %s", tt.expectName, scope.meta.Name)
			}
		})
	}
}

// Helper function to create test config
func createTestConfig() *config.Config {
	return &config.Config{
		Version: 2,
		Workspace: config.WorkspaceConfig{
			Name:     "test",
			BasePath: "workspaces",
		},
		Repositories: map[string]config.Repository{
			"test-repo": {
				URL:           "https://github.com/test/repo.git",
				DefaultBranch: "main",
			},
			"repo1": {
				URL:           "https://github.com/test/repo1.git",
				DefaultBranch: "main",
			},
			"repo2": {
				URL:           "https://github.com/test/repo2.git",
				DefaultBranch: "main",
			},
			"repo3": {
				URL:           "https://github.com/test/repo3.git",
				DefaultBranch: "main",
			},
		},
		Scopes: map[string]config.Scope{
			"test-scope": {
				Type:        "persistent",
				Repos:       []string{"test-repo"},
				Description: "Test scope",
			},
		},
	}
}