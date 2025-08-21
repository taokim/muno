package manager

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	
	"github.com/taokim/repo-claude/internal/config"
)

// StartOptions defines how agents should be started
type StartOptions struct {
	NewWindow  bool // Open in new window instead of tab (default: false, open in tab)
	LogOutput  bool // Log output to files (kept for debugging)
}

// StartAgentWithOptions starts an agent with specific options
func (m *Manager) StartAgentWithOptions(agentName string, opts StartOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if agent exists in config
	agentConfig, exists := m.Config.Agents[agentName]
	if !exists {
		return fmt.Errorf("agent %s not found in configuration (legacy mode)", agentName)
	}

	// Check if already running
	if agent, exists := m.agents[agentName]; exists && agent.Process != nil {
		return fmt.Errorf("agent %s is already running (legacy mode)", agentName)
	}

	// Find repository for this agent
	var repoPath string
	for _, project := range m.Config.Workspace.Manifest.Projects {
		if project.Agent == agentName {
			path := project.Name
			if project.Path != "" {
				path = project.Path
			}
			repoPath = filepath.Join(m.WorkspacePath, path)
			break
		}
	}

	if repoPath == "" {
		return fmt.Errorf("no repository assigned to agent %s (legacy mode)", agentName)
	}

	// Check if repository exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("repository %s not found, run 'rc sync' first", repoPath)
	}

	location := "new tab"
	if opts.NewWindow {
		location = "new window"
	}
	fmt.Printf("🚀 Starting %s in %s (%s) [legacy agent mode]\n", agentName, repoPath, location)

	// Build command - always open in new terminal window
	systemPrompt := fmt.Sprintf("You are %s, specialized in: %s. "+
		"You are working in a multi-agent environment. "+
		"Check shared-memory.md for coordination with other agents.",
		agentName, agentConfig.Specialization)

	// Set environment variables for the Claude session
	// TODO: Will be replaced with scope-based env vars in config refactor
	envVars := map[string]string{
		"RC_AGENT_NAME":    agentName,
		"RC_WORKSPACE_ROOT": m.WorkspacePath,
	}

	var cmd Cmd
	if m.CmdExecutor != nil {
		// Use the command executor interface for testing
		switch runtime.GOOS {
		case "darwin":
			script := fmt.Sprintf(`
				tell application "Terminal"
					do script "cd %s && export RC_AGENT_NAME='%s'; export RC_WORKSPACE_ROOT='%s'; claude --model %s --append-system-prompt '%s'"
					activate
				end tell
			`, repoPath, agentName, m.WorkspacePath, agentConfig.Model, systemPrompt)
			cmd = m.CmdExecutor.Command("osascript", "-e", script)
		default:
			// For other platforms, use a simple command for testing
			cmd = m.CmdExecutor.Command("claude", "--model", agentConfig.Model, "--append-system-prompt", systemPrompt)
			cmd.SetDir(repoPath)
		}
	} else {
		// Use the real implementation
		realCmd := createNewTerminalCommand(agentName, repoPath, agentConfig.Model, systemPrompt, envVars, opts.NewWindow)
		cmd = &RealCmd{cmd: realCmd}
	}

	// Start the command
	err := cmd.Start()

	if err != nil {
		// If we tried to open in a tab and it failed, try opening in a new window
		if !opts.NewWindow && runtime.GOOS == "darwin" {
			fmt.Printf("⚠️  Failed to open in tab, trying new window...\n")
			opts.NewWindow = true
			m.mu.Unlock() // Unlock before recursive call
			err := m.StartAgentWithOptions(agentName, opts)
			m.mu.Lock() // Re-lock after return
			return err
		}
		return fmt.Errorf("failed to start: %w", err)
	}

	// Track agent (always track since we always open in new window)
	m.agents[agentName] = &Agent{
		Name:    agentName,
		Process: cmd.Process(),
		Status:  "running",
	}

	// Update state
	if m.State == nil {
		m.State = &config.State{
			Agents: make(map[string]config.AgentStatus),
		}
	}
	m.State.Agents[agentName] = config.AgentStatus{
		Name:         agentName,
		Status:       "running",
		PID:          cmd.Process().Pid,
		Repository:   filepath.Base(repoPath),
		LastActivity: time.Now().Format(time.RFC3339),
	}
	m.State.Save(filepath.Join(m.WorkspacePath, ".repo-claude-state.json"))

	fmt.Printf("✅ %s started (PID: %d) [legacy agent mode]\n", agentName, cmd.Process().Pid)

	return nil
}

// StartAllAgentsWithOptions starts all auto-start agents with options
func (m *Manager) StartAllAgentsWithOptions(opts StartOptions) error {
	fmt.Println("🚀 Starting all auto-start agents... [legacy mode]")

	// Get agents that should auto-start
	var toStart []string
	for name, config := range m.Config.Agents {
		if config.AutoStart {
			toStart = append(toStart, name)
		}
	}

	if len(toStart) == 0 {
		fmt.Println("No auto-start agents configured [legacy mode]")
		return nil
	}

	// Start agents respecting dependencies
	started := make(map[string]bool)
	for len(started) < len(toStart) {
		progress := false
		for _, name := range toStart {
			if started[name] {
				continue
			}

			// Check dependencies
			agent := m.Config.Agents[name]
			ready := true
			for _, dep := range agent.Dependencies {
				if !started[dep] {
					ready = false
					break
				}
			}

			if ready {
				if err := m.StartAgentWithOptions(name, opts); err != nil {
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

// StartByRepos starts agents assigned to specific repositories
func (m *Manager) StartByRepos(repos []string, opts StartOptions) error {
	fmt.Printf("🚀 Starting agents for repositories: %s [legacy mode]\n", strings.Join(repos, ", "))
	
	agentsToStart := make(map[string]bool)
	
	// Find agents assigned to these repos
	for _, repo := range repos {
		for _, project := range m.Config.Workspace.Manifest.Projects {
			if project.Name == repo && project.Agent != "" {
				agentsToStart[project.Agent] = true
			}
		}
	}
	
	if len(agentsToStart) == 0 {
		return fmt.Errorf("no agents assigned to repositories: %s [legacy mode]", strings.Join(repos, ", "))
	}
	
	// Start the agents
	for agent := range agentsToStart {
		if err := m.StartAgentWithOptions(agent, opts); err != nil {
			fmt.Printf("Failed to start %s: %v\n", agent, err)
		}
	}
	
	return nil
}

// StartPreset starts agents based on a predefined preset
func (m *Manager) StartPreset(presetName string, opts StartOptions) error {
	// This would read from configuration presets
	// For now, implement some hardcoded presets as examples
	
	presets := map[string][]string{
		"fullstack": {"frontend", "backend"},
		"backend-all": {"auth-service", "api-gateway", "data-service"},
		"minimal": {"frontend"},
	}
	
	repos, exists := presets[presetName]
	if !exists {
		return fmt.Errorf("preset '%s' not found", presetName)
	}
	
	fmt.Printf("🚀 Starting preset '%s'\n", presetName)
	return m.StartByRepos(repos, opts)
}

// StartInteractive allows interactive selection of agents and repos
func (m *Manager) StartInteractive(opts StartOptions) error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("\n🎯 Interactive Agent Selection")
	fmt.Println(strings.Repeat("-", 40))
	
	// Show available repositories and their agents
	fmt.Println("\nAvailable repositories and agents:")
	repoAgentMap := make(map[string]string)
	for i, project := range m.Config.Workspace.Manifest.Projects {
		agent := project.Agent
		if agent == "" {
			agent = "(no agent)"
		} else {
			repoAgentMap[project.Name] = agent
		}
		fmt.Printf("  %d. %-20s → %s\n", i+1, project.Name, agent)
	}
	
	// Get selection
	fmt.Print("\nSelect repositories (comma-separated numbers or names): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" {
		return fmt.Errorf("no selection made")
	}
	
	// Parse selection
	selections := strings.Split(input, ",")
	selectedRepos := []string{}
	
	for _, sel := range selections {
		sel = strings.TrimSpace(sel)
		// Check if it's a number
		num := 0
		if n, _ := fmt.Sscanf(sel, "%d", &num); n == 1 && num > 0 && num <= len(m.Config.Workspace.Manifest.Projects) {
			selectedRepos = append(selectedRepos, m.Config.Workspace.Manifest.Projects[num-1].Name)
		} else {
			// Assume it's a repo name
			selectedRepos = append(selectedRepos, sel)
		}
	}
	
	return m.StartByRepos(selectedRepos, opts)
}

// createNewTerminalCommand creates a command to open claude in a new terminal tab or window
func createNewTerminalCommand(agentName, repoPath, model, systemPrompt string, envVars map[string]string, newWindow bool) *exec.Cmd {
	// Build environment variable exports
	var envExports []string
	for k, v := range envVars {
		envExports = append(envExports, fmt.Sprintf("export %s='%s'", k, v))
	}
	envExportsStr := strings.Join(envExports, "; ")
	
	claudeCmd := fmt.Sprintf("%s; claude --model %s --append-system-prompt '%s'", envExportsStr, model, systemPrompt)
	
	switch runtime.GOOS {
	case "darwin": // macOS
		if newWindow {
			// Open in new window
			script := fmt.Sprintf(`
				tell application "Terminal"
					do script "cd %s && %s"
					activate
				end tell
			`, repoPath, claudeCmd)
			return exec.Command("osascript", "-e", script)
		} else {
			// Try to open in new tab first
			fullCommand := fmt.Sprintf("cd %s && %s", repoPath, claudeCmd)
			
			// First try to open in a new tab
			tabScript := fmt.Sprintf(`
				tell application "Terminal"
					activate
					tell application "System Events" to keystroke "t" using command down
					delay 0.5
					do script "%s" in selected tab of the front window
				end tell
			`, fullCommand)
			
			// Create a command that tries tab first, falls back to window
			cmd := exec.Command("osascript", "-e", tabScript)
			// If tab creation fails, we'll handle it in the caller
			return cmd
		}
		
	case "linux":
		// Try common terminal emulators
		terminals := []struct {
			cmd  string
			args []string
		}{
			{"gnome-terminal", []string{"--", "bash", "-c", fmt.Sprintf("cd %s && %s; exec bash", repoPath, claudeCmd)}},
			{"konsole", []string{"-e", "bash", "-c", fmt.Sprintf("cd %s && %s; exec bash", repoPath, claudeCmd)}},
			{"xterm", []string{"-e", "bash", "-c", fmt.Sprintf("cd %s && %s; exec bash", repoPath, claudeCmd)}},
		}
		
		for _, term := range terminals {
			if _, err := exec.LookPath(term.cmd); err == nil {
				return exec.Command(term.cmd, term.args...)
			}
		}
		
		// Fallback
		return exec.Command("xterm", "-e", "bash", "-c", fmt.Sprintf("cd %s && %s; exec bash", repoPath, claudeCmd))
		
	case "windows":
		// Windows Terminal or cmd
		return exec.Command("cmd", "/c", "start", "cmd", "/k", fmt.Sprintf("cd /d %s && %s", repoPath, claudeCmd))
		
	default:
		// Fallback to running in same terminal
		cmd := exec.Command("claude", "--model", model, "--append-system-prompt", systemPrompt)
		cmd.Dir = repoPath
		return cmd
	}
}