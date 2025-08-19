package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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
		Short: "Multi-agent orchestration for Claude Code",
		Long: `Repo-Claude orchestrates multiple Claude Code agents for 
trunk-based development across multiple Git repositories.`,
		Version: version,
	}
	
	// Add all commands
	a.rootCmd.AddCommand(a.newInitCmd())
	a.rootCmd.AddCommand(a.newStartCmd())
	a.rootCmd.AddCommand(a.newStopCmd())
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
			}
			
			mgr = manager.New(".")
			return mgr.InitWorkspace(workspaceName, nonInteractive)
		},
	}
	
	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Skip interactive prompts and use defaults")
	
	return cmd
}

// newStartCmd creates the start command
func (a *App) newStartCmd() *cobra.Command {
	var foreground bool
	var newWindow bool
	var repos []string
	var preset string
	var interactive bool
	
	cmd := &cobra.Command{
		Use:   "start [agent-names...]",
		Short: "Start agents with flexible selection",
		Long: `Start one or more agents. Without arguments, starts all auto-start agents.
		
Examples:
  rc start                    # Start all auto-start agents
  rc start backend frontend   # Start specific agents
  rc start --repos backend    # Start agents for specific repos
  rc start --preset dev       # Start agents matching a preset
  rc start --interactive      # Choose agents interactively`,
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
			
			// Default behavior: start specified agents or all auto-start
			opts := manager.StartOptions{
				Foreground: foreground,
				NewWindow:  newWindow,
			}
			
			if len(args) == 0 {
				return mgr.StartAllAgentsWithOptions(opts)
			}
			
			// Start specific agents
			for _, agentName := range args {
				if err := mgr.StartAgentWithOptions(agentName, opts); err != nil {
					return err
				}
			}
			return nil
		},
	}
	
	cmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run agents in foreground")
	cmd.Flags().BoolVarP(&newWindow, "new-window", "n", false, "Start each agent in a new terminal window")
	cmd.Flags().StringSliceVarP(&repos, "repos", "r", nil, "Start agents for specific repositories")
	cmd.Flags().StringVarP(&preset, "preset", "p", "", "Start agents matching a preset tag")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Choose agents interactively")
	
	return cmd
}

// newStopCmd creates the stop command
func (a *App) newStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop [agent-names...]",
		Short: "Stop agents",
		Long:  `Stop one or more agents. Without arguments, stops all running agents.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			if len(args) == 0 {
				return mgr.StopAllAgents()
			}
			
			// Stop specific agents
			for _, agentName := range args {
				if err := mgr.StopAgent(agentName); err != nil {
					return err
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
		Short: "Show agent and repo status",
		Long:  `Display the status of all agents and repositories in the workspace`,
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
		Short: "List agent processes",
		Long:  `Display running agent processes with detailed information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return err
			}
			
			// For now, just show standard process info
			return mgr.ListAgents(manager.AgentListOptions{
				ShowAll:     all,
				ShowDetails: extended || full || long,
				Format:      "table",
				SortBy:      sortBy,
			})
		},
	}
	
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Show all processes including stopped agents")
	cmd.Flags().BoolVarP(&extended, "extended", "x", false, "Show extended information")
	cmd.Flags().BoolVarP(&full, "full", "f", false, "Show full command lines")
	cmd.Flags().BoolVarP(&long, "long", "l", false, "Long format with detailed info")
	cmd.Flags().StringVarP(&user, "user", "u", "", "Filter by user")
	cmd.Flags().StringVarP(&sortBy, "sort", "s", "name", "Sort by: name, cpu, memory, time")
	
	return cmd
}

// The main function is now simple and testable
func run() int {
	app := NewApp()
	if err := app.Execute(); err != nil {
		fmt.Fprintf(app.stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}