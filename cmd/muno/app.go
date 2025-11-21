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
- User Guide: https://%s.github.io/%s/guide
- Raw docs: https://raw.githubusercontent.com/%s/%s/main/docs/`,
		owner, repo, owner, repo, owner, repo)
}

// setupCommands initializes all commands
func (a *App) setupCommands() {
	a.rootCmd = &cobra.Command{
		Use:   "muno",
		Short: "Multi-repository orchestration with tree-based workspaces",
		Long: fmt.Sprintf(`MUNO (Multi-repository UNified Orchestration) orchestrates multiple 
repositories with tree-based navigation and lazy loading.

Features:
- Tree-based navigation: Navigate workspace like a filesystem
- Lazy loading: Repos clone on-demand when accessed
- CWD-first resolution: Commands operate based on current directory
- Simple configuration: Everything is just a repository

%s`, getDocumentationURLs()),
		Version: formatVersion(),
	}
	
	// Core commands
	a.rootCmd.AddCommand(a.newInitCmd())
	a.rootCmd.AddCommand(a.newListCmd())
	a.rootCmd.AddCommand(a.newStatusCmd())
	
	// Navigation commands
	a.rootCmd.AddCommand(a.newPathCmd())
	a.rootCmd.AddCommand(a.newShellInitCmd())
	a.rootCmd.AddCommand(a.newTreeCmd())
	
	// Repository management
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
			
			// Check if muno.yaml already exists
			configPath := filepath.Join(projectPath, "muno.yaml")
			if _, err := os.Stat(configPath); err == nil && !force {
				fmt.Fprintf(cmd.OutOrStdout(), "Project already initialized (muno.yaml exists)\n")
				fmt.Fprintf(cmd.OutOrStdout(), "Use --force to reinitialize\n")
				return nil
			}
			
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
	
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force initialization even if muno.yaml already exists")
	cmd.Flags().BoolVar(&smart, "smart", true, "Smart detection of existing git repos")
	cmd.Flags().Bool("no-smart", false, "Disable smart detection")
	cmd.Flags().BoolVarP(&nonInteractive, "non-interactive", "n", false, "Skip all prompts and use defaults")
	
	return cmd
}

// newListCmd creates the list command
func (a *App) newListCmd() *cobra.Command {
	var recursive bool
	var quiet bool
	
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
			
			if quiet {
				return mgr.ListNodesQuiet(recursive)
			}
			
			return mgr.ListNodesRecursive(recursive)
		},
	}
	
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "List recursively")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Output only node names, one per line")
	
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
	var includeLazy bool
	
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone non-lazy repositories at current node",
		Long:  `Clone repositories that haven't been cloned yet. By default, only clones non-lazy repositories.
Use --include-lazy to also clone lazy repositories.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := manager.LoadFromCurrentDir()
			if err != nil {
				return fmt.Errorf("loading workspace: %w", err)
			}
			
			return mgr.CloneRepos("", recursive, includeLazy)
		},
	}
	
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Clone recursively in subtree")
	cmd.Flags().BoolVar(&includeLazy, "include-lazy", false, "Include lazy repositories when cloning")
	
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
				force, _ := cmd.Flags().GetBool("force")
				home := os.Getenv("HOME")
				munoDir := filepath.Join(home, ".muno")
				
				// Create ~/.muno directory
				if err := os.MkdirAll(munoDir, 0755); err != nil {
					return fmt.Errorf("creating .muno directory: %w", err)
				}
				
				// Write all shell scripts to ~/.muno
				shellTypes := []string{"bash", "zsh", "fish"}
				var scriptFiles []string
				
				for _, st := range shellTypes {
					shellScript := generateShellScript(st, cmdName)
					scriptFile := filepath.Join(munoDir, fmt.Sprintf("shell-init-%s.%s", cmdName, st))
					if err := os.WriteFile(scriptFile, []byte(shellScript), 0644); err != nil {
						return fmt.Errorf("writing %s script: %w", st, err)
					}
					scriptFiles = append(scriptFiles, scriptFile)
				}
				
				// Find all existing shell RC files
				rcFiles := []struct {
					path    string
					shell   string
					srcFile string
				}{
					{filepath.Join(home, ".bashrc"), "bash", filepath.Join(munoDir, fmt.Sprintf("shell-init-%s.bash", cmdName))},
					{filepath.Join(home, ".bash_profile"), "bash", filepath.Join(munoDir, fmt.Sprintf("shell-init-%s.bash", cmdName))},
					{filepath.Join(home, ".zshrc"), "zsh", filepath.Join(munoDir, fmt.Sprintf("shell-init-%s.zsh", cmdName))},
					{filepath.Join(home, ".config", "fish", "config.fish"), "fish", filepath.Join(munoDir, fmt.Sprintf("shell-init-%s.fish", cmdName))},
				}
				
				var updatedFiles []string
				sourceLine := fmt.Sprintf("# MUNO %s integration", cmdName)
				
				for _, rc := range rcFiles {
					// Check if RC file exists
					if _, err := os.Stat(rc.path); os.IsNotExist(err) {
						continue
					}
					
					// Ensure parent directory exists for fish config
					if err := os.MkdirAll(filepath.Dir(rc.path), 0755); err != nil {
						continue
					}
					
					content, err := os.ReadFile(rc.path)
					if err != nil {
						continue
					}
					
					alreadyInstalled := strings.Contains(string(content), sourceLine)
					
					if alreadyInstalled && !force {
						continue
					}
					
					// Remove old source lines if force or already installed
					if alreadyInstalled {
						lines := strings.Split(string(content), "\n")
						var newLines []string
						skipNext := false
						
						for _, line := range lines {
							if skipNext {
								skipNext = false
								continue
							}
							if strings.Contains(line, sourceLine) {
								skipNext = true // Skip the source line too
								continue
							}
							newLines = append(newLines, line)
						}
						content = []byte(strings.Join(newLines, "\n"))
					}
					
					// Append source line
					newSourceLine := fmt.Sprintf("\n%s\nsource %s\n", sourceLine, rc.srcFile)
					content = append(content, []byte(newSourceLine)...)
					
					if err := os.WriteFile(rc.path, content, 0644); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to update %s: %v\n", rc.path, err)
						continue
					}
					
					updatedFiles = append(updatedFiles, rc.path)
				}
				
				// Output results
				fmt.Fprintf(a.stdout, "✅ Created shell scripts in %s:\n", munoDir)
				for _, f := range scriptFiles {
					fmt.Fprintf(a.stdout, "   %s\n", f)
				}
				
				if len(updatedFiles) > 0 {
					fmt.Fprintf(a.stdout, "✅ Updated shell configs:\n")
					for _, f := range updatedFiles {
						fmt.Fprintf(a.stdout, "   %s\n", f)
					}
					fmt.Fprintf(a.stdout, "Run 'source <config>' or restart your shell to use '%s'\n", cmdName)
				} else {
					fmt.Fprintf(a.stdout, "ℹ️  No shell config files found or all already up to date\n")
					fmt.Fprintf(a.stdout, "Add this line to your shell config:\n")
					fmt.Fprintf(a.stdout, "source %s\n", filepath.Join(munoDir, fmt.Sprintf("shell-init-%s.<shell>", cmdName)))
				}
				
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
	cmd.Flags().Bool("force", false, "Force reinstall/update even if already installed")
	cmd.Flags().StringVar(&shellType, "shell", "", "Shell type (bash, zsh, fish)")
	
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
		Long: `Pull latest changes for already cloned repositories at the current node.
		
Target is determined by:
1. Explicit path if provided
2. Current working directory mapping
3. Stored current node
4. Root node

Use --all to pull all cloned repositories in the workspace.
Use --force to override local changes.
Note: This command only pulls already cloned repositories. Use 'muno clone' first for new repositories.`,
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
			
			// Pull command never clones new repositories
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
	template := getShellTemplate(shellType)
	return renderShellTemplate(template, cmdName)
}

