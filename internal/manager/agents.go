package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StartAgent starts a specific agent (legacy method for compatibility)
func (m *Manager) StartAgent(agentName string) error {
	return m.StartAgentWithOptions(agentName, StartOptions{})
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

// StartAllAgents starts all auto-start agents (legacy method for compatibility)
func (m *Manager) StartAllAgents() error {
	return m.StartAllAgentsWithOptions(StartOptions{})
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
	fmt.Println("  rc ps        # Show agent processes")
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