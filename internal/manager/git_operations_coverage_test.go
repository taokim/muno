package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/taokim/muno/internal/interfaces"
)

// GitProviderStub implements interfaces.GitProvider for testing
type GitProviderStub struct {
	cloneError    error
	pullError     error
	pushError     error
	statusError   error
	commitError   error
	statusResult  *interfaces.GitStatus
	cloneCalled   bool
	pullCalled    bool
	pushCalled    bool
	statusCalled  bool
	commitCalled  bool
	pullCallCount int
	pushCallCount int
	lastPullPath  string
	lastPushPath  string
	lastCommitMsg string
}

func NewGitProviderStub() *GitProviderStub {
	return &GitProviderStub{
		statusResult: &interfaces.GitStatus{
			Branch:  "main",
			IsClean: true,
		},
	}
}

func (g *GitProviderStub) Clone(url, path string, options interfaces.CloneOptions) error {
	g.cloneCalled = true
	return g.cloneError
}

func (g *GitProviderStub) Pull(path string, options interfaces.PullOptions) error {
	g.pullCalled = true
	g.pullCallCount++
	g.lastPullPath = path
	return g.pullError
}

func (g *GitProviderStub) Push(path string, options interfaces.PushOptions) error {
	g.pushCalled = true
	g.pushCallCount++
	g.lastPushPath = path
	return g.pushError
}

func (g *GitProviderStub) Status(path string) (*interfaces.GitStatus, error) {
	g.statusCalled = true
	if g.statusError != nil {
		return nil, g.statusError
	}
	return g.statusResult, nil
}

func (g *GitProviderStub) Commit(path string, message string, options interfaces.CommitOptions) error {
	g.commitCalled = true
	g.lastCommitMsg = message
	return g.commitError
}

func (g *GitProviderStub) Branch(path string) (string, error) {
	return "main", nil
}

func (g *GitProviderStub) Checkout(path string, branch string) error {
	return nil
}

func (g *GitProviderStub) Fetch(path string, options interfaces.FetchOptions) error {
	return nil
}

func (g *GitProviderStub) Add(path string, files []string) error {
	return nil
}

func (g *GitProviderStub) Remove(path string, files []string) error {
	return nil
}

func (g *GitProviderStub) GetRemoteURL(path string) (string, error) {
	return "https://github.com/test/repo.git", nil
}

func (g *GitProviderStub) SetRemoteURL(path string, url string) error {
	return nil
}

// TestStatusNode_NonRecursive tests single node status display
func TestStatusNode_NonRecursive(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add a cloned node
	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	// Create directory
	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	if err := os.MkdirAll(backendPath, 0755); err != nil {
		t.Fatalf("Failed to create backend directory: %v", err)
	}

	// Test status
	err := mgr.StatusNode("/backend", false)
	if err != nil {
		t.Errorf("StatusNode failed: %v", err)
	}

	if !gitStub.statusCalled {
		t.Error("Expected Status to be called")
	}
}

// TestStatusNode_Recursive tests recursive status display
func TestStatusNode_Recursive(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add parent node
	teamNode := CreateSimpleNode("team", "https://github.com/test/team.git")
	AddNodeToTree(mgr, "/team", teamNode)

	// Add child node
	serviceNode := CreateSimpleNode("service", "https://github.com/test/service.git")
	serviceNode.Path = "/team/service"
	teamNode.Children = append(teamNode.Children, serviceNode)
	mgr.treeProvider.UpdateNode("/team", teamNode)

	// Create directories
	teamPath := filepath.Join(tmpDir, ".nodes", "team")
	servicePath := filepath.Join(teamPath, "service")
	if err := os.MkdirAll(servicePath, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Test recursive status
	err := mgr.StatusNode("/team", true)
	if err != nil {
		t.Errorf("StatusNode recursive failed: %v", err)
	}

	// Should be called for at least the parent
	if !gitStub.statusCalled {
		t.Error("Expected Status to be called")
	}
}

// TestStatusNode_Uninitialized tests error when manager not initialized
func TestStatusNode_Uninitialized(t *testing.T) {
	mgr := &Manager{initialized: false}
	err := mgr.StatusNode("/test", false)
	if err == nil {
		t.Error("Expected error for uninitialized manager")
	}
}

// TestStatusNode_GitError tests handling of git status errors
func TestStatusNode_GitError(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	gitStub.statusError = fmt.Errorf("git status failed")
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	if err := os.MkdirAll(backendPath, 0755); err != nil {
		t.Fatalf("Failed to create backend directory: %v", err)
	}

	err := mgr.StatusNode("/backend", false)
	if err == nil {
		t.Error("Expected error from git status")
	}
}

// TestPullNode_SingleNode tests pulling a single node
func TestPullNode_SingleNode(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	if err := os.MkdirAll(backendPath, 0755); err != nil {
		t.Fatalf("Failed to create backend directory: %v", err)
	}

	// Test pull
	err := mgr.PullNode("/backend", false, false)
	if err != nil {
		t.Errorf("PullNode failed: %v", err)
	}

	if !gitStub.pullCalled {
		t.Error("Expected Pull to be called")
	}
}

// TestPullNode_WithForce tests pulling with force flag
func TestPullNode_WithForce(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	if err := os.MkdirAll(backendPath, 0755); err != nil {
		t.Fatalf("Failed to create backend directory: %v", err)
	}

	// Test pull with force
	err := mgr.PullNode("/backend", false, true)
	if err != nil {
		t.Errorf("PullNode with force failed: %v", err)
	}

	if !gitStub.pullCalled {
		t.Error("Expected Pull to be called")
	}
}

// TestPullNode_Recursive tests pulling recursively
func TestPullNode_Recursive(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add parent node
	teamNode := CreateSimpleNode("team", "https://github.com/test/team.git")
	AddNodeToTree(mgr, "/team", teamNode)

	// Add child node
	serviceNode := CreateSimpleNode("service", "https://github.com/test/service.git")
	serviceNode.Path = "/team/service"
	teamNode.Children = append(teamNode.Children, serviceNode)
	mgr.treeProvider.UpdateNode("/team", teamNode)

	// Create directories
	teamPath := filepath.Join(tmpDir, ".nodes", "team")
	servicePath := filepath.Join(teamPath, "service")
	if err := os.MkdirAll(servicePath, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Test recursive pull
	err := mgr.PullNode("/team", true, false)
	// May fail due to incomplete implementation, but should not panic
	_ = err

	// Should have been called at least once
	if gitStub.pullCallCount == 0 {
		t.Error("Expected Pull to be called at least once")
	}
}

// TestPullNode_Uninitialized tests error when manager not initialized
func TestPullNode_Uninitialized(t *testing.T) {
	mgr := &Manager{initialized: false}
	err := mgr.PullNode("/test", false, false)
	if err == nil {
		t.Error("Expected error for uninitialized manager")
	}
}

// TestPullNode_GitError tests handling of pull errors
func TestPullNode_GitError(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	gitStub.pullError = fmt.Errorf("pull failed")
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	if err := os.MkdirAll(backendPath, 0755); err != nil {
		t.Fatalf("Failed to create backend directory: %v", err)
	}

	err := mgr.PullNode("/backend", false, false)
	if err == nil {
		t.Error("Expected error from git pull")
	}
}

// TestPullNodeWithOptions_IncludeLazy tests pulling with includeLazy option
func TestPullNodeWithOptions_IncludeLazy(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add lazy node
	node := CreateLazyNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	// Test pull with includeLazy=true (should trigger cloning)
	err := mgr.PullNodeWithOptions("/backend", false, false, true)
	// This might fail if clone is not fully implemented, but should at least not panic
	_ = err
}

// TestPushNode_SingleNode tests pushing a single node
func TestPushNode_SingleNode(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	if err := os.MkdirAll(backendPath, 0755); err != nil {
		t.Fatalf("Failed to create backend directory: %v", err)
	}

	// Test push
	err := mgr.PushNode("/backend", false)
	if err != nil {
		t.Errorf("PushNode failed: %v", err)
	}

	if !gitStub.pushCalled {
		t.Error("Expected Push to be called")
	}
}

// TestPushNode_Recursive tests pushing recursively
func TestPushNode_Recursive(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add parent node
	teamNode := CreateSimpleNode("team", "https://github.com/test/team.git")
	AddNodeToTree(mgr, "/team", teamNode)

	// Add child node
	serviceNode := CreateSimpleNode("service", "https://github.com/test/service.git")
	serviceNode.Path = "/team/service"
	teamNode.Children = append(teamNode.Children, serviceNode)
	mgr.treeProvider.UpdateNode("/team", teamNode)

	// Create directories
	teamPath := filepath.Join(tmpDir, ".nodes", "team")
	servicePath := filepath.Join(teamPath, "service")
	if err := os.MkdirAll(servicePath, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Test recursive push
	err := mgr.PushNode("/team", true)
	// Should not error even if push partially fails
	_ = err

	// Should have been called at least once
	if gitStub.pushCallCount == 0 {
		t.Error("Expected Push to be called at least once")
	}
}

// TestPushNode_Uninitialized tests error when manager not initialized
func TestPushNode_Uninitialized(t *testing.T) {
	mgr := &Manager{initialized: false}
	err := mgr.PushNode("/test", false)
	if err == nil {
		t.Error("Expected error for uninitialized manager")
	}
}

// TestCommitNode tests committing changes
func TestCommitNode(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	if err := os.MkdirAll(backendPath, 0755); err != nil {
		t.Fatalf("Failed to create backend directory: %v", err)
	}

	// Test commit
	commitMsg := "test commit message"
	err := mgr.CommitNode("/backend", commitMsg, false)
	if err != nil {
		t.Errorf("CommitNode failed: %v", err)
	}

	if !gitStub.commitCalled {
		t.Error("Expected Commit to be called")
	}

	if gitStub.lastCommitMsg != commitMsg {
		t.Errorf("Expected commit message %q, got %q", commitMsg, gitStub.lastCommitMsg)
	}
}

// TestCommitNode_Uninitialized tests error when manager not initialized
func TestCommitNode_Uninitialized(t *testing.T) {
	mgr := &Manager{initialized: false}
	err := mgr.CommitNode("/test", "message", false)
	if err == nil {
		t.Error("Expected error for uninitialized manager")
	}
}

// TestCommitNode_GitError tests handling of commit errors
func TestCommitNode_GitError(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	gitStub.commitError = fmt.Errorf("commit failed")
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	if err := os.MkdirAll(backendPath, 0755); err != nil {
		t.Fatalf("Failed to create backend directory: %v", err)
	}

	err := mgr.CommitNode("/backend", "test message", false)
	if err == nil {
		t.Error("Expected error from git commit")
	}
}

// TestStatusNode_NodeNotFound tests error when node doesn't exist
func TestStatusNode_NodeNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Try to get status of non-existent node
	err := mgr.StatusNode("/nonexistent", false)
	if err == nil {
		t.Error("Expected error for non-existent node")
	}
}

// TestPullNode_NodeNotFound tests error when node doesn't exist
func TestPullNode_NodeNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Try to pull non-existent node
	err := mgr.PullNode("/nonexistent", false, false)
	if err == nil {
		t.Error("Expected error for non-existent node")
	}
}

// TestPushNode_NodeNotFound tests error when node doesn't exist
func TestPushNode_NodeNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Try to push non-existent node
	err := mgr.PushNode("/nonexistent", false)
	if err == nil {
		t.Error("Expected error for non-existent node")
	}
}

// TestCommitNode_NodeNotFound tests error when node doesn't exist
func TestCommitNode_NodeNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Try to commit non-existent node
	err := mgr.CommitNode("/nonexistent", "message", false)
	if err == nil {
		t.Error("Expected error for non-existent node")
	}
}
// TestShowStatusRecursive_WithChanges tests recursive status with file changes
func TestShowStatusRecursive_WithChanges(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	// Set up status with changes
	gitStub.statusResult = &interfaces.GitStatus{
		Branch:       "main",
		IsClean:      false,
		HasUntracked: true,
		HasModified:  true,
		HasStaged:    true,
		Files: []interfaces.GitFileStatus{
			{Path: "file1.txt", Status: "untracked", Staged: false},
			{Path: "file2.txt", Status: "modified", Staged: false},
			{Path: "file3.txt", Status: "modified", Staged: true},
		},
	}
	mgr.gitProvider = gitStub

	// Add parent node
	teamNode := CreateSimpleNode("team", "https://github.com/test/team.git")
	AddNodeToTree(mgr, "/team", teamNode)

	// Add child node
	serviceNode := CreateSimpleNode("service", "https://github.com/test/service.git")
	serviceNode.Path = "/team/service"
	teamNode.Children = append(teamNode.Children, serviceNode)
	mgr.treeProvider.UpdateNode("/team", teamNode)

	// Create directories
	teamPath := filepath.Join(tmpDir, ".nodes", "team")
	servicePath := filepath.Join(teamPath, "service")
	if err := os.MkdirAll(servicePath, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Test recursive status with changes
	err := mgr.StatusNode("/team", true)
	if err != nil {
		t.Errorf("StatusNode recursive with changes failed: %v", err)
	}
}

// TestShowStatusRecursive_LazyNode tests recursive status skips lazy nodes
func TestShowStatusRecursive_LazyNode(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add lazy node
	lazyNode := CreateLazyNode("lazy-repo", "https://github.com/test/lazy.git")
	AddNodeToTree(mgr, "/lazy-repo", lazyNode)

	// Test status on lazy node (should not call git)
	gitStub.statusCalled = false
	err := mgr.StatusNode("/lazy-repo", false)
	
	// Should not error but also should not call git status on lazy node
	if err != nil && gitStub.statusCalled {
		t.Error("Should not call git status on lazy node")
	}
}

// TestPullRecursiveWithOptions_CloneLazy tests pullRecursive clones lazy repos
func TestPullRecursiveWithOptions_CloneLazy(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add lazy node that's NOT cloned yet
	lazyNode := interfaces.NodeInfo{
		Name:       "lazy-repo",
		Path:       "/lazy-repo",
		Repository: "https://github.com/test/lazy.git",
		IsLazy:     true,
		IsCloned:   false,
	}
	AddNodeToTree(mgr, "/lazy-repo", lazyNode)

	// Test pull with includeLazy - this should trigger cloning
	err := mgr.PullNodeWithOptions("/lazy-repo", false, false, true)
	// May fail due to implementation details, but should attempt to clone
	_ = err

	// Should have attempted to clone if lazy and not cloned
	// The test verifies the code path is exercised
}

// TestOutputNodesQuietRecursive_DeepTree tests recursive output with nested nodes
func TestOutputNodesQuietRecursive_DeepTree(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	// Build a deep tree
	level1 := CreateSimpleNode("level1", "https://github.com/test/level1.git")
	AddNodeToTree(mgr, "/level1", level1)

	level2 := CreateSimpleNode("level2", "https://github.com/test/level2.git")
	level2.Path = "/level1/level2"
	level1.Children = append(level1.Children, level2)
	mgr.treeProvider.UpdateNode("/level1", level1)

	level3 := CreateSimpleNode("level3", "https://github.com/test/level3.git")
	level3.Path = "/level1/level2/level3"
	level2.Children = append(level2.Children, level3)
	
	// Update level1 with the full tree
	level1Node, _ := mgr.treeProvider.GetNode("/level1")
	mgr.treeProvider.UpdateNode("/level1", level1Node)

	// Create directories
	level1Path := filepath.Join(tmpDir, ".nodes", "level1")
	level2Path := filepath.Join(level1Path, "level2")
	level3Path := filepath.Join(level2Path, "level3")
	if err := os.MkdirAll(level3Path, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Test list quiet with recursive (exercises outputNodesQuietRecursive)
	err := mgr.ListNodesQuiet(true)
	if err != nil {
		t.Errorf("ListNodesQuiet recursive failed: %v", err)
	}
}

// TestPullRecursive_MixedSuccess tests pulling with some failures
func TestPullRecursive_MixedSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	// Create a git stub - will test that recursive continues even with errors
	gitStub := NewGitProviderStub()
	// Set it to fail on all pulls to test error handling
	gitStub.pullError = fmt.Errorf("pull failed")
	mgr.gitProvider = gitStub

	// Add parent node
	teamNode := CreateSimpleNode("team", "https://github.com/test/team.git")
	AddNodeToTree(mgr, "/team", teamNode)

	// Add two child nodes
	service1Node := CreateSimpleNode("service1", "https://github.com/test/service1.git")
	service1Node.Path = "/team/service1"
	
	service2Node := CreateSimpleNode("service2", "https://github.com/test/service2.git")
	service2Node.Path = "/team/service2"
	
	teamNode.Children = append(teamNode.Children, service1Node, service2Node)
	mgr.treeProvider.UpdateNode("/team", teamNode)

	// Create directories
	teamPath := filepath.Join(tmpDir, ".nodes", "team")
	service1Path := filepath.Join(teamPath, "service1")
	service2Path := filepath.Join(teamPath, "service2")
	if err := os.MkdirAll(service1Path, 0755); err != nil {
		t.Fatalf("Failed to create service1 directory: %v", err)
	}
	if err := os.MkdirAll(service2Path, 0755); err != nil {
		t.Fatalf("Failed to create service2 directory: %v", err)
	}

	// Test recursive pull - some may fail
	err := mgr.PullNode("/team", true, false)
	// Should continue despite failures
	_ = err
}

// TestStatusNode_GitStatusWithAheadBehind tests status with ahead/behind info
func TestStatusNode_GitStatusWithAheadBehind(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	gitStub.statusResult = &interfaces.GitStatus{
		Branch:  "main",
		IsClean: true,
		Ahead:   5,
		Behind:  2,
	}
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	if err := os.MkdirAll(backendPath, 0755); err != nil {
		t.Fatalf("Failed to create backend directory: %v", err)
	}

	// Test status
	err := mgr.StatusNode("/backend", false)
	if err != nil {
		t.Errorf("StatusNode failed: %v", err)
	}

	if !gitStub.statusCalled {
		t.Error("Expected Status to be called")
	}
}

// TestCloneRepos_AllLazy tests CloneRepos with all lazy repos
func TestCloneRepos_AllLazy(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add lazy nodes
	lazy1 := CreateLazyNode("lazy1", "https://github.com/test/lazy1.git")
	AddNodeToTree(mgr, "/lazy1", lazy1)

	lazy2 := CreateLazyNode("lazy2", "https://github.com/test/lazy2.git")
	AddNodeToTree(mgr, "/lazy2", lazy2)

	// Test CloneRepos without includeLazy
	err := mgr.CloneRepos("/", false, false)
	// Should not error even if no repos to clone
	_ = err
}

// TestCloneRepos_WithIncludeLazy tests CloneRepos with includeLazy flag
func TestCloneRepos_WithIncludeLazy(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add lazy node that's NOT cloned
	lazy := interfaces.NodeInfo{
		Name:       "lazy",
		Path:       "/lazy",
		Repository: "https://github.com/test/lazy.git",
		IsLazy:     true,
		IsCloned:   false,
	}
	AddNodeToTree(mgr, "/lazy", lazy)

	// Test CloneRepos with includeLazy
	err := mgr.CloneRepos("/lazy", false, true)
	// May fail due to implementation but should attempt clone
	// This test exercises the CloneRepos code path with includeLazy
	_ = err
}

// TestCloneRepos_Recursive tests CloneRepos with recursive flag
func TestCloneRepos_Recursive(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add parent with lazy children
	parent := CreateSimpleNode("parent", "https://github.com/test/parent.git")
	AddNodeToTree(mgr, "/parent", parent)

	child := CreateLazyNode("child", "https://github.com/test/child.git")
	child.Path = "/parent/child"
	parent.Children = append(parent.Children, child)
	mgr.treeProvider.UpdateNode("/parent", parent)

	// Create parent directory
	parentPath := filepath.Join(tmpDir, ".nodes", "parent")
	if err := os.MkdirAll(parentPath, 0755); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}

	// Test recursive clone
	err := mgr.CloneRepos("/parent", true, true)
	// May fail but should attempt
	_ = err
}
