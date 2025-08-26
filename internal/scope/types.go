package scope

import (
	"time"
)

// Type represents the type of scope
type Type string

const (
	// TypePersistent represents a long-lived scope
	TypePersistent Type = "persistent"
	// TypeEphemeral represents a short-lived scope
	TypeEphemeral Type = "ephemeral"
)

// State represents the current state of a scope
type State string

const (
	// StateActive indicates the scope is currently in use
	StateActive State = "active"
	// StateInactive indicates the scope exists but is not in use
	StateInactive State = "inactive"
	// StateArchived indicates the scope has been archived
	StateArchived State = "archived"
)

// Meta contains metadata about a scope
type Meta struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Type         Type          `json:"type"`
	CreatedAt    time.Time     `json:"created_at"`
	LastAccessed time.Time     `json:"last_accessed"`
	Repos        []RepoState   `json:"repos"`
	State        State         `json:"state"`
	SessionPID   *int          `json:"session_pid,omitempty"`
	Description  string        `json:"description,omitempty"`
}

// RepoState tracks the state of a repository within a scope
type RepoState struct {
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Branch     string    `json:"branch"`
	Commit     string    `json:"commit"`
	ClonedAt   time.Time `json:"cloned_at"`
	LastPulled time.Time `json:"last_pulled"`
	IsDirty    bool      `json:"is_dirty,omitempty"`
}

// Info provides summary information about a scope
type Info struct {
	Name         string    `json:"name"`
	Type         Type      `json:"type"`
	State        State     `json:"state"`
	RepoCount    int       `json:"repo_count"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	Path         string    `json:"path"`
	Description  string    `json:"description,omitempty"`
}

// CreateOptions contains options for creating a new scope
type CreateOptions struct {
	Type        Type          `json:"type"`
	Repos       []string      `json:"repos"`
	Description string        `json:"description"`
	Branch      string        `json:"branch,omitempty"`
	CloneRepos  bool          `json:"clone_repos"`
}

// StartOptions contains options for starting a scope
type StartOptions struct {
	NewWindow   bool   `json:"new_window"`
	Pull        bool   `json:"pull"`
	Model       string `json:"model,omitempty"`
	WorkDir     string `json:"work_dir,omitempty"`
}

// PullOptions contains options for pulling repositories
type PullOptions struct {
	CloneMissing bool     `json:"clone_missing"`
	Repos        []string `json:"repos,omitempty"`
	Parallel     bool     `json:"parallel"`
	Force        bool     `json:"force"`
}

// StatusReport provides detailed status of a scope
type StatusReport struct {
	Name         string           `json:"name"`
	Type         Type             `json:"type"`
	State        State            `json:"state"`
	Path         string           `json:"path"`
	CreatedAt    time.Time        `json:"created_at"`
	LastAccessed time.Time        `json:"last_accessed"`
	Repos        []RepoStatus     `json:"repos"`
	DiskUsage    int64            `json:"disk_usage_bytes"`
	IsRunning    bool             `json:"is_running"`
	SessionPID   *int             `json:"session_pid,omitempty"`
}

// RepoStatus provides detailed status of a repository
type RepoStatus struct {
	Name           string   `json:"name"`
	Branch         string   `json:"branch"`
	Commit         string   `json:"commit"`
	IsDirty        bool     `json:"is_dirty"`
	UntrackedFiles int      `json:"untracked_files"`
	ModifiedFiles  int      `json:"modified_files"`
	AheadBy        int      `json:"ahead_by"`
	BehindBy       int      `json:"behind_by"`
	LastPulled     time.Time `json:"last_pulled"`
}

// ScopeDetail provides detailed information about a scope including its number
type ScopeDetail struct {
	Number       int       `json:"number"`
	Name         string    `json:"name"`
	Type         Type      `json:"type"`
	State        State     `json:"state"`
	Repos        []string  `json:"repos"`
	RepoCount    int       `json:"repo_count"`
	Description  string    `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	Path         string    `json:"path"`
}