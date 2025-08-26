package scope

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/taokim/repo-claude/internal/config"
)

const (
	// MetaFileName is the name of the scope metadata file
	MetaFileName = ".scope-meta.json"
	// SharedMemoryFileName is the name of the shared memory file
	SharedMemoryFileName = "shared-memory.md"
)

// Manager manages scope lifecycles
type Manager struct {
	config        *config.Config
	projectPath   string
	workspacePath string
}

// NewManager creates a new scope manager
func NewManager(cfg *config.Config, projectPath string) (*Manager, error) {
	workspacePath := filepath.Join(projectPath, cfg.Workspace.BasePath)
	if filepath.IsAbs(cfg.Workspace.BasePath) {
		workspacePath = cfg.Workspace.BasePath
	}

	// Create workspaces directory if it doesn't exist
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspaces directory: %w", err)
	}

	return &Manager{
		config:        cfg,
		projectPath:   projectPath,
		workspacePath: workspacePath,
	}, nil
}

// Create creates a new scope
func (m *Manager) Create(name string, opts CreateOptions) error {
	// Check if scope already exists
	scopePath := filepath.Join(m.workspacePath, name)
	if _, err := os.Stat(scopePath); err == nil {
		return fmt.Errorf("scope %s already exists", name)
	}

	// Create scope directory
	if err := os.MkdirAll(scopePath, 0755); err != nil {
		return fmt.Errorf("failed to create scope directory: %w", err)
	}

	// Generate scope ID
	scopeID := fmt.Sprintf("%s-%s-%d", name, time.Now().Format("20060102"), time.Now().Unix())

	// Create scope metadata
	meta := &Meta{
		ID:           scopeID,
		Name:         name,
		Type:         opts.Type,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		Repos:        []RepoState{},
		State:        StateInactive,
		Description:  opts.Description,
	}

	// Save metadata
	if err := m.saveMeta(scopePath, meta); err != nil {
		os.RemoveAll(scopePath) // Cleanup on failure
		return fmt.Errorf("failed to save scope metadata: %w", err)
	}

	// Initialize shared memory
	sharedMemPath := filepath.Join(scopePath, SharedMemoryFileName)
	initialContent := fmt.Sprintf("# Shared Memory for Scope: %s\n\nCreated: %s\n\n## Notes\n\n",
		name, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(sharedMemPath, []byte(initialContent), 0644); err != nil {
		return fmt.Errorf("failed to create shared memory: %w", err)
	}

	// Clone repositories if requested
	if opts.CloneRepos && len(opts.Repos) > 0 {
		scope := &Scope{
			meta:    meta,
			path:    scopePath,
			manager: m,
		}
		for _, repoName := range opts.Repos {
			if err := scope.cloneRepo(repoName, opts.Branch); err != nil {
				fmt.Printf("Warning: failed to clone %s: %v\n", repoName, err)
			}
		}
		// Update metadata with cloned repos
		if err := m.saveMeta(scopePath, meta); err != nil {
			return fmt.Errorf("failed to update metadata: %w", err)
		}
	}

	fmt.Printf("✅ Created scope '%s' at %s\n", name, scopePath)
	return nil
}

// Delete removes a scope and its workspace
func (m *Manager) Delete(name string) error {
	scopePath := filepath.Join(m.workspacePath, name)
	
	// Check if scope exists
	if _, err := os.Stat(scopePath); os.IsNotExist(err) {
		return fmt.Errorf("scope %s does not exist", name)
	}

	// Load metadata to check if scope is active
	meta, err := m.loadMeta(scopePath)
	if err == nil && meta.State == StateActive {
		return fmt.Errorf("cannot delete active scope %s", name)
	}

	// Remove the scope directory
	if err := os.RemoveAll(scopePath); err != nil {
		return fmt.Errorf("failed to delete scope: %w", err)
	}

	fmt.Printf("✅ Deleted scope '%s'\n", name)
	return nil
}

// List returns information about all scopes
func (m *Manager) List() ([]Info, error) {
	// Ensure workspace directory exists
	if err := os.MkdirAll(m.workspacePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	entries, err := os.ReadDir(m.workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace directory: %w", err)
	}

	var scopes []Info
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		scopePath := filepath.Join(m.workspacePath, entry.Name())
		meta, err := m.loadMeta(scopePath)
		if err != nil {
			// Not a valid scope directory
			continue
		}

		info := Info{
			Name:         meta.Name,
			Type:         meta.Type,
			State:        meta.State,
			RepoCount:    len(meta.Repos),
			CreatedAt:    meta.CreatedAt,
			LastAccessed: meta.LastAccessed,
			Path:         scopePath,
			Description:  meta.Description,
		}
		scopes = append(scopes, info)
	}

	return scopes, nil
}

// Get retrieves a specific scope
func (m *Manager) Get(name string) (*Scope, error) {
	scopePath := filepath.Join(m.workspacePath, name)
	
	// Check if scope exists
	if _, err := os.Stat(scopePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("scope %s does not exist", name)
	}

	// Load metadata
	meta, err := m.loadMeta(scopePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load scope metadata: %w", err)
	}

	// Update last accessed time
	meta.LastAccessed = time.Now()
	if err := m.saveMeta(scopePath, meta); err != nil {
		// Non-fatal error
		fmt.Printf("Warning: failed to update last accessed time: %v\n", err)
	}

	return &Scope{
		meta:    meta,
		path:    scopePath,
		manager: m,
	}, nil
}

// Archive archives a scope by compressing it
func (m *Manager) Archive(name string) error {
	scopePath := filepath.Join(m.workspacePath, name)
	
	// Check if scope exists
	if _, err := os.Stat(scopePath); os.IsNotExist(err) {
		return fmt.Errorf("scope %s does not exist", name)
	}

	// Load metadata
	meta, err := m.loadMeta(scopePath)
	if err != nil {
		return fmt.Errorf("failed to load scope metadata: %w", err)
	}

	// Check if scope is active
	if meta.State == StateActive {
		return fmt.Errorf("cannot archive active scope %s", name)
	}

	// Update state to archived
	meta.State = StateArchived
	if err := m.saveMeta(scopePath, meta); err != nil {
		return fmt.Errorf("failed to update scope state: %w", err)
	}

	// Create archive (tar.gz)
	archivePath := filepath.Join(m.workspacePath, fmt.Sprintf("%s.archive.tar.gz", name))
	// TODO: Implement actual archiving
	fmt.Printf("✅ Archived scope '%s' to %s\n", name, archivePath)
	
	return nil
}

// Cleanup removes expired ephemeral scopes
func (m *Manager) Cleanup() error {
	// Currently, ephemeral scopes are not automatically cleaned up
	// This method is kept for potential future use or manual cleanup logic
	return nil
}
// saveMeta saves scope metadata to disk
func (m *Manager) saveMeta(scopePath string, meta *Meta) error {
	metaPath := filepath.Join(scopePath, MetaFileName)
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, data, 0644)
}

// loadMeta loads scope metadata from disk
func (m *Manager) loadMeta(scopePath string) (*Meta, error) {
	metaPath := filepath.Join(scopePath, MetaFileName)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// GetWorkspacePath returns the workspace path
func (m *Manager) GetWorkspacePath() string {
	return m.workspacePath
}

// GetConfig returns the configuration
func (m *Manager) GetConfig() *config.Config {
	return m.config
}

// GetProjectPath returns the project root path
func (m *Manager) GetProjectPath() string {
	return m.projectPath
}

// ListWithRepos returns detailed information about all scopes including repositories
func (m *Manager) ListWithRepos() ([]ScopeDetail, error) {
	scopes, err := m.List()
	if err != nil {
		return nil, err
	}

	var details []ScopeDetail
	for i, info := range scopes {
		// Get repositories from config
		var repos []string
		if scopeConfig, exists := m.config.Scopes[info.Name]; exists {
			repos = scopeConfig.Repos
		}

		detail := ScopeDetail{
			Number:       i + 1, // 1-based numbering
			Name:         info.Name,
			Type:         info.Type,
			State:        info.State,
			Repos:        repos,
			RepoCount:    len(repos),
			Description:  info.Description,
			CreatedAt:    info.CreatedAt,
			LastAccessed: info.LastAccessed,
			Path:         info.Path,
		}
		details = append(details, detail)
	}

	return details, nil
}

// GetByNumberOrName retrieves a scope by its number (1-based) or name
func (m *Manager) GetByNumberOrName(identifier string) (*Scope, error) {
	// Try to parse as number first
	var scopeName string
	var num int
	if n, err := fmt.Sscanf(identifier, "%d", &num); err == nil && n == 1 && num > 0 {
		// Get list of scopes
		scopes, err := m.List()
		if err != nil {
			return nil, err
		}
		if num > len(scopes) {
			return nil, fmt.Errorf("scope number %d out of range (have %d scopes)", num, len(scopes))
		}
		scopeName = scopes[num-1].Name
	} else {
		// Use as name
		scopeName = identifier
	}

	return m.Get(scopeName)
}

// CreateFromConfig creates a scope from configuration definition
func (m *Manager) CreateFromConfig(name string) error {
	scopeConfig, exists := m.config.Scopes[name]
	if !exists {
		return fmt.Errorf("scope %s not defined in configuration", name)
	}


	opts := CreateOptions{
		Type:        Type(scopeConfig.Type),
		Repos:       scopeConfig.Repos,
		Description: scopeConfig.Description,
		CloneRepos:  true, // Always clone on creation
	}

	return m.Create(name, opts)
}