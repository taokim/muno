package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	
	"github.com/hashicorp/go-plugin"
	"github.com/taokim/muno/internal/interfaces"
)

// PluginManager manages plugin lifecycle
type PluginManager struct {
	mu          sync.RWMutex
	plugins     map[string]*LoadedPlugin
	commands    map[string]string // command -> plugin name mapping
	searchPaths []string
	config      *PluginConfig
}

// LoadedPlugin represents a loaded plugin
type LoadedPlugin struct {
	Name     string
	Path     string
	Plugin   interfaces.Plugin
	Client   *plugin.Client
	Metadata interfaces.PluginMetadata
}

// PluginConfig holds plugin manager configuration
type PluginConfig struct {
	PluginPaths []string
	Registry    string
	AutoLoad    bool
}

// HandshakeConfig is the handshake configuration for plugins
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "MUNO_PLUGIN",
	MagicCookieValue: "octopus",
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() (*PluginManager, error) {
	// Default search paths
	searchPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".muno", "plugins"),
		"/usr/local/lib/muno/plugins",
		"./plugins",
	}
	
	// Add paths from environment
	if envPaths := os.Getenv("MUNO_PLUGIN_PATH"); envPaths != "" {
		searchPaths = append(searchPaths, filepath.SplitList(envPaths)...)
	}
	
	return &PluginManager{
		plugins:     make(map[string]*LoadedPlugin),
		commands:    make(map[string]string),
		searchPaths: searchPaths,
		config:      &PluginConfig{},
	}, nil
}

// DiscoverPlugins discovers available plugins
func (pm *PluginManager) DiscoverPlugins(ctx context.Context) ([]interfaces.PluginMetadata, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	var discovered []interfaces.PluginMetadata
	
	for _, searchPath := range pm.searchPaths {
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			continue
		}
		
		// Look for plugin binaries
		entries, err := os.ReadDir(searchPath)
		if err != nil {
			continue
		}
		
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			
			// Check if it's an executable plugin
			pluginPath := filepath.Join(searchPath, entry.Name())
			if info, err := os.Stat(pluginPath); err == nil && info.Mode()&0111 != 0 {
				// Try to load plugin metadata
				if metadata, err := pm.getPluginMetadata(pluginPath); err == nil {
					discovered = append(discovered, metadata)
				}
			}
		}
	}
	
	return discovered, nil
}

// LoadPlugin loads a specific plugin
func (pm *PluginManager) LoadPlugin(ctx context.Context, name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// Check if already loaded
	if _, exists := pm.plugins[name]; exists {
		return nil
	}
	
	// Find plugin binary
	pluginPath, err := pm.findPluginBinary(name)
	if err != nil {
		return fmt.Errorf("plugin not found: %s", name)
	}
	
	// Create plugin client
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			name: &MunoPlugin{},
		},
		Cmd: exec.Command(pluginPath),
	})
	
	// Connect to plugin
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to connect to plugin: %w", err)
	}
	
	// Get plugin instance
	raw, err := rpcClient.Dispense(name)
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}
	
	pluginInstance, ok := raw.(interfaces.Plugin)
	if !ok {
		client.Kill()
		return fmt.Errorf("invalid plugin type")
	}
	
	// Initialize plugin
	if err := pluginInstance.Initialize(make(map[string]interface{})); err != nil {
		client.Kill()
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}
	
	// Get metadata
	metadata := pluginInstance.Metadata()
	
	// Store loaded plugin
	loaded := &LoadedPlugin{
		Name:     name,
		Path:     pluginPath,
		Plugin:   pluginInstance,
		Client:   client,
		Metadata: metadata,
	}
	pm.plugins[name] = loaded
	
	// Register commands
	for _, cmd := range pluginInstance.Commands() {
		pm.commands[cmd.Name] = name
		for _, alias := range cmd.Aliases {
			pm.commands[alias] = name
		}
	}
	
	return nil
}

// UnloadPlugin unloads a plugin
func (pm *PluginManager) UnloadPlugin(ctx context.Context, name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	loaded, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not loaded: %s", name)
	}
	
	// Cleanup plugin
	if err := loaded.Plugin.Cleanup(); err != nil {
		// Log error but continue
		fmt.Fprintf(os.Stderr, "Warning: plugin cleanup failed: %v\n", err)
	}
	
	// Kill client
	loaded.Client.Kill()
	
	// Remove from registry
	delete(pm.plugins, name)
	
	// Remove command mappings
	for cmd, pluginName := range pm.commands {
		if pluginName == name {
			delete(pm.commands, cmd)
		}
	}
	
	return nil
}

// ReloadPlugin reloads a plugin
func (pm *PluginManager) ReloadPlugin(ctx context.Context, name string) error {
	if err := pm.UnloadPlugin(ctx, name); err != nil {
		return err
	}
	return pm.LoadPlugin(ctx, name)
}

// GetPlugin gets a loaded plugin
func (pm *PluginManager) GetPlugin(name string) (interfaces.Plugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	if loaded, exists := pm.plugins[name]; exists {
		return loaded.Plugin, nil
	}
	
	return nil, fmt.Errorf("plugin not found: %s", name)
}

// ListPlugins lists loaded plugins
func (pm *PluginManager) ListPlugins() []interfaces.PluginMetadata {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	var plugins []interfaces.PluginMetadata
	for _, loaded := range pm.plugins {
		plugins = append(plugins, loaded.Metadata)
	}
	
	return plugins
}

// IsLoaded checks if a plugin is loaded
func (pm *PluginManager) IsLoaded(name string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	_, exists := pm.plugins[name]
	return exists
}

// GetCommand gets command definition and plugin
func (pm *PluginManager) GetCommand(name string) (*interfaces.CommandDefinition, interfaces.Plugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	pluginName, exists := pm.commands[name]
	if !exists {
		return nil, nil, fmt.Errorf("command not found: %s", name)
	}
	
	loaded, exists := pm.plugins[pluginName]
	if !exists {
		return nil, nil, fmt.Errorf("plugin not loaded: %s", pluginName)
	}
	
	// Find command definition
	for _, cmd := range loaded.Plugin.Commands() {
		if cmd.Name == name {
			return &cmd, loaded.Plugin, nil
		}
		for _, alias := range cmd.Aliases {
			if alias == name {
				return &cmd, loaded.Plugin, nil
			}
		}
	}
	
	return nil, nil, fmt.Errorf("command definition not found: %s", name)
}

// ListCommands lists all available commands
func (pm *PluginManager) ListCommands() []interfaces.CommandDefinition {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	var commands []interfaces.CommandDefinition
	seen := make(map[string]bool)
	
	for _, loaded := range pm.plugins {
		for _, cmd := range loaded.Plugin.Commands() {
			if !seen[cmd.Name] {
				commands = append(commands, cmd)
				seen[cmd.Name] = true
			}
		}
	}
	
	return commands
}

// ExecuteCommand executes a plugin command
func (pm *PluginManager) ExecuteCommand(ctx context.Context, name string, args []string) (interfaces.Result, error) {
	cmdDef, plugin, err := pm.GetCommand(name)
	if err != nil {
		return interfaces.Result{}, err
	}
	
	// Build environment
	cwd, _ := os.Getwd()
	env := interfaces.PluginEnvironment{
		WorkspacePath: cwd,
		Variables:     make(map[string]string),
		Config:        make(map[string]interface{}),
	}
	
	// Execute command
	return plugin.Execute(ctx, cmdDef.Name, args, env)
}

// InstallPlugin installs a plugin from source
func (pm *PluginManager) InstallPlugin(ctx context.Context, source string) error {
	// This would download and install a plugin
	// For now, just return not implemented
	return fmt.Errorf("plugin installation not implemented")
}

// UpdatePlugin updates a plugin
func (pm *PluginManager) UpdatePlugin(ctx context.Context, name string) error {
	// This would update an installed plugin
	// For now, just return not implemented
	return fmt.Errorf("plugin update not implemented")
}

// RemovePlugin removes an installed plugin
func (pm *PluginManager) RemovePlugin(ctx context.Context, name string) error {
	// Unload if loaded
	if pm.IsLoaded(name) {
		if err := pm.UnloadPlugin(ctx, name); err != nil {
			return err
		}
	}
	
	// Remove plugin binary
	pluginPath, err := pm.findPluginBinary(name)
	if err != nil {
		return err
	}
	
	return os.Remove(pluginPath)
}

// GetPluginConfig gets plugin configuration
func (pm *PluginManager) GetPluginConfig(name string) (map[string]interface{}, error) {
	// This would return plugin-specific configuration
	// For now, return empty config
	return make(map[string]interface{}), nil
}

// SetPluginConfig sets plugin configuration
func (pm *PluginManager) SetPluginConfig(name string, config map[string]interface{}) error {
	// This would update plugin-specific configuration
	// For now, just return nil
	return nil
}

// HealthCheck performs health check on all plugins
func (pm *PluginManager) HealthCheck(ctx context.Context) map[string]error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	results := make(map[string]error)
	
	for name, loaded := range pm.plugins {
		if err := loaded.Plugin.HealthCheck(ctx); err != nil {
			results[name] = err
		}
	}
	
	return results
}

// Helper methods

func (pm *PluginManager) findPluginBinary(name string) (string, error) {
	for _, searchPath := range pm.searchPaths {
		// Try exact name
		pluginPath := filepath.Join(searchPath, name)
		if info, err := os.Stat(pluginPath); err == nil && !info.IsDir() {
			return pluginPath, nil
		}
		
		// Try with muno-plugin- prefix
		pluginPath = filepath.Join(searchPath, "muno-plugin-"+name)
		if info, err := os.Stat(pluginPath); err == nil && !info.IsDir() {
			return pluginPath, nil
		}
		
		// Try with .exe extension on Windows
		if runtime.GOOS == "windows" {
			pluginPath = filepath.Join(searchPath, name+".exe")
			if info, err := os.Stat(pluginPath); err == nil && !info.IsDir() {
				return pluginPath, nil
			}
		}
	}
	
	return "", fmt.Errorf("plugin binary not found: %s", name)
}

func (pm *PluginManager) getPluginMetadata(path string) (interfaces.PluginMetadata, error) {
	// This would load plugin just to get metadata
	// For now, return a mock metadata
	return interfaces.PluginMetadata{
		Name:    filepath.Base(path),
		Version: "1.0.0",
	}, nil
}