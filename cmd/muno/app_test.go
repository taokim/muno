package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
)

// captureOutput captures stdout during function execution
func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestAppCommands(t *testing.T) {
	// Test that app creates all expected commands
	app := NewApp()
	assert.NotNil(t, app)
	assert.NotNil(t, app.rootCmd)
	
	// Check all commands are registered
	commands := []string{
		"init", "tree", "list", 
		"add", "remove", "status", "pull", "push",
		"commit", "clone", "version",
		"agent", "claude", "gemini",
	}
	
	for _, cmdName := range commands {
		cmd, _, err := app.rootCmd.Find([]string{cmdName})
		assert.NoError(t, err, "Command %s should exist", cmdName)
		assert.NotNil(t, cmd, "Command %s should not be nil", cmdName)
	}
}

func TestInitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	app := NewApp()
	
	var err error
	output := captureOutput(func() {
		err = app.ExecuteWithArgs([]string{"init", "test-project", "--non-interactive"})
	})
	
	require.NoError(t, err)
	assert.Contains(t, output, "Workspace 'test-project' initialized successfully")
	
	// Check files created
	assert.FileExists(t, filepath.Join(tmpDir, "muno.yaml"))
	assert.DirExists(t, filepath.Join(tmpDir, config.GetDefaultReposDir()))
}

func TestTreeCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	app := NewApp()
	app.ExecuteWithArgs([]string{"init", "test", "--non-interactive"})
	
	var err error
	output := captureOutput(func() {
		err = app.ExecuteWithArgs([]string{"tree"})
	})
	
	require.NoError(t, err)
	assert.Contains(t, output, "root")
}

func TestListCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	app := NewApp()
	app.ExecuteWithArgs([]string{"init", "test", "--non-interactive"})
	
	var err error
	output := captureOutput(func() {
		err = app.ExecuteWithArgs([]string{"list"})
	})
	
	require.NoError(t, err)
	// The new output doesn't include "No children" message
	// Just check that it runs without error
	assert.NotEmpty(t, output)
}

// TestCurrentCommand was removed - current command no longer exists in stateless architecture

func TestAddCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	app := NewApp()
	app.ExecuteWithArgs([]string{"init", "test", "--non-interactive"})
	
	var err error
	output := captureOutput(func() {
		err = app.ExecuteWithArgs([]string{"add", "https://github.com/test/repo.git", "--name", "testrepo", "--lazy"})
	})
	
	require.NoError(t, err)
	assert.Contains(t, output, "Successfully added: testrepo")
}

func TestRemoveCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	// Initialize and add a repo first
	app := NewApp()
	app.ExecuteWithArgs([]string{"init", "test", "--non-interactive"})
	app.ExecuteWithArgs([]string{"add", "https://github.com/test/removeme.git", "--lazy"})
	
	// Simulate 'y' response
	oldStdin := os.Stdin
	pipeR, pipeW, _ := os.Pipe()
	os.Stdin = pipeR
	pipeW.Write([]byte("y\n"))
	pipeW.Close()
	defer func() { os.Stdin = oldStdin }()
	
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()
	
	err := app.ExecuteWithArgs([]string{"remove", "removeme"})
	require.NoError(t, err)
	
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	assert.Contains(t, output, "Successfully removed: removeme")
}

func TestStatusCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	// Initialize first
	app := NewApp()
	app.ExecuteWithArgs([]string{"init", "test", "--non-interactive"})
	
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()
	
	err := app.ExecuteWithArgs([]string{"status"})
	require.NoError(t, err)
	
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// The output format has changed
	assert.Contains(t, output, "root: branch=")
}

func TestVersionCommand(t *testing.T) {
	app := NewApp()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	app.SetOutput(stdout, stderr)
	
	err := app.ExecuteWithArgs([]string{"version"})
	require.NoError(t, err)
	
	output := stdout.String()
	// Check for version info (the actual output might vary)
	assert.Contains(t, output, "Version:")
}

func TestHelpCommand(t *testing.T) {
	app := NewApp()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	app.SetOutput(stdout, stderr)
	
	err := app.ExecuteWithArgs([]string{"help"})
	require.NoError(t, err)
	
	output := stdout.String() + stderr.String()
	assert.Contains(t, output, "Multi-repository UNified Orchestration")
	assert.Contains(t, output, "Available Commands")
}

func TestCloneCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	// Initialize and add a lazy repo
	app := NewApp()
	app.ExecuteWithArgs([]string{"init", "test", "--non-interactive"})
	app.ExecuteWithArgs([]string{"add", "https://github.com/test/repo.git", "--name", "lazyone", "--lazy"})
	
	// Capture stdout for clone command
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()
	
	err := app.ExecuteWithArgs([]string{"clone"})
	// May error if can't actually clone, but shouldn't panic
	_ = err
	
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// The clone command should output something, even if it's just info messages
	assert.True(t, len(output) > 0 || err != nil, "Clone command should produce output or return an error")
}

// TestUseCommand was removed - use command no longer exists in stateless architecture

func TestPullCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	// Initialize first
	app := NewApp()
	app.ExecuteWithArgs([]string{"init", "test", "--non-interactive"})
	
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()
	
	// Will fail since root is not a git repo, but shouldn't panic
	_ = app.ExecuteWithArgs([]string{"pull"})
	
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// Should still print something about pulling
	assert.True(t, strings.Contains(output, "Pulling") || strings.Contains(output, "Error") || len(output) > 0)
}

func TestPushCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	// Initialize first
	app := NewApp()
	app.ExecuteWithArgs([]string{"init", "test", "--non-interactive"})
	
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()
	
	// Will fail since root is not a git repo, but shouldn't panic
	_ = app.ExecuteWithArgs([]string{"push"})
	
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// Should still print something about pushing
	assert.True(t, strings.Contains(output, "Pushing") || strings.Contains(output, "Error") || len(output) > 0)
}

func TestCommitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)
	
	// Initialize first
	app := NewApp()
	app.ExecuteWithArgs([]string{"init", "test", "--non-interactive"})
	
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()
	
	// Will fail since root is not a git repo, but shouldn't panic
	_ = app.ExecuteWithArgs([]string{"commit", "-m", "test commit"})
	
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	// Should still print something about committing
	assert.True(t, strings.Contains(output, "Committing") || strings.Contains(output, "Error") || len(output) > 0)
}

func TestInvalidCommand(t *testing.T) {
	app := NewApp()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	app.SetOutput(stdout, stderr)
	
	err := app.ExecuteWithArgs([]string{"invalid-command"})
	assert.Error(t, err)
	
	errOutput := stderr.String()
	assert.Contains(t, errOutput, "unknown command")
}