package adapters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
)

func TestNewTreeAdapter(t *testing.T) {
	adapter := NewTreeAdapter()
	assert.NotNil(t, adapter)
	
	// Check it returns the correct type
	stub, ok := adapter.(*TreeAdapterStub)
	assert.True(t, ok)
	assert.NotNil(t, stub)
}

func TestTreeAdapterStub_Load(t *testing.T) {
	adapter := &TreeAdapterStub{
		nodes: make(map[string]interfaces.NodeInfo),
	}

	// Load expects a *config.ConfigTree
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{
				Name:  "repo1",
				URL:   "https://github.com/test/repo1.git",
				Fetch: "eager",
			},
			{
				Name:  "repo2",
				URL:   "https://github.com/test/repo2.git",
				Fetch: "lazy",
			},
		},
	}

	err := adapter.Load(cfg)
	assert.NoError(t, err)

	// Verify root node was created
	root, err := adapter.GetTree()
	assert.NoError(t, err)
	assert.Equal(t, "test-workspace", root.Name)
	assert.Equal(t, "/", root.Path)

	// Verify child nodes were created
	node1, err := adapter.GetNode("/repo1")
	assert.NoError(t, err)
	assert.Equal(t, "repo1", node1.Name)
	assert.Equal(t, "https://github.com/test/repo1.git", node1.Repository)
	assert.False(t, node1.IsLazy)

	node2, err := adapter.GetNode("/repo2")
	assert.NoError(t, err)
	assert.Equal(t, "repo2", node2.Name)
	assert.True(t, node2.IsLazy)
}

func TestTreeAdapterStub_Navigate(t *testing.T) {
	adapter := &TreeAdapterStub{
		nodes: make(map[string]interfaces.NodeInfo),
	}
	
	// Navigate to non-existent node should error
	err := adapter.Navigate("test/path")
	assert.Error(t, err)
	
	// Add node and navigate successfully
	adapter.nodes["test/path"] = interfaces.NodeInfo{Path: "test/path"}
	err = adapter.Navigate("test/path")
	assert.NoError(t, err)
}

func TestTreeAdapterStub_GetCurrent(t *testing.T) {
	adapter := &TreeAdapterStub{}
	
	node, err := adapter.GetCurrent()
	assert.NoError(t, err)
	assert.Equal(t, interfaces.NodeInfo{}, node) // Returns empty struct by default
}

func TestTreeAdapterStub_GetTree(t *testing.T) {
	adapter := &TreeAdapterStub{
		nodes: make(map[string]interfaces.NodeInfo),
	}
	
	// When root doesn't exist, returns default root
	tree, err := adapter.GetTree()
	assert.NoError(t, err)
	assert.Equal(t, "root", tree.Name)
	assert.Equal(t, "/", tree.Path)
	
	// Add custom root node
	adapter.nodes["/"] = interfaces.NodeInfo{Name: "myroot", Path: "/", Repository: "test"}
	tree, err = adapter.GetTree()
	assert.NoError(t, err)
	assert.Equal(t, "myroot", tree.Name)
	assert.Equal(t, "/", tree.Path)
	assert.Equal(t, "test", tree.Repository)
}

func TestTreeAdapterStub_GetNode(t *testing.T) {
	adapter := &TreeAdapterStub{
		nodes: make(map[string]interfaces.NodeInfo),
	}
	
	// Non-existent node should error
	node, err := adapter.GetNode("test/path")
	assert.Error(t, err)
	assert.Equal(t, interfaces.NodeInfo{}, node)
	
	// Add node and get it successfully
	adapter.nodes["test/path"] = interfaces.NodeInfo{Name: "test", Path: "test/path"}
	node, err = adapter.GetNode("test/path")
	assert.NoError(t, err)
	assert.Equal(t, "test", node.Name)
	assert.Equal(t, "test/path", node.Path)
}

func TestTreeAdapterStub_AddNode(t *testing.T) {
	adapter := &TreeAdapterStub{
		nodes: make(map[string]interfaces.NodeInfo),
	}
	
	node := interfaces.NodeInfo{
		Name: "test",
		Path: "test/path",
	}
	
	err := adapter.AddNode("parent", node)
	assert.NoError(t, err)
}

func TestTreeAdapterStub_RemoveNode(t *testing.T) {
	adapter := &TreeAdapterStub{}
	
	err := adapter.RemoveNode("test/path")
	assert.NoError(t, err)
}

func TestTreeAdapterStub_UpdateNode(t *testing.T) {
	adapter := &TreeAdapterStub{
		nodes: make(map[string]interfaces.NodeInfo),
	}
	
	node := interfaces.NodeInfo{
		Name: "updated",
		Path: "test/path",
	}
	
	err := adapter.UpdateNode("test/path", node)
	assert.NoError(t, err)
}

func TestTreeAdapterStub_ListChildren(t *testing.T) {
	adapter := &TreeAdapterStub{}
	
	children, err := adapter.ListChildren("parent/path")
	assert.NoError(t, err)
	assert.Nil(t, children) // Stub returns nil
}

func TestTreeAdapterStub_GetPath(t *testing.T) {
	adapter := &TreeAdapterStub{}
	
	// GetPath takes no parameters and returns current path
	path := adapter.GetPath()
	assert.Empty(t, path) // Stub returns empty string when current is not set
}

func TestTreeAdapterStub_SetPath(t *testing.T) {
	adapter := &TreeAdapterStub{
		nodes: make(map[string]interfaces.NodeInfo),
	}
	
	// SetPath takes only path parameter
	err := adapter.SetPath("new/path")
	assert.Error(t, err) // Should error because node doesn't exist
	
	// Add a node first
	adapter.nodes["new/path"] = interfaces.NodeInfo{Path: "new/path"}
	err = adapter.SetPath("new/path")
	assert.NoError(t, err)
}

func TestTreeAdapterStub_GetState(t *testing.T) {
	adapter := &TreeAdapterStub{
		state: interfaces.TreeState{
			Nodes: make(map[string]interfaces.NodeInfo),
		},
	}
	
	state, err := adapter.GetState()
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.NotNil(t, state.Nodes)
}

func TestTreeAdapterStub_SetState(t *testing.T) {
	adapter := &TreeAdapterStub{}
	
	state := interfaces.TreeState{
		CurrentPath: "/current",
		Nodes: map[string]interfaces.NodeInfo{
			"/": {Name: "root", Path: "/"},
		},
	}
	
	err := adapter.SetState(state)
	assert.NoError(t, err)
}