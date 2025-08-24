package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/taokim/repo-claude/internal/manager"
)

// App encapsulates the application for testability
type App struct {
	rootCmd *cobra.Command
	stdout  io.Writer
	stderr  io.Writer
}

// NewApp creates a new application instance
func NewApp() *App {
	app := &App{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	app.setupCommands()
	return app
}

// SetOutput sets the output writers for testing
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

// ExecuteWithArgs runs the application with specific arguments (for testing)
func (a *App) ExecuteWithArgs(args []string) error {
	a.rootCmd.SetArgs(args)
	return a.Execute()
}

// setupCommands initializes all commands
func (a *App) setupCommands() {
	a.rootCmd = &cobra.Command{
		Use:   "rc",
		Short: "Multi-repository orchestration for Claude Code",
		Long: `Repo-Claude orchestrates Claude Code sessions across multiple 
Git repositories using a scope-based approach for collaborative development.`,
		Version: version,
	}
	
	// Add all commands
	a.rootCmd.AddCommand(a.newInitCmd())
	a.rootCmd.AddCommand(a.newStartCmd())
	a.rootCmd.AddCommand(a.newKillCmd())
	a.rootCmd.AddCommand(a.newStatusCmd())
	a.rootCmd.AddCommand(a.newSyncCmd())
	a.rootCmd.AddCommand(a.newForallCmd())
	a.rootCmd.AddCommand(a.newPsCmd())
	a.rootCmd.AddCommand(a.newBranchCmd())
	a.rootCmd.AddCommand(a.newPRCmd())
}

// newInitCmd creates the init command
func (a *App) newInitCmd() *cobra.Command {
	var nonInteractive bool
	
	cmd := &cobra.Command{
		Use:   "init [workspace-name]",
		Short: "Initialize a new workspace",
		Long:  `Initialize a new repo-claude workspace with configuration and directories`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceName := ""
			if len(args) > 0 {
				workspaceName = args[0]
			}
			
			mgr, err := manager.LoadFromCurrentDir()
			if err == nil {
				// Already initialized
				fmt.Fprintln(a.stdout, "Workspace already initialized")
				return nil
			}
			
			// Get current directory if no name provided
			if workspaceName == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getting current directory: %w", err)
				}
				workspaceName = filepath.Base(cwd)
				mgr = manager.New(".")
			} else {
				// Create new project directory
				mgr = manager.New(workspaceName)
			}
			
			return mgr.InitWorkspace(workspaceName, nonInteractive)
		},
	}
	
	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Skip interactive prompts and use defaults")
	
	return cmd
}

// newStartCmd creates the start command
func (a *App) newStartCmd() *cobra.Command {
	var repos []string
	var preset string
	var interactive bool
	var newWindow bool
	
	cmd := &cobra.Command{
		Use:   "start [scope-or-repo-names...]",
		Short: "Start scopes in new terminal windows",
		Long: `Start one or more scopes. Without arguments, starts all auto-start scopes.
All scopes are started in new terminal windows.
		
Examples:
  rc start                    # Start all auto-start scopes
  rc start backend frontend   # Start specific scopes
  rc start order-service      # Start scope containing order-service repo
  rc start --repos backend    # Start scopes for specific repos
  rc start --preset dev       # Start scopes matching a preset
  rc start --interactive      # Choose scopes interactively`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			// Handle different start modes
			if interactive {
				return fmt.Errorf("interactive mode not yet implemented")
			}
			
			if len(repos) > 0 {
				return fmt.Errorf("--repos filtering not yet implemented")
			}
			
			if preset != "" {
				return fmt.Errorf("--preset filtering not yet implemented")
			}
			
			// Default behavior: start specified scopes or all auto-start
			opts := manager.StartOptions{
				NewWindow: newWindow,
			}
			
			// Auto-enable new window when starting multiple items
			if !newWindow && len(args) > 1 {
				opts.NewWindow = true
				fmt.Println("🪟 Opening multiple sessions in new windows")
			}
			
			// Use scopes if configured, otherwise fall back to legacy agents
			if len(mgr.Config.Scopes) > 0 {
				if len(args) == 0 {
					return mgr.StartAllScopesWithOptions(opts)
				}
				
				// Start specific scopes or by repo name
				for _, name := range args {
					// First try as scope name
					if err := mgr.StartScopeWithOptions(name, opts); err != nil {
						// If not a scope, try as repo name
						if err2 := mgr.StartByRepoName(name); err2 != nil {
							return fmt.Errorf("'%s' is neither a scope nor a repository: %v", name, err)
						}
					}
				}
			} else {
				// Legacy agent support
				if len(args) == 0 {
					return mgr.StartAllAgentsWithOptions(opts)
				}
				
				// Start specific agents
				for _, agentName := range args {
					if err := mgr.StartAgentWithOptions(agentName, opts); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
	
	cmd.Flags().StringSliceVarP(&repos, "repos", "r", nil, "Start scopes for specific repositories")
	cmd.Flags().StringVarP(&preset, "preset", "p", "", "Start scopes matching a preset tag")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Choose scopes interactively")
	cmd.Flags().BoolVar(&newWindow, "new-window", false, "Open in new window instead of current terminal")
	
	return cmd
}

// newKillCmd creates the kill command
func (a *App) newKillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kill [scope-names-or-numbers...]",
		Short: "Kill scopes by name or number",
		Long: `Kill one or more scopes. Without arguments, kills all running scopes.
You can use scope names or numbers from 'rc ps' output.
		
Examples:
  rc kill              # Kill all running scopes
  rc kill backend      # Kill by name
  rc kill 1 2          # Kill by numbers from ps output
  rc kill backend 2    # Mix names and numbers`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			// Use scopes if configured, otherwise fall back to legacy agents
			if len(mgr.Config.Scopes) > 0 {
				if len(args) == 0 {
					return mgr.StopAllScopes()
				}
				
				// Kill specific scopes by name or number
				for _, arg := range args {
					// Try to parse as number first
					if num, err := strconv.Atoi(arg); err == nil {
						if err := mgr.KillScopeByNumber(num); err != nil {
							return err
						}
					} else {
						// Treat as scope name
						if err := mgr.StopScope(arg); err != nil {
							return err
						}
					}
				}
			} else {
				// Legacy agent support
				if len(args) == 0 {
					return mgr.StopAllAgents()
				}
				
				// Kill specific agents by name or number
				for _, arg := range args {
					// Try to parse as number first
					if num, err := strconv.Atoi(arg); err == nil {
						if err := mgr.KillByNumber(num); err != nil {
							return err
						}
					} else {
						// Treat as agent name
						if err := mgr.StopAgent(arg); err != nil {
							return err
						}
					}
				}
			}
			return nil
		},
	}
	
	return cmd
}

// newStatusCmd creates the status command
func (a *App) newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show scope and repo status",
		Long:  `Display the status of all scopes and repositories in the workspace`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			return mgr.ShowStatus()
		},
	}
	
	return cmd
}

// newSyncCmd creates the sync command
func (a *App) newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync all repositories",
		Long:  `Clone missing repositories and pull updates for existing ones using rebase`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			return mgr.Sync()
		},
	}
	
	return cmd
}

// newForallCmd creates the forall command
func (a *App) newForallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forall -- command [args...]",
		Short: "Run command in all repositories",
		Long: `Execute a command in all cloned repositories.
		
Example:
  rc forall -- git status
  rc forall -- git pull origin main`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			// First arg is command, rest are arguments
			if len(args) == 0 {
				return fmt.Errorf("no command specified")
			}
			command := args[0]
			cmdArgs := args[1:]
			return mgr.ForAll(command, cmdArgs)
		},
	}
	
	return cmd
}

// newPsCmd creates the ps command
func (a *App) newPsCmd() *cobra.Command {
	var all bool
	var extended bool
	var full bool
	var long bool
	var user string
	var sortBy string
	
	cmd := &cobra.Command{
		Use:   "ps",
		Short: "List scope processes with numbers",
		Long: `Display running scopes with numbers for easy reference.
		
Example output:
  #  SCOPE         STATUS  PID    REPOS
  1  backend       🟢      12345  auth-service, order-service, payment-service
  2  frontend      🟢      12346  web-app, mobile-app
  3  order-flow    ⚫      -      (not running)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			// Use scopes if configured, otherwise fall back to legacy agents
			if len(mgr.Config.Scopes) > 0 {
				return mgr.ListScopes(manager.AgentListOptions{
					ShowAll:     all,
					ShowDetails: extended || full || long,
					Format:      "numbered",  // Changed to numbered format
					SortBy:      sortBy,
				})
			} else {
				// Legacy agent support
				return mgr.ListAgents(manager.AgentListOptions{
					ShowAll:     all,
					ShowDetails: extended || full || long,
					Format:      "numbered",  // Changed to numbered format
					SortBy:      sortBy,
				})
			}
		},
	}
	
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Show all processes including stopped scopes")
	cmd.Flags().BoolVarP(&extended, "extended", "x", false, "Show extended information")
	cmd.Flags().BoolVarP(&full, "full", "f", false, "Show full command lines")
	cmd.Flags().BoolVarP(&long, "long", "l", false, "Long format with detailed info")
	cmd.Flags().StringVarP(&user, "user", "u", "", "Filter by user")
	cmd.Flags().StringVarP(&sortBy, "sort", "s", "name", "Sort by: name, cpu, memory, time")
	
	return cmd
}

// newBranchCmd creates the branch command with subcommands
func (a *App) newBranchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch",
		Short: "Manage branches across repositories",
		Long: `Branch management for all repositories in the workspace.
Useful for creating feature branches and checking branch status across repos.`,
	}
	
	// Add subcommands
	cmd.AddCommand(a.newBranchCreateCmd())
	cmd.AddCommand(a.newBranchListCmd())
	cmd.AddCommand(a.newBranchCheckoutCmd())
	cmd.AddCommand(a.newBranchDeleteCmd())
	
	return cmd
}

// newBranchCreateCmd creates the branch create subcommand
func (a *App) newBranchCreateCmd() *cobra.Command {
	var repos []string
	var fromBranch string
	
	cmd := &cobra.Command{
		Use:   "create <branch-name>",
		Short: "Create a branch in multiple repositories",
		Long: `Create a new branch with the same name across multiple repositories.
		
Examples:
  rc branch create feature/payment      # Create in all repos
  rc branch create feature/auth --repos backend,frontend
  rc branch create hotfix/security --from develop`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			opts := manager.BranchCreateOptions{
				BranchName: branchName,
				FromBranch: fromBranch,
				Repos:      repos,
			}
			
			return mgr.CreateBranch(opts)
		},
	}
	
	cmd.Flags().StringSliceVar(&repos, "repos", nil, "Specific repositories (default: all)")
	cmd.Flags().StringVar(&fromBranch, "from", "", "Base branch to create from (default: current branch)")
	
	return cmd
}

// newBranchListCmd creates the branch list subcommand
func (a *App) newBranchListCmd() *cobra.Command {
	var showAll bool
	var showCurrent bool
	
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List branches across repositories",
		Long: `Show current branch status for all repositories.
		
Examples:
  rc branch list             # Show current branches
  rc branch list --all       # Show all branches
  rc branch list --current   # Show only current branch names`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			opts := manager.BranchListOptions{
				ShowAll:     showAll,
				ShowCurrent: showCurrent,
			}
			
			return mgr.ListBranches(opts)
		},
	}
	
	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all branches")
	cmd.Flags().BoolVarP(&showCurrent, "current", "c", false, "Show only current branch")
	
	return cmd
}

// newBranchCheckoutCmd creates the branch checkout subcommand
func (a *App) newBranchCheckoutCmd() *cobra.Command {
	var repos []string
	var createIfMissing bool
	
	cmd := &cobra.Command{
		Use:   "checkout <branch-name>",
		Short: "Checkout a branch in multiple repositories",
		Long: `Checkout an existing branch across multiple repositories.
		
Examples:
  rc branch checkout main                  # Checkout main in all repos
  rc branch checkout feature/auth --repos backend,frontend
  rc branch checkout develop --create      # Create if doesn't exist`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			opts := manager.BranchCheckoutOptions{
				BranchName:      branchName,
				Repos:           repos,
				CreateIfMissing: createIfMissing,
			}
			
			return mgr.CheckoutBranch(opts)
		},
	}
	
	cmd.Flags().StringSliceVar(&repos, "repos", nil, "Specific repositories (default: all)")
	cmd.Flags().BoolVar(&createIfMissing, "create", false, "Create branch if it doesn't exist")
	
	return cmd
}

// newBranchDeleteCmd creates the branch delete subcommand
func (a *App) newBranchDeleteCmd() *cobra.Command {
	var repos []string
	var force bool
	var deleteRemote bool
	
	cmd := &cobra.Command{
		Use:   "delete <branch-name>",
		Short: "Delete a branch from multiple repositories",
		Long: `Delete a branch from multiple repositories.
		
Examples:
  rc branch delete feature/old          # Delete from all repos
  rc branch delete feature/old --force  # Force delete
  rc branch delete feature/old --remote # Also delete from remote`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			opts := manager.BranchDeleteOptions{
				BranchName:   branchName,
				Repos:        repos,
				Force:        force,
				DeleteRemote: deleteRemote,
			}
			
			return mgr.DeleteBranch(opts)
		},
	}
	
	cmd.Flags().StringSliceVar(&repos, "repos", nil, "Specific repositories (default: all)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force delete even if not merged")
	cmd.Flags().BoolVar(&deleteRemote, "remote", false, "Also delete from remote")
	
	return cmd
}

// newPRCmd creates the pr command with subcommands
func (a *App) newPRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Manage pull requests across repositories",
		Long: `Centralized pull request management for all repositories in the workspace.
Uses GitHub CLI (gh) to interact with GitHub API.

Requires:
  - GitHub CLI (gh) to be installed and authenticated
  - Repositories to have GitHub remotes configured`,
	}
	
	// Add subcommands
	cmd.AddCommand(a.newPRListCmd())
	cmd.AddCommand(a.newPRCreateCmd())
	cmd.AddCommand(a.newPRBatchCreateCmd())
	cmd.AddCommand(a.newPRStatusCmd())
	cmd.AddCommand(a.newPRCheckoutCmd())
	cmd.AddCommand(a.newPRMergeCmd())
	
	return cmd
}

// newPRListCmd creates the pr list subcommand
func (a *App) newPRListCmd() *cobra.Command {
	var state string
	var limit int
	var author string
	var assignee string
	var label string
	
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List PRs across all repositories",
		Long: `List pull requests from all repositories in the workspace.
		
Examples:
  rc pr list                    # List all open PRs
  rc pr list --state all        # List all PRs (open, closed, merged)
  rc pr list --author @me       # List your PRs
  rc pr list --limit 10         # Limit results per repo`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			opts := manager.PRListOptions{
				State:    state,
				Limit:    limit,
				Author:   author,
				Assignee: assignee,
				Label:    label,
			}
			
			return mgr.ListPRs(opts)
		},
	}
	
	cmd.Flags().StringVarP(&state, "state", "s", "open", "Filter by state: open, closed, merged, all")
	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Maximum number of PRs per repository")
	cmd.Flags().StringVarP(&author, "author", "a", "", "Filter by author (@me for current user)")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Filter by assignee")
	cmd.Flags().StringVar(&label, "label", "", "Filter by label")
	
	return cmd
}

// newPRCreateCmd creates the pr create subcommand
func (a *App) newPRCreateCmd() *cobra.Command {
	var title string
	var body string
	var base string
	var draft bool
	var reviewers []string
	var assignees []string
	var labels []string
	var repo string
	
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request in a repository",
		Long: `Create a new pull request in the specified repository.
		
Examples:
  rc pr create --repo backend --title "Fix auth bug"
  rc pr create --repo frontend --draft --title "WIP: New feature"
  rc pr create --repo backend --base develop --title "Feature X"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if repo == "" {
				return fmt.Errorf("--repo flag is required")
			}
			
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			opts := manager.PRCreateOptions{
				Repository: repo,
				Title:      title,
				Body:       body,
				Base:       base,
				Draft:      draft,
				Reviewers:  reviewers,
				Assignees:  assignees,
				Labels:     labels,
			}
			
			return mgr.CreatePR(opts)
		},
	}
	
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository to create PR in (required)")
	cmd.Flags().StringVarP(&title, "title", "t", "", "PR title")
	cmd.Flags().StringVarP(&body, "body", "b", "", "PR body/description")
	cmd.Flags().StringVar(&base, "base", "", "Base branch (default: repository default branch)")
	cmd.Flags().BoolVarP(&draft, "draft", "d", false, "Create as draft PR")
	cmd.Flags().StringSliceVar(&reviewers, "reviewers", nil, "Request reviews from users")
	cmd.Flags().StringSliceVar(&assignees, "assignees", nil, "Assign PR to users")
	cmd.Flags().StringSliceVar(&labels, "labels", nil, "Add labels to PR")
	
	cmd.MarkFlagRequired("repo")
	
	return cmd
}

// newPRBatchCreateCmd creates the pr batch-create subcommand
func (a *App) newPRBatchCreateCmd() *cobra.Command {
	var title string
	var body string
	var base string
	var draft bool
	var reviewers []string
	var assignees []string
	var labels []string
	var repos []string
	var skipMainCheck bool
	
	cmd := &cobra.Command{
		Use:   "batch-create",
		Short: "Create PRs in multiple repositories",
		Long: `Create pull requests in multiple repositories with the same title and body.
Only creates PRs for repositories that are on feature branches (not on main/master).
		
Examples:
  rc pr batch-create --title "Add payment integration"
  rc pr batch-create --title "Fix security issue" --base develop
  rc pr batch-create --title "Feature X" --repos backend,frontend,shared
  rc pr batch-create --title "Emergency fix" --skip-main-check`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return fmt.Errorf("--title flag is required")
			}
			
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			opts := manager.PRBatchCreateOptions{
				Title:         title,
				Body:          body,
				Base:          base,
				Draft:         draft,
				Reviewers:     reviewers,
				Assignees:     assignees,
				Labels:        labels,
				Repos:         repos,
				SkipMainCheck: skipMainCheck,
			}
			
			return mgr.BatchCreatePRs(opts)
		},
	}
	
	cmd.Flags().StringVarP(&title, "title", "t", "", "PR title (required)")
	cmd.Flags().StringVarP(&body, "body", "b", "", "PR body/description")
	cmd.Flags().StringVar(&base, "base", "", "Base branch (default: repository default branch)")
	cmd.Flags().BoolVarP(&draft, "draft", "d", false, "Create as draft PRs")
	cmd.Flags().StringSliceVar(&reviewers, "reviewers", nil, "Request reviews from users")
	cmd.Flags().StringSliceVar(&assignees, "assignees", nil, "Assign PRs to users")
	cmd.Flags().StringSliceVar(&labels, "labels", nil, "Add labels to PRs")
	cmd.Flags().StringSliceVar(&repos, "repos", nil, "Specific repositories (default: all non-main branches)")
	cmd.Flags().BoolVar(&skipMainCheck, "skip-main-check", false, "Skip check for main branch (use with caution)")
	
	cmd.MarkFlagRequired("title")
	
	return cmd
}

// newPRStatusCmd creates the pr status subcommand
func (a *App) newPRStatusCmd() *cobra.Command {
	var repo string
	var number int
	
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show detailed status of PRs",
		Long: `Show detailed status of pull requests including checks and reviews.
		
Examples:
  rc pr status                         # Status of all open PRs
  rc pr status --repo backend          # Status of PRs in backend repo
  rc pr status --repo backend --pr 42  # Status of specific PR`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			opts := manager.PRStatusOptions{
				Repository: repo,
				Number:     number,
			}
			
			return mgr.ShowPRStatus(opts)
		},
	}
	
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Filter by repository")
	cmd.Flags().IntVarP(&number, "pr", "n", 0, "PR number to show status for")
	
	return cmd
}

// newPRCheckoutCmd creates the pr checkout subcommand
func (a *App) newPRCheckoutCmd() *cobra.Command {
	var repo string
	
	cmd := &cobra.Command{
		Use:   "checkout <pr-number>",
		Short: "Checkout a PR branch locally",
		Long: `Checkout a pull request branch locally for review or testing.
		
Examples:
  rc pr checkout 42 --repo backend     # Checkout PR #42 from backend repo
  rc pr checkout 123 --repo frontend   # Checkout PR #123 from frontend repo`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if repo == "" {
				return fmt.Errorf("--repo flag is required")
			}
			
			prNumber, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid PR number: %s", args[0])
			}
			
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			return mgr.CheckoutPR(repo, prNumber)
		},
	}
	
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository containing the PR (required)")
	cmd.MarkFlagRequired("repo")
	
	return cmd
}

// newPRMergeCmd creates the pr merge subcommand
func (a *App) newPRMergeCmd() *cobra.Command {
	var repo string
	var method string
	var deleteRemoteBranch bool
	var deleteLocalBranch bool
	
	cmd := &cobra.Command{
		Use:   "merge <pr-number>",
		Short: "Merge a pull request",
		Long: `Merge a pull request using the specified merge method.
		
Examples:
  rc pr merge 42 --repo backend                    # Merge PR #42
  rc pr merge 42 --repo backend --squash           # Squash and merge
  rc pr merge 42 --repo backend --delete-branch    # Delete branch after merge`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if repo == "" {
				return fmt.Errorf("--repo flag is required")
			}
			
			prNumber, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid PR number: %s", args[0])
			}
			
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			opts := manager.PRMergeOptions{
				Repository:         repo,
				Number:             prNumber,
				Method:             method,
				DeleteRemoteBranch: deleteRemoteBranch,
				DeleteLocalBranch:  deleteLocalBranch,
			}
			
			return mgr.MergePR(opts)
		},
	}
	
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository containing the PR (required)")
	cmd.Flags().StringVarP(&method, "method", "m", "", "Merge method: merge, squash, rebase")
	cmd.Flags().BoolVar(&deleteRemoteBranch, "delete-branch", false, "Delete the remote branch after merge")
	cmd.Flags().BoolVar(&deleteLocalBranch, "delete-local", false, "Delete the local branch after merge")
	cmd.MarkFlagRequired("repo")
	
	return cmd
}
