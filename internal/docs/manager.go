package docs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Manager handles cross-repository documentation
type Manager struct {
	basePath string
	gitInit  bool
}

// NewManager creates a new documentation manager
func NewManager(projectPath string) (*Manager, error) {
	docsPath := filepath.Join(projectPath, "docs")
	
	// Create docs directory if it doesn't exist
	if err := os.MkdirAll(docsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create docs directory: %w", err)
	}

	// Create subdirectories
	globalPath := filepath.Join(docsPath, "global")
	scopesPath := filepath.Join(docsPath, "scopes")
	
	if err := os.MkdirAll(globalPath, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(scopesPath, 0755); err != nil {
		return nil, err
	}

	m := &Manager{
		basePath: docsPath,
	}

	// Initialize git if not already initialized
	gitPath := filepath.Join(docsPath, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		if err := m.initGit(); err != nil {
			fmt.Printf("Warning: failed to initialize git for docs: %v\n", err)
		} else {
			m.gitInit = true
		}
	} else {
		m.gitInit = true
	}

	// Create .gitignore if it doesn't exist
	gitignorePath := filepath.Join(docsPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gitignoreContent := `# Temporary files
*.tmp
*.swp
.DS_Store

# Ephemeral scope docs (uncomment to exclude)
# scopes/ephemeral-*
# scopes/hotfix-*
# scopes/temp-*
`
		os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	}

	// Create README if it doesn't exist
	readmePath := filepath.Join(globalPath, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		readmeContent := fmt.Sprintf(`# Project Documentation

This directory contains cross-repository documentation for the project.

## Structure

- **global/**: Documentation that applies to the entire project
- **scopes/**: Documentation specific to individual scopes

## Guidelines

1. Use global/ for architecture decisions, standards, and project-wide documentation
2. Use scopes/<scope-name>/ for scope-specific implementation details
3. Keep documentation in sync with code changes
4. Use meaningful commit messages when updating docs

Created: %s
`, time.Now().Format(time.RFC3339))
		os.WriteFile(readmePath, []byte(readmeContent), 0644)
	}

	return m, nil
}

// initGit initializes a git repository for documentation
func (m *Manager) initGit() error {
	cmd := exec.Command("git", "init")
	cmd.Dir = m.basePath
	if err := cmd.Run(); err != nil {
		return err
	}

	// Initial commit
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = m.basePath
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Initial documentation structure")
	cmd.Dir = m.basePath
	cmd.Run()

	return nil
}

// CreateGlobal creates a global documentation file
func (m *Manager) CreateGlobal(name, content string) error {
	// Ensure name has .md extension
	if !strings.HasSuffix(name, ".md") {
		name = name + ".md"
	}

	// Handle subdirectories
	fullPath := filepath.Join(m.basePath, "global", name)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write content
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✅ Created global documentation: %s\n", name)
	return nil
}

// CreateScope creates scope-specific documentation
func (m *Manager) CreateScope(scope, name, content string) error {
	// Ensure name has .md extension
	if !strings.HasSuffix(name, ".md") {
		name = name + ".md"
	}

	// Create scope directory if needed
	scopePath := filepath.Join(m.basePath, "scopes", scope)
	if err := os.MkdirAll(scopePath, 0755); err != nil {
		return fmt.Errorf("failed to create scope directory: %w", err)
	}

	// Handle subdirectories in name
	fullPath := filepath.Join(scopePath, name)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write content
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Create scope README if it doesn't exist
	scopeReadme := filepath.Join(scopePath, "README.md")
	if _, err := os.Stat(scopeReadme); os.IsNotExist(err) {
		readmeContent := fmt.Sprintf(`# Scope: %s

Documentation specific to the %s scope.

## Files
- %s

Created: %s
`, scope, scope, name, time.Now().Format(time.RFC3339))
		os.WriteFile(scopeReadme, []byte(readmeContent), 0644)
	}

	fmt.Printf("✅ Created scope documentation: %s/%s\n", scope, name)
	return nil
}

// Edit opens a documentation file for editing
func (m *Manager) Edit(path string) error {
	fullPath := filepath.Join(m.basePath, path)
	
	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("documentation not found: %s", path)
	}

	// Try to open with default editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // Default to vi
	}

	cmd := exec.Command(editor, fullPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// List lists available documentation
func (m *Manager) List(scope string) ([]DocInfo, error) {
	var docs []DocInfo
	
	if scope == "" {
		// List all documentation
		err := filepath.Walk(m.basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			
			// Skip .git directory
			if strings.Contains(path, ".git") {
				return filepath.SkipDir
			}
			
			// Only include .md files
			if !info.IsDir() && strings.HasSuffix(path, ".md") {
				relPath, _ := filepath.Rel(m.basePath, path)
				docs = append(docs, DocInfo{
					Path:     relPath,
					Size:     info.Size(),
					Modified: info.ModTime(),
					IsGlobal: strings.HasPrefix(relPath, "global/"),
				})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// List scope-specific documentation
		scopePath := filepath.Join(m.basePath, "scopes", scope)
		if _, err := os.Stat(scopePath); os.IsNotExist(err) {
			return docs, nil // Empty list if scope doesn't exist
		}
		
		err := filepath.Walk(scopePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			
			if !info.IsDir() && strings.HasSuffix(path, ".md") {
				relPath, _ := filepath.Rel(m.basePath, path)
				docs = append(docs, DocInfo{
					Path:     relPath,
					Size:     info.Size(),
					Modified: info.ModTime(),
					IsGlobal: false,
					Scope:    scope,
				})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	
	return docs, nil
}

// Sync commits and optionally pushes documentation changes
func (m *Manager) Sync(push bool) error {
	if !m.gitInit {
		return fmt.Errorf("git not initialized for documentation")
	}

	// Check for changes
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = m.basePath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if strings.TrimSpace(string(output)) == "" {
		fmt.Println("No documentation changes to sync")
		return nil
	}

	// Add all changes
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = m.basePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Commit changes
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("Documentation update - %s", timestamp)
	
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = m.basePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	fmt.Println("✅ Documentation changes committed")

	// Push if requested
	if push {
		cmd = exec.Command("git", "push")
		cmd.Dir = m.basePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to push: %w", err)
		}
		fmt.Println("✅ Documentation pushed to remote")
	}

	return nil
}

// GetPath returns a full path for a documentation file
func (m *Manager) GetPath(relativePath string) string {
	return filepath.Join(m.basePath, relativePath)
}

// DocInfo contains information about a documentation file
type DocInfo struct {
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	IsGlobal bool      `json:"is_global"`
	Scope    string    `json:"scope,omitempty"`
}