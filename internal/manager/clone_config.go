package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
)

// processConfigNode handles expansion of config reference nodes during clone
func (m *Manager) processConfigNode(node interfaces.NodeInfo, recursive bool, includeLazy bool, toClone *[]interfaces.NodeInfo) error {
	if !node.IsConfig || node.ConfigFile == "" {
		return nil
	}

	// Create directory for config node
	configNodePath := m.computeFilesystemPath(node.Path)
	if err := m.fsProvider.MkdirAll(configNodePath, 0755); err != nil {
		return fmt.Errorf("creating config node directory: %w", err)
	}

	// Resolve config file path
	configFilePath := m.resolveConfigPath(node.ConfigFile, node.Path)
	
	// Load the referenced configuration
	cfg, err := config.LoadTree(configFilePath)
	if err != nil {
		m.logProvider.Warn(fmt.Sprintf("Failed to load config from %s: %v", configFilePath, err))
		return nil // Don't fail entire clone if one config can't be loaded
	}

	// Copy or link the config file into the node directory for reference
	targetConfigPath := filepath.Join(configNodePath, "muno.yaml")
	if err := m.copyConfigFile(configFilePath, targetConfigPath); err != nil {
		m.logProvider.Warn(fmt.Sprintf("Failed to copy config file: %v", err))
	}

	// Process nodes from the loaded configuration
	for _, nodeDef := range cfg.Nodes {
		childPath := node.Path + "/" + nodeDef.Name
		if node.Path == "/" {
			childPath = "/" + nodeDef.Name
		}

		// Determine if this child should be cloned
		isLazy := nodeDef.IsLazy()
		shouldClone := false

		if nodeDef.URL != "" {
			// It's a repository node
			if !isLazy || includeLazy {
				shouldClone = true
			}
		} else if nodeDef.File != "" {
			// It's another config node - only recurse if in recursive mode
			if recursive {
				childNode := interfaces.NodeInfo{
					Name:       nodeDef.Name,
					Path:       childPath,
					ConfigFile: nodeDef.File,
					IsConfig:   true,
					IsLazy:     false,
					Children:   []interfaces.NodeInfo{},
				}
				// Recursively process nested config node
				if err := m.processConfigNode(childNode, recursive, includeLazy, toClone); err != nil {
					m.logProvider.Warn(fmt.Sprintf("Failed to process nested config %s: %v", nodeDef.Name, err))
				}
			}
			continue
		}

		if shouldClone {
			// Check if already cloned
			childFsPath := filepath.Join(configNodePath, nodeDef.Name)
			if _, err := os.Stat(filepath.Join(childFsPath, ".git")); err == nil {
				// Already cloned
				continue
			}

			// Add to clone list
			*toClone = append(*toClone, interfaces.NodeInfo{
				Name:       nodeDef.Name,
				Path:       childPath,
				Repository: nodeDef.URL,
				IsLazy:     isLazy,
				IsCloned:   false,
			})
		}
	}

	return nil
}

// resolveConfigPath resolves a config file path relative to the current node
func (m *Manager) resolveConfigPath(configFile string, nodePath string) string {
	// If absolute path, use as-is
	if filepath.IsAbs(configFile) {
		return configFile
	}

	// Check if it's a URL (remote config)
	if strings.HasPrefix(configFile, "http://") || strings.HasPrefix(configFile, "https://") {
		// TODO: Handle remote config files
		return configFile
	}

	// Relative path - resolve from workspace root or current node directory
	workspaceRoot := m.workspace
	
	// Try relative to workspace root first
	absPath := filepath.Join(workspaceRoot, configFile)
	if _, err := os.Stat(absPath); err == nil {
		return absPath
	}

	// Try relative to parent of workspace (for ../team/muno.yaml style references)
	parentPath := filepath.Join(filepath.Dir(workspaceRoot), configFile)
	if _, err := os.Stat(parentPath); err == nil {
		return parentPath
	}

	// Default to workspace-relative path
	return absPath
}

// copyConfigFile copies a configuration file to the target location
func (m *Manager) copyConfigFile(src, dst string) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	// Write to destination
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// expandConfigNodes processes all config nodes in the tree and returns repositories to clone
func (m *Manager) expandConfigNodes(node interfaces.NodeInfo, recursive bool, includeLazy bool, toClone *[]interfaces.NodeInfo) error {
	// Process this node if it's a config node
	if node.IsConfig {
		if err := m.processConfigNode(node, recursive, includeLazy, toClone); err != nil {
			return err
		}
		// Config nodes are always processed recursively to find repos
		return nil
	}

	// If it's a repository node, add it to clone list if appropriate
	if node.Repository != "" && !node.IsCloned {
		if !node.IsLazy || includeLazy {
			*toClone = append(*toClone, node)
		}
		// Don't recurse into repository children (repos don't have child repos)
		return nil
	}

	// Recurse into children if in recursive mode
	if recursive {
		for _, child := range node.Children {
			if err := m.expandConfigNodes(child, recursive, includeLazy, toClone); err != nil {
				m.logProvider.Warn(fmt.Sprintf("Error processing child %s: %v", child.Name, err))
				// Continue with other children even if one fails
			}
		}
	}

	return nil
}