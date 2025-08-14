package manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourusername/repo-claude/internal/config"
)

// StartAgent starts a specific agent
func (m *Manager) StartAgent(agentName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if agent exists in config
	agentConfig, exists := m.Config.Agents[agentName]
	if !exists {
		return fmt.Errorf("agent %s not found", agentName)
	}

	// Find repository for this agent
	project, err := m.Config.GetProjectForAgent(agentName)
	if err != nil {
		return fmt.Errorf("no repository found for agent %s", agentName)
	}

	repoPath := filepath.Join(m.WorkspacePath, project.Name)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository %s not found at %s\nRun: ./repo-claude sync", project.Name, repoPath)
	}

	fmt.Printf("🚀 Starting %s for %s\n", agentName, project.Name)

	// Build system prompt
	systemPrompt := fmt.Sprintf(
		"You are %s, specialized in: %s. "+
			"You are working in a trunk-based development environment managed by Repo tool. "+
			"Multiple coordinated agents are working in parallel across different repositories.",
		agentName, agentConfig.Specialization)

	// Start Claude Code
	cmd := exec.Command("claude",
		"--model", agentConfig.Model,
		"--append-system-prompt", systemPrompt)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", agentName, err)
	}

	// Track the agent
	m.agents[agentName] = &Agent{
		Name:    agentName,
		Process: cmd.Process,
		Status:  "running",
	}

	// Update state
	m.State.UpdateAgent(config.AgentStatus{
		Name:       agentName,
		Status:     "running",
		PID:        cmd.Process.Pid,
		Repository: project.Name,
	})

	statePath := filepath.Join(m.WorkspacePath, ".repo-claude-state.json")
	if err := m.State.Save(statePath); err != nil {
		fmt.Printf("⚠️  Failed to save state: %v\n", err)
	}

	fmt.Printf("✅ %s started (PID: %d)\n", agentName, cmd.Process.Pid)
	return nil
}

// StopAgent stops a specific agent
func (m *Manager) StopAgent(agentName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, exists := m.agents[agentName]
	if !exists {
		return fmt.Errorf("%s is not running", agentName)
	}

	// Terminate the process
	if err := agent.Process.Signal(os.Interrupt); err != nil {
		// If interrupt fails, try kill
		if err := agent.Process.Kill(); err != nil {
			return fmt.Errorf("failed to stop %s: %w", agentName, err)
		}
	}

	// Wait for process to exit
	_, _ = agent.Process.Wait()

	// Update tracking
	delete(m.agents, agentName)

	// Update state
	if status, exists := m.State.Agents[agentName]; exists {
		status.Status = "stopped"
		m.State.UpdateAgent(status)
	}

	statePath := filepath.Join(m.WorkspacePath, ".repo-claude-state.json")
	if err := m.State.Save(statePath); err != nil {
		fmt.Printf("⚠️  Failed to save state: %v\n", err)
	}

	fmt.Printf("🛑 %s stopped\n", agentName)
	return nil
}

// StartAllAgents starts all agents with auto_start=true
func (m *Manager) StartAllAgents() error {
	fmt.Println("🚀 Starting all auto-start agents...")

	// Ensure repositories are up to date
	if err := m.Sync(); err != nil {
		fmt.Printf("⚠️  Sync warning: %v\n", err)
	}

	// Get agents that should auto-start
	toStart := []string{}
	for name, agent := range m.Config.Agents {
		if agent.AutoStart {
			toStart = append(toStart, name)
		}
	}

	// Start agents in dependency order
	started := make(map[string]bool)
	remaining := make(map[string]bool)
	for _, name := range toStart {
		remaining[name] = true
	}

	for len(remaining) > 0 {
		ready := []string{}
		
		for name := range remaining {
			agent := m.Config.Agents[name]
			canStart := true
			
			// Check dependencies
			for _, dep := range agent.Dependencies {
				if !started[dep] && m.Config.Agents[dep].AutoStart {
					canStart = false
					break
				}
			}
			
			if canStart {
				ready = append(ready, name)
			}
		}

		if len(ready) == 0 && len(remaining) > 0 {
			return fmt.Errorf("circular dependency detected")
		}

		// Start ready agents
		for _, name := range ready {
			if err := m.StartAgent(name); err != nil {
				fmt.Printf("❌ Failed to start %s: %v\n", name, err)
			} else {
				started[name] = true
			}
			delete(remaining, name)
			
			// Small delay between starts
			time.Sleep(500 * time.Millisecond)
		}
	}

	return nil
}

// StopAllAgents stops all running agents
func (m *Manager) StopAllAgents() error {
	fmt.Println("🛑 Stopping all agents...")
	
	// Get list of running agents
	agentNames := []string{}
	m.mu.Lock()
	for name := range m.agents {
		agentNames = append(agentNames, name)
	}
	m.mu.Unlock()

	// Stop each agent
	for _, name := range agentNames {
		if err := m.StopAgent(name); err != nil {
			fmt.Printf("❌ Failed to stop %s: %v\n", name, err)
		}
	}

	return nil
}

// Sync runs repo sync
func (m *Manager) Sync() error {
	fmt.Println("🔄 Syncing all repositories with Repo...")
	return m.repoSync()
}

// ShowStatus displays the current status
func (m *Manager) ShowStatus() error {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println(" REPO-CLAUDE STATUS")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf(" Workspace: %s\n", m.Config.Workspace.Name)
	fmt.Printf(" Location:  %s\n", m.WorkspacePath)
	fmt.Printf(" Manifest:  %s\n", filepath.Join(m.WorkspacePath, ".manifest-repo"))
	fmt.Println(strings.Repeat("-", 70))

	// Show repo projects
	projects := m.getRepoProjects()
	fmt.Printf(" Repo Projects (%d):\n", len(projects))
	for _, project := range projects {
		agent := ""
		for _, p := range m.Config.Workspace.Manifest.Projects {
			if p.Name == project {
				if p.Agent != "" {
					agent = fmt.Sprintf("→ %s", p.Agent)
				} else {
					agent = "no agent"
				}
				break
			}
		}
		fmt.Printf("   📁 %-20s %s\n", project, agent)
	}

	fmt.Println(strings.Repeat("-", 70))

	// Show agent status
	if len(m.State.Agents) == 0 {
		fmt.Println(" Agents: No agents configured")
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
	fmt.Println("💡 Tip: Use 'repo status' to see detailed git status of all projects")

	return nil
}

// getRepoProjects gets list of projects from repo
func (m *Manager) getRepoProjects() []string {
	cmd := exec.Command("repo", "list")
	cmd.Dir = m.WorkspacePath
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	projects := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Format: "path : project"
			parts := strings.Split(line, " : ")
			if len(parts) > 0 {
				projects = append(projects, strings.TrimSpace(parts[0]))
			}
		}
	}

	return projects
}