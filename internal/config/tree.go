package config

import (
	"fmt"
	"os"
	"path/filepath"
	
	"gopkg.in/yaml.v3"
)

// ConfigTree represents the tree-based configuration
type ConfigTree struct {
	Version      int              `yaml:"version"`
	Workspace    WorkspaceTree  `yaml:"workspace"`
	Repositories []RepoDefinition `yaml:"repositories,omitempty"`
	
	// Runtime fields (not in YAML)
	Path     string `yaml:"-"`
}

// WorkspaceTree represents workspace configuration for v3
type WorkspaceTree struct {
	Name     string `yaml:"name"`
	RootRepo string `yaml:"root_repo,omitempty"` // Root is also a git repo
	ReposDir string `yaml:"repos_dir,omitempty"` // Directory for repositories (default: "repos")
}

// RepoDefinition represents a repository definition in config
type RepoDefinition struct {
	URL  string `yaml:"url"`
	Name string `yaml:"name"`
	Lazy bool   `yaml:"lazy,omitempty"`
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

// TreeState represents the current state of the tree
type TreeState struct {
	CurrentNodePath string              `json:"current_node_path"`
	Nodes          map[string]NodeMeta `json:"nodes"`
	LastUpdated    string              `json:"last_updated"`
}

// DefaultConfigTree returns the default tree configuration
func DefaultConfigTree(projectName string) *ConfigTree {
	return &ConfigTree{
		Version: 3,
		Workspace: WorkspaceTree{
			Name:     projectName,
			RootRepo: "", // Can be set if root is a repo
			ReposDir: "repos", // Default repos directory
		},
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
	if cfg.Version == 0 {
		cfg.Version = 3
	}
	if cfg.Workspace.ReposDir == "" {
		cfg.Workspace.ReposDir = "repos"
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
	if c.Version != 3 {
		return fmt.Errorf("invalid version: %d (expected 3)", c.Version)
	}
	
	if c.Workspace.Name == "" {
		return fmt.Errorf("workspace name is required")
	}

	return nil
}

// GetReposDir returns the configured repos directory name
func (c *ConfigTree) GetReposDir() string {
	if c.Workspace.ReposDir == "" {
		return "repos"
	}
	return c.Workspace.ReposDir
}

// LoadTreeState loads the tree state from JSON
func LoadTreeState(path string) (*TreeState, error) {
	// Implementation will be in tree package
	return nil, fmt.Errorf("not implemented")
}

// SaveTreeState saves the tree state to JSON
func SaveTreeState(state *TreeState, path string) error {
	// Implementation will be in tree package
	return fmt.Errorf("not implemented")
}