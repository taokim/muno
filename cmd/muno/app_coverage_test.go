package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test isReleaseVersion function
func TestIsReleaseVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "valid release version v1.0.0",
			version:  "v1.0.0",
			expected: true,
		},
		{
			name:     "valid release version v0.4.0",
			version:  "v0.4.0",
			expected: true,
		},
		{
			name:     "dev version with commit info",
			version:  "v0.4.0-5-gabcd123",
			expected: false,
		},
		{
			name:     "dev version with dash",
			version:  "v1.0.0-dev",
			expected: false,
		},
		{
			name:     "empty version",
			version:  "",
			expected: false,
		},
		{
			name:     "version without v prefix",
			version:  "1.0.0",
			expected: false,
		},
		{
			name:     "just v",
			version:  "v",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isReleaseVersion(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAgentCommand tests the agent command parsing
func TestAgentCommand(t *testing.T) {
	// Create temporary directory and change to it to ensure no workspace exists
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	app := NewApp()
	
	tests := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "agent with no args",
			args:        []string{"agent"},
			expectError: true, // Will fail because no workspace
			description: "Should try to use default agent (claude)",
		},
		{
			name:        "agent with specific name",
			args:        []string{"agent", "gemini"},
			expectError: true, // Will fail because no workspace
			description: "Should try to use gemini agent",
		},
		{
			name:        "agent with name and path",
			args:        []string{"agent", "claude", "backend"},
			expectError: true, // Will fail because no workspace
			description: "Should try to use claude at backend path",
		},
		{
			name:        "agent with pass-through args",
			args:        []string{"agent", "gemini", "--", "--model", "pro"},
			expectError: true, // Will fail because no workspace
			description: "Should pass args to gemini",
		},
		{
			name:        "agent help",
			args:        []string{"agent", "--help"},
			expectError: false,
			description: "Should show agent help",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			app.SetOutput(&stdout, &stderr)
			
			err := app.ExecuteWithArgs(tt.args)
			
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				// Check for help output
				if len(tt.args) > 1 && tt.args[1] == "--help" {
					output := stdout.String()
					assert.Contains(t, output, "Start an AI agent session")
					assert.Contains(t, output, "Available agents:")
					assert.Contains(t, output, "claude (default)")
					assert.Contains(t, output, "gemini")
				}
			}
		})
	}
}

// TestClaudeCommand tests the claude command
func TestClaudeCommand(t *testing.T) {
	// Create temporary directory and change to it to ensure no workspace exists
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	app := NewApp()
	
	tests := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "claude with no args",
			args:        []string{"claude"},
			expectError: true, // Will fail because no workspace
			description: "Should try to start claude at current location",
		},
		{
			name:        "claude with path",
			args:        []string{"claude", "backend"},
			expectError: true, // Will fail because no workspace
			description: "Should try to start claude at backend path",
		},
		{
			name:        "claude with pass-through args",
			args:        []string{"claude", "--", "--model", "opus"},
			expectError: true, // Will fail because no workspace
			description: "Should pass args to claude",
		},
		{
			name:        "claude help",
			args:        []string{"claude", "--help"},
			expectError: false,
			description: "Should show claude help",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			app.SetOutput(&stdout, &stderr)
			
			err := app.ExecuteWithArgs(tt.args)
			
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				// Check for help output
				if len(tt.args) > 1 && tt.args[1] == "--help" {
					output := stdout.String()
					assert.Contains(t, output, "Start a Claude")
				}
			}
		})
	}
}

// TestGeminiCommand tests the gemini command
func TestGeminiCommand(t *testing.T) {
	// Create temporary directory and change to it to ensure no workspace exists
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	app := NewApp()
	
	tests := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "gemini with no args",
			args:        []string{"gemini"},
			expectError: true, // Will fail because no workspace
			description: "Should try to start gemini at current location",
		},
		{
			name:        "gemini with path",
			args:        []string{"gemini", "frontend"},
			expectError: true, // Will fail because no workspace
			description: "Should try to start gemini at frontend path",
		},
		{
			name:        "gemini with pass-through args",
			args:        []string{"gemini", ".", "--", "--model", "pro", "--temperature", "0.7"},
			expectError: true, // Will fail because no workspace
			description: "Should pass args to gemini",
		},
		{
			name:        "gemini help",
			args:        []string{"gemini", "--help"},
			expectError: false,
			description: "Should show gemini help",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			app.SetOutput(&stdout, &stderr)
			
			err := app.ExecuteWithArgs(tt.args)
			
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				// Check for help output
				if len(tt.args) > 1 && tt.args[1] == "--help" {
					output := stdout.String()
					assert.Contains(t, output, "Start a Gemini")
				}
			}
		})
	}
}

// TestAgentCommandIntegration tests agent command with workspace
func TestAgentCommandIntegration(t *testing.T) {
	t.Skip("Skipping agent integration tests - requires mock process provider injection")
	// Create temporary directory with initialized workspace
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	// Initialize workspace first
	app := NewApp()
	err := app.ExecuteWithArgs([]string{"init", "test-workspace", "--non-interactive"})
	if err != nil {
		t.Fatalf("Failed to initialize workspace: %v", err)
	}
	
	// Now these should execute successfully (DefaultProcessProvider returns success)
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "agent default claude",
			args: []string{"agent"},
		},
		{
			name: "agent with gemini",
			args: []string{"agent", "gemini"},
		},
		{
			name: "claude command",
			args: []string{"claude"},
		},
		{
			name: "gemini command",
			args: []string{"gemini"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			app.SetOutput(&stdout, &stderr)
			
			err := app.ExecuteWithArgs(tt.args)
			// These should succeed with DefaultProcessProvider
			assert.NoError(t, err, "Command should succeed with mock process provider")
		})
	}
}

// Test formatVersionDetails with different scenarios
func TestFormatVersionDetailsMore(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		buildTime string
		gitCommit string
		gitBranch string
		expected  string
	}{
		{
			name:      "release version",
			version:   "v1.0.0",
			buildTime: "2025-01-01T12:00:00Z",
			gitCommit: "abc123",
			gitBranch: "main",
			expected:  `Version:     v1.0.0
Type:        release
Git Commit:  abc123
Git Branch:  main
Build Time:  2025-01-01T12:00:00Z`,
		},
		{
			name:      "dev version with commit",
			version:   "v1.0.0-5-gabc123",
			buildTime: "2025-01-01T12:00:00Z",
			gitCommit: "abc123",
			gitBranch: "develop",
			expected:  `Version:     v1.0.0-5-gabc123
Type:        dev
Git Commit:  abc123
Git Branch:  develop
Build Time:  2025-01-01T12:00:00Z`,
		},
		{
			name:      "dirty version",
			version:   "v1.0.0-dirty",
			buildTime: "2025-01-01T12:00:00Z",
			gitCommit: "abc123",
			gitBranch: "main",
			expected:  `Version:     v1.0.0-dirty
Type:        dev (uncommitted changes)
Git Commit:  abc123
Git Branch:  main
Build Time:  2025-01-01T12:00:00Z`,
		},
		{
			name:      "dev not in git",
			version:   "dev",
			buildTime: "unknown",
			gitCommit: "unknown",
			gitBranch: "unknown",
			expected:  `Version:     dev
Type:        dev (not in git repo)
Git Commit:  unknown
Git Branch:  unknown
Build Time:  unknown`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origVersion := version
			origBuildTime := buildTime
			origGitCommit := gitCommit
			origGitBranch := gitBranch

			// Set test values
			version = tt.version
			buildTime = tt.buildTime
			gitCommit = tt.gitCommit
			gitBranch = tt.gitBranch

			result := formatVersionDetails()
			assert.Equal(t, tt.expected, result)

			// Restore original values
			version = origVersion
			buildTime = origBuildTime
			gitCommit = origGitCommit
			gitBranch = origGitBranch
		})
	}
}

// Test newAddCmd with more edge cases
func TestNewAddCmdEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Change to temp dir
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	app := NewApp()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	app.SetOutput(stdout, stderr)

	// Test add command with no arguments (should error)
	err := app.ExecuteWithArgs([]string{"add"})
	assert.Error(t, err)
}

// Test newRemoveCmd with more edge cases
func TestNewRemoveCmdEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Change to temp dir
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	app := NewApp()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	app.SetOutput(stdout, stderr)

	// Test remove command with no arguments (should error)
	err := app.ExecuteWithArgs([]string{"remove"})
	assert.Error(t, err)
}

// Test newCurrentCmd with more scenarios
func TestNewCurrentCmdMore(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Change to temp dir
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	app := NewApp()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	app.SetOutput(stdout, stderr)

	// Test current command without initialization
	err := app.ExecuteWithArgs([]string{"current"})
	// Should error since no workspace is initialized
	assert.Error(t, err)
}