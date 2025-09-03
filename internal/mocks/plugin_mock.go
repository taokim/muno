package mocks

import (
	"context"
	"sync"
	
	"github.com/taokim/muno/internal/interfaces"
)

// MockPluginManager is a mock implementation of PluginManager
type MockPluginManager struct {
	mu       sync.RWMutex
	plugins  map[string]interfaces.Plugin
	results  map[string]interfaces.Result
	errors   map[string]error
	calls    []string
}

// NewMockPluginManager creates a new mock plugin manager
func NewMockPluginManager() *MockPluginManager {
	return &MockPluginManager{
		plugins: make(map[string]interfaces.Plugin),
		results: make(map[string]interfaces.Result),
		errors:  make(map[string]error),
		calls:   []string{},
	}
}

// DiscoverPlugins discovers available plugins
func (m *MockPluginManager) DiscoverPlugins(ctx context.Context) ([]interfaces.PluginMetadata, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "DiscoverPlugins()")
	
	if err, ok := m.errors["discover"]; ok && err != nil {
		return nil, err
	}
	
	var metadata []interfaces.PluginMetadata
	for _, p := range m.plugins {
		metadata = append(metadata, p.Metadata())
	}
	
	return metadata, nil
}

// LoadPlugin loads a specific plugin
func (m *MockPluginManager) LoadPlugin(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "LoadPlugin("+name+")")
	
	if err, ok := m.errors["load:"+name]; ok && err != nil {
		return err
	}
	
	return nil
}

// ExecuteCommand executes a plugin command
func (m *MockPluginManager) ExecuteCommand(ctx context.Context, name string, args []string) (interfaces.Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "ExecuteCommand("+name+")")
	
	if err, ok := m.errors[name]; ok && err != nil {
		return interfaces.Result{}, err
	}
	
	if result, ok := m.results[name]; ok {
		return result, nil
	}
	
	return interfaces.Result{
		Success: true,
		Message: "Command executed",
	}, nil
}

// ListPlugins returns available plugins
func (m *MockPluginManager) ListPlugins() []interfaces.PluginMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, "ListPlugins()")
	
	var plugins []interfaces.PluginMetadata
	for _, p := range m.plugins {
		plugins = append(plugins, p.Metadata())
	}
	
	return plugins
}

// GetPlugin returns a specific plugin
func (m *MockPluginManager) GetPlugin(name string) (interfaces.Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, "GetPlugin("+name+")")
	
	if p, ok := m.plugins[name]; ok {
		return p, nil
	}
	
	return nil, nil
}

// UnloadPlugin unloads a plugin
func (m *MockPluginManager) UnloadPlugin(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "UnloadPlugin("+name+")")
	
	delete(m.plugins, name)
	
	return nil
}

// ReloadPlugin reloads a plugin
func (m *MockPluginManager) ReloadPlugin(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "ReloadPlugin("+name+")")
	
	return nil
}

// IsLoaded checks if a plugin is loaded
func (m *MockPluginManager) IsLoaded(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	_, ok := m.plugins[name]
	return ok
}

// GetCommand gets a command definition
func (m *MockPluginManager) GetCommand(name string) (*interfaces.CommandDefinition, interfaces.Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, p := range m.plugins {
		for _, cmd := range p.Commands() {
			if cmd.Name == name {
				return &cmd, p, nil
			}
		}
	}
	
	return nil, nil, nil
}

// ListCommands lists all commands
func (m *MockPluginManager) ListCommands() []interfaces.CommandDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var commands []interfaces.CommandDefinition
	for _, p := range m.plugins {
		commands = append(commands, p.Commands()...)
	}
	
	return commands
}

// InstallPlugin installs a plugin
func (m *MockPluginManager) InstallPlugin(ctx context.Context, source string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "InstallPlugin("+source+")")
	
	return nil
}

// UpdatePlugin updates a plugin
func (m *MockPluginManager) UpdatePlugin(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "UpdatePlugin("+name+")")
	
	return nil
}

// RemovePlugin removes a plugin
func (m *MockPluginManager) RemovePlugin(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "RemovePlugin("+name+")")
	
	delete(m.plugins, name)
	
	return nil
}

// GetPluginConfig gets plugin configuration
func (m *MockPluginManager) GetPluginConfig(name string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return map[string]interface{}{}, nil
}

// SetPluginConfig sets plugin configuration
func (m *MockPluginManager) SetPluginConfig(name string, config map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "SetPluginConfig("+name+")")
	
	return nil
}

// Mock helper methods

// SetPlugin adds a mock plugin
func (m *MockPluginManager) SetPlugin(name string, plugin interfaces.Plugin) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.plugins[name] = plugin
}

// SetCommandResult sets the result for a command
func (m *MockPluginManager) SetCommandResult(name string, result interfaces.Result) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.results[name] = result
}

// SetCommandError sets an error for a command
func (m *MockPluginManager) SetCommandError(name string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.errors[name] = err
}

// SetError sets an error for an operation
func (m *MockPluginManager) SetError(operation string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.errors[operation] = err
}

// GetCalls returns all method calls made
func (m *MockPluginManager) GetCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	calls := make([]string, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// Reset resets the mock state
func (m *MockPluginManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.plugins = make(map[string]interfaces.Plugin)
	m.results = make(map[string]interfaces.Result)
	m.errors = make(map[string]error)
	m.calls = []string{}
}
// HealthCheck checks health of all plugins
func (m *MockPluginManager) HealthCheck(ctx context.Context) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, "HealthCheck()")
	
	result := make(map[string]error)
	for name := range m.plugins {
		if err, ok := m.errors["health:"+name]; ok {
			result[name] = err
		}
	}
	
	return result
}
