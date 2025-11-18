package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
)

func TestManager_GetTreePath_BasicConversions(t *testing.T) {
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
	
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	
	// Create physical directories
	tw.AddRepository("backend")
	tw.AddRepository("frontend")
	
	tests := []struct {
		name         string
		physicalPath string
		want         string
		wantErr      bool
	}{
		{
			name:         "workspace root to tree root",
			physicalPath: tw.Root,
			want:         "/",
			wantErr:      false,
		},
		{
			name:         "nodes directory to tree root",
			physicalPath: tw.NodesDir,
			want:         "/",
			wantErr:      false,
		},
		{
			name:         "backend repo to tree path",
			physicalPath: filepath.Join(tw.NodesDir, "backend"),
			want:         "/backend",
			wantErr:      false,
		},
		{
			name:         "frontend repo to tree path",
			physicalPath: filepath.Join(tw.NodesDir, "frontend"),
			want:         "/frontend",
			wantErr:      false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.GetTreePath(tt.physicalPath)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestManager_GetTreePath_NestedStructures(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create config with team structure
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "team", URL: "https://example.com/team"},
		},
	}
	
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	
	// Create nested directory structure
	tw.AddRepository("team")
	tw.AddRepository("team/.nodes/backend")
	tw.AddRepository("team/.nodes/backend/.nodes/api")
	tw.AddRepository("team/.nodes/backend/.nodes/db")
	tw.AddRepository("team/.nodes/frontend")
	tw.AddRepository("team/.nodes/frontend/.nodes/web")
	
	tests := []struct {
		name         string
		physicalPath string
		want         string
	}{
		{
			name:         "team repo",
			physicalPath: filepath.Join(tw.NodesDir, "team"),
			want:         "/team",
		},
		{
			name:         "nested backend",
			physicalPath: filepath.Join(tw.NodesDir, "team", ".nodes", "backend"),
			want:         "/team/backend",
		},
		{
			name:         "deeply nested api",
			physicalPath: filepath.Join(tw.NodesDir, "team", ".nodes", "backend", ".nodes", "api"),
			want:         "/team/backend/api",
		},
		{
			name:         "deeply nested db",
			physicalPath: filepath.Join(tw.NodesDir, "team", ".nodes", "backend", ".nodes", "db"),
			want:         "/team/backend/db",
		},
		{
			name:         "nested frontend",
			physicalPath: filepath.Join(tw.NodesDir, "team", ".nodes", "frontend"),
			want:         "/team/frontend",
		},
		{
			name:         "deeply nested web",
			physicalPath: filepath.Join(tw.NodesDir, "team", ".nodes", "frontend", ".nodes", "web"),
			want:         "/team/frontend/web",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.GetTreePath(tt.physicalPath)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestManager_GetTreePath_ConfigReferences(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create external config with custom repos_dir
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
	
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)

	// Create directory structure with custom repos_dir
	tw.AddRepository("config-ref")
	// Create muno.yaml in config-ref directory so buildTreePathFromFilesystem can find it
	configRefPath := filepath.Join(tw.NodesDir, "config-ref")
	externalCfg.Save(filepath.Join(configRefPath, "muno.yaml"))
	os.MkdirAll(filepath.Join(tw.NodesDir, "config-ref", "custom-repos", "service1"), 0755)
	os.MkdirAll(filepath.Join(tw.NodesDir, "config-ref", "custom-repos", "service2"), 0755)
	tw.AddRepository("regular")
	
	tests := []struct {
		name         string
		physicalPath string
		want         string
		description  string
	}{
		{
			name:         "config reference node",
			physicalPath: filepath.Join(tw.NodesDir, "config-ref"),
			want:         "/config-ref",
			description:  "config reference node should map to tree path",
		},
		{
			name:         "child under config ref with custom repos_dir",
			physicalPath: filepath.Join(tw.NodesDir, "config-ref", "custom-repos", "service1"),
			want:         "/config-ref/service1",
			description:  "should strip custom repos_dir from tree path",
		},
		{
			name:         "another child under config ref",
			physicalPath: filepath.Join(tw.NodesDir, "config-ref", "custom-repos", "service2"),
			want:         "/config-ref/service2",
			description:  "should handle multiple children under config ref",
		},
		{
			name:         "regular node with default repos_dir",
			physicalPath: filepath.Join(tw.NodesDir, "regular"),
			want:         "/regular",
			description:  "regular nodes should work normally",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.GetTreePath(tt.physicalPath)
			require.NoError(t, err, tt.description)
			assert.Equal(t, tt.want, result, tt.description)
		})
	}
}

func TestManager_GetTreePath_ErrorCases(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo", URL: "https://example.com/repo"},
		},
	}
	
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	
	tests := []struct {
		name         string
		physicalPath string
		wantErr      string
	}{
		{
			name:         "path outside workspace",
			physicalPath: "/tmp/outside",
			wantErr:      "not within workspace",
		},
		{
			name:         "parent of workspace",
			physicalPath: filepath.Dir(tw.Root),
			wantErr:      "not within workspace",
		},
		{
			name:         "non-existent path within workspace",
			physicalPath: filepath.Join(tw.NodesDir, "nonexistent"),
			wantErr:      "",  // Should still work, just returns computed path
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.GetTreePath(tt.physicalPath)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				// Non-existent paths should still compute a tree path
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestManager_GetTreePath_SymlinkResolution(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo", URL: "https://example.com/repo"},
		},
	}
	
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	
	// Create a repository
	repoPath := tw.AddRepository("repo")
	
	// Create a symlink to the repository
	symlinkPath := filepath.Join(tw.Root, "repo-link")
	err := os.Symlink(repoPath, symlinkPath)
	require.NoError(t, err)
	
	// Test that symlink resolves to the same tree path
	realResult, err := m.GetTreePath(repoPath)
	require.NoError(t, err)
	
	symlinkResult, err := m.GetTreePath(symlinkPath)
	require.NoError(t, err)
	
	assert.Equal(t, realResult, symlinkResult, "symlink should resolve to same tree path")
}

func TestManager_GetTreePath_EdgeCases(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo-with-dots", URL: "https://example.com/repo.with.dots"},
			{Name: "repo_with_underscores", URL: "https://example.com/repo_with_underscores"},
			{Name: "repo-with-dashes", URL: "https://example.com/repo-with-dashes"},
		},
	}
	
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	
	// Create repositories with special characters
	tw.AddRepository("repo-with-dots")
	tw.AddRepository("repo_with_underscores")
	tw.AddRepository("repo-with-dashes")
	
	tests := []struct {
		name         string
		physicalPath string
		want         string
	}{
		{
			name:         "repository with dots",
			physicalPath: filepath.Join(tw.NodesDir, "repo-with-dots"),
			want:         "/repo-with-dots",
		},
		{
			name:         "repository with underscores",
			physicalPath: filepath.Join(tw.NodesDir, "repo_with_underscores"),
			want:         "/repo_with_underscores",
		},
		{
			name:         "repository with dashes",
			physicalPath: filepath.Join(tw.NodesDir, "repo-with-dashes"),
			want:         "/repo-with-dashes",
		},
		{
			name:         "trailing slash should be handled",
			physicalPath: filepath.Join(tw.NodesDir, "repo-with-dashes") + string(filepath.Separator),
			want:         "/repo-with-dashes",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.GetTreePath(tt.physicalPath)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestManager_GetTreePath_NotInitialized(t *testing.T) {
	m := &Manager{
		initialized: false,
	}
	
	_, err := m.GetTreePath("/some/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}