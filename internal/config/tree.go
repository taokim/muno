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
	Workspace     WorkspaceTree          `yaml:"workspace"`
	Nodes         []NodeDefinition       `yaml:"nodes"`  // Flat list of direct children only
	Defaults      TreeDefaults           `yaml:"defaults,omitempty"` // Default settings for repositories
	Overrides     map[string]interface{} `yaml:"overrides,omitempty"` // Workspace-level config overrides
	
	// Runtime fields (not in YAML)
	Path string `yaml:"-"`  // Path to this config file
}

// TreeDefaults contains default settings for tree-based configuration
type TreeDefaults struct {
	SSHPreference bool `yaml:"ssh_preference,omitempty"` // Default: true (prefer SSH over HTTPS for GitHub)
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
	Name          string                 `yaml:"name"`
	URL           string                 `yaml:"url,omitempty"`           // Git repository URL
	File          string                 `yaml:"file,omitempty"`           // Path to sub-configuration file
	Fetch         string                 `yaml:"fetch,omitempty"`          // Fetch mode: "auto" (default), "lazy", or "eager"
	DefaultBranch string                 `yaml:"default_branch,omitempty"` // Node's default branch override
	Overrides     map[string]interface{} `yaml:"overrides,omitempty"`     // Node-level config overrides
	Metadata      map[string]string      `yaml:"metadata,omitempty"`       // Flexible metadata key-value pairs
}

// IsLazy determines if a node should be lazy based on its fetch mode
// extractRepoNameFromURL extracts the repository name from a git URL
func extractRepoNameFromURL(url string) string {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")
	
	// Get last path component
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	
	return url
}

func (n *NodeDefinition) IsLazy() bool {
	switch n.Fetch {
	case FetchEager:
		return false
	case FetchLazy:
		return true
	case FetchAuto, "":
		// Default to auto mode for smart detection
		// Check both node name AND repository URL for eager patterns
		
		// Check node name first
		name := strings.ToLower(n.Name)
		for _, pattern := range GetEagerLoadPatterns() {
			if strings.HasSuffix(name, pattern) {
				return false // It's a meta-repo, should be eager
			}
		}
		
		// Also check repository URL if present
		if n.URL != "" {
			repoName := extractRepoNameFromURL(n.URL)
			repoName = strings.ToLower(repoName)
			for _, pattern := range GetEagerLoadPatterns() {
				if strings.HasSuffix(repoName, pattern) {
					return false // Repository name matches meta-repo pattern, should be eager
				}
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
		
		// Also check repository URL for unknown fetch modes
		if n.URL != "" {
			repoName := extractRepoNameFromURL(n.URL)
			repoName = strings.ToLower(repoName)
			for _, pattern := range GetEagerLoadPatterns() {
				if strings.HasSuffix(repoName, pattern) {
					return false
				}
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
		Defaults: TreeDefaults{
			SSHPreference: true, // Enable SSH preference by default
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

// LoadTreeReposDir loads a config and returns the repos_dir as explicitly set in the YAML
// (without applying defaults). Returns empty string if repos_dir is not set.
// This is useful for path resolution where we need to distinguish between
// "user explicitly set repos_dir to .nodes" vs "repos_dir not set (use parent directory directly)"
func LoadTreeReposDir(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading config file: %w", err)
	}

	var cfg ConfigTree
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parsing config: %w", err)
	}

	// Return the repos_dir as specified in the YAML (without defaults)
	return cfg.Workspace.ReposDir, nil
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
		
		// Ensure node has either URL or File, not both
		hasURL := node.URL != ""
		hasFile := node.File != ""
		
		if hasURL && hasFile {
			return fmt.Errorf("node %s cannot have both URL and file fields", node.Name)
		}
		
		if !hasURL && !hasFile {
			return fmt.Errorf("node %s must have either URL or file field", node.Name)
		}
	}

	return nil
}

// GetNodesDir returns the configured nodes directory name
func (c *ConfigTree) GetNodesDir() string {
	// If ReposDir is empty, return the default
	if c.Workspace.ReposDir == "" {
		return GetDefaultNodesDir()
	}
	// Otherwise return the configured value
	return c.Workspace.ReposDir
}

// GetReposDir is deprecated, use GetNodesDir instead
// Kept for backward compatibility
func (c *ConfigTree) GetReposDir() string {
	return c.GetNodesDir()
}

// State management functions removed - now stateless operation