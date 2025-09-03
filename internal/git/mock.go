package git

// MockGit implements Interface for testing
type MockGit struct {
	CloneFunc  func(url, path string) error
	PullFunc   func(path string) error
	StatusFunc func(path string) (string, error)
	CommitFunc func(path, message string) error
	PushFunc   func(path string) error
	AddFunc    func(path, pattern string) error
	
	// Track calls for assertions
	CloneCalls  [][]string
	PullCalls   []string
	StatusCalls []string
	CommitCalls [][]string
	PushCalls   []string
	AddCalls    [][]string
}

// Clone mocks git clone
func (m *MockGit) Clone(url, path string) error {
	m.CloneCalls = append(m.CloneCalls, []string{url, path})
	if m.CloneFunc != nil {
		return m.CloneFunc(url, path)
	}
	return nil
}

// Pull mocks git pull
func (m *MockGit) Pull(path string) error {
	m.PullCalls = append(m.PullCalls, path)
	if m.PullFunc != nil {
		return m.PullFunc(path)
	}
	return nil
}

// Status mocks git status
func (m *MockGit) Status(path string) (string, error) {
	m.StatusCalls = append(m.StatusCalls, path)
	if m.StatusFunc != nil {
		return m.StatusFunc(path)
	}
	return "clean", nil
}

// Commit mocks git commit
func (m *MockGit) Commit(path, message string) error {
	m.CommitCalls = append(m.CommitCalls, []string{path, message})
	if m.CommitFunc != nil {
		return m.CommitFunc(path, message)
	}
	return nil
}

// Push mocks git push
func (m *MockGit) Push(path string) error {
	m.PushCalls = append(m.PushCalls, path)
	if m.PushFunc != nil {
		return m.PushFunc(path)
	}
	return nil
}

// Add mocks git add
func (m *MockGit) Add(path, pattern string) error {
	m.AddCalls = append(m.AddCalls, []string{path, pattern})
	if m.AddFunc != nil {
		return m.AddFunc(path, pattern)
	}
	return nil
}

// Ensure MockGit implements Interface
var _ Interface = (*MockGit)(nil)