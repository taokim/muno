package adapters

import (
	"os"
	"path/filepath"
	
	"github.com/taokim/muno/internal/interfaces"
)

// RealFileSystem implements interfaces.FileSystem using actual OS operations
type RealFileSystem struct{}

// NewRealFileSystem creates a new real filesystem implementation
func NewRealFileSystem() *RealFileSystem {
	return &RealFileSystem{}
}

func (fs *RealFileSystem) Open(name string) (interfaces.File, error) {
	return os.Open(name)
}

func (fs *RealFileSystem) Create(name string) (interfaces.File, error) {
	return os.Create(name)
}

func (fs *RealFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (fs *RealFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (fs *RealFileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (fs *RealFileSystem) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (fs *RealFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fs *RealFileSystem) ReadDir(dirname string) ([]os.DirEntry, error) {
	return os.ReadDir(dirname)
}

func (fs *RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (fs *RealFileSystem) Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(name)
}

func (fs *RealFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (fs *RealFileSystem) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (fs *RealFileSystem) Getwd() (string, error) {
	return os.Getwd()
}

func (fs *RealFileSystem) Chdir(dir string) error {
	return os.Chdir(dir)
}

func (fs *RealFileSystem) TempDir() string {
	return os.TempDir()
}

func (fs *RealFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (fs *RealFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (fs *RealFileSystem) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func (fs *RealFileSystem) Readlink(name string) (string, error) {
	return os.Readlink(name)
}

// FileWrapper wraps os.File to implement interfaces.File
type FileWrapper struct {
	*os.File
}

func (f *FileWrapper) Read(p []byte) (n int, err error) {
	return f.File.Read(p)
}

func (f *FileWrapper) Write(p []byte) (n int, err error) {
	return f.File.Write(p)
}

func (f *FileWrapper) Close() error {
	return f.File.Close()
}

func (f *FileWrapper) Seek(offset int64, whence int) (int64, error) {
	return f.File.Seek(offset, whence)
}

func (f *FileWrapper) Name() string {
	return f.File.Name()
}

func (f *FileWrapper) Stat() (os.FileInfo, error) {
	return f.File.Stat()
}

func (f *FileWrapper) Sync() error {
	return f.File.Sync()
}

func (f *FileWrapper) Truncate(size int64) error {
	return f.File.Truncate(size)
}

// Ensure os.File can be used as interfaces.File
var _ interfaces.File = (*os.File)(nil)

// Helper function to convert path to absolute
func AbsPath(path string) (string, error) {
	return filepath.Abs(path)
}