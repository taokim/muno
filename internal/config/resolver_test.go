package config

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigResolver(t *testing.T) {
	t.Run("NewConfigResolver", func(t *testing.T) {
		defaults := GetDefaults()
		resolver := NewConfigResolver(defaults)
		
		assert.NotNil(t, resolver)
		assert.NotNil(t, resolver.defaults)
		assert.NotNil(t, resolver.workspace)
		assert.NotNil(t, resolver.cli)
	})
	
	t.Run("SetWorkspaceConfig", func(t *testing.T) {
		resolver := NewConfigResolver(GetDefaults())
		
		workspaceConfig := map[string]interface{}{
			"git": map[string]interface{}{
				"default_branch": "develop",
			},
		}
		
		resolver.SetWorkspaceConfig(workspaceConfig)
		assert.Equal(t, workspaceConfig, resolver.workspace)
	})
	
	t.Run("SetCLIConfig", func(t *testing.T) {
		resolver := NewConfigResolver(GetDefaults())
		
		cliConfig := map[string]interface{}{
			"git": map[string]interface{}{
				"default_branch": "feature",
			},
		}
		
		resolver.SetCLIConfig(cliConfig)
		assert.Equal(t, cliConfig, resolver.cli)
	})
	
	t.Run("ResolveForWorkspace", func(t *testing.T) {
		resolver := NewConfigResolver(GetDefaults())
		
		// Set workspace config
		resolver.SetWorkspaceConfig(map[string]interface{}{
			"git": map[string]interface{}{
				"default_branch": "develop",
				"clone_timeout": 600,
			},
		})
		
		// Set CLI config (should override workspace)
		resolver.SetCLIConfig(map[string]interface{}{
			"git": map[string]interface{}{
				"default_branch": "hotfix",
			},
		})
		
		config := resolver.ResolveForWorkspace()
		
		// CLI should override workspace
		gitConfig := config["git"].(map[string]interface{})
		assert.Equal(t, "hotfix", gitConfig["default_branch"])
		assert.Equal(t, 600, gitConfig["clone_timeout"]) // From workspace
		assert.Equal(t, "origin", gitConfig["default_remote"]) // From defaults
	})
	
	t.Run("ResolveForNode", func(t *testing.T) {
		resolver := NewConfigResolver(GetDefaults())
		
		// Set workspace config
		resolver.SetWorkspaceConfig(map[string]interface{}{
			"git": map[string]interface{}{
				"default_branch": "develop",
			},
		})
		
		// Create node with config
		node := &NodeDefinition{
			Name: "test-node",
			DefaultBranch: "master", // First-class field
			Overrides: map[string]interface{}{
				"git": map[string]interface{}{
					"shallow_depth": 1,
				},
			},
		}
		
		config := resolver.ResolveForNode(node)
		
		gitConfig := config["git"].(map[string]interface{})
		assert.Equal(t, "develop", gitConfig["default_branch"]) // From workspace (node doesn't override in Config map)
		assert.Equal(t, 1, gitConfig["shallow_depth"]) // From node
	})
	
	t.Run("GetDefaultBranch", func(t *testing.T) {
		resolver := NewConfigResolver(GetDefaults())
		
		// Test with node that has DefaultBranch field
		node1 := &NodeDefinition{
			Name: "test-node",
			DefaultBranch: "master",
		}
		assert.Equal(t, "master", resolver.GetDefaultBranch(node1))
		
		// Test with node without DefaultBranch field
		node2 := &NodeDefinition{
			Name: "test-node2",
		}
		
		// Set workspace config
		resolver.SetWorkspaceConfig(map[string]interface{}{
			"git": map[string]interface{}{
				"default_branch": "develop",
			},
		})
		
		// Should use resolved config
		assert.Equal(t, "develop", resolver.GetDefaultBranch(node2))
		
		// Test with nil node
		assert.Equal(t, "develop", resolver.GetDefaultBranch(nil))
		
		// Test with CLI override
		resolver.SetCLIConfig(map[string]interface{}{
			"git": map[string]interface{}{
				"default_branch": "hotfix",
			},
		})
		assert.Equal(t, "hotfix", resolver.GetDefaultBranch(node2))
	})
	
	t.Run("GetValue", func(t *testing.T) {
		resolver := NewConfigResolver(GetDefaults())
		
		resolver.SetWorkspaceConfig(map[string]interface{}{
			"git": map[string]interface{}{
				"default_branch": "develop",
				"clone_timeout": 600,
			},
			"behavior": map[string]interface{}{
				"max_parallel_clones": 8,
			},
		})
		
		// Get nested value
		assert.Equal(t, "develop", resolver.GetValue("git.default_branch", nil))
		assert.Equal(t, 600, resolver.GetValue("git.clone_timeout", nil))
		assert.Equal(t, 8, resolver.GetValue("behavior.max_parallel_clones", nil))
		
		// Get non-existent value
		assert.Nil(t, resolver.GetValue("non.existent.path", nil))
		
		// Get with node override
		node := &NodeDefinition{
			Overrides: map[string]interface{}{
				"git": map[string]interface{}{
					"clone_timeout": 900,
				},
			},
		}
		assert.Equal(t, 900, resolver.GetValue("git.clone_timeout", node))
	})
	
	t.Run("ParseConfigOverrides", func(t *testing.T) {
		// Test valid overrides
		overrides := []string{
			"git.default_branch=develop",
			"git.clone_timeout=600",
			"behavior.interactive=false",
			"behavior.max_parallel_clones=8",
		}
		
		config, err := ParseConfigOverrides(overrides)
		require.NoError(t, err)
		
		gitConfig := config["git"].(map[string]interface{})
		assert.Equal(t, "develop", gitConfig["default_branch"])
		assert.Equal(t, 600, gitConfig["clone_timeout"])
		
		behaviorConfig := config["behavior"].(map[string]interface{})
		assert.Equal(t, false, behaviorConfig["interactive"])
		assert.Equal(t, 8, behaviorConfig["max_parallel_clones"])
		
		// Test invalid format
		_, err = ParseConfigOverrides([]string{"invalid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid config override format")
		
		// Test boolean parsing
		config, err = ParseConfigOverrides([]string{
			"feature.enabled=true",
			"feature.disabled=false",
		})
		require.NoError(t, err)
		featureConfig := config["feature"].(map[string]interface{})
		assert.Equal(t, true, featureConfig["enabled"])
		assert.Equal(t, false, featureConfig["disabled"])
	})
	
	t.Run("IsNodeSpecific", func(t *testing.T) {
		assert.True(t, IsNodeSpecific("git.default_branch"))
		assert.True(t, IsNodeSpecific("git.default_remote"))
		assert.True(t, IsNodeSpecific("git.shallow_depth"))
		assert.True(t, IsNodeSpecific("fetch"))
		
		assert.False(t, IsNodeSpecific("behavior.max_parallel_clones"))
		assert.False(t, IsNodeSpecific("behavior.interactive"))
		assert.False(t, IsNodeSpecific("git.clone_timeout"))
	})
	
	t.Run("deepMerge", func(t *testing.T) {
		dst := map[string]interface{}{
			"git": map[string]interface{}{
				"default_branch": "main",
				"clone_timeout": 300,
			},
			"behavior": map[string]interface{}{
				"interactive": true,
			},
		}
		
		src := map[string]interface{}{
			"git": map[string]interface{}{
				"default_branch": "develop", // Override
				"shallow_depth": 1, // New field
			},
			"files": map[string]interface{}{ // New section
				"state_file": ".state",
			},
		}
		
		deepMerge(dst, src)
		
		gitConfig := dst["git"].(map[string]interface{})
		assert.Equal(t, "develop", gitConfig["default_branch"]) // Overridden
		assert.Equal(t, 300, gitConfig["clone_timeout"]) // Preserved
		assert.Equal(t, 1, gitConfig["shallow_depth"]) // Added
		
		behaviorConfig := dst["behavior"].(map[string]interface{})
		assert.Equal(t, true, behaviorConfig["interactive"]) // Preserved
		
		filesConfig := dst["files"].(map[string]interface{})
		assert.Equal(t, ".state", filesConfig["state_file"]) // Added
	})
	
	t.Run("getByPath", func(t *testing.T) {
		m := map[string]interface{}{
			"git": map[string]interface{}{
				"default_branch": "main",
				"config": map[string]interface{}{
					"user": map[string]interface{}{
						"name": "Test User",
					},
				},
			},
		}
		
		assert.Equal(t, "main", getByPath(m, "git.default_branch"))
		assert.Equal(t, "Test User", getByPath(m, "git.config.user.name"))
		assert.Nil(t, getByPath(m, "non.existent"))
		assert.Nil(t, getByPath(m, "git.non.existent"))
	})
	
	t.Run("setByPath", func(t *testing.T) {
		m := make(map[string]interface{})
		
		setByPath(m, "git.default_branch", "develop")
		setByPath(m, "git.config.user.name", "Test User")
		setByPath(m, "behavior.interactive", true)
		
		gitConfig := m["git"].(map[string]interface{})
		assert.Equal(t, "develop", gitConfig["default_branch"])
		
		gitUserConfig := gitConfig["config"].(map[string]interface{})["user"].(map[string]interface{})
		assert.Equal(t, "Test User", gitUserConfig["name"])
		
		behaviorConfig := m["behavior"].(map[string]interface{})
		assert.Equal(t, true, behaviorConfig["interactive"])
	})
}

func TestNodeDefinitionIsLazy(t *testing.T) {
	tests := []struct {
		name     string
		node     NodeDefinition
		expected bool
	}{
		{
			name: "explicit lazy",
			node: NodeDefinition{Name: "test", Fetch: FetchLazy},
			expected: true,
		},
		{
			name: "explicit eager",
			node: NodeDefinition{Name: "test", Fetch: FetchEager},
			expected: false,
		},
		{
			name: "auto with monorepo suffix",
			node: NodeDefinition{Name: "test-monorepo", Fetch: FetchAuto},
			expected: false, // Should be eager
		},
		{
			name: "auto with regular name",
			node: NodeDefinition{Name: "regular-service", Fetch: FetchAuto},
			expected: true, // Should be lazy
		},
		{
			name: "empty fetch mode defaults to auto",
			node: NodeDefinition{Name: "regular-service", Fetch: ""},
			expected: true,
		},
		{
			name: "unknown fetch mode defaults to auto behavior",
			node: NodeDefinition{Name: "test", Fetch: "unknown"},
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.node.IsLazy())
		})
	}
}