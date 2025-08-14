package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Workspace WorkspaceConfig `yaml:"workspace"`
	Agents    map[string]Agent `yaml:"agents"`
}

// WorkspaceConfig represents workspace configuration
type WorkspaceConfig struct {
	Name     string   `yaml:"name"`
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
	Name   string `yaml:"name"`
	Groups string `yaml:"groups"`
	Agent  string `yaml:"agent,omitempty"`
}

// Agent represents an AI agent configuration
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
				DefaultRevision: "master",
				Projects: []Project{
					{Name: "backend", Groups: "core,services", Agent: "backend-agent"},
					{Name: "frontend", Groups: "core,ui", Agent: "frontend-agent"},
					{Name: "mobile", Groups: "mobile,ui", Agent: "mobile-agent"},
					{Name: "shared-libs", Groups: "shared,core"},
				},
			},
		},
		Agents: map[string]Agent{
			"backend-agent": {
				Model:          "claude-sonnet-4",
				Specialization: "API development, database design, backend services",
				AutoStart:      true,
			},
			"frontend-agent": {
				Model:          "claude-sonnet-4",
				Specialization: "React/Vue development, UI/UX implementation",
				AutoStart:      true,
				Dependencies:   []string{"backend-agent"},
			},
			"mobile-agent": {
				Model:          "claude-sonnet-4",
				Specialization: "Mobile app development, native iOS/Android",
				AutoStart:      false,
				Dependencies:   []string{"backend-agent"},
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