package config

import (
	"fmt"
	"os"
	"path/filepath"
	
	"gopkg.in/yaml.v3"
)

// ConfigV3Tree represents the v3 tree-based configuration
type ConfigV3Tree struct {
	Version   int              `yaml:"version"`
	Workspace WorkspaceV3Tree  `yaml:"workspace"`
	
	// Runtime fields (not in YAML)
	Path     string `yaml:"-"`
}

// WorkspaceV3Tree represents workspace configuration for v3
type WorkspaceV3Tree struct {
	Name     string `yaml:"name"`
	RootRepo string `yaml:"root_repo"` // Root is also a git repo
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

// DefaultConfigV3Tree returns the default tree configuration
func DefaultConfigV3Tree(projectName string) *ConfigV3Tree {
	return &ConfigV3Tree{
		Version: 3,
		Workspace: WorkspaceV3Tree{
			Name:     projectName,
			RootRepo: "", // Can be set if root is a repo
		},
	}
}

// LoadV3Tree reads a v3 tree configuration from a YAML file
func LoadV3Tree(path string) (*ConfigV3Tree, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg ConfigV3Tree
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Apply defaults
	if cfg.Version == 0 {
		cfg.Version = 3
	}
	
	// Store the path
	cfg.Path = filepath.Dir(path)
	
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Save writes a v3 tree configuration to a YAML file
func (c *ConfigV3Tree) Save(path string) error {
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

// Validate validates the v3 tree configuration
func (c *ConfigV3Tree) Validate() error {
	if c.Version != 3 {
		return fmt.Errorf("invalid version: %d (expected 3)", c.Version)
	}
	
	if c.Workspace.Name == "" {
		return fmt.Errorf("workspace name is required")
	}

	return nil
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