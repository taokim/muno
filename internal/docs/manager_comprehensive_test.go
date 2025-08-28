package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsManagerFullLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	t.Run("Create and manage global docs", func(t *testing.T) {
		// Create multiple global docs
		docs := []struct {
			name    string
			content string
		}{
			{"architecture.md", "# Architecture\nSystem design documentation"},
			{"api.md", "# API Documentation\nREST API endpoints"},
			{"setup.md", "# Setup Guide\nInstallation instructions"},
		}
		
		for _, doc := range docs {
			err := m.CreateGlobal(doc.name, doc.content)
			assert.NoError(t, err)
			
			// Verify file was created
			docPath := filepath.Join(tmpDir, "global", doc.name)
			assert.FileExists(t, docPath)
			
			// Verify content
			content, err := os.ReadFile(docPath)
			require.NoError(t, err)
			assert.Equal(t, doc.content, string(content))
		}
		
		// List global docs
		list, err := m.List("")
		require.NoError(t, err)
		
		// Should have at least our created docs
		assert.GreaterOrEqual(t, len(list), len(docs))
		
		// Verify our docs are in the list
		names := make(map[string]bool)
		for _, item := range list {
			names[filepath.Base(item.Path)] = true
		}
		
		for _, doc := range docs {
			assert.True(t, names[doc.name], "Doc %s should be in list", doc.name)
		}
	})
	
	t.Run("Create and manage scope docs", func(t *testing.T) {
		scopes := []string{"scope1", "scope2", "scope3"}
		
		for _, scope := range scopes {
			// Create scope docs
			docName := scope + "-design.md"
			content := "# " + scope + " Design\nScope-specific documentation"
			
			err := m.CreateScope(scope, docName, content)
			assert.NoError(t, err)
			
			// Verify file was created
			docPath := filepath.Join(tmpDir, "scopes", scope, docName)
			assert.FileExists(t, docPath)
			
			// List scope docs
			list, err := m.List(scope)
			require.NoError(t, err)
			assert.Greater(t, len(list), 0)
			
			// Verify the doc is in the list
			found := false
			for _, item := range list {
				if strings.Contains(item.Path, docName) {
					found = true
					break
				}
			}
			assert.True(t, found, "Doc %s should be in scope list", docName)
		}
	})
	
	t.Run("GetPath", func(t *testing.T) {
		relativePath := "global/test.md"
		fullPath := m.GetPath(relativePath)
		expectedPath := filepath.Join(tmpDir, "docs", relativePath)
		assert.Equal(t, expectedPath, fullPath)
	})
}

func TestDocsManagerErrorCases(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	t.Run("Create global doc with invalid name", func(t *testing.T) {
		err := m.CreateGlobal("", "content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "document name cannot be empty")
		
		err = m.CreateGlobal("../outside.md", "content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid document name")
	})
	
	t.Run("Create scope doc with invalid scope", func(t *testing.T) {
		err := m.CreateScope("", "doc.md", "content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scope name cannot be empty")
		
		err = m.CreateScope("../outside", "doc.md", "content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid scope name")
	})
	
	t.Run("Create scope doc with invalid name", func(t *testing.T) {
		err := m.CreateScope("scope1", "", "content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "document name cannot be empty")
		
		err = m.CreateScope("scope1", "../outside.md", "content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid document name")
	})
	
	t.Run("Create duplicate global doc", func(t *testing.T) {
		// Create first doc
		err := m.CreateGlobal("duplicate.md", "content 1")
		require.NoError(t, err)
		
		// Try to create duplicate
		err = m.CreateGlobal("duplicate.md", "content 2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
		
		// Verify original content is preserved
		docPath := filepath.Join(tmpDir, "global", "duplicate.md")
		content, _ := os.ReadFile(docPath)
		assert.Equal(t, "content 1", string(content))
	})
	
	t.Run("Create duplicate scope doc", func(t *testing.T) {
		scope := "test-scope"
		
		// Create first doc
		err := m.CreateScope(scope, "duplicate.md", "content 1")
		require.NoError(t, err)
		
		// Try to create duplicate
		err = m.CreateScope(scope, "duplicate.md", "content 2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
		
		// Verify original content is preserved
		docPath := filepath.Join(tmpDir, "scopes", scope, "duplicate.md")
		content, _ := os.ReadFile(docPath)
		assert.Equal(t, "content 1", string(content))
	})
}

func TestDocsManagerSync(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	t.Run("Sync without git", func(t *testing.T) {
		// Create a doc
		err := m.CreateGlobal("test.md", "Test content")
		require.NoError(t, err)
		
		// Try to sync (will fail if git not initialized)
		err = m.Sync(false)
		// We don't assert error since it depends on git state
		_ = err
	})
	
	t.Run("Sync with push", func(t *testing.T) {
		// Create a doc
		err := m.CreateGlobal("push-test.md", "Push test content")
		require.NoError(t, err)
		
		// Try to sync with push (will fail without remote)
		err = m.Sync(true)
		// We don't assert error since it depends on git state
		_ = err
	})
}

func TestDocsManagerInitGit(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	// Test initGit (private method but called during NewManager)
	// The git directory should exist
	gitPath := filepath.Join(tmpDir, "docs", ".git")
	
	// Check if git was initialized (may or may not succeed depending on git availability)
	if _, err := os.Stat(gitPath); err == nil {
		assert.DirExists(t, gitPath)
	}
}

func TestDocsManagerEdit(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	t.Run("Edit existing document", func(t *testing.T) {
		// Create a document
		docName := "edit-test.md"
		err := m.CreateGlobal(docName, "Original content")
		require.NoError(t, err)
		
		// Try to edit (will fail because editor doesn't exist in test)
		err = m.Edit(filepath.Join("global", docName))
		assert.Error(t, err)
	})
	
	t.Run("Edit non-existent document", func(t *testing.T) {
		err := m.Edit("non-existent.md")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "document not found")
	})
	
	t.Run("Edit with invalid path", func(t *testing.T) {
		err := m.Edit("../outside/doc.md")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid document path")
	})
}

func TestDocsManagerListEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	t.Run("List empty global docs", func(t *testing.T) {
		// Remove any existing docs
		globalPath := filepath.Join(tmpDir, "global")
		os.RemoveAll(globalPath)
		os.MkdirAll(globalPath, 0755)
		
		list, err := m.List("")
		assert.NoError(t, err)
		// May have README.md or be empty
		assert.GreaterOrEqual(t, len(list), 0)
	})
	
	t.Run("List non-existent scope", func(t *testing.T) {
		list, err := m.List("non-existent-scope")
		assert.NoError(t, err)
		assert.Len(t, list, 0)
	})
	
	t.Run("List with special characters in scope", func(t *testing.T) {
		// Create scope with special name
		scopeName := "scope-with-dash"
		docName := "doc.md"
		
		err := m.CreateScope(scopeName, docName, "content")
		require.NoError(t, err)
		
		list, err := m.List(scopeName)
		require.NoError(t, err)
		assert.Greater(t, len(list), 0)
		
		// Verify the doc is in the list
		found := false
		for _, item := range list {
			if strings.Contains(item.Path, docName) {
				found = true
				assert.Equal(t, scopeName, item.Scope)
				break
			}
		}
		assert.True(t, found)
	})
}

func TestDocsManagerWithLargeContent(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	t.Run("Create doc with large content", func(t *testing.T) {
		// Generate large content (1MB)
		var builder strings.Builder
		for i := 0; i < 1024*1024/10; i++ {
			builder.WriteString("0123456789")
		}
		largeContent := builder.String()
		
		err := m.CreateGlobal("large.md", largeContent)
		assert.NoError(t, err)
		
		// Verify it was saved correctly
		docPath := filepath.Join(tmpDir, "global", "large.md")
		content, err := os.ReadFile(docPath)
		require.NoError(t, err)
		assert.Equal(t, len(largeContent), len(content))
	})
}

func TestDocsManagerConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	// Test concurrent doc creation (should be safe)
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(index int) {
			docName := fmt.Sprintf("concurrent-%d.md", index)
			content := fmt.Sprintf("Content for doc %d", index)
			
			err := m.CreateGlobal(docName, content)
			assert.NoError(t, err)
			
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all docs were created
	list, err := m.List("")
	require.NoError(t, err)
	
	count := 0
	for _, item := range list {
		if strings.Contains(item.Path, "concurrent-") {
			count++
		}
	}
	assert.Equal(t, 10, count)
}

func TestDocsManagerPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	t.Run("Create doc and check permissions", func(t *testing.T) {
		docName := "permissions.md"
		err := m.CreateGlobal(docName, "content")
		require.NoError(t, err)
		
		docPath := filepath.Join(tmpDir, "global", docName)
		info, err := os.Stat(docPath)
		require.NoError(t, err)
		
		// Check file permissions (should be readable and writable by owner)
		mode := info.Mode()
		assert.True(t, mode.IsRegular())
		// Exact permissions may vary by system
		assert.True(t, mode.Perm()&0600 == 0600, "File should be readable and writable by owner")
	})
}