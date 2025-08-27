package config

// WorkspaceConfigV3 represents workspace configuration for v3 tree-based system
type WorkspaceConfigV3 struct {
	Name     string `yaml:"name"`
	RootRepo string `yaml:"root_repo,omitempty"` // Optional: root can be a git repo
}

// DocumentationConfigV3 represents documentation settings (simplified for v3)
// Documentation is now just README.md files in each node
type DocumentationConfigV3 struct {
	// Deprecated - each node has its own README.md
}

// Config represents the main configuration (v3 only)
type Config struct {
	Version   int               `yaml:"version"`
	Workspace WorkspaceConfigV3 `yaml:"workspace"`
	
	// Runtime fields
	Path string `yaml:"-"`
}