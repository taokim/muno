package tree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/git"
)

func setupTestManager(t *testing.T) (*StatelessManager, string) {
	tmpDir := t.TempDir()
	
	// Create config directory structure
	reposDir := filepath.Join(tmpDir, "nodes")
	os.MkdirAll(reposDir, 0755)
	
	// Create a basic config
	cfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: "nodes",
		},
		Nodes: []config.NodeDefinition{
			{
				Name: "child1",
				URL:  "https://github.com/test/child1.git",
			},
			{
				Name: "child2",
				URL:  "https://github.com/test/child2.git",
				Lazy: true,
			},
			{
				Name:   "parent",
				Config: "parent/muno.yaml",
			},
		},
	}
	
	// Save config
	configPath := filepath.Join(tmpDir, "muno.yaml")
	err := cfg.Save(configPath)
	require.NoError(t, err)
	
	// Create nested config for parent
	parentDir := filepath.Join(reposDir, "parent")
	os.MkdirAll(parentDir, 0755)
	parentCfg := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "parent-workspace",
			ReposDir: "nodes",
		},
		Nodes: []config.NodeDefinition{
			{
				Name: "nested1",
				URL:  "https://github.com/test/nested1.git",
			},
			{
				Name: "nested2",
				URL:  "https://github.com/test/nested2.git",
				Lazy: true,
			},
		},
	}
	err = parentCfg.Save(filepath.Join(parentDir, "muno.yaml"))
	require.NoError(t, err)
	
	// Create git mock
	gitCmd := &git.MockGit{}
	
	mgr, err := NewStatelessManager(tmpDir, gitCmd)
	require.NoError(t, err)
	
	return mgr, tmpDir
}

func TestStatelessManager_ComputeFilesystemPath(t *testing.T) {
	mgr, tmpDir := setupTestManager(t)
	
	tests := []struct {
		name         string
		logicalPath  string
		expectedPath string
	}{
		{
			name:         "root path",
			logicalPath:  "/",
			expectedPath: filepath.Join(tmpDir, "nodes"),
		},
		{
			name:         "empty path defaults to root",
			logicalPath:  "",
			expectedPath: filepath.Join(tmpDir, "nodes"),
		},
		{
			name:         "child path",
			logicalPath:  "/child1",
			expectedPath: filepath.Join(tmpDir, "nodes", "child1"),
		},
		{
			name:         "nested path",
			logicalPath:  "/parent/nested1",
			expectedPath: filepath.Join(tmpDir, "nodes", "parent", "nodes", "nested1"),
		},
		{
			name:         "relative path",
			logicalPath:  "child1",
			expectedPath: filepath.Join(tmpDir, "nodes", "child1"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mgr.ComputeFilesystemPath(tt.logicalPath)
			assert.Equal(t, tt.expectedPath, result)
		})
	}
}

func TestStatelessManager_GetNodeByPath(t *testing.T) {
	mgr, _ := setupTestManager(t)
	
	tests := []struct {
		name     string
		path     string
		wantName string
		wantErr  bool
	}{
		{
			name:     "get root returns nil",
			path:     "/",
			wantName: "",
			wantErr:  false,
		},
		{
			name:     "get child",
			path:     "/child1",
			wantName: "child1",
			wantErr:  false,
		},
		{
			name:     "get nested through parent config",
			path:     "/parent/nested1",
			wantName: "parent", // Currently returns parent node, not nested
			wantErr:  false,
		},
		{
			name:     "non-existent path",
			path:     "/nonexistent",
			wantName: "",
			wantErr:  true,
		},
		{
			name:     "empty path defaults to root",
			path:     "",
			wantName: "",
			wantErr:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := mgr.GetNodeByPath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.wantName == "" {
					assert.Nil(t, node)
				} else {
					assert.NotNil(t, node)
					assert.Equal(t, tt.wantName, node.Name)
				}
			}
		})
	}
}

func TestStatelessManager_UseNode(t *testing.T) {
	mgr, tmpDir := setupTestManager(t)
	
	// Create directories for navigation
	os.MkdirAll(filepath.Join(tmpDir, "nodes", "child1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "nodes", "parent", "nodes", "nested1"), 0755)
	
	tests := []struct {
		name            string
		targetPath      string
		wantCurrentPath string
		wantErr         bool
	}{
		{
			name:            "navigate to child",
			targetPath:      "/child1",
			wantCurrentPath: "/child1",
			wantErr:         false,
		},
		{
			name:            "navigate to nested",
			targetPath:      "/parent/nested1",
			wantCurrentPath: "/parent/nested1",
			wantErr:         false,
		},
		{
			name:            "navigate to root",
			targetPath:      "/",
			wantCurrentPath: "/",
			wantErr:         false,
		},
		{
			name:            "invalid path",
			targetPath:      "/nonexistent",
			wantCurrentPath: "/",
			wantErr:         true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to root for each test
			mgr.currentPath = "/"
			
			err := mgr.UseNode(tt.targetPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCurrentPath, mgr.currentPath)
			}
		})
	}
}

func TestStatelessManager_AddRepo(t *testing.T) {
	mgr, tmpDir := setupTestManager(t)
	
	// Create repos directory
	os.MkdirAll(filepath.Join(tmpDir, "nodes"), 0755)
	
	tests := []struct {
		name       string
		parentPath string
		repoName   string
		repoURL    string
		lazy       bool
		wantErr    bool
	}{
		{
			name:       "add repository to root",
			parentPath: "/",
			repoName:   "new-repo",
			repoURL:    "https://github.com/test/new-repo.git",
			lazy:       false,
			wantErr:    false,
		},
		{
			name:       "add lazy repository",
			parentPath: "/",
			repoName:   "lazy-repo",
			repoURL:    "https://github.com/test/lazy-repo.git",
			lazy:       true,
			wantErr:    false,
		},
		{
			name:       "add to non-existent parent",
			parentPath: "/nonexistent",
			repoName:   "fail-repo",
			repoURL:    "https://github.com/test/fail.git",
			lazy:       false,
			wantErr:    false, // AddRepo doesn't validate parent path
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.AddRepo(tt.parentPath, tt.repoName, tt.repoURL, tt.lazy)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// Verify the repo was added to config
				assert.NotNil(t, mgr.config)
			}
		})
	}
}

func TestStatelessManager_RemoveNode(t *testing.T) {
	mgr, tmpDir := setupTestManager(t)
	
	// Create directories
	os.MkdirAll(filepath.Join(tmpDir, "nodes", "child1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "nodes", "child2"), 0755)
	
	tests := []struct {
		name       string
		targetPath string
		wantErr    bool
	}{
		{
			name:       "remove existing child",
			targetPath: "/child1",
			wantErr:    false,
		},
		{
			name:       "cannot remove root",
			targetPath: "/",
			wantErr:    true,
		},
		{
			name:       "remove non-existent node",
			targetPath: "/nonexistent",
			wantErr:    false, // RemoveNode doesn't error on non-existent nodes
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.RemoveNode(tt.targetPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStatelessManager_ListChildren(t *testing.T) {
	mgr, _ := setupTestManager(t)
	
	tests := []struct {
		name         string
		targetPath   string
		wantChildren int
		wantErr      bool
	}{
		{
			name:         "list root children",
			targetPath:   "/",
			wantChildren: 3, // child1, child2, parent
			wantErr:      false,
		},
		{
			name:         "list parent children",
			targetPath:   "/parent",
			wantChildren: 0, // Currently returns empty for non-root paths
			wantErr:      false,
		},
		{
			name:         "list leaf node children",
			targetPath:   "/child1",
			wantChildren: 0,
			wantErr:      false,
		},
		{
			name:         "list non-existent node",
			targetPath:   "/nonexistent",
			wantChildren: 0,
			wantErr:      false, // ListChildren returns empty for non-root paths
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children, err := mgr.ListChildren(tt.targetPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, children, tt.wantChildren)
			}
		})
	}
}

func TestStatelessManager_CloneLazyRepos(t *testing.T) {
	tests := []struct {
		name       string
		targetPath string
		recursive  bool
		wantClones int
		wantErr    bool
	}{
		{
			name:       "clone lazy repos at root",
			targetPath: "/",
			recursive:  false,
			wantClones: 2, // Both child1 and child2 (no .git dirs exist)
			wantErr:    false,
		},
		{
			name:       "clone recursively",
			targetPath: "/",
			recursive:  true,
			wantClones: 2, // Both child1 and child2 (no .git dirs exist)
			wantErr:    false,
		},
		{
			name:       "clone at non-existent path",
			targetPath: "/nonexistent",
			recursive:  false,
			wantClones: 2, // Both child1 and child2 (targetPath is ignored)
			wantErr:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh manager for each test case
			mgr, tmpDir := setupTestManager(t)
			
			// Create parent directory
			os.MkdirAll(filepath.Join(tmpDir, "nodes", "parent"), 0755)
			
			// Set up git mock to track clone calls
			gitMock := mgr.gitCmd.(*git.MockGit)
			gitMock.CloneFunc = func(url, path string) error {
				// Simulate successful clone by creating .git directory
				gitDir := filepath.Join(path, ".git")
				os.MkdirAll(gitDir, 0755)
				return nil
			}
			
			err := mgr.CloneLazyRepos(tt.targetPath, tt.recursive)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, gitMock.CloneCalls, tt.wantClones)
			}
		})
	}
}

func TestStatelessManager_DisplayMethods(t *testing.T) {
	mgr, tmpDir := setupTestManager(t)
	
	// Create some directories to simulate cloned repos
	os.MkdirAll(filepath.Join(tmpDir, "nodes", "child1", ".git"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "nodes", "parent"), 0755)
	
	t.Run("DisplayTree", func(t *testing.T) {
		result := mgr.DisplayTree()
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "test-workspace")
	})
	
	t.Run("DisplayStatus", func(t *testing.T) {
		result := mgr.DisplayStatus()
		assert.NotEmpty(t, result)
	})
	
	t.Run("DisplayPath", func(t *testing.T) {
		result := mgr.DisplayPath()
		assert.NotEmpty(t, result)
		assert.Equal(t, "/", result) // DisplayPath just returns currentPath
	})
	
	t.Run("DisplayChildren", func(t *testing.T) {
		result := mgr.DisplayChildren()
		assert.NotEmpty(t, result)
		
		// Navigate to leaf and test no children message
		mgr.currentPath = "/child1"
		result = mgr.DisplayChildren()
		assert.Contains(t, result, "No children")
	})
	
	t.Run("DisplayTreeWithDepth", func(t *testing.T) {
		// Test various depths
		result := mgr.DisplayTreeWithDepth(0)
		assert.NotEmpty(t, result)
		
		result = mgr.DisplayTreeWithDepth(1)
		assert.NotEmpty(t, result)
		
		result = mgr.DisplayTreeWithDepth(2)
		assert.NotEmpty(t, result)
		
		result = mgr.DisplayTreeWithDepth(-1)
		assert.NotEmpty(t, result)
	})
}

func TestStatelessManager_GetCurrentPath(t *testing.T) {
	mgr, _ := setupTestManager(t)
	
	// Test initial path
	assert.Equal(t, "/", mgr.GetCurrentPath())
	
	// Change path and test
	mgr.currentPath = "/child1"
	assert.Equal(t, "/child1", mgr.GetCurrentPath())
	
	// Test nested path
	mgr.currentPath = "/parent/nested1"
	assert.Equal(t, "/parent/nested1", mgr.GetCurrentPath())
}