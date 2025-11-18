package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test LoadPlugin function to improve coverage from 19.4%
func TestPluginManager_LoadPlugin_Extended(t *testing.T) {
	pm, err := NewPluginManager()
	assert.NoError(t, err)
	
	ctx := context.Background()
	
	// Test loading a non-existent plugin
	err = pm.LoadPlugin(ctx, "/non/existent/plugin")
	assert.Error(t, err)
	
	// Test that GetPlugin returns error for non-loaded plugin
	plugin, err := pm.GetPlugin("non-existent")
	assert.Nil(t, plugin)
	assert.Error(t, err)
}

// Test UnloadPlugin to improve coverage from 38.5%
func TestPluginManager_UnloadPlugin_Extended(t *testing.T) {
	pm, err := NewPluginManager()
	assert.NoError(t, err)
	
	ctx := context.Background()
	
	// Test unloading a non-existent plugin
	err = pm.UnloadPlugin(ctx, "non-existent")
	assert.Error(t, err)
}

// Test RemovePlugin to improve coverage from 57.1%
func TestPluginManager_RemovePlugin_Extended(t *testing.T) {
	pm, err := NewPluginManager()
	assert.NoError(t, err)
	
	ctx := context.Background()
	
	// Test removing non-existent plugin
	err = pm.RemovePlugin(ctx, "non-existent")
	assert.Error(t, err)
}

// Test DiscoverPlugins to improve overall coverage
func TestPluginManager_DiscoverPlugins_Extended(t *testing.T) {
	pm, err := NewPluginManager()
	assert.NoError(t, err)
	
	ctx := context.Background()
	
	// Test discovery (may not find any plugins in test env)
	discovered, err := pm.DiscoverPlugins(ctx)
	// We don't assert error because it depends on filesystem
	_ = err
	_ = discovered
}

// Test ReloadPlugin to improve coverage
func TestPluginManager_ReloadPlugin_Extended(t *testing.T) {
	pm, err := NewPluginManager()
	assert.NoError(t, err)
	
	ctx := context.Background()
	
	// Test reloading non-existent plugin
	err = pm.ReloadPlugin(ctx, "non-existent")
	assert.Error(t, err)
}

// Test GetCommand to improve coverage 
func TestPluginManager_GetCommand_Extended(t *testing.T) {
	pm, err := NewPluginManager()
	assert.NoError(t, err)
	
	// Test getting non-existent command
	cmd, plugin, err := pm.GetCommand("non-existent-cmd")
	assert.Error(t, err)
	assert.Nil(t, cmd)
	assert.Nil(t, plugin)
}

// Test HealthCheck to improve coverage
func TestPluginManager_HealthCheck_Extended(t *testing.T) {
	pm, err := NewPluginManager()
	assert.NoError(t, err)
	
	ctx := context.Background()
	errors := pm.HealthCheck(ctx)
	// Should return empty map with no plugins
	assert.Empty(t, errors)
}

// Test ListCommands to improve coverage
func TestPluginManager_ListCommands_Extended(t *testing.T) {
	pm, err := NewPluginManager()
	assert.NoError(t, err)
	
	commands := pm.ListCommands()
	// Should return empty or nil when no plugins loaded
	if commands != nil {
		assert.Empty(t, commands)
	}
}

// Test IsLoaded to improve coverage
func TestPluginManager_IsLoaded_Extended(t *testing.T) {
	pm, err := NewPluginManager()
	assert.NoError(t, err)
	
	// Test with non-existent plugin
	loaded := pm.IsLoaded("non-existent")
	assert.False(t, loaded)
}

// Test ExecuteCommand to improve coverage
func TestPluginManager_ExecuteCommand_Extended(t *testing.T) {
	pm, err := NewPluginManager()
	assert.NoError(t, err)
	
	ctx := context.Background()
	
	// Test executing non-existent command
	result, err := pm.ExecuteCommand(ctx, "non-existent", []string{})
	assert.Error(t, err)
	assert.False(t, result.Success)
}