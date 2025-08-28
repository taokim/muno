package config

// WorkspaceConfig represents workspace configuration (shared between v2 and v3 for now)
type WorkspaceConfig struct {
	Name          string `yaml:"name"`
	IsolationMode bool   `yaml:"isolation_mode"`
	BasePath      string `yaml:"base_path"`
}

// DocumentationConfig represents documentation settings
type DocumentationConfig struct {
	Path      string `yaml:"path"`
	SyncToGit bool   `yaml:"sync_to_git"`
}

// Scope represents a scope configuration (v2 - to be removed)
// TODO: Remove when all code is migrated to ScopeV3
type Scope struct {
	Type            string              `yaml:"type"`
	Repos           []string            `yaml:"repos,omitempty"`
	Description     string              `yaml:"description"`
	Model           string              `yaml:"model,omitempty"`
	AutoStart       bool                `yaml:"auto_start"`
	WorkspaceScopes map[string][]string `yaml:"workspace_scopes,omitempty"`
}