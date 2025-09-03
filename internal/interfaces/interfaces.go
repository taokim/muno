// Package interfaces provides all interfaces for dependency injection and testing
package interfaces

import (
	"io"
	"os"
	"time"
)

// FileSystem interface abstracts file system operations
type FileSystem interface {
	// File operations
	Open(name string) (File, error)
	Create(name string) (File, error)
	Remove(name string) error
	RemoveAll(path string) error
	Rename(oldpath, newpath string) error
	
	// Directory operations
	Mkdir(name string, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	ReadDir(dirname string) ([]os.DirEntry, error)
	
	// File info operations
	Stat(name string) (os.FileInfo, error)
	Lstat(name string) (os.FileInfo, error)
	Exists(path string) bool
	IsDir(path string) bool
	
	// Path operations
	Getwd() (string, error)
	Chdir(dir string) error
	TempDir() string
	
	// File content operations
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	
	// Symlink operations
	Symlink(oldname, newname string) error
	Readlink(name string) (string, error)
}

// File interface abstracts file operations
type File interface {
	io.Reader
	io.Writer
	io.Closer
	io.Seeker
	
	Name() string
	Stat() (os.FileInfo, error)
	Sync() error
	Truncate(size int64) error
}

// CommandExecutor interface abstracts command execution
type CommandExecutor interface {
	// Execute runs a command and returns output
	Execute(name string, args ...string) ([]byte, error)
	
	// ExecuteWithInput runs a command with stdin input
	ExecuteWithInput(input string, name string, args ...string) ([]byte, error)
	
	// ExecuteInDir runs a command in a specific directory
	ExecuteInDir(dir string, name string, args ...string) ([]byte, error)
	
	// ExecuteWithEnv runs a command with environment variables
	ExecuteWithEnv(env []string, name string, args ...string) ([]byte, error)
	
	// Start starts a command without waiting for it to complete
	Start(name string, args ...string) (Process, error)
	
	// StartInDir starts a command in a specific directory
	StartInDir(dir string, name string, args ...string) (Process, error)
}

// Process interface abstracts process management
type Process interface {
	// Wait waits for the process to complete
	Wait() error
	
	// Kill terminates the process
	Kill() error
	
	// Signal sends a signal to the process
	Signal(sig os.Signal) error
	
	// Pid returns the process ID
	Pid() int
	
	// StdoutPipe returns a pipe for stdout
	StdoutPipe() (io.ReadCloser, error)
	
	// StderrPipe returns a pipe for stderr
	StderrPipe() (io.ReadCloser, error)
	
	// StdinPipe returns a pipe for stdin
	StdinPipe() (io.WriteCloser, error)
}

// GitInterface abstracts git operations
type GitInterface interface {
	// Basic operations
	Clone(url, path string) error
	CloneWithOptions(url, path string, options ...string) error
	Pull(path string) error
	PullWithOptions(path string, options ...string) error
	Push(path string) error
	PushWithOptions(path string, options ...string) error
	Fetch(path string) error
	FetchWithOptions(path string, options ...string) error
	
	// Status and info
	Status(path string) (string, error)
	StatusWithOptions(path string, options ...string) (string, error)
	Branch(path string) (string, error)
	CurrentBranch(path string) (string, error)
	RemoteURL(path string) (string, error)
	HasChanges(path string) (bool, error)
	IsRepo(path string) bool
	
	// Commit operations
	Add(path string, files ...string) error
	AddAll(path string) error
	Commit(path, message string) error
	CommitWithOptions(path, message string, options ...string) error
	
	// Branch operations
	Checkout(path, branch string) error
	CheckoutNew(path, branch string) error
	CreateBranch(path, branch string) error
	DeleteBranch(path, branch string) error
	ListBranches(path string) ([]string, error)
	
	// Tag operations
	Tag(path, tag string) error
	TagWithMessage(path, tag, message string) error
	ListTags(path string) ([]string, error)
	
	// Reset operations
	Reset(path string) error
	ResetHard(path string) error
	ResetSoft(path string) error
	
	// Diff operations
	Diff(path string) (string, error)
	DiffStaged(path string) (string, error)
	DiffWithBranch(path, branch string) (string, error)
	
	// Log operations
	Log(path string, limit int) ([]string, error)
	LogOneline(path string, limit int) ([]string, error)
	
	// Remote operations
	AddRemote(path, name, url string) error
	RemoveRemote(path, name string) error
	ListRemotes(path string) (map[string]string, error)
}

// Output interface abstracts output operations
type Output interface {
	io.Writer
	
	// Print methods
	Print(a ...interface{}) (n int, err error)
	Printf(format string, a ...interface{}) (n int, err error)
	Println(a ...interface{}) (n int, err error)
	
	// Error output
	Error(a ...interface{}) (n int, err error)
	Errorf(format string, a ...interface{}) (n int, err error)
	Errorln(a ...interface{}) (n int, err error)
	
	// Colored output
	Success(message string)
	Info(message string)
	Warning(message string)
	Danger(message string)
	
	// Formatting
	Bold(text string) string
	Italic(text string) string
	Underline(text string) string
	Color(text string, color string) string
}

// ConfigManager interface abstracts configuration management
type ConfigManager interface {
	// Load and save
	Load(path string) error
	Save(path string) error
	
	// Get operations
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetStringSlice(key string) []string
	GetStringMap(key string) map[string]interface{}
	
	// Set operations
	Set(key string, value interface{})
	SetDefault(key string, value interface{})
	
	// Check operations
	IsSet(key string) bool
	AllKeys() []string
	
	// Nested config
	Sub(key string) ConfigManager
	
	// Unmarshal
	Unmarshal(rawVal interface{}) error
	UnmarshalKey(key string, rawVal interface{}) error
}

// Logger interface abstracts logging operations
type Logger interface {
	// Log levels
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	
	// Structured logging
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
	
	// Log level control
	SetLevel(level string)
	GetLevel() string
}

// HTTPClient interface abstracts HTTP operations
type HTTPClient interface {
	Get(url string) ([]byte, error)
	Post(url string, contentType string, body []byte) ([]byte, error)
	Put(url string, contentType string, body []byte) ([]byte, error)
	Delete(url string) ([]byte, error)
	Head(url string) (int, error)
	
	// With headers
	GetWithHeaders(url string, headers map[string]string) ([]byte, error)
	PostWithHeaders(url string, body []byte, headers map[string]string) ([]byte, error)
}

// TimeProvider interface abstracts time operations
type TimeProvider interface {
	Now() time.Time
	Since(t time.Time) time.Duration
	Until(t time.Time) time.Duration
	Sleep(d time.Duration)
	After(d time.Duration) <-chan time.Time
	NewTimer(d time.Duration) Timer
	NewTicker(d time.Duration) Ticker
}

// Timer interface abstracts timer operations
type Timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(d time.Duration) bool
}

// Ticker interface abstracts ticker operations  
type Ticker interface {
	C() <-chan time.Time
	Stop()
	Reset(d time.Duration)
}

// Environment interface abstracts environment variable operations
type Environment interface {
	Get(key string) string
	Set(key, value string) error
	Unset(key string) error
	Lookup(key string) (string, bool)
	Environ() []string
	Expand(s string) string
}

// UserInterface abstracts user interaction
type UserInterface interface {
	// Input
	Prompt(message string) (string, error)
	PromptPassword(message string) (string, error)
	Confirm(message string) (bool, error)
	Select(message string, options []string) (int, error)
	MultiSelect(message string, options []string) ([]int, error)
	
	// Output
	Print(message string)
	Printf(format string, args ...interface{})
	Error(message string)
	Errorf(format string, args ...interface{})
	Success(message string)
	Warning(message string)
	Info(message string)
	
	// Progress
	StartProgress(message string, total int) ProgressBar
	StartSpinner(message string) Spinner
}

// ProgressBar interface abstracts progress bar operations
type ProgressBar interface {
	Add(n int)
	Set(n int)
	Finish()
	SetMessage(message string)
}

// Spinner interface abstracts spinner operations
type Spinner interface {
	Start()
	Stop()
	SetMessage(message string)
}