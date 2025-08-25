package manager

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// StartOptions defines how scopes should be started
type StartOptions struct {
	NewWindow  bool // Open in new window (default: false, run in current terminal)
}

// createNewTerminalCommand creates a command to run Claude Code
func createNewTerminalCommand(executor CommandExecutor, scopeName, workDir, model, systemPrompt string, envVars map[string]string, newWindow bool) Cmd {
	if newWindow {
		// New window mode - platform specific
		switch runtime.GOOS {
		case "darwin":
			// Build environment variables string
			envStr := ""
			for k, v := range envVars {
				envStr += fmt.Sprintf("export %s='%s'; ", k, v)
			}
			
			script := fmt.Sprintf(`
				tell application "Terminal"
					do script "cd %s && %s claude --model %s --append-system-prompt '%s'"
					activate
				end tell
			`, workDir, envStr, model, systemPrompt)
			return executor.Command("osascript", "-e", script)
		default:
			// For other platforms, use xterm or similar
			envStr := ""
			for k, v := range envVars {
				envStr += fmt.Sprintf("export %s='%s'; ", k, v)
			}
			return executor.Command("xterm", "-e", "bash", "-c", 
				fmt.Sprintf("cd %s && %s claude --model %s --append-system-prompt '%s'; exec bash", 
					workDir, envStr, model, systemPrompt))
		}
	} else {
		// Current terminal mode
		cmd := executor.Command("claude", "--model", model, "--append-system-prompt", systemPrompt)
		cmd.SetDir(workDir)
		
		// Set environment variables
		var env []string
		for k, v := range envVars {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		// Add parent environment
		env = append(env, os.Environ()...)
		cmd.SetEnv(env)
		
		// Attach to terminal for current window mode
		if realCmd, ok := cmd.(*RealCmd); ok {
			realCmd.AttachToTerminal()
		}
		
		return cmd
	}
}

// StartScope starts a specific scope
func (m *Manager) StartScope(scopeName string) error {
	return m.StartScopeWithOptions(scopeName, StartOptions{})
}

// StartScopeWithOptions starts a scope with specific options
func (m *Manager) StartScopeWithOptions(scopeName string, opts StartOptions) error {
	// Check if scope exists in config
	scopeConfig, exists := m.Config.Scopes[scopeName]
	if !exists {
		return fmt.Errorf("scope %s not found in configuration", scopeName)
	}

	// Resolve repositories for this scope
	repos := m.resolveScopeRepos(scopeConfig.Repos)
	if len(repos) == 0 {
		return fmt.Errorf("no repositories found for scope %s", scopeName)
	}

	// Use project root as working directory for consistency
	// This ensures all scopes start from the same location
	workDir := m.ProjectPath

	location := "current terminal"
	if opts.NewWindow {
		location = "new window"
	}
	fmt.Printf("🚀 Starting scope %s with %d repositories (%s)\n", scopeName, len(repos), location)

	// Build system prompt
	systemPrompt := fmt.Sprintf("You are working in repo-claude scope: %s\n"+
		"Description: %s\n"+
		"Repositories in scope: %s\n"+
		"Check shared-memory.md for coordination with other scopes.",
		scopeName, scopeConfig.Description, strings.Join(repos, ", "))

	// Set environment variables for the Claude session
	envVars := map[string]string{
		"RC_SCOPE_ID":        generateScopeID(scopeName),
		"RC_SCOPE_NAME":      scopeName,
		"RC_SCOPE_REPOS":     strings.Join(repos, ","),
		"RC_WORKSPACE_ROOT":  m.WorkspacePath,  // Path where repositories are cloned
		"RC_PROJECT_ROOT":    m.ProjectPath,     // Path where repo-claude.yaml is located
	}

	// Ensure CmdExecutor is initialized
	if m.CmdExecutor == nil {
		m.CmdExecutor = &RealCommandExecutor{}
	}
	
	cmd := createNewTerminalCommand(m.CmdExecutor, scopeName, workDir, scopeConfig.Model, systemPrompt, envVars, opts.NewWindow)

	// For current terminal, run in foreground and wait
	if !opts.NewWindow {
		// Run the command in foreground (blocking)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to run scope: %w", err)
		}
		// Command has completed when running in current terminal
		return nil
	}

	// For new window, start in background
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start scope: %w", err)
	}

	// No longer tracking processes - Claude Code manages its own lifecycle
	fmt.Printf("✅ Scope %s started in new window\n", scopeName)

	return nil
}

// REMOVED: StopScope - users should use Ctrl+C or OS commands to stop processes

// Dummy function to prevent compilation errors
func (m *Manager) StopScope(scopeName string) error {
	return fmt.Errorf("process management has been removed - use Ctrl+C to stop Claude Code")
}

// StartAllScopes starts all auto-start scopes
func (m *Manager) StartAllScopes() error {
	return m.StartAllScopesWithOptions(StartOptions{})
}

// StartAllScopesWithOptions starts all auto-start scopes with options
func (m *Manager) StartAllScopesWithOptions(opts StartOptions) error {
	fmt.Println("🚀 Starting all auto-start scopes...")

	// Get scopes that should auto-start
	var toStart []string
	for name, config := range m.Config.Scopes {
		if config.AutoStart {
			toStart = append(toStart, name)
		}
	}

	if len(toStart) == 0 {
		fmt.Println("No auto-start scopes configured")
		return nil
	}
	
	// Auto-enable new window when starting multiple scopes
	if !opts.NewWindow && len(toStart) > 1 {
		opts.NewWindow = true
		fmt.Printf("🪟 Opening %d scopes in new windows\n", len(toStart))
	}

	// Start scopes respecting dependencies
	started := make(map[string]bool)
	for len(started) < len(toStart) {
		progress := false
		for _, name := range toStart {
			if started[name] {
				continue
			}

			// Check dependencies
			scope := m.Config.Scopes[name]
			ready := true
			for _, dep := range scope.Dependencies {
				if !started[dep] {
					ready = false
					break
				}
			}

			if ready {
				if err := m.StartScopeWithOptions(name, opts); err != nil {
					fmt.Printf("❌ Failed to start %s: %v\n", name, err)
				} else {
					started[name] = true
					progress = true
				}
			}
		}

		if !progress && len(started) < len(toStart) {
			return fmt.Errorf("circular dependency detected")
		}
	}

	return nil
}

// REMOVED: StopAllScopes - process management has been removed
func (m *Manager) StopAllScopes() error {
	return fmt.Errorf("process management has been removed - use Ctrl+C to stop Claude Code")
}

// REMOVED: KillScopeByNumber - process management has been removed
func (m *Manager) KillScopeByNumber(num int) error {
	return fmt.Errorf("process management has been removed - use Ctrl+C to stop Claude Code")
}

// resolveScopeRepos resolves repository patterns to actual repository names
func (m *Manager) resolveScopeRepos(patterns []string) []string {
	repos := make(map[string]bool)
	
	for _, pattern := range patterns {
		// Handle wildcards
		if strings.Contains(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			for _, project := range m.Config.Workspace.Manifest.Projects {
				if strings.HasPrefix(project.Name, prefix) {
					repos[project.Name] = true
				}
			}
		} else {
			// Direct match
			for _, project := range m.Config.Workspace.Manifest.Projects {
				if project.Name == pattern {
					repos[project.Name] = true
					break
				}
			}
		}
	}
	
	// Convert map to slice
	result := make([]string, 0, len(repos))
	for repo := range repos {
		result = append(result, repo)
	}
	
	return result
}

// generateScopeID generates a unique ID for a scope instance
func generateScopeID(scopeName string) string {
	return fmt.Sprintf("%s-%d", scopeName, time.Now().Unix())
}

// StartByRepoName starts a scope that contains the specified repository
func (m *Manager) StartByRepoName(repoName string) error {
	// Find which scope contains this repository
	for scopeName, scopeConfig := range m.Config.Scopes {
		repos := m.resolveScopeRepos(scopeConfig.Repos)
		for _, repo := range repos {
			if repo == repoName {
				return m.StartScope(scopeName)
			}
		}
	}
	
	return fmt.Errorf("no scope found for repository %s", repoName)
}