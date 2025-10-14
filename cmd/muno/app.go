package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/manager"
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

// Build-time variables (set via ldflags)
var (
	// GitHubOwner is the GitHub repository owner (set at build time)
	GitHubOwner = "taokim"
	// GitHubRepo is the GitHub repository name (set at build time)
	GitHubRepo = "muno"
)

// getDocumentationURLs returns formatted documentation URLs for MUNO
func getDocumentationURLs() string {
	// Use build-time variables that were set when MUNO was compiled
	// These always point to the documentation of the repository where MUNO was built from
	owner := GitHubOwner
	repo := GitHubRepo
	
	return fmt.Sprintf(`Documentation:
- Web: https://%s.github.io/%s/
- AI Agents: https://%s.github.io/%s/AI_AGENT_GUIDE (READ THIS!)
- Raw docs: https://raw.githubusercontent.com/%s/%s/main/docs/`,
		owner, repo, owner, repo, owner, repo)
}

// setupCommands initializes all commands
func (a *App) setupCommands() {
	a.rootCmd = &cobra.Command{
		Use:   "muno",
		Short: "Multi-repository orchestration for Claude Code with tree-based workspaces",
		Long: fmt.Sprintf(`MUNO (Multi-repository UNified Orchestration) orchestrates Claude Code sessions across multiple 
repositories with tree-based navigation and lazy loading.

Features:
- Tree-based navigation: Navigate workspace like a filesystem
- Lazy loading: Repos clone on-demand when parent is used
- CWD-first resolution: Commands operate based on current directory
- Simple configuration: Everything is just a repository

%s`, getDocumentationURLs()),
		Version: formatVersion(),
	}
	
	// Core commands
	a.rootCmd.AddCommand(a.newInitCmd())
	a.rootCmd.AddCommand(a.newListCmd())
	a.rootCmd.AddCommand(a.newAgentCmd())
	a.rootCmd.AddCommand(a.newClaudeCmd())
	a.rootCmd.AddCommand(a.newGeminiCmd())
	a.rootCmd.AddCommand(a.newStatusCmd())
	
	// Navigation commands
	a.rootCmd.AddCommand(a.newUseCmd())
	a.rootCmd.AddCommand(a.newPathCmd())
	a.rootCmd.AddCommand(a.newCurrentCmd())
	a.rootCmd.AddCommand(a.newShellInitCmd())
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
	
	// Convenience alias
	a.rootCmd.AddCommand(a.newStartCmd())
}

// newInitCmd creates the init command
func (a *App) newInitCmd() *cobra.Command {
	var force bool
	var smart bool
	var nonInteractive bool
	
	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new MUNO project",
		Long: `Initialize a new MUNO project with tree-based workspace.
		
Smart mode (default):
- Detects existing git repositories
- Offers to add them to workspace
- Moves repositories to repos/ directory
- Creates muno.yaml with all repository definitions
		
Creates:
- muno.yaml (v3 configuration with repo list)
- repos/ directory for tree structure
- Root CLAUDE.md with project instructions`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := ""
			projectPath := "."
			
			if len(args) > 0 {
				projectName = args[0]
				// For init, we always use current directory
				projectPath = "."
			} else {
				// Use current directory name
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectName = filepath.Base(cwd)
			}
			
			// For init command, create Manager without loading config
			// since we're about to create it
			
			mgr, err := manager.NewManagerForInit(projectPath)
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
				// Use basic init method
				ctx := context.Background()
				if err := mgr.Initialize(ctx, projectPath); err != nil {
					return fmt.Errorf("initializing workspace: %w", err)
				}
			}
			
			fmt.Fprintf(cmd.OutOrStdout(), "Workspace '%s' initialized successfully\n", projectName)
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

// newAgentCmd creates the agent command
func (a *App) newAgentCmd() *cobra.Command {
	var withContext bool
	
	cmd := &cobra.Command{
		Use:   "agent [agent-name] [path] [-- agent-args]",
		Short: "Start an AI agent session (claude, gemini, cursor, etc.)",
		Long: `Start an AI agent session at the current node or a specified path.

Available agents:
- claude (default) - Anthropic's Claude CLI
- gemini - Google's Gemini CLI (npm install -g @google/gemini-cli)
- cursor - Cursor AI editor
- windsurf - Windsurf AI editor
- aider - Aider AI pair programmer
- Or any other agent CLI installed in your PATH

Usage examples:
  muno agent                    # Start default agent (claude) at current location
  muno agent gemini            # Start Gemini at current location
  muno agent claude backend    # Start Claude at backend node
  muno agent gemini . -- --model pro  # Start Gemini with extra args
  muno agent claude --with-context  # Include MUNO docs for repo organization

The working directory will be set to the node's directory.
Use --with-context when organizing repositories or building workspace hierarchies.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			// Parse arguments: [agent-name] [path] [-- agent-args]
			// Note: Cobra removes the "--" separator, so args after it are just positional args
			agentName := "claude" // default
			path := ""
			var agentArgs []string
			
			// Process arguments
			// If we have args, the first is always the agent name (unless it starts with -)
			argIndex := 0
			if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
				agentName = args[0]
				argIndex++
			}
			
			// The second arg (if present) could be a path or could be an agent arg
			// If it starts with "-", it's likely an agent arg that came after "--"
			if argIndex < len(args) {
				if strings.HasPrefix(args[argIndex], "-") {
					// This is likely an agent argument (user used -- separator)
					agentArgs = args[argIndex:]
				} else {
					// This is a path
					path = args[argIndex]
					argIndex++
					// Everything after the path is agent args
					if argIndex < len(args) {
						agentArgs = args[argIndex:]
					}
				}
			}
			
			return mgr.StartAgent(agentName, path, agentArgs, withContext)
		},
	}
	
	cmd.Flags().BoolVar(&withContext, "with-context", false, "Include MUNO documentation and workspace context for repository organization tasks")
	
	return cmd
}

// newClaudeCmd creates the claude command (alias for agent claude)
func (a *App) newClaudeCmd() *cobra.Command {
	var withContext bool
	
	cmd := &cobra.Command{
		Use:   "claude [path] [-- agent-args]",
		Short: "Start a Claude session (alias for 'agent claude')",
		Long:  `Start a Claude Code session at the current node or a specified path.`,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			// Parse arguments: [path] [-- agent-args]
			// Note: Cobra removes the "--" separator, so args after it are just positional args
			path := ""
			var agentArgs []string
			
			// Process arguments
			if len(args) > 0 {
				// If the first arg starts with "-", it's likely an agent argument (user used -- separator)
				if strings.HasPrefix(args[0], "-") {
					// All args are agent arguments
					agentArgs = args
				} else {
					// First arg is a path
					path = args[0]
					// Everything after the path is agent args
					if len(args) > 1 {
						agentArgs = args[1:]
					}
				}
			}
			
			return mgr.StartAgent("claude", path, agentArgs, withContext)
		},
	}
	
	cmd.Flags().BoolVar(&withContext, "with-context", false, "Include MUNO documentation and workspace context for repository organization tasks")
	
	return cmd
}

// newGeminiCmd creates the gemini command (alias for agent gemini)
func (a *App) newGeminiCmd() *cobra.Command {
	var withContext bool
	
	cmd := &cobra.Command{
		Use:   "gemini [path] [-- agent-args]",
		Short: "Start a Gemini session (alias for 'agent gemini')",
		Long:  `Start a Gemini AI session at the current node or a specified path.`,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			// Parse arguments: [path] [-- agent-args]
			// Note: Cobra removes the "--" separator, so args after it are just positional args
			path := ""
			var agentArgs []string
			
			// Process arguments
			if len(args) > 0 {
				// If the first arg starts with "-", it's likely an agent argument (user used -- separator)
				if strings.HasPrefix(args[0], "-") {
					// All args are agent arguments
					agentArgs = args
				} else {
					// First arg is a path
					path = args[0]
					// Everything after the path is agent args
					if len(args) > 1 {
						agentArgs = args[1:]
					}
				}
			}
			
			return mgr.StartAgent("gemini", path, agentArgs, withContext)
		},
	}
	
	cmd.Flags().BoolVar(&withContext, "with-context", false, "Include MUNO documentation and workspace context for repository organization tasks")
	
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

func (a *App) newPathCmd() *cobra.Command {
	var ensure bool
	var relative bool
	
	cmd := &cobra.Command{
		Use:   "path [target]",
		Short: "Resolve virtual path to physical filesystem path",
		Long: `Resolves a virtual tree path to its physical filesystem location.
		
Without arguments, shows the path of the current directory.
With --relative, shows the position in the tree instead of filesystem path.
With --ensure, clones lazy repositories if needed.
		
Path formats:
  .         Current directory (default)
  ..        Parent directory
  /         Root of workspace
  ~         Root of workspace
  /team/svc Absolute path in tree
  svc       Relative path from current position
		
Examples:
  muno path                     # Current directory's physical path
  muno path . --relative        # Current position in tree
  muno path team/backend        # Resolve to physical path
  muno path ../frontend --ensure # Resolve and clone if needed`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) > 0 {
				target = args[0]
			}
			
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			// Resolve the path
			physicalPath, err := mgr.ResolvePath(target, ensure)
			if err != nil {
				return fmt.Errorf("resolving path: %w", err)
			}
			
			if relative {
				// Show position in tree instead of physical path
				treePath, err := mgr.GetTreePath(physicalPath)
				if err != nil {
					return fmt.Errorf("getting tree path: %w", err)
				}
				fmt.Fprintln(a.stdout, treePath)
			} else {
				// Output the physical path
				fmt.Fprintln(a.stdout, physicalPath)
			}
			
			return nil
		},
	}
	
	cmd.Flags().BoolVar(&ensure, "ensure", false, "Clone lazy repositories if needed")
	cmd.Flags().BoolVar(&relative, "relative", false, "Show position in tree instead of filesystem path")
	
	return cmd
}

func (a *App) newShellInitCmd() *cobra.Command {
	var cmdName string
	var checkOnly bool
	var install bool
	var shellType string
	
	cmd := &cobra.Command{
		Use:   "shell-init",
		Short: "Generate shell integration script for easy navigation",
		Long: `Generate shell integration script that enables easy navigation with muno paths.
		
This creates a shell function (default: mcd) that combines path resolution with cd.
		
Examples:
  muno shell-init                        # Show script for current shell
  muno shell-init --install               # Auto-install to shell config
  muno shell-init --cmd-name goto        # Use 'goto' instead of 'mcd'
  muno shell-init --check                 # Check if command name is available
  muno shell-init --shell bash           # Force bash script`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmdName == "" {
				cmdName = "mcd"
			}
			
			// Detect shell if not specified
			if shellType == "" {
				shellType = detectShell()
			}
			
			// Check for conflicts
			if checkOnly || install {
				if commandExists(cmdName) {
					if checkOnly {
						return fmt.Errorf("command '%s' already exists in shell", cmdName)
					}
					// Suggest alternatives
					suggestions := findAvailableCommandNames(cmdName)
					fmt.Fprintf(os.Stderr, "⚠️  Command '%s' already exists\n", cmdName)
					fmt.Fprintf(os.Stderr, "Available alternatives: %s\n", 
						strings.Join(suggestions, ", "))
					fmt.Fprintf(os.Stderr, "Use --cmd-name to specify different name\n")
					return fmt.Errorf("command name conflict")
				}
				if checkOnly {
					fmt.Fprintf(a.stdout, "✓ Command '%s' is available\n", cmdName)
					return nil
				}
			}
			
			// Generate script
			script := generateShellScript(shellType, cmdName)
			
			if install {
				configFile := getShellConfigFile(shellType)
				
				// Check if already installed
				content, err := os.ReadFile(configFile)
				if err == nil && strings.Contains(string(content), fmt.Sprintf("# MUNO shell integration for %s", cmdName)) {
					fmt.Fprintf(a.stdout, "✓ '%s' is already installed in %s\n", cmdName, configFile)
					return nil
				}
				
				// Append to config file
				f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return fmt.Errorf("opening shell config: %w", err)
				}
				defer f.Close()
				
				if _, err := f.WriteString("\n" + script); err != nil {
					return fmt.Errorf("writing to shell config: %w", err)
				}
				
				fmt.Fprintf(a.stdout, "✅ Installed '%s' function to %s\n", cmdName, configFile)
				fmt.Fprintf(a.stdout, "Run 'source %s' or restart your shell to use it\n", configFile)
				return nil
			}
			
			// Just print it
			fmt.Fprint(a.stdout, script)
			if !checkOnly {
				configFile := getShellConfigFile(shellType)
				fmt.Fprintf(os.Stderr, "\n# To install, run:\n")
				fmt.Fprintf(os.Stderr, "# muno shell-init --cmd-name %s >> %s\n", cmdName, configFile)
				fmt.Fprintf(os.Stderr, "# source %s\n", configFile)
			}
			
			return nil
		},
	}
	
	cmd.Flags().StringVar(&cmdName, "cmd-name", "mcd", "Name for the navigation command")
	cmd.Flags().BoolVar(&checkOnly, "check", false, "Check if command name exists")
	cmd.Flags().BoolVar(&install, "install", false, "Auto-install to shell config")
	cmd.Flags().StringVar(&shellType, "shell", "", "Shell type (bash, zsh, fish)")
	
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
	var force bool
	var all bool
	var configOverrides []string
	var branch string
	var parallel int
	
	cmd := &cobra.Command{
		Use:   "pull [path]",
		Short: "Pull repositories at current or specified node",
		Long: `Pull latest changes for repositories at the current node.
		
Target is determined by:
1. Explicit path if provided
2. Current working directory mapping
3. Stored current node
4. Root node

Use --all to pull all cloned repositories in the workspace.
Use --force to override local changes.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			// Parse CLI config overrides
			if len(configOverrides) > 0 || branch != "" || parallel > 0 {
				cliConfig := make(map[string]interface{})
				
				if len(configOverrides) > 0 {
					parsed, err := config.ParseConfigOverrides(configOverrides)
					if err != nil {
						return fmt.Errorf("parsing config overrides: %w", err)
					}
					for k, v := range parsed {
						cliConfig[k] = v
					}
				}
				
				// Add shorthand flags to CLI config
				if branch != "" {
					if gitCfg, ok := cliConfig["git"].(map[string]interface{}); ok {
						gitCfg["default_branch"] = branch
					} else {
						cliConfig["git"] = map[string]interface{}{"default_branch": branch}
					}
				}
				if parallel > 0 {
					if behaviorCfg, ok := cliConfig["behavior"].(map[string]interface{}); ok {
						behaviorCfg["max_parallel_pulls"] = parallel
					} else {
						cliConfig["behavior"] = map[string]interface{}{"max_parallel_pulls": parallel}
					}
				}
				
				// Set CLI config on manager
				mgr.SetCLIConfig(cliConfig)
			}
			
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			
			// Handle --all flag
			if all {
				path = ""
				recursive = true
			}
			
			return mgr.PullNode(path, recursive, force)
		},
	}
	
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Pull recursively in subtree")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force pull, overriding local changes")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Pull all cloned repositories in workspace")
	cmd.Flags().StringSliceVar(&configOverrides, "config", nil, "Override config values (key=value)")
	cmd.Flags().StringVar(&branch, "branch", "", "Override default branch for this operation")
	cmd.Flags().IntVar(&parallel, "parallel", 0, "Max parallel pull operations")
	
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
func (a *App) newStartCmd() *cobra.Command {
	// Alias for starting default agent (claude)
	cmd := &cobra.Command{
		Use:   "start [path]",
		Short: "Start default agent (claude) at current or specified path",
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
			return mgr.StartAgent("claude", path, nil, false)
		},
	}
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

// Helper functions for shell-init command

func detectShell() string {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return "zsh"
	} else if strings.Contains(shell, "fish") {
		return "fish"
	}
	return "bash" // Default to bash
}

func commandExists(name string) bool {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("command -v %s", name))
	return cmd.Run() == nil
}

func findAvailableCommandNames(base string) []string {
	suggestions := []string{
		base + "d",      // mcd -> mcdd
		"m" + base,      // mcd -> mmcd  
		base + "2",      // mcd -> mcd2
		"go" + base,     // mcd -> gomcd
	}
	
	available := []string{}
	for _, name := range suggestions {
		if !commandExists(name) {
			available = append(available, name)
			if len(available) >= 3 {
				break  // Show max 3 suggestions
			}
		}
	}
	
	return available
}

func getShellConfigFile(shellType string) string {
	home := os.Getenv("HOME")
	switch shellType {
	case "zsh":
		return filepath.Join(home, ".zshrc")
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish")
	default:
		return filepath.Join(home, ".bashrc")
	}
}

func generateShellScript(shellType, cmdName string) string {
	switch shellType {
	case "fish":
		return generateFishScript(cmdName)
	default: // bash/zsh
		return generateBashScript(cmdName)
	}
}

func generateBashScript(cmdName string) string {
	return fmt.Sprintf(`# MUNO shell integration for %s
%s() {
    local target="${1:-.}"
    
    # Special navigation patterns
    case "$target" in
        -)  # Previous location
            target="${_MUNO_PREV:-}"
            [ -z "$target" ] && echo "No previous location" && return 1
            ;;
        ...)  # Grandparent  
            target="../.."
            ;;
    esac
    
    # Save current position for '-' navigation
    _MUNO_PREV="$(muno path . --relative 2>/dev/null || echo '/')"
    
    # Resolve path with lazy clone
    local resolved
    resolved=$(muno path "$target" --ensure 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        cd "$resolved"
        # Show current position in tree
        echo "📍 $(muno path . --relative 2>/dev/null || pwd)"
    else
        echo "❌ Failed to resolve: $target" >&2
        return 1
    fi
}

# Completion support for %s
_%s_complete() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local nodes=$(muno list --format simple 2>/dev/null | grep -v "^$")
    COMPREPLY=($(compgen -W "$nodes" -- "$cur"))
}
complete -F _%s_complete %s

# Optional aliases
alias %st='muno tree'
alias %ss='muno status --recursive'
alias %sl='muno list'
`, cmdName, cmdName, cmdName, cmdName, cmdName, cmdName, cmdName, cmdName, cmdName)
}

func generateFishScript(cmdName string) string {
	return fmt.Sprintf(`# MUNO shell integration for %s
function %s
    set -l target $argv[1]
    test -z "$target" && set target "."
    
    # Special navigation patterns
    switch $target
        case '-'  # Previous location
            if test -z "$_MUNO_PREV"
                echo "No previous location"
                return 1
            end
            set target $_MUNO_PREV
        case '...'  # Grandparent
            set target "../.."
    end
    
    # Save current position
    set -g _MUNO_PREV (muno path . --relative 2>/dev/null; or echo '/')
    
    # Resolve path with lazy clone
    set -l resolved (muno path $target --ensure 2>/dev/null)
    
    if test $status -eq 0
        cd $resolved
        echo "📍 "(muno path . --relative 2>/dev/null; or pwd)
    else
        echo "❌ Failed to resolve: $target" >&2
        return 1
    end
end

# Completion for %s
complete -c %s -a '(muno list --format simple 2>/dev/null)'

# Optional aliases  
alias %st='muno tree'
alias %ss='muno status --recursive'
alias %sl='muno list'
`, cmdName, cmdName, cmdName, cmdName, cmdName, cmdName, cmdName)
}
