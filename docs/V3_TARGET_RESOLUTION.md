# V3 Target Tree Resolution Design

## Problem Statement

In v3's tree-based architecture, every command needs to determine which node(s) to operate on. We need a consistent, predictable resolution strategy that:
1. Is intuitive for users
2. Provides clear feedback about what will be affected
3. Supports both explicit and implicit targeting
4. Works seamlessly with tree navigation

## Resolution Priority Order

### Proposed Strategy: CWD-First with Explicit Override

```
Priority Order (highest to lowest):
1. Explicit path argument (if provided)
2. Current Working Directory (CWD) mapping
3. Stored "current" node (from `rc use`)
4. Root node (fallback)
```

### Decision Matrix

| Scenario | Command | Resolution | Rationale |
|----------|---------|------------|-----------|
| User in `/repos/team/backend/payment` | `rc pull` | `team/backend/payment` | CWD maps directly |
| User in `/home/user` | `rc pull` | Stored current or root | Outside workspace |
| User anywhere | `rc pull team/frontend` | `team/frontend` | Explicit wins |
| After `rc use team/backend` | `rc pull` | `team/backend` OR CWD | **DECISION NEEDED** |

## Key Design Decision: CWD vs Use State

### Option A: CWD-First (Recommended) ✓
```bash
cd repos/team/backend
rc pull                    # Always pulls backend (CWD-based)
rc use team/frontend       # Changes stored state
rc pull                    # STILL pulls backend (CWD wins)
cd ../frontend
rc pull                    # Now pulls frontend (CWD changed)
```

**Pros:**
- Predictable: Location determines behavior
- Natural for shell users
- No hidden state confusion
- Works with standard shell navigation

**Cons:**
- `rc use` becomes less powerful
- Must cd to change context

### Option B: Use-State-First
```bash
cd repos/team/backend
rc pull                    # Pulls backend (CWD-based)
rc use team/frontend       # Changes context
rc pull                    # Pulls frontend (use state wins)
pwd                        # Still in backend directory!
```

**Pros:**
- `rc use` is powerful context switcher
- Can work on different nodes without cd

**Cons:**
- Confusing: CWD vs operational context mismatch
- Hidden state affects commands
- Violates principle of least surprise

## Recommended Implementation

### 1. Resolution Algorithm
```go
func (m *TreeManager) ResolveTarget(explicitPath string) (*Node, error) {
    // 1. Explicit path always wins
    if explicitPath != "" {
        return m.findNode(explicitPath)
    }
    
    // 2. Try CWD mapping
    cwd, _ := os.Getwd()
    if node := m.mapCWDToNode(cwd); node != nil {
        return node, nil
    }
    
    // 3. Use stored current (only if outside workspace)
    if m.currentNode != nil {
        return m.currentNode, nil
    }
    
    // 4. Default to root
    return m.rootNode, nil
}
```

### 2. Feedback Display (CRITICAL)
Every command MUST show what it's operating on:

```bash
$ rc pull
🎯 Target: team/backend/payment (from CWD)
Pulling 3 repositories...
✓ payment-service
✓ shared-libs  
✓ payment-docs

$ rc pull --recursive
🎯 Target: team/backend (from CWD)
🔄 Recursive: will pull 12 repositories in 4 nodes
Proceed? [Y/n]
```

### 3. The `use` Command Role
With CWD-first approach, `rc use` becomes:
1. **Navigator**: Changes both CWD and stored state
2. **Lazy-loader**: Triggers auto-clone of lazy repos
3. **Context shower**: Displays current position

```bash
$ rc use team/frontend
📍 Navigated to: team/frontend
📂 Changed directory to: /workspace/repos/team/frontend
🔄 Auto-cloning 2 lazy repositories...
✓ component-lib (cloned)
✓ design-system (cloned)
```

### 4. Special Cases

#### Outside Workspace
```bash
$ cd /home/user
$ rc pull
🎯 Target: team/backend (from stored current)
⚠️  Not in workspace directory, using last position
```

#### Repository Subdirectories
```bash
$ cd repos/team/backend/payment-service/src/handlers
$ rc pull
🎯 Target: team/backend/payment-service (mapped from CWD)
ℹ️  Resolved from: .../payment-service/src/handlers → payment-service
```

#### Root Directory
```bash
$ cd repos/
$ rc pull
🎯 Target: / (root - from CWD)
⚠️  This will affect root-level repositories only
    Use --recursive to include all children
```

## Implementation Plan

### Phase 1: Core Resolution (Week 1)
- [ ] Implement `ResolveTarget` function
- [ ] Add CWD to node mapping
- [ ] Store current node in state file
- [ ] Add target display to all commands

### Phase 2: Navigation Enhancement (Week 1-2)
- [ ] Make `rc use` change CWD
- [ ] Implement lazy loading on navigation
- [ ] Add breadcrumb display
- [ ] Handle relative paths

### Phase 3: Feedback System (Week 2)
- [ ] Standardize target display format
- [ ] Add confirmation for recursive operations
- [ ] Show affected repository count
- [ ] Implement --dry-run flag

### Phase 4: Edge Cases (Week 2-3)
- [ ] Handle symlinks properly
- [ ] Support repository subdirectories
- [ ] Add ../ and ./ path resolution
- [ ] Implement path shortcuts (-, ~, @)

## Command Examples with Target Resolution

```bash
# Explicit path (Priority 1)
rc pull team/backend          # 🎯 Target: team/backend (explicit)

# CWD-based (Priority 2)  
cd repos/team/frontend
rc pull                       # 🎯 Target: team/frontend (from CWD)

# Stored current (Priority 3)
cd /tmp
rc pull                       # 🎯 Target: team/backend (from stored current)

# Root fallback (Priority 4)
cd /tmp
rc current --clear            # Clear stored current
rc pull                       # 🎯 Target: / (fallback to root)
```

## Migration from v2 Scopes

| v2 Scope Concept | v3 Tree Equivalent |
|------------------|-------------------|
| `--scope backend` | `rc pull team/backend` or `cd repos/team/backend && rc pull` |
| Current scope | CWD-mapped node or stored current |
| Scope isolation | Natural directory isolation |
| Scope metadata | Node metadata in tree |

## Summary

**Decision: CWD-First Resolution**
- CWD mapping has highest priority (after explicit paths)
- `rc use` changes both CWD and stored state  
- Always display target clearly
- Stored state only used when outside workspace
- Predictable, shell-friendly behavior