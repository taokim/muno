package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/taokim/muno/internal/constants"
)

// TestableHelpers contains extracted functions that are easier to test
type TestableHelpers struct{}

// FindWorkspaceRoot finds the workspace root by looking for muno.yaml
// This is extracted from the private findWorkspaceRoot function
func (h *TestableHelpers) FindWorkspaceRoot(startPath string) string {
	if startPath == "" {
		startPath, _ = os.Getwd()
	}

	current := startPath
	for {
		configPath := filepath.Join(current, "muno.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached root
			return ""
		}
		current = parent
	}
}

// EnsureGitignoreEntry ensures an entry exists in .gitignore
// This is extracted from the private ensureGitignoreEntry function
func (h *TestableHelpers) EnsureGitignoreEntry(workspacePath, entry string) error {
	gitignorePath := filepath.Join(workspacePath, ".gitignore")
	
	// Read existing content
	content := ""
	if data, err := os.ReadFile(gitignorePath); err == nil {
		content = string(data)
	}

	// Check if entry already exists
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == entry || trimmed == "/"+entry {
			// Entry already exists
			return nil
		}
	}

	// Add entry
	if !strings.HasSuffix(content, "\n") && content != "" {
		content += "\n"
	}
	content += entry + "\n"

	// Write back
	return os.WriteFile(gitignorePath, []byte(content), 0644)
}

// GenerateTreeContext generates a markdown context for the tree
// This is extracted from generateTreeContext function
func (h *TestableHelpers) GenerateTreeContext(workspaceName, currentPath string) string {
	var sb strings.Builder
	
	sb.WriteString("# Repository Structure\n\n")
	sb.WriteString(fmt.Sprintf("Workspace: %s\n", workspaceName))
	
	if currentPath != "" {
		sb.WriteString(fmt.Sprintf("Current location: %s\n\n", currentPath))
	}
	
	sb.WriteString("## Commands\n\n")
	sb.WriteString("- `muno tree` - Display repository tree\n")
	sb.WriteString("- `muno use <path>` - Navigate to repository\n")
	sb.WriteString("- `muno add <url>` - Add new repository\n")
	sb.WriteString("- `muno list` - List repositories at current level\n")
	
	return sb.String()
}

// CreateAgentContextFile creates a context file for AI agents
func (h *TestableHelpers) CreateAgentContextFile(content string) (string, error) {
	// Create temp file
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("muno-context-%d.md", os.Getpid()))
	
	// Write context to file
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("writing context file: %w", err)
	}
	
	return tempFile, nil
}

// ComputeNodePath computes the filesystem path for a node
func (h *TestableHelpers) ComputeNodePath(workspacePath, nodePath string) string {
	// Use the default repos directory from constants
	reposDir := constants.DefaultReposDir
	
	if nodePath == "" || nodePath == "/" {
		return filepath.Join(workspacePath, reposDir)
	}
	
	// Remove leading slash
	nodePath = strings.TrimPrefix(nodePath, "/")
	
	return filepath.Join(workspacePath, reposDir, nodePath)
}

// IsGitRepository checks if a path contains a git repository
func (h *TestableHelpers) IsGitRepository(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	return err == nil && info.IsDir()
}

// ExtractRepoName extracts repository name from URL
func (h *TestableHelpers) ExtractRepoName(url string) string {
	// Remove trailing .git
	url = strings.TrimSuffix(url, ".git")
	
	// Get last part of path
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	
	return ""
}

// NormalizePath normalizes a path for consistent handling
func (h *TestableHelpers) NormalizePath(path string) string {
	// Ensure path starts with /
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	// Remove trailing slash except for root
	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	
	return path
}

// SplitPath splits a path into segments
func (h *TestableHelpers) SplitPath(path string) []string {
	path = h.NormalizePath(path)
	if path == "/" || path == "" {
		return []string{}
	}
	
	// Remove leading slash and split
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

// JoinPath joins path segments
func (h *TestableHelpers) JoinPath(segments ...string) string {
	if len(segments) == 0 {
		return "/"
	}
	
	path := "/" + strings.Join(segments, "/")
	return h.NormalizePath(path)
}