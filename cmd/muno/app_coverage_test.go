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