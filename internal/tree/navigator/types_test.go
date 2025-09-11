package navigator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNodeTypes(t *testing.T) {
	t.Run("NewNode", func(t *testing.T) {
		// Test creating a repository node
		node := NewNode("/test/repo", "repo", NodeTypeRepo)
		assert.NotNil(t, node)
		assert.Equal(t, "/test/repo", node.Path)
		assert.Equal(t, "repo", node.Name)
		assert.Equal(t, NodeTypeRepo, node.Type)
		assert.Empty(t, node.Children)
		assert.NotNil(t, node.Metadata)
		
		// Test creating a config node
		node = NewNode("/config", "config", NodeTypeConfig)
		assert.Equal(t, NodeTypeConfig, node.Type)
		
		// Test creating a root node
		node = NewNode("/", "root", NodeTypeRoot)
		assert.Equal(t, NodeTypeRoot, node.Type)
	})

	t.Run("IsRepository", func(t *testing.T) {
		repoNode := &Node{Type: NodeTypeRepo}
		assert.True(t, repoNode.IsRepository())
		
		configNode := &Node{Type: NodeTypeConfig}
		assert.False(t, configNode.IsRepository())
		
		rootNode := &Node{Type: NodeTypeRoot}
		assert.False(t, rootNode.IsRepository())
	})

	t.Run("IsConfig", func(t *testing.T) {
		configNode := &Node{Type: NodeTypeConfig}
		assert.True(t, configNode.IsConfig())
		
		repoNode := &Node{Type: NodeTypeRepo}
		assert.False(t, repoNode.IsConfig())
		
		rootNode := &Node{Type: NodeTypeRoot}
		assert.False(t, rootNode.IsConfig())
	})

	t.Run("IsRoot", func(t *testing.T) {
		rootNode := &Node{Type: NodeTypeRoot}
		assert.True(t, rootNode.IsRoot())
		
		// A repo node with non-root path
		repoNode := &Node{Type: NodeTypeRepo, Path: "/repo"}
		assert.False(t, repoNode.IsRoot())
		
		// A config node with non-root path
		configNode := &Node{Type: NodeTypeConfig, Path: "/config"}
		assert.False(t, configNode.IsRoot())
		
		// Any node with "/" path is considered root
		rootPathNode := &Node{Type: NodeTypeRepo, Path: "/"}
		assert.True(t, rootPathNode.IsRoot())
	})

	t.Run("HasChildren", func(t *testing.T) {
		// Node with children
		node := &Node{Children: []string{"child1", "child2"}}
		assert.True(t, node.HasChildren())
		
		// Node without children
		node = &Node{Children: []string{}}
		assert.False(t, node.HasChildren())
		
		// Node with nil children
		node = &Node{}
		assert.False(t, node.HasChildren())
	})

	t.Run("NodeStatus_NeedsClone", func(t *testing.T) {
		// Lazy repo not cloned
		status := &NodeStatus{Lazy: true, Cloned: false}
		assert.True(t, status.NeedsClone())
		
		// Lazy repo already cloned
		status = &NodeStatus{Lazy: true, Cloned: true}
		assert.False(t, status.NeedsClone())
		
		// Non-lazy repo not cloned
		status = &NodeStatus{Lazy: false, Cloned: false}
		assert.False(t, status.NeedsClone())
		
		// Non-lazy repo cloned
		status = &NodeStatus{Lazy: false, Cloned: true}
		assert.False(t, status.NeedsClone())
	})

	t.Run("NodeStatus_IsClean", func(t *testing.T) {
		// Clean repo (cloned state and not modified)
		status := &NodeStatus{State: RepoStateCloned, Modified: false}
		assert.True(t, status.IsClean())
		
		// Modified repo
		status = &NodeStatus{State: RepoStateCloned, Modified: true}
		assert.False(t, status.IsClean())
		
		// Different state even if not modified
		status = &NodeStatus{State: RepoStateModified, Modified: false}
		assert.False(t, status.IsClean())
	})

	t.Run("NodeStatus_HasRemoteChanges", func(t *testing.T) {
		// With the current structure, we can't track remote changes directly
		// This would need to be determined by comparing with remote
		status := &NodeStatus{State: RepoStateBehind}
		assert.True(t, status.HasRemoteChanges())
		
		// Repo without remote changes
		status = &NodeStatus{State: RepoStateCloned}
		assert.False(t, status.HasRemoteChanges())
	})

	t.Run("NodeStatus_HasLocalChanges", func(t *testing.T) {
		// Repo with local changes (modified)
		status := &NodeStatus{Modified: true}
		assert.True(t, status.HasLocalChanges())
		
		// Repo with local changes (ahead state)
		status = &NodeStatus{State: RepoStateAhead, Modified: false}
		assert.True(t, status.HasLocalChanges())
		
		// Clean repo
		status = &NodeStatus{Modified: false, State: RepoStateCloned}
		assert.False(t, status.HasLocalChanges())
	})

	t.Run("RepoState", func(t *testing.T) {
		// Test all repo state constants
		assert.Equal(t, RepoState("missing"), RepoStateMissing)
		assert.Equal(t, RepoState("cloned"), RepoStateCloned)
		assert.Equal(t, RepoState("modified"), RepoStateModified)
		assert.Equal(t, RepoState("ahead"), RepoStateAhead)
		assert.Equal(t, RepoState("behind"), RepoStateBehind)
		assert.Equal(t, RepoState("diverged"), RepoStateDiverged)
	})

	t.Run("TreeView", func(t *testing.T) {
		// Test TreeView structure
		root := &Node{
			Path:     "/",
			Name:     "root",
			Type:     NodeTypeRoot,
			Children: []string{"repo1", "repo2"},
		}
		
		view := &TreeView{
			Root:      root,
			Nodes:     make(map[string]*Node),
			Status:    make(map[string]*NodeStatus),
			Depth:     2,
			Generated: time.Now(),
		}
		
		// Add nodes
		view.Nodes["/repo1"] = &Node{
			Path: "/repo1",
			Name: "repo1",
			Type: NodeTypeRepo,
		}
		view.Nodes["/repo2"] = &Node{
			Path: "/repo2",
			Name: "repo2",
			Type: NodeTypeRepo,
		}
		
		// Add status
		view.Status["/repo1"] = &NodeStatus{
			Exists: true,
			Cloned: true,
			State:  RepoStateCloned,
		}
		view.Status["/repo2"] = &NodeStatus{
			Exists: false,
			Lazy:   true,
			State:  RepoStateMissing,
		}
		
		// Verify structure
		assert.Equal(t, 2, len(view.Nodes))
		assert.Equal(t, 2, len(view.Status))
		assert.Equal(t, 2, view.Depth)
		assert.NotNil(t, view.Generated)
		
		// Verify nodes
		repo1 := view.Nodes["/repo1"]
		assert.NotNil(t, repo1)
		assert.True(t, repo1.IsRepository())
		
		// Verify status
		status1 := view.Status["/repo1"]
		assert.NotNil(t, status1)
		assert.True(t, status1.Cloned)
		assert.False(t, status1.NeedsClone())
		
		status2 := view.Status["/repo2"]
		assert.NotNil(t, status2)
		assert.True(t, status2.NeedsClone())
	})
}