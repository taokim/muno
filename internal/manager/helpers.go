package manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/taokim/muno/internal/adapters"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
	"github.com/taokim/muno/internal/tree"
)

// LoadFromCurrentDir loads a manager from the current directory
func LoadFromCurrentDir() (*Manager, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting current directory: %w", err)
	}
	
	// Find workspace root by looking for muno.yaml
	workspaceRoot := findWorkspaceRoot(cwd)
	if workspaceRoot == "" {
		return nil, fmt.Errorf("not in a MUNO workspace (no muno.yaml found)")
	}
	
	// Load the config
	configPath := filepath.Join(workspaceRoot, "muno.yaml")
	cfg, err := config.LoadTree(configPath)
	if err != nil {
		// Try alternative names
		for _, name := range []string{"muno.yml", ".muno.yaml", ".muno.yml"} {
			altPath := filepath.Join(workspaceRoot, name)
			if altCfg, altErr := config.LoadTree(altPath); altErr == nil {
				cfg = altCfg
				configPath = altPath
				break
			}
		}
		if cfg == nil {
			return nil, fmt.Errorf("loading config: %w", err)
		}
	}
	
	// Create providers using adapters
	configAdapter := adapters.NewConfigAdapter()
	gitProvider := adapters.NewGitProvider()
	fsAdapter := adapters.NewFileSystemAdapter()
	uiAdapter := adapters.NewUIAdapter()
	
	// Create git interface for tree manager
	gitCmd := git.New()
	
	// Create tree manager
	treeManager, err := tree.NewManager(workspaceRoot, gitCmd)
	if err != nil {
		return nil, fmt.Errorf("creating tree manager: %w", err)
	}
	
	// Create tree adapter stub to implement TreeProvider interface
	treeAdapter := adapters.NewTreeAdapter()
	
	// Create manager with all required providers
	mgr, err := NewManager(ManagerOptions{
		ConfigProvider:  configAdapter,
		GitProvider:     gitProvider,
		FSProvider:      fsAdapter,
		UIProvider:      uiAdapter,
		TreeProvider:    treeAdapter,
		AutoLoadConfig:  true,
	})
	if err != nil {
		return nil, err
	}
	
	// Set workspace and config
	mgr.workspace = workspaceRoot
	mgr.config = cfg
	mgr.initialized = true
	
	// Store the actual tree manager if needed (optional)
	_ = treeManager // treeManager is available if needed for tree operations
	
	return mgr, nil
}

// NewManagerForInit creates a manager for initialization
func NewManagerForInit(projectPath string) (*Manager, error) {
	// Resolve the absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}
	
	// Create a minimal config for initialization
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     filepath.Base(absPath),
			ReposDir: "repos",
		},
		Nodes: []config.NodeDefinition{},
	}
	
	// Create providers using adapters
	configAdapter := adapters.NewConfigAdapter()
	gitProvider := adapters.NewGitProvider()
	fsAdapter := adapters.NewFileSystemAdapter()
	uiAdapter := adapters.NewUIAdapter()
	
	// Create tree adapter stub to implement TreeProvider interface
	treeAdapter := adapters.NewTreeAdapter()
	
	// Create manager without loading config (since we're initializing)
	mgr, err := NewManager(ManagerOptions{
		ConfigProvider:  configAdapter,
		GitProvider:     gitProvider,
		FSProvider:      fsAdapter,
		UIProvider:      uiAdapter,
		TreeProvider:    treeAdapter,
		AutoLoadConfig:  false, // Don't auto-load since we're initializing
	})
	if err != nil {
		return nil, err
	}
	
	// Set workspace and config
	mgr.workspace = absPath
	mgr.config = cfg
	mgr.initialized = false // Will be set to true during SmartInitWorkspace
	
	return mgr, nil
}

// findWorkspaceRoot finds the workspace root by looking for muno.yaml
func findWorkspaceRoot(startPath string) string {
	current := startPath
	
	for {
		// Check for muno.yaml in current directory
		configPath := filepath.Join(current, "muno.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return current
		}
		
		// Also check for alternative names
		for _, name := range []string{"muno.yml", ".muno.yaml", ".muno.yml"} {
			configPath := filepath.Join(current, name)
			if _, err := os.Stat(configPath); err == nil {
				return current
			}
		}
		
		// Move up one directory
		parent := filepath.Dir(current)
		if parent == current {
			// Reached root directory
			break
		}
		current = parent
	}
	
	return ""
}