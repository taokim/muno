# Interface Patterns for External Dependencies

This document describes the interface patterns used in the repo-claude codebase to ensure proper abstraction of external dependencies for testing and maintainability.

## Core Interfaces

### CommandExecutor
Used for executing external commands. This allows us to mock command execution in tests without actually running external programs.

```go
type CommandExecutor interface {
    Command(name string, arg ...string) Cmd
}
```

**Used in:**
- `manager/start_options.go` - For launching Claude sessions
- `manager/scopes.go` - For starting scope processes

**Should be used in (future refactoring):**
- `git/git.go` - For git operations
- `manager/process_info.go` - For system commands

### ProcessManager
Used for process management operations like sending signals to processes.

```go
type ProcessManager interface {
    FindProcess(pid int) (*os.Process, error)
    Signal(p *os.Process, sig os.Signal) error
}
```

**Used in:**
- `manager/agents.go` - For stopping agent processes
- `manager/scopes.go` - For stopping scope processes

### FileSystem
Used for file system operations.

```go
type FileSystem interface {
    ReadFile(name string) ([]byte, error)
    WriteFile(name string, data []byte, perm os.FileMode) error
    Stat(name string) (os.FileInfo, error)
    MkdirAll(path string, perm os.FileMode) error
    Remove(name string) error
    RemoveAll(path string) error
    ReadDir(name string) ([]os.DirEntry, error)
}
```

**Used in:**
- Various manager operations for file I/O

## Best Practices

1. **Always use interfaces for external operations**
   - Any operation that touches the OS, file system, network, or external processes should go through an interface
   - This ensures testability and maintainability

2. **Initialize interfaces properly**
   - Check if the interface is nil and initialize with the real implementation if needed
   - Example:
   ```go
   if m.CmdExecutor == nil {
       m.CmdExecutor = &RealCommandExecutor{}
   }
   ```

3. **Mock interfaces in tests**
   - Use mock implementations that simulate the behavior without actual external calls
   - Track calls and return predetermined responses

4. **Avoid direct usage of:**
   - `exec.Command()` - Use CommandExecutor instead
   - `process.Signal()`, `process.Kill()` - Use ProcessManager instead
   - Direct file operations - Use FileSystem instead
   - Direct network calls - Use appropriate network interfaces

## Areas for Future Improvement

1. **Git Package (`internal/git/git.go`)**
   - Currently uses `exec.Command` directly
   - Should be refactored to use CommandExecutor interface
   - Would allow better testing of git operations

2. **Process Info (`internal/manager/process_info.go`)**
   - Uses `exec.Command` for system commands like `sysctl`
   - Could benefit from CommandExecutor interface for consistency

3. **Integration Tests**
   - Currently use `exec.Command` directly
   - Consider whether these should also use interfaces or remain as true integration tests

## Testing Benefits

Using these interfaces provides:
- **Isolation**: Tests don't depend on external tools being installed
- **Speed**: No actual process spawning or file I/O in unit tests
- **Determinism**: Tests always produce the same results
- **Coverage**: Can test error conditions that are hard to reproduce with real systems