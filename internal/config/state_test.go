package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadState(t *testing.T) {
	tmpDir := t.TempDir()
	
	t.Run("NonExistentFile", func(t *testing.T) {
		statePath := filepath.Join(tmpDir, "nonexistent.json")
		state, err := LoadState(statePath)
		require.NoError(t, err)
		assert.NotNil(t, state)
		assert.NotNil(t, state.Agents)
		assert.Len(t, state.Agents, 0)
	})
	
	t.Run("ValidStateFile", func(t *testing.T) {
		statePath := filepath.Join(tmpDir, "valid-state.json")
		
		// Create a valid state file
		testState := State{
			Timestamp: time.Now().Format(time.RFC3339),
			Agents: map[string]AgentStatus{
				"test-agent": {
					Name:         "test-agent",
					Status:       "running",
					PID:          12345,
					Repository:   "test-repo",
					LastActivity: time.Now().Format(time.RFC3339),
				},
			},
		}
		
		data, err := json.MarshalIndent(testState, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(statePath, data, 0644)
		require.NoError(t, err)
		
		// Load and verify
		loaded, err := LoadState(statePath)
		require.NoError(t, err)
		assert.Equal(t, testState.Timestamp, loaded.Timestamp)
		assert.Len(t, loaded.Agents, 1)
		assert.Equal(t, testState.Agents["test-agent"].Name, loaded.Agents["test-agent"].Name)
		assert.Equal(t, testState.Agents["test-agent"].Status, loaded.Agents["test-agent"].Status)
		assert.Equal(t, testState.Agents["test-agent"].PID, loaded.Agents["test-agent"].PID)
	})
	
	t.Run("InvalidJSON", func(t *testing.T) {
		statePath := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(statePath, []byte("invalid json content"), 0644)
		require.NoError(t, err)
		
		_, err = LoadState(statePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parsing state")
	})
	
	t.Run("EmptyAgentsMap", func(t *testing.T) {
		statePath := filepath.Join(tmpDir, "empty-agents.json")
		testState := State{
			Timestamp: time.Now().Format(time.RFC3339),
			// Agents map is nil
		}
		
		data, err := json.MarshalIndent(testState, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(statePath, data, 0644)
		require.NoError(t, err)
		
		loaded, err := LoadState(statePath)
		require.NoError(t, err)
		assert.NotNil(t, loaded.Agents) // Should be initialized
	})
}

func TestStateSave(t *testing.T) {
	tmpDir := t.TempDir()
	
	t.Run("SaveNewState", func(t *testing.T) {
		statePath := filepath.Join(tmpDir, "new-state.json")
		
		state := &State{
			Agents: map[string]AgentStatus{
				"agent1": {
					Name:       "agent1",
					Status:     "running",
					PID:        9999,
					Repository: "repo1",
				},
			},
		}
		
		err := state.Save(statePath)
		require.NoError(t, err)
		
		// Verify file was created
		assert.FileExists(t, statePath)
		
		// Verify timestamp was set
		assert.NotEmpty(t, state.Timestamp)
		
		// Load and verify content
		loaded, err := LoadState(statePath)
		require.NoError(t, err)
		assert.Equal(t, state.Timestamp, loaded.Timestamp)
		assert.Len(t, loaded.Agents, 1)
	})
	
	t.Run("SaveToInvalidPath", func(t *testing.T) {
		state := &State{
			Agents: make(map[string]AgentStatus),
		}
		
		// Try to save to a path that cannot be created
		invalidPath := "/root/cannot-create/state.json"
		err := state.Save(invalidPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "writing state file")
	})
}

func TestUpdateAgent(t *testing.T) {
	state := &State{
		Agents: make(map[string]AgentStatus),
	}
	
	t.Run("AddNewAgent", func(t *testing.T) {
		agent := AgentStatus{
			Name:       "new-agent",
			Status:     "running",
			PID:        1234,
			Repository: "new-repo",
		}
		
		state.UpdateAgent(agent)
		
		assert.Len(t, state.Agents, 1)
		stored := state.Agents["new-agent"]
		assert.Equal(t, agent.Name, stored.Name)
		assert.Equal(t, agent.Status, stored.Status)
		assert.NotEmpty(t, stored.LastActivity)
	})
	
	t.Run("UpdateExistingAgent", func(t *testing.T) {
		// First add an agent
		agent1 := AgentStatus{
			Name:       "existing-agent",
			Status:     "running",
			PID:        5555,
			Repository: "repo1",
		}
		state.UpdateAgent(agent1)
		// firstActivity := state.Agents["existing-agent"].LastActivity
		
		// Wait a bit to ensure different timestamp
		time.Sleep(100 * time.Millisecond)
		
		// Update the agent
		agent2 := AgentStatus{
			Name:       "existing-agent",
			Status:     "stopped",
			PID:        0,
			Repository: "repo1",
		}
		state.UpdateAgent(agent2)
		
		assert.Len(t, state.Agents, 2) // Should still have 2 agents total
		updated := state.Agents["existing-agent"]
		assert.Equal(t, "stopped", updated.Status)
		assert.Equal(t, 0, updated.PID)
		// Check that LastActivity was updated (should have a value)
		assert.NotEmpty(t, updated.LastActivity)
	})
}

func TestUpdateScope(t *testing.T) {
	t.Run("AddNewScope", func(t *testing.T) {
		state := &State{
			Scopes: make(map[string]ScopeStatus),
		}
		scope := ScopeStatus{
			Name:   "new-scope",
			Status: "running",
			PID:    1234,
			Repos:  []string{"repo1", "repo2"},
		}
		
		state.UpdateScope(scope)
		
		assert.Len(t, state.Scopes, 1)
		stored := state.Scopes["new-scope"]
		assert.Equal(t, scope.Name, stored.Name)
		assert.Equal(t, scope.Status, stored.Status)
		assert.NotEmpty(t, stored.LastActivity)
		// Check that CurrentRepos and CurrentScope are initialized
		assert.Equal(t, scope.Repos, stored.CurrentRepos)
		assert.Equal(t, scope.Name, stored.CurrentScope)
	})
	
	t.Run("UpdateExistingScope", func(t *testing.T) {
		state := &State{
			Scopes: make(map[string]ScopeStatus),
		}
		// First add a scope
		scope1 := ScopeStatus{
			Name:   "existing-scope",
			Status: "running",
			PID:    5555,
			Repos:  []string{"repo1", "repo2"},
		}
		state.UpdateScope(scope1)
		
		// Wait a bit to ensure different timestamp
		time.Sleep(100 * time.Millisecond)
		
		// Update the scope
		scope2 := ScopeStatus{
			Name:         "existing-scope",
			Status:       "stopped",
			PID:          0,
			Repos:        []string{"repo1", "repo2"},
			CurrentRepos: []string{"repo3"}, // Changed repos
			CurrentScope: "different-scope",  // Changed scope
		}
		state.UpdateScope(scope2)
		
		assert.Len(t, state.Scopes, 1)
		updated := state.Scopes["existing-scope"]
		assert.Equal(t, "stopped", updated.Status)
		assert.Equal(t, 0, updated.PID)
		assert.Equal(t, []string{"repo3"}, updated.CurrentRepos)
		assert.Equal(t, "different-scope", updated.CurrentScope)
		assert.NotEmpty(t, updated.LastActivity)
	})
}

func TestChangeScopeContext(t *testing.T) {
	state := &State{
		Scopes: make(map[string]ScopeStatus),
	}
	
	t.Run("ChangeScopeForExistingSession", func(t *testing.T) {
		// Add a scope first
		scope := ScopeStatus{
			Name:         "test-session",
			Status:       "running",
			PID:          1234,
			Repos:        []string{"repo1", "repo2"},
			CurrentRepos: []string{"repo1", "repo2"},
			CurrentScope: "test-session",
		}
		state.UpdateScope(scope)
		
		// Change scope context
		err := state.ChangeScopeContext("test-session", "new-scope", []string{"repo3", "repo4"})
		require.NoError(t, err)
		
		// Verify changes
		updated := state.Scopes["test-session"]
		assert.Equal(t, "new-scope", updated.CurrentScope)
		assert.Equal(t, []string{"repo3", "repo4"}, updated.CurrentRepos)
		assert.NotEmpty(t, updated.LastChange)
		assert.NotEmpty(t, updated.LastActivity)
	})
	
	t.Run("ChangeScopeForNonExistentSession", func(t *testing.T) {
		err := state.ChangeScopeContext("non-existent", "new-scope", []string{"repo1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scope session non-existent not found")
	})
}

func TestScopeStateRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "scope-roundtrip.json")
	
	// Create a complex state with scopes
	original := &State{
		Scopes: map[string]ScopeStatus{
			"scope1": {
				Name:         "scope1",
				Status:       "running",
				PID:          1111,
				Repos:        []string{"repo1", "repo2"},
				CurrentRepos: []string{"repo3"},
				CurrentScope: "different-scope",
				LastActivity: time.Now().Format(time.RFC3339),
				LastChange:   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			},
			"scope2": {
				Name:         "scope2",
				Status:       "stopped",
				Repos:        []string{"repo4", "repo5"},
				CurrentRepos: []string{"repo4", "repo5"},
				CurrentScope: "scope2",
				LastActivity: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			},
		},
		// Also test backward compatibility with agents
		Agents: map[string]AgentStatus{
			"legacy-agent": {
				Name:         "legacy-agent",
				Status:       "running",
				PID:          9999,
				Repository:   "legacy-repo",
				LastActivity: time.Now().Format(time.RFC3339),
			},
		},
	}
	
	// Save
	err := original.Save(statePath)
	require.NoError(t, err)
	
	// Load
	loaded, err := LoadState(statePath)
	require.NoError(t, err)
	
	// Compare all fields
	assert.Equal(t, original.Timestamp, loaded.Timestamp)
	assert.Len(t, loaded.Scopes, len(original.Scopes))
	assert.Len(t, loaded.Agents, len(original.Agents))
	
	// Check scopes
	for name, originalScope := range original.Scopes {
		loadedScope, exists := loaded.Scopes[name]
		assert.True(t, exists, "Scope %s should exist", name)
		assert.Equal(t, originalScope.Name, loadedScope.Name)
		assert.Equal(t, originalScope.Status, loadedScope.Status)
		assert.Equal(t, originalScope.PID, loadedScope.PID)
		assert.Equal(t, originalScope.Repos, loadedScope.Repos)
		assert.Equal(t, originalScope.CurrentRepos, loadedScope.CurrentRepos)
		assert.Equal(t, originalScope.CurrentScope, loadedScope.CurrentScope)
		assert.Equal(t, originalScope.LastActivity, loadedScope.LastActivity)
		assert.Equal(t, originalScope.LastChange, loadedScope.LastChange)
	}
	
	// Check legacy agents
	for name, originalAgent := range original.Agents {
		loadedAgent, exists := loaded.Agents[name]
		assert.True(t, exists, "Agent %s should exist", name)
		assert.Equal(t, originalAgent.Name, loadedAgent.Name)
		assert.Equal(t, originalAgent.Status, loadedAgent.Status)
		assert.Equal(t, originalAgent.PID, loadedAgent.PID)
		assert.Equal(t, originalAgent.Repository, loadedAgent.Repository)
		assert.Equal(t, originalAgent.LastActivity, loadedAgent.LastActivity)
	}
}

func TestStateRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "roundtrip.json")
	
	// Create a complex state
	original := &State{
		Agents: map[string]AgentStatus{
			"agent1": {
				Name:         "agent1",
				Status:       "running",
				PID:          1111,
				Repository:   "repo1",
				LastActivity: time.Now().Format(time.RFC3339),
			},
			"agent2": {
				Name:         "agent2",
				Status:       "stopped",
				Repository:   "repo2",
				LastActivity: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			},
			"agent3": {
				Name:         "agent3",
				Status:       "error",
				PID:          3333,
				Repository:   "repo3",
				LastActivity: time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			},
		},
	}
	
	// Save
	err := original.Save(statePath)
	require.NoError(t, err)
	
	// Load
	loaded, err := LoadState(statePath)
	require.NoError(t, err)
	
	// Compare all fields
	assert.Equal(t, original.Timestamp, loaded.Timestamp)
	assert.Len(t, loaded.Agents, len(original.Agents))
	
	for name, originalAgent := range original.Agents {
		loadedAgent, exists := loaded.Agents[name]
		assert.True(t, exists, "Agent %s should exist", name)
		assert.Equal(t, originalAgent.Name, loadedAgent.Name)
		assert.Equal(t, originalAgent.Status, loadedAgent.Status)
		assert.Equal(t, originalAgent.PID, loadedAgent.PID)
		assert.Equal(t, originalAgent.Repository, loadedAgent.Repository)
		assert.Equal(t, originalAgent.LastActivity, loadedAgent.LastActivity)
	}
}