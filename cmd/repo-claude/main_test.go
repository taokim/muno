package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandStructure(t *testing.T) {
	// Test that all expected commands exist
	expectedCommands := []string{
		"init",
		"start", 
		"status",
		"list",
		"pull",
		"forall",
	}
	
	app := NewApp()
	for _, cmdName := range expectedCommands {
		t.Run("Command exists: "+cmdName, func(t *testing.T) {
			found := false
			for _, cmd := range app.rootCmd.Commands() {
				if cmd.Use == cmdName || strings.HasPrefix(cmd.Use, cmdName+" ") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected command %s not found", cmdName)
			}
		})
	}
}

func TestStartCommandFlags(t *testing.T) {
	app := NewApp()
	var startTestCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if strings.HasPrefix(cmd.Use, "start") {
			startTestCmd = cmd
			break
		}
	}
	
	if startTestCmd == nil {
		t.Fatal("Start command not found")
	}
	
	// Check that all expected flags exist
	expectedFlags := []string{
		"repos",
		"preset",
		"interactive",
	}
	
	for _, flagName := range expectedFlags {
		t.Run("Flag exists: "+flagName, func(t *testing.T) {
			flag := startTestCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("Expected flag %s not found", flagName)
			}
		})
	}
}


func TestVersionOutput(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	app := NewApp()
	app.SetOutput(&buf, &buf)
	err := app.ExecuteWithArgs([]string{"--version"})
	if err != nil {
		t.Fatalf("Failed to execute version command: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "rc version") {
		t.Error("Version output should contain 'rc version'")
	}
}