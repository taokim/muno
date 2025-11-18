package manager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
)

// TestNewManager_MissingConfigProvider tests NewManager with missing ConfigProvider
func TestNewManager_MissingConfigProvider(t *testing.T) {
	opts := ManagerOptions{
		ConfigProvider: nil,
	}

	_, err := NewManager(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ConfigProvider is required")
}

// TestNewManager_MissingGitProvider tests NewManager with missing GitProvider
func TestNewManager_MissingGitProvider(t *testing.T) {
	workspace := t.TempDir()
	opts := CreateTestManagerOptions(workspace)
	opts.GitProvider = nil

	_, err := NewManager(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitProvider is required")
}

// TestNewManager_MissingFSProvider tests NewManager with missing FSProvider
func TestNewManager_MissingFSProvider(t *testing.T) {
	workspace := t.TempDir()
	opts := CreateTestManagerOptions(workspace)
	opts.FSProvider = nil

	_, err := NewManager(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FSProvider is required")
}

// TestNewManager_MissingUIProvider tests NewManager with missing UIProvider
func TestNewManager_MissingUIProvider(t *testing.T) {
	workspace := t.TempDir()
	opts := CreateTestManagerOptions(workspace)
	opts.UIProvider = nil

	_, err := NewManager(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UIProvider is required")
}

// TestNewManager_MissingTreeProvider tests NewManager with missing TreeProvider
func TestNewManager_MissingTreeProvider(t *testing.T) {
	workspace := t.TempDir()
	opts := CreateTestManagerOptions(workspace)
	opts.TreeProvider = nil

	_, err := NewManager(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TreeProvider is required")
}

// TestNewManager_WithAllProviders tests NewManager with all required providers
func TestNewManager_WithAllProviders(t *testing.T) {
	workspace := t.TempDir()
	opts := CreateTestManagerOptions(workspace)

	mgr, err := NewManager(opts)
	assert.NoError(t, err)
	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr.configProvider)
	assert.NotNil(t, mgr.gitProvider)
	assert.NotNil(t, mgr.fsProvider)
	assert.NotNil(t, mgr.uiProvider)
	assert.NotNil(t, mgr.treeProvider)
}

// TestNewManager_WithOptionalProviders tests NewManager uses default providers
func TestNewManager_WithOptionalProviders(t *testing.T) {
	workspace := t.TempDir()
	opts := CreateTestManagerOptions(workspace)

	// Don't set optional providers
	opts.ProcessProvider = nil
	opts.LogProvider = nil
	opts.MetricsProvider = nil

	mgr, err := NewManager(opts)
	assert.NoError(t, err)
	assert.NotNil(t, mgr)

	// Verify default providers were created
	assert.NotNil(t, mgr.processProvider)
	assert.NotNil(t, mgr.logProvider)
	assert.NotNil(t, mgr.metricsProvider)
}

// TestNewManager_WithDebugMode tests NewManager with debug mode enabled
func TestNewManager_WithDebugMode(t *testing.T) {
	workspace := t.TempDir()
	opts := CreateTestManagerOptions(workspace)
	opts.DebugMode = true
	opts.LogProvider = nil // Let it create default with debug mode

	mgr, err := NewManager(opts)
	assert.NoError(t, err)
	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr.logProvider)
}

// TestClose_BasicCleanup tests Close performs basic cleanup
func TestClose_BasicCleanup(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	err := m.Close()
	assert.NoError(t, err)
}

// TestClose_WithNilLogProvider tests Close when logProvider is nil
func TestClose_WithNilLogProvider(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)
	m.logProvider = nil

	err := m.Close()
	assert.NoError(t, err)
}

// TestClose_WithNilPluginManager tests Close when pluginManager is nil
func TestClose_WithNilPluginManager(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)
	m.pluginManager = nil

	err := m.Close()
	assert.NoError(t, err)
}

// TestClose_WithNilMetricsProvider tests Close when metricsProvider is nil
func TestClose_WithNilMetricsProvider(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)
	m.metricsProvider = nil

	err := m.Close()
	assert.NoError(t, err)
}

// TestInitializeWithConfig_Success tests successful initialization with config
func TestInitializeWithConfig_Success(t *testing.T) {
	workspace := t.TempDir()

	opts := CreateTestManagerOptions(workspace)
	m, err := NewManager(opts)
	require.NoError(t, err)

	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Fetch: "eager"},
		},
	}

	ctx := context.Background()
	err = m.InitializeWithConfig(ctx, workspace, cfg)
	assert.NoError(t, err)
	assert.True(t, m.initialized)
	assert.Equal(t, cfg, m.config)
}

// TestInitializeWithConfig_WithExistingRepo tests initialization when workspace repo exists
func TestInitializeWithConfig_WithExistingRepo(t *testing.T) {
	workspace := t.TempDir()

	opts := CreateTestManagerOptions(workspace)
	m, err := NewManager(opts)
	require.NoError(t, err)

	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
			RootRepo: "https://github.com/test/root.git",
		},
		Nodes: []config.NodeDefinition{},
	}

	ctx := context.Background()
	err = m.InitializeWithConfig(ctx, workspace, cfg)
	assert.NoError(t, err)
	assert.True(t, m.initialized)
}

// TestInitializeWithConfig_LoadTreeFailure tests initialization when tree loading fails
func TestInitializeWithConfig_LoadTreeFailure(t *testing.T) {
	workspace := t.TempDir()

	opts := CreateTestManagerOptions(workspace)
	m, err := NewManager(opts)
	require.NoError(t, err)

	// Create an invalid config that will cause tree loading to fail
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "", // Invalid - empty name
			ReposDir: ".nodes",
		},
	}

	// Create a tree adapter that will fail on Load
	m.treeProvider = &FailingTreeAdapter{}

	ctx := context.Background()
	err = m.InitializeWithConfig(ctx, workspace, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load tree")
}

// TestInitializeWithConfig_SaveConfigFailure tests initialization when config saving fails
func TestInitializeWithConfig_SaveConfigFailure(t *testing.T) {
	workspace := t.TempDir()

	opts := CreateTestManagerOptions(workspace)
	m, err := NewManager(opts)
	require.NoError(t, err)

	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
	}

	// Use a config provider that fails on save
	m.configProvider = &FailingConfigProvider{}

	ctx := context.Background()
	err = m.InitializeWithConfig(ctx, workspace, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save config")
}

// TestSetCLIConfig_WithConfigResolver tests SetCLIConfig when configResolver exists
func TestSetCLIConfig_WithConfigResolver(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create a config resolver with defaults
	defaults := &config.DefaultConfiguration{}
	m.configResolver = config.NewConfigResolver(defaults)

	cliConfig := map[string]interface{}{
		"test_key": "test_value",
	}

	m.SetCLIConfig(cliConfig)
	// If no panic, test passes
}

// TestSetCLIConfig_WithNilConfigResolver tests SetCLIConfig when configResolver is nil
func TestSetCLIConfig_WithNilConfigResolver(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)
	m.configResolver = nil

	cliConfig := map[string]interface{}{
		"test_key": "test_value",
	}

	// Should not panic
	m.SetCLIConfig(cliConfig)
}

// TestSetCLIConfig_WithWorkspaceConfig tests SetCLIConfig with workspace config
func TestSetCLIConfig_WithWorkspaceConfig(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create a config resolver with defaults
	defaults := &config.DefaultConfiguration{}
	m.configResolver = config.NewConfigResolver(defaults)

	// Set workspace config with overrides
	m.config = &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name: "test",
		},
		Overrides: map[string]interface{}{
			"workspace_key": "workspace_value",
		},
	}

	cliConfig := map[string]interface{}{
		"test_key": "test_value",
	}

	m.SetCLIConfig(cliConfig)
}

// TestGetConfigResolver tests GetConfigResolver
func TestGetConfigResolver(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	defaults := &config.DefaultConfiguration{}
	resolver := config.NewConfigResolver(defaults)
	m.configResolver = resolver

	result := m.GetConfigResolver()
	assert.Equal(t, resolver, result)
}

// FailingTreeAdapter is a test adapter that fails on Load
type FailingTreeAdapter struct {
	nodes map[string]interfaces.NodeInfo
}

func (f *FailingTreeAdapter) Load(cfg interface{}) error {
	return assert.AnError
}

func (f *FailingTreeAdapter) Navigate(path string) error                                 { return nil }
func (f *FailingTreeAdapter) GetCurrent() (interfaces.NodeInfo, error)                   { return interfaces.NodeInfo{}, nil }
func (f *FailingTreeAdapter) GetTree() (interfaces.NodeInfo, error)                      { return interfaces.NodeInfo{}, nil }
func (f *FailingTreeAdapter) GetNode(path string) (interfaces.NodeInfo, error)           { return interfaces.NodeInfo{}, assert.AnError }
func (f *FailingTreeAdapter) AddNode(parent string, node interfaces.NodeInfo) error      { return nil }
func (f *FailingTreeAdapter) RemoveNode(path string) error                               { return nil }
func (f *FailingTreeAdapter) UpdateNode(path string, node interfaces.NodeInfo) error     { return nil }
func (f *FailingTreeAdapter) ListChildren(path string) ([]interfaces.NodeInfo, error)    { return nil, nil }
func (f *FailingTreeAdapter) GetPath() string                                            { return "/" }
func (f *FailingTreeAdapter) SetPath(path string) error                                  { return nil }
func (f *FailingTreeAdapter) GetState() (interfaces.TreeState, error)                    { return interfaces.TreeState{}, nil }
func (f *FailingTreeAdapter) SetState(state interfaces.TreeState) error                  { return nil }

// FailingConfigProvider is a test adapter that fails on Save
type FailingConfigProvider struct{}

func (f *FailingConfigProvider) Load(path string) (interface{}, error) {
	return nil, nil
}

func (f *FailingConfigProvider) Save(path string, cfg interface{}) error {
	return assert.AnError
}

func (f *FailingConfigProvider) Exists(path string) bool {
	return false
}

func (f *FailingConfigProvider) Watch(path string) (<-chan interfaces.ConfigEvent, error) {
	return nil, nil
}
