package manager

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"time"
	
	"github.com/mattn/go-isatty"
	"github.com/taokim/muno/internal/adapters"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/plugin"
	"github.com/taokim/muno/internal/tree"
)

// Embed the AI Agent Context documentation as a file system
//
//go:embed docs/AI_AGENT_CONTEXT.md
var agentContextFS embed.FS

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
			Recursive: options.Recursive,
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

// ResolvePath resolves a virtual tree path to its physical filesystem location
// If ensure is true, it will clone lazy repositories if needed
// ResolvePath resolves a virtual tree path to its physical filesystem location
// If ensure is true, it will clone lazy repositories if needed
func (m *Manager) ResolvePath(target string, ensure bool) (string, error) {
	if !m.initialized {
		return "", fmt.Errorf("manager not initialized")
	}
	
	// Get current directory to resolve relative paths
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %w", err)
	}
	
	// Determine current position in tree based on filesystem location
	currentTreePath := "/"
	reposDir := filepath.Join(m.workspace, m.config.GetReposDir())
	
	if strings.HasPrefix(cwd, reposDir) {
		// Extract relative path from repos directory
		relPath, err := filepath.Rel(reposDir, cwd)
		if err == nil && relPath != "." {
			currentTreePath = "/" + filepath.ToSlash(relPath)
		}
	}
	
	// Resolve target path
	resolvedPath := target
	if target == "." {
		resolvedPath = currentTreePath
	} else if target == ".." {
		parts := strings.Split(strings.TrimPrefix(currentTreePath, "/"), "/")
		if len(parts) > 1 {
			resolvedPath = "/" + strings.Join(parts[:len(parts)-1], "/")
		} else {
			resolvedPath = "/"
		}
	} else if target == "/" || target == "~" {
		resolvedPath = "/"
	} else if !strings.HasPrefix(target, "/") {
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
		node, err := m.treeProvider.GetNode(resolvedPath)
		if err == nil && node.IsLazy && !node.IsCloned && node.Repository != "" {
			// Clone the repository
			physPath := m.computeFilesystemPath(resolvedPath)
			if _, err := os.Stat(physPath); os.IsNotExist(err) {
				// Clone options
			opts := interfaces.CloneOptions{
				Recursive: false,
				Quiet:     false,
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
	
	// Compute filesystem path
	physicalPath := m.computeFilesystemPath(resolvedPath)
	
	return physicalPath, nil
}

// GetTreePath converts a physical filesystem path to its position in the tree
// GetTreePath converts a physical filesystem path to its position in the tree
func (m *Manager) GetTreePath(physicalPath string) (string, error) {
	if !m.initialized {
		return "", fmt.Errorf("manager not initialized")
	}
	
	reposDir := filepath.Join(m.workspace, m.config.GetReposDir())
	
	// Check if the path is within the workspace
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
	
	// Convert to tree path format
	treePath := "/" + filepath.ToSlash(relPath)
	
	return treePath, nil
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

// DefaultProcessProvider is a simple implementation of ProcessProvider
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

// getReposDir returns the configured repos directory name
func (m *Manager) getReposDir() string {
	if m.config != nil {
		return m.config.GetReposDir()
	}
	return config.GetDefaultReposDir()
}

// computeFilesystemPath computes the actual filesystem path from a logical tree path
// This replicates the logic from tree.Manager.ComputeFilesystemPath
func (m *Manager) computeFilesystemPath(logicalPath string) string {
	reposDir := m.getReposDir()
	
	// For root, always use repos directory
	if logicalPath == "/" || logicalPath == "" {
		return filepath.Join(m.workspace, reposDir)
	}
	
	// Split the path into parts
	parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")
	
	// For top-level repository
	if len(parts) == 1 {
		return filepath.Join(m.workspace, reposDir, parts[0])
	}
	
	// For nested paths, we need to check if parent is a git repo
	// If parent is a git repo, children go inside it
	// Otherwise, they go in parallel under repos dir
	
	// Check if the first part is a cloned repository
	parentPath := filepath.Join(m.workspace, reposDir, parts[0])
	gitPath := filepath.Join(parentPath, ".git")
	
	// If parent is a git repository, nest children inside it
	if m.fsProvider.Exists(gitPath) {
		// Build path inside the parent repository
		fsPath := parentPath
		for i := 1; i < len(parts); i++ {
			fsPath = filepath.Join(fsPath, parts[i])
		}
		return fsPath
	}
	
	// Otherwise, use flat structure under repos dir
	fsPath := filepath.Join(m.workspace, reposDir)
	for _, part := range parts {
		fsPath = filepath.Join(fsPath, part)
	}
	
	return fsPath
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
	isTerminal := len(node.Children) == 0
	
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
		// Non-terminal nodes (parent nodes with children)
		// Check if it's a config reference node or git parent node
		if m.config != nil {
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
func (m *Manager) CloneRepos(path string, recursive bool) error {
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
	
	// Clone lazy repos
	var toClone []interfaces.NodeInfo
	if recursive {
		toClone = collectLazyNodes(current)
	} else {
		for _, child := range current.Children {
			if child.IsLazy && !child.IsCloned {
				toClone = append(toClone, child)
			}
		}
	}
	
	for _, node := range toClone {
		m.logProvider.Info(fmt.Sprintf("Cloning repository %s from %s", node.Name, node.Repository))
		if err := m.gitProvider.Clone(node.Repository, m.computeFilesystemPath(node.Path), interfaces.CloneOptions{}); err != nil {
			return fmt.Errorf("cloning %s: %w", node.Name, err)
		}
		
		// Update node status
		node.IsCloned = true
		node.IsLazy = false
		if err := m.treeProvider.UpdateNode(node.Path, node); err != nil {
			return fmt.Errorf("updating node: %w", err)
		}
	}
	
	return m.saveConfig()
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
	
	// Only add terminal nodes that are cloned
	if len(node.Children) == 0 && node.IsCloned {
		repos = append(repos, node)
	}
	
	// Recurse into children
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
		if err := m.gitProvider.Clone(node.Repository, fullPath, interfaces.CloneOptions{}); err != nil {
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
	
	// Then pull if it's a cloned repository (terminal node)
	if len(node.Children) == 0 && node.IsCloned {
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
	
	// Recurse into children
	for _, child := range node.Children {
		if err := m.pullRecursiveWithOptions(child, force, includeLazy); err != nil {
			return err
		}
	}
	
	return nil
}

func (m *Manager) pullRecursive(node interfaces.NodeInfo, force bool) error {
	// Only pull terminal nodes (actual repositories)
	if len(node.Children) == 0 && node.IsCloned {
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
	
	// Recurse into children
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

// StartClaude starts a Claude session
// Deprecated: Use StartAgent("claude", path, nil) instead
func (m *Manager) StartClaude(path string) error {
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
	
	// Start Claude session using process provider
	// For root level, use workspace root instead of repos directory
	var fullPath string
	if node.Path == "/" || node.Path == "" {
		fullPath = m.workspace
	} else {
		fullPath = m.computeFilesystemPath(node.Path)
	}
	
	m.logProvider.Info("Starting Claude session")
	m.logProvider.Info(fmt.Sprintf("  Tree path: %s", node.Path))
	m.logProvider.Info(fmt.Sprintf("  Directory: %s", fullPath))
	
	result, err := m.processProvider.ExecuteShell(context.Background(), fmt.Sprintf("cd %s && claude", fullPath), interfaces.ProcessOptions{})
	if err != nil {
		return fmt.Errorf("starting Claude: %w", err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("Claude exited with code %d: %s", result.ExitCode, result.Stderr)
	}
	
	return nil
}

// createAgentContext creates a context file with MUNO documentation and workspace info
func (m *Manager) createAgentContext(node *interfaces.NodeInfo, workDir string) error {
	// Create .muno directory if it doesn't exist
	contextDir := filepath.Join(workDir, ".muno")
	if err := m.fsProvider.MkdirAll(contextDir, 0755); err != nil {
		return fmt.Errorf("creating context directory: %w", err)
	}
	
	// Add .muno to .gitignore if this is a git repository
	if err := m.ensureGitignoreEntry(workDir, ".muno/"); err != nil {
		// Log but don't fail - .gitignore update is optional
		m.logProvider.Debug(fmt.Sprintf("Could not update .gitignore: %v", err))
	}
	
	// Generate tree structure
	treeOutput := m.generateTreeContext(node)
	
	// Create context content
	contextContent := fmt.Sprintf(`# MUNO Agent Context

## Current Workspace Information

**Current Position**: %s
**Working Directory**: %s

## Workspace Tree Structure

%s

### Node Types in the Tree:
- **[git: URL]**: Git repository node - clones and manages a git repository
- **[config: PATH]**: Config reference node - delegates subtree management to another muno.yaml file
- **[lazy]**: Repository will be cloned on-demand when first accessed
- **[not cloned]**: Repository exists in config but hasn't been cloned yet

## Available MUNO Commands

- **Navigation**: muno use <path> - Navigate to a node in the tree
- **Status**: muno current - Show current position
- **Tree**: muno tree - Display full tree structure
- **Git**: muno pull/push/status - Git operations at current node
- **Repos**: muno add <url> - Add new repository

## Documentation Links

### For AI Agents - READ THESE:
- **AI Agent Guide**: https://raw.githubusercontent.com/taokim/muno/main/docs/AI_AGENT_GUIDE.md
- **Web Documentation**: https://taokim.github.io/muno/
- **Examples**: https://github.com/taokim/muno/tree/main/examples

## Configuration Documentation

For complete configuration schema, defaults, and examples, see:
- **Configuration Schema**: https://raw.githubusercontent.com/taokim/muno/main/docs/MUNO_CONFIG_SCHEMA.md

## Tips for AI Agents

- Use 'muno tree' to understand the full workspace structure
- Navigate with 'muno use <path>' to move between repositories
- Check status with 'muno status --recursive' for all repos
- The current directory is already set to the target repository

---
*This context was generated by MUNO to help AI agents understand the workspace structure.*
`, node.Path, workDir, treeOutput)
	
	// Read and append the embedded AI Agent Context documentation
	agentContextData, err := agentContextFS.ReadFile("docs/AI_AGENT_CONTEXT.md")
	if err != nil {
		m.logProvider.Warn(fmt.Sprintf("Could not read embedded agent context: %v", err))
		// Continue without the extended context
	} else {
		contextContent += "\n\n---\n\n" + string(agentContextData)
	}
	
	// Write context file to .muno directory
	contextFile := filepath.Join(contextDir, "agent-context.md")
	if err := m.fsProvider.WriteFile(contextFile, []byte(contextContent), 0644); err != nil {
		return fmt.Errorf("writing context file: %w", err)
	}
	
	m.logProvider.Info(fmt.Sprintf("  Context created at: %s", contextFile))
	return nil
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

// StartAgent starts an AI agent session (claude, gemini, cursor, etc.)
func (m *Manager) StartAgent(agentName string, path string, agentArgs []string, withMunoContext bool) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	// Default to claude if no agent specified
	if agentName == "" {
		agentName = "claude"
	}
	
	targetPath := path
	if targetPath == "" {
		// Use pwd-based resolution
		pwd, err := os.Getwd()
		if err != nil {
			// If we can't get pwd, use root
			targetPath = "/"
		} else {
			workspaceRoot := m.workspace
			reposDir := filepath.Join(workspaceRoot, m.getReposDir())
			
			if strings.HasPrefix(pwd, reposDir) {
				relPath, err := filepath.Rel(reposDir, pwd)
				if err != nil {
					targetPath = "/"
				} else if relPath == "." {
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
	}
	
	// Special case for root - create a simple node
	var node interfaces.NodeInfo
	if targetPath == "/" || targetPath == "" {
		// For root, just use workspace directly
		node = interfaces.NodeInfo{
			Path: "/",
			Name: "root",
		}
	} else {
		n, err := m.treeProvider.GetNode(targetPath)
		if err != nil {
			return fmt.Errorf("getting node: %w", err)
		}
		node = n
	}
	
	// Start agent session using process provider
	// For agents at root level, use workspace root instead of repos directory
	var fullPath string
	if node.Path == "/" || node.Path == "" {
		fullPath = m.workspace
	} else {
		fullPath = m.computeFilesystemPath(node.Path)
	}
	
	// Check if the agent command exists (but don't fail if it's an alias)
	agentPath, err := exec.LookPath(agentName)
	if err != nil {
		// It might be a shell alias, so we'll try to run it anyway
		m.logProvider.Debug(fmt.Sprintf("%s not found in PATH, might be a shell alias", agentName))
		agentPath = agentName // Use the name directly, shell will resolve it
	}
	
	m.logProvider.Info(fmt.Sprintf("Starting %s session", agentName))
	if err == nil {
		m.logProvider.Info(fmt.Sprintf("  Agent path: %s", agentPath))
	} else {
		m.logProvider.Info(fmt.Sprintf("  Agent: %s (possibly shell alias)", agentName))
	}
	m.logProvider.Info(fmt.Sprintf("  Tree path: %s", node.Path))
	m.logProvider.Info(fmt.Sprintf("  Directory: %s", fullPath))
	
	// Build the command arguments
	var contextFile string
	
	// If requested, inject MUNO context for the agent
	if withMunoContext {
		m.logProvider.Info("  Creating MUNO context for agent...")
		// Always create context in workspace root for consistency
		contextPath := m.workspace
		if err := m.createAgentContext(&node, contextPath); err != nil {
			m.logProvider.Error(fmt.Sprintf("Failed to create agent context: %v", err))
			// Continue anyway, context is optional
		} else {
			// For Claude, append the context as a system prompt
			if agentName == "claude" {
				contextFile := filepath.Join(contextPath, ".muno", "agent-context.md")
				// Read the context file and pass it via --append-system-prompt
				if _, err := m.fsProvider.ReadFile(contextFile); err == nil {
					// We'll use this contextFile path later when building the command
					m.logProvider.Info("  Appending MUNO context to Claude system prompt")
				}
			}
		}
	}
	
	// Build command arguments
	var allArgs []string
	if withMunoContext && agentName == "claude" && contextFile != "" {
		allArgs = append(allArgs, "--append-system-prompt", contextFile)
	}
	allArgs = append(allArgs, agentArgs...)
	
	// For interactive agents, we need to run them directly with stdin/stdout connected
	m.logProvider.Debug(fmt.Sprintf("Starting interactive agent: %s", agentName))
	
	// Check if we're in a TTY environment
	isInteractive := isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
	
	if !isInteractive {
		// Not in a TTY - agent command might fail or behave unexpectedly
		m.logProvider.Warn("Not running in a terminal. Some agents may not work properly.")
		m.logProvider.Warn("Consider using --print flag for non-interactive mode if supported by the agent.")
	}
	
	// Create the command
	// We need to use the user's shell to support aliases
	// Detect the user's shell
	userShell := os.Getenv("SHELL")
	if userShell == "" {
		userShell = "/bin/sh" // fallback to sh
	}
	
	// Build the command string for the shell
	// Use -i for interactive shell to load aliases and -c to execute command
	var shellCmd string
	if len(allArgs) > 0 {
		// Properly escape arguments
		quotedArgs := make([]string, len(allArgs))
		for i, arg := range allArgs {
			// Simple escaping - wrap in single quotes and escape any single quotes
			quotedArgs[i] = "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
		}
		shellCmd = fmt.Sprintf("cd %s && %s %s", fullPath, agentName, strings.Join(quotedArgs, " "))
	} else {
		shellCmd = fmt.Sprintf("cd %s && %s", fullPath, agentName)
	}
	
	// Use -i for interactive shell to load RC files (aliases)
	// But we also need -c to run a command
	// For zsh/bash, we can use: shell -i -c "command"
	cmd := exec.Command(userShell, "-i", "-c", shellCmd)
	cmd.Dir = fullPath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Set up the environment
	cmd.Env = os.Environ()
	
	// Run the interactive command
	m.logProvider.Debug(fmt.Sprintf("Executing via %s: %s", userShell, shellCmd))
	if err := cmd.Run(); err != nil {
		// Check if the error is due to various issues
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 127 {
				// Command not found - provide helpful message
				installMsg := ""
				switch agentName {
				case "claude":
					installMsg = "Install with: npm install -g @anthropic-ai/cli"
				case "gemini":
					installMsg = "Install with: npm install -g @google/gemini-cli"
				case "cursor":
					installMsg = "Download from: https://cursor.sh"
				case "windsurf":
					installMsg = "Download from: https://www.codeium.com/windsurf"
				case "aider":
					installMsg = "Install with: pip install aider-chat"
				default:
					installMsg = fmt.Sprintf("Make sure %s is installed or aliased in your shell", agentName)
				}
				return fmt.Errorf("%s not found. %s", agentName, installMsg)
			}
			if exitErr.ExitCode() == 126 {
				return fmt.Errorf("%s found but not executable. Check file permissions", agentName)
			}
			// For other exit codes, just pass through the error
			return fmt.Errorf("%s exited with error: %w", agentName, err)
		}
		return fmt.Errorf("failed to run %s: %w", agentName, err)
	}
	
	return nil
}
