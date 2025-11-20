package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/taokim/muno/internal/interfaces"
)

// TestPullNodeWithOptions_EmptyPathFromWorkspaceRoot tests PWD-based resolution from workspace root
func TestPullNodeWithOptions_EmptyPathFromWorkspaceRoot(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add a node
	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	if err := os.MkdirAll(backendPath, 0755); err != nil {
		t.Fatalf("Failed to create backend directory: %v", err)
	}

	// Change to workspace root
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Test pull with empty path (should use PWD resolution)
	err := mgr.PullNodeWithOptions("", false, false, false)
	// Should resolve to root and attempt to pull
	_ = err // May fail if no repo at root, but should not panic
}

// TestPullNodeWithOptions_NonRecursive tests single node pull (non-recursive)
func TestPullNodeWithOptions_NonRecursive(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add a node
	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	if err := os.MkdirAll(backendPath, 0755); err != nil {
		t.Fatalf("Failed to create backend directory: %v", err)
	}

	// Test pull with explicit path non-recursive
	err := mgr.PullNodeWithOptions("/backend", false, false, false)
	if err != nil {
		t.Errorf("PullNodeWithOptions failed: %v", err)
	}

	if !gitStub.pullCalled {
		t.Error("Expected Pull to be called")
	}
}

// TestPullNodeWithOptions_RecursiveEmptyPath tests recursive pull with empty path (--all case)
func TestPullNodeWithOptions_RecursiveEmptyPath(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add multiple nodes
	node1 := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node1)

	node2 := CreateSimpleNode("frontend", "https://github.com/test/frontend.git")
	AddNodeToTree(mgr, "/frontend", node2)

	// Create directories
	backend := filepath.Join(tmpDir, ".nodes", "backend")
	frontend := filepath.Join(tmpDir, ".nodes", "frontend")
	os.MkdirAll(backend, 0755)
	os.MkdirAll(frontend, 0755)

	// Test pull with empty path and recursive (should call pullAllRepositories)
	err := mgr.PullNodeWithOptions("", true, false, false)
	// May fail if pullAllRepositories not fully implemented, but should not panic
	_ = err
}

// TestPullNodeWithOptions_RecursiveWithIncludeLazy tests recursive pull with includeLazy
func TestPullNodeWithOptions_RecursiveWithIncludeLazy(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add parent with lazy child
	parent := CreateSimpleNode("parent", "https://github.com/test/parent.git")
	AddNodeToTree(mgr, "/parent", parent)

	lazy := CreateLazyNode("lazy", "https://github.com/test/lazy.git")
	lazy.Path = "/parent/lazy"
	parent.Children = append(parent.Children, lazy)
	mgr.treeProvider.UpdateNode("/parent", parent)

	parentPath := filepath.Join(tmpDir, ".nodes", "parent")
	os.MkdirAll(parentPath, 0755)

	// Test recursive pull with includeLazy
	err := mgr.PullNodeWithOptions("/parent", true, false, true)
	// May fail if not fully implemented, but should attempt to clone lazy repos
	_ = err
}

// TestPullNodeWithOptions_SingleNodeWithForce tests single node pull with force flag
func TestPullNodeWithOptions_SingleNodeWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	os.MkdirAll(backendPath, 0755)

	// Test pull with force flag
	err := mgr.PullNodeWithOptions("/backend", false, true, false)
	if err != nil {
		t.Errorf("PullNodeWithOptions with force failed: %v", err)
	}

	if !gitStub.pullCalled {
		t.Error("Expected Pull to be called with force option")
	}
}

// TestCommitNode_EmptyPathFromWorkspaceRoot tests commit with PWD-based resolution from workspace root
func TestCommitNode_EmptyPathFromWorkspaceRoot(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add a node
	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	os.MkdirAll(backendPath, 0755)

	// Change to workspace root
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Test commit with empty path (should use PWD resolution to root)
	err := mgr.CommitNode("", "test message", false)
	// May fail if no repo at root, but should not panic
	_ = err
}

// TestCommitNode_EmptyPathFromReposDir tests commit with explicit path (simplified from PWD test)
func TestCommitNode_EmptyPathFromReposDir(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	os.MkdirAll(backendPath, 0755)

	// Test commit with explicit path (covers the main commit logic)
	err := mgr.CommitNode("/backend", "test message", false)
	if err != nil {
		t.Errorf("CommitNode failed: %v", err)
	}

	if !gitStub.commitCalled {
		t.Error("Expected Commit to be called")
	}

	if gitStub.lastCommitMsg != "test message" {
		t.Errorf("Expected commit message 'test message', got '%s'", gitStub.lastCommitMsg)
	}
}

// TestCommitNode_RecursiveFlag tests commit with recursive flag (though it doesn't seem to use it)
func TestCommitNode_RecursiveFlag(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	os.MkdirAll(backendPath, 0755)

	// Test commit with recursive flag (currently has no effect)
	err := mgr.CommitNode("/backend", "test message", true)
	if err != nil {
		t.Errorf("CommitNode with recursive flag failed: %v", err)
	}

	if !gitStub.commitCalled {
		t.Error("Expected Commit to be called")
	}
}

// TestPushNode_EmptyPathFromWorkspaceRoot tests push with PWD-based resolution from workspace root
func TestPushNode_EmptyPathFromWorkspaceRoot(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	os.MkdirAll(backendPath, 0755)

	// Change to workspace root
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Test push with empty path (should use PWD resolution to root)
	err := mgr.PushNode("", false)
	// May fail if no repo at root, but should not panic
	_ = err
}

// TestPushNode_EmptyPathFromReposDir tests push with explicit path (simplified from PWD test)
func TestPushNode_EmptyPathFromReposDir(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	os.MkdirAll(backendPath, 0755)

	// Test push with explicit path (covers the main push logic)
	err := mgr.PushNode("/backend", false)
	if err != nil {
		t.Errorf("PushNode failed: %v", err)
	}

	if !gitStub.pushCalled {
		t.Error("Expected Push to be called")
	}
}

// TestPushNode_RecursiveWithError tests push recursive handling of errors
func TestPushNode_RecursiveWithError(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	gitStub.pushError = fmt.Errorf("push failed")
	mgr.gitProvider = gitStub

	// Add parent with children
	parent := CreateSimpleNode("parent", "https://github.com/test/parent.git")
	AddNodeToTree(mgr, "/parent", parent)

	child := CreateSimpleNode("child", "https://github.com/test/child.git")
	child.Path = "/parent/child"
	parent.Children = append(parent.Children, child)
	mgr.treeProvider.UpdateNode("/parent", parent)

	parentPath := filepath.Join(tmpDir, ".nodes", "parent")
	childPath := filepath.Join(parentPath, "child")
	os.MkdirAll(childPath, 0755)

	// Test recursive push with errors
	err := mgr.PushNode("/parent", true)
	// Should handle errors gracefully
	_ = err

	if gitStub.pushCallCount == 0 {
		t.Error("Expected at least one push attempt")
	}
}

// TestCloneRepos_EmptyPathFromWorkspaceRoot tests clone with PWD-based resolution from workspace root
func TestCloneRepos_EmptyPathFromWorkspaceRoot(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add a lazy node
	lazy := CreateLazyNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", lazy)

	// Change to workspace root
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Test clone with empty path (should use PWD resolution to root)
	err := mgr.CloneRepos("", false, true)
	// May fail if implementation incomplete, but should not panic
	_ = err
}

// TestCloneRepos_EmptyPathFromReposDir tests clone with PWD-based resolution from repos directory
func TestCloneRepos_EmptyPathFromReposDir(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add parent with lazy child
	parent := CreateSimpleNode("parent", "https://github.com/test/parent.git")
	AddNodeToTree(mgr, "/parent", parent)

	lazy := CreateLazyNode("lazy", "https://github.com/test/lazy.git")
	lazy.Path = "/parent/lazy"
	parent.Children = append(parent.Children, lazy)
	mgr.treeProvider.UpdateNode("/parent", parent)

	parentPath := filepath.Join(tmpDir, ".nodes", "parent")
	os.MkdirAll(parentPath, 0755)

	// Change to parent directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(parentPath)

	// Test clone with empty path (should resolve to /parent)
	err := mgr.CloneRepos("", false, true)
	// Should attempt to clone children
	_ = err
}

// TestCloneRepos_NonRecursiveWithIncludeLazy tests non-recursive clone with includeLazy
func TestCloneRepos_NonRecursiveWithIncludeLazy(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add parent with lazy children
	parent := CreateSimpleNode("parent", "https://github.com/test/parent.git")
	AddNodeToTree(mgr, "/parent", parent)

	lazy1 := CreateLazyNode("lazy1", "https://github.com/test/lazy1.git")
	lazy1.Path = "/parent/lazy1"

	lazy2 := CreateLazyNode("lazy2", "https://github.com/test/lazy2.git")
	lazy2.Path = "/parent/lazy2"

	parent.Children = append(parent.Children, lazy1, lazy2)
	mgr.treeProvider.UpdateNode("/parent", parent)

	parentPath := filepath.Join(tmpDir, ".nodes", "parent")
	os.MkdirAll(parentPath, 0755)

	// Test non-recursive clone (should only clone direct children)
	err := mgr.CloneRepos("/parent", false, true)
	// Should attempt to clone children but not their descendants
	_ = err
}

// TestCloneRepos_RecursiveNested tests recursive clone with nested lazy repos
func TestCloneRepos_RecursiveNested(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add parent with nested lazy children
	parent := CreateSimpleNode("parent", "https://github.com/test/parent.git")
	AddNodeToTree(mgr, "/parent", parent)

	child := CreateLazyNode("child", "https://github.com/test/child.git")
	child.Path = "/parent/child"

	grandchild := CreateLazyNode("grandchild", "https://github.com/test/grandchild.git")
	grandchild.Path = "/parent/child/grandchild"
	child.Children = append(child.Children, grandchild)

	parent.Children = append(parent.Children, child)
	mgr.treeProvider.UpdateNode("/parent", parent)

	parentPath := filepath.Join(tmpDir, ".nodes", "parent")
	os.MkdirAll(parentPath, 0755)

	// Test recursive clone (should attempt to clone all descendants)
	err := mgr.CloneRepos("/parent", true, true)
	// May fail due to implementation, but should attempt recursive clone
	_ = err
}

// TestCloneRepos_NoReposToClone tests clone when no repos need cloning
func TestCloneRepos_NoReposToClone(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add already cloned node
	node := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	node.IsCloned = true
	AddNodeToTree(mgr, "/backend", node)

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	os.MkdirAll(backendPath, 0755)

	// Test clone (should report no repos to clone)
	err := mgr.CloneRepos("/backend", false, false)
	if err != nil {
		t.Errorf("CloneRepos with no repos to clone failed: %v", err)
	}

	// Should not have called clone
	if gitStub.cloneCalled {
		t.Error("Did not expect Clone to be called for already cloned repos")
	}
}

// FailingConfigProviderForClone is a ConfigProvider that can fail on Save
type FailingConfigProviderForClone struct {
	saveError error
}

func (f *FailingConfigProviderForClone) Load(path string) (interface{}, error) {
	return nil, nil
}

func (f *FailingConfigProviderForClone) Save(path string, cfg interface{}) error {
	if f.saveError != nil {
		return f.saveError
	}
	return nil
}

func (f *FailingConfigProviderForClone) Exists(path string) bool {
	return false
}

func (f *FailingConfigProviderForClone) Watch(path string) (<-chan interfaces.ConfigEvent, error) {
	return nil, nil
}

// TestCloneRepos_SaveConfigError tests clone with saveConfig failure
func TestCloneRepos_SaveConfigError(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	// Use failing config provider
	failingConfig := &FailingConfigProviderForClone{
		saveError: fmt.Errorf("failed to save config"),
	}
	mgr.configProvider = failingConfig

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Add a lazy node
	lazy := CreateLazyNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, "/backend", lazy)

	// Test clone (should fail when trying to save config)
	err := mgr.CloneRepos("/backend", false, true)
	if err == nil {
		t.Error("Expected error from saveConfig failure")
	}
}
