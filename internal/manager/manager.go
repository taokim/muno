package manager

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/plugin"
	"github.com/taokim/muno/internal/tree"
)

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
	
	return &Manager{
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
	}, nil
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

// Use navigates to a specific node in the tree
func (m *Manager) Use(ctx context.Context, path string) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	m.logProvider.Info("Navigating to node", 
		interfaces.Field{Key: "path", Value: path})
	
	// Navigate in tree
	if err := m.treeProvider.Navigate(path); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}
	
	// Get node info
	node, err := m.treeProvider.GetCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current node: %w", err)
	}
	
	// Clone if lazy
	if node.IsLazy && !node.IsCloned {
		m.uiProvider.Info(fmt.Sprintf("Cloning repository: %s", node.Repository))
		
		repoPath := m.computeFilesystemPath(node.Path)
		if err := m.gitProvider.Clone(node.Repository, repoPath, interfaces.CloneOptions{
			Recursive: true,
		}); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
		
		// Update node state
		node.IsCloned = true
		if err := m.treeProvider.UpdateNode(node.Path, node); err != nil {
			m.logProvider.Warn("Failed to update node state", 
				interfaces.Field{Key: "error", Value: err})
		}
	}
	
	m.uiProvider.Success(fmt.Sprintf("Now at: %s", path))
	m.metricsProvider.Counter("manager.navigate", 1, "path:"+path)
	
	return nil
}

// Add adds a new repository to the tree
func (m *Manager) Add(ctx context.Context, repoURL string, options AddOptions) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	m.logProvider.Info("Adding repository", 
		interfaces.Field{Key: "url", Value: repoURL},
		interfaces.Field{Key: "lazy", Value: options.Lazy})
	
	// Get current node
	current, err := m.treeProvider.GetCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current node: %w", err)
	}
	
	// Extract repo name from URL
	repoName := extractRepoName(repoURL)
	
	// Create new node
	newNode := interfaces.NodeInfo{
		Name:       repoName,
		Repository: repoURL,
		IsLazy:     options.Lazy,
		IsCloned:   false,
	}
	
	// Add to tree
	if err := m.treeProvider.AddNode(current.Path, newNode); err != nil {
		return fmt.Errorf("failed to add node: %w", err)
	}
	
	// Clone immediately if not lazy
	if !options.Lazy {
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
	
	m.uiProvider.Success(fmt.Sprintf("Added repository: %s", repoName))
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
	
	// Get current node
	current, err := m.treeProvider.GetCurrent()
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
	
	// Save configuration
	if err := m.saveConfig(); err != nil {
		m.logProvider.Warn("Failed to save config", 
			interfaces.Field{Key: "error", Value: err})
	}
	
	m.uiProvider.Success(fmt.Sprintf("Removed repository: %s", name))
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
		// Navigate to a path
		return m.Use(ctx, action.Path)
		
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

// AddOptions for adding repositories
type AddOptions struct {
	Lazy      bool
	Recursive bool
	Branch    string
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
	fmt.Printf("[INFO] %s\n", msg)
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
func (t *NoOpTimer) C() <-chan time.Time { return nil }
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
	if m.config != nil && m.config.Workspace.ReposDir != "" {
		return m.config.Workspace.ReposDir
	}
	return config.GetDefaultReposDir()
}

// computeFilesystemPath computes the actual filesystem path from a logical tree path
// This replicates the logic from tree.Manager.ComputeFilesystemPath
func (m *Manager) computeFilesystemPath(logicalPath string) string {
	reposDir := m.getReposDir()
	
	if logicalPath == "/" {
		return filepath.Join(m.workspace, reposDir)
	}
	
	// Split path: /level1/level2/level3 -> [level1, level2, level3]
	parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")
	
	// Build filesystem path with repos subdirectories
	// workspace/[reposDir]/level1/[reposDir]/level2/[reposDir]/level3
	fsPath := filepath.Join(m.workspace, reposDir)
	for i, part := range parts {
		fsPath = filepath.Join(fsPath, part)
		// Add repos dir before next level (except last)
		if i < len(parts)-1 {
			fsPath = filepath.Join(fsPath, reposDir)
		}
	}
	
	return fsPath
}

// Helper function to display tree recursively
func (m *Manager) displayTreeRecursive(node interfaces.NodeInfo, indent int) {
	prefix := strings.Repeat("  ", indent)
	status := ""
	if node.IsLazy {
		status = " (lazy)"
	} else if node.HasChanges {
		status = " (modified)"
	}
	m.uiProvider.Info(fmt.Sprintf("%s%s%s", prefix, node.Name, status))
	
	for _, child := range node.Children {
		m.displayTreeRecursive(child, indent+1)
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
		m.uiProvider.Info(fmt.Sprintf("Tree at %s:", tree.Path))
		m.displayTreeRecursive(tree, 0)
		return nil
	}
	
	// Just list immediate children
	current, err := m.treeProvider.GetCurrent()
	if err != nil {
		return fmt.Errorf("getting current node: %w", err)
	}
	
	if len(current.Children) == 0 {
		m.uiProvider.Info("No children")
		return nil
	}
	
	for _, child := range current.Children {
		m.uiProvider.Info(fmt.Sprintf("  %s", child.Name))
	}
	
	return nil
}

// ShowCurrent shows the current position in the tree
func (m *Manager) ShowCurrent() error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	current, err := m.treeProvider.GetCurrent()
	if err != nil {
		return fmt.Errorf("getting current node: %w", err)
	}
	
	m.uiProvider.Info(fmt.Sprintf("Current: %s", current.Path))
	return nil
}

// ShowTreeAtPath shows the tree at a specific path
func (m *Manager) ShowTreeAtPath(path string, depth int) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	// Default to root if path is empty
	if path == "" {
		path = "/"
	}
	
	node, err := m.treeProvider.GetNode(path)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}
	
	m.uiProvider.Info(fmt.Sprintf("Tree at %s:", node.Path))
	m.displayTreeRecursive(node, 0)
	return nil
}

// UseNodeWithClone navigates to a node and clones if needed
func (m *Manager) UseNodeWithClone(path string, autoClone bool) error {
	ctx := context.Background()
	return m.Use(ctx, path)
}

// AddRepoSimple adds a repository
func (m *Manager) AddRepoSimple(repoURL string, name string, lazy bool) error {
	ctx := context.Background()
	return m.Add(ctx, repoURL, AddOptions{Lazy: lazy})
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
	
	// Get current node
	current, err := m.treeProvider.GetCurrent()
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
		current, err := m.treeProvider.GetCurrent()
		if err != nil {
			return fmt.Errorf("getting current node: %w", err)
		}
		targetPath = current.Path
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
	
	m.uiProvider.Info(fmt.Sprintf("%s: branch=%s clean=%v", node.Name, status.Branch, status.IsClean))
	return nil
}

func (m *Manager) showStatusRecursive(node interfaces.NodeInfo) error {
	status, err := m.gitProvider.Status(m.computeFilesystemPath(node.Path))
	if err != nil {
		m.uiProvider.Info(fmt.Sprintf("%s: error - %v", node.Name, err))
	} else {
		m.uiProvider.Info(fmt.Sprintf("%s: branch=%s clean=%v", node.Name, status.Branch, status.IsClean))
	}
	
	for _, child := range node.Children {
		if err := m.showStatusRecursive(child); err != nil {
			return err
		}
	}
	
	return nil
}

// PullNode pulls changes for a node
func (m *Manager) PullNode(path string, recursive bool) error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	targetPath := path
	if targetPath == "" {
		current, err := m.treeProvider.GetCurrent()
		if err != nil {
			return fmt.Errorf("getting current node: %w", err)
		}
		targetPath = current.Path
	}
	
	node, err := m.treeProvider.GetNode(targetPath)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}
	
	if recursive {
		return m.pullRecursive(node)
	}
	
	fullPath := m.computeFilesystemPath(node.Path)
	m.logProvider.Info("Pulling changes")
	m.logProvider.Info(fmt.Sprintf("  Tree path: %s", node.Path))
	m.logProvider.Info(fmt.Sprintf("  Directory: %s", fullPath))
	return m.gitProvider.Pull(fullPath, interfaces.PullOptions{})
}

func (m *Manager) pullRecursive(node interfaces.NodeInfo) error {
	if node.IsCloned {
		m.logProvider.Info(fmt.Sprintf("Pulling changes at %s", node.Path))
		if err := m.gitProvider.Pull(m.computeFilesystemPath(node.Path), interfaces.PullOptions{}); err != nil {
			m.logProvider.Error(fmt.Sprintf("Pull failed at %s: %v", node.Path, err))
		}
	}
	
	for _, child := range node.Children {
		if err := m.pullRecursive(child); err != nil {
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
		current, err := m.treeProvider.GetCurrent()
		if err != nil {
			return fmt.Errorf("getting current node: %w", err)
		}
		targetPath = current.Path
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
		current, err := m.treeProvider.GetCurrent()
		if err != nil {
			return fmt.Errorf("getting current node: %w", err)
		}
		targetPath = current.Path
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
		current, err := m.treeProvider.GetCurrent()
		if err != nil {
			return fmt.Errorf("getting current node: %w", err)
		}
		targetPath = current.Path
	}
	
	node, err := m.treeProvider.GetNode(targetPath)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}
	
	// Start Claude session using process provider
	// Compute filesystem path with repos/ directory pattern
	fullPath := m.computeFilesystemPath(node.Path)
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

## Organization Strategies

MUNO supports several repository organization patterns:

1. **Team-based**: Organize by team ownership
2. **Service-type**: Group by architectural layers (APIs, frontends, libraries)
3. **Domain-driven**: Follow domain boundaries (commerce, identity, etc.)
4. **Multi-cloud**: Organize by cloud provider

For detailed migration and organization guidance, see the AI Agent Guide above.

## Tips for AI Agents

- Use 'muno tree' to understand the full workspace structure
- Navigate with 'muno use <path>' to move between repositories
- Check status with 'muno status --recursive' for all repos
- The current directory is already set to the target repository

---
*This context was generated by MUNO to help AI agents understand the workspace structure.*
`, node.Path, workDir, treeOutput)
	
	// Write context file
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
	output.WriteString("```\n")
	
	// Get root node to show full tree
	root, err := m.treeProvider.GetNode("")
	if err != nil {
		// If we can't get root, just show current node
		root = *currentNode
	}
	
	// Generate tree with current node highlighted
	m.writeTreeNode(&output, &root, currentNode.Path, "", true)
	output.WriteString("```\n")
	
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
	
	// Add status indicators
	statusIndicator := ""
	if node.IsLazy && !node.IsCloned {
		statusIndicator = " [lazy]"
	} else if !node.IsCloned {
		statusIndicator = " [not cloned]"
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
		current, err := m.treeProvider.GetCurrent()
		if err != nil {
			return fmt.Errorf("getting current node: %w", err)
		}
		targetPath = current.Path
	}
	
	node, err := m.treeProvider.GetNode(targetPath)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}
	
	// Start agent session using process provider
	// Compute filesystem path with repos/ directory pattern
	fullPath := m.computeFilesystemPath(node.Path)
	m.logProvider.Info(fmt.Sprintf("Starting %s session", agentName))
	m.logProvider.Info(fmt.Sprintf("  Tree path: %s", node.Path))
	m.logProvider.Info(fmt.Sprintf("  Directory: %s", fullPath))
	
	// If requested, inject MUNO context for the agent
	if withMunoContext {
		m.logProvider.Info("  Creating MUNO context for agent...")
		if err := m.createAgentContext(&node, fullPath); err != nil {
			m.logProvider.Error(fmt.Sprintf("Failed to create agent context: %v", err))
			// Continue anyway, context is optional
		}
	}
	
	// Build the command
	command := fmt.Sprintf("cd %s && %s", fullPath, agentName)
	if len(agentArgs) > 0 {
		// Add agent-specific arguments
		for _, arg := range agentArgs {
			command += " " + arg
		}
	}
	
	result, err := m.processProvider.ExecuteShell(context.Background(), command, interfaces.ProcessOptions{})
	if err != nil {
		return fmt.Errorf("starting %s: %w", agentName, err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("%s exited with code %d: %s", agentName, result.ExitCode, result.Stderr)
	}
	
	return nil
}

// SmartInitWorkspace performs intelligent initialization for V2
func (m *Manager) SmartInitWorkspace(projectName string, options InitOptions) error {
	// Initialize the manager
	m.initialized = true
	
	// Try to load existing config first
	var existingConfig *config.ConfigTree
	configPath := filepath.Join(m.workspace, "muno.yaml")
	
	// Try different config file names
	for _, configName := range []string{"muno.yaml", "muno.yml", ".muno.yaml", ".muno.yml"} {
		testPath := filepath.Join(m.workspace, configName)
		if m.fsProvider.Exists(testPath) {
			cfg, err := config.LoadTree(testPath)
			if err == nil {
				existingConfig = cfg
				configPath = testPath
				m.logProvider.Info(fmt.Sprintf("Found existing config: %s", configName))
				break
			}
			m.logProvider.Warn(fmt.Sprintf("Failed to load %s: %v", configName, err))
		}
	}
	
	// Create or update configuration
	if existingConfig != nil {
		// Use existing config, but update project name if provided
		m.config = existingConfig
		if projectName != "" && projectName != existingConfig.Workspace.Name {
			m.config.Workspace.Name = projectName
		}
		m.logProvider.Info(fmt.Sprintf("Preserving %d existing nodes", len(m.config.Nodes)))
	} else {
		// Create new configuration
		m.config = &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     projectName,
				ReposDir: config.GetDefaultReposDir(),
			},
			Nodes: []config.NodeDefinition{},
			Path:  configPath,
		}
	}
	
	// Build a map of existing nodes to avoid duplicates
	existingNodes := make(map[string]bool)
	for _, node := range m.config.Nodes {
		existingNodes[node.Name] = true
	}
	
	// Get repos directory
	reposDir := m.getReposDir()
	if reposDir == "" {
		reposDir = "repos"
	}
	
	// Create repos directory if it doesn't exist
	reposDirPath := filepath.Join(m.workspace, reposDir)
	if err := m.fsProvider.MkdirAll(reposDirPath, 0755); err != nil {
		return fmt.Errorf("creating repos directory: %w", err)
	}
	
	// Define directories to ignore during scanning
	ignoreDirs := map[string]bool{
		".git":         true,
		".muno":        true,
		"node_modules": true,
		".idea":        true,
		".vscode":      true,
		"vendor":       true,
		".cache":       true,
		"dist":         true,
		"build":        true,
		"target":       true,
	}
	
	// Scan for existing git repositories
	m.logProvider.Info("Scanning for git repositories...")
	
	// First, scan the repos directory for existing repos
	reposInNodesDir := 0
	if m.fsProvider.Exists(reposDirPath) {
		entries, err := os.ReadDir(reposDirPath)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				
				repoName := entry.Name()
				if existingNodes[repoName] {
					continue // Already in config
				}
				
				repoPath := filepath.Join(reposDirPath, repoName)
				gitDir := filepath.Join(repoPath, ".git")
				
				if m.fsProvider.Exists(gitDir) {
					// Found a git repo in repos directory
					remote, _ := m.getGitRemote(repoPath)
					if remote != "" {
						m.logProvider.Info(fmt.Sprintf("Found repo in %s/: %s", reposDir, repoName))
						
						// Add to configuration
						m.config.Nodes = append(m.config.Nodes, config.NodeDefinition{
							Name: repoName,
							URL:  remote,
							Lazy: false,
						})
						existingNodes[repoName] = true
						reposInNodesDir++
					}
				}
			}
		}
	}
	
	// Now scan other directories (excluding repos dir and ignores)
	repos, err := m.findGitRepositoriesWithIgnore(".", append([]string{reposDir}, getMapKeys(ignoreDirs)...))
	if err != nil && !options.Force {
		return fmt.Errorf("scanning repositories: %w", err)
	}
	
	m.logProvider.Info(fmt.Sprintf("Found %d repositories in %s/, %d in other directories", 
		reposInNodesDir, reposDir, len(repos)))
	
	// Process found repositories outside repos directory
	movedRepos := 0
	for _, repo := range repos {
		repoName := filepath.Base(repo.Path)
		
		// Skip if already exists in config
		if existingNodes[repoName] {
			m.logProvider.Info(fmt.Sprintf("Skipping %s (already in config)", repoName))
			continue
		}
		
		// In interactive mode, ask user
		if !options.NonInteractive && !options.Force {
			m.uiProvider.Info(fmt.Sprintf("\nFound repository: %s", repo.Path))
			if repo.RemoteURL != "" {
				m.uiProvider.Info(fmt.Sprintf("  Remote URL: %s", repo.RemoteURL))
			}
			
			shouldAdd, err := m.uiProvider.Confirm("Add to workspace?")
			if err != nil || !shouldAdd {
				continue
			}
		}
		
		// Move to repos directory
		targetPath := filepath.Join(reposDirPath, repoName)
		if err := os.Rename(repo.Path, targetPath); err != nil {
			m.logProvider.Warn(fmt.Sprintf("Failed to move %s: %v", repoName, err))
			continue
		}
		
		// Add to configuration
		url := repo.RemoteURL
		if url == "" {
			url = "file://" + targetPath
		}
		
		m.config.Nodes = append(m.config.Nodes, config.NodeDefinition{
			Name: repoName,
			URL:  url,
			Lazy: false,
		})
		existingNodes[repoName] = true
		movedRepos++
		
		m.logProvider.Info(fmt.Sprintf("Moved and added: %s", repoName))
	}
	
	// Validate all nodes
	for i, node := range m.config.Nodes {
		// Ensure node has either URL or Config, not both
		if node.URL != "" && node.Config != "" {
			m.logProvider.Warn(fmt.Sprintf("Node %s has both URL and Config, clearing Config", node.Name))
			m.config.Nodes[i].Config = ""
		}
		
		// Validate config references exist
		if node.Config != "" {
			configPath := node.Config
			if !filepath.IsAbs(configPath) {
				configPath = filepath.Join(m.workspace, configPath)
			}
			if !m.fsProvider.Exists(configPath) {
				m.logProvider.Warn(fmt.Sprintf("Config file not found for %s: %s", node.Name, node.Config))
			}
		}
	}
	
	// Save the configuration
	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}
	
	// Create CLAUDE.md if it doesn't exist
	claudePath := filepath.Join(m.workspace, "CLAUDE.md")
	if !m.fsProvider.Exists(claudePath) {
		content := fmt.Sprintf("# %s\n\nMUNO workspace with tree-based navigation.\n", m.config.Workspace.Name)
		if err := m.fsProvider.WriteFile(claudePath, []byte(content), 0644); err != nil {
			m.logProvider.Warn(fmt.Sprintf("Failed to create CLAUDE.md: %v", err))
		}
	}
	
	// Summary
	m.uiProvider.Success(fmt.Sprintf("\nWorkspace '%s' initialized successfully!", m.config.Workspace.Name))
	m.uiProvider.Info(fmt.Sprintf("\nConfiguration summary:"))
	m.uiProvider.Info(fmt.Sprintf("  - Total nodes: %d", len(m.config.Nodes)))
	
	configRefs := 0
	gitRepos := 0
	for _, node := range m.config.Nodes {
		if node.Config != "" {
			configRefs++
		} else if node.URL != "" {
			gitRepos++
		}
	}
	
	m.uiProvider.Info(fmt.Sprintf("  - Git repositories: %d", gitRepos))
	m.uiProvider.Info(fmt.Sprintf("  - Config references: %d", configRefs))
	
	if movedRepos > 0 {
		m.uiProvider.Info(fmt.Sprintf("  - Moved to %s/: %d", reposDir, movedRepos))
	}
	
	m.uiProvider.Info("\nNext steps:")
	m.uiProvider.Info("  muno tree        # View repository tree")
	m.uiProvider.Info("  muno use <repo>  # Navigate to a repository")
	m.uiProvider.Info("  muno add <url>   # Add more repositories")
	
	return nil
}

// findGitRepositories searches for git repositories in the given path
func (m *Manager) findGitRepositories(rootPath string) ([]GitRepoInfo, error) {
	var repos []GitRepoInfo
	visited := make(map[string]bool)
	
	err := m.fsProvider.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		
		// Skip hidden directories except .git
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".git" {
			return filepath.SkipDir
		}
		
		// Skip common non-repo directories
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == "vendor") {
			return filepath.SkipDir
		}
		
		// Check if this is a .git directory
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			absRepoPath, _ := filepath.Abs(repoPath)
			
			if visited[absRepoPath] {
				return filepath.SkipDir
			}
			visited[absRepoPath] = true
			
			repoInfo := GitRepoInfo{
				Path: repoPath,
			}
			
			// Get remote URL
			if url, err := m.gitProvider.GetRemoteURL(repoPath); err == nil {
				repoInfo.RemoteURL = url
			}
			
			// Get current branch
			if branch, err := m.gitProvider.Branch(repoPath); err == nil {
				repoInfo.Branch = branch
			}
			
			repos = append(repos, repoInfo)
			return filepath.SkipDir
		}
		
		return nil
	})
	
	return repos, err
}

// ClearCurrent clears the current position in the tree
func (m *Manager) ClearCurrent() error {
	if !m.initialized {
		return fmt.Errorf("manager not initialized")
	}
	
	return m.treeProvider.SetPath("/")
}

// NewManagerForInit creates a new Manager for initialization without requiring existing config
func NewManagerForInit(projectPath string) (*Manager, error) {
	
	// Ensure project path exists
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return nil, fmt.Errorf("creating project directory: %w", err)
	}
	
	// Create git interface for tree manager
	gitInterface := git.New()
	
	// Create tree manager
	treeMgr, err := tree.NewManager(projectPath, gitInterface)
	if err != nil {
		return nil, fmt.Errorf("creating tree manager: %w", err)
	}
	
	// Create Manager options with default providers
	opts := ManagerOptions{
		ConfigProvider:  NewConfigAdapter(nil),
		GitProvider:     NewGitAdapter(nil),
		FSProvider:      NewFileSystemAdapter(nil),
		UIProvider:      NewUIAdapter(nil),
		TreeProvider:    NewTreeAdapter(treeMgr),
		ProcessProvider: NewDefaultProcessProvider(),
		LogProvider:     NewDefaultLogProvider(false),
		MetricsProvider: NewNoOpMetricsProvider(),
		EnablePlugins:   false,
		AutoLoadConfig:  false, // Don't try to load config for init
	}
	
	// Create Manager
	mgr, err := NewManager(opts)
	if err != nil {
		return nil, fmt.Errorf("creating manager: %w", err)
	}
	
	// Set workspace path
	mgr.workspace = projectPath
	
	return mgr, nil
}

// LoadFromCurrentDir loads a Manager from the current directory
func LoadFromCurrentDir() (*Manager, error) {
	return LoadFromCurrentDirWithOptions(nil)
}

// LoadFromCurrentDirWithOptions loads a Manager from the current directory with options
func LoadFromCurrentDirWithOptions(opts *ManagerOptions) (*Manager, error) {
	// Use defaults if no options provided
	if opts == nil {
		opts = DefaultManagerOptions()
	}
	
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Search upwards for muno.yaml
	searchDir := cwd
	configPath := ""
	for {
		candidate := filepath.Join(searchDir, "muno.yaml")
		if _, err := os.Stat(candidate); err == nil {
			configPath = candidate
			cwd = searchDir // Update cwd to the project root
			break
		}
		
		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			break // Reached root
		}
		searchDir = parent
	}
	
	if configPath == "" {
		return nil, fmt.Errorf("muno.yaml not found in current directory or any parent")
	}
	
	// Create git interface for tree manager
	gitInterface := git.New()
	
	// Create tree manager
	treeMgr, err := tree.NewManager(cwd, gitInterface)
	if err != nil {
		return nil, fmt.Errorf("creating tree manager: %w", err)
	}
	
	// Create Manager options with default providers
	managerOpts := ManagerOptions{
		ConfigProvider:  NewConfigAdapter(nil),
		GitProvider:     NewGitAdapter(nil),
		FSProvider:      NewFileSystemAdapter(nil),
		UIProvider:      NewUIAdapter(nil),
		TreeProvider:    NewTreeAdapter(treeMgr),
		ProcessProvider: NewDefaultProcessProvider(),
		LogProvider:     NewDefaultLogProvider(false),
		MetricsProvider: NewNoOpMetricsProvider(),
		EnablePlugins:   false,
		AutoLoadConfig:  true,
	}
	
	// Create Manager
	mgr, err := NewManager(managerOpts)
	if err != nil {
		return nil, fmt.Errorf("creating manager: %w", err)
	}
	
	// Initialize with the workspace path
	ctx := context.Background()
	if err := mgr.Initialize(ctx, cwd); err != nil {
		return nil, fmt.Errorf("initializing manager: %w", err)
	}
	
	return mgr, nil
}

// findGitRepositoriesWithIgnore searches for git repositories excluding specified directories
func (m *Manager) findGitRepositoriesWithIgnore(rootPath string, ignoreDirs []string) ([]GitRepoInfo, error) {
	var repos []GitRepoInfo
	visited := make(map[string]bool)
	
	// Convert ignore list to map for faster lookup
	ignoreMap := make(map[string]bool)
	for _, dir := range ignoreDirs {
		ignoreMap[dir] = true
	}
	
	err := m.fsProvider.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		
		// Skip ignored directories
		if info.IsDir() && ignoreMap[info.Name()] {
			return filepath.SkipDir
		}
		
		// Skip hidden directories except .git
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".git" {
			return filepath.SkipDir
		}
		
		// Check if this is a .git directory
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			absRepoPath, _ := filepath.Abs(repoPath)
			
			if visited[absRepoPath] {
				return filepath.SkipDir
			}
			visited[absRepoPath] = true
			
			repoInfo := GitRepoInfo{
				Path: repoPath,
			}
			
			// Get remote URL
			if url, err := m.gitProvider.GetRemoteURL(repoPath); err == nil {
				repoInfo.RemoteURL = url
			}
			
			// Get current branch
			if branch, err := m.gitProvider.Branch(repoPath); err == nil {
				repoInfo.Branch = branch
			}
			
			repos = append(repos, repoInfo)
			return filepath.SkipDir
		}
		
		return nil
	})
	
	return repos, err
}

// getGitRemote gets the remote URL for a git repository
func (m *Manager) getGitRemote(repoPath string) (string, error) {
	return m.gitProvider.GetRemoteURL(repoPath)
}

// getMapKeys returns the keys of a map as a slice
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}