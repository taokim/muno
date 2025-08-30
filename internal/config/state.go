package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// State represents the runtime state of the tree-based system
type State struct {
	Timestamp       string            `json:"timestamp"`
	CurrentNodePath string            `json:"current_node_path,omitempty"` // Current node in tree
	Sessions        map[string]Session `json:"sessions"`                    // Active Claude sessions
}

// Session represents an active Claude Code session
type Session struct {
	NodePath     string `json:"node_path"`
	Status       string `json:"status"` // running, stopped, error
	PID          int    `json:"pid,omitempty"`
	StartTime    string `json:"start_time"`
	LastActivity string `json:"last_activity"`
}

// LoadState reads state from JSON file
func LoadState(path string) (*State, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &State{
			Sessions: make(map[string]Session),
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

	return &state, nil
}

// SaveState writes state to JSON file
func (s *State) SaveState(path string) error {
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

// SetCurrentNode updates the current node path
func (s *State) SetCurrentNode(path string) {
	s.CurrentNodePath = path
}

// GetCurrentNode returns the current node path
func (s *State) GetCurrentNode() string {
	return s.CurrentNodePath
}

// AddSession adds a new session
func (s *State) AddSession(nodePath string, pid int) {
	s.Sessions[nodePath] = Session{
		NodePath:     nodePath,
		Status:       "running",
		PID:          pid,
		StartTime:    time.Now().Format(time.RFC3339),
		LastActivity: time.Now().Format(time.RFC3339),
	}
}

// RemoveSession removes a session
func (s *State) RemoveSession(nodePath string) {
	delete(s.Sessions, nodePath)
}

// UpdateSessionActivity updates the last activity time for a session
func (s *State) UpdateSessionActivity(nodePath string) {
	if session, exists := s.Sessions[nodePath]; exists {
		session.LastActivity = time.Now().Format(time.RFC3339)
		s.Sessions[nodePath] = session
	}
}