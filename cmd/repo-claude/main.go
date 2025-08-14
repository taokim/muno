package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/repo-claude/internal/manager"
)

var (
	version = "1.0.0"
	rootCmd = &cobra.Command{
		Use:   "repo-claude",
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
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		interactive, _ := cmd.Flags().GetBool("interactive")
		
		mgr := manager.New(projectName)
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
	Short: "Sync all repositories using repo tool",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.LoadFromCurrentDir()
		if err != nil {
			return fmt.Errorf("no repo-claude workspace found: %w", err)
		}
		return mgr.Sync()
	},
}

func init() {
	initCmd.Flags().Bool("non-interactive", false, "Use defaults without prompts")
}