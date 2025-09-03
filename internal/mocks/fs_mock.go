package mocks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	
	"github.com/taokim/muno/internal/interfaces"
)

// MockFileSystemProvider is a mock implementation of FileSystemProvider
type MockFileSystemProvider struct {
	mu       sync.RWMutex
	files    map[string][]byte
	dirs     map[string]bool
	exists   map[string]bool
	fileInfo map[string]interfaces.FileInfo
	errors   map[string]error
	calls    []string
	walkFunc func(root string, fn filepath.WalkFunc) error
}

// NewMockFileSystemProvider creates a new mock file system provider
func NewMockFileSystemProvider() *MockFileSystemProvider {
	return &MockFileSystemProvider{
		files:    make(map[string][]byte),
		dirs:     make(map[string]bool),
		exists:   make(map[string]bool),
		fileInfo: make(map[string]interfaces.FileInfo),
		errors:   make(map[string]error),
		calls:    []string{},
	}
}

// Exists checks if a path exists
func (m *MockFileSystemProvider) Exists(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Exists(%s)", path))
	
	// Check explicit exists map first
	if exists, ok := m.exists[path]; ok {
		return exists
	}
	
	// Then check files and dirs
	_, hasFile := m.files[path]
	_, hasDir := m.dirs[path]
	return hasFile || hasDir
}

// Create creates a new file
func (m *MockFileSystemProvider) Create(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Create(%s)", path))
	
	if err, ok := m.errors["create:"+path]; ok && err != nil {
		return err
	}
	
	m.files[path] = []byte{}
	m.fileInfo[path] = interfaces.FileInfo{
		Name:    path,
		Size:    0,
		Mode:    0644,
		ModTime: time.Now(),
		IsDir:   false,
	}
	
	return nil
}

// Remove removes a file or empty directory
func (m *MockFileSystemProvider) Remove(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Remove(%s)", path))
	
	if err, ok := m.errors["remove:"+path]; ok && err != nil {
		return err
	}
	
	delete(m.files, path)
	delete(m.dirs, path)
	delete(m.fileInfo, path)
	
	return nil
}

// RemoveAll removes a path and all its contents
func (m *MockFileSystemProvider) RemoveAll(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("RemoveAll(%s)", path))
	
	if err, ok := m.errors["removeall:"+path]; ok && err != nil {
		return err
	}
	
	// Remove path and all subpaths
	for p := range m.files {
		if p == path || len(p) > len(path) && p[:len(path)] == path {
			delete(m.files, p)
			delete(m.fileInfo, p)
		}
	}
	for p := range m.dirs {
		if p == path || len(p) > len(path) && p[:len(path)] == path {
			delete(m.dirs, p)
			delete(m.fileInfo, p)
		}
	}
	
	return nil
}

// ReadDir reads a directory
func (m *MockFileSystemProvider) ReadDir(path string) ([]interfaces.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("ReadDir(%s)", path))
	
	if err, ok := m.errors["readdir:"+path]; ok && err != nil {
		return nil, err
	}
	
	if !m.dirs[path] {
		return nil, fmt.Errorf("not a directory: %s", path)
	}
	
	var entries []interfaces.FileInfo
	for p, info := range m.fileInfo {
		// Check if this is a direct child of path
		if len(p) > len(path) && p[:len(path)] == path {
			// Simple check for direct children (not recursive)
			entries = append(entries, info)
		}
	}
	
	return entries, nil
}

// Mkdir creates a directory
func (m *MockFileSystemProvider) Mkdir(path string, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Mkdir(%s, %v)", path, perm))
	
	if err, ok := m.errors["mkdir:"+path]; ok && err != nil {
		return err
	}
	
	if m.dirs[path] {
		return fmt.Errorf("directory already exists: %s", path)
	}
	
	m.dirs[path] = true
	m.fileInfo[path] = interfaces.FileInfo{
		Name:    path,
		Size:    0,
		Mode:    perm | os.ModeDir,
		ModTime: time.Now(),
		IsDir:   true,
	}
	
	return nil
}

// MkdirAll creates a directory and all parents
func (m *MockFileSystemProvider) MkdirAll(path string, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("MkdirAll(%s, %v)", path, perm))
	
	if err, ok := m.errors["mkdirall:"+path]; ok && err != nil {
		return err
	}
	
	m.dirs[path] = true
	m.fileInfo[path] = interfaces.FileInfo{
		Name:    path,
		Size:    0,
		Mode:    perm | os.ModeDir,
		ModTime: time.Now(),
		IsDir:   true,
	}
	
	return nil
}

// ReadFile reads file contents
func (m *MockFileSystemProvider) ReadFile(path string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("ReadFile(%s)", path))
	
	if err, ok := m.errors["readfile:"+path]; ok && err != nil {
		return nil, err
	}
	
	if data, ok := m.files[path]; ok {
		return data, nil
	}
	
	return nil, fmt.Errorf("file not found: %s", path)
}

// WriteFile writes data to a file
func (m *MockFileSystemProvider) WriteFile(path string, data []byte, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("WriteFile(%s, %d bytes, %v)", path, len(data), perm))
	
	if err, ok := m.errors["writefile:"+path]; ok && err != nil {
		return err
	}
	
	m.files[path] = data
	m.fileInfo[path] = interfaces.FileInfo{
		Name:    path,
		Size:    int64(len(data)),
		Mode:    perm,
		ModTime: time.Now(),
		IsDir:   false,
	}
	
	return nil
}

// Stat returns file information
func (m *MockFileSystemProvider) Stat(path string) (interfaces.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Stat(%s)", path))
	
	if err, ok := m.errors["stat:"+path]; ok && err != nil {
		return interfaces.FileInfo{}, err
	}
	
	if info, ok := m.fileInfo[path]; ok {
		return info, nil
	}
	
	return interfaces.FileInfo{}, fmt.Errorf("no such file or directory: %s", path)
}

// Symlink creates a symbolic link
func (m *MockFileSystemProvider) Symlink(oldname, newname string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Symlink(%s, %s)", oldname, newname))
	
	if err, ok := m.errors["symlink:"+newname]; ok && err != nil {
		return err
	}
	
	// Simple mock: just mark as existing
	m.files[newname] = []byte(oldname) // Store target as content
	m.fileInfo[newname] = interfaces.FileInfo{
		Name:    newname,
		Size:    0,
		Mode:    0777 | os.ModeSymlink,
		ModTime: time.Now(),
		IsDir:   false,
	}
	
	return nil
}

// Rename renames a file or directory
func (m *MockFileSystemProvider) Rename(oldpath, newpath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Rename(%s, %s)", oldpath, newpath))
	
	if err, ok := m.errors["rename:"+oldpath]; ok && err != nil {
		return err
	}
	
	// Move file
	if data, ok := m.files[oldpath]; ok {
		m.files[newpath] = data
		delete(m.files, oldpath)
	}
	
	// Move directory
	if _, ok := m.dirs[oldpath]; ok {
		m.dirs[newpath] = true
		delete(m.dirs, oldpath)
	}
	
	// Move file info
	if info, ok := m.fileInfo[oldpath]; ok {
		info.Name = newpath
		m.fileInfo[newpath] = info
		delete(m.fileInfo, oldpath)
	}
	
	return nil
}

// Copy copies a file from src to dst
func (m *MockFileSystemProvider) Copy(src, dst string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Copy(%s, %s)", src, dst))
	
	if err, ok := m.errors["copy:"+src]; ok && err != nil {
		return err
	}
	
	if data, ok := m.files[src]; ok {
		m.files[dst] = append([]byte(nil), data...) // Deep copy
		if info, ok := m.fileInfo[src]; ok {
			dstInfo := info
			dstInfo.Name = dst
			m.fileInfo[dst] = dstInfo
		}
		return nil
	}
	
	return fmt.Errorf("source file not found: %s", src)
}

// Mock helper methods

// SetExists sets whether a path exists
func (m *MockFileSystemProvider) SetExists(path string, exists bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.exists[path] = exists
}

// SetFile sets mock file content
func (m *MockFileSystemProvider) SetFile(path string, content []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.files[path] = content
	m.fileInfo[path] = interfaces.FileInfo{
		Name:    path,
		Size:    int64(len(content)),
		Mode:    0644,
		ModTime: time.Now(),
		IsDir:   false,
	}
}

// SetDir marks a path as a directory
func (m *MockFileSystemProvider) SetDir(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.dirs[path] = true
	m.fileInfo[path] = interfaces.FileInfo{
		Name:    path,
		Size:    0,
		Mode:    0755 | os.ModeDir,
		ModTime: time.Now(),
		IsDir:   true,
	}
}

// SetError sets an error for a specific operation and path
func (m *MockFileSystemProvider) SetError(operation, path string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.errors[operation+":"+path] = err
}

// GetCalls returns all method calls made
func (m *MockFileSystemProvider) GetCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	calls := make([]string, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// SetWalkFunc sets a custom walk function for testing
func (m *MockFileSystemProvider) SetWalkFunc(walkFunc func(root string, fn filepath.WalkFunc) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.walkFunc = walkFunc
}

// Walk walks the file tree
func (m *MockFileSystemProvider) Walk(root string, fn filepath.WalkFunc) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Walk(%s)", root))
	
	if m.walkFunc != nil {
		return m.walkFunc(root, fn)
	}
	
	// Default implementation - walk through mock dirs
	for path := range m.dirs {
		if strings.HasPrefix(path, root) {
			info := m.fileInfo[path]
			if info.Name == "" {
				info = interfaces.FileInfo{
					Name:  filepath.Base(path),
					IsDir: true,
				}
			}
			// Convert to a type that implements fs.FileInfo
			fileInfo := &mockFileInfo{
				name:  info.Name,
				size:  info.Size,
				mode:  info.Mode,
				modTime: info.ModTime,
				isDir: info.IsDir,
			}
			if err := fn(path, fileInfo, nil); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// mockFileInfo implements fs.FileInfo
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// Reset resets the mock state
func (m *MockFileSystemProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.files = make(map[string][]byte)
	m.dirs = make(map[string]bool)
	m.fileInfo = make(map[string]interfaces.FileInfo)
	m.errors = make(map[string]error)
	m.calls = []string{}
}