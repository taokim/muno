package manager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
)

// TestListNodesQuiet_Uninitialized tests that ListNodesQuiet fails when manager is not initialized
func TestListNodesQuiet_Uninitialized(t *testing.T) {
	workspace := t.TempDir()
	opts := CreateTestManagerOptions(workspace)
	m, err := NewManager(opts)
	require.NoError(t, err)
	m.initialized = false

	err = m.ListNodesQuiet(false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestListNodesQuiet_NonRecursive tests non-recursive quiet listing
func TestListNodesQuiet_NonRecursive(t *testing.T) {
	workspace := t.TempDir()

	// Create manager
	m := CreateTestManager(t, workspace)

	// Create a tree with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Fetch: "eager"},
			{Name: "repo2", URL: "https://github.com/test/repo2.git", Fetch: "lazy"},
			{Name: "repo3", URL: "https://github.com/test/repo3.git", Fetch: "eager"},
		},
	}

	err := m.treeProvider.Load(cfg)
	require.NoError(t, err)

	// Change to workspace directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(workspace)

	// Test non-recursive listing
	err = m.ListNodesQuiet(false)
	assert.NoError(t, err)
}

// TestListNodesQuiet_Recursive tests recursive quiet listing
func TestListNodesQuiet_Recursive(t *testing.T) {
	workspace := t.TempDir()

	// Create manager
	m := CreateTestManager(t, workspace)

	// Create a tree with nested nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "parent1", URL: "https://github.com/test/parent1.git", Fetch: "eager"},
			{Name: "parent2", URL: "https://github.com/test/parent2.git", Fetch: "eager"},
		},
	}

	err := m.treeProvider.Load(cfg)
	require.NoError(t, err)

	// Change to workspace directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(workspace)

	// Test recursive listing
	err = m.ListNodesQuiet(true)
	assert.NoError(t, err)
}

// TestListNodesQuiet_FromSubdirectory tests listing from a subdirectory
func TestListNodesQuiet_FromSubdirectory(t *testing.T) {
	workspace := t.TempDir()

	// Create physical directory structure
	nodesDir := filepath.Join(workspace, ".nodes")
	repo1Dir := filepath.Join(nodesDir, "repo1")
	require.NoError(t, os.MkdirAll(repo1Dir, 0755))

	// Create manager
	m := CreateTestManager(t, workspace)

	// Create a tree with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Fetch: "eager"},
		},
	}

	err := m.treeProvider.Load(cfg)
	require.NoError(t, err)

	// Change to repo1 subdirectory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(repo1Dir)

	// Test listing from subdirectory
	err = m.ListNodesQuiet(false)
	assert.NoError(t, err)
}

// TestListNodesRecursive_Uninitialized tests that ListNodesRecursive fails when manager is not initialized
func TestListNodesRecursive_Uninitialized(t *testing.T) {
	workspace := t.TempDir()
	opts := CreateTestManagerOptions(workspace)
	m, err := NewManager(opts)
	require.NoError(t, err)
	m.initialized = false

	err = m.ListNodesRecursive(false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestListNodesRecursive_RecursiveMode tests recursive tree display
func TestListNodesRecursive_RecursiveMode(t *testing.T) {
	workspace := t.TempDir()

	// Create manager with UI adapter stub that captures output
	m := CreateTestManager(t, workspace)
	uiStub := &UIAdapterStub{messages: make([]string, 0)}
	m.uiProvider = uiStub

	// Create a tree with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Fetch: "eager"},
			{Name: "repo2", URL: "https://github.com/test/repo2.git", Fetch: "lazy"},
		},
	}

	err := m.treeProvider.Load(cfg)
	require.NoError(t, err)

	// Test recursive mode
	err = m.ListNodesRecursive(true)
	assert.NoError(t, err)

	// Verify UI output contains expected elements
	output := strings.Join(uiStub.messages, "\n")
	assert.Contains(t, output, "🌳 Repository Tree")
	assert.Contains(t, output, "📊 Summary")
}

// TestListNodesRecursive_NonRecursiveMode tests non-recursive tree display
func TestListNodesRecursive_NonRecursiveMode(t *testing.T) {
	workspace := t.TempDir()

	// Create manager with UI adapter stub
	m := CreateTestManager(t, workspace)
	uiStub := &UIAdapterStub{messages: make([]string, 0)}
	m.uiProvider = uiStub

	// Create a tree with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Fetch: "eager"},
			{Name: "repo2", URL: "https://github.com/test/repo2.git", Fetch: "lazy"},
		},
	}

	err := m.treeProvider.Load(cfg)
	require.NoError(t, err)

	// Change to workspace directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(workspace)

	// Test non-recursive mode
	err = m.ListNodesRecursive(false)
	assert.NoError(t, err)
}

// TestListNodesRecursive_WithLazyNodes tests that lazy nodes are counted correctly
func TestListNodesRecursive_WithLazyNodes(t *testing.T) {
	workspace := t.TempDir()

	// Create manager with UI adapter stub
	m := CreateTestManager(t, workspace)
	uiStub := &UIAdapterStub{messages: make([]string, 0)}
	m.uiProvider = uiStub

	// Create a tree with lazy nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Fetch: "lazy"},
			{Name: "repo2", URL: "https://github.com/test/repo2.git", Fetch: "lazy"},
			{Name: "repo3", URL: "https://github.com/test/repo3.git", Fetch: "eager"},
		},
	}

	err := m.treeProvider.Load(cfg)
	require.NoError(t, err)

	// Test recursive mode
	err = m.ListNodesRecursive(true)
	assert.NoError(t, err)

	// Verify summary includes lazy count
	output := strings.Join(uiStub.messages, "\n")
	assert.Contains(t, output, "lazy")
}

// TestShowTreeAtPath_Uninitialized tests that ShowTreeAtPath fails when manager is not initialized
func TestShowTreeAtPath_Uninitialized(t *testing.T) {
	workspace := t.TempDir()
	opts := CreateTestManagerOptions(workspace)
	m, err := NewManager(opts)
	require.NoError(t, err)
	m.initialized = false

	err = m.ShowTreeAtPath("/", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestShowTreeAtPath_RootPath tests showing tree from root path
func TestShowTreeAtPath_RootPath(t *testing.T) {
	workspace := t.TempDir()

	// Create manager with UI adapter stub
	m := CreateTestManager(t, workspace)
	uiStub := &UIAdapterStub{messages: make([]string, 0)}
	m.uiProvider = uiStub

	// Create a tree with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Fetch: "eager"},
			{Name: "repo2", URL: "https://github.com/test/repo2.git", Fetch: "lazy"},
		},
	}

	err := m.treeProvider.Load(cfg)
	require.NoError(t, err)

	// Test showing tree at root path
	err = m.ShowTreeAtPath("/", 0)
	assert.NoError(t, err)

	// Verify UI output
	output := strings.Join(uiStub.messages, "\n")
	assert.Contains(t, output, "🌳 Repository Tree")
	assert.Contains(t, output, "📍 Starting from: /")
	assert.Contains(t, output, "📊 Summary")
}

// TestShowTreeAtPath_EmptyPath tests showing tree with empty path (uses PWD)
func TestShowTreeAtPath_EmptyPath(t *testing.T) {
	workspace := t.TempDir()

	// Create manager with UI adapter stub
	m := CreateTestManager(t, workspace)
	uiStub := &UIAdapterStub{messages: make([]string, 0)}
	m.uiProvider = uiStub

	// Create a tree with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Fetch: "eager"},
		},
	}

	err := m.treeProvider.Load(cfg)
	require.NoError(t, err)

	// Change to workspace directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(workspace)

	// Test with empty path - should use PWD
	err = m.ShowTreeAtPath("", 0)
	assert.NoError(t, err)

	// Verify UI output
	output := strings.Join(uiStub.messages, "\n")
	assert.Contains(t, output, "📍 Starting from:")
}

// TestShowTreeAtPath_InvalidNode tests showing tree at non-existent path
func TestShowTreeAtPath_InvalidNode(t *testing.T) {
	workspace := t.TempDir()

	// Create manager
	m := CreateTestManager(t, workspace)

	// Create a tree with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Fetch: "eager"},
		},
	}

	err := m.treeProvider.Load(cfg)
	require.NoError(t, err)

	// Test with invalid path
	err = m.ShowTreeAtPath("/nonexistent", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node not found")
}

// TestShowTreeAtPath_SpecificNode tests showing tree at a specific node path
func TestShowTreeAtPath_SpecificNode(t *testing.T) {
	workspace := t.TempDir()

	// Create manager with UI adapter stub
	m := CreateTestManager(t, workspace)
	uiStub := &UIAdapterStub{messages: make([]string, 0)}
	m.uiProvider = uiStub

	// Create a tree with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Fetch: "eager"},
			{Name: "repo2", URL: "https://github.com/test/repo2.git", Fetch: "eager"},
		},
	}

	err := m.treeProvider.Load(cfg)
	require.NoError(t, err)

	// Test showing tree at specific node
	err = m.ShowTreeAtPath("/repo1", 0)
	assert.NoError(t, err)

	// Verify UI output shows the correct starting point
	output := strings.Join(uiStub.messages, "\n")
	assert.Contains(t, output, "📍 Starting from: /repo1")
}

// TestShowTreeAtPath_FromReposDir tests showing tree when PWD is in repos directory
func TestShowTreeAtPath_FromReposDir(t *testing.T) {
	workspace := t.TempDir()

	// Create physical directory structure
	nodesDir := filepath.Join(workspace, ".nodes")
	repo1Dir := filepath.Join(nodesDir, "repo1")
	require.NoError(t, os.MkdirAll(repo1Dir, 0755))

	// Create manager with UI adapter stub
	m := CreateTestManager(t, workspace)
	uiStub := &UIAdapterStub{messages: make([]string, 0)}
	m.uiProvider = uiStub

	// Create a tree with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", Fetch: "eager"},
		},
	}

	err := m.treeProvider.Load(cfg)
	require.NoError(t, err)

	// Change to repos directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(repo1Dir)

	// Test with empty path - should detect we're in repos dir
	err = m.ShowTreeAtPath("", 0)
	assert.NoError(t, err)

	// Verify UI output
	output := strings.Join(uiStub.messages, "\n")
	assert.Contains(t, output, "📍 Starting from:")
}

// UIAdapterStub is a test stub for UIProvider that captures messages
type UIAdapterStub struct {
	messages []string
}

func (u *UIAdapterStub) Prompt(message string) (string, error) {
	return "test", nil
}

func (u *UIAdapterStub) PromptPassword(message string) (string, error) {
	return "password", nil
}

func (u *UIAdapterStub) Confirm(message string) (bool, error) {
	return true, nil
}

func (u *UIAdapterStub) Select(message string, options []string) (string, error) {
	if len(options) > 0 {
		return options[0], nil
	}
	return "", nil
}

func (u *UIAdapterStub) MultiSelect(message string, options []string) ([]string, error) {
	return options, nil
}

func (u *UIAdapterStub) Progress(message string) interfaces.ProgressReporter {
	return &ProgressReporterStub{}
}

func (u *UIAdapterStub) Info(msg string) {
	u.messages = append(u.messages, msg)
}

func (u *UIAdapterStub) Success(msg string) {
	u.messages = append(u.messages, msg)
}

func (u *UIAdapterStub) Warning(msg string) {
	u.messages = append(u.messages, msg)
}

func (u *UIAdapterStub) Error(msg string) {
	u.messages = append(u.messages, msg)
}

func (u *UIAdapterStub) Debug(msg string) {
	u.messages = append(u.messages, msg)
}

// ProgressReporterStub is a test stub for ProgressReporter
type ProgressReporterStub struct{}

func (p *ProgressReporterStub) Start() {}

func (p *ProgressReporterStub) Update(current, total int) {}

func (p *ProgressReporterStub) SetMessage(message string) {}

func (p *ProgressReporterStub) Finish() {}

func (p *ProgressReporterStub) Error(err error) {}

// Ensure UIAdapterStub implements interfaces.UIProvider
var _ interfaces.UIProvider = (*UIAdapterStub)(nil)
var _ interfaces.ProgressReporter = (*ProgressReporterStub)(nil)
