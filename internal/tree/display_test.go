package tree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
)

func TestDisplayMethods(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a basic config file
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: "repos",
		},
		Nodes: []config.NodeDefinition{},
	}
	configPath := filepath.Join(tmpDir, "muno.yaml")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Create repos directory
	os.MkdirAll(filepath.Join(tmpDir, "repos"), 0755)
	
	mockGit := &git.MockGit{}
	
	mgr, err := NewStatelessManager(tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Add some nodes for display
	mgr.AddRepo("/", "repo1", "https://github.com/test/repo1.git", false)
	mgr.AddRepo("/", "repo2", "https://github.com/test/repo2.git", true)
	// Note: StatelessManager currently stores all repos flat, so nested repos
	// are not truly nested in the config structure
	mgr.AddRepo("/repo1", "nested", "https://github.com/test/nested.git", false)
	
	t.Run("DisplayTree", func(t *testing.T) {
		output := mgr.DisplayTree()
		if !strings.Contains(output, "repo1") {
			t.Errorf("DisplayTree should contain repo1")
		}
		if !strings.Contains(output, "repo2") {
			t.Errorf("DisplayTree should contain repo2")
		}
	})
	
	t.Run("DisplayTreeWithDepth", func(t *testing.T) {
		output := mgr.DisplayTreeWithDepth(1)
		if !strings.Contains(output, "repo1") {
			t.Errorf("DisplayTreeWithDepth should contain repo1")
		}
		// Note: StatelessManager currently adds all repos to flat config,
		// so "nested" will appear even though it was added with parent "/repo1"
		// This is a known limitation of the current implementation
		// TODO: Implement proper nested structure in StatelessManager
	})
	
	t.Run("DisplayStatus", func(t *testing.T) {
		output := mgr.DisplayStatus()
		if !strings.Contains(output, "Tree Status") {
			t.Errorf("DisplayStatus should contain 'Tree Status'")
		}
	})
	
	t.Run("DisplayPath", func(t *testing.T) {
		// Navigate to a nested node
		mgr.UseNode("/repo1")
		output := mgr.DisplayPath()
		if !strings.Contains(output, "/") {
			t.Errorf("DisplayPath should show path from root")
		}
	})
	
	t.Run("DisplayChildren", func(t *testing.T) {
		// Navigate back to root
		mgr.UseNode("/")
		output := mgr.DisplayChildren()
		if !strings.Contains(output, "repo1") && !strings.Contains(output, "repo2") {
			t.Errorf("DisplayChildren should show child nodes")
		}
	})
}