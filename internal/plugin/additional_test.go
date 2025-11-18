package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taokim/muno/internal/interfaces"
)

// Test additional functions to improve coverage
func TestPluginManager_AdditionalCoverage(t *testing.T) {
	pm := &PluginManager{
		plugins:  make(map[string]*LoadedPlugin),
		commands: make(map[string]string),
	}

	// Test with loaded plugin
	mockPlugin := &DuckPlugin{
		MetadataFunc: func() interfaces.PluginMetadata {
			return interfaces.PluginMetadata{
				Name:    "test",
				Version: "1.0",
			}
		},
		CommandsFunc: func() []interfaces.CommandDefinition {
			return []interfaces.CommandDefinition{
				{Name: "cmd1", Description: "Test command"},
			}
		},
	}

	pm.plugins["test"] = &LoadedPlugin{
		Name:   "test",
		Plugin: mockPlugin,
		Metadata: interfaces.PluginMetadata{
			Name:    "test",
			Version: "1.0",
		},
	}
	pm.commands["cmd1"] = "test"

	// Test GetCommand with existing command
	cmd, plugin, err := pm.GetCommand("cmd1")
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.NotNil(t, plugin)
	assert.Equal(t, "cmd1", cmd.Name)

	// Test ListCommands with commands
	commands := pm.ListCommands()
	assert.NotNil(t, commands)
	assert.Len(t, commands, 1)

	// Test ListPlugins with plugins
	plugins := pm.ListPlugins()
	assert.NotNil(t, plugins)
	assert.Len(t, plugins, 1)

	// Test ExecuteCommand with successful execution
	mockPlugin.ExecuteFunc = func(ctx context.Context, cmd string, args []string, env interfaces.PluginEnvironment) (interfaces.Result, error) {
		return interfaces.Result{
			Success: true,
			Message: "Executed",
		}, nil
	}

	result, err := pm.ExecuteCommand(context.Background(), "cmd1", []string{})
	assert.NoError(t, err)
	assert.True(t, result.Success)
}

// Test DiscoverPlugins with actual directory scanning
func TestPluginManager_DiscoverPlugins_Coverage(t *testing.T) {
	pm := &PluginManager{
		plugins:     make(map[string]*LoadedPlugin),
		searchPaths: []string{"."}, // Current directory
	}

	// This will scan current directory for plugins
	paths, err := pm.DiscoverPlugins(context.Background())
	assert.NoError(t, err)
	// We don't care about the result, just exercising the code
	_ = paths
}


// Test getPluginMetadata directly (if possible)
func TestPluginManager_getPluginMetadata_Coverage(t *testing.T) {
	// Since getPluginMetadata is private, we test it indirectly through LoadPlugin
	pm := &PluginManager{
		plugins:     make(map[string]*LoadedPlugin),
		searchPaths: []string{},
	}

	// Try to load a non-existent plugin (will call getPluginMetadata)
	err := pm.LoadPlugin(context.Background(), "testdata/fake-plugin")
	assert.Error(t, err)
}

// Test LoadPlugin with different paths
func TestPluginManager_LoadPlugin_Coverage(t *testing.T) {
	pm := &PluginManager{
		plugins:     make(map[string]*LoadedPlugin),
		searchPaths: []string{"testdata"}, // Non-existent directory
	}

	// Try to load from various paths
	err := pm.LoadPlugin(context.Background(), "testdata/plugin")
	assert.Error(t, err) // Will fail as file doesn't exist
}

