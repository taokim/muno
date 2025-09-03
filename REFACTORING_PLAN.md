# MUNO Refactoring Plan: Testability & Plugin Architecture

## Executive Summary

This document outlines a comprehensive refactoring plan for MUNO to achieve two major goals:
1. **Improve testability** of the Manager package to achieve 80%+ code coverage
2. **Introduce plugin architecture** for extensible command system

## Current State Analysis

### Testability Issues
- **Manager Package Coverage**: ~13% (blocking overall 80% target)
- **Root Causes**:
  - Tight coupling with file system (config files must exist)
  - Direct dependencies on TreeManager, Git operations
  - Interactive prompts blocking automated tests
  - No abstraction layer for external dependencies

### Extensibility Limitations
- All commands hardcoded in main application
- No mechanism for third-party extensions
- Monolithic architecture limiting community contributions
- No command namespace isolation

## Proposed Architecture

### 1. Dependency Abstraction Layer

```go
// internal/interfaces/providers.go
package interfaces

// ConfigProvider abstracts configuration operations
type ConfigProvider interface {
    Load(path string) (*config.ConfigTree, error)
    Save(path string, cfg *config.ConfigTree) error
    Exists(path string) bool
    Watch(path string) (<-chan ConfigEvent, error)
}

// GitProvider abstracts git operations
type GitProvider interface {
    Clone(url, path string, options CloneOptions) error
    Pull(path string, options PullOptions) error
    Push(path string, options PushOptions) error
    Status(path string) (*GitStatus, error)
    Commit(path string, message string) error
}

// FileSystemProvider abstracts file operations
type FileSystemProvider interface {
    Exists(path string) bool
    Create(path string) error
    Remove(path string) error
    ReadDir(path string) ([]FileInfo, error)
    Mkdir(path string, perm os.FileMode) error
}

// UIProvider abstracts user interactions
type UIProvider interface {
    Prompt(message string) string
    Confirm(message string) bool
    Select(message string, options []string) string
    Progress(message string) ProgressReporter
}

// TreeProvider abstracts tree operations
type TreeProvider interface {
    Load(config *config.ConfigTree) error
    Navigate(path string) error
    GetCurrent() (*Node, error)
    GetTree() (*Node, error)
    AddNode(parent, child *Node) error
}
```

### 2. Plugin System Architecture

```go
// internal/plugin/interface.go
package plugin

import (
    "context"
    "github.com/hashicorp/go-plugin"
)

// Plugin is the interface that must be implemented by all MUNO plugins
type Plugin interface {
    // Metadata returns plugin information
    Metadata() PluginMetadata
    
    // Commands returns the list of commands this plugin provides
    Commands() []CommandDefinition
    
    // Execute runs a specific command
    Execute(ctx context.Context, cmd string, args []string, env Environment) (Result, error)
    
    // Initialize is called when plugin is loaded
    Initialize(config map[string]interface{}) error
    
    // Cleanup is called before plugin is unloaded
    Cleanup() error
}

// PluginMetadata describes a plugin
type PluginMetadata struct {
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Author      string   `json:"author"`
    Description string   `json:"description"`
    MinMunoVer  string   `json:"min_muno_version"`
    MaxMunoVer  string   `json:"max_muno_version"`
    License     string   `json:"license"`
    Homepage    string   `json:"homepage"`
    Tags        []string `json:"tags"`
}

// CommandDefinition describes a plugin command
type CommandDefinition struct {
    Name        string            `json:"name"`
    Aliases     []string          `json:"aliases"`
    Description string            `json:"description"`
    Usage       string            `json:"usage"`
    Flags       []FlagDefinition  `json:"flags"`
    Examples    []string          `json:"examples"`
    Category    string            `json:"category"`
}

// Environment provides context to plugin execution
type Environment struct {
    WorkspacePath string            `json:"workspace_path"`
    CurrentNode   string            `json:"current_node"`
    TreeState     map[string]interface{} `json:"tree_state"`
    Config        map[string]interface{} `json:"config"`
    Variables     map[string]string `json:"variables"`
}

// Result represents plugin execution result
type Result struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Error   string      `json:"error,omitempty"`
}
```

### 3. Plugin Discovery & Loading

```go
// internal/plugin/manager.go
package plugin

type PluginManager struct {
    registry     map[string]*LoadedPlugin
    searchPaths  []string
    config       *PluginConfig
    loader       PluginLoader
}

// DiscoverPlugins scans for available plugins
func (pm *PluginManager) DiscoverPlugins() error {
    // Search in:
    // 1. $HOME/.muno/plugins/
    // 2. /usr/local/lib/muno/plugins/
    // 3. ./plugins/ (workspace local)
    // 4. Paths from MUNO_PLUGIN_PATH env var
}

// LoadPlugin loads a specific plugin
func (pm *PluginManager) LoadPlugin(name string) error {
    // 1. Find plugin binary/package
    // 2. Verify compatibility
    // 3. Start plugin process (for go-plugin)
    // 4. Register commands
    // 5. Initialize plugin
}

// Plugin communication using Hashicorp go-plugin
var HandshakeConfig = plugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "MUNO_PLUGIN",
    MagicCookieValue: "octopus",
}
```

### 4. Refactored Manager with DI

```go
// internal/manager/manager.go
package manager

type Manager struct {
    // Injected dependencies
    config   interfaces.ConfigProvider
    git      interfaces.GitProvider
    fs       interfaces.FileSystemProvider
    ui       interfaces.UIProvider
    tree     interfaces.TreeProvider
    plugins  *plugin.PluginManager
    
    // Internal state
    workspace string
    opts      ManagerOptions
}

// ManagerOptions for configuration
type ManagerOptions struct {
    ConfigProvider interfaces.ConfigProvider
    GitProvider    interfaces.GitProvider
    FSProvider     interfaces.FileSystemProvider
    UIProvider     interfaces.UIProvider
    TreeProvider   interfaces.TreeProvider
    PluginManager  *plugin.PluginManager
    
    // Optional: defaults will be used if nil
    Logger         Logger
    MetricsClient  MetricsClient
}

// NewManager creates a manager with injected dependencies
func NewManager(opts ManagerOptions) (*Manager, error) {
    // Use defaults for nil options
    if opts.ConfigProvider == nil {
        opts.ConfigProvider = config.NewDefaultProvider()
    }
    if opts.GitProvider == nil {
        opts.GitProvider = git.NewDefaultProvider()
    }
    // ... etc
    
    return &Manager{
        config:  opts.ConfigProvider,
        git:     opts.GitProvider,
        fs:      opts.FSProvider,
        ui:      opts.UIProvider,
        tree:    opts.TreeProvider,
        plugins: opts.PluginManager,
        opts:    opts,
    }, nil
}

// Initialize now optionally loads config
func (m *Manager) Initialize(workspace string, loadConfig bool) error {
    m.workspace = workspace
    
    if loadConfig {
        cfg, err := m.config.Load(filepath.Join(workspace, "muno.yaml"))
        if err != nil {
            return err
        }
        return m.tree.Load(cfg)
    }
    
    return nil
}
```

## Implementation Phases

### Phase 1: Core Abstractions (Week 1-2)
**Goal**: Create abstraction layer without breaking existing functionality

1. Define all interfaces in `internal/interfaces/`
2. Create default implementations wrapping existing code
3. Add interface fields to Manager struct
4. Update Manager methods to use interfaces
5. Ensure all existing tests still pass

**Deliverables**:
- [ ] interfaces package with all provider interfaces
- [ ] Default implementations for each interface
- [ ] Manager using interfaces with backward compatibility
- [ ] All existing tests passing

### Phase 2: Testing Infrastructure (Week 3)
**Goal**: Achieve 80%+ coverage for Manager package

1. Create comprehensive mock implementations
2. Add mock state tracking and assertions
3. Write Manager unit tests using mocks
4. Add integration test suite with real implementations
5. Achieve 80% coverage target

**Deliverables**:
- [ ] Mock implementations in `internal/mocks/`
- [ ] Manager unit tests with 80%+ coverage
- [ ] Integration test suite
- [ ] CI/CD pipeline updates for coverage reporting

### Phase 3: Plugin Foundation (Week 4-5)
**Goal**: Implement core plugin system

1. Implement plugin interface and types
2. Add Hashicorp go-plugin integration
3. Create plugin discovery mechanism
4. Implement plugin lifecycle management
5. Add plugin command routing

**Deliverables**:
- [ ] Plugin interface definition
- [ ] Plugin manager implementation
- [ ] Plugin discovery and loading
- [ ] Command routing to plugins
- [ ] Plugin configuration system

### Phase 4: Example Plugins (Week 6)
**Goal**: Validate plugin architecture with real examples

1. Create `muno-plugin-github` for GitHub integration
2. Create `muno-plugin-docker` for container management
3. Create `muno-plugin-k8s` for Kubernetes operations
4. Document plugin development guide
5. Create plugin template/generator

**Example Plugin Structure**:
```
muno-plugin-github/
├── main.go           # Plugin entry point
├── plugin.go         # Plugin implementation
├── commands/         # Command implementations
│   ├── pr.go        # Pull request commands
│   ├── issue.go     # Issue commands
│   └── release.go   # Release commands
├── manifest.yaml     # Plugin metadata
└── README.md        # Documentation
```

### Phase 5: Plugin Ecosystem (Week 7-8)
**Goal**: Enable community plugin development

1. Create plugin registry/marketplace
2. Add plugin installation commands
3. Implement plugin versioning and updates
4. Add plugin dependency management
5. Create plugin certification process

**New Commands**:
```bash
muno plugin install github     # Install from registry
muno plugin list               # List installed plugins
muno plugin update github      # Update plugin
muno plugin remove github      # Uninstall plugin
muno plugin search docker      # Search registry
muno plugin info github        # Show plugin details
```

## Configuration Changes

### Updated muno.yaml Structure

```yaml
workspace:
  name: my-platform
  repos_dir: repos

# New plugins section
plugins:
  enabled: true
  registry: https://plugins.muno.dev
  local_paths:
    - ~/.muno/plugins
    - ./plugins
  
  installed:
    - name: github
      version: 1.2.0
      config:
        token: ${GITHUB_TOKEN}
        default_org: my-org
    
    - name: docker
      version: 2.0.1
      config:
        registry: docker.io

# Existing tree configuration
tree:
  nodes:
    - name: backend
      repo: https://github.com/org/backend
      children:
        - name: service-a
          repo: https://github.com/org/service-a
```

## Migration Strategy

### Backward Compatibility
1. All existing commands remain unchanged
2. Plugin commands integrate seamlessly as native commands (e.g., `muno github`, `muno docker`)
3. Gradual migration of built-in commands to plugins
4. Configuration versioning for smooth upgrades
5. Deprecation warnings for legacy features

### Risk Mitigation
1. Feature flags for gradual rollout
2. Comprehensive test coverage before each phase
3. Beta testing program for plugin system
4. Rollback procedures for each phase
5. Performance benchmarking to prevent regression

## Success Metrics

### Testability Goals
- [ ] Manager package: 80%+ coverage
- [ ] Overall project: 80%+ coverage
- [ ] All tests run in <30 seconds
- [ ] Mock-based tests: <100ms each
- [ ] Zero flaky tests

### Plugin System Goals
- [ ] Plugin load time: <100ms
- [ ] Plugin command execution: <10ms overhead
- [ ] Support for 10+ concurrent plugins
- [ ] 5+ example plugins available
- [ ] Plugin development guide completed

## Example Plugin Implementation

### GitHub Integration Plugin

```go
// muno-plugin-github/main.go
package main

import (
    "github.com/hashicorp/go-plugin"
    munoplugin "github.com/muno/plugin"
)

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: munoplugin.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "github": &GitHubPlugin{},
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}

// plugin.go
type GitHubPlugin struct{}

func (p *GitHubPlugin) Metadata() munoplugin.PluginMetadata {
    return munoplugin.PluginMetadata{
        Name:        "github",
        Version:     "1.0.0",
        Author:      "MUNO Team",
        Description: "GitHub integration for MUNO",
        MinMunoVer:  "0.5.0",
    }
}

func (p *GitHubPlugin) Commands() []munoplugin.CommandDefinition {
    return []munoplugin.CommandDefinition{
        {
            Name:        "pr",
            Description: "Manage pull requests",
            Usage:       "muno pr [create|list|merge]",
        },
        {
            Name:        "issue",
            Description: "Manage issues",
            Usage:       "muno issue [create|list|close]",
        },
    }
}

func (p *GitHubPlugin) Execute(ctx context.Context, cmd string, args []string, env munoplugin.Environment) (munoplugin.Result, error) {
    switch cmd {
    case "pr":
        return p.handlePR(ctx, args, env)
    case "issue":
        return p.handleIssue(ctx, args, env)
    default:
        return munoplugin.Result{
            Success: false,
            Message: "Unknown command: " + cmd,
        }, nil
    }
}
```

## Testing Strategy

### Unit Testing with Mocks

```go
// internal/manager/manager_test.go
func TestManager_Initialize(t *testing.T) {
    // Setup mocks
    mockConfig := mocks.NewMockConfigProvider()
    mockGit := mocks.NewMockGitProvider()
    mockFS := mocks.NewMockFileSystemProvider()
    mockUI := mocks.NewMockUIProvider()
    mockTree := mocks.NewMockTreeProvider()
    
    // Configure mock behavior
    mockConfig.On("Load", "test/muno.yaml").Return(&config.ConfigTree{
        Workspace: config.WorkspaceTree{Name: "test"},
    }, nil)
    mockTree.On("Load", mock.Anything).Return(nil)
    
    // Create manager with mocks
    manager, err := NewManager(ManagerOptions{
        ConfigProvider: mockConfig,
        GitProvider:    mockGit,
        FSProvider:     mockFS,
        UIProvider:     mockUI,
        TreeProvider:   mockTree,
    })
    require.NoError(t, err)
    
    // Test initialization
    err = manager.Initialize("test", true)
    assert.NoError(t, err)
    
    // Verify mock interactions
    mockConfig.AssertExpectations(t)
    mockTree.AssertExpectations(t)
}
```

### Plugin Testing

```go
// internal/plugin/manager_test.go
func TestPluginManager_LoadPlugin(t *testing.T) {
    pm := NewPluginManager()
    
    // Create test plugin
    testPlugin := createTestPlugin(t)
    defer testPlugin.Cleanup()
    
    // Load plugin
    err := pm.LoadPlugin(testPlugin.Path)
    assert.NoError(t, err)
    
    // Verify plugin loaded
    plugin, exists := pm.GetPlugin("test-plugin")
    assert.True(t, exists)
    assert.Equal(t, "1.0.0", plugin.Metadata().Version)
    
    // Test command execution
    result, err := plugin.Execute(context.Background(), "test", []string{}, Environment{})
    assert.NoError(t, err)
    assert.True(t, result.Success)
}
```

## Timeline & Milestones

| Week | Phase | Milestone | Success Criteria |
|------|-------|-----------|------------------|
| 1-2 | Phase 1 | Core Abstractions | All interfaces defined, backward compatible |
| 3 | Phase 2 | Testing Infrastructure | Manager at 80% coverage |
| 4-5 | Phase 3 | Plugin Foundation | Plugin system functional |
| 6 | Phase 4 | Example Plugins | 3+ working plugins |
| 7-8 | Phase 5 | Plugin Ecosystem | Registry operational, 5+ plugins |

## Risk Analysis

### Technical Risks
1. **Plugin Security**: Malicious plugins could compromise system
   - Mitigation: Sandboxing, code signing, permission system
2. **Performance Impact**: Plugins could slow down operations
   - Mitigation: Performance budgets, timeout controls
3. **Compatibility Issues**: Version mismatches between core and plugins
   - Mitigation: Strict versioning, compatibility matrix

### Process Risks
1. **Scope Creep**: Feature additions delaying core refactoring
   - Mitigation: Strict phase boundaries, feature freeze
2. **Breaking Changes**: Existing users affected by refactoring
   - Mitigation: Comprehensive testing, gradual rollout
3. **Documentation Lag**: Plugin system complex without good docs
   - Mitigation: Documentation-first approach, examples

## Conclusion

This refactoring plan addresses both immediate testing needs and long-term extensibility goals. The phased approach minimizes risk while delivering incremental value. The plugin architecture positions MUNO as an extensible platform for multi-repository orchestration, enabling community contributions and custom workflows.

## Next Steps

1. Review and approve this plan
2. Create detailed technical design documents for each phase
3. Set up project tracking and milestones
4. Begin Phase 1 implementation
5. Establish plugin developer preview program

## Appendix A: Alternative Plugin Architectures Considered

1. **Lua Scripting**: Embedded Lua interpreter
   - Pros: Lightweight, sandboxed, popular in CLI tools
   - Cons: Another language to learn, limited ecosystem

2. **WebAssembly (WASM)**: Run WASM modules as plugins
   - Pros: Language agnostic, sandboxed, growing ecosystem
   - Cons: Immature tooling, complex for simple plugins

3. **JSON-RPC over stdio**: Simple process communication
   - Pros: Very simple, language agnostic
   - Cons: Limited features, manual protocol management

4. **Shared Libraries (.so/.dll)**: Dynamic loading
   - Pros: Fast, direct API access
   - Cons: Language locked, versioning hell, platform specific

**Decision**: Hashicorp go-plugin chosen for maturity, features, and Go ecosystem fit.

## Appendix B: Config Provider Implementations

```go
// For testing - in-memory config
type InMemoryConfigProvider struct {
    configs map[string]*config.ConfigTree
}

// For production - file-based config
type FileConfigProvider struct {
    cache map[string]*config.ConfigTree
}

// For plugins - remote config
type RemoteConfigProvider struct {
    endpoint string
    client   *http.Client
}
```