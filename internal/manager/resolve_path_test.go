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
	m := CreateTestManager(t, tw.Root)
	
	// Create nested structure
	tw.AddRepository("team")
	tw.AddRepository("team/.nodes/service")
	tw.AddRepository("team/.nodes/service/.nodes/module")
	
	// Add nodes to tree
	AddNodeToTree(m, "/team", CreateSimpleNode("team", "https://example.com/team"))
	AddNodeToTree(m, "/team/service", CreateSimpleNode("service", "https://example.com/service"))
	AddNodeToTree(m, "/team/service/module", CreateSimpleNode("module", "https://example.com/module"))
	
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
			currentDir: filepath.Join(tw.NodesDir, "team"),
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
	m := CreateTestManager(t, tw.Root)
	
	// Create structure
	tw.AddRepository("backend")
	tw.AddRepository("frontend")
	tw.AddRepository("backend/.nodes/api")
	
	// Add nodes to tree
	AddNodeToTree(m, "/backend", CreateSimpleNode("backend", "https://example.com/backend"))
	AddNodeToTree(m, "/frontend", CreateSimpleNode("frontend", "https://example.com/frontend"))
	AddNodeToTree(m, "/backend/api", CreateSimpleNode("api", "https://example.com/api"))
	
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
			name:   "absolute nested path",
			target: "/backend/api",
			want:   filepath.Join(tw.NodesDir, "backend", ".nodes", "api"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test from any location
			require.NoError(t, os.Chdir(tw.Root))
			
			result, err := m.ResolvePath(tt.target, false)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
			
			// Also test from nested location - absolute path should give same result
			require.NoError(t, os.Chdir(filepath.Join(tw.NodesDir, "backend")))
			result2, err2 := m.ResolvePath(tt.target, false)
			require.NoError(t, err2)
			assert.Equal(t, tt.want, result2)
		})
	}
}

func TestManager_ResolvePath_RelativePaths(t *testing.T) {
	tw := CreateTestWorkspace(t)
	m := CreateTestManager(t, tw.Root)
	
	// Create structure
	tw.AddRepository("team")
	tw.AddRepository("team/.nodes/service1")
	tw.AddRepository("team/.nodes/service2")
	
	// Add nodes to tree
	AddNodeToTree(m, "/team", CreateSimpleNode("team", "https://example.com/team"))
	AddNodeToTree(m, "/team/service1", CreateSimpleNode("service1", "https://example.com/service1"))
	AddNodeToTree(m, "/team/service2", CreateSimpleNode("service2", "https://example.com/service2"))
	
	tests := []struct {
		name       string
		currentDir string
		target     string
		want       string
	}{
		{
			name:       "relative child from root",
			currentDir: tw.Root,
			target:     "team",
			want:       filepath.Join(tw.NodesDir, "team"),
		},
		{
			name:       "relative nested from root",
			currentDir: tw.Root,
			target:     "team/service1",
			want:       filepath.Join(tw.NodesDir, "team", ".nodes", "service1"),
		},
		{
			name:       "relative sibling",
			currentDir: filepath.Join(tw.NodesDir, "team", ".nodes", "service1"),
			target:     "../service2",
			want:       filepath.Join(tw.NodesDir, "team", ".nodes", "service2"),
		},
		{
			name:       "relative child from parent",
			currentDir: filepath.Join(tw.NodesDir, "team"),
			target:     "service1",
			want:       filepath.Join(tw.NodesDir, "team", ".nodes", "service1"),
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

func TestManager_ResolvePath_ConfigReferences(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create external config with custom repos_dir
	extConfig := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "external",
			ReposDir: "custom-repos",
		},
		Nodes: []config.NodeDefinition{
			{Name: "service1", URL: "https://example.com/service1"},
			{Name: "service2", URL: "https://example.com/service2"},
		},
	}
	tw.CreateConfigReference("configs/external.yaml", extConfig)
	
	// Create main config
	mainConfig := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "config-ref", File: "configs/external.yaml"},
			{Name: "regular", URL: "https://example.com/regular"},
		},
	}
	tw.CreateConfig(mainConfig)
	
	m := CreateTestManagerWithConfig(t, tw.Root, mainConfig)
	
	// Create the directory structure
	configRefDir := tw.AddRepository("config-ref")
	customReposDir := filepath.Join(configRefDir, "custom-repos")
	require.NoError(t, os.MkdirAll(customReposDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(customReposDir, "service1"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(customReposDir, "service2"), 0755))
	tw.AddRepository("regular")
	
	// Add nodes to tree
	AddNodeToTree(m, "/config-ref", CreateConfigNode("config-ref", "configs/external.yaml"))
	AddNodeToTree(m, "/config-ref/service1", CreateSimpleNode("service1", "https://example.com/service1"))
	AddNodeToTree(m, "/config-ref/service2", CreateSimpleNode("service2", "https://example.com/service2"))
	AddNodeToTree(m, "/regular", CreateSimpleNode("regular", "https://example.com/regular"))
	
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{
			name:   "config reference node",
			target: "/config-ref",
			want:   filepath.Join(tw.NodesDir, "config-ref"),
		},
		{
			name:   "child under config reference with custom repos_dir",
			target: "/config-ref/service1",
			want:   filepath.Join(tw.NodesDir, "config-ref", "custom-repos", "service1"),
		},
		{
			name:   "another child under config reference",
			target: "/config-ref/service2",
			want:   filepath.Join(tw.NodesDir, "config-ref", "custom-repos", "service2"),
		},
		{
			name:   "regular node uses default repos_dir",
			target: "/regular",
			want:   filepath.Join(tw.NodesDir, "regular"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, os.Chdir(tw.Root))
			
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
	m := CreateTestManager(t, tw.Root)
	
	// Create a complex structure
	// /team
	//   /backend
	//     /api
	//     /db
	//   /frontend
	//     /web
	//     /mobile
	tw.AddRepository("team")
	tw.AddRepository("team/.nodes/backend")
	tw.AddRepository("team/.nodes/backend/.nodes/api")
	tw.AddRepository("team/.nodes/backend/.nodes/db")
	tw.AddRepository("team/.nodes/frontend")
	tw.AddRepository("team/.nodes/frontend/.nodes/web")
	tw.AddRepository("team/.nodes/frontend/.nodes/mobile")
	
	// Add all nodes to tree
	AddNodeToTree(m, "/team", CreateSimpleNode("team", ""))
	AddNodeToTree(m, "/team/backend", CreateSimpleNode("backend", ""))
	AddNodeToTree(m, "/team/backend/api", CreateSimpleNode("api", ""))
	AddNodeToTree(m, "/team/backend/db", CreateSimpleNode("db", ""))
	AddNodeToTree(m, "/team/frontend", CreateSimpleNode("frontend", ""))
	AddNodeToTree(m, "/team/frontend/web", CreateSimpleNode("web", ""))
	AddNodeToTree(m, "/team/frontend/mobile", CreateSimpleNode("mobile", ""))
	
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