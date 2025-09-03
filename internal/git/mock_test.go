package git

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockGit(t *testing.T) {
	t.Run("Clone", func(t *testing.T) {
		mock := &MockGit{
			CloneFunc: func(url, path string) error {
				if url == "fail" {
					return errors.New("clone failed")
				}
				return nil
			},
		}
		
		// Test successful clone
		err := mock.Clone("https://github.com/test/repo.git", "/tmp/repo")
		assert.NoError(t, err)
		
		// Test failed clone
		err = mock.Clone("fail", "/tmp/repo")
		assert.Error(t, err)
		assert.Equal(t, "clone failed", err.Error())
		
		// Test with nil CloneFunc
		mock.CloneFunc = nil
		err = mock.Clone("test", "path")
		assert.NoError(t, err)
	})

	t.Run("Pull", func(t *testing.T) {
		mock := &MockGit{
			PullFunc: func(path string) error {
				if path == "fail" {
					return errors.New("pull failed")
				}
				return nil
			},
		}
		
		// Test successful pull
		err := mock.Pull("/tmp/repo")
		assert.NoError(t, err)
		
		// Test failed pull
		err = mock.Pull("fail")
		assert.Error(t, err)
		
		// Test with nil PullFunc
		mock.PullFunc = nil
		err = mock.Pull("test")
		assert.NoError(t, err)
	})

	t.Run("Status", func(t *testing.T) {
		mock := &MockGit{
			StatusFunc: func(path string) (string, error) {
				if path == "fail" {
					return "", errors.New("status failed")
				}
				return "M file.txt", nil
			},
		}
		
		// Test successful status
		status, err := mock.Status("/tmp/repo")
		assert.NoError(t, err)
		assert.Equal(t, "M file.txt", status)
		
		// Test failed status
		status, err = mock.Status("fail")
		assert.Error(t, err)
		assert.Empty(t, status)
		
		// Test with nil StatusFunc
		mock.StatusFunc = nil
		status, err = mock.Status("test")
		assert.NoError(t, err)
		assert.Equal(t, "clean", status)
	})

	t.Run("Commit", func(t *testing.T) {
		mock := &MockGit{
			CommitFunc: func(path, message string) error {
				if message == "fail" {
					return errors.New("commit failed")
				}
				return nil
			},
		}
		
		// Test successful commit
		err := mock.Commit("/tmp/repo", "Test commit")
		assert.NoError(t, err)
		
		// Test failed commit
		err = mock.Commit("/tmp/repo", "fail")
		assert.Error(t, err)
		
		// Test with nil CommitFunc
		mock.CommitFunc = nil
		err = mock.Commit("path", "message")
		assert.NoError(t, err)
	})

	t.Run("Push", func(t *testing.T) {
		mock := &MockGit{
			PushFunc: func(path string) error {
				if path == "fail" {
					return errors.New("push failed")
				}
				return nil
			},
		}
		
		// Test successful push
		err := mock.Push("/tmp/repo")
		assert.NoError(t, err)
		
		// Test failed push
		err = mock.Push("fail")
		assert.Error(t, err)
		
		// Test with nil PushFunc
		mock.PushFunc = nil
		err = mock.Push("test")
		assert.NoError(t, err)
	})

	t.Run("Add", func(t *testing.T) {
		mock := &MockGit{
			AddFunc: func(path, pattern string) error {
				if pattern == "fail" {
					return errors.New("add failed")
				}
				return nil
			},
		}
		
		// Test successful add
		err := mock.Add("/tmp/repo", ".")
		assert.NoError(t, err)
		
		// Test failed add
		err = mock.Add("/tmp/repo", "fail")
		assert.Error(t, err)
		
		// Test with nil AddFunc
		mock.AddFunc = nil
		err = mock.Add("path", "pattern")
		assert.NoError(t, err)
	})
}