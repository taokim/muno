package git

import (
	"os"
	"os/exec"
	"testing"
)

// skipIfNoGit skips the test if git is not available
func skipIfNoGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in test environment")
	}
}

// skipIfCI skips the test if running in CI environment
func skipIfCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skipping test in CI environment")
	}
}

// createTestRepo creates a minimal git repository for testing
func createTestRepo(t *testing.T, path string) {
	skipIfNoGit(t)
	
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}
	
	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = path
	cmd.Run()
	
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = path
	cmd.Run()
	
	// Create initial commit
	testFile := path + "/README.md"
	os.WriteFile(testFile, []byte("# Test"), 0644)
	
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = path
	cmd.Run()
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = path
	cmd.Run()
}