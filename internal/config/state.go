package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// State represents the runtime state of the system
type State struct {
	Timestamp string                 `json:"timestamp"`
	Scopes    map[string]ScopeStatus `json:"scopes"`
	
	// Deprecated: Agents field for backwards compatibility
	Agents    map[string]AgentStatus `json:"agents,omitempty"`
}

// ScopeStatus represents the status of a running scope
type ScopeStatus struct {
	Name         string   `json:"name"`
	Status       string   `json:"status"` // running, stopped, error
	PID          int      `json:"pid,omitempty"`
	Repos        []string `json:"repos"`
	LastActivity string   `json:"last_activity"`
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
	s.Scopes[status.Name] = status
}

// UpdateAgent updates or adds an agent status (deprecated - for backwards compatibility)
func (s *State) UpdateAgent(status AgentStatus) {
	status.LastActivity = time.Now().Format(time.RFC3339)
	s.Agents[status.Name] = status
}