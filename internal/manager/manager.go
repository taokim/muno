package manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/yourusername/repo-claude/internal/config"
	"github.com/yourusername/repo-claude/internal/manifest"
)

// Manager handles the repo-claude workspace
type Manager struct {
	WorkspacePath string
	Config        *config.Config
	State         *config.State
	agents        map[string]*Agent
	mu            sync.Mutex
}

// Agent represents a running Claude Code instance
type Agent struct {
	Name    string
	Process *os.Process
	Status  string
}

// New creates a new manager for initialization
func New(workspacePath string) *Manager {
	absPath, _ := filepath.Abs(workspacePath)
	return &Manager{
		WorkspacePath: absPath,
		agents:        make(map[string]*Agent),
	}
}

// LoadFromCurrentDir loads an existing workspace from current directory
func LoadFromCurrentDir() (*Manager, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(cwd, "repo-claude.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no repo-claude.yaml found in current directory")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	statePath := filepath.Join(cwd, ".repo-claude-state.json")
	state, err := config.LoadState(statePath)
	if err != nil {
		return nil, fmt.Errorf("loading state: %w", err)
	}

	return &Manager{
		WorkspacePath: cwd,
		Config:        cfg,
		State:         state,
		agents:        make(map[string]*Agent),
	}, nil
}

// InitWorkspace initializes a new workspace
func (m *Manager) InitWorkspace(projectName string, interactive bool) error {
	fmt.Printf("🚀 Initializing Repo-Claude workspace: %s\n", projectName)

	// Create workspace directory
	if err := os.MkdirAll(m.WorkspacePath, 0755); err != nil {
		return fmt.Errorf("creating workspace: %w", err)
	}

	// Change to workspace directory
	if err := os.Chdir(m.WorkspacePath); err != nil {
		return fmt.Errorf("changing to workspace: %w", err)
	}

	// Create configuration
	m.Config = config.DefaultConfig(projectName)
	
	if interactive {
		if err := m.interactiveConfig(); err != nil {
			return fmt.Errorf("interactive configuration: %w", err)
		}
	}

	// Save configuration
	configPath := filepath.Join(m.WorkspacePath, "repo-claude.yaml")
	if err := m.Config.Save(configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	// Generate manifest XML
	manifestXML, err := manifest.Generate(m.Config)
	if err != nil {
		return fmt.Errorf("generating manifest: %w", err)
	}

	// Create manifest repository
	if err := manifest.CreateManifestRepo(m.WorkspacePath, manifestXML); err != nil {
		return fmt.Errorf("creating manifest repo: %w", err)
	}

	// Initialize repo workspace
	if err := m.initRepoWorkspace(); err != nil {
		return fmt.Errorf("initializing repo: %w", err)
	}

	// Sync repositories (allow failure for initial setup)
	if err := m.repoSync(); err != nil {
		fmt.Printf("⚠️  Initial sync skipped: %v\n", err)
		fmt.Println("You can run './repo-claude sync' after updating the configuration")
	}

	// Setup coordination files
	if err := m.setupCoordination(); err != nil {
		return fmt.Errorf("setting up coordination: %w", err)
	}

	// Copy executable to workspace
	if err := m.copyExecutable(); err != nil {
		return fmt.Errorf("copying executable: %w", err)
	}

	fmt.Println("✅ Workspace initialized using Repo tool!")
	fmt.Printf("📍 Location: %s\n", m.WorkspacePath)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  ./repo-claude start     # Start all agents")
	fmt.Println("  ./repo-claude status    # Check status")
	fmt.Println("  repo status             # Use repo tool directly")

	return nil
}

// initRepoWorkspace initializes the repo workspace
func (m *Manager) initRepoWorkspace() error {
	fmt.Println("📦 Initializing Repo workspace...")
	
	manifestPath := filepath.Join(m.WorkspacePath, ".manifest-repo")
	// Use file:// URL for local manifest repository
	manifestURL := "file://" + manifestPath
	cmd := exec.Command("repo", "init", "-u", manifestURL, "-b", "main")
	cmd.Dir = m.WorkspacePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("repo init failed: %w", err)
	}
	
	fmt.Println("✅ Repo workspace initialized")
	return nil
}

// repoSync syncs all repositories
func (m *Manager) repoSync() error {
	fmt.Println("🔄 Syncing repositories with Repo...")
	
	cmd := exec.Command("repo", "sync", "-j4")
	cmd.Dir = m.WorkspacePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		// For initial sync, repos might not exist yet
		fmt.Println("⚠️  Initial sync skipped (repositories may not exist yet)")
		return nil
	}
	
	fmt.Println("✅ Repo sync completed")
	return nil
}

// copyExecutable copies the current executable to the workspace
func (m *Manager) copyExecutable() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}

	destPath := filepath.Join(m.WorkspacePath, "repo-claude")
	
	// Read source file
	data, err := os.ReadFile(executable)
	if err != nil {
		return fmt.Errorf("reading executable: %w", err)
	}

	// Write to destination
	if err := os.WriteFile(destPath, data, 0755); err != nil {
		return fmt.Errorf("writing executable: %w", err)
	}

	return nil
}