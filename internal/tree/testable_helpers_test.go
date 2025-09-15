package tree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
	"github.com/taokim/muno/internal/tree/navigator"
)

func TestTreeHelpers_DisplayChildNodes(t *testing.T) {
	h := &TreeHelpers{}
	
	t.Run("nil node", func(t *testing.T) {
		result := h.DisplayChildNodes(nil)
		assert.Equal(t, "No children", result)
	})
	
	t.Run("no children", func(t *testing.T) {
		node := &navigator.Node{Name: "parent", Children: []string{}}
		result := h.DisplayChildNodes(node)
		assert.Equal(t, "No children", result)
	})
	
	t.Run("with children", func(t *testing.T) {
		node := &navigator.Node{
			Name: "parent",
			Children: []string{"child1", "child2", "child3"},
		}
		result := h.DisplayChildNodes(node)
		assert.Contains(t, result, "├── child1")
		assert.Contains(t, result, "├── child2")
		assert.Contains(t, result, "└── child3")
	})
}

func TestTreeHelpers_PrintNodeInfo(t *testing.T) {
	h := &TreeHelpers{}
	
	t.Run("nil node", func(t *testing.T) {
		result := h.PrintNodeInfo(nil)
		assert.Empty(t, result)
	})
	
	t.Run("basic node", func(t *testing.T) {
		node := &interfaces.NodeInfo{
			Name: "test-node",
			Path: "/test/path",
		}
		result := h.PrintNodeInfo(node)
		assert.Contains(t, result, "Node: test-node")
		assert.Contains(t, result, "Path: /test/path")
	})
	
	t.Run("lazy node", func(t *testing.T) {
		node := &interfaces.NodeInfo{
			Name:   "lazy-node",
			Path:   "/lazy",
			IsLazy: true,
		}
		result := h.PrintNodeInfo(node)
		assert.Contains(t, result, "Status: Lazy (not cloned)")
	})
	
	t.Run("cloned node with changes", func(t *testing.T) {
		node := &interfaces.NodeInfo{
			Name:       "repo",
			Path:       "/repo",
			Repository: "https://github.com/test/repo.git",
			IsCloned:   true,
			HasChanges: true,
		}
		result := h.PrintNodeInfo(node)
		assert.Contains(t, result, "Repository: https://github.com/test/repo.git")
		assert.Contains(t, result, "Status: Cloned")
		assert.Contains(t, result, "Has local changes")
	})
}

func TestTreeHelpers_PrintNodePath(t *testing.T) {
	h := &TreeHelpers{}
	
	tests := []struct {
		path     string
		expected string
	}{
		{"", "Current: / (root)"},
		{"/", "Current: / (root)"},
		{"/test", "Current: /test"},
		{"/parent/child", "Current: /parent/child"},
	}
	
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := h.PrintNodePath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTreeHelpers_PrintChildrenInfo(t *testing.T) {
	h := &TreeHelpers{}
	
	t.Run("no children", func(t *testing.T) {
		result := h.PrintChildrenInfo([]*interfaces.NodeInfo{})
		assert.Equal(t, "No children", result)
	})
	
	t.Run("with children", func(t *testing.T) {
		nodes := []*interfaces.NodeInfo{
			{Name: "child1", IsLazy: true},
			{Name: "child2", IsCloned: true},
			{Name: "child3"},
		}
		result := h.PrintChildrenInfo(nodes)
		assert.Contains(t, result, "Children (3):")
		assert.Contains(t, result, "- child1 [lazy]")
		assert.Contains(t, result, "- child2 [cloned]")
		assert.Contains(t, result, "- child3")
	})
}

func TestTreeHelpers_SaveTreeState(t *testing.T) {
	h := &TreeHelpers{}
	tmpDir := t.TempDir()
	
	t.Run("nil state", func(t *testing.T) {
		stateFile := filepath.Join(tmpDir, "state1.json")
		err := h.SaveTreeState(nil, stateFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state is nil")
	})
	
	t.Run("save state", func(t *testing.T) {
		state := &TreeState{
			CurrentPath: "/test",
		}
		stateFile := filepath.Join(tmpDir, "subdir", "state2.json")
		err := h.SaveTreeState(state, stateFile)
		assert.NoError(t, err)
		
		// Check file was created
		_, err = os.Stat(stateFile)
		assert.NoError(t, err)
	})
}

func TestTreeHelpers_LoadConfigReference(t *testing.T) {
	h := &TreeHelpers{}
	tmpDir := t.TempDir()
	
	t.Run("empty path", func(t *testing.T) {
		cfg, err := h.LoadConfigReference("")
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "config path is empty")
	})
	
	t.Run("non-existent file", func(t *testing.T) {
		cfg, err := h.LoadConfigReference("/non/existent/file.yaml")
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "config file not found")
	})
	
	t.Run("existing file", func(t *testing.T) {
		configFile := filepath.Join(tmpDir, "config.yaml")
		err := os.WriteFile(configFile, []byte("workspace:\n  name: test\n"), 0644)
		require.NoError(t, err)
		
		cfg, err := h.LoadConfigReference(configFile)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "test", cfg.Workspace.Name)
	})
}

func TestTreeHelpers_BuildNodeFromConfig(t *testing.T) {
	h := &TreeHelpers{}
	
	t.Run("root child", func(t *testing.T) {
		def := config.NodeDefinition{
			Name: "child",
			URL:  "https://github.com/test/repo.git",
		}
		node := h.BuildNodeFromConfig(def, "/")
		
		assert.Equal(t, "child", node.Name)
		assert.Equal(t, "/child", node.Path)
		assert.Equal(t, "https://github.com/test/repo.git", node.URL)
		assert.Equal(t, navigator.NodeTypeRepo, node.Type)
	})
	
	t.Run("nested child", func(t *testing.T) {
		def := config.NodeDefinition{
			Name:  "grandchild",
			Fetch: config.FetchLazy,
		}
		node := h.BuildNodeFromConfig(def, "/parent")
		
		assert.Equal(t, "grandchild", node.Name)
		assert.Equal(t, "/parent/grandchild", node.Path)
		// Check if it's marked as lazy in metadata
		assert.Equal(t, navigator.NodeTypeDirectory, node.Type) // No URL, so it's a directory
	})
	
	t.Run("config reference", func(t *testing.T) {
		def := config.NodeDefinition{
			Name:   "config-node",
			ConfigRef: "sub.yaml",
		}
		node := h.BuildNodeFromConfig(def, "/parent")
		
		assert.Equal(t, "config-node", node.Name)
		assert.Equal(t, "/parent/config-node", node.Path)
		assert.Equal(t, "sub.yaml", node.ConfigRef)
		assert.Equal(t, navigator.NodeTypeConfig, node.Type)
	})
}

func TestTreeHelpers_ComputeNodeStatus(t *testing.T) {
	h := &TreeHelpers{}
	
	t.Run("nil node", func(t *testing.T) {
		status := h.ComputeNodeStatus(nil, nil)
		assert.Nil(t, status)
	})
	
	t.Run("lazy node", func(t *testing.T) {
		node := &navigator.Node{
			Name: "lazy",
			Path: "/lazy",
			URL:  "https://github.com/test/repo.git",
			Type: navigator.NodeTypeRepo,
			Metadata: map[string]interface{}{"lazy": true},
		}
		status := h.ComputeNodeStatus(node, nil)
		
		assert.Equal(t, "/lazy", status.Path)
		assert.Equal(t, "lazy", status.Name)
		assert.True(t, status.IsLazy)
		assert.False(t, status.IsCloned)
	})
	
	t.Run("cloned node with git status", func(t *testing.T) {
		node := &navigator.Node{
			Name: "repo",
			Path: "/repo",
			URL:  "https://github.com/test/repo.git",
			Type: navigator.NodeTypeRepo,
		}
		gitStatus := &interfaces.GitStatus{
			HasChanges: true,
			Behind:     1,  // This indicates remote changes
			Branch:     "main",
		}
		status := h.ComputeNodeStatus(node, gitStatus)
		
		assert.False(t, status.IsLazy)
		assert.True(t, status.IsCloned)
		assert.True(t, status.HasLocalChanges)
		assert.True(t, status.HasRemoteChanges)
		assert.Equal(t, "main", status.Branch)
	})
}

func TestTreeHelpers_FormatTreeDisplay(t *testing.T) {
	h := &TreeHelpers{}
	
	t.Run("nil node", func(t *testing.T) {
		result := h.FormatTreeDisplay(nil, "", false)
		assert.Empty(t, result)
	})
	
	t.Run("single node", func(t *testing.T) {
		node := &navigator.Node{
			Name: "root",
			Path: "/",
		}
		result := h.FormatTreeDisplay(node, "", false)
		assert.Contains(t, result, "├── root [dir]")
	})
	
	t.Run("node with children", func(t *testing.T) {
		node := &navigator.Node{
			Name: "parent",
			Path: "/parent",
			Type: navigator.NodeTypeDirectory,
			Children: []string{"child1", "child2", "child3"},
		}
		result := h.FormatTreeDisplay(node, "", true)
		
		assert.Contains(t, result, "└── parent [dir]")
		// Note: We can't display children details as they're just strings, not full nodes
	})
	
	t.Run("deep tree", func(t *testing.T) {
		// Since navigator.Node has Children as []string, we can't create a deep tree
		// We'll just test the basic formatting
		node := &navigator.Node{
			Name:     "root",
			Type:     navigator.NodeTypeDirectory,
			Children: []string{"level1", "level2"},
		}
		result := h.FormatTreeDisplay(node, "", false)
		
		// Check that we got a result
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "root")
	})
}