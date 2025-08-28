package git

import (
	"os"
	"os/exec"
)

// CommandRunner interface for running git commands
type CommandRunner interface {
	Run(cmd *exec.Cmd) error
	Output(cmd *exec.Cmd) ([]byte, error)
	CombinedOutput(cmd *exec.Cmd) ([]byte, error)
}

// DefaultCommandRunner uses the actual exec package
type DefaultCommandRunner struct{}

func (r *DefaultCommandRunner) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

func (r *DefaultCommandRunner) Output(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}

func (r *DefaultCommandRunner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

// FileChecker interface for checking file existence
type FileChecker interface {
	FileExists(path string) bool
	DirExists(path string) bool
}

// DefaultFileChecker uses the actual file system
type DefaultFileChecker struct{}

func (c *DefaultFileChecker) FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func (c *DefaultFileChecker) DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}