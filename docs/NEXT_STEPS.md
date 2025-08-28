# Next Steps - Immediate Action Plan

## 🎯 Priority 1: Integration (This Week)

### Update Core Manager to V3
```go
// internal/manager/manager.go
// TODO: Replace Config with ConfigV3
type Manager struct {
    Config *config.ConfigV3  // Update to v3
    // ...
}
```

**Tasks:**
- [ ] Update manager.go to use ConfigV3
- [ ] Update LoadFromCurrentDir() to use AutoMigrate
- [ ] Update scope creation to handle recursive workspaces
- [ ] Test with existing commands

### Update CLI Commands
```bash
# These commands need v3 integration:
rc init      # Generate v3 config by default
rc status    # Show recursive workspace status
rc pull      # Handle recursive pulling
rc tree      # Already implemented, needs integration
```

**Tasks:**
- [ ] Update init command to generate v3 configs
- [ ] Integrate tree command into main CLI
- [ ] Update status to show workspace hierarchy
- [ ] Add --recursive flag to pull/push/commit

### Integration Tests
```go
// test/integration/v3_test.go
func TestV3WorkspaceOperations(t *testing.T) {
    // Test full workflow with v3 config
}
```

**Tasks:**
- [ ] Create integration test suite
- [ ] Test migration scenarios
- [ ] Test recursive operations
- [ ] Test lazy loading behavior

## 🎯 Priority 2: Performance Critical Path

### Implement Caching Layer
```go
// internal/cache/workspace_cache.go
type WorkspaceCache struct {
    entries map[string]*CacheEntry
    ttl     time.Duration
}
```

**Tasks:**
- [ ] Implement workspace caching
- [ ] Add TTL-based invalidation
- [ ] Cache git operations results
- [ ] Add cache warming for meta-repos

### Parallel Operations
```go
// internal/git/parallel.go
func ParallelClone(repos []Repository, workers int) error {
    // Clone multiple repos in parallel
}
```

**Tasks:**
- [ ] Implement parallel git operations
- [ ] Add progress reporting
- [ ] Handle rate limiting
- [ ] Add retry logic

## 🎯 Priority 3: User Experience

### Migration Tool
```bash
# Standalone migration command
rc migrate --backup     # Migrate v2 to v3 with backup
rc migrate --dry-run    # Show what would change
rc migrate --validate   # Validate migration
```

**Tasks:**
- [ ] Create migrate command
- [ ] Add dry-run mode
- [ ] Create backup before migration
- [ ] Validate migrated config

### Improved Error Messages
```go
// internal/errors/errors.go
type WorkspaceError struct {
    Op      string
    Path    string
    Err     error
    Hint    string  // Helpful suggestion
}
```

**Tasks:**
- [ ] Create structured error types
- [ ] Add helpful hints to errors
- [ ] Improve error context
- [ ] Add recovery suggestions

## 📊 Success Metrics

### Week 1 Goals
- [ ] V3 config integrated with main codebase
- [ ] All existing commands working with v3
- [ ] Migration tool complete
- [ ] Basic integration tests passing

### Week 2 Goals
- [ ] Caching layer implemented
- [ ] Parallel operations working
- [ ] Performance benchmarks complete
- [ ] Documentation updated

## 🚀 Quick Wins

1. **Update README**: Add v3 examples and benefits
2. **Create CHANGELOG**: Document v3 changes
3. **Add Benchmarks**: Compare v2 vs v3 performance
4. **Video Demo**: Show v3 simplification in action

## ⚠️ Blockers & Risks

### Current Blockers
- None identified

### Potential Risks
1. **Breaking Changes**: Existing users' workflows
   - *Mitigation*: Thorough migration testing
   
2. **Performance Regression**: V3 might be slower initially
   - *Mitigation*: Profile and optimize critical paths
   
3. **Complex Integration**: V3 touching many files
   - *Mitigation*: Incremental changes with tests

## 📝 Code Review Checklist

Before merging v3 integration:
- [ ] All tests passing
- [ ] Backward compatibility verified
- [ ] Migration tested with real configs
- [ ] Performance benchmarked
- [ ] Documentation updated
- [ ] Examples working
- [ ] Error handling improved

## 🎯 Definition of "Integration Complete"

- [ ] `rc init` generates v3 config
- [ ] `rc status` shows workspace tree
- [ ] `rc pull --recursive` works
- [ ] `rc tree` command integrated
- [ ] Migration tool complete and tested
- [ ] 10+ integration tests passing
- [ ] Performance equal or better than v2
- [ ] Documentation reflects v3 changes

## 📅 Timeline

| Day | Focus | Deliverable |
|-----|-------|------------|
| Mon | Manager integration | Core manager using v3 |
| Tue | CLI commands | Commands updated |
| Wed | Migration tool | Migration complete |
| Thu | Testing | Integration tests |
| Fri | Performance | Caching layer |
| Mon | Documentation | All docs updated |
| Tue | Polish | Bug fixes, improvements |

## 💻 Development Branch Strategy

```bash
# Create feature branch
git checkout -b feature/v3-integration

# Regular commits
git commit -m "feat: integrate v3 config with manager"
git commit -m "feat: add migration tool"
git commit -m "test: add v3 integration tests"

# Final merge
git checkout main
git merge --no-ff feature/v3-integration
```

## 🤝 Getting Help

- **Questions**: Open GitHub issue with `question` label
- **Bugs**: Open issue with reproduction steps
- **PRs**: Welcome! Follow contribution guidelines

## ✅ Ready to Start

1. Create feature branch
2. Start with manager integration
3. Run tests frequently
4. Update docs as you go
5. Ask for review early and often