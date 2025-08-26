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

func TestCreateClaudeMD(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Name: "test",
			Manifest: config.Manifest{
				Projects: []config.Project{
					{Name: "backend", Groups: "core", Agent: "backend-agent"},
					{Name: "frontend", Groups: "ui", Agent: "frontend-agent"},
					{Name: "shared", Groups: "lib"}, // No agent
				},
			},
		},
		Agents: map[string]config.Agent{
			"backend-agent": {
				Model:          "claude-sonnet-4",
				Specialization: "Backend development",
			},
			"frontend-agent": {
				Model:          "claude-sonnet-4",
				Specialization: "Frontend development",
			},
		},
	}
	
	mgr := &Manager{
		ProjectPath:   filepath.Dir(tmpDir),
		WorkspacePath: tmpDir,
		Config:        cfg,
	}
	
	t.Run("ValidProject", func(t *testing.T) {
		err := mgr.createClaudeMD(cfg.Workspace.Manifest.Projects[0])
		require.NoError(t, err)
		
		claudePath := filepath.Join(tmpDir, "backend", "CLAUDE.md")
		assert.FileExists(t, claudePath)
		
		content, err := os.ReadFile(claudePath)
		require.NoError(t, err)
		
		// Check content
		assert.Contains(t, string(content), "backend-agent")
		assert.Contains(t, string(content), "Backend development")
		assert.Contains(t, string(content), "backend")
		assert.Contains(t, string(content), "rc status")
		assert.Contains(t, string(content), "Cross-Repository Awareness")
		assert.Contains(t, string(content), "frontend")
	})
	
	t.Run("ProjectWithNoAgent", func(t *testing.T) {
		// This should not be called for projects without agents
		// but let's test it doesn't panic
		err := mgr.createClaudeMD(cfg.Workspace.Manifest.Projects[2])
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent  not found in config")
	})
	
	t.Run("NonExistentAgent", func(t *testing.T) {
		project := config.Project{
			Name:   "test",
			Groups: "test",
			Agent:  "non-existent-agent",
		}
		err := mgr.createClaudeMD(project)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent non-existent-agent not found")
	})
}

func TestSetupCoordinationEdgeCases(t *testing.T) {
	t.Run("NoAgentProjects", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		cfg := &config.Config{
			Workspace: config.WorkspaceConfig{
				Manifest: config.Manifest{
					Projects: []config.Project{
						{Name: "lib1", Groups: "lib"},
						{Name: "lib2", Groups: "lib"},
					},
				},
			},
			Agents: map[string]config.Agent{},
		}
		
		mgr := &Manager{
			WorkspacePath: tmpDir,
			Config:        cfg,
		}
		
		err := mgr.setupCoordination()
		require.NoError(t, err)
		
		// Only shared memory should be created
		assert.FileExists(t, filepath.Join(tmpDir, "shared-memory.md"))
		
		// No CLAUDE.md files should be created
		assert.NoDirExists(t, filepath.Join(tmpDir, "lib1", "CLAUDE.md"))
		assert.NoDirExists(t, filepath.Join(tmpDir, "lib2", "CLAUDE.md"))
	})
	
	t.Run("ExistingSharedMemory", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Create existing shared memory with custom content
		sharedMemPath := filepath.Join(tmpDir, "shared-memory.md")
		customContent := "# Custom Shared Memory\nDo not overwrite!"
		err := os.WriteFile(sharedMemPath, []byte(customContent), 0644)
		require.NoError(t, err)
		
		cfg := config.DefaultConfig("test")
		mgr := &Manager{
			WorkspacePath: tmpDir,
			Config:        cfg,
		}
		
		err = mgr.setupCoordination()
		require.NoError(t, err)
		
		// Check that existing shared memory was not overwritten
		content, err := os.ReadFile(sharedMemPath)
		require.NoError(t, err)
		assert.Equal(t, customContent, string(content))
	})
}

func TestCoordinationFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Name: "test-workspace",
			Manifest: config.Manifest{
				Projects: []config.Project{
					{Name: "service-a", Groups: "api,core", Agent: "api-agent"},
					{Name: "service-b", Groups: "api,core", Agent: "api-agent"},
					{Name: "web-app", Groups: "frontend", Agent: "web-agent"},
				},
			},
		},
		Agents: map[string]config.Agent{
			"api-agent": {
				Model:          "claude-opus-4",
				Specialization: "API and microservices",
			},
			"web-agent": {
				Model:          "claude-sonnet-4",
				Specialization: "Web application development",
			},
		},
	}
	
	mgr := &Manager{
		ProjectPath:   filepath.Dir(tmpDir),
		WorkspacePath: tmpDir,
		Config:        cfg,
	}
	
	// Create directories
	for _, project := range cfg.Workspace.Manifest.Projects {
		if project.Agent != "" {
			err := os.MkdirAll(filepath.Join(tmpDir, project.Name), 0755)
			require.NoError(t, err)
		}
	}
	
	err := mgr.setupCoordination()
	require.NoError(t, err)
	
	// Check service-a CLAUDE.md
	serviceAPath := filepath.Join(tmpDir, "service-a", "CLAUDE.md")
	content, err := os.ReadFile(serviceAPath)
	require.NoError(t, err)
	
	// Check all required elements
	assert.Contains(t, string(content), "api-agent - service-a")
	assert.Contains(t, string(content), "claude-opus-4")
	assert.Contains(t, string(content), "API and microservices")
	assert.Contains(t, string(content), "api,core")
	assert.Contains(t, string(content), "service-b")
	assert.Contains(t, string(content), "web-app")
	assert.Contains(t, string(content), "Multi-Repository Management")
	assert.Contains(t, string(content), "trunk-based development")
	assert.Contains(t, string(content), "shared-memory.md")
}

func TestRelativePathCalculation(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create nested project structure
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Manifest: config.Manifest{
				Projects: []config.Project{
					{Name: "services/api", Groups: "api", Agent: "api-agent"},
					{Name: "apps/web", Groups: "web", Agent: "web-agent"},
				},
			},
		},
		Agents: map[string]config.Agent{
			"api-agent": {Model: "claude-sonnet-4"},
			"web-agent": {Model: "claude-sonnet-4"},
		},
	}
	
	mgr := &Manager{
		ProjectPath:   filepath.Dir(tmpDir),
		WorkspacePath: tmpDir,
		Config:        cfg,
	}
	
	// Create nested directories
	for _, project := range cfg.Workspace.Manifest.Projects {
		if project.Agent != "" {
			err := os.MkdirAll(filepath.Join(tmpDir, project.Name), 0755)
			require.NoError(t, err)
		}
	}
	
	err := mgr.setupCoordination()
	require.NoError(t, err)
	
	// Check relative paths in nested structure
	apiPath := filepath.Join(tmpDir, "services/api", "CLAUDE.md")
	content, err := os.ReadFile(apiPath)
	require.NoError(t, err)
	
	// Should have correct relative paths
	assert.Contains(t, string(content), "../../shared-memory.md")
	assert.Contains(t, string(content), "../../apps/web")
}

func TestCreateClaudeMDWithCustomPaths(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Manifest: config.Manifest{
				Projects: []config.Project{
					{
						Name:   "service-a",
						Path:   "backend/services/a",
						Groups: "backend,api",
						Agent:  "backend-agent",
					},
					{
						Name:   "frontend",
						Path:   "apps/web",
						Groups: "frontend",
						Agent:  "frontend-agent",
					},
				},
			},
		},
		Agents: map[string]config.Agent{
			"backend-agent": {
				Model:          "claude-sonnet-4",
				Specialization: "Backend API development",
			},
			"frontend-agent": {
				Model:          "claude-sonnet-4",
				Specialization: "Frontend development",
			},
		},
	}
	
	mgr := &Manager{
		ProjectPath:   tmpDir,
		WorkspacePath: workspaceDir,
		Config:        cfg,
	}
	
	// Test creating CLAUDE.md for project with custom path
	project := cfg.Workspace.Manifest.Projects[0]
	err := mgr.createClaudeMD(project)
	require.NoError(t, err)
	
	// Check file was created at the custom path
	expectedPath := filepath.Join(workspaceDir, "backend/services/a/CLAUDE.md")
	assert.FileExists(t, expectedPath)
	
	// Check content has correct relative paths
	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err)
	
	// Should reference the correct relative path to shared memory
	assert.Contains(t, string(content), "../../../shared-memory.md")
	
	// Should reference other projects with correct relative paths
	assert.Contains(t, string(content), "../../../apps/web")
}