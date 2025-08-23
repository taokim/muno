package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	
	"github.com/taokim/repo-claude/internal/git"
)

// PRListOptions contains options for listing PRs
type PRListOptions struct {
	State    string
	Limit    int
	Author   string
	Assignee string
	Label    string
}

// PRCreateOptions contains options for creating a PR
type PRCreateOptions struct {
	Repository string
	Title      string
	Body       string
	Base       string
	Draft      bool
	Reviewers  []string
	Assignees  []string
	Labels     []string
}

// PRStatusOptions contains options for showing PR status
type PRStatusOptions struct {
	Repository string
	Number     int
}

// PRMergeOptions contains options for merging a PR
type PRMergeOptions struct {
	Repository         string
	Number             int
	Method             string
	DeleteRemoteBranch bool
	DeleteLocalBranch  bool
}

// PRBatchCreateOptions contains options for creating PRs in multiple repos
type PRBatchCreateOptions struct {
	Title         string
	Body          string
	Base          string
	Draft         bool
	Reviewers     []string
	Assignees     []string
	Labels        []string
	Repos         []string
	SkipMainCheck bool
}

// PRInfo represents information about a pull request
type PRInfo struct {
	Number     int         `json:"number"`
	Title      string      `json:"title"`
	Author     interface{} `json:"author"` // Can be string or object with login field
	State      string      `json:"state"`
	URL        string      `json:"url"`
	Repository string
	Draft      bool        `json:"isDraft"`
	CreatedAt  string      `json:"createdAt"`
	UpdatedAt  string      `json:"updatedAt"`
}

// ListPRs lists pull requests across all repositories
func (m *Manager) ListPRs(opts PRListOptions) error {
	fmt.Println("📋 Pull Requests Across Repositories")
	fmt.Println(strings.Repeat("=", 60))
	
	totalPRs := 0
	
	for _, repo := range m.GitManager.GetRepositories() {
		repoPath := filepath.Join(m.WorkspacePath, repo.Path)
		
		// Check if repository exists
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			fmt.Printf("\n⚠️  %s: not cloned\n", repo.Name)
			continue
		}
		
		// Build gh pr list command
		args := []string{"pr", "list", "--json", "number,title,author,state,url,isDraft,createdAt,updatedAt"}
		
		if opts.State != "" && opts.State != "all" {
			args = append(args, "--state", opts.State)
		}
		
		if opts.Limit > 0 {
			args = append(args, "--limit", fmt.Sprintf("%d", opts.Limit))
		}
		
		if opts.Author != "" {
			args = append(args, "--author", opts.Author)
		}
		
		if opts.Assignee != "" {
			args = append(args, "--assignee", opts.Assignee)
		}
		
		if opts.Label != "" {
			args = append(args, "--label", opts.Label)
		}
		
		cmd := exec.Command("gh", args...)
		cmd.Dir = repoPath
		
		output, err := cmd.Output()
		if err != nil {
			// Check if it's just no PRs or an actual error
			if exitErr, ok := err.(*exec.ExitError); ok {
				stderr := string(exitErr.Stderr)
				if strings.Contains(stderr, "no pull requests") {
					continue
				}
				fmt.Printf("\n❌ %s: %s\n", repo.Name, stderr)
			}
			continue
		}
		
		// Parse JSON output
		var prs []PRInfo
		if err := json.Unmarshal(output, &prs); err != nil {
			fmt.Printf("\n❌ %s: failed to parse PR data: %v\n", repo.Name, err)
			continue
		}
		
		if len(prs) == 0 {
			continue
		}
		
		// Display PRs for this repo
		fmt.Printf("\n🗂️  %s (%d PRs)\n", repo.Name, len(prs))
		fmt.Println(strings.Repeat("-", 60))
		
		for _, pr := range prs {
			status := "🟢"
			if pr.State == "CLOSED" {
				status = "🔴"
			} else if pr.State == "MERGED" {
				status = "🟣"
			} else if pr.Draft {
				status = "📝"
			}
			
			authorLogin := "unknown"
			if author, ok := pr.Author.(map[string]interface{}); ok {
				if login, ok := author["login"].(string); ok {
					authorLogin = login
				}
			}
			
			fmt.Printf("%s #%-4d %s\n", status, pr.Number, pr.Title)
			fmt.Printf("       by @%s • %s\n", authorLogin, pr.URL)
		}
		
		totalPRs += len(prs)
	}
	
	fmt.Printf("\n📊 Total: %d pull requests\n", totalPRs)
	return nil
}

// CreatePR creates a new pull request in the specified repository
func (m *Manager) CreatePR(opts PRCreateOptions) error {
	// Find the repository
	var targetRepo *git.Repository
	for _, repo := range m.GitManager.GetRepositories() {
		if repo.Name == opts.Repository || repo.Path == opts.Repository {
			r := repo // Create a copy to get a pointer
			targetRepo = &r
			break
		}
	}
	
	if targetRepo == nil {
		return fmt.Errorf("repository '%s' not found in workspace", opts.Repository)
	}
	
	repoPath := filepath.Join(m.WorkspacePath, targetRepo.Path)
	
	// Check if repository exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("repository '%s' is not cloned", targetRepo.Name)
	}
	
	fmt.Printf("🔧 Creating PR in %s...\n", targetRepo.Name)
	
	// Build gh pr create command
	args := []string{"pr", "create"}
	
	// Use interactive mode if title not provided
	if opts.Title == "" {
		args = append(args, "--fill")
	} else {
		args = append(args, "--title", opts.Title)
		
		if opts.Body != "" {
			args = append(args, "--body", opts.Body)
		} else {
			args = append(args, "--body", "")
		}
	}
	
	if opts.Base != "" {
		args = append(args, "--base", opts.Base)
	}
	
	if opts.Draft {
		args = append(args, "--draft")
	}
	
	for _, reviewer := range opts.Reviewers {
		args = append(args, "--reviewer", reviewer)
	}
	
	for _, assignee := range opts.Assignees {
		args = append(args, "--assignee", assignee)
	}
	
	for _, label := range opts.Labels {
		args = append(args, "--label", label)
	}
	
	cmd := exec.Command("gh", args...)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}
	
	fmt.Println("✅ PR created successfully")
	return nil
}

// ShowPRStatus shows detailed status of pull requests
func (m *Manager) ShowPRStatus(opts PRStatusOptions) error {
	fmt.Println("📊 Pull Request Status")
	fmt.Println(strings.Repeat("=", 60))
	
	// If specific PR is requested
	if opts.Number > 0 && opts.Repository != "" {
		return m.showSinglePRStatus(opts.Repository, opts.Number)
	}
	
	// Show status for all or filtered repos
	for _, repo := range m.GitManager.GetRepositories() {
		if opts.Repository != "" && repo.Name != opts.Repository && repo.Path != opts.Repository {
			continue
		}
		
		repoPath := filepath.Join(m.WorkspacePath, repo.Path)
		
		// Check if repository exists
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			if opts.Repository != "" {
				return fmt.Errorf("repository '%s' is not cloned", repo.Name)
			}
			continue
		}
		
		// Get open PRs with checks
		cmd := exec.Command("gh", "pr", "status")
		cmd.Dir = repoPath
		
		output, err := cmd.Output()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				stderr := string(exitErr.Stderr)
				if !strings.Contains(stderr, "no pull requests") {
					fmt.Printf("\n❌ %s: %s\n", repo.Name, stderr)
				}
			}
			continue
		}
		
		if len(output) > 0 {
			fmt.Printf("\n🗂️  %s\n", repo.Name)
			fmt.Println(strings.Repeat("-", 60))
			fmt.Print(string(output))
		}
	}
	
	return nil
}

// showSinglePRStatus shows status for a specific PR
func (m *Manager) showSinglePRStatus(repoName string, prNumber int) error {
	// Find the repository
	var targetRepo *git.Repository
	for _, repo := range m.GitManager.GetRepositories() {
		if repo.Name == repoName || repo.Path == repoName {
			r := repo
			targetRepo = &r
			break
		}
	}
	
	if targetRepo == nil {
		return fmt.Errorf("repository '%s' not found in workspace", repoName)
	}
	
	repoPath := filepath.Join(m.WorkspacePath, targetRepo.Path)
	
	// Check if repository exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("repository '%s' is not cloned", targetRepo.Name)
	}
	
	// Get PR details
	cmd := exec.Command("gh", "pr", "view", fmt.Sprintf("%d", prNumber))
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to get PR status: %w", err)
	}
	
	// Get checks status
	fmt.Println("\n📋 Checks Status:")
	fmt.Println(strings.Repeat("-", 40))
	
	cmd = exec.Command("gh", "pr", "checks", fmt.Sprintf("%d", prNumber))
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		// Checks might not be available for all PRs
		if !strings.Contains(err.Error(), "no checks") {
			fmt.Println("No checks configured for this PR")
		}
	}
	
	return nil
}

// CheckoutPR checks out a pull request branch locally
func (m *Manager) CheckoutPR(repoName string, prNumber int) error {
	// Find the repository
	var targetRepo *git.Repository
	for _, repo := range m.GitManager.GetRepositories() {
		if repo.Name == repoName || repo.Path == repoName {
			r := repo
			targetRepo = &r
			break
		}
	}
	
	if targetRepo == nil {
		return fmt.Errorf("repository '%s' not found in workspace", repoName)
	}
	
	repoPath := filepath.Join(m.WorkspacePath, targetRepo.Path)
	
	// Check if repository exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("repository '%s' is not cloned", targetRepo.Name)
	}
	
	fmt.Printf("🔄 Checking out PR #%d in %s...\n", prNumber, targetRepo.Name)
	
	// Use gh pr checkout command
	cmd := exec.Command("gh", "pr", "checkout", fmt.Sprintf("%d", prNumber))
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout PR: %w", err)
	}
	
	fmt.Printf("✅ Checked out PR #%d successfully\n", prNumber)
	return nil
}

// MergePR merges a pull request
func (m *Manager) MergePR(opts PRMergeOptions) error {
	// Find the repository
	var targetRepo *git.Repository
	for _, repo := range m.GitManager.GetRepositories() {
		if repo.Name == opts.Repository || repo.Path == opts.Repository {
			r := repo
			targetRepo = &r
			break
		}
	}
	
	if targetRepo == nil {
		return fmt.Errorf("repository '%s' not found in workspace", opts.Repository)
	}
	
	repoPath := filepath.Join(m.WorkspacePath, targetRepo.Path)
	
	// Check if repository exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("repository '%s' is not cloned", targetRepo.Name)
	}
	
	fmt.Printf("🔀 Merging PR #%d in %s...\n", opts.Number, targetRepo.Name)
	
	// Build gh pr merge command
	args := []string{"pr", "merge", fmt.Sprintf("%d", opts.Number)}
	
	// Add merge method if specified
	if opts.Method != "" {
		switch opts.Method {
		case "squash":
			args = append(args, "--squash")
		case "rebase":
			args = append(args, "--rebase")
		case "merge":
			args = append(args, "--merge")
		default:
			return fmt.Errorf("invalid merge method: %s (use merge, squash, or rebase)", opts.Method)
		}
	}
	
	if opts.DeleteRemoteBranch {
		args = append(args, "--delete-branch")
	}
	
	// Auto-confirm the merge
	args = append(args, "--auto")
	
	cmd := exec.Command("gh", args...)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to merge PR: %w", err)
	}
	
	// Delete local branch if requested
	if opts.DeleteLocalBranch {
		fmt.Println("🗑️  Deleting local branch...")
		
		// Get current branch
		cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = repoPath
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
		
		currentBranch := strings.TrimSpace(string(output))
		
		// Switch to main/master first
		cmd = exec.Command("git", "checkout", targetRepo.Branch)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to switch to %s branch: %w", targetRepo.Branch, err)
		}
		
		// Delete the branch
		cmd = exec.Command("git", "branch", "-D", currentBranch)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to delete local branch: %v\n", err)
		} else {
			fmt.Printf("✅ Deleted local branch: %s\n", currentBranch)
		}
	}
	
	fmt.Printf("✅ PR #%d merged successfully\n", opts.Number)
	return nil
}

// BatchCreatePRs creates pull requests in multiple repositories
func (m *Manager) BatchCreatePRs(opts PRBatchCreateOptions) error {
	fmt.Println("🚀 Batch Creating Pull Requests")
	fmt.Println(strings.Repeat("=", 60))
	
	// Determine target repositories
	var targetRepos []git.Repository
	if len(opts.Repos) > 0 {
		// Use specified repositories
		for _, repoName := range opts.Repos {
			for _, repo := range m.GitManager.GetRepositories() {
				if repo.Name == repoName || repo.Path == repoName {
					targetRepos = append(targetRepos, repo)
					break
				}
			}
		}
	} else {
		// Use all repositories
		targetRepos = m.GitManager.GetRepositories()
	}
	
	if len(targetRepos) == 0 {
		return fmt.Errorf("no repositories found")
	}
	
	// Track results
	successCount := 0
	skippedMain := 0
	failedCount := 0
	var results []string
	
	for _, repo := range targetRepos {
		repoPath := filepath.Join(m.WorkspacePath, repo.Path)
		
		// Check if repository exists
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			results = append(results, fmt.Sprintf("⚠️  %s: not cloned", repo.Name))
			failedCount++
			continue
		}
		
		// Get current branch
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = repoPath
		
		output, err := cmd.Output()
		if err != nil {
			results = append(results, fmt.Sprintf("❌ %s: failed to get current branch", repo.Name))
			failedCount++
			continue
		}
		
		currentBranch := strings.TrimSpace(string(output))
		
		// Check if on main/master branch
		isMainBranch := currentBranch == "main" || currentBranch == "master" || currentBranch == repo.Branch
		
		if isMainBranch && !opts.SkipMainCheck {
			results = append(results, fmt.Sprintf("🔵 %s: skipped (on %s branch)", repo.Name, currentBranch))
			skippedMain++
			continue
		}
		
		// Check if there are any changes to commit
		cmd = exec.Command("git", "status", "--porcelain")
		cmd.Dir = repoPath
		output, err = cmd.Output()
		if err == nil && len(output) > 0 {
			results = append(results, fmt.Sprintf("⚠️  %s: has uncommitted changes", repo.Name))
			failedCount++
			continue
		}
		
		// Check if branch has any commits ahead of remote
		cmd = exec.Command("git", "rev-list", "--count", "@{u}..HEAD")
		cmd.Dir = repoPath
		output, err = cmd.Output()
		if err != nil {
			// Branch might not be pushed yet, try to push it
			cmd = exec.Command("git", "push", "-u", "origin", currentBranch)
			cmd.Dir = repoPath
			if err := cmd.Run(); err != nil {
				results = append(results, fmt.Sprintf("❌ %s: failed to push branch", repo.Name))
				failedCount++
				continue
			}
		}
		
		// Build gh pr create command
		args := []string{"pr", "create"}
		
		// Add title and body
		args = append(args, "--title", opts.Title)
		
		if opts.Body != "" {
			args = append(args, "--body", opts.Body)
		} else {
			args = append(args, "--body", fmt.Sprintf("PR created as part of batch operation for: %s", currentBranch))
		}
		
		if opts.Base != "" {
			args = append(args, "--base", opts.Base)
		}
		
		if opts.Draft {
			args = append(args, "--draft")
		}
		
		for _, reviewer := range opts.Reviewers {
			args = append(args, "--reviewer", reviewer)
		}
		
		for _, assignee := range opts.Assignees {
			args = append(args, "--assignee", assignee)
		}
		
		for _, label := range opts.Labels {
			args = append(args, "--label", label)
		}
		
		// Create the PR
		cmd = exec.Command("gh", args...)
		cmd.Dir = repoPath
		
		output, err = cmd.CombinedOutput()
		if err != nil {
			// Check if PR already exists
			if strings.Contains(string(output), "already exists") {
				results = append(results, fmt.Sprintf("⚠️  %s: PR already exists for branch '%s'", repo.Name, currentBranch))
				skippedMain++
			} else {
				results = append(results, fmt.Sprintf("❌ %s: %s", repo.Name, strings.TrimSpace(string(output))))
				failedCount++
			}
			continue
		}
		
		// Extract PR URL from output
		prURL := strings.TrimSpace(string(output))
		results = append(results, fmt.Sprintf("✅ %s: PR created on branch '%s'\n       %s", repo.Name, currentBranch, prURL))
		successCount++
	}
	
	// Display results
	fmt.Println("\n📋 Results:")
	fmt.Println(strings.Repeat("-", 60))
	for _, result := range results {
		fmt.Println(result)
	}
	
	// Summary
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   ✅ Created: %d PRs\n", successCount)
	if skippedMain > 0 {
		fmt.Printf("   🔵 Skipped: %d (on main branch)\n", skippedMain)
	}
	if failedCount > 0 {
		fmt.Printf("   ❌ Failed: %d\n", failedCount)
	}
	
	if skippedMain > 0 && !opts.SkipMainCheck {
		fmt.Println("\n💡 Tip: Create feature branches first with 'rc branch create <name>'")
	}
	
	return nil
}