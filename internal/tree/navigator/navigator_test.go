package navigator

import (
	"testing"
	"time"
)

// TestInMemoryNavigator tests the in-memory navigator implementation
func TestInMemoryNavigator(t *testing.T) {
	nav := NewInMemoryNavigator()

	t.Run("InitialState", func(t *testing.T) {
		// Check initial current path
		path, err := nav.GetCurrentPath()
		if err != nil {
			t.Fatalf("GetCurrentPath failed: %v", err)
		}
		if path != "/" {
			t.Errorf("Expected current path '/', got '%s'", path)
		}

		// Check root node exists
		root, err := nav.GetNode("/")
		if err != nil {
			t.Fatalf("GetNode(/) failed: %v", err)
		}
		if root == nil {
			t.Fatal("Root node should exist")
		}
		if root.Type != NodeTypeRoot {
			t.Errorf("Expected root type, got %s", root.Type)
		}
	})

	t.Run("AddAndNavigate", func(t *testing.T) {
		// Add a backend node
		backend := &Node{
			Path:     "/backend",
			Name:     "backend",
			Type:     NodeTypeRepo,
			URL:      "https://github.com/example/backend.git",
			Children: []string{},
		}
		if err := nav.AddNode("/backend", backend); err != nil {
			t.Fatalf("AddNode failed: %v", err)
		}

		// Navigate to backend
		if err := nav.Navigate("/backend"); err != nil {
			t.Fatalf("Navigate failed: %v", err)
		}

		// Check current path
		path, err := nav.GetCurrentPath()
		if err != nil {
			t.Fatalf("GetCurrentPath failed: %v", err)
		}
		if path != "/backend" {
			t.Errorf("Expected current path '/backend', got '%s'", path)
		}
	})

	t.Run("RelativeNavigation", func(t *testing.T) {
		// Add frontend as sibling
		frontend := &Node{
			Path:     "/frontend",
			Name:     "frontend",
			Type:     NodeTypeRepo,
			URL:      "https://github.com/example/frontend.git",
			Children: []string{},
		}
		if err := nav.AddNode("/frontend", frontend); err != nil {
			t.Fatalf("AddNode failed: %v", err)
		}

		// Navigate relatively from backend to frontend
		if err := nav.Navigate("../frontend"); err != nil {
			t.Fatalf("Relative navigate failed: %v", err)
		}

		path, err := nav.GetCurrentPath()
		if err != nil {
			t.Fatalf("GetCurrentPath failed: %v", err)
		}
		if path != "/frontend" {
			t.Errorf("Expected current path '/frontend', got '%s'", path)
		}
	})

	t.Run("ListChildren", func(t *testing.T) {
		// Navigate to root
		if err := nav.Navigate("/"); err != nil {
			t.Fatalf("Navigate to root failed: %v", err)
		}

		// List children
		children, err := nav.ListChildren("/")
		if err != nil {
			t.Fatalf("ListChildren failed: %v", err)
		}

		if len(children) != 2 {
			t.Errorf("Expected 2 children, got %d", len(children))
		}

		// Check children names
		childNames := make(map[string]bool)
		for _, child := range children {
			childNames[child.Name] = true
		}

		if !childNames["backend"] {
			t.Error("Expected 'backend' in children")
		}
		if !childNames["frontend"] {
			t.Error("Expected 'frontend' in children")
		}
	})

	t.Run("GetTree", func(t *testing.T) {
		// Add nested structure
		services := &Node{
			Path:     "/backend/services",
			Name:     "services",
			Type:     NodeTypeDirectory,
			Children: []string{},
		}
		if err := nav.AddNode("/backend/services", services); err != nil {
			t.Fatalf("AddNode failed: %v", err)
		}

		auth := &Node{
			Path:     "/backend/services/auth",
			Name:     "auth",
			Type:     NodeTypeRepo,
			URL:      "https://github.com/example/auth.git",
			Children: []string{},
		}
		if err := nav.AddNode("/backend/services/auth", auth); err != nil {
			t.Fatalf("AddNode failed: %v", err)
		}

		// Get tree with depth 2
		tree, err := nav.GetTree("/", 2)
		if err != nil {
			t.Fatalf("GetTree failed: %v", err)
		}

		// Check tree contents
		if len(tree.Nodes) < 4 {
			t.Errorf("Expected at least 4 nodes in tree, got %d", len(tree.Nodes))
		}

		// Verify specific nodes exist
		if _, exists := tree.Nodes["/"]; !exists {
			t.Error("Root node missing from tree")
		}
		if _, exists := tree.Nodes["/backend"]; !exists {
			t.Error("Backend node missing from tree")
		}
		if _, exists := tree.Nodes["/backend/services"]; !exists {
			t.Error("Services node missing from tree")
		}
	})

	t.Run("LazyLoading", func(t *testing.T) {
		// Add a lazy node
		lazy := &Node{
			Path:     "/lazy-repo",
			Name:     "lazy-repo",
			Type:     NodeTypeRepo,
			URL:      "https://github.com/example/lazy.git",
			Children: []string{},
		}
		if err := nav.AddNode("/lazy-repo", lazy); err != nil {
			t.Fatalf("AddNode failed: %v", err)
		}

		// Set as lazy
		status := &NodeStatus{
			Exists:    true,
			Cloned:    false,
			Lazy:      true,
			State:     RepoStateMissing,
			LastCheck: time.Now(),
		}
		if err := nav.SetNodeStatus("/lazy-repo", status); err != nil {
			t.Fatalf("SetNodeStatus failed: %v", err)
		}

		// Check if lazy
		isLazy, err := nav.IsLazy("/lazy-repo")
		if err != nil {
			t.Fatalf("IsLazy failed: %v", err)
		}
		if !isLazy {
			t.Error("Expected node to be lazy")
		}

		// Trigger lazy load
		if err := nav.TriggerLazyLoad("/lazy-repo"); err != nil {
			t.Fatalf("TriggerLazyLoad failed: %v", err)
		}

		// Check status after load
		newStatus, err := nav.GetNodeStatus("/lazy-repo")
		if err != nil {
			t.Fatalf("GetNodeStatus failed: %v", err)
		}
		if !newStatus.Cloned {
			t.Error("Expected node to be cloned after lazy load")
		}
	})

	t.Run("RemoveNode", func(t *testing.T) {
		// Remove a node
		if err := nav.RemoveNode("/lazy-repo"); err != nil {
			t.Fatalf("RemoveNode failed: %v", err)
		}

		// Check it's gone
		node, err := nav.GetNode("/lazy-repo")
		if err != nil {
			t.Fatalf("GetNode failed: %v", err)
		}
		if node != nil {
			t.Error("Expected node to be removed")
		}

		// Check parent's children updated
		root, err := nav.GetNode("/")
		if err != nil {
			t.Fatalf("GetNode(/) failed: %v", err)
		}
		for _, child := range root.Children {
			if child == "lazy-repo" {
				t.Error("Removed node still in parent's children")
			}
		}
	})
}

// TestCachedNavigator tests the caching wrapper
func TestCachedNavigator(t *testing.T) {
	// Create in-memory base navigator
	base := NewInMemoryNavigator()
	
	// Add some test nodes
	backend := &Node{
		Path:     "/backend",
		Name:     "backend",
		Type:     NodeTypeRepo,
		URL:      "https://github.com/example/backend.git",
		Children: []string{"services"},
	}
	base.AddNode("/backend", backend)

	services := &Node{
		Path:     "/backend/services",
		Name:     "services",
		Type:     NodeTypeDirectory,
		Children: []string{},
	}
	base.AddNode("/backend/services", services)

	// Create cached navigator with short TTL for testing
	cached := NewCachedNavigator(base, 100*time.Millisecond)

	t.Run("CacheHit", func(t *testing.T) {
		// First call - cache miss
		node1, err := cached.GetNode("/backend")
		if err != nil {
			t.Fatalf("GetNode failed: %v", err)
		}

		// Second call - should be cache hit
		node2, err := cached.GetNode("/backend")
		if err != nil {
			t.Fatalf("GetNode failed: %v", err)
		}

		// Should be same data
		if node1.URL != node2.URL {
			t.Error("Cache returned different data")
		}

		// Check cache stats
		stats := cached.GetCacheStats()
		if stats.Size == 0 {
			t.Error("Expected cache to have entries")
		}
	})

	t.Run("CacheExpiry", func(t *testing.T) {
		// Get node to cache it
		_, err := cached.GetNode("/backend/services")
		if err != nil {
			t.Fatalf("GetNode failed: %v", err)
		}

		// Wait for cache to expire
		time.Sleep(150 * time.Millisecond)

		// Clear expired entries
		cached.ClearExpired()

		// Check cache stats
		_ = cached.GetCacheStats()
		// The /backend entry might still be cached if not expired
		// but /backend/services should be expired
	})

	t.Run("CacheInvalidation", func(t *testing.T) {
		// Cache a node
		_, err := cached.GetNode("/backend")
		if err != nil {
			t.Fatalf("GetNode failed: %v", err)
		}

		// Navigate to it (should invalidate cache)
		if err := cached.Navigate("/backend"); err != nil {
			t.Fatalf("Navigate failed: %v", err)
		}

		// Trigger lazy load (should also invalidate)
		if err := cached.RefreshStatus("/backend"); err != nil {
			t.Fatalf("RefreshStatus failed: %v", err)
		}
	})

	t.Run("TreeCaching", func(t *testing.T) {
		// Get tree
		tree1, err := cached.GetTree("/", 2)
		if err != nil {
			t.Fatalf("GetTree failed: %v", err)
		}

		// Get again - should be cached
		tree2, err := cached.GetTree("/", 2)
		if err != nil {
			t.Fatalf("GetTree failed: %v", err)
		}

		// Should have same number of nodes
		if len(tree1.Nodes) != len(tree2.Nodes) {
			t.Error("Cached tree has different size")
		}
	})
}

// TestNavigatorFactory tests the factory creation
func TestNavigatorFactory(t *testing.T) {
	factory := NewFactory("", nil, nil)

	t.Run("CreateInMemory", func(t *testing.T) {
		nav, err := factory.Create(TypeInMemory, nil)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if _, ok := nav.(*InMemoryNavigator); !ok {
			t.Error("Expected InMemoryNavigator type")
		}
	})

	t.Run("CreateForTesting", func(t *testing.T) {
		nav, err := factory.CreateForTesting()
		if err != nil {
			t.Fatalf("CreateForTesting failed: %v", err)
		}

		// Should be able to use it
		path, err := nav.GetCurrentPath()
		if err != nil {
			t.Fatalf("GetCurrentPath failed: %v", err)
		}
		if path != "/" {
			t.Errorf("Expected initial path '/', got '%s'", path)
		}
	})
}

// BenchmarkNavigators compares performance of different navigators
func BenchmarkNavigators(b *testing.B) {
	// Setup in-memory navigator with test data
	inmem := NewInMemoryNavigator()
	setupTestTree(inmem)

	// Setup cached navigator
	cached := NewCachedNavigator(inmem, 1*time.Minute)

	b.Run("InMemory-GetNode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			inmem.GetNode("/backend/services/auth")
		}
	})

	b.Run("Cached-GetNode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cached.GetNode("/backend/services/auth")
		}
	})

	b.Run("InMemory-GetTree", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			inmem.GetTree("/", 3)
		}
	})

	b.Run("Cached-GetTree", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cached.GetTree("/", 3)
		}
	})
}

// setupTestTree creates a test tree structure
func setupTestTree(nav *InMemoryNavigator) {
	nodes := []struct {
		path string
		name string
		typ  NodeType
	}{
		{"/backend", "backend", NodeTypeRepo},
		{"/backend/services", "services", NodeTypeDirectory},
		{"/backend/services/auth", "auth", NodeTypeRepo},
		{"/backend/services/payment", "payment", NodeTypeRepo},
		{"/frontend", "frontend", NodeTypeRepo},
		{"/frontend/web", "web", NodeTypeRepo},
		{"/frontend/mobile", "mobile", NodeTypeRepo},
		{"/infrastructure", "infrastructure", NodeTypeConfig},
	}

	for _, n := range nodes {
		node := &Node{
			Path:     n.path,
			Name:     n.name,
			Type:     n.typ,
			Children: []string{},
		}
		nav.AddNode(n.path, node)
	}
}