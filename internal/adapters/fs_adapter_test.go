package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystemAdapter_Exists(t *testing.T) {
	fs := NewFileSystemAdapter()
	tmpDir := t.TempDir()

	// Test existing directory
	exists := fs.Exists(tmpDir)
	assert.True(t, exists)

	// Test non-existing path
	exists = fs.Exists(filepath.Join(tmpDir, "nonexistent"))
	assert.False(t, exists)

	// Create a file and test
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)
	
	exists = fs.Exists(testFile)
	assert.True(t, exists)
}

func TestFileSystemAdapter_CreateRemove(t *testing.T) {
	fs := NewFileSystemAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create file
	err := fs.Create(testFile)
	assert.NoError(t, err)
	assert.True(t, fs.Exists(testFile))

	// Remove file
	err = fs.Remove(testFile)
	assert.NoError(t, err)
	assert.False(t, fs.Exists(testFile))

	// Remove non-existent should error
	err = fs.Remove(testFile)
	assert.Error(t, err)
}

func TestFileSystemAdapter_RemoveAll(t *testing.T) {
	fs := NewFileSystemAdapter()
	tmpDir := t.TempDir()
	
	// Create nested structure
	nestedDir := filepath.Join(tmpDir, "nested", "deep")
	err := os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)
	
	testFile := filepath.Join(nestedDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// RemoveAll should remove everything
	err = fs.RemoveAll(filepath.Join(tmpDir, "nested"))
	assert.NoError(t, err)
	assert.False(t, fs.Exists(filepath.Join(tmpDir, "nested")))
}

func TestFileSystemAdapter_Directories(t *testing.T) {
	fs := NewFileSystemAdapter()
	tmpDir := t.TempDir()

	// Test Mkdir
	dir1 := filepath.Join(tmpDir, "dir1")
	err := fs.Mkdir(dir1, 0755)
	assert.NoError(t, err)
	assert.True(t, fs.Exists(dir1))

	// Mkdir existing should error
	err = fs.Mkdir(dir1, 0755)
	assert.Error(t, err)

	// Test MkdirAll
	deepDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	err = fs.MkdirAll(deepDir, 0755)
	assert.NoError(t, err)
	assert.True(t, fs.Exists(deepDir))

	// MkdirAll existing should not error
	err = fs.MkdirAll(deepDir, 0755)
	assert.NoError(t, err)
}

func TestFileSystemAdapter_ReadWriteFile(t *testing.T) {
	fs := NewFileSystemAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write file
	content := []byte("test content")
	err := fs.WriteFile(testFile, content, 0644)
	assert.NoError(t, err)

	// Read file
	readContent, err := fs.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)

	// Read non-existent file
	_, err = fs.ReadFile(filepath.Join(tmpDir, "nonexistent.txt"))
	assert.Error(t, err)
}

func TestFileSystemAdapter_ReadDir(t *testing.T) {
	fs := NewFileSystemAdapter()
	tmpDir := t.TempDir()

	// Create some files and directories
	for i := 0; i < 3; i++ {
		fileName := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".txt")
		err := os.WriteFile(fileName, []byte("content"), 0644)
		require.NoError(t, err)
		
		dirName := filepath.Join(tmpDir, "dir"+string(rune('0'+i)))
		err = os.Mkdir(dirName, 0755)
		require.NoError(t, err)
	}

	// Read directory
	entries, err := fs.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, entries, 6) // 3 files + 3 directories

	// Read non-existent directory
	_, err = fs.ReadDir(filepath.Join(tmpDir, "nonexistent"))
	assert.Error(t, err)
}

func TestFileSystemAdapter_Stat(t *testing.T) {
	fs := NewFileSystemAdapter()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create file
	content := []byte("test content")
	err := os.WriteFile(testFile, content, 0644)
	require.NoError(t, err)

	// Stat file
	info, err := fs.Stat(testFile)
	assert.NoError(t, err)
	assert.Equal(t, "test.txt", info.Name)
	assert.Equal(t, int64(len(content)), info.Size)
	assert.False(t, info.IsDir)

	// Stat directory
	info, err = fs.Stat(tmpDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir)

	// Stat non-existent
	_, err = fs.Stat(filepath.Join(tmpDir, "nonexistent"))
	assert.Error(t, err)
}

func TestFileSystemAdapter_Symlink(t *testing.T) {
	fs := NewFileSystemAdapter()
	tmpDir := t.TempDir()
	
	// Create source file
	sourceFile := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("source content"), 0644)
	require.NoError(t, err)

	// Create symlink
	linkFile := filepath.Join(tmpDir, "link.txt")
	err = fs.Symlink(sourceFile, linkFile)
	assert.NoError(t, err)

	// Read through symlink
	content, err := os.ReadFile(linkFile)
	assert.NoError(t, err)
	assert.Equal(t, []byte("source content"), content)

	// Check if it's a symlink
	info, err := os.Lstat(linkFile)
	assert.NoError(t, err)
	assert.Equal(t, os.ModeSymlink, info.Mode()&os.ModeSymlink)
}

func TestFileSystemAdapter_Rename(t *testing.T) {
	fs := NewFileSystemAdapter()
	tmpDir := t.TempDir()
	
	// Create source file
	oldPath := filepath.Join(tmpDir, "old.txt")
	newPath := filepath.Join(tmpDir, "new.txt")
	err := os.WriteFile(oldPath, []byte("content"), 0644)
	require.NoError(t, err)

	// Rename file
	err = fs.Rename(oldPath, newPath)
	assert.NoError(t, err)
	assert.False(t, fs.Exists(oldPath))
	assert.True(t, fs.Exists(newPath))

	// Verify content preserved
	content, err := os.ReadFile(newPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("content"), content)

	// Rename non-existent should error
	err = fs.Rename(filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "other"))
	assert.Error(t, err)
}

func TestFileSystemAdapter_Copy(t *testing.T) {
	fs := NewFileSystemAdapter()
	tmpDir := t.TempDir()
	
	// Create source file
	sourceFile := filepath.Join(tmpDir, "source.txt")
	destFile := filepath.Join(tmpDir, "dest.txt")
	sourceContent := []byte("source content")
	err := os.WriteFile(sourceFile, sourceContent, 0644)
	require.NoError(t, err)

	// Copy file
	err = fs.Copy(sourceFile, destFile)
	assert.NoError(t, err)
	assert.True(t, fs.Exists(sourceFile))
	assert.True(t, fs.Exists(destFile))

	// Verify content
	destContent, err := os.ReadFile(destFile)
	assert.NoError(t, err)
	assert.Equal(t, sourceContent, destContent)

	// Copy non-existent should error
	err = fs.Copy(filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "other"))
	assert.Error(t, err)

	// Copy directory should error (not supported)
	sourceDir := filepath.Join(tmpDir, "sourcedir")
	destDir := filepath.Join(tmpDir, "destdir")
	err = os.Mkdir(sourceDir, 0755)
	require.NoError(t, err)
	
	// Trying to copy directory should fail
	err = fs.Copy(sourceDir, destDir)
	assert.Error(t, err) // Should error because it's a directory
}

func TestFileSystemAdapter_Walk(t *testing.T) {
	fs := NewFileSystemAdapter()
	tmpDir := t.TempDir()

	// Create directory structure
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	err := os.Mkdir(dir1, 0755)
	require.NoError(t, err)
	err = os.Mkdir(dir2, 0755)
	require.NoError(t, err)

	// Create files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(dir1, "file2.txt")
	file3 := filepath.Join(dir2, "file3.txt")
	
	for _, file := range []string{file1, file2, file3} {
		err = os.WriteFile(file, []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Walk and collect paths
	var paths []string
	err = fs.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Store relative path for easier comparison
		relPath, _ := filepath.Rel(tmpDir, path)
		if relPath == "." {
			relPath = ""
		}
		paths = append(paths, relPath)
		return nil
	})
	
	assert.NoError(t, err)
	assert.Contains(t, paths, "")
	assert.Contains(t, paths, "dir1")
	assert.Contains(t, paths, "dir2")
	assert.Contains(t, paths, "file1.txt")
	assert.Contains(t, paths, filepath.Join("dir1", "file2.txt"))
	assert.Contains(t, paths, filepath.Join("dir2", "file3.txt"))

	// Walk non-existent directory
	err = fs.Walk(filepath.Join(tmpDir, "nonexistent"), func(path string, info os.FileInfo, err error) error {
		return err
	})
	assert.Error(t, err)
}