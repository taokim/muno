package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// State represents the runtime state of the system
type State struct {
	Timestamp   string                 `json:"timestamp"`
	ActiveScope string                 `json:"active_scope,omitempty"` // Currently active scope
	Scopes      map[string]ScopeStatus `json:"scopes"`
	
	// Deprecated: Agents field for backwards compatibility
	Agents    map[string]AgentStatus `json:"agents,omitempty"`
}

// ScopeStatus represents the status of a running scope
type ScopeStatus struct {
	Name         string   `json:"name"`
	Status       string   `json:"status"` // running, stopped, error
	PID          int      `json:"pid,omitempty"`
	Repos        []string `json:"repos"`         // Initial repositories when started
	CurrentRepos []string `json:"current_repos,omitempty"` // Current repositories (may change dynamically)
	CurrentScope string   `json:"current_scope,omitempty"` // Current scope name (may differ from Name)
	LastActivity string   `json:"last_activity"`
	LastChange   string   `json:"last_change,omitempty"`   // When scope was last changed
}

// AgentStatus represents the status of a running agent (deprecated)
type AgentStatus struct {
	Name         string `json:"name"`
	Status       string `json:"status"` // running, stopped, error
	PID          int    `json:"pid,omitempty"`
	Repository   string `json:"repository"`
	LastActivity string `json:"last_activity"`
}

// LoadState reads state from JSON file
func LoadState(path string) (*State, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &State{
			Scopes: make(map[string]ScopeStatus),
			Agents: make(map[string]AgentStatus),
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing state: %w", err)
	}

	if state.Scopes == nil {
		state.Scopes = make(map[string]ScopeStatus)
	}
	if state.Agents == nil {
		state.Agents = make(map[string]AgentStatus)
	}

	return &state, nil
}

// Save writes state to JSON file
func (s *State) Save(path string) error {
	s.Timestamp = time.Now().Format(time.RFC3339)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	return nil
}

// UpdateScope updates or adds a scope status
func (s *State) UpdateScope(status ScopeStatus) {
	status.LastActivity = time.Now().Format(time.RFC3339)
	// Initialize CurrentRepos and CurrentScope if not set
	if len(status.CurrentRepos) == 0 && len(status.Repos) > 0 {
		status.CurrentRepos = status.Repos
	}
	if status.CurrentScope == "" {
		status.CurrentScope = status.Name
	}
	s.Scopes[status.Name] = status
}

// ChangeScopeContext updates the current scope context for a running session
func (s *State) ChangeScopeContext(sessionName, newScopeName string, newRepos []string) error {
	scope, exists := s.Scopes[sessionName]
	if !exists {
		return fmt.Errorf("scope session %s not found", sessionName)
	}
	
	scope.CurrentScope = newScopeName
	scope.CurrentRepos = newRepos
	scope.LastChange = time.Now().Format(time.RFC3339)
	scope.LastActivity = time.Now().Format(time.RFC3339)
	s.Scopes[sessionName] = scope
	
	return nil
}

// UpdateAgent updates or adds an agent status (deprecated - for backwards compatibility)
func (s *State) UpdateAgent(status AgentStatus) {
	status.LastActivity = time.Now().Format(time.RFC3339)
	s.Agents[status.Name] = status
}

// SetActiveScope sets the currently active scope
func (s *State) SetActiveScope(scopeName string) {
	s.ActiveScope = scopeName
	s.Timestamp = time.Now().Format(time.RFC3339)
}

// GetActiveScope returns the currently active scope
func (s *State) GetActiveScope() string {
	return s.ActiveScope
}