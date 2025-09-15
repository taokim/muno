package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taokim/muno/internal/config"
)

// TestNodeTypeConstants tests node type constants
func TestNodeTypeConstants(t *testing.T) {
	// Test NodeType constants
	assert.Equal(t, NodeType("root"), NodeTypeRoot)
	assert.Equal(t, NodeType("repo"), NodeTypeRepo)
	assert.Equal(t, NodeType("config"), NodeTypeConfig)
	
	// Test RepoState constants
	assert.Equal(t, RepoState("missing"), RepoStateMissing)
	assert.Equal(t, RepoState("cloned"), RepoStateCloned)
	assert.Equal(t, RepoState("modified"), RepoStateModified)
}

// TestTreeNodeSimpleMethods tests simple TreeNode methods
func TestTreeNodeSimpleMethods(t *testing.T) {
	t.Run("IsRoot", func(t *testing.T) {
		node := &TreeNode{Type: NodeTypeRoot}
		assert.True(t, node.IsRoot())
		
		node.Type = NodeTypeRepo
		assert.False(t, node.IsRoot())
	})
	
	t.Run("IsLeaf", func(t *testing.T) {
		node := &TreeNode{Children: []string{"child1"}}
		assert.False(t, node.IsLeaf())
		
		node.Children = []string{}
		assert.True(t, node.IsLeaf())
		
		node.Children = nil
		assert.True(t, node.IsLeaf())
	})
	
	t.Run("NeedsClone", func(t *testing.T) {
		node := &TreeNode{Type: NodeTypeRepo, URL: "repo.git", Lazy: true, State: RepoStateMissing}
		assert.True(t, node.NeedsClone())
		
		node.State = RepoStateCloned
		assert.False(t, node.NeedsClone())
		
		node.Type = NodeTypeRoot
		node.State = RepoStateMissing
		assert.False(t, node.NeedsClone())
	})
	
	t.Run("HasLazyRepos", func(t *testing.T) {
		// Node itself is lazy repo with missing state
		node := &TreeNode{Type: NodeTypeRepo, URL: "repo.git", Lazy: true, State: RepoStateMissing}
		assert.True(t, node.HasLazyRepos())
		
		// Node is cloned
		node.State = RepoStateCloned
		assert.False(t, node.HasLazyRepos())
	})
}

// TestResolverFunctions tests resolver helper functions
func TestResolverFunctions(t *testing.T) {
	t.Run("IsMetaRepo", func(t *testing.T) {
		// Test various meta repo suffixes that match eager patterns
		assert.True(t, IsMetaRepo("my-monorepo"))      // Ends with -monorepo
		assert.True(t, IsMetaRepo("dev-workspace"))    // Ends with -workspace
		assert.False(t, IsMetaRepo("root-repo"))       // Not -root-repo suffix
		assert.True(t, IsMetaRepo("my-root-repo"))     // Ends with -root-repo
		assert.True(t, IsMetaRepo("MY-MONOREPO"))      // Case insensitive
		
		// These match the patterns
		assert.True(t, IsMetaRepo("team-muno"))        // Ends with -muno
		assert.False(t, IsMetaRepo("metarepo"))        // Not -metarepo suffix
		assert.False(t, IsMetaRepo("main-platform"))   // Platform not in eager patterns
		
		// Test non-meta repos
		assert.False(t, IsMetaRepo("my-service"))
		assert.False(t, IsMetaRepo("api-gateway"))
		assert.False(t, IsMetaRepo(""))
		assert.False(t, IsMetaRepo("monorepo-tool")) // Contains but not suffix
	})
	
	t.Run("GetNodeKind", func(t *testing.T) {
		// Test repo node
		node := &config.NodeDefinition{URL: "repo.git"}
		assert.Equal(t, NodeKindRepo, GetNodeKind(node))
		
		// Test config node
		node = &config.NodeDefinition{ConfigRef: "config.yaml"}
		assert.Equal(t, NodeKindConfigRef, GetNodeKind(node))
		
		// Test invalid node (both URL and config)
		node = &config.NodeDefinition{URL: "repo.git", ConfigRef: "config.yaml"}
		assert.Equal(t, NodeKindInvalid, GetNodeKind(node))
		
		// Test invalid node (neither)
		node = &config.NodeDefinition{}
		assert.Equal(t, NodeKindInvalid, GetNodeKind(node))
	})
	
	t.Run("ResolveConfigPath", func(t *testing.T) {
		// Test with relative config
		node := &config.NodeDefinition{Name: "test", ConfigRef: "../config/muno.yaml"}
		result := ResolveConfigPath("/workspace", node)
		// filepath.Join simplifies the path
		assert.Equal(t, "/workspace/config/muno.yaml", result)
		
		// Test with absolute config
		node = &config.NodeDefinition{Name: "test", ConfigRef: "/absolute/path/config.yaml"}
		result = ResolveConfigPath("/workspace", node)
		assert.Equal(t, "/absolute/path/config.yaml", result)
		
		// Test with no config
		node = &config.NodeDefinition{Name: "test"}
		result = ResolveConfigPath("/workspace", node)
		assert.Equal(t, "", result)
		
		// Test with simple filename
		node = &config.NodeDefinition{Name: "test", ConfigRef: "muno.yaml"}
		result = ResolveConfigPath("/workspace", node)
		assert.Equal(t, "/workspace/test/muno.yaml", result)
	})
}

// TestTreeStateStructureMethods tests TreeState methods
func TestTreeStateStructureMethods(t *testing.T) {
	state := &TreeState{
		CurrentPath: "/test/path",
		Nodes: map[string]*TreeNode{
			"test": {Name: "test"},
		},
	}
	
	assert.Equal(t, "/test/path", state.CurrentPath)
	assert.NotNil(t, state.Nodes)
	assert.Contains(t, state.Nodes, "test")
}

// TestTargetResolution tests target resolution structure
func TestTargetResolution(t *testing.T) {
	resolution := &TargetResolution{
		Path:   "/test/path",
		Node:   &TreeNode{Name: "test"},
		Source: SourceExplicit,
	}
	
	assert.Equal(t, "/test/path", resolution.Path)
	assert.NotNil(t, resolution.Node)
	assert.Equal(t, "test", resolution.Node.Name)
	assert.Equal(t, SourceExplicit, resolution.Source)
}