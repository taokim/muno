package manager

import (
	"fmt"
	"path/filepath"
)

// SmartInitWorkspace performs intelligent initialization for the workspace
func (m *Manager) SmartInitWorkspace(projectName string, options InitOptions) error {
	if !m.initialized {
		m.initialized = true
	}
	
	// Update project name if provided
	if projectName != "" && m.config != nil {
		m.config.Workspace.Name = projectName
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
	// For now, just return nil
	// The actual implementation would depend on how the tree provider manages current position
	return nil
}