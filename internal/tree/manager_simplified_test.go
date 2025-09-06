package tree

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/taokim/muno/internal/config"
)

// MockGitInterface for testing
type MockGitInterface struct {
	CloneFunc func(url, path string) error
}

func (m *MockGitInterface) Clone(url, path string) error {
	if m.CloneFunc != nil {
		return m.CloneFunc(url, path)
	}
	// Default: create the directory to simulate clone
	return os.MkdirAll(path, 0755)
}

func (m *MockGitInterface) Status(path string) (string, error) {
	return "clean", nil
}

func (m *MockGitInterface) Pull(path string) error {
	return nil
}

func (m *MockGitInterface) Commit(path, message string) error {
	return nil
}

func (m *MockGitInterface) Push(path string) error {
	return nil
}

func (m *MockGitInterface) Add(path, pattern string) error {
	return nil
}

func TestComputeFilesystemPath(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	// Create a test configuration
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: "repos",
		},
		Nodes: []config.NodeDefinition{},
	}
	
	// Save config
	configPath := filepath.Join(tmpDir, "muno.yaml")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	mgr, err := CreateTestManagerWithConfig(t, tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	tests := []struct {
		logical  string
		expected string
	}{
		{"/", filepath.Join(tmpDir, "repos")},
		{"/level1", filepath.Join(tmpDir, "repos", "level1")},
		{"/level1/level2", filepath.Join(tmpDir, "repos", "level1", "level2")},
		{"/a/b/c", filepath.Join(tmpDir, "repos", "a", "b", "c")},
	}
	
	for _, test := range tests {
		result := mgr.ComputeFilesystemPath(test.logical)
		if result != test.expected {
			t.Errorf("ComputeFilesystemPath(%s) = %s, want %s", test.logical, result, test.expected)
		}
	}
}

func TestStateManagement(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a test configuration
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: "repos",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Lazy: true},
		},
	}
	
	// Save config
	configPath := filepath.Join(tmpDir, "muno.yaml")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	mockGit := &MockGitInterface{}
	
	// Create stateless manager
	mgr, err := NewStatelessManager(tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create stateless manager: %v", err)
	}
	
	// Verify manager was created
	if mgr == nil {
		t.Fatal("Manager is nil")
	}
	
	// Verify current path
	if mgr.GetCurrentPath() != "/" {
		t.Errorf("Initial current path = %s, want /", mgr.GetCurrentPath())
	}
	
	// Since we're stateless, there's no state file to check
	// All state is derived from filesystem and config
}

func TestAddRepo(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	mgr, err := CreateTestManagerWithConfig(t, tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Add a non-lazy repo
	err = mgr.AddRepo("/", "test-repo", "https://github.com/test/repo.git", false)
	if err != nil {
		t.Fatalf("Failed to add repo: %v", err)
	}
	
	// Verify repo was added to state
	repo := mgr.state.Nodes["/test-repo"]
	if repo == nil {
		t.Fatal("Repo node not found")
	}
	
	if repo.Type != NodeTypeRepo {
		t.Errorf("Repo type = %s, want %s", repo.Type, NodeTypeRepo)
	}
	
	if repo.URL != "https://github.com/test/repo.git" {
		t.Errorf("Repo URL = %s, want https://github.com/test/repo.git", repo.URL)
	}
	
	if repo.State != RepoStateCloned {
		t.Errorf("Repo state = %s, want %s", repo.State, RepoStateCloned)
	}
	
	// Add a lazy repo
	err = mgr.AddRepo("/", "lazy-repo", "https://github.com/test/lazy.git", true)
	if err != nil {
		t.Fatalf("Failed to add lazy repo: %v", err)
	}
	
	lazyRepo := mgr.state.Nodes["/lazy-repo"]
	if lazyRepo == nil {
		t.Fatal("Lazy repo node not found")
	}
	
	if !lazyRepo.Lazy {
		t.Error("Repo should be lazy")
	}
	
	if lazyRepo.State != RepoStateMissing {
		t.Errorf("Lazy repo state = %s, want %s", lazyRepo.State, RepoStateMissing)
	}
	
	// Verify root has both children
	root := mgr.state.Nodes["/"]
	if len(root.Children) != 2 {
		t.Errorf("Root children count = %d, want 2", len(root.Children))
	}
}

func TestTreeNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	mgr, err := CreateTestManagerWithConfig(t, tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Create a multi-level tree
	mgr.AddRepo("/", "level1", "https://github.com/test/level1.git", false)
	mgr.AddRepo("/level1", "level2", "https://github.com/test/level2.git", false)
	mgr.AddRepo("/level1/level2", "level3", "https://github.com/test/level3.git", true)
	
	// Test navigation
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	
	// Navigate to level2
	err = mgr.UseNode("/level1/level2")
	if err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}
	
	if mgr.state.CurrentPath != "/level1/level2" {
		t.Errorf("Current path = %s, want /level1/level2", mgr.state.CurrentPath)
	}
	
	// Verify filesystem path - should be under repos directory
	expectedFS := filepath.Join(tmpDir, config.GetDefaultReposDir(), "level1", "level2")
	actualFS := mgr.ComputeFilesystemPath("/level1/level2")
	if actualFS != expectedFS {
		t.Errorf("Filesystem path = %s, want %s", actualFS, expectedFS)
	}
	
	// Test auto-clone of lazy repo
	err = mgr.UseNode("/level1/level2/level3")
	if err != nil {
		t.Fatalf("Failed to navigate to lazy repo: %v", err)
	}
	
	// Verify lazy repo was cloned
	level3 := mgr.state.Nodes["/level1/level2/level3"]
	if level3.State != RepoStateCloned {
		t.Errorf("Lazy repo state after navigation = %s, want %s", level3.State, RepoStateCloned)
	}
}

func TestRemoveNode(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	mgr, err := CreateTestManagerWithConfig(t, tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Create a tree with multiple levels
	mgr.AddRepo("/", "parent", "https://github.com/test/parent.git", false)
	mgr.AddRepo("/parent", "child1", "https://github.com/test/child1.git", false)
	mgr.AddRepo("/parent", "child2", "https://github.com/test/child2.git", false)
	
	// Remove child1
	err = mgr.RemoveNode("/parent/child1")
	if err != nil {
		t.Fatalf("Failed to remove node: %v", err)
	}
	
	// Verify child1 is gone
	if mgr.state.Nodes["/parent/child1"] != nil {
		t.Error("child1 should be removed")
	}
	
	// Verify parent still has child2
	parent := mgr.state.Nodes["/parent"]
	if len(parent.Children) != 1 || parent.Children[0] != "child2" {
		t.Errorf("Parent children = %v, want [child2]", parent.Children)
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || len(substr) < len(s) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}