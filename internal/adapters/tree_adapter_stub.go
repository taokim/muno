package adapters

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
)

// TreeAdapterStub is a basic implementation of TreeProvider for testing
// It provides in-memory tree operations without filesystem interaction
type TreeAdapterStub struct {
	current    interfaces.NodeInfo
	nodes      map[string]interfaces.NodeInfo
	state      interfaces.TreeState
	workspace  string
	fsProvider interfaces.FileSystemProvider
	scannedPaths map[string]bool // Tracks which paths have been scanned for nested configs
}

// NewTreeAdapter creates a new tree adapter
func NewTreeAdapter() interfaces.TreeProvider {
	return &TreeAdapterStub{
		nodes: make(map[string]interfaces.NodeInfo),
		state: interfaces.TreeState{
			Nodes: make(map[string]interfaces.NodeInfo),
		},
		scannedPaths: make(map[string]bool),
	}
}

// SetWorkspaceContext sets the workspace and filesystem provider for nested config loading
func (t *TreeAdapterStub) SetWorkspaceContext(workspace string, fsProvider interfaces.FileSystemProvider) {
	t.workspace = workspace
	t.fsProvider = fsProvider
}

// DebugNodes returns all nodes in the tree for debugging purposes
func (t *TreeAdapterStub) DebugNodes() map[string]interfaces.NodeInfo {
	return t.nodes
}

// Load loads the tree from configuration
func (t *TreeAdapterStub) Load(cfg interface{}) error {
	// Parse ConfigTree
	configTree, ok := cfg.(*config.ConfigTree)
	if !ok {
		return fmt.Errorf("invalid config type")
	}

	// Create root node
	root := interfaces.NodeInfo{
		Name:     configTree.Workspace.Name,
		Path:     "/",
		IsCloned: true,
	}
	t.nodes["/"] = root
	t.current = root

	// Add nodes from config
	for _, nodeDef := range configTree.Nodes {
		isLazy := nodeDef.IsLazy()
		node := interfaces.NodeInfo{
			Name:       nodeDef.Name,
			Path:       "/" + nodeDef.Name,
			Repository: nodeDef.URL,
			ConfigFile: nodeDef.File,
			IsConfig:   nodeDef.File != "",
			IsLazy:     isLazy,
			IsCloned:   !isLazy,
		}
		t.nodes[node.Path] = node
		
		// If workspace and fsProvider are set, recursively load nested configs
		if t.workspace != "" && t.fsProvider != nil {
			t.loadNestedConfigs(node.Path, configTree.Workspace.ReposDir)
		}
	}

	return nil
}

// loadNestedConfigs recursively loads configs from nested directories
func (t *TreeAdapterStub) loadNestedConfigs(parentPath string, defaultReposDir string) {
	// Check if we've already scanned this path
	if t.scannedPaths[parentPath] {
		return
	}

	if defaultReposDir == "" {
		defaultReposDir = ".nodes"
	}

	// Construct the physical path for this tree node
	// Tree path "/" maps to workspace root
	// Tree path "/team" maps to workspace/.nodes/team
	var physicalPath string
	if parentPath == "/" {
		physicalPath = t.workspace
	} else {
		// Remove leading slash and replace / with /.nodes/
		treeParts := strings.Split(strings.TrimPrefix(parentPath, "/"), "/")
		physicalPath = t.workspace
		for i, part := range treeParts {
			if i > 0 {
				physicalPath = filepath.Join(physicalPath, defaultReposDir)
			} else {
				physicalPath = filepath.Join(physicalPath, defaultReposDir)
			}
			physicalPath = filepath.Join(physicalPath, part)
		}
	}

	// Check if there's a muno.yaml in this directory
	configPath := filepath.Join(physicalPath, "muno.yaml")
	if !t.fsProvider.Exists(configPath) {
		// No config found - don't mark as scanned so we can try again later
		return
	}

	// Mark as scanned only after we confirm config exists
	t.scannedPaths[parentPath] = true
	
	// Load the config
	nestedConfig, err := config.LoadTree(configPath)
	if err != nil {
		return
	}
	
	// Get repos_dir from the nested config
	reposDir := nestedConfig.Workspace.ReposDir
	if reposDir == "" {
		reposDir = ".nodes"
	}
	
	// Add child nodes
	for _, nodeDef := range nestedConfig.Nodes {
		isLazy := nodeDef.IsLazy()
		childPath := parentPath + "/" + nodeDef.Name
		if parentPath == "/" {
			childPath = "/" + nodeDef.Name
		}
		
		node := interfaces.NodeInfo{
			Name:       nodeDef.Name,
			Path:       childPath,
			Repository: nodeDef.URL,
			ConfigFile: nodeDef.File,
			IsConfig:   nodeDef.File != "",
			IsLazy:     isLazy,
			IsCloned:   !isLazy,
		}
		t.nodes[node.Path] = node
		
		// Recursively load configs for this child
		t.loadNestedConfigs(childPath, reposDir)
	}
}

// Navigate navigates to a node in the tree
func (t *TreeAdapterStub) Navigate(path string) error {
	if node, ok := t.nodes[path]; ok {
		t.current = node
		return nil
	}
	return fmt.Errorf("node not found: %s", path)
}

// GetCurrent returns the current node information
func (t *TreeAdapterStub) GetCurrent() (interfaces.NodeInfo, error) {
	return t.current, nil
}

// buildNodeWithChildren recursively builds a NodeInfo with populated Children array
// This matches the behavior of the real treeProviderAdapter
func (t *TreeAdapterStub) buildNodeWithChildren(path string) (interfaces.NodeInfo, error) {
	// Get the base node from the map
	node, ok := t.nodes[path]
	if !ok {
		return interfaces.NodeInfo{}, fmt.Errorf("node not found: %s", path)
	}

	// Create a copy to avoid modifying the stored node
	result := node
	result.Children = []interfaces.NodeInfo{}

	// Find and recursively build all direct children
	// A direct child has a path like "parent/child" where parent is the current path
	pathPrefix := path
	if path != "/" {
		pathPrefix = path + "/"
	}

	for childPath := range t.nodes {
		// Skip the node itself
		if childPath == path {
			continue
		}

		// Check if this is a direct child (not a grandchild)
		if strings.HasPrefix(childPath, pathPrefix) {
			// Get the relative path after the parent
			relativePath := strings.TrimPrefix(childPath, pathPrefix)

			// It's a direct child only if there are no more "/" in the relative path
			if !strings.Contains(relativePath, "/") {
				// Recursively build the child with its children
				childInfo, err := t.buildNodeWithChildren(childPath)
				if err == nil {
					result.Children = append(result.Children, childInfo)
				}
			}
		}
	}

	return result, nil
}

// GetTree returns the root node of the tree with Children populated recursively
func (t *TreeAdapterStub) GetTree() (interfaces.NodeInfo, error) {
	if _, ok := t.nodes["/"]; ok {
		return t.buildNodeWithChildren("/")
	}
	return interfaces.NodeInfo{Name: "root", Path: "/"}, nil
}

// GetNode gets a specific node by path with Children populated recursively
func (t *TreeAdapterStub) GetNode(path string) (interfaces.NodeInfo, error) {
	// If we have workspace context, ensure all ancestors are scanned for nested configs
	if t.workspace != "" && t.fsProvider != nil && path != "/" {
		parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

		// Scan each ancestor level if not already scanned
		for i := 1; i <= len(parts); i++ {
			ancestorPath := "/" + strings.Join(parts[:i], "/")
			if i == len(parts) {
				// For the requested path itself, try its parent
				if i > 1 {
					parentPath := "/" + strings.Join(parts[:i-1], "/")
					if !t.scannedPaths[parentPath] {
						t.loadNestedConfigs(parentPath, ".nodes")
					}
				}
			} else {
				// For ancestors, scan them
				if !t.scannedPaths[ancestorPath] && t.nodes[ancestorPath].Path != "" {
					t.loadNestedConfigs(ancestorPath, ".nodes")
				}
			}
		}
	}

	// Build node with children populated recursively
	return t.buildNodeWithChildren(path)
}

// AddNode adds a new node to the tree
func (t *TreeAdapterStub) AddNode(parentPath string, node interfaces.NodeInfo) error {
	// Special case: if the node is the root node itself (Path="/"), store it at "/"
	// This handles initializeRootNode scenarios where we add the root directly
	if node.Path == "/" {
		t.nodes["/"] = node
		return nil
	}

	// Regular node: compute full path from parent and name
	fullPath := parentPath + "/" + node.Name
	if parentPath == "/" || parentPath == "" {
		fullPath = "/" + node.Name
	}
	node.Path = fullPath
	t.nodes[fullPath] = node
	return nil
}

// RemoveNode removes a node from the tree
func (t *TreeAdapterStub) RemoveNode(path string) error {
	delete(t.nodes, path)
	return nil
}

// UpdateNode updates a node's information
func (t *TreeAdapterStub) UpdateNode(path string, info interfaces.NodeInfo) error {
	t.nodes[path] = info
	return nil
}

// ListChildren lists children of a node
func (t *TreeAdapterStub) ListChildren(path string) ([]interfaces.NodeInfo, error) {
	var children []interfaces.NodeInfo
	for nodePath, node := range t.nodes {
		// Simple check for direct children
		if len(nodePath) > len(path) && nodePath[:len(path)] == path {
			children = append(children, node)
		}
	}
	return children, nil
}

// GetPath returns the current path
func (t *TreeAdapterStub) GetPath() string {
	return t.current.Path
}

// SetPath sets the current path
func (t *TreeAdapterStub) SetPath(path string) error {
	return t.Navigate(path)
}

// GetState returns the tree state
func (t *TreeAdapterStub) GetState() (interfaces.TreeState, error) {
	return t.state, nil
}

// SetState sets the tree state
func (t *TreeAdapterStub) SetState(state interfaces.TreeState) error {
	t.state = state
	return nil
}