package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	
	"github.com/taokim/repo-claude/internal/config"
)

// StartScope starts a specific scope
func (m *Manager) StartScope(scopeName string) error {
	return m.StartScopeWithOptions(scopeName, StartOptions{})
}

// StartScopeWithOptions starts a scope with specific options
func (m *Manager) StartScopeWithOptions(scopeName string, opts StartOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if scope exists in config
	scopeConfig, exists := m.Config.Scopes[scopeName]
	if !exists {
		return fmt.Errorf("scope %s not found in configuration", scopeName)
	}

	// Check if already running
	if scope, exists := m.scopes[scopeName]; exists && scope.Process != nil {
		return fmt.Errorf("scope %s is already running", scopeName)
	}

	// Resolve repositories for this scope
	repos := m.resolveScopeRepos(scopeConfig.Repos)
	if len(repos) == 0 {
		return fmt.Errorf("no repositories found for scope %s", scopeName)
	}

	// Use project root as working directory for consistency
	// This ensures all scopes start from the same location
	workDir := m.ProjectPath

	location := "new tab"
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

	cmd := createNewTerminalCommand(scopeName, workDir, scopeConfig.Model, systemPrompt, envVars, opts.NewWindow)

	// Start the command
	err := cmd.Start()
	if err != nil {
		// If we tried to open in a tab and it failed, try opening in a new window
		if !opts.NewWindow && runtime.GOOS == "darwin" {
			fmt.Printf("⚠️  Failed to open in tab, trying new window...\n")
			opts.NewWindow = true
			return m.StartScopeWithOptions(scopeName, opts)
		}
		return fmt.Errorf("failed to start scope: %w", err)
	}

	// Track scope
	m.scopes[scopeName] = &Scope{
		Name:    scopeName,
		Process: cmd.Process,
		Status:  "running",
		Repos:   repos,
	}

	// Update state
	if m.State == nil {
		m.State = &config.State{
			Scopes: make(map[string]config.ScopeStatus),
		}
	}
	m.State.UpdateScope(config.ScopeStatus{
		Name:         scopeName,
		Status:       "running",
		PID:          cmd.Process.Pid,
		Repos:        repos,
		LastActivity: time.Now().Format(time.RFC3339),
	})
	m.State.Save(filepath.Join(m.ProjectPath, ".repo-claude-state.json"))

	fmt.Printf("✅ Scope %s started (PID: %d)\n", scopeName, cmd.Process.Pid)

	return nil
}

// StopScope stops a specific scope
func (m *Manager) StopScope(scopeName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	scope, exists := m.scopes[scopeName]
	if !exists || scope.Process == nil {
		return fmt.Errorf("scope %s is not running", scopeName)
	}

	// Terminate the process
	if err := scope.Process.Signal(os.Interrupt); err != nil {
		if err := scope.Process.Kill(); err != nil {
			return fmt.Errorf("failed to stop scope: %w", err)
		}
	}

	// Wait for process to exit
	scope.Process.Wait()

	// Update tracking
	delete(m.scopes, scopeName)
	
	// Update state
	if m.State != nil && m.State.Scopes != nil {
		if status, exists := m.State.Scopes[scopeName]; exists {
			status.Status = "stopped"
			m.State.Scopes[scopeName] = status
			m.State.Save(filepath.Join(m.ProjectPath, ".repo-claude-state.json"))
		}
	}

	fmt.Printf("🛑 Scope %s stopped\n", scopeName)
	return nil
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

// StopAllScopes stops all running scopes
func (m *Manager) StopAllScopes() error {
	fmt.Println("🛑 Stopping all scopes...")

	for name := range m.scopes {
		if err := m.StopScope(name); err != nil {
			fmt.Printf("❌ Failed to stop %s: %v\n", name, err)
		}
	}

	return nil
}

// KillScopeByNumber kills a scope by its number from ps output
func (m *Manager) KillScopeByNumber(num int) error {
	m.mu.Lock()
	scopeName, exists := m.numberToScope[num]
	m.mu.Unlock()
	
	if !exists {
		return fmt.Errorf("no scope with number %d", num)
	}
	
	return m.StopScope(scopeName)
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