package constants

// Default configuration values
const (
	// DefaultReposDir is the default directory name for repositories
	DefaultReposDir = "repos"

	// DefaultWorkspaceName is the default name when workspace name is not specified
	DefaultWorkspaceName = "muno-workspace"

	// DefaultConfigFileName is the primary configuration file name
	DefaultConfigFileName = "muno.yaml"

	// AlternativeConfigFileName is an alternative configuration file name (hidden)
	AlternativeConfigFileName = ".muno.yaml"

	// DefaultStateFileName is the state file name for tracking tree state
	DefaultStateFileName = ".muno-state.json"
)

// Configuration file variants for discovery
var (
	// ConfigFileNames are the possible configuration file names to search for
	ConfigFileNames = []string{
		"muno.yaml",
		".muno.yaml",
		"muno.yml",
		".muno.yml",
	}

	// EagerLoadPatterns are repository name patterns that trigger eager loading
	// These repositories are expected to contain child nodes and should be
	// cloned immediately to discover their structure
	EagerLoadPatterns = []string{
		"-monorepo",
		"-munorepo",
		"-muno",
		"-metarepo",
		"-platform",
		"-workspace",
		"-root-repo", // New pattern for root repositories
	}

	// DefaultIgnorePatterns are patterns to ignore when scanning for repositories
	DefaultIgnorePatterns = []string{
		".git",
		"node_modules",
		"vendor",
		"dist",
		"build",
		"target",
		".idea",
		".vscode",
	}
)

// Git-related constants
const (
	// GitDirName is the directory name for git metadata
	GitDirName = ".git"

	// DefaultRemoteName is the default git remote name
	DefaultRemoteName = "origin"

	// DefaultBranch is the default branch name
	DefaultBranch = "main"
)

// Command output formats
const (
	// TreeIndent is the indentation for tree display
	TreeIndent = "  "

	// TreeBranch is the branch character for tree display
	TreeBranch = "├── "

	// TreeLastBranch is the last branch character for tree display
	TreeLastBranch = "└── "

	// TreeVertical is the vertical line for tree display
	TreeVertical = "│   "

	// TreeSpace is the space for tree display
	TreeSpace = "    "
)
