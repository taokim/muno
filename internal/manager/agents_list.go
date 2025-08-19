package manager

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/taokim/repo-claude/internal/config"
)

// AgentListOptions defines options for listing agents
type AgentListOptions struct {
	ShowAll     bool   // Show all agents including stopped ones
	ShowDetails bool   // Show detailed process information
	ShowLogs    bool   // Show recent log entries
	Format      string // Output format: table, json, simple
	SortBy      string // Sort by: name, pid, cpu, mem, time
}

// AgentInfo combines agent configuration with runtime information
type AgentInfo struct {
	Name           string
	Status         string
	PID            int
	Repository     string
	Specialization string
	ProcessInfo    *ProcessInfo
	LastActivity   time.Time
	LogPreview     []string
}

// ListAgents displays information about agents
func (m *Manager) ListAgents(opts AgentListOptions) error {
	agents, err := m.GetAgentsInfo(opts)
	if err != nil {
		return err
	}
	
	// Sort agents based on options
	sortAgents(agents, opts.SortBy)
	
	// Display based on format
	switch opts.Format {
	case "json":
		return m.displayAgentsJSON(agents)
	case "simple":
		return m.displayAgentsSimple(agents)
	default:
		return m.displayAgentsTable(agents, opts)
	}
}

// GetAgentsInfo gathers information about all agents
func (m *Manager) GetAgentsInfo(opts AgentListOptions) ([]*AgentInfo, error) {
	var agents []*AgentInfo
	
	// Load current state
	statePath := fmt.Sprintf("%s/.repo-claude-state.json", m.WorkspacePath)
	state, err := config.LoadState(statePath)
	if err != nil {
		return nil, err
	}
	
	// Get info for each agent
	for name, agentConfig := range m.Config.Agents {
		info := &AgentInfo{
			Name:           name,
			Specialization: agentConfig.Specialization,
			Status:         "stopped",
		}
		
		// Check if agent is in state
		if agentStatus, exists := state.Agents[name]; exists {
			info.Status = agentStatus.Status
			info.PID = agentStatus.PID
			info.Repository = agentStatus.Repository
			
			if t, err := time.Parse(time.RFC3339, agentStatus.LastActivity); err == nil {
				info.LastActivity = t
			}
			
			// Get process info if running
			if info.Status == "running" && info.PID > 0 {
				// Verify process is actually running
				pm := m.ProcessManager
				if pm == nil {
					pm = RealProcessManager{}
				}
				if healthy, _ := CheckProcessHealthWithManager(info.PID, pm); healthy {
					provider := &ProcessInfoProvider{
						ProcessManager: pm,
						CmdExecutor:    m.CmdExecutor,
					}
					if provider.CmdExecutor == nil {
						provider.CmdExecutor = &RealCommandExecutor{}
					}
					if procInfo, err := provider.GetProcessInfo(info.PID); err == nil {
						info.ProcessInfo = procInfo
					}
				} else {
					// Process died, update status
					info.Status = "stopped"
					agentStatus.Status = "stopped"
					state.Agents[name] = agentStatus
					state.Save(statePath)
				}
			}
			
			// Get log preview if requested
			if opts.ShowLogs && info.Status == "running" {
				if logs, err := m.GetAgentLogs(name, 3); err == nil {
					info.LogPreview = logs
				}
			}
		}
		
		// Add to list if showing all or if running
		if opts.ShowAll || info.Status == "running" {
			agents = append(agents, info)
		}
	}
	
	return agents, nil
}

// displayAgentsTable shows agents in a formatted table
func (m *Manager) displayAgentsTable(agents []*AgentInfo, opts AgentListOptions) error {
	if len(agents) == 0 {
		fmt.Println("No agents running. Use 'rc start' to start agents.")
		return nil
	}
	
	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	// Print header
	fmt.Println("\n" + strings.Repeat("=", 90))
	fmt.Println(" CLAUDE CODE AGENTS")
	fmt.Println(strings.Repeat("=", 90))
	fmt.Printf(" Workspace: %s\n", m.Config.Workspace.Name)
	fmt.Printf(" Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("-", 90))
	
	if opts.ShowDetails {
		// Detailed view with process info
		fmt.Fprintln(w, "NAME\tSTATUS\tPID\tCPU%\tMEM(MB)\tTIME\tREPO\tCOMMAND")
		fmt.Fprintln(w, "----\t------\t---\t----\t-------\t----\t----\t-------")
		
		for _, agent := range agents {
			status := agent.Status
			if status == "running" {
				status = "🟢 " + status
			} else {
				status = "⚫ " + status
			}
			
			cpu := "-"
			mem := "-"
			elapsed := "-"
			command := "-"
			
			if agent.ProcessInfo != nil {
				cpu = fmt.Sprintf("%.1f", agent.ProcessInfo.CPUPercent)
				mem = fmt.Sprintf("%.1f", agent.ProcessInfo.MemoryMB)
				elapsed = agent.ProcessInfo.ElapsedTime
				
				// Truncate command if too long
				command = agent.ProcessInfo.Command
				if len(command) > 40 {
					command = command[:37] + "..."
				}
			}
			
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
				agent.Name, status, agent.PID, cpu, mem, elapsed, agent.Repository, command)
		}
	} else {
		// Simple view
		fmt.Fprintln(w, "NAME\tSTATUS\tPID\tREPOSITORY\tSPECIALIZATION")
		fmt.Fprintln(w, "----\t------\t---\t----------\t--------------")
		
		for _, agent := range agents {
			status := agent.Status
			if status == "running" {
				status = "🟢 " + status
			} else {
				status = "⚫ " + status
			}
			
			// Truncate specialization if too long
			spec := agent.Specialization
			if len(spec) > 40 {
				spec = spec[:37] + "..."
			}
			
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
				agent.Name, status, agent.PID, agent.Repository, spec)
		}
	}
	
	w.Flush()
	
	// Show log preview if requested
	if opts.ShowLogs {
		for _, agent := range agents {
			if len(agent.LogPreview) > 0 && agent.Status == "running" {
				fmt.Printf("\n📋 Recent logs for %s:\n", agent.Name)
				fmt.Println(strings.Repeat("-", 70))
				for _, line := range agent.LogPreview {
					if line != "" {
						fmt.Printf("  %s\n", line)
					}
				}
			}
		}
	}
	
	// Summary
	fmt.Println(strings.Repeat("=", 90))
	running := 0
	for _, agent := range agents {
		if agent.Status == "running" {
			running++
		}
	}
	fmt.Printf(" Summary: %d running, %d total\n", running, len(agents))
	
	// Tips
	fmt.Println("\n💡 Tips:")
	fmt.Println("  rc ps aux           # Show all agents with details")
	fmt.Println("  rc ps -ef           # Full format listing")
	fmt.Println("  rc ps --logs        # Show with recent logs")
	fmt.Println("  rc start <agent>    # Start a specific agent")
	fmt.Println("  rc stop <agent>     # Stop a specific agent")
	
	return nil
}

// displayAgentsSimple shows agents in simple format
func (m *Manager) displayAgentsSimple(agents []*AgentInfo) error {
	for _, agent := range agents {
		if agent.Status == "running" {
			fmt.Printf("%s (PID: %d) - %s\n", agent.Name, agent.PID, agent.Repository)
		}
	}
	return nil
}

// displayAgentsJSON shows agents in JSON format
func (m *Manager) displayAgentsJSON(agents []*AgentInfo) error {
	// Implementation for JSON output
	// This would use json.Marshal to output the agent information
	fmt.Println("JSON output not yet implemented")
	return nil
}

// sortAgents sorts the agent list based on the specified field
func sortAgents(agents []*AgentInfo, sortBy string) {
	sort.Slice(agents, func(i, j int) bool {
		switch sortBy {
		case "pid":
			return agents[i].PID < agents[j].PID
		case "cpu":
			if agents[i].ProcessInfo != nil && agents[j].ProcessInfo != nil {
				return agents[i].ProcessInfo.CPUPercent > agents[j].ProcessInfo.CPUPercent
			}
			return false
		case "mem":
			if agents[i].ProcessInfo != nil && agents[j].ProcessInfo != nil {
				return agents[i].ProcessInfo.MemoryMB > agents[j].ProcessInfo.MemoryMB
			}
			return false
		case "time":
			return agents[i].LastActivity.After(agents[j].LastActivity)
		default: // name
			return agents[i].Name < agents[j].Name
		}
	})
}