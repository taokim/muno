package manager

import (
	"fmt"
	"io"
	
	"github.com/taokim/muno/internal/adapters"
	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/tree"
)

// NewRealFileSystem creates a real filesystem implementation
func NewRealFileSystem() interfaces.FileSystem {
	return adapters.NewRealFileSystem()
}

// NewRealGit creates a real git implementation
func NewRealGit() interfaces.GitInterface {
	return adapters.NewRealGit()
}

// NewRealCommandExecutor creates a real command executor
func NewRealCommandExecutor() interfaces.CommandExecutor {
	return adapters.NewRealCommandExecutor()
}

// NewRealOutput creates a real output implementation
func NewRealOutput(stdout, stderr io.Writer) interfaces.Output {
	return adapters.NewRealOutput(stdout, stderr)
}

// For backward compatibility
func createSharedMemory(path string) error {
	fs := NewRealFileSystem()
	content := `# Shared Memory

This file is for shared context across Claude Code sessions.

## Project Overview


## Key Decisions


## Current Tasks


## Notes

`
	return fs.WriteFile(path, []byte(content), 0644)
}

// NewFileSystemAdapter creates a filesystem adapter wrapper
func NewFileSystemAdapter(fs interfaces.FileSystem) interfaces.FileSystemProvider {
	return adapters.NewFileSystemAdapter()
}

// NewConfigAdapter creates a config adapter wrapper
func NewConfigAdapter(cfg interface{}) interfaces.ConfigProvider {
	return adapters.NewConfigAdapter()
}

// NewGitAdapter creates a git adapter wrapper that adapts git.Interface to GitProvider
func NewGitAdapter(gitCmd interfaces.GitInterface) interfaces.GitProvider {
	if gitCmd == nil {
		// Create a real git implementation
		gitCmd = adapters.NewRealGit()
	}
	return &gitProviderAdapter{git: gitCmd}
}

// NewUIAdapter creates a UI adapter wrapper
func NewUIAdapter(output interfaces.Output) interfaces.UIProvider {
	return adapters.NewUIAdapter()
}

// NewTreeAdapter creates a tree adapter wrapper
func NewTreeAdapter(treeMgr *tree.Manager) interfaces.TreeProvider {
	return &treeProviderAdapter{mgr: treeMgr}
}

// gitProviderAdapter adapts interfaces.GitInterface to GitProvider
type gitProviderAdapter struct {
	git interfaces.GitInterface
}

func (g *gitProviderAdapter) Clone(url, path string, opts interfaces.CloneOptions) error {
	return g.git.Clone(url, path)
}

func (g *gitProviderAdapter) Pull(path string, opts interfaces.PullOptions) error {
	return g.git.Pull(path)
}

func (g *gitProviderAdapter) Push(path string, opts interfaces.PushOptions) error {
	return g.git.Push(path)
}

func (g *gitProviderAdapter) Commit(path, message string, opts interfaces.CommitOptions) error {
	return g.git.Commit(path, message)
}

func (g *gitProviderAdapter) Status(path string) (*interfaces.GitStatus, error) {
	// Parse status string to create GitStatus
	// This is a simplified implementation
	if g.git == nil {
		return &interfaces.GitStatus{
			Branch:       "main",
			IsClean:      true,
			HasUntracked: false,
			HasStaged:    false,
			HasModified:  false,
			Files:        []interfaces.GitFileStatus{},
			Ahead:        0,
			Behind:       0,
		}, nil
	}
	
	hasChanges, _ := g.git.HasChanges(path)
	branch, _ := g.git.CurrentBranch(path)
	
	return &interfaces.GitStatus{
		Branch:       branch,
		IsClean:      !hasChanges,
		HasUntracked: false, // Would need to parse from status
		HasStaged:    false, // Would need to parse from status
		HasModified:  hasChanges,
		Files:        []interfaces.GitFileStatus{},
		Ahead:        0, // Would need to parse from status
		Behind:       0, // Would need to parse from status
	}, nil
}

func (g *gitProviderAdapter) Add(path string, files []string) error {
	// Adapt the Add method signature
	for _, file := range files {
		if err := g.git.Add(path, file); err != nil {
			return err
		}
	}
	return nil
}

func (g *gitProviderAdapter) GetRemoteURL(path string) (string, error) {
	return g.git.RemoteURL(path)
}

func (g *gitProviderAdapter) Branch(path string) (string, error) {
	return g.git.CurrentBranch(path)
}

func (g *gitProviderAdapter) Checkout(path string, branch string) error {
	return g.git.Checkout(path, branch)
}

func (g *gitProviderAdapter) Fetch(path string, options interfaces.FetchOptions) error {
	return g.git.Fetch(path)
}

func (g *gitProviderAdapter) Remove(path string, files []string) error {
	// GitInterface doesn't have Remove, so we'll return an error
	return fmt.Errorf("remove not implemented")
}

func (g *gitProviderAdapter) SetRemoteURL(path string, url string) error {
	// GitInterface doesn't have SetRemoteURL, so we'll return an error
	return fmt.Errorf("set remote URL not implemented")
}

// treeProviderAdapter adapts tree.Manager to TreeProvider
type treeProviderAdapter struct {
	mgr *tree.Manager
}

func (t *treeProviderAdapter) Load(config interface{}) error {
	// Tree manager doesn't need configuration loading
	return nil
}

func (t *treeProviderAdapter) Navigate(path string) error {
	return t.mgr.UseNode(path)
}

func (t *treeProviderAdapter) GetCurrent() (interfaces.NodeInfo, error) {
	currentPath := t.mgr.GetCurrentPath()
	node := t.mgr.GetNode(currentPath)
	if node == nil {
		return interfaces.NodeInfo{}, fmt.Errorf("current node not found")
	}
	return t.nodeToNodeInfo(currentPath, node), nil
}

func (t *treeProviderAdapter) GetTree() (interfaces.NodeInfo, error) {
	rootNode := t.mgr.GetNode("/")
	if rootNode == nil {
		return interfaces.NodeInfo{}, fmt.Errorf("root node not found")
	}
	return t.buildTreeRecursive("/", rootNode), nil
}

func (t *treeProviderAdapter) GetNode(path string) (interfaces.NodeInfo, error) {
	node := t.mgr.GetNode(path)
	if node == nil {
		return interfaces.NodeInfo{}, fmt.Errorf("node not found: %s", path)
	}
	return t.nodeToNodeInfo(path, node), nil
}

func (t *treeProviderAdapter) AddNode(parentPath string, node interfaces.NodeInfo) error {
	return t.mgr.AddRepo(parentPath, node.Name, node.Repository, node.IsLazy)
}

func (t *treeProviderAdapter) RemoveNode(path string) error {
	return t.mgr.RemoveNode(path)
}

func (t *treeProviderAdapter) UpdateNode(path string, node interfaces.NodeInfo) error {
	// Get the underlying tree node
	treeNode := t.mgr.GetNode(path)
	if treeNode == nil {
		return fmt.Errorf("node not found: %s", path)
	}
	
	// Update the tree node state based on NodeInfo
	if node.IsCloned {
		treeNode.State = tree.RepoStateCloned
	} else {
		treeNode.State = tree.RepoStateMissing
	}
	treeNode.Lazy = node.IsLazy
	
	// Save the state
	return t.mgr.SaveState()
}

func (t *treeProviderAdapter) ListChildren(path string) ([]interfaces.NodeInfo, error) {
	children, err := t.mgr.ListChildren(path)
	if err != nil {
		return nil, err
	}
	
	var result []interfaces.NodeInfo
	for _, child := range children {
		childPath := path + "/" + child.Name
		if path == "/" {
			childPath = "/" + child.Name
		}
		result = append(result, t.nodeToNodeInfo(childPath, child))
	}
	return result, nil
}

func (t *treeProviderAdapter) GetPath() string {
	return t.mgr.GetCurrentPath()
}

func (t *treeProviderAdapter) SetPath(path string) error {
	return t.mgr.UseNode(path)
}

func (t *treeProviderAdapter) GetState() (interfaces.TreeState, error) {
	// Build tree state from manager state
	rootNode := t.mgr.GetNode("/")
	if rootNode == nil {
		return interfaces.TreeState{}, fmt.Errorf("no tree state")
	}
	
	nodes := make(map[string]interfaces.NodeInfo)
	t.collectNodesRecursive("/", rootNode, nodes)
	
	return interfaces.TreeState{
		CurrentPath: t.mgr.GetCurrentPath(),
		Nodes:       nodes,
		History:     []string{}, // Tree manager doesn't track history
	}, nil
}

func (t *treeProviderAdapter) SetState(state interfaces.TreeState) error {
	// Tree manager manages its own state
	return fmt.Errorf("SetState not supported")
}

// Helper methods for treeProviderAdapter
func (t *treeProviderAdapter) nodeToNodeInfo(path string, node *tree.TreeNode) interfaces.NodeInfo {
	info := interfaces.NodeInfo{
		Name:       node.Name,
		Path:       path,
		Repository: node.URL,
		IsLazy:     node.Lazy,
		IsCloned:   node.State == tree.RepoStateCloned,
		HasChanges: false, // Would need git status check
		Children:   []interfaces.NodeInfo{},
		Parent:     nil,
	}
	
	// Add children
	for _, childName := range node.Children {
		childPath := path + "/" + childName
		if path == "/" {
			childPath = "/" + childName
		}
		childNode := t.mgr.GetNode(childPath)
		if childNode != nil {
			childInfo := t.nodeToNodeInfo(childPath, childNode)
			info.Children = append(info.Children, childInfo)
		}
	}
	
	return info
}

func (t *treeProviderAdapter) buildTreeRecursive(path string, node *tree.TreeNode) interfaces.NodeInfo {
	return t.nodeToNodeInfo(path, node)
}

func (t *treeProviderAdapter) collectNodesRecursive(path string, node *tree.TreeNode, nodes map[string]interfaces.NodeInfo) {
	nodes[path] = t.nodeToNodeInfo(path, node)
	
	for _, childName := range node.Children {
		childPath := path + "/" + childName
		if path == "/" {
			childPath = "/" + childName
		}
		childNode := t.mgr.GetNode(childPath)
		if childNode != nil {
			t.collectNodesRecursive(childPath, childNode, nodes)
		}
	}
}

// gitInterfaceAdapter adapts interfaces.GitInterface to git.Interface
type gitInterfaceAdapter struct {
	git interfaces.GitInterface
}

func (g *gitInterfaceAdapter) Clone(url, path string) error {
	return g.git.Clone(url, path)
}

func (g *gitInterfaceAdapter) Pull(path string) error {
	return g.git.Pull(path)
}

func (g *gitInterfaceAdapter) Status(path string) (string, error) {
	return g.git.Status(path)
}

func (g *gitInterfaceAdapter) Commit(path, message string) error {
	return g.git.Commit(path, message)
}

func (g *gitInterfaceAdapter) Push(path string) error {
	return g.git.Push(path)
}

func (g *gitInterfaceAdapter) Add(path, pattern string) error {
	return g.git.Add(path, pattern)
}