package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/taokim/muno/internal/adapters"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
	"github.com/taokim/muno/internal/interfaces"
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
	
	// Create tree adapter with the actual tree manager
	treeAdapter := NewTreeAdapter(treeManager)
	
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
			ReposDir: config.GetDefaultNodesDir(),
		},
		Nodes: []config.NodeDefinition{},
	}
	
	// Create providers using adapters
	configAdapter := adapters.NewConfigAdapter()
	gitProvider := adapters.NewGitProvider()
	fsAdapter := adapters.NewFileSystemAdapter()
	uiAdapter := adapters.NewUIAdapter()
	
	// Create git interface for tree manager
	gitCmd := git.New()
	
	// Create tree manager for initialization
	treeManager, err := tree.NewManager(absPath, gitCmd)
	if err != nil {
		// For init, create with default tree state
		treeManager = nil // We'll handle this during initialization
	}
	
	// Create tree adapter - use stub for init since tree might not exist yet
	var treeProvider interfaces.TreeProvider
	if treeManager != nil {
		treeProvider = NewTreeAdapter(treeManager)
	} else {
		treeProvider = adapters.NewTreeAdapter()
	}
	
	// Create manager without loading config (since we're initializing)
	mgr, err := NewManager(ManagerOptions{
		ConfigProvider:  configAdapter,
		GitProvider:     gitProvider,
		FSProvider:      fsAdapter,
		UIProvider:      uiAdapter,
		TreeProvider:    treeProvider,
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
	var candidates []string
	
	for {
		// Check for muno.yaml in current directory (must be a regular file, not a symlink)
		configPath := filepath.Join(current, "muno.yaml")
		if info, err := os.Lstat(configPath); err == nil && info.Mode().IsRegular() {
			// Found a muno.yaml, but we need to check if it's the TRUE workspace root
			// A true workspace root is one that's not inside any .nodes or repos directory
			candidates = append(candidates, current)
		}
		
		// Also check for alternative names
		for _, name := range []string{"muno.yml", ".muno.yaml", ".muno.yml"} {
			configPath := filepath.Join(current, name)
			if info, err := os.Lstat(configPath); err == nil && info.Mode().IsRegular() {
				// Don't add duplicate if we already found muno.yaml
				if len(candidates) == 0 || candidates[len(candidates)-1] != current {
					candidates = append(candidates, current)
				}
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
	
	// Now find the TRUE workspace root from candidates
	// The true workspace root is the highest level muno.yaml that's not inside a repos directory
	for i := len(candidates) - 1; i >= 0; i-- {
		candidate := candidates[i]
		
		// Check if this candidate is inside a .nodes or other repos directory
		// by checking if any parent directory up to the next candidate contains .nodes
		isNested := false
		if i < len(candidates) - 1 {
			// There's a parent candidate, check if we're under its repos directory
			parentCandidate := candidates[i+1]
			relPath, _ := filepath.Rel(parentCandidate, candidate)
			// Check if the path contains .nodes or starts with .nodes
			if strings.Contains(relPath, ".nodes") || strings.HasPrefix(relPath, ".nodes") {
				isNested = true
			}
			// Also check for other common repos directory names
			if strings.Contains(relPath, "nodes") || strings.Contains(relPath, "repos") || 
			   strings.Contains(relPath, "subrepos") || strings.Contains(relPath, "custom-repos") {
				isNested = true
			}
		}
		
		if !isNested {
			return candidate
		}
	}
	
	// If all candidates are nested (shouldn't happen), return the topmost one
	if len(candidates) > 0 {
		return candidates[len(candidates)-1]
	}
	
	return ""
}

// getCurrentTreePath converts current working directory to tree path
// Returns "/" if outside workspace or at workspace root
func (m *Manager) getCurrentTreePath() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %w", err)
	}
	
	workspaceRoot := m.workspace
	reposDir := filepath.Join(workspaceRoot, m.getReposDir())
	
	// Convert pwd to tree path
	if strings.HasPrefix(pwd, reposDir) {
		// We're inside the repos directory
		relPath, err := filepath.Rel(reposDir, pwd)
		if err != nil {
			return "", fmt.Errorf("getting relative path: %w", err)
		}
		
		// Convert to tree path
		if relPath == "." {
			return "/", nil
		}
		// Clean and convert to forward slashes
		return "/" + strings.ReplaceAll(filepath.ToSlash(relPath), "\\", "/"), nil
	}
	
	if pwd == workspaceRoot {
		// We're at workspace root
		return "/", nil
	}
	
	// Outside workspace - use root
	return "/", nil
}