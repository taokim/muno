package tree

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
)

// Manager is the refactored tree manager with simplified state
type Manager struct {
	workspacePath string
	config        *config.ConfigTree
	state         *TreeState
	gitCmd        git.Interface
}

// NewManager creates a new tree manager
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
		// Check if it's a file not found error (might be wrapped)
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file or directory") {
			cfg = config.DefaultConfigTree("workspace")
		} else {
			return nil, fmt.Errorf("loading config: %w", err)
		}
	}
	
	m := &Manager{
		workspacePath: workspacePath,
		config:        cfg,
		gitCmd:        gitCmd,
	}
	
	// Try to load state, but if it doesn't exist, create a default one
	if err := m.loadState(); err != nil {
		// If file doesn't exist, create default state
		if os.IsNotExist(err) {
			m.state = &TreeState{
				Nodes: map[string]*TreeNode{
					"/": {
						Name:     "root",
						Type:     NodeTypeRoot,
						Children: []string{},
					},
				},
				CurrentPath: "/",
				LastUpdated: time.Now(),
			}
		} else {
			return nil, fmt.Errorf("failed to load state: %w", err)
		}
	}
	
	// Build tree from config
	if err := m.buildTreeFromConfig(); err != nil {
		return nil, fmt.Errorf("failed to build tree from config: %w", err)
	}
	
	return m, nil
}

// ComputeFilesystemPath derives filesystem path from logical path
// This is the ONLY place that knows about the repos directory pattern
func (m *Manager) ComputeFilesystemPath(logicalPath string) string {
	reposDir := m.config.GetReposDir()
	
	if logicalPath == "/" {
		return filepath.Join(m.workspacePath, reposDir)
	}
	
	// Split path: /level1/level2/level3 -> [level1, level2, level3]
	parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")
	
	// Build filesystem path with repos subdirectories
	// workspace/[reposDir]/level1/[reposDir]/level2/[reposDir]/level3
	fsPath := filepath.Join(m.workspacePath, reposDir)
	for i, part := range parts {
		fsPath = filepath.Join(fsPath, part)
		// Add repos dir before next level (except last)
		if i < len(parts)-1 {
			fsPath = filepath.Join(fsPath, reposDir)
		}
	}
	
	return fsPath
}

// UseNode navigates to a node in the tree
func (m *Manager) UseNode(logicalPath string) error {
	// Normalize path
	if logicalPath == "" {
		logicalPath = "/"
	}
	if !strings.HasPrefix(logicalPath, "/") {
		// Relative path - append to current
		logicalPath = path.Join(m.state.CurrentPath, logicalPath)
	}
	
	node := m.state.Nodes[logicalPath]
	if node == nil {
		return fmt.Errorf("node not found: %s", logicalPath)
	}
	
	fsPath := m.ComputeFilesystemPath(logicalPath)
	
	// Auto-clone if lazy
	if node.Type == NodeTypeRepo && node.State == RepoStateMissing {
		fmt.Printf("Auto-cloning lazy repository: %s\n", node.Name)
		if err := m.cloneToPath(node.URL, fsPath); err != nil {
			return fmt.Errorf("failed to clone %s: %w", node.Name, err)
		}
		node.State = RepoStateCloned
		m.saveState()
	}
	
	// Ensure directory exists (for root or intermediate nodes)
	if err := os.MkdirAll(fsPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	if err := os.Chdir(fsPath); err != nil {
		return fmt.Errorf("cannot navigate to %s: %w", logicalPath, err)
	}
	
	m.state.CurrentPath = logicalPath
	return m.saveState()
}

// AddRepo adds a repository as a child of the current or specified parent
func (m *Manager) AddRepo(parentPath, name, url string, lazy bool) error {
	// Default to current path if not specified
	if parentPath == "" {
		parentPath = m.state.CurrentPath
	}
	
	// Normalize parent path
	if !strings.HasPrefix(parentPath, "/") {
		parentPath = path.Join(m.state.CurrentPath, parentPath)
	}
	
	parent := m.state.Nodes[parentPath]
	if parent == nil {
		return fmt.Errorf("parent node not found: %s", parentPath)
	}
	
	// Check if child already exists
	for _, childName := range parent.Children {
		if childName == name {
			return fmt.Errorf("child %s already exists in %s", name, parentPath)
		}
	}
	
	childPath := path.Join(parentPath, name)
	
	// Create new node
	m.state.Nodes[childPath] = &TreeNode{
		Name:     name,
		Type:     NodeTypeRepo,
		URL:      url,
		Lazy:     lazy,
		State:    RepoStateMissing,
		Children: []string{},
	}
	
	// Add to parent's children
	parent.Children = append(parent.Children, name)
	
	// Clone if not lazy
	if !lazy {
		fsPath := m.ComputeFilesystemPath(childPath)
		fmt.Printf("Cloning %s to %s\n", url, fsPath)
		if err := m.cloneToPath(url, fsPath); err != nil {
			// Rollback on failure
			delete(m.state.Nodes, childPath)
			parent.Children = parent.Children[:len(parent.Children)-1]
			return fmt.Errorf("failed to clone: %w", err)
		}
		m.state.Nodes[childPath].State = RepoStateCloned
	}
	
	m.state.LastUpdated = time.Now()
	return m.saveState()
}

// RemoveNode removes a node and its subtree
func (m *Manager) RemoveNode(targetPath string) error {
	if targetPath == "/" {
		return fmt.Errorf("cannot remove root node")
	}
	
	// Normalize path
	if !strings.HasPrefix(targetPath, "/") {
		targetPath = path.Join(m.state.CurrentPath, targetPath)
	}
	
	node := m.state.Nodes[targetPath]
	if node == nil {
		return fmt.Errorf("node not found: %s", targetPath)
	}
	
	// Find parent
	parentPath := path.Dir(targetPath)
	parent := m.state.Nodes[parentPath]
	if parent == nil {
		return fmt.Errorf("parent node not found: %s", parentPath)
	}
	
	// Remove from parent's children
	newChildren := []string{}
	for _, child := range parent.Children {
		if child != node.Name {
			newChildren = append(newChildren, child)
		}
	}
	parent.Children = newChildren
	
	// Remove node and all descendants from state
	m.removeNodeRecursive(targetPath)
	
	// Remove from filesystem
	fsPath := m.ComputeFilesystemPath(targetPath)
	if err := os.RemoveAll(fsPath); err != nil {
		fmt.Printf("Warning: failed to remove directory %s: %v\n", fsPath, err)
	}
	
	// If we removed the current node, navigate to parent
	if strings.HasPrefix(m.state.CurrentPath, targetPath) {
		m.UseNode(parentPath)
	}
	
	m.state.LastUpdated = time.Now()
	return m.saveState()
}

// removeNodeRecursive removes a node and all its descendants from the state
func (m *Manager) removeNodeRecursive(logicalPath string) {
	node := m.state.Nodes[logicalPath]
	if node == nil {
		return
	}
	
	// Remove all children first
	for _, childName := range node.Children {
		childPath := path.Join(logicalPath, childName)
		m.removeNodeRecursive(childPath)
	}
	
	// Remove this node
	delete(m.state.Nodes, logicalPath)
}

// GetCurrentPath returns the current logical path
func (m *Manager) GetCurrentPath() string {
	return m.state.CurrentPath
}

// GetNode returns a node by its logical path
func (m *Manager) GetNode(logicalPath string) *TreeNode {
	return m.state.Nodes[logicalPath]
}

// ListChildren lists the children of the current or specified node
func (m *Manager) ListChildren(targetPath string) ([]*TreeNode, error) {
	if targetPath == "" {
		targetPath = m.state.CurrentPath
	}
	
	if !strings.HasPrefix(targetPath, "/") {
		targetPath = path.Join(m.state.CurrentPath, targetPath)
	}
	
	node := m.state.Nodes[targetPath]
	if node == nil {
		return nil, fmt.Errorf("node not found: %s", targetPath)
	}
	
	children := make([]*TreeNode, 0, len(node.Children))
	for _, childName := range node.Children {
		childPath := path.Join(targetPath, childName)
		if child := m.state.Nodes[childPath]; child != nil {
			children = append(children, child)
		}
	}
	
	return children, nil
}

// CloneLazyRepos clones all lazy repositories in the current or specified node
func (m *Manager) CloneLazyRepos(targetPath string, recursive bool) error {
	if targetPath == "" {
		targetPath = m.state.CurrentPath
	}
	
	if !strings.HasPrefix(targetPath, "/") {
		targetPath = path.Join(m.state.CurrentPath, targetPath)
	}
	
	return m.cloneLazyReposRecursive(targetPath, recursive)
}

func (m *Manager) cloneLazyReposRecursive(logicalPath string, recursive bool) error {
	node := m.state.Nodes[logicalPath]
	if node == nil {
		return fmt.Errorf("node not found: %s", logicalPath)
	}
	
	// Clone if this is a lazy repo
	if node.Type == NodeTypeRepo && node.State == RepoStateMissing {
		fsPath := m.ComputeFilesystemPath(logicalPath)
		fmt.Printf("Cloning %s to %s\n", node.URL, fsPath)
		if err := m.cloneToPath(node.URL, fsPath); err != nil {
			return fmt.Errorf("failed to clone %s: %w", node.Name, err)
		}
		node.State = RepoStateCloned
		m.saveState()
	}
	
	// Clone children - always clone direct children, recursively if requested
	for _, childName := range node.Children {
		childPath := path.Join(logicalPath, childName)
		if recursive {
			// Recursive: clone all descendants
			if err := m.cloneLazyReposRecursive(childPath, true); err != nil {
				return err
			}
		} else {
			// Non-recursive: only clone direct children if they're lazy
			childNode := m.state.Nodes[childPath]
			if childNode != nil && childNode.Type == NodeTypeRepo && childNode.State == RepoStateMissing {
				fsPath := m.ComputeFilesystemPath(childPath)
				fmt.Printf("Cloning %s to %s\n", childNode.URL, fsPath)
				if err := m.cloneToPath(childNode.URL, fsPath); err != nil {
					return fmt.Errorf("failed to clone %s: %w", childNode.Name, err)
				}
				childNode.State = RepoStateCloned
				m.saveState()
			}
		}
	}
	
	return nil
}

// cloneToPath clones a repository to the specified filesystem path
func (m *Manager) cloneToPath(url, fsPath string) error {
	// Create parent directory
	parentDir := filepath.Dir(fsPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	
	// Clone the repository
	return m.gitCmd.Clone(url, fsPath)
}

// saveState persists the tree state to disk
func (m *Manager) saveState() error {
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	
	statePath := filepath.Join(m.workspacePath, config.GetStateFileName())
	return os.WriteFile(statePath, data, 0644)
}

// loadState loads the tree state from disk
func (m *Manager) loadState() error {
	statePath := filepath.Join(m.workspacePath, config.GetStateFileName())
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return the error so NewManager can handle it
			return err
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}
	
	m.state = &TreeState{}
	if err := json.Unmarshal(data, m.state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}
	
	return nil
}

// buildTreeFromConfig builds the tree structure from the configuration
func (m *Manager) buildTreeFromConfig() error {
	if m.config == nil || len(m.config.Nodes) == 0 {
		// No nodes to build
		return nil
	}
	
	// Ensure root exists
	if m.state.Nodes["/"] == nil {
		m.state.Nodes["/"] = &TreeNode{
			Name:     "root",
			Type:     NodeTypeRoot,
			Children: []string{},
		}
	}
	
	// Process each node definition from config
	for _, nodeDef := range m.config.Nodes {
		// Check if node already exists in state
		nodePath := "/" + nodeDef.Name
		if existingNode, exists := m.state.Nodes[nodePath]; exists {
			// Update existing node with config info
			if nodeDef.URL != "" {
				existingNode.URL = nodeDef.URL
				existingNode.Type = NodeTypeRepository
			} else if nodeDef.Config != "" {
				existingNode.ConfigPath = nodeDef.Config
				existingNode.Type = NodeTypeConfig
			}
			existingNode.Lazy = nodeDef.Lazy
		} else {
			// Create new node
			newNode := &TreeNode{
				Name:     nodeDef.Name,
				Children: []string{},
				Lazy:     nodeDef.Lazy,
			}
			
			if nodeDef.URL != "" {
				newNode.Type = NodeTypeRepository
				newNode.URL = nodeDef.URL
			} else if nodeDef.Config != "" {
				newNode.Type = NodeTypeConfig
				newNode.ConfigPath = nodeDef.Config
			}
			
			// Add to state first
			m.state.Nodes[nodePath] = newNode
			
			// Add as child of root
			if !contains(m.state.Nodes["/"].Children, nodePath) {
				m.state.Nodes["/"].Children = append(m.state.Nodes["/"].Children, nodePath)
			}
		}
		
		// Create directory for node (lazy or config nodes get empty dirs)
		nodeDir := m.ComputeFilesystemPath(nodePath)
		if err := os.MkdirAll(nodeDir, 0755); err != nil {
			fmt.Printf("Warning: Failed to create directory for %s: %v\n", nodeDef.Name, err)
		}
		
		// Clone eager repositories
		if nodeDef.URL != "" && !nodeDef.Lazy {
			if _, err := os.Stat(filepath.Join(nodeDir, ".git")); os.IsNotExist(err) {
				fmt.Printf("Cloning %s from %s...\n", nodeDef.Name, nodeDef.URL)
				if err := m.cloneToPath(nodeDef.URL, nodeDir); err != nil {
					fmt.Printf("Warning: Failed to clone %s: %v\n", nodeDef.Name, err)
				} else {
					m.state.Nodes[nodePath].Cloned = true
				}
			} else {
				// Repository already exists
				m.state.Nodes[nodePath].Cloned = true
			}
		}
	}
	
	// Second pass: Load config references now that all parent nodes exist
	for _, nodeDef := range m.config.Nodes {
		if nodeDef.Config != "" {
			nodePath := "/" + nodeDef.Name
			if err := m.loadConfigReference(nodePath, nodeDef.Config); err != nil {
				fmt.Printf("Warning: Failed to load config %s: %v\n", nodeDef.Config, err)
			}
		}
	}
	
	// Save the updated state
	return m.saveState()
}

// loadConfigReference loads a config file and adds its nodes as children
func (m *Manager) loadConfigReference(parentPath string, configPath string) error {
	// Resolve config path
	fullConfigPath := configPath
	if !filepath.IsAbs(configPath) {
		fullConfigPath = filepath.Join(m.workspacePath, configPath)
	}
	
	// Load the config file
	refConfig, err := config.LoadTree(fullConfigPath)
	if err != nil {
		return fmt.Errorf("loading config reference %s: %w", configPath, err)
	}
	
	// Ensure parent node exists
	if m.state.Nodes[parentPath] == nil {
		return fmt.Errorf("parent node %s does not exist", parentPath)
	}
	
	// Process nodes from referenced config
	for _, nodeDef := range refConfig.Nodes {
		childPath := parentPath + "/" + nodeDef.Name
		
		// Create child node
		childNode := &TreeNode{
			Name:     nodeDef.Name,
			URL:      nodeDef.URL,
			Lazy:     nodeDef.Lazy,
			Children: []string{},
			Type:     NodeTypeRepository,
		}
		
		// Handle nested config references
		if nodeDef.Config != "" {
			childNode.Type = NodeTypeConfig
			childNode.ConfigPath = nodeDef.Config
			// Recursively load nested config
			if err := m.loadConfigReference(childPath, nodeDef.Config); err != nil {
				fmt.Printf("Warning: Failed to load nested config %s: %v\n", nodeDef.Config, err)
			}
		}
		
		// Add to state
		m.state.Nodes[childPath] = childNode
		
		// Add as child of parent
		if !contains(m.state.Nodes[parentPath].Children, childPath) {
			m.state.Nodes[parentPath].Children = append(m.state.Nodes[parentPath].Children, childPath)
		}
		
		// Create directory for child node
		childDir := m.ComputeFilesystemPath(childPath)
		if err := os.MkdirAll(childDir, 0755); err != nil {
			fmt.Printf("Warning: Failed to create directory for %s: %v\n", nodeDef.Name, err)
		}
		
		// Clone eager child repositories
		if nodeDef.URL != "" && !nodeDef.Lazy {
			if _, err := os.Stat(filepath.Join(childDir, ".git")); os.IsNotExist(err) {
				fmt.Printf("Cloning %s from %s...\n", nodeDef.Name, nodeDef.URL)
				if err := m.cloneToPath(nodeDef.URL, childDir); err != nil {
					fmt.Printf("Warning: Failed to clone %s: %v\n", nodeDef.Name, err)
				} else {
					childNode.Cloned = true
				}
			} else {
				// Repository already exists
				childNode.Cloned = true
			}
		}
	}
	
	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetState returns the current tree state
func (m *Manager) GetState() *TreeState {
	return m.state
}

// SaveState saves the current tree state to disk
func (m *Manager) SaveState() error {
	return m.saveState()
}