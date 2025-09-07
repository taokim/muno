package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/mocks"
)

func setupTestRepo(t *testing.T) (string, *RealGit) {
	tmpDir := t.TempDir()
	
	// Initialize a git repo
	cmd := NewRealCommandExecutor()
	_, err := cmd.ExecuteInDir(tmpDir, "git", "init")
	require.NoError(t, err)
	
	// Configure git user for commits
	_, err = cmd.ExecuteInDir(tmpDir, "git", "config", "user.email", "test@example.com")
	require.NoError(t, err)
	_, err = cmd.ExecuteInDir(tmpDir, "git", "config", "user.name", "Test User")
	require.NoError(t, err)
	
	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("initial content"), 0644)
	require.NoError(t, err)
	
	_, err = cmd.ExecuteInDir(tmpDir, "git", "add", ".")
	require.NoError(t, err)
	_, err = cmd.ExecuteInDir(tmpDir, "git", "commit", "-m", "Initial commit")
	require.NoError(t, err)
	
	return tmpDir, NewRealGit()
}

func TestRealGit_Clone(t *testing.T) {
	git := NewRealGit()
	tmpDir := t.TempDir()
	
	// We can't test real clone without a remote repo, but we can test error handling
	t.Run("Clone non-existent repo", func(t *testing.T) {
		dest := filepath.Join(tmpDir, "cloned")
		err := git.Clone("https://github.com/nonexistent/repo.git", dest)
		assert.Error(t, err)
	})
	
	t.Run("Clone with mock executor", func(t *testing.T) {
		mockCmd := &mocks.MockCommandExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "git" && len(args) > 0 && args[0] == "clone" {
					return []byte("Cloning into 'repo'..."), nil
				}
				return nil, nil
			},
		}
		
		git := NewRealGitWithExecutor(mockCmd)
		err := git.Clone("https://github.com/test/repo.git", filepath.Join(tmpDir, "test"))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(mockCmd.Calls))
	})
}

func TestRealGit_CloneWithOptions(t *testing.T) {
	t.Run("Clone with branch", func(t *testing.T) {
		mockCmd := &mocks.MockCommandExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "git" && len(args) > 2 {
					// Check that branch option is included
					for i, arg := range args {
						if arg == "--branch" && i+1 < len(args) {
							assert.Equal(t, "develop", args[i+1])
							return []byte("Cloning with branch..."), nil
						}
					}
				}
				return nil, nil
			},
		}
		
		git := NewRealGitWithExecutor(mockCmd)
		err := git.CloneWithOptions("https://github.com/test/repo.git", 
			filepath.Join(t.TempDir(), "test"), "--branch", "develop")
		assert.NoError(t, err)
	})
	
	t.Run("Clone with depth", func(t *testing.T) {
		mockCmd := &mocks.MockCommandExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "git" && len(args) > 2 {
					// Check that depth option is included
					for i, arg := range args {
						if arg == "--depth" && i+1 < len(args) {
							assert.Equal(t, "5", args[i+1])
							return []byte("Cloning with depth..."), nil
						}
					}
				}
				return nil, nil
			},
		}
		
		git := NewRealGitWithExecutor(mockCmd)
		err := git.CloneWithOptions("https://github.com/test/repo.git",
			filepath.Join(t.TempDir(), "test"), "--depth", "5")
		assert.NoError(t, err)
	})
}

func TestRealGit_Status(t *testing.T) {
	repoDir, git := setupTestRepo(t)
	
	t.Run("Clean repository", func(t *testing.T) {
		status, err := git.Status(repoDir)
		require.NoError(t, err)
		// Debug output
		t.Logf("Git status output: %q", status)
		assert.True(t, 
			strings.Contains(status, "nothing to commit") ||
			strings.Contains(status, "working tree clean") ||
			strings.Contains(status, "커밋할 사항 없음") ||
			strings.Contains(status, "작업 트리 깨끗함") ||
			strings.Contains(status, "작업 폴더 깨끗함"))
	})
	
	t.Run("Modified file", func(t *testing.T) {
		testFile := filepath.Join(repoDir, "test.txt")
		err := os.WriteFile(testFile, []byte("modified content"), 0644)
		require.NoError(t, err)
		
		status, err := git.Status(repoDir)
		require.NoError(t, err)
		assert.True(t,
			strings.Contains(status, "modified") || strings.Contains(status, "test.txt"))
	})
	
	t.Run("New file", func(t *testing.T) {
		newFile := filepath.Join(repoDir, "new.txt")
		err := os.WriteFile(newFile, []byte("new content"), 0644)
		require.NoError(t, err)
		
		status, err := git.Status(repoDir)
		require.NoError(t, err)
		assert.True(t,
			strings.Contains(status, "new.txt") || strings.Contains(status, "Untracked"))
	})
}

func TestRealGit_Add(t *testing.T) {
	repoDir, git := setupTestRepo(t)
	
	t.Run("Add specific file", func(t *testing.T) {
		newFile := filepath.Join(repoDir, "new.txt")
		err := os.WriteFile(newFile, []byte("new content"), 0644)
		require.NoError(t, err)
		
		err = git.Add(repoDir, "new.txt")
		require.NoError(t, err)
		
		status, err := git.Status(repoDir)
		require.NoError(t, err)
		assert.True(t,
			strings.Contains(status, "new file") || 
			strings.Contains(status, "Changes to be committed") ||
			strings.Contains(status, "새 파일") ||
			strings.Contains(status, "커밋할 변경 사항"))
	})
	
	t.Run("Add all files", func(t *testing.T) {
		// Create multiple files
		for i := 0; i < 3; i++ {
			file := filepath.Join(repoDir, "file"+string(rune('0'+i))+".txt")
			err := os.WriteFile(file, []byte("content"), 0644)
			require.NoError(t, err)
		}
		
		err := git.AddAll(repoDir)
		require.NoError(t, err)
		
		status, err := git.Status(repoDir)
		require.NoError(t, err)
		assert.True(t,
			strings.Contains(status, "new file") || 
			strings.Contains(status, "Changes to be committed") ||
			strings.Contains(status, "새 파일") ||
			strings.Contains(status, "커밋할 변경 사항"))
	})
}

func TestRealGit_Commit(t *testing.T) {
	repoDir, git := setupTestRepo(t)
	
	t.Run("Commit with message", func(t *testing.T) {
		// Make a change
		testFile := filepath.Join(repoDir, "commit-test.txt")
		err := os.WriteFile(testFile, []byte("commit test"), 0644)
		require.NoError(t, err)
		
		err = git.Add(repoDir, ".")
		require.NoError(t, err)
		
		err = git.Commit(repoDir, "Test commit message")
		require.NoError(t, err)
		
		// Check log contains the commit
		log, err := git.Log(repoDir, 1)
		require.NoError(t, err)
		assert.True(t, len(log) > 0, "Log should have at least one commit")
		found := false
		for _, entry := range log {
			if strings.Contains(entry, "Test commit message") {
				found = true
				break
			}
		}
		assert.True(t, found, "Log should contain the commit message")
	})
	
	t.Run("Commit with nothing staged", func(t *testing.T) {
		err := git.Commit(repoDir, "Empty commit")
		// This should error because nothing to commit
		assert.Error(t, err)
	})
}

func TestRealGit_Branch(t *testing.T) {
	repoDir, git := setupTestRepo(t)
	
	t.Run("Create branch", func(t *testing.T) {
		err := git.CreateBranch(repoDir, "feature-branch")
		require.NoError(t, err)
		
		branches, err := git.ListBranches(repoDir)
		require.NoError(t, err)
		assert.Contains(t, branches, "feature-branch")
	})
	
	t.Run("Switch branch", func(t *testing.T) {
		err := git.CreateBranch(repoDir, "develop")
		require.NoError(t, err)
		
		err = git.Checkout(repoDir, "develop")
		require.NoError(t, err)
		
		branch, err := git.CurrentBranch(repoDir)
		require.NoError(t, err)
		assert.Equal(t, "develop", strings.TrimSpace(branch))
	})
	
	t.Run("Current branch", func(t *testing.T) {
		branch, err := git.CurrentBranch(repoDir)
		require.NoError(t, err)
		assert.NotEmpty(t, branch)
	})
	
	t.Run("List branches", func(t *testing.T) {
		branches, err := git.ListBranches(repoDir)
		require.NoError(t, err)
		// branches is a slice, check if it contains the branch
		hasMasterOrMain := false
		for _, branch := range branches {
			if strings.Contains(branch, "master") || strings.Contains(branch, "main") {
				hasMasterOrMain = true
				break
			}
		}
		assert.True(t, hasMasterOrMain)
	})
	
	t.Run("Delete branch", func(t *testing.T) {
		// Create and switch away from branch to delete
		err := git.CreateBranch(repoDir, "to-delete")
		require.NoError(t, err)
		
		err = git.Checkout(repoDir, "master")
		if err != nil {
			err = git.Checkout(repoDir, "main")
		}
		require.NoError(t, err)
		
		err = git.DeleteBranch(repoDir, "to-delete")
		require.NoError(t, err)
		
		branches, err := git.ListBranches(repoDir)
		require.NoError(t, err)
		assert.NotContains(t, branches, "to-delete")
	})
}

func TestRealGit_Remote(t *testing.T) {
	repoDir, git := setupTestRepo(t)
	
	t.Run("Add remote", func(t *testing.T) {
		err := git.AddRemote(repoDir, "origin", "https://github.com/test/repo.git")
		require.NoError(t, err)
		
		remotes, err := git.ListRemotes(repoDir)
		require.NoError(t, err)
		assert.Contains(t, remotes, "origin")
	})
	
	t.Run("List remotes", func(t *testing.T) {
		remotes, err := git.ListRemotes(repoDir)
		require.NoError(t, err)
		// Should have origin from previous test
		assert.Contains(t, remotes, "origin")
	})
	
	t.Run("Remove remote", func(t *testing.T) {
		err := git.RemoveRemote(repoDir, "origin")
		require.NoError(t, err)
		
		remotes, err := git.ListRemotes(repoDir)
		require.NoError(t, err)
		assert.NotContains(t, remotes, "origin")
	})
}

func TestRealGit_Diff(t *testing.T) {
	repoDir, git := setupTestRepo(t)
	
	t.Run("Diff unstaged changes", func(t *testing.T) {
		testFile := filepath.Join(repoDir, "test.txt")
		err := os.WriteFile(testFile, []byte("modified content\nnew line"), 0644)
		require.NoError(t, err)
		
		diff, err := git.Diff(repoDir)
		require.NoError(t, err)
		assert.True(t,
			strings.Contains(diff, "modified content") || strings.Contains(diff, "@@"))
	})
	
	t.Run("Diff staged changes", func(t *testing.T) {
		testFile := filepath.Join(repoDir, "staged.txt")
		err := os.WriteFile(testFile, []byte("staged content"), 0644)
		require.NoError(t, err)
		
		err = git.Add(repoDir, "staged.txt")
		require.NoError(t, err)
		
		diff, err := git.DiffStaged(repoDir)
		require.NoError(t, err)
		assert.True(t,
			strings.Contains(diff, "staged") || strings.Contains(diff, "+++"))
	})
}

func TestRealGit_Log(t *testing.T) {
	repoDir, git := setupTestRepo(t)
	
	t.Run("Get log", func(t *testing.T) {
		logLines, err := git.Log(repoDir, 10)
		require.NoError(t, err)
		// Check if any line contains "Initial commit"
		foundInitialCommit := false
		for _, line := range logLines {
			if strings.Contains(line, "Initial commit") {
				foundInitialCommit = true
				break
			}
		}
		assert.True(t, foundInitialCommit, "Initial commit not found in log")
	})
	
	t.Run("Limited log", func(t *testing.T) {
		// Add more commits
		for i := 0; i < 3; i++ {
			file := filepath.Join(repoDir, "log"+string(rune('0'+i))+".txt")
			err := os.WriteFile(file, []byte("content"), 0644)
			require.NoError(t, err)
			err = git.Add(repoDir, ".")
			require.NoError(t, err)
			err = git.Commit(repoDir, "Commit "+string(rune('0'+i)))
			require.NoError(t, err)
		}
		
		logLines, err := git.Log(repoDir, 2)
		require.NoError(t, err)
		// Should have limited number of commits
		commitCount := 0
		for _, line := range logLines {
			if strings.HasPrefix(line, "commit ") {
				commitCount++
			}
		}
		assert.LessOrEqual(t, commitCount, 2)
	})
}

// TestRealGit_Stash - Stash methods not implemented in RealGit
// func TestRealGit_Stash(t *testing.T) {
// 	// Stash and StashPop methods need to be implemented
// }

func TestRealGit_IsRepo(t *testing.T) {
	git := NewRealGit()
	
	t.Run("Valid repo", func(t *testing.T) {
		repoDir, _ := setupTestRepo(t)
		isRepo := git.IsRepo(repoDir)
		assert.True(t, isRepo)
	})
	
	t.Run("Not a repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		isRepo := git.IsRepo(tmpDir)
		assert.False(t, isRepo)
	})
	
	t.Run("Non-existent directory", func(t *testing.T) {
		isRepo := git.IsRepo("/non/existent/path")
		assert.False(t, isRepo)
	})
}

func TestRealGit_RemoteURL(t *testing.T) {
	repoDir, git := setupTestRepo(t)
	
	t.Run("Get remote URL", func(t *testing.T) {
		// Add a remote first
		err := git.AddRemote(repoDir, "origin", "https://github.com/test/repo.git")
		require.NoError(t, err)
		
		url, err := git.RemoteURL(repoDir)
		require.NoError(t, err)
		assert.Contains(t, strings.TrimSpace(url), "github.com/test/repo.git")
	})
}

func TestRealGit_HasChanges(t *testing.T) {
	repoDir, git := setupTestRepo(t)
	
	t.Run("No changes", func(t *testing.T) {
		hasChanges, err := git.HasChanges(repoDir)
		require.NoError(t, err)
		assert.False(t, hasChanges)
	})
	
	t.Run("With changes", func(t *testing.T) {
		testFile := filepath.Join(repoDir, "test.txt")
		err := os.WriteFile(testFile, []byte("changed"), 0644)
		require.NoError(t, err)
		
		hasChanges, err := git.HasChanges(repoDir)
		require.NoError(t, err)
		assert.True(t, hasChanges)
	})
}

func TestRealGit_ListBranchesExtended(t *testing.T) {
	repoDir, git := setupTestRepo(t)
	
	// Create some branches
	branches := []string{"feature-1", "feature-2", "develop"}
	for _, branch := range branches {
		err := git.CreateBranch(repoDir, branch)
		require.NoError(t, err)
	}
	
	gotBranches, err := git.ListBranches(repoDir)
	require.NoError(t, err)
	
	// Should have at least the branches we created
	assert.GreaterOrEqual(t, len(gotBranches), len(branches))
	for _, branch := range branches {
		found := false
		for _, got := range gotBranches {
			if strings.Contains(got, branch) {
				found = true
				break
			}
		}
		assert.True(t, found, "Branch %s not found", branch)
	}
}