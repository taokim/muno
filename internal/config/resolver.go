package config

import (
	"fmt"
	"strings"
)

// ConfigResolver handles multi-level configuration resolution
type ConfigResolver struct {
	defaults  map[string]interface{} // From defaults.yaml
	workspace map[string]interface{} // From workspace config
	cli       map[string]interface{} // From CLI flags
}

// NewConfigResolver creates a new configuration resolver
func NewConfigResolver(defaults *DefaultConfiguration) *ConfigResolver {
	return &ConfigResolver{
		defaults:  defaultsToMap(defaults),
		workspace: make(map[string]interface{}),
		cli:       make(map[string]interface{}),
	}
}

// SetWorkspaceConfig sets the workspace-level configuration
func (r *ConfigResolver) SetWorkspaceConfig(config map[string]interface{}) {
	r.workspace = config
}

// SetCLIConfig sets the CLI-level configuration overrides
func (r *ConfigResolver) SetCLIConfig(config map[string]interface{}) {
	r.cli = config
}

// ResolveForNode resolves configuration for a specific node
// Priority: CLI > Node > Workspace > Defaults
func (r *ConfigResolver) ResolveForNode(node *NodeDefinition) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Start with defaults
	deepMerge(result, r.defaults)
	
	// Override with workspace config
	deepMerge(result, r.workspace)
	
	// Override with node config (if present)
	if node != nil && node.Overrides != nil {
		deepMerge(result, node.Overrides)
	}
	
	// Override with CLI config (highest priority)
	deepMerge(result, r.cli)
	
	return result
}

// ResolveForWorkspace resolves configuration for workspace-level operations
// Priority: CLI > Workspace > Defaults
func (r *ConfigResolver) ResolveForWorkspace() map[string]interface{} {
	result := make(map[string]interface{})
	
	// Start with defaults
	deepMerge(result, r.defaults)
	
	// Override with workspace config
	deepMerge(result, r.workspace)
	
	// Override with CLI config (highest priority)
	deepMerge(result, r.cli)
	
	return result
}

// GetValue gets a specific configuration value by path (e.g., "git.default_branch")
func (r *ConfigResolver) GetValue(path string, node *NodeDefinition) interface{} {
	var config map[string]interface{}
	
	if node != nil {
		config = r.ResolveForNode(node)
	} else {
		config = r.ResolveForWorkspace()
	}
	
	return getByPath(config, path)
}

// GetDefaultBranch resolves the default branch for a node
func (r *ConfigResolver) GetDefaultBranch(node *NodeDefinition) string {
	// First check node's explicit default_branch field
	if node != nil && node.DefaultBranch != "" {
		return node.DefaultBranch
	}
	
	// Then check resolved config hierarchy
	if branch := r.GetValue("git.default_branch", node); branch != nil {
		if str, ok := branch.(string); ok {
			return str
		}
	}
	
	// Final fallback
	return "main"
}

// Node-specific configuration keys
var nodeSpecificKeys = map[string]bool{
	"git.default_branch":  true,
	"git.default_remote":  true,
	"git.shallow_depth":   true,
	"fetch":              true,
}

// IsNodeSpecific checks if a configuration key is node-specific
func IsNodeSpecific(key string) bool {
	return nodeSpecificKeys[key]
}

// Helper functions

// deepMerge merges src into dst, overwriting values in dst
func deepMerge(dst, src map[string]interface{}) {
	for key, srcVal := range src {
		if srcMap, ok := srcVal.(map[string]interface{}); ok {
			if dstMap, ok := dst[key].(map[string]interface{}); ok {
				// Both are maps, merge recursively
				deepMerge(dstMap, srcMap)
			} else {
				// Source is map but dst is not, replace
				dst[key] = deepCopy(srcMap)
			}
		} else {
			// Not a map, simple override
			dst[key] = srcVal
		}
	}
}

// deepCopy creates a deep copy of a map
func deepCopy(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		if vm, ok := v.(map[string]interface{}); ok {
			result[k] = deepCopy(vm)
		} else {
			result[k] = v
		}
	}
	return result
}

// getByPath gets a value from a map by dot-separated path
func getByPath(m map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := m
	
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part, return the value
			return current[part]
		}
		
		// Navigate deeper
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}
	
	return nil
}

// setByPath sets a value in a map by dot-separated path
func setByPath(m map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := m
	
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part, set the value
			current[part] = value
			return
		}
		
		// Navigate deeper, creating maps as needed
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			next := make(map[string]interface{})
			current[part] = next
			current = next
		}
	}
}

// ParseConfigOverrides parses key=value config overrides into a map
func ParseConfigOverrides(overrides []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	for _, override := range overrides {
		parts := strings.SplitN(override, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid config override format: %s (expected key=value)", override)
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Try to parse value as bool or number
		var parsedValue interface{}
		if value == "true" {
			parsedValue = true
		} else if value == "false" {
			parsedValue = false
		} else if strings.Contains(key, "timeout") || strings.Contains(key, "depth") || 
		          strings.Contains(key, "parallel") || strings.Contains(key, "size") {
			// These are likely numeric values
			var intVal int
			if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
				parsedValue = intVal
			} else {
				parsedValue = value
			}
		} else {
			parsedValue = value
		}
		
		setByPath(result, key, parsedValue)
	}
	
	return result, nil
}

// defaultsToMap converts the DefaultConfiguration struct to a map for merging
func defaultsToMap(d *DefaultConfiguration) map[string]interface{} {
	// This would ideally use reflection or be generated
	// For now, manually construct the map based on defaults.yaml structure
	return map[string]interface{}{
		"workspace": map[string]interface{}{
			"name":      d.Workspace.Name,
			"repos_dir": d.Workspace.ReposDir,
			"root_repo": d.Workspace.RootRepo,
		},
		"git": map[string]interface{}{
			"default_remote": d.Git.DefaultRemote,
			"default_branch": d.Git.DefaultBranch,
			"clone_timeout":  d.Git.CloneTimeout,
			"shallow_depth":  d.Git.ShallowDepth,
		},
		"behavior": map[string]interface{}{
			"auto_clone_on_nav":    d.Behavior.AutoCloneOnNav,
			"show_progress":        d.Behavior.ShowProgress,
			"max_parallel_clones":  d.Behavior.MaxParallelClones,
			"max_parallel_pulls":   d.Behavior.MaxParallelPulls,
			"interactive":          d.Behavior.Interactive,
		},
		"detection": map[string]interface{}{
			"eager_patterns":  d.Detection.EagerPatterns,
			"ignore_patterns": d.Detection.IgnorePatterns,
		},
		"files": map[string]interface{}{
			"config_names":      d.Files.ConfigNames,
			"state_file":        d.Files.StateFile,
			"legacy_state_file": d.Files.LegacyStateFile,
		},
	}
}