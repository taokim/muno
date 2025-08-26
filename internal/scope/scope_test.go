package scope

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/repo-claude/internal/config"
)

func TestScopeOperations(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create a scope first
	scopeName := "test-ops"
	err := mgr.Create(scopeName, CreateOptions{
		Type:        TypePersistent,
		Repos:       []string{"test-repo"},
		Description: "Test operations",
	})
	require.NoError(t, err)
	
	scope, err := mgr.Get(scopeName)
	require.NoError(t, err)
	
	t.Run("GetPath", func(t *testing.T) {
		path := scope.GetPath()
		expectedPath := filepath.Join(mgr.workspacePath, scopeName)
		assert.Equal(t, expectedPath, path)
	})
	
	t.Run("GetMeta", func(t *testing.T) {
		meta := scope.GetMeta()
		assert.Equal(t, scopeName, meta.Name)
		assert.Equal(t, TypePersistent, meta.Type)
		assert.Equal(t, "Test operations", meta.Description)
	})
	
	t.Run("createScopeCLAUDE", func(t *testing.T) {
		// This is an internal method but we can test it indirectly
		// by checking if CLAUDE.md exists after scope creation
		claudePath := filepath.Join(scope.path, "CLAUDE.md")
		
		// Create the CLAUDE.md file
		err := scope.createScopeCLAUDE()
		assert.NoError(t, err)
		
		// Check file exists
		assert.FileExists(t, claudePath)
	})
}

func TestScopeStart(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create a scope
	scopeName := "test-start"
	err := mgr.Create(scopeName, CreateOptions{
		Type:  TypePersistent,
		Repos: []string{"test-repo"},
	})
	require.NoError(t, err)
	
	scope, err := mgr.Get(scopeName)
	require.NoError(t, err)
	
	// Test start (will fail on clone but should update metadata)
	opts := StartOptions{
		NewWindow: false,
		Pull:      false,
	}
	
	_ = scope.Start(opts) // Ignore error as repos don't exist
	
	// Check metadata was updated
	updatedScope, _ := mgr.Get(scopeName)
	assert.Equal(t, StateActive, updatedScope.meta.State)
}

func TestScopeStatus(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create a scope with repos
	scopeName := "test-status"
	err := mgr.Create(scopeName, CreateOptions{
		Type:        TypePersistent,
		Repos:       []string{"test-repo"},
		Description: "Status test scope",
	})
	require.NoError(t, err)
	
	scope, err := mgr.Get(scopeName)
	require.NoError(t, err)
	
	// Create a dummy repo directory (simulate clone)
	repoPath := filepath.Join(scope.path, "test-repo")
	err = os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)
	
	// Create a .git directory to simulate a git repo
	gitPath := filepath.Join(repoPath, ".git")
	err = os.MkdirAll(gitPath, 0755)
	require.NoError(t, err)
	
	// Get status
	report, err := scope.Status()
	require.NoError(t, err)
	
	assert.Equal(t, scopeName, report.Name)
	assert.Equal(t, TypePersistent, report.Type)
	assert.Equal(t, StateInactive, report.State)
	assert.NotNil(t, report.CreatedAt)
	assert.Len(t, report.Repos, 1)
}

func TestScopeCommit(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create a scope
	scopeName := "test-commit"
	err := mgr.Create(scopeName, CreateOptions{
		Type:  TypePersistent,
		Repos: []string{"test-repo"},
	})
	require.NoError(t, err)
	
	scope, err := mgr.Get(scopeName)
	require.NoError(t, err)
	
	// Test commit (will fail as repos don't exist, but should handle gracefully)
	err = scope.Commit("Test commit")
	// Expected to fail
	assert.Error(t, err)
}

func TestScopePull(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create a scope
	scopeName := "test-pull"
	err := mgr.Create(scopeName, CreateOptions{
		Type:  TypePersistent,
		Repos: []string{"test-repo"},
	})
	require.NoError(t, err)
	
	scope, err := mgr.Get(scopeName)
	require.NoError(t, err)
	
	// Test pull (will fail as repos don't exist, but should handle gracefully)
	opts := PullOptions{
		CloneMissing: true,
		Parallel:     false,
	}
	
	err = scope.Pull(opts)
	// Expected to fail as repos don't exist
	assert.Error(t, err)
}

func TestScopePush(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create a scope
	scopeName := "test-push"
	err := mgr.Create(scopeName, CreateOptions{
		Type:  TypePersistent,
		Repos: []string{"test-repo"},
	})
	require.NoError(t, err)
	
	scope, err := mgr.Get(scopeName)
	require.NoError(t, err)
	
	// Test push (will fail as repos don't exist, but should handle gracefully)
	err = scope.Push()
	// Expected to fail
	assert.Error(t, err)
}

func TestScopeSwitchBranch(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create a scope
	scopeName := "test-branch"
	err := mgr.Create(scopeName, CreateOptions{
		Type:  TypePersistent,
		Repos: []string{"test-repo"},
	})
	require.NoError(t, err)
	
	scope, err := mgr.Get(scopeName)
	require.NoError(t, err)
	
	// Test branch switch (will fail as repos don't exist, but should handle gracefully)
	err = scope.SwitchBranch("feature-branch")
	// Expected to fail
	assert.Error(t, err)
}

func TestScopeClone(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create a scope
	scopeName := "test-clone"
	err := mgr.Create(scopeName, CreateOptions{
		Type:        TypePersistent,
		Repos:       []string{"test-repo"},
		Description: "Clone test",
	})
	require.NoError(t, err)
	
	scope, err := mgr.Get(scopeName)
	require.NoError(t, err)
	
	// Test clone (will fail as repos don't exist remotely, but should handle gracefully)
	err = scope.Clone([]string{"test-repo"})
	// Expected to fail as the repo URL doesn't exist
	assert.Error(t, err)
}

func TestCreateFromConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	
	// Add a new scope to config
	cfg.Scopes["from-config"] = config.Scope{
		Type:        "ephemeral",
		Repos:       []string{"test-repo", "repo1"},
		Description: "Created from config",
		Model:       "claude-3-sonnet",
		AutoStart:   true,
	}
	
	mgr, _ := NewManager(cfg, tmpDir)
	
	// Create from config
	err := mgr.CreateFromConfig("from-config")
	require.NoError(t, err)
	
	// Verify it was created correctly
	scope, err := mgr.Get("from-config")
	require.NoError(t, err)
	
	assert.Equal(t, "from-config", scope.meta.Name)
	assert.Equal(t, TypeEphemeral, scope.meta.Type)
	assert.Equal(t, "Created from config", scope.meta.Description)
	assert.Len(t, scope.meta.Repos, 2)
	
	// Try to create non-existent scope from config
	err = mgr.CreateFromConfig("non-existent")
	assert.Error(t, err)
}

func TestGetWorkspacePath(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	mgr, _ := NewManager(cfg, tmpDir)
	
	expectedPath := filepath.Join(tmpDir, "workspaces")
	assert.Equal(t, expectedPath, mgr.GetWorkspacePath())
}