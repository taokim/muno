package manager

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// MockCommandExecutor allows mocking command execution
type MockCommandExecutor struct {
	Commands []MockCommand
	Index    int
}

type MockCommand struct {
	Cmd      string
	Args     []string
	Response string
	Error    error
	ExitCode int
	OnCall   func() // Callback when command is called
}

func (m *MockCommandExecutor) Execute(cmd string, args ...string) error {
	if m.Index >= len(m.Commands) {
		return fmt.Errorf("unexpected command: %s %v", cmd, args)
	}
	
	expected := m.Commands[m.Index]
	m.Index++
	
	if expected.Cmd != cmd {
		return fmt.Errorf("expected command %s, got %s", expected.Cmd, cmd)
	}
	
	// Simple args matching
	if len(expected.Args) > 0 && !equalArgs(expected.Args, args) {
		return fmt.Errorf("expected args %v, got %v", expected.Args, args)
	}
	
	if expected.Response != "" {
		fmt.Print(expected.Response)
	}
	
	return expected.Error
}

func equalArgs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// MockProcessManager mocks process operations
type MockProcessManager struct {
	Processes map[int]*MockProcess
	NextPID   int
}

type MockProcess struct {
	PID    int
	Name   string
	Status string
	CPU    float64
	Memory float64
}

func NewMockProcessManager() *MockProcessManager {
	return &MockProcessManager{
		Processes: make(map[int]*MockProcess),
		NextPID:   1000,
	}
}

func (m *MockProcessManager) StartProcess(name string) (int, error) {
	pid := m.NextPID
	m.NextPID++
	
	m.Processes[pid] = &MockProcess{
		PID:    pid,
		Name:   name,
		Status: "running",
		CPU:    10.5,
		Memory: 128.0,
	}
	
	return pid, nil
}

func (m *MockProcessManager) GetProcess(pid int) (*MockProcess, error) {
	proc, ok := m.Processes[pid]
	if !ok {
		return nil, fmt.Errorf("process not found")
	}
	return proc, nil
}

func (m *MockProcessManager) KillProcess(pid int) error {
	proc, ok := m.Processes[pid]
	if !ok {
		return fmt.Errorf("process not found")
	}
	proc.Status = "stopped"
	return nil
}

// Helper to mock exec.Command
type CommandRunner interface {
	Run(name string, args ...string) error
	Start(name string, args ...string) (*exec.Cmd, error)
}

type MockRunner struct {
	RunFunc   func(name string, args ...string) error
	StartFunc func(name string, args ...string) (*exec.Cmd, error)
}

func (m *MockRunner) Run(name string, args ...string) error {
	if m.RunFunc != nil {
		return m.RunFunc(name, args...)
	}
	return nil
}

func (m *MockRunner) Start(name string, args ...string) (*exec.Cmd, error) {
	if m.StartFunc != nil {
		return m.StartFunc(name, args...)
	}
	// Return a dummy command
	cmd := exec.Command("echo", "test")
	cmd.Process = &os.Process{Pid: 1234}
	return cmd, nil
}

// MockUserInput mocks interactive input
type MockUserInput struct {
	Responses []string
	Index     int
}

func NewMockUserInput(responses ...string) *MockUserInput {
	return &MockUserInput{
		Responses: responses,
	}
}

func (m *MockUserInput) ReadLine() (string, error) {
	if m.Index >= len(m.Responses) {
		return "", fmt.Errorf("no more responses")
	}
	resp := m.Responses[m.Index]
	m.Index++
	return resp, nil
}

// MockFileSystem mocks file operations
type MockFileSystem struct {
	Files map[string][]byte
	Dirs  map[string]bool
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string][]byte),
		Dirs:  make(map[string]bool),
	}
}

func (m *MockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	m.Files[name] = data
	return nil
}

func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	data, ok := m.Files[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	m.Dirs[path] = true
	// Also create parent directories
	parts := strings.Split(path, "/")
	for i := 1; i <= len(parts); i++ {
		m.Dirs[strings.Join(parts[:i], "/")] = true
	}
	return nil
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if _, ok := m.Files[name]; ok {
		return &mockFileInfo{name: name, isDir: false}, nil
	}
	if _, ok := m.Dirs[name]; ok {
		return &mockFileInfo{name: name, isDir: true}, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Remove(name string) error {
	delete(m.Files, name)
	delete(m.Dirs, name)
	return nil
}

func (m *MockFileSystem) RemoveAll(path string) error {
	// Remove all files/dirs starting with path
	for k := range m.Files {
		if strings.HasPrefix(k, path) {
			delete(m.Files, k)
		}
	}
	for k := range m.Dirs {
		if strings.HasPrefix(k, path) {
			delete(m.Dirs, k)
		}
	}
	return nil
}

func (m *MockFileSystem) Glob(pattern string) ([]string, error) {
	var matches []string
	// Simple pattern matching for tests
	for path := range m.Files {
		if strings.Contains(pattern, "*") {
			base := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(path, base) {
				matches = append(matches, path)
			}
		} else if path == pattern {
			matches = append(matches, path)
		}
	}
	return matches, nil
}

func (m *MockFileSystem) Exists(path string) bool {
	_, fileExists := m.Files[path]
	_, dirExists := m.Dirs[path]
	return fileExists || dirExists
}

// Add ProcessManager methods to MockProcessManager
func (m *MockProcessManager) FindProcess(pid int) (*os.Process, error) {
	proc, ok := m.Processes[pid]
	if !ok {
		return nil, fmt.Errorf("process not found")
	}
	if proc.Status == "stopped" {
		return &os.Process{Pid: pid}, nil
	}
	return &os.Process{Pid: pid}, nil
}

func (m *MockProcessManager) Signal(p *os.Process, sig os.Signal) error {
	proc, ok := m.Processes[p.Pid]
	if !ok {
		return fmt.Errorf("process not found")
	}
	if proc.Status != "running" {
		return fmt.Errorf("process not running")
	}
	return nil
}

// MockCmd implements the Cmd interface
type MockCmd struct {
	fullCmd  string
	commands map[string]struct {
		Output []byte
		Error  error
	}
	process *os.Process
}

func (c *MockCmd) Output() ([]byte, error) {
	// Find matching command pattern
	for pattern, response := range c.commands {
		if strings.Contains(c.fullCmd, pattern) {
			return response.Output, response.Error
		}
	}
	return nil, fmt.Errorf("command not found: %s", c.fullCmd)
}

func (c *MockCmd) Run() error {
	// Check if we have an error configured for this command
	for pattern, response := range c.commands {
		if strings.Contains(c.fullCmd, pattern) {
			return response.Error
		}
	}
	return nil
}
func (c *MockCmd) Start() error {
	// Set a fake process when starting
	if c.process == nil {
		c.process = &os.Process{Pid: 12345}
	}
	// Check if we have an error configured
	// First check exact command name match
	cmdName := strings.Split(c.fullCmd, " ")[0]
	if response, ok := c.commands[cmdName]; ok {
		return response.Error
	}
	// Fallback to pattern matching
	for pattern, response := range c.commands {
		if strings.Contains(c.fullCmd, pattern) {
			return response.Error
		}
	}
	return nil
}
func (c *MockCmd) Wait() error                         { return nil }
func (c *MockCmd) StdoutPipe() (io.ReadCloser, error) { return nil, nil }
func (c *MockCmd) StderrPipe() (io.ReadCloser, error) { return nil, nil }
func (c *MockCmd) SetDir(dir string)                  {}
func (c *MockCmd) SetEnv(env []string)                {}
func (c *MockCmd) Process() *os.Process                { return c.process }

// Command method for MockCommandExecutor
func (m *MockCommandExecutor) Command(name string, arg ...string) Cmd {
	cmdStr := name + " " + strings.Join(arg, " ")
	// Convert old format to new format
	commands := make(map[string]struct {
		Output []byte
		Error  error
	})
	
	// Find the matching command
	for _, mc := range m.Commands {
		if mc.Cmd == name {
			// Call the callback if present
			if mc.OnCall != nil {
				mc.OnCall()
			}
			
			// key := mc.Cmd
			// if len(mc.Args) > 0 {
			// 	key = mc.Cmd + " " + strings.Join(mc.Args, " ")
			// }
			commands[mc.Cmd] = struct {
				Output []byte
				Error  error
			}{
				Output: []byte(mc.Response),
				Error:  mc.Error,
			}
			// Use just the command name as key for simpler matching
			commands[name] = commands[mc.Cmd]
			break
		}
	}
	
	return &MockCmd{
		fullCmd:  cmdStr,
		commands: commands,
	}
}

// mockFileInfo implements os.FileInfo
type mockFileInfo struct {
	name  string
	isDir bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }