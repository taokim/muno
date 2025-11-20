package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
)

// =====================================================
// Custom Test Stubs
// =====================================================

// UIProviderStub is a UI provider that captures messages
type UIProviderStub struct {
	messages      []string
	confirmResult bool
}

func (u *UIProviderStub) Info(message string) {
	u.messages = append(u.messages, message)
}

func (u *UIProviderStub) Success(message string) {
	u.messages = append(u.messages, message)
}

func (u *UIProviderStub) Warning(message string) {
	u.messages = append(u.messages, message)
}

func (u *UIProviderStub) Error(message string) {
	u.messages = append(u.messages, message)
}

func (u *UIProviderStub) Debug(message string) {
	u.messages = append(u.messages, message)
}

func (u *UIProviderStub) Prompt(message string) (string, error) {
	return "", nil
}

func (u *UIProviderStub) PromptPassword(message string) (string, error) {
	return "", nil
}

func (u *UIProviderStub) Confirm(message string) (bool, error) {
	return u.confirmResult, nil
}

func (u *UIProviderStub) Select(message string, options []string) (string, error) {
	if len(options) > 0 {
		return options[0], nil
	}
	return "", nil
}

func (u *UIProviderStub) MultiSelect(message string, options []string) ([]string, error) {
	return options, nil
}

func (u *UIProviderStub) Progress(message string) interfaces.ProgressReporter {
	return &progressReporterStub{}
}

// progressReporterStub is a simple stub for progress reporting
type progressReporterStub struct{}

func (p *progressReporterStub) Start()                    {}
func (p *progressReporterStub) Update(current, total int) {}
func (p *progressReporterStub) SetMessage(message string) {}
func (p *progressReporterStub) Finish()                   {}
func (p *progressReporterStub) Error(err error)           {}

func (u *UIProviderStub) OpenInEditor(path string) error {
	return nil
}

func (u *UIProviderStub) OpenInBrowser(url string) error {
	return nil
}

// EnhancedGitProviderStub supports per-path pull errors
type EnhancedGitProviderStub struct {
	pullErrors      map[string]error
	pullCallCount   int
	lastPullOptions interfaces.PullOptions
}

func NewEnhancedGitProviderStub() *EnhancedGitProviderStub {
	return &EnhancedGitProviderStub{
		pullErrors: make(map[string]error),
	}
}

func (g *EnhancedGitProviderStub) Clone(url, path string, options interfaces.CloneOptions) error {
	return nil
}

func (g *EnhancedGitProviderStub) Pull(path string, options interfaces.PullOptions) error {
	g.pullCallCount++
	g.lastPullOptions = options
	if err, ok := g.pullErrors[path]; ok {
		return err
	}
	return nil
}

func (g *EnhancedGitProviderStub) Push(path string, options interfaces.PushOptions) error {
	return nil
}

func (g *EnhancedGitProviderStub) Status(path string) (*interfaces.GitStatus, error) {
	return &interfaces.GitStatus{Branch: "main", IsClean: true}, nil
}

func (g *EnhancedGitProviderStub) Commit(path string, message string, options interfaces.CommitOptions) error {
	return nil
}

func (g *EnhancedGitProviderStub) Branch(path string) (string, error) {
	return "main", nil
}

func (g *EnhancedGitProviderStub) Checkout(path string, branch string) error {
	return nil
}

func (g *EnhancedGitProviderStub) Fetch(path string, options interfaces.FetchOptions) error {
	return nil
}

func (g *EnhancedGitProviderStub) Add(path string, files []string) error {
	return nil
}

func (g *EnhancedGitProviderStub) Remove(path string, files []string) error {
	return nil
}

func (g *EnhancedGitProviderStub) GetRemoteURL(path string) (string, error) {
	return "https://github.com/test/repo.git", nil
}

func (g *EnhancedGitProviderStub) SetRemoteURL(path string, url string) error {
	return nil
}

// =====================================================
// Tests for pullAllRepositories and collectClonedRepos
// =====================================================

// TestPullAllRepositories_NoRepositories tests when no cloned repos exist
func TestPullAllRepositories_NoRepositories(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Create tree with only lazy (uncloned) repositories
	lazyNode1 := CreateLazyNode("lazy1", "https://github.com/test/lazy1.git")
	lazyNode2 := CreateLazyNode("lazy2", "https://github.com/test/lazy2.git")

	AddNodeToTree(mgr, lazyNode1.Path, lazyNode1)
	AddNodeToTree(mgr, lazyNode2.Path, lazyNode2)

	uiStub := &UIProviderStub{messages: []string{}}
	mgr.uiProvider = uiStub

	// Call pullAllRepositories - should handle empty repos list
	err := mgr.pullAllRepositories(false)
	if err != nil {
		t.Fatalf("pullAllRepositories failed: %v", err)
	}

	// Verify UI showed no repos message
	if !containsMessage(uiStub.messages, "No cloned repositories found") {
		t.Error("Expected 'No cloned repositories found' message")
	}
}

// TestPullAllRepositories_AllSuccess tests when all pulls succeed
func TestPullAllRepositories_AllSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Create cloned repositories with .git directories
	backend := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	frontend := CreateSimpleNode("frontend", "https://github.com/test/frontend.git")

	// Create .git directories to simulate cloned repos
	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	frontendPath := filepath.Join(tmpDir, ".nodes", "frontend")
	os.MkdirAll(filepath.Join(backendPath, ".git"), 0755)
	os.MkdirAll(filepath.Join(frontendPath, ".git"), 0755)

	AddNodeToTree(mgr, backend.Path, backend)
	AddNodeToTree(mgr, frontend.Path, frontend)

	// Setup git stub that succeeds on pull
	gitStub := NewEnhancedGitProviderStub()
	mgr.gitProvider = gitStub

	uiStub := &UIProviderStub{messages: []string{}}
	mgr.uiProvider = uiStub

	// Call pullAllRepositories
	err := mgr.pullAllRepositories(false)
	if err != nil {
		t.Fatalf("pullAllRepositories failed: %v", err)
	}

	// Verify both repos were pulled
	if gitStub.pullCallCount != 2 {
		t.Errorf("Expected 2 pull calls, got %d", gitStub.pullCallCount)
	}

	// Verify success message
	if !containsMessage(uiStub.messages, "2 succeeded, 0 failed") {
		t.Error("Expected success summary message")
	}
}

// TestPullAllRepositories_SomeFailures tests when some pulls fail
func TestPullAllRepositories_SomeFailures(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Create cloned repositories
	backend := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	frontend := CreateSimpleNode("frontend", "https://github.com/test/frontend.git")

	// Create .git directories
	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	frontendPath := filepath.Join(tmpDir, ".nodes", "frontend")
	os.MkdirAll(filepath.Join(backendPath, ".git"), 0755)
	os.MkdirAll(filepath.Join(frontendPath, ".git"), 0755)

	AddNodeToTree(mgr, backend.Path, backend)
	AddNodeToTree(mgr, frontend.Path, frontend)

	// Setup git stub that fails for backend
	gitStub := NewEnhancedGitProviderStub()
	gitStub.pullErrors[backendPath] = fmt.Errorf("pull failed: local changes")
	mgr.gitProvider = gitStub

	uiStub := &UIProviderStub{messages: []string{}}
	mgr.uiProvider = uiStub

	// Call pullAllRepositories without force
	err := mgr.pullAllRepositories(false)
	if err != nil {
		t.Fatalf("pullAllRepositories failed: %v", err)
	}

	// Verify summary shows failures
	if !containsMessage(uiStub.messages, "1 succeeded, 1 failed") {
		t.Error("Expected summary with 1 success and 1 failure")
	}

	// Verify failed repo is listed
	if !containsMessage(uiStub.messages, "backend") {
		t.Error("Expected backend in failed repos list")
	}

	// Verify force tip is shown
	if !containsMessage(uiStub.messages, "Use --force to override") {
		t.Error("Expected force flag tip")
	}
}

// TestPullAllRepositories_WithForceFlag tests force flag usage
func TestPullAllRepositories_WithForceFlag(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Create cloned repository
	backend := CreateSimpleNode("backend", "https://github.com/test/backend.git")

	backendPath := filepath.Join(tmpDir, ".nodes", "backend")
	os.MkdirAll(filepath.Join(backendPath, ".git"), 0755)

	AddNodeToTree(mgr, backend.Path, backend)

	// Setup git stub that fails
	gitStub := NewEnhancedGitProviderStub()
	gitStub.pullErrors[backendPath] = fmt.Errorf("pull failed")
	mgr.gitProvider = gitStub

	uiStub := &UIProviderStub{messages: []string{}}
	mgr.uiProvider = uiStub

	// Call with force=true
	err := mgr.pullAllRepositories(true)
	if err != nil {
		t.Fatalf("pullAllRepositories failed: %v", err)
	}

	// Verify force flag was passed to git
	if !gitStub.lastPullOptions.Force {
		t.Error("Expected force flag to be true")
	}

	// Verify force tip is NOT shown when using force
	if containsMessage(uiStub.messages, "Use --force") {
		t.Error("Should not show force tip when already using force")
	}
}

// =====================================================
// Tests for collectClonedRepos
// =====================================================

// TestCollectClonedRepos_ConfigNodeWithURLChildren tests config node expansion
func TestCollectClonedRepos_ConfigNodeWithURLChildren(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Create config file with URL children
	configContent := `
workspace:
    name: team-backend
    repos_dir: services
nodes:
  - name: payment-service
    url: https://github.com/test/payment.git
    fetch: eager
  - name: order-service
    url: https://github.com/test/order.git
    fetch: eager
`
	// Config goes under .nodes/team-backend (first level)
	configPath := filepath.Join(tmpDir, ".nodes", "team-backend", "muno.yaml")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create .git directories for cloned repos (under .nodes/team-backend/services/)
	paymentPath := filepath.Join(tmpDir, ".nodes", "team-backend", "services", "payment-service", ".git")
	orderPath := filepath.Join(tmpDir, ".nodes", "team-backend", "services", "order-service", ".git")
	os.MkdirAll(paymentPath, 0755)
	os.MkdirAll(orderPath, 0755)

	// Create config node
	configNode := interfaces.NodeInfo{
		Name:       "team-backend",
		Path:       "/team-backend",
		ConfigFile: configPath,
		IsConfig:   true,
	}

	// Add config node to tree so computeFilesystemPath can resolve it
	AddNodeToTree(mgr, "/team-backend", configNode)

	// Collect cloned repos
	repos := mgr.collectClonedRepos(configNode)

	// Verify both repos were collected
	if len(repos) != 2 {
		t.Errorf("Expected 2 repos, got %d", len(repos))
	}

	// Verify repo details
	repoNames := make(map[string]bool)
	for _, repo := range repos {
		repoNames[repo.Name] = true
		if !repo.IsCloned {
			t.Errorf("Repo %s should be marked as cloned", repo.Name)
		}
	}

	if !repoNames["payment-service"] || !repoNames["order-service"] {
		t.Error("Expected payment-service and order-service in collected repos")
	}
}

// TestCollectClonedRepos_ConfigNodeWithFileChildren tests nested config nodes
func TestCollectClonedRepos_ConfigNodeWithFileChildren(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Create nested config file (referenced by parent, can be anywhere)
	nestedConfigContent := `
workspace:
    name: frontend
    repos_dir: apps
nodes:
  - name: web-app
    url: https://github.com/test/web.git
    fetch: eager
`
	nestedConfigPath := filepath.Join(tmpDir, "frontend", "muno.yaml")
	os.MkdirAll(filepath.Dir(nestedConfigPath), 0755)
	if err := os.WriteFile(nestedConfigPath, []byte(nestedConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create nested config: %v", err)
	}

	// Create parent config with File child
	parentConfigContent := `
workspace:
    name: platform
    repos_dir: teams
nodes:
  - name: frontend
    file: frontend/muno.yaml
`
	parentConfigPath := filepath.Join(tmpDir, "muno.yaml")
	if err := os.WriteFile(parentConfigPath, []byte(parentConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create parent config: %v", err)
	}

	// Create .git directory for web-app
	// Frontend is first level, so goes under .nodes/frontend
	// Then repos_dir=apps means web-app is under .nodes/frontend/apps/web-app
	webAppPath := filepath.Join(tmpDir, ".nodes", "frontend", "apps", "web-app", ".git")
	os.MkdirAll(webAppPath, 0755)

	// Create config node
	configNode := interfaces.NodeInfo{
		Name:       "platform",
		Path:       "/",
		ConfigFile: parentConfigPath,
		IsConfig:   true,
	}

	// Add parent config node to tree (it's the root, so it's already there from initializeRootNode)
	// But we need to update it with the ConfigFile
	mgr.treeProvider.UpdateNode("/", configNode)

	// Also add the nested frontend config node so path resolution works
	frontendNode := interfaces.NodeInfo{
		Name:       "frontend",
		Path:       "/frontend",
		ConfigFile: nestedConfigPath,
		IsConfig:   true,
	}
	AddNodeToTree(mgr, "/frontend", frontendNode)

	// Collect cloned repos
	repos := mgr.collectClonedRepos(configNode)

	// Verify web-app was collected through nested config
	if len(repos) != 1 {
		t.Errorf("Expected 1 repo, got %d", len(repos))
	}

	if len(repos) > 0 && repos[0].Name != "web-app" {
		t.Errorf("Expected web-app, got %s", repos[0].Name)
	}
}

// TestCollectClonedRepos_TerminalClonedNode tests terminal node collection
func TestCollectClonedRepos_TerminalClonedNode(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Create terminal cloned node (no children)
	terminalNode := interfaces.NodeInfo{
		Name:       "standalone",
		Path:       "/standalone",
		Repository: "https://github.com/test/standalone.git",
		IsCloned:   true,
		Children:   []interfaces.NodeInfo{}, // Explicitly empty
	}

	// Collect cloned repos
	repos := mgr.collectClonedRepos(terminalNode)

	// Verify terminal node is collected
	if len(repos) != 1 {
		t.Errorf("Expected 1 repo, got %d", len(repos))
	}

	if len(repos) > 0 && repos[0].Name != "standalone" {
		t.Errorf("Expected standalone, got %s", repos[0].Name)
	}
}

// TestCollectClonedRepos_RecursiveChildren tests recursive child processing
func TestCollectClonedRepos_RecursiveChildren(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Create parent node with children
	child1 := CreateSimpleNode("child1", "https://github.com/test/child1.git")
	child1.Path = "/parent/child1"
	child1.Children = []interfaces.NodeInfo{}

	child2 := CreateSimpleNode("child2", "https://github.com/test/child2.git")
	child2.Path = "/parent/child2"
	child2.Children = []interfaces.NodeInfo{}

	parentNode := interfaces.NodeInfo{
		Name:     "parent",
		Path:     "/parent",
		Children: []interfaces.NodeInfo{child1, child2},
	}

	// Collect cloned repos
	repos := mgr.collectClonedRepos(parentNode)

	// Verify both children were collected
	if len(repos) != 2 {
		t.Errorf("Expected 2 repos, got %d", len(repos))
	}

	repoNames := make(map[string]bool)
	for _, repo := range repos {
		repoNames[repo.Name] = true
	}

	if !repoNames["child1"] || !repoNames["child2"] {
		t.Error("Expected child1 and child2 in collected repos")
	}
}

// TestCollectClonedRepos_MixedClonedAndUncloned tests filtering cloned repos
func TestCollectClonedRepos_MixedClonedAndUncloned(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Create mix of cloned and uncloned nodes
	cloned := CreateSimpleNode("cloned", "https://github.com/test/cloned.git")
	cloned.Path = "/parent/cloned"
	cloned.Children = []interfaces.NodeInfo{}

	uncloned := CreateSimpleNode("uncloned", "https://github.com/test/uncloned.git")
	uncloned.Path = "/parent/uncloned"
	uncloned.IsCloned = false
	uncloned.Children = []interfaces.NodeInfo{}

	parentNode := interfaces.NodeInfo{
		Name:     "parent",
		Path:     "/parent",
		Children: []interfaces.NodeInfo{cloned, uncloned},
	}

	// Collect cloned repos
	repos := mgr.collectClonedRepos(parentNode)

	// Verify only cloned repo is collected
	if len(repos) != 1 {
		t.Errorf("Expected 1 repo, got %d", len(repos))
	}

	if len(repos) > 0 && repos[0].Name != "cloned" {
		t.Errorf("Expected cloned repo only, got %s", repos[0].Name)
	}
}

// =====================================================
// Tests for Remove function
// =====================================================

// TestRemove_UserCancels tests when user cancels removal
func TestRemove_UserCancels(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Add a repository to remove
	backend := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, backend.Path, backend)

	// Setup UI to return false for confirm
	uiStub := &UIProviderStub{confirmResult: false, messages: []string{}}
	mgr.uiProvider = uiStub

	// Call Remove
	err := mgr.Remove(context.Background(), "backend")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify cancellation message
	if !containsMessage(uiStub.messages, "Removal cancelled") {
		t.Error("Expected cancellation message")
	}

	// Verify node still exists
	node, err := mgr.treeProvider.GetNode("/backend")
	if err != nil || node.Name != "backend" {
		t.Error("Node should still exist after cancellation")
	}
}

// TestRemove_UnclonedNode tests removing a lazy (uncloned) node
func TestRemove_UnclonedNode(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Add lazy (uncloned) repository
	lazy := CreateLazyNode("lazy-repo", "https://github.com/test/lazy.git")
	AddNodeToTree(mgr, lazy.Path, lazy)

	// Update config with the node
	mgr.config = &config.ConfigTree{
		Nodes: []config.NodeDefinition{
			{Name: "lazy-repo", URL: "https://github.com/test/lazy.git", Fetch: config.FetchLazy},
		},
	}

	// Setup UI to confirm
	uiStub := &UIProviderStub{confirmResult: true, messages: []string{}}
	mgr.uiProvider = uiStub

	// Call Remove
	err := mgr.Remove(context.Background(), "lazy-repo")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify node was removed from tree
	_, err = mgr.treeProvider.GetNode("/lazy-repo")
	if err == nil {
		t.Error("Node should be removed from tree")
	}

	// Verify config was updated
	if len(mgr.config.Nodes) != 0 {
		t.Error("Config should have empty nodes after removal")
	}
}

// TestRemove_FilesystemRemovalFailure tests when filesystem removal fails
func TestRemove_FilesystemRemovalFailure(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Add cloned repository
	backend := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, backend.Path, backend)

	// Update config
	mgr.config = &config.ConfigTree{
		Nodes: []config.NodeDefinition{
			{Name: "backend", URL: "https://github.com/test/backend.git"},
		},
	}

	// Setup filesystem stub that fails on RemoveAll
	fsStub := &FailingFileSystemProvider{
		removeAllError: fmt.Errorf("permission denied"),
		existsResult:   true,
	}
	mgr.fsProvider = fsStub

	// Setup UI to confirm
	uiStub := &UIProviderStub{confirmResult: true, messages: []string{}}
	mgr.uiProvider = uiStub

	// Call Remove - should not fail even if filesystem removal fails
	err := mgr.Remove(context.Background(), "backend")
	if err != nil {
		t.Fatalf("Remove should succeed despite filesystem failure: %v", err)
	}

	// Verify node was still removed from tree (graceful degradation)
	_, err = mgr.treeProvider.GetNode("/backend")
	if err == nil {
		t.Error("Node should be removed from tree even if filesystem removal failed")
	}
}

// TestRemove_ConfigSaveFailure tests when config save fails
func TestRemove_ConfigSaveFailure(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)
	initializeRootNode(mgr)

	// Add repository
	backend := CreateSimpleNode("backend", "https://github.com/test/backend.git")
	AddNodeToTree(mgr, backend.Path, backend)

	// Update config
	mgr.config = &config.ConfigTree{
		Nodes: []config.NodeDefinition{
			{Name: "backend", URL: "https://github.com/test/backend.git"},
		},
	}

	// Setup config provider that fails on save
	failingConfigProvider := &FailingConfigProviderForClone{
		saveError: fmt.Errorf("disk full"),
	}
	mgr.configProvider = failingConfigProvider

	// Setup UI to confirm
	uiStub := &UIProviderStub{confirmResult: true, messages: []string{}}
	mgr.uiProvider = uiStub

	// Call Remove - should not fail even if config save fails
	err := mgr.Remove(context.Background(), "backend")
	if err != nil {
		t.Fatalf("Remove should succeed despite config save failure: %v", err)
	}

	// Verify node was removed from tree (operation succeeded)
	_, err = mgr.treeProvider.GetNode("/backend")
	if err == nil {
		t.Error("Node should be removed from tree")
	}
}

// =====================================================
// Helper Types and Functions
// =====================================================

// FailingFileSystemProvider is a filesystem stub that can fail on operations
type FailingFileSystemProvider struct {
	removeAllError error
	existsResult   bool
}

func (f *FailingFileSystemProvider) RemoveAll(path string) error {
	if f.removeAllError != nil {
		return f.removeAllError
	}
	return nil
}

func (f *FailingFileSystemProvider) Exists(path string) bool {
	return f.existsResult
}

func (f *FailingFileSystemProvider) ReadDir(path string) ([]interfaces.FileInfo, error) {
	return nil, nil
}

func (f *FailingFileSystemProvider) Create(path string) error {
	return nil
}

func (f *FailingFileSystemProvider) Mkdir(path string, perm os.FileMode) error {
	return nil
}

func (f *FailingFileSystemProvider) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (f *FailingFileSystemProvider) WriteFile(path string, data []byte, perm os.FileMode) error {
	return nil
}

func (f *FailingFileSystemProvider) ReadFile(path string) ([]byte, error) {
	return nil, nil
}

func (f *FailingFileSystemProvider) Remove(path string) error {
	return nil
}

func (f *FailingFileSystemProvider) Rename(oldPath, newPath string) error {
	return nil
}

func (f *FailingFileSystemProvider) Symlink(oldName, newName string) error {
	return nil
}

func (f *FailingFileSystemProvider) Stat(path string) (interfaces.FileInfo, error) {
	return interfaces.FileInfo{
		Name:    "test",
		Size:    0,
		Mode:    0755,
		ModTime: time.Time{},
		IsDir:   false,
	}, nil
}

func (f *FailingFileSystemProvider) Walk(root string, walkFn filepath.WalkFunc) error {
	return nil
}

func (f *FailingFileSystemProvider) Copy(src, dst string) error {
	return nil
}

// initializeRootNode adds a root node to the tree
func initializeRootNode(mgr *Manager) {
	rootNode := interfaces.NodeInfo{Name: "root", Path: "/", Children: []interfaces.NodeInfo{}}
	AddNodeToTree(mgr, "/", rootNode)
}

// containsMessage checks if a message contains a substring
func containsMessage(messages []string, substring string) bool {
	for _, msg := range messages {
		if strings.Contains(msg, substring) {
			return true
		}
	}
	return false
}
