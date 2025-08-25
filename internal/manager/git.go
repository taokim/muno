package manager

import (
	"fmt"
	"strings"
	
	"github.com/taokim/repo-claude/internal/git"
)

// GitOptions contains common options for git commands
type GitOptions struct {
	// ExcludeRoot excludes the root repository
	ExcludeRoot bool
	// Parallel runs operations in parallel
	Parallel bool
	// MaxParallel limits parallel operations
	MaxParallel int
	// Quiet suppresses output
	Quiet bool
	// Verbose shows detailed output
	Verbose bool
}

// DefaultGitOptions returns default git options
func DefaultGitOptions() GitOptions {
	return GitOptions{
		Parallel:    true,
		MaxParallel: 4,
	}
}

// toExecutorOptions converts GitOptions to git.ExecutorOptions
func (o GitOptions) toExecutorOptions() git.ExecutorOptions {
	return git.ExecutorOptions{
		IncludeRoot: !o.ExcludeRoot, // Include root by default unless excluded
		ExcludeRoot: o.ExcludeRoot,
		Parallel:    o.Parallel,
		MaxParallel: o.MaxParallel,
		Quiet:       o.Quiet,
		Verbose:     o.Verbose,
	}
}

// ForAll runs a command in all repositories with options
func (m *Manager) ForAllWithOptions(command string, args []string, opts GitOptions) error {
	if m.GitManager == nil {
		return fmt.Errorf("no git manager initialized")
	}
	
	execOpts := opts.toExecutorOptions()
	return m.GitManager.ForAllWithOptions(command, args, execOpts)
}

// GitCommit commits changes across repositories
func (m *Manager) GitCommit(message string, opts GitOptions) error {
	if m.GitManager == nil {
		return fmt.Errorf("no git manager initialized")
	}
	
	if message == "" {
		return fmt.Errorf("commit message cannot be empty")
	}
	
	execOpts := opts.toExecutorOptions()
	results, err := m.GitManager.Commit(message, execOpts)
	if err != nil {
		return err
	}
	
	// Check for any failures
	var failures []string
	for _, result := range results {
		if !result.Success && result.Error != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", result.RepoName, result.Error))
		}
	}
	
	if len(failures) > 0 {
		return fmt.Errorf("commit failed in some repositories:\n%s", strings.Join(failures, "\n"))
	}
	
	return nil
}

// GitPush pushes changes to remote repositories
func (m *Manager) GitPush(opts GitOptions) error {
	if m.GitManager == nil {
		return fmt.Errorf("no git manager initialized")
	}
	
	execOpts := opts.toExecutorOptions()
	results, err := m.GitManager.Push(execOpts)
	if err != nil {
		return err
	}
	
	// Check for any failures
	var failures []string
	for _, result := range results {
		if !result.Success && result.Error != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", result.RepoName, result.Error))
		}
	}
	
	if len(failures) > 0 {
		return fmt.Errorf("push failed in some repositories:\n%s", strings.Join(failures, "\n"))
	}
	
	return nil
}

// GitPull pulls changes from remote repositories
func (m *Manager) GitPull(rebase bool, opts GitOptions) error {
	if m.GitManager == nil {
		return fmt.Errorf("no git manager initialized")
	}
	
	execOpts := opts.toExecutorOptions()
	results, err := m.GitManager.PullWithOptions("", "", rebase, execOpts)
	if err != nil {
		return err
	}
	
	// Check for any failures
	var failures []string
	for _, result := range results {
		if !result.Success && result.Error != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", result.RepoName, result.Error))
		}
	}
	
	if len(failures) > 0 {
		return fmt.Errorf("pull failed in some repositories:\n%s", strings.Join(failures, "\n"))
	}
	
	return nil
}

// GitFetch fetches changes from remote repositories
func (m *Manager) GitFetch(all, prune bool, opts GitOptions) error {
	if m.GitManager == nil {
		return fmt.Errorf("no git manager initialized")
	}
	
	execOpts := opts.toExecutorOptions()
	results, err := m.GitManager.FetchWithOptions("", all, prune, execOpts)
	if err != nil {
		return err
	}
	
	// Check for any failures
	var failures []string
	for _, result := range results {
		if !result.Success && result.Error != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", result.RepoName, result.Error))
		}
	}
	
	if len(failures) > 0 {
		return fmt.Errorf("fetch failed in some repositories:\n%s", strings.Join(failures, "\n"))
	}
	
	return nil
}

// GitStatus shows status across all repositories (enhanced version)
func (m *Manager) GitStatus(opts GitOptions) error {
	if m.GitManager == nil {
		return fmt.Errorf("no git manager initialized")
	}
	
	execOpts := opts.toExecutorOptions()
	execOpts.Verbose = true // Always show output for status
	
	results, err := m.GitManager.ExecuteInRepos("status", []string{"-sb"}, execOpts)
	if err != nil {
		return err
	}
	
	// Display results in a nice format
	fmt.Println("\n📊 Git Status Summary")
	fmt.Println(strings.Repeat("=", 60))
	
	cleanCount := 0
	dirtyCount := 0
	
	for _, result := range results {
		if result.Success {
			output := strings.TrimSpace(result.Output)
			lines := strings.Split(output, "\n")
			
			// Parse branch info from first line
			branchInfo := ""
			if len(lines) > 0 {
				branchInfo = strings.TrimPrefix(lines[0], "## ")
			}
			
			// Check if clean (only branch info line)
			isClean := len(lines) == 1
			
			if isClean {
				cleanCount++
				fmt.Printf("  ✅ %s [%s] - clean\n", result.RepoName, branchInfo)
			} else {
				dirtyCount++
				fmt.Printf("  ⚠️  %s [%s] - %d changes\n", result.RepoName, branchInfo, len(lines)-1)
				if opts.Verbose {
					for i := 1; i < len(lines) && i <= 5; i++ {
						fmt.Printf("      %s\n", lines[i])
					}
					if len(lines) > 6 {
						fmt.Printf("      ... and %d more\n", len(lines)-6)
					}
				}
			}
		} else {
			fmt.Printf("  ❌ %s - error: %v\n", result.RepoName, result.Error)
		}
	}
	
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Summary: %d clean, %d with changes\n", cleanCount, dirtyCount)
	
	return nil
}