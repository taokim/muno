# 🎯 Implementation Plan: Recursive Workspaces with Distributed Documentation

## Executive Summary

Transform repo-claude from a flat workspace manager into a **recursive, tree-based system** that naturally scales to enterprise level (500+ repositories) while maintaining autonomy at each level. Each workspace becomes a self-contained unit that can reference other workspaces, with documentation distributed throughout the tree.

---

## 🏗️ Architecture Overview

### Core Concept
```
Platform Root (repo-claude)
├── Backend Platform (repo-claude)
│   ├── Payment Services (repo-claude)
│   │   ├── payment-gateway (git repo)
│   │   └── fraud-detection (git repo)
│   └── Order Services (repo-claude)
│       ├── order-api (git repo)
│       └── inventory-api (git repo)
└── Frontend Platform (repo-claude)
    ├── Web Apps (repo-claude)
    │   └── customer-portal (git repo)
    └── Mobile Apps (repo-claude)
        ├── ios-app (git repo)
        └── android-app (git repo)
```

Each level:
- Has its own `repo-claude.yaml` configuration
- Maintains its own documentation in git
- Can work completely independently
- Can be composed by parent workspaces

---

## 📋 Phase 1: Foundation (Week 1-2)

### Objective
Establish core recursive data model while maintaining full backward compatibility.

### 1.1 Data Model Enhancement

```go
// internal/config/config.go
type WorkspaceType string

const (
    WorkspaceTypeLeaf      WorkspaceType = "leaf"      // Only repositories (current behavior)
    WorkspaceTypeAggregate WorkspaceType = "aggregate" // Only sub-workspaces
    WorkspaceTypeHybrid    WorkspaceType = "hybrid"    // Both repos and sub-workspaces
)

type Config struct {
    // Existing fields remain unchanged
    Version       int                     `yaml:"version"`
    WorkspaceName string                  `yaml:"workspace_name"`
    
    // New fields with omitempty for backward compatibility
    Type          WorkspaceType           `yaml:"type,omitempty"`
    SubWorkspaces []SubWorkspace          `yaml:"sub_workspaces,omitempty"`
    Documentation DocumentationConfig     `yaml:"documentation,omitempty"`
    
    // Runtime fields (not in YAML)
    Parent        *WorkspaceRef           `yaml:"-"`
    Path          string                  `yaml:"-"`
    Depth         int                     `yaml:"-"`
}

type SubWorkspace struct {
    Name          string                  `yaml:"name"`
    Type          string                  `yaml:"type"`  // "repo-claude" or "git"
    Source        string                  `yaml:"source"`
    Branch        string                  `yaml:"branch,omitempty"`
    LoadStrategy  string                  `yaml:"load_strategy,omitempty"` // "eager" or "lazy"
    OnlyScopes    []string                `yaml:"only_scopes,omitempty"`
    CacheTTL      int                     `yaml:"cache_ttl,omitempty"`
}
```

### 1.2 Backward Compatibility

```go
// internal/config/migration.go
func MigrateConfig(cfg *Config) *Config {
    // If type is not set, determine from structure
    if cfg.Type == "" {
        if len(cfg.SubWorkspaces) > 0 && len(cfg.Repositories) > 0 {
            cfg.Type = WorkspaceTypeHybrid
        } else if len(cfg.SubWorkspaces) > 0 {
            cfg.Type = WorkspaceTypeAggregate
        } else {
            cfg.Type = WorkspaceTypeLeaf
        }
    }
    
    // Ensure version compatibility
    if cfg.Version == 2 {
        // Current version, no migration needed
        return cfg
    }
    
    // Future: handle version 3 migration
    return cfg
}
```

### 1.3 Configuration Examples

```yaml
# Leaf workspace (current behavior) - fully compatible
version: 2
workspace_name: "payment-services"
type: "leaf"

repositories:
  - name: "payment-gateway"
    url: "https://github.com/acme/payment-gateway"
    
scopes:
  core:
    repos: ["payment-gateway"]
```

```yaml
# Aggregate workspace (new)
version: 2
workspace_name: "backend-platform"
type: "aggregate"

sub_workspaces:
  - name: "payments"
    type: "repo-claude"
    source: "https://github.com/acme/payment-services"
    
  - name: "orders"
    type: "repo-claude"
    source: "https://github.com/acme/order-services"
    
scopes:
  integration:
    workspace_scopes:
      payments: ["api"]
      orders: ["api"]
```

### Tasks
- [ ] Implement WorkspaceType enum and Config extensions
- [ ] Create SubWorkspace structure with lazy loading support
- [ ] Add backward compatibility layer
- [ ] Implement config migration logic
- [ ] Write comprehensive unit tests
- [ ] Update config validation

---

## 📋 Phase 2: Tree Traversal (Week 3-4)

### Objective
Implement tree navigation and scope resolution across workspace hierarchy.

### 2.1 Path Resolution System

```go
// internal/workspace/resolver.go
type PathResolver struct {
    root    *Manager
    current *Manager
    cache   map[string]*Manager
}

func (r *PathResolver) Resolve(path string) (*Manager, *Scope, error) {
    // Path types:
    // - "scope"                  -> Local scope
    // - "./scope"                -> Explicit local scope
    // - "payments/api"           -> Child traversal
    // - "../orders/api"          -> Parent traversal
    // - "//platform/payments"    -> Absolute from root
    
    if strings.HasPrefix(path, "//") {
        return r.resolveAbsolute(path[2:])
    }
    
    if strings.HasPrefix(path, "../") {
        return r.resolveParent(path[3:])
    }
    
    if strings.HasPrefix(path, "./") {
        return r.resolveLocal(path[2:])
    }
    
    parts := strings.Split(path, "/")
    if len(parts) == 1 {
        return r.resolveLocal(path)
    }
    
    return r.resolveChild(parts)
}
```

### 2.2 Recursive Commands

```go
// cmd/repo-claude/tree.go
func newTreeCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "tree",
        Short: "Display workspace tree structure",
        RunE: func(cmd *cobra.Command, args []string) error {
            mgr, _ := manager.LoadFromCurrentDir()
            return mgr.PrintTree(os.Stdout, 0)
        },
    }
}

// internal/manager/tree.go
func (m *Manager) PrintTree(w io.Writer, depth int) error {
    indent := strings.Repeat("  ", depth)
    
    // Print current workspace
    fmt.Fprintf(w, "%s📁 %s (%s)\n", indent, m.WorkspaceName, m.Type)
    
    // Print repositories
    for _, repo := range m.Repositories {
        fmt.Fprintf(w, "%s  📄 %s\n", indent, repo.Name)
    }
    
    // Recursively print sub-workspaces
    for _, sub := range m.SubWorkspaces {
        subMgr, _ := m.LoadSubWorkspace(sub.Name)
        subMgr.PrintTree(w, depth+1)
    }
    
    return nil
}
```

### 2.3 Enhanced Commands

```bash
# Navigation commands
rc tree                               # Show tree structure
rc tree --depth 2                     # Limit depth
rc tree --format json                 # Output as JSON

# Scope operations with paths
rc start payments/api                 # Start child scope
rc start ../shared/libs               # Start sibling scope
rc start //platform/integration       # Start from root

# Recursive operations
rc pull --recursive                    # Pull entire tree
rc pull payments/*                     # Pull all in payments
rc status --recursive --depth 2       # Status to depth 2
```

### Tasks
- [ ] Implement PathResolver with all path types
- [ ] Create tree traversal algorithms
- [ ] Add recursive Manager loading
- [ ] Implement tree command
- [ ] Add --recursive flag to existing commands
- [ ] Create scope path resolution
- [ ] Add wildcard support (payments/*)
- [ ] Write integration tests

---

## 📋 Phase 3: Distributed Documentation (Week 5-6)

### Objective
Implement distributed documentation system with tree-aware aggregation.

### 3.1 Documentation Model

```go
// internal/docs/distributed.go
type DistributedDocs struct {
    Local    *DocsManager           // Current workspace docs
    Parent   *DistributedDocs       // Parent workspace docs
    Children map[string]*DistributedDocs // Child workspace docs
    Index    *DocIndex              // Aggregated index
}

type DocIndex struct {
    Entries  map[string]DocEntry    // path -> entry
    Tree     *DocTree               // Hierarchical view
    Tags     map[string][]string    // tag -> paths
    Search   *SearchIndex           // Full-text search
}

type DocEntry struct {
    Path         string             // Full path in tree
    LocalPath    string             // Path in local repo
    Workspace    string             // Workspace name
    Title        string
    Description  string
    Tags         []string
    LastModified time.Time
    Hash         string             // Content hash for caching
}
```

### 3.2 Documentation Commands

```bash
# Basic operations
rc docs list                          # Local docs
rc docs list --recursive              # All docs in tree
rc docs list payments/                # Docs in child

# Viewing
rc docs view api.md                   # View local
rc docs view payments/api.md          # View in child
rc docs view //platform/governance.md # View from root

# Search
rc docs search "authentication"       # Search all
rc docs search "api" --scope payments # Search subtree
rc docs grep "TODO"                   # Find patterns

# Aggregation
rc docs compose onboarding            # Create composite view
rc docs export --format html          # Export tree as HTML
rc docs index --rebuild               # Rebuild search index
```

### 3.3 Documentation Configuration

```yaml
# In each repo-claude.yaml
documentation:
  path: "docs"                        # Local docs directory
  
  # Inheritance
  inherit_from_parent: true           # Include parent docs in context
  
  # Composite views
  views:
    overview:
      title: "System Overview"
      compose:
        - path: "./README.md"
        - path: "payments/README.md"
        - path: "orders/README.md"
      
  # AI Context
  claude_context:
    files:
      - "docs/architecture.md"
      - "docs/api.md"
    inherit: true
```

### Tasks
- [ ] Implement DistributedDocs structure
- [ ] Create DocIndex with search capabilities
- [ ] Add documentation inheritance
- [ ] Implement composite views
- [ ] Create doc traversal commands
- [ ] Add full-text search
- [ ] Build HTML export
- [ ] Integrate with CLAUDE.md generation
- [ ] Add caching layer

---

## 📋 Phase 4: Performance Optimization (Week 7-8)

### Objective
Optimize for large-scale deployments (500+ repositories).

### 4.1 Lazy Loading

```go
// internal/workspace/loader.go
type LazyLoader struct {
    source   string
    branch   string
    loaded   bool
    manager  *Manager
    mu       sync.RWMutex
}

func (l *LazyLoader) Load() (*Manager, error) {
    l.mu.RLock()
    if l.loaded {
        defer l.mu.RUnlock()
        return l.manager, nil
    }
    l.mu.RUnlock()
    
    l.mu.Lock()
    defer l.mu.Unlock()
    
    // Double-check after acquiring write lock
    if l.loaded {
        return l.manager, nil
    }
    
    // Load workspace
    mgr, err := loadWorkspace(l.source, l.branch)
    if err != nil {
        return nil, err
    }
    
    l.manager = mgr
    l.loaded = true
    return mgr, nil
}
```

### 4.2 Caching Strategy

```go
// internal/cache/workspace.go
type WorkspaceCache struct {
    entries map[string]*CacheEntry
    mu      sync.RWMutex
    maxSize int
    ttl     time.Duration
}

type CacheEntry struct {
    Manager     *Manager
    LoadedAt    time.Time
    AccessedAt  time.Time
    Size        int64
    Hash        string
}

func (c *WorkspaceCache) Get(key string) (*Manager, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    entry, ok := c.entries[key]
    if !ok {
        return nil, false
    }
    
    // Check TTL
    if time.Since(entry.LoadedAt) > c.ttl {
        return nil, false
    }
    
    entry.AccessedAt = time.Now()
    return entry.Manager, true
}
```

### 4.3 Parallel Operations

```go
// internal/manager/parallel.go
func (m *Manager) PullRecursive(maxDepth int) error {
    var wg sync.WaitGroup
    errors := make(chan error, 100)
    
    // Local repos in parallel
    for _, repo := range m.Repositories {
        wg.Add(1)
        go func(r Repository) {
            defer wg.Done()
            if err := m.pullRepo(r); err != nil {
                errors <- err
            }
        }(repo)
    }
    
    // Sub-workspaces in parallel
    if maxDepth > 0 {
        for _, sub := range m.SubWorkspaces {
            wg.Add(1)
            go func(s SubWorkspace) {
                defer wg.Done()
                subMgr, _ := m.LoadSubWorkspace(s.Name)
                if err := subMgr.PullRecursive(maxDepth - 1); err != nil {
                    errors <- err
                }
            }(sub)
        }
    }
    
    wg.Wait()
    close(errors)
    
    // Collect errors
    var allErrors []error
    for err := range errors {
        allErrors = append(allErrors, err)
    }
    
    if len(allErrors) > 0 {
        return fmt.Errorf("pull failed: %v", allErrors)
    }
    
    return nil
}
```

### Tasks
- [ ] Implement LazyLoader with thread safety
- [ ] Create WorkspaceCache with TTL
- [ ] Add parallel pull/push operations
- [ ] Implement progress reporting
- [ ] Add connection pooling
- [ ] Create batch operations
- [ ] Add resource limits
- [ ] Performance benchmarks

---

## 📋 Phase 5: Enterprise Features (Week 9-10)

### Objective
Add enterprise-scale features for large organizations.

### 5.1 Organization Boundaries

```yaml
# Enterprise configuration
version: 3
workspace_name: "acme-platform"
type: "aggregate"

# Organization-level settings
organizations:
  payments:
    owner: "payments-team"
    workspace: "payments-platform"
    policies:
      branch: "main"
      require_pr: true
      
  logistics:
    owner: "logistics-team"  
    workspace: "logistics-platform"
    policies:
      branch: "develop"
      
# Cross-org scopes
scopes:
  platform-integration:
    cross_org: true
    requires_approval: ["payments-team", "logistics-team"]
    workspace_scopes:
      payments: ["api"]
      logistics: ["api"]
```

### 5.2 Access Control

```go
// internal/auth/rbac.go
type AccessControl struct {
    Roles       map[string]Role
    Permissions map[string]Permission
}

type Role struct {
    Name        string
    Permissions []string
    Workspaces  []string  // Workspace patterns
}

func (ac *AccessControl) CanAccess(user User, workspace string, action string) bool {
    role := ac.Roles[user.Role]
    
    // Check workspace pattern
    for _, pattern := range role.Workspaces {
        if matches(pattern, workspace) {
            // Check permission
            return role.HasPermission(action)
        }
    }
    
    return false
}
```

### 5.3 Metrics & Monitoring

```go
// internal/metrics/collector.go
type MetricsCollector struct {
    RepoCount       int
    WorkspaceCount  int
    MaxDepth        int
    TotalSize       int64
    LoadTime        time.Duration
    Operations      map[string]OpMetrics
}

func (m *Manager) CollectMetrics() *MetricsCollector {
    metrics := &MetricsCollector{
        Operations: make(map[string]OpMetrics),
    }
    
    // Recursive collection
    m.collectMetricsRecursive(metrics, 0)
    
    return metrics
}
```

### Tasks
- [ ] Design organization boundary system
- [ ] Implement RBAC framework
- [ ] Add policy enforcement
- [ ] Create metrics collection
- [ ] Add audit logging
- [ ] Implement approval workflows
- [ ] Create admin commands
- [ ] Add compliance features

---

## 📊 Success Metrics

### Performance Targets
- Load 500+ repo tree: < 5 seconds
- Navigate to any scope: < 100ms  
- Documentation search: < 500ms
- Full tree pull: < 2 minutes

### Quality Metrics
- Zero breaking changes for v2 configs
- 90% test coverage for new code
- Documentation coverage > 95%

---

## 🔄 Implementation Timeline

### Week 1-2: Foundation
- Implement data model
- Add backward compatibility
- Release v0.7.0-alpha

### Week 3-4: Tree Traversal  
- Implement navigation
- Add recursive commands
- Release v0.7.0-beta

### Week 5-6: Documentation
- Distributed docs system
- Search and aggregation
- Release v0.7.0-rc1

### Week 7-8: Performance
- Optimization pass
- Load testing
- Release v0.7.0-rc2

### Week 9-10: Enterprise
- Advanced features
- Release v0.7.0

### Week 11-12: Stabilization
- Bug fixes
- Documentation
- Training materials
- Release v1.0.0

---

## 🎯 Key Decisions

1. **Version 3 Config**: New features require version 3, but version 2 remains fully supported
2. **Git-Native**: Each workspace is a git repository for versioning
3. **Lazy by Default**: Sub-workspaces load on-demand unless specified
4. **Documentation**: Lives in git repos, not separate system
5. **Caching**: Aggressive caching with TTL for performance
6. **Compatibility**: Zero breaking changes, gradual adoption path

---

## ⚠️ Risk Mitigation

### Technical Risks
- **Circular Dependencies**: Detect at load time, fail fast
- **Performance Degradation**: Implement caching, lazy loading
- **State Corruption**: Atomic operations, validation

### Organizational Risks
- **Adoption Resistance**: Maintain full backward compatibility
- **Training Needs**: Comprehensive documentation, examples

---

## 📝 Implementation Progress Tracking

### Phase 1: Foundation
- [ ] WorkspaceType enum implementation
- [ ] Config struct extensions
- [ ] SubWorkspace structure
- [ ] Backward compatibility layer
- [ ] Config migration logic
- [ ] Unit tests for recursive structures

### Phase 2: Tree Traversal
- [ ] PathResolver implementation
- [ ] Tree traversal algorithms
- [ ] Recursive Manager loading
- [ ] Tree command
- [ ] Recursive flags for commands
- [ ] Scope path resolution
- [ ] Wildcard support
- [ ] Integration tests

### Phase 3: Distributed Documentation
- [ ] DistributedDocs structure
- [ ] DocIndex with search
- [ ] Documentation inheritance
- [ ] Composite views
- [ ] Doc traversal commands
- [ ] Full-text search
- [ ] HTML export
- [ ] CLAUDE.md integration
- [ ] Caching layer

### Phase 4: Performance Optimization
- [ ] LazyLoader implementation
- [ ] WorkspaceCache with TTL
- [ ] Parallel operations
- [ ] Progress reporting
- [ ] Connection pooling
- [ ] Batch operations
- [ ] Resource limits
- [ ] Performance benchmarks

### Phase 5: Enterprise Features
- [ ] Organization boundaries
- [ ] RBAC framework
- [ ] Policy enforcement
- [ ] Metrics collection
- [ ] Audit logging
- [ ] Approval workflows
- [ ] Admin commands
- [ ] Compliance features

---

This plan provides a systematic approach to evolving repo-claude into an enterprise-scale tool while maintaining its simplicity and effectiveness at smaller scales.