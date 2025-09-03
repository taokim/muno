package mocks

import (
	"fmt"
	"sync"
	
	"github.com/taokim/muno/internal/interfaces"
)

// MockConfigProvider is a mock implementation of ConfigProvider
type MockConfigProvider struct {
	mu      sync.RWMutex
	configs map[string]interface{}
	exists  map[string]bool
	errors  map[string]error
	calls   []string
}

// NewMockConfigProvider creates a new mock config provider
func NewMockConfigProvider() *MockConfigProvider {
	return &MockConfigProvider{
		configs: make(map[string]interface{}),
		exists:  make(map[string]bool),
		errors:  make(map[string]error),
		calls:   []string{},
	}
}

// Load loads configuration from a file
func (m *MockConfigProvider) Load(path string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Load(%s)", path))
	
	if err, ok := m.errors[path]; ok && err != nil {
		return nil, err
	}
	
	if cfg, ok := m.configs[path]; ok {
		return cfg, nil
	}
	
	return nil, fmt.Errorf("config not found: %s", path)
}

// Save saves configuration to a file
func (m *MockConfigProvider) Save(path string, cfg interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Save(%s)", path))
	
	// Check for "save" operation error
	if err, ok := m.errors["save"]; ok && err != nil {
		return err
	}
	
	// Check for path-specific error
	if err, ok := m.errors[path]; ok && err != nil {
		return err
	}
	
	m.configs[path] = cfg
	m.exists[path] = true
	return nil
}

// Exists checks if a config file exists
func (m *MockConfigProvider) Exists(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Exists(%s)", path))
	return m.exists[path]
}

// Watch watches for configuration changes
func (m *MockConfigProvider) Watch(path string) (<-chan interfaces.ConfigEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Watch(%s)", path))
	
	ch := make(chan interfaces.ConfigEvent)
	close(ch) // Close immediately for testing
	return ch, nil
}

// Mock helper methods

// SetConfig sets a mock configuration
func (m *MockConfigProvider) SetConfig(path string, cfg interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.configs[path] = cfg
	m.exists[path] = true
}

// SetError sets an error for a specific path
func (m *MockConfigProvider) SetError(path string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.errors[path] = err
}

// SetExists sets whether a path exists
func (m *MockConfigProvider) SetExists(path string, exists bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.exists[path] = exists
}

// GetCalls returns all method calls made
func (m *MockConfigProvider) GetCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	calls := make([]string, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// Reset resets the mock state
func (m *MockConfigProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.configs = make(map[string]interface{})
	m.exists = make(map[string]bool)
	m.errors = make(map[string]error)
	m.calls = []string{}
}