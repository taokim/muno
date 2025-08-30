package manager

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
)

// SmartInitWorkspace performs intelligent initialization
// It detects existing git repos and offers to add them
func (m *Manager) SmartInitWorkspace(projectName string, options InitOptions) error {
	// Check if we're in an existing muno workspace
	configPath := filepath.Join(".", "muno.yaml")
	
	// Create base configuration early to get repos directory
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     projectName,
			ReposDir: "repos", // Default value
		},
		Repositories: []config.RepoDefinition{},
	}
	
	// Check if config exists and load it to get repos dir
	hasConfig := false
	if _, err := os.Stat(configPath); err == nil {
		hasConfig = true
		if existingCfg, err := config.LoadTree(configPath); err == nil {
			// Use existing repos dir if specified
			if existingCfg.Workspace.ReposDir != "" {
				cfg.Workspace.ReposDir = existingCfg.Workspace.ReposDir
			}
		}
	}
	
	reposDir := filepath.Join(".", cfg.GetReposDir())
	
	// If config exists, ask user what to do
	if hasConfig {
		fmt.Println("Found existing muno.yaml")
		
		if !options.NonInteractive {
			fmt.Print("Do you want to reinitialize with existing subdirectories? [Y/n]: ")
			
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			
			if response == "n" || response == "no" {
				return fmt.Errorf("initialization cancelled by user")
			}
		} else {
			// In non-interactive mode, proceed with reinitialization
			fmt.Println("Proceeding with reinitialization (non-interactive mode)")
		}
	}
	
	// Get current working directory to properly resolve paths
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}
	
	// Find all git repositories in current directory and subdirectories
	fmt.Printf("\nScanning for git repositories in: %s\n", cwd)
	gitRepos, err := m.findGitRepositoriesWithReposDir(cwd, cfg.GetReposDir())
	if err != nil && !options.Force {
		return fmt.Errorf("scanning for git repositories: %w", err)
	}
	fmt.Printf("Found %d repositories\n", len(gitRepos))
	
	// Create repos directory if it doesn't exist
	if err := os.MkdirAll(reposDir, 0755); err != nil {
		return fmt.Errorf("creating repos directory: %w", err)
	}
	
	// Process found repositories
	movedRepos := []string{}
	existingInRepos := []string{}
	
	// Get absolute path of repos directory for proper comparison
	absReposDir, err := filepath.Abs(reposDir)
	if err != nil {
		absReposDir = reposDir
	}
	fmt.Printf("Repos directory: %s (absolute: %s)\n", reposDir, absReposDir)
	
	for _, repo := range gitRepos {
		repoPath := repo.Path
		repoName := filepath.Base(repoPath)
		
		// Get absolute path of the repository for comparison
		absRepoPath, err := filepath.Abs(repoPath)
		if err != nil {
			absRepoPath = repoPath
		}
		
		// Check if repository is already in the repos directory
		// Compare using absolute paths to avoid relative path issues
		isInReposDir := false
		if strings.HasPrefix(absRepoPath, absReposDir+string(filepath.Separator)) {
			isInReposDir = true
		}
		
		// Additional check using Rel for edge cases
		if !isInReposDir {
			if relPath, err := filepath.Rel(absReposDir, absRepoPath); err == nil {
				isInReposDir = !strings.HasPrefix(relPath, "..") && !filepath.IsAbs(relPath)
			}
		}
		
		fmt.Printf("\nDebug: Found repo at: %s\n", repoPath)
		fmt.Printf("  Absolute path: %s\n", absRepoPath)
		fmt.Printf("  Repos dir: %s\n", absReposDir)
		fmt.Printf("  Is in repos dir: %v\n", isInReposDir)
		
		if isInReposDir {
			// Repository is already in the right place
			existingInRepos = append(existingInRepos, repoName)
			
			// Ask user if they want to include it
			fmt.Printf("Found repository in %s: %s\n", cfg.GetReposDir(), repoPath)
			if repo.RemoteURL != "" {
				fmt.Printf("  Remote URL: %s\n", repo.RemoteURL)
			}
			shouldAdd := true // Default to include
			
			if !options.NonInteractive {
				fmt.Printf("Include in workspace configuration? [Y/n]: ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))
				
				shouldAdd = response != "n" && response != "no"
			}
			
			if shouldAdd {
				// Add to configuration
				repoConfig := config.RepoDefinition{
					Name: repoName,
					Lazy: false, // Already cloned
				}
				
				if repo.RemoteURL != "" {
					repoConfig.URL = repo.RemoteURL
				} else {
					// Use absolute path for file URL
					absPath, _ := filepath.Abs(repoPath)
					repoConfig.URL = "file://" + absPath
				}
				
				cfg.Repositories = append(cfg.Repositories, repoConfig)
			}
			continue
		}
		
		// Skip the .git directory itself
		if repoName == ".git" {
			continue
		}
		
		// Ask user about repositories outside repos dir
		fmt.Printf("Found repository outside %s: %s\n", cfg.GetReposDir(), repoPath)
		if repo.RemoteURL != "" {
			fmt.Printf("  Remote URL: %s\n", repo.RemoteURL)
		}
		
		shouldAdd := false // Default to skip for repos outside repos dir
		
		if !options.NonInteractive {
			fmt.Printf("Add to workspace? [Y/n]: ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			
			shouldAdd = response != "n" && response != "no"
		} else {
			// In non-interactive mode, skip repos outside repos dir
			fmt.Printf("Skipping repository outside %s/ (non-interactive mode)\n", cfg.GetReposDir())
		}
		
		if !shouldAdd {
			continue
		}
		
		// Determine the target path in repos/
		targetPath := filepath.Join(reposDir, repoName)
		
		// Check if target already exists
		if _, err := os.Stat(targetPath); err == nil {
			fmt.Printf("  Warning: %s already exists in %s/, skipping move\n", repoName, cfg.GetReposDir())
			// Still add to config if it has a remote URL
			if repo.RemoteURL != "" {
				cfg.Repositories = append(cfg.Repositories, config.RepoDefinition{
					URL:  repo.RemoteURL,
					Name: repoName,
					Lazy: false, // Already cloned
				})
			}
			continue
		}
		
		// Move the repository to repos/ directory
		fmt.Printf("  Moving %s to %s\n", repoPath, targetPath)
		if err := os.Rename(repoPath, targetPath); err != nil {
			fmt.Printf("  Error moving repository: %v\n", err)
			continue
		}
		
		movedRepos = append(movedRepos, repoName)
		
		// Add to configuration
		repoConfig := config.RepoDefinition{
			Name: repoName,
			Lazy: false, // Already cloned
		}
		
		// Use remote URL if available, otherwise use file:// URL
		if repo.RemoteURL != "" {
			repoConfig.URL = repo.RemoteURL
		} else {
			// Use absolute path for file URL
			absPath, _ := filepath.Abs(targetPath)
			repoConfig.URL = "file://" + absPath
		}
		
		cfg.Repositories = append(cfg.Repositories, repoConfig)
	}
	
	// Save configuration
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}
	
	// Create other necessary files
	sharedMemoryPath := filepath.Join(".", "shared-memory.md")
	if _, err := os.Stat(sharedMemoryPath); os.IsNotExist(err) {
		sharedMemoryContent := fmt.Sprintf("# Shared Memory for %s\n\n", projectName)
		sharedMemoryContent += "This file is used for coordination between different parts of the project.\n\n"
		sharedMemoryContent += "## Repositories\n\n"
		
		for _, repo := range cfg.Repositories {
			sharedMemoryContent += fmt.Sprintf("- %s: %s\n", repo.Name, repo.URL)
		}
		
		if err := os.WriteFile(sharedMemoryPath, []byte(sharedMemoryContent), 0644); err != nil {
			return fmt.Errorf("creating shared memory: %w", err)
		}
	}
	
	// Create CLAUDE.md
	claudePath := filepath.Join(".", "CLAUDE.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		claudeContent := fmt.Sprintf("# %s\n\n", projectName)
		claudeContent += "This is a MUNO workspace with tree-based navigation.\n\n"
		claudeContent += "## Commands\n\n"
		claudeContent += "- `muno tree` - Display repository tree\n"
		claudeContent += "- `muno use <path>` - Navigate to repository\n"
		claudeContent += "- `muno add <url>` - Add new repository\n"
		claudeContent += "- `muno list` - List repositories at current level\n\n"
		
		if err := os.WriteFile(claudePath, []byte(claudeContent), 0644); err != nil {
			return fmt.Errorf("creating CLAUDE.md: %w", err)
		}
	}
	
	// Update the tree manager with the new configuration
	// TODO: The new tree manager doesn't have SetConfig
	// if m.TreeManager != nil {
	// 	m.TreeManager.SetConfig(cfg)
	// }
	
	// Initialize tree state (minimal, just runtime info)
	treeStatePath := filepath.Join(".", ".muno-tree.json")
	if _, err := os.Stat(treeStatePath); os.IsNotExist(err) {
		// Tree manager will handle this
		// For now, we just ensure the config has all repos
		// Also update the manager's config
		m.Config = cfg
	}
	
	// Print summary
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("Workspace '%s' initialized successfully!\n", projectName)
	if len(movedRepos) > 0 {
		fmt.Printf("\nMoved %d repositories to repos/:\n", len(movedRepos))
		for _, name := range movedRepos {
			fmt.Printf("  - %s\n", name)
		}
	}
	if len(existingInRepos) > 0 {
		fmt.Printf("\nFound %d existing repositories in repos/:\n", len(existingInRepos))
		for _, name := range existingInRepos {
			fmt.Printf("  - %s\n", name)
		}
	}
	if len(cfg.Repositories) > 0 {
		fmt.Printf("\nTotal repositories in workspace: %d\n", len(cfg.Repositories))
	}
	fmt.Println("\nNext steps:")
	fmt.Println("  rc tree        # View repository tree")
	fmt.Println("  rc use <repo>  # Navigate to a repository")
	fmt.Println("  rc add <url>   # Add more repositories")
	
	return nil
}

// GitRepoInfo contains information about a discovered git repository
type GitRepoInfo struct {
	Path      string
	RemoteURL string
	Branch    string
}

// findGitRepositoriesWithReposDir searches for git repositories in the given path
func (m *Manager) findGitRepositoriesWithReposDir(rootPath string, reposDir string) ([]GitRepoInfo, error) {
	var repos []GitRepoInfo
	
	// Track visited directories to avoid duplicates
	visited := make(map[string]bool)
	
	fmt.Printf("Walking directory tree from: %s\n", rootPath)
	
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		
		// Skip hidden directories except .git
		// But allow .git itself to be found
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".git" {
			return filepath.SkipDir
		}
		
		// Skip node_modules and other common non-repo directories
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == "vendor" || info.Name() == "target" || info.Name() == "dist" || info.Name() == "build") {
			return filepath.SkipDir
		}
		
		// Check if this is a .git directory
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			
			// Get absolute path for consistent handling
			absRepoPath, err := filepath.Abs(repoPath)
			if err != nil {
				absRepoPath = repoPath
			}
			
			// Skip if we've already processed this repo
			if visited[absRepoPath] {
				return filepath.SkipDir
			}
			visited[absRepoPath] = true
			
			fmt.Printf("  Found .git directory at: %s\n", repoPath)
			
			// Get repository information - use relative path from current dir
			// This ensures consistent path handling
			relPath := repoPath
			if rootPath != "." {
				// If we started from a specific path, make it relative
				if rel, err := filepath.Rel(rootPath, repoPath); err == nil {
					relPath = rel
				}
			}
			
			repoInfo := GitRepoInfo{
				Path: relPath,
			}
			
			// Try to get remote URL
			g := git.New()
			if remotes, err := g.GetRemotes(repoPath); err == nil && len(remotes) > 0 {
				// Use origin if available, otherwise first remote
				for name, url := range remotes {
					if name == "origin" {
						repoInfo.RemoteURL = url
						break
					}
					if repoInfo.RemoteURL == "" {
						repoInfo.RemoteURL = url
					}
				}
			}
			
			// Get current branch
			if branch, err := g.CurrentBranch(repoPath); err == nil {
				repoInfo.Branch = branch
			}
			
			repos = append(repos, repoInfo)
			
			// Don't descend into the .git directory
			return filepath.SkipDir
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return repos, nil
}

// InitOptions contains options for initialization
type InitOptions struct {
	Force         bool // Force initialization even if errors occur
	NonInteractive bool // Skip all prompts and use defaults
}