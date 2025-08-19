package manager

import (
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
			name:      "Valid agent with foreground option",
			agentName: "test-agent",
			opts:      StartOptions{Foreground: true},
			setupFunc: func(m *Manager, tmpDir string) error {
				// Create repository directory with .git
				repoPath := filepath.Join(tmpDir, "test-repo")
				os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
				return nil
			},
			expectError: true, // Will fail because claude CLI doesn't exist
			errorMsg:   "failed to start agent",
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
			expectError: true, // Will fail because claude CLI doesn't exist
			errorMsg:    "failed to start agent",
		},
		{
			name:        "Multiple repositories",
			repos:       []string{"frontend", "backend"},
			opts:        StartOptions{},
			expectError: true, // Will fail because claude CLI doesn't exist
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
			expectError: true, // Will fail because repos not found
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
				"frontend-dev": {Model: "test", Specialization: "frontend", AutoStart: true},
				"backend-dev":  {Model: "test", Specialization: "backend", AutoStart: false},
				"other-dev":    {Model: "test", Specialization: "other", AutoStart: true},
			},
		},
		agents: make(map[string]*Agent),
	}

	// Create repositories
	for _, project := range mgr.Config.Workspace.Manifest.Projects {
		repoPath := filepath.Join(tmpDir, project.Name)
		os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
	}

	err := mgr.StartAllAgentsWithOptions(StartOptions{})
	
	// Should try to start auto-start agents but fail due to missing claude CLI
	if err == nil {
		// Should have attempted to start frontend-dev
		// Can't really test success without mocking the CLI
	}
}

func TestCreateNewTerminalCommand(t *testing.T) {
	tests := []struct {
		name     string
		platform string
	}{
		{
			name:     "macOS command",
			platform: "darwin",
		},
		{
			name:     "Linux command",
			platform: "linux",
		},
		{
			name:     "Windows command",
			platform: "windows",
		},
		{
			name:     "Unknown platform",
			platform: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: We can't actually change runtime.GOOS, so this is more of a
			// documentation test to show the function handles different platforms
			cmd := createNewTerminalCommand("test-agent", "/path/to/repo", "claude-3", "test prompt")
			
			if cmd == nil {
				t.Error("Expected command, got nil")
			}
			
			// Check that command is constructed
			if len(cmd.Args) == 0 {
				t.Error("Expected command arguments")
			}
		})
	}
}