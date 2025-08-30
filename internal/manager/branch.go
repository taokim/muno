//go:build legacy
// +build legacy

package manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	
	"github.com/taokim/muno/internal/git"
)

// BranchCreateOptions contains options for creating branches
type BranchCreateOptions struct {
	BranchName string
	FromBranch string
	Repos      []string
}

// BranchListOptions contains options for listing branches
type BranchListOptions struct {
	ShowAll     bool
	ShowCurrent bool
}

// BranchCheckoutOptions contains options for checking out branches
type BranchCheckoutOptions struct {
	BranchName      string
	Repos           []string
	CreateIfMissing bool
}

// BranchDeleteOptions contains options for deleting branches
type BranchDeleteOptions struct {
	BranchName   string
	Repos        []string
	Force        bool
	DeleteRemote bool
}

// CreateBranch creates a new branch in multiple repositories
func (m *Manager) CreateBranch(opts BranchCreateOptions) error {
	fmt.Printf("🌿 Creating branch '%s' across repositories...\n", opts.BranchName)
	fmt.Println(strings.Repeat("=", 60))
	
	repos := m.getTargetRepos(opts.Repos)
	successCount := 0
	
	for _, repo := range repos {
		repoPath := filepath.Join(m.WorkspacePath, repo.Path)
		
		// Check if repository exists
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			fmt.Printf("⚠️  %s: not cloned\n", repo.Name)
			continue
		}
		
		// Create the branch
		args := []string{"checkout", "-b", opts.BranchName}
		if opts.FromBranch != "" {
			args = append(args, opts.FromBranch)
		}
		
		cmd := exec.Command("git", args...)
		cmd.Dir = repoPath
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Check if branch already exists
			if strings.Contains(string(output), "already exists") {
				fmt.Printf("⚠️  %s: branch already exists\n", repo.Name)
			} else {
				fmt.Printf("❌ %s: %s\n", repo.Name, strings.TrimSpace(string(output)))
			}
			continue
		}
		
		fmt.Printf("✅ %s: created branch '%s'\n", repo.Name, opts.BranchName)
		successCount++
	}
	
	fmt.Printf("\n📊 Created branch in %d/%d repositories\n", successCount, len(repos))
	return nil
}

// ListBranches lists branches across all repositories
func (m *Manager) ListBranches(opts BranchListOptions) error {
	fmt.Println("🌿 Branch Status Across Repositories")
	fmt.Println(strings.Repeat("=", 60))
	
	mainBranchCount := 0
	featureBranchCount := 0
	
	for _, repo := range m.GitManager.GetRepositories() {
		repoPath := filepath.Join(m.WorkspacePath, repo.Path)
		
		// Check if repository exists
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			fmt.Printf("\n⚠️  %s: not cloned\n", repo.Name)
			continue
		}
		
		// Get current branch
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = repoPath
		
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("\n❌ %s: failed to get branch\n", repo.Name)
			continue
		}
		
		currentBranch := strings.TrimSpace(string(output))
		
		if opts.ShowCurrent {
			fmt.Printf("%s: %s\n", repo.Name, currentBranch)
			continue
		}
		
		// Check if on main/master branch
		isMainBranch := currentBranch == "main" || currentBranch == "master" || currentBranch == repo.Branch
		
		status := "🟢"
		if isMainBranch {
			status = "🔵"
			mainBranchCount++
		} else {
			featureBranchCount++
		}
		
		fmt.Printf("%s %-20s → %s\n", status, repo.Name, currentBranch)
		
		// Show all branches if requested
		if opts.ShowAll {
			cmd = exec.Command("git", "branch", "-a")
			cmd.Dir = repoPath
			
			output, err = cmd.Output()
			if err == nil {
				branches := strings.Split(string(output), "\n")
				for _, branch := range branches {
					branch = strings.TrimSpace(branch)
					if branch != "" && !strings.HasPrefix(branch, "* ") {
						fmt.Printf("    %s\n", branch)
					}
				}
			}
		}
	}
	
	fmt.Printf("\n📊 Summary: %d on main branch, %d on feature branches\n", 
		mainBranchCount, featureBranchCount)
	
	if featureBranchCount > 0 {
		fmt.Println("\n💡 Tip: Use 'rc pr batch-create' to create PRs for all feature branches")
	}
	
	return nil
}

// CheckoutBranch checks out a branch in multiple repositories
func (m *Manager) CheckoutBranch(opts BranchCheckoutOptions) error {
	fmt.Printf("🔄 Checking out branch '%s' across repositories...\n", opts.BranchName)
	fmt.Println(strings.Repeat("=", 60))
	
	repos := m.getTargetRepos(opts.Repos)
	successCount := 0
	
	for _, repo := range repos {
		repoPath := filepath.Join(m.WorkspacePath, repo.Path)
		
		// Check if repository exists
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			fmt.Printf("⚠️  %s: not cloned\n", repo.Name)
			continue
		}
		
		// Try to checkout the branch
		cmd := exec.Command("git", "checkout", opts.BranchName)
		cmd.Dir = repoPath
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			if strings.Contains(string(output), "did not match any file") && opts.CreateIfMissing {
				// Create the branch if it doesn't exist
				cmd = exec.Command("git", "checkout", "-b", opts.BranchName)
				cmd.Dir = repoPath
				
				output, err = cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("❌ %s: %s\n", repo.Name, strings.TrimSpace(string(output)))
					continue
				}
				fmt.Printf("✅ %s: created and checked out '%s'\n", repo.Name, opts.BranchName)
			} else {
				fmt.Printf("❌ %s: %s\n", repo.Name, strings.TrimSpace(string(output)))
				continue
			}
		} else {
			fmt.Printf("✅ %s: checked out '%s'\n", repo.Name, opts.BranchName)
		}
		successCount++
	}
	
	fmt.Printf("\n📊 Checked out branch in %d/%d repositories\n", successCount, len(repos))
	return nil
}

// DeleteBranch deletes a branch from multiple repositories
func (m *Manager) DeleteBranch(opts BranchDeleteOptions) error {
	fmt.Printf("🗑️  Deleting branch '%s' from repositories...\n", opts.BranchName)
	fmt.Println(strings.Repeat("=", 60))
	
	repos := m.getTargetRepos(opts.Repos)
	successCount := 0
	
	for _, repo := range repos {
		repoPath := filepath.Join(m.WorkspacePath, repo.Path)
		
		// Check if repository exists
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			fmt.Printf("⚠️  %s: not cloned\n", repo.Name)
			continue
		}
		
		// Get current branch to ensure we're not on the branch being deleted
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = repoPath
		
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("❌ %s: failed to get current branch\n", repo.Name)
			continue
		}
		
		currentBranch := strings.TrimSpace(string(output))
		if currentBranch == opts.BranchName {
			// Switch to main/master first
			cmd = exec.Command("git", "checkout", repo.Branch)
			cmd.Dir = repoPath
			if err := cmd.Run(); err != nil {
				fmt.Printf("❌ %s: cannot delete current branch\n", repo.Name)
				continue
			}
		}
		
		// Delete local branch
		deleteFlag := "-d"
		if opts.Force {
			deleteFlag = "-D"
		}
		
		cmd = exec.Command("git", "branch", deleteFlag, opts.BranchName)
		cmd.Dir = repoPath
		
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("❌ %s: %s\n", repo.Name, strings.TrimSpace(string(output)))
			continue
		}
		
		fmt.Printf("✅ %s: deleted local branch '%s'\n", repo.Name, opts.BranchName)
		
		// Delete remote branch if requested
		if opts.DeleteRemote {
			cmd = exec.Command("git", "push", "origin", "--delete", opts.BranchName)
			cmd.Dir = repoPath
			
			output, err = cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("⚠️  %s: failed to delete remote branch: %s\n", 
					repo.Name, strings.TrimSpace(string(output)))
			} else {
				fmt.Printf("✅ %s: deleted remote branch '%s'\n", repo.Name, opts.BranchName)
			}
		}
		
		successCount++
	}
	
	fmt.Printf("\n📊 Deleted branch from %d/%d repositories\n", successCount, len(repos))
	return nil
}

// getTargetRepos returns the repositories to operate on
func (m *Manager) getTargetRepos(repoNames []string) []git.Repository {
	if len(repoNames) == 0 {
		return m.GitManager.GetRepositories()
	}
	
	var repos []git.Repository
	for _, name := range repoNames {
		for _, repo := range m.GitManager.GetRepositories() {
			if repo.Name == name || repo.Path == name {
				repos = append(repos, repo)
				break
			}
		}
	}
	
	return repos
}