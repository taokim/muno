package mocks

import (
	"fmt"
	"sync"
	
	"github.com/taokim/muno/internal/interfaces"
)

// MockGitProvider is a mock implementation of GitProvider
type MockGitProvider struct {
	mu          sync.RWMutex
	statuses    map[string]*interfaces.GitStatus
	branches    map[string]string
	remoteURLs  map[string]string
	errors      map[string]error
	calls       []string
	pullResults map[string]interfaces.GitPullResult
	pushResults map[string]interfaces.GitPushResult
}

// NewMockGitProvider creates a new mock git provider
func NewMockGitProvider() *MockGitProvider {
	return &MockGitProvider{
		statuses:   make(map[string]*interfaces.GitStatus),
		branches:   make(map[string]string),
		remoteURLs: make(map[string]string),
		errors:     make(map[string]error),
		calls:      []string{},
	}
}

// Clone clones a repository
func (m *MockGitProvider) Clone(url, path string, options interfaces.CloneOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Clone(%s, %s)", url, path))
	
	if err, ok := m.errors["clone:"+path]; ok && err != nil {
		return err
	}
	
	// Set default status for cloned repo
	m.statuses[path] = &interfaces.GitStatus{
		Branch:   "main",
		IsClean:  true,
		HasUntracked: false,
		HasStaged: false,
		HasModified: false,
	}
	m.branches[path] = "main"
	m.remoteURLs[path] = url
	
	return nil
}

// Pull pulls changes from remote
func (m *MockGitProvider) Pull(path string, options interfaces.PullOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Pull(%s)", path))
	
	if err, ok := m.errors["pull:"+path]; ok && err != nil {
		return err
	}
	
	// Return mock pull result if set
	if result, ok := m.pullResults[path]; ok {
		// In a real implementation, we might update some state based on result
		// For mock, we just track that it was called
		_ = result
	}
	
	return nil
}

// Push pushes changes to remote
func (m *MockGitProvider) Push(path string, options interfaces.PushOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Push(%s)", path))
	
	if err, ok := m.errors["push:"+path]; ok && err != nil {
		return err
	}
	
	// Return mock push result if set
	if result, ok := m.pushResults[path]; ok {
		// In a real implementation, we might update some state based on result
		// For mock, we just track that it was called
		_ = result
	}
	
	return nil
}

// Status returns repository status
func (m *MockGitProvider) Status(path string) (*interfaces.GitStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Status(%s)", path))
	
	if err, ok := m.errors["status:"+path]; ok && err != nil {
		return nil, err
	}
	
	if status, ok := m.statuses[path]; ok {
		return status, nil
	}
	
	return &interfaces.GitStatus{
		Branch:   "main",
		IsClean:  true,
	}, nil
}

// Commit creates a commit
func (m *MockGitProvider) Commit(path string, message string, options interfaces.CommitOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Commit(%s, %s)", path, message))
	
	if err, ok := m.errors["commit:"+path]; ok && err != nil {
		return err
	}
	
	// Update status to clean after commit
	if status, ok := m.statuses[path]; ok {
		status.HasStaged = false
		status.HasModified = false
	}
	
	return nil
}

// Branch returns current branch
func (m *MockGitProvider) Branch(path string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Branch(%s)", path))
	
	if err, ok := m.errors["branch:"+path]; ok && err != nil {
		return "", err
	}
	
	if branch, ok := m.branches[path]; ok {
		return branch, nil
	}
	
	return "main", nil
}

// Checkout switches branches
func (m *MockGitProvider) Checkout(path string, branch string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Checkout(%s, %s)", path, branch))
	
	if err, ok := m.errors["checkout:"+path]; ok && err != nil {
		return err
	}
	
	m.branches[path] = branch
	if status, ok := m.statuses[path]; ok {
		status.Branch = branch
	}
	
	return nil
}

// Fetch fetches from remote
func (m *MockGitProvider) Fetch(path string, options interfaces.FetchOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Fetch(%s)", path))
	
	if err, ok := m.errors["fetch:"+path]; ok && err != nil {
		return err
	}
	
	return nil
}

// Add stages files
func (m *MockGitProvider) Add(path string, files []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Add(%s, %v)", path, files))
	
	if err, ok := m.errors["add:"+path]; ok && err != nil {
		return err
	}
	
	// Update status
	if status, ok := m.statuses[path]; ok {
		status.HasStaged = true
	}
	
	return nil
}

// Remove removes files
func (m *MockGitProvider) Remove(path string, files []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Remove(%s, %v)", path, files))
	
	if err, ok := m.errors["remove:"+path]; ok && err != nil {
		return err
	}
	
	return nil
}

// GetRemoteURL returns remote URL
func (m *MockGitProvider) GetRemoteURL(path string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("GetRemoteURL(%s)", path))
	
	if err, ok := m.errors["getremoteurl:"+path]; ok && err != nil {
		return "", err
	}
	
	if url, ok := m.remoteURLs[path]; ok {
		return url, nil
	}
	
	return "", fmt.Errorf("no remote URL for %s", path)
}

// SetRemoteURL sets remote URL
func (m *MockGitProvider) SetRemoteURL(path string, url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("SetRemoteURL(%s, %s)", path, url))
	
	if err, ok := m.errors["setremoteurl:"+path]; ok && err != nil {
		return err
	}
	
	m.remoteURLs[path] = url
	return nil
}

// SetPullResult sets the result for Pull operation
func (m *MockGitProvider) SetPullResult(path string, result interfaces.GitPullResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.pullResults == nil {
		m.pullResults = make(map[string]interfaces.GitPullResult)
	}
	m.pullResults[path] = result
}

// SetPushResult sets the result for Push operation
func (m *MockGitProvider) SetPushResult(path string, result interfaces.GitPushResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.pushResults == nil {
		m.pushResults = make(map[string]interfaces.GitPushResult)
	}
	m.pushResults[path] = result
}

// Mock helper methods

// SetStatus sets mock status for a path
func (m *MockGitProvider) SetStatus(path string, status *interfaces.GitStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.statuses[path] = status
	if status != nil {
		m.branches[path] = status.Branch
	}
}

// SetError sets an error for a specific operation and path
func (m *MockGitProvider) SetError(operation, path string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.errors[operation+":"+path] = err
}

// GetCalls returns all method calls made
func (m *MockGitProvider) GetCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	calls := make([]string, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// Reset resets the mock state
func (m *MockGitProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.statuses = make(map[string]*interfaces.GitStatus)
	m.branches = make(map[string]string)
	m.remoteURLs = make(map[string]string)
	m.errors = make(map[string]error)
	m.calls = []string{}
}