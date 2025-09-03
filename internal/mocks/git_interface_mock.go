package mocks

import (
	"sync"
)

// CallRecord represents a method call record
type CallRecord struct {
	Method string
	Path   string
	Args   []interface{}
}

// MockGitInterface is a mock implementation of GitInterface
type MockGitInterface struct {
	mu         sync.RWMutex
	Calls      []CallRecord
	StatusFunc func(path string) (string, error)
	errors     map[string]error
}

// NewMockGitInterface creates a new mock git interface
func NewMockGitInterface() *MockGitInterface {
	return &MockGitInterface{
		Calls:  []CallRecord{},
		errors: make(map[string]error),
	}
}

// ResetMock resets the mock state
func (m *MockGitInterface) ResetMock() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = []CallRecord{}
}

// SetError sets an error for a specific operation
func (m *MockGitInterface) SetError(operation string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[operation] = err
}

// Clone clones a repository
func (m *MockGitInterface) Clone(url, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Clone",
		Path:   path,
		Args:   []interface{}{url},
	})
	
	if err, ok := m.errors["Clone"]; ok && err != nil {
		return err
	}
	return nil
}

// CloneWithOptions clones with options
func (m *MockGitInterface) CloneWithOptions(url, path string, options ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "CloneWithOptions",
		Path:   path,
		Args:   []interface{}{url, options},
	})
	
	if err, ok := m.errors["CloneWithOptions"]; ok && err != nil {
		return err
	}
	return nil
}

// Pull pulls changes
func (m *MockGitInterface) Pull(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Pull",
		Path:   path,
	})
	
	if err, ok := m.errors["Pull"]; ok && err != nil {
		return err
	}
	return nil
}

// PullWithOptions pulls with options
func (m *MockGitInterface) PullWithOptions(path string, options ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "PullWithOptions",
		Path:   path,
		Args:   []interface{}{options},
	})
	
	if err, ok := m.errors["PullWithOptions"]; ok && err != nil {
		return err
	}
	return nil
}

// Push pushes changes
func (m *MockGitInterface) Push(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Push",
		Path:   path,
	})
	
	if err, ok := m.errors["Push"]; ok && err != nil {
		return err
	}
	return nil
}

// PushWithOptions pushes with options
func (m *MockGitInterface) PushWithOptions(path string, options ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "PushWithOptions",
		Path:   path,
		Args:   []interface{}{options},
	})
	
	if err, ok := m.errors["PushWithOptions"]; ok && err != nil {
		return err
	}
	return nil
}

// Fetch fetches changes
func (m *MockGitInterface) Fetch(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Fetch",
		Path:   path,
	})
	
	if err, ok := m.errors["Fetch"]; ok && err != nil {
		return err
	}
	return nil
}

// FetchWithOptions fetches with options
func (m *MockGitInterface) FetchWithOptions(path string, options ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "FetchWithOptions",
		Path:   path,
		Args:   []interface{}{options},
	})
	
	if err, ok := m.errors["FetchWithOptions"]; ok && err != nil {
		return err
	}
	return nil
}

// Status returns repository status
func (m *MockGitInterface) Status(path string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Status",
		Path:   path,
	})
	
	if m.StatusFunc != nil {
		return m.StatusFunc(path)
	}
	
	if err, ok := m.errors["Status"]; ok && err != nil {
		return "", err
	}
	
	return "On branch master\nnothing to commit, working tree clean", nil
}

// StatusWithOptions returns status with options
func (m *MockGitInterface) StatusWithOptions(path string, options ...string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "StatusWithOptions",
		Path:   path,
		Args:   []interface{}{options},
	})
	
	if err, ok := m.errors["StatusWithOptions"]; ok && err != nil {
		return "", err
	}
	
	return "On branch master\nnothing to commit, working tree clean", nil
}

// Branch returns current branch
func (m *MockGitInterface) Branch(path string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Branch",
		Path:   path,
	})
	
	if err, ok := m.errors["Branch"]; ok && err != nil {
		return "", err
	}
	
	return "master", nil
}

// CurrentBranch returns current branch
func (m *MockGitInterface) CurrentBranch(path string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "CurrentBranch",
		Path:   path,
	})
	
	if err, ok := m.errors["CurrentBranch"]; ok && err != nil {
		return "", err
	}
	
	return "master", nil
}

// RemoteURL returns remote URL
func (m *MockGitInterface) RemoteURL(path string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "RemoteURL",
		Path:   path,
	})
	
	if err, ok := m.errors["RemoteURL"]; ok && err != nil {
		return "", err
	}
	
	return "https://github.com/test/repo.git", nil
}

// HasChanges checks if repo has changes
func (m *MockGitInterface) HasChanges(path string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "HasChanges",
		Path:   path,
	})
	
	if err, ok := m.errors["HasChanges"]; ok && err != nil {
		return false, err
	}
	
	return false, nil
}

// IsRepo checks if path is a git repository
func (m *MockGitInterface) IsRepo(path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "IsRepo",
		Path:   path,
	})
	
	return true
}

// Add adds files to staging
func (m *MockGitInterface) Add(path string, files ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	args := make([]interface{}, len(files))
	for i, f := range files {
		args[i] = f
	}
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Add",
		Path:   path,
		Args:   args,
	})
	
	if err, ok := m.errors["Add"]; ok && err != nil {
		return err
	}
	
	return nil
}

// AddAll adds all files to staging
func (m *MockGitInterface) AddAll(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "AddAll",
		Path:   path,
	})
	
	if err, ok := m.errors["AddAll"]; ok && err != nil {
		return err
	}
	
	return nil
}

// Commit creates a commit
func (m *MockGitInterface) Commit(path, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Commit",
		Path:   path,
		Args:   []interface{}{message},
	})
	
	if err, ok := m.errors["Commit"]; ok && err != nil {
		return err
	}
	
	return nil
}

// CommitWithOptions commits with options
func (m *MockGitInterface) CommitWithOptions(path, message string, options ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "CommitWithOptions",
		Path:   path,
		Args:   []interface{}{message, options},
	})
	
	if err, ok := m.errors["CommitWithOptions"]; ok && err != nil {
		return err
	}
	
	return nil
}

// Checkout switches branches
func (m *MockGitInterface) Checkout(path, branch string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Checkout",
		Path:   path,
		Args:   []interface{}{branch},
	})
	
	if err, ok := m.errors["Checkout"]; ok && err != nil {
		return err
	}
	
	return nil
}

// CheckoutNew creates and switches to new branch
func (m *MockGitInterface) CheckoutNew(path, branch string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "CheckoutNew",
		Path:   path,
		Args:   []interface{}{branch},
	})
	
	if err, ok := m.errors["CheckoutNew"]; ok && err != nil {
		return err
	}
	
	return nil
}
// CreateBranch creates a new branch
func (m *MockGitInterface) CreateBranch(path, branch string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "CreateBranch",
		Path:   path,
		Args:   []interface{}{branch},
	})
	
	if err, ok := m.errors["CreateBranch"]; ok && err != nil {
		return err
	}
	
	return nil
}

// DeleteBranch deletes a branch
func (m *MockGitInterface) DeleteBranch(path, branch string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "DeleteBranch",
		Path:   path,
		Args:   []interface{}{branch},
	})
	
	if err, ok := m.errors["DeleteBranch"]; ok && err != nil {
		return err
	}
	
	return nil
}

// ListBranches lists all branches
func (m *MockGitInterface) ListBranches(path string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "ListBranches",
		Path:   path,
	})
	
	if err, ok := m.errors["ListBranches"]; ok && err != nil {
		return nil, err
	}
	
	return []string{"master", "develop"}, nil
}

// Tag creates a tag
func (m *MockGitInterface) Tag(path, tag string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Tag",
		Path:   path,
		Args:   []interface{}{tag},
	})
	
	if err, ok := m.errors["Tag"]; ok && err != nil {
		return err
	}
	
	return nil
}

// TagWithMessage creates a tag with message
func (m *MockGitInterface) TagWithMessage(path, tag, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "TagWithMessage",
		Path:   path,
		Args:   []interface{}{tag, message},
	})
	
	if err, ok := m.errors["TagWithMessage"]; ok && err != nil {
		return err
	}
	
	return nil
}

// ListTags lists all tags
func (m *MockGitInterface) ListTags(path string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "ListTags",
		Path:   path,
	})
	
	if err, ok := m.errors["ListTags"]; ok && err != nil {
		return nil, err
	}
	
	return []string{"v1.0.0", "v1.1.0"}, nil
}

// Reset resets repository
func (m *MockGitInterface) Reset(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Reset",
		Path:   path,
	})
	
	if err, ok := m.errors["Reset"]; ok && err != nil {
		return err
	}
	
	return nil
}

// ResetHard performs hard reset
func (m *MockGitInterface) ResetHard(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "ResetHard",
		Path:   path,
	})
	
	if err, ok := m.errors["ResetHard"]; ok && err != nil {
		return err
	}
	
	return nil
}

// ResetSoft performs soft reset
func (m *MockGitInterface) ResetSoft(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "ResetSoft",
		Path:   path,
	})
	
	if err, ok := m.errors["ResetSoft"]; ok && err != nil {
		return err
	}
	
	return nil
}

// Diff shows diff
func (m *MockGitInterface) Diff(path string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Diff",
		Path:   path,
	})
	
	if err, ok := m.errors["Diff"]; ok && err != nil {
		return "", err
	}
	
	return "diff --git a/file.txt b/file.txt", nil
}

// DiffStaged shows staged diff
func (m *MockGitInterface) DiffStaged(path string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "DiffStaged",
		Path:   path,
	})
	
	if err, ok := m.errors["DiffStaged"]; ok && err != nil {
		return "", err
	}
	
	return "diff --staged", nil
}

// DiffWithBranch shows diff with branch
func (m *MockGitInterface) DiffWithBranch(path, branch string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "DiffWithBranch",
		Path:   path,
		Args:   []interface{}{branch},
	})
	
	if err, ok := m.errors["DiffWithBranch"]; ok && err != nil {
		return "", err
	}
	
	return "diff with " + branch, nil
}

// Log shows git log
func (m *MockGitInterface) Log(path string, limit int) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "Log",
		Path:   path,
		Args:   []interface{}{limit},
	})
	
	if err, ok := m.errors["Log"]; ok && err != nil {
		return nil, err
	}
	
	return []string{"commit 1", "commit 2"}, nil
}

// LogOneline shows oneline log
func (m *MockGitInterface) LogOneline(path string, limit int) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "LogOneline",
		Path:   path,
		Args:   []interface{}{limit},
	})
	
	if err, ok := m.errors["LogOneline"]; ok && err != nil {
		return nil, err
	}
	
	return []string{"abc123 commit 1", "def456 commit 2"}, nil
}

// AddRemote adds a remote
func (m *MockGitInterface) AddRemote(path, name, url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "AddRemote",
		Path:   path,
		Args:   []interface{}{name, url},
	})
	
	if err, ok := m.errors["AddRemote"]; ok && err != nil {
		return err
	}
	
	return nil
}

// RemoveRemote removes a remote
func (m *MockGitInterface) RemoveRemote(path, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "RemoveRemote",
		Path:   path,
		Args:   []interface{}{name},
	})
	
	if err, ok := m.errors["RemoveRemote"]; ok && err != nil {
		return err
	}
	
	return nil
}

// ListRemotes lists all remotes
func (m *MockGitInterface) ListRemotes(path string) (map[string]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CallRecord{
		Method: "ListRemotes",
		Path:   path,
	})
	
	if err, ok := m.errors["ListRemotes"]; ok && err != nil {
		return nil, err
	}
	
	return map[string]string{"origin": "https://github.com/test/repo.git"}, nil
}

// CloneFunc field for custom behavior
var CloneFunc func(url, dest string) error

// CommitFunc field for custom behavior  
var CommitFunc func(path, message string) error
