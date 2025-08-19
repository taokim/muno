package manager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/taokim/repo-claude/internal/config"
)

func TestListAgents(t *testing.T) {
	tests := []struct {
		name        string
		opts        AgentListOptions
		setupFunc   func(*Manager, string) error
		expectError bool
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "Show running agents only (default)",
			opts: AgentListOptions{
				ShowAll:     false,
				ShowDetails: false,
				Format:      "table",
			},
			setupFunc: func(m *Manager, tmpDir string) error {
				// Set up state with one running and one stopped agent
				state := &config.State{
					Agents: map[string]config.AgentStatus{
						"frontend-dev": {
							Name:         "frontend-dev",
							Status:       "running",
							PID:          1234, // Use mocked PID that's recognized as running
							Repository:   "frontend",
							LastActivity: time.Now().Format(time.RFC3339),
						},
						"backend-dev": {
							Name:         "backend-dev",
							Status:       "stopped",
							PID:          0,
							Repository:   "backend",
							LastActivity: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
						},
					},
				}
				m.State = state
				// Save state to the correct location - m is the Manager instance in setupFunc
				statePath := filepath.Join(m.WorkspacePath, ".repo-claude-state.json")
				return state.Save(statePath)
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "frontend-dev") {
					t.Error("Expected to see frontend-dev in output")
				}
				if strings.Contains(output, "backend-dev") {
					t.Error("Should not show stopped agent by default")
				}
			},
		},
		{
			name: "Show all agents including stopped",
			opts: AgentListOptions{
				ShowAll:     true,
				ShowDetails: false,
				Format:      "table",
			},
			setupFunc: func(m *Manager, tmpDir string) error {
				state := &config.State{
					Agents: map[string]config.AgentStatus{
						"frontend-dev": {
							Name:       "frontend-dev",
							Status:     "running",
							PID:        1234, // Use mocked PID
							Repository: "frontend",
						},
						"backend-dev": {
							Name:       "backend-dev",
							Status:     "stopped",
							PID:        0,
							Repository: "backend",
						},
					},
				}
				m.State = state
				// Save state to the correct location - m is the Manager instance in setupFunc
				statePath := filepath.Join(m.WorkspacePath, ".repo-claude-state.json")
				return state.Save(statePath)
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "frontend-dev") {
					t.Error("Expected to see frontend-dev in output")
				}
				if !strings.Contains(output, "backend-dev") {
					t.Error("Expected to see backend-dev when ShowAll is true")
				}
			},
		},
		{
			name: "Show details with process info",
			opts: AgentListOptions{
				ShowAll:     false,
				ShowDetails: true,
				Format:      "table",
			},
			setupFunc: func(m *Manager, tmpDir string) error {
				state := &config.State{
					Agents: map[string]config.AgentStatus{
						"frontend-dev": {
							Name:       "frontend-dev",
							Status:     "running",
							PID:        1234, // Use mocked PID
							Repository: "frontend",
						},
					},
				}
				m.State = state
				// Save state to the correct location - m is the Manager instance in setupFunc
				statePath := filepath.Join(m.WorkspacePath, ".repo-claude-state.json")
				return state.Save(statePath)
			},
			checkOutput: func(t *testing.T, output string) {
				// Should show column headers for detailed view
				if !strings.Contains(output, "CPU%") {
					t.Error("Expected CPU% column in detailed view")
				}
				if !strings.Contains(output, "MEM(MB)") {
					t.Error("Expected MEM(MB) column in detailed view")
				}
			},
		},
		{
			name: "Simple format output",
			opts: AgentListOptions{
				ShowAll:     false,
				ShowDetails: false,
				Format:      "simple",
			},
			setupFunc: func(m *Manager, tmpDir string) error {
				state := &config.State{
					Agents: map[string]config.AgentStatus{
						"frontend-dev": {
							Name:       "frontend-dev",
							Status:     "running",
							PID:        1234, // Use mocked PID
							Repository: "frontend",
						},
					},
				}
				m.State = state
				// Save state to the correct location - m is the Manager instance in setupFunc
				statePath := filepath.Join(m.WorkspacePath, ".repo-claude-state.json")
				return state.Save(statePath)
			},
			checkOutput: func(t *testing.T, output string) {
				// Simple format should just list agent name and PID
				if !strings.Contains(output, "frontend-dev (PID: 1234)") {
					t.Error("Expected simple format output")
				}
			},
		},
		{
			name: "No agents running",
			opts: AgentListOptions{
				ShowAll:     false,
				ShowDetails: false,
				Format:      "table",
			},
			setupFunc: func(m *Manager, tmpDir string) error {
				// Empty state or all stopped
				state := &config.State{
					Agents: map[string]config.AgentStatus{
						"frontend-dev": {
							Name:   "frontend-dev",
							Status: "stopped",
						},
					},
				}
				m.State = state
				// Save state to the correct location - m is the Manager instance in setupFunc
				statePath := filepath.Join(m.WorkspacePath, ".repo-claude-state.json")
				return state.Save(statePath)
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "No agents running") {
					t.Error("Expected 'No agents running' message")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			// Create mock process manager
			mockPM := &MockProcessManager{
				Processes: map[int]*MockProcess{
					1234: {PID: 1234, Status: "running"},
					5678: {PID: 5678, Status: "stopped"},
				},
			}
			
			// Create mock command executor for process info
			mockCmd := &MockCommandExecutor{
				Commands: []MockCommand{
					{
						Cmd:  "ps",
						Args: []string{"-p", "1234", "-o", "pid,ppid,%cpu,%mem,rss,vsz,etime,command"},
						Response: `  PID  PPID  %CPU %MEM      RSS      VSZ ELAPSED COMMAND
 1234  1000  25.5  3.2     1024     2048 10:30   claude --model sonnet`,
					},
				},
			}
			
			mgr := &Manager{
				WorkspacePath:  tmpDir,
				ProcessManager: mockPM,
				CmdExecutor:    mockCmd,
				Config: &config.Config{
					Workspace: config.WorkspaceConfig{
						Name: "test-workspace",
					},
					Agents: map[string]config.Agent{
						"frontend-dev": {
							Model:          "test-model",
							Specialization: "Frontend development",
						},
						"backend-dev": {
							Model:          "test-model",
							Specialization: "Backend development",
						},
					},
				},
				agents: make(map[string]*Agent),
			}

			if tt.setupFunc != nil {
				if err := tt.setupFunc(mgr, tmpDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := mgr.ListAgents(tt.opts)

			w.Close()
			output := make([]byte, 10000)
			n, _ := r.Read(output)
			os.Stdout = oldStdout

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if tt.checkOutput != nil {
				outputStr := string(output[:n])
				// Debug output for failing test
				if t.Failed() || testing.Verbose() {
					t.Logf("Actual output:\n%s", outputStr)
				}
				tt.checkOutput(t, outputStr)
			}
		})
	}
}

func TestGetAgentsInfo(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create mock process manager with current process
	currentPID := os.Getpid()
	mockPM := &MockProcessManager{
		Processes: map[int]*MockProcess{
			currentPID: {PID: currentPID, Status: "running"},
		},
	}
	
	mgr := &Manager{
		WorkspacePath:  tmpDir,
		ProcessManager: mockPM,
		Config: &config.Config{
			Agents: map[string]config.Agent{
				"test-agent": {
					Model:          "test-model",
					Specialization: "Test specialization",
				},
			},
		},
	}

	// Create state file
	state := &config.State{
		Agents: map[string]config.AgentStatus{
			"test-agent": {
				Name:         "test-agent",
				Status:       "running",
				PID:          os.Getpid(), // Use current process
				Repository:   "test-repo",
				LastActivity: time.Now().Format(time.RFC3339),
			},
		},
	}
	state.Save(filepath.Join(tmpDir, ".repo-claude-state.json"))

	opts := AgentListOptions{ShowAll: true}
	agents, err := mgr.GetAgentsInfo(opts)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents))
	}
	
	if agents[0].Name != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got %s", agents[0].Name)
	}
	
	// Process should be detected as running since we used current PID
	if agents[0].Status != "running" {
		t.Errorf("Expected status 'running', got %s", agents[0].Status)
	}
}

func TestSortAgents(t *testing.T) {
	agents := []*AgentInfo{
		{Name: "c-agent", PID: 300, LastActivity: time.Now()},
		{Name: "a-agent", PID: 100, LastActivity: time.Now().Add(-1 * time.Hour)},
		{Name: "b-agent", PID: 200, LastActivity: time.Now().Add(-30 * time.Minute)},
	}

	tests := []struct {
		sortBy       string
		expectedFirst string
	}{
		{sortBy: "name", expectedFirst: "a-agent"},
		{sortBy: "pid", expectedFirst: "a-agent"},
		{sortBy: "time", expectedFirst: "c-agent"},
	}

	for _, tt := range tests {
		t.Run("Sort by "+tt.sortBy, func(t *testing.T) {
			// Make a copy to avoid modifying original
			testAgents := make([]*AgentInfo, len(agents))
			copy(testAgents, agents)
			
			sortAgents(testAgents, tt.sortBy)
			
			if testAgents[0].Name != tt.expectedFirst {
				t.Errorf("Expected %s first, got %s", tt.expectedFirst, testAgents[0].Name)
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	tmpDir := t.TempDir()
	
	mgr := &Manager{
		WorkspacePath: tmpDir,
		Config: &config.Config{
			Agents: map[string]config.Agent{
				"dead-agent": {
					Model:          "test-model",
					Specialization: "Test",
				},
			},
		},
	}

	// Create state with dead process
	state := &config.State{
		Agents: map[string]config.AgentStatus{
			"dead-agent": {
				Name:       "dead-agent",
				Status:     "running",
				PID:        99999, // Non-existent PID
				Repository: "test-repo",
			},
		},
	}
	state.Save(filepath.Join(tmpDir, ".repo-claude-state.json"))

	opts := AgentListOptions{ShowAll: true}
	agents, _ := mgr.GetAgentsInfo(opts)
	
	// The dead process should be detected and marked as stopped
	if len(agents) > 0 && agents[0].Status != "stopped" {
		t.Error("Expected dead process to be marked as stopped")
	}
	
	// Verify state was updated
	updatedState, _ := config.LoadState(filepath.Join(tmpDir, ".repo-claude-state.json"))
	if updatedState.Agents["dead-agent"].Status != "stopped" {
		t.Error("Expected state file to be updated with stopped status")
	}
}