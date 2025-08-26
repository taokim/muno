package scope

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Scope represents an isolated workspace with its own repository checkouts
type Scope struct {
	meta    *Meta
	path    string
	manager *Manager
}

// Clone clones missing repositories into the scope
func (s *Scope) Clone(repoNames []string) error {
	for _, repoName := range repoNames {
		if err := s.cloneRepo(repoName, ""); err != nil {
			return fmt.Errorf("failed to clone %s: %w", repoName, err)
		}
	}
	return s.manager.saveMeta(s.path, s.meta)
}

// cloneRepo clones a single repository
func (s *Scope) cloneRepo(repoName, branch string) error {
	// Find repository configuration
	repo, exists := s.manager.config.Repositories[repoName]
	if !exists {
		return fmt.Errorf("repository %s not found in configuration", repoName)
	}

	repoURL := repo.URL
	defaultBranch := repo.DefaultBranch
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	// Use specified branch or default
	if branch == "" {
		branch = defaultBranch
	}

	// Clone the repository
	repoPath := filepath.Join(s.path, repoName)
	if _, err := os.Stat(repoPath); err == nil {
		return fmt.Errorf("repository %s already exists in scope", repoName)
	}

	fmt.Printf("Cloning %s (branch: %s)...\n", repoName, branch)
	cmd := exec.Command("git", "clone", "-b", branch, repoURL, repoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Get current commit
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get commit hash: %w", err)
	}
	commit := strings.TrimSpace(string(output))

	// Add to metadata
	s.meta.Repos = append(s.meta.Repos, RepoState{
		Name:       repoName,
		URL:        repoURL,
		Branch:     branch,
		Commit:     commit,
		ClonedAt:   time.Now(),
		LastPulled: time.Now(),
	})

	// Create CLAUDE.md in the repository with links to scope and root
	claudePath := filepath.Join(repoPath, "CLAUDE.md")
	claudeContent := fmt.Sprintf(`# CLAUDE.md - Repository Level

This repository (%s) is part of scope: %s

## Three-Level Structure
- **Root Level**: %s (Project root with repo-claude.yaml)
- **Scope Level**: %s (Isolated workspace, not under git)
- **Repo Level**: %s (This repository)

## Scope Information
- Name: %s
- Type: %s
- Description: %s
- Created: %s

## Links
- [Root CLAUDE.md](../../CLAUDE.md) - Project-wide instructions
- [Scope CLAUDE.md](../CLAUDE.md) - Scope-specific context (not in git)
- [Shared Memory](../shared-memory.md) - Inter-scope coordination

## Documentation Guidelines

### IMPORTANT: Documentation Location Rules
1. **Cross-repository documentation**: MUST be stored in %s/docs/scopes/%s/
2. **Repository-specific docs**: Store in this repository's docs/ folder
3. **Global project docs**: Store in %s/docs/global/

### Documentation Commands
- Create scope doc: rc docs create %s <filename>
- Create global doc: rc docs create global <filename>
- List docs: rc docs list [%s]
- Sync docs to git: rc docs sync

## Other Repositories in This Scope
%s

## Working Directory Note
Your current working directory is the scope directory, not this repository.
To work with files in this repo, use relative paths like: %s/<file>
`, repoName, s.meta.Name,
		s.manager.projectPath, s.path, repoPath,
		s.meta.Name, s.meta.Type, s.meta.Description,
		s.meta.CreatedAt.Format(time.RFC3339),
		s.manager.projectPath, s.meta.Name,
		s.manager.projectPath,
		s.meta.Name, s.meta.Name,
		s.getRepoList(),
		repoName)

	if err := os.WriteFile(claudePath, []byte(claudeContent), 0644); err != nil {
		fmt.Printf("Warning: failed to create CLAUDE.md in repo: %v\n", err)
	}

	fmt.Printf("✅ Cloned %s to %s\n", repoName, repoPath)
	return nil
}

// Pull pulls latest changes for repositories
func (s *Scope) Pull(opts PullOptions) error {
	// Update metadata state
	s.meta.LastAccessed = time.Now()
	defer s.manager.saveMeta(s.path, s.meta)

	repos := opts.Repos
	if len(repos) == 0 {
		// Pull all repos in scope
		for _, repo := range s.meta.Repos {
			repos = append(repos, repo.Name)
		}
	}

	var errors []string
	for _, repoName := range repos {
		repoPath := filepath.Join(s.path, repoName)
		
		// Check if repo exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			if opts.CloneMissing {
				if err := s.cloneRepo(repoName, ""); err != nil {
					errors = append(errors, fmt.Sprintf("%s: %v", repoName, err))
				}
			} else {
				errors = append(errors, fmt.Sprintf("%s: not cloned", repoName))
			}
			continue
		}

		// Pull latest changes
		fmt.Printf("Pulling %s...\n", repoName)
		cmd := exec.Command("git", "pull")
		if opts.Force {
			cmd = exec.Command("git", "pull", "--force")
		}
		cmd.Dir = repoPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", repoName, err))
			continue
		}

		// Update metadata
		for i, repo := range s.meta.Repos {
			if repo.Name == repoName {
				s.meta.Repos[i].LastPulled = time.Now()
				
				// Get current commit
				cmd = exec.Command("git", "rev-parse", "HEAD")
				cmd.Dir = repoPath
				if output, err := cmd.Output(); err == nil {
					s.meta.Repos[i].Commit = strings.TrimSpace(string(output))
				}
				
				// Get current branch
				cmd = exec.Command("git", "branch", "--show-current")
				cmd.Dir = repoPath
				if output, err := cmd.Output(); err == nil {
					s.meta.Repos[i].Branch = strings.TrimSpace(string(output))
				}
				break
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("pull failed for some repositories:\n%s", strings.Join(errors, "\n"))
	}

	fmt.Printf("✅ Pulled latest changes for %d repositories\n", len(repos))
	return nil
}

// Status returns the status of the scope
func (s *Scope) Status() (*StatusReport, error) {
	report := &StatusReport{
		Name:         s.meta.Name,
		Type:         s.meta.Type,
		State:        s.meta.State,
		Path:         s.path,
		CreatedAt:    s.meta.CreatedAt,
		LastAccessed: s.meta.LastAccessed,
		Repos:        []RepoStatus{},
		IsRunning:    s.meta.State == StateActive,
		SessionPID:   s.meta.SessionPID,
	}

	// Calculate disk usage
	var totalSize int64
	filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	report.DiskUsage = totalSize

	// Get status for each repository
	for _, repo := range s.meta.Repos {
		repoPath := filepath.Join(s.path, repo.Name)
		repoStatus := RepoStatus{
			Name:       repo.Name,
			Branch:     repo.Branch,
			Commit:     repo.Commit,
			LastPulled: repo.LastPulled,
		}

		// Check if repo exists
		if _, err := os.Stat(repoPath); err == nil {
			// Get git status
			cmd := exec.Command("git", "status", "--porcelain")
			cmd.Dir = repoPath
			if output, err := cmd.Output(); err == nil {
				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					if strings.TrimSpace(line) == "" {
						continue
					}
					if strings.HasPrefix(line, "??") {
						repoStatus.UntrackedFiles++
					} else if strings.HasPrefix(line, " M") || strings.HasPrefix(line, "M ") {
						repoStatus.ModifiedFiles++
					}
				}
				repoStatus.IsDirty = len(lines) > 1 // More than just empty line
			}

			// Get ahead/behind status
			cmd = exec.Command("git", "rev-list", "--count", "--left-right", "@{upstream}...HEAD")
			cmd.Dir = repoPath
			if output, err := cmd.Output(); err == nil {
				var behind, ahead int
				fmt.Sscanf(string(output), "%d\t%d", &behind, &ahead)
				repoStatus.AheadBy = ahead
				repoStatus.BehindBy = behind
			}
		}

		report.Repos = append(report.Repos, repoStatus)
	}

	return report, nil
}

// Commit commits changes across repositories
func (s *Scope) Commit(message string) error {
	if message == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	var committed []string
	var errors []string

	for _, repo := range s.meta.Repos {
		repoPath := filepath.Join(s.path, repo.Name)
		
		// Check if repo exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			continue
		}

		// Check if there are changes
		cmd := exec.Command("git", "status", "--porcelain")
		cmd.Dir = repoPath
		output, err := cmd.Output()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to check status", repo.Name))
			continue
		}

		if strings.TrimSpace(string(output)) == "" {
			// No changes to commit
			continue
		}

		// Add all changes
		cmd = exec.Command("git", "add", "-A")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to add changes", repo.Name))
			continue
		}

		// Commit changes
		cmd = exec.Command("git", "commit", "-m", message)
		cmd.Dir = repoPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to commit", repo.Name))
			continue
		}

		committed = append(committed, repo.Name)
	}

	if len(errors) > 0 {
		return fmt.Errorf("commit failed in some repositories:\n%s", strings.Join(errors, "\n"))
	}

	if len(committed) > 0 {
		fmt.Printf("✅ Committed changes in %d repositories: %s\n", 
			len(committed), strings.Join(committed, ", "))
	} else {
		fmt.Println("No changes to commit")
	}

	return nil
}

// Push pushes changes to remote repositories
func (s *Scope) Push() error {
	var pushed []string
	var errors []string

	for _, repo := range s.meta.Repos {
		repoPath := filepath.Join(s.path, repo.Name)
		
		// Check if repo exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			continue
		}

		// Push changes
		cmd := exec.Command("git", "push")
		cmd.Dir = repoPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", repo.Name, err))
			continue
		}

		pushed = append(pushed, repo.Name)
	}

	if len(errors) > 0 {
		return fmt.Errorf("push failed in some repositories:\n%s", strings.Join(errors, "\n"))
	}

	if len(pushed) > 0 {
		fmt.Printf("✅ Pushed changes in %d repositories: %s\n", 
			len(pushed), strings.Join(pushed, ", "))
	} else {
		fmt.Println("Nothing to push")
	}

	return nil
}

// SwitchBranch switches all repositories to a specific branch
func (s *Scope) SwitchBranch(branch string) error {
	var switched []string
	var errors []string

	for _, repo := range s.meta.Repos {
		repoPath := filepath.Join(s.path, repo.Name)
		
		// Check if repo exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			continue
		}

		// Check if branch exists locally
		cmd := exec.Command("git", "show-ref", "--verify", fmt.Sprintf("refs/heads/%s", branch))
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			// Try to create branch from remote
			cmd = exec.Command("git", "checkout", "-b", branch, fmt.Sprintf("origin/%s", branch))
		} else {
			// Switch to existing branch
			cmd = exec.Command("git", "checkout", branch)
		}
		
		cmd.Dir = repoPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", repo.Name, err))
			continue
		}

		// Update metadata
		for i, r := range s.meta.Repos {
			if r.Name == repo.Name {
				s.meta.Repos[i].Branch = branch
				break
			}
		}

		switched = append(switched, repo.Name)
	}

	// Save updated metadata
	if err := s.manager.saveMeta(s.path, s.meta); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("branch switch failed in some repositories:\n%s", strings.Join(errors, "\n"))
	}

	if len(switched) > 0 {
		fmt.Printf("✅ Switched to branch '%s' in %d repositories: %s\n", 
			branch, len(switched), strings.Join(switched, ", "))
	}

	return nil
}

// getRepoList returns a formatted list of repositories in the scope
func (s *Scope) getRepoList() string {
	var repos []string
	for _, repo := range s.meta.Repos {
		repos = append(repos, fmt.Sprintf("- %s (branch: %s)", repo.Name, repo.Branch))
	}
	if len(repos) == 0 {
		return "- No repositories cloned yet"
	}
	return strings.Join(repos, "\n")
}

// GetMeta returns the scope metadata
func (s *Scope) GetMeta() *Meta {
	return s.meta
}

// GetPath returns the scope path
func (s *Scope) GetPath() string {
	return s.path
}

// Start starts a Claude session for this scope
func (s *Scope) Start(opts StartOptions) error {
	// Check if repositories need to be cloned
	if !s.hasClonedRepos() {
		fmt.Printf("First time starting scope '%s'. Cloning repositories...\n", s.meta.Name)
		if err := s.cloneAllRepos(); err != nil {
			return fmt.Errorf("failed to clone repositories: %w", err)
		}
	}

	// Pull latest changes if requested
	if opts.Pull {
		fmt.Println("Pulling latest changes...")
		if err := s.Pull(PullOptions{CloneMissing: true, Parallel: true}); err != nil {
			fmt.Printf("Warning: some pulls failed: %v\n", err)
		}
	}

	// Create scope-level CLAUDE.md
	if err := s.createScopeCLAUDE(); err != nil {
		fmt.Printf("Warning: failed to create scope CLAUDE.md: %v\n", err)
	}

	// Update state to active
	s.meta.State = StateActive
	s.meta.LastAccessed = time.Now()
	if err := s.manager.saveMeta(s.path, s.meta); err != nil {
		fmt.Printf("Warning: failed to update scope state: %v\n", err)
	}

	fmt.Printf("✅ Scope '%s' is ready at %s\n", s.meta.Name, s.path)
	fmt.Printf("Working directory will be: %s\n", s.path)
	return nil
}

// hasClonedRepos checks if any repositories have been cloned
func (s *Scope) hasClonedRepos() bool {
	return len(s.meta.Repos) > 0
}

// cloneAllRepos clones all repositories defined for this scope
func (s *Scope) cloneAllRepos() error {
	// Get repos from config
	scopeConfig, exists := s.manager.config.Scopes[s.meta.Name]
	if !exists {
		return fmt.Errorf("scope %s not found in configuration", s.meta.Name)
	}

	for _, repoName := range scopeConfig.Repos {
		if err := s.cloneRepo(repoName, ""); err != nil {
			fmt.Printf("Warning: failed to clone %s: %v\n", repoName, err)
		}
	}

	// Save updated metadata
	return s.manager.saveMeta(s.path, s.meta)
}

// createScopeCLAUDE creates the scope-level CLAUDE.md file
func (s *Scope) createScopeCLAUDE() error {
	claudePath := filepath.Join(s.path, "CLAUDE.md")
	gitignorePath := filepath.Join(s.path, ".gitignore")

	// Create .gitignore to exclude scope-level files
	gitignoreContent := `# Scope-level files (not in git)
CLAUDE.md
shared-memory.md
.scope-meta.json
.gitignore
`
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	claudeContent := fmt.Sprintf(`# CLAUDE.md - Scope Level

## IMPORTANT: This file is at the SCOPE level, not under git control

You are working in scope: %s

## Three-Level Structure
1. **Root Level**: %s (Project root with repo-claude.yaml)
2. **Scope Level**: %s (THIS DIRECTORY - isolated workspace)
3. **Repo Level**: Individual repositories cloned below

## Current Working Directory
Your CWD is: %s

## Scope Details
- Name: %s
- Type: %s
- Description: %s
- Created: %s
- Last Accessed: %s

## Repositories in This Scope
%s

## Important Links
- [Root CLAUDE.md](%s/CLAUDE.md) - Main project instructions
- [Shared Memory](./shared-memory.md) - Coordination with other scopes

## Documentation Structure

### CRITICAL: All cross-repository documentation MUST be stored in:
%s/docs/scopes/%s/

### Documentation Guidelines:
1. **Cross-repo docs**: Store in %s/docs/scopes/%s/
2. **Repo-specific docs**: Store in each repository's docs/ folder
3. **Global docs**: Store in %s/docs/global/

### Why This Structure?
- Scope directory is NOT under git (temporary workspace)
- Documentation needs to persist in git
- Clear separation between temporary work and permanent docs

## Available Commands (run from this directory)

### Scope Management
- rc pull %s --clone-missing    # Pull/clone repositories
- rc status %s                  # Check scope status
- rc commit %s -m "message"     # Commit changes
- rc push %s                    # Push changes
- rc branch %s <branch-name>    # Switch branches

### Documentation
- rc docs create %s <file>      # Create scope documentation
- rc docs list %s               # List scope docs
- rc docs sync                  # Commit docs to git

## Notes
- All changes in THIS directory (%s) are ignored by git
- Only changes within repository subdirectories are tracked
- Use shared-memory.md for inter-scope communication
`, s.meta.Name,
		s.manager.projectPath, s.path, s.path,
		s.meta.Name, s.meta.Type, s.meta.Description,
		s.meta.CreatedAt.Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
		s.getRepoList(),
		s.manager.projectPath,
		s.manager.projectPath, s.meta.Name,
		s.manager.projectPath, s.meta.Name,
		s.manager.projectPath,
		s.meta.Name, s.meta.Name, s.meta.Name, s.meta.Name, s.meta.Name,
		s.meta.Name, s.meta.Name,
		s.path)

	return os.WriteFile(claudePath, []byte(claudeContent), 0644)
}