package interfaces

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

// ConfigProvider abstracts configuration operations
type ConfigProvider interface {
	Load(path string) (interface{}, error)
	Save(path string, cfg interface{}) error
	Exists(path string) bool
	Watch(path string) (<-chan ConfigEvent, error)
}

// ConfigEvent represents a configuration change event
type ConfigEvent struct {
	Type      string
	Path      string
	Timestamp time.Time
	Error     error
}

// GitProvider abstracts git operations
type GitProvider interface {
	Clone(url, path string, options CloneOptions) error
	Pull(path string, options PullOptions) error
	Push(path string, options PushOptions) error
	Status(path string) (*GitStatus, error)
	Commit(path string, message string, options CommitOptions) error
	Branch(path string) (string, error)
	Checkout(path string, branch string) error
	Fetch(path string, options FetchOptions) error
	Add(path string, files []string) error
	Remove(path string, files []string) error
	GetRemoteURL(path string) (string, error)
	SetRemoteURL(path string, url string) error
}

// CloneOptions for git clone operations
type CloneOptions struct {
	Branch    string
	Depth     int
	Recursive bool
	Quiet     bool
}

// PullOptions for git pull operations
type PullOptions struct {
	Rebase    bool
	Force     bool
	Recursive bool
	Quiet     bool
}

// PushOptions for git push operations
type PushOptions struct {
	Force     bool
	SetUpstream bool
	Branch    string
	Quiet     bool
}

// FetchOptions for git fetch operations
type FetchOptions struct {
	All       bool
	Prune     bool
	Tags      bool
	Quiet     bool
}

// CommitOptions for git commit operations
type CommitOptions struct {
	All       bool
	Amend     bool
	NoVerify  bool
}

// GitStatus represents the status of a git repository
type GitStatus struct {
	Branch        string
	IsClean       bool
	HasUntracked  bool
	HasStaged     bool
	HasModified   bool
	HasChanges    bool  // Added for compatibility
	FilesModified int   // Added for compatibility
	FilesAdded    int   // Added for compatibility
	Files         []GitFileStatus
	Ahead         int
	Behind        int
}

// GitPullResult represents the result of a git pull operation
type GitPullResult struct {
	UpdatesReceived bool
	FilesChanged    int
	Message         string
}

// GitPushResult represents the result of a git push operation  
type GitPushResult struct {
	Success bool
	Message string
}

// GitFileStatus represents the status of a single file
type GitFileStatus struct {
	Path      string
	Status    string // "modified", "added", "deleted", "untracked", etc.
	Staged    bool
}

// FileSystemProvider abstracts file operations
type FileSystemProvider interface {
	Exists(path string) bool
	Create(path string) error
	Remove(path string) error
	RemoveAll(path string) error
	ReadDir(path string) ([]FileInfo, error)
	Mkdir(path string, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	Stat(path string) (FileInfo, error)
	Symlink(oldname, newname string) error
	Rename(oldpath, newpath string) error
	Copy(src, dst string) error
	Walk(root string, fn filepath.WalkFunc) error
}

// FileInfo represents file information
type FileInfo struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
}

// UIProvider abstracts user interactions
type UIProvider interface {
	Prompt(message string) (string, error)
	PromptPassword(message string) (string, error)
	Confirm(message string) (bool, error)
	Select(message string, options []string) (string, error)
	MultiSelect(message string, options []string) ([]string, error)
	Progress(message string) ProgressReporter
	Info(message string)
	Success(message string)
	Warning(message string)
	Error(message string)
	Debug(message string)
}

// ProgressReporter provides progress reporting capabilities
type ProgressReporter interface {
	Start()
	Update(current, total int)
	SetMessage(message string)
	Finish()
	Error(err error)
}

// TreeProvider abstracts tree operations
type TreeProvider interface {
	Load(config interface{}) error
	Navigate(path string) error
	GetCurrent() (NodeInfo, error)
	GetTree() (NodeInfo, error)
	GetNode(path string) (NodeInfo, error)
	AddNode(parentPath string, node NodeInfo) error
	RemoveNode(path string) error
	UpdateNode(path string, node NodeInfo) error
	ListChildren(path string) ([]NodeInfo, error)
	GetPath() string
	SetPath(path string) error
	GetState() (TreeState, error)
	SetState(state TreeState) error
}

// NodeInfo represents information about a tree node
type NodeInfo struct {
	Name        string
	Path        string
	Repository  string
	ConfigFile  string     // Path to config file for config nodes
	IsConfig    bool       // True if this is a config reference node
	IsLazy      bool
	IsCloned    bool
	HasChanges  bool
	Children    []NodeInfo
	Parent      *NodeInfo
}

// TreeState represents the state of the entire tree
type TreeState struct {
	CurrentPath string
	Nodes       map[string]NodeInfo
	History     []string
}

// ProcessProvider abstracts process and shell operations
type ProcessProvider interface {
	Execute(ctx context.Context, command string, args []string, options ProcessOptions) (*ProcessResult, error)
	ExecuteShell(ctx context.Context, command string, options ProcessOptions) (*ProcessResult, error)
	StartBackground(ctx context.Context, command string, args []string, options ProcessOptions) (Process, error)
	OpenInEditor(path string) error
	OpenInBrowser(url string) error
}

// ProcessOptions for process execution
type ProcessOptions struct {
	WorkingDir string
	Env        []string
	Stdin      string
	Timeout    time.Duration
	Silent     bool
}

// ProcessResult represents the result of process execution
type ProcessResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// Process interface is defined in interfaces.go

// LogProvider abstracts logging operations
type LogProvider interface {
	Debug(message string, fields ...Field)
	Info(message string, fields ...Field)
	Warn(message string, fields ...Field)
	Error(message string, fields ...Field)
	Fatal(message string, fields ...Field)
	WithFields(fields ...Field) LogProvider
	SetLevel(level LogLevel)
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// LogLevel represents logging levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// MetricsProvider abstracts metrics collection
type MetricsProvider interface {
	Counter(name string, value int64, tags ...string)
	Gauge(name string, value float64, tags ...string)
	Histogram(name string, value float64, tags ...string)
	Timer(name string) TimerMetric
	Flush() error
}

// TimerMetric represents a timer metric
type TimerMetric interface {
	Start()
	Stop() time.Duration
	Record(duration time.Duration)
}