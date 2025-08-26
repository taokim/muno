package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Version       int                    `yaml:"version"`
	Workspace     WorkspaceConfig        `yaml:"workspace"`
	Repositories  map[string]Repository  `yaml:"repositories"`
	Scopes        map[string]Scope       `yaml:"scopes"`
	Documentation DocumentationConfig    `yaml:"documentation,omitempty"`
}

// WorkspaceConfig represents workspace configuration
type WorkspaceConfig struct {
	Name          string `yaml:"name"`
	IsolationMode bool   `yaml:"isolation_mode"`       // Always true in v2
	BasePath      string `yaml:"base_path,omitempty"`  // Base path for scopes (defaults to "workspaces")
}

// Repository represents a repository definition
type Repository struct {
	URL           string   `yaml:"url"`
	DefaultBranch string   `yaml:"default_branch"`
	Groups        []string `yaml:"groups,omitempty"`
}

// Scope represents an isolated working context
type Scope struct {
	Type         string   `yaml:"type"`               // "persistent" or "ephemeral"
	Repos        []string `yaml:"repos"`              // List of repository names
	Description  string   `yaml:"description"`        // Human-readable description
	Model        string   `yaml:"model"`              // Claude model to use
	AutoStart    bool     `yaml:"auto_start"`         // Whether to auto-start this scope
}

// DocumentationConfig represents documentation settings
type DocumentationConfig struct {
	Path       string `yaml:"path"`          // Documentation root path (defaults to "docs")
	SyncToGit  bool   `yaml:"sync_to_git"`   // Auto-commit documentation changes
	RemoteURL  string `yaml:"remote_url,omitempty"` // Optional separate docs repo
}

// DefaultConfig returns the default configuration template
func DefaultConfig(projectName string) *Config {
	return &Config{
		Version: 2,
		Workspace: WorkspaceConfig{
			Name:          projectName,
			IsolationMode: true,
			BasePath:      "workspaces",
		},
		Repositories: map[string]Repository{
			// WMS (Warehouse Management System) components
			"wms-core": {
				URL:           "https://github.com/yourorg/wms-core.git",
				DefaultBranch: "main",
				Groups:        []string{"wms", "backend", "core"},
			},
			"wms-inventory": {
				URL:           "https://github.com/yourorg/wms-inventory.git",
				DefaultBranch: "main",
				Groups:        []string{"wms", "backend", "inventory"},
			},
			"wms-shipping": {
				URL:           "https://github.com/yourorg/wms-shipping.git",
				DefaultBranch: "main",
				Groups:        []string{"wms", "backend", "logistics"},
			},
			"wms-ui": {
				URL:           "https://github.com/yourorg/wms-ui.git",
				DefaultBranch: "main",
				Groups:        []string{"wms", "frontend", "ui"},
			},
			// OMS (Order Management System) components
			"oms-core": {
				URL:           "https://github.com/yourorg/oms-core.git",
				DefaultBranch: "main",
				Groups:        []string{"oms", "backend", "core"},
			},
			"oms-payment": {
				URL:           "https://github.com/yourorg/oms-payment.git",
				DefaultBranch: "main",
				Groups:        []string{"oms", "backend", "payment"},
			},
			"oms-fulfillment": {
				URL:           "https://github.com/yourorg/oms-fulfillment.git",
				DefaultBranch: "main",
				Groups:        []string{"oms", "backend", "fulfillment"},
			},
			"oms-ui": {
				URL:           "https://github.com/yourorg/oms-ui.git",
				DefaultBranch: "main",
				Groups:        []string{"oms", "frontend", "ui"},
			},
			// Search components
			"search-engine": {
				URL:           "https://github.com/yourorg/search-engine.git",
				DefaultBranch: "main",
				Groups:        []string{"search", "backend", "core"},
			},
			"search-indexer": {
				URL:           "https://github.com/yourorg/search-indexer.git",
				DefaultBranch: "main",
				Groups:        []string{"search", "backend", "data"},
			},
			"search-ui": {
				URL:           "https://github.com/yourorg/search-ui.git",
				DefaultBranch: "main",
				Groups:        []string{"search", "frontend", "ui"},
			},
			// Catalog components
			"catalog-service": {
				URL:           "https://github.com/yourorg/catalog-service.git",
				DefaultBranch: "main",
				Groups:        []string{"catalog", "backend", "core"},
			},
			"catalog-admin": {
				URL:           "https://github.com/yourorg/catalog-admin.git",
				DefaultBranch: "main",
				Groups:        []string{"catalog", "frontend", "admin"},
			},
			"catalog-api": {
				URL:           "https://github.com/yourorg/catalog-api.git",
				DefaultBranch: "main",
				Groups:        []string{"catalog", "backend", "api"},
			},
			// Shared components
			"shared-libs": {
				URL:           "https://github.com/yourorg/shared-libs.git",
				DefaultBranch: "main",
				Groups:        []string{"shared", "core"},
			},
			"api-gateway": {
				URL:           "https://github.com/yourorg/api-gateway.git",
				DefaultBranch: "main",
				Groups:        []string{"shared", "backend", "gateway"},
			},
			"web-storefront": {
				URL:           "https://github.com/yourorg/web-storefront.git",
				DefaultBranch: "main",
				Groups:        []string{"storefront", "frontend", "web"},
			},
			"mobile-app": {
				URL:           "https://github.com/yourorg/mobile-app.git",
				DefaultBranch: "main",
				Groups:        []string{"storefront", "mobile", "ui"},
			},
		},
		Scopes: map[string]Scope{
			// WMS scope - Warehouse Management System
			"wms": {
				Type:        "persistent",
				Repos:       []string{"wms-core", "wms-inventory", "wms-shipping", "wms-ui", "shared-libs"},
				Description: "Warehouse Management System - inventory, shipping, and logistics",
				Model:       "claude-3-5-sonnet-20241022",
				AutoStart:   false,
			},
			// OMS scope - Order Management System
			"oms": {
				Type:        "persistent",
				Repos:       []string{"oms-core", "oms-payment", "oms-fulfillment", "oms-ui", "shared-libs"},
				Description: "Order Management System - order processing, payments, and fulfillment",
				Model:       "claude-3-5-sonnet-20241022",
				AutoStart:   false,
			},
			// Search scope - Search and Discovery
			"search": {
				Type:        "persistent",
				Repos:       []string{"search-engine", "search-indexer", "search-ui", "catalog-api", "shared-libs"},
				Description: "Search and Discovery - search engine, indexing, and relevance",
				Model:       "claude-3-5-sonnet-20241022",
				AutoStart:   false,
			},
			// Catalog scope - Product Catalog Management
			"catalog": {
				Type:        "persistent",
				Repos:       []string{"catalog-service", "catalog-admin", "catalog-api", "search-indexer", "shared-libs"},
				Description: "Product Catalog Management - products, categories, and attributes",
				Model:       "claude-3-5-sonnet-20241022",
				AutoStart:   false,
			},
			// Storefront scope - Customer-facing applications
			"storefront": {
				Type:        "persistent",
				Repos:       []string{"web-storefront", "mobile-app", "api-gateway", "catalog-api", "search-ui"},
				Description: "Customer-facing storefront applications and APIs",
				Model:       "claude-3-5-sonnet-20241022",
				AutoStart:   false,
			},
			// Full-stack scope - End-to-end development
			"fullstack": {
				Type:        "persistent",
				Repos:       []string{"web-storefront", "oms-core", "catalog-api", "search-engine", "api-gateway", "shared-libs"},
				Description: "Full-stack e-commerce development - frontend to backend",
				Model:       "claude-3-5-sonnet-20241022",
				AutoStart:   false,
			},
			// Integration scope - Cross-system integration
			"integration": {
				Type:        "persistent",
				Repos:       []string{"api-gateway", "oms-core", "wms-core", "catalog-service", "shared-libs"},
				Description: "System integration - APIs, events, and data flow",
				Model:       "claude-3-5-sonnet-20241022",
				AutoStart:   false,
			},
			// Ephemeral scope templates
			"hotfix": {
				Type:        "ephemeral",
				Repos:       []string{},
				Description: "Emergency hotfix template - select repos at creation",
				Model:       "claude-3-5-sonnet-20241022",
				AutoStart:   false,
			},
			"feature": {
				Type:        "ephemeral",
				Repos:       []string{},
				Description: "Feature development template - cross-service features",
				Model:       "claude-3-5-sonnet-20241022",
				AutoStart:   false,
			},
		},
		Documentation: DocumentationConfig{
			Path:      "docs",
			SyncToGit: true,
		},
	}
}

// Load reads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Set defaults
	if cfg.Version == 0 {
		cfg.Version = 2
	}
	if cfg.Workspace.BasePath == "" {
		cfg.Workspace.BasePath = "workspaces"
	}
	cfg.Workspace.IsolationMode = true // Always true in v2

	if cfg.Documentation.Path == "" {
		cfg.Documentation.Path = "docs"
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Save writes configuration to a YAML file
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

	if len(c.Scopes) == 0 {
		return fmt.Errorf("at least one scope must be defined")
	}

	// Validate scope repositories exist
	for scopeName, scope := range c.Scopes {
		for _, repoName := range scope.Repos {
			if _, exists := c.Repositories[repoName]; !exists {
				return fmt.Errorf("scope %s references undefined repository: %s", scopeName, repoName)
			}
		}

		// Validate scope type
		if scope.Type != "persistent" && scope.Type != "ephemeral" {
			return fmt.Errorf("scope %s has invalid type: %s (must be 'persistent' or 'ephemeral')", scopeName, scope.Type)
		}
	}

	return nil
}

// GetRepository returns a repository by name
func (c *Config) GetRepository(name string) (*Repository, error) {
	repo, exists := c.Repositories[name]
	if !exists {
		return nil, fmt.Errorf("repository %s not found", name)
	}
	return &repo, nil
}

// GetScope returns a scope by name
func (c *Config) GetScope(name string) (*Scope, error) {
	scope, exists := c.Scopes[name]
	if !exists {
		return nil, fmt.Errorf("scope %s not found", name)
	}
	return &scope, nil
}

// GetScopesForRepo returns all scopes that contain a given repository
func (c *Config) GetScopesForRepo(repoName string) []string {
	var scopes []string
	for scopeName, scope := range c.Scopes {
		for _, repo := range scope.Repos {
			if repo == repoName {
				scopes = append(scopes, scopeName)
				break
			}
		}
	}
	return scopes
}