package mocks

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	
	"github.com/taokim/muno/internal/interfaces"
)

// MockCommandExecutor implements interfaces.CommandExecutor for testing
type MockCommandExecutor struct {
	mu    sync.RWMutex
	Calls []CommandCall
	
	// Function overrides for custom behavior
	ExecuteFunc         func(name string, args ...string) ([]byte, error)
	ExecuteWithInputFunc func(input string, name string, args ...string) ([]byte, error)
	ExecuteInDirFunc    func(dir string, name string, args ...string) ([]byte, error)
	ExecuteWithEnvFunc  func(env []string, name string, args ...string) ([]byte, error)
	StartFunc           func(name string, args ...string) (interfaces.Process, error)
	StartInDirFunc      func(dir string, name string, args ...string) (interfaces.Process, error)
	
	// Default responses
	DefaultOutput []byte
	DefaultError  error
	
	// Command-specific responses
	Responses map[string]CommandResponse
}

// CommandCall records a command execution
type CommandCall struct {
	Name  string
	Args  []string
	Dir   string
	Input string
	Env   []string
}

// CommandResponse defines a response for a specific command
type CommandResponse struct {
	Output []byte
	Error  error
}

// NewMockCommandExecutor creates a new mock command executor
func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		Calls:         []CommandCall{},
		Responses:     make(map[string]CommandResponse),
		DefaultOutput: []byte("mock output"),
	}
}

func (m *MockCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CommandCall{
		Name: name,
		Args: args,
	})
	
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(name, args...)
	}
	
	// Check for command-specific response
	key := fmt.Sprintf("%s %v", name, args)
	if resp, ok := m.Responses[key]; ok {
		return resp.Output, resp.Error
	}
	
	return m.DefaultOutput, m.DefaultError
}

func (m *MockCommandExecutor) ExecuteWithInput(input string, name string, args ...string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CommandCall{
		Name:  name,
		Args:  args,
		Input: input,
	})
	
	if m.ExecuteWithInputFunc != nil {
		return m.ExecuteWithInputFunc(input, name, args...)
	}
	
	return m.DefaultOutput, m.DefaultError
}

func (m *MockCommandExecutor) ExecuteInDir(dir string, name string, args ...string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CommandCall{
		Name: name,
		Args: args,
		Dir:  dir,
	})
	
	if m.ExecuteInDirFunc != nil {
		return m.ExecuteInDirFunc(dir, name, args...)
	}
	
	return m.DefaultOutput, m.DefaultError
}

func (m *MockCommandExecutor) ExecuteWithEnv(env []string, name string, args ...string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CommandCall{
		Name: name,
		Args: args,
		Env:  env,
	})
	
	if m.ExecuteWithEnvFunc != nil {
		return m.ExecuteWithEnvFunc(env, name, args...)
	}
	
	return m.DefaultOutput, m.DefaultError
}

func (m *MockCommandExecutor) Start(name string, args ...string) (interfaces.Process, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CommandCall{
		Name: name,
		Args: args,
	})
	
	if m.StartFunc != nil {
		return m.StartFunc(name, args...)
	}
	
	return &MockProcess{
		pid:    12345,
		stdout: bytes.NewBufferString(string(m.DefaultOutput)),
		stderr: bytes.NewBuffer(nil),
		stdin:  bytes.NewBuffer(nil),
	}, m.DefaultError
}

func (m *MockCommandExecutor) StartInDir(dir string, name string, args ...string) (interfaces.Process, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = append(m.Calls, CommandCall{
		Name: name,
		Args: args,
		Dir:  dir,
	})
	
	if m.StartInDirFunc != nil {
		return m.StartInDirFunc(dir, name, args...)
	}
	
	return &MockProcess{
		pid:    12345,
		stdout: bytes.NewBufferString(string(m.DefaultOutput)),
		stderr: bytes.NewBuffer(nil),
		stdin:  bytes.NewBuffer(nil),
	}, m.DefaultError
}

// Helper methods for testing

// SetResponse sets a response for a specific command
func (m *MockCommandExecutor) SetResponse(command string, args []string, output []byte, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := fmt.Sprintf("%s %v", command, args)
	m.Responses[key] = CommandResponse{
		Output: output,
		Error:  err,
	}
}

// GetCalls returns all command calls
func (m *MockCommandExecutor) GetCalls() []CommandCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	calls := make([]CommandCall, len(m.Calls))
	copy(calls, m.Calls)
	return calls
}

// Reset clears all calls and responses
func (m *MockCommandExecutor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Calls = []CommandCall{}
	m.Responses = make(map[string]CommandResponse)
}

// MockProcess implements interfaces.Process
type MockProcess struct {
	pid      int
	stdout   *bytes.Buffer
	stderr   *bytes.Buffer
	stdin    *bytes.Buffer
	waitErr  error
	killErr  error
	signalErr error
}

func (p *MockProcess) Wait() error {
	return p.waitErr
}

func (p *MockProcess) Kill() error {
	return p.killErr
}

func (p *MockProcess) Signal(sig os.Signal) error {
	return p.signalErr
}

func (p *MockProcess) Pid() int {
	return p.pid
}

func (p *MockProcess) StdoutPipe() (io.ReadCloser, error) {
	return io.NopCloser(p.stdout), nil
}

func (p *MockProcess) StderrPipe() (io.ReadCloser, error) {
	return io.NopCloser(p.stderr), nil
}

func (p *MockProcess) StdinPipe() (io.WriteCloser, error) {
	return nopWriteCloser{p.stdin}, nil
}

// nopWriteCloser wraps a Writer to add a no-op Close method
type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }