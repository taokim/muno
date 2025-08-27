package tree

import (
	"time"
)

// Node represents a node in the workspace tree
type Node struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Path     string           `json:"path"`     // Relative path from workspace root
	FullPath string           `json:"full_path"` // Absolute filesystem path
	Parent   *Node            `json:"-"`
	Children map[string]*Node `json:"children,omitempty"`
	Repos    []RepoConfig     `json:"repos,omitempty"`
	Meta     NodeMeta         `json:"meta"`
}

// RepoConfig represents a repository configuration within a node
type RepoConfig struct {
	URL   string `json:"url"`
	Path  string `json:"path"`
	Name  string `json:"name"`
	Lazy  bool   `json:"lazy"`
	State string `json:"state"` // "missing", "cloned", "modified"
}

// NodeMeta contains metadata about a node
type NodeMeta struct {
	Type        string    `json:"type"` // "persistent" or "ephemeral"
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
	Description string    `json:"description,omitempty"`
	README      string    `json:"readme,omitempty"` // README.md content
}

// TreeState represents the persistent state of the tree
type TreeState struct {
	CurrentNodePath string           `json:"current_node_path"`
	PreviousNodePath string          `json:"previous_node_path,omitempty"`
	Nodes          map[string]*Node `json:"nodes"`
	LastUpdated    time.Time        `json:"last_updated"`
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
	SourceExplicit ResolutionSource = "explicit"     // User provided explicit path
	SourceCWD      ResolutionSource = "CWD"         // Resolved from current working directory
	SourceStored   ResolutionSource = "stored"      // From stored current node
	SourceRoot     ResolutionSource = "root"        // Fallback to root
)

// TargetResolution represents how a target node was resolved
type TargetResolution struct {
	Node   *Node
	Source ResolutionSource
}

// NodeType represents the type of node
type NodeType string

const (
	NodeTypePersistent NodeType = "persistent"
	NodeTypeEphemeral  NodeType = "ephemeral"
)

// RepoState represents the state of a repository
type RepoState string

const (
	RepoStateMissing  RepoState = "missing"
	RepoStateCloned   RepoState = "cloned"
	RepoStateModified RepoState = "modified"
)

// TreeDisplay represents options for displaying the tree
type TreeDisplay struct {
	MaxDepth     int
	ShowRepos    bool
	ShowLazy     bool
	ShowModified bool
}

// IsRoot returns true if this is the root node
func (n *Node) IsRoot() bool {
	return n.Parent == nil
}

// IsLeaf returns true if this node has no children
func (n *Node) IsLeaf() bool {
	return len(n.Children) == 0
}

// GetPath returns the tree path (e.g., "/team/backend")
func (n *Node) GetPath() string {
	if n.IsRoot() {
		return "/"
	}
	return n.Path
}

// GetDepth returns the depth of this node in the tree
func (n *Node) GetDepth() int {
	depth := 0
	current := n
	for current.Parent != nil {
		depth++
		current = current.Parent
	}
	return depth
}

// HasLazyRepos returns true if this node has any lazy repositories
func (n *Node) HasLazyRepos() bool {
	for _, repo := range n.Repos {
		if repo.Lazy && repo.State == string(RepoStateMissing) {
			return true
		}
	}
	return false
}

// CountRepos counts the total number of repositories in this node and its subtree
func (n *Node) CountRepos(recursive bool) int {
	count := len(n.Repos)
	
	if recursive {
		for _, child := range n.Children {
			count += child.CountRepos(true)
		}
	}
	
	return count
}