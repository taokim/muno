package navigator

import (
	"time"
)

// NodeType represents the type of a tree node
type NodeType string

const (
	// NodeTypeRoot represents the root of the tree
	NodeTypeRoot NodeType = "root"
	
	// NodeTypeRepo represents a git repository
	NodeTypeRepo NodeType = "repo"
	
	// NodeTypeFile represents a configuration reference
	NodeTypeFile NodeType = "config"
	
	// NodeTypeDirectory represents a plain directory
	NodeTypeDirectory NodeType = "directory"
)

// Node represents a single node in the tree structure
type Node struct {
	// Path is the absolute path in the tree (e.g., /backend/services/auth)
	Path string `json:"path"`
	
	// Name is the node's name (e.g., "auth")
	Name string `json:"name"`
	
	// Type indicates what kind of node this is
	Type NodeType `json:"type"`
	
	// URL is the git repository URL (only for NodeTypeRepo)
	URL string `json:"url,omitempty"`
	
	// File is the path to a configuration file (only for NodeTypeFile)
	File string `json:"file,omitempty"`
	
	// Children contains the names of child nodes
	Children []string `json:"children"`
	
	// Metadata for additional properties
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RepoState represents the state of a repository node
type RepoState string

const (
	// RepoStateMissing indicates the repository hasn't been cloned
	RepoStateMissing RepoState = "missing"
	
	// RepoStateCloned indicates the repository is cloned and clean
	RepoStateCloned RepoState = "cloned"
	
	// RepoStateModified indicates the repository has uncommitted changes
	RepoStateModified RepoState = "modified"
	
	// RepoStateAhead indicates the repository is ahead of remote
	RepoStateAhead RepoState = "ahead"
	
	// RepoStateBehind indicates the repository is behind remote
	RepoStateBehind RepoState = "behind"
	
	// RepoStateDiverged indicates the repository has diverged from remote
	RepoStateDiverged RepoState = "diverged"
)

// NodeStatus represents the runtime status of a node
type NodeStatus struct {
	// Exists indicates if the node exists on the filesystem
	Exists bool `json:"exists"`
	
	// Cloned indicates if a repository has been cloned
	Cloned bool `json:"cloned"`
	
	// State represents the repository state
	State RepoState `json:"state"`
	
	// Modified indicates if there are uncommitted changes
	Modified bool `json:"modified"`
	
	// Lazy indicates if this node is configured for lazy loading
	Lazy bool `json:"lazy"`
	
	// Branch is the current git branch (for repositories)
	Branch string `json:"branch,omitempty"`
	
	// RemoteURL is the configured remote URL (for repositories)
	RemoteURL string `json:"remote_url,omitempty"`
	
	// LastCheck is when this status was last updated
	LastCheck time.Time `json:"last_check"`
	
	// Error contains any error encountered during status check
	Error string `json:"error,omitempty"`
}

// TreeView represents a hierarchical view of nodes
type TreeView struct {
	// Root is the starting node of this view
	Root *Node `json:"root"`
	
	// Nodes contains all nodes in the tree, keyed by path
	Nodes map[string]*Node `json:"nodes"`
	
	// Status contains status for each node, keyed by path
	Status map[string]*NodeStatus `json:"status"`
	
	// Depth indicates how deep this view goes (-1 for unlimited)
	Depth int `json:"depth"`
	
	// Generated indicates when this view was created
	Generated time.Time `json:"generated"`
}

// NavigatorOptions configures navigator behavior
type NavigatorOptions struct {
	// WorkspacePath is the root directory for the workspace
	WorkspacePath string
	
	// FilePath is the path to the configuration file
	FilePath string
	
	// CacheEnabled enables caching for performance
	CacheEnabled bool
	
	// CacheTTL is how long cached data remains valid
	CacheTTL time.Duration
	
	// MaxCacheSize is the maximum number of nodes to cache
	MaxCacheSize int
	
	// LazyLoadTimeout is the timeout for lazy loading operations
	LazyLoadTimeout time.Duration
	
	// RefreshInterval is how often to refresh status automatically
	RefreshInterval time.Duration
}

// NewNode creates a new node with the given parameters
func NewNode(path, name string, nodeType NodeType) *Node {
	return &Node{
		Path:     path,
		Name:     name,
		Type:     nodeType,
		Children: []string{},
		Metadata: make(map[string]interface{}),
	}
}

// IsRepository returns true if this is a repository node
func (n *Node) IsRepository() bool {
	return n.Type == NodeTypeRepo
}

// IsFile returns true if this is a config reference node
func (n *Node) IsFile() bool {
	return n.Type == NodeTypeFile
}

// IsRoot returns true if this is the root node
func (n *Node) IsRoot() bool {
	return n.Type == NodeTypeRoot || n.Path == "/" || n.Path == ""
}

// HasChildren returns true if the node has children
func (n *Node) HasChildren() bool {
	return len(n.Children) > 0
}

// NeedsClone returns true if this node needs to be cloned
func (ns *NodeStatus) NeedsClone() bool {
	return !ns.Cloned && ns.Lazy
}

// IsClean returns true if the node has no modifications
func (ns *NodeStatus) IsClean() bool {
	return ns.State == RepoStateCloned && !ns.Modified
}

// HasRemoteChanges returns true if there are remote changes to pull
func (ns *NodeStatus) HasRemoteChanges() bool {
	return ns.State == RepoStateBehind || ns.State == RepoStateDiverged
}

// HasLocalChanges returns true if there are local changes to push
func (ns *NodeStatus) HasLocalChanges() bool {
	return ns.State == RepoStateAhead || ns.State == RepoStateDiverged || ns.Modified
}