package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// StateV3 represents the runtime state of the v3 tree-based system
type StateV3 struct {
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

// LoadStateV3 reads v3 state from JSON file
func LoadStateV3(path string) (*StateV3, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &StateV3{
			Sessions: make(map[string]Session),
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var state StateV3
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing state: %w", err)
	}

	return &state, nil
}

// SaveStateV3 writes v3 state to JSON file
func (s *StateV3) SaveStateV3(path string) error {
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
func (s *StateV3) SetCurrentNode(path string) {
	s.CurrentNodePath = path
}

// GetCurrentNode returns the current node path
func (s *StateV3) GetCurrentNode() string {
	return s.CurrentNodePath
}

// AddSession adds a new session
func (s *StateV3) AddSession(nodePath string, pid int) {
	s.Sessions[nodePath] = Session{
		NodePath:     nodePath,
		Status:       "running",
		PID:          pid,
		StartTime:    time.Now().Format(time.RFC3339),
		LastActivity: time.Now().Format(time.RFC3339),
	}
}

// RemoveSession removes a session
func (s *StateV3) RemoveSession(nodePath string) {
	delete(s.Sessions, nodePath)
}

// UpdateSessionActivity updates the last activity time for a session
func (s *StateV3) UpdateSessionActivity(nodePath string) {
	if session, exists := s.Sessions[nodePath]; exists {
		session.LastActivity = time.Now().Format(time.RFC3339)
		s.Sessions[nodePath] = session
	}
}