package manager

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/mocks"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name    string
		opts    ManagerOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "success with all required providers",
			opts: ManagerOptions{
				ConfigProvider: mocks.NewMockConfigProvider(),
				GitProvider:    mocks.NewMockGitProvider(),
				FSProvider:     mocks.NewMockFileSystemProvider(),
				UIProvider:     mocks.NewMockUIProvider(),
				TreeProvider:   mocks.NewMockTreeProvider(),
			},
			wantErr: false,
		},
		{
			name: "error when ConfigProvider missing",
			opts: ManagerOptions{
				GitProvider:  mocks.NewMockGitProvider(),
				FSProvider:   mocks.NewMockFileSystemProvider(),
				UIProvider:   mocks.NewMockUIProvider(),
				TreeProvider: mocks.NewMockTreeProvider(),
			},
			wantErr: true,
			errMsg:  "ConfigProvider is required",
		},
		{
			name: "error when GitProvider missing",
			opts: ManagerOptions{
				ConfigProvider: mocks.NewMockConfigProvider(),
				FSProvider:     mocks.NewMockFileSystemProvider(),
				UIProvider:     mocks.NewMockUIProvider(),
				TreeProvider:   mocks.NewMockTreeProvider(),
			},
			wantErr: true,
			errMsg:  "GitProvider is required",
		},
		{
			name: "success with optional providers defaulted",
			opts: ManagerOptions{
				ConfigProvider: mocks.NewMockConfigProvider(),
				GitProvider:    mocks.NewMockGitProvider(),
				FSProvider:     mocks.NewMockFileSystemProvider(),
				UIProvider:     mocks.NewMockUIProvider(),
				TreeProvider:   mocks.NewMockTreeProvider(),
				// Optional providers will be defaulted
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.opts)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, manager)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
				
				// Verify optional providers were defaulted
				assert.NotNil(t, manager.processProvider)
				assert.NotNil(t, manager.logProvider)
				assert.NotNil(t, manager.metricsProvider)
			}
		})
	}
}

func TestManager_Initialize(t *testing.T) {
	tests := []struct {
		name      string
		workspace string
		setupMock func(*mocks.MockConfigProvider, *mocks.MockFileSystemProvider, *mocks.MockTreeProvider)
		autoLoad  bool
		wantErr   bool
	}{
		{
			name:      "success creating new workspace",
			workspace: "/test/workspace",
			setupMock: func(cfg *mocks.MockConfigProvider, fs *mocks.MockFileSystemProvider, tree *mocks.MockTreeProvider) {
				fs.SetExists("/test/workspace", false)
				// MkdirAll will be called
			},
			autoLoad: false,
			wantErr:  false,
		},
		{
			name:      "success with existing workspace",
			workspace: "/test/workspace",
			setupMock: func(cfg *mocks.MockConfigProvider, fs *mocks.MockFileSystemProvider, tree *mocks.MockTreeProvider) {
				fs.SetExists("/test/workspace", true)
			},
			autoLoad: false,
			wantErr:  false,
		},
		{
			name:      "success with auto-load config",
			workspace: "/test/workspace",
			setupMock: func(cfg *mocks.MockConfigProvider, fs *mocks.MockFileSystemProvider, tree *mocks.MockTreeProvider) {
				fs.SetExists("/test/workspace", true)
				cfg.SetExists("/test/workspace/muno.yaml", true)
				cfg.SetConfig("/test/workspace/muno.yaml", &config.ConfigTree{
					Workspace: config.WorkspaceTree{
						Name: "test",
					},
				})
			},
			autoLoad: true,
			wantErr:  false,
		},
		{
			name:      "error when config load fails",
			workspace: "/test/workspace",
			setupMock: func(cfg *mocks.MockConfigProvider, fs *mocks.MockFileSystemProvider, tree *mocks.MockTreeProvider) {
				fs.SetExists("/test/workspace", true)
				cfg.SetExists("/test/workspace/muno.yaml", true)
				cfg.SetError("/test/workspace/muno.yaml", errors.New("test error"))
			},
			autoLoad: true,
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockConfig := mocks.NewMockConfigProvider()
			mockGit := mocks.NewMockGitProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockUI := mocks.NewMockUIProvider()
			mockTree := mocks.NewMockTreeProvider()
			
			// Setup mocks
			tt.setupMock(mockConfig, mockFS, mockTree)
			
			// Create manager
			manager, err := NewManager(ManagerOptions{
				ConfigProvider: mockConfig,
				GitProvider:    mockGit,
				FSProvider:     mockFS,
				UIProvider:     mockUI,
				TreeProvider:   mockTree,
				AutoLoadConfig: tt.autoLoad,
			})
			require.NoError(t, err)
			
			// Test Initialize
			ctx := context.Background()
			err = manager.Initialize(ctx, tt.workspace)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.workspace, manager.workspace)
				assert.True(t, manager.initialized)
			}
			
			// Verify mock calls
			calls := mockFS.GetCalls()
			assert.Contains(t, calls, "Exists(/test/workspace)")
			
			if !tt.wantErr && tt.autoLoad {
				configCalls := mockConfig.GetCalls()
				assert.Contains(t, configCalls, "Exists(/test/workspace/muno.yaml)")
			}
		})
	}
}

// TestManager_Use was removed - Use method no longer exists in stateless architecture
/*
func TestManager_Use(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		setupMock func(*mocks.MockTreeProvider, *mocks.MockGitProvider, *mocks.MockUIProvider)
		wantErr   bool
	}{
		{
			name: "success navigating to cloned node",
			path: "backend/service-a",
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				tree.SetCurrent(interfaces.NodeInfo{
					Name:       "service-a",
					Path:       "backend/service-a",
					Repository: "https://github.com/org/service-a",
					IsCloned:   true,
					IsLazy:     false,
				})
			},
			wantErr: false,
		},
		{
			name: "success cloning lazy repository",
			path: "backend/service-b",
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				tree.SetCurrent(interfaces.NodeInfo{
					Name:       "service-b",
					Path:       "backend/service-b",
					Repository: "https://github.com/org/service-b",
					IsCloned:   false,
					IsLazy: true,
				})
				// Git clone will be called
			},
			wantErr: false,
		},
		{
			name: "error when not initialized",
			path: "backend/service-a",
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				// Manager not initialized
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockConfig := mocks.NewMockConfigProvider()
			mockGit := mocks.NewMockGitProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockUI := mocks.NewMockUIProvider()
			mockTree := mocks.NewMockTreeProvider()
			
			// Setup mocks
			tt.setupMock(mockTree, mockGit, mockUI)
			
			// Create and initialize manager
			manager, err := NewManager(ManagerOptions{
				ConfigProvider: mockConfig,
				GitProvider:    mockGit,
				FSProvider:     mockFS,
				UIProvider:     mockUI,
				TreeProvider:   mockTree,
			})
			require.NoError(t, err)
			
			// Initialize if needed for test
			if tt.name != "error when not initialized" {
				mockFS.SetExists("/test", true)
				err = manager.Initialize(context.Background(), "/test")
				require.NoError(t, err)
			}
			
			// Test Use
			ctx := context.Background()
			err = nil // Use method was removed in stateless migration
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify navigation happened
				treeCalls := mockTree.GetCalls()
				assert.Contains(t, treeCalls, "Navigate("+tt.path+")")
				
				// Verify success message was shown (not necessarily the last message)
				uiMessages := mockUI.GetMessages()
				hasSuccess := false
				for _, msg := range uiMessages {
					if strings.Contains(msg, "SUCCESS") || strings.Contains(msg, "Added") || strings.Contains(msg, "Repository") {
						hasSuccess = true
						break
					}
				}
				assert.True(t, hasSuccess, "Should have success message in output")
			}
		})
	}
}
*/

// TestManager_Use_LazyChildCloning was removed - Use method no longer exists
/*
func TestManager_Use_LazyChildCloning(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		setupMock    func(*mocks.MockTreeProvider, *mocks.MockGitProvider, *mocks.MockUIProvider)
		wantClone    bool
		wantErr      bool
		description  string
	}{
		{
			name: "navigating to lazy node triggers clone",
			path: "backend/lazy-service",
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				// Set current node as lazy and not cloned
				tree.SetCurrent(interfaces.NodeInfo{
					Name:       "lazy-service",
					Path:       "backend/lazy-service",
					Repository: "https://github.com/org/lazy-service",
					IsCloned:   false,
					IsLazy:     true,
					Children:   []interfaces.NodeInfo{},
				})
			},
			wantClone:   true,
			wantErr:     false,
			description: "Should clone lazy repository when navigating to it",
		},
		{
			name: "navigating to parent with lazy children triggers child clones",
			path: "backend",
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				// Parent node with lazy children
				tree.SetCurrent(interfaces.NodeInfo{
					Name:       "backend",
					Path:       "backend",
					Repository: "",
					IsCloned:   true,
					IsLazy:     false,
					Children: []interfaces.NodeInfo{
						{
							Name:       "service-a",
							Path:       "backend/service-a",
							Repository: "https://github.com/org/service-a",
							IsCloned:   false,
							IsLazy:     true,
						},
						{
							Name:       "service-b",
							Path:       "backend/service-b",
							Repository: "https://github.com/org/service-b",
							IsCloned:   false,
							IsLazy:     true,
						},
					},
				})
			},
			wantClone:   false, // Lazy children are NOT auto-cloned anymore (commented out in manager.go)
			wantErr:     false,
			description: "Should NOT auto-clone lazy children when navigating to parent",
		},
		{
			name: "navigating to already cloned node does not trigger clone",
			path: "backend/cloned-service",
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				tree.SetCurrent(interfaces.NodeInfo{
					Name:       "cloned-service",
					Path:       "backend/cloned-service",
					Repository: "https://github.com/org/cloned-service",
					IsCloned:   true,
					IsLazy:     false,
					Children:   []interfaces.NodeInfo{},
				})
			},
			wantClone:   false,
			wantErr:     false,
			description: "Should not clone already cloned repository",
		},
		{
			name: "navigating to node without repository does not trigger clone",
			path: "backend/config-node",
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				tree.SetCurrent(interfaces.NodeInfo{
					Name:       "config-node",
					Path:       "backend/config-node",
					Repository: "", // Nodes without repository URL (config nodes, parent folders)
					IsCloned:   false,
					IsLazy:     true,
					Children:   []interfaces.NodeInfo{},
				})
			},
			wantClone:   false,
			wantErr:     false,
			description: "Should not clone nodes without repository URL",
		},
		{
			name: "navigating to parent with mixed children only clones lazy ones",
			path: "platform",
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				tree.SetCurrent(interfaces.NodeInfo{
					Name:       "platform",
					Path:       "platform",
					Repository: "",
					IsCloned:   true,
					IsLazy:     false,
					Children: []interfaces.NodeInfo{
						{
							Name:       "already-cloned-service",
							Path:       "platform/already-cloned-service",
							Repository: "https://github.com/org/cloned",
							IsCloned:   true,
							IsLazy:     false,
						},
						{
							Name:       "lazy-service",
							Path:       "platform/lazy-service",
							Repository: "https://github.com/org/lazy",
							IsCloned:   false,
							IsLazy:     true,
						},
						{
							Name:       "config-child",
							Path:       "platform/config-child",
							Repository: "", // Config node
							IsCloned:   false,
							IsLazy:     true,
						},
					},
				})
			},
			wantClone:   false, // Lazy children are NOT auto-cloned anymore (commented out in manager.go)
			wantErr:     false,
			description: "Should NOT auto-clone lazy children even with mixed types",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockConfig := mocks.NewMockConfigProvider()
			mockGit := mocks.NewMockGitProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockUI := mocks.NewMockUIProvider()
			mockTree := mocks.NewMockTreeProvider()
			
			// Setup mocks
			tt.setupMock(mockTree, mockGit, mockUI)
			
			// Create and initialize manager
			manager, err := NewManager(ManagerOptions{
				ConfigProvider: mockConfig,
				GitProvider:    mockGit,
				FSProvider:     mockFS,
				UIProvider:     mockUI,
				TreeProvider:   mockTree,
			})
			require.NoError(t, err)
			
			// Initialize manager
			mockFS.SetExists("/test", true)
			err = manager.Initialize(context.Background(), "/test")
			require.NoError(t, err)
			
			// Test Use
			ctx := context.Background()
			err = nil // Use method was removed in stateless migration
			
			if tt.wantErr {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				
				// Verify navigation happened
				treeCalls := mockTree.GetCalls()
				assert.Contains(t, treeCalls, "Navigate("+tt.path+")", "Should navigate to path")
				
				// Verify clone was called if expected
				gitCalls := mockGit.GetCalls()
				if tt.wantClone {
					// Should have called Clone
					hasClone := false
					for _, call := range gitCalls {
						if strings.Contains(call, "Clone(") {
							hasClone = true
							break
						}
					}
					assert.True(t, hasClone, "Should have called git.Clone for lazy repository")
					
					// Verify node state was updated
					hasUpdate := false
					for _, call := range treeCalls {
						if strings.Contains(call, "UpdateNode(") {
							hasUpdate = true
							break
						}
					}
					assert.True(t, hasUpdate, "Should update node state after cloning")
				} else {
					// Should NOT have called Clone
					for _, call := range gitCalls {
						assert.NotContains(t, call, "Clone(", "Should not call git.Clone for "+tt.description)
					}
				}
			}
		})
	}
}
*/

func TestManager_Add(t *testing.T) {
	tests := []struct {
		name      string
		repoURL   string
		options   AddOptions
		setupMock func(*mocks.MockTreeProvider, *mocks.MockGitProvider, *mocks.MockUIProvider)
		wantErr   bool
	}{
		{
			name:    "success adding lazy repository",
			repoURL: "https://github.com/org/new-service",
			options: AddOptions{
				Fetch: "lazy",
			},
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				// In stateless mode, always at root
				tree.SetCurrent(interfaces.NodeInfo{
					Name: "root",
					Path: "/",
				})
			},
			wantErr: false,
		},
		{
			name:    "success adding and cloning repository",
			repoURL: "https://github.com/org/new-service",
			options: AddOptions{
				Fetch:     "eager",
				Recursive: true,
			},
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				// In stateless mode, always at root
				tree.SetCurrent(interfaces.NodeInfo{
					Name: "root",
					Path: "/",
				})
				// Git clone will be called
			},
			wantErr: false,
		},
		{
			name:    "error when not initialized",
			repoURL: "https://github.com/org/new-service",
			options: AddOptions{},
			setupMock: func(tree *mocks.MockTreeProvider, git *mocks.MockGitProvider, ui *mocks.MockUIProvider) {
				// Manager not initialized
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockConfig := mocks.NewMockConfigProvider()
			mockGit := mocks.NewMockGitProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockUI := mocks.NewMockUIProvider()
			mockTree := mocks.NewMockTreeProvider()
			
			// Setup mocks
			tt.setupMock(mockTree, mockGit, mockUI)
			
			// Create manager
			manager, err := NewManager(ManagerOptions{
				ConfigProvider: mockConfig,
				GitProvider:    mockGit,
				FSProvider:     mockFS,
				UIProvider:     mockUI,
				TreeProvider:   mockTree,
			})
			require.NoError(t, err)
			
			// Initialize if needed for test
			if tt.name != "error when not initialized" {
				mockFS.SetExists("/test", true)
				err = manager.Initialize(context.Background(), "/test")
				require.NoError(t, err)
			}
			
			// Test Add
			ctx := context.Background()
			err = manager.Add(ctx, tt.repoURL, tt.options)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify node was added
				treeCalls := mockTree.GetCalls()
				// Check that AddNode was called (may not be the last call if UpdateNode followed)
				hasAddNode := false
				for _, call := range treeCalls {
					if strings.Contains(call, "AddNode") {
						hasAddNode = true
						break
					}
				}
				assert.True(t, hasAddNode, "AddNode should have been called")
				
				// Verify clone was called if not lazy
				if tt.options.Fetch == "eager" {
					gitCalls := mockGit.GetCalls()
					assert.Contains(t, gitCalls[len(gitCalls)-1], "Clone")
				}
				
				// Verify success message was shown (not necessarily the last message)
				uiMessages := mockUI.GetMessages()
				hasSuccess := false
				for _, msg := range uiMessages {
					if strings.Contains(msg, "SUCCESS") || strings.Contains(msg, "Added") || strings.Contains(msg, "Repository") {
						hasSuccess = true
						break
					}
				}
				assert.True(t, hasSuccess, "Should have success message in output")
			}
		})
	}
}

func TestManager_Remove(t *testing.T) {
	tests := []struct {
		name      string
		repoName  string
		confirm   bool
		setupMock func(*mocks.MockTreeProvider, *mocks.MockFileSystemProvider, *mocks.MockUIProvider)
		wantErr   bool
	}{
		{
			name:     "success removing repository",
			repoName: "service-a",
			confirm:  true,
			setupMock: func(tree *mocks.MockTreeProvider, fs *mocks.MockFileSystemProvider, ui *mocks.MockUIProvider) {
				// In stateless mode, always at root
				tree.SetCurrent(interfaces.NodeInfo{
					Name: "root",
					Path: "/",
				})
				tree.SetNode("/service-a", interfaces.NodeInfo{
					Name:     "service-a",
					Path:     "/service-a",
					IsCloned: true,
				})
				fs.SetExists("/test/repos/service-a", true)
				ui.SetConfirm("Remove service-a and all its contents?", true)
			},
			wantErr: false,
		},
		{
			name:     "cancelled when user declines",
			repoName: "service-a",
			confirm:  false,
			setupMock: func(tree *mocks.MockTreeProvider, fs *mocks.MockFileSystemProvider, ui *mocks.MockUIProvider) {
				// In stateless mode, always at root
				tree.SetCurrent(interfaces.NodeInfo{
					Name: "root",
					Path: "/",
				})
				tree.SetNode("/service-a", interfaces.NodeInfo{
					Name: "service-a",
					Path: "/service-a",
				})
				ui.SetConfirm("Remove service-a and all its contents?", false)
			},
			wantErr: false,
		},
		{
			name:     "error when repository not found",
			repoName: "non-existent",
			confirm:  true,
			setupMock: func(tree *mocks.MockTreeProvider, fs *mocks.MockFileSystemProvider, ui *mocks.MockUIProvider) {
				// In stateless mode, always at root
				tree.SetCurrent(interfaces.NodeInfo{
					Name: "root",
					Path: "/",
				})
				tree.SetError("GetNode", "/non-existent", errors.New("test error"))
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockConfig := mocks.NewMockConfigProvider()
			mockGit := mocks.NewMockGitProvider()
			mockFS := mocks.NewMockFileSystemProvider()
			mockUI := mocks.NewMockUIProvider()
			mockTree := mocks.NewMockTreeProvider()
			
			// Setup mocks
			tt.setupMock(mockTree, mockFS, mockUI)
			
			// Create and initialize manager
			manager, err := NewManager(ManagerOptions{
				ConfigProvider: mockConfig,
				GitProvider:    mockGit,
				FSProvider:     mockFS,
				UIProvider:     mockUI,
				TreeProvider:   mockTree,
			})
			require.NoError(t, err)
			
			mockFS.SetExists("/test", true)
			err = manager.Initialize(context.Background(), "/test")
			require.NoError(t, err)
			
			// Test Remove
			ctx := context.Background()
			err = manager.Remove(ctx, tt.repoName)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				if tt.confirm {
					// Verify node was removed
					treeCalls := mockTree.GetCalls()
					assert.Contains(t, treeCalls[len(treeCalls)-1], "RemoveNode")
					
					// Verify success message - check all messages for SUCCESS
					uiMessages := mockUI.GetMessages()
					found := false
					for _, msg := range uiMessages {
						if strings.Contains(msg, "SUCCESS") {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected SUCCESS message in output")
				} else {
					// Verify cancellation message
					uiMessages := mockUI.GetMessages()
					assert.Contains(t, uiMessages[len(uiMessages)-1], "cancelled")
				}
			}
		})
	}
}
func TestManager_InitializeWithConfig(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	
	// Set up the config to be loaded
	existingConfig := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "existing-workspace",
			ReposDir: "nodes",
		},
	}
	mockConfig.SetConfig("/test/workspace/muno.yaml", existingConfig)
	mockFS.SetExists("/test/workspace/muno.yaml", true)
	mockFS.SetExists("/test/workspace/repos", true)
	
	manager, err := NewManager(ManagerOptions{
		ConfigProvider: mockConfig,
		GitProvider:    mockGit,
		FSProvider:     mockFS,
		UIProvider:     mockUI,
		TreeProvider:   mockTree,
	})
	require.NoError(t, err)
	
	// Initialize with existing config
	err = manager.InitializeWithConfig(context.Background(), "/test/workspace", existingConfig)
	assert.NoError(t, err)
	assert.True(t, manager.initialized)
	assert.Equal(t, "/test/workspace", manager.workspace)
}

func TestManager_ExecutePluginCommand(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	mockPlugin := mocks.NewMockPluginManager()
	
	manager, err := NewManager(ManagerOptions{
		ConfigProvider: mockConfig,
		GitProvider:    mockGit,
		FSProvider:     mockFS,
		UIProvider:     mockUI,
		TreeProvider:   mockTree,
		PluginManager:  mockPlugin,
	})
	require.NoError(t, err)
	
	// Initialize manager first
	mockFS.SetExists("/test/workspace/muno.yaml", false)
	mockFS.SetExists("/test/workspace/repos", false)
	err = manager.Initialize(context.Background(), "/test/workspace")
	require.NoError(t, err)
	
	// Set up plugin response
	mockPlugin.SetCommandResult("github", interfaces.Result{
		Success: true,
		Message: "Pull request created",
		Data:    map[string]interface{}{"pr_number": 123},
	})
	
	// Execute plugin command
	err = manager.ExecutePluginCommand(context.Background(), "github", []string{"pr", "create"})
	assert.NoError(t, err)
	
	// Test error case
	mockPlugin.SetCommandError("github", errors.New("plugin error"))
	err = manager.ExecutePluginCommand(context.Background(), "github", []string{"pr", "close"})
	assert.Error(t, err)
}

func TestManager_Close(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	
	manager, err := NewManager(ManagerOptions{
		ConfigProvider: mockConfig,
		GitProvider:    mockGit,
		FSProvider:     mockFS,
		UIProvider:     mockUI,
		TreeProvider:   mockTree,
	})
	require.NoError(t, err)
	
	// Close should not error
	err = manager.Close()
	assert.NoError(t, err)
}

func TestManager_ExtractRepoName(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com/user/repo.git", "repo"},
		{"https://github.com/user/repo", "repo"},
		{"git@github.com:user/repo.git", "repo"},
		{"https://gitlab.com/user/my-project.git", "my-project"},
		{"ssh://git@bitbucket.org/user/repo", "repo"},
		{"https://github.com/user/repo-with-dash.git", "repo-with-dash"},
		{"repo", "repo"},
		{"", ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := extractRepoName(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_SaveConfig(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	
	// Create manager with AutoLoadConfig enabled
	manager, err := NewManager(ManagerOptions{
		ConfigProvider: mockConfig,
		GitProvider:    mockGit,
		FSProvider:     mockFS,
		UIProvider:     mockUI,
		TreeProvider:   mockTree,
		AutoLoadConfig: true,  // Enable auto-loading
	})
	require.NoError(t, err)
	
	// Set up existing config
	existingConfig := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: "nodes",
		},
	}
	mockConfig.SetConfig("/test/workspace/muno.yaml", existingConfig)
	mockFS.SetExists("/test/workspace/muno.yaml", true)
	mockFS.SetExists("/test/workspace/repos", true)
	mockConfig.SetExists("/test/workspace/muno.yaml", true)
	
	err = manager.Initialize(context.Background(), "/test/workspace")
	require.NoError(t, err)
	
	// Save config should work
	err = manager.saveConfig()
	assert.NoError(t, err)
	
	// Test save error
	mockConfig.SetError("/test/workspace/muno.yaml", errors.New("save error"))
	err = manager.saveConfig()
	assert.Error(t, err)
}

func TestManager_HandlePluginAction(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	
	manager, err := NewManager(ManagerOptions{
		ConfigProvider: mockConfig,
		GitProvider:    mockGit,
		FSProvider:     mockFS,
		UIProvider:     mockUI,
		TreeProvider:   mockTree,
	})
	require.NoError(t, err)
	
	// Initialize manager
	mockFS.SetExists("/test/workspace/muno.yaml", false)
	mockFS.SetExists("/test/workspace/repos", false)
	err = manager.Initialize(context.Background(), "/test/workspace")
	require.NoError(t, err)
	
	// Test various action types
	actions := []struct {
		action  interfaces.Action
		wantErr bool
	}{
		{interfaces.Action{Type: "navigate", Path: "/backend"}, true}, // Navigation not supported in stateless
		{interfaces.Action{Type: "clone", URL: "https://github.com/test/repo.git"}, false},
		{interfaces.Action{Type: "message", Message: "Operation completed"}, false},
		{interfaces.Action{Type: "unknown", Message: "ignored"}, false},
	}
	
	for _, tc := range actions {
		err = manager.handlePluginAction(context.Background(), tc.action)
		if tc.wantErr {
			assert.Error(t, err, "Action %s should error in stateless mode", tc.action.Type)
		} else {
			assert.NoError(t, err, "Action %s should not error", tc.action.Type)
		}
	}
}

func TestManager_EdgeCases(t *testing.T) {
	t.Run("NewManager with nil providers", func(t *testing.T) {
		_, err := NewManager(ManagerOptions{
			ConfigProvider: nil,
			GitProvider:    mocks.NewMockGitProvider(),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ConfigProvider is required")
	})
	
	t.Run("Operations before initialization", func(t *testing.T) {
		manager, err := NewManager(ManagerOptions{
			ConfigProvider: mocks.NewMockConfigProvider(),
			GitProvider:    mocks.NewMockGitProvider(),
			FSProvider:     mocks.NewMockFileSystemProvider(),
			UIProvider:     mocks.NewMockUIProvider(),
			TreeProvider:   mocks.NewMockTreeProvider(),
		})
		require.NoError(t, err)
		
		// All operations should fail before initialization
		err = fmt.Errorf("not initialized") // Use method was removed in stateless migration
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		
		err = manager.Add(context.Background(), "https://github.com/test/repo", AddOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		
		err = manager.Remove(context.Background(), "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}

func TestManager_PluginLifecycle(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	mockPlugin := mocks.NewMockPluginManager()
	mockProcess := mocks.NewMockProcessProvider()
	
	manager, err := NewManager(ManagerOptions{
		ConfigProvider:  mockConfig,
		GitProvider:     mockGit,
		FSProvider:      mockFS,
		UIProvider:      mockUI,
		TreeProvider:    mockTree,
		PluginManager:   mockPlugin,
		ProcessProvider: mockProcess,
	})
	require.NoError(t, err)
	
	// Test Close with plugin manager
	err = manager.Close()
	assert.NoError(t, err)
}

func TestManager_HandlePluginActionNavigation(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockGit := mocks.NewMockGitProvider()
	mockFS := mocks.NewMockFileSystemProvider()
	mockUI := mocks.NewMockUIProvider()
	mockTree := mocks.NewMockTreeProvider()
	
	manager, err := NewManager(ManagerOptions{
		ConfigProvider: mockConfig,
		GitProvider:    mockGit,
		FSProvider:     mockFS,
		UIProvider:     mockUI,
		TreeProvider:   mockTree,
		AutoLoadConfig: true,
	})
	require.NoError(t, err)
	
	// Initialize with config
	existingConfig := &config.ConfigTree{
		Workspace: config.WorkspaceTree{
			Name:     "test-workspace",
			ReposDir: "nodes",
		},
	}
	mockConfig.SetConfig("/test/workspace/muno.yaml", existingConfig)
	mockFS.SetExists("/test/workspace/muno.yaml", true)
	mockFS.SetExists("/test/workspace/repos", true)
	mockConfig.SetExists("/test/workspace/muno.yaml", true)
	
	err = manager.Initialize(context.Background(), "/test/workspace")
	require.NoError(t, err)
	
	// Set up tree navigation
	mockTree.SetCurrent(interfaces.NodeInfo{
		Name: "root",
		Path: "/",
	})
	mockTree.SetNode("/backend", interfaces.NodeInfo{
		Name:     "backend",
		Path:     "/backend",
		IsCloned: true,
	})
	
	// Test navigate action (not supported in stateless mode)
	action := interfaces.Action{
		Type: "navigate",
		Path: "/backend",
	}
	
	err = manager.handlePluginAction(context.Background(), action)
	assert.Error(t, err, "Navigation should error in stateless mode")
	assert.Contains(t, err.Error(), "navigation action not supported")
	
	// Test command action
	action = interfaces.Action{
		Type:      "command",
		Command:   "test",
		Arguments: []string{"arg1", "arg2"},
	}
	
	// Without plugin manager, command action returns error
	err = manager.handlePluginAction(context.Background(), action)
	assert.Error(t, err)  // Should error when plugin manager is nil
	assert.Contains(t, err.Error(), "plugins not enabled")
}

func TestManager_NewManagerErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		opts    ManagerOptions
		wantErr string
	}{
		{
			name: "missing GitProvider",
			opts: ManagerOptions{
				ConfigProvider: mocks.NewMockConfigProvider(),
			},
			wantErr: "GitProvider is required",
		},
		{
			name: "missing FSProvider",
			opts: ManagerOptions{
				ConfigProvider: mocks.NewMockConfigProvider(),
				GitProvider:    mocks.NewMockGitProvider(),
			},
			wantErr: "FSProvider is required",
		},
		{
			name: "missing UIProvider",
			opts: ManagerOptions{
				ConfigProvider: mocks.NewMockConfigProvider(),
				GitProvider:    mocks.NewMockGitProvider(),
				FSProvider:     mocks.NewMockFileSystemProvider(),
			},
			wantErr: "UIProvider is required",
		},
		{
			name: "missing TreeProvider",
			opts: ManagerOptions{
				ConfigProvider: mocks.NewMockConfigProvider(),
				GitProvider:    mocks.NewMockGitProvider(),
				FSProvider:     mocks.NewMockFileSystemProvider(),
				UIProvider:     mocks.NewMockUIProvider(),
			},
			wantErr: "TreeProvider is required",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewManager(tt.opts)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
