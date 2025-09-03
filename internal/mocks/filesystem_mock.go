package mocks

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
	
	"github.com/taokim/muno/internal/interfaces"
)

// MockFileSystem implements interfaces.FileSystem for testing
type MockFileSystem struct {
	mu      sync.RWMutex
	files   map[string]*MockFile
	dirs    map[string]bool
	cwd     string
	tempDir string
	
	// Function overrides for custom behavior
	OpenFunc      func(name string) (interfaces.File, error)
	CreateFunc    func(name string) (interfaces.File, error)
	RemoveFunc    func(name string) error
	RemoveAllFunc func(path string) error
	RenameFunc    func(oldpath, newpath string) error
	MkdirFunc     func(name string, perm os.FileMode) error
	MkdirAllFunc  func(path string, perm os.FileMode) error
	ReadDirFunc   func(dirname string) ([]os.DirEntry, error)
	StatFunc      func(name string) (os.FileInfo, error)
	LstatFunc     func(name string) (os.FileInfo, error)
	ExistsFunc    func(path string) bool
	IsDirFunc     func(path string) bool
	GetwdFunc     func() (string, error)
	ChdirFunc     func(dir string) error
	TempDirFunc   func() string
	ReadFileFunc  func(filename string) ([]byte, error)
	WriteFileFunc func(filename string, data []byte, perm os.FileMode) error
	SymlinkFunc   func(oldname, newname string) error
	ReadlinkFunc  func(name string) (string, error)
	
	// Call tracking
	Calls []string
}

// NewMockFileSystem creates a new mock filesystem
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files:   make(map[string]*MockFile),
		dirs:    make(map[string]bool),
		cwd:     "/mock",
		tempDir: "/tmp/mock",
	}
}

// MockFile implements interfaces.File
type MockFile struct {
	name   string
	buffer *bytes.Buffer
	closed bool
	pos    int64
	mode   os.FileMode
	
	// Mock file info
	size    int64
	modTime time.Time
	isDir   bool
}

// NewMockFile creates a new mock file
func NewMockFile(name string) *MockFile {
	return &MockFile{
		name:    name,
		buffer:  &bytes.Buffer{},
		mode:    0644,
		modTime: time.Now(),
	}
}

// Read implements io.Reader
func (f *MockFile) Read(p []byte) (n int, err error) {
	if f.closed {
		return 0, fmt.Errorf("file closed")
	}
	return f.buffer.Read(p)
}

// Write implements io.Writer
func (f *MockFile) Write(p []byte) (n int, err error) {
	if f.closed {
		return 0, fmt.Errorf("file closed")
	}
	n, err = f.buffer.Write(p)
	f.size = int64(f.buffer.Len())
	return n, err
}

// Close implements io.Closer
func (f *MockFile) Close() error {
	if f.closed {
		return fmt.Errorf("file already closed")
	}
	f.closed = true
	return nil
}

// Seek implements io.Seeker
func (f *MockFile) Seek(offset int64, whence int) (int64, error) {
	if f.closed {
		return 0, fmt.Errorf("file closed")
	}
	
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = f.pos + offset
	case io.SeekEnd:
		newPos = f.size + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}
	
	if newPos < 0 {
		return 0, fmt.Errorf("negative position")
	}
	
	f.pos = newPos
	return newPos, nil
}

// Name returns the file name
func (f *MockFile) Name() string {
	return f.name
}

// Stat returns file info
func (f *MockFile) Stat() (os.FileInfo, error) {
	return &MockFileInfo{
		name:    filepath.Base(f.name),
		size:    f.size,
		mode:    f.mode,
		modTime: f.modTime,
		isDir:   f.isDir,
	}, nil
}

// Sync flushes the file
func (f *MockFile) Sync() error {
	if f.closed {
		return fmt.Errorf("file closed")
	}
	return nil
}

// Truncate changes the size of the file
func (f *MockFile) Truncate(size int64) error {
	if f.closed {
		return fmt.Errorf("file closed")
	}
	if size < 0 {
		return fmt.Errorf("negative size")
	}
	
	if size < f.size {
		data := f.buffer.Bytes()[:size]
		f.buffer = bytes.NewBuffer(data)
	}
	f.size = size
	return nil
}

// MockFileInfo implements os.FileInfo
type MockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi *MockFileInfo) Name() string       { return fi.name }
func (fi *MockFileInfo) Size() int64        { return fi.size }
func (fi *MockFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *MockFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *MockFileInfo) IsDir() bool        { return fi.isDir }
func (fi *MockFileInfo) Sys() interface{}   { return nil }

// MockDirEntry implements os.DirEntry
type MockDirEntry struct {
	name  string
	isDir bool
	mode  os.FileMode
	info  os.FileInfo
}

func (de *MockDirEntry) Name() string               { return de.name }
func (de *MockDirEntry) IsDir() bool                { return de.isDir }
func (de *MockDirEntry) Type() os.FileMode          { return de.mode }
func (de *MockDirEntry) Info() (os.FileInfo, error) { return de.info, nil }

// FileSystem interface implementation

func (fs *MockFileSystem) Open(name string) (interfaces.File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("Open(%s)", name))
	
	if fs.OpenFunc != nil {
		return fs.OpenFunc(name)
	}
	
	if file, ok := fs.files[name]; ok {
		return file, nil
	}
	
	return nil, os.ErrNotExist
}

func (fs *MockFileSystem) Create(name string) (interfaces.File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("Create(%s)", name))
	
	if fs.CreateFunc != nil {
		return fs.CreateFunc(name)
	}
	
	file := NewMockFile(name)
	fs.files[name] = file
	return file, nil
}

func (fs *MockFileSystem) Remove(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("Remove(%s)", name))
	
	if fs.RemoveFunc != nil {
		return fs.RemoveFunc(name)
	}
	
	delete(fs.files, name)
	delete(fs.dirs, name)
	return nil
}

func (fs *MockFileSystem) RemoveAll(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("RemoveAll(%s)", path))
	
	if fs.RemoveAllFunc != nil {
		return fs.RemoveAllFunc(path)
	}
	
	// Remove all files and dirs with prefix
	for k := range fs.files {
		if filepath.HasPrefix(k, path) {
			delete(fs.files, k)
		}
	}
	for k := range fs.dirs {
		if filepath.HasPrefix(k, path) {
			delete(fs.dirs, k)
		}
	}
	
	return nil
}

func (fs *MockFileSystem) Rename(oldpath, newpath string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("Rename(%s, %s)", oldpath, newpath))
	
	if fs.RenameFunc != nil {
		return fs.RenameFunc(oldpath, newpath)
	}
	
	if file, ok := fs.files[oldpath]; ok {
		fs.files[newpath] = file
		delete(fs.files, oldpath)
	}
	
	if _, ok := fs.dirs[oldpath]; ok {
		fs.dirs[newpath] = true
		delete(fs.dirs, oldpath)
	}
	
	return nil
}

func (fs *MockFileSystem) Mkdir(name string, perm os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("Mkdir(%s, %v)", name, perm))
	
	if fs.MkdirFunc != nil {
		return fs.MkdirFunc(name, perm)
	}
	
	if fs.dirs[name] {
		return os.ErrExist
	}
	
	fs.dirs[name] = true
	return nil
}

func (fs *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("MkdirAll(%s, %v)", path, perm))
	
	if fs.MkdirAllFunc != nil {
		return fs.MkdirAllFunc(path, perm)
	}
	
	// Create all parent directories
	parts := filepath.SplitList(path)
	current := ""
	for _, part := range parts {
		current = filepath.Join(current, part)
		fs.dirs[current] = true
	}
	
	return nil
}

func (fs *MockFileSystem) ReadDir(dirname string) ([]os.DirEntry, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("ReadDir(%s)", dirname))
	
	if fs.ReadDirFunc != nil {
		return fs.ReadDirFunc(dirname)
	}
	
	var entries []os.DirEntry
	
	// Add directories
	for path := range fs.dirs {
		if filepath.Dir(path) == dirname {
			entries = append(entries, &MockDirEntry{
				name:  filepath.Base(path),
				isDir: true,
				mode:  os.ModeDir | 0755,
			})
		}
	}
	
	// Add files
	for path, file := range fs.files {
		if filepath.Dir(path) == dirname {
			info, _ := file.Stat()
			entries = append(entries, &MockDirEntry{
				name:  filepath.Base(path),
				isDir: false,
				mode:  0644,
				info:  info,
			})
		}
	}
	
	return entries, nil
}

func (fs *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("Stat(%s)", name))
	
	if fs.StatFunc != nil {
		return fs.StatFunc(name)
	}
	
	if file, ok := fs.files[name]; ok {
		return file.Stat()
	}
	
	if _, ok := fs.dirs[name]; ok {
		return &MockFileInfo{
			name:  filepath.Base(name),
			isDir: true,
			mode:  os.ModeDir | 0755,
		}, nil
	}
	
	return nil, os.ErrNotExist
}

func (fs *MockFileSystem) Lstat(name string) (os.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("Lstat(%s)", name))
	
	if fs.LstatFunc != nil {
		return fs.LstatFunc(name)
	}
	
	return fs.Stat(name)
}

func (fs *MockFileSystem) Exists(path string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("Exists(%s)", path))
	
	if fs.ExistsFunc != nil {
		return fs.ExistsFunc(path)
	}
	
	_, fileExists := fs.files[path]
	_, dirExists := fs.dirs[path]
	
	return fileExists || dirExists
}

func (fs *MockFileSystem) IsDir(path string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("IsDir(%s)", path))
	
	if fs.IsDirFunc != nil {
		return fs.IsDirFunc(path)
	}
	
	return fs.dirs[path]
}

func (fs *MockFileSystem) Getwd() (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	fs.Calls = append(fs.Calls, "Getwd()")
	
	if fs.GetwdFunc != nil {
		return fs.GetwdFunc()
	}
	
	return fs.cwd, nil
}

func (fs *MockFileSystem) Chdir(dir string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("Chdir(%s)", dir))
	
	if fs.ChdirFunc != nil {
		return fs.ChdirFunc(dir)
	}
	
	if !fs.dirs[dir] {
		return os.ErrNotExist
	}
	
	fs.cwd = dir
	return nil
}

func (fs *MockFileSystem) TempDir() string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	fs.Calls = append(fs.Calls, "TempDir()")
	
	if fs.TempDirFunc != nil {
		return fs.TempDirFunc()
	}
	
	return fs.tempDir
}

func (fs *MockFileSystem) ReadFile(filename string) ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("ReadFile(%s)", filename))
	
	if fs.ReadFileFunc != nil {
		return fs.ReadFileFunc(filename)
	}
	
	if file, ok := fs.files[filename]; ok {
		return file.buffer.Bytes(), nil
	}
	
	return nil, os.ErrNotExist
}

func (fs *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("WriteFile(%s, %d bytes, %v)", filename, len(data), perm))
	
	if fs.WriteFileFunc != nil {
		return fs.WriteFileFunc(filename, data, perm)
	}
	
	file := NewMockFile(filename)
	file.buffer = bytes.NewBuffer(data)
	file.size = int64(len(data))
	file.mode = perm
	fs.files[filename] = file
	
	return nil
}

func (fs *MockFileSystem) Symlink(oldname, newname string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("Symlink(%s, %s)", oldname, newname))
	
	if fs.SymlinkFunc != nil {
		return fs.SymlinkFunc(oldname, newname)
	}
	
	// Simple symlink simulation
	fs.files[newname] = &MockFile{
		name:   newname,
		buffer: bytes.NewBufferString(oldname),
		mode:   os.ModeSymlink | 0777,
	}
	
	return nil
}

func (fs *MockFileSystem) Readlink(name string) (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	fs.Calls = append(fs.Calls, fmt.Sprintf("Readlink(%s)", name))
	
	if fs.ReadlinkFunc != nil {
		return fs.ReadlinkFunc(name)
	}
	
	if file, ok := fs.files[name]; ok {
		if file.mode&os.ModeSymlink != 0 {
			return file.buffer.String(), nil
		}
	}
	
	return "", fmt.Errorf("not a symlink")
}

// Helper methods for testing

// AddFile adds a file to the mock filesystem
func (fs *MockFileSystem) AddFile(path string, content []byte) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	file := NewMockFile(path)
	file.buffer = bytes.NewBuffer(content)
	file.size = int64(len(content))
	fs.files[path] = file
	
	// Ensure parent directories exist
	dir := filepath.Dir(path)
	for dir != "." && dir != "/" {
		fs.dirs[dir] = true
		dir = filepath.Dir(dir)
	}
}

// AddDir adds a directory to the mock filesystem
func (fs *MockFileSystem) AddDir(path string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	fs.dirs[path] = true
	
	// Ensure parent directories exist
	dir := filepath.Dir(path)
	for dir != "." && dir != "/" {
		fs.dirs[dir] = true
		dir = filepath.Dir(dir)
	}
}

// GetFile returns the content of a file
func (fs *MockFileSystem) GetFile(path string) ([]byte, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	
	if file, ok := fs.files[path]; ok {
		return file.buffer.Bytes(), true
	}
	
	return nil, false
}

// Reset clears the mock filesystem
func (fs *MockFileSystem) Reset() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	fs.files = make(map[string]*MockFile)
	fs.dirs = make(map[string]bool)
	fs.cwd = "/mock"
	fs.Calls = nil
}