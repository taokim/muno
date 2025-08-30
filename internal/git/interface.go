package git

// Interface defines the methods for git operations
type Interface interface {
	Clone(url, path string) error
	Pull(path string) error
	Status(path string) (string, error)
	Commit(path, message string) error
	Push(path string) error
	Add(path, pattern string) error
}

// Ensure Git implements Interface
var _ Interface = (*Git)(nil)