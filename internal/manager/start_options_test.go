package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/taokim/repo-claude/internal/config"
)

func TestStartAgentWithOptions(t *testing.T) {
	tests := []struct {
		name        string
		agentName   string
		opts        StartOptions
		setupFunc   func(*Manager, string) error
		expectError bool
		errorMsg    string
	}{
		{
			name:      "Agent not found",
			agentName: "non-existent",
			opts:      StartOptions{},
			expectError: true,
			errorMsg:   "agent non-existent not found in configuration",
		},
		{
			name:      "Agent already running",
			agentName: "test-agent",
			opts:      StartOptions{},
			setupFunc: func(m *Manager, tmpDir string) error {
				// Simulate running agent
				m.agents["test-agent"] = &Agent{
					Name:    "test-agent",
					Process: &os.Process{Pid: 12345},
					Status:  "running",
				}
				return nil
			},
			expectError: true,
			errorMsg:   "agent test-agent is already running",
		},
		{
			name:      "No repository assigned",
			agentName: "orphan-agent",
			opts:      StartOptions{},
			setupFunc: func(m *Manager, tmpDir string) error {
				// Add agent without repository assignment
				m.Config.Agents["orphan-agent"] = config.Agent{
					Model:          "test-model",
					Specialization: "test",
					AutoStart:      false,
				}
				return nil
			},
			expectError: true,
			errorMsg:   "no repository assigned to agent orphan-agent",
		},
		{
			name:      "Repository not found",
			agentName: "test-agent",
			opts:      StartOptions{},
			setupFunc: func(m *Manager, tmpDir string) error {
				// Don't create the .git directory
				return nil
			},
			expectError: true,
			errorMsg:   "repository",
		},
		{
			name:      "Valid agent - current terminal",
			agentName: "test-agent",
			opts:      StartOptions{},
			setupFunc: func(m *Manager, tmpDir string) error {
				// Create repository directory with .git
				repoPath := filepath.Join(tmpDir, "test-repo")
				os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
				
				// Use mock command executor that prevents real terminal operations
				m.CmdExecutor = &MockCommandExecutor{
					Commands: []MockCommand{
						{
							Cmd:   "osascript", // On macOS, it uses osascript
							Error: fmt.Errorf("failed to open terminal"),
						},
						{
							Cmd:   "claude",
							Error: fmt.Errorf("claude command not found"),
						},
						{
							Cmd:   "xterm", // Linux fallback
							Error: fmt.Errorf("xterm not found"),
						},
					},
				}
				return nil
			},
			expectError: true, // Will fail because claude CLI doesn't exist
			errorMsg:   "failed to start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			// Create test manager
			mgr := &Manager{
				WorkspacePath: tmpDir,
				Config: &config.Config{
					Workspace: config.WorkspaceConfig{
						Name: "test",
						Manifest: config.Manifest{
							Projects: []config.Project{
								{
									Name:  "test-repo",
									// URL field removed from Project struct
									Agent: "test-agent",
								},
							},
						},
					},
					Agents: map[string]config.Agent{
						"test-agent": {
							Model:          "test-model",
							Specialization: "test",
							AutoStart:      false,
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

			err := mgr.StartAgentWithOptions(tt.agentName, tt.opts)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestStartByRepos(t *testing.T) {
	tests := []struct {
		name        string
		repos       []string
		opts        StartOptions
		expectError bool
		errorMsg    string
	}{
		{
			name:        "No agents assigned",
			repos:       []string{"repo-without-agent"},
			opts:        StartOptions{},
			expectError: true,
			errorMsg:    "no agents assigned to repositories",
		},
		{
			name:        "Valid repository",
			repos:       []string{"frontend"},
			opts:        StartOptions{},
			expectError: false, // StartByRepos doesn't return error for individual failures
			errorMsg:    "",
		},
		{
			name:        "Multiple repositories",
			repos:       []string{"frontend", "backend"},
			opts:        StartOptions{},
			expectError: false, // StartByRepos doesn't return error for individual failures
			errorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			mgr := &Manager{
				WorkspacePath: tmpDir,
				Config: &config.Config{
					Workspace: config.WorkspaceConfig{
						Name: "test",
						Manifest: config.Manifest{
							Projects: []config.Project{
								{Name: "frontend", Agent: "frontend-dev"},
								{Name: "backend", Agent: "backend-dev"},
								{Name: "repo-without-agent"},
							},
						},
					},
					Agents: map[string]config.Agent{
						"frontend-dev": {Model: "test", Specialization: "frontend"},
						"backend-dev":  {Model: "test", Specialization: "backend"},
					},
				},
				agents: make(map[string]*Agent),
			}

			// Create .git directories
			for _, project := range mgr.Config.Workspace.Manifest.Projects {
				repoPath := filepath.Join(tmpDir, project.Name)
				os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
			}
			
			// Use mock command executor for tests that expect errors
			if tt.expectError && tt.errorMsg == "failed to start" {
				mgr.CmdExecutor = &MockCommandExecutor{
					Commands: []MockCommand{
						{
							Cmd:   "claude",
							Error: fmt.Errorf("claude command not found"),
						},
					},
				}
			}

			err := mgr.StartByRepos(tt.repos, tt.opts)
			
			if tt.expectError && tt.errorMsg != "" {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestStartPreset(t *testing.T) {
	tests := []struct {
		name        string
		preset      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Unknown preset",
			preset:      "unknown",
			expectError: true,
			errorMsg:    "preset 'unknown' not found",
		},
		{
			name:        "Valid preset",
			preset:      "fullstack",
			expectError: false, // Won't error, but won't start anything either
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			mgr := &Manager{
				WorkspacePath: tmpDir,
				Config: &config.Config{
					Workspace: config.WorkspaceConfig{
						Name: "test",
						Manifest: config.Manifest{
							Projects: []config.Project{
								{Name: "frontend", Agent: "frontend-dev"},
								{Name: "backend", Agent: "backend-dev"},
							},
						},
					},
					Agents: map[string]config.Agent{
						"frontend-dev": {Model: "test", Specialization: "frontend"},
						"backend-dev":  {Model: "test", Specialization: "backend"},
					},
				},
				agents: make(map[string]*Agent),
			}

			err := mgr.StartPreset(tt.preset, StartOptions{})
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestStartAllAgentsWithOptions(t *testing.T) {
	tests := []struct {
		name              string
		autoStartCount    int
		initialNewWindow  bool
		expectedNewWindow bool
	}{
		{
			name:              "Single auto-start agent - current terminal",
			autoStartCount:    1,
			initialNewWindow:  false,
			expectedNewWindow: false,
		},
		{
			name:              "Multiple auto-start agents - auto new windows",
			autoStartCount:    2,
			initialNewWindow:  false,
			expectedNewWindow: true, // Should auto-enable new windows
		},
		{
			name:              "Multiple agents with explicit new window",
			autoStartCount:    2,
			initialNewWindow:  true,
			expectedNewWindow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			agents := map[string]config.Agent{}
			projects := []config.Project{}
			
			// Create the specified number of auto-start agents
			for i := 0; i < tt.autoStartCount; i++ {
				agentName := fmt.Sprintf("agent-%d", i)
				repoName := fmt.Sprintf("repo-%d", i)
				agents[agentName] = config.Agent{
					Model:          "test",
					Specialization: "test",
					AutoStart:      true,
				}
				projects = append(projects, config.Project{
					Name:  repoName,
					Agent: agentName,
				})
			}
			
			// Add one non-auto-start agent
			agents["manual-agent"] = config.Agent{
				Model:          "test",
				Specialization: "test",
				AutoStart:      false,
			}
			
			mgr := &Manager{
				WorkspacePath: tmpDir,
				Config: &config.Config{
					Workspace: config.WorkspaceConfig{
						Name: "test",
						Manifest: config.Manifest{
							Projects: projects,
						},
					},
					Agents: agents,
				},
				agents: make(map[string]*Agent),
			}

			// Create repositories
			for _, project := range mgr.Config.Workspace.Manifest.Projects {
				repoPath := filepath.Join(tmpDir, project.Name)
				os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
			}

			// Track command calls
			var capturedNewWindow bool
			// var claudeCalled bool
			mgr.CmdExecutor = &MockCommandExecutor{
				Commands: []MockCommand{
					{
						Cmd:   "osascript",
						OnCall: func() {
							capturedNewWindow = true
						},
						Error: fmt.Errorf("mock osascript error"),
					},
					{
						Cmd:   "claude",
						OnCall: func() {
							// claudeCalled = true
						},
						Error: fmt.Errorf("mock claude error"),
					},
				},
			}

			opts := StartOptions{NewWindow: tt.initialNewWindow}
			err := mgr.StartAllAgentsWithOptions(opts)
			
			// We expect errors since claude CLI doesn't exist, but we can check the behavior
			if err == nil && tt.autoStartCount > 0 {
				t.Error("Expected error when starting agents without claude CLI")
			}
			
			// Check if new window was used as expected
			if tt.expectedNewWindow && !capturedNewWindow {
				t.Error("Expected new window to be used but it wasn't")
			}
			if !tt.expectedNewWindow && capturedNewWindow {
				t.Error("Did not expect new window but it was used")
			}
		})
	}
}

func TestCreateNewTerminalCommand(t *testing.T) {
	tests := []struct {
		name            string
		platform        string
		newWindow       bool
		expectOsascript bool // For macOS window creation
	}{
		{
			name:            "current terminal",
			platform:        "darwin",
			newWindow:       false,
			expectOsascript: false,
		},
		{
			name:            "new window on macOS",
			platform:        "darwin",
			newWindow:       true,
			expectOsascript: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := map[string]string{
				"RC_AGENT_NAME": "test-agent",
				"RC_WORKSPACE_ROOT": "/test/workspace",
			}
			cmd := createNewTerminalCommand("test-agent", "/path/to/repo", "claude-3", "test prompt", envVars, tt.newWindow)
			
			if cmd == nil {
				t.Error("Expected command, got nil")
				return
			}
			
			// Check that command is constructed correctly
			if len(cmd.Args) == 0 {
				t.Error("Expected command arguments")
				return
			}
			
			// For current terminal, should use claude directly
			if !tt.newWindow && !strings.Contains(cmd.Path, "claude") {
				t.Errorf("Expected claude command for current terminal, got %s", cmd.Path)
			}
			
			// For new window on macOS, should use osascript
			if tt.expectOsascript && !strings.Contains(cmd.Path, "osascript") {
				t.Errorf("Expected osascript for new window on macOS, got %s", cmd.Path)
			}
		})
	}
}