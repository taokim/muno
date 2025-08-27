package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/taokim/repo-claude/internal/config"
	"github.com/taokim/repo-claude/internal/git"
	"github.com/taokim/repo-claude/internal/tree"
)

// ManagerV3 is the tree-based manager implementation
type ManagerV3 struct {
	ProjectPath  string              // Project root path
	Config       *config.ConfigV3Tree // Tree configuration
	TreeManager  *tree.Manager       // Tree manager
	GitCmd       *git.Git            // Git operations
	CmdExecutor  CommandExecutor
	State        *config.StateV3     // Runtime state
	statePath    string              // Path to state file
}

// NewV3 creates a new tree-based manager
func NewV3(projectPath string) (*ManagerV3, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("resolving project path: %w", err)
	}
	
	treeMgr, err := tree.NewManager(absPath)
	if err != nil {
		return nil, fmt.Errorf("creating tree manager: %w", err)
	}
	
	return &ManagerV3{
		ProjectPath: absPath,
		TreeManager: treeMgr,
		GitCmd:      git.New(),
		CmdExecutor: &RealCommandExecutor{},
		statePath:   filepath.Join(absPath, ".repo-claude-state.json"),
	}, nil
}

// LoadFromCurrentDirV3 loads an existing workspace
func LoadFromCurrentDirV3() (*ManagerV3, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Search upwards for repo-claude.yaml
	searchDir := cwd
	configPath := ""
	for {
		candidate := filepath.Join(searchDir, "repo-claude.yaml")
		if _, err := os.Stat(candidate); err == nil {
			configPath = candidate
			cwd = searchDir // Update cwd to the project root
			break
		}
		
		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			break // Reached root
		}
		searchDir = parent
	}
	
	if configPath == "" {
		return nil, fmt.Errorf("repo-claude.yaml not found in current directory or any parent")
	}
	
	// Load configuration
	cfg, err := config.LoadV3Tree(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading configuration: %w", err)
	}
	
	// Create manager
	mgr, err := NewV3(cwd)
	if err != nil {
		return nil, err
	}
	
	mgr.Config = cfg
	
	// Load tree
	if err := mgr.TreeManager.LoadTree(); err != nil {
		return nil, fmt.Errorf("loading tree: %w", err)
	}
	
	// Load state
	state, err := config.LoadStateV3(mgr.statePath)
	if err != nil {
		return nil, fmt.Errorf("loading state: %w", err)
	}
	mgr.State = state
	
	return mgr, nil
}

// InitWorkspace initializes a new v3 workspace
func (m *ManagerV3) InitWorkspace(projectName string, interactive bool) error {
	// Check if already initialized
	configPath := filepath.Join(m.ProjectPath, "repo-claude.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("workspace already initialized (repo-claude.yaml exists)")
	}
	
	// Create project directory if needed
	if err := os.MkdirAll(m.ProjectPath, 0755); err != nil {
		return fmt.Errorf("creating project directory: %w", err)
	}
	
	// Create configuration
	cfg := config.DefaultConfigV3Tree(projectName)
	
	// Interactive configuration if requested
	if interactive {
		fmt.Println("🚀 Interactive workspace configuration")
		fmt.Print("Enter root repository URL (optional): ")
		var rootRepo string
		fmt.Scanln(&rootRepo)
		cfg.Workspace.RootRepo = rootRepo
	}
	
	// Save configuration
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}
	
	m.Config = cfg
	
	// Initialize tree
	if err := m.TreeManager.Initialize(projectName, cfg.Workspace.RootRepo); err != nil {
		return fmt.Errorf("initializing tree: %w", err)
	}
	
	// Create shared memory file
	sharedMemPath := filepath.Join(m.ProjectPath, "shared-memory.md")
	sharedContent := fmt.Sprintf("# Shared Memory for %s\n\nCreated: %s\n\n## Active Nodes\n\n## Notes\n\n",
		projectName, "2024-01-01") // Using fixed date for consistency
	if err := os.WriteFile(sharedMemPath, []byte(sharedContent), 0644); err != nil {
		return fmt.Errorf("creating shared memory: %w", err)
	}
	
	// Create CLAUDE.md
	claudePath := filepath.Join(m.ProjectPath, "CLAUDE.md")
	claudeContent := fmt.Sprintf(`# CLAUDE.md - %s

This file provides context for Claude Code sessions in this workspace.

## Workspace Structure

This is a v3 tree-based workspace. Navigate using:
- rc use <path> - Navigate to a node
- rc tree - Display workspace structure
- rc add <repo-url> - Add child repository
- rc current - Show current position

## Working Directory

Your current position determines what repositories you're working with.
All git operations are relative to your current node.

## Shared Memory

The shared-memory.md file in the project root is used for coordination
between different Claude sessions.
`, projectName)
	
	if err := os.WriteFile(claudePath, []byte(claudeContent), 0644); err != nil {
		return fmt.Errorf("creating CLAUDE.md: %w", err)
	}
	
	fmt.Println("✅ Workspace initialized successfully!")
	fmt.Printf("📁 Project root: %s\n", m.ProjectPath)
	fmt.Printf("🌳 Tree structure created in: workspaces/\n")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. rc add <repo-url> - Add repositories")
	fmt.Println("  2. rc tree - View structure")
	fmt.Println("  3. rc start - Start Claude session")
	
	return nil
}

// Navigation Commands

// UseNode navigates to a node in the tree
func (m *ManagerV3) UseNode(path string, autoClone bool) error {
	node, err := m.TreeManager.UseNode(path, autoClone)
	if err != nil {
		return err
	}
	
	// Display feedback
	fmt.Printf("📍 Navigated to: %s\n", node.GetPath())
	fmt.Printf("📂 Changed directory to: %s\n", node.FullPath)
	
	if autoClone && node.HasLazyRepos() {
		// Already handled by TreeManager, just show count
		lazyCount := 0
		for _, repo := range node.Repos {
			if repo.Lazy && repo.State == "cloned" {
				lazyCount++
			}
		}
		if lazyCount > 0 {
			fmt.Printf("🔄 Auto-cloned %d lazy repositories\n", lazyCount)
		}
	}
	
	return nil
}

// ShowCurrent displays the current node
func (m *ManagerV3) ShowCurrent() error {
	resolution, err := m.TreeManager.ResolveTarget("")
	if err != nil {
		return err
	}
	
	fmt.Printf("📍 Current: %s (from %s)\n", resolution.Node.GetPath(), resolution.Source)
	
	if resolution.Node.Meta.Description != "" {
		fmt.Printf("   Description: %s\n", resolution.Node.Meta.Description)
	}
	
	fmt.Printf("   Repositories: %d\n", len(resolution.Node.Repos))
	fmt.Printf("   Child nodes: %d\n", len(resolution.Node.Children))
	
	return nil
}

// ClearCurrent clears the stored current node
func (m *ManagerV3) ClearCurrent() error {
	if m.State != nil {
		m.State.SetCurrentNode("")
		if err := m.State.SaveStateV3(m.statePath); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}
	}
	fmt.Println("✅ Cleared current node")
	return nil
}

// Display Commands

// ShowTree displays the tree structure
func (m *ManagerV3) ShowTree(path string, depth int) error {
	resolution, err := m.TreeManager.ResolveTarget(path)
	if err != nil {
		return err
	}
	
	fmt.Printf("🎯 Target: %s (from %s)\n", resolution.Node.GetPath(), resolution.Source)
	fmt.Println()
	
	options := tree.TreeDisplay{
		MaxDepth:     depth,
		ShowRepos:    true,
		ShowLazy:     true,
		ShowModified: true,
	}
	
	output := m.TreeManager.DisplayTree(resolution.Node, options)
	fmt.Print(output)
	
	return nil
}

// ListNodes lists child nodes
func (m *ManagerV3) ListNodes(recursive bool) error {
	resolution, err := m.TreeManager.ResolveTarget("")
	if err != nil {
		return err
	}
	
	fmt.Printf("🎯 Target: %s (from %s)\n", resolution.Node.GetPath(), resolution.Source)
	
	output := m.TreeManager.DisplayList(resolution.Node, recursive)
	fmt.Print(output)
	
	return nil
}

// StatusNode shows node status
func (m *ManagerV3) StatusNode(path string, recursive bool) error {
	resolution, err := m.TreeManager.ResolveTarget(path)
	if err != nil {
		return err
	}
	
	fmt.Printf("🎯 Target: %s (from %s)\n", resolution.Node.GetPath(), resolution.Source)
	fmt.Println()
	
	output := m.TreeManager.DisplayStatus(resolution.Node, recursive)
	fmt.Print(output)
	
	return nil
}

// Repository Management

// AddRepo adds a repository to the current node
func (m *ManagerV3) AddRepo(repoURL, name string, lazy bool) error {
	resolution, err := m.TreeManager.ResolveTarget("")
	if err != nil {
		return err
	}
	
	fmt.Printf("🎯 Target: %s (from %s)\n", resolution.Node.GetPath(), resolution.Source)
	
	options := tree.AddOptions{
		Name: name,
		Lazy: lazy,
	}
	
	repo, err := m.TreeManager.AddRepo(repoURL, options)
	if err != nil {
		return err
	}
	
	if lazy {
		fmt.Printf("📦 Added %s (lazy - will clone on use)\n", repo.Name)
	} else {
		fmt.Printf("✅ Added and cloned %s\n", repo.Name)
	}
	
	return nil
}

// RemoveRepo removes a repository from the current node
func (m *ManagerV3) RemoveRepo(name string) error {
	resolution, err := m.TreeManager.ResolveTarget("")
	if err != nil {
		return err
	}
	
	fmt.Printf("🎯 Target: %s (from %s)\n", resolution.Node.GetPath(), resolution.Source)
	
	if err := m.TreeManager.RemoveRepo(name); err != nil {
		return err
	}
	
	fmt.Printf("✅ Removed %s\n", name)
	return nil
}

// CloneLazy clones lazy repositories
func (m *ManagerV3) CloneLazy(recursive bool) error {
	resolution, err := m.TreeManager.ResolveTarget("")
	if err != nil {
		return err
	}
	
	fmt.Printf("🎯 Target: %s (from %s)\n", resolution.Node.GetPath(), resolution.Source)
	
	if recursive {
		fmt.Printf("🔄 Cloning lazy repositories recursively...\n")
	} else {
		fmt.Printf("🔄 Cloning lazy repositories...\n")
	}
	
	if err := m.TreeManager.CloneLazy(recursive); err != nil {
		return err
	}
	
	fmt.Println("✅ Clone complete")
	return nil
}

// Git Operations

// PullNode pulls repositories at a node
func (m *ManagerV3) PullNode(path string, recursive bool) error {
	resolution, err := m.TreeManager.ResolveTarget(path)
	if err != nil {
		return err
	}
	
	fmt.Printf("🎯 Target: %s (from %s)\n", resolution.Node.GetPath(), resolution.Source)
	
	if recursive {
		count := resolution.Node.CountRepos(true)
		fmt.Printf("🔄 Recursive: will pull %d repositories\n", count)
		fmt.Print("Proceed? [Y/n] ")
		
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) == "n" {
			fmt.Println("❌ Cancelled")
			return nil
		}
	}
	
	// Pull repos at this node
	return m.pullNodeRepos(resolution.Node, recursive)
}

// CommitNode commits changes at a node
func (m *ManagerV3) CommitNode(path string, message string, recursive bool) error {
	resolution, err := m.TreeManager.ResolveTarget(path)
	if err != nil {
		return err
	}
	
	fmt.Printf("🎯 Target: %s (from %s)\n", resolution.Node.GetPath(), resolution.Source)
	
	if recursive {
		count := resolution.Node.CountRepos(true)
		fmt.Printf("🔄 Recursive: will commit in %d repositories\n", count)
	}
	
	return m.commitNodeRepos(resolution.Node, message, recursive)
}

// PushNode pushes changes from a node
func (m *ManagerV3) PushNode(path string, recursive bool) error {
	resolution, err := m.TreeManager.ResolveTarget(path)
	if err != nil {
		return err
	}
	
	fmt.Printf("🎯 Target: %s (from %s)\n", resolution.Node.GetPath(), resolution.Source)
	
	if recursive {
		count := resolution.Node.CountRepos(true)
		fmt.Printf("🔄 Recursive: will push %d repositories\n", count)
	}
	
	return m.pushNodeRepos(resolution.Node, recursive)
}

// Session Management

// StartNode starts a Claude session at a node
func (m *ManagerV3) StartNode(path string, newWindow bool) error {
	resolution, err := m.TreeManager.ResolveTarget(path)
	if err != nil {
		return err
	}
	
	fmt.Printf("🎯 Starting session at: %s (from %s)\n", resolution.Node.GetPath(), resolution.Source)
	
	// Generate CLAUDE.md for the node
	claudePath := filepath.Join(resolution.Node.FullPath, "CLAUDE.md")
	claudeContent := fmt.Sprintf(`# CLAUDE.md - %s

This Claude Code session is working at: %s

## Current Node
- Path: %s
- Repositories: %d
- Children: %d

## Available Repositories
`, resolution.Node.Name, resolution.Node.GetPath(), resolution.Node.GetPath(), 
		len(resolution.Node.Repos), len(resolution.Node.Children))
	
	for _, repo := range resolution.Node.Repos {
		claudeContent += fmt.Sprintf("- %s (%s)\n", repo.Name, repo.State)
	}
	
	claudeContent += "\n## Navigation\nUse 'rc use <path>' to navigate the tree.\n"
	
	if err := os.WriteFile(claudePath, []byte(claudeContent), 0644); err != nil {
		return fmt.Errorf("creating CLAUDE.md: %w", err)
	}
	
	// Start Claude session
	cmd := m.CmdExecutor.Command("claude", "--local-dir", resolution.Node.FullPath)
	cmd.SetDir(resolution.Node.FullPath)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("starting Claude session: %w", err)
	}
	
	// Update state
	if m.State != nil {
		m.State.AddSession(resolution.Node.GetPath(), 0) // PID would be set if we tracked it
		m.State.SaveStateV3(m.statePath)
	}
	
	fmt.Printf("✅ Started Claude session at %s\n", resolution.Node.GetPath())
	return nil
}

// Helper functions

func (m *ManagerV3) pullNodeRepos(node *tree.Node, recursive bool) error {
	// Pull repos at this node
	for _, repo := range node.Repos {
		if repo.State == "cloned" || repo.State == "modified" {
			fmt.Printf("  Pulling %s...\n", repo.Name)
			if err := m.GitCmd.Pull(repo.Path); err != nil {
				fmt.Printf("  ⚠️  Failed to pull %s: %v\n", repo.Name, err)
			} else {
				fmt.Printf("  ✅ %s\n", repo.Name)
			}
		}
	}
	
	// Recursively pull children if requested
	if recursive {
		for _, child := range node.Children {
			fmt.Printf("\n📁 %s:\n", child.Name)
			if err := m.pullNodeRepos(child, true); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (m *ManagerV3) commitNodeRepos(node *tree.Node, message string, recursive bool) error {
	// Commit in repos at this node
	for _, repo := range node.Repos {
		if repo.State == "cloned" || repo.State == "modified" {
			fmt.Printf("  Committing in %s...\n", repo.Name)
			
			// Add all changes
			if err := m.GitCmd.Add(repo.Path, "."); err != nil {
				fmt.Printf("  ⚠️  Failed to add changes in %s: %v\n", repo.Name, err)
				continue
			}
			
			// Commit
			if err := m.GitCmd.Commit(repo.Path, message); err != nil {
				// Check if it's because there's nothing to commit
				if strings.Contains(err.Error(), "nothing to commit") {
					fmt.Printf("  ℹ️  %s: no changes to commit\n", repo.Name)
				} else {
					fmt.Printf("  ⚠️  Failed to commit in %s: %v\n", repo.Name, err)
				}
			} else {
				fmt.Printf("  ✅ %s\n", repo.Name)
			}
		}
	}
	
	// Recursively commit in children if requested
	if recursive {
		for _, child := range node.Children {
			fmt.Printf("\n📁 %s:\n", child.Name)
			if err := m.commitNodeRepos(child, message, true); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (m *ManagerV3) pushNodeRepos(node *tree.Node, recursive bool) error {
	// Push repos at this node
	for _, repo := range node.Repos {
		if repo.State == "cloned" || repo.State == "modified" {
			fmt.Printf("  Pushing %s...\n", repo.Name)
			if err := m.GitCmd.Push(repo.Path); err != nil {
				fmt.Printf("  ⚠️  Failed to push %s: %v\n", repo.Name, err)
			} else {
				fmt.Printf("  ✅ %s\n", repo.Name)
			}
		}
	}
	
	// Recursively push children if requested
	if recursive {
		for _, child := range node.Children {
			fmt.Printf("\n📁 %s:\n", child.Name)
			if err := m.pushNodeRepos(child, true); err != nil {
				return err
			}
		}
	}
	
	return nil
}