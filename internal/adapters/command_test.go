package adapters

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealCommandExecutor_Execute(t *testing.T) {
	executor := NewRealCommandExecutor()

	tests := []struct {
		name        string
		command     string
		args        []string
		wantErr     bool
		checkOutput func(t *testing.T, output []byte)
	}{
		{
			name:    "Simple echo command",
			command: "echo",
			args:    []string{"hello world"},
			wantErr: false,
			checkOutput: func(t *testing.T, output []byte) {
				assert.Contains(t, string(output), "hello world")
			},
		},
		{
			name:    "Command with exit code",
			command: "false",
			wantErr: true,
		},
		{
			name:    "Non-existent command",
			command: "nonexistentcommand123",
			wantErr: true,
		},
		{
			name:    "Print working directory",
			command: "pwd",
			wantErr: false,
			checkOutput: func(t *testing.T, output []byte) {
				assert.NotEmpty(t, output)
				outStr := string(output)
				assert.True(t, strings.HasPrefix(outStr, "/") || strings.HasPrefix(outStr, "C:"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.command == "false" && runtime.GOOS == "windows" {
				tt.command = "cmd"
				tt.args = []string{"/c", "exit", "1"}
			}
			if tt.command == "pwd" && runtime.GOOS == "windows" {
				tt.command = "cd"
				tt.args = []string{}
			}

			output, err := executor.Execute(tt.command, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkOutput != nil {
					tt.checkOutput(t, output)
				}
			}
		})
	}
}

func TestRealCommandExecutor_ExecuteWithInput(t *testing.T) {
	executor := NewRealCommandExecutor()

	tests := []struct {
		name        string
		input       string
		command     string
		args        []string
		wantErr     bool
		checkOutput func(t *testing.T, output []byte)
	}{
		{
			name:    "Command with stdin input",
			input:   "hello from stdin",
			command: "cat",
			args:    []string{},
			wantErr: false,
			checkOutput: func(t *testing.T, output []byte) {
				assert.Contains(t, string(output), "hello from stdin")
			},
		},
		{
			name:    "Grep with stdin",
			input:   "line1\nline2\nline3",
			command: "grep",
			args:    []string{"line2"},
			wantErr: false,
			checkOutput: func(t *testing.T, output []byte) {
				assert.Contains(t, string(output), "line2")
				assert.NotContains(t, string(output), "line1")
				assert.NotContains(t, string(output), "line3")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS == "windows" {
				if tt.command == "cat" {
					tt.command = "more"
				}
				if tt.command == "grep" {
					tt.command = "findstr"
				}
			}

			output, err := executor.ExecuteWithInput(tt.input, tt.command, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkOutput != nil {
					tt.checkOutput(t, output)
				}
			}
		})
	}
}

func TestRealCommandExecutor_ExecuteInDir(t *testing.T) {
	executor := NewRealCommandExecutor()

	t.Run("Execute in specific directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Create a test file in the directory
		testFile := "test.txt"
		err := os.WriteFile(filepath.Join(tmpDir, testFile), []byte("test content"), 0644)
		require.NoError(t, err)

		// List files in the directory
		lsCmd := "ls"
		lsArgs := []string{"-la"}
		if runtime.GOOS == "windows" {
			lsCmd = "dir"
			lsArgs = []string{}
		}

		output, err := executor.ExecuteInDir(tmpDir, lsCmd, lsArgs...)
		require.NoError(t, err)
		assert.Contains(t, string(output), testFile)
	})

	t.Run("Execute pwd in directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		pwdCmd := "pwd"
		if runtime.GOOS == "windows" {
			pwdCmd = "cd"
		}

		output, err := executor.ExecuteInDir(tmpDir, pwdCmd)
		require.NoError(t, err)
		assert.Contains(t, string(output), filepath.Base(tmpDir))
	})
}

func TestRealCommandExecutor_ExecuteWithEnv(t *testing.T) {
	executor := NewRealCommandExecutor()

	t.Run("Execute with custom environment", func(t *testing.T) {
		env := []string{"CUSTOM_VAR=test_value"}
		
		printEnvCmd := "sh"
		printEnvArgs := []string{"-c", "echo $CUSTOM_VAR"}
		if runtime.GOOS == "windows" {
			printEnvCmd = "cmd"
			printEnvArgs = []string{"/c", "echo %CUSTOM_VAR%"}
		}

		output, err := executor.ExecuteWithEnv(env, printEnvCmd, printEnvArgs...)
		require.NoError(t, err)
		assert.Contains(t, string(output), "test_value")
	})

	t.Run("Multiple environment variables", func(t *testing.T) {
		env := []string{
			"VAR1=value1",
			"VAR2=value2",
		}
		
		printEnvCmd := "sh"
		printEnvArgs := []string{"-c", "echo $VAR1:$VAR2"}
		if runtime.GOOS == "windows" {
			printEnvCmd = "cmd"
			printEnvArgs = []string{"/c", "echo %VAR1%:%VAR2%"}
		}

		output, err := executor.ExecuteWithEnv(env, printEnvCmd, printEnvArgs...)
		require.NoError(t, err)
		assert.Contains(t, string(output), "value1")
		assert.Contains(t, string(output), "value2")
	})
}

func TestRealCommandExecutor_Start(t *testing.T) {
	executor := NewRealCommandExecutor()

	t.Run("Start and wait for command", func(t *testing.T) {
		proc, err := executor.Start("echo", "async test")
		require.NoError(t, err)
		assert.NotNil(t, proc)

		err = proc.Wait()
		assert.NoError(t, err)
	})

	t.Run("Start command and get PID", func(t *testing.T) {
		sleepCmd := "sleep"
		sleepArgs := []string{"0.1"}
		if runtime.GOOS == "windows" {
			sleepCmd = "timeout"
			sleepArgs = []string{"/t", "1", "/nobreak"}
		}

		proc, err := executor.Start(sleepCmd, sleepArgs...)
		require.NoError(t, err)
		assert.NotNil(t, proc)

		pid := proc.Pid()
		assert.Greater(t, pid, 0)

		// Clean up
		proc.Kill()
	})

	t.Run("Start non-existent command", func(t *testing.T) {
		proc, err := executor.Start("nonexistentcommand456")
		assert.Error(t, err)
		assert.Nil(t, proc)
	})
}

func TestRealCommandExecutor_StartInDir(t *testing.T) {
	executor := NewRealCommandExecutor()

	t.Run("Start command in directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		pwdCmd := "pwd"
		if runtime.GOOS == "windows" {
			pwdCmd = "cd"
		}

		proc, err := executor.StartInDir(tmpDir, pwdCmd)
		require.NoError(t, err)
		assert.NotNil(t, proc)

		err = proc.Wait()
		assert.NoError(t, err)
	})
}

func TestProcessWrapper(t *testing.T) {
	executor := NewRealCommandExecutor()

	t.Run("Process operations", func(t *testing.T) {
		// Start a long-running process
		sleepCmd := "sleep"
		sleepArgs := []string{"2"}
		if runtime.GOOS == "windows" {
			sleepCmd = "timeout"
			sleepArgs = []string{"/t", "2", "/nobreak"}
		}

		proc, err := executor.Start(sleepCmd, sleepArgs...)
		require.NoError(t, err)
		assert.NotNil(t, proc)

		// Check PID
		pid := proc.Pid()
		assert.Greater(t, pid, 0)

		// Kill the process
		err = proc.Kill()
		assert.NoError(t, err)

		// Wait should return immediately after kill
		err = proc.Wait()
		// Error is expected as process was killed
		assert.Error(t, err)
	})

	t.Run("Process pipes", func(t *testing.T) {
		// Create process but don't start yet
		proc := NewProcessWrapper("echo", "test output")
		require.NotNil(t, proc)

		// Get pipes before starting
		stdout, err := proc.StdoutPipe()
		assert.NoError(t, err)
		assert.NotNil(t, stdout)

		stderr, err := proc.StderrPipe()
		assert.NoError(t, err)
		assert.NotNil(t, stderr)

		stdin, err := proc.StdinPipe()
		assert.NoError(t, err)
		assert.NotNil(t, stdin)

		// Now start the process
		err = proc.Start()
		assert.NoError(t, err)

		// Wait for completion
		err = proc.Wait()
		assert.NoError(t, err)
	})
}

func TestRealCommandExecutor_Integration(t *testing.T) {
	executor := NewRealCommandExecutor()

	t.Run("Complex command chain", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a test file
		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("Hello\nWorld\nTest"), 0644)
		require.NoError(t, err)

		// Run commands in the directory
		catCmd := "cat"
		if runtime.GOOS == "windows" {
			catCmd = "type"
		}

		output, err := executor.ExecuteInDir(tmpDir, catCmd, testFile)
		require.NoError(t, err)
		assert.Contains(t, string(output), "Hello")
		assert.Contains(t, string(output), "World")
		assert.Contains(t, string(output), "Test")
	})

	t.Run("Environment variable usage", func(t *testing.T) {
		// Set a custom environment variable
		output, err := executor.ExecuteWithEnv([]string{"MY_TEST_VAR=test123"}, "sh", "-c", "echo $MY_TEST_VAR")
		if runtime.GOOS == "windows" {
			output, err = executor.ExecuteWithEnv([]string{"MY_TEST_VAR=test123"}, "cmd", "/c", "echo %MY_TEST_VAR%")
		}

		require.NoError(t, err)
		assert.Contains(t, string(output), "test123")
	})
}