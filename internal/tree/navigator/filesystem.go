package navigator

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/taokim/muno/internal/config"

	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/tree"
)

// FilesystemNavigator implements TreeNavigator by reading directly from the filesystem.
// This is the default navigator that ensures accuracy by always checking actual state.
type FilesystemNavigator struct {
	workspace    string
	config       *config.ConfigTree
	resolver     *tree.ConfigResolver
	gitCmd       interfaces.GitInterface
	currentPath  string
	currentFile  string // Path to .muno/current file
}

// NewFilesystemNavigator creates a new filesystem-based navigator
func NewFilesystemNavigator(workspace string, cfg *config.ConfigTree, gitCmd interfaces.GitInterface) (*FilesystemNavigator, error) {
	if workspace == "" {
		return nil, fmt.Errorf("workspace path cannot be empty")
	}

	// Ensure workspace exists
	if _, err := os.Stat(workspace); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace does not exist: %s", workspace)
	}

	if cfg == nil {
		// Try to load config from workspace
		var err error
		for _, configName := range config.GetConfigFileNames() {
			configPath := filepath.Join(workspace, configName)
			cfg, err = config.LoadTree(configPath)
			if err == nil {
				break
			}
		}
		if cfg == nil {
			cfg = config.DefaultConfigTree("workspace")
		}
	}

	resolver := tree.NewConfigResolver(workspace)
	currentFile := filepath.Join(workspace, ".muno", "current")

	nav := &FilesystemNavigator{
		workspace:    workspace,
		config:       cfg,
		resolver:     resolver,
		gitCmd:       gitCmd,
		currentPath:  "/",
		currentFile:  currentFile,
	}

	// Load current path from file
	nav.loadCurrentPath()

	return nav, nil
}

// GetCurrentPath returns the current position in the tree
func (n *FilesystemNavigator) GetCurrentPath() (string, error) {
	return n.currentPath, nil
}

// Navigate changes the current position to the specified path
func (n *FilesystemNavigator) Navigate(path string) error {
	// Normalize path
	targetPath := n.normalizePath(path)

	// Check if node exists
	node, err := n.GetNode(targetPath)
	if err != nil {
		return fmt.Errorf("cannot navigate to %s: %w", targetPath, err)
	}
	if node == nil {
		return fmt.Errorf("node not found: %s", targetPath)
	}

	// Get filesystem path
	fsPath := n.computeFilesystemPath(targetPath)

	// Check if it's a lazy repository that needs cloning
	status, err := n.GetNodeStatus(targetPath)
	if err == nil && status.Lazy && !status.Cloned {
		if err := n.TriggerLazyLoad(targetPath); err != nil {
			return fmt.Errorf("failed to auto-clone lazy repository: %w", err)
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(fsPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Change directory
	if err := os.Chdir(fsPath); err != nil {
		return fmt.Errorf("cannot change directory to %s: %w", fsPath, err)
	}

	// Update current path
	n.currentPath = targetPath
	return n.saveCurrentPath()
}

// GetNode retrieves a single node by its path
func (n *FilesystemNavigator) GetNode(nodePath string) (*Node, error) {
	nodePath = n.normalizePath(nodePath)

	// Root node
	if nodePath == "/" || nodePath == "" {
		return &Node{
			Path:     "/",
			Name:     "root",
			Type:     NodeTypeRoot,
			Children: n.getRootChildren(),
		}, nil
	}

	// Check if this path exists in config or filesystem
	node := n.buildNodeFromPath(nodePath)
	if node == nil {
		return nil, nil
	}

	return node, nil
}

// ListChildren returns all direct children of a node
func (n *FilesystemNavigator) ListChildren(nodePath string) ([]*Node, error) {
	nodePath = n.normalizePath(nodePath)

	parent, err := n.GetNode(nodePath)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, fmt.Errorf("node not found: %s", nodePath)
	}

	children := make([]*Node, 0, len(parent.Children))
	for _, childName := range parent.Children {
		childPath := path.Join(nodePath, childName)
		child, err := n.GetNode(childPath)
		if err != nil {
			continue // Skip problematic children
		}
		if child != nil {
			children = append(children, child)
		}
	}

	return children, nil
}

// GetTree returns a tree view starting from path with specified depth
func (n *FilesystemNavigator) GetTree(startPath string, depth int) (*TreeView, error) {
	startPath = n.normalizePath(startPath)

	root, err := n.GetNode(startPath)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, fmt.Errorf("node not found: %s", startPath)
	}

	view := &TreeView{
		Root:      root,
		Nodes:     make(map[string]*Node),
		Status:    make(map[string]*NodeStatus),
		Depth:     depth,
		Generated: time.Now(),
	}

	// Build tree recursively
	if err := n.buildTreeView(view, root, 0, depth); err != nil {
		return nil, err
	}

	return view, nil
}

// GetNodeStatus returns the current status of a node
func (n *FilesystemNavigator) GetNodeStatus(nodePath string) (*NodeStatus, error) {
	nodePath = n.normalizePath(nodePath)

	node, err := n.GetNode(nodePath)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, fmt.Errorf("node not found: %s", nodePath)
	}

	fsPath := n.computeFilesystemPath(nodePath)
	
	status := &NodeStatus{
		Exists:    n.pathExists(fsPath),
		LastCheck: time.Now(),
	}

	// Check if it's a repository
	if node.Type == NodeTypeRepo {
		status.State = RepoState(tree.GetRepoState(fsPath))
		status.Cloned = (status.State != RepoStateMissing)
		
		// Check if it's configured as lazy
		if nodeDef := n.findNodeDefinition(nodePath); nodeDef != nil {
			status.Lazy = nodeDef.IsLazy()
		}

		// Get git status if cloned
		if status.Cloned && n.gitCmd != nil {
			if branch, err := n.gitCmd.CurrentBranch(fsPath); err == nil {
				status.Branch = branch
			}
			if url, err := n.gitCmd.RemoteURL(fsPath); err == nil {
				status.RemoteURL = url
			}
			if modified, err := n.gitCmd.HasChanges(fsPath); err == nil {
				status.Modified = modified
				if modified {
					status.State = RepoStateModified
				}
			}
		}
	}

	return status, nil
}

// RefreshStatus forces a status refresh for a node and its children
func (n *FilesystemNavigator) RefreshStatus(nodePath string) error {
	// In filesystem navigator, status is always fresh
	// This is a no-op but could trigger cache invalidation in other implementations
	return nil
}

// IsLazy checks if a node is configured for lazy loading
func (n *FilesystemNavigator) IsLazy(nodePath string) (bool, error) {
	nodeDef := n.findNodeDefinition(nodePath)
	if nodeDef == nil {
		return false, nil
	}
	return nodeDef.IsLazy(), nil
}

// TriggerLazyLoad initiates loading of a lazy node
func (n *FilesystemNavigator) TriggerLazyLoad(nodePath string) error {
	node, err := n.GetNode(nodePath)
	if err != nil {
		return err
	}
	if node == nil {
		return fmt.Errorf("node not found: %s", nodePath)
	}

	if node.Type != NodeTypeRepo || node.URL == "" {
		return fmt.Errorf("node %s is not a repository", nodePath)
	}

	fsPath := n.computeFilesystemPath(nodePath)
	
	// Check if already cloned
	if n.pathExists(filepath.Join(fsPath, ".git")) {
		return nil
	}

	// Create parent directory
	parentDir := filepath.Dir(fsPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Clone the repository
	fmt.Printf("Cloning %s to %s\n", node.URL, fsPath)
	if err := n.gitCmd.Clone(node.URL, fsPath); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// Helper methods

func (n *FilesystemNavigator) normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		// Relative path - append to current
		p = path.Join(n.currentPath, p)
	}
	// Clean the path
	p = path.Clean(p)
	if p == "" || p == "." {
		return "/"
	}
	return p
}

func (n *FilesystemNavigator) computeFilesystemPath(logicalPath string) string {
	reposDir := n.config.GetReposDir()

	// For root, use repos directory
	if logicalPath == "/" || logicalPath == "" {
		return filepath.Join(n.workspace, reposDir)
	}

	// Split path into components
	parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")
	
	// Check if this is a top-level repository (directly under root)
	if len(parts) == 1 {
		// Top-level repos go directly in workspace
		return filepath.Join(n.workspace, parts[0])
	}

	// For nested paths, build the full path
	pathComponents := []string{n.workspace}
	
	// Check if first component is a repo or uses repos dir
	firstNodeDef := n.findNodeDefinition("/" + parts[0])
	if firstNodeDef != nil && firstNodeDef.URL != "" {
		// It's a repository at top level
		pathComponents = append(pathComponents, parts...)
	} else {
		// Use repos directory for non-repo top-level nodes
		pathComponents = append(pathComponents, reposDir)
		pathComponents = append(pathComponents, parts...)
	}

	return filepath.Join(pathComponents...)
}

func (n *FilesystemNavigator) getRootChildren() []string {
	children := []string{}
	
	// Get children from config
	for _, node := range n.config.Nodes {
		children = append(children, node.Name)
	}

	// Also check filesystem for directories not in config
	reposDir := filepath.Join(n.workspace, n.config.GetReposDir())
	if entries, err := os.ReadDir(reposDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				// Check if not already in children
				found := false
				for _, child := range children {
					if child == entry.Name() {
						found = true
						break
					}
				}
				if !found {
					children = append(children, entry.Name())
				}
			}
		}
	}

	return children
}

func (n *FilesystemNavigator) buildNodeFromPath(nodePath string) *Node {
	// Try to find in config first
	nodeDef := n.findNodeDefinition(nodePath)
	
	node := &Node{
		Path:     nodePath,
		Name:     path.Base(nodePath),
		Children: []string{},
	}

	if nodeDef != nil {
		if nodeDef.URL != "" {
			node.Type = NodeTypeRepo
			node.URL = nodeDef.URL
		} else if nodeDef.Config != "" {
			node.Type = NodeTypeConfig
			node.ConfigRef = nodeDef.Config
		} else {
			node.Type = NodeTypeDirectory
		}
	} else {
		// Check filesystem
		fsPath := n.computeFilesystemPath(nodePath)
		if n.pathExists(filepath.Join(fsPath, ".git")) {
			node.Type = NodeTypeRepo
			// Try to get URL from git
			if n.gitCmd != nil {
				if url, err := n.gitCmd.RemoteURL(fsPath); err == nil {
					node.URL = url
				}
			}
		} else if n.pathExists(fsPath) {
			node.Type = NodeTypeDirectory
		} else {
			return nil // Node doesn't exist
		}
	}

	// Get children
	node.Children = n.getNodeChildren(nodePath)

	return node
}

func (n *FilesystemNavigator) getNodeChildren(nodePath string) []string {
	children := []string{}

	// Check if this is a config reference
	nodeDef := n.findNodeDefinition(nodePath)
	if nodeDef != nil && nodeDef.Config != "" {
		// Load referenced config and get its nodes
		if cfg, err := n.resolver.LoadNodeConfig(nodeDef.Config, nodeDef); err == nil {
			for _, child := range cfg.Nodes {
				children = append(children, child.Name)
			}
		}
	}

	// Check filesystem
	fsPath := n.computeFilesystemPath(nodePath)
	if entries, err := os.ReadDir(fsPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				// Check if not already in children
				found := false
				for _, child := range children {
					if child == entry.Name() {
						found = true
						break
					}
				}
				if !found {
					children = append(children, entry.Name())
				}
			}
		}
	}

	return children
}

func (n *FilesystemNavigator) findNodeDefinition(nodePath string) *config.NodeDefinition {
	// For root-level nodes
	parts := strings.Split(strings.TrimPrefix(nodePath, "/"), "/")
	if len(parts) == 1 {
		for _, node := range n.config.Nodes {
			if node.Name == parts[0] {
				return &node
			}
		}
	}
	
	// For nested nodes, would need to traverse config references
	// This is simplified for now
	return nil
}

func (n *FilesystemNavigator) buildTreeView(view *TreeView, node *Node, currentDepth, maxDepth int) error {
	// Add node to view
	view.Nodes[node.Path] = node

	// Get status
	if status, err := n.GetNodeStatus(node.Path); err == nil {
		view.Status[node.Path] = status
	}

	// Check depth limit
	if maxDepth >= 0 && currentDepth >= maxDepth {
		return nil
	}

	// Recurse into children
	for _, childName := range node.Children {
		childPath := path.Join(node.Path, childName)
		child, err := n.GetNode(childPath)
		if err != nil {
			continue
		}
		if child != nil {
			if err := n.buildTreeView(view, child, currentDepth+1, maxDepth); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *FilesystemNavigator) pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (n *FilesystemNavigator) loadCurrentPath() {
	data, err := os.ReadFile(n.currentFile)
	if err != nil {
		n.currentPath = "/"
		return
	}

	path := strings.TrimSpace(string(data))
	if path == "" || !strings.HasPrefix(path, "/") {
		n.currentPath = "/"
	} else {
		n.currentPath = path
	}
}

func (n *FilesystemNavigator) saveCurrentPath() error {
	// Ensure .muno directory exists
	munoDir := filepath.Dir(n.currentFile)
	if err := os.MkdirAll(munoDir, 0755); err != nil {
		return fmt.Errorf("creating .muno directory: %w", err)
	}

	return os.WriteFile(n.currentFile, []byte(n.currentPath), 0644)
}