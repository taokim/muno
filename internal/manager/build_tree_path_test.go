package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
)

func TestManager_buildTreePathFromFilesystem_Simple(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create config with basic structure
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo1", URL: "https://example.com/repo1"},
			{Name: "repo2", URL: "https://example.com/repo2"},
		},
	}
	
	// Create manager and save config
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	tw.CreateConfig(cfg)
	
	// Create repository directories
	repo1Path := tw.AddRepository("repo1")
	repo2Path := tw.AddRepository("repo2")
	
	tests := []struct {
		name         string
		physicalPath string
		want         string
	}{
		{
			name:         "simple repo1 path",
			physicalPath: repo1Path,
			want:         "/repo1",
		},
		{
			name:         "simple repo2 path",
			physicalPath: repo2Path,
			want:         "/repo2",
		},
		{
			name:         "workspace root",
			physicalPath: tw.Root,
			want:         "/",
		},
		{
			name:         "nodes directory",
			physicalPath: tw.NodesDir,
			want:         "/",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildTreePathFromFilesystem(tt.physicalPath)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestManager_buildTreePathFromFilesystem_NestedWithConfigs(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create main config
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
	
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	tw.CreateConfig(cfg)
	
	// Create platform with its own muno.yaml
	platformPath := tw.AddRepository("platform")
	platformCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "platform",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "core", URL: "https://example.com/core"},
			{Name: "utils", URL: "https://example.com/utils"},
		},
	}
	tw.AddRepositoryWithConfig("platform", platformCfg)
	
	// Create nested repositories
	corePath := tw.AddRepository("platform/.nodes/core")
	utilsPath := tw.AddRepository("platform/.nodes/utils")
	
	tests := []struct {
		name         string
		physicalPath string
		want         string
		description  string
	}{
		{
			name:         "platform repo",
			physicalPath: platformPath,
			want:         "/platform",
			description:  "top-level repository",
		},
		{
			name:         "nested core under platform",
			physicalPath: corePath,
			want:         "/platform/core",
			description:  "should build path from nested muno.yaml",
		},
		{
			name:         "nested utils under platform",
			physicalPath: utilsPath,
			want:         "/platform/utils",
			description:  "should handle sibling nested repo",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildTreePathFromFilesystem(tt.physicalPath)
			assert.Equal(t, tt.want, result, tt.description)
		})
	}
}

func TestManager_buildTreePathFromFilesystem_ConfigReferences(t *testing.T) {
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
	externalConfigPath := tw.CreateConfigReference("configs/external.yaml", externalCfg)
	
	// Create main config
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
	tw.CreateConfig(cfg)
	
	// Create config-ref directory
	configRefPath := tw.AddRepository("config-ref")
	
	// Create symlink to external config inside config-ref
	symlinkPath := filepath.Join(configRefPath, "muno.yaml")
	err := os.Symlink(externalConfigPath, symlinkPath)
	require.NoError(t, err)
	
	// Create custom repos directories
	service1Path := filepath.Join(configRefPath, "custom-repos", "service1")
	service2Path := filepath.Join(configRefPath, "custom-repos", "service2")
	os.MkdirAll(service1Path, 0755)
	os.MkdirAll(service2Path, 0755)
	
	tests := []struct {
		name         string
		physicalPath string
		want         string
		description  string
	}{
		{
			name:         "config reference node itself",
			physicalPath: configRefPath,
			want:         "/config-ref",
			description:  "config ref node should map correctly",
		},
		{
			name:         "service1 under config ref",
			physicalPath: service1Path,
			want:         "/config-ref/service1",
			description:  "should handle custom repos_dir from config ref",
		},
		{
			name:         "service2 under config ref",
			physicalPath: service2Path,
			want:         "/config-ref/service2",
			description:  "should handle multiple children under config ref",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildTreePathFromFilesystem(tt.physicalPath)
			assert.Equal(t, tt.want, result, tt.description)
		})
	}
}

func TestManager_buildTreePathFromFilesystem_DeepNesting(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create main config
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "org", URL: "https://example.com/org"},
		},
	}
	
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	tw.CreateConfig(cfg)
	
	// Create org with muno.yaml
	orgPath := tw.AddRepository("org")
	orgCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "org",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "team", URL: "https://example.com/team"},
		},
	}
	tw.AddRepositoryWithConfig("org", orgCfg)
	
	// Create team with muno.yaml
	teamPath := tw.AddRepository("org/.nodes/team")
	teamCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "team",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "project", URL: "https://example.com/project"},
		},
	}
	tw.AddRepositoryWithConfig("org/.nodes/team", teamCfg)
	
	// Create project with muno.yaml
	projectPath := tw.AddRepository("org/.nodes/team/.nodes/project")
	projectCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "project",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "module", URL: "https://example.com/module"},
		},
	}
	tw.AddRepositoryWithConfig("org/.nodes/team/.nodes/project", projectCfg)
	
	// Create module
	modulePath := tw.AddRepository("org/.nodes/team/.nodes/project/.nodes/module")
	
	tests := []struct {
		name         string
		physicalPath string
		want         string
	}{
		{
			name:         "org level",
			physicalPath: orgPath,
			want:         "/org",
		},
		{
			name:         "team level",
			physicalPath: teamPath,
			want:         "/org/team",
		},
		{
			name:         "project level",
			physicalPath: projectPath,
			want:         "/org/team/project",
		},
		{
			name:         "module level (deepest)",
			physicalPath: modulePath,
			want:         "/org/team/project/module",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildTreePathFromFilesystem(tt.physicalPath)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestManager_buildTreePathFromFilesystem_MixedReposDirs(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	// Create main config with default repos_dir
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "standard", URL: "https://example.com/standard"},
			{Name: "custom", URL: "https://example.com/custom"},
		},
	}
	
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	tw.CreateConfig(cfg)
	
	// Create standard repo with default .nodes
	standardPath := tw.AddRepository("standard")
	standardCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "standard",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "child1", URL: "https://example.com/child1"},
		},
	}
	tw.AddRepositoryWithConfig("standard", standardCfg)
	child1Path := tw.AddRepository("standard/.nodes/child1")
	
	// Create custom repo with different repos_dir
	customPath := tw.AddRepository("custom")
	customCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "custom",
			ReposDir: "repos",
		},
		Nodes: []config.NodeDefinition{
			{Name: "child2", URL: "https://example.com/child2"},
		},
	}
	tw.AddRepositoryWithConfig("custom", customCfg)
	
	// Create child2 under custom repos dir
	child2Path := filepath.Join(customPath, "repos", "child2")
	os.MkdirAll(child2Path, 0755)
	
	tests := []struct {
		name         string
		physicalPath string
		want         string
		description  string
	}{
		{
			name:         "standard repo",
			physicalPath: standardPath,
			want:         "/standard",
			description:  "standard repo with default repos_dir",
		},
		{
			name:         "child under standard",
			physicalPath: child1Path,
			want:         "/standard/child1",
			description:  "child using default .nodes",
		},
		{
			name:         "custom repo",
			physicalPath: customPath,
			want:         "/custom",
			description:  "custom repo with different repos_dir",
		},
		{
			name:         "child under custom",
			physicalPath: child2Path,
			want:         "/custom/child2",
			description:  "child using custom repos dir",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildTreePathFromFilesystem(tt.physicalPath)
			assert.Equal(t, tt.want, result, tt.description)
		})
	}
}

func TestManager_buildTreePathFromFilesystem_PathsOutsideWorkspace(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
	}
	
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	tw.CreateConfig(cfg)
	
	tests := []struct {
		name         string
		physicalPath string
		want         string
		description  string
	}{
		{
			name:         "completely outside workspace",
			physicalPath: "/tmp/random/path",
			want:         "",
			description:  "should return empty for paths outside workspace",
		},
		{
			name:         "parent of workspace",
			physicalPath: filepath.Dir(tw.Root),
			want:         "",
			description:  "should return empty for parent directory",
		},
		{
			name:         "sibling of workspace",
			physicalPath: filepath.Join(filepath.Dir(tw.Root), "sibling"),
			want:         "",
			description:  "should return empty for sibling paths",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildTreePathFromFilesystem(tt.physicalPath)
			assert.Equal(t, tt.want, result, tt.description)
		})
	}
}

func TestManager_buildTreePathFromFilesystem_SpecialCharacters(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "repo.with.dots", URL: "https://example.com/dots"},
			{Name: "repo_with_underscores", URL: "https://example.com/underscores"},
			{Name: "repo-with-dashes", URL: "https://example.com/dashes"},
			{Name: "repo with spaces", URL: "https://example.com/spaces"},
		},
	}
	
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	tw.CreateConfig(cfg)
	
	// Create repositories with special characters
	dotsPath := tw.AddRepository("repo.with.dots")
	underscoresPath := tw.AddRepository("repo_with_underscores")
	dashesPath := tw.AddRepository("repo-with-dashes")
	spacesPath := tw.AddRepository("repo with spaces")
	
	tests := []struct {
		name         string
		physicalPath string
		want         string
	}{
		{
			name:         "repository with dots",
			physicalPath: dotsPath,
			want:         "/repo.with.dots",
		},
		{
			name:         "repository with underscores",
			physicalPath: underscoresPath,
			want:         "/repo_with_underscores",
		},
		{
			name:         "repository with dashes",
			physicalPath: dashesPath,
			want:         "/repo-with-dashes",
		},
		{
			name:         "repository with spaces",
			physicalPath: spacesPath,
			want:         "/repo with spaces",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildTreePathFromFilesystem(tt.physicalPath)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestManager_buildTreePathFromFilesystem_NonExistentPaths(t *testing.T) {
	tw := CreateTestWorkspace(t)
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "existing", URL: "https://example.com/existing"},
		},
	}
	
	m := CreateTestManagerWithConfig(t, tw.Root, cfg)
	tw.CreateConfig(cfg)
	
	// Create one existing repo for comparison
	existingPath := tw.AddRepository("existing")
	existingCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "existing",
			ReposDir: ".nodes",
		},
		Nodes: []config.NodeDefinition{
			{Name: "child", URL: "https://example.com/child"},
		},
	}
	tw.AddRepositoryWithConfig("existing", existingCfg)
	
	tests := []struct {
		name         string
		physicalPath string
		want         string
		description  string
	}{
		{
			name:         "existing repository",
			physicalPath: existingPath,
			want:         "/existing",
			description:  "existing repo should work",
		},
		{
			name:         "non-existent repo in nodes",
			physicalPath: filepath.Join(tw.NodesDir, "nonexistent"),
			want:         "",
			description:  "non-existent path should return empty (no muno.yaml to find)",
		},
		{
			name:         "non-existent nested path",
			physicalPath: filepath.Join(existingPath, ".nodes", "nonexistent"),
			want:         "",
			description:  "non-existent nested path should return empty",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildTreePathFromFilesystem(tt.physicalPath)
			assert.Equal(t, tt.want, result, tt.description)
		})
	}
}