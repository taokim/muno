package manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/taokim/repo-claude/internal/config"
)

// StartAgent starts a specific agent
func (m *Manager) StartAgent(agentName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if agent exists in config
	agentConfig, exists := m.Config.Agents[agentName]
	if !exists {
		return fmt.Errorf("agent %s not found in configuration", agentName)
	}

	// Check if already running
	if agent, exists := m.agents[agentName]; exists && agent.Process != nil {
		return fmt.Errorf("agent %s is already running", agentName)
	}

	// Find repository for this agent
	var repoPath string
	for _, project := range m.Config.Workspace.Manifest.Projects {
		if project.Agent == agentName {
			// Use project path if specified, otherwise use name
			path := project.Name
			if project.Path != "" {
				path = project.Path
			}
			repoPath = filepath.Join(m.WorkspacePath, path)
			break
		}
	}

	if repoPath == "" {
		return fmt.Errorf("no repository assigned to agent %s", agentName)
	}

	// Check if repository exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("repository %s not found, run 'rc sync' first", repoPath)
	}

	fmt.Printf("🚀 Starting %s in %s\n", agentName, repoPath)

	// Start Claude Code
	cmd := exec.Command("claude",
		"--model", agentConfig.Model,
		"--append-system-prompt",
		fmt.Sprintf("You are %s, specialized in: %s. "+
			"You are working in a multi-agent environment. "+
			"Check shared-memory.md for coordination with other agents.",
			agentName, agentConfig.Specialization))
	cmd.Dir = repoPath

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	// Track agent
	m.agents[agentName] = &Agent{
		Name:    agentName,
		Process: cmd.Process,
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
		PID:          cmd.Process.Pid,
		Repository:   filepath.Base(repoPath),
		LastActivity: time.Now().Format(time.RFC3339),
	}
	m.State.Save(filepath.Join(m.WorkspacePath, ".repo-claude-state.json"))

	fmt.Printf("✅ %s started (PID: %d)\n", agentName, cmd.Process.Pid)
	return nil
}

// StopAgent stops a specific agent
func (m *Manager) StopAgent(agentName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, exists := m.agents[agentName]
	if !exists || agent.Process == nil {
		return fmt.Errorf("agent %s is not running", agentName)
	}

	// Terminate the process
	if err := agent.Process.Signal(os.Interrupt); err != nil {
		if err := agent.Process.Kill(); err != nil {
			return fmt.Errorf("failed to stop agent: %w", err)
		}
	}

	// Wait for process to exit
	agent.Process.Wait()

	// Update tracking
	delete(m.agents, agentName)
	
	// Update state
	if m.State != nil && m.State.Agents != nil {
		if status, exists := m.State.Agents[agentName]; exists {
			status.Status = "stopped"
			m.State.Agents[agentName] = status
			m.State.Save(filepath.Join(m.WorkspacePath, ".repo-claude-state.json"))
		}
	}

	fmt.Printf("🛑 %s stopped\n", agentName)
	return nil
}

// StartAllAgents starts all auto-start agents
func (m *Manager) StartAllAgents() error {
	fmt.Println("🚀 Starting all auto-start agents...")

	// Get agents that should auto-start
	var toStart []string
	for name, config := range m.Config.Agents {
		if config.AutoStart {
			toStart = append(toStart, name)
		}
	}

	if len(toStart) == 0 {
		fmt.Println("No auto-start agents configured")
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
				if err := m.StartAgent(name); err != nil {
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

// StopAllAgents stops all running agents
func (m *Manager) StopAllAgents() error {
	fmt.Println("🛑 Stopping all agents...")

	for name := range m.agents {
		if err := m.StopAgent(name); err != nil {
			fmt.Printf("❌ Failed to stop %s: %v\n", name, err)
		}
	}

	return nil
}

// ShowStatus displays the current status
func (m *Manager) ShowStatus() error {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println(" REPO-CLAUDE STATUS")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf(" Workspace: %s\n", m.Config.Workspace.Name)
	fmt.Printf(" Location:  %s\n", m.WorkspacePath)
	fmt.Println(strings.Repeat("-", 70))

	// Show repositories
	if m.GitManager != nil {
		statuses, _ := m.GitManager.Status()
		fmt.Printf(" Repositories (%d):\n", len(statuses))
		for _, status := range statuses {
			statusIcon := "✓"
			statusText := "clean"
			if status.Error != nil {
				statusIcon = "✗"
				statusText = status.Error.Error()
			} else if !status.Clean {
				statusIcon = "●"
				statusText = fmt.Sprintf("%d modified", len(status.Modified))
			}
			
			// Find agent for this repo
			agent := "no agent"
			for _, p := range m.Config.Workspace.Manifest.Projects {
				if p.Name == status.Name {
					if p.Agent != "" {
						agent = fmt.Sprintf("→ %s", p.Agent)
					}
					break
				}
			}
			
			fmt.Printf("   %s %-20s %-15s %s\n", statusIcon, status.Name, statusText, agent)
		}
	}

	fmt.Println(strings.Repeat("-", 70))

	// Show agent status
	if m.State == nil || len(m.State.Agents) == 0 {
		fmt.Println(" Agents: No agents running")
	} else {
		fmt.Println(" Agent Status:")
		for name, status := range m.State.Agents {
			emoji := "🟢"
			if status.Status != "running" {
				emoji = "⚫"
			}
			fmt.Printf("   %s %-20s %-10s %s\n", emoji, name, status.Status, status.Repository)
		}
	}

	fmt.Println(strings.Repeat("=", 70) + "\n")
	fmt.Println("💡 Commands:")
	fmt.Println("  rc sync      # Sync all repositories")
	fmt.Println("  rc start     # Start agents")
	fmt.Println("  rc forall    # Run command in all repos")

	return nil
}

// ForAll runs a command in all repositories
func (m *Manager) ForAll(command string, args []string) error {
	if m.GitManager == nil {
		return fmt.Errorf("no git manager initialized")
	}
	return m.GitManager.ForAll(command, args...)
}