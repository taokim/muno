package tree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestStatusFunctions(t *testing.T) {
	t.Run("GetRepoState", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Test missing directory
		state := GetRepoState(filepath.Join(tmpDir, "nonexistent"))
		if state != RepoStateMissing {
			t.Errorf("Expected RepoStateMissing for nonexistent dir, got %s", state)
		}
		
		// Test directory without .git
		plainDir := filepath.Join(tmpDir, "plain")
		os.MkdirAll(plainDir, 0755)
		state = GetRepoState(plainDir)
		if state != RepoStateMissing {
			t.Errorf("Expected RepoStateMissing for non-git dir, got %s", state)
		}
		
		// Test directory with .git (initialize real git repo)
		gitDir := filepath.Join(tmpDir, "repo")
		os.MkdirAll(gitDir, 0755)
		// Initialize as a real git repository
		cmd := exec.Command("git", "init")
		cmd.Dir = gitDir
		if err := cmd.Run(); err != nil {
			t.Skipf("Skipping test: git init failed: %v", err)
		}
		// Configure git user for the test repo
		exec.Command("git", "config", "user.email", "test@example.com").Dir = gitDir
		exec.Command("git", "config", "user.name", "Test User").Dir = gitDir
		
		state = GetRepoState(gitDir)
		if state != RepoStateCloned {
			t.Errorf("Expected RepoStateCloned for git dir, got %s", state)
		}
		
		// Test modified repo (with untracked file)
		os.WriteFile(filepath.Join(gitDir, "new.txt"), []byte("test"), 0644)
		state = GetRepoState(gitDir)
		// With untracked file, should show as modified
		if state != RepoStateModified {
			t.Errorf("Expected RepoStateModified, got %s", state)
		}
	})
	
	t.Run("GetFileStatus", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Test missing config
		status := GetFileStatus(filepath.Join(tmpDir, "nonexistent"))
		if status {
			t.Error("Expected false for nonexistent path")
		}
		
		// Test existing directory
		dirPath := filepath.Join(tmpDir, "existing")
		os.MkdirAll(dirPath, 0755)
		status = GetFileStatus(dirPath)
		if !status {
			t.Error("Expected true for existing directory")
		}
	})
	
	t.Run("CreateFileMarker", func(t *testing.T) {
		tmpDir := t.TempDir()
		repoPath := filepath.Join(tmpDir, "repo")
		
		err := CreateFileMarker(repoPath)
		if err != nil {
			t.Errorf("Failed to create config ref marker: %v", err)
		}
		
		// Check marker file was created
		markerPath := filepath.Join(repoPath, ".muno-ref")
		if _, err := os.Stat(markerPath); os.IsNotExist(err) {
			t.Error("Config ref marker file was not created")
		}
	})
}