package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/taokim/repo-claude/internal/config"
	"github.com/taokim/repo-claude/internal/tui"
)

// StartInteractiveTUIV2 launches the improved Bubbletea interactive UI for selecting what to start
func (m *Manager) StartInteractiveTUIV2() error {
	// Create the improved TUI model
	model := tui.NewStartModelV2(m.Config, m.State)
	
	// Run the TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("running interactive UI: %w", err)
	}
	
	// Check if user cancelled
	startModel, ok := finalModel.(*tui.StartModelV2)
	if !ok || !startModel.IsLaunching() {
		return nil // User cancelled, not an error
	}
	
	// Get selection mode and items
	selectionMode := startModel.GetSelectionMode()
	
	// Handle based on selection mode
	switch selectionMode {
	case tui.ModeScope:
		// Single scope selected (radio button behavior)
		scopeName := startModel.GetSelectedScope()
		if scopeName == "" {
			return nil
		}
		
		// For single scope, run in current terminal (no new window)
		opts := StartOptions{
			NewWindow: false,
		}
		
		fmt.Printf("🚀 Starting scope '%s' in current terminal...\n", scopeName)
		return m.StartScopeWithOptions(scopeName, opts)
		
	case tui.ModeRepo:
		// One or more repos selected (checkbox behavior)
		repos := startModel.GetSelectedRepos()
		if len(repos) == 0 {
			return nil
		}
		
		// Determine if we need new windows
		opts := StartOptions{
			NewWindow: len(repos) > 1, // Only use new window for multiple repos
		}
		
		if len(repos) == 1 {
			// Single repo - run in current terminal
			fmt.Printf("🚀 Starting repository '%s' in current terminal...\n", repos[0])
			
			// Try to find a scope that contains only this repo
			for name, scope := range m.Config.Scopes {
				resolvedRepos := m.resolveScopeRepos(scope.Repos)
				if len(resolvedRepos) == 1 && resolvedRepos[0] == repos[0] {
					// Found a scope with just this repo
					return m.StartScopeWithOptions(name, opts)
				}
			}
			
			// No single-repo scope found, create a temporary one
			return m.StartRepoAsSingleScope(repos[0], opts)
		} else {
			// Multiple repos - create a combined scope
			fmt.Printf("🪟 Starting %d repositories in new windows...\n", len(repos))
			return m.StartReposAsScope(repos, opts)
		}
		
	default:
		return fmt.Errorf("no selection made")
	}
}

// StartRepoAsSingleScope starts a Claude session with a single repository
func (m *Manager) StartRepoAsSingleScope(repo string, opts StartOptions) error {
	// Find the project configuration
	var project *config.Project
	for _, p := range m.Config.Workspace.Manifest.Projects {
		if p.Name == repo {
			proj := p
			project = &proj
			break
		}
	}
	
	if project == nil {
		return fmt.Errorf("repository %s not found in configuration", repo)
	}
	
	// Build the repository path
	repoPath := filepath.Join(m.WorkspacePath, project.Path)
	if project.Path == "" {
		repoPath = filepath.Join(m.WorkspacePath, project.Name)
	}
	
	// Check if repository exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("repository %s not found at %s, run 'rc sync' first", repo, repoPath)
	}
	
	// Build command for single repo
	systemPrompt := fmt.Sprintf(
		"You are working on the %s repository. "+
		"This is part of the %s workspace. "+
		"The shared memory file is at: %s/shared-memory.md",
		repo, m.Config.Workspace.Name, m.WorkspacePath)
	
	// Set environment variables
	envVars := map[string]string{
		"RC_SCOPE_ID":       fmt.Sprintf("repo-%s", repo),
		"RC_SCOPE_NAME":     repo,
		"RC_SCOPE_REPOS":    repo,
		"RC_WORKSPACE_ROOT": m.WorkspacePath,
		"RC_PROJECT_ROOT":   m.WorkspacePath,
	}
	
	// Determine model to use
	model := "claude-sonnet-4" // default
	
	// Check if any scope contains this repo and use its model
	for _, scope := range m.Config.Scopes {
		repos := m.resolveScopeRepos(scope.Repos)
		for _, r := range repos {
			if r == repo {
				if scope.Model != "" {
					model = scope.Model
				}
				break
			}
		}
	}
	
	// Create and start the command
	var cmd Cmd
	if m.CmdExecutor != nil {
		if opts.NewWindow {
			cmd = createNewTerminalCommand(m.CmdExecutor, repo, repoPath, model, systemPrompt, envVars, true)
		} else {
			// Run in current terminal
			cmd = m.CmdExecutor.Command("claude", "--model", model, "--append-system-prompt", systemPrompt)
			cmd.SetDir(repoPath)
			
			// Set environment variables
			envList := os.Environ()
			for k, v := range envVars {
				envList = append(envList, fmt.Sprintf("%s=%s", k, v))
			}
			cmd.SetEnv(envList)
		}
	} else {
		// Fallback to real implementation
		if m.CmdExecutor == nil {
			m.CmdExecutor = &RealCommandExecutor{}
		}
		cmd = createNewTerminalCommand(m.CmdExecutor, repo, repoPath, model, systemPrompt, envVars, opts.NewWindow)
	}
	
	// Start the command
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start Claude for %s: %w", repo, err)
	}
	
	fmt.Printf("✅ Started Claude session for repository: %s (PID: %d)\n", repo, cmd.Process().Pid)
	
	// Update state
	if m.State == nil {
		m.State = &config.State{
			Scopes: make(map[string]config.ScopeStatus),
		}
	}
	
	m.State.Scopes[fmt.Sprintf("repo-%s", repo)] = config.ScopeStatus{
		Name:         repo,
		Status:       "running",
		PID:          cmd.Process().Pid,
		Repos:        []string{repo},
		LastActivity: time.Now().Format(time.RFC3339),
	}
	
	return m.State.Save(filepath.Join(m.WorkspacePath, ".repo-claude-state.json"))
}