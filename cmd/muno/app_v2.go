package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/taokim/muno/internal/tree"
)

// AppV2 represents the refactored application with navigator-based architecture
type AppV2 struct {
	rootCmd *cobra.Command
	stdout  io.Writer
	stderr  io.Writer
}

// NewAppV2 creates a new application instance
func NewAppV2() *AppV2 {
	app := &AppV2{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	app.setupCommands()
	return app
}

// SetOutput sets the output writers
func (a *AppV2) SetOutput(stdout, stderr io.Writer) {
	a.stdout = stdout
	a.stderr = stderr
	a.rootCmd.SetOut(stdout)
	a.rootCmd.SetErr(stderr)
}

// Execute runs the application
func (a *AppV2) Execute() error {
	return a.rootCmd.Execute()
}

// ExecuteWithArgs runs with specific arguments
func (a *AppV2) ExecuteWithArgs(args []string) error {
	a.rootCmd.SetArgs(args)
	return a.Execute()
}

// loadManagerV2 loads a ManagerV2 from the current directory
func loadManagerV2() (*tree.ManagerV2, error) {
	// Find workspace root by looking for muno.yaml
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting current directory: %w", err)
	}

	// Walk up to find workspace root
	workspace := cwd
	for {
		// Check for muno.yaml
		configPath := filepath.Join(workspace, "muno.yaml")
		if _, err := os.Stat(configPath); err == nil {
			// Found workspace
			return tree.NewManagerV2(workspace)
		}

		// Move up
		parent := filepath.Dir(workspace)
		if parent == workspace {
			// Reached root without finding config
			return nil, fmt.Errorf("not in a MUNO workspace (no muno.yaml found)")
		}
		workspace = parent
	}
}

// setupCommands initializes all commands
func (a *AppV2) setupCommands() {
	a.rootCmd = &cobra.Command{
		Use:   "muno",
		Short: "Multi-repository orchestration with tree-based navigation",
		Long: `MUNO (Multi-repository UNified Orchestration) orchestrates multiple 
repositories with tree-based navigation and lazy loading.

Features:
- Tree-based navigation with clean interface
- Filesystem-first accuracy (default)
- Optional caching for performance
- Lazy loading support
- Multiple backend strategies`,
		Version: "v2.0.0",
	}

	// Core commands
	a.rootCmd.AddCommand(a.newInitCmd())
	a.rootCmd.AddCommand(a.newTreeCmd())
	a.rootCmd.AddCommand(a.newStatusCmd())
	a.rootCmd.AddCommand(a.newListCmd())

	// Navigation commands
	a.rootCmd.AddCommand(a.newUseCmd())
	a.rootCmd.AddCommand(a.newCurrentCmd())

	// Repository management
	a.rootCmd.AddCommand(a.newAddCmd())
	a.rootCmd.AddCommand(a.newRemoveCmd())
	a.rootCmd.AddCommand(a.newCloneCmd())

	// Git operations
	a.rootCmd.AddCommand(a.newPullCmd())
	a.rootCmd.AddCommand(a.newCommitCmd())
	a.rootCmd.AddCommand(a.newPushCmd())
}

// newInitCmd creates the init command
func (a *AppV2) newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [workspace-name]",
		Short: "Initialize a new MUNO workspace",
		Long:  `Initialize a new MUNO workspace with tree-based navigation.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceName := "workspace"
			if len(args) > 0 {
				workspaceName = args[0]
			}

			// Get current directory
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			// Check if already initialized
			configPath := filepath.Join(cwd, "muno.yaml")
			if _, err := os.Stat(configPath); err == nil {
				return fmt.Errorf("workspace already initialized")
			}

			// Create default config
			// This is simplified - would need proper config creation
			fmt.Fprintf(a.stdout, "Initializing workspace '%s'...\n", workspaceName)
			
			// For now, just create an empty manager
			mgr, err := tree.NewManagerV2(cwd)
			if err != nil {
				return fmt.Errorf("creating manager: %w", err)
			}

			// Save initial config
			if mgr.GetConfig() != nil {
				if err := mgr.GetConfig().Save(configPath); err != nil {
					return fmt.Errorf("saving config: %w", err)
				}
			}

			fmt.Fprintf(a.stdout, "Workspace '%s' initialized successfully\n", workspaceName)
			return nil
		},
	}
	return cmd
}

// newTreeCmd creates the tree command
func (a *AppV2) newTreeCmd() *cobra.Command {
	var depth int

	cmd := &cobra.Command{
		Use:   "tree [path]",
		Short: "Display workspace tree structure",
		Long:  `Display the tree structure of the workspace.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := loadManagerV2()
			if err != nil {
				return err
			}

			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			return mgr.DisplayTree(path, depth)
		},
	}

	cmd.Flags().IntVarP(&depth, "depth", "d", -1, "Maximum depth to display (-1 for unlimited)")
	return cmd
}

// newStatusCmd creates the status command
func (a *AppV2) newStatusCmd() *cobra.Command {
	var recursive bool

	cmd := &cobra.Command{
		Use:   "status [path]",
		Short: "Show repository status",
		Long:  `Show status of repositories in the workspace.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := loadManagerV2()
			if err != nil {
				return err
			}

			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			return mgr.DisplayStatus(path, recursive)
		},
	}

	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Show status recursively")
	return cmd
}

// newListCmd creates the list command
func (a *AppV2) newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [path]",
		Aliases: []string{"ls"},
		Short:   "List child nodes",
		Long:    `List child nodes at the current or specified path.`,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := loadManagerV2()
			if err != nil {
				return err
			}

			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			children, err := mgr.ListChildren(path)
			if err != nil {
				return err
			}

			// Display children
			display := tree.NewDisplay(a.stdout, nil)
			return display.PrintChildren(children)
		},
	}
	return cmd
}

// newUseCmd creates the use command
func (a *AppV2) newUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <path>",
		Short: "Navigate to a node",
		Long: `Navigate to a node in the workspace tree.
		
Path formats:
- Absolute: /backend
- Relative: ../frontend
- Parent: ..
- Root: /`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := loadManagerV2()
			if err != nil {
				return err
			}

			if err := mgr.UseNode(args[0]); err != nil {
				return err
			}

			// Show new location
			currentPath, err := mgr.GetCurrentPath()
			if err != nil {
				return err
			}

			fmt.Fprintf(a.stdout, "Now at: %s\n", currentPath)
			return nil
		},
	}
	return cmd
}

// newCurrentCmd creates the current command
func (a *AppV2) newCurrentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show current position",
		Long:  `Display the current position in the workspace tree.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := loadManagerV2()
			if err != nil {
				return err
			}

			currentPath, err := mgr.GetCurrentPath()
			if err != nil {
				return err
			}

			fmt.Fprintln(a.stdout, currentPath)
			return nil
		},
	}
	return cmd
}

// newAddCmd creates the add command
func (a *AppV2) newAddCmd() *cobra.Command {
	var name string
	var lazy bool

	cmd := &cobra.Command{
		Use:   "add <repo-url>",
		Short: "Add a repository",
		Long:  `Add a repository as a child of the current node.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := loadManagerV2()
			if err != nil {
				return err
			}

			url := args[0]

			// Extract name from URL if not provided
			if name == "" {
				parts := strings.Split(url, "/")
				name = strings.TrimSuffix(parts[len(parts)-1], ".git")
			}

			if err := mgr.AddRepo("", name, url, lazy); err != nil {
				return err
			}

			fmt.Fprintf(a.stdout, "Added repository '%s'\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Custom name for the repository")
	cmd.Flags().BoolVarP(&lazy, "lazy", "l", false, "Don't clone until needed")
	return cmd
}

// newRemoveCmd creates the remove command
func (a *AppV2) newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a node",
		Long:  `Remove a node and its subtree.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := loadManagerV2()
			if err != nil {
				return err
			}

			if err := mgr.RemoveNode(args[0]); err != nil {
				return err
			}

			fmt.Fprintf(a.stdout, "Removed node '%s'\n", args[0])
			return nil
		},
	}
	return cmd
}

// newCloneCmd creates the clone command
func (a *AppV2) newCloneCmd() *cobra.Command {
	var recursive bool

	cmd := &cobra.Command{
		Use:   "clone [path]",
		Short: "Clone lazy repositories",
		Long:  `Clone repositories marked as lazy.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := loadManagerV2()
			if err != nil {
				return err
			}

			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			return mgr.CloneLazyRepos(path, recursive)
		},
	}

	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Clone recursively")
	return cmd
}

// newPullCmd creates the pull command
func (a *AppV2) newPullCmd() *cobra.Command {
	var recursive bool

	cmd := &cobra.Command{
		Use:   "pull [path]",
		Short: "Pull repositories",
		Long:  `Pull latest changes from repositories.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := loadManagerV2()
			if err != nil {
				return err
			}

			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			return mgr.Pull(path, recursive)
		},
	}

	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Pull recursively")
	return cmd
}

// newCommitCmd creates the commit command
func (a *AppV2) newCommitCmd() *cobra.Command {
	var message string
	var recursive bool

	cmd := &cobra.Command{
		Use:   "commit [path]",
		Short: "Commit changes",
		Long:  `Commit changes in repositories.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if message == "" {
				return fmt.Errorf("commit message is required")
			}

			mgr, err := loadManagerV2()
			if err != nil {
				return err
			}

			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			return mgr.Commit(path, message, recursive)
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message (required)")
	cmd.MarkFlagRequired("message")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Commit recursively")
	return cmd
}

// newPushCmd creates the push command
func (a *AppV2) newPushCmd() *cobra.Command {
	var recursive bool

	cmd := &cobra.Command{
		Use:   "push [path]",
		Short: "Push changes",
		Long:  `Push committed changes to remote repositories.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := loadManagerV2()
			if err != nil {
				return err
			}

			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			return mgr.Push(path, recursive)
		},
	}

	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Push recursively")
	return cmd
}