package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/taokim/repo-claude/internal/manager"
)

// App is the tree-based application
type App struct {
	rootCmd *cobra.Command
	stdout  io.Writer
	stderr  io.Writer
}

// NewApp creates a new tree-based application
func NewApp() *App {
	app := &App{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	app.setupCommands()
	return app
}

// SetOutput sets the output writers
func (a *App) SetOutput(stdout, stderr io.Writer) {
	a.stdout = stdout
	a.stderr = stderr
	a.rootCmd.SetOut(stdout)
	a.rootCmd.SetErr(stderr)
}

// Execute runs the application
func (a *App) Execute() error {
	return a.rootCmd.Execute()
}

// ExecuteWithArgs runs with specific arguments
func (a *App) ExecuteWithArgs(args []string) error {
	a.rootCmd.SetArgs(args)
	return a.Execute()
}

// setupCommands initializes all commands
func (a *App) setupCommands() {
	a.rootCmd = &cobra.Command{
		Use:   "rc",
		Short: "Multi-repository orchestration for Claude Code with tree-based workspaces",
		Long: `Repo-Claude v3 orchestrates Claude Code sessions across multiple 
repositories with tree-based navigation and lazy loading.

Features:
- Tree-based navigation: Navigate workspace like a filesystem
- Lazy loading: Repos clone on-demand when parent is used
- CWD-first resolution: Commands operate based on current directory
- Simple configuration: Everything is just a repository`,
		Version: formatVersion(),
	}
	
	// Core commands
	a.rootCmd.AddCommand(a.newInitCmd())
	a.rootCmd.AddCommand(a.newListCmd())
	a.rootCmd.AddCommand(a.newStartCmd())
	a.rootCmd.AddCommand(a.newStatusCmd())
	
	// Navigation commands
	a.rootCmd.AddCommand(a.newUseCmd())
	a.rootCmd.AddCommand(a.newCurrentCmd())
	a.rootCmd.AddCommand(a.newTreeCmd())
	
	// Repository management
	a.rootCmd.AddCommand(a.newAddCmd())
	a.rootCmd.AddCommand(a.newRemoveCmd())
	a.rootCmd.AddCommand(a.newCloneCmd())
	
	// Git operations
	a.rootCmd.AddCommand(a.newPullCmd())
	a.rootCmd.AddCommand(a.newCommitCmd())
	a.rootCmd.AddCommand(a.newPushCmd())
	
	// Version
	a.rootCmd.AddCommand(a.newVersionCmd())
}

// newInitCmd creates the init command
func (a *App) newInitCmd() *cobra.Command {
	var force bool
	var smart bool
	var nonInteractive bool
	
	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new repo-claude v3 project",
		Long: `Initialize a new repo-claude v3 project with tree-based workspace.
		
Smart mode (default):
- Detects existing git repositories
- Offers to add them to workspace
- Moves repositories to repos/ directory
- Creates repo-claude.yaml with all repository definitions
		
Creates:
- repo-claude.yaml (v3 configuration with repo list)
- repos/ directory for tree structure
- Root CLAUDE.md with project instructions`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := ""
			projectPath := "."
			
			if len(args) > 0 {
				projectName = args[0]
				projectPath = projectName
			} else {
				// Use current directory name
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectName = filepath.Base(cwd)
			}
			
			mgr, err := manager.NewManager(projectPath)
			if err != nil {
				return fmt.Errorf("creating manager: %w", err)
			}
			
			// Use smart init by default
			if smart || !cmd.Flags().Changed("no-smart") {
				options := manager.InitOptions{
					Force:          force,
					NonInteractive: nonInteractive,
				}
				if err := mgr.SmartInitWorkspace(projectName, options); err != nil {
					return fmt.Errorf("smart init workspace: %w", err)
				}
			} else {
				// Use old init method (always interactive)
				if err := mgr.Initialize(projectName, true); err != nil {
					return fmt.Errorf("initializing workspace: %w", err)
				}
			}
			
			return nil
		},
	}
	
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force initialization even if errors occur")
	cmd.Flags().BoolVar(&smart, "smart", true, "Smart detection of existing git repos")
	cmd.Flags().Bool("no-smart", false, "Disable smart detection")
	cmd.Flags().BoolVarP(&nonInteractive, "non-interactive", "n", false, "Skip all prompts and use defaults")
	
	return cmd
}

// newListCmd creates the list command
func (a *App) newListCmd() *cobra.Command {
	var recursive bool
	
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List child nodes and repositories",
		Long: `List child nodes and their repositories from the current position.
		
Shows:
- Child node names
- Repository count and status
- Lazy/cloned state`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.ListNodesRecursive(recursive)
		},
	}
	
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "List recursively")
	
	return cmd
}

// newStartCmd creates the start command
func (a *App) newStartCmd() *cobra.Command {
	var newWindow bool
	
	cmd := &cobra.Command{
		Use:   "start [path]",
		Short: "Start a Claude session at current or specified node",
		Long: `Start a Claude Code session at the current node or a specified path.
		
If no path is provided, starts at the current node based on CWD.
The working directory will be set to the node's directory.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			
			return mgr.StartClaude(path)
		},
	}
	
	cmd.Flags().BoolVarP(&newWindow, "new-window", "n", false, "Open in new terminal window")
	
	return cmd
}

// newStatusCmd creates the status command
func (a *App) newStatusCmd() *cobra.Command {
	var recursive bool
	
	cmd := &cobra.Command{
		Use:   "status [path]",
		Short: "Show tree and repository status",
		Long: `Show status of the current node or specified path including:
- Tree structure
- Repository states (clean/dirty)
- Branch information
- Uncommitted changes`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			
			return mgr.StatusNode(path, recursive)
		},
	}
	
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Show status recursively")
	
	return cmd
}

// newTreeCmd creates the tree command
func (a *App) newTreeCmd() *cobra.Command {
	var depth int
	
	cmd := &cobra.Command{
		Use:   "tree [path]",
		Short: "Display workspace tree structure",
		Long:  `Display the tree structure of the workspace from current or specified node.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			
			return mgr.ShowTreeAtPath(path, depth)
		},
	}
	
	cmd.Flags().IntVarP(&depth, "depth", "d", 0, "Maximum depth to display (0 for unlimited)")
	
	return cmd
}

// newAddCmd creates the add command
func (a *App) newAddCmd() *cobra.Command {
	var name string
	var lazy bool
	
	cmd := &cobra.Command{
		Use:   "add <repo-url>",
		Short: "Add a child repository to current node",
		Long: `Add a repository as a child of the current node.
		
The repository will be cloned immediately unless --lazy is specified.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.AddRepoSimple(args[0], name, lazy)
		},
	}
	
	cmd.Flags().StringVarP(&name, "name", "n", "", "Custom name for the repository")
	cmd.Flags().BoolVarP(&lazy, "lazy", "l", false, "Don't clone until needed")
	
	return cmd
}

// newRemoveCmd creates the remove command
func (a *App) newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a child repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.RemoveNode(args[0])
		},
	}
}

// newCloneCmd creates the clone command
func (a *App) newCloneCmd() *cobra.Command {
	var recursive bool
	
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone lazy repositories at current node",
		Long:  `Clone repositories marked as lazy that haven't been cloned yet.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.CloneRepos("", recursive)
		},
	}
	
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Clone recursively in subtree")
	
	return cmd
}

// newUseCmd creates the use command
func (a *App) newUseCmd() *cobra.Command {
	var noClone bool
	
	cmd := &cobra.Command{
		Use:   "use <path>",
		Short: "Navigate to a node in the tree",
		Long: `Navigate to a node in the workspace tree.
		
Changes both the current working directory and stored context.
Auto-clones lazy repositories unless --no-clone is specified.
		
Path formats:
- Absolute: /team/backend
- Relative: ../frontend
- Parent: ..
- Current: .
- Previous: -
- Root: ~ or /`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.UseNodeWithClone(args[0], !noClone)
		},
	}
	
	cmd.Flags().BoolVarP(&noClone, "no-clone", "n", false, "Skip auto-cloning lazy repositories")
	
	return cmd
}

// newCurrentCmd creates the current command  
func (a *App) newCurrentCmd() *cobra.Command {
	var clear bool
	
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show current node position",
		Long:  `Display the current node path in the workspace tree.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			if clear {
				return mgr.ClearCurrent()
			}
			
			return mgr.ShowCurrent()
		},
	}
	
	cmd.Flags().BoolVar(&clear, "clear", false, "Clear stored current node")
	
	return cmd
}

// newPullCmd creates the pull command
func (a *App) newPullCmd() *cobra.Command {
	var recursive bool
	
	cmd := &cobra.Command{
		Use:   "pull [path]",
		Short: "Pull repositories at current or specified node",
		Long: `Pull latest changes for repositories at the current node.
		
Target is determined by:
1. Explicit path if provided
2. Current working directory mapping
3. Stored current node
4. Root node`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			
			return mgr.PullNode(path, recursive)
		},
	}
	
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Pull recursively in subtree")
	
	return cmd
}

// newCommitCmd creates the commit command
func (a *App) newCommitCmd() *cobra.Command {
	var message string
	var recursive bool
	
	cmd := &cobra.Command{
		Use:   "commit [path]",
		Short: "Commit changes at current or specified node",
		Long: `Commit changes across repositories at the current node.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if message == "" {
				return fmt.Errorf("commit message is required")
			}
			
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			
			return mgr.CommitNode(path, message, recursive)
		},
	}
	
	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message (required)")
	cmd.MarkFlagRequired("message")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Commit recursively in subtree")
	
	return cmd
}

// newPushCmd creates the push command
func (a *App) newPushCmd() *cobra.Command {
	var recursive bool
	
	cmd := &cobra.Command{
		Use:   "push [path]",
		Short: "Push changes from current or specified node",
		Long: `Push committed changes from repositories at the current node.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			
			return mgr.PushNode(path, recursive)
		},
	}
	
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Push recursively in subtree")
	
	return cmd
}


// newVersionCmd creates the version command
func (a *App) newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(a.stdout, formatVersionDetails())
		},
	}
}

// formatVersion returns the formatted version string
func formatVersion() string {
	return version
}

// formatVersionDetails returns detailed version information
func formatVersionDetails() string {
	versionType := "release"
	if len(version) >= 6 && version[len(version)-6:] == "-dirty" {
		versionType = "dev (uncommitted changes)"
	} else if version == "dev" {
		versionType = "dev (not in git repo)"
	} else if !isReleaseVersion(version) {
		versionType = "dev"
	}
	
	return fmt.Sprintf(`Version:     %s
Type:        %s
Git Commit:  %s
Git Branch:  %s
Build Time:  %s`, version, versionType, gitCommit, gitBranch, buildTime)
}

// isReleaseVersion checks if this is a release version (exact tag)
func isReleaseVersion(v string) bool {
	// Release versions start with 'v' and contain only version numbers
	// e.g., v0.4.0, v1.0.0
	// Dev versions contain additional info like v0.4.0-5-gabcd123
	if len(v) == 0 || v[0] != 'v' {
		return false
	}
	
	// Check if it contains commit info (dash after version number)
	for i := 1; i < len(v); i++ {
		if v[i] == '-' {
			return false
		}
	}
	return true
}