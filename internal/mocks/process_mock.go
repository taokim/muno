package mocks

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
	
	"github.com/taokim/muno/internal/interfaces"
)

// MockProcessProvider is a mock implementation of ProcessProvider
type MockProcessProvider struct {
	mu      sync.RWMutex
	results map[string]*interfaces.ProcessResult
	errors  map[string]error
	calls   []string
}

// NewMockProcessProvider creates a new mock process provider
func NewMockProcessProvider() *MockProcessProvider {
	return &MockProcessProvider{
		results: make(map[string]*interfaces.ProcessResult),
		errors:  make(map[string]error),
		calls:   []string{},
	}
}

// Execute executes a command
func (m *MockProcessProvider) Execute(ctx context.Context, command string, args []string, options interfaces.ProcessOptions) (*interfaces.ProcessResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "Execute("+command+")")
	
	if err, ok := m.errors[command]; ok && err != nil {
		return nil, err
	}
	
	if result, ok := m.results[command]; ok {
		return result, nil
	}
	
	return &interfaces.ProcessResult{
		Stdout:   "mock output",
		ExitCode: 0,
	}, nil
}

// ExecuteShell executes a shell command
func (m *MockProcessProvider) ExecuteShell(ctx context.Context, command string, options interfaces.ProcessOptions) (*interfaces.ProcessResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "ExecuteShell("+command+")")
	
	if err, ok := m.errors["shell:"+command]; ok && err != nil {
		return nil, err
	}
	
	return &interfaces.ProcessResult{
		Stdout:   "mock shell output",
		ExitCode: 0,
	}, nil
}

// StartBackground starts a background process
func (m *MockProcessProvider) StartBackground(ctx context.Context, command string, args []string, options interfaces.ProcessOptions) (interfaces.Process, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "StartBackground("+command+")")
	
	if err, ok := m.errors["bg:"+command]; ok && err != nil {
		return nil, err
	}
	
	return &mockProcess{pid: 12345}, nil
}

// OpenInEditor opens a file in the default editor
func (m *MockProcessProvider) OpenInEditor(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "OpenInEditor("+path+")")
	
	if err, ok := m.errors["editor:"+path]; ok && err != nil {
		return err
	}
	
	return nil
}

// OpenInBrowser opens a URL in the default browser
func (m *MockProcessProvider) OpenInBrowser(url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, "OpenInBrowser("+url+")")
	
	if err, ok := m.errors["browser:"+url]; ok && err != nil {
		return err
	}
	
	return nil
}

// Mock helper methods

// SetResult sets the result for a command
func (m *MockProcessProvider) SetResult(command string, result *interfaces.ProcessResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.results[command] = result
}

// SetError sets an error for a command
func (m *MockProcessProvider) SetError(key string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.errors[key] = err
}

// GetCalls returns all method calls made
func (m *MockProcessProvider) GetCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	calls := make([]string, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// Reset resets the mock state
func (m *MockProcessProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.results = make(map[string]*interfaces.ProcessResult)
	m.errors = make(map[string]error)
	m.calls = []string{}
}

// mockProcess is a mock Process implementation
type mockProcess struct {
	pid int
}

func (p *mockProcess) Pid() int {
	return p.pid
}

func (p *mockProcess) Wait() error {
	return nil
}

func (p *mockProcess) Kill() error {
	return nil
}

func (p *mockProcess) Signal(sig os.Signal) error {
	return nil
}

func (p *mockProcess) StdoutPipe() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock stdout")), nil
}

func (p *mockProcess) StderrPipe() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock stderr")), nil
}

func (p *mockProcess) StdinPipe() (io.WriteCloser, error) {
	return &mockWriteCloser{}, nil
}

// Use mockWriteCloser to avoid conflict
type mockWriteCloser struct{}

func (mwc *mockWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (mwc *mockWriteCloser) Close() error {
	return nil
}