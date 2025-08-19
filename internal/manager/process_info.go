package manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ProcessInfo contains detailed information about a running process
type ProcessInfo struct {
	PID         int
	Name        string
	Status      string
	CPUPercent  float64
	MemoryMB    float64
	StartTime   time.Time
	ElapsedTime string
	Command     string
}

// GetProcessInfo retrieves detailed process information
func GetProcessInfo(pid int) (*ProcessInfo, error) {
	// Use default provider with real implementations
	provider := &ProcessInfoProvider{
		CmdExecutor:    RealCommandExecutor{},
		ProcessManager: RealProcessManager{},
	}
	return provider.GetProcessInfo(pid)
}

// ProcessInfoProvider uses dependency injection for testability
type ProcessInfoProvider struct {
	CmdExecutor    CommandExecutor
	ProcessManager ProcessManager
}

// GetProcessInfo retrieves process information using injected dependencies
func (p *ProcessInfoProvider) GetProcessInfo(pid int) (*ProcessInfo, error) {
	info := &ProcessInfo{PID: pid}
	
	switch runtime.GOOS {
	case "darwin", "linux":
		return p.getUnixProcessInfo(pid)
	case "windows":
		return p.getWindowsProcessInfo(pid)
	default:
		return info, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// getUnixProcessInfo gets process info on Unix-like systems
func (p *ProcessInfoProvider) getUnixProcessInfo(pid int) (*ProcessInfo, error) {
	info := &ProcessInfo{PID: pid}
	
	// Check if process exists
	process, err := p.ProcessManager.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("process not found: %w", err)
	}
	
	// Try to send signal 0 to check if process is alive
	err = p.ProcessManager.Signal(process, os.Signal(nil))
	if err != nil {
		info.Status = "stopped"
		return info, nil
	}
	
	info.Status = "running"
	
	// Get detailed info using ps command
	cmd := p.CmdExecutor.Command("ps", "-p", strconv.Itoa(pid), "-o", "pid,ppid,%cpu,%mem,rss,vsz,etime,command")
	output, err := cmd.Output()
	if err != nil {
		return info, nil // Return basic info if ps fails
	}
	
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return info, nil
	}
	
	// Parse the output (skip header)
	fields := strings.Fields(lines[1])
	if len(fields) >= 8 {
		// Parse CPU percentage
		if cpu, err := strconv.ParseFloat(fields[2], 64); err == nil {
			info.CPUPercent = cpu
		}
		
		// Parse memory percentage and convert to MB (approximate)
		if memPercent, err := strconv.ParseFloat(fields[3], 64); err == nil {
			// Get total system memory to calculate MB from percentage
			if totalMem := getSystemMemory(); totalMem > 0 {
				info.MemoryMB = (memPercent / 100) * float64(totalMem) / 1024 / 1024
			}
		}
		
		// Parse elapsed time
		info.ElapsedTime = fields[6]
		
		// Get command
		info.Command = strings.Join(fields[7:], " ")
	}
	
	// Try to get more accurate start time
	if runtime.GOOS == "darwin" {
		// macOS specific: use ps with lstart option
		cmd = p.CmdExecutor.Command("ps", "-p", strconv.Itoa(pid), "-o", "lstart")
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) >= 2 && lines[1] != "" {
				// Parse the start time (format varies by system)
				if t, err := parseStartTime(strings.TrimSpace(lines[1])); err == nil {
					info.StartTime = t
				}
			}
		}
	}
	
	return info, nil
}

// getWindowsProcessInfo gets process info on Windows
func (p *ProcessInfoProvider) getWindowsProcessInfo(pid int) (*ProcessInfo, error) {
	info := &ProcessInfo{PID: pid}
	
	// Use wmic to get process information
	cmd := p.CmdExecutor.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%d", pid), 
		"get", "ProcessId,Name,Status,WorkingSetSize,KernelModeTime,UserModeTime,CommandLine", "/format:csv")
	output, err := cmd.Output()
	if err != nil {
		info.Status = "stopped"
		return info, nil
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, strconv.Itoa(pid)) {
			fields := strings.Split(line, ",")
			if len(fields) >= 7 {
				info.Status = "running"
				// Parse memory (WorkingSetSize is in bytes)
				if mem, err := strconv.ParseFloat(fields[7], 64); err == nil {
					info.MemoryMB = mem / 1024 / 1024
				}
				info.Command = fields[2] // CommandLine field
			}
		}
	}
	
	return info, nil
}

// getSystemMemory tries to get total system memory
func getSystemMemory() int64 {
	switch runtime.GOOS {
	case "darwin":
		// macOS: use sysctl
		cmd := exec.Command("sysctl", "-n", "hw.memsize")
		if output, err := cmd.Output(); err == nil {
			if mem, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64); err == nil {
				return mem
			}
		}
	case "linux":
		// Linux: read from /proc/meminfo
		if data, err := os.ReadFile("/proc/meminfo"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "MemTotal:") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						if mem, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
							return mem * 1024 // Convert KB to bytes
						}
					}
				}
			}
		}
	}
	return 0
}

// parseStartTime attempts to parse various time formats
func parseStartTime(timeStr string) (time.Time, error) {
	// Try common formats
	formats := []string{
		"Mon Jan _2 15:04:05 2006",
		"Mon Jan _2 15:04:05 MST 2006",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd%dh%dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

// CheckProcessHealth checks if a process is healthy
func CheckProcessHealth(pid int) (bool, error) {
	pm := RealProcessManager{}
	return CheckProcessHealthWithManager(pid, pm)
}

// CheckProcessHealthWithManager checks if a process is healthy using provided manager
func CheckProcessHealthWithManager(pid int, pm ProcessManager) (bool, error) {
	process, err := pm.FindProcess(pid)
	if err != nil {
		return false, err
	}
	
	// Send signal 0 to check if process exists
	err = pm.Signal(process, os.Signal(nil))
	return err == nil, nil
}

// GetAgentLogs retrieves recent logs for an agent
func (m *Manager) GetAgentLogs(agentName string, lines int) ([]string, error) {
	logDir := filepath.Join(m.WorkspacePath, ".logs")
	pattern := filepath.Join(logDir, fmt.Sprintf("%s-*.log", agentName))
	
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("no logs found for agent %s", agentName)
	}
	
	// Get the most recent log file
	latestLog := matches[len(matches)-1]
	
	// Use tail command on Unix-like systems
	if runtime.GOOS != "windows" {
		cmd := exec.Command("tail", "-n", strconv.Itoa(lines), latestLog)
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		return strings.Split(string(output), "\n"), nil
	}
	
	// For Windows, read the file and get last N lines
	data, err := os.ReadFile(latestLog)
	if err != nil {
		return nil, err
	}
	
	allLines := strings.Split(string(data), "\n")
	start := len(allLines) - lines
	if start < 0 {
		start = 0
	}
	
	return allLines[start:], nil
}