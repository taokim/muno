package manager

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// CommandExecutor interface for executing system commands
type CommandExecutor interface {
	Command(name string, arg ...string) Cmd
}

// Cmd interface wraps exec.Cmd functionality
type Cmd interface {
	Output() ([]byte, error)
	Run() error
	Start() error
	Wait() error
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	SetDir(dir string)
	SetEnv(env []string)
	Process() *os.Process
}

// FileSystem interface for file operations
type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
	Remove(name string) error
	RemoveAll(path string) error
	Glob(pattern string) ([]string, error)
}

// ProcessManager interface for process operations
type ProcessManager interface {
	FindProcess(pid int) (*os.Process, error)
	Signal(p *os.Process, sig os.Signal) error
}

// Default implementations that use real system calls
type RealCommandExecutor struct{}

func (r RealCommandExecutor) Command(name string, arg ...string) Cmd {
	return &RealCmd{cmd: exec.Command(name, arg...)}
}

type RealCmd struct {
	cmd *exec.Cmd
}

func (c *RealCmd) Output() ([]byte, error)         { return c.cmd.Output() }
func (c *RealCmd) Run() error                      { return c.cmd.Run() }
func (c *RealCmd) Start() error                    { return c.cmd.Start() }
func (c *RealCmd) Wait() error                     { return c.cmd.Wait() }
func (c *RealCmd) StdoutPipe() (io.ReadCloser, error) { return c.cmd.StdoutPipe() }
func (c *RealCmd) StderrPipe() (io.ReadCloser, error) { return c.cmd.StderrPipe() }
func (c *RealCmd) SetDir(dir string)               { c.cmd.Dir = dir }
func (c *RealCmd) SetEnv(env []string)             { c.cmd.Env = env }
func (c *RealCmd) Process() *os.Process            { return c.cmd.Process }

type RealFileSystem struct{}

func (r RealFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (r RealFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (r RealFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (r RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (r RealFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (r RealFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (r RealFileSystem) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

type RealProcessManager struct{}

func (r RealProcessManager) FindProcess(pid int) (*os.Process, error) {
	return os.FindProcess(pid)
}

func (r RealProcessManager) Signal(p *os.Process, sig os.Signal) error {
	return p.Signal(sig)
}