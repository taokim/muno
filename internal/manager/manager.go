package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/taokim/repo-claude/internal/config"
	"github.com/taokim/repo-claude/internal/git"
)

// Manager handles the repo-claude workspace
type Manager struct {
	ProjectPath   string  // Path to the project root (where repo-claude.yaml is)
	WorkspacePath string  // Path to the workspace subdirectory (where repos are cloned)
	Config        *config.Config
	State         *config.State
	GitManager    *git.Manager
	scopes        map[string]*Scope
	numberToScope map[int]string  // Maps numbers from ps output to scope names
	mu            sync.Mutex
	
	// Legacy support
	agents        map[string]*Agent   // Deprecated: for backwards compatibility
	numberToAgent map[int]string      // Deprecated: for backwards compatibility
	
	// Interfaces for external dependencies (default to real implementations)
	CmdExecutor     CommandExecutor
	FileSystem      FileSystem
	ProcessManager  ProcessManager
}

// Scope represents a running Claude Code instance with scope context
type Scope struct {
	Name    string
	Process *os.Process
	Status  string
	Repos   []string  // List of repositories in this scope
}

// Agent represents a running Claude Code instance (deprecated)
type Agent struct {
	Name    string
	Process *os.Process
	Status  string
}

// New creates a new manager for initialization
func New(projectPath string) *Manager {
	absPath, _ := filepath.Abs(projectPath)
	return &Manager{
		ProjectPath:   absPath,
		WorkspacePath: filepath.Join(absPath, "workspace"), // Default workspace path
		scopes:        make(map[string]*Scope),
		numberToScope: make(map[int]string),
		agents:        make(map[string]*Agent),  // Legacy support
		numberToAgent: make(map[int]string),     // Legacy support
		CmdExecutor:   RealCommandExecutor{},
		FileSystem:    RealFileSystem{},
		ProcessManager: RealProcessManager{},
	}
}

// LoadFromCurrentDir loads an existing workspace from current directory
func LoadFromCurrentDir() (*Manager, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(cwd, "repo-claude.yaml")
	fs := RealFileSystem{}
	if _, err := fs.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no repo-claude.yaml found in current directory")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// Determine workspace path
	workspacePath := filepath.Join(cwd, "workspace") // Default
	if cfg.Workspace.Path != "" {
		if filepath.IsAbs(cfg.Workspace.Path) {
			workspacePath = cfg.Workspace.Path
		} else {
			workspacePath = filepath.Join(cwd, cfg.Workspace.Path)
		}
	}

	// Convert config projects to git repositories
	repos := configToRepos(cfg)
	gitMgr := git.NewManager(workspacePath, repos)

	statePath := filepath.Join(cwd, ".repo-claude-state.json")
	state, _ := config.LoadState(statePath) // Ignore error if state doesn't exist

	return &Manager{
		ProjectPath:   cwd,
		WorkspacePath: workspacePath,
		Config:        cfg,
		State:         state,
		GitManager:    gitMgr,
		scopes:        make(map[string]*Scope),
		numberToScope: make(map[int]string),
		agents:        make(map[string]*Agent),  // Legacy support
		numberToAgent: make(map[int]string),     // Legacy support
		CmdExecutor:   RealCommandExecutor{},
		FileSystem:    RealFileSystem{},
		ProcessManager: RealProcessManager{},
	}, nil
}

// InitWorkspace initializes a new workspace
func (m *Manager) InitWorkspace(projectName string, interactive bool) error {
	if projectName == "." {
		fmt.Println("🚀 Initializing Repo-Claude in current directory")
	} else {
		fmt.Printf("🚀 Initializing Repo-Claude workspace: %s\n", projectName)
	}

	// Create project directory if needed
	if m.FileSystem == nil {
		m.FileSystem = RealFileSystem{}
	}
	if err := m.FileSystem.MkdirAll(m.ProjectPath, 0755); err != nil {
		return fmt.Errorf("creating project directory: %w", err)
	}

	// Check if repo-claude.yaml already exists in project root
	configPath := filepath.Join(m.ProjectPath, "repo-claude.yaml")
	if _, err := m.FileSystem.Stat(configPath); err == nil {
		// Configuration exists, load it
		fmt.Println("📄 Found existing repo-claude.yaml, loading configuration...")
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("loading existing config: %w", err)
		}
		m.Config = cfg
	} else {
		// No existing config, create new one
		fmt.Println("📝 Creating new configuration...")
		m.Config = config.DefaultConfig(projectName)
		
		if interactive {
			if err := m.interactiveConfig(); err != nil {
				return fmt.Errorf("interactive configuration: %w", err)
			}
		}

		// Save configuration
		if err := m.Config.Save(configPath); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
	}

	// Update workspace path based on config
	if m.Config.Workspace.Path != "" {
		if filepath.IsAbs(m.Config.Workspace.Path) {
			m.WorkspacePath = m.Config.Workspace.Path
		} else {
			m.WorkspacePath = filepath.Join(m.ProjectPath, m.Config.Workspace.Path)
		}
	}

	// Create workspace directory
	if err := m.FileSystem.MkdirAll(m.WorkspacePath, 0755); err != nil {
		return fmt.Errorf("creating workspace: %w", err)
	}

	// Initialize GitManager
	repos := configToRepos(m.Config)
	m.GitManager = git.NewManager(m.WorkspacePath, repos)

	// Clone repositories
	fmt.Println("📦 Cloning repositories...")
	if err := m.GitManager.Clone(); err != nil {
		fmt.Printf("⚠️  Some repositories failed to clone: %v\n", err)
		fmt.Println("You can run 'rc sync' to retry")
	}

	// Setup coordination files
	if err := m.setupCoordination(); err != nil {
		return fmt.Errorf("setting up coordination: %w", err)
	}

	// Add workspace to .gitignore if we're in a git repository
	if err := m.updateGitignore(); err != nil {
		// Non-fatal error, just warn
		fmt.Printf("⚠️  Warning: %v\n", err)
	}

	fmt.Println("✅ Workspace initialized!")
	fmt.Printf("📍 Project root: %s\n", m.ProjectPath)
	fmt.Printf("📁 Workspace: %s\n", m.WorkspacePath)
	fmt.Println("\nNext steps:")
	if projectName != "." && m.ProjectPath != "." {
		fmt.Printf("  cd %s\n", filepath.Base(m.ProjectPath))
	}
	fmt.Println("  rc start     # Start all agents")
	fmt.Println("  rc status    # Check status")

	return nil
}

// CloneMissing clones any missing repositories
func (m *Manager) CloneMissing() error {
	if m.GitManager == nil {
		return fmt.Errorf("no git manager initialized")
	}
	return m.GitManager.CloneMissing()
}


// updateGitignore adds the workspace directory to .gitignore if needed
func (m *Manager) updateGitignore() error {
	// Check if we're in a git repository
	if m.FileSystem == nil {
		m.FileSystem = RealFileSystem{}
	}
	gitDir := filepath.Join(m.ProjectPath, ".git")
	if _, err := m.FileSystem.Stat(gitDir); os.IsNotExist(err) {
		// Not in a git repo, nothing to do
		return nil
	}

	gitignorePath := filepath.Join(m.ProjectPath, ".gitignore")
	
	// Calculate workspace path relative to project root
	workspaceRel, err := filepath.Rel(m.ProjectPath, m.WorkspacePath)
	if err != nil {
		return fmt.Errorf("calculating relative workspace path: %w", err)
	}
	
	// Ensure it ends with / to indicate directory
	if !strings.HasSuffix(workspaceRel, "/") {
		workspaceRel += "/"
	}
	
	// Check if .gitignore exists
	content := []byte{}
	if data, err := m.FileSystem.ReadFile(gitignorePath); err == nil {
		content = data
	}
	
	// Check what needs to be added
	lines := strings.Split(string(content), "\n")
	workspaceIgnored := false
	stateIgnored := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == workspaceRel || trimmed == strings.TrimSuffix(workspaceRel, "/") {
			workspaceIgnored = true
		}
		if trimmed == ".repo-claude-state.json" {
			stateIgnored = true
		}
	}
	
	// If both are already ignored, nothing to do
	if workspaceIgnored && stateIgnored {
		return nil
	}
	
	// Add to .gitignore
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		content = append(content, '\n')
	}
	
	// Build addition based on what's needed
	var addition strings.Builder
	addition.WriteString("\n# Repo-Claude workspace\n")
	
	if !workspaceIgnored {
		addition.WriteString(workspaceRel)
		addition.WriteString("\n")
	}
	
	if !stateIgnored {
		addition.WriteString(".repo-claude-state.json\n")
	}
	
	content = append(content, []byte(addition.String())...)
	
	if err := m.FileSystem.WriteFile(gitignorePath, content, 0644); err != nil {
		return fmt.Errorf("updating .gitignore: %w", err)
	}
	
	// Report what was added
	added := []string{}
	if !workspaceIgnored {
		added = append(added, workspaceRel)
	}
	if !stateIgnored {
		added = append(added, ".repo-claude-state.json")
	}
	
	fmt.Printf("📝 Added to .gitignore: %s\n", strings.Join(added, ", "))
	return nil
}

// configToRepos converts config projects to git repositories
func configToRepos(cfg *config.Config) []git.Repository {
	var repos []git.Repository
	
	for _, project := range cfg.Workspace.Manifest.Projects {
		// Build full URL from remote fetch + project name
		url := cfg.Workspace.Manifest.RemoteFetch
		if url == "" {
			// Skip if no remote fetch URL
			continue
		}
		if len(url) > 0 && url[len(url)-1] != '/' {
			url += "/"
		}
		url += project.Name
		
		// Handle .git suffix
		if len(url) > 4 && url[len(url)-4:] != ".git" {
			url += ".git"
		}

		// Use project-specific revision if provided, otherwise use default
		branch := cfg.Workspace.Manifest.DefaultRevision
		if project.Revision != "" {
			branch = project.Revision
		}
		
		// Use project-specific path if provided, otherwise use name
		path := project.Name
		if project.Path != "" {
			path = project.Path
		}
		
		repo := git.Repository{
			Name:   project.Name,
			Path:   path,
			URL:    url,
			Branch: branch,
			Agent:  project.Agent,
		}
		
		// Parse groups
		if project.Groups != "" {
			repo.Groups = strings.Split(project.Groups, ",")
			for i := range repo.Groups {
				repo.Groups[i] = strings.TrimSpace(repo.Groups[i])
			}
		}
		
		repos = append(repos, repo)
	}
	
	return repos
}