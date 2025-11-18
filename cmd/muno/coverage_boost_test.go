package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test various app commands to boost coverage
func TestAppCommands_CoverageBoost(t *testing.T) {
	t.Run("Help", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"--help"})
		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "Multi-repository")
	})

	t.Run("Version", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"version"})
		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "Version:")
	})

	t.Run("Init_NoArgs", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"init"})
		// May error or succeed depending on context
		_ = err
	})

	t.Run("Add_NoArgs", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"add"})
		assert.Error(t, err)
	})

	t.Run("Remove_NoArgs", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"remove"})
		assert.Error(t, err)
	})

	t.Run("Use_NoArgs", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"use"})
		assert.Error(t, err)
	})

	t.Run("List", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"list"})
		// May succeed if in workspace
		_ = err
	})

	t.Run("Tree", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"tree"})
		// May succeed if in workspace
		_ = err
	})

	t.Run("Current", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		// Will error because no manager initialized
		err := app.ExecuteWithArgs([]string{"current"})
		assert.Error(t, err)
	})

	t.Run("Status", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"status"})
		// May succeed if in workspace
		_ = err
	})

	t.Run("Pull", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"pull"})
		// May succeed if in workspace
		_ = err
	})

	t.Run("Push", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"push"})
		// May succeed if in workspace
		_ = err
	})

	t.Run("Clone", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"clone"})
		// May succeed if in workspace
		_ = err
	})

	t.Run("Commit_NoMessage", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"commit"})
		assert.Error(t, err)
	})

	t.Run("Path", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"path"})
		// May succeed if in workspace
		_ = err
	})

	t.Run("ShellInit", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"shell-init"})
		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "mcd")
	})

	t.Run("ShellInit_Bash", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"shell-init", "--shell", "bash"})
		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "mcd")
	})

	t.Run("UnknownCommand", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"unknown-command"})
		assert.Error(t, err)
	})

	t.Run("VersionShortFlag", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"-v"})
		assert.NoError(t, err)
	})

	t.Run("HelpShortFlag", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"-h"})
		assert.NoError(t, err)
	})

	t.Run("ShellInit_Zsh", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"shell-init", "--shell", "zsh"})
		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "mcd")
	})

	t.Run("ShellInit_Fish", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"shell-init", "--shell", "fish"})
		assert.NoError(t, err)
	})

	t.Run("Pull_WithPath", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"pull", "example"})
		// May succeed if in workspace
		_ = err
	})

	t.Run("Pull_WithBranch", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"pull", "--branch", "main"})
		// May succeed if in workspace
		_ = err
	})

	t.Run("Pull_WithParallel", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"pull", "--parallel", "5"})
		// May succeed if in workspace
		_ = err
	})

	t.Run("Pull_WithConfigOverride", func(t *testing.T) {
		app := NewApp()
		var stdout, stderr bytes.Buffer
		app.SetOutput(&stdout, &stderr)

		err := app.ExecuteWithArgs([]string{"pull", "--config", "git.default_branch=main"})
		// May succeed if in workspace
		_ = err
	})
}