package config

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/taokim/muno/internal/constants"
	"gopkg.in/yaml.v3"
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
	Lazy   bool   `yaml:"lazy,omitempty"`    // Clone on-demand
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
	return &ConfigTree{
		Workspace: WorkspaceTree{
			Name:     projectName,
			RootRepo: "", // Can be set if root is a repo
			ReposDir: constants.DefaultReposDir, // Default nodes directory
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

	// Apply defaults
	if cfg.Workspace.ReposDir == "" {
		cfg.Workspace.ReposDir = constants.DefaultReposDir
	}
	
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
	if c.Workspace.ReposDir == "" {
		return constants.DefaultReposDir
	}
	return c.Workspace.ReposDir
}

// State management functions removed - now stateless operation