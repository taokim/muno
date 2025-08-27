package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Git provides simple git operations
type Git struct{}

// New creates a new Git instance
func New() *Git {
	return &Git{}
}

// Clone clones a repository
func (g *Git) Clone(url, path string) error {
	cmd := exec.Command("git", "clone", url, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s\n%s", err, string(output))
	}
	return nil
}

// Pull pulls latest changes
func (g *Git) Pull(repoPath string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %s\n%s", err, string(output))
	}
	return nil
}

// Add stages changes
func (g *Git) Add(repoPath, pattern string) error {
	cmd := exec.Command("git", "add", pattern)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %s\n%s", err, string(output))
	}
	return nil
}

// Commit creates a commit
func (g *Git) Commit(repoPath, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's just "nothing to commit"
		if strings.Contains(string(output), "nothing to commit") {
			return fmt.Errorf("nothing to commit")
		}
		return fmt.Errorf("git commit failed: %s\n%s", err, string(output))
	}
	return nil
}

// Push pushes changes to remote
func (g *Git) Push(repoPath string) error {
	cmd := exec.Command("git", "push")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %s\n%s", err, string(output))
	}
	return nil
}

// Status gets the status of a repository
func (g *Git) Status(repoPath string) (string, error) {
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git status failed: %s\n%s", err, string(output))
	}
	return string(output), nil
}