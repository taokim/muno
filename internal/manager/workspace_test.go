//go:build legacy
// +build legacy

package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/repo-claude/internal/config"
)

func TestWorkspacePathHandling(t *testing.T) {
	tests := []struct {
		name              string
		projectPath       string
		configWorkspace   string
		expectedWorkspace string
	}{
		{
			name:              "default workspace",
			projectPath:       "/tmp/project",
			configWorkspace:   "",
			expectedWorkspace: "/tmp/project/workspace",
		},
		{
			name:              "custom relative workspace",
			projectPath:       "/tmp/project",
			configWorkspace:   "code",
			expectedWorkspace: "/tmp/project/code",
		},
		{
			name:              "nested relative workspace",
			projectPath:       "/tmp/project",
			configWorkspace:   "src/repos",
			expectedWorkspace: "/tmp/project/src/repos",
		},
		{
			name:              "absolute workspace path",
			projectPath:       "/tmp/project",
			configWorkspace:   "/opt/workspace",
			expectedWorkspace: "/opt/workspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Workspace: config.WorkspaceConfig{
					Name: "test",
					Path: tt.configWorkspace,
					Manifest: config.Manifest{
						RemoteFetch:     "https://github.com/test/",
						DefaultRevision: "main",
					},
				},
			}

			mgr := &Manager{
				ProjectPath:   tt.projectPath,
				WorkspacePath: filepath.Join(tt.projectPath, "workspace"), // Default
				Config:        cfg,
			}

			// Simulate what InitWorkspace does
			if cfg.Workspace.Path != "" {
				if filepath.IsAbs(cfg.Workspace.Path) {
					mgr.WorkspacePath = cfg.Workspace.Path
				} else {
					mgr.WorkspacePath = filepath.Join(mgr.ProjectPath, cfg.Workspace.Path)
				}
			}

			assert.Equal(t, tt.expectedWorkspace, mgr.WorkspacePath)
		})
	}
}

func TestConfigToReposWithPaths(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Manifest: config.Manifest{
				RemoteFetch:     "git@github.com:org",
				DefaultRevision: "main",
				Projects: []config.Project{
					{
						Name:     "service-a",
						Path:     "backend/service-a",
						Groups:   "backend,core",
						Agent:    "backend-agent",
						Revision: "develop",
					},
					{
						Name:   "service-b",
						Path:   "", // Should default to name
						Groups: "backend",
						Agent:  "backend-agent",
					},
					{
						Name:   "frontend-app",
						Path:   "apps/frontend",
						Groups: "frontend,web",
						Agent:  "frontend-agent",
					},
				},
			},
		},
	}

	repos := configToRepos(cfg)
	require.Len(t, repos, 3)

	// Test repo with custom path
	assert.Equal(t, "service-a", repos[0].Name)
	assert.Equal(t, "backend/service-a", repos[0].Path)
	assert.Equal(t, "develop", repos[0].Branch)
	assert.Equal(t, "git@github.com:org/service-a.git", repos[0].URL)

	// Test repo with default path
	assert.Equal(t, "service-b", repos[1].Name)
	assert.Equal(t, "service-b", repos[1].Path) // Defaults to name

	// Test another custom path
	assert.Equal(t, "frontend-app", repos[2].Name)
	assert.Equal(t, "apps/frontend", repos[2].Path)
}

func TestLoadFromCurrentDirWithWorkspacePath(t *testing.T) {
	tests := []struct {
		name          string
		workspacePath string
		expectError   bool
	}{
		{
			name:          "default workspace",
			workspacePath: "",
			expectError:   false,
		},
		{
			name:          "custom workspace",
			workspacePath: "repos",
			expectError:   false,
		},
		{
			name:          "nested workspace",
			workspacePath: "src/external/repos",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			// Create config with custom workspace
			cfg := &config.Config{
				Workspace: config.WorkspaceConfig{
					Name: "test",
					Path: tt.workspacePath,
					Manifest: config.Manifest{
						RemoteFetch:     "https://github.com/test/",
						DefaultRevision: "main",
						Projects: []config.Project{
							{Name: "test-repo", Groups: "core"},
						},
					},
				},
				Agents: map[string]config.Agent{
					"test-agent": {
						Model:          "claude-sonnet-4",
						Specialization: "testing",
					},
				},
			}

			// Save config
			configPath := filepath.Join(tmpDir, "repo-claude.yaml")
			err := cfg.Save(configPath)
			require.NoError(t, err)

			// Change to temp directory
			oldCwd, _ := os.Getwd()
			defer os.Chdir(oldCwd)
			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			// Load manager
			mgr, err := LoadFromCurrentDir()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				
				// Verify workspace path (use EvalSymlinks to handle /var vs /private/var on macOS)
				expectedWorkspace := filepath.Join(tmpDir, "workspace")
				if tt.workspacePath != "" {
					expectedWorkspace = filepath.Join(tmpDir, tt.workspacePath)
				}
				expectedWorkspace, _ = filepath.EvalSymlinks(expectedWorkspace)
				actualWorkspace, _ := filepath.EvalSymlinks(mgr.WorkspacePath)
				assert.Equal(t, expectedWorkspace, actualWorkspace)
			}
		})
	}
}