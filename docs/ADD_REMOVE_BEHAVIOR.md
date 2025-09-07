# MUNO Add/Remove Commands - Current Behavior Analysis

## Summary
**The `add` and `remove` commands in MUNO currently DO NOT persist repositories to any storage.** The added repositories exist only in memory during the current session and are lost when the session ends.

## Test Results

### Test Scenario
```bash
# Add a repository
muno add 'file:///tmp/test-persist-repo' --name test-persist

# Result shows success
✅ Successfully added: test-persist-repo
   URL: file:///tmp/test-persist-repo
   Status: 💤 Lazy (will clone on first use)
   Location: /test-persist-repo
```

### Persistence Check
```bash
# Check muno.yaml - NOT ADDED
grep "test-persist" muno.yaml
# Result: No matches

# Check list command - NOT SHOWN
muno list | grep "test-persist"
# Result: No matches

# Check tree command - NOT SHOWN
muno tree | grep "test-persist"
# Result: No matches
```

## Technical Analysis

### Code Path
1. **Command Layer** (`cmd/muno/app.go`):
   ```go
   func (a *App) newAddCmd() *cobra.Command {
       // ...
       return mgr.AddRepoSimple(args[0], name, lazy)
   }
   ```

2. **Manager Layer** (`internal/manager/manager.go`):
   ```go
   func (m *Manager) AddRepoSimple(repoURL string, name string, lazy bool) error {
       // ...
       return m.Add(ctx, repoURL, AddOptions{Fetch: fetchMode})
   }
   ```

3. **Add Implementation**:
   ```go
   func (m *Manager) Add(ctx context.Context, repoURL string, options AddOptions) error {
       // Creates node in memory
       newNode := interfaces.NodeInfo{
           Name:       repoName,
           Repository: repoURL,
           IsLazy:     isLazy,
           IsCloned:   false,
       }
       
       // Adds to tree provider (memory only)
       if err := m.treeProvider.AddNode(current.Path, newNode); err != nil {
           return fmt.Errorf("failed to add node: %w", err)
       }
       
       // NO SAVE TO FILE OPERATION
   }
   ```

## Current Behavior

### What Happens
1. User runs `muno add <repo-url>`
2. Repository is added to in-memory tree structure
3. Success message is displayed
4. Repository can be used during current session (maybe)
5. **Repository is NOT saved to muno.yaml**
6. **Repository is lost when muno exits**

### What Should Happen
1. User runs `muno add <repo-url>`
2. Repository is added to in-memory tree structure
3. **Repository is saved to muno.yaml**
4. Success message is displayed
5. Repository persists across sessions

## Impact

### User Experience Issues
- Users think they've added repositories permanently
- Repositories disappear after restart
- Confusion about why added repos don't persist
- Manual editing of muno.yaml required for permanent changes

### Bug Classification
- **Severity**: High
- **Type**: Data Loss / Persistence Failure
- **Affected Commands**: `add`, `remove`
- **Workaround**: Manually edit muno.yaml

## Recommendations

### Short-term Fix
Add a warning message to the add/remove commands:
```
⚠️  Note: Repository added to current session only. 
   Edit muno.yaml to make permanent.
```

### Long-term Fix
Implement proper persistence:
1. After successful add, update muno.yaml
2. After successful remove, update muno.yaml
3. Add transaction support for rollback on failure
4. Consider using a separate state file for dynamic changes

## Test Coverage
The regression test suite correctly identifies this issue:
- Test: "Add updates config" - FAILS
- Test: "Remove updates config" - FAILS
- These are marked as "known issues" in the test suite

## Conclusion
The `add` and `remove` commands are essentially broken for their intended purpose. They give users the illusion of adding repositories but don't actually persist the changes. This is a critical bug that should be fixed before the next release.