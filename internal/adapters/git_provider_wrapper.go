package adapters

import (
	"strings"
	
	"github.com/taokim/muno/internal/git"
	"github.com/taokim/muno/internal/interfaces"
)

// GitProviderWrapper wraps RealGit to implement GitProvider interface
type GitProviderWrapper struct {
	*RealGit
	simpleGit *git.Git // SSH-aware git implementation
}

// NewGitProvider creates a new GitProvider
func NewGitProvider() interfaces.GitProvider {
	return &GitProviderWrapper{
		RealGit:   NewRealGit(),
		simpleGit: git.New(),
	}
}

// Clone implements GitProvider.Clone
func (g *GitProviderWrapper) Clone(url, path string, options interfaces.CloneOptions) error {
	// Use SSH-aware clone method with SSH preference from options
	return g.simpleGit.CloneWithSSHPreference(url, path, options.SSHPreference)
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
	// Initialize result
	status := &interfaces.GitStatus{
		Branch:       "main",
		IsClean:      true,
		HasUntracked: false,
		HasStaged:    false,
		HasModified:  false,
		HasChanges:   false,
		Files:        []interfaces.GitFileStatus{},
		Ahead:        0,
		Behind:       0,
	}
	
	// Get current branch
	branch, err := g.RealGit.CurrentBranch(path)
	if err == nil && branch != "" {
		status.Branch = branch
	}
	
	// Get the raw git status output
	statusOutput, err := g.RealGit.Status(path)
	if err != nil {
		// If git command fails, return default status
		return status, nil
	}
	
	// Parse the git status output
	// Git status --short format:
	// XY filename
	// X is the staged status, Y is the unstaged status
	// ?? for untracked files
	// M  for staged modifications
	//  M for unstaged modifications
	// A  for added files
	// D  for deleted files
	
	if statusOutput != "" {
		lines := strings.Split(strings.TrimSpace(statusOutput), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			
			// The status codes are in the first two characters
			if len(line) >= 2 {
				stagedStatus := line[0]
				unstagedStatus := line[1]
				
				// Check for untracked files
				if stagedStatus == '?' && unstagedStatus == '?' {
					status.HasUntracked = true
				}
				
				// Check for staged changes
				if stagedStatus != ' ' && stagedStatus != '?' {
					status.HasStaged = true
				}
				
				// Check for unstaged modifications
				if unstagedStatus != ' ' && unstagedStatus != '?' {
					status.HasModified = true
				}
				
				// Parse file information if needed
				if len(line) > 3 {
					fileName := strings.TrimSpace(line[3:])
					fileStatus := interfaces.GitFileStatus{
						Path: fileName,
					}
					
					// Determine file status
					if stagedStatus == '?' && unstagedStatus == '?' {
						fileStatus.Status = "untracked"
						fileStatus.Staged = false
					} else if stagedStatus == 'A' {
						fileStatus.Status = "added"
						fileStatus.Staged = true
					} else if stagedStatus == 'M' || unstagedStatus == 'M' {
						fileStatus.Status = "modified"
						fileStatus.Staged = (stagedStatus == 'M')
					} else if stagedStatus == 'D' || unstagedStatus == 'D' {
						fileStatus.Status = "deleted"
						fileStatus.Staged = (stagedStatus == 'D')
					}
					
					status.Files = append(status.Files, fileStatus)
				}
			}
		}
		
		// Set overall status flags
		status.IsClean = !status.HasUntracked && !status.HasStaged && !status.HasModified
		status.HasChanges = !status.IsClean
		
		// Count files for compatibility
		for _, file := range status.Files {
			if file.Status == "modified" {
				status.FilesModified++
			} else if file.Status == "added" {
				status.FilesAdded++
			}
		}
	}
	
	return status, nil
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