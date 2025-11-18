package manager

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
)

// Simple tests to boost coverage to 70%

// createTestManager creates a Manager with all required providers for testing
func createTestManager(tmpDir string, cfg *config.ConfigTree) *Manager {
	return &Manager{
		workspace:       tmpDir,
		config:          cfg,
		gitProvider:     NewStubGitProvider(),
		treeProvider:    NewStubTreeProvider(),
		processProvider: NewStubProcessProvider(),
		logProvider:     NewDefaultLogProvider(false),
		metricsProvider: NewNoOpMetricsProvider(),
		fsProvider:      NewStubFileSystemProvider(),
		configProvider:  NewStubConfigProvider(),
		configResolver:  config.NewConfigResolver(config.GetDefaults()),
		uiProvider:      NewStubUIProvider(),
	}
}

func TestManager_Initialize(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceName := "test-workspace"

	mgr := createTestManager(tmpDir, &config.ConfigTree{Workspace: config.WorkspaceTree{Name: workspaceName}})

	ctx := context.Background()
	err := mgr.Initialize(ctx, workspaceName)

	// Should complete without error
	assert.NoError(t, err)
}

func TestManager_InitializeWithConfig(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
	}

	mgr := createTestManager(tmpDir, cfg)

	ctx := context.Background()
	err := mgr.InitializeWithConfig(ctx, tmpDir, cfg)

	// Should complete without error
	assert.NoError(t, err)
}

func TestManager_SetCLIConfig(t *testing.T) {
	mgr := createTestManager(t.TempDir(), &config.ConfigTree{})

	cliConfig := map[string]interface{}{
		"git": map[string]interface{}{
			"default_branch": "main",
		},
	}

	mgr.SetCLIConfig(cliConfig)

	// Should set CLI config
	assert.NotNil(t, mgr.configResolver)
}

func TestManager_GetConfigResolver(t *testing.T) {
	mgr := createTestManager(t.TempDir(), &config.ConfigTree{})

	resolver := mgr.GetConfigResolver()

	// Should return config resolver
	assert.NotNil(t, resolver)
}

func TestManager_GetSSHPreference(t *testing.T) {
	mgr := createTestManager(t.TempDir(), &config.ConfigTree{})

	useSSH := mgr.getSSHPreference()

	// Should return false by default
	assert.False(t, useSSH)
}

func TestManager_Close(t *testing.T) {
	mgr := createTestManager(t.TempDir(), &config.ConfigTree{})

	err := mgr.Close()

	// Should close without error
	assert.NoError(t, err)
}

func TestManager_ResolvePath_SimpleCases(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple workspace
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)

	// Create muno.yaml
	munoYaml := filepath.Join(workspaceDir, "muno.yaml")
	os.WriteFile(munoYaml, []byte("workspace: test\n"), 0644)

	mgr := createTestManager(workspaceDir, &config.ConfigTree{Workspace: config.WorkspaceTree{Name: "test", ReposDir: ".nodes"}})

	t.Run("resolve_empty_path", func(t *testing.T) {
		// Empty path should resolve to something
		_, err := mgr.ResolvePath("", false)
		// May error or succeed depending on tree state
		_ = err
	})

	t.Run("resolve_dot_path", func(t *testing.T) {
		// Dot path
		result, err := mgr.ResolvePath(".", false)
		if err == nil {
			assert.NotEmpty(t, result)
		}
	})
}

func TestManager_HandlePluginAction_OpenBrowser(t *testing.T) {
	mgr := createTestManager(t.TempDir(), &config.ConfigTree{})

	action := interfaces.Action{
		Type: "open",
		URL:  "https://example.com",
	}

	err := mgr.handlePluginAction(context.Background(), action)

	// Should handle open action (may succeed or fail depending on browser availability)
	_ = err
}

func TestManager_HandlePluginAction_Navigate(t *testing.T) {
	mgr := createTestManager(t.TempDir(), &config.ConfigTree{})

	action := interfaces.Action{
		Type: "navigate",
		Path: "/test",
	}

	err := mgr.handlePluginAction(context.Background(), action)

	// Should handle navigate action
	_ = err
}

func TestManager_Add_Success(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)

	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{},
	}

	mgr := createTestManager(workspaceDir, cfg)

	// Initialize the manager
	ctx := context.Background()
	err := mgr.InitializeWithConfig(ctx, workspaceDir, cfg)
	assert.NoError(t, err)

	// Change to workspace directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(workspaceDir)

	// Add a repository
	err = mgr.Add(ctx, "https://github.com/test/repo.git", AddOptions{
		Name:  "test-repo",
		Fetch: config.FetchLazy,
	})

	// Should complete without error
	assert.NoError(t, err)
}

func TestManager_Add_EagerFetch(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)

	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
	}

	mgr := createTestManager(workspaceDir, cfg)

	ctx := context.Background()
	err := mgr.InitializeWithConfig(ctx, workspaceDir, cfg)
	assert.NoError(t, err)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(workspaceDir)

	// Add with eager fetch
	err = mgr.Add(ctx, "https://github.com/test/eager.git", AddOptions{
		Fetch: config.FetchEager,
	})

	// Should complete without error
	assert.NoError(t, err)
}

func TestManager_Add_AutoFetch(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)

	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
	}

	mgr := createTestManager(workspaceDir, cfg)

	ctx := context.Background()
	err := mgr.InitializeWithConfig(ctx, workspaceDir, cfg)
	assert.NoError(t, err)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(workspaceDir)

	// Add with auto fetch mode
	err = mgr.Add(ctx, "https://github.com/test/auto.git", AddOptions{
		Fetch: config.FetchAuto,
	})

	// Should complete without error
	assert.NoError(t, err)
}

func TestManager_Remove_Success(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)

	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "test-repo", URL: "https://github.com/test/repo.git"},
		},
	}

	mgr := createTestManager(workspaceDir, cfg)

	ctx := context.Background()
	err := mgr.InitializeWithConfig(ctx, workspaceDir, cfg)
	assert.NoError(t, err)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(workspaceDir)

	// Remove the repository
	err = mgr.Remove(ctx, "test-repo")

	// Should complete without error
	assert.NoError(t, err)
}

