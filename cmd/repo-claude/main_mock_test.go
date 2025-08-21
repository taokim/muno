package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Test main function
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

// Test the run function execution
func TestRunFunction(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{
			name:     "Help command",
			args:     []string{"--help"},
			wantExit: 0,
		},
		{
			name:     "Version command", 
			args:     []string{"--version"},
			wantExit: 0,
		},
		{
			name:     "Invalid command",
			args:     []string{"invalid-command"},
			wantExit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create app with test output
			app := NewApp()
			var stdout, stderr bytes.Buffer
			app.SetOutput(&stdout, &stderr)
			
			// Execute with test args
			err := app.ExecuteWithArgs(tt.args)
			
			if tt.wantExit == 0 && err != nil {
				t.Errorf("Expected success but got error: %v", err)
			}
			if tt.wantExit != 0 && err == nil {
				t.Errorf("Expected error but got success")
			}
		})
	}
}


// Test command initialization
func TestCommandInitialization(t *testing.T) {
	// Test that all commands are properly initialized
	commands := []string{"init", "start", "kill", "status", "sync", "forall", "ps"}
	
	app := NewApp()
	
	for _, cmdName := range commands {
		t.Run(cmdName, func(t *testing.T) {
			found := false
			for _, cmd := range app.rootCmd.Commands() {
				if strings.HasPrefix(cmd.Use, cmdName) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Command %s not found in root command", cmdName)
			}
		})
	}
}

// Test command flags
func TestCommandFlags(t *testing.T) {
	tests := []struct {
		cmdName   string
		flagName  string
		flagType  string
		shorthand string
	}{
		{"init", "non-interactive", "bool", ""},
		{"start", "repos", "string", ""},
		{"start", "preset", "string", ""},
		{"start", "interactive", "bool", ""},
		{"ps", "all", "bool", "a"},
		{"ps", "extended", "bool", "x"},
		{"ps", "full", "bool", "f"},
		{"ps", "long", "bool", "l"},
		{"ps", "user", "bool", "u"},
		{"ps", "sort", "string", "s"},
	}

	app := NewApp()

	for _, tt := range tests {
		t.Run(tt.cmdName+"_"+tt.flagName, func(t *testing.T) {
			var cmd *cobra.Command
			
			// Find the command
			if tt.cmdName == "root" {
				cmd = app.rootCmd
			} else {
				for _, c := range app.rootCmd.Commands() {
					if strings.HasPrefix(c.Use, tt.cmdName) {
						cmd = c
						break
					}
				}
			}
			
			if cmd == nil {
				t.Fatalf("Command %s not found", tt.cmdName)
			}
			
			// Check flag exists
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag --%s not found in %s command", tt.flagName, tt.cmdName)
				return
			}
			
			// Check shorthand if specified
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("Flag --%s shorthand = %s, want %s", tt.flagName, flag.Shorthand, tt.shorthand)
			}
		})
	}
}

// Test help output
func TestHelpOutput(t *testing.T) {
	commands := []string{"", "init", "start", "kill", "status", "sync", "forall", "ps"}
	
	for _, cmd := range commands {
		t.Run("help_"+cmd, func(t *testing.T) {
			app := NewApp()
			var stdout, stderr bytes.Buffer
			app.SetOutput(&stdout, &stderr)
			
			args := []string{"--help"}
			if cmd != "" {
				args = []string{cmd, "--help"}
			}
			
			err := app.ExecuteWithArgs(args)
			
			if err != nil {
				t.Errorf("Help command failed: %v", err)
			}
			
			output := stdout.String()
			if len(output) == 0 {
				t.Error("Help output is empty")
			}
			
			// Check for expected content
			if cmd == "" {
				// Root help should mention repo-claude or rc
				if !strings.Contains(output, "Multi-repository orchestration") && !strings.Contains(output, "rc") {
					t.Error("Root help missing repo-claude reference")
				}
			} else {
				// Command help should mention the command
				if !strings.Contains(output, cmd) {
					t.Errorf("Command help for %s missing command name", cmd)
				}
			}
		})
	}
}

// Test version output format
func TestVersionFlag(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer
	app.SetOutput(&stdout, &stderr)
	
	err := app.ExecuteWithArgs([]string{"--version"})
	if err != nil {
		t.Errorf("Version command failed: %v", err)
	}
	
	output := stdout.String()
	if !strings.Contains(output, version) {
		t.Errorf("Version output doesn't contain version string %s", version)
	}
}

// Test command error handling
func TestMainCommandErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Unknown command",
			args:    []string{"unknown"},
			wantErr: true,
			errMsg:  "unknown command",
		},
		{
			name:    "Init with too many args",
			args:    []string{"init", "arg1", "arg2"},
			wantErr: true,
			errMsg:  "accepts at most 1 arg",
		},
		{
			name:    "Forall without command",
			args:    []string{"forall"},
			wantErr: true,
			errMsg:  "requires at least 1 arg",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			var stdout, stderr bytes.Buffer
			app.SetOutput(&stdout, &stderr)
			
			err := app.ExecuteWithArgs(tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if err != nil && tt.errMsg != "" {
				errStr := err.Error()
				if !strings.Contains(errStr, tt.errMsg) {
					t.Errorf("Error message = %v, want to contain %v", errStr, tt.errMsg)
				}
			}
		})
	}
}