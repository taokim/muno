package main

import (
	"os"
	"strings"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestGetShellTemplate(t *testing.T) {
	tests := []struct {
		name      string
		shellType string
		wantEmpty bool
		contains  []string
	}{
		{
			name:      "bash template",
			shellType: "bash",
			wantEmpty: false,
			contains:  []string{"{{CMD_NAME}}", "cd", "path", "--ensure"},
		},
		{
			name:      "zsh template",
			shellType: "zsh",
			wantEmpty: false,
			contains:  []string{"{{CMD_NAME}}", "cd", "path", "--ensure"},
		},
		{
			name:      "fish template",
			shellType: "fish",
			wantEmpty: false,
			contains:  []string{"{{CMD_NAME}}", "cd", "path", "--ensure"},
		},
		{
			name:      "unknown shell",
			shellType: "unknown",
			wantEmpty: false,  // getShellTemplate returns bash template by default
			contains:  []string{"{{CMD_NAME}}", "cd", "path", "--ensure"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getShellTemplate(tt.shellType)
			if tt.wantEmpty {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
				for _, substr := range tt.contains {
					assert.Contains(t, result, substr, "template should contain %s", substr)
				}
			}
		})
	}
}

func TestRenderShellTemplate(t *testing.T) {
	template := "function {{CMD_NAME}}() { echo hello }"
	cmdName := "test_cmd"
	
	result := renderShellTemplate(template, cmdName)
	assert.Contains(t, result, "test_cmd")
	assert.NotContains(t, result, "{{CMD_NAME}}")
	
	// Test with empty template
	result = renderShellTemplate("", cmdName)
	assert.Empty(t, result)
	
	// Test with template without placeholder
	result = renderShellTemplate("no placeholder here", cmdName)
	assert.Equal(t, "no placeholder here", result)
}

func TestFindAvailableCommandNames(t *testing.T) {
	baseCmd := "test-cmd"
	
	names := findAvailableCommandNames(baseCmd)
	
	// Should return at least one name
	assert.NotEmpty(t, names)
	
	// The function returns specific variants: base+"d", "m"+base, base+"2", "go"+base
	// But only includes those that don't exist as commands
	// We can't predict which ones will be available, but we know the pattern
	for _, name := range names {
		// Check it's one of the expected patterns
		validPattern := false
		if name == baseCmd+"d" || name == "m"+baseCmd || 
		   name == baseCmd+"2" || name == "go"+baseCmd {
			validPattern = true
		}
		assert.True(t, validPattern, "Name %s should match expected pattern", name)
	}
}

func TestGenerateShellScript(t *testing.T) {
	tests := []struct {
		name      string
		shellType string
		cmdName   string
		wantEmpty bool
	}{
		{
			name:      "bash script",
			shellType: "bash",
			cmdName:   "mcd",
			wantEmpty: false,
		},
		{
			name:      "zsh script",
			shellType: "zsh",
			cmdName:   "mcd",
			wantEmpty: false,
		},
		{
			name:      "fish script",
			shellType: "fish",
			cmdName:   "mcd",
			wantEmpty: false,
		},
		{
			name:      "unknown shell",
			shellType: "unknown",
			cmdName:   "mcd",
			wantEmpty: false,  // generateShellScript returns bash template by default
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateShellScript(tt.shellType, tt.cmdName)
			if tt.wantEmpty {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, tt.cmdName)
				// Should not contain template placeholders
				assert.NotContains(t, result, "{{CMD_NAME}}")
			}
		})
	}
}

func TestGetShellConfigFile(t *testing.T) {
	tests := []struct {
		name  string
		shell string
		want  string
	}{
		{
			name:  "bash config",
			shell: "bash",
			want:  ".bashrc",
		},
		{
			name:  "zsh config",
			shell: "zsh",
			want:  ".zshrc",
		},
		{
			name:  "fish config",
			shell: "fish",
			want:  ".config/fish/config.fish",
		},
		{
			name:  "unknown shell",
			shell: "unknown",
			want:  ".bashrc",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getShellConfigFile(tt.shell)
			assert.True(t, strings.HasSuffix(result, tt.want))
		})
	}
}

func TestDetectShell(t *testing.T) {
	// Save original SHELL env
	oldShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", oldShell)
	
	tests := []struct {
		name     string
		shellEnv string
		want     string
	}{
		{
			name:     "bash shell",
			shellEnv: "/bin/bash",
			want:     "bash",
		},
		{
			name:     "zsh shell",
			shellEnv: "/bin/zsh",
			want:     "zsh",
		},
		{
			name:     "fish shell",
			shellEnv: "/usr/bin/fish",
			want:     "fish",
		},
		{
			name:     "unknown shell",
			shellEnv: "/bin/unknown",
			want:     "bash",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("SHELL", tt.shellEnv)
			result := detectShell()
			assert.Equal(t, tt.want, result)
		})
	}
}