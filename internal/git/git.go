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

// Sync pulls latest changes for all repositories
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

	// Pull latest changes
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = repoPath
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w\n%s", err, output)
	}

	return nil
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

// ForAll runs a command in all repositories
func (m *Manager) ForAll(command string, args ...string) error {
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

// GetRepositories returns the list of repositories
func (m *Manager) GetRepositories() []Repository {
	return m.Repositories
}