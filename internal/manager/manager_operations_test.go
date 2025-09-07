package manager

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/mocks"
)

// TestCloneRepos tests the CloneRepos function
func TestManager_CloneRepos(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		recursive bool
		setupMock func(*mocks.MockTreeProvider, *mocks.MockGitProvider, *mocks.MockConfigProvider)
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "clone lazy repos non-recursive",
			path:      "",
			recursive: false,
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, cfg *mocks.MockConfigProvider) {
				currentNode := interfaces.NodeInfo{
					Name: "root",
					Path: "/",
					Children: []interfaces.NodeInfo{
						{Name: "repo1", Path: "/repo1", IsLazy: true, IsCloned: false, Repository: "https://github.com/test/repo1.git"},
						{Name: "repo2", Path: "/repo2", IsLazy: false, IsCloned: true, Repository: "https://github.com/test/repo2.git"},
					},
				}
				tree.SetCurrent(currentNode)
				// repo1 should be cloned - mock will succeed by default
				tree.SetNode("/repo1", interfaces.NodeInfo{
					Name: "repo1", Path: "/repo1", IsLazy: false, IsCloned: true, Repository: "https://github.com/test/repo1.git",
				})
			},
			wantErr: false,
		},
		{
			name:      "clone recursive with nested lazy repos",
			path:      "",
			recursive: true,
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, cfg *mocks.MockConfigProvider) {
				currentNode := interfaces.NodeInfo{
					Name: "root",
					Path: "/",
					Children: []interfaces.NodeInfo{
						{
							Name: "parent", Path: "/parent", IsLazy: true, IsCloned: false, Repository: "https://github.com/test/parent.git",
							Children: []interfaces.NodeInfo{
								{Name: "child", Path: "/parent/child", IsLazy: true, IsCloned: false, Repository: "https://github.com/test/child.git"},
							},
						},
					},
				}
				tree.SetCurrent(currentNode)
				// Both parent and child should be cloned - mock will succeed by default
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTree := mocks.NewMockTreeProvider()
			mockGit := mocks.NewMockGitProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockUI := mocks.NewMockUIProvider()
			mockConfig := mocks.NewMockConfigProvider()
			mockLog := NewDefaultLogProvider(false)

			// Setup mocks
			tt.setupMock(mockTree, mockGit, mockConfig)

			// Create manager
			manager := &Manager{
				treeProvider:   mockTree,
				gitProvider:    mockGit,
				fsProvider:     mockFS,
				uiProvider:     mockUI,
				configProvider: mockConfig,
				logProvider:    mockLog,
				workspace:      "workspace",
				initialized:    true,
				config:         &config.ConfigTree{},
			}

			// Execute
			err := manager.CloneRepos(tt.path, tt.recursive)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestStatusNode tests the StatusNode function
func TestManager_StatusNode(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		recursive bool
		setupMock func(*mocks.MockTreeProvider, *mocks.MockGitProvider, *mocks.MockUIProvider)
		wantErr   bool
	}{
		{
			name:      "status for current node non-recursive",
			path:      "",
			recursive: false,
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				currentNode := interfaces.NodeInfo{
					Name: "repo1",
					Path: "/repo1",
				}
				tree.SetCurrent(currentNode)
				tree.SetNode("/repo1", currentNode)
				
				status := interfaces.GitStatus{
					Branch:  "main",
					IsClean: true,
				}
				git.SetStatus(filepath.Join("workspace", "/repo1"), &status)
			},
			wantErr: false,
		},
		{
			name:      "status for specific path",
			path:      "/repo2",
			recursive: false,
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				node := interfaces.NodeInfo{
					Name: "repo2",
					Path: "/repo2",
				}
				tree.SetNode("/repo2", node)
				
				status := interfaces.GitStatus{
					Branch:  "develop",
					IsClean: false,
				}
				git.SetStatus(filepath.Join("workspace", "/repo2"), &status)
			},
			wantErr: false,
		},
		{
			name:      "recursive status",
			path:      "",
			recursive: true,
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				currentNode := interfaces.NodeInfo{
					Name: "root",
					Path: "/",
					Children: []interfaces.NodeInfo{
						{Name: "repo1", Path: "/repo1"},
						{Name: "repo2", Path: "/repo2"},
					},
				}
				tree.SetCurrent(currentNode)
				tree.SetNode("/", currentNode)
				
				// Set status for all repos
				status1 := &interfaces.GitStatus{Branch: "main", IsClean: true}
				status2 := &interfaces.GitStatus{Branch: "feature", IsClean: false}
				status3 := &interfaces.GitStatus{Branch: "develop", IsClean: true}
				git.SetStatus(filepath.Join("workspace", "/"), status1)
				git.SetStatus(filepath.Join("workspace", "/repo1"), status2)
				git.SetStatus(filepath.Join("workspace", "/repo2"), status3)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTree := mocks.NewMockTreeProvider()
			mockGit := mocks.NewMockGitProvider()
			mockUI := mocks.NewMockUIProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockConfig := mocks.NewMockConfigProvider()

			// Setup mocks
			tt.setupMock(mockTree, mockGit, mockUI)

			// Create manager
			manager := &Manager{
				treeProvider:   mockTree,
				gitProvider:    mockGit,
				fsProvider:     mockFS,
				uiProvider:     mockUI,
				configProvider: mockConfig,
				workspace:      "workspace",
				initialized:    true,
			}

			// Execute
			err := manager.StatusNode(tt.path, tt.recursive)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPullNode tests the PullNode function
func TestManager_PullNode(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		recursive bool
		setupMock func(*mocks.MockTreeProvider, *mocks.MockGitProvider)
		wantErr   bool
	}{
		{
			name:      "pull current node non-recursive",
			path:      "",
			recursive: false,
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider) {
				currentNode := interfaces.NodeInfo{
					Name: "repo1",
					Path: "/repo1",
				}
				tree.SetCurrent(currentNode)
				tree.SetNode("/repo1", currentNode)
				// Pull will succeed by default in mock
			},
			wantErr: false,
		},
		{
			name:      "pull recursive",
			path:      "",
			recursive: true,
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider) {
				currentNode := interfaces.NodeInfo{
					Name:     "root",
					Path:     "/",
					IsCloned: true,
					Children: []interfaces.NodeInfo{
						{Name: "repo1", Path: "/repo1", IsCloned: true},
						{Name: "repo2", Path: "/repo2", IsCloned: false}, // Lazy, shouldn't be pulled
					},
				}
				tree.SetCurrent(currentNode)
				tree.SetNode("/", currentNode)
				
				// Only cloned repos should be pulled - mock will succeed by default
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTree := mocks.NewMockTreeProvider()
			mockGit := mocks.NewMockGitProvider()
			mockLog := NewDefaultLogProvider(false)
			mockFS := mocks.NewMockFileSystemProvider()
			mockUI := mocks.NewMockUIProvider()
			mockConfig := mocks.NewMockConfigProvider()

			// Setup mocks
			tt.setupMock(mockTree, mockGit)

			// Create manager
			manager := &Manager{
				treeProvider:   mockTree,
				gitProvider:    mockGit,
				fsProvider:     mockFS,
				uiProvider:     mockUI,
				configProvider: mockConfig,
				logProvider:    mockLog,
				workspace:      "workspace",
				initialized:    true,
			}

			// Execute
			err := manager.PullNode(tt.path, tt.recursive, false)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPushNode tests the PushNode function
func TestManager_PushNode(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		recursive bool
		setupMock func(*mocks.MockTreeProvider, *mocks.MockGitProvider)
		wantErr   bool
	}{
		{
			name:      "push current node non-recursive",
			path:      "",
			recursive: false,
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider) {
				currentNode := interfaces.NodeInfo{
					Name: "repo1",
					Path: "/repo1",
				}
				tree.SetCurrent(currentNode)
				tree.SetNode("/repo1", currentNode)
				// Push will succeed by default in mock
			},
			wantErr: false,
		},
		{
			name:      "push specific path",
			path:      "/repo2",
			recursive: false,
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider) {
				node := interfaces.NodeInfo{
					Name: "repo2",
					Path: "/repo2",
				}
				tree.SetNode("/repo2", node)
				// Push will succeed by default in mock
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTree := mocks.NewMockTreeProvider()
			mockGit := mocks.NewMockGitProvider()
			mockLog := NewDefaultLogProvider(false)
			mockFS := mocks.NewMockFileSystemProvider()
			mockUI := mocks.NewMockUIProvider()
			mockConfig := mocks.NewMockConfigProvider()

			// Setup mocks
			tt.setupMock(mockTree, mockGit)

			// Create manager
			manager := &Manager{
				treeProvider:   mockTree,
				gitProvider:    mockGit,
				fsProvider:     mockFS,
				uiProvider:     mockUI,
				configProvider: mockConfig,
				logProvider:    mockLog,
				workspace:      "workspace",
				initialized:    true,
			}

			// Execute
			err := manager.PushNode(tt.path, tt.recursive)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCommitNode tests the CommitNode function
func TestManager_CommitNode(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		message   string
		recursive bool
		setupMock func(*mocks.MockTreeProvider, *mocks.MockGitProvider)
		wantErr   bool
	}{
		{
			name:      "commit current node",
			path:      "",
			message:   "test commit",
			recursive: false,
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider) {
				currentNode := interfaces.NodeInfo{
					Name: "repo1",
					Path: "/repo1",
				}
				tree.SetCurrent(currentNode)
				tree.SetNode("/repo1", currentNode)
				// Commit will succeed by default in mock
			},
			wantErr: false,
		},
		{
			name:      "commit specific path",
			path:      "/repo2",
			message:   "fix: bug fix",
			recursive: false,
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider) {
				node := interfaces.NodeInfo{
					Name: "repo2",
					Path: "/repo2",
				}
				tree.SetNode("/repo2", node)
				// Commit will succeed by default in mock
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTree := mocks.NewMockTreeProvider()
			mockGit := mocks.NewMockGitProvider()
			mockLog := NewDefaultLogProvider(false)
			mockFS := mocks.NewMockFileSystemProvider()
			mockUI := mocks.NewMockUIProvider()
			mockConfig := mocks.NewMockConfigProvider()

			// Setup mocks
			tt.setupMock(mockTree, mockGit)

			// Create manager
			manager := &Manager{
				treeProvider:   mockTree,
				gitProvider:    mockGit,
				fsProvider:     mockFS,
				uiProvider:     mockUI,
				configProvider: mockConfig,
				logProvider:    mockLog,
				workspace:      "workspace",
				initialized:    true,
			}

			// Execute
			err := manager.CommitNode(tt.path, tt.message, tt.recursive)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestListNodesRecursive tests the ListNodesRecursive function
func TestManager_ListNodesRecursive(t *testing.T) {
	tests := []struct {
		name      string
		recursive bool
		setupMock func(*mocks.MockTreeProvider, *mocks.MockUIProvider)
		wantErr   bool
	}{
		{
			name:      "list non-recursive",
			recursive: false,
			setupMock: func(tree *mocks.MockTreeProvider, ui *mocks.MockUIProvider) {
				currentNode := interfaces.NodeInfo{
					Name: "root",
					Path: "/",
					Children: []interfaces.NodeInfo{
						{Name: "repo1"},
						{Name: "repo2"},
						{Name: "repo3"},
					},
				}
				tree.SetCurrent(currentNode)
			},
			wantErr: false,
		},
		{
			name:      "list recursive",
			recursive: true,
			setupMock: func(tree *mocks.MockTreeProvider, ui *mocks.MockUIProvider) {
				treeNode := interfaces.NodeInfo{
					Name: "root",
					Path: "/",
					Children: []interfaces.NodeInfo{
						{
							Name: "parent1",
							Children: []interfaces.NodeInfo{
								{Name: "child1", IsLazy: true},
								{Name: "child2", HasChanges: true},
							},
						},
						{
							Name:       "parent2",
							HasChanges: true,
						},
					},
				}
				tree.SetCurrent(treeNode) // Use SetCurrent for tree root
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTree := mocks.NewMockTreeProvider()
			mockUI := mocks.NewMockUIProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockGit := mocks.NewMockGitProvider()
			mockConfig := mocks.NewMockConfigProvider()

			// Setup mocks
			tt.setupMock(mockTree, mockUI)

			// Create manager
			manager := &Manager{
				treeProvider:   mockTree,
				uiProvider:     mockUI,
				fsProvider:     mockFS,
				gitProvider:    mockGit,
				configProvider: mockConfig,
				initialized:    true,
			}

			// Execute
			err := manager.ListNodesRecursive(tt.recursive)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestShowCurrent tests the ShowCurrent function
func TestManager_ShowCurrent(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MockTreeProvider, *mocks.MockUIProvider)
		wantErr   bool
	}{
		{
			name: "show current position",
			setupMock: func(tree *mocks.MockTreeProvider, ui *mocks.MockUIProvider) {
				currentNode := interfaces.NodeInfo{
					Name: "repo1",
					Path: "/workspace/repo1",
				}
				tree.SetCurrent(currentNode)
			},
			wantErr: false,
		},
		{
			name: "show root position",
			setupMock: func(tree *mocks.MockTreeProvider, ui *mocks.MockUIProvider) {
				currentNode := interfaces.NodeInfo{
					Name: "root",
					Path: "/",
				}
				tree.SetCurrent(currentNode)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTree := mocks.NewMockTreeProvider()
			mockUI := mocks.NewMockUIProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockGit := mocks.NewMockGitProvider()
			mockConfig := mocks.NewMockConfigProvider()

			// Setup mocks
			tt.setupMock(mockTree, mockUI)

			// Create manager
			manager := &Manager{
				treeProvider:   mockTree,
				uiProvider:     mockUI,
				fsProvider:     mockFS,
				gitProvider:    mockGit,
				configProvider: mockConfig,
				initialized:    true,
			}

			// Execute
			err := manager.ShowCurrent()

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestShowTreeAtPath tests the ShowTreeAtPath function
func TestManager_ShowTreeAtPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		depth     int
		setupMock func(*mocks.MockTreeProvider, *mocks.MockUIProvider)
		wantErr   bool
	}{
		{
			name:  "show tree at specific path",
			path:  "/repo1",
			depth: 2,
			setupMock: func(tree *mocks.MockTreeProvider, ui *mocks.MockUIProvider) {
				node := interfaces.NodeInfo{
					Name: "repo1",
					Path: "/repo1",
					Children: []interfaces.NodeInfo{
						{Name: "submodule1"},
						{Name: "submodule2", IsLazy: true},
					},
				}
				tree.SetNode("/repo1", node)
			},
			wantErr: false,
		},
		{
			name:  "empty path defaults to root",
			path:  "",
			depth: 1,
			setupMock: func(tree *mocks.MockTreeProvider, ui *mocks.MockUIProvider) {
				rootNode := interfaces.NodeInfo{
					Name: "root",
					Path: "/",
					Children: []interfaces.NodeInfo{
						{Name: "repo1"},
						{Name: "repo2"},
					},
				}
				tree.SetNode("/", rootNode)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTree := mocks.NewMockTreeProvider()
			mockUI := mocks.NewMockUIProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockGit := mocks.NewMockGitProvider()
			mockConfig := mocks.NewMockConfigProvider()

			// Setup mocks
			tt.setupMock(mockTree, mockUI)

			// Create manager
			manager := &Manager{
				treeProvider:   mockTree,
				uiProvider:     mockUI,
				fsProvider:     mockFS,
				gitProvider:    mockGit,
				configProvider: mockConfig,
				initialized:    true,
			}

			// Execute
			err := manager.ShowTreeAtPath(tt.path, tt.depth)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test the metrics provider
func TestNoOpMetricsProvider(t *testing.T) {
	provider := NewNoOpMetricsProvider()
	
	// Test Counter
	assert.NotPanics(t, func() {
		provider.Counter("test.counter", 1)
		provider.Counter("test.counter", 100, "tag1:value1")
	})
	
	// Test Gauge
	assert.NotPanics(t, func() {
		provider.Gauge("test.gauge", 1.5)
		provider.Gauge("test.gauge", 100.0, "tag1:value1")
	})
	
	// Test Histogram
	assert.NotPanics(t, func() {
		provider.Histogram("test.histogram", 1.5)
		provider.Histogram("test.histogram", 100.0, "tag1:value1")
	})
	
	// Test Timer
	timer := provider.Timer("test.timer")
	assert.NotNil(t, timer)
	assert.NotPanics(t, func() {
		timer.Start()
		duration := timer.Stop()
		assert.Equal(t, duration, timer.Stop())
		// Reset not available in interface
		timer.Record(duration)
	})
	
	// Test Flush
	err := provider.Flush()
	assert.NoError(t, err)
}

// TestCloseMethodExtra tests the Close function
func TestManager_CloseExtra(t *testing.T) {
	// Create manager with nil providers
	manager := &Manager{
		logProvider:     nil,
		metricsProvider: nil,
		pluginManager:   nil,
	}
	
	// Should not panic
	err := manager.Close()
	assert.NoError(t, err)
	
	// Create manager with providers
	manager2 := &Manager{
		logProvider:     NewDefaultLogProvider(false),
		metricsProvider: NewNoOpMetricsProvider(),
		pluginManager:   nil,
	}
	
	// Should not panic
	err = manager2.Close()
	assert.NoError(t, err)
}