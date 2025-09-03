package adapters

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/taokim/muno/internal/mocks"
)

func TestGitInterfaceWrapper_BasicOperations(t *testing.T) {
	mockGit := mocks.NewMockGitInterface()
	wrapper := NewGitInterfaceWrapper(mockGit)

	t.Run("Clone", func(t *testing.T) {
		mockGit.ResetMock()
		err := wrapper.Clone("https://github.com/test/repo.git", "/path/to/dest")
		assert.NoError(t, err)
		assert.Len(t, mockGit.Calls, 1)
		assert.Equal(t, "Clone", mockGit.Calls[0].Method)
		assert.Equal(t, "https://github.com/test/repo.git", mockGit.Calls[0].Args[0])
		assert.Equal(t, "/path/to/dest", mockGit.Calls[0].Path)
	})

	t.Run("Status", func(t *testing.T) {
		mockGit.ResetMock()
		mockGit.StatusFunc = func(path string) (string, error) {
			return "On branch master\nnothing to commit", nil
		}
		
		status, err := wrapper.Status("/repo")
		assert.NoError(t, err)
		assert.Contains(t, status, "On branch master")
		assert.Len(t, mockGit.Calls, 1)
		assert.Equal(t, "Status", mockGit.Calls[0].Method)
	})

	t.Run("Commit", func(t *testing.T) {
		mockGit.ResetMock()
		err := wrapper.Commit("/repo", "Test commit message")
		assert.NoError(t, err)
		assert.Len(t, mockGit.Calls, 1)
		assert.Equal(t, "Commit", mockGit.Calls[0].Method)
		assert.Equal(t, "/repo", mockGit.Calls[0].Path)
		if len(mockGit.Calls[0].Args) > 0 {
			assert.Equal(t, "Test commit message", mockGit.Calls[0].Args[0])
		}
	})

	t.Run("Push", func(t *testing.T) {
		mockGit.ResetMock()
		err := wrapper.Push("/repo")
		assert.NoError(t, err)
		assert.Len(t, mockGit.Calls, 1)
		assert.Equal(t, "Push", mockGit.Calls[0].Method)
	})

	t.Run("Pull", func(t *testing.T) {
		mockGit.ResetMock()
		err := wrapper.Pull("/repo")
		assert.NoError(t, err)
		assert.Len(t, mockGit.Calls, 1)
		assert.Equal(t, "Pull", mockGit.Calls[0].Method)
	})
}

func TestGitInterfaceWrapper_AddOperations(t *testing.T) {
	mockGit := mocks.NewMockGitInterface()
	wrapper := NewGitInterfaceWrapper(mockGit)

	t.Run("Add with dot pattern calls AddAll", func(t *testing.T) {
		mockGit.ResetMock()
		err := wrapper.Add("/repo", ".")
		assert.NoError(t, err)
		assert.Len(t, mockGit.Calls, 1)
		assert.Equal(t, "AddAll", mockGit.Calls[0].Method)
		assert.Equal(t, "/repo", mockGit.Calls[0].Path)
	})

	t.Run("Add with specific file", func(t *testing.T) {
		mockGit.ResetMock()
		err := wrapper.Add("/repo", "file.txt")
		assert.NoError(t, err)
		assert.Len(t, mockGit.Calls, 1)
		assert.Equal(t, "Add", mockGit.Calls[0].Method)
		assert.Equal(t, "/repo", mockGit.Calls[0].Path)
		if len(mockGit.Calls[0].Args) > 0 {
			assert.Equal(t, "file.txt", mockGit.Calls[0].Args[0])
		}
	})

	t.Run("Add with wildcard pattern", func(t *testing.T) {
		mockGit.ResetMock()
		err := wrapper.Add("/repo", "*.txt")
		assert.NoError(t, err)
		assert.Len(t, mockGit.Calls, 1)
		assert.Equal(t, "Add", mockGit.Calls[0].Method)
		if len(mockGit.Calls[0].Args) > 0 {
			assert.Equal(t, "*.txt", mockGit.Calls[0].Args[0])
		}
	})
}

func TestGitInterfaceWrapper_ErrorHandling(t *testing.T) {
	mockGit := mocks.NewMockGitInterface()
	wrapper := NewGitInterfaceWrapper(mockGit)

	t.Run("Clone error", func(t *testing.T) {
		mockGit.ResetMock()
		mockGit.SetError("Clone", assert.AnError)

		err := wrapper.Clone("https://github.com/test/repo.git", "/dest")
		assert.Error(t, err)
	})

	t.Run("Commit error", func(t *testing.T) {
		mockGit.ResetMock()
		mockGit.SetError("Commit", assert.AnError)

		err := wrapper.Commit("/repo", "test")
		assert.Error(t, err)
	})

	t.Run("Status error", func(t *testing.T) {
		mockGit.ResetMock()
		mockGit.StatusFunc = func(path string) (string, error) {
			return "", assert.AnError
		}

		_, err := wrapper.Status("/repo")
		assert.Error(t, err)
	})
}

func TestGitInterfaceWrapper_NilGit(t *testing.T) {
	wrapper := &GitInterfaceWrapper{git: nil}

	// These should return errors when git is nil
	t.Run("Operations with nil git", func(t *testing.T) {
		// Test that operations with nil git return errors gracefully
		_, err := wrapper.Status("/repo")
		assert.Error(t, err)
		
		err = wrapper.Add("/repo", ".")
		assert.Error(t, err)
		
		err = wrapper.Commit("/repo", "test message")
		assert.Error(t, err)
	})
}