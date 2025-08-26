# Scope Isolation Implementation Plan

## Overview

This document provides a detailed implementation plan for transitioning repo-claude from a shared workspace model to an isolated scope-based architecture.

## Core Architecture Changes

### 1. Directory Structure Refactoring

#### Current Structure
```
project/
├── workspace/           # Shared by all scopes
│   ├── repo1/
│   ├── repo2/
│   └── shared-memory.md
```

#### New Structure
```
project/
├── workspaces/         # Isolated scope directories
│   ├── backend-dev/    # Persistent scope
│   │   ├── .scope-meta.json
│   │   ├── auth-service/
│   │   └── order-service/
│   └── hotfix-123/     # Ephemeral scope
│       ├── .scope-meta.json
│       └── auth-service/
├── docs/               # Shared documentation
│   ├── global/
│   └── scopes/
```

### 2. Configuration Schema Updates

#### Updated repo-claude.yaml
```yaml
version: 2  # Configuration version

workspace:
  name: my-project
  isolation_mode: true  # Enable scope isolation
  base_path: workspaces
  
# Global repository definitions
repositories:
  auth-service:
    url: git@github.com:org/auth-service.git
    default_branch: main
    groups: [backend, services]
  order-service:
    url: git@github.com:org/order-service.git
    default_branch: main
    groups: [backend, services]
  payment-service:
    url: git@github.com:org/payment-service.git
    default_branch: main
    groups: [backend, services]
  frontend:
    url: git@github.com:org/frontend.git
    default_branch: develop
    groups: [ui, web]

# Scope definitions
scopes:
  # Persistent scopes
  backend:
    type: persistent
    repos: [auth-service, order-service, payment-service]
    description: "Backend services development"
    model: claude-3-sonnet
    auto_start: false
    
  frontend:
    type: persistent
    repos: [frontend, shared-ui]
    description: "Frontend development"
    model: claude-3-sonnet
    
  # Ephemeral scope templates
  hotfix:
    type: ephemeral
    repos: []  # Selected at creation time
    description: "Hotfix template"
    model: claude-3-5-sonnet-20241022
    
  feature:
    type: ephemeral
    repos: []  # Selected at creation time
    description: "Feature development template"
    model: claude-3-5-sonnet-20241022

# Documentation settings
documentation:
  path: docs
  sync_to_git: true
  structure:
    - global: "Shared across all scopes"
    - scopes: "Scope-specific documentation"
```

### 3. New Data Structures

#### Scope Metadata (.scope-meta.json)
```go
type ScopeMeta struct {
    ID          string              `json:"id"`
    Name        string              `json:"name"`
    Type        ScopeType          `json:"type"`
    CreatedAt   time.Time          `json:"created_at"`
    LastAccessed time.Time         `json:"last_accessed"`
    Repos       []RepoState        `json:"repos"`
    State       ScopeState         `json:"state"`
    SessionPID  *int               `json:"session_pid,omitempty"`
}

type RepoState struct {
    Name       string    `json:"name"`
    URL        string    `json:"url"`
    Branch     string    `json:"branch"`
    Commit     string    `json:"commit"`
    ClonedAt   time.Time `json:"cloned_at"`
    LastPulled time.Time `json:"last_pulled"`
}

type ScopeType string
const (
    ScopeTypePersistent ScopeType = "persistent"
    ScopeTypeEphemeral  ScopeType = "ephemeral"
)

type ScopeState string
const (
    ScopeStateActive   ScopeState = "active"
    ScopeStateInactive ScopeState = "inactive"
    ScopeStateArchived ScopeState = "archived"
)
```

### 4. Core Components to Refactor

#### A. Scope Manager (new)
```go
// internal/scope/manager.go
package scope

type Manager struct {
    config      *config.Config
    workspacePath string
    gitManager  git.Interface
}

func (m *Manager) Create(name string, opts CreateOptions) error
func (m *Manager) Delete(name string) error
func (m *Manager) List() ([]ScopeInfo, error)
func (m *Manager) Get(name string) (*Scope, error)
func (m *Manager) Archive(name string) error
func (m *Manager) Cleanup() error  // Remove expired ephemeral scopes
```

#### B. Scope Operations
```go
// internal/scope/scope.go
package scope

type Scope struct {
    meta     *ScopeMeta
    path     string
    repos    map[string]*Repository
    manager  *Manager
}

func (s *Scope) Clone(repoName string) error
func (s *Scope) Pull(opts PullOptions) error
func (s *Scope) Status() (*StatusReport, error)
func (s *Scope) Commit(message string) error
func (s *Scope) Push() error
func (s *Scope) SwitchBranch(branch string) error
func (s *Scope) Start(opts StartOptions) error
func (s *Scope) Stop() error
```

#### C. Command Updates
```go
// cmd/repo-claude/app.go

// Update command structure
func setupCommands(app *cli.App) {
    app.Commands = []*cli.Command{
        {
            Name: "init",
            // No change needed
        },
        {
            Name: "scope",
            Subcommands: []*cli.Command{
                {
                    Name: "create",
                    Action: createScope,
                },
                {
                    Name: "delete",
                    Action: deleteScope,
                },
                {
                    Name: "list",
                    Action: listScopes,
                },
                {
                    Name: "archive",
                    Action: archiveScope,
                },
            },
        },
        {
            Name: "start",
            ArgsUsage: "<scope>",
            Action: startScope,
            Flags: []cli.Flag{
                &cli.BoolFlag{
                    Name: "new-window",
                },
                &cli.BoolFlag{
                    Name: "pull",
                    Usage: "Pull latest changes before starting",
                },
            },
        },
        {
            Name: "pull",
            ArgsUsage: "<scope>",
            Action: pullScope,
            Flags: []cli.Flag{
                &cli.BoolFlag{
                    Name: "clone-missing",
                },
            },
        },
        {
            Name: "status",
            ArgsUsage: "<scope>",
            Action: statusScope,
        },
        {
            Name: "commit",
            ArgsUsage: "<scope>",
            Action: commitScope,
            Flags: []cli.Flag{
                &cli.StringFlag{
                    Name: "message",
                    Aliases: []string{"m"},
                    Required: true,
                },
            },
        },
        {
            Name: "push",
            ArgsUsage: "<scope>",
            Action: pushScope,
        },
        {
            Name: "docs",
            Subcommands: []*cli.Command{
                {
                    Name: "create",
                    Action: createDoc,
                },
                {
                    Name: "edit",
                    Action: editDoc,
                },
                {
                    Name: "list",
                    Action: listDocs,
                },
                {
                    Name: "sync",
                    Action: syncDocs,
                },
            },
        },
    }
}
```

### 5. Documentation System

#### Documentation Manager
```go
// internal/docs/manager.go
package docs

type Manager struct {
    basePath string
    config   *config.Config
}

func (m *Manager) CreateGlobal(name, content string) error
func (m *Manager) CreateScope(scope, name, content string) error
func (m *Manager) Edit(path string) error
func (m *Manager) List(scope string) ([]DocInfo, error)
func (m *Manager) Sync() error  // Commit to git
```

#### Documentation Structure
```
docs/
├── .git/                   # Separate git repo for docs
├── global/
│   ├── README.md
│   ├── architecture/
│   │   ├── overview.md
│   │   └── patterns.md
│   └── standards/
│       ├── coding.md
│       └── testing.md
├── scopes/
│   ├── backend/
│   │   ├── api-design.md
│   │   └── database.md
│   └── frontend/
│       ├── components.md
│       └── state-management.md
└── .scope-docs.yaml        # Metadata for scope docs
```

### 6. Migration Implementation

#### Migration Command
```go
// internal/migration/v2.go
package migration

type V2Migrator struct {
    oldConfig *config.V1Config
    newConfig *config.Config
}

func (m *V2Migrator) Migrate() error {
    // 1. Backup existing configuration
    // 2. Convert shared workspace to default scope
    // 3. Update configuration format
    // 4. Create new directory structure
    // 5. Move repositories to scope workspace
    // 6. Update state files
}
```

## Implementation Phases

### Phase 1: Foundation (Week 1)
- [x] Create ADR document
- [ ] Design new configuration schema
- [ ] Implement scope metadata structures
- [ ] Create scope manager package
- [ ] Add configuration versioning

### Phase 2: Core Logic (Week 2)
- [ ] Implement scope creation/deletion
- [ ] Add isolated git operations
- [ ] Update command structure
- [ ] Implement scope lifecycle management
- [ ] Add scope state persistence

### Phase 3: Documentation System (Week 3)
- [ ] Design documentation structure
- [ ] Implement docs manager
- [ ] Add documentation commands
- [ ] Create sync mechanism
- [ ] Add doc templates

### Phase 4: Migration & Compatibility (Week 4)
- [ ] Implement migration command
- [ ] Add backward compatibility layer
- [ ] Create migration tests
- [ ] Update user documentation
- [ ] Add deprecation warnings

### Phase 5: Testing & Polish (Week 5)
- [ ] Comprehensive unit tests
- [ ] Integration testing
- [ ] Performance optimization
- [ ] Error handling improvements
- [ ] CLI UX improvements

## Testing Strategy

### Unit Tests
```go
// internal/scope/manager_test.go
func TestScopeCreation(t *testing.T)
func TestScopeIsolation(t *testing.T)
func TestEphemeralScopeCleanup(t *testing.T)

// internal/docs/manager_test.go
func TestDocumentationHierarchy(t *testing.T)
func TestDocumentationSync(t *testing.T)
```

### Integration Tests
```go
// test/integration/scope_test.go
func TestFullScopeLifecycle(t *testing.T)
func TestParallelScopes(t *testing.T)
func TestScopeGitOperations(t *testing.T)
```

### Migration Tests
```go
// test/migration/v2_test.go
func TestV1ToV2Migration(t *testing.T)
func TestBackwardCompatibility(t *testing.T)
```

## Rollout Plan

### Alpha Release (Week 6)
- Feature flag: `--enable-isolation`
- Limited to new projects
- Gather feedback

### Beta Release (Week 7-8)
- Default for new projects
- Migration tool available
- Documentation complete

### GA Release (Week 9)
- Version 2.0
- Full migration support
- Deprecate old mode

## Risk Mitigation

### Risks and Mitigations

1. **Data Loss During Migration**
   - Mitigation: Automatic backups before migration
   - Rollback capability

2. **Disk Space Issues**
   - Mitigation: Disk space checks before cloning
   - Cleanup commands for old scopes

3. **User Confusion**
   - Mitigation: Clear migration guide
   - Interactive migration wizard
   - Comprehensive documentation

4. **Performance Degradation**
   - Mitigation: Parallel git operations
   - Lazy loading of scope data
   - Efficient file operations

## Success Metrics

- Migration success rate > 95%
- No data loss incidents
- User satisfaction (survey)
- Performance benchmarks met
- Test coverage > 80%

## Open Implementation Questions

1. **Scope Naming**: Enforce kebab-case? Allow spaces?
2. **Git Strategy**: Use worktrees where possible?
3. **Caching**: Cache git objects between scopes?
4. **Networking**: Share git credentials between scopes?
5. **UI**: Add TUI for scope management?

## Next Steps

1. Review and approve design
2. Set up feature branch
3. Begin Phase 1 implementation
4. Weekly progress reviews
5. User feedback sessions