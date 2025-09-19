package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/taokim/muno/internal/config"
)

// ConfigResolver handles distributed configuration resolution
type ConfigResolver struct {
	cache map[string]*config.ConfigTree  // Cache loaded configs
	root  string                          // Root workspace path
}

// NewConfigResolver creates a new config resolver
func NewConfigResolver(rootPath string) *ConfigResolver {
	return &ConfigResolver{
		cache: make(map[string]*config.ConfigTree),
		root:  rootPath,
	}
}

// LoadNodeFile loads the configuration for a node if it has one
func (r *ConfigResolver) LoadNodeFile(basePath string, node *config.NodeDefinition) (*config.ConfigTree, error) {
	if node.File == "" {
		return nil, nil // No config to load
	}
	
	configPath := r.resolveFilePath(basePath, node)
	
	// Check cache first
	if cached, ok := r.cache[configPath]; ok {
		return cached, nil
	}
	
	// Load the config
	cfg, err := config.LoadTree(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config %s: %w", configPath, err)
	}
	
	// Cache it
	r.cache[configPath] = cfg
	return cfg, nil
}

// BuildDistributedTree recursively builds the tree from distributed configs
func (r *ConfigResolver) BuildDistributedTree(cfg *config.ConfigTree, currentPath string) (map[string]*TreeNode, error) {
	nodes := make(map[string]*TreeNode)
	
	// Create root node for this level
	root := &TreeNode{
		Name:     cfg.Workspace.Name,
		Type:     NodeTypeRoot,
		Children: make([]string, 0),
	}
	nodes[""] = root  // Root node at empty path
	
	// Process each node
	for _, nodeDef := range cfg.Nodes {
		node, err := r.buildNode(currentPath, &nodeDef)
		if err != nil {
			return nil, fmt.Errorf("failed to build node %s: %w", nodeDef.Name, err)
		}
		nodes[nodeDef.Name] = node
		root.Children = append(root.Children, nodeDef.Name)
	}
	
	return nodes, nil
}

func (r *ConfigResolver) buildNode(basePath string, nodeDef *config.NodeDefinition) (*TreeNode, error) {
	nodePath := filepath.Join(basePath, nodeDef.Name)
	
	node := &TreeNode{
		Name:     nodeDef.Name,
		URL:      nodeDef.URL,
		Children: make([]string, 0),
	}
	
	// Validate node configuration
	nodeKind := GetNodeKind(nodeDef)
	if nodeKind == NodeKindInvalid {
		return nil, fmt.Errorf("invalid node %s: cannot have both URL and config fields", nodeDef.Name)
	}
	
	switch nodeKind {
	case NodeKindRepo:
		node.Type = NodeTypeRepo
		node.Lazy = GetEffectiveLazy(nodeDef)  // Smart lazy defaults
		node.State = GetRepoState(nodePath)
		
		// If repo is cloned, auto-discover its config
		if node.State != RepoStateMissing {
			if configPath, found := AutoDiscoverConfig(nodePath); found {
				// Load the auto-discovered config
				cfg, err := config.LoadTree(configPath)
				if err != nil {
					// Log warning but don't fail
					fmt.Fprintf(os.Stderr, "Warning: failed to load auto-discovered config in %s: %v\n", nodeDef.Name, err)
				} else if cfg != nil {
					// Cache the discovered config
					r.cache[configPath] = cfg
					
					// Build sub-tree from discovered config
					subNodes, err := r.BuildDistributedTree(cfg, nodePath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to build tree for %s: %v\n", nodeDef.Name, err)
					} else if rootNode, ok := subNodes[""]; ok {
						// Add children names from the sub-tree root
						node.Children = rootNode.Children
					}
				}
			}
		}
		
	case NodeKindFile:
		node.Type = "config"  // Config-only node type
		
		// Load the referenced config
		configPath := r.resolveFilePath(basePath, nodeDef)
		cfg, err := config.LoadTree(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config reference %s: %w", configPath, err)
		}
		
		// Cache it
		r.cache[configPath] = cfg
		
		// Build sub-tree from referenced config
		subNodes, err := r.BuildDistributedTree(cfg, nodePath)
		if err != nil {
			return nil, err
		}
		if rootNode, ok := subNodes[""]; ok {
			node.Children = rootNode.Children
		}
		
		// Create marker for config-only node if needed
		if !GetFileStatus(nodePath) {
			if err := CreateFileMarker(nodePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create marker for %s: %v\n", nodeDef.Name, err)
			}
		}
	}
	
	return node, nil
}

func (r *ConfigResolver) resolveFilePath(basePath string, node *config.NodeDefinition) string {
	if filepath.IsAbs(node.File) {
		return node.File
	}
	
	// Config path is relative to the node's directory
	nodePath := filepath.Join(basePath, node.Name)
	return filepath.Join(nodePath, node.File)
}

// The TreeNode and NodeType types are defined in types.go