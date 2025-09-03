package adapters

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/interfaces"
)

func TestRealFileSystem_FileOperations(t *testing.T) {
	fs := NewRealFileSystem()
	tmpDir := t.TempDir()
	
	t.Run("Create and Open", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "test.txt")
		
		// Create file
		file, err := fs.Create(filePath)
		require.NoError(t, err)
		assert.NotNil(t, file)
		
		// Write content
		content := []byte("test content")
		n, err := file.Write(content)
		assert.NoError(t, err)
		assert.Equal(t, len(content), n)
		
		err = file.Close()
		assert.NoError(t, err)
		
		// Open and read
		file, err = fs.Open(filePath)
		require.NoError(t, err)
		
		readContent := make([]byte, len(content))
		n, err = file.Read(readContent)
		assert.NoError(t, err)
		assert.Equal(t, len(content), n)
		assert.Equal(t, content, readContent)
		
		err = file.Close()
		assert.NoError(t, err)
	})
	
	t.Run("Remove", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "remove.txt")
		
		// Create file
		err := os.WriteFile(filePath, []byte("test"), 0644)
		require.NoError(t, err)
		
		// Remove
		err = fs.Remove(filePath)
		assert.NoError(t, err)
		
		// Verify removed
		assert.False(t, fs.Exists(filePath))
	})
	
	t.Run("RemoveAll", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "removedir")
		filePath := filepath.Join(dirPath, "file.txt")
		
		// Create directory with file
		err := os.MkdirAll(dirPath, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filePath, []byte("test"), 0644)
		require.NoError(t, err)
		
		// RemoveAll
		err = fs.RemoveAll(dirPath)
		assert.NoError(t, err)
		
		// Verify removed
		assert.False(t, fs.Exists(dirPath))
	})
	
	t.Run("Rename", func(t *testing.T) {
		oldPath := filepath.Join(tmpDir, "old.txt")
		newPath := filepath.Join(tmpDir, "new.txt")
		
		// Create file
		err := os.WriteFile(oldPath, []byte("test"), 0644)
		require.NoError(t, err)
		
		// Rename
		err = fs.Rename(oldPath, newPath)
		assert.NoError(t, err)
		
		// Verify
		assert.False(t, fs.Exists(oldPath))
		assert.True(t, fs.Exists(newPath))
	})
}

func TestRealFileSystem_DirectoryOperations(t *testing.T) {
	fs := NewRealFileSystem()
	tmpDir := t.TempDir()
	
	t.Run("Mkdir", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "newdir")
		
		err := fs.Mkdir(dirPath, 0755)
		assert.NoError(t, err)
		
		assert.True(t, fs.IsDir(dirPath))
	})
	
	t.Run("MkdirAll", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "deep", "nested", "dir")
		
		err := fs.MkdirAll(dirPath, 0755)
		assert.NoError(t, err)
		
		assert.True(t, fs.IsDir(dirPath))
	})
	
	t.Run("ReadDir", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "readdir")
		err := os.MkdirAll(dirPath, 0755)
		require.NoError(t, err)
		
		// Create some files and directories
		err = os.WriteFile(filepath.Join(dirPath, "file1.txt"), []byte("test"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dirPath, "file2.txt"), []byte("test"), 0644)
		require.NoError(t, err)
		err = os.Mkdir(filepath.Join(dirPath, "subdir"), 0755)
		require.NoError(t, err)
		
		entries, err := fs.ReadDir(dirPath)
		assert.NoError(t, err)
		assert.Len(t, entries, 3)
		
		// Check entries
		names := make(map[string]bool)
		for _, entry := range entries {
			names[entry.Name()] = true
		}
		assert.True(t, names["file1.txt"])
		assert.True(t, names["file2.txt"])
		assert.True(t, names["subdir"])
	})
}

func TestRealFileSystem_FileInfo(t *testing.T) {
	fs := NewRealFileSystem()
	tmpDir := t.TempDir()
	
	t.Run("Stat", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "stat.txt")
		content := []byte("test content")
		err := os.WriteFile(filePath, content, 0644)
		require.NoError(t, err)
		
		info, err := fs.Stat(filePath)
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "stat.txt", info.Name())
		assert.Equal(t, int64(len(content)), info.Size())
		assert.False(t, info.IsDir())
	})
	
	t.Run("Lstat", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "lstat.txt")
		err := os.WriteFile(filePath, []byte("test"), 0644)
		require.NoError(t, err)
		
		info, err := fs.Lstat(filePath)
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "lstat.txt", info.Name())
	})
	
	t.Run("Exists", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "exists.txt")
		
		// File doesn't exist
		assert.False(t, fs.Exists(filePath))
		
		// Create file
		err := os.WriteFile(filePath, []byte("test"), 0644)
		require.NoError(t, err)
		
		// File exists
		assert.True(t, fs.Exists(filePath))
	})
	
	t.Run("IsDir", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "file.txt")
		dirPath := filepath.Join(tmpDir, "dir")
		
		err := os.WriteFile(filePath, []byte("test"), 0644)
		require.NoError(t, err)
		err = os.Mkdir(dirPath, 0755)
		require.NoError(t, err)
		
		assert.False(t, fs.IsDir(filePath))
		assert.True(t, fs.IsDir(dirPath))
		assert.False(t, fs.IsDir(filepath.Join(tmpDir, "nonexistent")))
	})
}

func TestRealFileSystem_PathOperations(t *testing.T) {
	fs := NewRealFileSystem()
	
	t.Run("Getwd", func(t *testing.T) {
		cwd, err := fs.Getwd()
		assert.NoError(t, err)
		assert.NotEmpty(t, cwd)
	})
	
	t.Run("Chdir", func(t *testing.T) {
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir) // Restore original directory
		
		tmpDir := t.TempDir()
		err = fs.Chdir(tmpDir)
		assert.NoError(t, err)
		
		currentDir, err := os.Getwd()
		assert.NoError(t, err)
		// On macOS, /var is a symlink to /private/var, so we need to resolve both paths
		resolvedTmpDir, _ := filepath.EvalSymlinks(tmpDir)
		resolvedCurrentDir, _ := filepath.EvalSymlinks(currentDir)
		assert.Equal(t, resolvedTmpDir, resolvedCurrentDir)
	})
	
	t.Run("TempDir", func(t *testing.T) {
		tempDir := fs.TempDir()
		assert.NotEmpty(t, tempDir)
		assert.True(t, fs.IsDir(tempDir))
	})
}

func TestRealFileSystem_FileContent(t *testing.T) {
	fs := NewRealFileSystem()
	tmpDir := t.TempDir()
	
	t.Run("ReadFile and WriteFile", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "content.txt")
		content := []byte("test content\nwith newlines\n")
		
		// Write file
		err := fs.WriteFile(filePath, content, 0644)
		assert.NoError(t, err)
		
		// Read file
		readContent, err := fs.ReadFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, content, readContent)
	})
	
	t.Run("WriteFile with permissions", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "perms.txt")
		
		err := fs.WriteFile(filePath, []byte("test"), 0600)
		assert.NoError(t, err)
		
		info, err := os.Stat(filePath)
		require.NoError(t, err)
		
		// Check permissions (only checking user read/write on Unix-like systems)
		mode := info.Mode()
		assert.True(t, mode.Perm()&0600 == 0600)
	})
}

func TestRealFileSystem_Symlink(t *testing.T) {
	// Skip on Windows as symlinks require special permissions
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink tests on Windows")
	}
	
	fs := NewRealFileSystem()
	tmpDir := t.TempDir()
	
	t.Run("Symlink and Readlink", func(t *testing.T) {
		targetPath := filepath.Join(tmpDir, "target.txt")
		linkPath := filepath.Join(tmpDir, "link.txt")
		
		// Create target file
		err := os.WriteFile(targetPath, []byte("target content"), 0644)
		require.NoError(t, err)
		
		// Create symlink
		err = fs.Symlink(targetPath, linkPath)
		assert.NoError(t, err)
		
		// Read symlink
		target, err := fs.Readlink(linkPath)
		assert.NoError(t, err)
		assert.Equal(t, targetPath, target)
		
		// Verify link works
		content, err := os.ReadFile(linkPath)
		assert.NoError(t, err)
		assert.Equal(t, []byte("target content"), content)
	})
}

func TestFileWrapper(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "wrapper.txt")
	
	// Create file with os.Create
	file, err := os.Create(filePath)
	require.NoError(t, err)
	defer file.Close()
	
	// Test that os.File implements our File interface
	var _ interfaces.File = file
	
	t.Run("Write and Read", func(t *testing.T) {
		content := []byte("test content")
		
		n, err := file.Write(content)
		assert.NoError(t, err)
		assert.Equal(t, len(content), n)
		
		// Seek to beginning
		pos, err := file.Seek(0, io.SeekStart)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), pos)
		
		// Read
		readContent := make([]byte, len(content))
		n, err = file.Read(readContent)
		assert.NoError(t, err)
		assert.Equal(t, len(content), n)
		assert.Equal(t, content, readContent)
	})
	
	t.Run("File methods", func(t *testing.T) {
		// Name
		assert.Equal(t, filePath, file.Name())
		
		// Stat
		info, err := file.Stat()
		assert.NoError(t, err)
		assert.NotNil(t, info)
		
		// Sync
		err = file.Sync()
		assert.NoError(t, err)
		
		// Truncate
		err = file.Truncate(5)
		assert.NoError(t, err)
		
		info, err = file.Stat()
		assert.NoError(t, err)
		assert.Equal(t, int64(5), info.Size())
	})
}

func TestAbsPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
	}{
		{
			name:    "relative path",
			input:   "./test",
			wantErr: false,
		},
		{
			name:    "absolute path",
			input:   "/tmp/test",
			wantErr: false,
		},
		{
			name:    "home directory",
			input:   "~/test",
			wantErr: false,
		},
		{
			name:    "current directory",
			input:   ".",
			wantErr: false,
		},
		{
			name:    "parent directory",
			input:   "..",
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absPath, err := AbsPath(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, absPath)
				assert.True(t, filepath.IsAbs(absPath))
			}
		})
	}
}