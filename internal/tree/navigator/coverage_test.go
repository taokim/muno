package navigator

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
)

// TestGitHubHTTPSToSSH tests the GitHub HTTPS to SSH URL conversion
func TestGitHubHTTPSToSSH(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedURL string
		isGitHub    bool
	}{
		{
			name:        "standard_github_https",
			input:       "https://github.com/user/repo.git",
			expectedURL: "git@github.com:user/repo.git",
			isGitHub:    true,
		},
		{
			name:        "github_https_without_git_suffix",
			input:       "https://github.com/user/repo",
			expectedURL: "git@github.com:user/repo.git",
			isGitHub:    true,
		},
		{
			name:        "github_https_with_trailing_slash",
			input:       "https://github.com/user/repo/",
			expectedURL: "git@github.com:user/repo.git",
			isGitHub:    true,
		},
		{
			name:        "non_github_url",
			input:       "https://gitlab.com/user/repo.git",
			expectedURL: "https://gitlab.com/user/repo.git",
			isGitHub:    false,
		},
		{
			name:        "ssh_url_unchanged",
			input:       "git@github.com:user/repo.git",
			expectedURL: "git@github.com:user/repo.git",
			isGitHub:    false,
		},
		{
			name:        "invalid_url",
			input:       "not-a-url",
			expectedURL: "not-a-url",
			isGitHub:    false,
		},
		{
			name:        "github_enterprise_not_matched",
			input:       "https://github.enterprise.com/user/repo.git",
			expectedURL: "https://github.enterprise.com/user/repo.git",
			isGitHub:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, isGitHub := gitHubHTTPSToSSH(tt.input)
			assert.Equal(t, tt.expectedURL, result)
			assert.Equal(t, tt.isGitHub, isGitHub)
		})
	}
}

// TestIsSSHAuthError tests SSH authentication error detection
func TestIsSSHAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil_error",
			err:      nil,
			expected: false,
		},
		{
			name:     "permission_denied",
			err:      errors.New("Permission denied (publickey)"),
			expected: true,
		},
		{
			name:     "host_key_verification",
			err:      errors.New("Host key verification failed"),
			expected: true,
		},
		{
			name:     "ssh_connect_error",
			err:      errors.New("ssh: connect to host github.com port 22: Connection refused"),
			expected: true,
		},
		{
			name:     "git_permission_denied",
			err:      errors.New("git@github.com: Permission denied"),
			expected: true,
		},
		{
			name:     "could_not_read_remote",
			err:      errors.New("fatal: Could not read from remote repository"),
			expected: true,
		},
		{
			name:     "remote_hung_up",
			err:      errors.New("fatal: The remote end hung up unexpectedly"),
			expected: true,
		},
		{
			name:     "generic_error",
			err:      errors.New("generic error message"),
			expected: false,
		},
		{
			name:     "network_error_no_ssh",
			err:      errors.New("network timeout"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSSHAuthError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCachedNavigatorPutInCache tests cache insertion with size limits and eviction
func TestCachedNavigatorPutInCache(t *testing.T) {
	t.Run("cache_respects_max_size", func(t *testing.T) {
		base := NewInMemoryNavigator()
		setupTestTree(base)
		cached := NewCachedNavigator(base, 1*time.Hour)
		cached.WithMaxSize(3)

		// Fill cache to max size
		cached.putInCache("key1", "value1")
		cached.putInCache("key2", "value2")
		cached.putInCache("key3", "value3")

		// Verify all cached
		assert.NotNil(t, cached.getFromCache("key1"))
		assert.NotNil(t, cached.getFromCache("key2"))
		assert.NotNil(t, cached.getFromCache("key3"))

		// Add one more, should evict oldest
		cached.putInCache("key4", "value4")

		// key4 should exist
		assert.NotNil(t, cached.getFromCache("key4"))
		// Cache should still be at max size
		assert.LessOrEqual(t, len(cached.cache), 3)
	})

	t.Run("cache_evicts_expired_first", func(t *testing.T) {
		base := NewInMemoryNavigator()
		setupTestTree(base)
		cached := NewCachedNavigator(base, 1*time.Millisecond) // Very short TTL
		cached.WithMaxSize(2)

		// Add entries
		cached.putInCache("key1", "value1")
		time.Sleep(5 * time.Millisecond) // Let key1 expire

		cached.putInCache("key2", "value2")
		cached.putInCache("key3", "value3") // Should evict expired key1

		// key1 should be evicted (expired)
		assert.Nil(t, cached.getFromCache("key1"))
	})
}

// TestCachedNavigatorStartCleanupRoutine tests the background cleanup routine
func TestCachedNavigatorStartCleanupRoutine(t *testing.T) {
	base := NewInMemoryNavigator()
	setupTestTree(base)
	cached := NewCachedNavigator(base, 10*time.Millisecond)

	// Add some entries
	cached.putInCache("test1", "value1")
	cached.putInCache("test2", "value2")

	// Start cleanup routine with short interval
	cached.StartCleanupRoutine(20 * time.Millisecond)

	// Wait for TTL to expire and cleanup to run
	time.Sleep(50 * time.Millisecond)

	// Entries should be cleaned up
	assert.Nil(t, cached.getFromCache("test1"))
	assert.Nil(t, cached.getFromCache("test2"))
}

// TestCachedNavigatorGetCacheStats tests cache statistics
func TestCachedNavigatorGetCacheStats(t *testing.T) {
	base := NewInMemoryNavigator()
	setupTestTree(base)
	cached := NewCachedNavigator(base, 1*time.Hour)
	cached.WithMaxSize(100)

	// Add entries
	cached.putInCache("key1", "value1")
	cached.putInCache("key2", "value2")

	stats := cached.GetCacheStats()
	assert.Equal(t, 2, stats.Size)
	assert.Equal(t, 100, stats.MaxSize)
}

// TestFactoryCreateWithCache tests cached navigator creation via factory
func TestFactoryCreateWithCache(t *testing.T) {
	workspace := t.TempDir()
	cfg := config.DefaultConfigTree("test")
	factory := NewFactory(workspace, cfg, nil)

	nav, err := factory.CreateWithCache(30 * time.Second)
	require.NoError(t, err)
	assert.NotNil(t, nav)

	// Verify it's a cached navigator
	_, isCached := nav.(*CachedNavigator)
	assert.True(t, isCached)
}

// TestFactoryCreateFromConfig tests navigator creation from config
func TestFactoryCreateFromConfig(t *testing.T) {
	workspace := t.TempDir()
	cfg := config.DefaultConfigTree("test")
	factory := NewFactory(workspace, cfg, nil)

	nav, err := factory.CreateFromConfig()
	require.NoError(t, err)
	assert.NotNil(t, nav)
}

// TestFactoryCreateCachedNavigator tests the private createCachedNavigator method
func TestFactoryCreateCachedNavigator(t *testing.T) {
	workspace := t.TempDir()
	cfg := config.DefaultConfigTree("test")
	factory := NewFactory(workspace, cfg, nil)

	opts := &NavigatorOptions{
		WorkspacePath:   workspace,
		CacheEnabled:    true,
		CacheTTL:        1 * time.Minute,
		MaxCacheSize:    50,
		RefreshInterval: 30 * time.Second,
	}

	nav, err := factory.Create(TypeCached, opts)
	require.NoError(t, err)
	assert.NotNil(t, nav)

	// Verify it's a cached navigator
	_, isCached := nav.(*CachedNavigator)
	assert.True(t, isCached)
}

// TestFactoryCreateUnknownType tests error handling for unknown navigator type
func TestFactoryCreateUnknownType(t *testing.T) {
	workspace := t.TempDir()
	factory := NewFactory(workspace, nil, nil)

	_, err := factory.Create(NavigatorType("unknown"), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown navigator type")
}

// TestInMemoryNavigatorRefreshStatus tests RefreshStatus with config sync
func TestInMemoryNavigatorRefreshStatus(t *testing.T) {
	t.Run("refresh_status_syncs_with_config", func(t *testing.T) {
		nav := NewInMemoryNavigator()

		// Set up initial config with one node
		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: "nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "repo1", URL: "https://github.com/test/repo1.git"},
			},
		}
		nav.SetConfig(cfg)

		// Add initial node
		err := nav.AddNode("/repo1", &Node{
			Path: "/repo1",
			Name: "repo1",
			Type: NodeTypeRepo,
		})
		require.NoError(t, err)

		// Refresh status at root
		err = nav.RefreshStatus("/")
		assert.NoError(t, err)

		// Node should still exist
		node, err := nav.GetNode("/repo1")
		assert.NoError(t, err)
		assert.NotNil(t, node)
	})

	t.Run("refresh_status_removes_nodes_not_in_config", func(t *testing.T) {
		nav := NewInMemoryNavigator()

		// Add nodes first
		err := nav.AddNode("/repo1", &Node{
			Path: "/repo1",
			Name: "repo1",
			Type: NodeTypeRepo,
		})
		require.NoError(t, err)
		err = nav.AddNode("/repo2", &Node{
			Path: "/repo2",
			Name: "repo2",
			Type: NodeTypeRepo,
		})
		require.NoError(t, err)

		// Set config with only repo1
		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: "nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "repo1", URL: "https://github.com/test/repo1.git"},
			},
		}
		nav.SetConfig(cfg)

		// Refresh status at root - should sync with config
		err = nav.RefreshStatus("/")
		assert.NoError(t, err)

		// repo1 should exist
		node1, _ := nav.GetNode("/repo1")
		assert.NotNil(t, node1)

		// repo2 should be removed (not in config)
		node2, _ := nav.GetNode("/repo2")
		assert.Nil(t, node2)
	})

	t.Run("refresh_status_adds_missing_nodes_from_config", func(t *testing.T) {
		nav := NewInMemoryNavigator()

		// Set config with two nodes (use Fetch: "lazy" instead of Lazy field)
		cfg := &config.ConfigTree{
			Workspace: config.WorkspaceTree{
				Name:     "test",
				ReposDir: "nodes",
			},
			Nodes: []config.NodeDefinition{
				{Name: "repo1", URL: "https://github.com/test/repo1.git"},
				{Name: "repo2", URL: "https://github.com/test/repo2.git", Fetch: "lazy"},
			},
		}
		nav.SetConfig(cfg)

		// Refresh status at root - should add nodes from config
		err := nav.RefreshStatus("/")
		assert.NoError(t, err)

		// Both nodes should exist
		node1, _ := nav.GetNode("/repo1")
		assert.NotNil(t, node1)

		node2, _ := nav.GetNode("/repo2")
		assert.NotNil(t, node2)

		// repo2 should be lazy
		status, _ := nav.GetNodeStatus("/repo2")
		assert.True(t, status.Lazy)
	})

	t.Run("refresh_status_updates_last_check", func(t *testing.T) {
		nav := NewInMemoryNavigator()
		setupTestTree(nav)

		// Set status for a node
		nav.SetNodeStatus("/backend", &NodeStatus{
			Exists:    true,
			LastCheck: time.Now().Add(-1 * time.Hour),
		})

		// Refresh status
		err := nav.RefreshStatus("/backend")
		assert.NoError(t, err)

		// LastCheck should be updated
		status, err := nav.GetNodeStatus("/backend")
		assert.NoError(t, err)
		assert.True(t, status.LastCheck.After(time.Now().Add(-1*time.Minute)))
	})
}

// TestCachedNavigatorListChildren tests cached list children including error path
func TestCachedNavigatorListChildren(t *testing.T) {
	t.Run("list_children_caches_results", func(t *testing.T) {
		base := NewInMemoryNavigator()
		setupTestTree(base)
		cached := NewCachedNavigator(base, 1*time.Hour)

		// First call
		children1, err := cached.ListChildren("/")
		require.NoError(t, err)

		// Second call should use cache
		children2, err := cached.ListChildren("/")
		require.NoError(t, err)

		assert.Equal(t, len(children1), len(children2))
	})

	t.Run("list_children_for_non_existent_path", func(t *testing.T) {
		base := NewInMemoryNavigator()
		cached := NewCachedNavigator(base, 1*time.Hour)

		children, err := cached.ListChildren("/non-existent")
		assert.Error(t, err) // Non-existent path should return error
		assert.Nil(t, children)
	})
}

// TestCachedNavigatorIsLazy tests cached IsLazy including non-existent path
func TestCachedNavigatorIsLazy(t *testing.T) {
	t.Run("is_lazy_for_non_existent_path", func(t *testing.T) {
		base := NewInMemoryNavigator()
		cached := NewCachedNavigator(base, 1*time.Hour)

		isLazy, err := cached.IsLazy("/non-existent")
		assert.NoError(t, err)
		assert.False(t, isLazy)
	})
}

// TestNewCachedFactory tests the NewCached factory function
func TestNewCachedFactory(t *testing.T) {
	workspace := t.TempDir()
	cfg := config.DefaultConfigTree("test")

	nav, err := NewCached(workspace, cfg, nil)
	require.NoError(t, err)
	assert.NotNil(t, nav)
}

// TestInMemoryNavigatorNavigateEdgeCases tests Navigate edge cases
func TestInMemoryNavigatorNavigateEdgeCases(t *testing.T) {
	t.Run("navigate_to_non_existent_path", func(t *testing.T) {
		nav := NewInMemoryNavigator()
		setupTestTree(nav)

		err := nav.Navigate("/non-existent")
		assert.Error(t, err)
	})

	t.Run("navigate_normalizes_path", func(t *testing.T) {
		nav := NewInMemoryNavigator()
		setupTestTree(nav)

		err := nav.Navigate("backend") // Without leading slash
		assert.NoError(t, err)

		path, _ := nav.GetCurrentPath()
		assert.Equal(t, "/backend", path)
	})
}

// TestInMemoryNavigatorNormalizePath tests path normalization
func TestInMemoryNavigatorNormalizePath(t *testing.T) {
	nav := NewInMemoryNavigator()

	tests := []struct {
		input    string
		expected string
	}{
		{"", "/"},
		{"/", "/"},
		{"backend", "/backend"},
		{"/backend", "/backend"},
		{"/backend/", "/backend"},
		{"//backend//", "/backend"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := nav.normalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilesystemNavigatorIsRepositoryAlreadyCloned tests repository detection
func TestFilesystemNavigatorIsRepositoryAlreadyCloned(t *testing.T) {
	workspace := t.TempDir()
	cfg := config.DefaultConfigTree("test")
	nav, err := NewFilesystemNavigator(workspace, cfg, nil)
	require.NoError(t, err)

	// Non-existent path
	assert.False(t, nav.isRepositoryAlreadyCloned("/non/existent/path"))

	// Existing path without .git
	assert.False(t, nav.isRepositoryAlreadyCloned(workspace))
}

// TestFilesystemNavigatorShouldFallbackToHTTPS tests HTTPS fallback logic
func TestFilesystemNavigatorShouldFallbackToHTTPS(t *testing.T) {
	workspace := t.TempDir()
	cfg := config.DefaultConfigTree("test")
	nav, err := NewFilesystemNavigator(workspace, cfg, nil)
	require.NoError(t, err)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil_error",
			err:      nil,
			expected: false,
		},
		{
			name:     "permission_denied",
			err:      errors.New("Permission denied (publickey)"),
			expected: true,
		},
		{
			name:     "host_key_verification",
			err:      errors.New("Host key verification failed"),
			expected: true,
		},
		{
			name:     "ssh_connect",
			err:      errors.New("ssh: connect to host github.com"),
			expected: true,
		},
		{
			name:     "connection_refused",
			err:      errors.New("Connection refused"),
			expected: true,
		},
		{
			name:     "no_route_to_host",
			err:      errors.New("No route to host"),
			expected: true,
		},
		{
			name:     "destination_exists",
			err:      errors.New("destination path already exists"),
			expected: false,
		},
		{
			name:     "repo_not_found",
			err:      errors.New("Repository not found"),
			expected: false,
		},
		{
			name:     "directory_not_empty",
			err:      errors.New("directory not empty"),
			expected: false,
		},
		{
			name:     "generic_error",
			err:      errors.New("some random error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nav.shouldFallbackToHTTPS(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
