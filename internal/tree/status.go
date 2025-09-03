package tree

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetRepoState checks the filesystem to determine repository status
func GetRepoState(repoPath string) RepoState {
	// Check if .git directory exists
	gitPath := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return RepoStateMissing
	}
	
	// Check for modifications using git status
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		// If git command fails, assume missing
		return RepoStateMissing
	}
	
	// If output is empty, repo is clean
	if len(strings.TrimSpace(string(output))) == 0 {
		return RepoStateCloned
	}
	
	return RepoStateModified
}

// GetConfigRefStatus checks if a config-only node exists
func GetConfigRefStatus(nodePath string) bool {
	// Check for .muno-ref marker file
	markerPath := filepath.Join(nodePath, ".muno-ref")
	if _, err := os.Stat(markerPath); err == nil {
		return true
	}
	
	// Or just check if directory exists
	if info, err := os.Stat(nodePath); err == nil && info.IsDir() {
		return true
	}
	
	return false
}

// CreateConfigRefMarker creates a marker for config-only nodes
func CreateConfigRefMarker(nodePath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(nodePath, 0755); err != nil {
		return err
	}
	
	// Create .muno-ref marker
	markerPath := filepath.Join(nodePath, ".muno-ref")
	file, err := os.Create(markerPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = file.WriteString("# This directory is a MUNO config reference node\n")
	return err
}