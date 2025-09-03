package adapters

import (
	"fmt"
	
	"github.com/taokim/muno/internal/interfaces"
)

// TreeAdapter is a stub implementation of TreeProvider
// TODO: Implement actual tree operations when tree package is refactored
type TreeAdapterStub struct {
	current interfaces.NodeInfo
	nodes   map[string]interfaces.NodeInfo
	state   interfaces.TreeState
}

// NewTreeAdapter creates a new tree adapter
func NewTreeAdapter() interfaces.TreeProvider {
	return &TreeAdapterStub{
		nodes: make(map[string]interfaces.NodeInfo),
		state: interfaces.TreeState{
			Nodes: make(map[string]interfaces.NodeInfo),
		},
	}
}

// Load loads the tree from configuration
func (t *TreeAdapterStub) Load(cfg interface{}) error {
	// Stub implementation
	return nil
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

// GetTree returns the root node of the tree
func (t *TreeAdapterStub) GetTree() (interfaces.NodeInfo, error) {
	if root, ok := t.nodes["/"]; ok {
		return root, nil
	}
	return interfaces.NodeInfo{Name: "root", Path: "/"}, nil
}

// GetNode gets a specific node by path
func (t *TreeAdapterStub) GetNode(path string) (interfaces.NodeInfo, error) {
	if node, ok := t.nodes[path]; ok {
		return node, nil
	}
	return interfaces.NodeInfo{}, fmt.Errorf("node not found: %s", path)
}

// AddNode adds a new node to the tree
func (t *TreeAdapterStub) AddNode(parentPath string, node interfaces.NodeInfo) error {
	fullPath := parentPath + "/" + node.Name
	if parentPath == "/" || parentPath == "" {
		fullPath = node.Name
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