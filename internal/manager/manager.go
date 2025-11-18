package manager

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"time"
	
	"github.com/taokim/muno/internal/adapters"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/constants"
	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/plugin"
	"github.com/taokim/muno/internal/tree"
)

// Embed the AI Agent Context documentation as a file system
//
// Manager is the refactored manager with full dependency injection
type Manager struct {
	// Core dependencies (injected)
	configProvider interfaces.ConfigProvider
	gitProvider    interfaces.GitProvider
	fsProvider     interfaces.FileSystemProvider
	uiProvider     interfaces.UIProvider
	treeProvider   interfaces.TreeProvider
	processProvider interfaces.ProcessProvider
	logProvider    interfaces.LogProvider
	metricsProvider interfaces.MetricsProvider
	pluginManager  interfaces.PluginManager
	
	// Internal state
	workspace    string
	config       *config.ConfigTree
	initialized  bool
	
	// Configuration resolver
	configResolver *config.ConfigResolver
	
	// Options
	opts         ManagerOptions
}

// ManagerOptions for comprehensive dependency injection
type ManagerOptions struct {
	// Required providers
	ConfigProvider  interfaces.ConfigProvider
	GitProvider     interfaces.GitProvider
	FSProvider      interfaces.FileSystemProvider
	UIProvider      interfaces.UIProvider
	TreeProvider    interfaces.TreeProvider
	
	// Optional providers (will use defaults if nil)
	ProcessProvider interfaces.ProcessProvider
	LogProvider     interfaces.LogProvider
	MetricsProvider interfaces.MetricsProvider
	PluginManager   interfaces.PluginManager
	
	// Configuration options
	AutoLoadConfig  bool // Automatically load config on initialization
	EnablePlugins   bool // Enable plugin system
	DebugMode       bool // Enable debug logging
}

// NewManager creates a new manager with dependency injection
func NewManager(opts ManagerOptions) (*Manager, error) {
	// Validate required dependencies
	if opts.ConfigProvider == nil {
		return nil, fmt.Errorf("ConfigProvider is required")
	}
	if opts.GitProvider == nil {
		return nil, fmt.Errorf("GitProvider is required")
	}
	if opts.FSProvider == nil {
		return nil, fmt.Errorf("FSProvider is required")
	}
	if opts.UIProvider == nil {
		return nil, fmt.Errorf("UIProvider is required")
	}
	if opts.TreeProvider == nil {
		return nil, fmt.Errorf("TreeProvider is required")
	}
	
	// Use default providers for optional dependencies
	if opts.ProcessProvider == nil {
		opts.ProcessProvider = NewDefaultProcessProvider()
	}
	if opts.LogProvider == nil {
		opts.LogProvider = NewDefaultLogProvider(opts.DebugMode)
	}
	if opts.MetricsProvider == nil {
		opts.MetricsProvider = NewNoOpMetricsProvider()
	}
	
	// Initialize plugin manager if enabled
	if opts.EnablePlugins && opts.PluginManager == nil {
		pm, err := plugin.NewPluginManager()
		if err != nil {
			return nil, fmt.Errorf("failed to create plugin manager: %w", err)
		}
		opts.PluginManager = pm
	}
	
	mgr := &Manager{
		configProvider:  opts.ConfigProvider,
		gitProvider:     opts.GitProvider,
		fsProvider:      opts.FSProvider,
		uiProvider:      opts.UIProvider,
		treeProvider:    opts.TreeProvider,
		processProvider: opts.ProcessProvider,
		logProvider:     opts.LogProvider,
		metricsProvider: opts.MetricsProvider,
		pluginManager:   opts.PluginManager,
		opts:            opts,
	}
	
	// Initialize config resolver with defaults
	mgr.configResolver = config.NewConfigResolver(config.GetDefaults())
	
	return mgr, nil
}

// Initialize initializes the manager with a workspace
func (m *Manager) Initialize(ctx context.Context, workspace string) error {
	m.logProvider.Info("Initializing manager", 
		interfaces.Field{Key: "workspace", Value: workspace})
	
	timer := m.metricsProvider.Timer("manager.initialize")
	timer.Start()
	defer timer.Stop()
	
	// Set workspace
	m.workspace = workspace
	
	// Ensure workspace directory exists
	if !m.fsProvider.Exists(workspace) {
		m.logProvider.Debug("Creating workspace directory")
		if err := m.fsProvider.MkdirAll(workspace, 0755); err != nil {
			return fmt.Errorf("failed to create workspace: %w", err)
		}
	}
	
	// Load configuration if auto-load is enabled
	if m.opts.AutoLoadConfig {
		configPath := filepath.Join(workspace, "muno.yaml")
		if m.configProvider.Exists(configPath) {
			m.logProvider.Debug("Loading configuration", 
				interfaces.Field{Key: "path", Value: configPath})
			
			cfg, err := m.configProvider.Load(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			
			// Type assert to ConfigTree
			configTree, ok := cfg.(*config.ConfigTree)
			if !ok {
				return fmt.Errorf("invalid config type")
			}
			m.config = configTree
			
			// Set workspace config in resolver
			if m.configResolver != nil && configTree.Overrides != nil {
					m.configResolver.SetWorkspaceConfig(configTree.Overrides)
			}
			
			// Load tree
			if err := m.treeProvider.Load(configTree); err != nil {
				return fmt.Errorf("failed to load tree: %w", err)
			}
		}
	}
	
	// Initialize plugins if enabled
	if m.pluginManager != nil {
		m.logProvider.Debug("Discovering plugins")
		if _, err := m.pluginManager.DiscoverPlugins(ctx); err != nil {
			m.logProvider.Warn("Failed to discover plugins", 
				interfaces.Field{Key: "error", Value: err})
		}
	}
	
	m.initialized = true
	m.metricsProvider.Counter("manager.initialized", 1)
	
	return nil
}

// InitializeWithConfig initializes with a specific configuration
func (m *Manager) InitializeWithConfig(ctx context.Context, workspace string, cfg *config.ConfigTree) error {
	// Initialize workspace
	if err := m.Initialize(ctx, workspace); err != nil {
		return err
	}
	
	// Set config
	m.config = cfg
	
	// Load tree
	if err := m.treeProvider.Load(cfg); err != nil {
		return fmt.Errorf("failed to load tree: %w", err)
	}
	
	// Save config
	configPath := filepath.Join(workspace, "muno.yaml")
	if err := m.configProvider.Save(configPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	
	return nil
}

// SetCLIConfig sets CLI configuration overrides
func (m *Manager) SetCLIConfig(cliConfig map[string]interface{}) {
	if m.configResolver != nil {
		m.configResolver.SetCLIConfig(cliConfig)
	}
	
	// Also update workspace config if already loaded
	if m.config != nil && m.config.Overrides != nil {
			m.configResolver.SetWorkspaceConfig(m.config.Overrides)
	}
}

// GetConfigResolver returns the configuration resolver
func (m *Manager) GetConfigResolver() *config.ConfigResolver {
	return m.configResolver
}



// Add adds a new repository to the tree
func (m *Manager) Add(ctx context.Context, repoURL string, options AddOptions) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	m.logProvider.Info("Adding repository", 
		interfaces.Field{Key: "url", Value: repoURL},
		interfaces.Field{Key: "fetch", Value: options.Fetch})
	
	// Get current node based on pwd
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}
	
	workspaceRoot := m.workspace
	reposDir := filepath.Join(workspaceRoot, m.getReposDir())
	
	var currentPath string
	if strings.HasPrefix(pwd, reposDir) {
		relPath, err := filepath.Rel(reposDir, pwd)
		if err != nil {
			return fmt.Errorf("getting relative path: %w", err)
		}
		if relPath == "." {
			currentPath = "/"
		} else {
			currentPath = "/" + strings.ReplaceAll(filepath.ToSlash(relPath), "\\", "/")
		}
	} else if pwd == workspaceRoot {
		currentPath = "/"
	} else {
		currentPath = "/"
	}
	
	current, err := m.treeProvider.GetNode(currentPath)
	if err != nil {
		return fmt.Errorf("failed to get current node: %w", err)
	}
	
	// Use custom name if provided, otherwise extract from URL
	repoName := options.Name
	if repoName == "" {
		repoName = extractRepoName(repoURL)
	}
	
	// Determine if node should be lazy based on fetch mode
	isLazy := true // Default to lazy
	switch options.Fetch {
	case config.FetchEager:
		isLazy = false
	case config.FetchLazy:
		isLazy = true
	case config.FetchAuto, "":
		// Default to auto mode for smart detection
		// Use smart detection
		isLazy = !tree.IsMetaRepo(repoName)
	}
	
	// Create new node
	newNode := interfaces.NodeInfo{
		Name:       repoName,
		Repository: repoURL,
		IsLazy:     isLazy,
		IsCloned:   false,
	}
	
	// Add to tree
	if err := m.treeProvider.AddNode(current.Path, newNode); err != nil {
		return fmt.Errorf("failed to add node: %w", err)
	}
	
	// Update config to persist the change
	if m.config != nil && current.Path == "/" {
		// Only support adding to root for now
		m.config.Nodes = append(m.config.Nodes, config.NodeDefinition{
			Name:  repoName,
			URL:   repoURL,
			Fetch: options.Fetch,
		})
	}
	
	// Clone immediately if not lazy
	if !isLazy {
		// Compute filesystem path for the new child node
		childPath := filepath.Join(current.Path, repoName)
		repoPath := m.computeFilesystemPath(childPath)
		
		progress := m.uiProvider.Progress(fmt.Sprintf("Cloning %s", repoName))
		progress.Start()
		
		if err := m.gitProvider.Clone(repoURL, repoPath, interfaces.CloneOptions{
			Recursive:     options.Recursive,
			SSHPreference: m.getSSHPreference(),
		}); err != nil {
			progress.Error(err)
			return fmt.Errorf("failed to clone: %w", err)
		}
		
		progress.Finish()
		newNode.IsCloned = true
		
		// Update node state
		if err := m.treeProvider.UpdateNode(filepath.Join(current.Path, repoName), newNode); err != nil {
			m.logProvider.Warn("Failed to update node state", 
				interfaces.Field{Key: "error", Value: err})
		}
	}
	
	// Save configuration
	if err := m.saveConfig(); err != nil {
		m.logProvider.Warn("Failed to save config", 
			interfaces.Field{Key: "error", Value: err})
	}
	
	// Show success with more details
	m.uiProvider.Info("")
	m.uiProvider.Success(fmt.Sprintf("✅ Successfully added: %s", repoName))
	m.uiProvider.Info(fmt.Sprintf("   URL: %s", repoURL))
	if isLazy {
		m.uiProvider.Info("   Status: 💤 Lazy (will clone on first use)")
	} else {
		m.uiProvider.Info("   Status: ✅ Cloned and ready")
	}
	m.uiProvider.Info(fmt.Sprintf("   Location: %s", filepath.Join(current.Path, repoName)))
	m.metricsProvider.Counter("manager.add_repo", 1)
	
	return nil
}

// Remove removes a repository from the tree
func (m *Manager) Remove(ctx context.Context, name string) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	m.logProvider.Info("Removing repository", 
		interfaces.Field{Key: "name", Value: name})
	
	// Get current node based on pwd
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}
	
	workspaceRoot := m.workspace
	reposDir := filepath.Join(workspaceRoot, m.getReposDir())
	
	var currentPath string
	if strings.HasPrefix(pwd, reposDir) {
		relPath, err := filepath.Rel(reposDir, pwd)
		if err != nil {
			return fmt.Errorf("getting relative path: %w", err)
		}
		if relPath == "." {
			currentPath = "/"
		} else {
			currentPath = "/" + strings.ReplaceAll(filepath.ToSlash(relPath), "\\", "/")
		}
	} else if pwd == workspaceRoot {
		currentPath = "/"
	} else {
		currentPath = "/"
	}
	
	current, err := m.treeProvider.GetNode(currentPath)
	if err != nil {
		return fmt.Errorf("failed to get current node: %w", err)
	}
	
	// Find child node
	nodePath := filepath.Join(current.Path, name)
	node, err := m.treeProvider.GetNode(nodePath)
	if err != nil {
		return fmt.Errorf("repository not found: %s", name)
	}
	
	// Confirm removal
	confirm, err := m.uiProvider.Confirm(fmt.Sprintf("Remove %s and all its contents?", name))
	if err != nil {
		return err
	}
	if !confirm {
		m.uiProvider.Info("Removal cancelled")
		return nil
	}
	
	// Remove from filesystem if cloned
	if node.IsCloned {
		repoPath := m.computeFilesystemPath(nodePath)
		if m.fsProvider.Exists(repoPath) {
			m.logProvider.Debug("Removing repository files", 
				interfaces.Field{Key: "path", Value: repoPath})
			
			if err := m.fsProvider.RemoveAll(repoPath); err != nil {
				m.logProvider.Warn("Failed to remove files", 
					interfaces.Field{Key: "error", Value: err})
			}
		}
	}
	
	// Remove from tree
	if err := m.treeProvider.RemoveNode(nodePath); err != nil {
		return fmt.Errorf("failed to remove node: %w", err)
	}
	
	// Update config to persist the change
	if m.config != nil && current.Path == "/" {
		// Only support removing from root for now
		newNodes := []config.NodeDefinition{}
		for _, nodeDef := range m.config.Nodes {
			if nodeDef.Name != name {
				newNodes = append(newNodes, nodeDef)
			}
		}
		m.config.Nodes = newNodes
	}
	
	// Save configuration
	if err := m.saveConfig(); err != nil {
		m.logProvider.Warn("Failed to save config", 
			interfaces.Field{Key: "error", Value: err})
	}
	
	m.uiProvider.Info("")
	m.uiProvider.Success(fmt.Sprintf("🗑️  Successfully removed: %s", name))
	m.uiProvider.Info(fmt.Sprintf("   Path was: %s", nodePath))
	if node.IsCloned {
		m.uiProvider.Info("   Files: Deleted from filesystem")
	}
	m.uiProvider.Info("   File: Updated")
	m.metricsProvider.Counter("manager.remove_repo", 1)
	
	return nil
}

// ExecutePluginCommand executes a plugin command
func (m *Manager) ExecutePluginCommand(ctx context.Context, command string, args []string) error {
	if m.pluginManager == nil {
		return fmt.Errorf("plugins not enabled")
	}
	
	m.logProvider.Info("Executing plugin command", 
		interfaces.Field{Key: "command", Value: command})
	
	result, err := m.pluginManager.ExecuteCommand(ctx, command, args)
	if err != nil {
		return fmt.Errorf("plugin execution failed: %w", err)
	}
	
	if result.Success {
		m.uiProvider.Success(result.Message)
	} else {
		m.uiProvider.Error(result.Message)
		if result.Error != "" {
			return fmt.Errorf("%s", result.Error)
		}
	}
	
	// Handle follow-up actions
	for _, action := range result.Actions {
		if err := m.handlePluginAction(ctx, action); err != nil {
			m.logProvider.Warn("Failed to handle plugin action", 
				interfaces.Field{Key: "action", Value: action.Type},
				interfaces.Field{Key: "error", Value: err})
		}
	}
	
	return nil
}

// handlePluginAction handles a plugin action
func (m *Manager) handlePluginAction(ctx context.Context, action interfaces.Action) error {
	switch action.Type {
	case "command":
		// Execute another command
		return m.ExecutePluginCommand(ctx, action.Command, action.Arguments)
		
	case "navigate":
		// Navigation is no longer supported in stateless mode
		return fmt.Errorf("navigation action not supported in stateless mode")
		
	case "open":
		// Open URL in browser
		if action.URL != "" {
			return m.processProvider.OpenInBrowser(action.URL)
		}
		// Open file in editor
		if action.Path != "" {
			return m.processProvider.OpenInEditor(action.Path)
		}
		
	case "prompt":
		// Prompt user
		response, err := m.uiProvider.Prompt(action.Message)
		if err != nil {
			return err
		}
		m.logProvider.Debug("User response", 
			interfaces.Field{Key: "response", Value: response})
		
	default:
		m.logProvider.Warn("Unknown action type", 
			interfaces.Field{Key: "type", Value: action.Type})
	}
	
	return nil
}

// saveConfig saves the current configuration
func (m *Manager) saveConfig() error {
	if m.config == nil {
		return nil
	}
	
	configPath := filepath.Join(m.workspace, "muno.yaml")
	return m.configProvider.Save(configPath, m.config)
}

// getSSHPreference returns the SSH preference setting from configuration
func (m *Manager) getSSHPreference() bool {
	if m.config != nil {
		return m.config.Defaults.SSHPreference
	}
	// Default to true if config is not available
	return true
}

// Close performs cleanup
func (m *Manager) Close() error {
	// Check if logProvider exists before using it
	if m.logProvider != nil {
		m.logProvider.Info("Closing manager")
	}
	
	// Cleanup plugins
	if m.pluginManager != nil {
		if plugins := m.pluginManager.ListPlugins(); len(plugins) > 0 {
			ctx := context.Background()
			for _, plugin := range plugins {
				if err := m.pluginManager.UnloadPlugin(ctx, plugin.Name); err != nil {
					if m.logProvider != nil {
						m.logProvider.Warn("Failed to unload plugin", 
							interfaces.Field{Key: "plugin", Value: plugin.Name},
							interfaces.Field{Key: "error", Value: err})
					}
				}
			}
		}
	}
	
	// Flush metrics
	if m.metricsProvider != nil {
		if err := m.metricsProvider.Flush(); err != nil {
			if m.logProvider != nil {
				m.logProvider.Warn("Failed to flush metrics", 
					interfaces.Field{Key: "error", Value: err})
			}
		}
	}
	
	return nil
}

// ResolvePath translates MUNO tree paths to filesystem paths
// 
// IMPORTANT CONCEPT: The path command bridges the MUNO tree structure and filesystem.
// In the MUNO tree, there is NO ".nodes" directory - it's hidden as an implementation detail.
// The tree structure is:
//   "/" = workspace root (where muno.yaml lives)
//   "/repo1" = a repository node (physically at .nodes/repo1)
//   "/team/service" = nested node (physically at .nodes/team/repos_dir/service)
//
// This command translates tree positions to filesystem paths, hiding the complexity
// of the actual filesystem structure (like .nodes) from users.
//
// If ensure is true, it will clone lazy repositories if needed
// normalizePathToWorkspaceFormat ensures path has same symlink format as workspace
// On macOS, /var might be /private/var, and we want to match workspace format
func (m *Manager) normalizePathToWorkspaceFormat(path string) string {
	// Check if workspace starts with /private but path doesn't (or vice versa)
	if strings.HasPrefix(m.workspace, "/private/var") && strings.HasPrefix(path, "/var/") && !strings.HasPrefix(path, "/private") {
		// Workspace has /private, path doesn't - add it
		return "/private" + path
	} else if strings.HasPrefix(m.workspace, "/var/") && !strings.HasPrefix(m.workspace, "/private") && strings.HasPrefix(path, "/private/var/") {
		// Workspace doesn't have /private, path does - remove it
		return strings.TrimPrefix(path, "/private")
	}
	return path
}

func (m *Manager) ResolvePath(target string, ensure bool) (string, error) {
	if !m.initialized {
		return "", fmt.Errorf("manager not initialized")
	}

	// Get current directory to resolve relative paths
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %w", err)
	}

	// Normalize cwd to match workspace symlink format
	cwd = m.normalizePathToWorkspaceFormat(cwd)
	
	// Determine current position in tree based on filesystem location
	currentTreePath := "/"
	
	// Use GetTreePath to properly convert filesystem path to tree path
	if m.config != nil {
		treePath, err := m.GetTreePath(cwd)
		if err == nil {
			currentTreePath = treePath
		} else {
			// Fallback: try basic resolution
			reposDir := filepath.Join(m.workspace, m.config.GetReposDir())
			if cwd == reposDir {
				currentTreePath = "/"
			} else if strings.HasPrefix(cwd, reposDir) {
				// This is a rough fallback - GetTreePath should handle most cases
				relPath, _ := filepath.Rel(reposDir, cwd)
				if relPath != "." {
					// Very basic: just remove known implementation directories
					parts := strings.Split(filepath.ToSlash(relPath), "/")
					var treeParts []string
					for _, part := range parts {
						if part != ".nodes" && part != "repos" && part != "" {
							treeParts = append(treeParts, part)
						}
					}
					if len(treeParts) > 0 {
						currentTreePath = "/" + strings.Join(treeParts, "/")
					}
				}
			}
		}
	}
	
	// Resolve target path
	resolvedPath := target
	if target == "." || target == "" {
		// For current directory
		if !ensure {
			// Special case: if we're at the repos directory itself (.nodes),
			// return workspace root since .nodes doesn't exist in the tree
			if m.config != nil {
				reposDir := filepath.Join(m.workspace, m.config.GetReposDir())
				if cwd == reposDir {
					return m.workspace, nil
				}
			}
			return cwd, nil
		}
		resolvedPath = currentTreePath
	} else if target == ".." {
		// For "..", navigate up in the MUNO tree structure
		// If we're at the root ("/"), we can't go up further
		if currentTreePath == "/" || currentTreePath == "" {
			// At workspace root, can't go up
			return m.workspace, nil
		}
		
		// Go up one level in the tree
		parts := strings.Split(strings.TrimPrefix(currentTreePath, "/"), "/")
		if len(parts) > 1 {
			// Go to parent node in the tree
			resolvedPath = "/" + strings.Join(parts[:len(parts)-1], "/")
		} else {
			// Going up from a top-level node goes to root
			resolvedPath = "/"
		}
		// Continue to resolve this path normally
	} else if target == "/" || target == "~" {
		resolvedPath = "/"
	} else if strings.HasPrefix(target, "/") {
		// Absolute path - use as is
		resolvedPath = target
	} else {
		// Relative path
		if currentTreePath == "/" {
			resolvedPath = "/" + target
		} else {
			resolvedPath = currentTreePath + "/" + target
		}
	}
	
	// Clean the path
	resolvedPath = filepath.Clean(resolvedPath)
	resolvedPath = strings.ReplaceAll(resolvedPath, "\\", "/")
	if !strings.HasPrefix(resolvedPath, "/") {
		resolvedPath = "/" + resolvedPath
	}
	
	// If ensure is true, check if the node exists and clone if needed
	if ensure {
		// First, try to find the exact node
		node, err := m.treeProvider.GetNode(resolvedPath)
		// If the exact node wasn't found, check if a parent is a config node
		if err != nil && strings.Contains(resolvedPath, "/") {
			parts := strings.Split(strings.TrimPrefix(resolvedPath, "/"), "/")
			// Check each parent level to find config nodes
			for i := len(parts) - 1; i > 0; i-- {
				parentPath := "/" + strings.Join(parts[:i], "/")
				parentNode, parentErr := m.treeProvider.GetNode(parentPath)
		if parentErr == nil && parentNode.IsConfig && parentNode.ConfigFile != "" {
					// Found a config parent - process it to expand its children
					// When navigating to a specific child, we want to clone it even if lazy
					var toClone []interfaces.NodeInfo
					if err := m.processConfigNode(parentNode, false, true, &toClone); err != nil {
						m.logProvider.Warn(fmt.Sprintf("Failed to process config node during navigation: %v", err))
					}
					// Find and clone the specific child we're navigating to
					targetName := parts[i] // The child name under the config node
					for _, repo := range toClone {
						if repo.Name == targetName {
							clonePath := m.computeFilesystemPath(repo.Path)
							if _, err := os.Stat(clonePath); os.IsNotExist(err) {
								opts := interfaces.CloneOptions{
									Recursive:     false,
									Quiet:         false,
									SSHPreference: m.getSSHPreference(),
								}
								if err := m.gitProvider.Clone(repo.Repository, clonePath, opts); err != nil {
									m.logProvider.Warn(fmt.Sprintf("Failed to clone %s: %v", repo.Name, err))
								}
							}
							break
						}
					}
					break // Found and processed the config parent
				}
			}
		} else if err == nil {
			// Found the exact node
if node.IsConfig && node.ConfigFile != "" {
				// Process config node to find repositories to clone
				var toClone []interfaces.NodeInfo
				if err := m.processConfigNode(node, false, false, &toClone); err != nil {
					m.logProvider.Warn(fmt.Sprintf("Failed to process config node during navigation: %v", err))
				}
				// Clone any found repositories
				for _, repo := range toClone {
					clonePath := m.computeFilesystemPath(repo.Path)
					if _, err := os.Stat(clonePath); os.IsNotExist(err) {
						opts := interfaces.CloneOptions{
							Recursive:     false,
							Quiet:         false,
							SSHPreference: m.getSSHPreference(),
						}
						if err := m.gitProvider.Clone(repo.Repository, clonePath, opts); err != nil {
							m.logProvider.Warn(fmt.Sprintf("Failed to clone %s: %v", repo.Name, err))
						}
					}
				}
			} else if !node.IsCloned && node.Repository != "" {
				// Clone repository
				physPath := m.computeFilesystemPath(resolvedPath)
				// Check if it's actually a cloned repository (has .git directory)
				gitPath := filepath.Join(physPath, ".git")
				if _, err := os.Stat(gitPath); os.IsNotExist(err) {
					// Clone options
					opts := interfaces.CloneOptions{
						Recursive:     false,
						Quiet:         false,
						SSHPreference: m.getSSHPreference(),
					}
					if err := m.gitProvider.Clone(node.Repository, physPath, opts); err != nil {
						return "", fmt.Errorf("cloning lazy repository: %w", err)
					}
					// Update node state
					node.IsLazy = false
					node.IsCloned = true
					m.treeProvider.UpdateNode(resolvedPath, node)
				}
			}
		}
	}
	
	// Validate that the path exists in the tree structure
	// Special case: root always exists
	if resolvedPath != "/" && resolvedPath != "" {
		// Check if the node exists in the tree
		_, err := m.treeProvider.GetNode(resolvedPath)
		if err != nil {
			// Node doesn't exist - check if a parent is a config node that might define this child
			parts := strings.Split(strings.TrimPrefix(resolvedPath, "/"), "/")
			found := false
			
			// Check each parent level for config nodes
			for i := len(parts) - 1; i > 0; i-- {
				parentPath := "/" + strings.Join(parts[:i], "/")
				parentNode, parentErr := m.treeProvider.GetNode(parentPath)
				if parentErr == nil && parentNode.IsConfig && parentNode.ConfigFile != "" {
					// Found a config parent - check if it defines this child
					configPath := parentNode.ConfigFile
					if !filepath.IsAbs(configPath) {
						configPath = filepath.Join(m.workspace, configPath)
					}
					if cfg, loadErr := config.LoadTree(configPath); loadErr == nil && cfg != nil {
						// Check if the config defines the child we're looking for
						childName := parts[i]
						for _, node := range cfg.Nodes {
							if node.Name == childName {
								found = true
								break
							}
						}
					}
					if found {
						break
					}
				}
			}
			
			if !found {
				return "", fmt.Errorf("path does not exist in tree: %s", resolvedPath)
			}
		}
	}
	
	// Compute filesystem path
	physicalPath := m.computeFilesystemPath(resolvedPath)

	// Normalize to match workspace symlink format
	physicalPath = m.normalizePathToWorkspaceFormat(physicalPath)

	return physicalPath, nil
}

// GetTreePath converts a physical filesystem path to its position in the tree
// GetTreePath converts a physical filesystem path to its position in the tree
func (m *Manager) GetTreePath(physicalPath string) (string, error) {
	if !m.initialized {
		return "", fmt.Errorf("manager not initialized")
	}
	
	// Special case: workspace root maps to "/" in the tree
	if physicalPath == m.workspace {
		return "/", nil
	}
	
	// Build the tree path by walking up the filesystem and checking muno.yaml files
	// This properly handles config nodes, repos_dir variations, etc.
	treePath := m.buildTreePathFromFilesystem(physicalPath)
	if treePath != "" {
		return treePath, nil
	}
	
	// Fallback to simple logic if advanced resolution fails
	reposDir := filepath.Join(m.workspace, m.config.GetReposDir())
	
	// Special case: repos directory (.nodes) also maps to "/" since .nodes doesn't exist in the tree
	if physicalPath == reposDir {
		return "/", nil
	}
	
	// Check if the path is within the repos directory
	if !strings.HasPrefix(physicalPath, reposDir) {
		return "", fmt.Errorf("path is not within workspace")
	}
	
	// Extract relative path from repos directory
	relPath, err := filepath.Rel(reposDir, physicalPath)
	if err != nil {
		return "", fmt.Errorf("computing relative path: %w", err)
	}
	
	if relPath == "." {
		return "/", nil
	}
	
	// Simple conversion - just remove .nodes and repos directories
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	var cleanParts []string
	for _, part := range parts {
		if part != ".nodes" && part != "repos" && part != "" {
			cleanParts = append(cleanParts, part)
		}
	}
	
	if len(cleanParts) > 0 {
		treePath = "/" + strings.Join(cleanParts, "/")
	} else {
		treePath = "/"
	}
	
	return treePath, nil
}

// buildTreePathFromFilesystem walks up the filesystem path and builds the correct tree path
// by examining muno.yaml files and understanding the node relationships
func (m *Manager) buildTreePathFromFilesystem(physicalPath string) string {
	// Resolve symlinks in paths for accurate comparison (macOS /var -> /private/var)
	resolvedPath, err := filepath.EvalSymlinks(physicalPath)
	if err != nil {
		resolvedPath = physicalPath
	}

	resolvedWorkspace, err := filepath.EvalSymlinks(m.workspace)
	if err != nil {
		resolvedWorkspace = m.workspace
	}

	// Start from the given path and walk up to find the workspace root
	currentPath := resolvedPath
	pathComponents := []string{}

	// Walk up the directory tree
	for currentPath != "" && currentPath != "/" && currentPath != resolvedWorkspace {
		// Get the directory name
		dirName := filepath.Base(currentPath)
		parentPath := filepath.Dir(currentPath)
		var configFound bool
		var shouldSkip bool
		
		// First check if this node is defined in parent's or grandparent's muno.yaml
		configFound = false
		
		// Check parent's muno.yaml
		parentMunoYaml := filepath.Join(parentPath, "muno.yaml")
		if m.fsProvider != nil && m.fsProvider.Exists(parentMunoYaml) {
			if cfg, err := config.LoadTree(parentMunoYaml); err == nil && cfg != nil {
				for _, node := range cfg.Nodes {
					if node.Name == dirName {
						pathComponents = append([]string{dirName}, pathComponents...)
						currentPath = parentPath
						configFound = true
						goto next_iteration
					}
				}
			}
		}
		
		// If parent is a repos directory (like .nodes), check grandparent's muno.yaml
		if !configFound {
			grandparentPath := filepath.Dir(parentPath)
			grandparentMunoYaml := filepath.Join(grandparentPath, "muno.yaml")
			if m.fsProvider != nil && m.fsProvider.Exists(grandparentMunoYaml) {
				if cfg, err := config.LoadTree(grandparentMunoYaml); err == nil && cfg != nil {
					reposDir := cfg.Workspace.ReposDir
					if reposDir == "" {
						reposDir = ".nodes"
					}
					// Check if parent is indeed the repos directory
					if filepath.Base(parentPath) == reposDir {
						// Check if this node is defined in grandparent's config
						for _, node := range cfg.Nodes {
							if node.Name == dirName {
								pathComponents = append([]string{dirName}, pathComponents...)
								currentPath = grandparentPath
								configFound = true
								goto next_iteration
							}
						}
					}
				}
			}
		}
		
		// Check for special directories to skip
		shouldSkip = false
		if m.fsProvider != nil && m.fsProvider.Exists(filepath.Join(parentPath, "muno.yaml")) {
			if cfg, err := config.LoadTree(filepath.Join(parentPath, "muno.yaml")); err == nil && cfg != nil {
				reposDir := cfg.Workspace.ReposDir
				if reposDir == "" {
					reposDir = ".nodes"
				}
				if dirName == reposDir {
					shouldSkip = true
				}
			}
		}
		
		// Also skip hardcoded implementation directories
		if shouldSkip || dirName == ".nodes" || dirName == "repos" {
			// Skip these implementation directories
			currentPath = parentPath
			continue
		}
		
		// Normal directory - add to path
		if dirName != "" && dirName != "." {
			pathComponents = append([]string{dirName}, pathComponents...)
		}
		currentPath = parentPath
		
		next_iteration:
		// Check if we've reached the workspace
		if currentPath == resolvedWorkspace {
			break
		}
	}

	// Build the tree path only if we reached the workspace
	// The loop exits when currentPath == resolvedWorkspace or we go outside workspace

	if currentPath != resolvedWorkspace {
		return ""
	}
	
	if len(pathComponents) > 0 {
		result := "/" + strings.Join(pathComponents, "/")
		return result
	}
	
	// We reached workspace root with no components (at root)
	return "/"
}

// AddOptions for adding repositories
type AddOptions struct {
	Fetch     string // Fetch mode: "lazy", "eager", or "auto"
	Recursive bool
	Branch    string
	Name      string // Custom name for the repository
}

// InitOptions for workspace initialization
type InitOptions struct {
	CloneOnInit    bool
	Force          bool
	NonInteractive bool
}

// GitRepoInfo contains information about a discovered git repository
type GitRepoInfo struct {
	Path      string
	RemoteURL string
	Branch    string
}


// NewDefaultProcessProvider creates a default process provider
func NewDefaultProcessProvider() interfaces.ProcessProvider {
	// Use the real process adapter for actual command execution
	return adapters.NewProcessAdapter()
}

// NewStubProcessProvider creates a stub process provider for testing
func NewStubProcessProvider() interfaces.ProcessProvider {
	return &DefaultProcessProvider{}
}

// NewStubGitProvider creates a stub git provider for testing
func NewStubGitProvider() interfaces.GitProvider {
	return &StubGitProvider{}
}

// StubGitProvider is a stub implementation of GitProvider for testing
type StubGitProvider struct{}

func (g *StubGitProvider) Clone(url, path string, options interfaces.CloneOptions) error {
	return nil
}

func (g *StubGitProvider) Pull(path string, options interfaces.PullOptions) error {
	return nil
}

func (g *StubGitProvider) Push(path string, options interfaces.PushOptions) error {
	return nil
}

func (g *StubGitProvider) Status(path string) (*interfaces.GitStatus, error) {
	return &interfaces.GitStatus{}, nil
}

func (g *StubGitProvider) Commit(path string, message string, options interfaces.CommitOptions) error {
	return nil
}

func (g *StubGitProvider) Branch(path string) (string, error) {
	return "main", nil
}

func (g *StubGitProvider) Checkout(path string, branch string) error {
	return nil
}

func (g *StubGitProvider) Fetch(path string, options interfaces.FetchOptions) error {
	return nil
}

func (g *StubGitProvider) Add(path string, files []string) error {
	return nil
}

func (g *StubGitProvider) Remove(path string, files []string) error {
	return nil
}

func (g *StubGitProvider) GetRemoteURL(path string) (string, error) {
	return "", nil
}

func (g *StubGitProvider) SetRemoteURL(path string, url string) error {
	return nil
}

// NewStubTreeProvider creates a stub tree provider for testing
func NewStubTreeProvider() interfaces.TreeProvider {
	return &StubTreeProvider{}
}

// NewStubFileSystemProvider creates a stub filesystem provider for testing
func NewStubFileSystemProvider() interfaces.FileSystemProvider {
	return &StubFileSystemProvider{}
}

// NewStubConfigProvider creates a stub config provider for testing
func NewStubConfigProvider() interfaces.ConfigProvider {
	return &StubConfigProvider{}
}

// StubConfigProvider is a stub implementation of ConfigProvider for testing
type StubConfigProvider struct{}

func (c *StubConfigProvider) Load(path string) (interface{}, error) {
	return &config.ConfigTree{}, nil
}

func (c *StubConfigProvider) Save(path string, cfg interface{}) error {
	return nil
}

func (c *StubConfigProvider) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (c *StubConfigProvider) Watch(path string) (<-chan interfaces.ConfigEvent, error) {
	ch := make(chan interfaces.ConfigEvent)
	close(ch)
	return ch, nil
}

// StubFileSystemProvider is a stub implementation of FileSystemProvider for testing
type StubFileSystemProvider struct{}

func (f *StubFileSystemProvider) Exists(path string) bool {
	// Check if path actually exists for basic functionality
	_, err := os.Stat(path)
	return err == nil
}

func (f *StubFileSystemProvider) Create(path string) error {
	return nil
}

func (f *StubFileSystemProvider) Remove(path string) error {
	return nil
}

func (f *StubFileSystemProvider) RemoveAll(path string) error {
	return nil
}

func (f *StubFileSystemProvider) ReadDir(path string) ([]interfaces.FileInfo, error) {
	return []interfaces.FileInfo{}, nil
}

func (f *StubFileSystemProvider) Mkdir(path string, perm os.FileMode) error {
	return os.Mkdir(path, perm)
}

func (f *StubFileSystemProvider) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *StubFileSystemProvider) ReadFile(path string) ([]byte, error) {
	return []byte{}, nil
}

func (f *StubFileSystemProvider) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (f *StubFileSystemProvider) Stat(path string) (interfaces.FileInfo, error) {
	return interfaces.FileInfo{}, nil
}

func (f *StubFileSystemProvider) Symlink(oldname, newname string) error {
	return nil
}

func (f *StubFileSystemProvider) Rename(oldpath, newpath string) error {
	return nil
}

func (f *StubFileSystemProvider) Copy(src, dst string) error {
	return nil
}

func (f *StubFileSystemProvider) Walk(root string, fn filepath.WalkFunc) error {
	return nil
}

// StubTreeProvider is a stub implementation of TreeProvider for testing
type StubTreeProvider struct{}

func (t *StubTreeProvider) Load(config interface{}) error {
	return nil
}

func (t *StubTreeProvider) Navigate(path string) error {
	return nil
}

func (t *StubTreeProvider) GetCurrent() (interfaces.NodeInfo, error) {
	return interfaces.NodeInfo{}, nil
}

func (t *StubTreeProvider) GetTree() (interfaces.NodeInfo, error) {
	return interfaces.NodeInfo{}, nil
}

func (t *StubTreeProvider) GetNode(path string) (interfaces.NodeInfo, error) {
	return interfaces.NodeInfo{}, nil
}

func (t *StubTreeProvider) AddNode(parentPath string, node interfaces.NodeInfo) error {
	return nil
}

func (t *StubTreeProvider) RemoveNode(path string) error {
	return nil
}

func (t *StubTreeProvider) UpdateNode(path string, node interfaces.NodeInfo) error {
	return nil
}

func (t *StubTreeProvider) ListChildren(path string) ([]interfaces.NodeInfo, error) {
	return []interfaces.NodeInfo{}, nil
}

func (t *StubTreeProvider) GetPath() string {
	return ""
}

func (t *StubTreeProvider) SetPath(path string) error {
	return nil
}

func (t *StubTreeProvider) GetState() (interfaces.TreeState, error) {
	return interfaces.TreeState{}, nil
}

func (t *StubTreeProvider) SetState(state interfaces.TreeState) error {
	return nil
}


// DefaultProcessProvider is a simple implementation of ProcessProvider
// NewStubUIProvider creates a stub UI provider for testing
func NewStubUIProvider() interfaces.UIProvider {
	return &StubUIProvider{}
}

// StubUIProvider is a stub implementation of UIProvider for testing
type StubUIProvider struct{}

func (u *StubUIProvider) Prompt(message string) (string, error) {
	return "", nil
}

func (u *StubUIProvider) PromptPassword(message string) (string, error) {
	return "", nil
}

func (u *StubUIProvider) Confirm(message string) (bool, error) {
	// Always confirm in tests
	return true, nil
}

func (u *StubUIProvider) Select(message string, options []string) (string, error) {
	if len(options) > 0 {
		return options[0], nil
	}
	return "", nil
}

func (u *StubUIProvider) MultiSelect(message string, options []string) ([]string, error) {
	return options, nil
}

func (u *StubUIProvider) Progress(message string) interfaces.ProgressReporter {
	return &StubProgressReporter{}
}

func (u *StubUIProvider) Info(message string) {
	// No-op
}

func (u *StubUIProvider) Success(message string) {
	// No-op
}

func (u *StubUIProvider) Warning(message string) {
	// No-op
}

func (u *StubUIProvider) Error(message string) {
	// No-op
}

func (u *StubUIProvider) Debug(message string) {
	// No-op
}


// StubProgressReporter is a stub implementation of ProgressReporter for testing
type StubProgressReporter struct{}

func (p *StubProgressReporter) Start() {
	// No-op
}

func (p *StubProgressReporter) Update(current, total int) {
	// No-op
}

func (p *StubProgressReporter) SetMessage(message string) {
	// No-op
}

func (p *StubProgressReporter) Finish() {
	// No-op
}

func (p *StubProgressReporter) Error(err error) {
	// No-op
}

type DefaultProcessProvider struct{}

func (p *DefaultProcessProvider) ExecuteShell(ctx context.Context, command string, opts interfaces.ProcessOptions) (*interfaces.ProcessResult, error) {
	// Simple implementation
	return &interfaces.ProcessResult{
		ExitCode: 0,
		Stdout:   "",
		Stderr:   "",
	}, nil
}

func (p *DefaultProcessProvider) OpenInBrowser(url string) error {
	return nil
}

func (p *DefaultProcessProvider) OpenInEditor(path string) error {
	return nil
}

func (p *DefaultProcessProvider) Execute(ctx context.Context, name string, args []string, opts interfaces.ProcessOptions) (*interfaces.ProcessResult, error) {
	return &interfaces.ProcessResult{
		ExitCode: 0,
		Stdout:   "",
		Stderr:   "",
	}, nil
}

func (p *DefaultProcessProvider) StartBackground(ctx context.Context, name string, args []string, opts interfaces.ProcessOptions) (interfaces.Process, error) {
	// Return a simple mock process
	return &MockProcess{}, nil
}

// MockProcess is a simple implementation of Process interface
type MockProcess struct{}

func (m *MockProcess) Wait() error { return nil }
func (m *MockProcess) Kill() error { return nil }
func (m *MockProcess) Signal(sig os.Signal) error { return nil }
func (m *MockProcess) Pid() int { return 0 }
func (m *MockProcess) StdoutPipe() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("")), nil }
func (m *MockProcess) StderrPipe() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("")), nil }
func (m *MockProcess) StdinPipe() (io.WriteCloser, error) { return &nopWriteCloser{}, nil }

// nopWriteCloser is a no-op io.WriteCloser
type nopWriteCloser struct{}

func (n *nopWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (n *nopWriteCloser) Close() error { return nil }

// NewDefaultLogProvider creates a default log provider
func NewDefaultLogProvider(debug bool) interfaces.LogProvider {
	return &DefaultLogProvider{debug: debug}
}

// DefaultLogProvider is a simple implementation of LogProvider
type DefaultLogProvider struct {
	debug bool
}

func (l *DefaultLogProvider) Info(msg string, fields ...interfaces.Field) {
	fmt.Printf("%s\n", msg)
}

func (l *DefaultLogProvider) Debug(msg string, fields ...interfaces.Field) {
	if l.debug {
		fmt.Printf("[DEBUG] %s\n", msg)
	}
}

func (l *DefaultLogProvider) Warn(msg string, fields ...interfaces.Field) {
	fmt.Printf("[WARN] %s\n", msg)
}

func (l *DefaultLogProvider) Error(msg string, fields ...interfaces.Field) {
	fmt.Printf("[ERROR] %s\n", msg)
}

func (l *DefaultLogProvider) Fatal(msg string, fields ...interfaces.Field) {
	fmt.Printf("[FATAL] %s\n", msg)
	os.Exit(1)
}

func (l *DefaultLogProvider) SetLevel(level interfaces.LogLevel) {
	// Simple implementation - just store the level
}

func (l *DefaultLogProvider) WithFields(fields ...interfaces.Field) interfaces.LogProvider {
	return l // Just return self for simplicity
}

// NewNoOpMetricsProvider creates a no-op metrics provider
func NewNoOpMetricsProvider() interfaces.MetricsProvider {
	return &NoOpMetricsProvider{}
}

// NoOpMetricsProvider is a no-op implementation of MetricsProvider
type NoOpMetricsProvider struct{}

func (m *NoOpMetricsProvider) Counter(name string, value int64, tags ...string) {}
func (m *NoOpMetricsProvider) Gauge(name string, value float64, tags ...string) {}
func (m *NoOpMetricsProvider) Timer(name string) interfaces.TimerMetric {
	return &NoOpTimer{}
}
func (m *NoOpMetricsProvider) Flush() error {
	return nil
}

func (m *NoOpMetricsProvider) Histogram(name string, value float64, tags ...string) {}

// NoOpTimer is a no-op timer
type NoOpTimer struct{}

func (t *NoOpTimer) Start() {}
func (t *NoOpTimer) Stop() time.Duration { return 0 }
func (t *NoOpTimer) C() <-chan time.Time { 
	// Return a closed channel so it's non-nil but won't block
	ch := make(chan time.Time)
	close(ch)
	return ch
}
func (t *NoOpTimer) Reset() {}
func (t *NoOpTimer) Record(duration time.Duration) {}

// DefaultManagerOptions returns default options for backward compatibility
func DefaultManagerOptions() *ManagerOptions {
	return &ManagerOptions{
		FSProvider:      nil,
		GitProvider:     nil,
		ConfigProvider:  nil,
		UIProvider:      nil,
		ProcessProvider: NewDefaultProcessProvider(),
		LogProvider:     NewDefaultLogProvider(false),
		MetricsProvider: NewNoOpMetricsProvider(),
	}
}

// Helper function to extract repository name from URL
func extractRepoName(url string) string {
	// Remove .git suffix
	if idx := strings.LastIndex(url, ".git"); idx > 0 {
		url = url[:idx]
	}
	
	// Get last path component
	if idx := strings.LastIndex(url, "/"); idx >= 0 {
		return url[idx+1:]
	}
	
	return url
}

// getNodesDir returns the configured nodes directory name
func (m *Manager) getNodesDir() string {
	if m.config != nil {
		return m.config.GetNodesDir()
	}
	return config.GetDefaultNodesDir()
}

// getReposDir is deprecated, use getNodesDir instead
// Kept for backward compatibility
func (m *Manager) getReposDir() string {
	return m.getNodesDir()
}

// computeFilesystemPath computes the actual filesystem path from a logical tree path
// This replicates the logic from tree.Manager.ComputeFilesystemPath
func (m *Manager) computeFilesystemPath(logicalPath string) string {
	nodesDir := m.getNodesDir()
	
	// For root, return the workspace directory (the node that contains muno.yaml)
	// NOT the children directory (.nodes)
	if logicalPath == "/" || logicalPath == "" {
		return m.workspace
	}
	
	// Split the path into parts
	parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")
	
	// Build path iteratively, checking each level for muno.yaml
	currentPath := filepath.Join(m.workspace, nodesDir)
	
	for i, part := range parts {
		if i == 0 {
			// First level goes directly under workspace nodes dir
			currentPath = filepath.Join(currentPath, part)
		} else {
			// For subsequent levels, check if parent has a custom repos_dir
			childReposDir := ""
			
			// Check if the PARENT (not current) is a config node
			if m.treeProvider != nil && i > 0 {
				// The parent is the path up to (but not including) the current part
				parentLogicalPath := "/" + strings.Join(parts[:i], "/")
				parentNode, err := m.treeProvider.GetNode(parentLogicalPath)
				if err == nil && parentNode.IsConfig && parentNode.ConfigFile != "" {
					// Parent is a config reference node, read its config file
					configPath := parentNode.ConfigFile
					if !filepath.IsAbs(configPath) {
						// Make it absolute relative to workspace
						configPath = filepath.Join(m.workspace, configPath)
					}
					if cfg, err := config.LoadTree(configPath); err == nil && cfg != nil && cfg.Workspace.ReposDir != "" {
						childReposDir = cfg.Workspace.ReposDir
					} else {
						childReposDir = constants.DefaultReposDir
					}
				} else if err == nil && parentNode.Repository != "" {
					// Parent is a git repository, check if it has a muno.yaml
					parentMunoYaml := filepath.Join(currentPath, "muno.yaml")
					if m.fsProvider.Exists(parentMunoYaml) {
						// Parent has muno.yaml, use its repos_dir
						childReposDir = constants.DefaultReposDir // default from constants
						if cfg, err := config.LoadTree(parentMunoYaml); err == nil && cfg != nil && cfg.Workspace.ReposDir != "" {
							childReposDir = cfg.Workspace.ReposDir
						}
					} else if m.fsProvider.Exists(filepath.Join(currentPath, ".git")) {
						// Parent is a git repo without muno.yaml, use default
						childReposDir = constants.DefaultReposDir
					}
				}
			}
			
			// If we still don't have a repos_dir and parent is not a config node,
			// check for muno.yaml at the current path
			if childReposDir == "" {
				parentMunoYaml := filepath.Join(currentPath, "muno.yaml")
				if m.fsProvider.Exists(parentMunoYaml) {
					// Parent has muno.yaml, use its repos_dir
					childReposDir = constants.DefaultReposDir // default from constants
					if cfg, err := config.LoadTree(parentMunoYaml); err == nil && cfg != nil && cfg.Workspace.ReposDir != "" {
						childReposDir = cfg.Workspace.ReposDir
					}
				} else if m.fsProvider.Exists(filepath.Join(currentPath, ".git")) {
					// Parent is a git repo without muno.yaml, use default
					childReposDir = constants.DefaultReposDir
				}
			}
			
			// Apply the repos_dir if we found one
			if childReposDir != "" {
				currentPath = filepath.Join(currentPath, childReposDir, part)
			} else {
				// Parent is neither config nor git nor has muno.yaml, continue directly
				currentPath = filepath.Join(currentPath, part)
			}
		}
	}
	
	return currentPath
}

// Helper function to display tree recursively
func (m *Manager) displayTreeRecursive(node interfaces.NodeInfo, indent int) {
	m.displayTreeRecursiveWithPrefix(node, "", true, true)
}

// displayTreeRecursiveWithPrefix displays tree with proper tree characters
func (m *Manager) displayTreeRecursiveWithPrefix(node interfaces.NodeInfo, prefix string, isRoot bool, isLast bool) {
	// Prepare node display
	var output string
	if isRoot {
		output = node.Name
	} else {
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		output = prefix + connector + node.Name
	}
	
	// Add status indicators based on node type
	var status []string
	
	// Check if this is a terminal node (leaf) or non-terminal (parent)
	// Config nodes are never terminal, even if they don't have children loaded yet
	isTerminal := len(node.Children) == 0 && !node.IsConfig
	
	if isTerminal {
		// Terminal nodes (actual repositories)
		if node.Repository != "" {
			status = append(status, "📦")
		}
		if node.IsLazy && !node.IsCloned {
			status = append(status, "💤 lazy")
		} else if !node.IsCloned {
			status = append(status, "⏳ not cloned")
		} else if node.IsCloned {
			status = append(status, "✅")
		}
		if node.HasChanges {
			status = append(status, "📝 modified")
		}
	} else {
		// Non-terminal nodes (parent nodes with children or config nodes)
		// Check if node is explicitly marked as config
		if node.IsConfig {
			status = append(status, "📄 config")
		} else if m.config != nil {
			// Look up in config to determine node type
			nodeFound := false
			for _, nodeDef := range m.config.Nodes {
				if nodeDef.Name == node.Name {
					if nodeDef.File != "" {
						status = append(status, "📄 config")
					} else if nodeDef.URL != "" {
						status = append(status, "📁 git parent")
					}
					nodeFound = true
					break
				}
			}
			if !nodeFound && node.Repository != "" {
				// It's a git parent node
				status = append(status, "📁 git parent")
			} else if !nodeFound {
				// Generic parent node
				status = append(status, "📂 parent")
			}
		} else {
			// No config available, use generic icon
			status = append(status, "📂 parent")
		}
		
		// Show child count for parent nodes
		childCount := len(node.Children)
		if childCount > 0 {
			status = append(status, fmt.Sprintf("%d children", childCount))
		}
	}
	
	if len(status) > 0 {
		output += " [" + strings.Join(status, " ") + "]"
	}
	
	m.uiProvider.Info(output)
	
	// Process children
	childCount := len(node.Children)
	for i, child := range node.Children {
		childPrefix := prefix
		if !isRoot {
			if isLast {
				childPrefix += "    "
			} else {
				childPrefix += "│   "
			}
		}
		isLastChild := (i == childCount - 1)
		m.displayTreeRecursiveWithPrefix(child, childPrefix, false, isLastChild)
	}
}

// Compatibility methods for app.go usage

// ListNodesRecursive lists nodes in the tree
func (m *Manager) ListNodesRecursive(recursive bool) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	tree, err := m.treeProvider.GetTree()
	if err != nil {
		return fmt.Errorf("getting tree: %w", err)
	}
	
	// Display tree using UI provider
	if recursive {
		m.uiProvider.Info("🌳 Repository Tree")
		m.uiProvider.Info("─────────────────")
		m.displayTreeRecursive(tree, 0)
		
		// Show summary
		totalNodes, clonedNodes, lazyNodes := m.countNodes(tree)
		m.uiProvider.Info("")
		m.uiProvider.Info(fmt.Sprintf("📊 Summary: %d total • %d cloned • %d lazy", 
			totalNodes, clonedNodes, lazyNodes))
		return nil
	}
	
	// Get current node based on pwd
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}
	
	// Convert pwd to tree path
	workspaceRoot := m.workspace
	reposDir := filepath.Join(workspaceRoot, m.getReposDir())
	
	var currentPath string
	if strings.HasPrefix(pwd, reposDir) {
		relPath, err := filepath.Rel(reposDir, pwd)
		if err != nil {
			return fmt.Errorf("getting relative path: %w", err)
		}
		if relPath == "." {
			currentPath = "/"
		} else {
			currentPath = "/" + strings.ReplaceAll(filepath.ToSlash(relPath), "\\", "/")
		}
	} else if pwd == workspaceRoot {
		currentPath = "/"
	} else {
		currentPath = "/"
	}
	
	current, err := m.treeProvider.GetNode(currentPath)
	if err != nil {
		return fmt.Errorf("getting current node: %w", err)
	}
	
	m.uiProvider.Info(fmt.Sprintf("📂 Current: %s", current.Path))
	m.uiProvider.Info("─────────────────")
	
	if len(current.Children) == 0 {
		m.uiProvider.Info("  📭 No repositories at this level")
		m.uiProvider.Info("")
		m.uiProvider.Info("  💡 Tip: Use 'muno add <repo-url>' to add repositories")
		return nil
	}
	
	m.uiProvider.Info(fmt.Sprintf("  Found %d repositories:", len(current.Children)))
	m.uiProvider.Info("")
	
	for i, child := range current.Children {
		icon := "📦"
		status := ""
		
		if child.IsLazy && !child.IsCloned {
			icon = "💤"
			status = " (lazy - not cloned)"
		} else if !child.IsCloned {
			icon = "⏳"
			status = " (not cloned)"
		} else if child.HasChanges {
			icon = "📝"
			status = " (modified)"
		} else if child.IsCloned {
			icon = "✅"
			status = " (cloned)"
		}
		
		m.uiProvider.Info(fmt.Sprintf("  %s %s%s", icon, child.Name, status))
		
		if child.Repository != "" && i < 5 { // Show URLs for first 5 repos
			m.uiProvider.Info(fmt.Sprintf("     └─ %s", child.Repository))
		}
	}
	
	if len(current.Children) > 5 {
		m.uiProvider.Info(fmt.Sprintf("\n  ... and %d more", len(current.Children)-5))
	}
	
	return nil
}

// ListNodesQuiet lists node names in quiet mode (one per line, names only)
func (m *Manager) ListNodesQuiet(recursive bool) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	tree, err := m.treeProvider.GetTree()
	if err != nil {
		return fmt.Errorf("getting tree: %w", err)
	}
	
	// Get current location in tree
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}
	
	// Convert pwd to tree path
	workspaceRoot := m.workspace
	reposDir := filepath.Join(workspaceRoot, m.getReposDir())
	
	var currentPath string
	if strings.HasPrefix(pwd, reposDir) {
		relPath, err := filepath.Rel(reposDir, pwd)
		if err != nil {
			return fmt.Errorf("getting relative path: %w", err)
		}
		
		if relPath == "." {
			currentPath = "/"
		} else {
			currentPath = "/" + strings.ReplaceAll(relPath, string(filepath.Separator), "/")
		}
	} else {
		// We're at workspace root
		currentPath = "/"
	}
	
	// Find current node in tree
	current := tree
	if currentPath != "/" {
		parts := strings.Split(strings.Trim(currentPath, "/"), "/")
		for _, part := range parts {
			found := false
			for _, child := range current.Children {
				if child.Name == part {
					current = child
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("path not found in tree: %s", currentPath)
			}
		}
	}
	
	// Output node names
	if recursive {
		m.outputNodesQuietRecursive(current, "")
	} else {
		for _, child := range current.Children {
			fmt.Println(child.Name)
		}
	}
	
	return nil
}

// outputNodesQuietRecursive recursively outputs node names with paths
func (m *Manager) outputNodesQuietRecursive(node interfaces.NodeInfo, pathPrefix string) {
	// Output current node's children
	for _, child := range node.Children {
		childPath := child.Name
		if pathPrefix != "" {
			childPath = pathPrefix + "/" + child.Name
		}
		fmt.Println(childPath)
		
		// Recursively output children
		if len(child.Children) > 0 {
			m.outputNodesQuietRecursive(child, childPath)
		}
	}
}

// ShowTreeAtPath shows the tree at a specific path
func (m *Manager) ShowTreeAtPath(path string, depth int) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	// Use pwd-based resolution if path is empty
	if path == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
		
		workspaceRoot := m.workspace
		reposDir := filepath.Join(workspaceRoot, m.getReposDir())
		
		if strings.HasPrefix(pwd, reposDir) {
			relPath, err := filepath.Rel(reposDir, pwd)
			if err != nil {
				return fmt.Errorf("getting relative path: %w", err)
			}
			if relPath == "." {
				path = "/"
			} else {
				path = "/" + strings.ReplaceAll(filepath.ToSlash(relPath), "\\", "/")
			}
		} else if pwd == workspaceRoot {
			path = "/"
		} else {
			path = "/"
		}
	}
	
	node, err := m.treeProvider.GetNode(path)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}
	
	m.uiProvider.Info("🌳 Repository Tree")
	m.uiProvider.Info("─────────────────")
	m.uiProvider.Info(fmt.Sprintf("📍 Starting from: %s", node.Path))
	m.uiProvider.Info("")
	m.displayTreeRecursive(node, 0)
	
	// Show summary
	totalNodes, clonedNodes, lazyNodes := m.countNodes(node)
	m.uiProvider.Info("")
	m.uiProvider.Info("─────────────────")
	m.uiProvider.Info(fmt.Sprintf("📊 Summary: %d total • %d cloned • %d lazy", 
		totalNodes, clonedNodes, lazyNodes))
		
	return nil
}



// AddRepoSimple adds a repository
func (m *Manager) AddRepoSimple(repoURL string, name string, lazy bool) error {
	// Determine fetch mode based on lazy flag
	fetchMode := config.FetchAuto // Default to auto for smart detection
	if lazy {
		fetchMode = config.FetchLazy
	}
	
	ctx := context.Background()
	return m.Add(ctx, repoURL, AddOptions{Fetch: fetchMode, Name: name})
}

// RemoveNode removes a repository
func (m *Manager) RemoveNode(name string) error {
	ctx := context.Background()
	return m.Remove(ctx, name)
}

// CloneRepos clones lazy repositories
func (m *Manager) CloneRepos(path string, recursive bool, includeLazy bool) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	// Get current node based on pwd
	targetPath := path
	if targetPath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
		
		workspaceRoot := m.workspace
		reposDir := filepath.Join(workspaceRoot, m.getReposDir())
		
		if strings.HasPrefix(pwd, reposDir) {
			relPath, err := filepath.Rel(reposDir, pwd)
			if err != nil {
				return fmt.Errorf("getting relative path: %w", err)
			}
			if relPath == "." {
				targetPath = "/"
			} else {
				targetPath = "/" + strings.ReplaceAll(filepath.ToSlash(relPath), "\\", "/")
			}
		} else if pwd == workspaceRoot {
			targetPath = "/"
		} else {
			targetPath = "/"
		}
	}
	
	current, err := m.treeProvider.GetNode(targetPath)
	if err != nil {
		return fmt.Errorf("getting current node: %w", err)
	}
	
	// Use the unified visit pattern for cloning
	// The visitNodeForClone function handles both config and git nodes uniformly
	clonedCount := 0
	
	if recursive {
		// Visit the current node's children and all their descendants
		for _, child := range current.Children {
			if err := m.visitNodeForClone(child, recursive, includeLazy); err != nil {
				m.logProvider.Warn(fmt.Sprintf("Failed to process child %s: %v", child.Name, err))
			}
		}
		// Count how many repos were cloned (this is a simplified count, could be improved)
		clonedCount = m.countClonedInSubtree(current)
	} else {
		// Non-recursive: only visit direct children
		for _, child := range current.Children {
			if err := m.visitNodeForClone(child, false, includeLazy); err != nil {
				m.logProvider.Warn(fmt.Sprintf("Failed to process child %s: %v", child.Name, err))
			}
		}
		// Count cloned children
		for _, child := range current.Children {
			if child.IsCloned {
				clonedCount++
			}
		}
	}
	
	if clonedCount == 0 {
		m.logProvider.Info("No repositories to clone")
	} else {
		m.logProvider.Info(fmt.Sprintf("Successfully cloned %d repositories", clonedCount))
	}
	
	return m.saveConfig()
}

// countClonedInSubtree counts how many repositories are cloned in a subtree
func (m *Manager) countClonedInSubtree(node interfaces.NodeInfo) int {
	count := 0
	if node.IsCloned && node.Repository != "" {
		count = 1
	}
	
	for _, child := range node.Children {
		count += m.countClonedInSubtree(child)
	}
	
	return count
}

// cloneAndProcessRepositories is deprecated in favor of visitNodeForClone
// This function is kept for backward compatibility in case other parts of the code use it
func (m *Manager) cloneAndProcessRepositories(toClone []interfaces.NodeInfo, recursive bool, includeLazy bool) int {
	clonedCount := 0
	
	// Use the unified visit pattern for each node to clone
	for _, node := range toClone {
		if err := m.visitNodeForClone(node, recursive, includeLazy); err != nil {
			m.logProvider.Warn(fmt.Sprintf("Failed to process node %s: %v", node.Name, err))
		} else if node.Repository != "" {
			clonedCount++ // Count successful repository clones
		}
	}
	
	return clonedCount
}

// countNodes counts total, cloned, and lazy nodes in the tree
func (m *Manager) countNodes(node interfaces.NodeInfo) (total, cloned, lazy int) {
	total = 1
	if node.IsCloned {
		cloned = 1
	}
	if node.IsLazy && !node.IsCloned {
		lazy = 1
	}
	
	for _, child := range node.Children {
		t, c, l := m.countNodes(child)
		total += t
		cloned += c
		lazy += l
	}
	
	return total, cloned, lazy
}

// Helper function to collect lazy nodes recursively
func collectLazyNodes(node interfaces.NodeInfo) []interfaces.NodeInfo {
	var nodes []interfaces.NodeInfo
	
	if node.IsLazy && !node.IsCloned {
		nodes = append(nodes, node)
	}
	
	for _, child := range node.Children {
		nodes = append(nodes, collectLazyNodes(child)...)
	}
	
	return nodes
}

// StatusNode shows git status for a node
func (m *Manager) StatusNode(path string, recursive bool) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	targetPath := path
	if targetPath == "" {
		// Use pwd-based resolution instead of stored state
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
		
		// Check if we're in the workspace
		workspaceRoot := m.workspace
		reposDir := filepath.Join(workspaceRoot, m.getReposDir())
		
		// Convert pwd to tree path
		if strings.HasPrefix(pwd, reposDir) {
			// We're inside the repos directory
			relPath, err := filepath.Rel(reposDir, pwd)
			if err != nil {
				return fmt.Errorf("getting relative path: %w", err)
			}
			
			// Convert to tree path
			if relPath == "." {
				targetPath = "/"
			} else {
				// Clean and convert to forward slashes
				targetPath = "/" + strings.ReplaceAll(filepath.ToSlash(relPath), "\\", "/")
			}
		} else if pwd == workspaceRoot {
			// We're at workspace root
			targetPath = "/"
		} else {
			// Outside workspace - use root
			targetPath = "/"
		}
	}
	
	node, err := m.treeProvider.GetNode(targetPath)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}
	
	// Show status
	m.uiProvider.Info("Tree Status")
	if recursive {
		return m.showStatusRecursive(node)
	}
	
	status, err := m.gitProvider.Status(m.computeFilesystemPath(node.Path))
	if err != nil {
		return fmt.Errorf("getting status: %w", err)
	}
	
	// Display basic status
	statusMsg := fmt.Sprintf("%s: branch=%s", node.Name, status.Branch)
	
	// Add file counts if there are changes
	if !status.IsClean {
		details := []string{}
		if status.HasUntracked {
			untrackedCount := 0
			for _, f := range status.Files {
				if f.Status == "untracked" {
					untrackedCount++
				}
			}
			if untrackedCount > 0 {
				details = append(details, fmt.Sprintf("%d untracked", untrackedCount))
			}
		}
		if status.HasModified {
			modifiedCount := 0
			for _, f := range status.Files {
				if f.Status == "modified" && !f.Staged {
					modifiedCount++
				}
			}
			if modifiedCount > 0 {
				details = append(details, fmt.Sprintf("%d modified", modifiedCount))
			}
		}
		if status.HasStaged {
			stagedCount := 0
			for _, f := range status.Files {
				if f.Staged {
					stagedCount++
				}
			}
			if stagedCount > 0 {
				details = append(details, fmt.Sprintf("%d staged", stagedCount))
			}
		}
		
		if len(details) > 0 {
			statusMsg += " (" + strings.Join(details, ", ") + ")"
		}
		
		// List changed files
		for _, file := range status.Files {
			prefix := "  "
			if file.Staged {
				prefix += "+"
			} else if file.Status == "untracked" {
				prefix += "?"
			} else if file.Status == "modified" {
				prefix += "M"
			} else if file.Status == "deleted" {
				prefix += "D"
			} else if file.Status == "added" {
				prefix += "A"
			}
			m.uiProvider.Info(fmt.Sprintf("%s %s", prefix, file.Path))
		}
	} else {
		statusMsg += " (clean)"
	}
	
	m.uiProvider.Info(statusMsg)
	return nil
}

func (m *Manager) showStatusRecursive(node interfaces.NodeInfo) error {
	// Only check status for cloned repositories
	if node.IsCloned && !node.IsLazy {
		status, err := m.gitProvider.Status(m.computeFilesystemPath(node.Path))
		if err != nil {
			m.uiProvider.Info(fmt.Sprintf("%s: error - %v", node.Name, err))
		} else {
			// Display basic status
			statusMsg := fmt.Sprintf("%s: branch=%s", node.Name, status.Branch)
			
			// Add summary if there are changes
			if !status.IsClean {
				details := []string{}
				if status.HasUntracked {
					untrackedCount := 0
					for _, f := range status.Files {
						if f.Status == "untracked" {
							untrackedCount++
						}
					}
					if untrackedCount > 0 {
						details = append(details, fmt.Sprintf("%d untracked", untrackedCount))
					}
				}
				if status.HasModified {
					modifiedCount := 0
					for _, f := range status.Files {
						if f.Status == "modified" && !f.Staged {
							modifiedCount++
						}
					}
					if modifiedCount > 0 {
						details = append(details, fmt.Sprintf("%d modified", modifiedCount))
					}
				}
				if status.HasStaged {
					stagedCount := 0
					for _, f := range status.Files {
						if f.Staged {
							stagedCount++
						}
					}
					if stagedCount > 0 {
						details = append(details, fmt.Sprintf("%d staged", stagedCount))
					}
				}
				
				if len(details) > 0 {
					statusMsg += " (" + strings.Join(details, ", ") + ")"
				}
			} else {
				statusMsg += " (clean)"
			}
			
			m.uiProvider.Info(statusMsg)
		}
	} else if node.IsLazy && !node.IsCloned {
		// Show lazy repositories that haven't been cloned
		m.uiProvider.Info(fmt.Sprintf("%s: (lazy - not cloned)", node.Name))
	}
	
	for _, child := range node.Children {
		if err := m.showStatusRecursive(child); err != nil {
			return err
		}
	}
	
	return nil
}

// PullNodeWithOptions pulls changes for a node with additional options
func (m *Manager) PullNodeWithOptions(path string, recursive bool, force bool, includeLazy bool) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	// Handle --all case (empty path with recursive flag)
	if path == "" && recursive {
		return m.pullAllRepositories(force)
	}
	
	targetPath := path
	if targetPath == "" {
		// Use pwd-based resolution
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
		
		workspaceRoot := m.workspace
		reposDir := filepath.Join(workspaceRoot, m.getReposDir())
		
		if strings.HasPrefix(pwd, reposDir) {
			relPath, err := filepath.Rel(reposDir, pwd)
			if err != nil {
				return fmt.Errorf("getting relative path: %w", err)
			}
			if relPath == "." {
				targetPath = "/"
			} else {
				targetPath = "/" + strings.ReplaceAll(filepath.ToSlash(relPath), "\\", "/")
			}
		} else if pwd == workspaceRoot {
			targetPath = "/"
		} else {
			targetPath = "/"
		}
	}
	
	node, err := m.treeProvider.GetNode(targetPath)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}
	
	if recursive {
		return m.pullRecursiveWithOptions(node, force, includeLazy)
	}
	
	// Single node pull
	fullPath := m.computeFilesystemPath(node.Path)
	m.uiProvider.Info(fmt.Sprintf("📦 Pulling: %s", node.Name))
	m.uiProvider.Info(fmt.Sprintf("   Path: %s", fullPath))
	
	pullOpts := interfaces.PullOptions{Force: force}
	if err := m.gitProvider.Pull(fullPath, pullOpts); err != nil {
		m.uiProvider.Error(fmt.Sprintf("   ❌ Failed: %v", err))
		return err
	}
	
	m.uiProvider.Success("   ✅ Success")
	return nil
}

// PullNode pulls changes for a node (or all nodes if path is empty and recursive is true)
func (m *Manager) PullNode(path string, recursive bool, force bool) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	// Handle --all case (empty path with recursive flag)
	if path == "" && recursive {
		return m.pullAllRepositories(force)
	}
	
	targetPath := path
	if targetPath == "" {
		// Use pwd-based resolution
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
		
		workspaceRoot := m.workspace
		reposDir := filepath.Join(workspaceRoot, m.getReposDir())
		
		if strings.HasPrefix(pwd, reposDir) {
			relPath, err := filepath.Rel(reposDir, pwd)
			if err != nil {
				return fmt.Errorf("getting relative path: %w", err)
			}
			if relPath == "." {
				targetPath = "/"
			} else {
				targetPath = "/" + strings.ReplaceAll(filepath.ToSlash(relPath), "\\", "/")
			}
		} else if pwd == workspaceRoot {
			targetPath = "/"
		} else {
			targetPath = "/"
		}
	}
	
	node, err := m.treeProvider.GetNode(targetPath)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}
	
	if recursive {
		return m.pullRecursive(node, force)
	}
	
	// Single node pull
	fullPath := m.computeFilesystemPath(node.Path)
	m.uiProvider.Info(fmt.Sprintf("📦 Pulling: %s", node.Name))
	m.uiProvider.Info(fmt.Sprintf("   Path: %s", fullPath))
	
	pullOpts := interfaces.PullOptions{Force: force}
	if err := m.gitProvider.Pull(fullPath, pullOpts); err != nil {
		m.uiProvider.Error(fmt.Sprintf("   ❌ Failed: %v", err))
		return err
	}
	
	m.uiProvider.Success("   ✅ Success")
	return nil
}

// pullAllRepositories pulls all cloned repositories in the workspace
func (m *Manager) pullAllRepositories(force bool) error {
	// Get root node
	root, err := m.treeProvider.GetNode("/")
	if err != nil {
		return fmt.Errorf("getting root node: %w", err)
	}
	
	m.uiProvider.Info("🔄 Pulling all repositories...")
	m.uiProvider.Info("─────────────────")
	
	// Collect all cloned repositories
	allRepos := m.collectClonedRepos(root)
	if len(allRepos) == 0 {
		m.uiProvider.Info("📭 No cloned repositories found")
		return nil
	}
	
	m.uiProvider.Info(fmt.Sprintf("Found %d repositories to pull", len(allRepos)))
	m.uiProvider.Info("")
	
	successCount := 0
	failedRepos := []string{}
	
	for _, node := range allRepos {
		fullPath := m.computeFilesystemPath(node.Path)
		m.uiProvider.Info(fmt.Sprintf("📦 Pulling: %s", node.Name))
		
		pullOpts := interfaces.PullOptions{Force: force}
		
		if err := m.gitProvider.Pull(fullPath, pullOpts); err != nil {
			m.uiProvider.Error(fmt.Sprintf("   ❌ Failed: %v", err))
			failedRepos = append(failedRepos, node.Name)
		} else {
			m.uiProvider.Success("   ✅ Success")
			successCount++
		}
	}
	
	// Show summary
	m.uiProvider.Info("")
	m.uiProvider.Info("─────────────────")
	m.uiProvider.Info(fmt.Sprintf("📊 Results: %d succeeded, %d failed", 
		successCount, len(failedRepos)))
	
	if len(failedRepos) > 0 {
		m.uiProvider.Info("")
		m.uiProvider.Warning("⚠️  Failed repositories:")
		for _, repo := range failedRepos {
			m.uiProvider.Info(fmt.Sprintf("   - %s", repo))
		}
		if !force {
			m.uiProvider.Info("")
			m.uiProvider.Info("💡 Tip: Use --force to override local changes")
		}
	}
	
	return nil
}

// collectClonedRepos collects all cloned repositories recursively
func (m *Manager) collectClonedRepos(node interfaces.NodeInfo) []interfaces.NodeInfo {
	var repos []interfaces.NodeInfo
	
	// Handle config nodes - expand them to find repositories
	if node.IsConfig && node.ConfigFile != "" {
		// Load the config and process its children
		configFilePath := node.ConfigFile
		if !filepath.IsAbs(configFilePath) && !strings.HasPrefix(configFilePath, "http") {
			configFilePath = m.resolveConfigPath(node.ConfigFile, node.Path)
		}
		
		if cfg, err := config.LoadTree(configFilePath); err == nil {
			// Process each child node from the config
			for _, nodeDef := range cfg.Nodes {
				childPath := node.Path + "/" + nodeDef.Name
				if node.Path == "/" {
					childPath = "/" + nodeDef.Name
				}
				
				if nodeDef.URL != "" {
					// It's a repository - check if it's cloned
					childFsPath := m.computeFilesystemPath(childPath)
					if _, err := os.Stat(filepath.Join(childFsPath, ".git")); err == nil {
						// Repository is cloned, add it
						repos = append(repos, interfaces.NodeInfo{
							Name:       nodeDef.Name,
							Path:       childPath,
							Repository: nodeDef.URL,
							IsCloned:   true,
						})
					}
				} else if nodeDef.File != "" {
					// It's another config node - recurse into it
					childConfigFile := nodeDef.File
					if !filepath.IsAbs(childConfigFile) && !strings.HasPrefix(childConfigFile, "http") {
						currentConfigDir := filepath.Dir(configFilePath)
						childConfigFile = filepath.Join(currentConfigDir, nodeDef.File)
					}
					
					childNode := interfaces.NodeInfo{
						Name:       nodeDef.Name,
						Path:       childPath,
						ConfigFile: childConfigFile,
						IsConfig:   true,
					}
					repos = append(repos, m.collectClonedRepos(childNode)...)
				}
			}
		}
	} else if len(node.Children) == 0 && node.IsCloned {
		// Terminal node that is cloned
		repos = append(repos, node)
	}
	
	// Recurse into children (for non-config nodes)
	for _, child := range node.Children {
		repos = append(repos, m.collectClonedRepos(child)...)
	}
	
	return repos
}

func (m *Manager) pullRecursiveWithOptions(node interfaces.NodeInfo, force bool, includeLazy bool) error {
	// First, clone lazy repositories if includeLazy is true
	if includeLazy && node.IsLazy && !node.IsCloned && node.Repository != "" {
		fullPath := m.computeFilesystemPath(node.Path)
		m.uiProvider.Info(fmt.Sprintf("📥 Cloning lazy repository: %s", node.Name))
		m.uiProvider.Info(fmt.Sprintf("   URL: %s", node.Repository))
		
		// Clone the repository
		cloneOptions := interfaces.CloneOptions{
			SSHPreference: m.getSSHPreference(),
		}
		if err := m.gitProvider.Clone(node.Repository, fullPath, cloneOptions); err != nil {
			m.uiProvider.Error(fmt.Sprintf("   ❌ Clone failed: %v", err))
			return fmt.Errorf("cloning %s: %w", node.Name, err)
		}
		
		// Update node status in tree
		node.IsCloned = true
		node.IsLazy = false
		if err := m.treeProvider.UpdateNode(node.Path, node); err != nil {
			m.uiProvider.Warning(fmt.Sprintf("   ⚠️  Failed to update node status: %v", err))
			// Don't fail the operation, just warn
		}
		
		// Save configuration to persist the change
		if err := m.saveConfig(); err != nil {
			m.uiProvider.Warning(fmt.Sprintf("   ⚠️  Failed to save config: %v", err))
			// Don't fail the operation, just warn
		}
		
		m.uiProvider.Success(fmt.Sprintf("   ✅ Cloned successfully: %s", node.Name))
	}
	
	// Handle config nodes - expand them to find repositories
	if node.IsConfig && node.ConfigFile != "" {
		// Process config node similar to pullRecursive
		configFilePath := node.ConfigFile
		if !filepath.IsAbs(configFilePath) && !strings.HasPrefix(configFilePath, "http") {
			configFilePath = m.resolveConfigPath(node.ConfigFile, node.Path)
		}
		
		if cfg, err := config.LoadTree(configFilePath); err == nil {
			for _, nodeDef := range cfg.Nodes {
				childPath := node.Path + "/" + nodeDef.Name
				if node.Path == "/" {
					childPath = "/" + nodeDef.Name
				}
				
				if nodeDef.URL != "" {
					childNode := interfaces.NodeInfo{
						Name:       nodeDef.Name,
						Path:       childPath,
						Repository: nodeDef.URL,
						IsLazy:     nodeDef.IsLazy(),
						IsCloned:   false,
					}
					// Check if cloned
					childFsPath := m.computeFilesystemPath(childPath)
					if _, err := os.Stat(filepath.Join(childFsPath, ".git")); err == nil {
						childNode.IsCloned = true
					}
					// Recursively process this child
					if err := m.pullRecursiveWithOptions(childNode, force, includeLazy); err != nil {
						return err
					}
				} else if nodeDef.File != "" {
					// Nested config node
					childConfigFile := nodeDef.File
					if !filepath.IsAbs(childConfigFile) && !strings.HasPrefix(childConfigFile, "http") {
						currentConfigDir := filepath.Dir(configFilePath)
						childConfigFile = filepath.Join(currentConfigDir, nodeDef.File)
					}
					
					childNode := interfaces.NodeInfo{
						Name:       nodeDef.Name,
						Path:       childPath,
						ConfigFile: childConfigFile,
						IsConfig:   true,
					}
					if err := m.pullRecursiveWithOptions(childNode, force, includeLazy); err != nil {
						return err
					}
				}
			}
		}
	} else if len(node.Children) == 0 && node.IsCloned && !node.IsLazy {
		// Pull only if it's a cloned repository (not lazy or already cloned)
		// Only pull terminal nodes (repositories without children)
		fullPath := m.computeFilesystemPath(node.Path)
		m.uiProvider.Info(fmt.Sprintf("📦 Pulling: %s", node.Name))
		
		pullOpts := interfaces.PullOptions{Force: force}
		if err := m.gitProvider.Pull(fullPath, pullOpts); err != nil {
			m.uiProvider.Error(fmt.Sprintf("   ❌ Failed at %s: %v", node.Path, err))
			// Don't stop on error, continue with other repos
		} else {
			m.uiProvider.Success(fmt.Sprintf("   ✅ Success: %s", node.Name))
		}
	} else if len(node.Children) == 0 && node.IsLazy && !node.IsCloned {
		// Skip lazy repositories that haven't been cloned
		m.logProvider.Debug(fmt.Sprintf("Skipping lazy repository: %s", node.Name))
	}
	
	// Recurse into children (for non-config nodes)
	for _, child := range node.Children {
		if err := m.pullRecursiveWithOptions(child, force, includeLazy); err != nil {
			return err
		}
	}
	
	return nil
}

func (m *Manager) pullRecursive(node interfaces.NodeInfo, force bool) error {
	// Handle config nodes - expand them to find repositories
	if node.IsConfig && node.ConfigFile != "" {
		// Load the config and process its children
		configFilePath := node.ConfigFile
		if !filepath.IsAbs(configFilePath) && !strings.HasPrefix(configFilePath, "http") {
			configFilePath = m.resolveConfigPath(node.ConfigFile, node.Path)
		}
		
		if cfg, err := config.LoadTree(configFilePath); err == nil {
			// Process each child node from the config
			for _, nodeDef := range cfg.Nodes {
				childPath := node.Path + "/" + nodeDef.Name
				if node.Path == "/" {
					childPath = "/" + nodeDef.Name
				}
				
				if nodeDef.URL != "" {
					// It's a repository - check if it's cloned and pull it
					childFsPath := m.computeFilesystemPath(childPath)
					if _, err := os.Stat(filepath.Join(childFsPath, ".git")); err == nil {
						// Repository is cloned, pull it
						m.uiProvider.Info(fmt.Sprintf("📦 Pulling: %s", nodeDef.Name))
						pullOpts := interfaces.PullOptions{Force: force}
						if err := m.gitProvider.Pull(childFsPath, pullOpts); err != nil {
							m.uiProvider.Error(fmt.Sprintf("   ❌ Failed at %s: %v", childPath, err))
						} else {
							m.uiProvider.Success(fmt.Sprintf("   ✅ Success: %s", nodeDef.Name))
						}
					}
				} else if nodeDef.File != "" {
					// It's another config node - recurse into it
					childConfigFile := nodeDef.File
					if !filepath.IsAbs(childConfigFile) && !strings.HasPrefix(childConfigFile, "http") {
						currentConfigDir := filepath.Dir(configFilePath)
						childConfigFile = filepath.Join(currentConfigDir, nodeDef.File)
					}
					
					childNode := interfaces.NodeInfo{
						Name:       nodeDef.Name,
						Path:       childPath,
						ConfigFile: childConfigFile,
						IsConfig:   true,
					}
					if err := m.pullRecursive(childNode, force); err != nil {
						return err
					}
				}
			}
		}
	} else if len(node.Children) == 0 && node.IsCloned {
		// Terminal node that is cloned - pull it
		fullPath := m.computeFilesystemPath(node.Path)
		m.uiProvider.Info(fmt.Sprintf("📦 Pulling: %s", node.Name))
		
		pullOpts := interfaces.PullOptions{Force: force}
		if err := m.gitProvider.Pull(fullPath, pullOpts); err != nil {
			m.uiProvider.Error(fmt.Sprintf("   ❌ Failed at %s: %v", node.Path, err))
			// Don't stop on error, continue with other repos
		} else {
			m.uiProvider.Success(fmt.Sprintf("   ✅ Success: %s", node.Name))
		}
	}
	
	// Recurse into children (for non-config nodes)
	for _, child := range node.Children {
		if err := m.pullRecursive(child, force); err != nil {
			return err
		}
	}
	
	return nil
}

// PushNode pushes changes for a node
func (m *Manager) PushNode(path string, recursive bool) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	targetPath := path
	if targetPath == "" {
		// Use pwd-based resolution
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
		
		workspaceRoot := m.workspace
		reposDir := filepath.Join(workspaceRoot, m.getReposDir())
		
		if strings.HasPrefix(pwd, reposDir) {
			relPath, err := filepath.Rel(reposDir, pwd)
			if err != nil {
				return fmt.Errorf("getting relative path: %w", err)
			}
			if relPath == "." {
				targetPath = "/"
			} else {
				targetPath = "/" + strings.ReplaceAll(filepath.ToSlash(relPath), "\\", "/")
			}
		} else if pwd == workspaceRoot {
			targetPath = "/"
		} else {
			targetPath = "/"
		}
	}
	
	node, err := m.treeProvider.GetNode(targetPath)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}
	
	if recursive {
		return m.pushRecursive(node)
	}
	
	fullPath := m.computeFilesystemPath(node.Path)
	m.logProvider.Info("Pushing changes")
	m.logProvider.Info(fmt.Sprintf("  Tree path: %s", node.Path))
	m.logProvider.Info(fmt.Sprintf("  Directory: %s", fullPath))
	return m.gitProvider.Push(fullPath, interfaces.PushOptions{})
}

func (m *Manager) pushRecursive(node interfaces.NodeInfo) error {
	if node.IsCloned {
		m.logProvider.Info(fmt.Sprintf("Pushing changes at %s", node.Path))
		if err := m.gitProvider.Push(m.computeFilesystemPath(node.Path), interfaces.PushOptions{}); err != nil {
			m.logProvider.Error(fmt.Sprintf("Push failed at %s: %v", node.Path, err))
		}
	}
	
	for _, child := range node.Children {
		if err := m.pushRecursive(child); err != nil {
			return err
		}
	}
	
	return nil
}

// CommitNode commits changes for a node
func (m *Manager) CommitNode(path string, message string, recursive bool) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	targetPath := path
	if targetPath == "" {
		// Use pwd-based resolution
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
		
		workspaceRoot := m.workspace
		reposDir := filepath.Join(workspaceRoot, m.getReposDir())
		
		if strings.HasPrefix(pwd, reposDir) {
			relPath, err := filepath.Rel(reposDir, pwd)
			if err != nil {
				return fmt.Errorf("getting relative path: %w", err)
			}
			if relPath == "." {
				targetPath = "/"
			} else {
				targetPath = "/" + strings.ReplaceAll(filepath.ToSlash(relPath), "\\", "/")
			}
		} else if pwd == workspaceRoot {
			targetPath = "/"
		} else {
			targetPath = "/"
		}
	}
	
	node, err := m.treeProvider.GetNode(targetPath)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}
	
	fullPath := m.computeFilesystemPath(node.Path)
	m.logProvider.Info(fmt.Sprintf("Committing changes: %s", message))
	m.logProvider.Info(fmt.Sprintf("  Tree path: %s", node.Path))
	m.logProvider.Info(fmt.Sprintf("  Directory: %s", fullPath))
	return m.gitProvider.Commit(fullPath, message, interfaces.CommitOptions{})
}





// ensureGitignoreEntry adds an entry to .gitignore if not already present
func (m *Manager) ensureGitignoreEntry(workDir string, entry string) error {
	// Check if this is a git repository
	gitDir := filepath.Join(workDir, ".git")
	if !m.fsProvider.Exists(gitDir) {
		// Not a git repository, skip
		return nil
	}
	
	gitignorePath := filepath.Join(workDir, ".gitignore")
	
	// Read existing .gitignore content
	var content []byte
	var err error
	if m.fsProvider.Exists(gitignorePath) {
		content, err = m.fsProvider.ReadFile(gitignorePath)
		if err != nil {
			return fmt.Errorf("reading .gitignore: %w", err)
		}
	}
	
	// Check if entry already exists
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == entry || trimmed == strings.TrimSuffix(entry, "/") {
			// Entry already exists
			return nil
		}
	}
	
	// Add the entry
	var newContent string
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		newContent = string(content) + "\n"
	} else {
		newContent = string(content)
	}
	
	// Add comment and entry
	newContent += "\n# MUNO agent context (auto-added)\n" + entry + "\n"
	
	// Write updated .gitignore
	if err := m.fsProvider.WriteFile(gitignorePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("writing .gitignore: %w", err)
	}
	
	m.logProvider.Debug(fmt.Sprintf("Added '%s' to .gitignore", entry))
	return nil
}

// generateTreeContext generates a tree representation for the context
func (m *Manager) generateTreeContext(currentNode *interfaces.NodeInfo) string {
	var output strings.Builder
	
	// Get root node to show full tree
	root, err := m.treeProvider.GetNode("")
	if err != nil {
		// If we can't get root, just show current node
		root = *currentNode
	}
	
	// Generate tree with current node highlighted
	m.writeTreeNode(&output, &root, currentNode.Path, "", true)
	
	return output.String()
}

// writeTreeNode writes a tree node with optional highlighting
func (m *Manager) writeTreeNode(output *strings.Builder, node *interfaces.NodeInfo, highlightPath string, prefix string, isLast bool) {
	// Determine the connector
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if prefix == "" {
		connector = ""
	}
	
	// Write the node with potential highlighting
	nodeName := node.Name
	if node.Path == highlightPath {
		nodeName = fmt.Sprintf("%s  <-- YOU ARE HERE", nodeName)
	}
	
	// Add status and type indicators
	statusIndicator := ""
	
	// First, add node type information from config if available
	if m.config != nil {
		for _, nodeDef := range m.config.Nodes {
			if nodeDef.Name == node.Name {
				if nodeDef.URL != "" {
					statusIndicator += fmt.Sprintf(" [git: %s]", nodeDef.URL)
				} else if nodeDef.File != "" {
					statusIndicator += fmt.Sprintf(" [config: %s]", nodeDef.File)
				}
				break
			}
		}
	}
	
	// Also show repository URL from NodeInfo if available
	if node.Repository != "" && !strings.Contains(statusIndicator, node.Repository) {
		statusIndicator += fmt.Sprintf(" [repo: %s]", node.Repository)
	}
	
	// Add clone status
	if node.IsLazy && !node.IsCloned {
		statusIndicator += " [lazy]"
	} else if !node.IsCloned {
		statusIndicator += " [not cloned]"
	}
	
	output.WriteString(fmt.Sprintf("%s%s%s%s\n", prefix, connector, nodeName, statusIndicator))
	
	// Process children
	if len(node.Children) > 0 {
		childPrefix := prefix
		if prefix == "" {
			childPrefix = ""
		} else if isLast {
			childPrefix = prefix + "    "
		} else {
			childPrefix = prefix + "│   "
		}
		
		for i, child := range node.Children {
			isLastChild := i == len(node.Children)-1
			// Create a copy of child to pass as pointer
			childCopy := child
			m.writeTreeNode(output, &childCopy, highlightPath, childPrefix, isLastChild)
		}
	}
}


