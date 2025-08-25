package main

import (
	"bytes"
	"strings"
	"testing"
)

// Test the App structure and command execution
func TestApp_Execute(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantOutput string
		wantErrMsg string
	}{
		{
			name:       "Help command",
			args:       []string{"--help"},
			wantErr:    false,
			wantOutput: "Repo-Claude orchestrates Claude Code",
		},
		{
			name:       "Version command",
			args:       []string{"--version"},
			wantErr:    false,
			wantOutput: "rc version",
		},
		{
			name:       "Unknown command",
			args:       []string{"unknown-command"},
			wantErr:    true,
			wantErrMsg: "unknown command",
		},
		{
			name:       "Init help",
			args:       []string{"init", "--help"},
			wantErr:    false,
			wantOutput: "Initialize a new repo-claude workspace",
		},
		{
			name:       "Start help",
			args:       []string{"start", "--help"},
			wantErr:    false,
			wantOutput: "Start one or more scopes",
		},
		{
			name:       "Status help",
			args:       []string{"status", "--help"},
			wantErr:    false,
			wantOutput: "Display detailed status of all scopes and repositories",
		},
		{
			name:       "Ps help",
			args:       []string{"ps", "--help"},
			wantErr:    false,
			wantOutput: "Display running scopes with numbers",
		},
		{
			name:       "Forall without arguments",
			args:       []string{"forall"},
			wantErr:    true,
			wantErrMsg: "requires at least 1 arg",
		},
		{
			name:       "Init with too many arguments",
			args:       []string{"init", "arg1", "arg2"},
			wantErr:    true,
			wantErrMsg: "accepts at most 1 arg",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			
			// Capture output
			var stdout, stderr bytes.Buffer
			app.SetOutput(&stdout, &stderr)
			
			// Execute with test args
			err := app.ExecuteWithArgs(tt.args)
			
			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			// Check output
			output := stdout.String() + stderr.String()
			
			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Output missing expected string %q\nGot: %s", tt.wantOutput, output)
			}
			
			if tt.wantErrMsg != "" {
				if err == nil {
					t.Errorf("Expected error containing %q but got no error", tt.wantErrMsg)
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) && !strings.Contains(output, tt.wantErrMsg) {
					t.Errorf("Error message missing %q\nGot error: %v\nOutput: %s", tt.wantErrMsg, err, output)
				}
			}
		})
	}
}

// Test version output specifically
func TestApp_Version(t *testing.T) {
	// Save original version
	oldVersion := version
	version = "1.2.3"
	defer func() { version = oldVersion }()
	
	app := NewApp()
	var stdout, stderr bytes.Buffer
	app.SetOutput(&stdout, &stderr)
	
	err := app.ExecuteWithArgs([]string{"--version"})
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}
	
	output := stdout.String()
	if !strings.Contains(output, "1.2.3") {
		t.Errorf("Version output missing version number: %s", output)
	}
}

// Test flag parsing
func TestApp_FlagParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "Init non-interactive flag",
			args: []string{"init", "--help"},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "--non-interactive") {
					t.Error("Init help missing --non-interactive flag")
				}
			},
		},
		{
			name: "Start flags",
			args: []string{"start", "--help"},
			checkOutput: func(t *testing.T, output string) {
				expectedFlags := []string{
					"--repos",
					"--preset",
					"--interactive",
				}
				for _, flag := range expectedFlags {
					if !strings.Contains(output, flag) {
						t.Errorf("Start help missing %s flag", flag)
					}
				}
			},
		},
		{
			name: "Ps flags",
			args: []string{"ps", "--help"},
			checkOutput: func(t *testing.T, output string) {
				expectedFlags := []string{
					"-a, --all",
					"-x, --extended",
					"-f, --full",
					"-l, --long",
					"-u, --user",
					"-s, --sort",
				}
				for _, flag := range expectedFlags {
					if !strings.Contains(output, flag) {
						t.Errorf("Ps help missing %s flag", flag)
					}
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			var stdout, stderr bytes.Buffer
			app.SetOutput(&stdout, &stderr)
			
			err := app.ExecuteWithArgs(tt.args)
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}
			
			output := stdout.String()
			tt.checkOutput(t, output)
		})
	}
}

// Test command structure
func TestApp_CommandStructure(t *testing.T) {
	app := NewApp()
	
	// Check root command
	if app.rootCmd == nil {
		t.Fatal("Root command not initialized")
	}
	
	// Check all expected commands exist
	expectedCommands := []string{
		"init",
		"start",
		"kill",
		"status",
		"sync",
		"forall",
		"ps",
	}
	
	commands := app.rootCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}
	
	for _, expected := range expectedCommands {
		found := false
		for cmdUse := range commandMap {
			if strings.HasPrefix(cmdUse, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Command %s not found in root command", expected)
		}
	}
}

// Test main function integration via App
func TestMainIntegration(t *testing.T) {
	// Test successful run
	app := NewApp()
	app.rootCmd.SetArgs([]string{"--help"})
	err := app.Execute()
	if err != nil {
		t.Errorf("Execute() with --help returned error: %v", err)
	}
	
	// Test error case (unknown command)
	app2 := NewApp()
	app2.rootCmd.SetArgs([]string{"unknown-command"})
	err = app2.Execute()
	if err == nil {
		t.Error("Execute() with unknown command should return error")
	}
}