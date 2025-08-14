package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/repo-claude/internal/config"
)

func TestStartAgent(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := config.DefaultConfig("test")
	state := &config.State{
		Agents: make(map[string]config.AgentStatus),
	}
	
	mgr := &Manager{
		WorkspacePath: tmpDir,
		Config:        cfg,
		State:         state,
		agents:        make(map[string]*Agent),
	}
	
	t.Run("AgentNotFound", func(t *testing.T) {
		err := mgr.StartAgent("non-existent-agent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent non-existent-agent not found")
	})
	
	t.Run("RepositoryNotFound", func(t *testing.T) {
		err := mgr.StartAgent("backend-agent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository backend not found")
	})
	
	t.Run("ValidAgentButNoClaudeCLI", func(t *testing.T) {
		// Create repository directory
		backendDir := filepath.Join(tmpDir, "backend")
		err := os.MkdirAll(backendDir, 0755)
		require.NoError(t, err)
		
		// This will fail because claude CLI is not installed
		err = mgr.StartAgent("backend-agent")
		assert.Error(t, err)
		// The error will be about claude command not found
	})
}

func TestStopAgent(t *testing.T) {
	mgr := &Manager{
		agents: make(map[string]*Agent),
		State: &config.State{
			Agents: make(map[string]config.AgentStatus),
		},
		WorkspacePath: t.TempDir(),
	}
	
	t.Run("AgentNotRunning", func(t *testing.T) {
		err := mgr.StopAgent("not-running")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not-running is not running")
	})
	
	// We can't easily test stopping a real process without actually starting one
}

func TestStartAllAgents(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a config with dependencies
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Name: "test",
			Manifest: config.Manifest{
				Projects: []config.Project{
					{Name: "backend", Agent: "backend-agent"},
					{Name: "frontend", Agent: "frontend-agent"},
				},
			},
		},
		Agents: map[string]config.Agent{
			"backend-agent": {
				Model:        "claude-sonnet-4",
				AutoStart:    true,
				Dependencies: []string{},
			},
			"frontend-agent": {
				Model:        "claude-sonnet-4",
				AutoStart:    true,
				Dependencies: []string{"backend-agent"},
			},
			"manual-agent": {
				Model:     "claude-sonnet-4",
				AutoStart: false, // Should not start
			},
		},
	}
	
	mgr := &Manager{
		WorkspacePath: tmpDir,
		Config:        cfg,
		State: &config.State{
			Agents: make(map[string]config.AgentStatus),
		},
		agents: make(map[string]*Agent),
	}
	
	// This will fail because repos don't exist, but we can check the logic
	err := mgr.StartAllAgents()
	// Will have errors but shouldn't panic
	assert.NotNil(t, err) // Sync might fail
}

func TestCircularDependencyDetection(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]config.Agent{
			"agent-a": {
				AutoStart:    true,
				Dependencies: []string{"agent-b"},
			},
			"agent-b": {
				AutoStart:    true,
				Dependencies: []string{"agent-a"}, // Circular!
			},
		},
	}
	
	mgr := &Manager{
		Config: cfg,
		agents: make(map[string]*Agent),
	}
	
	err := mgr.StartAllAgents()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency detected")
}

func TestStopAllAgents(t *testing.T) {
	mgr := &Manager{
		agents: map[string]*Agent{
			// Empty map, so nothing to stop
		},
		State: &config.State{
			Agents: make(map[string]config.AgentStatus),
		},
		WorkspacePath: t.TempDir(),
	}
	
	// Should complete without error even with no agents
	err := mgr.StopAllAgents()
	assert.NoError(t, err)
}

func TestShowStatus(t *testing.T) {
	tmpDir := t.TempDir()
	
	cfg := config.DefaultConfig("test")
	state := &config.State{
		Agents: map[string]config.AgentStatus{
			"backend-agent": {
				Name:       "backend-agent",
				Status:     "running",
				Repository: "backend",
			},
			"frontend-agent": {
				Name:       "frontend-agent",
				Status:     "stopped",
				Repository: "frontend",
			},
		},
	}
	
	mgr := &Manager{
		WorkspacePath: tmpDir,
		Config:        cfg,
		State:         state,
		agents:        make(map[string]*Agent),
	}
	
	// This should not panic and complete successfully
	err := mgr.ShowStatus()
	assert.NoError(t, err)
}