package manager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/constants"
)

func TestTestableHelpers_FindWorkspaceRoot(t *testing.T) {
	h := &TestableHelpers{}
	
	t.Run("find from subdirectory", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Create workspace structure
		wsDir := filepath.Join(tmpDir, "workspace")
		subDir := filepath.Join(wsDir, "sub", "dir")
		err := os.MkdirAll(subDir, 0755)
		require.NoError(t, err)
		
		// Create muno.yaml at workspace root
		configPath := filepath.Join(wsDir, "muno.yaml")
		err = os.WriteFile(configPath, []byte("workspace:\n  name: test\n"), 0644)
		require.NoError(t, err)
		
		// Test finding from subdirectory
		root := h.FindWorkspaceRoot(subDir)
		assert.Equal(t, wsDir, root)
	})
	
	t.Run("no workspace found", func(t *testing.T) {
		tmpDir := t.TempDir()
		root := h.FindWorkspaceRoot(tmpDir)
		assert.Empty(t, root)
	})
	
	t.Run("from current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldCwd, _ := os.Getwd()
		defer os.Chdir(oldCwd)
		
		err := os.Chdir(tmpDir)
		require.NoError(t, err)
		
		// Create muno.yaml in current dir
		err = os.WriteFile("muno.yaml", []byte("workspace:\n  name: test\n"), 0644)
		require.NoError(t, err)
		
		root := h.FindWorkspaceRoot("")
		// Compare using EvalSymlinks to handle /var vs /private/var on macOS
		expectedPath, _ := filepath.EvalSymlinks(tmpDir)
		actualPath, _ := filepath.EvalSymlinks(root)
		assert.Equal(t, expectedPath, actualPath)
	})
}

func TestTestableHelpers_EnsureGitignoreEntry(t *testing.T) {
	h := &TestableHelpers{}
	
	t.Run("add to new file", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		err := h.EnsureGitignoreEntry(tmpDir, ".muno-state.json")
		assert.NoError(t, err)
		
		// Check file was created
		content, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
		assert.NoError(t, err)
		assert.Equal(t, ".muno-state.json\n", string(content))
	})
	
	t.Run("add to existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		
		// Create existing .gitignore
		err := os.WriteFile(gitignorePath, []byte("node_modules/\n*.log\n"), 0644)
		require.NoError(t, err)
		
		err = h.EnsureGitignoreEntry(tmpDir, ".muno-state.json")
		assert.NoError(t, err)
		
		// Check entry was added
		content, err := os.ReadFile(gitignorePath)
		assert.NoError(t, err)
		assert.Contains(t, string(content), "node_modules/")
		assert.Contains(t, string(content), ".muno-state.json")
	})
	
	t.Run("entry already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		
		// Create .gitignore with entry
		err := os.WriteFile(gitignorePath, []byte(".muno-state.json\nnode_modules/\n"), 0644)
		require.NoError(t, err)
		
		err = h.EnsureGitignoreEntry(tmpDir, ".muno-state.json")
		assert.NoError(t, err)
		
		// Check file wasn't modified
		content, err := os.ReadFile(gitignorePath)
		assert.NoError(t, err)
		assert.Equal(t, ".muno-state.json\nnode_modules/\n", string(content))
	})
	
	t.Run("entry exists with slash", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		
		// Create .gitignore with /entry
		err := os.WriteFile(gitignorePath, []byte("/.muno-state.json\n"), 0644)
		require.NoError(t, err)
		
		err = h.EnsureGitignoreEntry(tmpDir, ".muno-state.json")
		assert.NoError(t, err)
		
		// Should not add duplicate
		content, err := os.ReadFile(gitignorePath)
		assert.NoError(t, err)
		assert.Equal(t, "/.muno-state.json\n", string(content))
	})
}

func TestTestableHelpers_GenerateTreeContext(t *testing.T) {
	h := &TestableHelpers{}
	
	t.Run("with current path", func(t *testing.T) {
		content := h.GenerateTreeContext("test-workspace", "/current/path")
		
		assert.Contains(t, content, "Repository Structure")
		assert.Contains(t, content, "Workspace: test-workspace")
		assert.Contains(t, content, "Current location: /current/path")
		assert.Contains(t, content, "muno tree")
		assert.Contains(t, content, "muno use")
	})
	
	t.Run("without current path", func(t *testing.T) {
		content := h.GenerateTreeContext("workspace", "")
		
		assert.Contains(t, content, "Workspace: workspace")
		assert.NotContains(t, content, "Current location:")
	})
}

func TestTestableHelpers_CreateAgentContextFile(t *testing.T) {
	h := &TestableHelpers{}
	
	content := "# Test Context\n\nThis is a test."
	
	file, err := h.CreateAgentContextFile(content)
	assert.NoError(t, err)
	assert.Contains(t, file, "muno-context")
	
	// Check file was created
	data, err := os.ReadFile(file)
	assert.NoError(t, err)
	assert.Equal(t, content, string(data))
	
	// Clean up
	os.Remove(file)
}

func TestTestableHelpers_ComputeNodePath(t *testing.T) {
	h := &TestableHelpers{}
	
	tests := []struct {
		name      string
		workspace string
		nodePath  string
		expected  string
	}{
		{
			name:      "root path",
			workspace: "/workspace",
			nodePath:  "/",
			expected:  filepath.Join("/workspace", constants.DefaultReposDir),
		},
		{
			name:      "empty path",
			workspace: "/workspace",
			nodePath:  "",
			expected:  filepath.Join("/workspace", constants.DefaultReposDir),
		},
		{
			name:      "nested path",
			workspace: "/workspace",
			nodePath:  "/parent/child",
			expected:  filepath.Join("/workspace", constants.DefaultReposDir, "parent/child"),
		},
		{
			name:      "path without leading slash",
			workspace: "/workspace",
			nodePath:  "parent/child",
			expected:  filepath.Join("/workspace", constants.DefaultReposDir, "parent/child"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.ComputeNodePath(tt.workspace, tt.nodePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTestableHelpers_IsGitRepository(t *testing.T) {
	h := &TestableHelpers{}
	
	t.Run("is git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.MkdirAll(gitDir, 0755)
		require.NoError(t, err)
		
		assert.True(t, h.IsGitRepository(tmpDir))
	})
	
	t.Run("not git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		assert.False(t, h.IsGitRepository(tmpDir))
	})
	
	t.Run(".git is file not dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitFile := filepath.Join(tmpDir, ".git")
		err := os.WriteFile(gitFile, []byte("gitdir: somewhere"), 0644)
		require.NoError(t, err)
		
		assert.False(t, h.IsGitRepository(tmpDir))
	})
}

func TestTestableHelpers_ExtractRepoName(t *testing.T) {
	h := &TestableHelpers{}
	
	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com/user/repo.git", "repo"},
		{"https://github.com/user/repo", "repo"},
		{"git@github.com:user/repo.git", "repo"},
		{"https://gitlab.com/group/subgroup/project.git", "project"},
		{"repo.git", "repo"},
		{"", ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := h.ExtractRepoName(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTestableHelpers_NormalizePath(t *testing.T) {
	h := &TestableHelpers{}
	
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"/", "/"},
		{"path", "/path"},
		{"/path", "/path"},
		{"/path/", "/path"},
		{"path/", "/path"},
		{"/parent/child/", "/parent/child"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.NormalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTestableHelpers_SplitPath(t *testing.T) {
	h := &TestableHelpers{}
	
	tests := []struct {
		input    string
		expected []string
	}{
		{"/", []string{}},
		{"", []string{}},
		{"/parent", []string{"parent"}},
		{"/parent/child", []string{"parent", "child"}},
		{"parent/child", []string{"parent", "child"}},
		{"/a/b/c/d", []string{"a", "b", "c", "d"}},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := h.SplitPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTestableHelpers_JoinPath(t *testing.T) {
	h := &TestableHelpers{}
	
	tests := []struct {
		segments []string
		expected string
	}{
		{[]string{}, "/"},
		{[]string{"parent"}, "/parent"},
		{[]string{"parent", "child"}, "/parent/child"},
		{[]string{"a", "b", "c"}, "/a/b/c"},
	}
	
	for _, tt := range tests {
		t.Run(strings.Join(tt.segments, ","), func(t *testing.T) {
			result := h.JoinPath(tt.segments...)
			assert.Equal(t, tt.expected, result)
		})
	}
}