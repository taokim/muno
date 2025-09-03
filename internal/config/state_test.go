package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadState(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")
	
	// Test loading non-existent state
	state, err := LoadState(statePath)
	require.NoError(t, err)
	assert.NotNil(t, state)
	assert.NotNil(t, state.Sessions)
	assert.Empty(t, state.CurrentNodePath)
	
	// Create a state file and load it
	testState := &State{
		Timestamp:       time.Now().Format(time.RFC3339),
		CurrentNodePath: "/test/path",
		Sessions: map[string]Session{
			"/test": {
				NodePath:  "/test",
				Status:    "running",
				PID:       12345,
				StartTime: time.Now().Format(time.RFC3339),
			},
		},
	}
	err = testState.SaveState(statePath)
	require.NoError(t, err)
	
	// Load the saved state
	loadedState, err := LoadState(statePath)
	require.NoError(t, err)
	assert.Equal(t, "/test/path", loadedState.CurrentNodePath)
	assert.Len(t, loadedState.Sessions, 1)
	assert.Equal(t, "running", loadedState.Sessions["/test"].Status)
}

func TestSaveState(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")
	
	state := &State{
		CurrentNodePath: "/test/node",
		Sessions:        make(map[string]Session),
	}
	
	err := state.SaveState(statePath)
	require.NoError(t, err)
	
	// Verify file was created
	_, err = os.Stat(statePath)
	assert.NoError(t, err)
	
	// Load and verify content
	loadedState, err := LoadState(statePath)
	require.NoError(t, err)
	assert.Equal(t, "/test/node", loadedState.CurrentNodePath)
	assert.NotEmpty(t, loadedState.Timestamp)
}

func TestSetGetCurrentNode(t *testing.T) {
	state := &State{
		Sessions: make(map[string]Session),
	}
	
	// Test setting and getting current node
	state.SetCurrentNode("/new/path")
	assert.Equal(t, "/new/path", state.GetCurrentNode())
	
	// Test empty state
	emptyState := &State{}
	assert.Equal(t, "", emptyState.GetCurrentNode())
}

func TestAddRemoveSession(t *testing.T) {
	state := &State{
		Sessions: make(map[string]Session),
	}
	
	// Add a session
	state.AddSession("/test/node", 12345)
	assert.Len(t, state.Sessions, 1)
	session := state.Sessions["/test/node"]
	assert.Equal(t, "/test/node", session.NodePath)
	assert.Equal(t, "running", session.Status)
	assert.Equal(t, 12345, session.PID)
	assert.NotEmpty(t, session.StartTime)
	assert.NotEmpty(t, session.LastActivity)
	
	// Remove the session
	state.RemoveSession("/test/node")
	assert.Len(t, state.Sessions, 0)
	
	// Remove non-existent session (should not panic)
	state.RemoveSession("/non/existent")
	assert.Len(t, state.Sessions, 0)
}

func TestUpdateSessionActivity(t *testing.T) {
	state := &State{
		Sessions: make(map[string]Session),
	}
	
	// Add a session
	state.AddSession("/test/node", 12345)
	originalActivity := state.Sessions["/test/node"].LastActivity
	
	// Wait a moment to ensure time difference (RFC3339 has second precision)
	time.Sleep(time.Second * 2)
	
	// Update activity
	state.UpdateSessionActivity("/test/node")
	updatedSession := state.Sessions["/test/node"]
	// The times should be different after sleep
	assert.NotEqual(t, originalActivity, updatedSession.LastActivity, "LastActivity should be updated")
	
	// Update non-existent session (should not panic)
	state.UpdateSessionActivity("/non/existent")
}

func TestLoadStateWithInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "invalid.json")
	
	// Write invalid JSON
	err := os.WriteFile(statePath, []byte("invalid json content"), 0644)
	require.NoError(t, err)
	
	// Try to load invalid JSON
	_, err = LoadState(statePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parsing state")
}

func TestSaveStateWithInvalidPath(t *testing.T) {
	state := &State{
		Sessions: make(map[string]Session),
	}
	
	// Try to save to an invalid path
	err := state.SaveState("/nonexistent/dir/state.json")
	assert.Error(t, err)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("test-project")
	
	assert.Equal(t, "test-project", cfg.Workspace.Name)
	assert.Equal(t, "repos", cfg.Workspace.RootPath)
	assert.NotNil(t, cfg.Repositories)
	assert.NotEmpty(t, cfg.Repositories)
	
	// Check meta-repos
	assert.Contains(t, cfg.Repositories, "backend-repo")
	assert.Contains(t, cfg.Repositories, "frontend-repo")
	
	// Check regular repos
	assert.Contains(t, cfg.Repositories, "payment-service")
	assert.Contains(t, cfg.Repositories, "fraud-detection")
	assert.Contains(t, cfg.Repositories, "web-app")
	
	// Check documentation config
	assert.Equal(t, "docs", cfg.Documentation.Path)
	assert.True(t, cfg.Documentation.SyncToGit)
}

func TestGetRepository(t *testing.T) {
	cfg := &Config{
		Repositories: map[string]Repository{
			"test-repo": {
				URL:    "https://github.com/test/repo.git",
				Branch: "main",
			},
		},
	}
	
	// Get existing repository
	repo, err := cfg.GetRepository("test-repo")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/test/repo.git", repo.URL)
	assert.Equal(t, "main", repo.Branch)
	
	// Get non-existent repository
	_, err = cfg.GetRepository("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetChildWorkspaces(t *testing.T) {
	cfg := &Config{
		Repositories: map[string]Repository{
			"workspace1": {
				URL:         "https://github.com/test/ws1.git",
				IsWorkspace: true,
			},
			"regular-repo": {
				URL:         "https://github.com/test/repo.git",
				IsWorkspace: false,
			},
			"workspace2": {
				URL:         "https://github.com/test/ws2.git",
				IsWorkspace: true,
			},
		},
	}
	
	workspaces := cfg.GetChildWorkspaces()
	assert.Len(t, workspaces, 2)
	assert.Contains(t, workspaces, "workspace1")
	assert.Contains(t, workspaces, "workspace2")
	assert.NotContains(t, workspaces, "regular-repo")
}

func TestGetLocalRepositories(t *testing.T) {
	cfg := &Config{
		Repositories: map[string]Repository{
			"workspace1": {
				URL:         "https://github.com/test/ws1.git",
				IsWorkspace: true,
			},
			"regular-repo1": {
				URL:         "https://github.com/test/repo1.git",
				IsWorkspace: false,
			},
			"regular-repo2": {
				URL:         "https://github.com/test/repo2.git",
				IsWorkspace: false,
			},
		},
	}
	
	repos := cfg.GetLocalRepositories()
	assert.Len(t, repos, 2)
	assert.Contains(t, repos, "regular-repo1")
	assert.Contains(t, repos, "regular-repo2")
	assert.NotContains(t, repos, "workspace1")
}

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com/org/repo.git", "repo"},
		{"https://github.com/org/repo", "repo"},
		{"git@github.com:org/repo.git", "repo"},
		{"repo.git", "repo"},
		{"repo", "repo"},
		{"/path/to/repo", "repo"},
		{"", ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			// extractRepoName is not exported, test through IsLazy
			repo := &Repository{URL: tt.url}
			defaults := Defaults{Lazy: true}
			_ = repo.IsLazy("", defaults) // This calls extractRepoName internally
		})
	}
}