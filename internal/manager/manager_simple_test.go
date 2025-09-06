package manager

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/taokim/muno/internal/interfaces"
)

// Test simple functions that don't require complex mock setup

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com/example/repo.git", "repo"},
		{"https://github.com/example/another-repo", "another-repo"},
		{"git@github.com:user/project.git", "project"},
		{"simple-name", "simple-name"},
		{"/path/to/repo", "repo"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractRepoName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultProcessProvider_SimpleOperations(t *testing.T) {
	provider := NewStubProcessProvider()
	ctx := context.Background()
	
	// Test ExecuteShell
	result, err := provider.ExecuteShell(ctx, "echo test", interfaces.ProcessOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)
	
	// Test Execute  
	result, err = provider.Execute(ctx, "echo", []string{"test"}, interfaces.ProcessOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	
	// Test OpenInBrowser
	err = provider.OpenInBrowser("https://example.com")
	assert.NoError(t, err)
	
	// Test OpenInEditor
	err = provider.OpenInEditor("/tmp/test.txt")
	assert.NoError(t, err)
}

func TestDefaultLogProvider_AllLevels(t *testing.T) {
	// Test with debug enabled
	provider := NewDefaultLogProvider(true)
	
	// These shouldn't panic
	provider.Debug("debug message")
	provider.Info("info message")
	provider.Warn("warn message")
	provider.Error("error message")
	
	// Test SetLevel
	provider.SetLevel(interfaces.LogLevelInfo)
	
	// Test WithFields
	newProvider := provider.WithFields(interfaces.Field{
		Key:   "test",
		Value: "value",
	})
	assert.NotNil(t, newProvider)
	assert.Equal(t, provider, newProvider) // Simple implementation returns self
}

func TestNoOpMetricsProvider_Operations(t *testing.T) {
	provider := NewNoOpMetricsProvider()
	
	// Test counter
	provider.Counter("test.counter", 1, "tag1", "tag2")
	
	// Test gauge
	provider.Gauge("test.gauge", 100.5, "tag1")
	
	// Test histogram
	provider.Histogram("test.histogram", 50.25, "tag1")
	
	// Test timer
	timer := provider.Timer("test.timer")
	assert.NotNil(t, timer)
	
	timer.Start()
	duration := timer.Stop()
	assert.Equal(t, int64(0), int64(duration))
	
	// Test flush
	err := provider.Flush()
	assert.NoError(t, err)
}

func TestDefaultManagerOptions_Creation(t *testing.T) {
	opts := DefaultManagerOptions()
	
	assert.NotNil(t, opts)
	assert.NotNil(t, opts.ProcessProvider)
	assert.NotNil(t, opts.LogProvider)
	assert.NotNil(t, opts.MetricsProvider)
	
	// FileSystem, Git, Config, UI should be nil (to be provided later)
	assert.Nil(t, opts.FSProvider)
	assert.Nil(t, opts.GitProvider)
	assert.Nil(t, opts.ConfigProvider)
	assert.Nil(t, opts.UIProvider)
}

func TestManager_SaveConfig_Simple(t *testing.T) {
	m := &Manager{
		config: nil, // No config, should return nil
	}
	
	err := m.saveConfig()
	assert.NoError(t, err)
}

func TestManager_Initialize_NotInitialized(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*Manager) error
	}{
		{
			"ShowCurrent",
			func(m *Manager) error { return m.ShowCurrent() },
		},
		{
			"ShowTreeAtPath",
			func(m *Manager) error { return m.ShowTreeAtPath("", 1) },
		},
		{
			"ListNodesRecursive",
			func(m *Manager) error { return m.ListNodesRecursive(false) },
		},
		{
			"StatusNode",
			func(m *Manager) error { return m.StatusNode("", false) },
		},
		{
			"PullNode",
			func(m *Manager) error { return m.PullNode("", false) },
		},
		{
			"PushNode",
			func(m *Manager) error { return m.PushNode("", false) },
		},
		{
			"CommitNode",
			func(m *Manager) error { return m.CommitNode("", "test", false) },
		},
		{
			"StartClaude",
			func(m *Manager) error { return m.StartClaude("") },
		},
		{
			"UseNodeWithClone",
			func(m *Manager) error { return m.UseNodeWithClone("", false) },
		},
		{
			"AddRepoSimple",
			func(m *Manager) error { return m.AddRepoSimple("url", "name", false) },
		},
		{
			"RemoveNode",
			func(m *Manager) error { return m.RemoveNode("name") },
		},
		{
			"CloneRepos",
			func(m *Manager) error { return m.CloneRepos("", false) },
		},
		{
			"ClearCurrent",
			func(m *Manager) error { return m.ClearCurrent() },
		},

	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				initialized: false, // Not initialized
			}
			
			err := tt.testFunc(m)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not initialized")
		})
	}
}

func TestManager_Close_Empty(t *testing.T) {
	m := &Manager{
		// No plugin manager or metrics provider
	}
	
	err := m.Close()
	assert.NoError(t, err)
}

func TestManager_ExecutePluginCommand_NoPlugin(t *testing.T) {
	m := &Manager{
		pluginManager: nil,
	}
	
	err := m.ExecutePluginCommand(context.Background(), "test", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugins not enabled")
}

func TestManager_Use_NotInitialized(t *testing.T) {
	m := &Manager{
		initialized: false,
	}
	
	err := m.Use(context.Background(), "path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestManager_Add_NotInitialized(t *testing.T) {
	m := &Manager{
		initialized: false,
	}
	
	err := m.Add(context.Background(), "url", AddOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestManager_Remove_NotInitialized(t *testing.T) {
	m := &Manager{
		initialized: false,
	}
	
	err := m.Remove(context.Background(), "name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}