package tree

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
	"gopkg.in/yaml.v3"
)

// Manager manages tree without persistent state file
// This is the stateless implementation that derives state from filesystem
type Manager struct {
	workspacePath string
	config        *config.ConfigTree
	resolver      *ConfigResolver
	gitCmd        git.Interface
	currentPath   string  // Current logical path (session only)
}

// NewManager creates a manager that derives state from filesystem
func NewManager(workspacePath string, gitCmd git.Interface) (*Manager, error) {
	// Load config from workspace - try various config file names
	var cfg *config.ConfigTree
	var err error
	
	for _, configName := range config.GetConfigFileNames() {
		configPath := filepath.Join(workspacePath, configName)
		cfg, err = config.LoadTree(configPath)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) && !strings.Contains(err.Error(), "no such file or directory") {
			return nil, fmt.Errorf("loading config: %w", err)
		}
	}
	if err != nil {
		// If no config, create default
		if os.IsNotExist(err) {
			cfg = config.DefaultConfigTree("workspace")
		} else {
			return nil, fmt.Errorf("loading config: %w", err)
		}
	}
	
	resolver := NewConfigResolver(workspacePath)
	
	return &Manager{
		workspacePath: workspacePath,
		config:        cfg,
		resolver:      resolver,
		gitCmd:        gitCmd,
		currentPath:   "/",
	}, nil
}

// ComputeFilesystemPath converts logical path to filesystem path
func (m *Manager) ComputeFilesystemPath(logicalPath string) string {
	reposDir := m.config.GetReposDir()
	
	// For root, always use repos directory
	if logicalPath == "/" || logicalPath == "" {
		return filepath.Join(m.workspacePath, reposDir)
	}
	
	// Try to get the node to determine its type
	node, err := m.GetNodeByPath(logicalPath)
	if err == nil && node != nil {
		// Check if this is a git repository node (not a config node)
		// A node is a git repository if it has a URL (not a config path)
		if node.URL != "" {
			// For git repository nodes, place them in the repos directory
			parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")
			
			// If it's a top-level repo, put it in the repos directory
			if len(parts) == 1 {
				return filepath.Join(m.workspacePath, reposDir, parts[0])
			}
			
			// For nested repos, we need to compute the parent path
			// and check if parent has muno.yaml with different repos_dir
			parentPath := "/" + strings.Join(parts[:len(parts)-1], "/")
			parentFsPath := m.ComputeFilesystemPath(parentPath)
			
			// Check if parent has muno.yaml with custom repos_dir
			parentMunoYaml := filepath.Join(parentFsPath, "muno.yaml")
			if _, err := os.Stat(parentMunoYaml); err == nil {
				// Parent has muno.yaml, check its repos_dir
				childReposDir := ".nodes" // default
				if cfg, loadErr := config.LoadTree(parentMunoYaml); loadErr == nil && cfg != nil && cfg.Workspace.ReposDir != "" {
					childReposDir = cfg.Workspace.ReposDir
				}
				return filepath.Join(parentFsPath, childReposDir, parts[len(parts)-1])
			}
			
			return filepath.Join(parentFsPath, parts[len(parts)-1])
		}
	}
	
	// For config nodes and intermediate directories, check if they have nested structure
	// Split path: /level1/level2 -> [level1, level2]
	parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")
	
	// Build the path considering each parent might have its own repos_dir
	currentPath := filepath.Join(m.workspacePath, reposDir)
	
	for i, part := range parts {
		if i == 0 {
			// First level goes directly under workspace repos dir
			currentPath = filepath.Join(currentPath, part)
		} else {
			// For nested levels, check if parent has muno.yaml
			parentMunoYaml := filepath.Join(currentPath, "muno.yaml")
			if _, err := os.Stat(parentMunoYaml); err == nil {
				// Parent has muno.yaml, use its repos_dir
				childReposDir := ".nodes" // default
				if cfg, loadErr := config.LoadTree(parentMunoYaml); loadErr == nil && cfg != nil && cfg.Workspace.ReposDir != "" {
					childReposDir = cfg.Workspace.ReposDir
				}
				currentPath = filepath.Join(currentPath, childReposDir, part)
			} else {
				// No muno.yaml, continue directly
				currentPath = filepath.Join(currentPath, part)
			}
		}
	}
	
	return currentPath
}

// GetNodeByPath finds a node by its logical path
func (m *Manager) GetNodeByPath(logicalPath string) (*config.NodeDefinition, error) {
	// Treat empty or only-slashes paths as root
	trimmed := strings.Trim(logicalPath, "/")
	if trimmed == "" {
		return nil, nil // Root node
	}
	
	// Split and ignore empty parts (e.g., "///")
	rawParts := strings.Split(trimmed, "/")
	parts := make([]string, 0, len(rawParts))
	for _, p := range rawParts {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		return nil, nil
	}
	
	// Find in top-level nodes
	for _, node := range m.config.Nodes {
		if node.Name == parts[0] {
			// For now, return the matched top-level node regardless of deeper parts
			return &node, nil
		}
	}
	
	return nil, fmt.Errorf("node not found: %s", logicalPath)
}



// AddRepo adds a new repository to config
func (m *Manager) AddRepo(parentPath, name, url string, lazy bool) error {
	// Check for duplicates at the same level
	targetPath := path.Join(parentPath, name)
	
	// Check if already exists in config (for top-level repos)
	if parentPath == "/" {
		for _, existing := range m.config.Nodes {
			if existing.Name == name {
				return fmt.Errorf("repository '%s' already exists", name)
			}
		}
	} else {
		// For nested repos, check filesystem
		fsPath := m.ComputeFilesystemPath(targetPath)
		if _, err := os.Stat(fsPath); err == nil {
			return fmt.Errorf("repository '%s' already exists at %s", name, parentPath)
		}
	}
	
	// For stateless operation, we add to config and save
	fetchMode := config.FetchLazy
	if !lazy {
		fetchMode = config.FetchEager
	}
	node := config.NodeDefinition{
		Name:  name,
		URL:   url,
		Fetch: fetchMode,
	}
	
	// Add to nodes (only for top-level)
	if parentPath == "/" {
		m.config.Nodes = append(m.config.Nodes, node)
	}
	
	// Save config (only for top-level repos)
	if parentPath == "/" {
		configPath := filepath.Join(m.workspacePath, "muno.yaml")
		if err := m.config.Save(configPath); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
	}
	
	// Clone if not lazy
	if !lazy {
		targetPath := path.Join(parentPath, name)
		fsPath := m.ComputeFilesystemPath(targetPath)
		fmt.Printf("Cloning %s to %s\n", url, fsPath)
		if err := m.cloneWithSSHPreference(url, fsPath); err != nil {
			return fmt.Errorf("cloning: %w", err)
		}
	}
	
	return nil
}

// RemoveNode removes a node from config
func (m *Manager) RemoveNode(targetPath string) error {
	// Handle relative paths
	if !strings.HasPrefix(targetPath, "/") {
		targetPath = path.Join(m.currentPath, targetPath)
	}
	
	if targetPath == "/" {
		return fmt.Errorf("cannot remove root")
	}
	
	// Parse the path
	parts := strings.Split(strings.TrimPrefix(targetPath, "/"), "/")
	if len(parts) == 0 {
		return fmt.Errorf("invalid path")
	}
	
	// Track if node was found in config
	found := false
	
	// If it's a top-level node, remove from config
	if len(parts) == 1 {
		targetName := parts[0]
		
		// Remove from config
		newNodes := []config.NodeDefinition{}
		for _, node := range m.config.Nodes {
			if node.Name != targetName {
				newNodes = append(newNodes, node)
			} else {
				found = true
			}
		}
		
		if found {
			m.config.Nodes = newNodes
			
			// Save config
			configPath := filepath.Join(m.workspacePath, "muno.yaml")
			if err := m.config.Save(configPath); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
		}
	}
	
	// Remove from filesystem (works for both top-level and nested nodes)
	fsPath := m.ComputeFilesystemPath(targetPath)
	exists := false
	if _, err := os.Stat(fsPath); err == nil {
		exists = true
		if err := os.RemoveAll(fsPath); err != nil {
			return fmt.Errorf("failed to remove %s: %v", fsPath, err)
		}
	}
	
	// If not found in config and not on filesystem, return error
	if !found && !exists {
		return fmt.Errorf("node not found: %s", targetPath)
	}
	
	// If we're at this path or below it, navigate to parent or root
	if strings.HasPrefix(m.currentPath, targetPath) {
		// Navigate to parent of removed node
		parentPath := filepath.Dir(targetPath)
		if parentPath == "." || parentPath == targetPath {
			parentPath = "/"
		}
		m.currentPath = parentPath
	}
	
	return nil
}

// GetCurrentPath returns current logical path
func (m *Manager) GetCurrentPath() string {
	return m.currentPath
}

// GetNode returns a TreeNode for the given logical path
// This builds the node dynamically from config and filesystem state
func (m *Manager) GetNode(logicalPath string) *TreeNode {
	// Normalize path
	if logicalPath == "" {
		logicalPath = "/"
	}
	
	
	// For root node
	if logicalPath == "/" || logicalPath == "" {
		rootNode := &TreeNode{
			Name:     "root",
			Type:     NodeTypeRoot,
			Children: []string{},
		}
		
		// Add children from config
		for _, node := range m.config.Nodes {
			rootNode.Children = append(rootNode.Children, node.Name)
		}
		
		return rootNode
	}
	
	// Parse the path to get node name and check filesystem
	parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")
	if len(parts) == 0 {
		return nil
	}
	
	nodeName := parts[len(parts)-1]
	
	// Check if the node exists in the filesystem
	fsPath := m.ComputeFilesystemPath(logicalPath)
	fileExists := true
	if _, err := os.Stat(fsPath); os.IsNotExist(err) {
		fileExists = false
	}
	
	// Try to find it in config for URL and lazy status
	nodeDef, err := m.GetNodeByPath(logicalPath)
	
	// For nested paths, check if parent has a muno.yaml first
	if len(parts) > 1 {
		// This is a nested node - try to find its definition
		// First, check if parent has a muno.yaml
		parentPath := "/" + strings.Join(parts[:len(parts)-1], "/")
		parentFsPath := m.ComputeFilesystemPath(parentPath)
		munoYamlPath := filepath.Join(parentFsPath, "muno.yaml")
		
		if _, statErr := os.Stat(munoYamlPath); statErr == nil {
			// Parent has muno.yaml, load it to find child definition
			if cfg, loadErr := config.LoadTree(munoYamlPath); loadErr == nil && cfg != nil {
				for _, child := range cfg.Nodes {
					if child.Name == nodeName {
						nodeDef = &child
						err = nil
						break
					}
				}
			}
		} else if nodeDef != nil && nodeDef.Name != nodeName {
			// GetNodeByPath returns the top-level parent
			// Check if parent is a config reference
			parentDef := nodeDef
			nodeDef = nil // Reset nodeDef for the actual node
			
			// If parent is a config reference, try to find the child in that config
			if parentDef.File != "" {
				refConfig, loadErr := m.loadExternalConfig(parentDef.File)
				if loadErr == nil && refConfig != nil {
					// Look for the child node in the referenced config
					for i := range refConfig.Nodes {
						if refConfig.Nodes[i].Name == nodeName {
							// Found the child node definition
							nodeDef = &refConfig.Nodes[i]
							err = nil
							break
						}
					}
				}
			}
			
			if nodeDef == nil {
				err = fmt.Errorf("node not found")
			}
		}
	}
	
	if err != nil && !fileExists {
		// Node not in config and not on filesystem
		return nil
	}
	
	// Build the TreeNode
	node := &TreeNode{
		Name:     nodeName,
		Type:     NodeTypeRepo,
		Children: []string{},
	}
	
	if nodeDef != nil {
		// Check if this is a config reference node
		if nodeDef.File != "" {
			node.Type = NodeTypeFile
			node.FilePath = nodeDef.File
		} else {
			node.URL = nodeDef.URL
			node.Lazy = nodeDef.Fetch == "lazy"
		}
	}
	
	// Check filesystem state
	if fileExists {
		if _, err := os.Stat(filepath.Join(fsPath, ".git")); err == nil {
			node.State = RepoStateCloned
			// Check if modified
			if m.gitCmd != nil {
				if status, err := m.gitCmd.Status(fsPath); err == nil && strings.Contains(status, "modified") {
					node.State = RepoStateModified
				}
			}
		} else {
			node.State = RepoStateCloned  // Directory exists but no git
		}
	} else {
		// Directory doesn't exist - it's a lazy repo that hasn't been cloned
		node.State = RepoStateMissing
	}
	
	// Build children list from both config and filesystem
	childMap := make(map[string]bool)
	
	// First, add children from config if this is the root node
	if logicalPath == "/" {
		for _, configNode := range m.config.Nodes {
			node.Children = append(node.Children, configNode.Name)
			childMap[configNode.Name] = true
		}
	} else if len(parts) == 1 {
		// For top-level nodes, check if it's a config reference in the main config
		for _, configNode := range m.config.Nodes {
			if configNode.Name == nodeName && configNode.File != "" {
				// This is a config reference node, load its children
				refConfig, err := m.loadExternalConfig(configNode.File)
				if err == nil && refConfig != nil {
					for _, childNode := range refConfig.Nodes {
						if !childMap[childNode.Name] {
							node.Children = append(node.Children, childNode.Name)
							childMap[childNode.Name] = true
						}
					}
				}
				break
			}
		}
	}
	
	// For config reference nodes found via GetNodeByPath, also load the referenced config
	if nodeDef != nil && nodeDef.File != "" {
		// Try to load the referenced config file
		refConfig, err := m.loadExternalConfig(nodeDef.File)
		if err == nil && refConfig != nil {
			for _, childNode := range refConfig.Nodes {
				if !childMap[childNode.Name] {
					node.Children = append(node.Children, childNode.Name)
					childMap[childNode.Name] = true
				}
			}
		}
	}
	
	// Then, add children from filesystem (if they're not already in the list)
	if fileExists {
		// Check if this node has a muno.yaml - if so, load children from it
		munoYamlPath := filepath.Join(fsPath, "muno.yaml")
		if _, err := os.Stat(munoYamlPath); err == nil {
			// Load the muno.yaml to get child definitions and repos_dir
			reposDir := ".nodes" // default
			if cfg, err := config.LoadTree(munoYamlPath); err == nil && cfg != nil {
				for _, childNode := range cfg.Nodes {
					if !childMap[childNode.Name] {
						node.Children = append(node.Children, childNode.Name)
						childMap[childNode.Name] = true
					}
				}
				
				// Get the configured repos directory
				if cfg.Workspace.ReposDir != "" {
					reposDir = cfg.Workspace.ReposDir
				}
			}
			
			// Also check configured repos subdirectory for actual cloned repos
			checkPath := filepath.Join(fsPath, reposDir)
			if entries, err := os.ReadDir(checkPath); err == nil {
				for _, entry := range entries {
					if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
						// Check if it's a repo (has .git dir)
						childPath := filepath.Join(checkPath, entry.Name())
						if _, err := os.Stat(filepath.Join(childPath, ".git")); err == nil {
							if !childMap[entry.Name()] {
								node.Children = append(node.Children, entry.Name())
								childMap[entry.Name()] = true
							}
						}
					}
				}
			}
		} else {
			// No muno.yaml, check for child repos directly in this directory
			if entries, err := os.ReadDir(fsPath); err == nil {
				for _, entry := range entries {
					if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
						// Check if it's a repo (has .git dir)
						childPath := filepath.Join(fsPath, entry.Name())
						if _, err := os.Stat(filepath.Join(childPath, ".git")); err == nil {
							if !childMap[entry.Name()] {
								node.Children = append(node.Children, entry.Name())
								childMap[entry.Name()] = true
							}
						}
					}
				}
			}
		}
	}
	
	return node
}

// ListChildren lists children nodes
func (m *Manager) ListChildren(targetPath string) ([]*TreeNode, error) {
	if targetPath == "" {
		targetPath = m.currentPath
	}
	
	// For root or current level, return config nodes
	if targetPath == "/" {
		children := make([]*TreeNode, 0, len(m.config.Nodes))
		for _, node := range m.config.Nodes {
			fsPath := m.ComputeFilesystemPath("/" + node.Name)
			state := GetRepoState(fsPath)
			
			children = append(children, &TreeNode{
				Name:  node.Name,
				Type:  NodeTypeRepo,
				URL:   node.URL,
				Lazy:  node.IsLazy(),
				State: state,
			})
		}
		return children, nil
	}
	
	// For non-root paths, check if the node exists
	node := m.GetNode(targetPath)
	if node == nil {
		return nil, fmt.Errorf("node not found: %s", targetPath)
	}
	
	// For sub-nodes, scan filesystem for children
	children := make([]*TreeNode, 0)
	for _, childName := range node.Children {
		childPath := path.Join(targetPath, childName)
		if childNode := m.GetNode(childPath); childNode != nil {
			children = append(children, childNode)
		}
	}
	return children, nil
}

// CloneLazyRepos clones lazy repositories
func (m *Manager) CloneLazyRepos(targetPath string, recursive bool) error {
	for _, node := range m.config.Nodes {
		if node.URL != "" {
			nodePath := "/" + node.Name
			fsPath := m.ComputeFilesystemPath(nodePath)
			
			// Check if needs cloning
			if _, err := os.Stat(filepath.Join(fsPath, ".git")); os.IsNotExist(err) {
				fmt.Printf("Cloning %s to %s\n", node.URL, fsPath)
				if err := m.cloneWithSSHPreference(node.URL, fsPath); err != nil {
					return fmt.Errorf("cloning %s: %w", node.Name, err)
				}
			}
			
			// If recursive and node has config, process sub-nodes
			if recursive && node.File != "" {
				// Would load sub-config and process
			}
		}
	}
	
	return nil
}

// cloneWithSSHPreference clones a repository using SSH preference from config
func (m *Manager) cloneWithSSHPreference(url, path string) error {
	// Get SSH preference from config defaults
	sshPreference := m.config.Defaults.SSHPreference
	
	// Use the Git interface with SSH preference support
	if gitWithSSH, ok := m.gitCmd.(*git.Git); ok {
		return gitWithSSH.CloneWithSSHPreference(url, path, sshPreference)
	}
	
	// Fallback to regular clone if not the standard Git implementation
	return m.gitCmd.Clone(url, path)
}

// The display methods are implemented in display.go
// These are alternate implementations kept for reference
/*
// DisplayTree shows tree structure
func (m *Manager) DisplayTree() string {
	var output strings.Builder
	output.WriteString(fmt.Sprintf("🌳 %s\n", m.config.Workspace.Name))
	
	for i, node := range m.config.Nodes {
		prefix := "├─"
		if i == len(m.config.Nodes)-1 {
			prefix = "└─"
		}
		
		icon := "📦"
		fsPath := m.ComputeFilesystemPath("/" + node.Name)
		state := GetRepoState(fsPath)
		
		if state == RepoStateMissing {
			icon = "💤"
		} else if state == RepoStateModified {
			icon = "📝"
		}
		
		if node.File != "" {
			icon = "📁" // Config reference
		}
		
		status := ""
		if node.IsLazy() && state == RepoStateMissing {
			status = " (lazy)"
		}
		
		output.WriteString(fmt.Sprintf("%s %s %s%s\n", prefix, icon, node.Name, status))
	}
	
	return output.String()
}

// DisplayStatus shows current status
func (m *Manager) DisplayStatus() string {
	var output strings.Builder
	output.WriteString("=== Tree Status ===\n")
	output.WriteString(fmt.Sprintf("Current Path: %s\n", m.currentPath))
	output.WriteString(fmt.Sprintf("Workspace: %s\n", m.config.Workspace.Name))
	output.WriteString(fmt.Sprintf("Nodes: %d\n", len(m.config.Nodes)))
	
	// Count states
	var cloned, missing, modified int
	for _, node := range m.config.Nodes {
		if node.URL != "" {
			fsPath := m.ComputeFilesystemPath("/" + node.Name)
			state := GetRepoState(fsPath)
			switch state {
			case RepoStateCloned:
				cloned++
			case RepoStateMissing:
				missing++
			case RepoStateModified:
				modified++
			}
		}
	}
	
	output.WriteString(fmt.Sprintf("Repositories: %d cloned, %d missing, %d modified\n", 
		cloned, missing, modified))
	
	return output.String()
}

// DisplayPath shows current path
func (m *Manager) DisplayPath() string {
	return m.currentPath
}

// DisplayChildren shows children of current node
func (m *Manager) DisplayChildren() string {
	children, _ := m.ListChildren(m.currentPath)
	if len(children) == 0 {
		return "No children\n"
	}
	
	var output strings.Builder
	output.WriteString("Children:\n")
	for _, child := range children {
		status := ""
		if child.State == RepoStateMissing {
			status = " (lazy)"
		} else if child.State == RepoStateModified {
			status = " (modified)"
		}
		output.WriteString(fmt.Sprintf("  - %s%s\n", child.Name, status))
	}
	
	return output.String()
}

*/

// GetState returns a dynamically generated TreeState for compatibility
// This is a stateless manager, so state is computed on demand
func (m *Manager) GetState() *TreeState {
	// Build a minimal state for compatibility
	state := &TreeState{
		CurrentPath: m.currentPath,
		Nodes:       make(map[string]*TreeNode),
	}
	
	// Add root node
	state.Nodes["/"] = m.GetNode("/")
	
	// Add config nodes
	for _, node := range m.config.Nodes {
		nodePath := "/" + node.Name
		if treeNode := m.GetNode(nodePath); treeNode != nil {
			state.Nodes[nodePath] = treeNode
		}
	}
	
	return state
}

// loadExternalConfig loads an external config file
func (m *Manager) loadExternalConfig(filePath string) (*config.ConfigTree, error) {
	// Check if it's a remote URL
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		// For now, we don't support remote configs in tree display
		// This could be implemented in the future
		return nil, fmt.Errorf("remote config not supported")
	}
	
	// Load the local config file
	configPath := filePath
	if !filepath.IsAbs(filePath) {
		// Make it relative to the workspace path
		configPath = filepath.Join(m.workspacePath, filePath)
	}
	
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}
	
	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	
	// Parse the YAML
	var extConfig config.ConfigTree
	if err := yaml.Unmarshal(data, &extConfig); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	
	return &extConfig, nil
}

// SaveState does nothing as this is a stateless manager
// This method exists for compatibility but does nothing
func (m *Manager) SaveState() error {
	// Stateless manager doesn't save state
	return nil
}