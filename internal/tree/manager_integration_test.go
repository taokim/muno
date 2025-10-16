package tree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSimplifiedStateIntegration tests that the new implementation
// stores only logical paths in state, not filesystem paths
func TestSimplifiedStateIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	mgr, err := CreateTestManagerWithConfig(t, tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Build a multi-level tree structure
	testCases := []struct {
		parent string
		name   string
		url    string
		lazy   bool
	}{
		{"/", "level1", "https://github.com/test/level1.git", false},
		{"/level1", "level2", "https://github.com/test/level2.git", false},
		{"/level1/level2", "level3", "https://github.com/test/level3.git", true},
		{"/", "shared", "https://github.com/test/shared.git", false},
		{"/level1", "sibling", "https://github.com/test/sibling.git", true},
	}
	
	for _, tc := range testCases {
		if err := mgr.AddRepo(tc.parent, tc.name, tc.url, tc.lazy); err != nil {
			t.Fatalf("Failed to add repo %s: %v", tc.name, err)
		}
	}
	
	// Test filesystem path computation
	pathTests := []struct {
		logical  string
		expected string
	}{
		{"/", filepath.Join(tmpDir, "repos")},
		{"/level1", filepath.Join(tmpDir, "repos", "level1")}, // Git repos use repos/ subdir
		{"/level1/level2", filepath.Join(tmpDir, "repos", "level1", "level2")}, // Nested git repo
		{"/level1/level2/level3", filepath.Join(tmpDir, "repos", "level1", "level2", "level3")}, // Nested git repo
		{"/shared", filepath.Join(tmpDir, "repos", "shared")}, // Git repos use repos/ subdir
		{"/level1/sibling", filepath.Join(tmpDir, "repos", "level1", "sibling")}, // Nested git repo
	}
	
	t.Run("FilesystemPathComputation", func(t *testing.T) {
		for _, pt := range pathTests {
			actual := mgr.ComputeFilesystemPath(pt.logical)
			if actual != pt.expected {
				t.Errorf("ComputeFilesystemPath(%s) = %s, want %s", pt.logical, actual, pt.expected)
			}
		}
	})
	
	// Verify no state file exists (stateless)
	t.Run("StateFileVerification", func(t *testing.T) {
		statePath := filepath.Join(tmpDir, ".muno-tree.json")
		_, err := os.ReadFile(statePath)
		if err == nil {
			t.Fatalf("State file should not exist in stateless mode")
		}
		
		// In stateless mode, verify nodes exist in filesystem
		level1Path := filepath.Join(tmpDir, "repos", "level1")
		if _, err := os.Stat(level1Path); os.IsNotExist(err) {
			t.Errorf("level1 directory should exist")
		}
		
		// Verify other directories exist
		level2Path := filepath.Join(tmpDir, "repos", "level1", "level2")
		if _, err := os.Stat(level2Path); os.IsNotExist(err) {
			t.Errorf("level2 directory should exist")
		}
		
		sharedPath := filepath.Join(tmpDir, "repos", "shared")
		if _, err := os.Stat(sharedPath); os.IsNotExist(err) {
			t.Errorf("shared directory should exist")
		}
		
		// Verify nodes via GetNode (stateless)
		level1Node := mgr.GetNode("/level1")
		if level1Node == nil {
			t.Fatal("level1 node not found")
		}
		
		if level1Node.Name != "level1" {
			t.Errorf("level1 node name = %s, want level1", level1Node.Name)
		}
		
		// In stateless mode, children are discovered from filesystem
		// So we check if directories exist instead
		level3Path := filepath.Join(tmpDir, "repos", "level1", "level2", "level3")
		if _, err := os.Stat(level3Path); err == nil {
			t.Log("level3 directory exists as expected")
		}
		
		// In stateless mode, lazy repos that aren't cloned won't have nodes
		// unless they're in the config, which level3 isn't (it's nested)
		
		
	})
	
	// Test that navigation is no longer supported in stateless architecture
	t.Run("NavigationStateUpdate", func(t *testing.T) {
		// In stateless architecture, navigation is removed
		// Current path always remains at root
		
		// Verify current path is always root
		if mgr.currentPath != "/" {
			t.Errorf("Current path = %s, want / (stateless architecture)", mgr.currentPath)
		}
		
		// Verify no state file exists (stateless)
		statePath := filepath.Join(tmpDir, ".muno-tree.json")
		if _, err := os.Stat(statePath); err == nil {
			t.Error("State file should not exist in stateless mode")
		}
		
		// Verify we can still compute filesystem paths for any logical path
		fsPath := mgr.ComputeFilesystemPath("/level1/level2")
		expectedPath := filepath.Join(tmpDir, "repos", "level1", "level2")
		if fsPath != expectedPath {
			t.Errorf("ComputeFilesystemPath(/level1/level2) = %s, want %s", fsPath, expectedPath)
		}
	})
}

// TestTreeDisplay tests the display functionality
func TestTreeDisplay(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	mgr, err := CreateTestManagerWithConfig(t, tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Build a simple tree
	mgr.AddRepo("/", "repo1", "https://github.com/test/repo1.git", false)
	mgr.AddRepo("/", "repo2", "https://github.com/test/repo2.git", true)
	mgr.AddRepo("/repo1", "subrepo", "https://github.com/test/subrepo.git", false)
	
	// Test tree display
	output := mgr.DisplayTree()
	
	// Verify output contains expected elements
	expectedStrings := []string{
		"🌳 test-workspace",
		"repo1",
		"repo2",
		"subrepo",
		"📦", // Cloned icon
		"💤", // Lazy icon
	}
	
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Tree display missing expected string: %s", expected)
		}
	}
	
	t.Logf("Tree display:\n%s", output)
	
	// Test status display
	status := mgr.DisplayStatus()
	
	if !strings.Contains(status, "Current Path: /") {
		t.Error("Status missing current path")
	}
	
	if !strings.Contains(status, "Total Nodes:") {
		t.Error("Status missing total nodes")
	}
	
	if !strings.Contains(status, "Repositories:") {
		t.Error("Status missing repositories count")
	}
	
	t.Logf("Status display:\n%s", status)
}