package tree

import (
	"path/filepath"
	"testing"
	
	"github.com/taokim/muno/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestCustomReposDir verifies that custom repos_dir from config is used correctly
func TestCustomReposDir(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	// Test with custom repos_dir
	customDir := "repositories"
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: customDir, // Custom directory name
		},
		Nodes: []config.NodeDefinition{},
	}
	
	// Save config
	configPath := filepath.Join(tmpDir, "muno.yaml")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	mgr, err := NewManager(tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Test that custom repos_dir is used
	tests := []struct {
		logical  string
		expected string
	}{
		{"/", filepath.Join(tmpDir, customDir)},
		{"/level1", filepath.Join(tmpDir, customDir, "level1")},
		{"/level1/level2", filepath.Join(tmpDir, customDir, "level1", "level2")},
		{"/a/b/c", filepath.Join(tmpDir, customDir, "a", "b", "c")},
	}
	
	for _, test := range tests {
		result := mgr.ComputeFilesystemPath(test.logical)
		assert.Equal(t, test.expected, result, "Path for %s should use custom repos_dir '%s'", test.logical, customDir)
	}
}

// TestStatelessManagerCustomReposDir verifies StatelessManager uses custom repos_dir
func TestStatelessManagerCustomReposDir(t *testing.T) {
	tmpDir := t.TempDir()
	mockGit := &MockGitInterface{}
	
	// Test with custom repos_dir
	customDir := "repos"
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: customDir,
		},
		Nodes: []config.NodeDefinition{},
	}
	
	// Save config
	configPath := filepath.Join(tmpDir, "muno.yaml")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	mgr, err := NewStatelessManager(tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create stateless manager: %v", err)
	}
	
	// Test that custom repos_dir is used
	tests := []struct {
		logical  string
		expected string
	}{
		{"/", filepath.Join(tmpDir, customDir)},
		{"/level1", filepath.Join(tmpDir, customDir, "level1")},
		{"/level1/level2", filepath.Join(tmpDir, customDir, "level1", "level2")},
		{"/parent/child", filepath.Join(tmpDir, customDir, "parent", "child")},
	}
	
	for _, test := range tests {
		result := mgr.ComputeFilesystemPath(test.logical)
		assert.Equal(t, test.expected, result, "StatelessManager path for %s should use custom repos_dir '%s'", test.logical, customDir)
	}
}