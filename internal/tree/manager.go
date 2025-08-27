package tree

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/taokim/repo-claude/internal/config"
	"github.com/taokim/repo-claude/internal/git"
)

// Manager manages the workspace tree
type Manager struct {
	rootPath     string
	reposPath    string  // repos/ directory
	rootNode     *Node
	currentNode  *Node
	config       *config.ConfigV3Tree
	state        *TreeState
	statePath    string
	gitCmd       *git.Git
}

// NewManager creates a new tree manager
func NewManager(projectPath string) (*Manager, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("resolving project path: %w", err)
	}
	
	return &Manager{
		rootPath:      absPath,
		reposPath:     filepath.Join(absPath, "repos"),
		statePath:     filepath.Join(absPath, ".repo-claude-tree.json"),
		gitCmd:        git.New(),
	}, nil
}

// Initialize creates a new workspace tree
func (m *Manager) Initialize(projectName string, rootRepoURL string) error {
	// Create repos directory
	if err := os.MkdirAll(m.reposPath, 0755); err != nil {
		return fmt.Errorf("creating repos directory: %w", err)
	}
	
	// Create root node
	m.rootNode = &Node{
		ID:       "root",
		Name:     projectName,
		Path:     "/",
		FullPath: m.reposPath,
		Children: make(map[string]*Node),
		Meta: NodeMeta{
			Type:      string(NodeTypePersistent),
			CreatedAt: time.Now(),
		},
	}
	
	// If root is also a repo, clone it
	if rootRepoURL != "" {
		rootRepo := RepoConfig{
			URL:   rootRepoURL,
			Path:  m.reposPath,
			Name:  "root",
			Lazy:  false,
			State: string(RepoStateMissing),
		}
		
		if err := m.cloneRepo(&rootRepo); err != nil {
			return fmt.Errorf("cloning root repository: %w", err)
		}
		
		m.rootNode.Repos = append(m.rootNode.Repos, rootRepo)
	}
	
	// Set as current
	m.currentNode = m.rootNode
	
	// Save initial state
	return m.saveState()
}

// LoadTree loads the tree from saved state
func (m *Manager) LoadTree() error {
	// Load tree state
	state, err := m.loadState()
	if err != nil {
		return fmt.Errorf("loading tree state: %w", err)
	}
	
	m.state = state
	
	// Reconstruct tree from state
	if err := m.reconstructTree(); err != nil {
		return fmt.Errorf("reconstructing tree: %w", err)
	}
	
	// Set current node
	if state.CurrentNodePath != "" {
		node := m.findNodeByPath(state.CurrentNodePath)
		if node != nil {
			m.currentNode = node
		}
	}
	
	if m.currentNode == nil {
		m.currentNode = m.rootNode
	}
	
	return nil
}

// ResolveTarget determines the target node based on priority rules
func (m *Manager) ResolveTarget(explicitPath string) (*TargetResolution, error) {
	// 1. Explicit path always wins
	if explicitPath != "" {
		node := m.findNodeByPath(explicitPath)
		if node == nil {
			return nil, fmt.Errorf("node not found: %s", explicitPath)
		}
		return &TargetResolution{
			Node:   node,
			Source: SourceExplicit,
		}, nil
	}
	
	// 2. Try CWD mapping
	cwd, _ := os.Getwd()
	if node := m.mapCWDToNode(cwd); node != nil {
		return &TargetResolution{
			Node:   node,
			Source: SourceCWD,
		}, nil
	}
	
	// 3. Use stored current (only if outside workspace)
	if m.currentNode != nil && !strings.HasPrefix(cwd, m.reposPath) {
		return &TargetResolution{
			Node:   m.currentNode,
			Source: SourceStored,
		}, nil
	}
	
	// 4. Default to root
	return &TargetResolution{
		Node:   m.rootNode,
		Source: SourceRoot,
	}, nil
}

// UseNode navigates to a node and optionally clones lazy repos
func (m *Manager) UseNode(path string, autoClone bool) (*Node, error) {
	// Resolve path
	node := m.resolvePath(path)
	if node == nil {
		return nil, fmt.Errorf("node not found: %s", path)
	}
	
	// Change working directory
	if err := os.Chdir(node.FullPath); err != nil {
		return nil, fmt.Errorf("changing directory: %w", err)
	}
	
	// Update previous and current
	if m.state != nil && m.currentNode != nil {
		m.state.PreviousNodePath = m.currentNode.Path
	}
	
	m.currentNode = node
	if m.state != nil {
		m.state.CurrentNodePath = node.Path
	}
	
	// Auto-clone lazy repos if requested
	if autoClone {
		cloned := 0
		for i, repo := range node.Repos {
			if repo.Lazy && repo.State == string(RepoStateMissing) {
				if err := m.cloneRepo(&node.Repos[i]); err != nil {
					fmt.Printf("⚠️  Failed to clone %s: %v\n", repo.Name, err)
				} else {
					cloned++
				}
			}
		}
		if cloned > 0 {
			fmt.Printf("🔄 Auto-cloned %d lazy repositories\n", cloned)
		}
	}
	
	// Save state
	return node, m.saveState()
}

// AddRepo adds a repository to the current node
func (m *Manager) AddRepo(repoURL string, options AddOptions) (*RepoConfig, error) {
	if m.currentNode == nil {
		return nil, fmt.Errorf("no current node set")
	}
	
	// Extract name from URL if not provided
	name := options.Name
	if name == "" {
		name = extractRepoName(repoURL)
	}
	
	// Check for duplicates
	for _, repo := range m.currentNode.Repos {
		if repo.Name == name {
			return nil, fmt.Errorf("repository %s already exists", name)
		}
	}
	
	// Create repo config
	repo := RepoConfig{
		URL:   repoURL,
		Path:  filepath.Join(m.currentNode.FullPath, name),
		Name:  name,
		Lazy:  options.Lazy,
		State: string(RepoStateMissing),
	}
	
	// Clone if not lazy
	if !options.Lazy {
		if err := m.cloneRepo(&repo); err != nil {
			return nil, fmt.Errorf("cloning repository: %w", err)
		}
	}
	
	// Add to current node
	m.currentNode.Repos = append(m.currentNode.Repos, repo)
	
	// Create child node for this repo
	childNode := &Node{
		ID:       fmt.Sprintf("%s/%s", m.currentNode.ID, name),
		Name:     name,
		Path:     filepath.Join(m.currentNode.Path, name),
		FullPath: repo.Path,
		Parent:   m.currentNode,
		Children: make(map[string]*Node),
		Meta: NodeMeta{
			Type:      string(NodeTypePersistent),
			CreatedAt: time.Now(),
		},
	}
	
	m.currentNode.Children[name] = childNode
	
	// Save state
	if err := m.saveState(); err != nil {
		return nil, fmt.Errorf("saving state: %w", err)
	}
	
	return &repo, nil
}

// RemoveRepo removes a repository from the current node
func (m *Manager) RemoveRepo(name string) error {
	if m.currentNode == nil {
		return fmt.Errorf("no current node set")
	}
	
	// Find and remove repo
	found := false
	newRepos := []RepoConfig{}
	for _, repo := range m.currentNode.Repos {
		if repo.Name != name {
			newRepos = append(newRepos, repo)
		} else {
			found = true
			// Remove from filesystem
			if repo.State == string(RepoStateCloned) {
				if err := os.RemoveAll(repo.Path); err != nil {
					return fmt.Errorf("removing repository directory: %w", err)
				}
			}
		}
	}
	
	if !found {
		return fmt.Errorf("repository %s not found", name)
	}
	
	m.currentNode.Repos = newRepos
	
	// Remove child node
	delete(m.currentNode.Children, name)
	
	// Save state
	return m.saveState()
}

// CloneLazy clones lazy repositories at the current node
func (m *Manager) CloneLazy(recursive bool) error {
	if m.currentNode == nil {
		return fmt.Errorf("no current node set")
	}
	
	return m.cloneLazyInNode(m.currentNode, recursive)
}

// Helper functions

func (m *Manager) cloneLazyInNode(node *Node, recursive bool) error {
	// Clone lazy repos in this node
	for i, repo := range node.Repos {
		if repo.Lazy && repo.State == string(RepoStateMissing) {
			fmt.Printf("🔄 Cloning %s...\n", repo.Name)
			if err := m.cloneRepo(&node.Repos[i]); err != nil {
				return fmt.Errorf("cloning %s: %w", repo.Name, err)
			}
		}
	}
	
	// Recursively clone in children if requested
	if recursive {
		for _, child := range node.Children {
			if err := m.cloneLazyInNode(child, true); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (m *Manager) cloneRepo(repo *RepoConfig) error {
	// Create parent directory if needed
	parentDir := filepath.Dir(repo.Path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("creating parent directory: %w", err)
	}
	
	// Clone repository
	if err := m.gitCmd.Clone(repo.URL, repo.Path); err != nil {
		return err
	}
	
	repo.State = string(RepoStateCloned)
	return nil
}

func (m *Manager) findNodeByPath(path string) *Node {
	// Handle special paths
	switch path {
	case "", ".", "/", "~":
		return m.rootNode
	case "-":
		if m.state != nil && m.state.PreviousNodePath != "" {
			return m.findNodeByPath(m.state.PreviousNodePath)
		}
		return m.currentNode
	}
	
	// Clean the path
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return m.rootNode
	}
	
	// Traverse the tree
	parts := strings.Split(path, "/")
	current := m.rootNode
	
	for _, part := range parts {
		if part == ".." {
			if current.Parent != nil {
				current = current.Parent
			}
		} else if part != "." && part != "" {
			if child, exists := current.Children[part]; exists {
				current = child
			} else {
				return nil
			}
		}
	}
	
	return current
}

func (m *Manager) resolvePath(path string) *Node {
	// Handle relative paths
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		// Start from current node
		current := m.currentNode
		if current == nil {
			current = m.rootNode
		}
		
		parts := strings.Split(path, "/")
		for _, part := range parts {
			switch part {
			case ".", "":
				// Stay at current
			case "..":
				if current.Parent != nil {
					current = current.Parent
				}
			default:
				if child, exists := current.Children[part]; exists {
					current = child
				} else {
					return nil
				}
			}
		}
		return current
	}
	
	// Absolute or special paths
	return m.findNodeByPath(path)
}

func (m *Manager) mapCWDToNode(cwd string) *Node {
	// Check if CWD is within workspace
	if !strings.HasPrefix(cwd, m.reposPath) {
		return nil
	}
	
	// Get relative path from workspace root
	relPath, err := filepath.Rel(m.reposPath, cwd)
	if err != nil {
		return nil
	}
	
	// Find corresponding node
	if relPath == "." {
		return m.rootNode
	}
	
	return m.findNodeByPath(relPath)
}

func (m *Manager) loadState() (*TreeState, error) {
	if _, err := os.Stat(m.statePath); os.IsNotExist(err) {
		return &TreeState{
			Nodes:       make(map[string]*Node),
			LastUpdated: time.Now(),
		}, nil
	}
	
	data, err := os.ReadFile(m.statePath)
	if err != nil {
		return nil, err
	}
	
	var state TreeState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	
	return &state, nil
}

func (m *Manager) saveState() error {
	if m.state == nil {
		m.state = &TreeState{
			Nodes: make(map[string]*Node),
		}
	}
	
	// Update state
	m.state.LastUpdated = time.Now()
	if m.currentNode != nil {
		m.state.CurrentNodePath = m.currentNode.Path
	}
	
	// Collect all nodes
	m.state.Nodes = make(map[string]*Node)
	m.collectNodes(m.rootNode, m.state.Nodes)
	
	// Save to file
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(m.statePath, data, 0644)
}

func (m *Manager) collectNodes(node *Node, collection map[string]*Node) {
	collection[node.Path] = node
	for _, child := range node.Children {
		m.collectNodes(child, collection)
	}
}

func (m *Manager) reconstructTree() error {
	// Find root node
	rootNode, exists := m.state.Nodes["/"]
	if !exists {
		// Create default root if not found
		m.rootNode = &Node{
			ID:       "root",
			Name:     "workspace",
			Path:     "/",
			FullPath: m.reposPath,
			Children: make(map[string]*Node),
			Meta: NodeMeta{
				Type:      string(NodeTypePersistent),
				CreatedAt: time.Now(),
			},
		}
		return nil
	}
	
	m.rootNode = rootNode
	m.rootNode.FullPath = m.reposPath
	
	// Reconstruct parent-child relationships
	for path, node := range m.state.Nodes {
		if path == "/" {
			continue
		}
		
		// Find parent path
		parentPath := filepath.Dir(path)
		if parentPath == "." {
			parentPath = "/"
		}
		
		parent, exists := m.state.Nodes[parentPath]
		if exists {
			node.Parent = parent
			if parent.Children == nil {
				parent.Children = make(map[string]*Node)
			}
			parent.Children[node.Name] = node
		}
		
		// Set full path
		node.FullPath = filepath.Join(m.reposPath, strings.TrimPrefix(path, "/"))
	}
	
	return nil
}

func extractRepoName(url string) string {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")
	
	// Get last path component
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	
	return url
}

// GetCurrentNode returns the current node
func (m *Manager) GetCurrentNode() *Node {
	return m.currentNode
}

// GetRootNode returns the root node
func (m *Manager) GetRootNode() *Node {
	return m.rootNode
}