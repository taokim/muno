package adapters

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	
	"github.com/taokim/muno/internal/interfaces"
)

// RealCommandExecutor implements interfaces.CommandExecutor using os/exec
type RealCommandExecutor struct{}

// NewRealCommandExecutor creates a new real command executor
func NewRealCommandExecutor() *RealCommandExecutor {
	return &RealCommandExecutor{}
}

func (e *RealCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

func (e *RealCommandExecutor) ExecuteWithInput(input string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = bytes.NewBufferString(input)
	return cmd.CombinedOutput()
}

func (e *RealCommandExecutor) ExecuteInDir(dir string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func (e *RealCommandExecutor) ExecuteWithEnv(env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	return cmd.CombinedOutput()
}

func (e *RealCommandExecutor) Start(name string, args ...string) (interfaces.Process, error) {
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &ProcessWrapper{cmd}, nil
}

func (e *RealCommandExecutor) StartInDir(dir string, name string, args ...string) (interfaces.Process, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &ProcessWrapper{cmd}, nil
}

// ProcessWrapper wraps exec.Cmd to implement interfaces.Process
type ProcessWrapper struct {
	cmd *exec.Cmd
}

// NewProcessWrapper creates a ProcessWrapper without starting the process
func NewProcessWrapper(name string, args ...string) *ProcessWrapper {
	cmd := exec.Command(name, args...)
	return &ProcessWrapper{cmd}
}

// Start starts the wrapped process
func (p *ProcessWrapper) Start() error {
	return p.cmd.Start()
}

func (p *ProcessWrapper) Wait() error {
	return p.cmd.Wait()
}

func (p *ProcessWrapper) Kill() error {
	if p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}

func (p *ProcessWrapper) Signal(sig os.Signal) error {
	if p.cmd.Process != nil {
		return p.cmd.Process.Signal(sig)
	}
	return nil
}

func (p *ProcessWrapper) Pid() int {
	if p.cmd.Process != nil {
		return p.cmd.Process.Pid
	}
	return -1
}

func (p *ProcessWrapper) StdoutPipe() (io.ReadCloser, error) {
	return p.cmd.StdoutPipe()
}

func (p *ProcessWrapper) StderrPipe() (io.ReadCloser, error) {
	return p.cmd.StderrPipe()
}

func (p *ProcessWrapper) StdinPipe() (io.WriteCloser, error) {
	return p.cmd.StdinPipe()
}