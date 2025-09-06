package adapters

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	
	"github.com/taokim/muno/internal/interfaces"
)

// ProcessAdapter implements interfaces.ProcessProvider using os/exec
type ProcessAdapter struct{}

// NewProcessAdapter creates a new process adapter
func NewProcessAdapter() *ProcessAdapter {
	return &ProcessAdapter{}
}

// Execute runs a command with arguments
func (p *ProcessAdapter) Execute(ctx context.Context, name string, args []string, opts interfaces.ProcessOptions) (*interfaces.ProcessResult, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	
	if opts.WorkingDir != "" {
		cmd.Dir = opts.WorkingDir
	}
	
	if opts.Env != nil {
		cmd.Env = append(os.Environ(), opts.Env...)
	}
	
	output, err := cmd.CombinedOutput()
	
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}
	
	return &interfaces.ProcessResult{
		ExitCode: exitCode,
		Stdout:   string(output),
		Stderr:   "",
	}, nil
}

// ExecuteShell runs a shell command
func (p *ProcessAdapter) ExecuteShell(ctx context.Context, command string, opts interfaces.ProcessOptions) (*interfaces.ProcessResult, error) {
	// For interactive commands like claude, we need to run them with stdin/stdout/stderr connected to terminal
	// Check if the command is an interactive CLI
	if isInteractiveCommand(command) {
		return p.executeInteractive(ctx, command, opts)
	}
	
	// For non-interactive commands, run normally
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	
	if opts.WorkingDir != "" {
		cmd.Dir = opts.WorkingDir
	}
	
	if opts.Env != nil {
		cmd.Env = append(os.Environ(), opts.Env...)
	}
	
	output, err := cmd.CombinedOutput()
	
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}
	
	return &interfaces.ProcessResult{
		ExitCode: exitCode,
		Stdout:   string(output),
		Stderr:   "",
	}, nil
}

// executeInteractive runs an interactive command with terminal attached
func (p *ProcessAdapter) executeInteractive(ctx context.Context, command string, opts interfaces.ProcessOptions) (*interfaces.ProcessResult, error) {
	// Run the command interactively
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	
	if opts.WorkingDir != "" {
		cmd.Dir = opts.WorkingDir
	}
	
	if opts.Env != nil {
		cmd.Env = append(os.Environ(), opts.Env...)
	}
	
	// Connect to terminal for interactive commands
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Run the command
	err := cmd.Run()
	
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}
	
	return &interfaces.ProcessResult{
		ExitCode: exitCode,
		Stdout:   "",
		Stderr:   "",
	}, nil
}

// isInteractiveCommand checks if a command should be run interactively
func isInteractiveCommand(command string) bool {
	// List of known interactive commands
	interactiveCommands := []string{
		"claude",
		"gemini",
		"cursor",
		"windsurf",
		"aider",
		"vim",
		"vi",
		"nano",
		"emacs",
		"less",
		"more",
		"top",
		"htop",
	}
	
	// Check if the command contains any of the interactive commands
	for _, ic := range interactiveCommands {
		// Check for the command after "cd ... && <command>"
		if strings.Contains(command, "&& "+ic) {
			return true
		}
		// Check for command with arguments
		if strings.Contains(command, "&& "+ic+" ") {
			return true
		}
		// Check if it's the whole command
		if command == ic || strings.HasPrefix(command, ic+" ") {
			return true
		}
	}
	
	return false
}

// StartBackground starts a command in the background
func (p *ProcessAdapter) StartBackground(ctx context.Context, name string, args []string, opts interfaces.ProcessOptions) (interfaces.Process, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	
	if opts.WorkingDir != "" {
		cmd.Dir = opts.WorkingDir
	}
	
	if opts.Env != nil {
		cmd.Env = append(os.Environ(), opts.Env...)
	}
	
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	
	return &ProcessWrapper{cmd}, nil
}

// OpenInEditor opens a file in the default editor
func (p *ProcessAdapter) OpenInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // Default to vi if no editor is set
	}
	
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// OpenInBrowser opens a URL in the default browser
func (p *ProcessAdapter) OpenInBrowser(url string) error {
	var cmd *exec.Cmd
	
	switch os := getOS(); os {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", os)
	}
	
	return cmd.Start()
}

// getOS returns the operating system
func getOS() string {
	// This is a simplified version. In production, use runtime.GOOS
	if _, err := exec.LookPath("open"); err == nil {
		return "darwin"
	}
	if _, err := exec.LookPath("xdg-open"); err == nil {
		return "linux"
	}
	return "windows"
}

