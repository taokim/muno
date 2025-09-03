package tree

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
)

// StatelessManager manages tree without persistent state file
type StatelessManager struct {
	workspacePath string
	config        *config.ConfigTree
	resolver      *ConfigResolver
	gitCmd        git.Interface
	currentPath   string  // Current logical path (session only)
}

// NewStatelessManager creates a manager that derives state from filesystem
func NewStatelessManager(workspacePath string, gitCmd git.Interface) (*StatelessManager, error) {
	// Load config from workspace
	configPath := filepath.Join(workspacePath, "muno.yaml")
	cfg, err := config.LoadTree(configPath)
	if err != nil {
		// If no config, create default
		if os.IsNotExist(err) {
			cfg = config.DefaultConfigTree("workspace")
		} else {
			return nil, fmt.Errorf("loading config: %w", err)
		}
	}
	
	resolver := NewConfigResolver(workspacePath)
	
	return &StatelessManager{
		workspacePath: workspacePath,
		config:        cfg,
		resolver:      resolver,
		gitCmd:        gitCmd,
		currentPath:   "/",
	}, nil
}

// ComputeFilesystemPath converts logical path to filesystem path
func (m *StatelessManager) ComputeFilesystemPath(logicalPath string) string {
	if logicalPath == "/" || logicalPath == "" {
		return filepath.Join(m.workspacePath, m.config.GetReposDir())
	}
	
	// Split path: /level1/level2 -> [level1, level2]
	parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")
	
	// Build filesystem path with repos/ subdirectories
	fsPath := filepath.Join(m.workspacePath, m.config.GetReposDir())
	for i, part := range parts {
		fsPath = filepath.Join(fsPath, part)
		// Add repos/ before next level (except last)
		if i < len(parts)-1 {
			fsPath = filepath.Join(fsPath, "repos")
		}
	}
	
	return fsPath
}

// GetNodeByPath finds a node by its logical path
func (m *StatelessManager) GetNodeByPath(logicalPath string) (*config.NodeDefinition, error) {
	if logicalPath == "/" || logicalPath == "" {
		return nil, nil // Root node
	}
	
	parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")
	if len(parts) == 0 {
		return nil, nil
	}
	
	// Find in top-level nodes
	for _, node := range m.config.Nodes {
		if node.Name == parts[0] {
			if len(parts) == 1 {
				return &node, nil
			}
			// For deeper paths, would need to load sub-configs
			// For now, just check if node exists on filesystem
			return &node, nil
		}
	}
	
	return nil, fmt.Errorf("node not found: %s", logicalPath)
}

// UseNode navigates to a node
func (m *StatelessManager) UseNode(logicalPath string) error {
	// Normalize path
	if logicalPath == "" {
		logicalPath = "/"
	}
	if !strings.HasPrefix(logicalPath, "/") {
		// Relative path
		logicalPath = path.Join(m.currentPath, logicalPath)
	}
	
	fsPath := m.ComputeFilesystemPath(logicalPath)
	
	// If it's a repository node, check if it needs cloning
	if logicalPath != "/" {
		node, err := m.GetNodeByPath(logicalPath)
		if err != nil {
			return err
		}
		
		if node != nil && node.URL != "" {
			// Check if repo exists
			gitPath := filepath.Join(fsPath, ".git")
			if _, err := os.Stat(gitPath); os.IsNotExist(err) {
				// Auto-clone if lazy
				fmt.Printf("Auto-cloning repository: %s\n", node.Name)
				if err := m.gitCmd.Clone(node.URL, fsPath); err != nil {
					return fmt.Errorf("failed to clone %s: %w", node.Name, err)
				}
			}
		}
	}
	
	// Ensure directory exists
	if err := os.MkdirAll(fsPath, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	
	// Change directory
	if err := os.Chdir(fsPath); err != nil {
		return fmt.Errorf("changing directory to %s: %w", fsPath, err)
	}
	
	m.currentPath = logicalPath
	return nil
}

// AddRepo adds a new repository to config
func (m *StatelessManager) AddRepo(parentPath, name, url string, lazy bool) error {
	// For stateless operation, we add to config and save
	node := config.NodeDefinition{
		Name: name,
		URL:  url,
		Lazy: lazy,
	}
	
	// Add to nodes
	m.config.Nodes = append(m.config.Nodes, node)
	
	// Save config
	configPath := filepath.Join(m.workspacePath, "muno.yaml")
	if err := m.config.Save(configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	
	// Clone if not lazy
	if !lazy {
		targetPath := path.Join(parentPath, name)
		fsPath := m.ComputeFilesystemPath(targetPath)
		fmt.Printf("Cloning %s to %s\n", url, fsPath)
		if err := m.gitCmd.Clone(url, fsPath); err != nil {
			return fmt.Errorf("cloning: %w", err)
		}
	}
	
	return nil
}

// RemoveNode removes a node from config
func (m *StatelessManager) RemoveNode(targetPath string) error {
	if targetPath == "/" {
		return fmt.Errorf("cannot remove root")
	}
	
	// Parse node name from path
	parts := strings.Split(strings.TrimPrefix(targetPath, "/"), "/")
	if len(parts) == 0 {
		return fmt.Errorf("invalid path")
	}
	
	targetName := parts[0]
	
	// Remove from config
	newNodes := []config.NodeDefinition{}
	for _, node := range m.config.Nodes {
		if node.Name != targetName {
			newNodes = append(newNodes, node)
		}
	}
	m.config.Nodes = newNodes
	
	// Save config
	configPath := filepath.Join(m.workspacePath, "muno.yaml")
	if err := m.config.Save(configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	
	// Remove from filesystem
	fsPath := m.ComputeFilesystemPath(targetPath)
	if err := os.RemoveAll(fsPath); err != nil {
		fmt.Printf("Warning: failed to remove %s: %v\n", fsPath, err)
	}
	
	return nil
}

// GetCurrentPath returns current logical path
func (m *StatelessManager) GetCurrentPath() string {
	return m.currentPath
}

// ListChildren lists children nodes
func (m *StatelessManager) ListChildren(targetPath string) ([]*TreeNode, error) {
	if targetPath == "" {
		targetPath = m.currentPath
	}
	
	// For root or current level, return config nodes
	if targetPath == "/" {
		children := make([]*TreeNode, 0, len(m.config.Nodes))
		for _, node := range m.config.Nodes {
			fsPath := m.ComputeFilesystemPath("/" + node.Name)
			state := GetRepoState(fsPath)
			
			children = append(children, &TreeNode{
				Name:  node.Name,
				Type:  NodeTypeRepo,
				URL:   node.URL,
				Lazy:  node.Lazy,
				State: state,
			})
		}
		return children, nil
	}
	
	// For sub-nodes, would need to load their configs
	// For now, return empty
	return []*TreeNode{}, nil
}

// CloneLazyRepos clones lazy repositories
func (m *StatelessManager) CloneLazyRepos(targetPath string, recursive bool) error {
	for _, node := range m.config.Nodes {
		if node.URL != "" {
			nodePath := "/" + node.Name
			fsPath := m.ComputeFilesystemPath(nodePath)
			
			// Check if needs cloning
			if _, err := os.Stat(filepath.Join(fsPath, ".git")); os.IsNotExist(err) {
				fmt.Printf("Cloning %s to %s\n", node.URL, fsPath)
				if err := m.gitCmd.Clone(node.URL, fsPath); err != nil {
					return fmt.Errorf("cloning %s: %w", node.Name, err)
				}
			}
			
			// If recursive and node has config, process sub-nodes
			if recursive && node.Config != "" {
				// Would load sub-config and process
			}
		}
	}
	
	return nil
}

// DisplayTree shows tree structure
func (m *StatelessManager) DisplayTree() string {
	var output strings.Builder
	output.WriteString(fmt.Sprintf("🌳 %s\n", m.config.Workspace.Name))
	
	for i, node := range m.config.Nodes {
		prefix := "├─"
		if i == len(m.config.Nodes)-1 {
			prefix = "└─"
		}
		
		icon := "📦"
		fsPath := m.ComputeFilesystemPath("/" + node.Name)
		state := GetRepoState(fsPath)
		
		if state == RepoStateMissing {
			icon = "💤"
		} else if state == RepoStateModified {
			icon = "📝"
		}
		
		if node.Config != "" {
			icon = "📁" // Config reference
		}
		
		status := ""
		if node.Lazy && state == RepoStateMissing {
			status = " (lazy)"
		}
		
		output.WriteString(fmt.Sprintf("%s %s %s%s\n", prefix, icon, node.Name, status))
	}
	
	return output.String()
}

// DisplayStatus shows current status
func (m *StatelessManager) DisplayStatus() string {
	var output strings.Builder
	output.WriteString("=== Tree Status ===\n")
	output.WriteString(fmt.Sprintf("Current Path: %s\n", m.currentPath))
	output.WriteString(fmt.Sprintf("Workspace: %s\n", m.config.Workspace.Name))
	output.WriteString(fmt.Sprintf("Nodes: %d\n", len(m.config.Nodes)))
	
	// Count states
	var cloned, missing, modified int
	for _, node := range m.config.Nodes {
		if node.URL != "" {
			fsPath := m.ComputeFilesystemPath("/" + node.Name)
			state := GetRepoState(fsPath)
			switch state {
			case RepoStateCloned:
				cloned++
			case RepoStateMissing:
				missing++
			case RepoStateModified:
				modified++
			}
		}
	}
	
	output.WriteString(fmt.Sprintf("Repositories: %d cloned, %d missing, %d modified\n", 
		cloned, missing, modified))
	
	return output.String()
}

// DisplayPath shows current path
func (m *StatelessManager) DisplayPath() string {
	return m.currentPath
}

// DisplayChildren shows children of current node
func (m *StatelessManager) DisplayChildren() string {
	children, _ := m.ListChildren(m.currentPath)
	if len(children) == 0 {
		return "No children\n"
	}
	
	var output strings.Builder
	output.WriteString("Children:\n")
	for _, child := range children {
		status := ""
		if child.State == RepoStateMissing {
			status = " (lazy)"
		} else if child.State == RepoStateModified {
			status = " (modified)"
		}
		output.WriteString(fmt.Sprintf("  - %s%s\n", child.Name, status))
	}
	
	return output.String()
}

// DisplayTreeWithDepth is implemented in display_tree_with_depth.go