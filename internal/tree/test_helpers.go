package tree

import (
	"path/filepath"
	"testing"
	
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
)

// CreateTestConfig creates a test configuration file in the given directory
func CreateTestConfig(t *testing.T, dir string, reposDir string) {
	t.Helper()
	
	if reposDir == "" {
		reposDir = config.GetDefaultReposDir() // use config default
	}
	
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: reposDir,
		},
		Nodes: []config.NodeDefinition{},
	}
	
	configPath := filepath.Join(dir, "muno.yaml")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}
}

// CreateTestManagerWithConfig creates a Manager with a test config
func CreateTestManagerWithConfig(t *testing.T, dir string, gitCmd git.Interface) (*Manager, error) {
	t.Helper()
	CreateTestConfig(t, dir, "")
	return NewManager(dir, gitCmd)
}

// CreateTestStatelessManagerWithConfig creates a StatelessManager with a test config
func CreateTestStatelessManagerWithConfig(t *testing.T, dir string, gitCmd git.Interface) (*StatelessManager, error) {
	t.Helper()
	CreateTestConfig(t, dir, "")
	return NewStatelessManager(dir, gitCmd)
}