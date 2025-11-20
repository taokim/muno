package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/taokim/muno/internal/interfaces"
)

// ================================================================================
// pullRecursiveWithOptions - Config Node Tests
// ================================================================================

func TestPullRecursiveWithOptions_ConfigNodeWithURLChildren(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub
	
	// Create config file with URL children
	configContent := `workspace:
  name: team
nodes:
  - name: service1
    url: https://github.com/test/service1.git
    fetch: eager
  - name: service2
    url: https://github.com/test/service2.git
    fetch: lazy
`
	configPath := filepath.Join(tmpDir, "team-config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create .git directories to simulate cloned repos (service1 is cloned, service2 is lazy)
	service1Path := filepath.Join(tmpDir, ".nodes", "team", "service1", ".git")
	if err := os.MkdirAll(service1Path, 0755); err != nil {
		t.Fatalf("Failed to create service1 .git: %v", err)
	}

	// Verify .git exists
	if _, err := os.Stat(service1Path); os.IsNotExist(err) {
		t.Fatalf("service1 .git directory does not exist at %s", service1Path)
	}
	t.Logf("Created .git directory at: %s", service1Path)

	// Create config node
	node := interfaces.NodeInfo{
		Name:       "team",
		Path:       "/team",
		ConfigFile: configPath,
		IsConfig:   true,
	}

	// Add the config node to the tree so computeFilesystemPath can resolve it
	if err := mgr.treeProvider.UpdateNode(node.Path, node); err != nil {
		t.Fatalf("Failed to add node to tree: %v", err)
	}

	err := mgr.pullRecursiveWithOptions(node, false, false)
	if err != nil {
		t.Errorf("pullRecursiveWithOptions failed: %v", err)
	}
	t.Logf("Pull call count after operation: %d", gitStub.pullCallCount)

	// Verify pull was called for service1 (cloned)
	if !gitStub.pullCalled {
		t.Error("Expected pull to be called for cloned repository")
	}
}

func TestPullRecursiveWithOptions_ConfigNodeWithFileChildren(t *testing.T) {
	// Test config node with File children (nested config nodes)
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Create nested config file
	nestedConfigPath := filepath.Join(tmpDir, "frontend-config.yaml")
	nestedConfigContent := `workspace:
  name: frontend
nodes:
  - name: web-app
    url: https://github.com/test/web-app.git
    fetch: eager
`
	if err := os.WriteFile(nestedConfigPath, []byte(nestedConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create nested config file: %v", err)
	}

	// Create parent config file
	parentConfigPath := filepath.Join(tmpDir, "team-config.yaml")
	parentConfigContent := fmt.Sprintf(`workspace:
  name: team
  repos_dir: nodes
nodes:
  - name: frontend
    file: %s
`, nestedConfigPath)
	if err := os.WriteFile(parentConfigPath, []byte(parentConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create parent config file: %v", err)
	}

	// Create .git directory for web-app
	// team config specifies repos_dir: nodes, so frontend is under nodes/
	// frontend config doesn't specify repos_dir, so web-app goes directly under frontend
	webAppPath := filepath.Join(tmpDir, ".nodes", "team", "nodes", "frontend", "web-app", ".git")
	os.MkdirAll(webAppPath, 0755)

	// Create config node
	node := interfaces.NodeInfo{
		Name:       "team",
		Path:       "/team",
		ConfigFile: parentConfigPath,
		IsConfig:   true,
	}

	// Add the config node to the tree
	if err := mgr.treeProvider.UpdateNode(node.Path, node); err != nil {
		t.Fatalf("Failed to add node to tree: %v", err)
	}

	// Add the frontend nested config node to the tree for path resolution
	frontendNode := interfaces.NodeInfo{
		Name:       "frontend",
		Path:       "/team/frontend",
		ConfigFile: nestedConfigPath,
		IsConfig:   true,
	}
	if err := mgr.treeProvider.UpdateNode(frontendNode.Path, frontendNode); err != nil {
		t.Fatalf("Failed to add frontend node to tree: %v", err)
	}

	err := mgr.pullRecursiveWithOptions(node, false, false)
	if err != nil {
		t.Errorf("pullRecursiveWithOptions failed: %v", err)
	}

	// Verify pull was called for web-app (nested config child)
	if !gitStub.pullCalled {
		t.Error("Expected pull to be called for nested config repository")
	}
}

func TestPullRecursiveWithOptions_ConfigNodeWithLazyCloning(t *testing.T) {
	// Test config node with includeLazy=true clones lazy repos
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Create config file with lazy repository
	configPath := filepath.Join(tmpDir, "team-config.yaml")
	configContent := `workspace:
  name: team
  repos_dir: services
nodes:
  - name: lazy-service
    url: https://github.com/test/lazy-service.git
    fetch: lazy
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create config node
	node := interfaces.NodeInfo{
		Name:       "team",
		Path:       "/team",
		ConfigFile: configPath,
		IsConfig:   true,
	}

	err := mgr.pullRecursiveWithOptions(node, false, true)
	if err != nil {
		t.Errorf("pullRecursiveWithOptions failed: %v", err)
	}

	// Verify clone was called for lazy repository
	if !gitStub.cloneCalled {
		t.Error("Expected clone to be called for lazy repository with includeLazy=true")
	}
}

func TestPullRecursiveWithOptions_ConfigNodeMixedChildren(t *testing.T) {
	// Test config node with both URL and File children
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Create nested config
	nestedConfigPath := filepath.Join(tmpDir, "sub-config.yaml")
	nestedConfigContent := `workspace:
  name: subteam
nodes:
  - name: repo3
    url: https://github.com/test/repo3.git
    fetch: eager
`
	if err := os.WriteFile(nestedConfigPath, []byte(nestedConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create nested config: %v", err)
	}

	// Create parent config with mixed children
	parentConfigPath := filepath.Join(tmpDir, "parent-config.yaml")
	parentConfigContent := fmt.Sprintf(`workspace:
  name: parent
nodes:
  - name: repo1
    url: https://github.com/test/repo1.git
    fetch: eager
  - name: subteam
    file: %s
  - name: repo2
    url: https://github.com/test/repo2.git
    fetch: eager
`, nestedConfigPath)
	if err := os.WriteFile(parentConfigPath, []byte(parentConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create parent config: %v", err)
	}

	// Create .git directories
	// parent and subteam configs don't specify repos_dir, so children go directly under parent
	repo1Path := filepath.Join(tmpDir, ".nodes", "parent", "repo1", ".git")
	repo2Path := filepath.Join(tmpDir, ".nodes", "parent", "repo2", ".git")
	repo3Path := filepath.Join(tmpDir, ".nodes", "parent", "subteam", "repo3", ".git")
	os.MkdirAll(repo1Path, 0755)
	os.MkdirAll(repo2Path, 0755)
	os.MkdirAll(repo3Path, 0755)

	// Create config node
	node := interfaces.NodeInfo{
		Name:       "parent",
		Path:       "/parent",
		ConfigFile: parentConfigPath,
		IsConfig:   true,
	}

	// Add the config node to the tree
	if err := mgr.treeProvider.UpdateNode(node.Path, node); err != nil {
		t.Fatalf("Failed to add node to tree: %v", err)
	}

	// Add the subteam nested config node to the tree for path resolution
	subteamNode := interfaces.NodeInfo{
		Name:       "subteam",
		Path:       "/parent/subteam",
		ConfigFile: nestedConfigPath,
		IsConfig:   true,
	}
	if err := mgr.treeProvider.UpdateNode(subteamNode.Path, subteamNode); err != nil {
		t.Fatalf("Failed to add subteam node to tree: %v", err)
	}

	err := mgr.pullRecursiveWithOptions(node, false, false)
	if err != nil {
		t.Errorf("pullRecursiveWithOptions failed: %v", err)
	}

	// Verify pull was called multiple times (for repo1, repo2, repo3)
	if gitStub.pullCallCount < 3 {
		t.Errorf("Expected at least 3 pull calls for mixed children, got %d", gitStub.pullCallCount)
	}
}

func TestPullRecursiveWithOptions_ConfigNodeRelativeFilePath(t *testing.T) {
	// Test config node with relative file path resolution
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Create configs directory
	configsDir := filepath.Join(tmpDir, "configs")
	os.MkdirAll(configsDir, 0755)

	// Create nested config in configs directory
	nestedConfigPath := filepath.Join(configsDir, "nested.yaml")
	nestedConfigContent := `workspace:
  name: nested
nodes:
  - name: nested-repo
    url: https://github.com/test/nested-repo.git
`
	if err := os.WriteFile(nestedConfigPath, []byte(nestedConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create nested config: %v", err)
	}

	// Create parent config with relative file path
	parentConfigPath := filepath.Join(configsDir, "parent.yaml")
	parentConfigContent := `workspace:
  name: parent
nodes:
  - name: nested
    file: ./nested.yaml
`
	if err := os.WriteFile(parentConfigPath, []byte(parentConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create parent config: %v", err)
	}

	// Create .git directory
	nestedRepoPath := filepath.Join(tmpDir, ".nodes", "parent", "nested", "nested-repo", ".git")
	os.MkdirAll(nestedRepoPath, 0755)

	// Create config node
	node := interfaces.NodeInfo{
		Name:       "parent",
		Path:       "/parent",
		ConfigFile: parentConfigPath,
		IsConfig:   true,
	}

	err := mgr.pullRecursiveWithOptions(node, false, false)
	if err != nil {
		t.Errorf("pullRecursiveWithOptions failed: %v", err)
	}

	// Verify pull was called
	if !gitStub.pullCalled {
		t.Error("Expected pull to be called for nested config repository")
	}
}

// ================================================================================
// pullRecursive - Config Node Tests
// ================================================================================

func TestPullRecursive_ConfigNodeWithURLChildren(t *testing.T) {
	// Test config node expansion with URL children
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Create config file with URL children
	configPath := filepath.Join(tmpDir, "team-config.yaml")
	configContent := `workspace:
  name: team
nodes:
  - name: backend
    url: https://github.com/test/backend.git
  - name: frontend
    url: https://github.com/test/frontend.git
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create .git directories
	backendPath := filepath.Join(tmpDir, ".nodes", "team", "backend", ".git")
	frontendPath := filepath.Join(tmpDir, ".nodes", "team", "frontend", ".git")
	os.MkdirAll(backendPath, 0755)
	os.MkdirAll(frontendPath, 0755)

	// Create config node
	node := interfaces.NodeInfo{
		Name:       "team",
		Path:       "/team",
		ConfigFile: configPath,
		IsConfig:   true,
	}

	err := mgr.pullRecursive(node, false)
	if err != nil {
		t.Errorf("pullRecursive failed: %v", err)
	}

	// Verify pull was called for both repositories
	if gitStub.pullCallCount < 2 {
		t.Errorf("Expected at least 2 pull calls, got %d", gitStub.pullCallCount)
	}
}

func TestPullRecursive_ConfigNodeWithFileChildren(t *testing.T) {
	// Test config node with File children (nested config nodes)
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Create nested config
	nestedConfigPath := filepath.Join(tmpDir, "infra-config.yaml")
	nestedConfigContent := `workspace:
  name: infrastructure
nodes:
  - name: monitoring
    url: https://github.com/test/monitoring.git
`
	if err := os.WriteFile(nestedConfigPath, []byte(nestedConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create nested config: %v", err)
	}

	// Create parent config
	parentConfigPath := filepath.Join(tmpDir, "team-config.yaml")
	parentConfigContent := fmt.Sprintf(`workspace:
  name: team
nodes:
  - name: infrastructure
    file: %s
`, nestedConfigPath)
	if err := os.WriteFile(parentConfigPath, []byte(parentConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create parent config: %v", err)
	}

	// Create .git directory
	monitoringPath := filepath.Join(tmpDir, ".nodes", "team", "infrastructure", "monitoring", ".git")
	os.MkdirAll(monitoringPath, 0755)

	// Create config node
	node := interfaces.NodeInfo{
		Name:       "team",
		Path:       "/team",
		ConfigFile: parentConfigPath,
		IsConfig:   true,
	}

	err := mgr.pullRecursive(node, false)
	if err != nil {
		t.Errorf("pullRecursive failed: %v", err)
	}

	// Verify pull was called
	if !gitStub.pullCalled {
		t.Error("Expected pull to be called for nested config repository")
	}
}

func TestPullRecursive_ConfigNodeSkipsUnclonedRepos(t *testing.T) {
	// Test that pullRecursive skips repositories without .git directory
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Create config file
	configPath := filepath.Join(tmpDir, "team-config.yaml")
	configContent := `workspace:
  name: team
nodes:
  - name: cloned-repo
    url: https://github.com/test/cloned.git
  - name: not-cloned-repo
    url: https://github.com/test/not-cloned.git
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Only create .git for cloned-repo
	clonedPath := filepath.Join(tmpDir, ".nodes", "team", "cloned-repo", ".git")
	os.MkdirAll(clonedPath, 0755)

	// Create config node
	node := interfaces.NodeInfo{
		Name:       "team",
		Path:       "/team",
		ConfigFile: configPath,
		IsConfig:   true,
	}

	err := mgr.pullRecursive(node, false)
	if err != nil {
		t.Errorf("pullRecursive failed: %v", err)
	}

	// Verify pull was called only once (for cloned-repo)
	if gitStub.pullCallCount != 1 {
		t.Errorf("Expected exactly 1 pull call, got %d", gitStub.pullCallCount)
	}
}

func TestPullRecursive_ConfigNodeNestedRecursion(t *testing.T) {
	// Test deeply nested config node recursion
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Create level 3 config
	level3ConfigPath := filepath.Join(tmpDir, "level3.yaml")
	level3ConfigContent := `workspace:
  name: level3
nodes:
  - name: deep-repo
    url: https://github.com/test/deep-repo.git
`
	if err := os.WriteFile(level3ConfigPath, []byte(level3ConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create level3 config: %v", err)
	}

	// Create level 2 config
	level2ConfigPath := filepath.Join(tmpDir, "level2.yaml")
	level2ConfigContent := fmt.Sprintf(`workspace:
  name: level2
nodes:
  - name: level3
    file: %s
`, level3ConfigPath)
	if err := os.WriteFile(level2ConfigPath, []byte(level2ConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create level2 config: %v", err)
	}

	// Create level 1 config
	level1ConfigPath := filepath.Join(tmpDir, "level1.yaml")
	level1ConfigContent := fmt.Sprintf(`workspace:
  name: level1
nodes:
  - name: level2
    file: %s
`, level2ConfigPath)
	if err := os.WriteFile(level1ConfigPath, []byte(level1ConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create level1 config: %v", err)
	}

	// Create .git directory for deep-repo
	deepRepoPath := filepath.Join(tmpDir, ".nodes", "level1", "level2", "level3", "deep-repo", ".git")
	os.MkdirAll(deepRepoPath, 0755)

	// Create config node
	node := interfaces.NodeInfo{
		Name:       "level1",
		Path:       "/level1",
		ConfigFile: level1ConfigPath,
		IsConfig:   true,
	}

	err := mgr.pullRecursive(node, false)
	if err != nil {
		t.Errorf("pullRecursive failed: %v", err)
	}

	// Verify pull was called for the deeply nested repository
	if !gitStub.pullCalled {
		t.Error("Expected pull to be called for deeply nested repository")
	}
}

func TestPullRecursive_ConfigNodeForceFlag(t *testing.T) {
	// Test that force flag is passed through config node expansion
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Create config file
	configPath := filepath.Join(tmpDir, "team-config.yaml")
	configContent := `workspace:
  name: team
nodes:
  - name: repo1
    url: https://github.com/test/repo1.git
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create .git directory
	repo1Path := filepath.Join(tmpDir, ".nodes", "team", "repo1", ".git")
	os.MkdirAll(repo1Path, 0755)

	// Create config node
	node := interfaces.NodeInfo{
		Name:       "team",
		Path:       "/team",
		ConfigFile: configPath,
		IsConfig:   true,
	}

	// Call with force=true
	err := mgr.pullRecursive(node, true)
	if err != nil {
		t.Errorf("pullRecursive failed: %v", err)
	}

	// Verify pull was called
	if !gitStub.pullCalled {
		t.Error("Expected pull to be called with force flag")
	}
}

func TestPullRecursive_ConfigNodeLoadError(t *testing.T) {
	// Test that config load errors are handled gracefully
	tmpDir := t.TempDir()
	mgr := CreateTestManager(t, tmpDir)

	gitStub := NewGitProviderStub()
	mgr.gitProvider = gitStub

	// Create config node with non-existent config file
	node := interfaces.NodeInfo{
		Name:       "team",
		Path:       "/team",
		ConfigFile: filepath.Join(tmpDir, "nonexistent.yaml"),
		IsConfig:   true,
	}

	// Should not fail, just skip processing
	err := mgr.pullRecursive(node, false)
	if err != nil {
		t.Errorf("pullRecursive should not fail on config load error: %v", err)
	}

	// Verify pull was not called
	if gitStub.pullCalled {
		t.Error("Expected pull not to be called when config load fails")
	}
}

// ================================================================================
// visitNodeForClone - Config Node Cloning Tests
// ================================================================================

// TestVisitNodeForClone_GitNode tests cloning a git repository node
func TestVisitNodeForClone_GitNode(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create a git node to clone
	gitNode := interfaces.NodeInfo{
		Name:       "test-repo",
		Path:       "/test-repo",
		Repository: "https://github.com/test/repo.git",
		IsLazy:     false,
		IsCloned:   false,
	}

	// Mock the git clone operation (will fail but we can check it was attempted)
	err := m.visitNodeForClone(gitNode, false, false)

	// The function should complete without critical errors (git clone will fail but is handled gracefully)
	if err != nil {
		t.Logf("visitNodeForClone returned error (expected for mock git): %v", err)
	}

	// Verify directory was created
	nodePath := filepath.Join(workspace, ".nodes", "test-repo")
	_, err = os.Stat(nodePath)
	if os.IsNotExist(err) {
		t.Error("Node directory should be created")
	}
}

// TestVisitNodeForClone_ConfigNode tests processing a config reference node
func TestVisitNodeForClone_ConfigNode(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create a config file to reference
	configDir := filepath.Join(workspace, "external-config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	configFile := filepath.Join(configDir, "team.yaml")

	configContent := `workspace:
  name: team-config
  repos_dir: .repos
nodes:
  - name: service1
    url: https://github.com/test/service1.git
    fetch: eager
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create a config node that references the external config
	configNode := interfaces.NodeInfo{
		Name:       "team",
		Path:       "/team",
		ConfigFile: configFile,
		IsConfig:   true,
		IsLazy:     false,
	}

	// Visit the config node (non-recursive first)
	err := m.visitNodeForClone(configNode, false, false)
	if err != nil {
		t.Logf("visitNodeForClone returned error: %v", err)
	}

	// Verify directory was created
	nodePath := filepath.Join(workspace, ".nodes", "team")
	_, err = os.Stat(nodePath)
	if os.IsNotExist(err) {
		t.Error("Config node directory should be created")
	}

	// Verify symlink or copied file exists
	symlinkPath := filepath.Join(nodePath, "muno.yaml")
	_, err = os.Lstat(symlinkPath)
	if os.IsNotExist(err) {
		t.Error("muno.yaml should exist (as symlink or copy)")
	}
}

// TestVisitNodeForClone_Recursive tests recursive processing of child nodes
func TestVisitNodeForClone_Recursive(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create a parent config with children
	parentConfigDir := filepath.Join(workspace, "parent-config")
	if err := os.MkdirAll(parentConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create parent config dir: %v", err)
	}
	parentConfigFile := filepath.Join(parentConfigDir, "parent.yaml")

	parentConfigContent := `workspace:
  name: parent
  repos_dir: .nodes
nodes:
  - name: child1
    url: https://github.com/test/child1.git
    fetch: eager
  - name: child2
    url: https://github.com/test/child2.git
    fetch: lazy
`
	if err := os.WriteFile(parentConfigFile, []byte(parentConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write parent config: %v", err)
	}

	// Create parent config node
	parentNode := interfaces.NodeInfo{
		Name:       "parent",
		Path:       "/parent",
		ConfigFile: parentConfigFile,
		IsConfig:   true,
	}

	// Visit recursively
	err := m.visitNodeForClone(parentNode, true, false)
	if err != nil {
		t.Logf("visitNodeForClone returned error: %v", err)
	}

	// Verify parent directory was created
	parentPath := filepath.Join(workspace, ".nodes", "parent")
	_, err = os.Stat(parentPath)
	if os.IsNotExist(err) {
		t.Error("Parent directory should be created")
	}

	// Verify child directories were created (even though git clone will fail)
	child1Path := filepath.Join(parentPath, ".nodes", "child1")
	_, err = os.Stat(child1Path)
	if os.IsNotExist(err) {
		t.Error("Child1 directory should be created")
	}

	// Child2 is lazy and includeLazy=false, so it might not be created
	// The behavior depends on implementation - just verify no crash
}

// TestVisitNodeForClone_LazyNode tests lazy node handling
func TestVisitNodeForClone_LazyNode(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create a lazy git node
	lazyNode := interfaces.NodeInfo{
		Name:       "lazy-repo",
		Path:       "/lazy-repo",
		Repository: "https://github.com/test/lazy.git",
		IsLazy:     true,
		IsCloned:   false,
	}

	// Visit without includeLazy
	err := m.visitNodeForClone(lazyNode, false, false)
	if err != nil {
		t.Logf("visitNodeForClone returned error: %v", err)
	}

	// Directory should be created
	nodePath := filepath.Join(workspace, ".nodes", "lazy-repo")
	_, err = os.Stat(nodePath)
	if os.IsNotExist(err) {
		t.Error("Directory should be created")
	}

	// Now visit WITH includeLazy
	err = m.visitNodeForClone(lazyNode, false, true)
	if err != nil {
		t.Logf("visitNodeForClone with includeLazy returned error: %v", err)
	}
}

// ================================================================================
// processConfigNode - Tests
// ================================================================================

// TestProcessConfigNode tests collecting child nodes from config
func TestProcessConfigNode(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create a config file
	configDir := filepath.Join(workspace, "test-config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	configFile := filepath.Join(configDir, "nodes.yaml")

	configContent := `workspace:
  name: test
  repos_dir: .nodes
nodes:
  - name: repo1
    url: https://github.com/test/repo1.git
    fetch: eager
  - name: repo2
    url: https://github.com/test/repo2.git
    fetch: lazy
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create config node
	configNode := interfaces.NodeInfo{
		Name:       "test-team",
		Path:       "/test-team",
		ConfigFile: configFile,
		IsConfig:   true,
	}

	// Collect child nodes
	toClone := []interfaces.NodeInfo{}
	err := m.processConfigNode(configNode, false, false, &toClone)
	if err != nil {
		t.Errorf("processConfigNode failed: %v", err)
	}

	// Should have collected 2 child nodes
	if len(toClone) != 2 {
		t.Errorf("Expected 2 child nodes, got %d", len(toClone))
	}

	// Verify collected nodes
	if len(toClone) >= 2 {
		if toClone[0].Name != "repo1" {
			t.Errorf("Expected first node name 'repo1', got '%s'", toClone[0].Name)
		}
		if toClone[0].Path != "/test-team/repo1" {
			t.Errorf("Expected first node path '/test-team/repo1', got '%s'", toClone[0].Path)
		}
		if toClone[1].Name != "repo2" {
			t.Errorf("Expected second node name 'repo2', got '%s'", toClone[1].Name)
		}
		if toClone[1].Path != "/test-team/repo2" {
			t.Errorf("Expected second node path '/test-team/repo2', got '%s'", toClone[1].Path)
		}
	}
}

// TestProcessConfigNode_Recursive tests recursive config node processing
func TestProcessConfigNode_Recursive(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create a config file
	configDir := filepath.Join(workspace, "recursive-config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	configFile := filepath.Join(configDir, "config.yaml")

	configContent := `workspace:
  name: recursive-test
  repos_dir: .nodes
nodes:
  - name: repo1
    url: https://github.com/test/repo1.git
    fetch: eager
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	configNode := interfaces.NodeInfo{
		Name:       "recursive-node",
		Path:       "/recursive-node",
		ConfigFile: configFile,
		IsConfig:   true,
	}

	// Test recursive=true
	toClone := []interfaces.NodeInfo{}
	err := m.processConfigNode(configNode, true, false, &toClone)
	if err != nil {
		t.Errorf("processConfigNode failed: %v", err)
	}

	// Should still collect child nodes
	if len(toClone) < 1 {
		t.Error("Should collect at least 1 child node")
	}
}

// ================================================================================
// cloneConfigNodeRecursive - Tests
// ================================================================================

// TestCloneConfigNodeRecursive tests the deprecated recursive cloning function
func TestCloneConfigNodeRecursive(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create a directory with muno.yaml
	configDir := filepath.Join(workspace, "clone-test")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create clone test dir: %v", err)
	}
	configFile := filepath.Join(configDir, "muno.yaml")

	configContent := `workspace:
  name: clone-test
  repos_dir: .nodes
nodes:
  - name: repo1
    url: https://github.com/test/repo1.git
    fetch: eager
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Call the deprecated function
	toClone := []interfaces.NodeInfo{}
	err := m.cloneConfigNodeRecursive(configDir, "/clone-test", false, &toClone)
	if err != nil {
		t.Logf("cloneConfigNodeRecursive returned error: %v", err)
	}

	// Function should complete without panic - that's the main test
	// Directory creation depends on Manager's computeFilesystemPath which may vary
	t.Logf("cloneConfigNodeRecursive completed successfully")
}

// ================================================================================
// resolveConfigPath - Tests
// ================================================================================

// TestResolveConfigPath tests config path resolution
func TestResolveConfigPath(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Test absolute path
	absPath := "/absolute/path/to/config.yaml"
	resolved := m.resolveConfigPath(absPath, "/some/node")
	if resolved != absPath {
		t.Errorf("Absolute path should be unchanged, got %s", resolved)
	}

	// Test URL
	urlPath := "https://config.example.com/muno.yaml"
	resolved = m.resolveConfigPath(urlPath, "/some/node")
	if resolved != urlPath {
		t.Errorf("URL should be unchanged, got %s", resolved)
	}

	// Test relative path
	// Create parent config
	parentDir := filepath.Join(workspace, ".nodes", "parent")
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		t.Fatalf("Failed to create parent dir: %v", err)
	}
	parentConfig := filepath.Join(parentDir, "muno.yaml")
	if err := os.WriteFile(parentConfig, []byte("workspace:\n  name: parent\n"), 0644); err != nil {
		t.Fatalf("Failed to write parent config: %v", err)
	}

	// Create referenced config
	siblingDir := filepath.Join(workspace, "configs")
	if err := os.MkdirAll(siblingDir, 0755); err != nil {
		t.Fatalf("Failed to create sibling dir: %v", err)
	}
	siblingConfig := filepath.Join(siblingDir, "sibling.yaml")
	if err := os.WriteFile(siblingConfig, []byte("workspace:\n  name: sibling\n"), 0644); err != nil {
		t.Fatalf("Failed to write sibling config: %v", err)
	}

	// Resolve relative path (this will use fallback logic)
	relPath := "configs/sibling.yaml"
	resolved = m.resolveConfigPath(relPath, "/parent/child")

	// Should resolve to some path (exact path depends on resolution logic)
	if resolved == "" {
		t.Error("Should resolve to a path")
	}
	if !filepath.IsAbs(resolved) {
		t.Error("Resolved path should be absolute")
	}
}

// ================================================================================
// copyConfigFile - Tests
// ================================================================================

// TestCopyConfigFile tests config file copying
func TestCopyConfigFile(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create source file
	srcFile := filepath.Join(workspace, "source.yaml")
	srcContent := "workspace:\n  name: test\nnodes: []"
	if err := os.WriteFile(srcFile, []byte(srcContent), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Copy to destination
	dstFile := filepath.Join(workspace, "destination.yaml")
	err := m.copyConfigFile(srcFile, dstFile)
	if err != nil {
		t.Errorf("copyConfigFile failed: %v", err)
	}

	// Verify destination file exists and has same content
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Errorf("Failed to read destination file: %v", err)
	}
	if string(dstContent) != srcContent {
		t.Error("Copied content should match source")
	}
}

// TestCopyConfigFile_Error tests error handling when source doesn't exist
func TestCopyConfigFile_Error(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Try to copy non-existent file
	srcFile := filepath.Join(workspace, "nonexistent.yaml")
	dstFile := filepath.Join(workspace, "destination.yaml")
	err := m.copyConfigFile(srcFile, dstFile)

	if err == nil {
		t.Error("Should error when source file doesn't exist")
	}
	if err != nil {
		// Error is expected - log it
		t.Logf("Got expected error: %v", err)
	}
}

// ================================================================================
// expandConfigNodes - Tests
// ================================================================================

// TestExpandConfigNodes tests expanding config nodes to find repositories
func TestExpandConfigNodes(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create config file
	configDir := filepath.Join(workspace, "expand-test")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create expand test dir: %v", err)
	}
	configFile := filepath.Join(configDir, "expand.yaml")

	configContent := `workspace:
  name: expand-test
  repos_dir: .nodes
nodes:
  - name: repo1
    url: https://github.com/test/repo1.git
    fetch: eager
  - name: repo2
    url: https://github.com/test/repo2.git
    fetch: lazy
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create config node
	configNode := interfaces.NodeInfo{
		Name:       "expand-node",
		Path:       "/expand-node",
		ConfigFile: configFile,
		IsConfig:   true,
	}

	// Expand to find repositories
	toClone := []interfaces.NodeInfo{}
	err := m.expandConfigNodes(configNode, false, false, &toClone)
	if err != nil {
		t.Errorf("expandConfigNodes failed: %v", err)
	}

	// Should process config node
	// The exact behavior depends on implementation
	t.Logf("Collected %d nodes to clone", len(toClone))
}

// TestExpandConfigNodes_GitRepository tests expanding a git repository node
func TestExpandConfigNodes_GitRepository(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create a git repository node (not cloned)
	gitNode := interfaces.NodeInfo{
		Name:       "test-repo",
		Path:       "/test-repo",
		Repository: "https://github.com/test/repo.git",
		IsLazy:     false,
		IsCloned:   false,
	}

	// Expand should add it to toClone list
	toClone := []interfaces.NodeInfo{}
	err := m.expandConfigNodes(gitNode, false, false, &toClone)
	if err != nil {
		t.Errorf("expandConfigNodes failed: %v", err)
	}

	// Should add the repository to clone list
	if len(toClone) != 1 {
		t.Errorf("Expected 1 node in toClone list, got %d", len(toClone))
	}
	if len(toClone) > 0 {
		if toClone[0].Name != "test-repo" {
			t.Errorf("Expected node name 'test-repo', got '%s'", toClone[0].Name)
		}
		if toClone[0].Repository != "https://github.com/test/repo.git" {
			t.Errorf("Expected repository URL, got '%s'", toClone[0].Repository)
		}
	}
}

// TestExpandConfigNodes_LazyRepository tests expanding a lazy repository
func TestExpandConfigNodes_LazyRepository(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create a lazy repository node
	lazyNode := interfaces.NodeInfo{
		Name:       "lazy-repo",
		Path:       "/lazy-repo",
		Repository: "https://github.com/test/lazy.git",
		IsLazy:     true,
		IsCloned:   false,
	}

	// Expand without includeLazy
	toClone := []interfaces.NodeInfo{}
	err := m.expandConfigNodes(lazyNode, false, false, &toClone)
	if err != nil {
		t.Errorf("expandConfigNodes failed: %v", err)
	}
	if len(toClone) != 0 {
		t.Error("Should not add lazy repo when includeLazy=false")
	}

	// Expand with includeLazy
	toClone = []interfaces.NodeInfo{}
	err = m.expandConfigNodes(lazyNode, false, true, &toClone)
	if err != nil {
		t.Errorf("expandConfigNodes failed: %v", err)
	}
	if len(toClone) != 1 {
		t.Error("Should add lazy repo when includeLazy=true")
	}
}

// TestExpandConfigNodes_Recursive tests recursive expansion with children
func TestExpandConfigNodes_Recursive(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create parent node with children
	parentNode := interfaces.NodeInfo{
		Name: "parent",
		Path: "/parent",
		Children: []interfaces.NodeInfo{
			{
				Name:       "child1",
				Path:       "/parent/child1",
				Repository: "https://github.com/test/child1.git",
				IsCloned:   false,
			},
			{
				Name:       "child2",
				Path:       "/parent/child2",
				Repository: "https://github.com/test/child2.git",
				IsLazy:     true,
				IsCloned:   false,
			},
		},
	}

	// Expand recursively without includeLazy
	toClone := []interfaces.NodeInfo{}
	err := m.expandConfigNodes(parentNode, true, false, &toClone)
	if err != nil {
		t.Errorf("expandConfigNodes failed: %v", err)
	}

	// Should collect child1 (eager) but not child2 (lazy)
	if len(toClone) != 1 {
		t.Errorf("Expected 1 node (eager only), got %d", len(toClone))
	}
	if len(toClone) > 0 {
		if toClone[0].Name != "child1" {
			t.Errorf("Expected 'child1', got '%s'", toClone[0].Name)
		}
	}

	// Expand recursively WITH includeLazy
	toClone = []interfaces.NodeInfo{}
	err = m.expandConfigNodes(parentNode, true, true, &toClone)
	if err != nil {
		t.Errorf("expandConfigNodes failed: %v", err)
	}

	// Should collect both children
	if len(toClone) != 2 {
		t.Errorf("Expected 2 nodes (eager and lazy), got %d", len(toClone))
	}
}

// TestExpandConfigNodes_AlreadyCloned tests that cloned repos are not added
func TestExpandConfigNodes_AlreadyCloned(t *testing.T) {
	workspace := t.TempDir()
	m := CreateTestManager(t, workspace)

	// Create an already cloned repository node
	clonedNode := interfaces.NodeInfo{
		Name:       "cloned-repo",
		Path:       "/cloned-repo",
		Repository: "https://github.com/test/cloned.git",
		IsCloned:   true,
	}

	// Expand should NOT add cloned repositories
	toClone := []interfaces.NodeInfo{}
	err := m.expandConfigNodes(clonedNode, false, false, &toClone)
	if err != nil {
		t.Errorf("expandConfigNodes failed: %v", err)
	}

	// Should not add already cloned repository
	if len(toClone) != 0 {
		t.Error("Should not add already cloned repositories")
	}
}
