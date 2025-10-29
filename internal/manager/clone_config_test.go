package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/mocks"
)

func TestResolveConfigPath(t *testing.T) {
	// Create a temporary workspace for testing
	tmpDir := t.TempDir()
	
	// Create test directory structure
	// workspace/
	//   ├── muno.yaml
	//   ├── configs/
	//   │   └── test.yaml
	//   └── .nodes/
	//       └── parent/
	//           ├── muno.yaml
	//           └── configs/
	//               └── child.yaml
	
	// Create directories
	os.MkdirAll(filepath.Join(tmpDir, "configs"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, ".nodes", "parent", "configs"), 0755)
	
	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "muno.yaml"), []byte("workspace:\n  name: test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "configs", "test.yaml"), []byte("test: config"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".nodes", "parent", "muno.yaml"), []byte("parent: config"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".nodes", "parent", "configs", "child.yaml"), []byte("child: config"), 0644)
	
	// Create manager with test workspace
	fsProvider := mocks.NewMockFileSystemProvider()
	
	// Setup the mock to check real files for this test
	fsProvider.SetExists(filepath.Join(tmpDir, ".nodes", "parent", "muno.yaml"), true)
	fsProvider.SetExists(filepath.Join(tmpDir, ".nodes", "parent", "configs", "child.yaml"), true)
	fsProvider.SetExists(filepath.Join(tmpDir, "configs", "test.yaml"), true)
	
	mgr := &Manager{
		workspace: tmpDir,
		fsProvider: fsProvider,
	}
	
	tests := []struct {
		name         string
		configFile   string
		nodePath     string
		setupFunc    func()
		expectedPath string
		description  string
	}{
		{
			name:         "absolute_path",
			configFile:   "/absolute/path/to/config.yaml",
			nodePath:     "/node",
			expectedPath: "/absolute/path/to/config.yaml",
			description:  "Absolute paths should be returned as-is",
		},
		{
			name:         "http_url",
			configFile:   "http://example.com/config.yaml",
			nodePath:     "/node",
			expectedPath: "http://example.com/config.yaml",
			description:  "HTTP URLs should be returned as-is",
		},
		{
			name:         "relative_to_parent_config",
			configFile:   "configs/child.yaml",
			nodePath:     "/parent/child",
			expectedPath: filepath.Join(tmpDir, ".nodes", "parent", "configs", "child.yaml"),
			description:  "Should resolve relative to parent config directory",
		},
		{
			name:         "relative_to_workspace",
			configFile:   "configs/test.yaml",
			nodePath:     "/test",
			expectedPath: filepath.Join(tmpDir, "configs", "test.yaml"),
			description:  "Should resolve relative to workspace when parent config doesn't exist",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}
			
			result := mgr.resolveConfigPath(tt.configFile, tt.nodePath)
			
			if result != tt.expectedPath {
				t.Errorf("%s: expected %s, got %s", tt.description, tt.expectedPath, result)
			}
		})
	}
}

func TestSymlinkOverwrite(t *testing.T) {
	// Test that existing wrong symlinks are properly overwritten
	tmpDir := t.TempDir()
	
	// Create test structure
	configDir := filepath.Join(tmpDir, "configs")
	nodeDir := filepath.Join(tmpDir, ".nodes", "test")
	os.MkdirAll(configDir, 0755)
	os.MkdirAll(nodeDir, 0755)
	
	// Create config file
	configPath := filepath.Join(configDir, "test.yaml")
	os.WriteFile(configPath, []byte("test: config"), 0644)
	
	// Create a wrong symlink (simulating committed symlink from git)
	wrongTarget := "/wrong/path/to/config.yaml"
	symlinkPath := filepath.Join(nodeDir, "muno.yaml")
	os.Symlink(wrongTarget, symlinkPath)
	
	// Verify wrong symlink exists
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read initial symlink: %v", err)
	}
	if target != wrongTarget {
		t.Fatalf("Initial symlink target mismatch: expected %s, got %s", wrongTarget, target)
	}
	
	// Create mock providers
	fsProvider := mocks.NewMockFileSystemProvider()
	logProvider := &DefaultLogProvider{debug: false}
	treeProvider := mocks.NewMockTreeProvider()
	
	// Create manager
	mgr := &Manager{
		workspace:    tmpDir,
		fsProvider:   fsProvider,
		logProvider:  logProvider,
		treeProvider: treeProvider,
	}
	
	// Create node info for config node
	node := interfaces.NodeInfo{
		Name:       "test",
		Path:       "/test",
		ConfigFile: configPath,
		IsConfig:   true,
	}
	
	// Process the node (should overwrite the symlink)
	err = mgr.visitNodeForClone(node, false, false)
	if err != nil {
		t.Fatalf("visitNodeForClone failed: %v", err)
	}
	
	// Verify symlink now points to correct location
	newTarget, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read updated symlink: %v", err)
	}
	if newTarget != configPath {
		t.Errorf("Symlink not updated correctly: expected %s, got %s", configPath, newTarget)
	}
}

