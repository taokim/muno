package adapters

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/taokim/muno/internal/interfaces"
)

func TestProcessAdapter_Execute(t *testing.T) {
	adapter := NewProcessAdapter()
	ctx := context.Background()
	
	// Test successful command
	result, err := adapter.Execute(ctx, "echo", []string{"hello"}, interfaces.ProcessOptions{})
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "hello")
	
	// Test command with non-zero exit
	result, err = adapter.Execute(ctx, "sh", []string{"-c", "exit 1"}, interfaces.ProcessOptions{})
	assert.NoError(t, err)
	assert.Equal(t, 1, result.ExitCode)
	
	// Test command that doesn't exist
	result, err = adapter.Execute(ctx, "nonexistent-command", []string{}, interfaces.ProcessOptions{})
	assert.Error(t, err)
}

func TestProcessAdapter_ExecuteShell(t *testing.T) {
	adapter := NewProcessAdapter()
	ctx := context.Background()
	
	// Test non-interactive command
	result, err := adapter.ExecuteShell(ctx, "echo hello", interfaces.ProcessOptions{})
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "hello")
	
	// Test command with exit code
	result, err = adapter.ExecuteShell(ctx, "exit 2", interfaces.ProcessOptions{})
	assert.NoError(t, err)
	assert.Equal(t, 2, result.ExitCode)
}

func TestIsInteractiveCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "claude after cd",
			command:  "cd /path && claude",
			expected: true,
		},
		{
			name:     "claude with args",
			command:  "cd /path && claude --help",
			expected: true,
		},
		{
			name:     "gemini",
			command:  "cd /path && gemini",
			expected: true,
		},
		{
			name:     "vim",
			command:  "cd /path && vim file.txt",
			expected: true,
		},
		{
			name:     "non-interactive",
			command:  "cd /path && ls",
			expected: false,
		},
		{
			name:     "echo",
			command:  "echo hello",
			expected: false,
		},
		{
			name:     "standalone claude",
			command:  "claude",
			expected: true,
		},
		{
			name:     "standalone claude with args",
			command:  "claude --help",
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInteractiveCommand(tt.command)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessAdapter_OpenInBrowser(t *testing.T) {
	// Skip this test to prevent opening browser during test runs
	t.Skip("Skipping browser opening test to prevent actual browser launch")
	
	// If we want to test this without opening browser, we would need to:
	// 1. Mock exec.Command
	// 2. Or add a test mode flag to the adapter
	// For now, we skip to prevent annoying browser popups during tests
}

func TestProcessAdapter_StartBackground(t *testing.T) {
	adapter := NewProcessAdapter()
	ctx := context.Background()
	
	// Start a simple background process
	proc, err := adapter.StartBackground(ctx, "sleep", []string{"0.01"}, interfaces.ProcessOptions{})
	if err != nil {
		// Skip test if sleep command is not available
		t.Skip("sleep command not available")
	}
	
	assert.NotNil(t, proc)
	
	// Wait for it to complete
	err = proc.Wait()
	assert.NoError(t, err)
}