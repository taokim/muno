package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/taokim/repo-claude/internal/manager"
)

var (
	version = "dev"
	rootCmd = &cobra.Command{
		Use:   "rc",
		Short: "Multi-agent orchestration using Repo tool and Claude Code",
		Long: `Repo-Claude integrates the Repo tool with Claude Code for 
trunk-based multi-agent development across multiple repositories.`,
		Version: version,
	}
)

func init() {
	// Add commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(forallCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new workspace or reinitialize from existing config",
	Long:  "Initialize a new repo-claude workspace with the given project name, or use current directory if no name provided",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectName string
		var projectPath string
		
		if len(args) > 0 {
			// Project name provided - create new project directory
			projectName = args[0]
			projectPath = projectName
		} else {
			// No project name - use current directory
			projectName = filepath.Base(".")
			projectPath = "."
		}
		
		interactive, _ := cmd.Flags().GetBool("interactive")
		
		mgr := manager.New(projectPath)
		return mgr.InitWorkspace(projectName, interactive)
	},
}

var startCmd = &cobra.Command{
	Use:   "start [agent-name]",
	Short: "Start agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.LoadFromCurrentDir()
		if err != nil {
			return fmt.Errorf("no repo-claude workspace found: %w", err)
		}
		
		if len(args) > 0 {
			return mgr.StartAgent(args[0])
		}
		return mgr.StartAllAgents()
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop [agent-name]",
	Short: "Stop agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.LoadFromCurrentDir()
		if err != nil {
			return fmt.Errorf("no repo-claude workspace found: %w", err)
		}
		
		if len(args) > 0 {
			return mgr.StopAgent(args[0])
		}
		return mgr.StopAllAgents()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show agent and repo status",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.LoadFromCurrentDir()
		if err != nil {
			return fmt.Errorf("no repo-claude workspace found: %w", err)
		}
		return mgr.ShowStatus()
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.LoadFromCurrentDir()
		if err != nil {
			return fmt.Errorf("no repo-claude workspace found: %w", err)
		}
		return mgr.Sync()
	},
}

var forallCmd = &cobra.Command{
	Use:   "forall [command]",
	Short: "Run a command in all repositories",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.LoadFromCurrentDir()
		if err != nil {
			return fmt.Errorf("no repo-claude workspace found: %w", err)
		}
		return mgr.ForAll(args[0], args[1:])
	},
}

func init() {
	initCmd.Flags().Bool("non-interactive", false, "Use defaults without prompts")
}