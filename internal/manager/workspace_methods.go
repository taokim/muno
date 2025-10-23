package manager

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/constants"
)

// SmartInitWorkspace performs intelligent initialization for the workspace
func (m *Manager) SmartInitWorkspace(projectName string, options InitOptions) error {
	if !m.initialized {
		m.initialized = true
	}
	
	// Initialize config if not present
	if m.config == nil {
		m.config = &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     projectName,
				ReposDir: config.GetDefaultNodesDir(),
			},
			Nodes: []config.NodeDefinition{},
		}
	} else if projectName != "" {
		// Update project name if provided
		m.config.Workspace.Name = projectName
	}
	
	// Scan for existing repositories if not Force mode
	if !options.Force && m.fsProvider != nil {
		// Check if Walk method exists (it might be a mock with extended functionality)
		if walker, ok := m.fsProvider.(interface {
			Walk(root string, fn filepath.WalkFunc) error
		}); ok {
			err := walker.Walk(m.workspace, func(path string, info os.FileInfo, err error) error {
				// Basic scanning logic - just check for .git directories
				if info != nil && info.IsDir() && info.Name() == ".git" {
					// Found a git repository
					return filepath.SkipDir
				}
				return nil
			})
			if err != nil && !options.Force {
				return fmt.Errorf("scanning repositories: %w", err)
			}
		}
	}
	
	// Save the configuration
	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	
	// Create the nodes directory if it doesn't exist
	nodesDir := m.getNodesDir()
	if nodesDir == "" {
		nodesDir = constants.DefaultReposDir
	}
	nodesDirPath := filepath.Join(m.workspace, nodesDir)
	if err := m.fsProvider.MkdirAll(nodesDirPath, 0755); err != nil {
		return fmt.Errorf("creating nodes directory: %w", err)
	}
	
	// Add MUNO-specific entries to .gitignore if this is a git repository
	gitignoreEntries := []string{
		".muno/",                       // Agent context directory
		nodesDir + "/",                 // Nodes directory (repositories and config nodes)
	}
	
	for _, entry := range gitignoreEntries {
		if err := m.ensureGitignoreEntry(m.workspace, entry); err != nil {
			// Log but don't fail - .gitignore update is optional
			m.logProvider.Debug(fmt.Sprintf("Could not add '%s' to .gitignore: %v", entry, err))
		}
	}
	
	m.logProvider.Info(fmt.Sprintf("Initialized workspace: %s", projectName))
	
	return nil
}

