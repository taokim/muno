package manager

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/taokim/muno/internal/config"
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
				ReposDir: config.GetDefaultReposDir(),
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
	
	// Create the repos directory if it doesn't exist
	reposDir := m.getReposDir()
	if reposDir == "" {
		reposDir = "repos"
	}
	reposDirPath := filepath.Join(m.workspace, reposDir)
	if err := m.fsProvider.MkdirAll(reposDirPath, 0755); err != nil {
		return fmt.Errorf("creating repos directory: %w", err)
	}
	
	m.logProvider.Info(fmt.Sprintf("Initialized workspace: %s", projectName))
	
	return nil
}

// ClearCurrent clears the current position in the tree
func (m *Manager) ClearCurrent() error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	// Reset to root
	return m.treeProvider.SetPath("/")
}