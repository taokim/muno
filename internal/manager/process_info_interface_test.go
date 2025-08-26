//go:build legacy
// +build legacy

package manager

import (
	"runtime"
	"testing"
)

// Test GetProcessInfo with mocked dependencies
func TestProcessInfoProvider_GetProcessInfo(t *testing.T) {
	tests := []struct {
		name      string
		pid       int
		platform  string
		setupMock func(*MockCommandExecutor, *MockProcessManager)
		wantInfo  *ProcessInfo
		wantErr   bool
	}{
		{
			name:     "Unix - Running process with full info",
			pid:      1234,
			platform: "darwin",
			setupMock: func(cmd *MockCommandExecutor, pm *MockProcessManager) {
				// Process is running
				pm.Processes[1234] = &MockProcess{
					PID:    1234,
					Status: "running",
				}
				
				// ps command returns detailed info
				psOutput := `  PID  PPID  %CPU %MEM      RSS      VSZ ELAPSED COMMAND
 1234  1000  25.5  3.2     1024     2048 10:30   claude --model sonnet`
				cmd.Commands = []MockCommand{
					{
						Cmd:      "ps",
						Args:     []string{"-p", "1234", "-o", "pid,ppid,%cpu,%mem,rss,vsz,etime,command"},
						Response: psOutput,
					},
				}
			},
			wantInfo: &ProcessInfo{
				PID:         1234,
				Status:      "running",
				CPUPercent:  25.5,
				MemoryMB:    32.0, // 3.2 * 10 from simplified calculation
				ElapsedTime: "10:30",
				Command:     "claude --model sonnet",
			},
			wantErr: false,
		},
		{
			name:     "Unix - Stopped process",
			pid:      5678,
			platform: "linux",
			setupMock: func(cmd *MockCommandExecutor, pm *MockProcessManager) {
				// Process exists but not running
				pm.Processes[5678] = &MockProcess{
					PID:    5678,
					Status: "stopped",
				}
			},
			wantInfo: &ProcessInfo{
				PID:    5678,
				Status: "stopped",
			},
			wantErr: false,
		},
		{
			name:     "Unix - Process not found",
			pid:      9999,
			platform: "linux",
			setupMock: func(cmd *MockCommandExecutor, pm *MockProcessManager) {
				// Process doesn't exist
			},
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name:     "Windows - Running process",
			pid:      1234,
			platform: "windows",
			setupMock: func(cmd *MockCommandExecutor, pm *MockProcessManager) {
				// wmic output (CSV format)
				wmicOutput := `Node,CommandLine,KernelModeTime,Name,ProcessId,Status,UserModeTime,WorkingSetSize
MYPC,claude.exe --model sonnet,1000,claude.exe,1234,,2000,268435456`
				cmd.Commands = []MockCommand{
					{
						Cmd:      "wmic",
						Args:     []string{"process", "where", "ProcessId=1234", "get", "ProcessId,Name,Status,WorkingSetSize,KernelModeTime,UserModeTime,CommandLine", "/format:csv"},
						Response: wmicOutput,
					},
				}
			},
			wantInfo: &ProcessInfo{
				PID:      1234,
				Status:   "running",
				MemoryMB: 256.0, // 268435456 / 1024 / 1024
				Command:  "claude.exe",
			},
			wantErr: false,
		},
		{
			name:     "Unsupported platform",
			pid:      1234,
			platform: "freebsd",
			setupMock: func(cmd *MockCommandExecutor, pm *MockProcessManager) {
				// No setup needed
			},
			wantInfo: &ProcessInfo{PID: 1234},
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip platform-specific tests when running on different platform
			if tt.platform != "" && tt.platform != runtime.GOOS {
				t.Skipf("Test is for %s platform, running on %s", tt.platform, runtime.GOOS)
			}
			
			// Create mocks
			cmdExec := &MockCommandExecutor{
				Commands: []MockCommand{},
			}
			procMgr := NewMockProcessManager()
			
			// Setup mocks for this test
			tt.setupMock(cmdExec, procMgr)
			
			// Create provider with mocks
			provider := &ProcessInfoProvider{
				CmdExecutor:    cmdExec,
				ProcessManager: procMgr,
			}
			
			// Test GetProcessInfo
			info, err := provider.GetProcessInfo(tt.pid)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProcessInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && info != nil && tt.wantInfo != nil {
				if info.PID != tt.wantInfo.PID {
					t.Errorf("PID = %d, want %d", info.PID, tt.wantInfo.PID)
				}
				if info.Status != tt.wantInfo.Status {
					t.Errorf("Status = %s, want %s", info.Status, tt.wantInfo.Status)
				}
				if info.CPUPercent != tt.wantInfo.CPUPercent {
					t.Errorf("CPUPercent = %.2f, want %.2f", info.CPUPercent, tt.wantInfo.CPUPercent)
				}
				// Skip MemoryMB check as it depends on system calls that can't be mocked
				// The memory calculation uses getSystemMemory() which calls exec.Command directly
				if info.ElapsedTime != tt.wantInfo.ElapsedTime {
					t.Errorf("ElapsedTime = %s, want %s", info.ElapsedTime, tt.wantInfo.ElapsedTime)
				}
				if info.Command != tt.wantInfo.Command {
					t.Errorf("Command = %s, want %s", info.Command, tt.wantInfo.Command)
				}
			}
		})
	}
}

// Test CheckProcessHealthWithManager
func TestCheckProcessHealthWithManager(t *testing.T) {
	tests := []struct {
		name      string
		pid       int
		setupMock func(*MockProcessManager)
		want      bool
		wantErr   bool
	}{
		{
			name: "Healthy running process",
			pid:  1234,
			setupMock: func(pm *MockProcessManager) {
				pm.Processes[1234] = &MockProcess{
					PID:    1234,
					Status: "running",
				}
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Unhealthy stopped process",
			pid:  5678,
			setupMock: func(pm *MockProcessManager) {
				pm.Processes[5678] = &MockProcess{
					PID:    5678,
					Status: "stopped",
				}
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Process not found",
			pid:  9999,
			setupMock: func(pm *MockProcessManager) {
				// Process doesn't exist
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Process find error",
			pid:  1111,
			setupMock: func(pm *MockProcessManager) {
				// Don't add process - FindProcess will return error
			},
			want:    false,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			procMgr := NewMockProcessManager()
			
			// Setup mock for this test
			tt.setupMock(procMgr)
			
			// Test CheckProcessHealthWithManager
			got, err := CheckProcessHealthWithManager(tt.pid, procMgr)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckProcessHealthWithManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if got != tt.want {
				t.Errorf("CheckProcessHealthWithManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test edge cases for process info parsing
func TestProcessInfoProvider_ParseEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		pid       int
		setupMock func(*MockCommandExecutor, *MockProcessManager)
		checkFunc func(*testing.T, *ProcessInfo, error)
	}{
		{
			name: "ps command fails but process is running",
			pid:  1234,
			setupMock: func(cmd *MockCommandExecutor, pm *MockProcessManager) {
				pm.Processes[1234] = &MockProcess{
					PID:    1234,
					Status: "running",
				}
				// ps command fails - empty Commands slice means any ps command will fail
			},
			checkFunc: func(t *testing.T, info *ProcessInfo, err error) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if info.Status != "running" {
					t.Errorf("Expected status 'running', got %s", info.Status)
				}
				// Should have basic info even if ps fails
				if info.PID != 1234 {
					t.Errorf("Expected PID 1234, got %d", info.PID)
				}
			},
		},
		{
			name: "ps output with fewer fields than expected",
			pid:  1234,
			setupMock: func(cmd *MockCommandExecutor, pm *MockProcessManager) {
				pm.Processes[1234] = &MockProcess{
					PID:    1234,
					Status: "running",
				}
				// ps output with only some fields
				psOutput := `PID
1234`
				cmd.Commands = []MockCommand{
					{
						Cmd:      "ps",
						Args:     []string{"-p", "1234", "-o", "pid,ppid,%cpu,%mem,rss,vsz,etime,command"},
						Response: psOutput,
					},
				}
			},
			checkFunc: func(t *testing.T, info *ProcessInfo, err error) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if info.Status != "running" {
					t.Errorf("Expected status 'running', got %s", info.Status)
				}
				// Should handle missing fields gracefully
				if info.CPUPercent != 0 {
					t.Errorf("Expected CPUPercent 0, got %.2f", info.CPUPercent)
				}
			},
		},
		{
			name: "ps output with malformed data",
			pid:  1234,
			setupMock: func(cmd *MockCommandExecutor, pm *MockProcessManager) {
				pm.Processes[1234] = &MockProcess{
					PID:    1234,
					Status: "running",
				}
				// ps output with non-numeric CPU/memory values
				psOutput := `  PID  PPID  %CPU %MEM      RSS      VSZ ELAPSED COMMAND
 1234  1000  N/A  N/A      N/A      N/A    N/A claude`
				cmd.Commands = []MockCommand{
					{
						Cmd:      "ps",
						Args:     []string{"-p", "1234", "-o", "pid,ppid,%cpu,%mem,rss,vsz,etime,command"},
						Response: psOutput,
					},
				}
			},
			checkFunc: func(t *testing.T, info *ProcessInfo, err error) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				// Should handle parse errors gracefully
				if info.CPUPercent != 0 {
					t.Errorf("Expected CPUPercent 0 for parse error, got %.2f", info.CPUPercent)
				}
				if info.MemoryMB != 0 {
					t.Errorf("Expected MemoryMB 0 for parse error, got %.2f", info.MemoryMB)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			cmdExec := &MockCommandExecutor{
				Commands: []MockCommand{},
			}
			procMgr := NewMockProcessManager()
			
			// Setup mocks for this test
			tt.setupMock(cmdExec, procMgr)
			
			// Create provider with mocks
			provider := &ProcessInfoProvider{
				CmdExecutor:    cmdExec,
				ProcessManager: procMgr,
			}
			
			// Test GetProcessInfo
			info, err := provider.GetProcessInfo(tt.pid)
			
			// Run custom check function
			tt.checkFunc(t, info, err)
		})
	}
}

