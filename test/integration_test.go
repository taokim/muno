package test

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
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
	
	// Build muno binary
	binary := buildBinary(t, tmpDir)
	
	// Initialize project first (required for all tests)
	projectName := "test-project"
	
	cmd := exec.Command(binary, "init", projectName)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	t.Logf("Init output: %s", string(output))
	require.NoError(t, err, "Init failed: %s", string(output))
	
	// Test init command results
	t.Run("Init", func(t *testing.T) {
		assert.Contains(t, string(output), "Workspace 'test-project' initialized successfully")
		
		// Check created files - init creates in current dir, not subdirectory
		assert.FileExists(t, filepath.Join(tmpDir, "muno.yaml"))
		// CLAUDE.md is no longer created by default
		assert.DirExists(t, filepath.Join(tmpDir, config.GetDefaultReposDir()))
	})
	
	// Test list command
	t.Run("List", func(t *testing.T) {
		cmd := exec.Command(binary, "list")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		
		require.NoError(t, err, "List failed: %s", string(output))
		// Should show empty list for new workspace
		// Check for either the old or new format
		assert.True(t, strings.Contains(string(output), "No children") || strings.Contains(string(output), "No repositories at this level"))
	})
	
	// Test status command
	t.Run("Status", func(t *testing.T) {
		cmd := exec.Command(binary, "status")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
				
		require.NoError(t, err, "Status failed: %s", string(output))
		assert.Contains(t, string(output), "Tree Status")
	})
	
	// Test pull command
	t.Run("Pull", func(t *testing.T) {
		cmd := exec.Command(binary, "pull")
		cmd.Dir = tmpDir
		output, _ := cmd.CombinedOutput()
		
		// Pull will fail on root (not a git repo), but command should be recognized
		assert.NotContains(t, string(output), "unknown command")
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
			// Check for either error message format
			assert.True(t, strings.Contains(string(output), "muno.yaml not found") || strings.Contains(string(output), "not in a MUNO workspace"))
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

// buildBinary builds the muno binary
func buildBinary(t *testing.T, tmpDir string) string {
	binary := filepath.Join(tmpDir, "muno")
	
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/muno")
	cmd.Dir = filepath.Join("..") // Go up one level from test/ to muno-go root
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	err := cmd.Run()
	require.NoError(t, err, "Build failed: %s", out.String())
	
	return binary
}

