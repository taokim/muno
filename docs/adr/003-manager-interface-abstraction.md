# ADR-003: Manager Interface Abstraction for Testability and Plugin Architecture

## Status
Accepted

## Context
The original manager implementation had tight coupling between the orchestration logic and concrete implementations of various subsystems (git operations, file system, UI, etc.). This created several problems:

1. **Poor Testability**: Difficult to write unit tests without involving real file system, git operations, or UI interactions
2. **Rigid Architecture**: No clear extension points for future plugin system
3. **High Coupling**: Changes to implementations required changes to the manager
4. **Mixed Responsibilities**: Manager contained both orchestration logic and implementation details

## Decision
Refactor the manager to use **Dependency Injection** with **Interface Abstraction** pattern:

1. Define provider interfaces for all external dependencies
2. Inject implementations through constructor
3. Manager only depends on abstractions (interfaces)
4. Concrete implementations wrapped in adapters

### Interface Providers
- `ConfigProvider` - Configuration management
- `GitProvider` - Git operations  
- `FileSystemProvider` - File system operations
- `UIProvider` - User interactions
- `TreeProvider` - Tree navigation and state
- `ProcessProvider` - Process and shell execution
- `LogProvider` - Logging operations
- `MetricsProvider` - Metrics collection

### Implementation Strategy
```go
// Interface definition
type GitProvider interface {
    Clone(url, path string, options CloneOptions) error
    Pull(path string, options PullOptions) error
    // ... other methods
}

// Manager uses interface
type ManagerV2 struct {
    gitProvider interfaces.GitProvider
    // ... other providers
}

// Constructor injection
func NewManagerV2(opts ManagerOptionsV2) (*ManagerV2, error) {
    return &ManagerV2{
        gitProvider: opts.GitProvider,  // Injected dependency
    }
}

// Commands use interfaces
func (m *ManagerV2) Add(ctx context.Context, url string, opts AddOptions) error {
    return m.gitProvider.Clone(url, path, cloneOpts)  // Via interface
}
```

## Consequences

### Positive
1. **Improved Testability**: 
   - Easy to inject mocks for unit testing
   - Test coverage increased from 56.9% to 71.7%
   - Can test manager logic in isolation

2. **Plugin Architecture Foundation**:
   - Clear extension points via interfaces
   - Providers can be replaced at runtime
   - Third-party implementations possible

3. **Better Separation of Concerns**:
   - Manager focuses on orchestration
   - Implementations isolated in providers
   - Single responsibility per provider

4. **Flexibility**:
   - Can swap implementations without changing manager
   - Support for different backends (e.g., different git libraries)
   - Environment-specific implementations (testing vs production)

5. **Maintainability**:
   - Changes to implementations don't affect manager
   - Clear contracts via interfaces
   - Easier to understand and modify

### Negative
1. **Increased Complexity**:
   - More files and interfaces to maintain
   - Additional abstraction layer
   - Need to understand dependency injection pattern

2. **Initial Development Overhead**:
   - Required significant refactoring effort
   - Need to maintain interface compatibility

3. **Potential Performance Impact**:
   - Interface calls have minimal overhead
   - Additional object allocations for adapters

## Implementation Notes

### Testing Strategy
- Use mock providers for unit tests (see `internal/mocks/`)
- Real providers for integration tests
- Table-driven tests for comprehensive coverage

### Future Plugin System
This architecture enables future plugin system:
```go
// Future: Plugin can implement providers
type Plugin interface {
    GetProviders() map[string]interface{}
}

// Plugin could provide custom GitProvider
plugin.GetProviders()["git"] // Returns GitProvider implementation
```

### Migration Path
1. ✅ Create provider interfaces
2. ✅ Refactor ManagerV2 to use interfaces
3. ✅ Implement adapters for existing code
4. ✅ Update tests to use mocks
5. ⏳ Future: Add plugin loading mechanism
6. ⏳ Future: Support dynamic provider registration

## References
- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)
- [Dependency Injection in Go](https://blog.drewolson.org/dependency-injection-in-go)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)

## Decision Date
2024-12-XX

## Participants
- Development Team
- Architecture Team

## Revision History
- 2024-12-XX: Initial decision and implementation