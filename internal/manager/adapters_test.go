package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/tree"
)

// TestGitProviderAdapter tests the gitProviderAdapter implementation
func TestGitProviderAdapter(t *testing.T) {
	adapter := &gitProviderAdapter{
		git: &MockGitInterfaceForAdapter{},
	}

	t.Run("Clone", func(t *testing.T) {
		err := adapter.Clone("https://github.com/test/repo.git", "/path", interfaces.CloneOptions{})
		assert.NoError(t, err)
	})

	t.Run("Pull", func(t *testing.T) {
		err := adapter.Pull("/path", interfaces.PullOptions{})
		assert.NoError(t, err)
	})

	t.Run("Push", func(t *testing.T) {
		err := adapter.Push("/path", interfaces.PushOptions{})
		assert.NoError(t, err)
	})

	t.Run("Status", func(t *testing.T) {
		status, err := adapter.Status("/path")
		assert.NoError(t, err)
		assert.NotNil(t, status)
	})

	t.Run("Commit", func(t *testing.T) {
		err := adapter.Commit("/path", "test commit", interfaces.CommitOptions{})
		assert.NoError(t, err)
	})

	t.Run("Add", func(t *testing.T) {
		err := adapter.Add("/path", []string{"file.txt"})
		assert.NoError(t, err)
	})

	t.Run("GetRemoteURL", func(t *testing.T) {
		url, err := adapter.GetRemoteURL("/path")
		assert.NoError(t, err)
		assert.Equal(t, "https://github.com/test/repo.git", url)
	})

	t.Run("Branch", func(t *testing.T) {
		branch, err := adapter.Branch("/path")
		assert.NoError(t, err)
		assert.Equal(t, "main", branch)
	})

	t.Run("Checkout", func(t *testing.T) {
		err := adapter.Checkout("/path", "develop")
		assert.NoError(t, err)
	})

	t.Run("Fetch", func(t *testing.T) {
		err := adapter.Fetch("/path", interfaces.FetchOptions{})
		assert.NoError(t, err)
	})

	t.Run("Remove", func(t *testing.T) {
		err := adapter.Remove("/path", []string{"file.txt"})
		assert.Error(t, err) // Not implemented
	})

	t.Run("SetRemoteURL", func(t *testing.T) {
		err := adapter.SetRemoteURL("/path", "https://github.com/test/new.git")
		assert.Error(t, err) // Not implemented
	})
}

// MockGitInterfaceForAdapter is a mock for GitInterface used in adapter tests
type MockGitInterfaceForAdapter struct{}

func (m *MockGitInterfaceForAdapter) Clone(url, path string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) Pull(path string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) Push(path string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) Status(path string) (string, error) {
	return "clean", nil
}

func (m *MockGitInterfaceForAdapter) Commit(path, message string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) Add(path string, files ...string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) CurrentBranch(path string) (string, error) {
	return "main", nil
}

func (m *MockGitInterfaceForAdapter) Checkout(path, branch string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) CheckoutNew(path, branch string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) Fetch(path string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) Remove(path string, files ...string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) IsRepo(path string) bool {
	return true
}

func (m *MockGitInterfaceForAdapter) RemoteURL(path string) (string, error) {
	return "https://github.com/test/repo.git", nil
}

func (m *MockGitInterfaceForAdapter) HasChanges(path string) (bool, error) {
	return false, nil
}

func (m *MockGitInterfaceForAdapter) AddAll(path string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) CloneWithOptions(url, path string, options ...string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) PullWithOptions(path string, options ...string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) PushWithOptions(path string, options ...string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) FetchWithOptions(path string, options ...string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) StatusWithOptions(path string, options ...string) (string, error) {
	return "clean", nil
}

func (m *MockGitInterfaceForAdapter) Branch(path string) (string, error) {
	return "main", nil
}

func (m *MockGitInterfaceForAdapter) CommitWithOptions(path, message string, options ...string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) CreateBranch(path, branch string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) DeleteBranch(path, branch string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) ListBranches(path string) ([]string, error) {
	return []string{"main"}, nil
}

func (m *MockGitInterfaceForAdapter) Tag(path, tag string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) TagWithMessage(path, tag, message string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) ListTags(path string) ([]string, error) {
	return []string{}, nil
}

func (m *MockGitInterfaceForAdapter) Reset(path string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) ResetHard(path string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) ResetSoft(path string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) Diff(path string) (string, error) {
	return "", nil
}

func (m *MockGitInterfaceForAdapter) DiffStaged(path string) (string, error) {
	return "", nil
}

func (m *MockGitInterfaceForAdapter) DiffWithBranch(path, branch string) (string, error) {
	return "", nil
}

func (m *MockGitInterfaceForAdapter) Log(path string, limit int) ([]string, error) {
	return []string{}, nil
}

func (m *MockGitInterfaceForAdapter) LogOneline(path string, limit int) ([]string, error) {
	return []string{}, nil
}

func (m *MockGitInterfaceForAdapter) AddRemote(path, name, url string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) RemoveRemote(path, name string) error {
	return nil
}

func (m *MockGitInterfaceForAdapter) ListRemotes(path string) (map[string]string, error) {
	return map[string]string{"origin": "https://github.com/test/repo.git"}, nil
}

// Test adapter with nil git
func TestGitProviderAdapterWithNilGit(t *testing.T) {
	adapter := &gitProviderAdapter{
		git: nil,
	}

	t.Run("Status with nil git", func(t *testing.T) {
		status, err := adapter.Status("/path")
		assert.NoError(t, err)
		assert.NotNil(t, status)
		assert.True(t, status.IsClean)
	})
}

// TestTreeProviderAdapter tests the treeProviderAdapter implementation  
func TestTreeProviderAdapter(t *testing.T) {
	adapter := &treeProviderAdapter{
		mgr: &tree.Manager{},  // Basic test with real manager
	}
	
	t.Run("Load", func(t *testing.T) {
		cfg := &config.ConfigTree{
			Nodes: []config.NodeDefinition{
				{Name: "test", URL: "https://github.com/test/repo.git"},
			},
		}
		// Load doesn't do anything in the adapter
		err := adapter.Load(cfg)
		assert.NoError(t, err)
	})
}

