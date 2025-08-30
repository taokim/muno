package tree

import (
	"testing"
	"time"
)

func TestTreeNodeMethods(t *testing.T) {
	t.Run("IsRoot", func(t *testing.T) {
		rootNode := &TreeNode{
			Name: "root",
			Type: NodeTypeRoot,
		}
		
		repoNode := &TreeNode{
			Name: "repo",
			Type: NodeTypeRepo,
		}
		
		if !rootNode.IsRoot() {
			t.Error("Root node should return true for IsRoot()")
		}
		
		if repoNode.IsRoot() {
			t.Error("Repo node should return false for IsRoot()")
		}
	})
	
	t.Run("IsLeaf", func(t *testing.T) {
		leafNode := &TreeNode{
			Name:     "leaf",
			Type:     NodeTypeRepo,
			Children: []string{},
		}
		
		parentNode := &TreeNode{
			Name:     "parent",
			Type:     NodeTypeRepo,
			Children: []string{"child1", "child2"},
		}
		
		if !leafNode.IsLeaf() {
			t.Error("Node with no children should be a leaf")
		}
		
		if parentNode.IsLeaf() {
			t.Error("Node with children should not be a leaf")
		}
	})
	
	t.Run("HasLazyRepos", func(t *testing.T) {
		lazyRepo := &TreeNode{
			Name:  "lazy",
			Type:  NodeTypeRepo,
			Lazy:  true,
			State: RepoStateMissing,
		}
		
		clonedLazyRepo := &TreeNode{
			Name:  "cloned-lazy",
			Type:  NodeTypeRepo,
			Lazy:  true,
			State: RepoStateCloned,
		}
		
		nonLazyRepo := &TreeNode{
			Name:  "non-lazy",
			Type:  NodeTypeRepo,
			Lazy:  false,
			State: RepoStateMissing,
		}
		
		rootNode := &TreeNode{
			Name: "root",
			Type: NodeTypeRoot,
		}
		
		if !lazyRepo.HasLazyRepos() {
			t.Error("Lazy repo with missing state should return true")
		}
		
		if clonedLazyRepo.HasLazyRepos() {
			t.Error("Cloned lazy repo should return false")
		}
		
		if nonLazyRepo.HasLazyRepos() {
			t.Error("Non-lazy repo should return false")
		}
		
		if rootNode.HasLazyRepos() {
			t.Error("Root node should return false")
		}
	})
	
	t.Run("NeedsClone", func(t *testing.T) {
		missingRepo := &TreeNode{
			Name:  "missing",
			Type:  NodeTypeRepo,
			State: RepoStateMissing,
		}
		
		clonedRepo := &TreeNode{
			Name:  "cloned",
			Type:  NodeTypeRepo,
			State: RepoStateCloned,
		}
		
		rootNode := &TreeNode{
			Name: "root",
			Type: NodeTypeRoot,
		}
		
		if !missingRepo.NeedsClone() {
			t.Error("Missing repo should need clone")
		}
		
		if clonedRepo.NeedsClone() {
			t.Error("Cloned repo should not need clone")
		}
		
		if rootNode.NeedsClone() {
			t.Error("Root node should not need clone")
		}
	})
}

func TestTreeStateStructure(t *testing.T) {
	// Test that TreeState can be properly created and used
	state := &TreeState{
		CurrentPath: "/test/path",
		Nodes: map[string]*TreeNode{
			"/": {
				Name:     "root",
				Type:     NodeTypeRoot,
				Children: []string{"child1", "child2"},
			},
			"/child1": {
				Name:     "child1",
				Type:     NodeTypeRepo,
				URL:      "https://github.com/test/child1.git",
				Lazy:     false,
				State:    RepoStateCloned,
				Children: []string{},
			},
		},
		LastUpdated: time.Now(),
	}
	
	if state.CurrentPath != "/test/path" {
		t.Errorf("CurrentPath = %s, want /test/path", state.CurrentPath)
	}
	
	if len(state.Nodes) != 2 {
		t.Errorf("Nodes count = %d, want 2", len(state.Nodes))
	}
	
	root := state.Nodes["/"]
	if root == nil {
		t.Fatal("Root node not found")
	}
	
	if !root.IsRoot() {
		t.Error("Root node should be identified as root")
	}
	
	if len(root.Children) != 2 {
		t.Errorf("Root children = %d, want 2", len(root.Children))
	}
}