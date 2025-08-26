package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/taokim/repo-claude/internal/config"
	"github.com/taokim/repo-claude/internal/docs"
	"github.com/taokim/repo-claude/internal/scope"
)

// Manager is the scope-based manager implementation
type Manager struct {
	ProjectPath   string         // Project root path
	Config        *config.Config // Configuration
	ScopeManager  *scope.Manager // Scope manager
	DocsManager   *docs.Manager  // Documentation manager
	CmdExecutor   CommandExecutor
}

// New creates a new scope-based manager
func New(projectPath string) (*Manager, error) {
	absPath, _ := filepath.Abs(projectPath)
	return &Manager{
		ProjectPath: absPath,
		CmdExecutor: &RealCommandExecutor{},
	}, nil
}

// LoadFromCurrentDir loads an existing workspace
func LoadFromCurrentDir() (*Manager, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(cwd, "repo-claude.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no repo-claude.yaml found in current directory")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// Create scope manager
	scopeMgr, err := scope.NewManager(cfg, cwd)
	if err != nil {
		return nil, fmt.Errorf("creating scope manager: %w", err)
	}

	// Create docs manager
	docsMgr, err := docs.NewManager(cwd)
	if err != nil {
		return nil, fmt.Errorf("creating docs manager: %w", err)
	}

	return &Manager{
		ProjectPath:  cwd,
		Config:       cfg,
		ScopeManager: scopeMgr,
		DocsManager:  docsMgr,
		CmdExecutor:  &RealCommandExecutor{},
	}, nil
}

// InitWorkspace initializes a new workspace
func (m *Manager) InitWorkspace(projectName string, interactive bool) error {
	fmt.Printf("🚀 Initializing Repo-Claude workspace (v2): %s\n", projectName)

	// Create project directory
	if err := os.MkdirAll(m.ProjectPath, 0755); err != nil {
		return fmt.Errorf("creating project directory: %w", err)
	}

	// Create or load configuration
	configPath := filepath.Join(m.ProjectPath, "repo-claude.yaml")
	var cfg *config.Config
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if interactive {
			cfg = m.interactiveConfig(projectName)
		} else {
			cfg = config.DefaultConfig(projectName)
		}
		
		if err := cfg.Save(configPath); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Println("✅ Created repo-claude.yaml")
	} else {
		cfg, err = config.Load(configPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		fmt.Println("✅ Using existing repo-claude.yaml")
	}
	
	m.Config = cfg

	// Create workspaces directory
	workspacesPath := filepath.Join(m.ProjectPath, cfg.Workspace.BasePath)
	if err := os.MkdirAll(workspacesPath, 0755); err != nil {
		return fmt.Errorf("creating workspaces directory: %w", err)
	}
	fmt.Printf("✅ Created workspaces directory: %s\n", cfg.Workspace.BasePath)

	// Create documentation structure
	docsManager, err := docs.NewManager(m.ProjectPath)
	if err != nil {
		return fmt.Errorf("creating docs manager: %w", err)
	}
	m.DocsManager = docsManager
	fmt.Println("✅ Created documentation structure: docs/")

	// Create root CLAUDE.md
	if err := m.createRootCLAUDE(); err != nil {
		return fmt.Errorf("creating root CLAUDE.md: %w", err)
	}
	fmt.Println("✅ Created root CLAUDE.md")

	// Initialize scope manager
	scopeMgr, err := scope.NewManager(cfg, m.ProjectPath)
	if err != nil {
		return fmt.Errorf("creating scope manager: %w", err)
	}
	m.ScopeManager = scopeMgr

	fmt.Println("\n✨ Workspace initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit repo-claude.yaml to configure your repositories and scopes")
	fmt.Println("2. Run 'rc list' to see available scopes")
	fmt.Println("3. Run 'rc start <scope>' to start working")
	
	return nil
}

// ListScopes lists all available scopes with details
func (m *Manager) ListScopes() error {
	// Get both configured scopes and created scopes
	configuredScopes := m.Config.Scopes
	createdScopes, err := m.ScopeManager.List()
	if err != nil {
		return fmt.Errorf("listing created scopes: %w", err)
	}

	// Build combined list
	scopeMap := make(map[string]scope.ScopeDetail)
	num := 1
	
	// Add all configured scopes
	for name, scopeConfig := range configuredScopes {
		detail := scope.ScopeDetail{
			Number:      num,
			Name:        name,
			Type:        scope.Type(scopeConfig.Type),
			State:       scope.StateInactive,
			Repos:       scopeConfig.Repos,
			RepoCount:   len(scopeConfig.Repos),
			Description: scopeConfig.Description,
		}
		
		// Check if scope is created
		for _, created := range createdScopes {
			if created.Name == name {
				detail.State = created.State
				detail.CreatedAt = created.CreatedAt
				detail.LastAccessed = created.LastAccessed
				detail.Path = created.Path
				break
			}
		}
		
		scopeMap[name] = detail
		num++
	}

	if len(scopeMap) == 0 {
		fmt.Println("No scopes found. Check your repo-claude.yaml configuration.")
		return nil
	}

	fmt.Println("\n📋 Available Scopes:")
	fmt.Println(strings.Repeat("-", 80))
	
	// Convert map to slice and sort by number
	var details []scope.ScopeDetail
	for _, detail := range scopeMap {
		details = append(details, detail)
	}
	// Sort by number (already set correctly)
	
	for _, detail := range details {
		status := "○"
		if detail.State == scope.StateActive {
			status = "●"
		}
		
		fmt.Printf("\n%s [%d] %s (%s)\n", status, detail.Number, detail.Name, detail.Type)
		if detail.Description != "" {
			fmt.Printf("    📝 %s\n", detail.Description)
		}
		fmt.Printf("    📦 Repos (%d): %s\n", detail.RepoCount, strings.Join(detail.Repos, ", "))
		
		if detail.State != scope.StateActive && detail.CreatedAt.IsZero() {
			fmt.Println("    ⚠️  Not initialized (will clone on first start)")
		} else {
			fmt.Printf("    📅 Created: %s, Last used: %s\n", 
				detail.CreatedAt.Format("2006-01-02"),
				detail.LastAccessed.Format("2006-01-02 15:04"))
		}
	}
	
	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Println("Use 'rc start <number|name>' to start a scope")
	
	return nil
}

// StartScope starts a scope session
func (m *Manager) StartScope(identifier string, newWindow bool) error {
	// Try to get scope by number or name
	var targetScope *scope.Scope
	var scopeName string
	
	// Check if identifier is a number
	if num, err := strconv.Atoi(identifier); err == nil {
		details, err := m.ScopeManager.ListWithRepos()
		if err != nil {
			return fmt.Errorf("listing scopes: %w", err)
		}
		
		if num < 1 || num > len(details) {
			return fmt.Errorf("scope number %d out of range (1-%d)", num, len(details))
		}
		
		scopeName = details[num-1].Name
	} else {
		scopeName = identifier
	}
	
	// Check if scope exists or needs to be created
	targetScope, err := m.ScopeManager.Get(scopeName)
	if err != nil {
		// Scope doesn't exist, check if it's defined in config
		if _, exists := m.Config.Scopes[scopeName]; exists {
			fmt.Printf("Creating new scope '%s' from configuration...\n", scopeName)
			if err := m.ScopeManager.CreateFromConfig(scopeName); err != nil {
				return fmt.Errorf("creating scope: %w", err)
			}
			targetScope, err = m.ScopeManager.Get(scopeName)
			if err != nil {
				return fmt.Errorf("getting created scope: %w", err)
			}
		} else {
			return fmt.Errorf("scope '%s' not found in configuration", scopeName)
		}
	}
	
	// Start the scope
	startOpts := scope.StartOptions{
		NewWindow: newWindow,
		Pull:      false, // Don't auto-pull by default
	}
	
	if err := targetScope.Start(startOpts); err != nil {
		return fmt.Errorf("starting scope: %w", err)
	}
	
	// Get scope config for model
	scopeConfig := m.Config.Scopes[scopeName]
	model := scopeConfig.Model
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}
	
	// Build system prompt
	systemPrompt := fmt.Sprintf(`You are working in repo-claude scope: %s
Description: %s
Working directory: %s
Check CLAUDE.md files for detailed instructions and documentation guidelines.`,
		scopeName, scopeConfig.Description, targetScope.GetPath())
	
	// Start Claude session
	scopePath := targetScope.GetPath()
	cmd := m.CmdExecutor.Command("claude", "--model", model, "--append-system-prompt", systemPrompt)
	
	if realCmd, ok := cmd.(*RealCmd); ok {
		realCmd.cmd.Dir = scopePath
		
		// Set environment variables
		realCmd.cmd.Env = append(os.Environ(),
			fmt.Sprintf("RC_SCOPE_NAME=%s", scopeName),
			fmt.Sprintf("RC_SCOPE_PATH=%s", scopePath),
			fmt.Sprintf("RC_PROJECT_ROOT=%s", m.ProjectPath),
			fmt.Sprintf("RC_WORKSPACE_ROOT=%s", m.ScopeManager.GetWorkspacePath()),
		)
		
		if newWindow {
			// Open in new terminal window
			fmt.Printf("Opening scope '%s' in new window...\n", scopeName)
			// Platform-specific new window logic would go here
			return realCmd.Start()
		} else {
			// Run in current terminal
			fmt.Printf("Starting scope '%s' in current terminal...\n", scopeName)
			realCmd.AttachToTerminal()
			return realCmd.Run()
		}
	}
	
	return fmt.Errorf("failed to create command")
}

// PullScope pulls repositories in a scope
func (m *Manager) PullScope(scopeName string, cloneMissing bool) error {
	s, err := m.ScopeManager.GetByNumberOrName(scopeName)
	if err != nil {
		return fmt.Errorf("getting scope: %w", err)
	}
	
	opts := scope.PullOptions{
		CloneMissing: cloneMissing,
		Parallel:     true,
	}
	
	return s.Pull(opts)
}

// CommitScope commits changes in a scope
func (m *Manager) CommitScope(scopeName, message string) error {
	s, err := m.ScopeManager.GetByNumberOrName(scopeName)
	if err != nil {
		return fmt.Errorf("getting scope: %w", err)
	}
	
	return s.Commit(message)
}

// PushScope pushes changes from a scope
func (m *Manager) PushScope(scopeName string) error {
	s, err := m.ScopeManager.GetByNumberOrName(scopeName)
	if err != nil {
		return fmt.Errorf("getting scope: %w", err)
	}
	
	return s.Push()
}

// BranchScope switches branches in a scope
func (m *Manager) BranchScope(scopeName, branchName string) error {
	s, err := m.ScopeManager.GetByNumberOrName(scopeName)
	if err != nil {
		return fmt.Errorf("getting scope: %w", err)
	}
	
	return s.SwitchBranch(branchName)
}

// StatusScope shows status of a scope
func (m *Manager) StatusScope(scopeName string) error {
	s, err := m.ScopeManager.GetByNumberOrName(scopeName)
	if err != nil {
		return fmt.Errorf("getting scope: %w", err)
	}
	
	report, err := s.Status()
	if err != nil {
		return fmt.Errorf("getting status: %w", err)
	}
	
	fmt.Printf("\n📊 Scope Status: %s\n", report.Name)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Type: %s\n", report.Type)
	fmt.Printf("State: %s\n", report.State)
	fmt.Printf("Path: %s\n", report.Path)
	fmt.Printf("Created: %s\n", report.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Printf("Last accessed: %s\n", report.LastAccessed.Format("2006-01-02 15:04"))
	fmt.Printf("Disk usage: %.2f MB\n", float64(report.DiskUsage)/(1024*1024))
	
	if len(report.Repos) > 0 {
		fmt.Println("\n📦 Repositories:")
		for _, repo := range report.Repos {
			status := "✅"
			if repo.IsDirty {
				status = "⚠️"
			}
			
			fmt.Printf("\n  %s %s (branch: %s)\n", status, repo.Name, repo.Branch)
			if repo.IsDirty {
				fmt.Printf("     Modified: %d, Untracked: %d\n", 
					repo.ModifiedFiles, repo.UntrackedFiles)
			}
			if repo.AheadBy > 0 || repo.BehindBy > 0 {
				fmt.Printf("     Ahead: %d, Behind: %d\n", repo.AheadBy, repo.BehindBy)
			}
		}
	}
	
	return nil
}

// createRootCLAUDE creates the root-level CLAUDE.md
func (m *Manager) createRootCLAUDE() error {
	claudePath := filepath.Join(m.ProjectPath, "CLAUDE.md")
	
	content := fmt.Sprintf(`# CLAUDE.md - Root Level

This file provides guidance to Claude Code when working with this repo-claude project.

## Project Structure

This is a **repo-claude v2** project with scope-based isolation.

### Three-Level Architecture:
1. **Root Level** (this directory): Project configuration and management
2. **Scope Level** (%s/): Isolated workspaces for different contexts
3. **Repo Level**: Individual git repositories within each scope

## Configuration

- Main config: repo-claude.yaml
- Workspaces directory: %s/
- Documentation: docs/

## Available Scopes

Check repo-claude.yaml for configured scopes. Each scope provides:
- Isolated workspace
- Independent repository clones
- Separate branch management
- Own documentation context

## Documentation Structure

### IMPORTANT: Documentation Location Rules

1. **Global documentation**: Store in docs/global/
   - Architecture decisions
   - Coding standards
   - Project-wide guidelines

2. **Scope documentation**: Store in docs/scopes/<scope-name>/
   - Scope-specific designs
   - Implementation notes
   - Cross-repository documentation for that scope

3. **Repository documentation**: Store in each repository's docs/
   - Repository-specific documentation
   - API documentation
   - Module guides

## Commands

All commands are scope-aware:

### Scope Management
- rc list                       # List all scopes
- rc start <scope>             # Start a scope (by name or number)
- rc status <scope>            # Check scope status

### Git Operations (scope-specific)
- rc pull <scope> --clone-missing
- rc commit <scope> -m "message"
- rc push <scope>
- rc branch <scope> <branch-name>

### Documentation
- rc docs create <scope> <file>
- rc docs list [<scope>]
- rc docs sync

## Working with Scopes

When you start a scope:
1. The working directory will be the scope directory
2. Repositories are cloned on first use
3. Each scope maintains independent state
4. Use shared-memory.md for inter-scope communication

## Best Practices

1. **Isolation**: Each scope is independent - changes don't affect other scopes
2. **Documentation**: Keep cross-repo docs in docs/scopes/<scope>/
3. **Cleanup**: Delete unused ephemeral scopes to save disk space
4. **Branching**: Each scope can work on different branches

Created: %s
`, m.Config.Workspace.BasePath, m.Config.Workspace.BasePath, 
		time.Now().Format("2006-01-02 15:04:05"))
	
	return os.WriteFile(claudePath, []byte(content), 0644)
}

// interactiveConfig creates configuration interactively
func (m *Manager) interactiveConfig(projectName string) *config.Config {
	// For now, return default config
	// TODO: Implement interactive configuration
	return config.DefaultConfig(projectName)
}