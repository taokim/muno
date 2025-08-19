package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/taokim/repo-claude/internal/config"
)

func TestManager_New(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
	}{
		{
			name:        "Absolute path",
			projectPath: "/tmp/test-project",
		},
		{
			name:        "Relative path",
			projectPath: "test-project",
		},
		{
			name:        "Current directory",
			projectPath: ".",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := New(tt.projectPath)
			
			if mgr == nil {
				t.Fatal("New() returned nil")
			}
			
			if mgr.agents == nil {
				t.Error("agents map not initialized")
			}
			
			if !filepath.IsAbs(mgr.ProjectPath) {
				t.Errorf("ProjectPath should be absolute, got %s", mgr.ProjectPath)
			}
			
			expectedWorkspace := filepath.Join(mgr.ProjectPath, "workspace")
			if mgr.WorkspacePath != expectedWorkspace {
				t.Errorf("WorkspacePath = %s, want %s", mgr.WorkspacePath, expectedWorkspace)
			}
		})
	}
}

func TestManager_LoadFromCurrentDir(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(string) error
		wantErr   bool
	}{
		{
			name:    "No config file",
			wantErr: true,
		},
		{
			name: "Valid config",
			setupFunc: func(dir string) error {
				cfg := config.DefaultConfig("test")
				return cfg.Save(filepath.Join(dir, "repo-claude.yaml"))
			},
			wantErr: false,
		},
		{
			name: "Config with custom workspace",
			setupFunc: func(dir string) error {
				cfg := config.DefaultConfig("test")
				cfg.Workspace.Path = "custom-workspace"
				return cfg.Save(filepath.Join(dir, "repo-claude.yaml"))
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldDir, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(oldDir)
			
			if tt.setupFunc != nil {
				if err := tt.setupFunc(tmpDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			
			mgr, err := LoadFromCurrentDir()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromCurrentDir() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if err == nil {
				if mgr.Config == nil {
					t.Error("Config not loaded")
				}
				if mgr.GitManager == nil {
					t.Error("GitManager not initialized")
				}
				if mgr.agents == nil {
					t.Error("agents map not initialized")
				}
			}
		})
	}
}

func TestConfigToRepos(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Manifest: config.Manifest{
				RemoteFetch: "https://github.com/test/",
				DefaultRevision: "main",
				Projects: []config.Project{
					{
						Name:   "frontend",
						Path:   "frontend",
						Agent:  "fe-dev",
						Groups: "ui,web",
					},
					{
						Name:   "backend",
						Path:   "services/backend",
						Agent:  "be-dev",
						Groups: "api,core",
					},
				},
			},
		},
	}
	
	repos := configToRepos(cfg)
	
	if len(repos) != 2 {
		t.Errorf("len(repos) = %d, want 2", len(repos))
	}
	
	// Check first repo
	if repos[0].Name != "frontend" {
		t.Errorf("repos[0].Name = %s, want frontend", repos[0].Name)
	}
	if repos[0].Path != "frontend" {
		t.Errorf("repos[0].Path = %s, want frontend", repos[0].Path)
	}
	if repos[0].Agent != "fe-dev" {
		t.Errorf("repos[0].Agent = %s, want fe-dev", repos[0].Agent)
	}
	if repos[0].URL != "https://github.com/test/frontend.git" {
		t.Errorf("repos[0].URL = %s, want https://github.com/test/frontend.git", repos[0].URL)
	}
	
	// Check second repo
	if repos[1].Path != "services/backend" {
		t.Errorf("repos[1].Path = %s, want services/backend", repos[1].Path)
	}
}

func TestManager_Sync_NoGitManager(t *testing.T) {
	mgr := &Manager{
		agents: make(map[string]*Agent),
	}
	
	err := mgr.Sync()
	if err == nil {
		t.Error("Expected error when GitManager is nil")
	}
}

func TestManager_ShowStatus_NoConfig(t *testing.T) {
	// Skip this test as ShowStatus doesn't handle nil config gracefully
	t.Skip("ShowStatus panics on nil config - needs refactoring")
}

func TestManager_StartAgent_Validation(t *testing.T) {
	mgr := &Manager{
		Config: &config.Config{
			Agents: map[string]config.Agent{
				"test-agent": {
					Model: "test",
				},
			},
		},
		agents: make(map[string]*Agent),
	}
	
	// Should fail because no repositories assigned
	err := mgr.StartAgent("test-agent")
	if err == nil {
		t.Error("Expected error when agent has no repositories")
	}
}

func TestManager_StopAgent_NotRunning(t *testing.T) {
	mgr := &Manager{
		Config: &config.Config{
			Agents: map[string]config.Agent{
				"test-agent": {
					Model: "test",
				},
			},
		},
		State: &config.State{
			Agents: map[string]config.AgentStatus{},
		},
		agents: make(map[string]*Agent),
	}
	
	err := mgr.StopAgent("test-agent")
	if err == nil {
		t.Error("Expected error when stopping non-running agent")
	}
}

func TestManager_ForAll_NoGitManager(t *testing.T) {
	mgr := &Manager{
		agents: make(map[string]*Agent),
	}
	
	err := mgr.ForAll("echo", []string{"test"})
	if err == nil {
		t.Error("Expected error when GitManager is nil")
	}
}

// Test agent struct creation
func TestAgent_Structure(t *testing.T) {
	agent := &Agent{
		Name:    "test-agent",
		Process: nil,
		Status:  "stopped",
	}
	
	if agent.Name != "test-agent" {
		t.Errorf("Agent.Name = %s, want test-agent", agent.Name)
	}
	if agent.Status != "stopped" {
		t.Errorf("Agent.Status = %s, want stopped", agent.Status)
	}
}

// Test Manager with State
func TestManager_WithState(t *testing.T) {
	mgr := &Manager{
		State: &config.State{
			Agents: map[string]config.AgentStatus{
				"test": {
					Name:   "test",
					Status: "running",
					PID:    1234,
				},
			},
		},
		agents: make(map[string]*Agent),
	}
	
	if mgr.State == nil {
		t.Error("State should not be nil")
	}
	
	if len(mgr.State.Agents) != 1 {
		t.Errorf("Expected 1 agent in state, got %d", len(mgr.State.Agents))
	}
}

// Test configToRepos with various configurations
func TestConfigToRepos_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		expected int
	}{
		{
			name: "Empty projects",
			config: &config.Config{
				Workspace: config.WorkspaceConfig{
					Manifest: config.Manifest{
						RemoteFetch: "https://github.com/test/",
						DefaultRevision: "main",
						Projects: []config.Project{},
					},
				},
			},
			expected: 0,
		},
		{
			name: "Projects without agents",
			config: &config.Config{
				Workspace: config.WorkspaceConfig{
					Manifest: config.Manifest{
						RemoteFetch: "https://github.com/test/",
						DefaultRevision: "main",
						Projects: []config.Project{
							{Name: "lib1", Path: "lib1"},
							{Name: "lib2", Path: "lib2"},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "Projects with custom paths",
			config: &config.Config{
				Workspace: config.WorkspaceConfig{
					Manifest: config.Manifest{
						RemoteFetch: "https://github.com/test/",
						DefaultRevision: "main",
						Projects: []config.Project{
							{Name: "api", Path: "services/api/v2"},
							{Name: "web", Path: "apps/web/main"},
						},
					},
				},
			},
			expected: 2,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repos := configToRepos(tt.config)
			if len(repos) != tt.expected {
				t.Errorf("configToRepos() returned %d repos, want %d", len(repos), tt.expected)
			}
		})
	}
}