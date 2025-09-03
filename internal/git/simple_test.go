package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit_Simple(t *testing.T) {
	g := New()
	assert.NotNil(t, g)

	t.Run("Clone", func(t *testing.T) {
		tmpDir := t.TempDir()
		repoPath := filepath.Join(tmpDir, "test-repo")
		
		// Try to clone a non-existent repo (will fail)
		err := g.Clone("https://github.com/non-existent/repo.git", repoPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "git clone failed")
	})

	t.Run("Pull", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Initialize a git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		err := cmd.Run()
		require.NoError(t, err)
		
		// Try to pull (will fail because no remote)
		err = g.Pull(tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "git pull failed")
	})

	t.Run("Add", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Initialize a git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		err := cmd.Run()
		require.NoError(t, err)
		
		// Create a file
		testFile := filepath.Join(tmpDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)
		
		// Add the file
		err = g.Add(tmpDir, "test.txt")
		assert.NoError(t, err)
		
		// Add all files
		err = g.Add(tmpDir, ".")
		assert.NoError(t, err)
	})

	t.Run("Commit", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Initialize a git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		err := cmd.Run()
		require.NoError(t, err)
		
		// Try to commit with no changes
		err = g.Commit(tmpDir, "Empty commit")
		assert.Error(t, err)
		// The error message might be in different languages, just check for error
		
		// Create and add a file
		testFile := filepath.Join(tmpDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)
		
		err = g.Add(tmpDir, ".")
		require.NoError(t, err)
		
		// Now commit should work
		err = g.Commit(tmpDir, "Initial commit")
		assert.NoError(t, err)
	})

	t.Run("Push", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Initialize a git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		err := cmd.Run()
		require.NoError(t, err)
		
		// Try to push (will fail because no remote)
		err = g.Push(tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "git push failed")
	})

	t.Run("Status", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Initialize a git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		err := cmd.Run()
		require.NoError(t, err)
		
		// Get status of empty repo
		status, err := g.Status(tmpDir)
		assert.NoError(t, err)
		assert.Empty(t, status)
		
		// Create a file
		testFile := filepath.Join(tmpDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)
		
		// Get status with untracked file
		status, err = g.Status(tmpDir)
		assert.NoError(t, err)
		assert.Contains(t, status, "test.txt")
	})

	t.Run("Status with non-git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Try to get status of non-git directory
		status, err := g.Status(tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "git status failed")
		assert.Empty(t, status)
	})
}

func TestGit_GetRemotes(t *testing.T) {
	g := &Git{}

	t.Run("GetRemotes with valid repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Initialize a git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		err := cmd.Run()
		require.NoError(t, err)
		
		// Add a remote
		cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
		cmd.Dir = tmpDir
		err = cmd.Run()
		require.NoError(t, err)
		
		// Get remotes
		remotes, err := g.GetRemotes(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, remotes)
		assert.Equal(t, "https://github.com/test/repo.git", remotes["origin"])
	})

	t.Run("GetRemotes with no remotes", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Initialize a git repo without remotes
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		err := cmd.Run()
		require.NoError(t, err)
		
		// Get remotes
		remotes, err := g.GetRemotes(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, remotes)
		assert.Empty(t, remotes)
	})

	t.Run("GetRemotes with non-git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Get remotes should fail
		remotes, err := g.GetRemotes(tmpDir)
		assert.Error(t, err)
		assert.Nil(t, remotes)
	})
}

func TestGit_CurrentBranch(t *testing.T) {
	g := &Git{}

	t.Run("CurrentBranch with valid repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Initialize a git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		err := cmd.Run()
		require.NoError(t, err)
		
		// Create an initial commit to establish a branch
		testFile := filepath.Join(tmpDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)
		
		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		err = cmd.Run()
		require.NoError(t, err)
		
		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = tmpDir
		err = cmd.Run()
		require.NoError(t, err)
		
		// Get current branch
		branch, err := g.CurrentBranch(tmpDir)
		assert.NoError(t, err)
		// On new repos, the default branch could be master or main
		assert.Contains(t, []string{"master", "main"}, branch)
	})

	t.Run("CurrentBranch with detached HEAD", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Initialize a git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		err := cmd.Run()
		require.NoError(t, err)
		
		// Get current branch before any commits
		branch, err := g.CurrentBranch(tmpDir)
		// This might error or return HEAD
		if err == nil {
			assert.NotEmpty(t, branch)
		}
	})

	t.Run("CurrentBranch with non-git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Get current branch should fail
		branch, err := g.CurrentBranch(tmpDir)
		assert.Error(t, err)
		assert.Empty(t, branch)
	})
}