package tree

import (
	"time"
)

// NodeType represents the type of node in the tree
type NodeType string

const (
	NodeTypeRoot NodeType = "root"
	NodeTypeRepo NodeType = "repo"
)

// RepoState represents the state of a repository
type RepoState string

const (
	RepoStateMissing  RepoState = "missing"
	RepoStateCloned   RepoState = "cloned"
	RepoStateModified RepoState = "modified"
)

// TreeNode represents a node in the workspace tree
// Contains ONLY logical structure and repository metadata
// NO filesystem paths are stored
type TreeNode struct {
	Name     string   `json:"name"`
	Type     NodeType `json:"type"`     // "root" or "repo"
	Children []string `json:"children"` // Just names, not paths
	
	// Repository metadata (only for type="repo")
	URL   string    `json:"url,omitempty"`
	Lazy  bool      `json:"lazy,omitempty"`
	State RepoState `json:"state,omitempty"` // "missing", "cloned", "modified"
}

// TreeState represents the persistent state of the tree
// Contains ONLY logical paths and tree structure
type TreeState struct {
	CurrentPath string                `json:"current_path"`
	Nodes       map[string]*TreeNode  `json:"nodes"`
	LastUpdated time.Time             `json:"last_updated"`
}

// UseOptions represents options for the Use command
type UseOptions struct {
	NoClone bool // Skip auto-cloning of lazy repos
}

// AddOptions represents options for the Add command
type AddOptions struct {
	Name string // Custom name for the repository
	Lazy bool   // Don't clone immediately
}

// ResolutionSource indicates how a target was resolved
type ResolutionSource string

const (
	SourceExplicit ResolutionSource = "explicit" // User provided explicit path
	SourceCWD      ResolutionSource = "CWD"      // Resolved from current working directory
	SourceStored   ResolutionSource = "stored"   // From stored current node
	SourceRoot     ResolutionSource = "root"     // Fallback to root
)

// TargetResolution represents how a target node was resolved
type TargetResolution struct {
	Path   string           // Logical path to the node
	Node   *TreeNode        // The resolved node
	Source ResolutionSource // How it was resolved
}

// TreeDisplay represents options for displaying the tree
type TreeDisplay struct {
	MaxDepth     int
	ShowRepos    bool
	ShowLazy     bool
	ShowModified bool
}

// Helper methods for TreeNode

// IsRoot returns true if this is the root node
func (n *TreeNode) IsRoot() bool {
	return n.Type == NodeTypeRoot
}

// IsLeaf returns true if this node has no children
func (n *TreeNode) IsLeaf() bool {
	return len(n.Children) == 0
}

// HasLazyRepos returns true if this is a lazy repo that hasn't been cloned
func (n *TreeNode) HasLazyRepos() bool {
	return n.Type == NodeTypeRepo && n.Lazy && n.State == RepoStateMissing
}

// NeedsClone returns true if this repo needs to be cloned
func (n *TreeNode) NeedsClone() bool {
	return n.Type == NodeTypeRepo && n.State == RepoStateMissing
}