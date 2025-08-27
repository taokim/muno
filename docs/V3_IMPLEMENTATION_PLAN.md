# V3 Implementation Plan: Tree-Based Architecture

## Overview
Complete transformation from v2 scope-based to v3 tree-based architecture with CWD-first resolution.

## Phase 1: Scope Removal & Cleanup (Day 1-2)

### 1.1 Remove Scope Package
```bash
# Actions:
- Delete internal/scope/ directory
- Remove scope imports from all files
- Delete scope-related test files
- Remove scope commands from CLI
```

### 1.2 Clean Configuration
```go
// Remove from internal/config/types.go:
- ScopeConfig struct
- ScopeType constants  
- All scope-related fields from Config struct

// Update config to v3:
type ConfigV3 struct {
    Version   int    `yaml:"version"`
    Workspace struct {
        Name     string `yaml:"name"`
        RootRepo string `yaml:"root_repo"`
    } `yaml:"workspace"`
}
```

### 1.3 Update Documentation
```bash
# Files to update:
- README.md - Remove scope mentions
- CLAUDE.md - Remove scope concepts
- docs/*.md - Clean all scope references
- examples/*.yaml - Update to v3 format
```

### Checklist Phase 1:
- [ ] Delete internal/scope/ package
- [ ] Remove scope imports from manager
- [ ] Clean CLI commands (remove scope subcommands)
- [ ] Update config structures
- [ ] Clean documentation files
- [ ] Update example configs
- [ ] Fix compilation errors
- [ ] Run tests to identify broken areas

## Phase 2: Tree Foundation (Day 3-5)

### 2.1 Create Tree Package
```go
// internal/tree/types.go
package tree

type Node struct {
    ID       string           `json:"id"`
    Name     string           `json:"name"`
    Path     string           `json:"path"`
    Parent   *Node            `json:"-"`
    Children map[string]*Node `json:"children,omitempty"`
    Repos    []RepoConfig     `json:"repos,omitempty"`
    Meta     NodeMeta         `json:"meta"`
}

type RepoConfig struct {
    URL    string `json:"url"`
    Path   string `json:"path"`
    Name   string `json:"name"`
    Lazy   bool   `json:"lazy"`
    State  string `json:"state"` // "missing", "cloned", "modified"
}

type NodeMeta struct {
    CreatedAt string `json:"created_at"`
    Type      string `json:"type"` // "persistent", "ephemeral"
    README    string `json:"readme,omitempty"`
}
```

### 2.2 Tree Manager Core
```go
// internal/tree/manager.go
type Manager struct {
    rootPath    string
    rootNode    *Node
    currentNode *Node
    config      *config.ConfigV3
    statePath   string
}

func (m *Manager) ResolveTarget(explicitPath string) (*Node, error)
func (m *Manager) LoadTree() error
func (m *Manager) SaveTree() error
func (m *Manager) mapCWDToNode(cwd string) *Node
```

### 2.3 State Management
```go
// internal/tree/state.go
type TreeState struct {
    CurrentNodePath string            `json:"current_node_path"`
    Nodes          map[string]*Node   `json:"nodes"`
    LastUpdated    string            `json:"last_updated"`
}
```

### Checklist Phase 2:
- [ ] Create internal/tree package structure
- [ ] Define Node and RepoConfig types
- [ ] Implement Manager struct
- [ ] Create ResolveTarget with CWD-first logic
- [ ] Implement tree persistence (load/save)
- [ ] Add CWD to node mapping
- [ ] Create unit tests for tree operations
- [ ] Integrate with manager package

## Phase 3: Navigation Commands (Day 6-8)

### 3.1 Core Navigation
```go
// cmd/repo-claude/commands/use.go
func UseCommand(c *cli.Context) error {
    target := c.Args().First()
    noClone := c.Bool("no-clone")
    
    // Navigate and change CWD
    node, err := mgr.Use(target, UseOptions{NoClone: noClone})
    
    // Display feedback
    fmt.Printf("🎯 Navigated to: %s\n", node.Path)
    fmt.Printf("📂 Changed directory to: %s\n", node.FullPath())
    
    // Auto-clone lazy repos
    if !noClone {
        cloned := mgr.CloneLazyRepos(node)
        fmt.Printf("🔄 Auto-cloned %d lazy repositories\n", cloned)
    }
}
```

### 3.2 Display Commands
```go
// Tree display with depth control
func TreeCommand(c *cli.Context) error
func ListCommand(c *cli.Context) error  
func CurrentCommand(c *cli.Context) error
```

### 3.3 Path Resolution
```go
// Support various path formats:
- Absolute: /team/backend
- Relative: ../frontend  
- Current: .
- Parent: ..
- Previous: -
- Root: ~ or /
```

### Checklist Phase 3:
- [ ] Implement use command with CWD change
- [ ] Add --no-clone flag for navigation
- [ ] Create current command
- [ ] Implement tree command with --depth
- [ ] Add list command for children
- [ ] Support all path formats
- [ ] Add navigation feedback displays
- [ ] Update help text and documentation

## Phase 4: Repository Management (Day 9-10)

### 4.1 Simplified Add Command
```go
func AddCommand(c *cli.Context) error {
    repoURL := c.Args().First()
    name := c.String("name")
    lazy := c.Bool("lazy")
    
    // Resolve current node
    node, _ := mgr.ResolveTarget("")
    fmt.Printf("🎯 Target: %s (from %s)\n", node.Path, getResolutionSource())
    
    // Add repository
    repo := mgr.AddRepo(node, repoURL, AddOptions{
        Name: name,
        Lazy: lazy,
    })
    
    if !lazy {
        fmt.Printf("✓ Cloned %s\n", repo.Name)
    } else {
        fmt.Printf("📦 Added %s (lazy)\n", repo.Name)
    }
}
```

### 4.2 Remove & Clone Commands
```go
func RemoveCommand(c *cli.Context) error
func CloneCommand(c *cli.Context) error  // For lazy repos
```

### Checklist Phase 4:
- [ ] Implement add command with lazy support
- [ ] Add remove command
- [ ] Create clone command for lazy repos
- [ ] Support --recursive flag
- [ ] Add auto-clone on navigation
- [ ] Update node metadata on changes
- [ ] Handle duplicate detection
- [ ] Add validation for URLs

## Phase 5: Git Operations (Day 11-12)

### 5.1 Target Resolution Display
```go
// Every git command shows target
func gitOperation(op string, recursive bool) error {
    node, source := mgr.ResolveTarget("")
    
    fmt.Printf("🎯 Target: %s (from %s)\n", node.Path, source)
    if recursive {
        count := countRepos(node, recursive)
        fmt.Printf("🔄 Recursive: %d repositories\n", count)
    }
    
    // Execute operation...
}
```

### 5.2 Git Commands
```go
func PullCommand(c *cli.Context) error
func PushCommand(c *cli.Context) error
func CommitCommand(c *cli.Context) error
func StatusCommand(c *cli.Context) error
```

### Checklist Phase 5:
- [ ] Update all git commands with target display
- [ ] Add --recursive flag support
- [ ] Implement parallel operations
- [ ] Add confirmation for recursive operations
- [ ] Update status with tree display
- [ ] Handle git operation errors
- [ ] Add progress indicators
- [ ] Test with various tree depths

## Phase 6: Session Management (Day 13)

### 6.1 Start Command Update
```go
func StartCommand(c *cli.Context) error {
    target := c.Args().First()
    node, source := mgr.ResolveTarget(target)
    
    fmt.Printf("🎯 Starting session at: %s (from %s)\n", node.Path, source)
    
    // Generate CLAUDE.md for context
    generateClaudeContext(node)
    
    // Start Claude session
    return mgr.StartSession(node)
}
```

### Checklist Phase 6:
- [ ] Update start command with target resolution
- [ ] Generate tree-aware CLAUDE.md
- [ ] Update session tracking for nodes
- [ ] Add stop command
- [ ] Handle recursive session starts
- [ ] Update interactive mode for tree
- [ ] Test session management
- [ ] Update documentation

## Phase 7: Testing & Polish (Day 14-15)

### 7.1 Comprehensive Testing
```bash
# Test scenarios:
- CWD resolution from various locations
- Lazy loading behavior
- Recursive operations
- Path resolution edge cases
- Tree persistence across restarts
- Large tree performance
```

### 7.2 User Experience
```bash
# Improvements:
- Consistent target display format
- Clear error messages
- Progress indicators
- Confirmation prompts
- Help text updates
```

### Checklist Phase 7:
- [ ] Write comprehensive tests
- [ ] Test CWD resolution thoroughly  
- [ ] Verify lazy loading works
- [ ] Test recursive operations
- [ ] Check tree persistence
- [ ⠀Performance testing
- [ ] Update all documentation
- [ ] Create migration guide from v2

## Success Metrics

1. **All v2 scope references removed**
2. **Tree navigation working smoothly**
3. **CWD-first resolution consistent**
4. **Target always displayed clearly**
5. **Lazy loading functional**
6. **Tests passing with >70% coverage**
7. **Documentation fully updated**

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Breaking existing users | Keep v2 binary available |
| CWD resolution confusion | Clear target display always |
| Performance with large trees | Implement lazy tree loading |
| Lost work during migration | Create backup before starting |

## Timeline Summary

- **Week 1**: Cleanup + Tree Foundation (Phase 1-2)
- **Week 2**: Commands + Features (Phase 3-5)  
- **Week 3**: Polish + Testing (Phase 6-7)

Total estimated time: 15 working days