package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Workspace WorkspaceConfig  `yaml:"workspace"`
	Scopes    map[string]Scope `yaml:"scopes"`
	
	// Deprecated: Agents field for backwards compatibility (will be removed)
	Agents    map[string]Agent `yaml:"agents,omitempty"`
}

// WorkspaceConfig represents workspace configuration
type WorkspaceConfig struct {
	Name     string   `yaml:"name"`
	Path     string   `yaml:"path,omitempty"`     // Optional workspace path (defaults to "workspace")
	Manifest Manifest `yaml:"manifest"`
}

// Manifest represents repo manifest configuration
type Manifest struct {
	RemoteName     string    `yaml:"remote_name"`
	RemoteFetch    string    `yaml:"remote_fetch"`
	DefaultRevision string   `yaml:"default_revision"`
	Projects       []Project `yaml:"projects"`
}

// Project represents a repository project
type Project struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path,omitempty"`     // Custom path (defaults to name)
	Groups   string `yaml:"groups"`
	Agent    string `yaml:"agent,omitempty"`
	Revision string `yaml:"revision,omitempty"` // Custom branch/revision
}

// Scope represents a working context with associated repositories
type Scope struct {
	Repos        []string `yaml:"repos"`              // List of repository names or patterns
	Description  string   `yaml:"description"`        // Human-readable description
	Model        string   `yaml:"model"`              // Claude model to use
	AutoStart    bool     `yaml:"auto_start"`         // Whether to auto-start this scope
	Dependencies []string `yaml:"dependencies,omitempty"` // Other scopes that must be running
}

// Agent represents an AI agent configuration (deprecated)
type Agent struct {
	Model           string   `yaml:"model"`
	Specialization  string   `yaml:"specialization"`
	AutoStart       bool     `yaml:"auto_start"`
	Dependencies    []string `yaml:"dependencies"`
}

// DefaultConfig returns the default configuration template
func DefaultConfig(projectName string) *Config {
	return &Config{
		Workspace: WorkspaceConfig{
			Name: projectName,
			Manifest: Manifest{
				RemoteName:      "origin",
				RemoteFetch:     "https://github.com/yourorg/",
				DefaultRevision: "main",
				Projects: []Project{
					{Name: "auth-service", Groups: "backend,services"},
					{Name: "order-service", Groups: "backend,services"},
					{Name: "payment-service", Groups: "backend,services"},
					{Name: "web-app", Groups: "frontend,ui"},
					{Name: "mobile-app", Groups: "mobile,ui"},
					{Name: "shared-libs", Groups: "shared,core"},
				},
			},
		},
		Scopes: map[string]Scope{
			"backend": {
				Repos:       []string{"auth-service", "order-service", "payment-service"},
				Description: "Backend services development",
				Model:       "claude-sonnet-4", 
				AutoStart:   true,
			},
			"frontend": {
				Repos:       []string{"web-app", "mobile-app"},
				Description: "Frontend and mobile development",
				Model:       "claude-sonnet-4",
				AutoStart:   true,
			},
			"fullstack": {
				Repos:       []string{"backend/*", "frontend", "shared-libs"},
				Description: "Full-stack development",
				Model:       "claude-sonnet-4",
				AutoStart:   false,
			},
			"order-flow": {
				Repos:       []string{"order-service", "payment-service", "shipping-service"},
				Description: "Order processing pipeline",
				Model:       "claude-sonnet-4",
				AutoStart:   false,
			},
			"infra": {
				Repos:       []string{"shared-libs"},
				Description: "Infrastructure and shared libraries",
				Model:       "claude-sonnet-4",
				AutoStart:   false,
			},
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

// GetProjectForAgent returns the project associated with an agent
func (c *Config) GetProjectForAgent(agentName string) (*Project, error) {
	for i, project := range c.Workspace.Manifest.Projects {
		if project.Agent == agentName {
			return &c.Workspace.Manifest.Projects[i], nil
		}
	}
	return nil, fmt.Errorf("no project found for agent %s", agentName)
}