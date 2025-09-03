package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealGit_ExtendedOperations(t *testing.T) {
	git := NewRealGit()
	tmpDir := t.TempDir()

	// Initialize a git repository for testing
	repoPath := filepath.Join(tmpDir, "test-repo")
	err := os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)

	// Initialize git repo
	cmd := NewRealCommandExecutor()
	_, err = cmd.ExecuteInDir(repoPath, "git", "init")
	require.NoError(t, err)

	// Configure git user
	_, err = cmd.ExecuteInDir(repoPath, "git", "config", "user.email", "test@example.com")
	require.NoError(t, err)
	_, err = cmd.ExecuteInDir(repoPath, "git", "config", "user.name", "Test User")
	require.NoError(t, err)

	// Create initial commit
	testFile := filepath.Join(repoPath, "test.txt")
	err = os.WriteFile(testFile, []byte("initial content"), 0644)
	require.NoError(t, err)
	
	err = git.AddAll(repoPath)
	require.NoError(t, err)
	
	err = git.Commit(repoPath, "Initial commit")
	require.NoError(t, err)

	t.Run("Branch operations", func(t *testing.T) {
		// Get current branch
		branch, err := git.CurrentBranch(repoPath)
		assert.NoError(t, err)
		assert.Contains(t, []string{"main", "master"}, branch)

		// Create new branch
		err = git.CreateBranch(repoPath, "feature-branch")
		assert.NoError(t, err)

		// Checkout new branch
		err = git.CheckoutNew(repoPath, "another-branch")
		assert.NoError(t, err)

		// List branches
		branches, err := git.ListBranches(repoPath)
		assert.NoError(t, err)
		assert.Contains(t, branches, "another-branch")
		assert.Contains(t, branches, "feature-branch")

		// Delete branch
		err = git.Checkout(repoPath, branch) // Go back to main/master
		assert.NoError(t, err)
		
		err = git.DeleteBranch(repoPath, "feature-branch")
		assert.NoError(t, err)
	})

	t.Run("Tag operations", func(t *testing.T) {
		// Create tag
		err := git.Tag(repoPath, "v1.0.0")
		assert.NoError(t, err)

		// Create annotated tag
		err = git.TagWithMessage(repoPath, "v1.1.0", "Release version 1.1.0")
		assert.NoError(t, err)

		// List tags
		tags, err := git.ListTags(repoPath)
		assert.NoError(t, err)
		assert.Contains(t, tags, "v1.0.0")
		assert.Contains(t, tags, "v1.1.0")
	})

	t.Run("Reset operations", func(t *testing.T) {
		// Make a change
		err := os.WriteFile(testFile, []byte("modified content"), 0644)
		require.NoError(t, err)
		
		// Soft reset
		err = git.ResetSoft(repoPath)
		assert.NoError(t, err)

		// Add and commit the change
		err = git.AddAll(repoPath)
		assert.NoError(t, err)
		err = git.Commit(repoPath, "Modified content")
		assert.NoError(t, err)

		// Hard reset
		err = git.ResetHard(repoPath)
		assert.NoError(t, err)

		// Regular reset
		err = git.Reset(repoPath)
		assert.NoError(t, err)
	})

	t.Run("Diff operations", func(t *testing.T) {
		// Make changes
		err := os.WriteFile(testFile, []byte("new content for diff"), 0644)
		require.NoError(t, err)

		// Get diff
		diff, err := git.Diff(repoPath)
		assert.NoError(t, err)
		assert.Contains(t, diff, "new content for diff")

		// Stage changes
		err = git.AddAll(repoPath)
		assert.NoError(t, err)

		// Get staged diff
		diffStaged, err := git.DiffStaged(repoPath)
		assert.NoError(t, err)
		assert.Contains(t, diffStaged, "new content for diff")
	})

	t.Run("Log operations", func(t *testing.T) {
		// Commit the staged changes first
		err := git.Commit(repoPath, "Test commit for log")
		assert.NoError(t, err)

		// Get log
		log, err := git.Log(repoPath, 5)
		assert.NoError(t, err)
		// Log returns a slice of strings, check if any contains our commit message
		found := false
		for _, line := range log {
			if strings.Contains(line, "Test commit for log") {
				found = true
				break
			}
		}
		assert.True(t, found, "Commit message not found in log")

		// Get one-line log
		logOneline, err := git.LogOneline(repoPath, 5)
		assert.NoError(t, err)
		// Check if any line contains our commit message
		foundOneline := false
		for _, line := range logOneline {
			if strings.Contains(line, "Test commit for log") {
				foundOneline = true
				break
			}
		}
		assert.True(t, foundOneline, "Commit message not found in oneline log")
	})

	t.Run("Remote operations", func(t *testing.T) {
		// Add remote
		err := git.AddRemote(repoPath, "origin", "https://github.com/test/repo.git")
		assert.NoError(t, err)

		// List remotes
		remotes, err := git.ListRemotes(repoPath)
		assert.NoError(t, err)
		assert.Contains(t, remotes, "origin")

		// Get remote URL
		url, err := git.RemoteURL(repoPath)
		assert.NoError(t, err)
		assert.Contains(t, url, "github.com/test/repo.git")

		// Remove remote
		err = git.RemoveRemote(repoPath, "origin")
		assert.NoError(t, err)
	})

	t.Run("Status operations", func(t *testing.T) {
		// Check if has changes
		hasChanges, err := git.HasChanges(repoPath)
		assert.NoError(t, err)
		assert.False(t, hasChanges)

		// Make a change
		err = os.WriteFile(testFile, []byte("changed for status test"), 0644)
		require.NoError(t, err)

		// Check again
		hasChanges, err = git.HasChanges(repoPath)
		assert.NoError(t, err)
		assert.True(t, hasChanges)

		// Get status with options (this tests the StatusWithOptions method)
		status, err := git.StatusWithOptions(repoPath, "--short")
		assert.NoError(t, err)
		assert.Contains(t, status, " M")
	})

	t.Run("CommitWithOptions", func(t *testing.T) {
		// Stage changes
		err := git.AddAll(repoPath)
		assert.NoError(t, err)

		// Commit with options
		err = git.CommitWithOptions(repoPath, "Commit with options", "--author=\"Another User <another@example.com>\"")
		assert.NoError(t, err)

		// Verify the commit was made
		log, err := git.Log(repoPath, 1)
		assert.NoError(t, err)
		// Check if any log line contains our commit message
		found := false
		for _, line := range log {
			if strings.Contains(line, "Commit with options") {
				found = true
				break
			}
		}
		assert.True(t, found, "Commit message not found in log")
	})
}

func TestRealGit_NetworkOperations(t *testing.T) {
	t.Skip("Skipping network operations in unit tests")
	
	git := NewRealGit()
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// These would require actual network access
	t.Run("Pull operations", func(t *testing.T) {
		err := git.Pull(repoPath)
		assert.Error(t, err) // Expected to fail without proper setup

		err = git.PullWithOptions(repoPath, "--rebase")
		assert.Error(t, err)
	})

	t.Run("Push operations", func(t *testing.T) {
		err := git.Push(repoPath)
		assert.Error(t, err)

		err = git.PushWithOptions(repoPath, "--force")
		assert.Error(t, err)
	})

	t.Run("Fetch operations", func(t *testing.T) {
		err := git.Fetch(repoPath)
		assert.Error(t, err)

		err = git.FetchWithOptions(repoPath, "--all")
		assert.Error(t, err)
	})
}

func TestRealGit_DiffWithBranch(t *testing.T) {
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

	// Create initial commit on main branch
	testFile := filepath.Join(repoPath, "test.txt")
	err = os.WriteFile(testFile, []byte("main content"), 0644)
	require.NoError(t, err)
	err = git.AddAll(repoPath)
	require.NoError(t, err)
	err = git.Commit(repoPath, "Main commit")
	require.NoError(t, err)

	// Create and switch to feature branch
	err = git.CheckoutNew(repoPath, "feature")
	require.NoError(t, err)
	
	// Make changes on feature branch
	err = os.WriteFile(testFile, []byte("feature content"), 0644)
	require.NoError(t, err)
	err = git.AddAll(repoPath)
	require.NoError(t, err)
	err = git.Commit(repoPath, "Feature commit")
	require.NoError(t, err)

	// Test CurrentBranch method  
	_, err = git.CurrentBranch(repoPath)
	require.NoError(t, err)
	
	// Determine main branch name
	mainBranch := "master"
	branches, _ := git.ListBranches(repoPath)
	for _, b := range branches {
		if b == "main" {
			mainBranch = "main"
			break
		}
	}
	
	// Switch back to main branch to see the diff
	err = git.Checkout(repoPath, mainBranch)
	require.NoError(t, err)
	
	diff, err := git.DiffWithBranch(repoPath, "feature")
	assert.NoError(t, err)
	assert.Contains(t, diff, "feature content")
}