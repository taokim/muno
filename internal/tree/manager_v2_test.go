package tree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/tree/navigator"
)

// TestManagerV2Creation tests creating a new ManagerV2
func TestManagerV2Creation(t *testing.T) {
	// Create temp workspace
	workspace := t.TempDir()

	t.Run("DefaultCreation", func(t *testing.T) {
		mgr, err := NewManagerV2(workspace)
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		if mgr.GetWorkspacePath() != workspace {
			t.Errorf("Expected workspace %s, got %s", workspace, mgr.GetWorkspacePath())
		}

		// Should have a navigator
		path, err := mgr.GetCurrentPath()
		if err != nil {
			t.Fatalf("Failed to get current path: %v", err)
		}
		if path != "/" {
			t.Errorf("Expected initial path '/', got '%s'", path)
		}
	})

	t.Run("WithInMemoryNavigator", func(t *testing.T) {
		// Create with in-memory navigator for testing
		nav := navigator.NewInMemory()
		mgr, err := NewManagerV2(workspace, WithNavigator(nav))
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// Should use the provided navigator
		path, err := mgr.GetCurrentPath()
		if err != nil {
			t.Fatalf("Failed to get current path: %v", err)
		}
		if path != "/" {
			t.Errorf("Expected initial path '/', got '%s'", path)
		}
	})

	t.Run("WithConfig", func(t *testing.T) {
		// Create with custom config
		cfg := config.DefaultConfigTree("test")
		mgr, err := NewManagerV2(workspace, WithConfig(cfg))
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		if mgr.GetConfig().Workspace.Name != "test" {
			t.Errorf("Expected workspace name 'test', got '%s'", mgr.GetConfig().Workspace.Name)
		}
	})
}

// TestManagerV2Navigation tests navigation functionality
func TestManagerV2Navigation(t *testing.T) {
	workspace := t.TempDir()
	
	// Use in-memory navigator for predictable testing
	nav := navigator.NewInMemoryNavigator()
	
	// Add test nodes
	backend := &navigator.Node{
		Path:     "/backend",
		Name:     "backend",
		Type:     navigator.NodeTypeRepo,
		URL:      "https://github.com/example/backend.git",
		Children: []string{"services"},
	}
	nav.AddNode("/backend", backend)

	services := &navigator.Node{
		Path:     "/backend/services",
		Name:     "services",
		Type:     navigator.NodeTypeDirectory,
		Children: []string{"auth", "payment"},
	}
	nav.AddNode("/backend/services", services)

	auth := &navigator.Node{
		Path:     "/backend/services/auth",
		Name:     "auth",
		Type:     navigator.NodeTypeRepo,
		URL:      "https://github.com/example/auth.git",
		Children: []string{},
	}
	nav.AddNode("/backend/services/auth", auth)

	payment := &navigator.Node{
		Path:     "/backend/services/payment",
		Name:     "payment",
		Type:     navigator.NodeTypeRepo,
		URL:      "https://github.com/example/payment.git",
		Children: []string{},
	}
	nav.AddNode("/backend/services/payment", payment)

	mgr, err := NewManagerV2(workspace, WithNavigator(nav))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	t.Run("NavigateAbsolute", func(t *testing.T) {
		err := mgr.UseNode("/backend")
		if err != nil {
			t.Fatalf("Failed to navigate: %v", err)
		}

		path, err := mgr.GetCurrentPath()
		if err != nil {
			t.Fatalf("Failed to get current path: %v", err)
		}
		if path != "/backend" {
			t.Errorf("Expected path '/backend', got '%s'", path)
		}
	})

	t.Run("NavigateRelative", func(t *testing.T) {
		err := mgr.UseNode("services")
		if err != nil {
			t.Fatalf("Failed to navigate: %v", err)
		}

		path, err := mgr.GetCurrentPath()
		if err != nil {
			t.Fatalf("Failed to get current path: %v", err)
		}
		if path != "/backend/services" {
			t.Errorf("Expected path '/backend/services', got '%s'", path)
		}
	})

	t.Run("NavigateParent", func(t *testing.T) {
		err := mgr.UseNode("..")
		if err != nil {
			t.Fatalf("Failed to navigate: %v", err)
		}

		path, err := mgr.GetCurrentPath()
		if err != nil {
			t.Fatalf("Failed to get current path: %v", err)
		}
		if path != "/backend" {
			t.Errorf("Expected path '/backend', got '%s'", path)
		}
	})
}

// TestManagerV2TreeOperations tests tree query operations
func TestManagerV2TreeOperations(t *testing.T) {
	workspace := t.TempDir()
	nav := navigator.NewInMemoryNavigator()
	
	// Setup test tree
	setupTestTree(nav)

	mgr, err := NewManagerV2(workspace, WithNavigator(nav))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	t.Run("GetNode", func(t *testing.T) {
		node, err := mgr.GetNode("/backend")
		if err != nil {
			t.Fatalf("Failed to get node: %v", err)
		}
		if node == nil {
			t.Fatal("Node should exist")
		}
		if node.Name != "backend" {
			t.Errorf("Expected name 'backend', got '%s'", node.Name)
		}
	})

	t.Run("ListChildren", func(t *testing.T) {
		children, err := mgr.ListChildren("/")
		if err != nil {
			t.Fatalf("Failed to list children: %v", err)
		}
		if len(children) != 3 {
			t.Errorf("Expected 3 children, got %d", len(children))
		}
	})

	t.Run("GetTree", func(t *testing.T) {
		tree, err := mgr.GetTree("/", 2)
		if err != nil {
			t.Fatalf("Failed to get tree: %v", err)
		}
		if tree == nil {
			t.Fatal("Tree should not be nil")
		}
		if len(tree.Nodes) < 4 {
			t.Errorf("Expected at least 4 nodes, got %d", len(tree.Nodes))
		}
	})

	t.Run("GetStatus", func(t *testing.T) {
		status, err := mgr.GetStatus("/backend")
		if err != nil {
			t.Fatalf("Failed to get status: %v", err)
		}
		if status == nil {
			t.Fatal("Status should not be nil")
		}
	})
}

// TestManagerV2RepoOperations tests repository management
func TestManagerV2RepoOperations(t *testing.T) {
	workspace := t.TempDir()
	
	// Create a config file
	cfg := config.DefaultConfigTree("test")
	configPath := filepath.Join(workspace, "muno.yaml")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Use in-memory navigator with config
	nav := navigator.NewInMemoryNavigator()
	nav.SetConfig(cfg)
	mgr, err := NewManagerV2(workspace, WithNavigator(nav), WithConfig(cfg))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	t.Run("AddRepo", func(t *testing.T) {
		err := mgr.AddRepo("/", "test-repo", "https://github.com/example/test.git", true)
		if err != nil {
			t.Fatalf("Failed to add repo: %v", err)
		}

		// Check it was added
		node, err := mgr.GetNode("/test-repo")
		if err != nil {
			t.Fatalf("Failed to get node: %v", err)
		}
		if node == nil {
			t.Fatal("Node should exist after adding")
		}
		if node.URL != "https://github.com/example/test.git" {
			t.Errorf("Expected URL to match")
		}

		// Check config was updated
		cfg, err = config.LoadTree(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		if len(cfg.Nodes) != 1 {
			t.Errorf("Expected 1 node in config, got %d", len(cfg.Nodes))
		}
	})

	t.Run("RemoveNode", func(t *testing.T) {
		// First add a node
		nav.AddNode("/to-remove", &navigator.Node{
			Path:     "/to-remove",
			Name:     "to-remove",
			Type:     navigator.NodeTypeRepo,
			Children: []string{},
		})

		// Remove it
		err := mgr.RemoveNode("/to-remove")
		if err != nil {
			t.Fatalf("Failed to remove node: %v", err)
		}

		// Check it's gone
		node, err := mgr.GetNode("/to-remove")
		if err != nil {
			t.Fatalf("Failed to get node: %v", err)
		}
		if node != nil {
			t.Error("Node should not exist after removal")
		}
	})
}

// TestManagerV2Display tests display functionality
func TestManagerV2Display(t *testing.T) {
	workspace := t.TempDir()
	nav := navigator.NewInMemoryNavigator()
	setupTestTree(nav)

	mgr, err := NewManagerV2(workspace, WithNavigator(nav))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	t.Run("DisplayTree", func(t *testing.T) {
		// Redirect stdout for testing
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := mgr.DisplayTree("/", 2)
		if err != nil {
			t.Fatalf("Failed to display tree: %v", err)
		}

		w.Close()
		os.Stdout = old

		// Read output
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		// Check output contains expected elements
		if !strings.Contains(output, "backend") {
			t.Error("Output should contain 'backend'")
		}
	})

	t.Run("DisplayStatus", func(t *testing.T) {
		// Redirect stdout for testing
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := mgr.DisplayStatus("/", false)
		if err != nil {
			t.Fatalf("Failed to display status: %v", err)
		}

		w.Close()
		os.Stdout = old

		// Read output
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		// Check output contains expected elements
		if !strings.Contains(output, "Repository Status") {
			t.Error("Output should contain 'Repository Status'")
		}
	})
}

// Helper function to setup test tree
func setupTestTree(nav *navigator.InMemoryNavigator) {
	nodes := []struct {
		path string
		name string
		typ  navigator.NodeType
		url  string
		children []string
	}{
		{"/backend", "backend", navigator.NodeTypeRepo, "https://github.com/example/backend.git", []string{"services"}},
		{"/backend/services", "services", navigator.NodeTypeDirectory, "", []string{"auth", "payment"}},
		{"/backend/services/auth", "auth", navigator.NodeTypeRepo, "https://github.com/example/auth.git", []string{}},
		{"/backend/services/payment", "payment", navigator.NodeTypeRepo, "https://github.com/example/payment.git", []string{}},
		{"/frontend", "frontend", navigator.NodeTypeRepo, "https://github.com/example/frontend.git", []string{"web", "mobile"}},
		{"/frontend/web", "web", navigator.NodeTypeRepo, "https://github.com/example/web.git", []string{}},
		{"/frontend/mobile", "mobile", navigator.NodeTypeRepo, "https://github.com/example/mobile.git", []string{}},
		{"/infrastructure", "infrastructure", navigator.NodeTypeFile, "", []string{}},
	}

	for _, n := range nodes {
		node := &navigator.Node{
			Path:      n.path,
			Name:      n.name,
			Type:      n.typ,
			URL:       n.url,
			Children:  n.children,
		}
		nav.AddNode(n.path, node)

		// Set status
		status := &navigator.NodeStatus{
			Exists:    true,
			Cloned:    n.typ != navigator.NodeTypeRepo || n.name != "payment", // payment is not cloned
			State:     navigator.RepoStateCloned,
			Modified:  n.name == "auth", // auth is modified
			Lazy:      n.name == "payment", // payment is lazy
			LastCheck: time.Now(),
		}
		nav.SetNodeStatus(n.path, status)
	}
}