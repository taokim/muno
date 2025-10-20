package config

import (
	_ "embed"
	"fmt"
	"reflect"
	
	"gopkg.in/yaml.v3"
)

// Embed the default configuration file at compile time
//go:embed defaults.yaml
var defaultConfigYAML string

// DefaultConfiguration holds the complete default configuration settings
type DefaultConfiguration struct {
	Workspace WorkspaceDefaults `yaml:"workspace"`
	Detection DetectionDefaults `yaml:"detection"`
	Files     FilesDefaults     `yaml:"files"`
	Git       GitDefaults       `yaml:"git"`
	Display   DisplayDefaults   `yaml:"display"`
	Behavior  BehaviorDefaults  `yaml:"behavior"`
}

// WorkspaceDefaults contains default workspace settings
type WorkspaceDefaults struct {
	Name     string `yaml:"name"`
	ReposDir string `yaml:"repos_dir"`
	RootRepo string `yaml:"root_repo"`
}

// DetectionDefaults contains repository detection patterns
type DetectionDefaults struct {
	EagerPatterns  []string `yaml:"eager_patterns"`
	IgnorePatterns []string `yaml:"ignore_patterns"`
}

// FilesDefaults contains file discovery settings
type FilesDefaults struct {
	ConfigNames      []string `yaml:"config_names"`
	StateFile        string   `yaml:"state_file"`
	LegacyStateFile  string   `yaml:"legacy_state_file"`
}

// GitDefaults contains git-related defaults
type GitDefaults struct {
	DefaultRemote string `yaml:"default_remote"`
	DefaultBranch string `yaml:"default_branch"`
	CloneTimeout  int    `yaml:"clone_timeout"`
	ShallowDepth  int    `yaml:"shallow_depth"`
}

// DisplayDefaults contains display settings
type DisplayDefaults struct {
	Tree  TreeDisplay  `yaml:"tree"`
	Icons IconDisplay  `yaml:"icons"`
}

// TreeDisplay contains tree display characters
type TreeDisplay struct {
	Indent     string `yaml:"indent"`
	Branch     string `yaml:"branch"`
	LastBranch string `yaml:"last_branch"`
	Vertical   string `yaml:"vertical"`
	Space      string `yaml:"space"`
}

// IconDisplay contains status icons
type IconDisplay struct {
	Workspace string `yaml:"workspace"`
	Cloned    string `yaml:"cloned"`
	Lazy      string `yaml:"lazy"`
	Modified  string `yaml:"modified"`
	Ahead     string `yaml:"ahead"`
	Behind    string `yaml:"behind"`
	Diverged  string `yaml:"diverged"`
	Error     string `yaml:"error"`
	Success   string `yaml:"success"`
	Info      string `yaml:"info"`
}

// BehaviorDefaults contains behavior settings
type BehaviorDefaults struct {
	AutoCloneOnNav    bool `yaml:"auto_clone_on_nav"`
	ShowProgress      bool `yaml:"show_progress"`
	MaxParallelClones int  `yaml:"max_parallel_clones"`
	MaxParallelPulls  int  `yaml:"max_parallel_pulls"`
	Interactive       bool `yaml:"interactive"`
}

var (
	// defaultConfig is the parsed default configuration
	defaultConfig *DefaultConfiguration
)

// init loads and parses the embedded default configuration
func init() {
	defaultConfig = &DefaultConfiguration{}
	if err := yaml.Unmarshal([]byte(defaultConfigYAML), defaultConfig); err != nil {
		// This should never happen since we control the embedded YAML
		panic(fmt.Sprintf("failed to parse embedded default config: %v", err))
	}
}

// GetDefaults returns a copy of the default configuration
func GetDefaults() *DefaultConfiguration {
	// Return a copy to prevent modification of the global default
	copied := *defaultConfig
	return &copied
}

// GetDefaultWorkspaceName returns the default workspace name
func GetDefaultWorkspaceName() string {
	return defaultConfig.Workspace.Name
}

// GetDefaultNodesDir returns the default nodes directory name (for all node types)
func GetDefaultNodesDir() string {
	return defaultConfig.Workspace.ReposDir
}

// GetDefaultReposDir is deprecated, use GetDefaultNodesDir instead
// Kept for backward compatibility
func GetDefaultReposDir() string {
	return GetDefaultNodesDir()
}

// GetEagerLoadPatterns returns the patterns that trigger eager loading
func GetEagerLoadPatterns() []string {
	// Return a copy of the slice
	patterns := make([]string, len(defaultConfig.Detection.EagerPatterns))
	copy(patterns, defaultConfig.Detection.EagerPatterns)
	return patterns
}

// GetIgnorePatterns returns the patterns to ignore during scanning
func GetIgnorePatterns() []string {
	patterns := make([]string, len(defaultConfig.Detection.IgnorePatterns))
	copy(patterns, defaultConfig.Detection.IgnorePatterns)
	return patterns
}

// GetConfigFileNames returns the config file names to search for
func GetConfigFileNames() []string {
	names := make([]string, len(defaultConfig.Files.ConfigNames))
	copy(names, defaultConfig.Files.ConfigNames)
	return names
}

// GetStateFileName returns the state file name
func GetStateFileName() string {
	return defaultConfig.Files.StateFile
}

// GetTreeDisplay returns tree display settings
func GetTreeDisplay() TreeDisplay {
	return defaultConfig.Display.Tree
}

// GetIcons returns display icons
func GetIcons() IconDisplay {
	return defaultConfig.Display.Icons
}

// MergeWithDefaults merges a project configuration with defaults
// Project config values override defaults where specified
func MergeWithDefaults(projectConfig *ConfigTree) *ConfigTree {
	if projectConfig == nil {
		projectConfig = &ConfigTree{}
	}
	
	// Apply defaults to workspace
	if projectConfig.Workspace.Name == "" {
		projectConfig.Workspace.Name = defaultConfig.Workspace.Name
	}
	// Only apply default repos_dir if not explicitly set
	// "." means use workspace root, empty string gets default
	if projectConfig.Workspace.ReposDir == "" {
		projectConfig.Workspace.ReposDir = defaultConfig.Workspace.ReposDir
	}
	
	// Note: We don't override Nodes as those are project-specific
	
	return projectConfig
}

// MergeConfigs performs a deep merge of two configurations
// The source config values override the destination where both exist
func MergeConfigs(dest, src interface{}) interface{} {
	// Use reflection for deep merging
	if src == nil {
		return dest
	}
	if dest == nil {
		return src
	}
	
	destValue := reflect.ValueOf(dest)
	srcValue := reflect.ValueOf(src)
	
	// Handle pointers
	if destValue.Kind() == reflect.Ptr {
		destValue = destValue.Elem()
	}
	if srcValue.Kind() == reflect.Ptr {
		srcValue = srcValue.Elem()
	}
	
	// Only merge if both are structs
	if destValue.Kind() != reflect.Struct || srcValue.Kind() != reflect.Struct {
		return src
	}
	
	// Iterate through source fields
	for i := 0; i < srcValue.NumField(); i++ {
		srcField := srcValue.Field(i)
		destField := destValue.Field(i)
		
		// Skip zero values in source (use defaults)
		if isZeroValue(srcField) {
			continue
		}
		
		// For nested structs, recurse
		if srcField.Kind() == reflect.Struct {
			merged := MergeConfigs(destField.Interface(), srcField.Interface())
			if destField.CanSet() {
				destField.Set(reflect.ValueOf(merged))
			}
		} else if destField.CanSet() {
			// For non-structs, just override
			destField.Set(srcField)
		}
	}
	
	return dest
}

// isZeroValue checks if a reflect.Value is a zero value
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}