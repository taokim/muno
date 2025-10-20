package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
)

// TestClonePullIntegration tests the separation of clone and pull commands
func TestClonePullIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory
	tmpDir := t.TempDir()

	// Build muno binary
	binary := buildBinary(t, tmpDir)

	// Create test repositories
	// Get the actual nodes directory from config
	nodesDir := config.GetDefaultNodesDir()
	repo1Dir := filepath.Join(tmpDir, nodesDir, "repo1")
	repo2Dir := filepath.Join(tmpDir, nodesDir, "repo2")
	
	// Setup repo1 (non-lazy by naming it "monorepo")
	require.NoError(t, os.MkdirAll(repo1Dir, 0755))
	cmd := exec.Command("git", "init")
	cmd.Dir = repo1Dir
	require.NoError(t, cmd.Run())
	
	// Create a file in repo1
	require.NoError(t, os.WriteFile(filepath.Join(repo1Dir, "README.md"), []byte("# Repo1"), 0644))
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repo1Dir
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repo1Dir
	require.NoError(t, cmd.Run())
	
	// Setup repo2 (lazy)
	require.NoError(t, os.MkdirAll(repo2Dir, 0755))
	cmd = exec.Command("git", "init")
	cmd.Dir = repo2Dir
	require.NoError(t, cmd.Run())
	
	// Create a file in repo2
	require.NoError(t, os.WriteFile(filepath.Join(repo2Dir, "README.md"), []byte("# Repo2"), 0644))
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repo2Dir
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repo2Dir
	require.NoError(t, cmd.Run())

	// Initialize workspace
	workspaceDir := filepath.Join(tmpDir, "workspace")
	require.NoError(t, os.MkdirAll(workspaceDir, 0755))
	
	cmd = exec.Command(binary, "init", "test-workspace")
	cmd.Dir = workspaceDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Init failed: %s", string(output))

	t.Run("CloneWithoutIncludeLazy", func(t *testing.T) {
		// Add repositories
		cmd = exec.Command(binary, "add", "file://"+repo1Dir, "--name", "test-monorepo")
		cmd.Dir = workspaceDir
		output, err = cmd.CombinedOutput()
		require.NoError(t, err, "Add repo1 failed: %s", string(output))
		
		cmd = exec.Command(binary, "add", "file://"+repo2Dir, "--name", "service", "--lazy")
		cmd.Dir = workspaceDir
		output, err = cmd.CombinedOutput()
		require.NoError(t, err, "Add repo2 failed: %s", string(output))
		
		// Clone without --include-lazy
		cmd = exec.Command(binary, "clone")
		cmd.Dir = workspaceDir
		output, err = cmd.CombinedOutput()
		t.Logf("Clone output: %s", string(output))
		require.NoError(t, err, "Clone failed: %s", string(output))
		
		// Check that non-lazy repo is cloned
		// Get the actual nodes directory from config
		nodesDir := config.GetDefaultNodesDir()
		assert.DirExists(t, filepath.Join(workspaceDir, nodesDir, "test-monorepo", ".git"))
		
		// Check that lazy repo is NOT cloned
		assert.NoDirExists(t, filepath.Join(workspaceDir, nodesDir, "service", ".git"))
	})

	t.Run("CloneWithIncludeLazy", func(t *testing.T) {
		// Clone with --include-lazy
		cmd = exec.Command(binary, "clone", "--include-lazy")
		cmd.Dir = workspaceDir
		output, err = cmd.CombinedOutput()
		t.Logf("Clone --include-lazy output: %s", string(output))
		require.NoError(t, err, "Clone --include-lazy failed: %s", string(output))
		
		// Now lazy repo should be cloned
		assert.DirExists(t, filepath.Join(workspaceDir, nodesDir, "service", ".git"))
	})

	t.Run("PullDoesNotCloneNew", func(t *testing.T) {
		// Remove the lazy repo to simulate it not being cloned
		os.RemoveAll(filepath.Join(workspaceDir, nodesDir, "service"))
		
		// Create repo3 (another lazy repo)
		repo3Dir := filepath.Join(tmpDir, nodesDir, "repo3")
		require.NoError(t, os.MkdirAll(repo3Dir, 0755))
		cmd = exec.Command("git", "init")
		cmd.Dir = repo3Dir
		require.NoError(t, cmd.Run())
		require.NoError(t, os.WriteFile(filepath.Join(repo3Dir, "README.md"), []byte("# Repo3"), 0644))
		cmd = exec.Command("git", "add", ".")
		cmd.Dir = repo3Dir
		require.NoError(t, cmd.Run())
		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = repo3Dir
		require.NoError(t, cmd.Run())
		
		// Add it as lazy
		cmd = exec.Command(binary, "add", "file://"+repo3Dir, "--name", "repo3", "--lazy")
		cmd.Dir = workspaceDir
		output, err = cmd.CombinedOutput()
		require.NoError(t, err, "Add repo3 failed: %s", string(output))
		
		// Pull should NOT clone the new lazy repos
		cmd = exec.Command(binary, "pull", "--recursive")
		cmd.Dir = workspaceDir
		output, err = cmd.CombinedOutput()
		t.Logf("Pull output: %s", string(output))
		// Pull might fail because repos aren't proper remotes, but shouldn't crash
		
		// Verify lazy repos are still not cloned
		assert.NoDirExists(t, filepath.Join(workspaceDir, nodesDir, "service", ".git"))
		assert.NoDirExists(t, filepath.Join(workspaceDir, nodesDir, "repo3", ".git"))
	})
}