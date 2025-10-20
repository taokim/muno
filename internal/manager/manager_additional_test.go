package manager

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/mocks"
)

// Test computeFilesystemPath
func TestComputeFilesystemPath(t *testing.T) {
	mockFS := mocks.NewMockFileSystemProvider()
	// Get the actual default nodes directory from config
	nodesDir := config.GetDefaultNodesDir()
	
	mgr := &Manager{
		workspace: "/test/workspace",
		fsProvider: mockFS,
		config: &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				ReposDir: nodesDir,
			},
		},
	}
	
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "root path",
			path: "/",
			want: filepath.Join("/test/workspace", nodesDir),
		},
		{
			name: "empty path",
			path: "",
			want: filepath.Join("/test/workspace", nodesDir),
		},
		{
			name: "top-level repo",
			path: "/backend",
			want: filepath.Join("/test/workspace", nodesDir, "backend"),
		},
		{
			name: "nested path without git",
			path: "/backend/service",
			want: filepath.Join("/test/workspace", nodesDir, "backend", "service"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mgr.computeFilesystemPath(tt.path)
			assert.Equal(t, tt.want, result)
		})
	}
}

// Test getReposDir
func TestGetReposDir(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		mgr := &Manager{
			config: &config.ConfigTree{
				Workspace: config.WorkspaceTree{
					ReposDir: "custom-repos",
				},
			},
		}
		assert.Equal(t, "custom-repos", mgr.getReposDir())
	})
	
	t.Run("without config", func(t *testing.T) {
		mgr := &Manager{}
		// Get the actual default nodes directory from config
		expected := config.GetDefaultNodesDir()
		assert.Equal(t, expected, mgr.getReposDir())
	})
}

// Test extractRepoName is already in manager_simple_test.go

// Test displayTreeRecursive
func TestDisplayTreeRecursive(t *testing.T) {
	mockUI := mocks.NewMockUIProvider()
	
	mgr := &Manager{
		uiProvider: mockUI,
		config: &config.ConfigTree{
			Nodes: []config.NodeDefinition{
				{Name: "backend", URL: "https://github.com/test/backend.git"},
			},
		},
	}
	
	node := interfaces.NodeInfo{
		Name: "root",
		Path: "/",
		Children: []interfaces.NodeInfo{
			{
				Name:       "backend",
				Path:       "/backend",
				Repository: "https://github.com/test/backend.git",
				IsCloned:   true,
			},
			{
				Name:       "frontend",
				Path:       "/frontend",
				Repository: "https://github.com/test/frontend.git",
				IsLazy:     true,
				IsCloned:   false,
			},
		},
	}
	
	// This should not panic
	mgr.displayTreeRecursive(node, 0)
}

// Test countNodes
func TestCountNodes(t *testing.T) {
	mgr := &Manager{}
	
	node := interfaces.NodeInfo{
		Name: "root",
		Children: []interfaces.NodeInfo{
			{
				Name:     "backend",
				IsCloned: true,
			},
			{
				Name:     "frontend",
				IsLazy:   true,
				IsCloned: false,
			},
			{
				Name:     "service",
				IsCloned: true,
				Children: []interfaces.NodeInfo{
					{
						Name:     "auth",
						IsCloned: true,
					},
				},
			},
		},
	}
	
	total, cloned, lazy := mgr.countNodes(node)
	assert.Equal(t, 5, total) // root + 4 children
	assert.Equal(t, 3, cloned) // backend, service, auth
	assert.Equal(t, 1, lazy) // frontend
}

// Test saveConfig
func TestSaveConfig(t *testing.T) {
	t.Run("no config", func(t *testing.T) {
		mgr := &Manager{}
		err := mgr.saveConfig()
		assert.NoError(t, err)
	})
	
	t.Run("with config", func(t *testing.T) {
		mockConfig := mocks.NewMockConfigProvider()
		
		mgr := &Manager{
			workspace: "/test/workspace",
			config: &config.ConfigTree{
				Workspace: config.WorkspaceTree{
					Name: "test",
				},
			},
			configProvider: mockConfig,
		}
		
		err := mgr.saveConfig()
		assert.NoError(t, err)
	})
}

// Test Close
func TestClose(t *testing.T) {
	t.Run("with nil providers", func(t *testing.T) {
		mgr := &Manager{}
		err := mgr.Close()
		assert.NoError(t, err)
	})
	
	t.Run("with providers", func(t *testing.T) {
		mgr := &Manager{
			logProvider:     NewDefaultLogProvider(false),
			metricsProvider: NewNoOpMetricsProvider(),
		}
		err := mgr.Close()
		assert.NoError(t, err)
	})
}

// Test handlePluginAction
func TestHandlePluginActionMore(t *testing.T) {
	t.Run("command action", func(t *testing.T) {
		mgr := &Manager{}
		action := interfaces.Action{
			Type:      "command",
			Command:   "test",
			Arguments: []string{"arg1"},
		}
		// Should not panic even without plugin manager
		err := mgr.handlePluginAction(context.Background(), action)
		assert.Error(t, err)
	})
	
	t.Run("prompt action", func(t *testing.T) {
		mockUI := mocks.NewMockUIProvider()
		// Mock will return empty response by default
		
		mgr := &Manager{
			uiProvider:  mockUI,
			logProvider: NewDefaultLogProvider(false),
		}
		
		action := interfaces.Action{
			Type:    "prompt",
			Message: "Enter value:",
		}
		
		err := mgr.handlePluginAction(context.Background(), action)
		assert.NoError(t, err)
	})
}

// Test DefaultManagerOptions
func TestDefaultManagerOptions(t *testing.T) {
	opts := DefaultManagerOptions()
	assert.NotNil(t, opts)
	assert.NotNil(t, opts.ProcessProvider)
	assert.NotNil(t, opts.LogProvider)
	assert.NotNil(t, opts.MetricsProvider)
	assert.Nil(t, opts.FSProvider)
	assert.Nil(t, opts.GitProvider)
}

// Test DefaultProcessProvider
func TestDefaultProcessProvider(t *testing.T) {
	p := NewDefaultProcessProvider()
	assert.NotNil(t, p)
	
	// Test stub provider
	stub := NewStubProcessProvider()
	assert.NotNil(t, stub)
	
	// Test methods don't panic
	ctx := context.Background()
	result, err := stub.ExecuteShell(ctx, "echo test", interfaces.ProcessOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	
	err = stub.OpenInBrowser("https://example.com")
	assert.NoError(t, err)
	
	err = stub.OpenInEditor("/path/to/file")
	assert.NoError(t, err)
	
	proc, err := stub.StartBackground(ctx, "test", []string{}, interfaces.ProcessOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, proc)
}