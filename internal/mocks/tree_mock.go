package mocks

import (
	"fmt"
	"sync"
	
	"github.com/taokim/muno/internal/interfaces"
)

// MockTreeProvider is a mock implementation of TreeProvider
type MockTreeProvider struct {
	mu       sync.RWMutex
	current  interfaces.NodeInfo
	nodes    map[string]interfaces.NodeInfo
	state    interfaces.TreeState
	errors   map[string]error
	calls    []string
}

// NewMockTreeProvider creates a new mock tree provider
func NewMockTreeProvider() *MockTreeProvider {
	return &MockTreeProvider{
		nodes:  make(map[string]interfaces.NodeInfo),
		errors: make(map[string]error),
		calls:  []string{},
	}
}

// Load loads the tree from configuration
func (m *MockTreeProvider) Load(config interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "Load(config)")
	
	if err, ok := m.errors["Load"]; ok && err != nil {
		return err
	}
	
	return nil
}

// Navigate navigates to a node in the tree
func (m *MockTreeProvider) Navigate(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Navigate(%s)", path))
	
	if err, ok := m.errors["Navigate:"+path]; ok && err != nil {
		return err
	}
	
	if node, ok := m.nodes[path]; ok {
		m.current = node
	}
	
	return nil
}

// GetCurrent returns the current node information
func (m *MockTreeProvider) GetCurrent() (interfaces.NodeInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, "GetCurrent()")
	
	if err, ok := m.errors["GetCurrent"]; ok && err != nil {
		return interfaces.NodeInfo{}, err
	}
	
	return m.current, nil
}

// GetTree returns the root node of the tree
func (m *MockTreeProvider) GetTree() (interfaces.NodeInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, "GetTree()")
	
	if err, ok := m.errors["GetTree"]; ok && err != nil {
		return interfaces.NodeInfo{}, err
	}
	
	if root, ok := m.nodes["/"]; ok {
		return root, nil
	}
	
	return interfaces.NodeInfo{Name: "root", Path: "/"}, nil
}

// GetNode gets a specific node by path
func (m *MockTreeProvider) GetNode(path string) (interfaces.NodeInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("GetNode(%s)", path))
	
	if err, ok := m.errors["GetNode:"+path]; ok && err != nil {
		return interfaces.NodeInfo{}, err
	}
	
	if node, ok := m.nodes[path]; ok {
		return node, nil
	}
	
	return interfaces.NodeInfo{}, fmt.Errorf("node not found: %s", path)
}

// AddNode adds a new node to the tree
func (m *MockTreeProvider) AddNode(parentPath string, node interfaces.NodeInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("AddNode(%s, %s)", parentPath, node.Name))
	
	if err, ok := m.errors["AddNode:"+parentPath]; ok && err != nil {
		return err
	}
	
	// Store the node
	fullPath := parentPath + "/" + node.Name
	if parentPath == "/" || parentPath == "" {
		fullPath = node.Name
	}
	node.Path = fullPath
	m.nodes[fullPath] = node
	
	return nil
}

// RemoveNode removes a node from the tree
func (m *MockTreeProvider) RemoveNode(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("RemoveNode(%s)", path))
	
	if err, ok := m.errors["RemoveNode:"+path]; ok && err != nil {
		return err
	}
	
	delete(m.nodes, path)
	
	return nil
}

// UpdateNode updates a node's information
func (m *MockTreeProvider) UpdateNode(path string, node interfaces.NodeInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("UpdateNode(%s)", path))
	
	if err, ok := m.errors["UpdateNode:"+path]; ok && err != nil {
		return err
	}
	
	m.nodes[path] = node
	
	return nil
}

// ListChildren lists children of a node
func (m *MockTreeProvider) ListChildren(path string) ([]interfaces.NodeInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("ListChildren(%s)", path))
	
	if err, ok := m.errors["ListChildren:"+path]; ok && err != nil {
		return nil, err
	}
	
	var children []interfaces.NodeInfo
	for nodePath, node := range m.nodes {
		// Simple check for direct children
		if len(nodePath) > len(path) && nodePath[:len(path)] == path {
			children = append(children, node)
		}
	}
	
	return children, nil
}

// GetPath returns the current path
func (m *MockTreeProvider) GetPath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, "GetPath()")
	
	return m.current.Path
}

// SetPath sets the current path
func (m *MockTreeProvider) SetPath(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("SetPath(%s)", path))
	
	if err, ok := m.errors["SetPath:"+path]; ok && err != nil {
		return err
	}
	
	if node, ok := m.nodes[path]; ok {
		m.current = node
	}
	
	return nil
}

// GetState returns the tree state
func (m *MockTreeProvider) GetState() (interfaces.TreeState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, "GetState()")
	
	if err, ok := m.errors["GetState"]; ok && err != nil {
		return interfaces.TreeState{}, err
	}
	
	return m.state, nil
}

// SetState sets the tree state
func (m *MockTreeProvider) SetState(state interfaces.TreeState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "SetState()")
	
	if err, ok := m.errors["SetState"]; ok && err != nil {
		return err
	}
	
	m.state = state
	
	return nil
}

// Mock helper methods

// SetCurrent sets the current node
func (m *MockTreeProvider) SetCurrent(node interfaces.NodeInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.current = node
	m.nodes[node.Path] = node
}

// SetNode sets a specific node
func (m *MockTreeProvider) SetNode(path string, node interfaces.NodeInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	node.Path = path
	m.nodes[path] = node
}

// SetError sets an error for a specific operation
func (m *MockTreeProvider) SetError(operation, path string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := operation
	if path != "" {
		key = operation + ":" + path
	}
	m.errors[key] = err
}

// GetCalls returns all method calls made
func (m *MockTreeProvider) GetCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	calls := make([]string, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// Reset resets the mock state
func (m *MockTreeProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.current = interfaces.NodeInfo{}
	m.nodes = make(map[string]interfaces.NodeInfo)
	m.state = interfaces.TreeState{}
	m.errors = make(map[string]error)
	m.calls = []string{}
}