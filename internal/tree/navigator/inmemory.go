package navigator

import (
	"fmt"
	"path"
	"strings"
	"sync"
	"time"
	
	"github.com/taokim/muno/internal/config"
)

// InMemoryNavigator implements TreeNavigator entirely in memory.
// This is primarily used for testing but can also be used for
// read-only operations or isolated scenarios.
type InMemoryNavigator struct {
	nodes       map[string]*Node
	status      map[string]*NodeStatus
	currentPath string
	config      *config.ConfigTree
	mu          sync.RWMutex
}

// NewInMemoryNavigator creates a new in-memory navigator
func NewInMemoryNavigator() *InMemoryNavigator {
	nav := &InMemoryNavigator{
		nodes:       make(map[string]*Node),
		status:      make(map[string]*NodeStatus),
		currentPath: "/",
	}

	// Initialize with root node
	nav.nodes["/"] = &Node{
		Path:     "/",
		Name:     "root",
		Type:     NodeTypeRoot,
		Children: []string{},
	}

	nav.status["/"] = &NodeStatus{
		Exists:    true,
		Cloned:    true,
		State:     RepoStateCloned,
		LastCheck: time.Now(),
	}

	return nav
}

// GetCurrentPath returns the current position in the tree
func (n *InMemoryNavigator) GetCurrentPath() (string, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.currentPath, nil
}

// Navigate changes the current position to the specified path
func (n *InMemoryNavigator) Navigate(path string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	targetPath := n.normalizePath(path)
	
	// Check if node exists
	if _, exists := n.nodes[targetPath]; !exists {
		return fmt.Errorf("node not found: %s", targetPath)
	}

	// Check if it's a lazy node that needs "loading"
	if status, exists := n.status[targetPath]; exists && status.Lazy && !status.Cloned {
		// Simulate lazy loading
		status.Cloned = true
		status.State = RepoStateCloned
		status.LastCheck = time.Now()
	}

	n.currentPath = targetPath
	return nil
}

// GetNode retrieves a single node by its path
func (n *InMemoryNavigator) GetNode(path string) (*Node, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	path = n.normalizePath(path)
	
	node, exists := n.nodes[path]
	if !exists {
		return nil, nil
	}

	// Return a copy to prevent external modification
	nodeCopy := *node
	nodeCopy.Children = make([]string, len(node.Children))
	copy(nodeCopy.Children, node.Children)
	
	return &nodeCopy, nil
}

// ListChildren returns all direct children of a node
func (n *InMemoryNavigator) ListChildren(path string) ([]*Node, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	path = n.normalizePath(path)
	
	parent, exists := n.nodes[path]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", path)
	}

	children := make([]*Node, 0, len(parent.Children))
	for _, childName := range parent.Children {
		childPath := joinPath(path, childName)
		if child, exists := n.nodes[childPath]; exists {
			// Return copies
			childCopy := *child
			childCopy.Children = make([]string, len(child.Children))
			copy(childCopy.Children, child.Children)
			children = append(children, &childCopy)
		}
	}

	return children, nil
}

// GetTree returns a tree view starting from path with specified depth
func (n *InMemoryNavigator) GetTree(startPath string, depth int) (*TreeView, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	startPath = n.normalizePath(startPath)
	
	root, exists := n.nodes[startPath]
	if !exists {
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
	n.buildTreeView(view, root, 0, depth)

	return view, nil
}

// GetNodeStatus returns the current status of a node
func (n *InMemoryNavigator) GetNodeStatus(path string) (*NodeStatus, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	path = n.normalizePath(path)
	
	status, exists := n.status[path]
	if !exists {
		// Create default status
		return &NodeStatus{
			Exists:    n.nodes[path] != nil,
			LastCheck: time.Now(),
		}, nil
	}

	// Return a copy
	statusCopy := *status
	statusCopy.LastCheck = time.Now()
	
	return &statusCopy, nil
}

// RefreshStatus forces a status refresh for a node and its children
func (n *InMemoryNavigator) RefreshStatus(path string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	path = n.normalizePath(path)
	
	// For InMemoryNavigator in tests, RefreshStatus should sync with config
	// This allows AddRepo and RemoveNode to work properly in tests
	if n.config != nil && path == "/" {
		// Sync root children with config nodes
		rootNode := n.nodes["/"]
		if rootNode != nil {
			// Build a map of config nodes
			configChildren := make(map[string]bool)
			for _, nodeDef := range n.config.Nodes {
				configChildren[nodeDef.Name] = true
			}
			
			// Remove nodes that are no longer in config
			newChildren := []string{}
			for _, child := range rootNode.Children {
				if configChildren[child] {
					newChildren = append(newChildren, child)
				} else {
					// Remove the node and its status
					childPath := "/" + child
					delete(n.nodes, childPath)
					delete(n.status, childPath)
					// Also remove all descendants
					n.removeNodeRecursive(childPath)
				}
			}
			rootNode.Children = newChildren
			
			// Add any missing nodes from config
			existingChildren := make(map[string]bool)
			for _, child := range rootNode.Children {
				existingChildren[child] = true
			}
			
			for _, nodeDef := range n.config.Nodes {
				if !existingChildren[nodeDef.Name] {
					// Add child to root
					rootNode.Children = append(rootNode.Children, nodeDef.Name)
					
					// Create the child node
					childPath := "/" + nodeDef.Name
					nodeType := NodeTypeRepo
					if nodeDef.ConfigRef != "" {
						nodeType = NodeTypeConfig
					}
					
					n.nodes[childPath] = &Node{
						Path:      childPath,
						Name:      nodeDef.Name,
						Type:      nodeType,
						URL:       nodeDef.URL,
						ConfigRef: nodeDef.ConfigRef,
						Children:  []string{},
					}
					
					// Create status
					n.status[childPath] = &NodeStatus{
						Exists:    true,
						Lazy:      nodeDef.IsLazy(),
						Cloned:    !nodeDef.IsLazy(),
						State:     RepoStateCloned,
						RemoteURL: nodeDef.URL,
						LastCheck: time.Now(),
					}
				}
			}
		}
	}
	
	// Update last check time
	if status, exists := n.status[path]; exists {
		status.LastCheck = time.Now()
	}

	return nil
}

// IsLazy checks if a node is configured for lazy loading
func (n *InMemoryNavigator) IsLazy(path string) (bool, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	path = n.normalizePath(path)
	
	if status, exists := n.status[path]; exists {
		return status.Lazy, nil
	}

	return false, nil
}

// TriggerLazyLoad initiates loading of a lazy node
func (n *InMemoryNavigator) TriggerLazyLoad(path string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	path = n.normalizePath(path)
	
	node, exists := n.nodes[path]
	if !exists {
		return fmt.Errorf("node not found: %s", path)
	}

	if node.Type != NodeTypeRepo {
		return fmt.Errorf("node %s is not a repository", path)
	}

	// Simulate lazy loading
	if status, exists := n.status[path]; exists {
		if !status.Lazy {
			return fmt.Errorf("node %s is not configured for lazy loading", path)
		}
		status.Cloned = true
		status.State = RepoStateCloned
		status.LastCheck = time.Now()
	}

	return nil
}

// Test helper methods for setting up the in-memory tree

// SetConfig sets the configuration for the in-memory navigator
func (n *InMemoryNavigator) SetConfig(cfg *config.ConfigTree) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.config = cfg
}

// AddNode adds a node to the in-memory tree
func (n *InMemoryNavigator) AddNode(nodePath string, node *Node) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	nodePath = n.normalizePath(nodePath)
	
	// Ensure node path matches
	node.Path = nodePath
	
	// Add to nodes map
	n.nodes[nodePath] = node

	// Create default status
	if _, exists := n.status[nodePath]; !exists {
		n.status[nodePath] = &NodeStatus{
			Exists:    true,
			Cloned:    node.Type != NodeTypeRepo,
			State:     RepoStateCloned,
			LastCheck: time.Now(),
		}
	}

	// Update parent's children
	if nodePath != "/" {
		parentPath := path.Dir(nodePath)
		if parentPath == "." {
			parentPath = "/"
		}
		
		if parent, exists := n.nodes[parentPath]; exists {
			nodeName := path.Base(nodePath)
			// Check if child already exists
			found := false
			for _, child := range parent.Children {
				if child == nodeName {
					found = true
					break
				}
			}
			if !found {
				parent.Children = append(parent.Children, nodeName)
			}
		}
	}

	return nil
}

// SetNodeStatus sets the status for a node
func (n *InMemoryNavigator) SetNodeStatus(path string, status *NodeStatus) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	path = n.normalizePath(path)
	
	if _, exists := n.nodes[path]; !exists {
		return fmt.Errorf("node not found: %s", path)
	}

	n.status[path] = status
	return nil
}

// RemoveNode removes a node and its children from the tree
func (n *InMemoryNavigator) RemoveNode(nodePath string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	nodePath = n.normalizePath(nodePath)
	
	if nodePath == "/" {
		return fmt.Errorf("cannot remove root node")
	}

	// Remove from parent's children
	parentPath := path.Dir(nodePath)
	if parentPath == "." {
		parentPath = "/"
	}
	
	if parent, exists := n.nodes[parentPath]; exists {
		nodeName := path.Base(nodePath)
		newChildren := []string{}
		for _, child := range parent.Children {
			if child != nodeName {
				newChildren = append(newChildren, child)
			}
		}
		parent.Children = newChildren
	}

	// Remove node and all descendants
	n.removeNodeRecursive(nodePath)

	return nil
}

// Clear removes all nodes except root
func (n *InMemoryNavigator) Clear() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.nodes = make(map[string]*Node)
	n.status = make(map[string]*NodeStatus)
	n.currentPath = "/"

	// Re-initialize root
	n.nodes["/"] = &Node{
		Path:     "/",
		Name:     "root",
		Type:     NodeTypeRoot,
		Children: []string{},
	}

	n.status["/"] = &NodeStatus{
		Exists:    true,
		Cloned:    true,
		State:     RepoStateCloned,
		LastCheck: time.Now(),
	}
}

// Helper methods

func (n *InMemoryNavigator) normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		// Relative path - append to current
		p = joinPath(n.currentPath, p)
	}
	// Clean the path
	p = path.Clean(p)
	if p == "" || p == "." {
		return "/"
	}
	return p
}

func (n *InMemoryNavigator) buildTreeView(view *TreeView, node *Node, currentDepth, maxDepth int) {
	// Add node to view
	view.Nodes[node.Path] = node

	// Add status
	if status, exists := n.status[node.Path]; exists {
		view.Status[node.Path] = status
	}

	// Check depth limit
	if maxDepth >= 0 && currentDepth >= maxDepth {
		return
	}

	// Recurse into children
	for _, childName := range node.Children {
		childPath := joinPath(node.Path, childName)
		if child, exists := n.nodes[childPath]; exists {
			n.buildTreeView(view, child, currentDepth+1, maxDepth)
		}
	}
}

func (n *InMemoryNavigator) removeNodeRecursive(nodePath string) {
	node, exists := n.nodes[nodePath]
	if !exists {
		return
	}

	// Remove all children first
	for _, childName := range node.Children {
		childPath := joinPath(nodePath, childName)
		n.removeNodeRecursive(childPath)
	}

	// Remove this node
	delete(n.nodes, nodePath)
	delete(n.status, nodePath)
}

func joinPath(base, child string) string {
	if base == "/" {
		return "/" + child
	}
	return base + "/" + child
}