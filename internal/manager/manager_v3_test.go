package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewV3(t *testing.T) {
	tmpDir := t.TempDir()
	
	mgr, err := NewV3(tmpDir)
	require.NoError(t, err)
	assert.NotNil(t, mgr)
	assert.Equal(t, tmpDir, mgr.ProjectPath)
	assert.NotNil(t, mgr.TreeManager)
	assert.NotNil(t, mgr.GitCmd)
}

func TestInitWorkspaceV3(t *testing.T) {
	tmpDir := t.TempDir()
	
	mgr, err := NewV3(tmpDir)
	require.NoError(t, err)
	
	t.Run("BasicInitialization", func(t *testing.T) {
		err := mgr.InitWorkspace("test-project", false)
		require.NoError(t, err)
		
		// Check config file was saved
		configPath := filepath.Join(tmpDir, "repo-claude.yaml")
		assert.FileExists(t, configPath)
		
		// Check shared memory was created
		sharedMemPath := filepath.Join(tmpDir, "shared-memory.md")
		assert.FileExists(t, sharedMemPath)
		
		// Check repos directory exists
		reposPath := filepath.Join(tmpDir, "repos")
		assert.DirExists(t, reposPath)
		
		// Check CLAUDE.md was created
		claudePath := filepath.Join(tmpDir, "CLAUDE.md")
		assert.FileExists(t, claudePath)
	})
	
	t.Run("AlreadyInitialized", func(t *testing.T) {
		// Try to initialize again
		err := mgr.InitWorkspace("test-project", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already initialized")
	})
}

func TestLoadFromCurrentDirV3(t *testing.T) {
	t.Run("LoadExistingProject", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Initialize project first
		mgr, err := NewV3(tmpDir)
		require.NoError(t, err)
		
		err = mgr.InitWorkspace("existing-project", false)
		require.NoError(t, err)
		
		// Change to project dir
		oldCwd, _ := os.Getwd()
		defer os.Chdir(oldCwd)
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		
		// Load manager
		mgr2, err := LoadFromCurrentDirV3()
		require.NoError(t, err)
		assert.NotNil(t, mgr2)
		assert.NotNil(t, mgr2.TreeManager)
		assert.NotNil(t, mgr2.Config)
	})
	
	t.Run("NoConfig", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		oldCwd, _ := os.Getwd()
		defer os.Chdir(oldCwd)
		err := os.Chdir(tmpDir)
		require.NoError(t, err)
		
		mgr, err := LoadFromCurrentDirV3()
		assert.Error(t, err)
		assert.Nil(t, mgr)
	})
}

func TestTreeOperationsV3(t *testing.T) {
	tmpDir := t.TempDir()
	
	mgr, err := NewV3(tmpDir)
	require.NoError(t, err)
	
	err = mgr.InitWorkspace("tree-ops-test", false)
	require.NoError(t, err)
	
	t.Run("UseNode", func(t *testing.T) {
		err := mgr.UseNode("/", false)
		require.NoError(t, err)
		
		// Current directory should change
		cwd, _ := os.Getwd()
		assert.Contains(t, cwd, "repos")
	})
	
	t.Run("AddRepo", func(t *testing.T) {
		err := mgr.AddRepo("https://github.com/test/new-repo.git", "new-repo", false)
		// Will fail to clone but should add to tree
		assert.Error(t, err)
	})
	
	t.Run("RemoveRepo", func(t *testing.T) {
		// First add a lazy repo (won't try to clone)
		err := mgr.AddRepo("https://github.com/test/remove-me.git", "remove-me", true)
		require.NoError(t, err)
		
		// Now remove it
		err = mgr.RemoveRepo("remove-me")
		require.NoError(t, err)
	})
	
	t.Run("ListNodes", func(t *testing.T) {
		err := mgr.ListNodes(false)
		// Just check it doesn't panic
		_ = err
	})
	
	t.Run("ShowTree", func(t *testing.T) {
		err := mgr.ShowTree("/", 0)
		// Just check it doesn't panic
		_ = err
	})
	
	t.Run("StatusNode", func(t *testing.T) {
		err := mgr.StatusNode("/", false)
		// Just check it doesn't panic
		_ = err
	})
	
	t.Run("CloneLazy", func(t *testing.T) {
		// Add a lazy repo
		err := mgr.AddRepo("https://github.com/test/lazy.git", "lazy-test", true)
		require.NoError(t, err)
		
		// Try to clone it
		err = mgr.CloneLazy(false)
		// Will fail with invalid URL
		assert.Error(t, err)
	})
}

func TestShowCurrentV3(t *testing.T) {
	tmpDir := t.TempDir()
	
	mgr, err := NewV3(tmpDir)
	require.NoError(t, err)
	
	err = mgr.InitWorkspace("show-test", false)
	require.NoError(t, err)
	
	t.Run("ShowCurrent", func(t *testing.T) {
		err := mgr.ShowCurrent()
		// Just check it doesn't panic
		_ = err
	})
	
	t.Run("ClearCurrent", func(t *testing.T) {
		err := mgr.ClearCurrent()
		// Just check it doesn't panic
		_ = err
	})
}

func TestStartNodeV3(t *testing.T) {
	tmpDir := t.TempDir()
	
	mgr, err := NewV3(tmpDir)
	require.NoError(t, err)
	
	err = mgr.InitWorkspace("claude-test", false)
	require.NoError(t, err)
	
	t.Run("StartWithoutNode", func(t *testing.T) {
		// This will try to start claude command which may not exist
		err := mgr.StartNode("", false)
		// Don't assert error as claude may or may not be installed
		_ = err
	})
	
	t.Run("StartWithExplicitPath", func(t *testing.T) {
		err := mgr.StartNode("/", true)
		// Don't assert error as claude may or may not be installed
		_ = err
	})
}

func TestGitOperationsV3(t *testing.T) {
	tmpDir := t.TempDir()
	
	mgr, err := NewV3(tmpDir)
	require.NoError(t, err)
	
	err = mgr.InitWorkspace("git-test", false)
	require.NoError(t, err)
	
	// Most git operations will fail without actual repos
	// but we can test they don't panic
	
	t.Run("PullNode", func(t *testing.T) {
		err := mgr.PullNode("/", false)
		// Will error but shouldn't panic
		_ = err
	})
	
	t.Run("CommitNode", func(t *testing.T) {
		err := mgr.CommitNode("/", "test commit", false)
		// Will error but shouldn't panic
		_ = err
	})
	
	t.Run("PushNode", func(t *testing.T) {
		err := mgr.PushNode("/", false)
		// Will error but shouldn't panic
		_ = err
	})
}

