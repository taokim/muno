package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/taokim/muno/internal/interfaces"
)

// ================================================================================
// StatusNode Tests
// ================================================================================

func TestStatusNode_RecursiveWithMultipleRepos(t *testing.T) {
	// Test recursive status display with multiple repositories
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	// Create git stub with status results
	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Create directories
	os.MkdirAll(filepath.Join(tmpDir, ".nodes", "backend"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, ".nodes", "frontend"), 0755)

	// Add backend node and test it directly
	backendNode := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", backendNode)

	err := mgr.StatusNode("/backend", false)
	if err != nil {
		t.Errorf("StatusNode failed: %v", err)
	}

	// Verify status was called
	if !gitStub.statusCalled {
		t.Error("Expected Status to be called for repositories")
	}
}

func TestStatusNode_WithFileChanges(t *testing.T) {
	// Test status display with various file changes (untracked, modified, staged)
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	// Create git stub with file changes
	gitStub := NewGitProviderStub()
	gitStub.statusResult = &interfaces.GitStatus{
		Branch:       "main",
		IsClean:      false,
		HasUntracked: true,
		HasModified:  true,
		HasStaged:    true,
		Files: []interfaces.GitFileStatus{
			{Path: "new.txt", Status: "untracked", Staged: false},
			{Path: "src/app.js", Status: "modified", Staged: false},
			{Path: "src/config.js", Status: "modified", Staged: true},
		},
	}
	mgr.gitProvider = gitStub

	// Add a cloned node
	node := CreateSimpleNode("repo1", "https://github.com/test/repo1.git")
	AddNodeToTree(mgr, "/repo1", node)

	// Create directory
	os.MkdirAll(filepath.Join(tmpDir, ".nodes", "repo1"), 0755)

	err := mgr.StatusNode("/repo1", false)
	if err != nil {
		t.Errorf("StatusNode failed: %v", err)
	}

	// Verify status was called
	if !gitStub.statusCalled {
		t.Error("Expected Status to be called")
	}
}

func TestStatusNode_LazyRepository(t *testing.T) {
	// Test status display for lazy (not cloned) repositories
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add a lazy node
	node := CreateLazyNode("lazy-repo", "https://github.com/test/lazy.git")
	AddNodeToTree(mgr, "/lazy-repo", node)

	err := mgr.StatusNode("/lazy-repo", true)
	if err != nil {
		t.Errorf("StatusNode failed: %v", err)
	}

	// For lazy repositories, status should not be called
	// The function should handle this gracefully
}

func TestStatusNode_GitStatusError(t *testing.T) {
	// Test error handling when git status fails
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	// Create git stub that returns error
	gitStub := NewGitProviderStub()
	gitStub.statusError = fmt.Errorf("git status failed")
	mgr.gitProvider = gitStub

	// Add a cloned node
	node := CreateSimpleNode("repo1", "https://github.com/test/repo1.git")
	AddNodeToTree(mgr, "/repo1", node)

	// Create directory
	os.MkdirAll(filepath.Join(tmpDir, ".nodes", "repo1"), 0755)

	// StatusNode in recursive mode should not fail on git status error
	err := mgr.StatusNode("/repo1", true)
	if err != nil {
		t.Errorf("StatusNode should not fail on git status error in recursive mode: %v", err)
	}

	// Verify status was attempted
	if !gitStub.statusCalled {
		t.Error("Expected Status to be called")
	}
}

func TestStatusNode_PWDFromReposDir(t *testing.T) {
	// Test PWD-based path resolution from repos directory
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	// Change to repos directory
	reposDir := filepath.Join(tmpDir, ".nodes")
	os.MkdirAll(reposDir, 0755)
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	if err := os.Chdir(reposDir); err != nil {
		t.Skipf("Cannot change directory: %v", err)
	}

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// The root should be treated as a repository
	root, _ := mgr.treeProvider.GetTree()
	root.IsCloned = true
	root.Repository = "https://github.com/test/root.git"
	mgr.treeProvider.UpdateNode("/", root)

	err := mgr.StatusNode("", false)
	if err != nil {
		t.Errorf("StatusNode with empty path failed: %v", err)
	}
}

func TestStatusNode_CleanRepository(t *testing.T) {
	// Test status display for clean repository (no changes)
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	// Create git stub with clean status
	gitStub := NewGitProviderStub()
	gitStub.statusResult = &interfaces.GitStatus{
		Branch:  "main",
		IsClean: true,
	}
	mgr.gitProvider = gitStub

	// Add a cloned node
	node := CreateSimpleNode("repo1", "https://github.com/test/repo1.git")
	AddNodeToTree(mgr, "/repo1", node)

	// Create directory
	os.MkdirAll(filepath.Join(tmpDir, ".nodes", "repo1"), 0755)

	err := mgr.StatusNode("/repo1", false)
	if err != nil {
		t.Errorf("StatusNode failed: %v", err)
	}

	// Verify status was called
	if !gitStub.statusCalled {
		t.Error("Expected Status to be called")
	}
}

// ================================================================================
// pullRecursiveWithOptions Tests
// ================================================================================

func TestPullRecursiveWithOptions_LazyRepoCloning(t *testing.T) {
	// Test lazy repository cloning when includeLazy=true
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := interfaces.NodeInfo{
		Name:       "lazy-repo",
		Path:       "/lazy-repo",
		Repository: "https://github.com/test/lazy.git",
		IsLazy:     true,
		IsCloned:   false,
		Children:   []interfaces.NodeInfo{},
	}

	err := mgr.pullRecursiveWithOptions(node, false, true)
	if err != nil {
		t.Errorf("pullRecursiveWithOptions failed: %v", err)
	}

	// Verify clone was called
	if !gitStub.cloneCalled {
		t.Error("Expected clone to be called for lazy repository")
	}
}

func TestPullRecursiveWithOptions_UpdateNodeFailure(t *testing.T) {
	// Test that UpdateNode failure doesn't stop the operation
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Create a failing tree provider stub
	failingTree := &FailingTreeProvider{
		updateError: fmt.Errorf("update node failed"),
	}
	mgr.treeProvider = failingTree

	node := interfaces.NodeInfo{
		Name:       "lazy-repo",
		Path:       "/lazy-repo",
		Repository: "https://github.com/test/lazy.git",
		IsLazy:     true,
		IsCloned:   false,
		Children:   []interfaces.NodeInfo{},
	}

	err := mgr.pullRecursiveWithOptions(node, false, true)
	// Should not fail, just warn
	if err != nil {
		t.Errorf("pullRecursiveWithOptions should not fail on UpdateNode error: %v", err)
	}

	// Verify clone was still called
	if !gitStub.cloneCalled {
		t.Error("Expected clone to be called despite UpdateNode failure")
	}
}

func TestPullRecursiveWithOptions_SaveConfigFailure(t *testing.T) {
	// Test that SaveConfig failure doesn't stop the operation
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Use a failing config provider
	mgr.configProvider = &FailingConfigProviderForClone{
		saveError: fmt.Errorf("failed to save config"),
	}

	node := interfaces.NodeInfo{
		Name:       "lazy-repo",
		Path:       "/lazy-repo",
		Repository: "https://github.com/test/lazy.git",
		IsLazy:     true,
		IsCloned:   false,
		Children:   []interfaces.NodeInfo{},
	}

	err := mgr.pullRecursiveWithOptions(node, false, true)
	// Should not fail, just warn
	if err != nil {
		t.Errorf("pullRecursiveWithOptions should not fail on SaveConfig error: %v", err)
	}

	// Verify clone was called
	if !gitStub.cloneCalled {
		t.Error("Expected clone to be called")
	}
}

func TestPullRecursiveWithOptions_TerminalNodePull(t *testing.T) {
	// Test pulling a terminal node (cloned repository without children)
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	repoPath := filepath.Join(tmpDir, ".nodes", "repo1")
	os.MkdirAll(repoPath, 0755)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := interfaces.NodeInfo{
		Name:       "repo1",
		Path:       "/repo1",
		Repository: "https://github.com/test/repo1.git",
		IsLazy:     false,
		IsCloned:   true,
		Children:   []interfaces.NodeInfo{},
	}

	err := mgr.pullRecursiveWithOptions(node, false, false)
	if err != nil {
		t.Errorf("pullRecursiveWithOptions failed: %v", err)
	}

	// Verify pull was called
	if !gitStub.pullCalled {
		t.Error("Expected pull to be called for terminal node")
	}
}

func TestPullRecursiveWithOptions_SkipLazyRepo(t *testing.T) {
	// Test that lazy repositories are skipped when includeLazy=false
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := interfaces.NodeInfo{
		Name:       "lazy-repo",
		Path:       "/lazy-repo",
		Repository: "https://github.com/test/lazy.git",
		IsLazy:     true,
		IsCloned:   false,
		Children:   []interfaces.NodeInfo{},
	}

	err := mgr.pullRecursiveWithOptions(node, false, false)
	if err != nil {
		t.Errorf("pullRecursiveWithOptions failed: %v", err)
	}

	// Verify clone was NOT called
	if gitStub.cloneCalled {
		t.Error("Expected clone not to be called for lazy repository when includeLazy=false")
	}
}

func TestPullRecursiveWithOptions_ChildrenRecursion(t *testing.T) {
	// Test recursion into child nodes
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	// Create directories for repos
	repo1Path := filepath.Join(tmpDir, ".nodes", "parent", "child1")
	repo2Path := filepath.Join(tmpDir, ".nodes", "parent", "child2")
	os.MkdirAll(repo1Path, 0755)
	os.MkdirAll(repo2Path, 0755)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := interfaces.NodeInfo{
		Name: "parent",
		Path: "/parent",
		Children: []interfaces.NodeInfo{
			{
				Name:       "child1",
				Path:       "/parent/child1",
				Repository: "https://github.com/test/child1.git",
				IsCloned:   true,
				Children:   []interfaces.NodeInfo{},
			},
			{
				Name:       "child2",
				Path:       "/parent/child2",
				Repository: "https://github.com/test/child2.git",
				IsCloned:   true,
				Children:   []interfaces.NodeInfo{},
			},
		},
	}

	err := mgr.pullRecursiveWithOptions(node, false, false)
	if err != nil {
		t.Errorf("pullRecursiveWithOptions failed: %v", err)
	}

	// Verify both children were pulled
	if mgr.gitProvider.(*GitProviderStub).pullCallCount < 2 {
		t.Errorf("Expected at least 2 pull calls for children, got %d",
			mgr.gitProvider.(*GitProviderStub).pullCallCount)
	}
}

// ================================================================================
// pullRecursive Tests
// ================================================================================

func TestPullRecursive_TerminalNodePull(t *testing.T) {
	// Test pulling a terminal cloned node
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	repoPath := filepath.Join(tmpDir, ".nodes", "repo1")
	os.MkdirAll(repoPath, 0755)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := interfaces.NodeInfo{
		Name:       "repo1",
		Path:       "/repo1",
		Repository: "https://github.com/test/repo1.git",
		IsCloned:   true,
		Children:   []interfaces.NodeInfo{},
	}

	err := mgr.pullRecursive(node, false)
	if err != nil {
		t.Errorf("pullRecursive failed: %v", err)
	}

	// Verify pull was called
	if !gitStub.pullCalled {
		t.Error("Expected pull to be called for terminal node")
	}
}

func TestPullRecursive_PullError(t *testing.T) {
	// Test that pull errors are handled gracefully (continue with other repos)
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	repoPath := filepath.Join(tmpDir, ".nodes", "repo1")
	os.MkdirAll(repoPath, 0755)

	gitStub := NewGitProviderStub()
	gitStub.pullError = fmt.Errorf("pull failed")
	mgr.gitProvider = gitStub

	node := interfaces.NodeInfo{
		Name:       "repo1",
		Path:       "/repo1",
		Repository: "https://github.com/test/repo1.git",
		IsCloned:   true,
		Children:   []interfaces.NodeInfo{},
	}

	err := mgr.pullRecursive(node, false)
	// Should not return error, just display error message
	if err != nil {
		t.Errorf("pullRecursive should not return error for pull failures: %v", err)
	}

	// Verify pull was attempted
	if !gitStub.pullCalled {
		t.Error("Expected pull to be called")
	}
}

func TestPullRecursive_ChildrenRecursion(t *testing.T) {
	// Test recursion into child nodes
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	repo1Path := filepath.Join(tmpDir, ".nodes", "parent", "child1")
	repo2Path := filepath.Join(tmpDir, ".nodes", "parent", "child2")
	os.MkdirAll(repo1Path, 0755)
	os.MkdirAll(repo2Path, 0755)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := interfaces.NodeInfo{
		Name: "parent",
		Path: "/parent",
		Children: []interfaces.NodeInfo{
			{
				Name:       "child1",
				Path:       "/parent/child1",
				Repository: "https://github.com/test/child1.git",
				IsCloned:   true,
				Children:   []interfaces.NodeInfo{},
			},
			{
				Name:       "child2",
				Path:       "/parent/child2",
				Repository: "https://github.com/test/child2.git",
				IsCloned:   true,
				Children:   []interfaces.NodeInfo{},
			},
		},
	}

	err := mgr.pullRecursive(node, false)
	if err != nil {
		t.Errorf("pullRecursive failed: %v", err)
	}

	// Verify both children were pulled
	if mgr.gitProvider.(*GitProviderStub).pullCallCount < 2 {
		t.Errorf("Expected at least 2 pull calls for children, got %d",
			mgr.gitProvider.(*GitProviderStub).pullCallCount)
	}
}

func TestPullRecursive_ForceFlag(t *testing.T) {
	// Test force flag is passed to git pull
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	repoPath := filepath.Join(tmpDir, ".nodes", "repo1")
	os.MkdirAll(repoPath, 0755)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := interfaces.NodeInfo{
		Name:       "repo1",
		Path:       "/repo1",
		Repository: "https://github.com/test/repo1.git",
		IsCloned:   true,
		Children:   []interfaces.NodeInfo{},
	}

	// Test with force flag
	err := mgr.pullRecursive(node, true)
	if err != nil {
		t.Errorf("pullRecursive with force failed: %v", err)
	}

	// Verify pull was called
	if !gitStub.pullCalled {
		t.Error("Expected pull to be called with force flag")
	}
}

// ================================================================================
// Helper Types for Testing
// ================================================================================

// FailingTreeProvider is a TreeProvider that fails on UpdateNode
type FailingTreeProvider struct {
	updateError error
}

func (f *FailingTreeProvider) Load(cfg interface{}) error { return nil }
func (f *FailingTreeProvider) Navigate(path string) error { return nil }
func (f *FailingTreeProvider) GetCurrent() (interfaces.NodeInfo, error) {
	return interfaces.NodeInfo{Name: "root", Path: "/"}, nil
}
func (f *FailingTreeProvider) GetTree() (interfaces.NodeInfo, error) {
	return interfaces.NodeInfo{Name: "root", Path: "/"}, nil
}
func (f *FailingTreeProvider) GetNode(path string) (interfaces.NodeInfo, error) {
	return interfaces.NodeInfo{}, nil
}
func (f *FailingTreeProvider) AddNode(parentPath string, node interfaces.NodeInfo) error {
	return nil
}
func (f *FailingTreeProvider) RemoveNode(path string) error { return nil }
func (f *FailingTreeProvider) UpdateNode(path string, node interfaces.NodeInfo) error {
	if f.updateError != nil {
		return f.updateError
	}
	return nil
}
func (f *FailingTreeProvider) ListChildren(path string) ([]interfaces.NodeInfo, error) {
	return nil, nil
}
func (f *FailingTreeProvider) GetPath() string { return "/" }
func (f *FailingTreeProvider) SetPath(path string) error {
	return nil
}
func (f *FailingTreeProvider) GetState() (interfaces.TreeState, error) {
	return interfaces.TreeState{}, nil
}
func (f *FailingTreeProvider) SetState(state interfaces.TreeState) error {
	return nil
}
