package git

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultCommandRunner(t *testing.T) {
	runner := &DefaultCommandRunner{}

	t.Run("Run command", func(t *testing.T) {
		// Use a simple command that should work on all systems
		cmd := exec.Command("echo", "test")
		err := runner.Run(cmd)
		assert.NoError(t, err)
	})

	t.Run("Run invalid command", func(t *testing.T) {
		cmd := exec.Command("/non/existent/command")
		err := runner.Run(cmd)
		assert.Error(t, err)
	})

	t.Run("Output command", func(t *testing.T) {
		cmd := exec.Command("echo", "test")
		output, err := runner.Output(cmd)
		assert.NoError(t, err)
		assert.Contains(t, string(output), "test")
	})

	t.Run("Output invalid command", func(t *testing.T) {
		cmd := exec.Command("/non/existent/command")
		output, err := runner.Output(cmd)
		assert.Error(t, err)
		assert.Nil(t, output)
	})

	t.Run("CombinedOutput command", func(t *testing.T) {
		cmd := exec.Command("echo", "test")
		output, err := runner.CombinedOutput(cmd)
		assert.NoError(t, err)
		assert.Contains(t, string(output), "test")
	})

	t.Run("CombinedOutput invalid command", func(t *testing.T) {
		cmd := exec.Command("/non/existent/command")
		_, err := runner.CombinedOutput(cmd)
		assert.Error(t, err)
		// CombinedOutput may return empty output on error for non-existent command
		// The behavior depends on the OS and command
	})
}

func TestDefaultFileChecker(t *testing.T) {
	checker := &DefaultFileChecker{}

	t.Run("FileExists with existing file", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		exists := checker.FileExists(tmpFile.Name())
		assert.True(t, exists)
	})

	t.Run("FileExists with non-existent file", func(t *testing.T) {
		exists := checker.FileExists("/non/existent/file")
		assert.False(t, exists)
	})

	t.Run("FileExists with directory", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Directory should not be considered a file
		exists := checker.FileExists(tmpDir)
		assert.False(t, exists)
	})

	t.Run("DirExists with existing directory", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		exists := checker.DirExists(tmpDir)
		assert.True(t, exists)
	})

	t.Run("DirExists with non-existent directory", func(t *testing.T) {
		exists := checker.DirExists("/non/existent/directory")
		assert.False(t, exists)
	})

	t.Run("DirExists with file", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		// File should not be considered a directory
		exists := checker.DirExists(tmpFile.Name())
		assert.False(t, exists)
	})
}

// Mock implementations for testing
type MockCommandRunner struct {
	RunFunc           func(*exec.Cmd) error
	OutputFunc        func(*exec.Cmd) ([]byte, error)
	CombinedOutputFunc func(*exec.Cmd) ([]byte, error)
}

func (m *MockCommandRunner) Run(cmd *exec.Cmd) error {
	if m.RunFunc != nil {
		return m.RunFunc(cmd)
	}
	return nil
}

func (m *MockCommandRunner) Output(cmd *exec.Cmd) ([]byte, error) {
	if m.OutputFunc != nil {
		return m.OutputFunc(cmd)
	}
	return []byte("mock output"), nil
}

func (m *MockCommandRunner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	if m.CombinedOutputFunc != nil {
		return m.CombinedOutputFunc(cmd)
	}
	return []byte("mock combined output"), nil
}

type MockFileChecker struct {
	FileExistsFunc func(string) bool
	DirExistsFunc  func(string) bool
}

func (m *MockFileChecker) FileExists(path string) bool {
	if m.FileExistsFunc != nil {
		return m.FileExistsFunc(path)
	}
	return false
}

func (m *MockFileChecker) DirExists(path string) bool {
	if m.DirExistsFunc != nil {
		return m.DirExistsFunc(path)
	}
	return false
}

func TestInterfaceImplementation(t *testing.T) {
	// Ensure our types implement the interfaces
	var _ CommandRunner = (*DefaultCommandRunner)(nil)
	var _ CommandRunner = (*MockCommandRunner)(nil)
	var _ FileChecker = (*DefaultFileChecker)(nil)
	var _ FileChecker = (*MockFileChecker)(nil)

	// Test mock implementations
	t.Run("MockCommandRunner", func(t *testing.T) {
		mock := &MockCommandRunner{
			RunFunc: func(cmd *exec.Cmd) error {
				if cmd.Path == "/fail" {
					return errors.New("mock error")
				}
				return nil
			},
		}

		cmd := exec.Command("test")
		err := mock.Run(cmd)
		assert.NoError(t, err)

		cmd = exec.Command("/fail")
		err = mock.Run(cmd)
		assert.Error(t, err)
	})

	t.Run("MockFileChecker", func(t *testing.T) {
		mock := &MockFileChecker{
			FileExistsFunc: func(path string) bool {
				return path == "/exists"
			},
			DirExistsFunc: func(path string) bool {
				return path == "/dir"
			},
		}

		assert.True(t, mock.FileExists("/exists"))
		assert.False(t, mock.FileExists("/not-exists"))
		assert.True(t, mock.DirExists("/dir"))
		assert.False(t, mock.DirExists("/not-dir"))
	})
}