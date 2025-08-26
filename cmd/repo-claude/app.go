package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/taokim/repo-claude/internal/manager"
)

// App is the scope-based application
type App struct {
	rootCmd *cobra.Command
	stdout  io.Writer
	stderr  io.Writer
}

// NewApp creates a new scope-based application
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
		Short: "Scope-based multi-repository orchestration for Claude Code",
		Long: `Repo-Claude v2 orchestrates Claude Code sessions across multiple 
repositories using isolated scopes for parallel development.

Each scope provides:
- Isolated workspace with independent repository clones
- Separate branch management
- Cross-repository documentation
- Inter-scope coordination via shared memory`,
		Version: formatVersion(),
	}
	
	// Core commands
	a.rootCmd.AddCommand(a.newInitCmd())
	a.rootCmd.AddCommand(a.newListCmd())
	a.rootCmd.AddCommand(a.newStartCmd())
	a.rootCmd.AddCommand(a.newStatusCmd())
	
	// Scope commands
	a.rootCmd.AddCommand(a.newScopeCmd())
	
	// Git operations (all scope-aware)
	a.rootCmd.AddCommand(a.newPullCmd())
	a.rootCmd.AddCommand(a.newCommitCmd())
	a.rootCmd.AddCommand(a.newPushCmd())
	a.rootCmd.AddCommand(a.newBranchCmd())
	a.rootCmd.AddCommand(a.newPRCmd())
	
	// Documentation
	a.rootCmd.AddCommand(a.newDocsCmd())
	
	// Version
	a.rootCmd.AddCommand(a.newVersionCmd())
}

// newInitCmd creates the init command
func (a *App) newInitCmd() *cobra.Command {
	var interactive bool
	
	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new repo-claude project",
		Long: `Initialize a new repo-claude project with scope-based architecture.
		
Creates:
- repo-claude.yaml configuration
- workspaces/ directory for isolated scopes  
- docs/ directory for documentation
- Root CLAUDE.md with instructions`,
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
			
			mgr, err := manager.New(projectPath)
			if err != nil {
				return fmt.Errorf("creating manager: %w", err)
			}
			
			if err := mgr.InitWorkspace(projectName, interactive); err != nil {
				return fmt.Errorf("initializing workspace: %w", err)
			}
			
			return nil
		},
	}
	
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive configuration")
	
	return cmd
}

// newListCmd creates the list command
func (a *App) newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available scopes",
		Long: `List all configured scopes with their repositories and status.
		
Shows:
- Scope number (for easy selection)
- Scope name and type
- Repositories in each scope
- Initialization status`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.ListScopes()
		},
	}
}

// newStartCmd creates the start command
func (a *App) newStartCmd() *cobra.Command {
	var newWindow bool
	
	cmd := &cobra.Command{
		Use:   "start <scope>",
		Short: "Start a Claude session for a scope",
		Long: `Start a Claude Code session for a specific scope.
		
The scope can be specified by:
- Name: rc start backend
- Number: rc start 1 (from 'rc list')
		
On first start, repositories will be cloned.
The working directory will be set to the scope directory.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.StartScope(args[0], newWindow)
		},
	}
	
	cmd.Flags().BoolVarP(&newWindow, "new-window", "n", false, "Open in new terminal window")
	
	return cmd
}

// newStatusCmd creates the status command
func (a *App) newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <scope>",
		Short: "Show scope status",
		Long: `Show detailed status of a scope including:
- Repository states (clean/dirty)
- Branch information
- Uncommitted changes
- Ahead/behind status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.StatusScope(args[0])
		},
	}
}

// newScopeCmd creates the scope management command
func (a *App) newScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scope",
		Short: "Manage scopes",
		Long:  `Create, delete, and manage isolated scopes`,
	}
	
	// Subcommands
	cmd.AddCommand(a.newScopeCreateCmd())
	cmd.AddCommand(a.newScopeDeleteCmd())
	cmd.AddCommand(a.newScopeArchiveCmd())
	
	return cmd
}

// newScopeCreateCmd creates the scope create command
func (a *App) newScopeCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new scope",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.ScopeManager.CreateFromConfig(args[0])
		},
	}
}

// newScopeDeleteCmd creates the scope delete command
func (a *App) newScopeDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a scope",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.ScopeManager.Delete(args[0])
		},
	}
}

// newScopeArchiveCmd creates the scope archive command
func (a *App) newScopeArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <name>",
		Short: "Archive a scope",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.ScopeManager.Archive(args[0])
		},
	}
}

// newPullCmd creates the pull command
func (a *App) newPullCmd() *cobra.Command {
	var cloneMissing bool
	
	cmd := &cobra.Command{
		Use:   "pull <scope>",
		Short: "Pull repositories in a scope",
		Long: `Pull latest changes for all repositories in a scope.
		
Use --clone-missing to clone repositories that haven't been cloned yet.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.PullScope(args[0], cloneMissing)
		},
	}
	
	cmd.Flags().BoolVarP(&cloneMissing, "clone-missing", "c", false, "Clone missing repositories")
	
	return cmd
}

// newCommitCmd creates the commit command
func (a *App) newCommitCmd() *cobra.Command {
	var message string
	
	cmd := &cobra.Command{
		Use:   "commit <scope>",
		Short: "Commit changes in a scope",
		Long:  `Commit all changes across repositories in a scope with the same message.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if message == "" {
				return fmt.Errorf("commit message is required")
			}
			
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.CommitScope(args[0], message)
		},
	}
	
	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message (required)")
	cmd.MarkFlagRequired("message")
	
	return cmd
}

// newPushCmd creates the push command
func (a *App) newPushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push <scope>",
		Short: "Push changes from a scope",
		Long:  `Push all committed changes from repositories in a scope to their remotes.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.PushScope(args[0])
		},
	}
}

// newBranchCmd creates the branch command
func (a *App) newBranchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "branch <scope> <branch-name>",
		Short: "Switch branches in a scope",
		Long: `Switch all repositories in a scope to a specific branch.
		
Creates the branch if it doesn't exist.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.BranchScope(args[0], args[1])
		},
	}
}

// newPRCmd creates the PR command
func (a *App) newPRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Manage pull requests",
		Long:  `Create and manage pull requests for scope repositories`,
	}
	
	// Add PR subcommands here
	// TODO: Implement PR functionality
	
	return cmd
}

// newDocsCmd creates the documentation command
func (a *App) newDocsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Manage documentation",
		Long: `Manage cross-repository documentation.
		
Documentation is stored in:
- docs/global/ for project-wide docs
- docs/scopes/<scope>/ for scope-specific docs`,
	}
	
	cmd.AddCommand(a.newDocsCreateCmd())
	cmd.AddCommand(a.newDocsListCmd())
	cmd.AddCommand(a.newDocsSyncCmd())
	
	return cmd
}

// newDocsCreateCmd creates the docs create command
func (a *App) newDocsCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <scope|global> <filename>",
		Short: "Create documentation",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			scope := args[0]
			filename := args[1]
			
			content := fmt.Sprintf("# %s\n\nCreated: %s\n\n## Overview\n\n",
				filename, time.Now().Format("2006-01-02"))
			
			if scope == "global" {
				return mgr.DocsManager.CreateGlobal(filename, content)
			} else {
				return mgr.DocsManager.CreateScope(scope, filename, content)
			}
		},
	}
}

// newDocsListCmd creates the docs list command
func (a *App) newDocsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [scope]",
		Short: "List documentation",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			scope := ""
			if len(args) > 0 {
				scope = args[0]
			}
			
			docs, err := mgr.DocsManager.List(scope)
			if err != nil {
				return err
			}
			
			if len(docs) == 0 {
				fmt.Fprintln(a.stdout, "No documentation found")
				return nil
			}
			
			fmt.Fprintln(a.stdout, "\n📚 Documentation:")
			for _, doc := range docs {
				fmt.Fprintf(a.stdout, "  %s (%.1f KB)\n", 
					doc.Path, float64(doc.Size)/1024)
			}
			
			return nil
		},
	}
}

// newDocsSyncCmd creates the docs sync command
func (a *App) newDocsSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync documentation to git",
		Long:  `Commit and optionally push documentation changes to git.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.DocsManager.Sync(false)
		},
	}
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