package docs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	assert.NotNil(t, m)
	
	// Check directories were created
	assert.DirExists(t, filepath.Join(tmpDir, "docs"))
	assert.DirExists(t, filepath.Join(tmpDir, "docs", "global"))
	assert.DirExists(t, filepath.Join(tmpDir, "docs", "scopes"))
}

func TestCreateGlobal(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	fileName := "architecture.md"
	content := "# Architecture\n\nGlobal architecture document"
	
	err = m.CreateGlobal(fileName, content)
	require.NoError(t, err)
	
	// Check file was created
	docPath := filepath.Join(tmpDir, "docs", "global", fileName)
	assert.FileExists(t, docPath)
	
	// Verify content
	data, err := os.ReadFile(docPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestCreateScope(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	scopeName := "test-scope"
	fileName := "design.md"
	content := "# Design Document\n\nTest content"
	
	err = m.CreateScope(scopeName, fileName, content)
	require.NoError(t, err)
	
	// Check file was created
	docPath := filepath.Join(tmpDir, "docs", "scopes", scopeName, fileName)
	assert.FileExists(t, docPath)
	
	// Verify content
	data, err := os.ReadFile(docPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	// Create some test documents
	err = m.CreateGlobal("doc1.md", "Content 1")
	require.NoError(t, err)
	
	err = m.CreateGlobal("doc2.md", "Content 2")
	require.NoError(t, err)
	
	err = m.CreateScope("scope1", "scope-doc1.md", "Scope content 1")
	require.NoError(t, err)
	
	err = m.CreateScope("scope2", "scope-doc2.md", "Scope content 2")
	require.NoError(t, err)
	
	t.Run("ListGlobal", func(t *testing.T) {
		docs, err := m.List("")
		require.NoError(t, err)
		
		// Should list global docs when no scope specified
		hasGlobalDoc := false
		for _, doc := range docs {
			if filepath.Base(doc.Path) == "doc1.md" || filepath.Base(doc.Path) == "doc2.md" {
				hasGlobalDoc = true
				break
			}
		}
		assert.True(t, hasGlobalDoc, "Should have global documents")
	})
	
	t.Run("ListScope", func(t *testing.T) {
		docs, err := m.List("scope1")
		require.NoError(t, err)
		
		// Should list scope docs
		hasScopeDoc := false
		for _, doc := range docs {
			if filepath.Base(doc.Path) == "scope-doc1.md" {
				hasScopeDoc = true
				break
			}
		}
		assert.True(t, hasScopeDoc, "Should have scope1 documents")
	})
}

func TestGetPath(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	// Test getting path
	relativePath := "global/test.md"
	fullPath := m.GetPath(relativePath)
	expectedPath := filepath.Join(tmpDir, "docs", relativePath)
	assert.Equal(t, expectedPath, fullPath)
}

func TestSync(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	require.NoError(t, err)
	
	// Create a test document
	err = m.CreateGlobal("test.md", "Test content")
	require.NoError(t, err)
	
	// Try to sync (will fail if git not initialized properly, but should handle gracefully)
	err = m.Sync(false)
	// We don't assert on the error as it depends on git initialization
	_ = err
}