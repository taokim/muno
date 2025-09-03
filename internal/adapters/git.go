package adapters

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	
	"github.com/taokim/muno/internal/interfaces"
)

// RealGit implements interfaces.GitInterface using git command
type RealGit struct {
	executor interfaces.CommandExecutor
}

// NewRealGit creates a new real git implementation
func NewRealGit() *RealGit {
	return &RealGit{
		executor: NewRealCommandExecutor(),
	}
}

// NewRealGitWithExecutor creates a git implementation with a custom executor
func NewRealGitWithExecutor(executor interfaces.CommandExecutor) *RealGit {
	return &RealGit{
		executor: executor,
	}
}

// Clone implements GitInterface.Clone
func (g *RealGit) Clone(url, path string) error {
	_, err := g.executor.Execute("git", "clone", url, path)
	return err
}

// CloneWithOptions implements GitInterface.CloneWithOptions
func (g *RealGit) CloneWithOptions(url, path string, options ...string) error {
	args := append([]string{"clone"}, options...)
	args = append(args, url, path)
	_, err := g.executor.Execute("git", args...)
	return err
}

// Pull implements GitInterface.Pull
func (g *RealGit) Pull(path string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "pull")
	return err
}

// PullWithOptions implements GitInterface.PullWithOptions
func (g *RealGit) PullWithOptions(path string, options ...string) error {
	args := append([]string{"pull"}, options...)
	_, err := g.executor.ExecuteInDir(path, "git", args...)
	return err
}

// Push implements GitInterface.Push
func (g *RealGit) Push(path string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "push")
	return err
}

// PushWithOptions implements GitInterface.PushWithOptions
func (g *RealGit) PushWithOptions(path string, options ...string) error {
	args := append([]string{"push"}, options...)
	_, err := g.executor.ExecuteInDir(path, "git", args...)
	return err
}

// Fetch implements GitInterface.Fetch
func (g *RealGit) Fetch(path string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "fetch")
	return err
}

// FetchWithOptions implements GitInterface.FetchWithOptions
func (g *RealGit) FetchWithOptions(path string, options ...string) error {
	args := append([]string{"fetch"}, options...)
	_, err := g.executor.ExecuteInDir(path, "git", args...)
	return err
}

// Status implements GitInterface.Status
func (g *RealGit) Status(path string) (string, error) {
	output, err := g.executor.ExecuteInDir(path, "git", "status")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// StatusWithOptions implements GitInterface.StatusWithOptions
func (g *RealGit) StatusWithOptions(path string, options ...string) (string, error) {
	args := append([]string{"status"}, options...)
	output, err := g.executor.ExecuteInDir(path, "git", args...)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Branch implements GitInterface.Branch
func (g *RealGit) Branch(path string) (string, error) {
	output, err := g.executor.ExecuteInDir(path, "git", "branch")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// CurrentBranch implements GitInterface.CurrentBranch
func (g *RealGit) CurrentBranch(path string) (string, error) {
	output, err := g.executor.ExecuteInDir(path, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// RemoteURL implements GitInterface.RemoteURL
func (g *RealGit) RemoteURL(path string) (string, error) {
	output, err := g.executor.ExecuteInDir(path, "git", "remote", "get-url", "origin")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// HasChanges implements GitInterface.HasChanges
func (g *RealGit) HasChanges(path string) (bool, error) {
	// Check for unstaged changes
	output, err := g.executor.ExecuteInDir(path, "git", "diff", "--exit-code")
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return true, nil // Has unstaged changes
		}
		return false, err
	}
	
	// Check for staged changes
	output, err = g.executor.ExecuteInDir(path, "git", "diff", "--cached", "--exit-code")
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return true, nil // Has staged changes
		}
		return false, err
	}
	
	// Check for untracked files
	output, err = g.executor.ExecuteInDir(path, "git", "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return false, err
	}
	
	return len(bytes.TrimSpace(output)) > 0, nil
}

// IsRepo implements GitInterface.IsRepo
func (g *RealGit) IsRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Add implements GitInterface.Add
func (g *RealGit) Add(path string, files ...string) error {
	args := append([]string{"add"}, files...)
	_, err := g.executor.ExecuteInDir(path, "git", args...)
	return err
}

// AddAll implements GitInterface.AddAll
func (g *RealGit) AddAll(path string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "add", "-A")
	return err
}

// Commit implements GitInterface.Commit
func (g *RealGit) Commit(path, message string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "commit", "-m", message)
	return err
}

// CommitWithOptions implements GitInterface.CommitWithOptions
func (g *RealGit) CommitWithOptions(path, message string, options ...string) error {
	args := append([]string{"commit", "-m", message}, options...)
	_, err := g.executor.ExecuteInDir(path, "git", args...)
	return err
}

// Checkout implements GitInterface.Checkout
func (g *RealGit) Checkout(path, branch string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "checkout", branch)
	return err
}

// CheckoutNew implements GitInterface.CheckoutNew
func (g *RealGit) CheckoutNew(path, branch string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "checkout", "-b", branch)
	return err
}

// CreateBranch implements GitInterface.CreateBranch
func (g *RealGit) CreateBranch(path, branch string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "branch", branch)
	return err
}

// DeleteBranch implements GitInterface.DeleteBranch
func (g *RealGit) DeleteBranch(path, branch string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "branch", "-d", branch)
	return err
}

// ListBranches implements GitInterface.ListBranches
func (g *RealGit) ListBranches(path string) ([]string, error) {
	output, err := g.executor.ExecuteInDir(path, "git", "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	for _, line := range lines {
		if line != "" {
			branches = append(branches, line)
		}
	}
	
	return branches, nil
}

// Tag implements GitInterface.Tag
func (g *RealGit) Tag(path, tag string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "tag", tag)
	return err
}

// TagWithMessage implements GitInterface.TagWithMessage
func (g *RealGit) TagWithMessage(path, tag, message string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "tag", "-a", tag, "-m", message)
	return err
}

// ListTags implements GitInterface.ListTags
func (g *RealGit) ListTags(path string) ([]string, error) {
	output, err := g.executor.ExecuteInDir(path, "git", "tag")
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var tags []string
	for _, line := range lines {
		if line != "" {
			tags = append(tags, line)
		}
	}
	
	return tags, nil
}

// Reset implements GitInterface.Reset
func (g *RealGit) Reset(path string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "reset")
	return err
}

// ResetHard implements GitInterface.ResetHard
func (g *RealGit) ResetHard(path string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "reset", "--hard")
	return err
}

// ResetSoft implements GitInterface.ResetSoft
func (g *RealGit) ResetSoft(path string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "reset", "--soft")
	return err
}

// Diff implements GitInterface.Diff
func (g *RealGit) Diff(path string) (string, error) {
	output, err := g.executor.ExecuteInDir(path, "git", "diff")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// DiffStaged implements GitInterface.DiffStaged
func (g *RealGit) DiffStaged(path string) (string, error) {
	output, err := g.executor.ExecuteInDir(path, "git", "diff", "--staged")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// DiffWithBranch implements GitInterface.DiffWithBranch
func (g *RealGit) DiffWithBranch(path, branch string) (string, error) {
	output, err := g.executor.ExecuteInDir(path, "git", "diff", branch)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Log implements GitInterface.Log
func (g *RealGit) Log(path string, limit int) ([]string, error) {
	args := []string{"log", "--oneline"}
	if limit > 0 {
		args = append(args, fmt.Sprintf("-n%d", limit))
	}
	
	output, err := g.executor.ExecuteInDir(path, "git", args...)
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var logs []string
	for _, line := range lines {
		if line != "" {
			logs = append(logs, line)
		}
	}
	
	return logs, nil
}

// LogOneline implements GitInterface.LogOneline
func (g *RealGit) LogOneline(path string, limit int) ([]string, error) {
	return g.Log(path, limit)
}

// AddRemote implements GitInterface.AddRemote
func (g *RealGit) AddRemote(path, name, url string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "remote", "add", name, url)
	return err
}

// RemoveRemote implements GitInterface.RemoveRemote
func (g *RealGit) RemoveRemote(path, name string) error {
	_, err := g.executor.ExecuteInDir(path, "git", "remote", "remove", name)
	return err
}

// ListRemotes implements GitInterface.ListRemotes
func (g *RealGit) ListRemotes(path string) (map[string]string, error) {
	output, err := g.executor.ExecuteInDir(path, "git", "remote", "-v")
	if err != nil {
		return nil, err
	}
	
	remotes := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Fields(line)
		if len(parts) >= 2 && strings.HasSuffix(parts[2], "(fetch)") {
			remotes[parts[0]] = parts[1]
		}
	}
	
	return remotes, nil
}