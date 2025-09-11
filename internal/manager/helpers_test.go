package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
)

func TestLoadFromCurrentDir(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (cleanup func())
		wantErr bool
		check   func(t *testing.T, m *Manager)
	}{
		{
			name: "load from workspace root",
			setup: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				wsDir := filepath.Join(tmpDir, "workspace")
				require.NoError(t, os.MkdirAll(wsDir, 0755))
				
				// Create muno.yaml
				cfg := &config.ConfigTree{
					Workspace: config.WorkspaceTree{
						Name: "test-workspace",
					},
					Nodes: []config.NodeDefinition{},
				}
				configPath := filepath.Join(wsDir, "muno.yaml")
				require.NoError(t, cfg.Save(configPath))
				
				// Save current directory and change to workspace
				oldWd, _ := os.Getwd()
				os.Chdir(wsDir)
				
				return func() {
					os.Chdir(oldWd)
				}
			},
			wantErr: false,
			check: func(t *testing.T, m *Manager) {
				assert.NotNil(t, m)
			},
		},
		{
			name: "load from subdirectory",
			setup: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				wsDir := filepath.Join(tmpDir, "workspace")
				subDir := filepath.Join(wsDir, "sub", "dir")
				require.NoError(t, os.MkdirAll(subDir, 0755))
				
				// Create muno.yaml at workspace root
				cfg := &config.ConfigTree{
					Workspace: config.WorkspaceTree{
						Name: "test-workspace",
					},
					Nodes: []config.NodeDefinition{},
				}
				configPath := filepath.Join(wsDir, "muno.yaml")
				require.NoError(t, cfg.Save(configPath))
				
				// Change to subdirectory
				oldWd, _ := os.Getwd()
				os.Chdir(subDir)
				
				return func() {
					os.Chdir(oldWd)
				}
			},
			wantErr: false,
			check: func(t *testing.T, m *Manager) {
				assert.NotNil(t, m)
			},
		},
		{
			name: "no workspace found",
			setup: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				os.Chdir(tmpDir)
				
				return func() {
					os.Chdir(oldWd)
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()
			
			m, err := LoadFromCurrentDir()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, m)
			}
		})
	}
}

func TestNewManagerForInit(t *testing.T) {
	tests := []struct {
		name      string
		workspace string
		setup     func(t *testing.T) (cleanup func())
		wantErr   bool
		check     func(t *testing.T, m *Manager)
	}{
		{
			name:      "create new workspace",
			workspace: "test-workspace",
			setup: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				os.Chdir(tmpDir)
				
				return func() {
					os.Chdir(oldWd)
				}
			},
			wantErr: false,
			check: func(t *testing.T, m *Manager) {
				assert.NotNil(t, m)
				// Check workspace directory was created
				assert.DirExists(t, "test-workspace")
			},
		},
		{
			name:      "empty workspace name",
			workspace: "",
			setup: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				os.Chdir(tmpDir)
				
				return func() {
					os.Chdir(oldWd)
				}
			},
			wantErr: true,
		},
		{
			name:      "workspace already exists",
			workspace: "existing",
			setup: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				os.Chdir(tmpDir)
				
				// Create existing workspace
				require.NoError(t, os.MkdirAll("existing", 0755))
				cfg := &config.ConfigTree{
					Workspace: config.WorkspaceTree{
						Name: "existing",
					},
				}
				require.NoError(t, cfg.Save(filepath.Join("existing", "muno.yaml")))
				
				return func() {
					os.Chdir(oldWd)
				}
			},
			wantErr: false, // Should succeed even if workspace exists
			check: func(t *testing.T, m *Manager) {
				assert.NotNil(t, m)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()
			
			m, err := NewManagerForInit(tt.workspace)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, m)
			}
		})
	}
}

func TestFindWorkspaceRoot(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (startDir string, cleanup func())
		want    string
		wantErr bool
	}{
		{
			name: "find from workspace root",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				wsDir := filepath.Join(tmpDir, "workspace")
				require.NoError(t, os.MkdirAll(wsDir, 0755))
				
				// Create muno.yaml
				cfg := &config.ConfigTree{
					Workspace: config.WorkspaceTree{Name: "test"},
				}
				require.NoError(t, cfg.Save(filepath.Join(wsDir, "muno.yaml")))
				
				return wsDir, func() {}
			},
			wantErr: false,
		},
		{
			name: "find from subdirectory",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				wsDir := filepath.Join(tmpDir, "workspace")
				subDir := filepath.Join(wsDir, "sub", "deep", "dir")
				require.NoError(t, os.MkdirAll(subDir, 0755))
				
				// Create muno.yaml at workspace root
				cfg := &config.ConfigTree{
					Workspace: config.WorkspaceTree{Name: "test"},
				}
				require.NoError(t, cfg.Save(filepath.Join(wsDir, "muno.yaml")))
				
				return subDir, func() {}
			},
			wantErr: false,
		},
		{
			name: "no workspace found",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				return tmpDir, func() {}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDir, cleanup := tt.setup(t)
			defer cleanup()
			
			got := findWorkspaceRoot(startDir)
			if tt.wantErr {
				assert.Empty(t, got)
				return
			}
			
			assert.NotEmpty(t, got)
			// Should have muno.yaml
			assert.FileExists(t, filepath.Join(got, "muno.yaml"))
		})
	}
}