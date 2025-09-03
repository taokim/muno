package adapters

import (
	"fmt"
	"github.com/taokim/muno/internal/git"
	"github.com/taokim/muno/internal/interfaces"
)

// GitInterfaceWrapper wraps interfaces.GitInterface to implement git.Interface
type GitInterfaceWrapper struct {
	git interfaces.GitInterface
}

// NewGitInterfaceWrapper creates a new wrapper
func NewGitInterfaceWrapper(gitImpl interfaces.GitInterface) git.Interface {
	return &GitInterfaceWrapper{git: gitImpl}
}

// Clone implements git.Interface
func (w *GitInterfaceWrapper) Clone(url, path string) error {
	if w.git == nil {
		return fmt.Errorf("git interface is nil")
	}
	return w.git.Clone(url, path)
}

// Pull implements git.Interface
func (w *GitInterfaceWrapper) Pull(path string) error {
	if w.git == nil {
		return fmt.Errorf("git interface is nil")
	}
	return w.git.Pull(path)
}

// Status implements git.Interface
func (w *GitInterfaceWrapper) Status(path string) (string, error) {
	if w.git == nil {
		return "", fmt.Errorf("git interface is nil")
	}
	return w.git.Status(path)
}

// Commit implements git.Interface
func (w *GitInterfaceWrapper) Commit(path, message string) error {
	if w.git == nil {
		return fmt.Errorf("git interface is nil")
	}
	return w.git.Commit(path, message)
}

// Push implements git.Interface
func (w *GitInterfaceWrapper) Push(path string) error {
	if w.git == nil {
		return fmt.Errorf("git interface is nil")
	}
	return w.git.Push(path)
}

// Add implements git.Interface - note the signature difference
func (w *GitInterfaceWrapper) Add(path, pattern string) error {
	if w.git == nil {
		return fmt.Errorf("git interface is nil")
	}
	// Convert the pattern-based Add to file-based Add
	// In most cases, pattern is a single file or "."
	if pattern == "." {
		return w.git.AddAll(path)
	}
	return w.git.Add(path, pattern)
}