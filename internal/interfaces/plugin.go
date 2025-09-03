package interfaces

import (
	"context"
)

// Plugin is the interface that must be implemented by all MUNO plugins
type Plugin interface {
	// Metadata returns plugin information
	Metadata() PluginMetadata
	
	// Commands returns the list of commands this plugin provides
	Commands() []CommandDefinition
	
	// Execute runs a specific command
	Execute(ctx context.Context, cmd string, args []string, env PluginEnvironment) (Result, error)
	
	// Initialize is called when plugin is loaded
	Initialize(config map[string]interface{}) error
	
	// Cleanup is called before plugin is unloaded
	Cleanup() error
	
	// HealthCheck verifies plugin is operational
	HealthCheck(ctx context.Context) error
}

// PluginMetadata describes a plugin
type PluginMetadata struct {
	Name        string   `json:"name" yaml:"name"`
	Version     string   `json:"version" yaml:"version"`
	Author      string   `json:"author" yaml:"author"`
	Description string   `json:"description" yaml:"description"`
	MinMunoVer  string   `json:"min_muno_version" yaml:"min_muno_version"`
	MaxMunoVer  string   `json:"max_muno_version" yaml:"max_muno_version"`
	License     string   `json:"license" yaml:"license"`
	Homepage    string   `json:"homepage" yaml:"homepage"`
	Tags        []string `json:"tags" yaml:"tags"`
	Icon        string   `json:"icon" yaml:"icon"`
}

// CommandDefinition describes a plugin command
type CommandDefinition struct {
	Name        string           `json:"name" yaml:"name"`
	Aliases     []string         `json:"aliases" yaml:"aliases"`
	Description string           `json:"description" yaml:"description"`
	Usage       string           `json:"usage" yaml:"usage"`
	Flags       []FlagDefinition `json:"flags" yaml:"flags"`
	Examples    []string         `json:"examples" yaml:"examples"`
	Category    string           `json:"category" yaml:"category"`
	Hidden      bool             `json:"hidden" yaml:"hidden"`
}

// FlagDefinition describes a command flag
type FlagDefinition struct {
	Name        string   `json:"name" yaml:"name"`
	Short       string   `json:"short" yaml:"short"`
	Description string   `json:"description" yaml:"description"`
	Type        string   `json:"type" yaml:"type"` // "string", "bool", "int", "float", "array"
	Default     string   `json:"default" yaml:"default"`
	Required    bool     `json:"required" yaml:"required"`
	Choices     []string `json:"choices" yaml:"choices"`
}

// PluginEnvironment provides context to plugin execution
type PluginEnvironment struct {
	WorkspacePath string                 `json:"workspace_path" yaml:"workspace_path"`
	CurrentNode   string                 `json:"current_node" yaml:"current_node"`
	TreeState     map[string]interface{} `json:"tree_state" yaml:"tree_state"`
	Config        map[string]interface{} `json:"config" yaml:"config"`
	Variables     map[string]string      `json:"variables" yaml:"variables"`
	Platform      string                 `json:"platform" yaml:"platform"`
	Version       string                 `json:"version" yaml:"version"`
}

// Result represents plugin execution result
type Result struct {
	Success bool        `json:"success" yaml:"success"`
	Message string      `json:"message" yaml:"message"`
	Data    interface{} `json:"data" yaml:"data"`
	Error   string      `json:"error,omitempty" yaml:"error,omitempty"`
	Actions []Action    `json:"actions,omitempty" yaml:"actions,omitempty"`
}

// Action represents a follow-up action from plugin execution
type Action struct {
	Type        string                 `json:"type" yaml:"type"` // "command", "navigate", "open", "prompt"
	Command     string                 `json:"command,omitempty" yaml:"command,omitempty"`
	Arguments   []string               `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	Path        string                 `json:"path,omitempty" yaml:"path,omitempty"`
	URL         string                 `json:"url,omitempty" yaml:"url,omitempty"`
	Message     string                 `json:"message,omitempty" yaml:"message,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty" yaml:"options,omitempty"`
}

// PluginManager manages plugin lifecycle
type PluginManager interface {
	// Discovery and loading
	DiscoverPlugins(ctx context.Context) ([]PluginMetadata, error)
	LoadPlugin(ctx context.Context, name string) error
	UnloadPlugin(ctx context.Context, name string) error
	ReloadPlugin(ctx context.Context, name string) error
	
	// Plugin queries
	GetPlugin(name string) (Plugin, error)
	ListPlugins() []PluginMetadata
	IsLoaded(name string) bool
	
	// Command routing
	GetCommand(name string) (*CommandDefinition, Plugin, error)
	ListCommands() []CommandDefinition
	ExecuteCommand(ctx context.Context, name string, args []string) (Result, error)
	
	// Plugin installation
	InstallPlugin(ctx context.Context, source string) error
	UpdatePlugin(ctx context.Context, name string) error
	RemovePlugin(ctx context.Context, name string) error
	
	// Configuration
	GetPluginConfig(name string) (map[string]interface{}, error)
	SetPluginConfig(name string, config map[string]interface{}) error
	
	// Health monitoring
	HealthCheck(ctx context.Context) map[string]error
}

// PluginRegistry provides plugin discovery and distribution
type PluginRegistry interface {
	// Search and discovery
	Search(ctx context.Context, query string) ([]PluginMetadata, error)
	GetPluginInfo(ctx context.Context, name string) (*PluginMetadata, error)
	GetPluginVersions(ctx context.Context, name string) ([]string, error)
	
	// Download and installation
	Download(ctx context.Context, name, version string) ([]byte, error)
	GetInstallScript(ctx context.Context, name, version string) (string, error)
	
	// Publishing (for plugin developers)
	Publish(ctx context.Context, metadata PluginMetadata, data []byte) error
	Unpublish(ctx context.Context, name, version string) error
	
	// Registry management
	AddRegistry(url string) error
	RemoveRegistry(url string) error
	ListRegistries() []string
	RefreshCache(ctx context.Context) error
}

// PluginLoader handles the actual loading mechanism
type PluginLoader interface {
	Load(path string) (Plugin, error)
	Unload(plugin Plugin) error
	Type() string // "grpc", "http", "native", "wasm"
}