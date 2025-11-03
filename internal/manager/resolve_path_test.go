package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
)

func TestManager_ResolvePath_RootPaths(t *testing.T) {
	tw := CreateTestWorkspace(t)
	m := CreateTestManager(t, tw.Root)
	
	// Resolve symlinks for macOS compatibility
	expectedRoot, _ := filepath.EvalSymlinks(tw.Root)
	
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{"slash root", "/", expectedRoot},
		{"tilde root", "~", expectedRoot},
		{"empty string", "", expectedRoot},
		{"dot current", ".", expectedRoot},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to workspace directory
			require.NoError(t, os.Chdir(tw.Root))
			
			result, err := m.ResolvePath(tt.target, false)
			require.NoError(t, err)
			// Resolve symlinks in result too for comparison
			resolvedResult, _ := filepath.EvalSymlinks(result)
			assert.Equal(t, tt.want, resolvedResult)
		})
	}
}

func TestManager_ResolvePath_ParentNavigation(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create config with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "platform", URL: "https://example.com/platform"},
			{Name: "team", URL: "https://example.com/team"},
		},
	}
	
	// Create manager with config
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	
	// Create directory structure
	tw.AddRepository("platform")
	tw.AddRepository("team")
	
	// Create nested structure for team repo
	teamRepoConfig := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "team",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "service", URL: "https://example.com/service"},
		},
	}
	tw.AddRepositoryWithConfig("team", teamRepoConfig)
	tw.AddRepository("team/.nodes/service")
	
	// Create nested structure for service repo
	serviceRepoConfig := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "service",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "module", URL: "https://example.com/module"},
		},
	}
	tw.AddRepositoryWithConfig("team/.nodes/service", serviceRepoConfig)
	tw.AddRepository("team/.nodes/service/.nodes/module")
	
	tests := []struct {
		name       string
		currentDir string
		target     string
		want       string
	}{
		{
			name:       "parent from root stays at root",
			currentDir: tw.Root,
			target:     "..",
			want:       tw.Root,
		},
		{
			name:       "parent from top-level repo",
			currentDir: filepath.Join(tw.NodesDir, "platform"),
			target:     "..",
			want:       tw.Root,
		},
		{
			name:       "parent from nested service",
			currentDir: filepath.Join(tw.NodesDir, "team", ".nodes", "service"),
			target:     "..",
			want:       filepath.Join(tw.NodesDir, "team"),
		},
		{
			name:       "parent from deeply nested module",
			currentDir: filepath.Join(tw.NodesDir, "team", ".nodes", "service", ".nodes", "module"),
			target:     "..",
			want:       filepath.Join(tw.NodesDir, "team", ".nodes", "service"),
		},
		{
			name:       "double parent from module",
			currentDir: filepath.Join(tw.NodesDir, "team", ".nodes", "service", ".nodes", "module"),
			target:     "../..",
			want:       filepath.Join(tw.NodesDir, "team"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to test directory
			require.NoError(t, os.Chdir(tt.currentDir))
			
			result, err := m.ResolvePath(tt.target, false)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestManager_ResolvePath_AbsolutePaths(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create config with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "backend", URL: "https://example.com/backend"},
			{Name: "frontend", URL: "https://example.com/frontend"},
		},
	}
	
	// Create manager with config
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	
	// Create structure
	tw.AddRepository("backend")
	tw.AddRepository("frontend")
	
	// Add backend's nested structure
	backendConfig := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "backend",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "api", URL: "https://example.com/api"},
		},
	}
	tw.AddRepositoryWithConfig("backend", backendConfig)
	tw.AddRepository("backend/.nodes/api")
	
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{
			name:   "absolute path to backend",
			target: "/backend",
			want:   filepath.Join(tw.NodesDir, "backend"),
		},
		{
			name:   "absolute path to frontend",
			target: "/frontend",
			want:   filepath.Join(tw.NodesDir, "frontend"),
		},
		{
			name:   "absolute path to nested api",
			target: "/backend/api",
			want:   filepath.Join(tw.NodesDir, "backend", ".nodes", "api"),
		},
		{
			name:   "absolute path to root",
			target: "/",
			want:   tw.Root,
		},
	}
	
	// Start from workspace root
	require.NoError(t, os.Chdir(tw.Root))
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.ResolvePath(tt.target, false)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestManager_ResolvePath_RelativePaths(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create config with nodes
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "backend", URL: "https://example.com/backend"},
			{Name: "frontend", URL: "https://example.com/frontend"},
			{Name: "shared", URL: "https://example.com/shared"},
		},
	}
	
	// Create manager with config
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	
	// Create structure
	tw.AddRepository("backend")
	tw.AddRepository("frontend")
	tw.AddRepository("shared")
	
	// Add backend's nested structure
	backendConfig := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "backend",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "api", URL: "https://example.com/api"},
			{Name: "db", URL: "https://example.com/db"},
		},
	}
	tw.AddRepositoryWithConfig("backend", backendConfig)
	tw.AddRepository("backend/.nodes/api")
	tw.AddRepository("backend/.nodes/db")
	
	tests := []struct {
		name       string
		currentDir string
		target     string
		want       string
	}{
		{
			name:       "relative from root to backend",
			currentDir: tw.Root,
			target:     "backend",
			want:       filepath.Join(tw.NodesDir, "backend"),
		},
		{
			name:       "relative from backend to api",
			currentDir: filepath.Join(tw.NodesDir, "backend"),
			target:     "api",
			want:       filepath.Join(tw.NodesDir, "backend", ".nodes", "api"),
		},
		{
			name:       "relative sibling from api to db",
			currentDir: filepath.Join(tw.NodesDir, "backend", ".nodes", "api"),
			target:     "../db",
			want:       filepath.Join(tw.NodesDir, "backend", ".nodes", "db"),
		},
		{
			name:       "current directory from api",
			currentDir: filepath.Join(tw.NodesDir, "backend", ".nodes", "api"),
			target:     ".",
			want:       filepath.Join(tw.NodesDir, "backend", ".nodes", "api"),
		},
		{
			name:       "navigate up to root from nested",
			currentDir: filepath.Join(tw.NodesDir, "backend", ".nodes", "db"),
			target:     "../..",
			want:       tw.Root,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to test directory
			require.NoError(t, os.Chdir(tt.currentDir))
			
			result, err := m.ResolvePath(tt.target, false)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestManager_ResolvePath_ConfigReferences(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create external config file for config reference
	externalCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "external",
			ReposDir: "custom-repos",
		},
		Nodes: []config.NodeDefinition{
			{Name: "service1", URL: "https://example.com/service1"},
			{Name: "service2", URL: "https://example.com/service2"},
		},
	}
	tw.CreateConfigReference("configs/external.yaml", externalCfg)
	
	// Create main config with config reference
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "config-ref", File: "configs/external.yaml"},
			{Name: "regular", URL: "https://example.com/regular"},
		},
	}
	
	// Create manager with config
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	
	// Create directory structure
	tw.AddRepository("config-ref")
	tw.AddRepository("regular")
	
	// Since config-ref uses custom repos_dir, create those directories
	os.MkdirAll(filepath.Join(tw.NodesDir, "config-ref", "custom-repos", "service1"), 0755)
	os.MkdirAll(filepath.Join(tw.NodesDir, "config-ref", "custom-repos", "service2"), 0755)
	
	tests := []struct {
		name       string
		currentDir string
		target     string
		want       string
	}{
		{
			name:       "navigate to config reference node",
			currentDir: tw.Root,
			target:     "config-ref",
			want:       filepath.Join(tw.NodesDir, "config-ref"),
		},
		{
			name:       "child under config reference",
			currentDir: filepath.Join(tw.NodesDir, "config-ref"),
			target:     "service1",
			want:       filepath.Join(tw.NodesDir, "config-ref", "custom-repos", "service1"),
		},
		{
			name:       "another child under config reference",
			currentDir: filepath.Join(tw.NodesDir, "config-ref"),
			target:     "service2",
			want:       filepath.Join(tw.NodesDir, "config-ref", "custom-repos", "service2"),
		},
		{
			name:       "regular node uses default repos_dir",
			currentDir: tw.Root,
			target:     "regular",
			want:       filepath.Join(tw.NodesDir, "regular"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to test directory
			require.NoError(t, os.Chdir(tt.currentDir))
			
			result, err := m.ResolvePath(tt.target, false)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestManager_ResolvePath_EnsureFlag(t *testing.T) {
	// This test would require mock git provider to test cloning
	// For now, we'll test the basic logic without actual cloning
	
	tw := CreateTestWorkspace(t)
	m := CreateTestManager(t, tw.Root)
	
	// Add a lazy node
	AddNodeToTree(m, "/lazy-repo", CreateLazyNode("lazy-repo", "https://example.com/lazy"))
	
	// With ensure=false, should fail for non-existent path
	_, err := m.ResolvePath("/lazy-repo", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")
	
	// Create the directory manually to simulate it exists
	tw.AddRepository("lazy-repo")
	
	// Now it should work
	result, err := m.ResolvePath("/lazy-repo", false)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tw.NodesDir, "lazy-repo"), result)
}

func TestManager_ResolvePath_Errors(t *testing.T) {
	tw := CreateTestWorkspace(t)
	m := CreateTestManager(t, tw.Root)
	
	tests := []struct {
		name          string
		target        string
		errorContains string
	}{
		{
			name:          "non-existent absolute path",
			target:        "/nonexistent",
			errorContains: "does not exist",
		},
		{
			name:          "non-existent relative path",
			target:        "nonexistent",
			errorContains: "does not exist",
		},
		{
			name:          "non-existent nested path",
			target:        "/team/nonexistent/module",
			errorContains: "does not exist",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, os.Chdir(tw.Root))
			
			_, err := m.ResolvePath(tt.target, false)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errorContains)
		})
	}
}

func TestManager_ResolvePath_ComplexRelatives(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create main config with team structure
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "team", URL: "https://example.com/team"},
		},
	}
	
	// Create manager with config
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	
	// Create team structure
	tw.AddRepository("team")
	
	// Create team config with backend and frontend
	teamCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "team",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "backend", URL: "https://example.com/backend"},
			{Name: "frontend", URL: "https://example.com/frontend"},
		},
	}
	tw.AddRepositoryWithConfig("team", teamCfg)
	tw.AddRepository("team/.nodes/backend")
	tw.AddRepository("team/.nodes/frontend")
	
	// Create backend config with api and db
	backendCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "backend",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "api", URL: "https://example.com/api"},
			{Name: "db", URL: "https://example.com/db"},
		},
	}
	tw.AddRepositoryWithConfig("team/.nodes/backend", backendCfg)
	tw.AddRepository("team/.nodes/backend/.nodes/api")
	tw.AddRepository("team/.nodes/backend/.nodes/db")
	
	// Create frontend config with web and mobile
	frontendCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "frontend",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "web", URL: "https://example.com/web"},
			{Name: "mobile", URL: "https://example.com/mobile"},
		},
	}
	tw.AddRepositoryWithConfig("team/.nodes/frontend", frontendCfg)
	tw.AddRepository("team/.nodes/frontend/.nodes/web")
	tw.AddRepository("team/.nodes/frontend/.nodes/mobile")
	
	tests := []struct {
		name       string
		currentDir string
		target     string
		want       string
	}{
		{
			name:       "sibling from api to db",
			currentDir: filepath.Join(tw.NodesDir, "team", ".nodes", "backend", ".nodes", "api"),
			target:     "../db",
			want:       filepath.Join(tw.NodesDir, "team", ".nodes", "backend", ".nodes", "db"),
		},
		{
			name:       "cousin from api to web",
			currentDir: filepath.Join(tw.NodesDir, "team", ".nodes", "backend", ".nodes", "api"),
			target:     "../../frontend/web",
			want:       filepath.Join(tw.NodesDir, "team", ".nodes", "frontend", ".nodes", "web"),
		},
		{
			name:       "complex navigation from mobile to api",
			currentDir: filepath.Join(tw.NodesDir, "team", ".nodes", "frontend", ".nodes", "mobile"),
			target:     "../../backend/api",
			want:       filepath.Join(tw.NodesDir, "team", ".nodes", "backend", ".nodes", "api"),
		},
		{
			name:       "up and down from backend to mobile",
			currentDir: filepath.Join(tw.NodesDir, "team", ".nodes", "backend"),
			target:     "../frontend/mobile",
			want:       filepath.Join(tw.NodesDir, "team", ".nodes", "frontend", ".nodes", "mobile"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, os.Chdir(tt.currentDir))
			
			result, err := m.ResolvePath(tt.target, false)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestManager_ResolvePath_NotInitialized(t *testing.T) {
	tw := CreateTestWorkspace(t)
	m := CreateTestManager(t, tw.Root)
	m.initialized = false // Mark as not initialized
	
	_, err := m.ResolvePath("/test", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not initialized")
}

func TestManager_ResolvePath_EdgeCases(t *testing.T) {
	tw := CreateTestWorkspace(t)
	m := CreateTestManager(t, tw.Root)
	
	// Test from .nodes directory itself
	require.NoError(t, os.Chdir(tw.NodesDir))
	result, err := m.ResolvePath(".", false)
	require.NoError(t, err)
	assert.Equal(t, tw.Root, result) // .nodes resolves to workspace root
	
	// Test with trailing slashes
	AddNodeToTree(m, "/repo", CreateSimpleNode("repo", ""))
	tw.AddRepository("repo")
	
	result, err = m.ResolvePath("/repo/", false)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tw.NodesDir, "repo"), result)
	
	// Test with multiple slashes
	result, err = m.ResolvePath("//repo", false)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tw.NodesDir, "repo"), result)
}