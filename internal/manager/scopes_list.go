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

// ScopeInfo combines scope configuration with runtime information
type ScopeInfo struct {
	Name         string
	Status       string
	PID          int
	Repos        []string
	Description  string
	ProcessInfo  *ProcessInfo
	LastActivity time.Time
}

// ListScopes displays information about scopes
func (m *Manager) ListScopes(opts AgentListOptions) error {
	scopes, err := m.GetScopesInfo(opts)
	if err != nil {
		return err
	}
	
	// Sort scopes based on options
	sortScopes(scopes, opts.SortBy)
	
	// Display based on format
	switch opts.Format {
	case "numbered":
		return m.displayScopesNumbered(scopes, opts)
	default:
		return m.displayScopesTable(scopes, opts)
	}
}

// GetScopesInfo gathers information about all scopes
func (m *Manager) GetScopesInfo(opts AgentListOptions) ([]*ScopeInfo, error) {
	var scopes []*ScopeInfo
	
	// Load current state
	statePath := fmt.Sprintf("%s/.repo-claude-state.json", m.ProjectPath)
	state, err := config.LoadState(statePath)
	if err != nil {
		return nil, err
	}
	
	// Get info for each scope
	for name, scopeConfig := range m.Config.Scopes {
		info := &ScopeInfo{
			Name:        name,
			Description: scopeConfig.Description,
			Repos:       m.resolveScopeRepos(scopeConfig.Repos),
			Status:      "stopped",
		}
		
		// Check if scope is in state
		if scopeStatus, exists := state.Scopes[name]; exists {
			info.Status = scopeStatus.Status
			info.PID = scopeStatus.PID
			
			if t, err := time.Parse(time.RFC3339, scopeStatus.LastActivity); err == nil {
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
					scopeStatus.Status = "stopped"
					state.Scopes[name] = scopeStatus
					state.Save(statePath)
				}
			}
		}
		
		// Add to list if showing all or if running
		if opts.ShowAll || info.Status == "running" {
			scopes = append(scopes, info)
		}
	}
	
	return scopes, nil
}

// displayScopesNumbered shows scopes in numbered format for easy kill reference
func (m *Manager) displayScopesNumbered(scopes []*ScopeInfo, opts AgentListOptions) error {
	if len(scopes) == 0 {
		fmt.Println("No scopes to display. Use 'rc start' to start scopes.")
		return nil
	}
	
	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	// Print header
	fmt.Fprintln(w, "#\tSCOPE\tSTATUS\tPID\tREPOS")
	fmt.Fprintln(w, "-\t-----\t------\t---\t-----")
	
	// Store scope mapping for kill command
	m.mu.Lock()
	m.numberToScope = make(map[int]string)
	m.mu.Unlock()
	
	for i, scope := range scopes {
		num := i + 1
		
		// Store mapping
		m.mu.Lock()
		m.numberToScope[num] = scope.Name
		m.mu.Unlock()
		
		status := "⚫"
		if scope.Status == "running" {
			status = "🟢"
		}
		
		pidStr := "-"
		if scope.PID > 0 {
			pidStr = fmt.Sprintf("%d", scope.PID)
		}
		
		// Format repos
		reposStr := strings.Join(scope.Repos, ", ")
		if reposStr == "" {
			reposStr = "(no repos)"
		}
		if len(reposStr) > 50 {
			reposStr = reposStr[:47] + "..."
		}
		
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", num, scope.Name, status, pidStr, reposStr)
	}
	
	w.Flush()
	
	// Tips
	fmt.Println("\n💡 Usage:")
	fmt.Println("  rc kill 1        # Kill by number")
	fmt.Println("  rc kill backend  # Kill by name")
	
	return nil
}

// displayScopesTable shows scopes in a formatted table
func (m *Manager) displayScopesTable(scopes []*ScopeInfo, opts AgentListOptions) error {
	if len(scopes) == 0 {
		fmt.Println("No scopes running. Use 'rc start' to start scopes.")
		return nil
	}
	
	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	// Print header
	fmt.Println("\n" + strings.Repeat("=", 90))
	fmt.Println(" REPO-CLAUDE SCOPES")
	fmt.Println(strings.Repeat("=", 90))
	fmt.Printf(" Workspace: %s\n", m.Config.Workspace.Name)
	fmt.Printf(" Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("-", 90))
	
	if opts.ShowDetails {
		// Detailed view with process info
		fmt.Fprintln(w, "SCOPE\tSTATUS\tPID\tCPU%\tMEM(MB)\tTIME\tDESCRIPTION")
		fmt.Fprintln(w, "-----\t------\t---\t----\t-------\t----\t-----------")
		
		for _, scope := range scopes {
			status := scope.Status
			if status == "running" {
				status = "🟢 " + status
			} else {
				status = "⚫ " + status
			}
			
			cpu := "-"
			mem := "-"
			elapsed := "-"
			
			if scope.ProcessInfo != nil {
				cpu = fmt.Sprintf("%.1f", scope.ProcessInfo.CPUPercent)
				mem = fmt.Sprintf("%.1f", scope.ProcessInfo.MemoryMB)
				elapsed = scope.ProcessInfo.ElapsedTime
			}
			
			// Truncate description if too long
			desc := scope.Description
			if len(desc) > 30 {
				desc = desc[:27] + "..."
			}
			
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\n",
				scope.Name, status, scope.PID, cpu, mem, elapsed, desc)
		}
	} else {
		// Simple view
		fmt.Fprintln(w, "SCOPE\tSTATUS\tPID\tREPOS\tDESCRIPTION")
		fmt.Fprintln(w, "-----\t------\t---\t-----\t-----------")
		
		for _, scope := range scopes {
			status := scope.Status
			if status == "running" {
				status = "🟢 " + status
			} else {
				status = "⚫ " + status
			}
			
			// Format repos
			reposStr := strings.Join(scope.Repos, ", ")
			if len(reposStr) > 30 {
				reposStr = reposStr[:27] + "..."
			}
			
			// Truncate description
			desc := scope.Description
			if len(desc) > 30 {
				desc = desc[:27] + "..."
			}
			
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
				scope.Name, status, scope.PID, reposStr, desc)
		}
	}
	
	w.Flush()
	
	// Summary
	fmt.Println(strings.Repeat("=", 90))
	running := 0
	for _, scope := range scopes {
		if scope.Status == "running" {
			running++
		}
	}
	fmt.Printf(" Summary: %d running, %d total\n", running, len(scopes))
	
	// Tips
	fmt.Println("\n💡 Tips:")
	fmt.Println("  rc ps -a            # Show all scopes")
	fmt.Println("  rc ps -x            # Show extended information")
	fmt.Println("  rc start <scope>    # Start a specific scope")
	fmt.Println("  rc kill <scope>     # Kill a specific scope")
	
	return nil
}

// sortScopes sorts the scope list based on the specified field
func sortScopes(scopes []*ScopeInfo, sortBy string) {
	sort.Slice(scopes, func(i, j int) bool {
		switch sortBy {
		case "pid":
			return scopes[i].PID < scopes[j].PID
		case "cpu":
			if scopes[i].ProcessInfo != nil && scopes[j].ProcessInfo != nil {
				return scopes[i].ProcessInfo.CPUPercent > scopes[j].ProcessInfo.CPUPercent
			}
			return false
		case "mem":
			if scopes[i].ProcessInfo != nil && scopes[j].ProcessInfo != nil {
				return scopes[i].ProcessInfo.MemoryMB > scopes[j].ProcessInfo.MemoryMB
			}
			return false
		case "time":
			return scopes[i].LastActivity.After(scopes[j].LastActivity)
		default: // name
			return scopes[i].Name < scopes[j].Name
		}
	})
}