package config

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestEmbeddedDefaults(t *testing.T) {
	// Test that defaults are loaded
	defaults := GetDefaults()
	
	assert.NotNil(t, defaults)
	assert.Equal(t, "muno-workspace", defaults.Workspace.Name)
	// Get the actual default nodes directory  
	expectedReposDir := GetDefaultNodesDir()
	assert.Equal(t, expectedReposDir, defaults.Workspace.ReposDir)
	
	// Test eager patterns
	patterns := GetEagerLoadPatterns()
	assert.Contains(t, patterns, "-monorepo")
	assert.Contains(t, patterns, "-root-repo")
	// DO NOT TEST FOR "-platform" - intentionally excluded to avoid conflicts with real platform repos
	assert.Contains(t, patterns, "-workspace")
	
	// Test config file names
	names := GetConfigFileNames()
	assert.Contains(t, names, "muno.yaml")
	assert.Contains(t, names, ".muno.yaml")
	
	// Test state file name
	assert.Equal(t, ".muno-tree.json", GetStateFileName())
	
	// Test tree display
	display := GetTreeDisplay()
	assert.Equal(t, "  ", display.Indent)
	assert.Equal(t, "├── ", display.Branch)
	
	// Test icons
	icons := GetIcons()
	assert.Equal(t, "🌳", icons.Workspace)
	assert.Equal(t, "📦", icons.Cloned)
	assert.Equal(t, "💤", icons.Lazy)
}

func TestConfigurationOverride(t *testing.T) {
	tmpDir := t.TempDir()
	
	tests := []struct {
		name           string
		configContent  string
		expectedName   string
		expectedRepos  string
	}{
		{
			name: "empty config uses defaults",
			configContent: ``,
			expectedName:  "muno-workspace",
			expectedRepos: GetDefaultNodesDir(),
		},
		{
			name: "partial override - name only",
			configContent: `workspace:
  name: my-project`,
			expectedName:  "my-project",
			expectedRepos: GetDefaultNodesDir(), // should use default
		},
		{
			name: "partial override - repos_dir only",
			configContent: `workspace:
  repos_dir: repositories`,
			expectedName:  "muno-workspace", // should use default
			expectedRepos: "repositories",
		},
		{
			name: "full override",
			configContent: `workspace:
  name: custom-name
  repos_dir: custom-dir`,
			expectedName:  "custom-name",
			expectedRepos: "custom-dir",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test config
			configPath := filepath.Join(tmpDir, "test.yaml")
			err := os.WriteFile(configPath, []byte(tt.configContent), 0644)
			require.NoError(t, err)
			
			// Load and verify
			cfg, err := LoadTree(configPath)
			require.NoError(t, err)
			
			assert.Equal(t, tt.expectedName, cfg.Workspace.Name)
			assert.Equal(t, tt.expectedRepos, cfg.Workspace.ReposDir)
			assert.Equal(t, tt.expectedRepos, cfg.GetReposDir())
		})
	}
}

func TestMergeWithDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    *ConfigTree
		expected *ConfigTree
	}{
		{
			name:  "nil config gets defaults",
			input: nil,
			expected: &ConfigTree{
				Workspace: WorkspaceTree{
					Name:     "muno-workspace",
					ReposDir: GetDefaultNodesDir(),
				},
			},
		},
		{
			name: "empty config gets defaults",
			input: &ConfigTree{},
			expected: &ConfigTree{
				Workspace: WorkspaceTree{
					Name:     "muno-workspace",
					ReposDir: GetDefaultNodesDir(),
				},
			},
		},
		{
			name: "partial config merges with defaults",
			input: &ConfigTree{
				Workspace: WorkspaceTree{
					Name: "my-project",
				},
			},
			expected: &ConfigTree{
				Workspace: WorkspaceTree{
					Name:     "my-project",
					ReposDir: GetDefaultNodesDir(),
				},
			},
		},
		{
			name: "full config overrides defaults",
			input: &ConfigTree{
				Workspace: WorkspaceTree{
					Name:     "custom",
					ReposDir: GetDefaultNodesDir(),
				},
			},
			expected: &ConfigTree{
				Workspace: WorkspaceTree{
					Name:     "custom",
					ReposDir: GetDefaultNodesDir(),
				},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeWithDefaults(tt.input)
			assert.Equal(t, tt.expected.Workspace.Name, result.Workspace.Name)
			assert.Equal(t, tt.expected.Workspace.ReposDir, result.Workspace.ReposDir)
		})
	}
}

func TestDefaultsYAMLParsing(t *testing.T) {
	// Test that the embedded YAML can be parsed
	var cfg DefaultConfiguration
	err := yaml.Unmarshal([]byte(defaultConfigYAML), &cfg)
	require.NoError(t, err)
	
	// Verify structure
	assert.NotEmpty(t, cfg.Workspace.Name)
	assert.NotEmpty(t, cfg.Workspace.ReposDir)
	assert.NotEmpty(t, cfg.Detection.EagerPatterns)
	assert.NotEmpty(t, cfg.Detection.IgnorePatterns)
	assert.NotEmpty(t, cfg.Files.ConfigNames)
	assert.NotEmpty(t, cfg.Files.StateFile)
	assert.NotEmpty(t, cfg.Git.DefaultBranch)
	assert.NotEmpty(t, cfg.Display.Tree.Indent)
	assert.NotEmpty(t, cfg.Display.Icons.Workspace)
	assert.True(t, cfg.Behavior.AutoCloneOnNav)
}