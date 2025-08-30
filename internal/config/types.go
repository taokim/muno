package config

// WorkspaceConfig represents workspace configuration
type WorkspaceConfig struct {
	Name     string `yaml:"name"`
	RootPath string `yaml:"root_path,omitempty"`
}

// DocumentationConfig represents documentation settings
type DocumentationConfig struct {
	Path      string `yaml:"path"`
	SyncToGit bool   `yaml:"sync_to_git"`
}

