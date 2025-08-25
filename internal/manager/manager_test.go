package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/repo-claude/internal/config"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	
	mgr := New(tmpDir)
	
	assert.NotNil(t, mgr)
	assert.Equal(t, tmpDir, mgr.ProjectPath)
	assert.Equal(t, filepath.Join(tmpDir, "workspace"), mgr.WorkspacePath)
}

func TestLoadFromCurrentDir(t *testing.T) {
	t.Run("NoConfigFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldCwd, _ := os.Getwd()
		defer os.Chdir(oldCwd)
		
		err := os.Chdir(tmpDir)
		require.NoError(t, err)
		
		_, err = LoadFromCurrentDir()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no repo-claude.yaml found")
	})
	
	t.Run("ValidConfig", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldCwd, _ := os.Getwd()
		defer os.Chdir(oldCwd)
		
		// Create a valid config
		cfg := config.DefaultConfig("test")
		configPath := filepath.Join(tmpDir, "repo-claude.yaml")
		err := cfg.Save(configPath)
		require.NoError(t, err)
		
		// Create empty state file
		statePath := filepath.Join(tmpDir, ".repo-claude-state.json")
		err = os.WriteFile(statePath, []byte("{}"), 0644)
		require.NoError(t, err)
		
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		
		mgr, err := LoadFromCurrentDir()
		require.NoError(t, err)
		assert.NotNil(t, mgr)
		
		// Use EvalSymlinks to handle /var vs /private/var on macOS
		expectedProject, _ := filepath.EvalSymlinks(tmpDir)
		actualProject, _ := filepath.EvalSymlinks(mgr.ProjectPath)
		assert.Equal(t, expectedProject, actualProject)
		
		expectedWorkspace, _ := filepath.EvalSymlinks(filepath.Join(tmpDir, "workspace"))
		actualWorkspace, _ := filepath.EvalSymlinks(mgr.WorkspacePath)
		assert.Equal(t, expectedWorkspace, actualWorkspace)
		
		assert.NotNil(t, mgr.Config)
		// State tracking removed - no longer checking State
	})
	
	t.Run("InvalidConfig", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldCwd, _ := os.Getwd()
		defer os.Chdir(oldCwd)
		
		// Create invalid config
		configPath := filepath.Join(tmpDir, "repo-claude.yaml")
		err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644)
		require.NoError(t, err)
		
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		
		_, err = LoadFromCurrentDir()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "loading config")
	})
}

func TestUpdateGitignore(t *testing.T) {
	t.Run("NotInGitRepo", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr := &Manager{
			ProjectPath:   tmpDir,
			WorkspacePath: filepath.Join(tmpDir, "workspace"),
		}
		
		// Should not error when not in a git repo
		err := mgr.updateGitignore()
		assert.NoError(t, err)
		
		// Should not create .gitignore
		_, err = os.Stat(filepath.Join(tmpDir, ".gitignore"))
		assert.True(t, os.IsNotExist(err))
	})
	
	t.Run("InGitRepo", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Create a fake .git directory
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)
		
		mgr := &Manager{
			ProjectPath:   tmpDir,
			WorkspacePath: filepath.Join(tmpDir, "workspace"),
		}
		
		err = mgr.updateGitignore()
		require.NoError(t, err)
		
		// Check .gitignore was created
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		
		// Check content
		assert.Contains(t, string(content), "# Repo-Claude workspace")
		assert.Contains(t, string(content), "workspace/")
		assert.Contains(t, string(content), ".repo-claude-state.json")
	})
	
	t.Run("ExistingGitignore", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Create a fake .git directory
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)
		
		// Create existing .gitignore
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		existingContent := "node_modules/\n*.log\n"
		err = os.WriteFile(gitignorePath, []byte(existingContent), 0644)
		require.NoError(t, err)
		
		mgr := &Manager{
			ProjectPath:   tmpDir,
			WorkspacePath: filepath.Join(tmpDir, "workspace"),
		}
		
		err = mgr.updateGitignore()
		require.NoError(t, err)
		
		// Check content
		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		
		// Should preserve existing content
		assert.Contains(t, string(content), "node_modules/")
		assert.Contains(t, string(content), "*.log")
		
		// Should add new content
		assert.Contains(t, string(content), "# Repo-Claude workspace")
		assert.Contains(t, string(content), "workspace/")
		assert.Contains(t, string(content), ".repo-claude-state.json")
	})
	
	t.Run("AlreadyIgnored", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Create a fake .git directory
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)
		
		// Create .gitignore with workspace already ignored
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		existingContent := "workspace/\n.repo-claude-state.json\n"
		err = os.WriteFile(gitignorePath, []byte(existingContent), 0644)
		require.NoError(t, err)
		
		mgr := &Manager{
			ProjectPath:   tmpDir,
			WorkspacePath: filepath.Join(tmpDir, "workspace"),
		}
		
		err = mgr.updateGitignore()
		require.NoError(t, err)
		
		// Check content didn't change
		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		assert.Equal(t, existingContent, string(content))
	})
}

func TestSetupCoordination(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := config.DefaultConfig("test")
	mgr := &Manager{
		ProjectPath:   filepath.Dir(tmpDir),
		WorkspacePath: tmpDir,
		Config:        cfg,
	}
	
	// Create repository directories
	for _, project := range cfg.Workspace.Manifest.Projects {
		if project.Agent != "" {
			err := os.MkdirAll(filepath.Join(tmpDir, project.Name), 0755)
			require.NoError(t, err)
		}
	}
	
	err := mgr.setupCoordination()
	require.NoError(t, err)
	
	// Check shared memory was created in workspace
	sharedMemPath := filepath.Join(tmpDir, "shared-memory.md")
	assert.FileExists(t, sharedMemPath)
	
	content, err := os.ReadFile(sharedMemPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Shared Agent Memory")
	assert.Contains(t, string(content), "rc status")
	
	// Check CLAUDE.md files were created
	for _, project := range cfg.Workspace.Manifest.Projects {
		if project.Agent != "" {
			claudePath := filepath.Join(tmpDir, project.Name, "CLAUDE.md")
			assert.FileExists(t, claudePath)
			
			content, err := os.ReadFile(claudePath)
			require.NoError(t, err)
			assert.Contains(t, string(content), project.Agent)
			assert.Contains(t, string(content), project.Name)
			assert.Contains(t, string(content), "Multi-Repository Management")
		}
	}
}

// TestGetRepoProjects removed - no longer using repo tool

func TestManagerCloneMissing(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := &Manager{
		ProjectPath:   tmpDir,
		WorkspacePath: filepath.Join(tmpDir, "workspace"),
		GitManager:    nil, // Will fail gracefully
	}
	
	// This will fail since GitManager is not initialized
	err := mgr.CloneMissing()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no git manager")
}