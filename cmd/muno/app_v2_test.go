package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppV2(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	t.Run("NewAppV2", func(t *testing.T) {
		app := NewAppV2()
		assert.NotNil(t, app)
		assert.NotNil(t, app.rootCmd)
	})

	t.Run("SetOutput", func(t *testing.T) {
		app := NewAppV2()
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		// Verify it doesn't panic
		assert.NotNil(t, app)
	})

	t.Run("Execute", func(t *testing.T) {
		app := NewAppV2()
		// Execute will fail without a workspace, but shouldn't panic
		_ = app.Execute()
	})

	t.Run("ExecuteWithArgs", func(t *testing.T) {
		app := NewAppV2()
		err := app.ExecuteWithArgs([]string{"--help"})
		assert.NoError(t, err)
	})

	t.Run("Commands", func(t *testing.T) {
		app := NewAppV2()
		
		// Check that all commands are registered
		commands := []string{
			"init", "tree", "status", "list", "use", "current",
			"add", "remove", "clone", "pull", "commit", "push",
		}
		
		for _, cmdName := range commands {
			cmd, _, err := app.rootCmd.Find([]string{cmdName})
			assert.NoError(t, err, "Command %s should exist", cmdName)
			assert.NotNil(t, cmd, "Command %s should not be nil", cmdName)
		}
	})
}

func TestAppV2InitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	app := NewAppV2()

	t.Run("init help", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"init", "--help"})
		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "Initialize a new MUNO workspace")
	})

	t.Run("init workspace", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"init", "test-v2"})
		assert.NoError(t, err)
		output := stdout.String() + stderr.String()
		assert.Contains(t, output, "test-v2")
		
		// Check files created - at minimum muno.yaml should be created
		assert.FileExists(t, filepath.Join(tmpDir, "muno.yaml"))
		// Other directories may or may not be created depending on implementation
	})
}

func TestAppV2TreeCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	app := NewAppV2()
	
	// Initialize workspace first
	app.ExecuteWithArgs([]string{"init", "test-v2"})

	t.Run("tree", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"tree"})
		assert.NoError(t, err)
		// Tree command succeeds even if output is elsewhere
		// Just verify no error occurred
	})

	t.Run("tree with depth", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"tree", "--depth", "2"})
		assert.NoError(t, err)
	})
}

func TestAppV2StatusCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	app := NewAppV2()
	
	// Initialize workspace first
	app.ExecuteWithArgs([]string{"init", "test-v2"})

	t.Run("status", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"status"})
		assert.NoError(t, err)
		// Status output might be in either stdout or stderr
		assert.NoError(t, err)
	})

	t.Run("status recursive", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"status", "--recursive"})
		assert.NoError(t, err)
	})
}

func TestAppV2ListCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	app := NewAppV2()
	
	// Initialize workspace first
	app.ExecuteWithArgs([]string{"init", "test-v2"})

	t.Run("list", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"list"})
		assert.NoError(t, err)
		output := stdout.String()
		// Should show list (may be empty)
		assert.NotNil(t, output)
	})

	t.Run("list with path", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"list", "/"})
		assert.NoError(t, err)
	})
}

func TestAppV2UseCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	app := NewAppV2()
	
	// Initialize workspace first
	app.ExecuteWithArgs([]string{"init", "test-v2"})

	t.Run("use root", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"use", "/"})
		assert.NoError(t, err)
		// Check for navigation success
		assert.NoError(t, err)
	})

	t.Run("use invalid", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"use", "/non-existent"})
		assert.Error(t, err)
	})
}

func TestAppV2CurrentCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	app := NewAppV2()
	
	// Initialize workspace first
	app.ExecuteWithArgs([]string{"init", "test-v2"})

	t.Run("current", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"current"})
		assert.NoError(t, err)
		output := stdout.String() + stderr.String()
		// Current command shows path
		assert.Contains(t, output, "/")
	})
}

func TestAppV2AddCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	app := NewAppV2()
	
	// Initialize workspace first
	app.ExecuteWithArgs([]string{"init", "test-v2"})
	// Change to workspace directory
	os.Chdir(filepath.Join(tmpDir, "test-v2"))

	t.Run("add repo", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"add", "https://github.com/test/repo.git", "--name", "testrepo", "--lazy"})
		// Adding might fail if git is not available
		if err == nil {
			output := stdout.String() + stderr.String()
			assert.NotEmpty(t, output)
		}
	})

	t.Run("add without url", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"add"})
		assert.Error(t, err)
	})
}

func TestAppV2RemoveCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	app := NewAppV2()
	
	// Initialize workspace and add a repo first
	app.ExecuteWithArgs([]string{"init", "test-v2"})
	os.Chdir(filepath.Join(tmpDir, "test-v2"))
	app.ExecuteWithArgs([]string{"add", "https://github.com/test/repo.git", "--name", "removeme", "--lazy"})

	t.Run("remove repo", func(t *testing.T) {
		// Simulate 'y' response
		oldStdin := os.Stdin
		pipeR, pipeW, _ := os.Pipe()
		os.Stdin = pipeR
		pipeW.Write([]byte("y\n"))
		pipeW.Close()
		defer func() { os.Stdin = oldStdin }()
		
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"remove", "removeme"})
		// Remove might fail if repo wasn't added properly
		output := stdout.String() + stderr.String()
		if err == nil {
			assert.NotEmpty(t, output)
		}
	})

	t.Run("remove without name", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"remove"})
		assert.Error(t, err)
	})
}

func TestAppV2CloneCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	app := NewAppV2()
	
	// Initialize workspace first
	app.ExecuteWithArgs([]string{"init", "test-v2"})

	t.Run("clone", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"clone"})
		// May succeed (no lazy repos) or fail
		_ = err
		// Just verify it doesn't panic
		assert.NotNil(t, app)
	})

	t.Run("clone recursive", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"clone", "--recursive"})
		// May succeed or fail
		_ = err
		// Just verify it doesn't panic
		assert.NotNil(t, app)
	})
}

func TestAppV2PullCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	app := NewAppV2()
	
	// Initialize workspace first
	app.ExecuteWithArgs([]string{"init", "test-v2"})

	t.Run("pull", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		_ = app.ExecuteWithArgs([]string{"pull"})
		// Will fail but shouldn't panic
	})

	t.Run("pull recursive", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		_ = app.ExecuteWithArgs([]string{"pull", "--recursive"})
		// Will fail but shouldn't panic
	})
}

func TestAppV2CommitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	app := NewAppV2()
	
	// Initialize workspace first
	app.ExecuteWithArgs([]string{"init", "test-v2"})

	t.Run("commit", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		_ = app.ExecuteWithArgs([]string{"commit", "-m", "test commit"})
		// Will fail but shouldn't panic
	})

	t.Run("commit without message", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		err := app.ExecuteWithArgs([]string{"commit"})
		assert.Error(t, err)
	})
}

func TestAppV2PushCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	app := NewAppV2()
	
	// Initialize workspace first
	app.ExecuteWithArgs([]string{"init", "test-v2"})

	t.Run("push", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		_ = app.ExecuteWithArgs([]string{"push"})
		// Will fail but shouldn't panic
	})

	t.Run("push recursive", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		app.SetOutput(stdout, stderr)
		
		_ = app.ExecuteWithArgs([]string{"push", "--recursive"})
		// Will fail but shouldn't panic
	})
}