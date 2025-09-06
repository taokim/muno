package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"gopkg.in/yaml.v3"
)

// Fetch mode constants
const (
	FetchLazy  = "lazy"  // Clone on first use
	FetchEager = "eager" // Clone immediately
	FetchAuto  = "auto"  // Use smart detection based on repo name (default)
)

// ConfigTree represents the tree-based configuration
type ConfigTree struct {
	Workspace WorkspaceTree    `yaml:"workspace"`
	Nodes     []NodeDefinition `yaml:"nodes"`  // Flat list of direct children only
	
	// Runtime fields (not in YAML)
	Path string `yaml:"-"`  // Path to this config file
}

// WorkspaceTree represents workspace configuration for v3
type WorkspaceTree struct {
	Name     string `yaml:"name"`
	RootRepo string `yaml:"root_repo,omitempty"` // Root is also a git repo
	ReposDir string `yaml:"repos_dir,omitempty"` // Directory for repositories (default: "nodes")
}

// NodeDefinition represents a node in the distributed tree
// Node type is determined by field presence (mutually exclusive):
// - URL only: Git repository (may auto-discover muno.yaml)
// - Config only: Pure config delegation (no repository)
type NodeDefinition struct {
	Name   string `yaml:"name"`
	URL    string `yaml:"url,omitempty"`     // Git repository URL
	Config string `yaml:"config,omitempty"`  // Path to sub-configuration
	Fetch  string `yaml:"fetch,omitempty"`   // Fetch mode: "auto" (default), "lazy", or "eager"
}

// IsLazy determines if a node should be lazy based on its fetch mode
func (n *NodeDefinition) IsLazy() bool {
	switch n.Fetch {
	case FetchEager:
		return false
	case FetchLazy:
		return true
	case FetchAuto, "":
		// Default to auto mode for smart detection
		// Use smart detection based on repo name patterns
		// Check if name ends with meta-repo patterns
		name := strings.ToLower(n.Name)
		for _, pattern := range GetEagerLoadPatterns() {
			if strings.HasSuffix(name, pattern) {
				return false // It's a meta-repo, should be eager
			}
		}
		return true // Regular repo, should be lazy
	default:
		// Unknown fetch mode defaults to auto behavior
		name := strings.ToLower(n.Name)
		for _, pattern := range GetEagerLoadPatterns() {
			if strings.HasSuffix(name, pattern) {
				return false
			}
		}
		return true
	}
}

// NodeMeta represents metadata for a tree node
type NodeMeta struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	Type      string       `json:"type"` // "persistent" or "ephemeral"
	CreatedAt string       `json:"created_at"`
	Repos     []RepoConfig `json:"repos,omitempty"`
}

// RepoConfig represents a repository configuration within a node
type RepoConfig struct {
	URL   string `json:"url"`
	Path  string `json:"path"`
	Name  string `json:"name"`
	Lazy  bool   `json:"lazy"`
	State string `json:"state"` // "missing", "cloned", "modified"
}

// TreeState is deprecated - we now derive state from filesystem

// DefaultConfigTree returns the default tree configuration
func DefaultConfigTree(projectName string) *ConfigTree {
	defaults := GetDefaults()
	return &ConfigTree{
		Workspace: WorkspaceTree{
			Name:     projectName,
			RootRepo: defaults.Workspace.RootRepo,
			ReposDir: defaults.Workspace.ReposDir,
		},
		Nodes: []NodeDefinition{}, // Empty nodes list
	}
}

// LoadTree reads a tree configuration from a YAML file
func LoadTree(path string) (*ConfigTree, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg ConfigTree
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Merge with defaults - project config overrides defaults
	cfg = *MergeWithDefaults(&cfg)
	
	// Store the path
	cfg.Path = filepath.Dir(path)
	
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Save writes a tree configuration to a YAML file
func (c *ConfigTree) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// Validate validates the tree configuration
func (c *ConfigTree) Validate() error {
	if c.Workspace.Name == "" {
		return fmt.Errorf("workspace name is required")
	}
	
	// Validate each node
	for _, node := range c.Nodes {
		if node.Name == "" {
			return fmt.Errorf("node name is required")
		}
		
		// Ensure node has either URL or Config, not both
		hasURL := node.URL != ""
		hasConfig := node.Config != ""
		
		if hasURL && hasConfig {
			return fmt.Errorf("node %s cannot have both URL and config fields", node.Name)
		}
		
		if !hasURL && !hasConfig {
			return fmt.Errorf("node %s must have either URL or config field", node.Name)
		}
	}

	return nil
}

// GetReposDir returns the configured repos directory name
func (c *ConfigTree) GetReposDir() string {
	// If ReposDir is explicitly set (even to empty string or "."), use it
	// Only use default if it's not set at all (which happens after merging with defaults)
	// An empty string or "." means the workspace root is the repos directory
	return c.Workspace.ReposDir
}

// State management functions removed - now stateless operation