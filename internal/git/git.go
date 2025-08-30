package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Repository represents a git repository
type Repository struct {
	Name   string
	Path   string
	URL    string
	Branch string
	Groups []string
	Agent  string
}

// Status represents the git status of a repository
type Status struct {
	Name     string
	Path     string
	Branch   string
	Clean    bool
	Modified []string
	Ahead    int
	Behind   int
	Error    error
}

// Manager handles git operations for multiple repositories
type Manager struct {
	WorkspacePath string
	Repositories  []Repository
}

// NewManager creates a new git manager
func NewManager(workspacePath string, repos []Repository) *Manager {
	return &Manager{
		WorkspacePath: workspacePath,
		Repositories:  repos,
	}
}

// Clone clones all repositories in parallel
func (m *Manager) Clone() error {
	var wg sync.WaitGroup
	errors := make(chan error, len(m.Repositories))

	for _, repo := range m.Repositories {
		wg.Add(1)
		go func(r Repository) {
			defer wg.Done()
			if err := m.cloneRepo(r); err != nil {
				errors <- fmt.Errorf("%s: %w", r.Name, err)
			}
		}(repo)
	}

	wg.Wait()
	close(errors)

	// Collect any errors
	var errs []string
	for err := range errors {
		errs = append(errs, err.Error())
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("clone errors:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}

// cloneRepo clones a single repository
func (m *Manager) cloneRepo(repo Repository) error {
	repoPath := filepath.Join(m.WorkspacePath, repo.Path)
	
	// Check if already exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		fmt.Printf("  ✓ %s (already exists)\n", repo.Name)
		return nil
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(repoPath), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Clone repository
	fmt.Printf("  → Cloning %s...\n", repo.Name)
	cmd := exec.Command("git", "clone", "-b", repo.Branch, repo.URL, repoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	fmt.Printf("  ✓ %s cloned\n", repo.Name)
	return nil
}

// DEPRECATED: Use CloneMissing() + Pull operations instead
// Sync pulls latest changes for all repositories using rebase
func (m *Manager) Sync() error {
	fmt.Println("🔄 Syncing repositories...")
	
	var wg sync.WaitGroup
	errors := make(chan error, len(m.Repositories))

	for _, repo := range m.Repositories {
		wg.Add(1)
		go func(r Repository) {
			defer wg.Done()
			if err := m.syncRepo(r); err != nil {
				errors <- fmt.Errorf("%s: %w", r.Name, err)
			}
		}(repo)
	}

	wg.Wait()
	close(errors)

	// Collect any errors
	var errs []string
	for err := range errors {
		errs = append(errs, err.Error())
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("sync errors:\n%s", strings.Join(errs, "\n"))
	}

	fmt.Println("✅ Sync completed")
	return nil
}

// syncRepo syncs a single repository
func (m *Manager) syncRepo(repo Repository) error {
	repoPath := filepath.Join(m.WorkspacePath, repo.Path)
	
	// Check if exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		// Clone if doesn't exist
		return m.cloneRepo(repo)
	}

	// Pull latest changes with rebase
	cmd := exec.Command("git", "pull", "--rebase")
	cmd.Dir = repoPath
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w\n%s", err, output)
	}

	return nil
}

// CloneMissing clones any missing repositories
func (m *Manager) CloneMissing() error {
	fmt.Println("🔍 Checking for missing repositories...")
	
	var wg sync.WaitGroup
	errors := make(chan error, len(m.Repositories))
	cloned := 0
	var mu sync.Mutex

	for _, repo := range m.Repositories {
		wg.Add(1)
		go func(r Repository) {
			defer wg.Done()
			repoPath := filepath.Join(m.WorkspacePath, r.Path)
			
			// Check if repository exists
			if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
				// Clone if doesn't exist
				if err := m.cloneRepo(r); err != nil {
					errors <- fmt.Errorf("%s: %w", r.Name, err)
				} else {
					mu.Lock()
					cloned++
					mu.Unlock()
				}
			}
		}(repo)
	}

	wg.Wait()
	close(errors)

	// Collect any errors
	var errs []string
	for err := range errors {
		errs = append(errs, err.Error())
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("clone errors:\n%s", strings.Join(errs, "\n"))
	}

	if cloned > 0 {
		fmt.Printf("✅ Cloned %d missing repositories\n", cloned)
	} else {
		fmt.Println("✅ All repositories already cloned")
	}
	return nil
}

// GetRemotes gets all remote URLs for a repository
func (g *Git) GetRemotes(repoPath string) (map[string]string, error) {
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	remotes := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && strings.Contains(line, "(fetch)") {
			remotes[parts[0]] = parts[1]
		}
	}

	return remotes, nil
}

// CurrentBranch gets the current branch name
func (g *Git) CurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		// Try alternative method for older git versions
		cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = repoPath
		output, err = cmd.Output()
		if err != nil {
			return "", err
		}
	}
	return strings.TrimSpace(string(output)), nil
}

// Status gets the status of all repositories
func (m *Manager) Status() ([]Status, error) {
	var statuses []Status
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, repo := range m.Repositories {
		wg.Add(1)
		go func(r Repository) {
			defer wg.Done()
			status := m.getRepoStatus(r)
			mu.Lock()
			statuses = append(statuses, status)
			mu.Unlock()
		}(repo)
	}

	wg.Wait()
	return statuses, nil
}

// getRepoStatus gets the status of a single repository
func (m *Manager) getRepoStatus(repo Repository) Status {
	repoPath := filepath.Join(m.WorkspacePath, repo.Path)
	status := Status{
		Name: repo.Name,
		Path: repo.Path,
	}

	// Check if repository exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		status.Error = fmt.Errorf("not cloned")
		return status
	}

	// Get current branch
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	if output, err := cmd.Output(); err == nil {
		status.Branch = strings.TrimSpace(string(output))
	}

	// Check if working directory is clean
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	if output, err := cmd.Output(); err == nil {
		if len(output) == 0 {
			status.Clean = true
		} else {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if line != "" {
					status.Modified = append(status.Modified, line)
				}
			}
		}
	}

	// Get ahead/behind counts
	cmd = exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
	cmd.Dir = repoPath
	if output, err := cmd.Output(); err == nil {
		fmt.Sscanf(string(output), "%d\t%d", &status.Ahead, &status.Behind)
	}

	return status
}

// ForAll runs a command in all repositories (now with parallel execution by default)
func (m *Manager) ForAll(command string, args ...string) error {
	return m.ForAllWithOptions(command, args, DefaultExecutorOptions())
}

// ForAllWithOptions runs a command in all repositories with options
func (m *Manager) ForAllWithOptions(command string, args []string, opts ExecutorOptions) error {
	// If the command is not git, run it directly
	if command != "git" {
		// For non-git commands, run them directly in each repo
		fmt.Printf("🔧 Running command in all repositories: %s %s\n", command, strings.Join(args, " "))
		
		for _, repo := range m.Repositories {
			repoPath := filepath.Join(m.WorkspacePath, repo.Path)
			
			// Check if repository exists
			if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
				fmt.Printf("  ⚠️  %s: not cloned\n", repo.Name)
				continue
			}

			fmt.Printf("\n[%s]\n", repo.Name)
			
			cmd := exec.Command(command, args...)
			cmd.Dir = repoPath
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			if err := cmd.Run(); err != nil {
				fmt.Printf("  ❌ Error: %v\n", err)
			}
		}
		return nil
	}
	
	// For git commands, use the new parallel executor
	if len(args) == 0 {
		return fmt.Errorf("no git command specified")
	}
	
	gitCommand := args[0]
	gitArgs := args[1:]
	
	results, err := m.ExecuteInRepos(gitCommand, gitArgs, opts)
	if err != nil {
		return err
	}
	
	// Check if any failed
	for _, result := range results {
		if !result.Success {
			return fmt.Errorf("command failed in one or more repositories")
		}
	}
	
	return nil
}

// Commit creates a commit in all repositories with changes
func (m *Manager) Commit(message string, opts ExecutorOptions) ([]CommandResult, error) {
	if message == "" {
		return nil, fmt.Errorf("commit message cannot be empty")
	}
	
	// First, check which repos have changes
	statusResults, err := m.ExecuteInRepos("status", []string{"--porcelain"}, opts)
	if err != nil {
		return nil, fmt.Errorf("checking status: %w", err)
	}
	
	// Filter repos with changes
	reposWithChanges := []string{}
	for _, result := range statusResults {
		if result.Success && len(strings.TrimSpace(result.Output)) > 0 {
			reposWithChanges = append(reposWithChanges, result.RepoName)
		}
	}
	
	if len(reposWithChanges) == 0 {
		fmt.Println("ℹ️  No repositories have uncommitted changes")
		return []CommandResult{}, nil
	}
	
	fmt.Printf("📝 Found changes in %d repositories: %s\n", 
		len(reposWithChanges), strings.Join(reposWithChanges, ", "))
	
	// Add all changes
	fmt.Println("\n➕ Adding changes...")
	addResults, err := m.ExecuteInRepos("add", []string{"-A"}, opts)
	if err != nil {
		return addResults, fmt.Errorf("adding changes: %w", err)
	}
	
	// Commit with message
	fmt.Printf("\n💾 Committing with message: %q\n", message)
	commitResults, err := m.ExecuteInRepos("commit", []string{"-m", message}, opts)
	if err != nil {
		return commitResults, fmt.Errorf("committing changes: %w", err)
	}
	
	return commitResults, nil
}

// Push pushes changes to remote repositories
func (m *Manager) Push(opts ExecutorOptions) ([]CommandResult, error) {
	return m.PushWithOptions("", "", opts)
}

// PushWithOptions pushes changes with specific remote and branch
func (m *Manager) PushWithOptions(remote, branch string, opts ExecutorOptions) ([]CommandResult, error) {
	args := []string{}
	
	// Add remote if specified
	if remote != "" {
		args = append(args, remote)
		// Add branch if specified
		if branch != "" {
			args = append(args, branch)
		}
	}
	
	// Add any additional push options
	if len(args) == 0 {
		// Default push (to tracked branch)
		fmt.Println("⬆️  Pushing to remote repositories...")
	} else {
		fmt.Printf("⬆️  Pushing to %s...\n", strings.Join(args, " "))
	}
	
	return m.ExecuteInRepos("push", args, opts)
}

// Pull pulls changes from remote repositories
func (m *Manager) Pull(opts ExecutorOptions) ([]CommandResult, error) {
	return m.PullWithOptions("", "", false, opts)
}

// PullWithOptions pulls changes with specific options
func (m *Manager) PullWithOptions(remote, branch string, rebase bool, opts ExecutorOptions) ([]CommandResult, error) {
	args := []string{}
	
	// Add rebase flag if requested
	if rebase {
		args = append(args, "--rebase")
	}
	
	// Add remote if specified
	if remote != "" {
		args = append(args, remote)
		// Add branch if specified
		if branch != "" {
			args = append(args, branch)
		}
	}
	
	if rebase {
		fmt.Println("⬇️  Pulling with rebase from remote repositories...")
	} else {
		fmt.Println("⬇️  Pulling from remote repositories...")
	}
	
	return m.ExecuteInRepos("pull", args, opts)
}

// Fetch fetches changes from remote repositories
func (m *Manager) Fetch(opts ExecutorOptions) ([]CommandResult, error) {
	return m.FetchWithOptions("", false, false, opts)
}

// FetchWithOptions fetches changes with specific options
func (m *Manager) FetchWithOptions(remote string, all, prune bool, opts ExecutorOptions) ([]CommandResult, error) {
	args := []string{}
	
	// Add flags
	if all {
		args = append(args, "--all")
	}
	if prune {
		args = append(args, "--prune")
	}
	
	// Add remote if specified and not using --all
	if remote != "" && !all {
		args = append(args, remote)
	}
	
	fmt.Println("🔄 Fetching from remote repositories...")
	return m.ExecuteInRepos("fetch", args, opts)
}

// GetRepositories returns the list of repositories
func (m *Manager) GetRepositories() []Repository {
	// Return a copy to prevent external modifications
	repos := make([]Repository, len(m.Repositories))
	copy(repos, m.Repositories)
	return repos
}