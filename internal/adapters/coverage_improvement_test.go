package adapters

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealOutput_ColoredMethods(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewRealOutput(&stdout, &stderr)

	t.Run("Success method", func(t *testing.T) {
		stdout.Reset()
		output.Success("Operation completed")
		assert.Contains(t, stdout.String(), "✅")
		assert.Contains(t, stdout.String(), "Operation completed")
	})

	t.Run("Info method", func(t *testing.T) {
		stdout.Reset()
		output.Info("Information message")
		assert.Contains(t, stdout.String(), "ℹ️")
		assert.Contains(t, stdout.String(), "Information message")
	})

	t.Run("Warning method", func(t *testing.T) {
		stdout.Reset()
		output.Warning("Warning message")
		assert.Contains(t, stdout.String(), "⚠️")
		assert.Contains(t, stdout.String(), "Warning message")
	})

	t.Run("Danger method", func(t *testing.T) {
		stderr.Reset()
		output.Danger("Danger message")
		assert.Contains(t, stderr.String(), "🚨")
		assert.Contains(t, stderr.String(), "Danger message")
	})
}

func TestRealOutput_FormattingMethods(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewRealOutput(&stdout, &stderr)

	t.Run("Bold method", func(t *testing.T) {
		result := output.Bold("bold text")
		assert.Equal(t, "bold text", result) // Currently just returns the text
	})

	t.Run("Italic method", func(t *testing.T) {
		result := output.Italic("italic text")
		assert.Equal(t, "italic text", result)
	})

	t.Run("Underline method", func(t *testing.T) {
		result := output.Underline("underlined text")
		assert.Equal(t, "underlined text", result)
	})

	t.Run("Color method", func(t *testing.T) {
		result := output.Color("colored text", "red")
		assert.Equal(t, "colored text", result)
	})
}

func TestRealOutput_WriteMethod(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := NewRealOutput(&stdout, &stderr)

	t.Run("Write method", func(t *testing.T) {
		n, err := output.Write([]byte("test write"))
		assert.NoError(t, err)
		assert.Equal(t, 10, n)
		assert.Equal(t, "test write", stdout.String())
	})
}

func TestRealFileWrapper_Methods(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	
	// Create a test file
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	fs := NewRealFileSystem()
	
	t.Run("File operations", func(t *testing.T) {
		// Open file
		file, err := fs.Open(testFile)
		require.NoError(t, err)
		defer file.Close()

		// Read from file
		buf := make([]byte, 12)
		n, err := file.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, 12, n)
		assert.Equal(t, "test content", string(buf))

		// Seek in file
		pos, err := file.Seek(0, 0)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), pos)

		// Get file name
		name := file.Name()
		assert.Equal(t, testFile, name)

		// Get file stat
		info, err := file.Stat()
		assert.NoError(t, err)
		assert.Equal(t, "test.txt", info.Name())

		// Sync file
		err = file.Sync()
		assert.NoError(t, err)
	})

	t.Run("Write operations", func(t *testing.T) {
		writeFile := filepath.Join(tmpDir, "write.txt")
		file, err := fs.Create(writeFile)
		require.NoError(t, err)
		defer file.Close()

		// Write to file
		n, err := file.Write([]byte("written content"))
		assert.NoError(t, err)
		assert.Equal(t, 15, n)

		// Truncate file
		err = file.Truncate(10)
		assert.NoError(t, err)
	})
}

func TestProcessWrapper_AdditionalMethods(t *testing.T) {
	executor := NewRealCommandExecutor()
	
	t.Run("Signal method", func(t *testing.T) {
		// Start a long-running process
		proc, err := executor.Start("sleep", "10")
		require.NoError(t, err)
		
		// Test getting PID
		pid := proc.Pid()
		assert.Greater(t, pid, 0)
		
		// Send interrupt signal
		err = proc.Signal(os.Interrupt)
		assert.NoError(t, err)
		
		// Wait should return an error since we interrupted it
		err = proc.Wait()
		assert.Error(t, err)
	})

	t.Run("Kill already terminated process", func(t *testing.T) {
		// Start a quick process
		proc, err := executor.Start("echo", "test")
		require.NoError(t, err)
		
		// Wait for it to finish
		err = proc.Wait()
		assert.NoError(t, err)
		
		// Kill should handle gracefully
		err = proc.Kill()
		// Could be nil or error depending on OS behavior
		_ = err
	})
}

func TestRealCommandExecutor_StartInDir_Additional(t *testing.T) {
	executor := NewRealCommandExecutor()
	tmpDir := t.TempDir()
	
	t.Run("StartInDir with valid directory", func(t *testing.T) {
		// Create a test file in tmpDir
		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)
		
		// Start ls command in tmpDir - need to get stdout before starting
		proc, err := executor.StartInDir(tmpDir, "ls")
		require.NoError(t, err)
		
		// Process should be started, just wait for it
		err = proc.Wait()
		// ls should complete successfully
		assert.NoError(t, err)
	})
}

func TestGitBranchMethod(t *testing.T) {
	git := NewRealGit()
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	
	// Setup repo
	err := os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)
	
	cmd := NewRealCommandExecutor()
	_, err = cmd.ExecuteInDir(repoPath, "git", "init")
	require.NoError(t, err)
	_, err = cmd.ExecuteInDir(repoPath, "git", "config", "user.email", "test@example.com")
	require.NoError(t, err)
	_, err = cmd.ExecuteInDir(repoPath, "git", "config", "user.name", "Test User")
	require.NoError(t, err)
	
	// Create initial commit
	testFile := filepath.Join(repoPath, "test.txt")
	err = os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)
	err = git.AddAll(repoPath)
	require.NoError(t, err)
	err = git.Commit(repoPath, "Initial")
	require.NoError(t, err)
	
	// Test Branch method (which has 0% coverage)
	branch, err := git.Branch(repoPath)
	assert.NoError(t, err)
	// git branch returns output like "* master\n" or "* main\n"
	assert.True(t, strings.Contains(branch, "master") || strings.Contains(branch, "main"))
}