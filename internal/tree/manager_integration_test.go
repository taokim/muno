package tree

import (
	"encoding/json"
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
	
	// Read and verify state file
	t.Run("StateFileVerification", func(t *testing.T) {
		statePath := filepath.Join(tmpDir, ".muno-tree.json")
		data, err := os.ReadFile(statePath)
		if err != nil {
			t.Fatalf("Failed to read state file: %v", err)
		}
		
		var state TreeState
		if err := json.Unmarshal(data, &state); err != nil {
			t.Fatalf("Failed to unmarshal state: %v", err)
		}
		
		// Convert to string for searching
		stateStr := string(data)
		
		// Verify NO filesystem paths in state
		if strings.Contains(stateStr, tmpDir) {
			t.Errorf("State file contains filesystem path: %s", tmpDir)
			t.Logf("State content:\n%s", stateStr)
		}
		
		if strings.Contains(stateStr, "nodes/") {
			t.Errorf("State file contains filesystem directory structure 'nodes/'")
		}
		
		// Verify logical paths ARE present
		expectedPaths := []string{
			`"/level1"`,
			`"/level1/level2"`,
			`"/level1/level2/level3"`,
			`"/shared"`,
			`"/level1/sibling"`,
		}
		
		for _, path := range expectedPaths {
			if !strings.Contains(stateStr, path) {
				t.Errorf("State file missing logical path: %s", path)
			}
		}
		
		// Verify tree structure
		if !strings.Contains(stateStr, `"children"`) {
			t.Error("State file missing 'children' field")
		}
		
		if !strings.Contains(stateStr, `"type"`) {
			t.Error("State file missing 'type' field")
		}
		
		// Verify specific node properties
		level1Node := state.Nodes["/level1"]
		if level1Node == nil {
			t.Fatal("level1 node not found in state")
		}
		
		if level1Node.Name != "level1" {
			t.Errorf("level1 node name = %s, want level1", level1Node.Name)
		}
		
		if len(level1Node.Children) != 2 {
			t.Errorf("level1 children count = %d, want 2", len(level1Node.Children))
		}
		
		// Verify lazy repo state
		level3Node := state.Nodes["/level1/level2/level3"]
		if level3Node == nil {
			t.Fatal("level3 node not found in state")
		}
		
		if !level3Node.Lazy {
			t.Error("level3 should be marked as lazy")
		}
		
		if level3Node.State != RepoStateMissing {
			t.Errorf("level3 state = %s, want %s", level3Node.State, RepoStateMissing)
		}
	})
	
	// Test navigation and state updates
	t.Run("NavigationStateUpdate", func(t *testing.T) {
		// Navigate to level2
		if err := mgr.UseNode("/level1/level2"); err != nil {
			t.Fatalf("Failed to navigate: %v", err)
		}
		
		if mgr.state.CurrentPath != "/level1/level2" {
			t.Errorf("Current path = %s, want /level1/level2", mgr.state.CurrentPath)
		}
		
		// Verify state file still has no filesystem paths after navigation
		statePath := filepath.Join(tmpDir, ".muno-tree.json")
		data, err := os.ReadFile(statePath)
		if err != nil {
			t.Fatalf("Failed to read state file after navigation: %v", err)
		}
		
		if strings.Contains(string(data), tmpDir) {
			t.Error("State file contains filesystem path after navigation")
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
		"🌳 Workspace",
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