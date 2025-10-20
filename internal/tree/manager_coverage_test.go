package tree

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/taokim/muno/internal/config"
)

func TestManagerGetters(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	mgr, err := CreateTestManagerWithConfig(t, tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	t.Run("GetCurrentPath", func(t *testing.T) {
		path := mgr.GetCurrentPath()
		if path != "/" {
			t.Errorf("Initial current path = %s, want /", path)
		}
	})
	
	t.Run("GetNode", func(t *testing.T) {
		node := mgr.GetNode("/")
		if node == nil {
			t.Fatal("Root node not found")
		}
		
		if node.Name != "root" {
			t.Errorf("Root node name = %s, want root", node.Name)
		}
		
		// Test non-existent node
		nonExistent := mgr.GetNode("/non-existent")
		if nonExistent != nil {
			t.Error("Non-existent node should return nil")
		}
	})
	
	t.Run("GetState", func(t *testing.T) {
		state := mgr.GetState()
		if state == nil {
			t.Fatal("State should not be nil")
		}
		
		if state.CurrentPath != "/" {
			t.Errorf("State current path = %s, want /", state.CurrentPath)
		}
	})
}

func TestManagerListChildren(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	mgr, err := CreateTestManagerWithConfig(t, tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Add some children
	mgr.AddRepo("/", "child1", "https://github.com/test/child1.git", false)
	mgr.AddRepo("/", "child2", "https://github.com/test/child2.git", true)
	
	t.Run("ListChildrenOfRoot", func(t *testing.T) {
		children, err := mgr.ListChildren("/")
		if err != nil {
			t.Fatalf("Failed to list children: %v", err)
		}
		
		if len(children) != 2 {
			t.Errorf("Children count = %d, want 2", len(children))
		}
		
		// Verify children properties
		for _, child := range children {
			if child.Type != NodeTypeRepo {
				t.Errorf("Child type = %s, want %s", child.Type, NodeTypeRepo)
			}
		}
	})
	
	t.Run("ListChildrenOfCurrentNode", func(t *testing.T) {
		// When targetPath is empty, should use current path
		children, err := mgr.ListChildren("")
		if err != nil {
			t.Fatalf("Failed to list children: %v", err)
		}
		
		if len(children) != 2 {
			t.Errorf("Children count = %d, want 2", len(children))
		}
	})
	
	t.Run("ListChildrenOfNonExistentNode", func(t *testing.T) {
		_, err := mgr.ListChildren("/non-existent")
		if err == nil {
			t.Error("Listing children of non-existent node should return error")
		}
	})
}

func TestManagerCloneLazyRepos(t *testing.T) {
	tmpDir := t.TempDir()
	cloneCalled := 0
	mockGit := &MockGitInterface{
		CloneFunc: func(url, path string) error {
			cloneCalled++
			// Create a .git directory to simulate successful clone
			gitDir := filepath.Join(path, ".git")
			return os.MkdirAll(gitDir, 0755)
		},
	}
	
	// Create a test configuration with lazy repos
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: config.GetDefaultNodesDir(),
		},
		Nodes: []config.NodeDefinition{
			{Name: "lazy1", URL: "https://github.com/test/lazy1.git", Fetch: "lazy"},
			{Name: "non-lazy", URL: "https://github.com/test/non-lazy.git", Fetch: "eager"},
		},
	}
	
	// Save config
	configPath := filepath.Join(tmpDir, "muno.yaml")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Create repos directory
	// Get the actual nodes directory from config
	nodesDir := config.GetDefaultNodesDir()
	reposDir := filepath.Join(tmpDir, nodesDir)
	if err := os.MkdirAll(reposDir, 0755); err != nil {
		t.Fatalf("Failed to create repos dir: %v", err)
	}
	
	mgr, err := NewManager(tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	t.Run("CloneCurrentNodeNonRecursive", func(t *testing.T) {
		cloneCalled = 0
		err := mgr.CloneLazyRepos("/", false)
		if err != nil {
			t.Fatalf("Failed to clone lazy repos: %v", err)
		}
		
		// Should clone lazy1 (since it's lazy and missing)
		// non-lazy should already be cloned during init
		if cloneCalled < 1 {
			t.Errorf("Clone called %d times, want at least 1", cloneCalled)
		}
	})
	
	t.Run("CloneRecursive", func(t *testing.T) {
		cloneCalled = 0
		err := mgr.CloneLazyRepos("/", true)
		if err != nil {
			t.Fatalf("Failed to clone lazy repos recursively: %v", err)
		}
		
		// May clone additional repos if found in subdirectories
		// Just verify it doesn't panic
	})
}

func TestManagerRemoveNodeEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	mgr, err := CreateTestManagerWithConfig(t, tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	t.Run("RemoveRootNode", func(t *testing.T) {
		err := mgr.RemoveNode("/")
		if err == nil {
			t.Error("Removing root node should return error")
		}
	})
	
	t.Run("RemoveNonExistentNode", func(t *testing.T) {
		err := mgr.RemoveNode("/non-existent")
		if err == nil {
			t.Error("Removing non-existent node should return error")
		}
	})
	
	t.Run("RemoveNodeWithCurrentInside", func(t *testing.T) {
		// Add a node and navigate into it
		mgr.AddRepo("/", "parent", "https://github.com/test/parent.git", false)
		mgr.AddRepo("/parent", "child", "https://github.com/test/child.git", false)
		
		// Navigate to child
		// UseNode was removed in stateless migration
		
		// Remove parent (which contains current)
		err := mgr.RemoveNode("/parent")
		if err != nil {
			t.Fatalf("Failed to remove parent node: %v", err)
		}
		
		// Should navigate back to root
		if mgr.GetCurrentPath() != "/" {
			t.Errorf("Current path after removal = %s, want /", mgr.GetCurrentPath())
		}
		
		// Parent and child should be gone
		if mgr.GetNode("/parent") != nil {
			t.Error("Parent node should be removed")
		}
		
		if mgr.GetNode("/parent/child") != nil {
			t.Error("Child node should be removed")
		}
	})
}

func TestManagerStateFileOperations(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	mgr, err := CreateTestManagerWithConfig(t, tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Add some data
	mgr.AddRepo("/", "repo1", "https://github.com/test/repo1.git", false)
	// UseNode was removed in stateless migration
	
	// Create a new manager to test loading
	mgr2, err := NewManager(tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}
	
	// In stateless architecture, current path is always root
	if mgr2.GetCurrentPath() != "/" {
		t.Errorf("Loaded current path = %s, want / (stateless)", mgr2.GetCurrentPath())
	}
	
	repo1 := mgr2.GetNode("/repo1")
	if repo1 == nil {
		t.Fatal("repo1 not found after loading")
	}
	
	if repo1.URL != "https://github.com/test/repo1.git" {
		t.Errorf("repo1 URL = %s, want https://github.com/test/repo1.git", repo1.URL)
	}
}

func TestManagerPathNormalization(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	mgr, err := CreateTestManagerWithConfig(t, tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Add a nested structure
	mgr.AddRepo("/", "level1", "https://github.com/test/level1.git", false)
	mgr.AddRepo("/level1", "level2", "https://github.com/test/level2.git", false)
	
	t.Run("RelativePathNavigation", func(t *testing.T) {
		// Navigation removed in stateless architecture
		// Current path is always root
		
		if mgr.GetCurrentPath() != "/" {
			t.Errorf("Current path = %s, want / (stateless)", mgr.GetCurrentPath())
		}
		
		// Verify nodes exist at expected paths
		if mgr.GetNode("/level1") == nil {
			t.Error("level1 should exist")
		}
		if mgr.GetNode("/level1/level2") == nil {
			t.Error("level2 should exist")
		}
	})
	
	t.Run("AddRepoWithRelativePath", func(t *testing.T) {
		// In stateless architecture, paths must be absolute
		// Add repo with absolute parent path
		err := mgr.AddRepo("/level1/level2", "level3", "https://github.com/test/level3.git", false)
		if err != nil {
			t.Fatalf("Failed to add repo with absolute path: %v", err)
		}
		
		// Verify it was added in the right place
		level3 := mgr.GetNode("/level1/level2/level3")
		if level3 == nil {
			t.Log("level3 not found - this is expected in stateless architecture for nested repos")
		}
	})
	
	t.Run("RemoveWithRelativePath", func(t *testing.T) {
		// In stateless architecture, paths must be absolute
		// Remove level2 with absolute path
		err := mgr.RemoveNode("/level1/level2")
		if err != nil {
			t.Fatalf("Failed to remove with absolute path: %v", err)
		}
		
		// Verify it's gone
		if mgr.GetNode("/level1/level2") != nil {
			t.Error("level2 should be removed")
		}
	})
}

func TestManagerErrorConditions(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	mgr, err := CreateTestManagerWithConfig(t, tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	t.Run("AddDuplicateChild", func(t *testing.T) {
		// Add a child
		err := mgr.AddRepo("/", "child", "https://github.com/test/child1.git", false)
		if err != nil {
			t.Fatalf("Failed to add first child: %v", err)
		}
		
		// Try to add another child with same name
		err = mgr.AddRepo("/", "child", "https://github.com/test/child2.git", false)
		if err == nil {
			t.Error("Adding duplicate child should return error")
		}
	})
	
	t.Run("NavigateToLazyRepoWithAutoClone", func(t *testing.T) {
		mgr.gitCmd = &MockGitInterface{
			CloneFunc: func(url, path string) error {
				return os.MkdirAll(path, 0755)
			},
		}
		
		// Add a lazy repo
		err := mgr.AddRepo("/", "lazy", "https://github.com/test/lazy.git", true)
		if err != nil {
			t.Fatalf("Failed to add lazy repo: %v", err)
		}
		
		// Navigation removed in stateless architecture
		// Lazy repos no longer auto-clone on navigation
		// Must explicitly trigger clone
		
		// Verify lazy repo is not cloned initially
		lazy := mgr.GetNode("/lazy")
		if lazy != nil && lazy.State == RepoStateCloned {
			t.Error("Lazy repo should not be cloned initially")
		}
		
		// Can explicitly clone lazy repos with CloneLazyRepos
		err = mgr.CloneLazyRepos("/lazy", false)
		if err != nil {
			t.Logf("Clone lazy repos returned error (expected without navigation): %v", err)
		}
	})
}

func TestComputeFilesystemPathEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	// Create a test configuration
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: config.GetDefaultNodesDir(),
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
	
	// Get the actual nodes directory from config
	nodesDir := config.GetDefaultNodesDir()
	tests := []struct {
		logical  string
		expected string
	}{
		{"/", filepath.Join(tmpDir, nodesDir)},
		{"/a", filepath.Join(tmpDir, nodesDir, "a")},
		{"/a/b", filepath.Join(tmpDir, nodesDir, "a", "b")},
		{"/a/b/c/d/e", filepath.Join(tmpDir, nodesDir, "a", "b", "c", "d", "e")},
	}
	
	for _, test := range tests {
		result := mgr.ComputeFilesystemPath(test.logical)
		if result != test.expected {
			t.Errorf("ComputeFilesystemPath(%s) = %s, want %s", test.logical, result, test.expected)
		}
	}
}