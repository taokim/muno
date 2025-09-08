// Package navigator provides tree traversal implementations for MUNO
package navigator

// TreeNavigator is the primary interface for tree traversal operations.
// Implementations can provide different strategies for node discovery
// and status tracking (filesystem, in-memory, remote, etc).
type TreeNavigator interface {
	// Navigation operations
	
	// GetCurrentPath returns the current position in the tree
	GetCurrentPath() (string, error)
	
	// Navigate changes the current position to the specified path
	// Path can be absolute (/backend) or relative (../frontend)
	Navigate(path string) error
	
	// Node discovery operations
	
	// GetNode retrieves a single node by its path
	// Returns nil if node doesn't exist
	GetNode(path string) (*Node, error)
	
	// ListChildren returns all direct children of a node
	// Returns empty slice if node has no children
	ListChildren(path string) ([]*Node, error)
	
	// GetTree returns a tree view starting from path with specified depth
	// Depth of -1 means unlimited depth
	GetTree(path string, depth int) (*TreeView, error)
	
	// Status and state operations
	
	// GetNodeStatus returns the current status of a node
	// This includes clone state, modification status, etc.
	GetNodeStatus(path string) (*NodeStatus, error)
	
	// RefreshStatus forces a status refresh for a node and its children
	// Useful when external changes may have occurred
	RefreshStatus(path string) error
	
	// Lazy loading operations
	
	// IsLazy checks if a node is configured for lazy loading
	IsLazy(path string) (bool, error)
	
	// TriggerLazyLoad initiates loading of a lazy node
	// This typically means cloning a repository
	TriggerLazyLoad(path string) error
}

// NodeVisitor is a callback for tree traversal operations
type NodeVisitor func(node *Node, depth int) error

// TreeWalker provides advanced tree traversal operations
type TreeWalker interface {
	// Walk traverses the tree starting from path, calling visitor for each node
	Walk(nav TreeNavigator, path string, visitor NodeVisitor) error
	
	// WalkBreadthFirst performs breadth-first traversal
	WalkBreadthFirst(nav TreeNavigator, path string, visitor NodeVisitor) error
	
	// WalkDepthFirst performs depth-first traversal
	WalkDepthFirst(nav TreeNavigator, path string, visitor NodeVisitor) error
}