package adapters

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	
	"github.com/taokim/muno/internal/interfaces"
)

// FileSystemAdapter wraps standard file system operations
type FileSystemAdapter struct{}

// NewFileSystemAdapter creates a new file system adapter
func NewFileSystemAdapter() interfaces.FileSystemProvider {
	return &FileSystemAdapter{}
}

// Exists checks if a path exists
func (f *FileSystemAdapter) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Create creates a new file
func (f *FileSystemAdapter) Create(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return file.Close()
}

// Remove removes a file or empty directory
func (f *FileSystemAdapter) Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll removes a path and all its contents
func (f *FileSystemAdapter) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// ReadDir reads a directory
func (f *FileSystemAdapter) ReadDir(path string) ([]interfaces.FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	
	var files []interfaces.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		files = append(files, interfaces.FileInfo{
			Name:    info.Name(),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})
	}
	
	return files, nil
}

// Mkdir creates a directory
func (f *FileSystemAdapter) Mkdir(path string, perm os.FileMode) error {
	return os.Mkdir(path, perm)
}

// MkdirAll creates a directory and all parents
func (f *FileSystemAdapter) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// ReadFile reads file contents
func (f *FileSystemAdapter) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to a file
func (f *FileSystemAdapter) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// Stat returns file information
func (f *FileSystemAdapter) Stat(path string) (interfaces.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return interfaces.FileInfo{}, err
	}
	
	return interfaces.FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

// Symlink creates a symbolic link
func (f *FileSystemAdapter) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

// Rename renames a file or directory
func (f *FileSystemAdapter) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Copy copies a file from src to dst
func (f *FileSystemAdapter) Copy(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()
	
	// Get source file info
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}
	
	// Create destination file
	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()
	
	// Copy contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}
	
	// Preserve timestamps
	return os.Chtimes(dst, time.Now(), sourceInfo.ModTime())
}
// Walk walks the file tree rooted at root
func (f *FileSystemAdapter) Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}
