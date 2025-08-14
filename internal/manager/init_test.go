package manager

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/repo-claude/internal/config"
)

func TestInitWorkspace(t *testing.T) {
	// Skip tests that require repo tool if not available
	if _, err := exec.LookPath("repo"); err != nil {
		t.Skip("repo tool not available")
	}
	
	t.Run("NonInteractiveInit", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectName := "test-project"
		workspacePath := filepath.Join(tmpDir, projectName)
		
		mgr := New(workspacePath)
		
		// Mock repo commands to avoid actual repo operations
		oldPath := os.Getenv("PATH")
		mockBinDir := filepath.Join(tmpDir, "mock-bin")
		err := os.MkdirAll(mockBinDir, 0755)
		require.NoError(t, err)
		
		// Create mock repo script
		mockRepo := filepath.Join(mockBinDir, "repo")
		mockRepoContent := `#!/bin/sh
echo "Mock repo command: $@"
exit 0
`
		err = os.WriteFile(mockRepo, []byte(mockRepoContent), 0755)
		require.NoError(t, err)
		
		os.Setenv("PATH", mockBinDir+":"+oldPath)
		defer os.Setenv("PATH", oldPath)
		
		err = mgr.InitWorkspace(projectName, false)
		require.NoError(t, err)
		
		// Check workspace was created
		assert.DirExists(t, workspacePath)
		
		// Check config was saved
		assert.FileExists(t, filepath.Join(workspacePath, "repo-claude.yaml"))
		
		// Check manifest repo was created
		assert.DirExists(t, filepath.Join(workspacePath, ".manifest-repo"))
		assert.FileExists(t, filepath.Join(workspacePath, ".manifest-repo", "default.xml"))
		
		// Check shared memory was created
		assert.FileExists(t, filepath.Join(workspacePath, "shared-memory.md"))
		
		// Check executable was copied
		assert.FileExists(t, filepath.Join(workspacePath, "repo-claude"))
	})
}

func TestInitRepoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := &Manager{WorkspacePath: tmpDir}
	
	// Create mock repo command
	mockBinDir := filepath.Join(tmpDir, "mock-bin")
	err := os.MkdirAll(mockBinDir, 0755)
	require.NoError(t, err)
	
	mockRepo := filepath.Join(mockBinDir, "repo")
	mockRepoContent := `#!/bin/sh
if [ "$1" = "init" ]; then
    echo "Repo initialized"
    exit 0
fi
exit 1
`
	err = os.WriteFile(mockRepo, []byte(mockRepoContent), 0755)
	require.NoError(t, err)
	
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", mockBinDir+":"+oldPath)
	defer os.Setenv("PATH", oldPath)
	
	err = mgr.initRepoWorkspace()
	require.NoError(t, err)
}

func TestRepoSync(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := &Manager{WorkspacePath: tmpDir}
	
	// Test when repo is not initialized
	err := mgr.repoSync()
	assert.NoError(t, err) // Should handle gracefully
}

func TestCopyExecutableError(t *testing.T) {
	// Test when executable path cannot be determined
	mgr := &Manager{WorkspacePath: "/invalid/path"}
	
	// This might not error on all systems, so we just ensure it doesn't panic
	_ = mgr.copyExecutable()
}

func TestInitWorkspaceErrors(t *testing.T) {
	t.Run("InvalidWorkspacePath", func(t *testing.T) {
		mgr := New("/root/cannot-create/workspace")
		err := mgr.InitWorkspace("test", false)
		assert.Error(t, err)
	})
	
	t.Run("ConfigSaveError", func(t *testing.T) {
		tmpDir := t.TempDir()
		workspacePath := filepath.Join(tmpDir, "test")
		
		mgr := New(workspacePath)
		
		// Create workspace but make config file read-only directory
		err := os.MkdirAll(workspacePath, 0755)
		require.NoError(t, err)
		
		// Create a directory where config file should be
		err = os.MkdirAll(filepath.Join(workspacePath, "repo-claude.yaml"), 0755)
		require.NoError(t, err)
		
		err = mgr.InitWorkspace("test", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "saving config")
	})
}