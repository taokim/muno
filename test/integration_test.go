package test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationWorkflow tests the complete workflow
func TestIntegrationWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Check required tools
	checkTool(t, "git")
	checkTool(t, "repo")
	
	// Create temporary directory
	tmpDir := t.TempDir()
	
	// Build repo-claude binary
	binary := buildBinary(t, tmpDir)
	
	// Test init command
	t.Run("Init", func(t *testing.T) {
		projectName := "test-project"
		projectDir := filepath.Join(tmpDir, projectName)
		
		cmd := exec.Command(binary, "init", projectName, "--non-interactive")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		
		require.NoError(t, err, "Init failed: %s", string(output))
		assert.Contains(t, string(output), "Initializing Repo-Claude workspace")
		assert.Contains(t, string(output), "Workspace initialized")
		
		// Check created files
		assert.FileExists(t, filepath.Join(projectDir, "repo-claude.yaml"))
		assert.FileExists(t, filepath.Join(projectDir, "repo-claude"))
		assert.FileExists(t, filepath.Join(projectDir, "shared-memory.md"))
		assert.DirExists(t, filepath.Join(projectDir, ".manifest-repo"))
		
		// Check executable was copied and is executable
		copiedBinary := filepath.Join(projectDir, "repo-claude")
		info, err := os.Stat(copiedBinary)
		require.NoError(t, err)
		assert.True(t, info.Mode()&0111 != 0, "Copied binary should be executable")
	})
	
	// Test status command
	t.Run("Status", func(t *testing.T) {
		projectDir := filepath.Join(tmpDir, "test-project")
		
		cmd := exec.Command(filepath.Join(projectDir, "repo-claude"), "status")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		
		require.NoError(t, err, "Status failed: %s", string(output))
		assert.Contains(t, string(output), "REPO-CLAUDE STATUS")
		assert.Contains(t, string(output), "Workspace: test-project")
		assert.Contains(t, string(output), "Repo Projects")
	})
	
	// Test sync command
	t.Run("Sync", func(t *testing.T) {
		projectDir := filepath.Join(tmpDir, "test-project")
		
		// Create dummy repositories for testing
		createDummyRepos(t, projectDir)
		
		cmd := exec.Command(filepath.Join(projectDir, "repo-claude"), "sync")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		
		// Sync might fail if repos don't exist, but command should run
		_ = err // Ignore error since repos don't exist
		assert.Contains(t, string(output), "Syncing all repositories")
	})
}

// checkTool verifies a tool is available
func checkTool(t *testing.T, tool string) {
	_, err := exec.LookPath(tool)
	if err != nil {
		t.Skipf("%s not found in PATH", tool)
	}
}

// buildBinary builds the repo-claude binary
func buildBinary(t *testing.T, tmpDir string) string {
	binary := filepath.Join(tmpDir, "repo-claude")
	
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/repo-claude")
	cmd.Dir = filepath.Join("..", "..", "repo-claude-go") // Adjust path as needed
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	err := cmd.Run()
	require.NoError(t, err, "Build failed: %s", out.String())
	
	return binary
}

// createDummyRepos creates dummy git repositories for testing
func createDummyRepos(t *testing.T, workspaceDir string) {
	repos := []string{"backend", "frontend", "mobile", "shared-libs"}
	
	for _, repo := range repos {
		repoPath := filepath.Join(workspaceDir, repo)
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			continue
		}
		
		// Initialize as git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = repoPath
		cmd.Run()
		
		// Create a dummy file
		dummyFile := filepath.Join(repoPath, "README.md")
		os.WriteFile(dummyFile, []byte(fmt.Sprintf("# %s\n", repo)), 0644)
		
		// Configure git if needed
		exec.Command("git", "config", "user.email", "test@example.com").Dir = repoPath
		exec.Command("git", "config", "user.name", "Test User").Dir = repoPath
		
		// Add and commit
		exec.Command("git", "add", ".").Dir = repoPath
		exec.Command("git", "commit", "-m", "Initial commit").Dir = repoPath
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tmpDir := t.TempDir()
	binary := buildBinary(t, tmpDir)
	
	// Test init with invalid project name
	cmd := exec.Command(binary, "init")
	output, err := cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(output), "Error")
	
	// Test commands without workspace
	cmds := []string{"start", "stop", "status", "sync"}
	for _, cmdName := range cmds {
		cmd := exec.Command(binary, cmdName)
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		assert.Error(t, err, "%s should fail without workspace", cmdName)
		assert.Contains(t, string(output), "no repo-claude workspace found")
	}
}