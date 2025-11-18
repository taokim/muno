package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test shell-init with zsh shell to increase coverage
func TestShellInit_Zsh(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer
	app.SetOutput(&stdout, &stderr)

	err := app.ExecuteWithArgs([]string{"shell-init", "--shell", "zsh"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "mcd()")
}

// Test help command output
func TestHelpCommandOutput(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer
	app.SetOutput(&stdout, &stderr)

	err := app.ExecuteWithArgs([]string{"--help"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Multi-repository")
}

// Test version command output
func TestVersionOutput(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer
	app.SetOutput(&stdout, &stderr)

	err := app.ExecuteWithArgs([]string{"version"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Version:")
}