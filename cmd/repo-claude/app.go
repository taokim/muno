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
		Long:  `Clone missing repositories and pull updates for existing ones`,
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

