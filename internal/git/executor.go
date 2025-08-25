package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ExecutorOptions provides common options for git operations
type ExecutorOptions struct {
	// IncludeRoot indicates whether to include the root project directory
	IncludeRoot bool
	// ExcludeRoot explicitly excludes root even if IncludeRoot is true
	ExcludeRoot bool
	// Parallel indicates whether to run operations in parallel
	Parallel bool
	// MaxParallel limits the number of parallel operations (0 = unlimited)
	MaxParallel int
	// FailFast stops on first error when running in parallel
	FailFast bool
	// Quiet suppresses non-error output
	Quiet bool
	// Verbose shows detailed command output
	Verbose bool
}

// DefaultExecutorOptions returns default options with root included and parallel execution
func DefaultExecutorOptions() ExecutorOptions {
	return ExecutorOptions{
		IncludeRoot: true,
		Parallel:    true,
		MaxParallel: 4, // Default to 4 parallel operations
		FailFast:    false,
	}
}

// CommandResult represents the result of a git command execution
type CommandResult struct {
	RepoName string
	RepoPath string
	Success  bool
	Output   string
	Error    error
}

// ExecuteInRepos runs a git command across repositories with unified options
func (m *Manager) ExecuteInRepos(command string, args []string, opts ExecutorOptions) ([]CommandResult, error) {
	// Build list of target paths
	targets := m.buildTargetPaths(opts)
	
	if len(targets) == 0 {
		return nil, fmt.Errorf("no repositories to operate on")
	}
	
	// Display what we're doing
	if !opts.Quiet {
		fmt.Printf("🔧 Running: git %s %s\n", command, strings.Join(args, " "))
		if opts.IncludeRoot && !opts.ExcludeRoot {
			fmt.Println("   Including root repository")
		}
		fmt.Printf("   Targets: %d repositories\n", len(targets))
		if opts.Parallel {
			fmt.Printf("   Mode: Parallel (max %d)\n", opts.MaxParallel)
		} else {
			fmt.Println("   Mode: Sequential")
		}
		fmt.Println(strings.Repeat("=", 60))
	}
	
	// Execute based on mode
	if opts.Parallel {
		return m.executeParallel(command, args, targets, opts)
	}
	return m.executeSequential(command, args, targets, opts)
}

// buildTargetPaths builds the list of paths to operate on
func (m *Manager) buildTargetPaths(opts ExecutorOptions) []targetPath {
	var targets []targetPath
	
	// Add root repository if requested
	if opts.IncludeRoot && !opts.ExcludeRoot {
		// Get the project root (parent of workspace)
		projectRoot := filepath.Dir(m.WorkspacePath)
		if _, err := os.Stat(filepath.Join(projectRoot, ".git")); err == nil {
			targets = append(targets, targetPath{
				Name: "root",
				Path: projectRoot,
			})
		}
	}
	
	// Add workspace repositories
	for _, repo := range m.Repositories {
		repoPath := filepath.Join(m.WorkspacePath, repo.Path)
		// Check if repository exists
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
			targets = append(targets, targetPath{
				Name: repo.Name,
				Path: repoPath,
			})
		} else if !opts.Quiet {
			fmt.Printf("  ⚠️  %s: not cloned, skipping\n", repo.Name)
		}
	}
	
	return targets
}

type targetPath struct {
	Name string
	Path string
}

// executeParallel runs commands in parallel across repositories
func (m *Manager) executeParallel(command string, args []string, targets []targetPath, opts ExecutorOptions) ([]CommandResult, error) {
	results := make([]CommandResult, 0, len(targets))
	resultChan := make(chan CommandResult, len(targets))
	errorChan := make(chan error, 1)
	
	// Use semaphore to limit parallelism
	var sem chan struct{}
	if opts.MaxParallel > 0 {
		sem = make(chan struct{}, opts.MaxParallel)
	}
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	stopped := false
	
	for _, target := range targets {
		if opts.FailFast {
			mu.Lock()
			if stopped {
				mu.Unlock()
				break
			}
			mu.Unlock()
		}
		
		wg.Add(1)
		go func(t targetPath) {
			defer wg.Done()
			
			// Acquire semaphore if using limited parallelism
			if sem != nil {
				sem <- struct{}{}
				defer func() { <-sem }()
			}
			
			// Check if we should stop (fail-fast mode)
			if opts.FailFast {
				mu.Lock()
				if stopped {
					mu.Unlock()
					return
				}
				mu.Unlock()
			}
			
			// Execute command
			result := m.executeCommand(command, args, t, opts)
			
			// Handle fail-fast
			if opts.FailFast && result.Error != nil {
				mu.Lock()
				if !stopped {
					stopped = true
					errorChan <- fmt.Errorf("%s: %w", t.Name, result.Error)
				}
				mu.Unlock()
			}
			
			resultChan <- result
		}(target)
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
	close(resultChan)
	close(errorChan)
	
	// Check for fail-fast error
	select {
	case err := <-errorChan:
		if err != nil {
			return nil, err
		}
	default:
	}
	
	// Collect results
	for result := range resultChan {
		results = append(results, result)
	}
	
	// Print summary if not quiet
	if !opts.Quiet {
		m.printExecutionSummary(results)
	}
	
	return results, nil
}

// executeSequential runs commands sequentially across repositories
func (m *Manager) executeSequential(command string, args []string, targets []targetPath, opts ExecutorOptions) ([]CommandResult, error) {
	results := make([]CommandResult, 0, len(targets))
	
	for _, target := range targets {
		result := m.executeCommand(command, args, target, opts)
		results = append(results, result)
		
		// Handle fail-fast
		if opts.FailFast && result.Error != nil {
			return results, fmt.Errorf("%s: %w", target.Name, result.Error)
		}
	}
	
	// Print summary if not quiet
	if !opts.Quiet {
		m.printExecutionSummary(results)
	}
	
	return results, nil
}

// executeCommand executes a single git command in a repository
func (m *Manager) executeCommand(command string, args []string, target targetPath, opts ExecutorOptions) CommandResult {
	result := CommandResult{
		RepoName: target.Name,
		RepoPath: target.Path,
	}
	
	// Build full command
	fullArgs := append([]string{command}, args...)
	cmd := exec.Command("git", fullArgs...)
	cmd.Dir = target.Path
	
	// Show progress if not quiet
	if !opts.Quiet {
		if opts.Verbose {
			fmt.Printf("\n[%s] %s\n", target.Name, target.Path)
			fmt.Printf("  → git %s\n", strings.Join(fullArgs, " "))
		} else {
			fmt.Printf("  → %s: ", target.Name)
		}
	}
	
	// Execute command
	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	
	if err != nil {
		result.Error = err
		result.Success = false
		if !opts.Quiet {
			if opts.Verbose {
				fmt.Printf("  ❌ Error: %v\n", err)
				if len(output) > 0 {
					fmt.Printf("  Output: %s\n", strings.TrimSpace(string(output)))
				}
			} else {
				fmt.Printf("❌\n")
			}
		}
	} else {
		result.Success = true
		if !opts.Quiet {
			if opts.Verbose {
				fmt.Printf("  ✅ Success\n")
				if len(output) > 0 && opts.Verbose {
					fmt.Printf("  Output: %s\n", strings.TrimSpace(string(output)))
				}
			} else {
				fmt.Printf("✅\n")
			}
		}
	}
	
	// Always show output in verbose mode or on error
	if opts.Verbose && len(output) > 0 {
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			fmt.Printf("    %s\n", line)
		}
	}
	
	return result
}

// printExecutionSummary prints a summary of execution results
func (m *Manager) printExecutionSummary(results []CommandResult) {
	successCount := 0
	failureCount := 0
	
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}
	
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📊 Summary: %d succeeded, %d failed (total: %d)\n", 
		successCount, failureCount, len(results))
	
	// Show failures if any
	if failureCount > 0 {
		fmt.Println("\n❌ Failed repositories:")
		for _, result := range results {
			if !result.Success {
				fmt.Printf("  - %s: %v\n", result.RepoName, result.Error)
			}
		}
	}
}