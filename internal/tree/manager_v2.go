// Package tree provides tree-based repository management
package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/tree/navigator"
)

// ManagerV2 is the refactored manager using the navigator interface
type ManagerV2 struct {
	workspace  string
	config     *config.ConfigTree
	navigator  navigator.TreeNavigator
	gitCmd     interfaces.GitInterface
}

// Option is a functional option for configuring ManagerV2
type Option func(*ManagerV2)

// WithNavigator sets a custom navigator
func WithNavigator(nav navigator.TreeNavigator) Option {
	return func(m *ManagerV2) {
		m.navigator = nav
	}
}

// WithGitCommand sets a custom git interface
func WithGitCommand(git interfaces.GitInterface) Option {
	return func(m *ManagerV2) {
		m.gitCmd = git
	}
}

// WithConfig sets a custom configuration
func WithConfig(cfg *config.ConfigTree) Option {
	return func(m *ManagerV2) {
		m.config = cfg
	}
}

// NewManagerV2 creates a new manager with navigator-based architecture
func NewManagerV2(workspace string, opts ...Option) (*ManagerV2, error) {
	if workspace == "" {
		return nil, fmt.Errorf("workspace path cannot be empty")
	}

	// Ensure workspace exists
	if err := os.MkdirAll(workspace, 0755); err != nil {
		return nil, fmt.Errorf("creating workspace: %w", err)
	}

	m := &ManagerV2{
		workspace: workspace,
	}

	// Apply options
	for _, opt := range opts {
		opt(m)
	}

	// Load config if not provided
	if m.config == nil {
		var err error
		for _, configName := range config.GetConfigFileNames() {
			configPath := filepath.Join(workspace, configName)
			m.config, err = config.LoadTree(configPath)
			if err == nil {
				break
			}
		}
		if m.config == nil {
			m.config = config.DefaultConfigTree("workspace")
		}
	}

	// Create git command if not provided
	if m.gitCmd == nil {
		// This would normally use the real git implementation
		// For now, we'll leave it nil and handle in methods
	}

	// Create navigator if not provided
	if m.navigator == nil {
		// Default to filesystem navigator
		factory := navigator.NewFactory(workspace, m.config, m.gitCmd)
		nav, err := factory.CreateDefault()
		if err != nil {
			return nil, fmt.Errorf("creating navigator: %w", err)
		}
		m.navigator = nav
	}

	return m, nil
}

// Navigation methods

// GetCurrentPath returns the current position in the tree
func (m *ManagerV2) GetCurrentPath() (string, error) {
	return m.navigator.GetCurrentPath()
}

// UseNode navigates to a node in the tree
func (m *ManagerV2) UseNode(path string) error {
	return m.navigator.Navigate(path)
}

// Tree query methods

// GetNode returns a node by its path
func (m *ManagerV2) GetNode(path string) (*navigator.Node, error) {
	return m.navigator.GetNode(path)
}

// ListChildren lists the children of the current or specified node
func (m *ManagerV2) ListChildren(path string) ([]*navigator.Node, error) {
	if path == "" {
		path, _ = m.navigator.GetCurrentPath()
	}
	return m.navigator.ListChildren(path)
}

// GetTree returns a tree view from the specified path
func (m *ManagerV2) GetTree(path string, depth int) (*navigator.TreeView, error) {
	if path == "" {
		path = "/"
	}
	return m.navigator.GetTree(path, depth)
}

// Repository management methods

// AddRepo adds a repository as a child of the current or specified parent
func (m *ManagerV2) AddRepo(parentPath, name, url string, lazy bool) error {
	// Default to current path if not specified
	if parentPath == "" {
		parentPath, _ = m.navigator.GetCurrentPath()
	}

	// Normalize parent path
	if !strings.HasPrefix(parentPath, "/") {
		currentPath, _ := m.navigator.GetCurrentPath()
		parentPath = filepath.Join(currentPath, parentPath)
	}
	parentPath = filepath.Clean(parentPath)
	if parentPath == "." {
		parentPath = "/"
	}

	// Check parent exists
	parent, err := m.navigator.GetNode(parentPath)
	if err != nil {
		return fmt.Errorf("getting parent node: %w", err)
	}
	if parent == nil {
		return fmt.Errorf("parent node not found: %s", parentPath)
	}

	// Check if child already exists
	for _, childName := range parent.Children {
		if childName == name {
			return fmt.Errorf("child %s already exists in %s", name, parentPath)
		}
	}

	// Add to config if adding to root
	if parentPath == "/" {
		fetchMode := "lazy"
		if !lazy {
			fetchMode = "eager"
		}

		// Add to config
		m.config.Nodes = append(m.config.Nodes, config.NodeDefinition{
			Name:  name,
			URL:   url,
			Fetch: fetchMode,
		})

		// Save config to file
		configPath := filepath.Join(m.workspace, "muno.yaml")
		if err := m.config.Save(configPath); err != nil {
			// Rollback
			m.config.Nodes = m.config.Nodes[:len(m.config.Nodes)-1]
			return fmt.Errorf("failed to save config: %w", err)
		}
	} else {
		// For non-root additions, we need to handle nested configs
		// This is a limitation for now
		fmt.Printf("Warning: Repository added but not persisted (non-root additions not yet supported)\n")
	}

	// Clone if not lazy
	if !lazy {
		childPath := filepath.Join(parentPath, name)
		if err := m.cloneRepository(url, childPath); err != nil {
			// Rollback config change
			if parentPath == "/" {
				m.config.Nodes = m.config.Nodes[:len(m.config.Nodes)-1]
				configPath := filepath.Join(m.workspace, "muno.yaml")
				m.config.Save(configPath)
			}
			return fmt.Errorf("failed to clone: %w", err)
		}
	}

	// Refresh navigator to pick up changes
	return m.navigator.RefreshStatus(parentPath)
}

// RemoveNode removes a node and its subtree
func (m *ManagerV2) RemoveNode(targetPath string) error {
	if targetPath == "/" {
		return fmt.Errorf("cannot remove root node")
	}

	// Normalize path
	if !strings.HasPrefix(targetPath, "/") {
		currentPath, _ := m.navigator.GetCurrentPath()
		targetPath = filepath.Join(currentPath, targetPath)
	}

	node, err := m.navigator.GetNode(targetPath)
	if err != nil {
		return err
	}
	if node == nil {
		return fmt.Errorf("node not found: %s", targetPath)
	}

	// Find parent path
	parentPath := filepath.Dir(targetPath)
	if parentPath == "." {
		parentPath = "/"
	}

	// Remove from filesystem
	fsPath := m.computeFilesystemPath(targetPath)
	if err := os.RemoveAll(fsPath); err != nil {
		fmt.Printf("Warning: failed to remove directory %s: %v\n", fsPath, err)
	}

	// If we removed the current node, navigate to parent
	currentPath, _ := m.navigator.GetCurrentPath()
	if strings.HasPrefix(currentPath, targetPath) {
		m.navigator.Navigate(parentPath)
	}

	// Update config if removing from root
	if parentPath == "/" {
		// Find and remove from config
		newNodes := []config.NodeDefinition{}
		for _, nodeDef := range m.config.Nodes {
			if nodeDef.Name != node.Name {
				newNodes = append(newNodes, nodeDef)
			}
		}
		m.config.Nodes = newNodes

		// Save config to file
		configPath := filepath.Join(m.workspace, "muno.yaml")
		if err := m.config.Save(configPath); err != nil {
			fmt.Printf("Warning: failed to save config after removal: %v\n", err)
		}
	}

	// Refresh navigator
	return m.navigator.RefreshStatus(parentPath)
}

// CloneLazyRepos clones all lazy repositories in the current or specified node
func (m *ManagerV2) CloneLazyRepos(targetPath string, recursive bool) error {
	if targetPath == "" {
		targetPath, _ = m.navigator.GetCurrentPath()
	}

	// Get tree view to find lazy repos
	depth := 1
	if recursive {
		depth = -1
	}

	tree, err := m.navigator.GetTree(targetPath, depth)
	if err != nil {
		return err
	}

	// Clone each lazy repo
	for path, status := range tree.Status {
		if status.Lazy && !status.Cloned {
			fmt.Printf("Cloning lazy repository at %s\n", path)
			if err := m.navigator.TriggerLazyLoad(path); err != nil {
				return fmt.Errorf("failed to clone %s: %w", path, err)
			}
		}
	}

	return nil
}

// Status methods

// GetStatus returns the status of a node
func (m *ManagerV2) GetStatus(path string) (*navigator.NodeStatus, error) {
	if path == "" {
		path, _ = m.navigator.GetCurrentPath()
	}
	return m.navigator.GetNodeStatus(path)
}

// RefreshStatus forces a status refresh
func (m *ManagerV2) RefreshStatus(path string) error {
	if path == "" {
		path = "/"
	}
	return m.navigator.RefreshStatus(path)
}

// Display methods

// DisplayTree displays the tree structure
func (m *ManagerV2) DisplayTree(path string, depth int) error {
	if path == "" {
		path = "/"
	}

	tree, err := m.navigator.GetTree(path, depth)
	if err != nil {
		return err
	}

	// Use existing display functionality
	display := NewDisplay(os.Stdout, &DisplayOptions{
		ShowStatus: true,
		ShowLazy:   true,
	})

	return display.PrintTreeView(tree)
}

// DisplayStatus displays the status of repositories
func (m *ManagerV2) DisplayStatus(path string, recursive bool) error {
	if path == "" {
		path, _ = m.navigator.GetCurrentPath()
	}

	depth := 1
	if recursive {
		depth = -1
	}

	tree, err := m.navigator.GetTree(path, depth)
	if err != nil {
		return err
	}

	// Use existing display functionality
	display := NewDisplay(os.Stdout, &DisplayOptions{
		ShowStatus: true,
		ShowBranch: true,
	})

	return display.PrintStatusView(tree)
}

// Git operations

// Pull performs git pull on repositories
func (m *ManagerV2) Pull(path string, recursive bool) error {
	if m.gitCmd == nil {
		return fmt.Errorf("git command not available")
	}

	if path == "" {
		path, _ = m.navigator.GetCurrentPath()
	}

	depth := 1
	if recursive {
		depth = -1
	}

	tree, err := m.navigator.GetTree(path, depth)
	if err != nil {
		return err
	}

	for nodePath, node := range tree.Nodes {
		if node.Type == navigator.NodeTypeRepo {
			status := tree.Status[nodePath]
			if status != nil && status.Cloned {
				fsPath := m.computeFilesystemPath(nodePath)
				fmt.Printf("Pulling %s...\n", nodePath)
				if err := m.gitCmd.Pull(fsPath); err != nil {
					fmt.Printf("Warning: failed to pull %s: %v\n", nodePath, err)
				}
			}
		}
	}

	return nil
}

// Push performs git push on repositories
func (m *ManagerV2) Push(path string, recursive bool) error {
	if m.gitCmd == nil {
		return fmt.Errorf("git command not available")
	}

	if path == "" {
		path, _ = m.navigator.GetCurrentPath()
	}

	depth := 1
	if recursive {
		depth = -1
	}

	tree, err := m.navigator.GetTree(path, depth)
	if err != nil {
		return err
	}

	for nodePath, node := range tree.Nodes {
		if node.Type == navigator.NodeTypeRepo {
			status := tree.Status[nodePath]
			if status != nil && status.Cloned && status.HasLocalChanges() {
				fsPath := m.computeFilesystemPath(nodePath)
				fmt.Printf("Pushing %s...\n", nodePath)
				if err := m.gitCmd.Push(fsPath); err != nil {
					fmt.Printf("Warning: failed to push %s: %v\n", nodePath, err)
				}
			}
		}
	}

	return nil
}

// Commit performs git commit on repositories
func (m *ManagerV2) Commit(path string, message string, recursive bool) error {
	if m.gitCmd == nil {
		return fmt.Errorf("git command not available")
	}

	if path == "" {
		path, _ = m.navigator.GetCurrentPath()
	}

	depth := 1
	if recursive {
		depth = -1
	}

	tree, err := m.navigator.GetTree(path, depth)
	if err != nil {
		return err
	}

	for nodePath, node := range tree.Nodes {
		if node.Type == navigator.NodeTypeRepo {
			status := tree.Status[nodePath]
			if status != nil && status.Cloned && status.Modified {
				fsPath := m.computeFilesystemPath(nodePath)
				fmt.Printf("Committing %s...\n", nodePath)
				if err := m.gitCmd.AddAll(fsPath); err != nil {
					fmt.Printf("Warning: failed to add files in %s: %v\n", nodePath, err)
					continue
				}
				if err := m.gitCmd.Commit(fsPath, message); err != nil {
					fmt.Printf("Warning: failed to commit %s: %v\n", nodePath, err)
				}
			}
		}
	}

	return nil
}

// Helper methods

func (m *ManagerV2) computeFilesystemPath(logicalPath string) string {
	reposDir := m.config.GetReposDir()

	// For root, use repos directory
	if logicalPath == "/" || logicalPath == "" {
		return filepath.Join(m.workspace, reposDir)
	}

	// Split path into components
	parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")

	// Check if this is a top-level repository
	if len(parts) == 1 {
		// Top-level repos go directly in workspace
		return filepath.Join(m.workspace, parts[0])
	}

	// For nested paths, build the full path
	pathComponents := []string{m.workspace}

	// Check if first component is a repo
	// This is simplified - would need to check actual node type
	firstNode, _ := m.navigator.GetNode("/" + parts[0])
	if firstNode != nil && firstNode.Type == navigator.NodeTypeRepo {
		// It's a repository at top level
		pathComponents = append(pathComponents, parts...)
	} else {
		// Use repos directory for non-repo top-level nodes
		pathComponents = append(pathComponents, reposDir)
		pathComponents = append(pathComponents, parts...)
	}

	return filepath.Join(pathComponents...)
}

func (m *ManagerV2) cloneRepository(url, path string) error {
	if m.gitCmd == nil {
		return fmt.Errorf("git command not available")
	}

	fsPath := m.computeFilesystemPath(path)

	// Create parent directory
	parentDir := filepath.Dir(fsPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("creating parent directory: %w", err)
	}

	// Clone the repository
	fmt.Printf("Cloning %s to %s\n", url, fsPath)
	return m.gitCmd.Clone(url, fsPath)
}

// GetWorkspacePath returns the workspace path
func (m *ManagerV2) GetWorkspacePath() string {
	return m.workspace
}

// GetConfig returns the configuration
func (m *ManagerV2) GetConfig() *config.ConfigTree {
	return m.config
}