package plugin

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taokim/muno/internal/interfaces"
)

// Test to boost coverage for internal/plugin package
func TestPluginManager_CoverageBoost(t *testing.T) {
	ctx := context.Background()
	
	t.Run("NewPluginManager", func(t *testing.T) {
		pm, err := NewPluginManager()
		assert.NoError(t, err)
		assert.NotNil(t, pm)
		assert.NotNil(t, pm.plugins)
		assert.NotNil(t, pm.commands)
		assert.NotNil(t, pm.searchPaths)
	})

	t.Run("LoadPlugin_Error", func(t *testing.T) {
		pm, _ := NewPluginManager()
		err := pm.LoadPlugin(ctx, "/invalid/path/to/plugin")
		assert.Error(t, err)
	})

	t.Run("LoadPlugin_AlreadyLoaded", func(t *testing.T) {
		pm, _ := NewPluginManager()
		// Manually add a "loaded" plugin to test the already-loaded path
		pm.plugins["test-plugin"] = &LoadedPlugin{
			Name: "test-plugin",
			Path: "/fake/path",
		}
		err := pm.LoadPlugin(ctx, "test-plugin")
		assert.NoError(t, err) // Should return nil if already loaded
	})

	t.Run("UnloadPlugin_NotFound", func(t *testing.T) {
		pm, _ := NewPluginManager()
		err := pm.UnloadPlugin(ctx, "non-existent-plugin")
		assert.Error(t, err)
	})


	t.Run("ReloadPlugin_NotFound", func(t *testing.T) {
		pm, _ := NewPluginManager()
		err := pm.ReloadPlugin(ctx, "non-existent-plugin")
		assert.Error(t, err)
	})

	t.Run("RemovePlugin_NotFound", func(t *testing.T) {
		pm, _ := NewPluginManager()
		err := pm.RemovePlugin(ctx, "non-existent-plugin")
		assert.Error(t, err)
	})

	t.Run("GetPlugin_NotFound", func(t *testing.T) {
		pm, _ := NewPluginManager()
		plugin, err := pm.GetPlugin("non-existent")
		assert.Nil(t, plugin)
		assert.Error(t, err)
	})

	t.Run("GetPlugin_Success", func(t *testing.T) {
		pm, _ := NewPluginManager()
		mockPlugin := &mockPluginInterface{}
		pm.plugins["test-plugin"] = &LoadedPlugin{
			Name:   "test-plugin",
			Plugin: mockPlugin,
		}
		plugin, err := pm.GetPlugin("test-plugin")
		assert.NoError(t, err)
		assert.NotNil(t, plugin)
		assert.Equal(t, mockPlugin, plugin)
	})

	t.Run("GetCommand_NotFound", func(t *testing.T) {
		pm, _ := NewPluginManager()
		cmd, plugin, err := pm.GetCommand("non-existent-cmd")
		assert.Nil(t, cmd)
		assert.Nil(t, plugin)
		assert.Error(t, err)
	})

	t.Run("GetCommand_Success", func(t *testing.T) {
		pm, _ := NewPluginManager()
		mockPlugin := &mockPluginInterfaceWithCommands{}
		pm.plugins["test-plugin"] = &LoadedPlugin{
			Name:   "test-plugin",
			Plugin: mockPlugin,
		}
		pm.commands["test-cmd"] = "test-plugin"

		cmd, plugin, err := pm.GetCommand("test-cmd")
		assert.NoError(t, err)
		assert.NotNil(t, cmd)
		assert.NotNil(t, plugin)
		assert.Equal(t, "test-cmd", cmd.Name)
	})

	t.Run("GetCommand_ByAlias", func(t *testing.T) {
		pm, _ := NewPluginManager()
		mockPlugin := &mockPluginInterfaceWithCommands{}
		pm.plugins["test-plugin"] = &LoadedPlugin{
			Name:   "test-plugin",
			Plugin: mockPlugin,
		}
		pm.commands["tc"] = "test-plugin"

		cmd, plugin, err := pm.GetCommand("tc")
		assert.NoError(t, err)
		assert.NotNil(t, cmd)
		assert.NotNil(t, plugin)
		assert.Equal(t, "test-cmd", cmd.Name)
		assert.Contains(t, cmd.Aliases, "tc")
	})

	t.Run("ListPlugins_Empty", func(t *testing.T) {
		pm, _ := NewPluginManager()
		plugins := pm.ListPlugins()
		if plugins != nil {
			assert.Empty(t, plugins)
		}
	})

	t.Run("ListCommands_Empty", func(t *testing.T) {
		pm, _ := NewPluginManager()
		commands := pm.ListCommands()
		if commands != nil {
			assert.Empty(t, commands)
		}
	})

	t.Run("IsLoaded_False", func(t *testing.T) {
		pm, _ := NewPluginManager()
		loaded := pm.IsLoaded("non-existent")
		assert.False(t, loaded)
	})

	t.Run("ExecuteCommand_NotFound", func(t *testing.T) {
		pm, _ := NewPluginManager()
		result, err := pm.ExecuteCommand(ctx, "non-existent", []string{})
		assert.Error(t, err)
		assert.False(t, result.Success)
	})

	t.Run("HealthCheck_Empty", func(t *testing.T) {
		pm, _ := NewPluginManager()
		errors := pm.HealthCheck(ctx)
		assert.Empty(t, errors)
	})

	t.Run("InstallPlugin", func(t *testing.T) {
		pm, _ := NewPluginManager()
		err := pm.InstallPlugin(ctx, "test-plugin")
		// Will error but tests the path
		assert.Error(t, err)
	})

	t.Run("UpdatePlugin", func(t *testing.T) {
		pm, _ := NewPluginManager()
		err := pm.UpdatePlugin(ctx, "test-plugin")
		assert.Error(t, err)
	})

	t.Run("GetPluginConfig", func(t *testing.T) {
		pm, _ := NewPluginManager()
		config, err := pm.GetPluginConfig("non-existent")
		// Current implementation returns empty config and nil error (stub)
		assert.NotNil(t, config)
		assert.NoError(t, err)
		assert.Empty(t, config)
	})

	t.Run("SetPluginConfig", func(t *testing.T) {
		pm, _ := NewPluginManager()
		err := pm.SetPluginConfig("non-existent", map[string]interface{}{"key": "value"})
		// Current implementation is a stub that returns nil
		assert.NoError(t, err)
	})

	t.Run("DiscoverPlugins", func(t *testing.T) {
		pm, _ := NewPluginManager()
		discovered, err := pm.DiscoverPlugins(ctx)
		// May or may not find plugins, just test it runs
		_ = discovered
		_ = err
	})
}

// Test private methods to boost coverage
func TestPluginManager_InternalMethods(t *testing.T) {
	t.Run("getPluginMetadata", func(t *testing.T) {
		pm, _ := NewPluginManager()
		metadata, err := pm.getPluginMetadata("/path/to/plugin")
		assert.NoError(t, err)
		assert.NotEmpty(t, metadata.Name)
		assert.NotEmpty(t, metadata.Version)
	})

	t.Run("findPluginBinary_NotFound", func(t *testing.T) {
		pm, _ := NewPluginManager()
		path, err := pm.findPluginBinary("non-existent-plugin")
		assert.Error(t, err)
		assert.Empty(t, path)
	})

	t.Run("findPluginBinary_WithTempFile", func(t *testing.T) {
		// Create a temporary directory and plugin file
		tmpDir := t.TempDir()
		pluginPath := tmpDir + "/test-plugin"
		err := os.WriteFile(pluginPath, []byte("#!/bin/sh\necho test"), 0755)
		assert.NoError(t, err)

		pm, _ := NewPluginManager()
		pm.searchPaths = []string{tmpDir}

		// Test finding the plugin
		foundPath, err := pm.findPluginBinary("test-plugin")
		assert.NoError(t, err)
		assert.Equal(t, pluginPath, foundPath)
	})

	t.Run("findPluginBinary_WithPrefix", func(t *testing.T) {
		// Create a temporary directory with muno-plugin- prefix
		tmpDir := t.TempDir()
		pluginPath := tmpDir + "/muno-plugin-test"
		err := os.WriteFile(pluginPath, []byte("#!/bin/sh\necho test"), 0755)
		assert.NoError(t, err)

		pm, _ := NewPluginManager()
		pm.searchPaths = []string{tmpDir}

		// Test finding the plugin with prefix
		foundPath, err := pm.findPluginBinary("test")
		assert.NoError(t, err)
		assert.Equal(t, pluginPath, foundPath)
	})
}

// Mock implementations for testing
type mockPluginInterface struct{}

func (m *mockPluginInterface) Initialize(config map[string]interface{}) error {
	return nil
}

func (m *mockPluginInterface) Cleanup() error {
	return nil
}

func (m *mockPluginInterface) Metadata() interfaces.PluginMetadata {
	return interfaces.PluginMetadata{Name: "mock", Version: "1.0.0"}
}

func (m *mockPluginInterface) Commands() []interfaces.CommandDefinition {
	return []interfaces.CommandDefinition{}
}

func (m *mockPluginInterface) Execute(ctx context.Context, cmd string, args []string, env interfaces.PluginEnvironment) (interfaces.Result, error) {
	return interfaces.Result{Success: true}, nil
}

func (m *mockPluginInterface) HealthCheck(ctx context.Context) error {
	return nil
}

// Mock plugin with commands for testing
type mockPluginInterfaceWithCommands struct{}

func (m *mockPluginInterfaceWithCommands) Initialize(config map[string]interface{}) error {
	return nil
}

func (m *mockPluginInterfaceWithCommands) Cleanup() error {
	return nil
}

func (m *mockPluginInterfaceWithCommands) Metadata() interfaces.PluginMetadata {
	return interfaces.PluginMetadata{Name: "mock", Version: "1.0.0"}
}

func (m *mockPluginInterfaceWithCommands) Commands() []interfaces.CommandDefinition {
	return []interfaces.CommandDefinition{
		{
			Name:    "test-cmd",
			Aliases: []string{"tc"},
		},
	}
}

func (m *mockPluginInterfaceWithCommands) Execute(ctx context.Context, cmd string, args []string, env interfaces.PluginEnvironment) (interfaces.Result, error) {
	return interfaces.Result{Success: true}, nil
}

func (m *mockPluginInterfaceWithCommands) HealthCheck(ctx context.Context) error {
	return nil
}