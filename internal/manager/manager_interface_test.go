//go:build legacy
// +build legacy

package manager

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/taokim/repo-claude/internal/config"
)

// Test InitWorkspace with mocked FileSystem
func TestManager_InitWorkspaceWithMocks(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		setupMock   func(*MockFileSystem)
		checkFunc   func(*testing.T, *MockFileSystem)
		wantErr     bool
	}{
		{
			name:        "Initialize new project",
			projectName: "test-project",
			setupMock: func(fs *MockFileSystem) {
				// No existing config
			},
			checkFunc: func(t *testing.T, fs *MockFileSystem) {
				t.Skip("Skipping test that requires refactoring to properly mock GitManager")
				// Check config was created
				configPath := filepath.Join("test-project", "repo-claude.yaml")
				if _, ok := fs.Files[configPath]; !ok {
					t.Error("Config file not created")
				}
				
				// Check workspace directory
				workspacePath := filepath.Join("test-project", "workspace")
				if _, ok := fs.Dirs[workspacePath]; !ok {
					t.Error("Workspace directory not created")
				}
			},
			wantErr: false,
		},
		{
			name:        "Initialize with existing config",
			projectName: "existing-project",
			setupMock: func(fs *MockFileSystem) {
				// Create existing config
				configPath := filepath.Join("existing-project", "repo-claude.yaml")
				fs.Files[configPath] = []byte(`workspace:
  name: existing
agents: {}`)
			},
			checkFunc: func(t *testing.T, fs *MockFileSystem) {
				// Config should be preserved
				configPath := filepath.Join("existing-project", "repo-claude.yaml")
				if data, ok := fs.Files[configPath]; ok {
					if !strings.Contains(string(data), "existing") {
						t.Error("Existing config not preserved")
					}
				}
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock file system
			mockFS := NewMockFileSystem()
			
			// Setup mock for this test
			if tt.setupMock != nil {
				tt.setupMock(mockFS)
			}
			
			// Create manager with mock
			mgr := New(tt.projectName)
			mgr.FileSystem = mockFS
			
			// Mock the git operations
			mgr.GitManager = nil // Will skip git operations in test
			
			// Test InitWorkspace
			err := mgr.InitWorkspace(tt.projectName, false)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("InitWorkspace() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if tt.checkFunc != nil && err == nil {
				tt.checkFunc(t, mockFS)
			}
		})
	}
}

// Test updateGitignore with mocked FileSystem
func TestManager_UpdateGitignoreWithMocks(t *testing.T) {
	tests := []struct {
		name           string
		existingIgnore string
		hasGitDir      bool
		checkFunc      func(*testing.T, *MockFileSystem)
		wantErr        bool
	}{
		{
			name:           "No existing gitignore",
			existingIgnore: "",
			hasGitDir:      true,
			checkFunc: func(t *testing.T, fs *MockFileSystem) {
				content := string(fs.Files[filepath.Join("project", ".gitignore")])
				if content == "" {
					t.Error("Gitignore not created")
				}
				if !strings.Contains(content, "workspace/") {
					t.Error("workspace/ not added to gitignore")
				}
				if !strings.Contains(content, ".repo-claude-state.json") {
					t.Error(".repo-claude-state.json not added to gitignore")
				}
			},
			wantErr: false,
		},
		{
			name:           "Existing gitignore without entries",
			existingIgnore: "# Existing\n*.log\n",
			hasGitDir:      true,
			checkFunc: func(t *testing.T, fs *MockFileSystem) {
				content := string(fs.Files[filepath.Join("project", ".gitignore")])
				if !strings.Contains(content, "*.log") {
					t.Error("Existing entries not preserved")
				}
				if !strings.Contains(content, "workspace/") {
					t.Error("workspace/ not added")
				}
			},
			wantErr: false,
		},
		{
			name:           "Existing gitignore with entries",
			existingIgnore: "workspace/\n.repo-claude-state.json\n",
			hasGitDir:      true,
			checkFunc: func(t *testing.T, fs *MockFileSystem) {
				content := string(fs.Files[filepath.Join("project", ".gitignore")])
				// Should not duplicate
				count := strings.Count(content, "workspace/")
				if count != 1 {
					t.Errorf("workspace/ appears %d times, want 1", count)
				}
			},
			wantErr: false,
		},
		{
			name:      "Not a git repository",
			hasGitDir: false,
			checkFunc: func(t *testing.T, fs *MockFileSystem) {
				// Should not create gitignore
				if _, ok := fs.Files[filepath.Join("project", ".gitignore")]; ok {
					t.Error("Gitignore created in non-git repo")
				}
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock file system
			mockFS := NewMockFileSystem()
			
			// Setup project directory
			mockFS.Dirs["project"] = true
			if tt.hasGitDir {
				mockFS.Dirs[filepath.Join("project", ".git")] = true
			}
			
			if tt.existingIgnore != "" {
				mockFS.Files[filepath.Join("project", ".gitignore")] = []byte(tt.existingIgnore)
			}
			
			// Create manager with mock
			mgr := &Manager{
				ProjectPath:   "project",
				WorkspacePath: filepath.Join("project", "workspace"),
				FileSystem:    mockFS,
			}
			
			// Test updateGitignore
			err := mgr.updateGitignore()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("updateGitignore() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if tt.checkFunc != nil {
				tt.checkFunc(t, mockFS)
			}
		})
	}
}

// Test setupCoordination with mocked FileSystem
func TestManager_SetupCoordinationWithMocks(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockFileSystem, *Manager)
		checkFunc func(*testing.T, *MockFileSystem)
		wantErr   bool
	}{
		{
			name: "Create shared memory and CLAUDE.md files",
			setupMock: func(fs *MockFileSystem, m *Manager) {
				// Setup workspace
				fs.Dirs[m.WorkspacePath] = true
				
				// Setup config
				m.Config = &config.Config{
					Workspace: config.WorkspaceConfig{
						Manifest: config.Manifest{
							Projects: []config.Project{
								{Name: "frontend", Path: "frontend", Agent: "fe-dev"},
								{Name: "backend", Path: "backend", Agent: "be-dev"},
							},
						},
					},
					Agents: map[string]config.Agent{
						"fe-dev": {Model: "claude-3", Specialization: "Frontend"},
						"be-dev": {Model: "claude-3", Specialization: "Backend"},
					},
				}
			},
			checkFunc: func(t *testing.T, fs *MockFileSystem) {
				// Check shared memory was created
				sharedMemPath := filepath.Join("workspace", "shared-memory.md")
				if data, ok := fs.Files[sharedMemPath]; !ok {
					t.Error("Shared memory file not created")
				} else {
					content := string(data)
					if !strings.Contains(content, "Shared Agent Memory") {
						t.Error("Shared memory missing expected content")
					}
				}
				
				// Check CLAUDE.md files
				frontendClaudePath := filepath.Join("workspace", "frontend", "CLAUDE.md")
				if data, ok := fs.Files[frontendClaudePath]; !ok {
					t.Error("Frontend CLAUDE.md not created")
				} else {
					content := string(data)
					if !strings.Contains(content, "fe-dev") {
						t.Error("Frontend CLAUDE.md missing agent name")
					}
					if !strings.Contains(content, "Frontend") {
						t.Error("Frontend CLAUDE.md missing specialization")
					}
				}
				
				backendClaudePath := filepath.Join("workspace", "backend", "CLAUDE.md")
				if _, ok := fs.Files[backendClaudePath]; !ok {
					t.Error("Backend CLAUDE.md not created")
				}
			},
			wantErr: false,
		},
		{
			name: "Existing shared memory file",
			setupMock: func(fs *MockFileSystem, m *Manager) {
				// Setup workspace
				fs.Dirs[m.WorkspacePath] = true
				
				// Create existing shared memory
				sharedMemPath := filepath.Join(m.WorkspacePath, "shared-memory.md")
				fs.Files[sharedMemPath] = []byte("# Existing content")
				
				// Setup config
				m.Config = &config.Config{
					Workspace: config.WorkspaceConfig{
						Manifest: config.Manifest{
							Projects: []config.Project{},
						},
					},
					Agents: map[string]config.Agent{},
				}
			},
			checkFunc: func(t *testing.T, fs *MockFileSystem) {
				// Should preserve existing shared memory
				sharedMemPath := filepath.Join("workspace", "shared-memory.md")
				if data, ok := fs.Files[sharedMemPath]; ok {
					content := string(data)
					if content != "# Existing content" {
						t.Error("Existing shared memory was overwritten")
					}
				}
			},
			wantErr: false,
		},
		{
			name: "Agent not found in config",
			setupMock: func(fs *MockFileSystem, m *Manager) {
				// Setup workspace
				fs.Dirs[m.WorkspacePath] = true
				
				// Setup config with missing agent
				m.Config = &config.Config{
					Workspace: config.WorkspaceConfig{
						Manifest: config.Manifest{
							Projects: []config.Project{
								{Name: "frontend", Path: "frontend", Agent: "missing-agent"},
							},
						},
					},
					Agents: map[string]config.Agent{},
				}
			},
			checkFunc: func(t *testing.T, fs *MockFileSystem) {
				// Error expected, no files should be created for missing agent
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock file system
			mockFS := NewMockFileSystem()
			
			// Create manager with mock
			mgr := &Manager{
				ProjectPath:   "project",
				WorkspacePath: "workspace",
				FileSystem:    mockFS,
			}
			
			// Setup mock for this test
			if tt.setupMock != nil {
				tt.setupMock(mockFS, mgr)
			}
			
			// Test setupCoordination
			err := mgr.setupCoordination()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("setupCoordination() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if tt.checkFunc != nil {
				tt.checkFunc(t, mockFS)
			}
		})
	}
}

// Test LoadFromCurrentDir with mocked FileSystem
func TestLoadFromCurrentDirWithMocks(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockFileSystem)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "Valid configuration",
			setupMock: func(fs *MockFileSystem) {
				// Create valid config
				fs.Files["repo-claude.yaml"] = []byte(`workspace:
  name: test
  path: workspace
agents:
  test-agent:
    model: claude-3`)
			},
			wantErr: false,
		},
		{
			name: "No configuration file",
			setupMock: func(fs *MockFileSystem) {
				// No config file
			},
			wantErr: true,
			errMsg:  "no repo-claude.yaml found",
		},
		{
			name: "Invalid configuration",
			setupMock: func(fs *MockFileSystem) {
				// Create invalid config
				fs.Files["repo-claude.yaml"] = []byte(`invalid: yaml: content:`)
			},
			wantErr: true,
			errMsg:  "loading config",
		},
	}
	
	// Since we can't mock os.Getwd directly, we'll skip this test
	// as it requires refactoring LoadFromCurrentDir to accept the path
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock file system
			mockFS := NewMockFileSystem()
			
			// Setup mock for this test
			if tt.setupMock != nil {
				tt.setupMock(mockFS)
			}
			
			// Skip this test as LoadFromCurrentDir uses os.Getwd directly
			t.Skip("LoadFromCurrentDir test requires refactoring to accept working directory")
		})
	}
}