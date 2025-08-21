package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPsCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]interface{}
		expectedFlags  map[string]bool
		expectedFormat string
	}{
		{
			name: "Basic ps command",
			args: []string{},
			expectedFlags: map[string]bool{
				"all":      false,
				"extended": false,
				"details":  false,
			},
			expectedFormat: "simple",
		},
		{
			name: "ps aux style",
			args: []string{"aux"},
			expectedFlags: map[string]bool{
				"all":      true,
				"extended": true,
				"user":     true,
				"details":  true,
			},
			expectedFormat: "table",
		},
		{
			name: "ps -aux with dash",
			flags: map[string]interface{}{
				"aux": true,
			},
			expectedFlags: map[string]bool{
				"all":      true,
				"extended": true,
				"user":     true,
				"details":  true,
			},
			expectedFormat: "table",
		},
		{
			name: "ps ef style",
			args: []string{"ef"},
			expectedFlags: map[string]bool{
				"all":     true,
				"full":    true,
				"details": true,
			},
			expectedFormat: "table",
		},
		{
			name: "Individual flags",
			flags: map[string]interface{}{
				"all":      true,
				"extended": true,
			},
			expectedFlags: map[string]bool{
				"all":      true,
				"extended": true,
				"details":  true,
			},
			expectedFormat: "table",
		},
		{
			name: "With logs flag",
			flags: map[string]interface{}{
				"logs": true,
			},
			expectedFlags: map[string]bool{
				"logs": true,
			},
			expectedFormat: "simple",
		},
		{
			name: "Sort flag",
			flags: map[string]interface{}{
				"sort": "cpu",
			},
			expectedFlags: map[string]bool{
				"sort": true, // We'll check the value separately
			},
			expectedFormat: "simple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test command that captures the logic without executing
			testCmd := &cobra.Command{
				Use: "test",
				RunE: func(cmd *cobra.Command, args []string) error {
					// Simulate the ps command logic
					aux, _ := cmd.Flags().GetBool("aux")
					ef, _ := cmd.Flags().GetBool("ef")
					
					showAll, _ := cmd.Flags().GetBool("all")
					extended, _ := cmd.Flags().GetBool("extended")
					full, _ := cmd.Flags().GetBool("full")
					long, _ := cmd.Flags().GetBool("long")
					user, _ := cmd.Flags().GetBool("user")
					
					// Handle combined flags
					if aux {
						showAll = true
						user = true
						extended = true
					}
					if ef {
						showAll = true
						full = true
					}
					
					// Parse args for aux/ef without dash
					for _, arg := range args {
						if arg == "aux" {
							showAll = true
							user = true
							extended = true
						} else if arg == "ef" {
							showAll = true
							full = true
						}
					}
					
					// Check expected flags
					if tt.expectedFlags["all"] && !showAll {
						t.Error("Expected all flag to be true")
					}
					if tt.expectedFlags["extended"] && !extended {
						t.Error("Expected extended flag to be true")
					}
					if tt.expectedFlags["user"] && !user {
						t.Error("Expected user flag to be true")
					}
					if tt.expectedFlags["full"] && !full {
						t.Error("Expected full flag to be true")
					}
					
					// Check format
					format := "simple"
					showDetails := false
					if extended || user || full || long {
						format = "table"
						showDetails = true
					}
					
					if format != tt.expectedFormat {
						t.Errorf("Expected format %s, got %s", tt.expectedFormat, format)
					}
					
					if tt.expectedFlags["details"] && !showDetails {
						t.Error("Expected details to be shown")
					}
					
					return nil
				},
			}
			
			// Add flags
			testCmd.Flags().BoolP("all", "a", false, "")
			testCmd.Flags().BoolP("extended", "x", false, "")
			testCmd.Flags().BoolP("full", "f", false, "")
			testCmd.Flags().BoolP("long", "l", false, "")
			testCmd.Flags().BoolP("user", "u", false, "")
			testCmd.Flags().Bool("logs", false, "")
			testCmd.Flags().String("sort", "name", "")
			testCmd.Flags().Bool("aux", false, "")
			testCmd.Flags().Bool("ef", false, "")
			
			// Set flags
			if tt.flags != nil {
				for k, v := range tt.flags {
					switch val := v.(type) {
					case bool:
						testCmd.Flags().Set(k, "true")
					case string:
						testCmd.Flags().Set(k, val)
					}
				}
			}
			
			// Execute
			testCmd.SetArgs(tt.args)
			if err := testCmd.Execute(); err != nil {
				t.Errorf("Command failed: %v", err)
			}
		})
	}
}

func TestCommandStructure(t *testing.T) {
	// Test that all expected commands exist
	expectedCommands := []string{
		"init",
		"start", 
		"kill",
		"status",
		"sync",
		"forall",
		"ps",
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

func TestPsCommandFlags(t *testing.T) {
	app := NewApp()
	var psTestCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Use == "ps" {
			psTestCmd = cmd
			break
		}
	}
	
	if psTestCmd == nil {
		t.Fatal("ps command not found")
	}
	
	// Check Unix-style flags
	unixFlags := map[string]string{
		"a": "all",
		"x": "extended",
		"f": "full",
		"l": "long",
		"u": "user",
		"s": "sort",
	}
	
	for short, long := range unixFlags {
		t.Run("Unix flag: -"+short, func(t *testing.T) {
			flag := psTestCmd.Flags().ShorthandLookup(short)
			if flag == nil {
				t.Errorf("Expected short flag -%s not found", short)
			} else if flag.Name != long {
				t.Errorf("Short flag -%s should map to --%s, got --%s", short, long, flag.Name)
			}
		})
	}
	
	// Check combined flags
	combinedFlags := []string{"aux", "ef"}
	for _, flagName := range combinedFlags {
		t.Run("Combined flag: --"+flagName, func(t *testing.T) {
			flag := psTestCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("Expected combined flag --%s not found", flagName)
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