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
	
	// Create temporary directory
	tmpDir := t.TempDir()
	
	// Build repo-claude binary
	binary := buildBinary(t, tmpDir)
	
	// Initialize project first (required for all tests)
	projectName := "test-project"
	projectDir := filepath.Join(tmpDir, projectName)
	
	cmd := exec.Command(binary, "init", projectName)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	t.Logf("Init output: %s", string(output))
	require.NoError(t, err, "Init failed: %s", string(output))
	
	// Test init command results
	t.Run("Init", func(t *testing.T) {
		assert.Contains(t, string(output), "🚀 Initializing Repo-Claude workspace")
		assert.Contains(t, string(output), "✨ Workspace initialized")
		
		// Check created files for v2 structure
		assert.FileExists(t, filepath.Join(projectDir, "repo-claude.yaml"))
		assert.FileExists(t, filepath.Join(projectDir, "CLAUDE.md"))
		assert.DirExists(t, filepath.Join(projectDir, "workspaces"))
		assert.DirExists(t, filepath.Join(projectDir, "docs"))
		assert.DirExists(t, filepath.Join(projectDir, "docs", "global"))
		assert.DirExists(t, filepath.Join(projectDir, "docs", "scopes"))
	})
	
	// Test list command
	t.Run("List", func(t *testing.T) {
		cmd := exec.Command(binary, "list")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		
		require.NoError(t, err, "List failed: %s", string(output))
		assert.Contains(t, string(output), "Available Scopes")
		// Should show default scopes from config
		assert.Contains(t, string(output), "wms")
		assert.Contains(t, string(output), "oms")
	})
	
	// Test status command for a scope
	t.Run("Status", func(t *testing.T) {
		// Create a scope first
		cmd := exec.Command(binary, "scope", "create", "wms")
		cmd.Dir = projectDir
		output, _ := cmd.CombinedOutput()
		t.Logf("Scope create output: %s", string(output))
		
		// Now check status
		cmd = exec.Command(binary, "status", "wms")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			// Status might fail if repos aren't cloned yet, but command should be recognized
			assert.NotContains(t, string(output), "unknown command")
		} else {
			assert.Contains(t, string(output), "Scope Status")
		}
	})
	
	// Test pull command
	t.Run("Pull", func(t *testing.T) {
		cmd := exec.Command(binary, "pull", "wms", "--clone-missing")
		cmd.Dir = projectDir
		output, _ := cmd.CombinedOutput()
		
		// Pull will fail since the repos don't actually exist, but command should be recognized
		assert.NotContains(t, string(output), "unknown command")
		// Should see attempt to clone
		t.Logf("Pull output: %s", string(output))
	})
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	tmpDir := t.TempDir()
	binary := buildBinary(t, tmpDir)
	
	// Try to run commands without config
	commands := []string{"list", "start wms", "status wms", "pull wms"}
	
	for _, cmdStr := range commands {
		t.Run(fmt.Sprintf("Command_%s", cmdStr), func(t *testing.T) {
			args := []string{}
			for _, arg := range bytes.Fields([]byte(cmdStr)) {
				args = append(args, string(arg))
			}
			
			cmd := exec.Command(binary, args...)
			cmd.Dir = tmpDir
			output, _ := cmd.CombinedOutput()
			
			// Should fail with config not found message
			assert.Contains(t, string(output), "no repo-claude.yaml found")
		})
	}
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
	cmd.Dir = filepath.Join("..") // Go up one level from test/ to repo-claude-go root
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	err := cmd.Run()
	require.NoError(t, err, "Build failed: %s", out.String())
	
	return binary
}

// createDummyRepos creates dummy git repositories for testing
func createDummyRepos(t *testing.T, projectDir string) {
	// Read config to get repo URLs
	// For testing, we'll just create local git repos
	reposDir := filepath.Join(projectDir, ".test-repos")
	os.MkdirAll(reposDir, 0755)
	
	repos := []string{"wms-core", "wms-inventory", "oms-core"}
	
	for _, repo := range repos {
		repoPath := filepath.Join(reposDir, repo)
		os.MkdirAll(repoPath, 0755)
		
		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = repoPath
		cmd.Run()
		
		// Create a dummy file
		dummyFile := filepath.Join(repoPath, "README.md")
		os.WriteFile(dummyFile, []byte("# "+repo), 0644)
		
		// Add and commit
		cmd = exec.Command("git", "add", ".")
		cmd.Dir = repoPath
		cmd.Run()
		
		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = repoPath
		cmd.Run()
	}
}