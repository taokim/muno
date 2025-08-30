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

// ManagerV3 is the refactored tree-based manager implementation with simplified state
type ManagerV3 struct {
	ProjectPath  string              // Project root path
	Config       *config.ConfigV3Tree // Tree configuration
	TreeManager  *tree.Manager     // Refactored tree manager
	GitCmd       *git.Git            // Git operations
	CmdExecutor  CommandExecutor
	State        *config.StateV3     // Runtime state
	statePath    string              // Path to state file
}

// NewV3 creates a new tree-based manager with simplified state
func NewV3(projectPath string) (*ManagerV3, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("resolving project path: %w", err)
	}
	
	// Create git command interface
	gitCmd := git.New()
	
	// Create the refactored tree manager
	treeMgr, err := tree.NewManager(absPath, gitCmd)
	if err != nil {
		return nil, fmt.Errorf("creating tree manager: %w", err)
	}
	
	return &ManagerV3{
		ProjectPath: absPath,
		TreeManager: treeMgr,
		GitCmd:      gitCmd,
		CmdExecutor: &RealCommandExecutor{},
		statePath:   filepath.Join(absPath, ".repo-claude-state.json"),
	}, nil
}

// LoadFromCurrentDirV3 loads an existing workspace with the new manager
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
	
	// The tree manager automatically loads state in NewManagerNew
	// No need to call LoadTree explicitly
	
	return mgr, nil
}

// InitializeV3 initializes a new v3 tree workspace
func (m *ManagerV3) InitializeV3(name string, interactive bool) error {
	// Create workspace directory
	reposDir := filepath.Join(m.ProjectPath, "repos")
	if err := os.MkdirAll(reposDir, 0755); err != nil {
		return fmt.Errorf("creating repos directory: %w", err)
	}
	
	// Create default configuration
	m.Config = config.DefaultConfigV3Tree(name)
	
	// Save configuration
	configPath := filepath.Join(m.ProjectPath, "repo-claude.yaml")
	if err := m.Config.Save(configPath); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}
	
	// Initialize tree state (already done in NewManagerNew)
	// The tree manager creates initial state with root node
	
	// Create shared memory file
	sharedMemPath := filepath.Join(m.ProjectPath, "shared-memory.md")
	if err := createSharedMemory(sharedMemPath); err != nil {
		return fmt.Errorf("creating shared memory: %w", err)
	}
	
	fmt.Printf("✅ Initialized v3 tree workspace: %s\n", name)
	fmt.Printf("📁 Project path: %s\n", m.ProjectPath)
	fmt.Printf("🌳 Tree structure initialized with root node\n")
	
	return nil
}

// UseNode navigates to a node in the tree
func (m *ManagerV3) UseNode(target string) error {
	// If target is empty, stay at current
	if target == "" {
		currentPath := m.TreeManager.GetCurrentPath()
		fmt.Printf("Current node: %s\n", currentPath)
		
		// Display filesystem path
		fsPath := m.TreeManager.ComputeFilesystemPath(currentPath)
		fmt.Printf("Filesystem path: %s\n", fsPath)
		
		return nil
	}
	
	// Use the tree manager's navigation
	if err := m.TreeManager.UseNode(target); err != nil {
		return fmt.Errorf("navigating to %s: %w", target, err)
	}
	
	newPath := m.TreeManager.GetCurrentPath()
	fmt.Printf("✅ Navigated to: %s\n", newPath)
	
	// Show filesystem path for clarity
	fsPath := m.TreeManager.ComputeFilesystemPath(newPath)
	fmt.Printf("📁 Filesystem path: %s\n", fsPath)
	
	// Show children if any
	children, _ := m.TreeManager.ListChildren("")
	if len(children) > 0 {
		fmt.Printf("\nChildren:\n")
		for _, child := range children {
			icon := "📦"
			if child.Type == tree.NodeTypeRepo && child.State == tree.RepoStateMissing {
				icon = "💤"
			}
			fmt.Printf("  %s %s\n", icon, child.Name)
		}
	}
	
	return nil
}

// AddRepo adds a repository to the current or specified node
func (m *ManagerV3) AddRepo(parentPath, url string, options tree.AddOptions) error {
	// Extract repo name from URL if not provided
	name := options.Name
	if name == "" {
		parts := strings.Split(url, "/")
		name = strings.TrimSuffix(parts[len(parts)-1], ".git")
	}
	
	// Add the repository
	if err := m.TreeManager.AddRepo(parentPath, name, url, options.Lazy); err != nil {
		return fmt.Errorf("adding repository: %w", err)
	}
	
	fmt.Printf("✅ Added repository: %s\n", name)
	if options.Lazy {
		fmt.Printf("💤 Repository marked as lazy (will clone on first use)\n")
	} else {
		fmt.Printf("📦 Repository cloned successfully\n")
	}
	
	return nil
}

// RemoveNode removes a node from the tree
func (m *ManagerV3) RemoveNode(target string) error {
	if target == "" {
		return fmt.Errorf("target path required")
	}
	
	// Confirm before removing
	fmt.Printf("⚠️  This will remove the node and all its children\n")
	fmt.Printf("Are you sure you want to remove %s? (y/N): ", target)
	
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Cancelled")
		return nil
	}
	
	if err := m.TreeManager.RemoveNode(target); err != nil {
		return fmt.Errorf("removing node: %w", err)
	}
	
	fmt.Printf("✅ Removed node: %s\n", target)
	return nil
}

// ShowTree displays the tree structure
func (m *ManagerV3) ShowTree(maxDepth int) error {
	var output string
	if maxDepth > 0 {
		output = m.TreeManager.DisplayTreeWithDepth(maxDepth)
	} else {
		output = m.TreeManager.DisplayTree()
	}
	
	fmt.Print(output)
	return nil
}

// ShowStatus displays the current status
func (m *ManagerV3) ShowStatus() error {
	fmt.Print(m.TreeManager.DisplayStatus())
	return nil
}

// ShowCurrent displays the current node information
func (m *ManagerV3) ShowCurrent() error {
	currentPath := m.TreeManager.GetCurrentPath()
	fmt.Printf("Current node: %s\n", currentPath)
	
	// Show filesystem path
	fsPath := m.TreeManager.ComputeFilesystemPath(currentPath)
	fmt.Printf("Filesystem path: %s\n", fsPath)
	
	// Show path from root
	fmt.Printf("Path: %s\n", m.TreeManager.DisplayPath())
	
	// Show children
	fmt.Printf("\n%s\n", m.TreeManager.DisplayChildren())
	
	return nil
}

// ListNodes lists children of current or specified node
func (m *ManagerV3) ListNodes(target string) error {
	children, err := m.TreeManager.ListChildren(target)
	if err != nil {
		return fmt.Errorf("listing children: %w", err)
	}
	
	if len(children) == 0 {
		fmt.Println("No children")
		return nil
	}
	
	targetPath := target
	if targetPath == "" {
		targetPath = m.TreeManager.GetCurrentPath()
	}
	
	fmt.Printf("Children of %s:\n", targetPath)
	for _, child := range children {
		status := ""
		if child.Type == tree.NodeTypeRepo {
			switch child.State {
			case tree.RepoStateMissing:
				status = " (lazy)"
			case tree.RepoStateModified:
				status = " (modified)"
			}
		}
		fmt.Printf("  - %s%s\n", child.Name, status)
	}
	
	return nil
}

// CloneRepos clones lazy repositories
func (m *ManagerV3) CloneRepos(target string, recursive bool) error {
	if target == "" {
		target = m.TreeManager.GetCurrentPath()
	}
	
	fmt.Printf("Cloning lazy repositories in %s", target)
	if recursive {
		fmt.Printf(" (recursive)")
	}
	fmt.Println()
	
	if err := m.TreeManager.CloneLazyRepos(target, recursive); err != nil {
		return fmt.Errorf("cloning repositories: %w", err)
	}
	
	fmt.Println("✅ All lazy repositories cloned")
	return nil
}

// StartClaude starts Claude Code at the current or specified node
func (m *ManagerV3) StartClaude(target string) error {
	targetPath := target
	if targetPath == "" {
		targetPath = m.TreeManager.GetCurrentPath()
	}
	
	// Navigate to the target node first
	if targetPath != m.TreeManager.GetCurrentPath() {
		if err := m.TreeManager.UseNode(targetPath); err != nil {
			return fmt.Errorf("navigating to %s: %w", targetPath, err)
		}
	}
	
	// Get filesystem path
	fsPath := m.TreeManager.ComputeFilesystemPath(targetPath)
	
	fmt.Printf("Starting Claude Code at: %s\n", targetPath)
	fmt.Printf("Filesystem path: %s\n", fsPath)
	
	// Change to the directory
	if err := os.Chdir(fsPath); err != nil {
		return fmt.Errorf("changing directory: %w", err)
	}
	
	// Execute claude command
	cmd := m.CmdExecutor.Command("claude")
	cmd.SetDir(fsPath)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("starting Claude: %w", err)
	}
	
	return nil
}

// Helper function to create shared memory file
func createSharedMemory(path string) error {
	content := `# Shared Memory

This file serves as a shared memory space for coordination across the repository tree.

## Current Focus
- Tree-based navigation system
- Simplified state management
- CWD-first resolution

## Notes
- Each node in the tree can contain repositories
- Navigation changes the current working directory
- Lazy repositories are cloned on first access
`
	return os.WriteFile(path, []byte(content), 0644)
}

// InitWorkspace initializes a new workspace (compatibility method)
func (m *ManagerV3) InitWorkspace(projectName string, interactive bool) error {
	return m.InitializeV3(projectName, interactive)
}

// UseNodeWithClone navigates with optional auto-clone (compatibility method)
func (m *ManagerV3) UseNodeWithClone(path string, autoClone bool) error {
	// For now, always auto-clone if needed
	return m.UseNode(path)
}

// ClearCurrent clears the current node position
func (m *ManagerV3) ClearCurrent() error {
	// Reset to root
	if err := m.TreeManager.UseNode("/"); err != nil {
		return fmt.Errorf("resetting to root: %w", err)
	}
	fmt.Println("Current position cleared (reset to root)")
	return nil
}

// ShowTreeAtPath shows tree at a specific path (compatibility method)
func (m *ManagerV3) ShowTreeAtPath(path string, depth int) error {
	// Navigate to path first if specified
	if path != "" {
		if err := m.TreeManager.UseNode(path); err != nil {
			return fmt.Errorf("navigating to %s: %w", path, err)
		}
	}
	return m.ShowTree(depth)
}

// ListNodesRecursive lists nodes with recursive option (compatibility method)
func (m *ManagerV3) ListNodesRecursive(recursive bool) error {
	// List from current position
	return m.ListNodes("")
}

// StatusNode shows status of a node
func (m *ManagerV3) StatusNode(path string, recursive bool) error {
	if path != "" {
		if err := m.TreeManager.UseNode(path); err != nil {
			return fmt.Errorf("navigating to %s: %w", path, err)
		}
	}
	return m.ShowStatus()
}

// AddRepoSimple adds repo with individual parameters (compatibility method)
func (m *ManagerV3) AddRepoSimple(repoURL, name string, lazy bool) error {
	options := tree.AddOptions{
		Name: name,
		Lazy: lazy,
	}
	return m.AddRepo("", repoURL, options)
}

// RemoveRepo removes a repository (compatibility method)
func (m *ManagerV3) RemoveRepo(name string) error {
	return m.RemoveNode(name)
}

// CloneLazy clones lazy repositories
func (m *ManagerV3) CloneLazy(recursive bool) error {
	return m.CloneRepos("", recursive)
}

// PullNode pulls repositories at a node
func (m *ManagerV3) PullNode(path string, recursive bool) error {
	if path != "" {
		if err := m.TreeManager.UseNode(path); err != nil {
			return fmt.Errorf("navigating to %s: %w", path, err)
		}
	}
	
	currentPath := m.TreeManager.GetCurrentPath()
	fsPath := m.TreeManager.ComputeFilesystemPath(currentPath)
	
	fmt.Printf("Pulling repositories at %s\n", currentPath)
	
	// Pull in the current directory
	if err := m.GitCmd.Pull(fsPath); err != nil {
		return fmt.Errorf("pulling at %s: %w", currentPath, err)
	}
	
	if recursive {
		// TODO: Implement recursive pull
		fmt.Println("Recursive pull not yet implemented")
	}
	
	fmt.Printf("✅ Pulled repositories at %s\n", currentPath)
	return nil
}

// CommitNode commits changes at a node
func (m *ManagerV3) CommitNode(path string, message string, recursive bool) error {
	if path != "" {
		if err := m.TreeManager.UseNode(path); err != nil {
			return fmt.Errorf("navigating to %s: %w", path, err)
		}
	}
	
	currentPath := m.TreeManager.GetCurrentPath()
	fsPath := m.TreeManager.ComputeFilesystemPath(currentPath)
	
	fmt.Printf("Committing changes at %s\n", currentPath)
	
	// Add all changes
	if err := m.GitCmd.Add(fsPath, "."); err != nil {
		return fmt.Errorf("staging changes at %s: %w", currentPath, err)
	}
	
	// Commit
	if err := m.GitCmd.Commit(fsPath, message); err != nil {
		return fmt.Errorf("committing at %s: %w", currentPath, err)
	}
	
	if recursive {
		// TODO: Implement recursive commit
		fmt.Println("Recursive commit not yet implemented")
	}
	
	fmt.Printf("✅ Committed changes at %s\n", currentPath)
	return nil
}

// PushNode pushes changes from a node
func (m *ManagerV3) PushNode(path string, recursive bool) error {
	if path != "" {
		if err := m.TreeManager.UseNode(path); err != nil {
			return fmt.Errorf("navigating to %s: %w", path, err)
		}
	}
	
	currentPath := m.TreeManager.GetCurrentPath()
	fsPath := m.TreeManager.ComputeFilesystemPath(currentPath)
	
	fmt.Printf("Pushing changes from %s\n", currentPath)
	
	// Push
	if err := m.GitCmd.Push(fsPath); err != nil {
		return fmt.Errorf("pushing from %s: %w", currentPath, err)
		}
	
	if recursive {
		// TODO: Implement recursive push
		fmt.Println("Recursive push not yet implemented")
	}
	
	fmt.Printf("✅ Pushed changes from %s\n", currentPath)
	return nil
}

// StartNode starts Claude at a node
func (m *ManagerV3) StartNode(path string, newWindow bool) error {
	// Use StartClaude internally
	return m.StartClaude(path)
}