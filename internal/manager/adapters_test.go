package manager

import (
	"os"
	"path/filepath"
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

// MockGitInterface is a mock for git.Interface used in tree tests
type MockGitInterface struct{}

func (m *MockGitInterface) Clone(url, path string) error {
	return nil
}

func (m *MockGitInterface) Pull(path string) error {
	return nil
}

func (m *MockGitInterface) Status(path string) (string, error) {
	return "clean", nil
}

func (m *MockGitInterface) Commit(path, message string) error {
	return nil
}

func (m *MockGitInterface) Push(path string) error {
	return nil
}

func (m *MockGitInterface) Add(path, pattern string) error {
	return nil
}

// TestAdapterCreation tests the adapter creation functions
func TestAdapterCreation(t *testing.T) {
	t.Run("NewRealFileSystem", func(t *testing.T) {
		fs := NewRealFileSystem()
		assert.NotNil(t, fs)
		// Just check it returns something
	})

	t.Run("NewRealGit", func(t *testing.T) {
		git := NewRealGit()
		assert.NotNil(t, git)
	})

	t.Run("NewRealCommandExecutor", func(t *testing.T) {
		executor := NewRealCommandExecutor()
		assert.NotNil(t, executor)
	})

	t.Run("NewRealOutput", func(t *testing.T) {
		output := NewRealOutput(os.Stdout, os.Stderr)
		assert.NotNil(t, output)
	})

	t.Run("NewFileSystemAdapter", func(t *testing.T) {
		fs := NewRealFileSystem()
		adapter := NewFileSystemAdapter(fs)
		assert.NotNil(t, adapter)
	})

	t.Run("NewConfigAdapter", func(t *testing.T) {
		cfg := &config.ConfigTree{}
		adapter := NewConfigAdapter(cfg)
		assert.NotNil(t, adapter)
	})

	t.Run("NewGitAdapter", func(t *testing.T) {
		gitCmd := NewRealGit()
		adapter := NewGitAdapter(gitCmd)
		assert.NotNil(t, adapter)
	})

	t.Run("NewGitAdapter with nil", func(t *testing.T) {
		adapter := NewGitAdapter(nil)
		assert.NotNil(t, adapter)
	})

	t.Run("NewUIAdapter", func(t *testing.T) {
		output := NewRealOutput(os.Stdout, os.Stderr)
		adapter := NewUIAdapter(output)
		assert.NotNil(t, adapter)
	})

	t.Run("NewTreeAdapter", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create a minimal config for the manager
		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name: "test",
				ReposDir: "repos",
			},
			Nodes: []config.NodeDefinition{},
		}
		configPath := filepath.Join(tmpDir, "muno.yaml")
		err := cfg.Save(configPath)
		assert.NoError(t, err)
		
		mockGit := &MockGitInterface{}
		mgr, err := tree.NewManager(tmpDir, mockGit)
		assert.NoError(t, err)
		adapter := NewTreeAdapter(mgr)
		assert.NotNil(t, adapter)
	})
}

// TestTreeProviderAdapter tests the treeProviderAdapter implementation  
func TestTreeProviderAdapter(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a minimal config for the manager
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name: "test",
			ReposDir: "repos",
		},
		Nodes: []config.NodeDefinition{},
	}
	configPath := filepath.Join(tmpDir, "muno.yaml")
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	mockGit := &MockGitInterface{}
	mgr, err := tree.NewManager(tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	adapter := &treeProviderAdapter{
		mgr: mgr,
	}
	
	t.Run("Load", func(t *testing.T) {
		cfg := &config.ConfigTree{
			Nodes: []config.NodeDefinition{
				{Name: "test", URL: "https://github.com/test/repo.git"},
			},
		}
		err := adapter.Load(cfg)
		assert.NoError(t, err)
	})

	t.Run("Navigate", func(t *testing.T) {
		err := adapter.Navigate("/")
		// Navigation is not supported in stateless mode
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
	})

	t.Run("GetCurrent", func(t *testing.T) {
		node, err := adapter.GetCurrent()
		assert.NoError(t, err)
		assert.NotNil(t, node)
	})

	t.Run("GetTree", func(t *testing.T) {
		tree, err := adapter.GetTree()
		assert.NoError(t, err)
		assert.NotNil(t, tree)
	})

	t.Run("GetNode", func(t *testing.T) {
		node, err := adapter.GetNode("/")
		assert.NoError(t, err)
		assert.NotNil(t, node)
	})

	t.Run("ListChildren", func(t *testing.T) {
		_, err := adapter.ListChildren("/")
		assert.NoError(t, err)
		// Children can be empty list, that's ok
	})

	t.Run("AddNode", func(t *testing.T) {
		node := interfaces.NodeInfo{
			Name:       "test",
			Repository: "https://github.com/test/repo.git",
			IsLazy:     false,
		}
		err := adapter.AddNode("/", node)
		assert.NoError(t, err)
	})

	t.Run("RemoveNode", func(t *testing.T) {
		// First add a node
		node := interfaces.NodeInfo{
			Name:       "toremove",
			Repository: "https://github.com/test/remove.git",
			IsLazy:     false,
		}
		adapter.AddNode("/", node)
		err := adapter.RemoveNode("/toremove")
		assert.NoError(t, err)
	})

	t.Run("UpdateNode", func(t *testing.T) {
		// First add a node
		node := interfaces.NodeInfo{
			Name:       "toupdate",
			Repository: "https://github.com/test/update.git",
			IsLazy:     false,
		}
		adapter.AddNode("/", node)
		
		// Update it
		node.IsCloned = true
		err := adapter.UpdateNode("/toupdate", node)
		assert.NoError(t, err)
	})

	t.Run("GetPath", func(t *testing.T) {
		path := adapter.GetPath()
		assert.Equal(t, "/", path)
	})

	t.Run("SetPath", func(t *testing.T) {
		err := adapter.SetPath("/")
		// Setting path is not supported in stateless mode
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
	})

	t.Run("GetState", func(t *testing.T) {
		state, err := adapter.GetState()
		assert.NoError(t, err)
		assert.NotNil(t, state)
	})

	t.Run("SetState", func(t *testing.T) {
		state := interfaces.TreeState{}
		err := adapter.SetState(state)
		assert.Error(t, err) // Not supported
	})
}

// TestGitInterfaceAdapter tests the gitInterfaceAdapter implementation
func TestGitInterfaceAdapter(t *testing.T) {
	adapter := &gitInterfaceAdapter{
		git: &MockGitInterfaceForAdapter{},
	}

	t.Run("Clone", func(t *testing.T) {
		err := adapter.Clone("https://github.com/test/repo.git", "/path")
		assert.NoError(t, err)
	})

	t.Run("Pull", func(t *testing.T) {
		err := adapter.Pull("/path")
		assert.NoError(t, err)
	})

	t.Run("Push", func(t *testing.T) {
		err := adapter.Push("/path")
		assert.NoError(t, err)
	})

	t.Run("Status", func(t *testing.T) {
		status, err := adapter.Status("/path")
		assert.NoError(t, err)
		assert.Equal(t, "clean", status)
	})

	t.Run("Commit", func(t *testing.T) {
		err := adapter.Commit("/path", "test commit")
		assert.NoError(t, err)
	})

	t.Run("Add", func(t *testing.T) {
		err := adapter.Add("/path", "file.txt")
		assert.NoError(t, err)
	})
}

// TestCreateSharedMemory tests the createSharedMemory function
func TestCreateSharedMemory(t *testing.T) {
	tmpDir := t.TempDir()
	memoryPath := filepath.Join(tmpDir, "shared-memory.md")
	
	err := createSharedMemory(memoryPath)
	assert.NoError(t, err)
	
	// Check file was created
	_, err = os.Stat(memoryPath)
	assert.NoError(t, err)
	
	// Check content
	content, err := os.ReadFile(memoryPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "# Shared Memory")
}

// TestTreeProviderAdapterHelpers tests helper methods
func TestTreeProviderAdapterHelpers(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a minimal config for the manager
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name: "test",
			ReposDir: "repos",
		},
		Nodes: []config.NodeDefinition{},
	}
	configPath := filepath.Join(tmpDir, "muno.yaml")
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	mockGit := &MockGitInterface{}
	mgr, err := tree.NewManager(tmpDir, mockGit)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	adapter := &treeProviderAdapter{
		mgr: mgr,
	}
	
	t.Run("nodeToNodeInfo", func(t *testing.T) {
		node := &tree.TreeNode{
			Name:     "test",
			URL:      "https://github.com/test/repo.git",
			Lazy:     true,
			State:    tree.RepoStateCloned,
			Children: []string{},
		}
		
		info := adapter.nodeToNodeInfo("/test", node)
		assert.Equal(t, "test", info.Name)
		assert.Equal(t, "/test", info.Path)
		assert.Equal(t, "https://github.com/test/repo.git", info.Repository)
		assert.True(t, info.IsLazy)
	})
	
	t.Run("buildTreeRecursive", func(t *testing.T) {
		node := &tree.TreeNode{
			Name:     "root",
			Children: []string{},
		}
		
		info := adapter.buildTreeRecursive("/", node)
		assert.Equal(t, "root", info.Name)
		assert.Equal(t, "/", info.Path)
	})
	
	t.Run("collectNodesRecursive", func(t *testing.T) {
		node := &tree.TreeNode{
			Name:     "root",
			Children: []string{},
		}
		
		nodes := make(map[string]interfaces.NodeInfo)
		adapter.collectNodesRecursive("/", node, nodes)
		assert.Contains(t, nodes, "/")
	})
}