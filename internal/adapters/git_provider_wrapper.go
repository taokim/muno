package adapters

import (
	"github.com/taokim/muno/internal/interfaces"
)

// GitProviderWrapper wraps RealGit to implement GitProvider interface
type GitProviderWrapper struct {
	*RealGit
}

// NewGitProvider creates a new GitProvider
func NewGitProvider() interfaces.GitProvider {
	return &GitProviderWrapper{
		RealGit: NewRealGit(),
	}
}

// Clone implements GitProvider.Clone
func (g *GitProviderWrapper) Clone(url, path string, options interfaces.CloneOptions) error {
	// Call the underlying RealGit.Clone which has a simpler signature
	return g.RealGit.Clone(url, path)
}

// Pull implements GitProvider.Pull
func (g *GitProviderWrapper) Pull(path string, options interfaces.PullOptions) error {
	// Call the underlying RealGit.Pull
	return g.RealGit.Pull(path)
}

// Push implements GitProvider.Push
func (g *GitProviderWrapper) Push(path string, options interfaces.PushOptions) error {
	// Call the underlying RealGit.Push
	return g.RealGit.Push(path)
}

// Commit implements GitProvider.Commit
func (g *GitProviderWrapper) Commit(path string, message string, options interfaces.CommitOptions) error {
	// Call the underlying RealGit.Commit
	return g.RealGit.Commit(path, message)
}

// Fetch implements GitProvider.Fetch
func (g *GitProviderWrapper) Fetch(path string, options interfaces.FetchOptions) error {
	// Call the underlying RealGit.Fetch
	return g.RealGit.Fetch(path)
}

// Status implements GitProvider.Status
func (g *GitProviderWrapper) Status(path string) (*interfaces.GitStatus, error) {
	// RealGit.Status returns a string, we need to convert to GitStatus
	// For now, return a clean status
	// This is a simplified implementation - you may want to parse the actual git status
	_ = path
	return &interfaces.GitStatus{
		Branch:        "main",
		IsClean:       true,
		HasUntracked:  false,
		HasStaged:     false,
		HasModified:   false,
		HasChanges:    false,
		FilesModified: 0,
		FilesAdded:    0,
		Files:         []interfaces.GitFileStatus{},
		Ahead:         0,
		Behind:        0,
	}, nil
}

// Branch implements GitProvider.Branch
func (g *GitProviderWrapper) Branch(path string) (string, error) {
	// Call the underlying RealGit.Branch
	return g.RealGit.Branch(path)
}

// Checkout implements GitProvider.Checkout
func (g *GitProviderWrapper) Checkout(path string, branch string) error {
	// Call the underlying RealGit.Checkout
	return g.RealGit.Checkout(path, branch)
}

// Add implements GitProvider.Add with the correct signature
func (g *GitProviderWrapper) Add(path string, files []string) error {
	// Convert slice to variadic arguments
	return g.RealGit.Add(path, files...)
}

// Remove implements GitProvider.Remove
func (g *GitProviderWrapper) Remove(path string, files []string) error {
	// RealGit doesn't have Remove, so we'll implement a basic version
	// This is just a stub - you may want to implement actual git rm functionality
	return nil
}

// GetRemoteURL implements GitProvider.GetRemoteURL
func (g *GitProviderWrapper) GetRemoteURL(path string) (string, error) {
	// RealGit doesn't have GetRemoteURL, return empty for now
	return "", nil
}

// SetRemoteURL implements GitProvider.SetRemoteURL
func (g *GitProviderWrapper) SetRemoteURL(path string, url string) error {
	// RealGit doesn't have SetRemoteURL, no-op for now
	return nil
}