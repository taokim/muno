package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	
	"gopkg.in/yaml.v3"
)

// DefaultMetaRepoPattern is the default regex pattern to detect meta-repos
const DefaultMetaRepoPattern = `(?i)(-(repo|monorepo|rc|meta)$)`

// Config represents the simplified configuration structure
type Config struct {
	Workspace     WorkspaceConfig          `yaml:"workspace"`
	Defaults      Defaults                 `yaml:"defaults,omitempty"`
	Repositories  map[string]Repository  `yaml:"repositories"`
	// Scopes - removed in tree-based architecture
	// Scopes        map[string]Scope       `yaml:"scopes"`
	Documentation DocumentationConfig      `yaml:"documentation,omitempty"`
	
	// Runtime fields (not in YAML)
	Path          string                 `yaml:"-"`
	Parent        *Config              `yaml:"-"`
	Depth         int                    `yaml:"-"`
	Children      map[string]*Config   `yaml:"-"` // Lazy-loaded child workspaces
}

// Defaults contains default settings for repositories
type Defaults struct {
	Lazy         bool   `yaml:"lazy,omitempty"`          // Default: true (lazy load by default)
	LazyPattern  string `yaml:"lazy_pattern,omitempty"`  // Regex for repos to lazy-load
	EagerPattern string `yaml:"eager_pattern,omitempty"` // Regex for repos to eager-load (meta-repos)
}

// Repository represents a simplified repository definition
type Repository struct {
	URL    string  `yaml:"url"`
	Branch string  `yaml:"branch,omitempty"`
	Groups []string `yaml:"groups,omitempty"`
	Lazy   *bool   `yaml:"lazy,omitempty"`  // nil = auto-detect based on patterns
	
	// Runtime fields (not in YAML)
	IsWorkspace bool      `yaml:"-"` // Auto-detected: true if contains muno.yaml
	Config      *Config `yaml:"-"` // Loaded config if it's a workspace
	Path        string    `yaml:"-"` // Local path after cloning
}

// Scope represents a scope that can reference both local repos and child workspace scopes
type Scope struct {
	Type            string              `yaml:"type"`                       // "persistent" or "ephemeral"
	Repos           []string            `yaml:"repos,omitempty"`            // Local repository references
	Description     string              `yaml:"description"`                // Human-readable description
	Model           string              `yaml:"model"`                      // Claude model to use
	AutoStart       bool                `yaml:"auto_start"`                 // Whether to auto-start this scope
	WorkspaceScopes map[string][]string `yaml:"workspace_scopes,omitempty"` // workspace -> scopes mapping
}

// DefaultDefaults returns the default configuration for repository loading
func DefaultDefaults() Defaults {
	return Defaults{
		Lazy:         true,                            // Everything lazy by default
		EagerPattern: DefaultMetaRepoPattern,          // Meta-repos fetched eagerly
	}
}

// DefaultConfig returns the default configuration template
func DefaultConfig(projectName string) *Config {
	return &Config{
		Workspace: WorkspaceConfig{
			Name:     projectName,
			RootPath: "repos",
		},
		Defaults: DefaultDefaults(),
		Repositories: map[string]Repository{
			// Meta-repos (will be eager-loaded due to naming)
			"backend-repo": {
				URL:    "https://github.com/yourorg/backend-repo.git",
				Branch: "main",
			},
			"frontend-repo": {
				URL:    "https://github.com/yourorg/frontend-repo.git",
				Branch: "main",
			},
			
			// Regular repos (will be lazy-loaded)
			"payment-service": {
				URL:    "https://github.com/yourorg/payment-service.git",
				Branch: "main",
				Groups: []string{"backend", "services"},
			},
			"fraud-detection": {
				URL:    "https://github.com/yourorg/fraud-detection.git",
				Branch: "main",
				Groups: []string{"backend", "ml"},
			},
			"web-app": {
				URL:    "https://github.com/yourorg/web-app.git",
				Branch: "main",
				Groups: []string{"frontend", "ui"},
			},
		},
		// Scopes removed in tree-based architecture
		// Scopes: map[string]Scope{
		// 	"backend": {
		// 		Type:        "persistent",
		// 		Repos:       []string{"payment-service", "fraud-detection"},
		// 		Description: "Backend services development",
		// 		Model:       "claude-3-5-sonnet-20241022",
		// 	},
		// 	"fullstack": {
		// 		Type:        "persistent",
		// 		Repos:       []string{"payment-service", "web-app"},
		// 		Description: "Full-stack development",
		// 		Model:       "claude-3-5-sonnet-20241022",
		// 	},
		// },
		Documentation: DocumentationConfig{
			Path:      "docs",
			SyncToGit: true,
		},
	}
}

// IsLazy determines if a repository should be lazy-loaded
func (r *Repository) IsLazy(name string, defaults Defaults) bool {
	// Explicit configuration wins
	if r.Lazy != nil {
		return *r.Lazy
	}
	
	// Extract repo name from URL if needed
	repoName := name
	if repoName == "" {
		repoName = extractRepoName(r.URL)
	}
	
	// Check eager pattern (meta-repos)
	if defaults.EagerPattern != "" {
		if matched, _ := regexp.MatchString(defaults.EagerPattern, repoName); matched {
			return false  // Not lazy, fetch eagerly
		}
	}
	
	// Check lazy pattern
	if defaults.LazyPattern != "" {
		if matched, _ := regexp.MatchString(defaults.LazyPattern, repoName); matched {
			return true
		}
	}
	
	// Use global default (typically true)
	return defaults.Lazy
}

// extractRepoName extracts repository name from URL
func extractRepoName(url string) string {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")
	
	// Get last path component
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	
	return url
}

// ValidateDefaults checks if patterns are valid regex
func (d *Defaults) ValidateDefaults() error {
	if d.EagerPattern != "" {
		if _, err := regexp.Compile(d.EagerPattern); err != nil {
			return fmt.Errorf("invalid eager_pattern regex: %w", err)
		}
	}
	
	if d.LazyPattern != "" {
		if _, err := regexp.Compile(d.LazyPattern); err != nil {
			return fmt.Errorf("invalid lazy_pattern regex: %w", err)
		}
	}
	
	return nil
}

// Load reads a configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Apply defaults
	
	// Set default workspace settings
	if cfg.Workspace.RootPath == "" {
		cfg.Workspace.RootPath = "repos"
	}
	
	// Set default repository loading settings
	if cfg.Defaults.EagerPattern == "" {
		cfg.Defaults = DefaultDefaults()
	}
	
	// Set documentation defaults
	if cfg.Documentation.Path == "" {
		cfg.Documentation.Path = "docs"
	}
	
	// Store the path
	cfg.Path = filepath.Dir(path)
	
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Save writes a configuration to a YAML file
func (c *Config) Save(path string) error {
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

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Workspace.Name == "" {
		return fmt.Errorf("workspace name is required")
	}

	if len(c.Repositories) == 0 {
		return fmt.Errorf("at least one repository must be defined")
	}

	// Scopes validation removed - tree-based doesn't use scopes
	// if len(c.Scopes) == 0 {
	// 	return fmt.Errorf("at least one scope must be defined")
	// }
	
	// Validate defaults
	if err := c.Defaults.ValidateDefaults(); err != nil {
		return fmt.Errorf("invalid defaults: %w", err)
	}

	// Scope validation removed - tree-based doesn't use scopes
	// for scopeName, scope := range c.Scopes {
	// 	// Check local repos
	// 	for _, repoName := range scope.Repos {
	// 		if _, exists := c.Repositories[repoName]; !exists {
	// 			return fmt.Errorf("scope %s references undefined repository: %s", scopeName, repoName)
	// 		}
	// 	}
	// 	
	// 	// Note: We can't validate workspace scopes until we load the child workspaces
	// 	// This will be done at runtime
	// 	
	// 	// Validate scope type
	// 	if scope.Type != "persistent" && scope.Type != "ephemeral" {
	// 		return fmt.Errorf("scope %s has invalid type: %s (must be 'persistent' or 'ephemeral')", scopeName, scope.Type)
	// 	}
	// }

	return nil
}

// GetRepository returns a repository by name
func (c *Config) GetRepository(name string) (*Repository, error) {
	repo, exists := c.Repositories[name]
	if !exists {
		return nil, fmt.Errorf("repository %s not found", name)
	}
	repoCopy := repo
	return &repoCopy, nil
}

// GetScope - removed in tree-based architecture
// func (c *Config) GetScope(name string) (*Scope, error) {
// 	scope, exists := c.Scopes[name]
// 	if !exists {
// 		return nil, fmt.Errorf("scope %s not found", name)
// 	}
// 	return &scope, nil
// }

// IsMetaRepo checks if a repository is a meta-repo based on patterns
func IsMetaRepo(name string, url string, defaults Defaults) bool {
	// Extract repo name from URL if needed
	repoName := name
	if repoName == "" {
		repoName = extractRepoName(url)
	}
	
	// Check eager pattern (meta-repos)
	if defaults.EagerPattern != "" {
		if matched, _ := regexp.MatchString(defaults.EagerPattern, repoName); matched {
			return true
		}
	}
	
	return false
}

// GetChildWorkspaces returns repositories that are workspaces
func (c *Config) GetChildWorkspaces() map[string]*Repository {
	workspaces := make(map[string]*Repository)
	
	for name, repo := range c.Repositories {
		if repo.IsWorkspace {
			repoCopy := repo
			workspaces[name] = &repoCopy
		}
	}
	
	return workspaces
}

// GetLocalRepositories returns repositories that are not workspaces
func (c *Config) GetLocalRepositories() map[string]*Repository {
	repos := make(map[string]*Repository)
	
	for name, repo := range c.Repositories {
		if !repo.IsWorkspace {
			repoCopy := repo
			repos[name] = &repoCopy
		}
	}
	
	return repos
}