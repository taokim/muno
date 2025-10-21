package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
)

// visitNodeForClone is the unified function that visits and processes any node (config or git)
// It follows these steps:
// 1. Create directory for the node if needed
// 2. If config node: symlink the muno.yaml; If git node: clone the repository
// 3. Load muno.yaml from the directory (if exists)
// 4. Recursively visit each child node
func (m *Manager) visitNodeForClone(node interfaces.NodeInfo, recursive bool, includeLazy bool) error {
	nodeFsPath := m.computeFilesystemPath(node.Path)
	
	// Step 1: Create directory if needed
	if err := m.fsProvider.MkdirAll(nodeFsPath, 0755); err != nil {
		return fmt.Errorf("creating node directory: %w", err)
	}

	// Step 2: Process based on node type
	if node.IsConfig && node.ConfigFile != "" {
		// Config node: symlink the external muno.yaml
		configFilePath := node.ConfigFile
		if !filepath.IsAbs(configFilePath) && !strings.HasPrefix(configFilePath, "http") {
			configFilePath = m.resolveConfigPath(node.ConfigFile, node.Path)
		}
		
		targetConfigPath := filepath.Join(nodeFsPath, "muno.yaml")
		if err := os.Symlink(configFilePath, targetConfigPath); err != nil {
			// If symlink fails, try to copy as fallback
			if err := m.copyConfigFile(configFilePath, targetConfigPath); err != nil {
				m.logProvider.Warn(fmt.Sprintf("Failed to link/copy config file: %v", err))
				return nil
			}
		}
	} else if node.Repository != "" && !node.IsCloned {
		// Git node: clone the repository
		if node.IsLazy && !includeLazy {
			// Skip lazy repositories unless includeLazy is set
			return nil
		}
		
		m.logProvider.Info(fmt.Sprintf("Cloning repository %s from %s", node.Name, node.Repository))
		cloneOptions := interfaces.CloneOptions{
			SSHPreference: m.getSSHPreference(),
		}
		if err := m.gitProvider.Clone(node.Repository, nodeFsPath, cloneOptions); err != nil {
			m.logProvider.Warn(fmt.Sprintf("Failed to clone %s: %v", node.Name, err))
			return nil
		}
		
		// Update node status
		node.IsCloned = true
		node.IsLazy = false
		if err := m.treeProvider.UpdateNode(node.Path, node); err != nil {
			m.logProvider.Debug(fmt.Sprintf("Could not update node status for %s: %v", node.Path, err))
		}
	}
	
	// Step 3: Stop here if not recursive
	if !recursive {
		return nil
	}
	
	// Step 4: Load muno.yaml from the directory (if it exists)
	munoYamlPath := filepath.Join(nodeFsPath, "muno.yaml")
	if _, err := os.Stat(munoYamlPath); err != nil {
		// No muno.yaml, nothing more to do
		return nil
	}
	
	cfg, err := config.LoadTree(munoYamlPath)
	if err != nil {
		m.logProvider.Warn(fmt.Sprintf("Failed to load muno.yaml from %s: %v", node.Name, err))
		return nil
	}
	
	// Step 5: Recursively visit each child node defined in muno.yaml
	for _, nodeDef := range cfg.Nodes {
		childPath := node.Path + "/" + nodeDef.Name
		if node.Path == "/" {
			childPath = "/" + nodeDef.Name
		}
		
		// Create child node info
		var childNode interfaces.NodeInfo
		
		if nodeDef.URL != "" {
			// Git repository node
			childNode = interfaces.NodeInfo{
				Name:       nodeDef.Name,
				Path:       childPath,
				Repository: nodeDef.URL,
				IsLazy:     nodeDef.IsLazy(),
				IsCloned:   false,
			}
		} else if nodeDef.File != "" {
			// Config reference node
			childConfigFile := nodeDef.File
			if !filepath.IsAbs(childConfigFile) && !strings.HasPrefix(childConfigFile, "http") {
				// Resolve relative to the real path of the current config
				realConfigPath, err := filepath.EvalSymlinks(munoYamlPath)
				if err != nil {
					realConfigPath = munoYamlPath
				}
				currentConfigDir := filepath.Dir(realConfigPath)
				childConfigFile = filepath.Join(currentConfigDir, nodeDef.File)
			}
			
			childNode = interfaces.NodeInfo{
				Name:       nodeDef.Name,
				Path:       childPath,
				ConfigFile: childConfigFile,
				IsConfig:   true,
				IsLazy:     false,
			}
		} else {
			// Invalid node definition, skip
			continue
		}
		
		// Recursively visit the child node
		if err := m.visitNodeForClone(childNode, recursive, includeLazy); err != nil {
			m.logProvider.Warn(fmt.Sprintf("Failed to process child %s: %v", nodeDef.Name, err))
			// Continue with other children even if one fails
		}
	}
	
	return nil
}

// processConfigNode is now a simple wrapper that calls visitNodeForClone
func (m *Manager) processConfigNode(node interfaces.NodeInfo, recursive bool, includeLazy bool, toClone *[]interfaces.NodeInfo) error {
	// This function is kept for backward compatibility but now just delegates to visitNodeForClone
	return m.visitNodeForClone(node, recursive, includeLazy)
}

// cloneConfigNodeRecursive is now deprecated in favor of visitNodeForClone
// but kept for backward compatibility
func (m *Manager) cloneConfigNodeRecursive(dirPath string, nodePath string, includeLazy bool, toClone *[]interfaces.NodeInfo) error {
	// Load the muno.yaml from the directory
	munoYamlPath := filepath.Join(dirPath, "muno.yaml")
	cfg, err := config.LoadTree(munoYamlPath)
	if err != nil {
		m.logProvider.Warn(fmt.Sprintf("Failed to load config from %s: %v", munoYamlPath, err))
		return nil
	}

	// Process each child node defined in the config
	for _, nodeDef := range cfg.Nodes {
		childPath := nodePath + "/" + nodeDef.Name
		if nodePath == "/" {
			childPath = "/" + nodeDef.Name
		}

		var childNode interfaces.NodeInfo
		
		if nodeDef.URL != "" {
			childNode = interfaces.NodeInfo{
				Name:       nodeDef.Name,
				Path:       childPath,
				Repository: nodeDef.URL,
				IsLazy:     nodeDef.IsLazy(),
				IsCloned:   false,
			}
		} else if nodeDef.File != "" {
			childConfigFile := nodeDef.File
			if !filepath.IsAbs(childConfigFile) && !strings.HasPrefix(childConfigFile, "http") {
				realConfigPath, err := filepath.EvalSymlinks(munoYamlPath)
				if err != nil {
					realConfigPath = munoYamlPath
				}
				currentConfigDir := filepath.Dir(realConfigPath)
				childConfigFile = filepath.Join(currentConfigDir, nodeDef.File)
			}
			
			childNode = interfaces.NodeInfo{
				Name:       nodeDef.Name,
				Path:       childPath,
				ConfigFile: childConfigFile,
				IsConfig:   true,
				IsLazy:     false,
			}
		} else {
			continue
		}
		
		// Use the unified visit function
		if err := m.visitNodeForClone(childNode, true, includeLazy); err != nil {
			m.logProvider.Warn(fmt.Sprintf("Failed to process child %s: %v", nodeDef.Name, err))
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

	// Relative path - resolve from workspace root or parent directory
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