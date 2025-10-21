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
	return g.CloneWithSSHPreference(url, path, true) // Default to SSH preference enabled
}

// CloneWithSSHPreference clones a repository with SSH preference support
func (g *Git) CloneWithSSHPreference(url, path string, sshPreference bool) error {
	originalURL := url
	
	// Check if repository already exists
	if isRepositoryAlreadyCloned(path) {
		fmt.Printf("ℹ️  Repository already exists at %s, skipping clone\n", path)
		return nil
	}
	
	// If SSH preference is enabled and this is a GitHub HTTPS URL, try SSH first
	if sshPreference {
		if sshURL, isGitHub := GitHubHTTPSToSSH(url); isGitHub {
			fmt.Printf("🔑 Trying SSH clone: %s\n", sshURL)
			err := g.cloneWithURL(sshURL, path)
			if err == nil {
				fmt.Printf("✅ SSH clone successful\n")
				return nil
			}
			
			// Log the SSH failure reason
			fmt.Printf("❌ SSH clone failed: %v\n", err)
			
			// Check if this is a recoverable SSH error that should trigger fallback
			if shouldFallbackToHTTPS(err) {
				if IsSSHAuthError(err) {
					fmt.Printf("🔄 SSH authentication failed, falling back to HTTPS: %s\n", originalURL)
				} else {
					fmt.Printf("🔄 SSH connection failed, falling back to HTTPS: %s\n", originalURL)
				}
				
				// Attempt HTTPS fallback
				fmt.Printf("🌐 Trying HTTPS clone: %s\n", originalURL)
				fallbackErr := g.cloneWithURL(originalURL, path)
				if fallbackErr == nil {
					fmt.Printf("✅ HTTPS clone successful\n")
					return nil
				}
				
				fmt.Printf("❌ HTTPS clone also failed: %v\n", fallbackErr)
				return fallbackErr
			} else {
				// Non-recoverable error (e.g., repository doesn't exist, already exists, etc.)
				return err
			}
		}
	}
	
	// Default: clone with original URL (non-GitHub or SSH disabled)
	fmt.Printf("🌐 Cloning with original URL: %s\n", originalURL)
	return g.cloneWithURL(originalURL, path)
}

// cloneWithURL performs the actual clone operation
func (g *Git) cloneWithURL(url, path string) error {
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