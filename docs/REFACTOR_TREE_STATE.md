# Tree State Simplification Refactoring Plan

## Problem Statement

The current v3 tree implementation mixes logical tree structure with filesystem implementation details, causing:
1. **Path sync issues**: Node's `FullPath` in state doesn't match actual filesystem structure
2. **Complex state management**: State contains absolute filesystem paths
3. **Navigation failures**: `/level1/level2` can't resolve to `repos/level1/repos/level2/`

## Current State (Wrong)

### State Contains Filesystem Paths
```json
{
  "nodes": {
    "/level1/level2": {
      "full_path": "/tmp/workspace/repos/level1/level2",  // WRONG - has filesystem path
      "path": "/level1/level2"
    }
  }
}
```

## Target Solution

### Core Principle
**Tree model as pure logical structure - filesystem is just an implementation detail**

### 1. New Clean State Structure
```json
{
  "current_node": "/level1/level2",
  "nodes": {
    "/": {
      "type": "root",
      "children": ["level1", "shared"]
    },
    "/level1": {
      "type": "repo",
      "url": "https://github.com/org/level1.git",
      "lazy": false,
      "state": "cloned",
      "children": ["level2"]
    },
    "/level1/level2": {
      "type": "repo",
      "url": "https://github.com/org/level2.git", 
      "lazy": false,
      "state": "cloned",
      "children": []
    }
  }
}
```

**State contains ONLY:**
- Logical tree structure (parent-child relationships)
- Repository metadata (URL, lazy flag, clone state)
- Current navigation position

**State does NOT contain:**
- Filesystem paths
- Absolute paths
- Directory names

### 2. Filesystem Path Computation

```go
// ComputeFilesystemPath derives filesystem path from logical path
// This is the ONLY place that knows about the repos/ directory pattern
func (m *Manager) ComputeFilesystemPath(logicalPath string) string {
    if logicalPath == "/" {
        return m.workspacePath
    }
    
    // Split path: /level1/level2/level3 -> [level1, level2, level3]
    parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")
    
    // Build filesystem path with repos/ subdirectories
    // workspace/repos/level1/repos/level2/repos/level3
    fsPath := filepath.Join(m.workspacePath, "repos")
    for i, part := range parts {
        fsPath = filepath.Join(fsPath, part)
        // Add repos/ before next level (except last)
        if i < len(parts)-1 {
            fsPath = filepath.Join(fsPath, "repos")
        }
    }
    
    return fsPath
}
```

### 3. Simplified Node Structure

```go
type Node struct {
    // Logical structure only
    Name     string           `json:"name"`
    Type     string           `json:"type"`     // "root" or "repo"
    Children []string         `json:"children"` // Just names, not paths
    
    // Repository metadata (only for type="repo")
    URL      string           `json:"url,omitempty"`
    Lazy     bool             `json:"lazy,omitempty"`
    State    string           `json:"state,omitempty"` // "missing", "cloned", "modified"
}

// No filesystem paths stored anywhere in Node
```

### 4. Manager Structure Changes

```go
type Manager struct {
    workspacePath string             // Base workspace directory
    nodes         map[string]*Node   // Map from logical path to node
    currentPath   string             // Current logical path
    
    // NO stored filesystem paths
}

// Navigation uses logical paths
func (m *Manager) UseNode(logicalPath string) error {
    node := m.nodes[logicalPath]
    if node == nil {
        return fmt.Errorf("node not found: %s", logicalPath)
    }
    
    // Compute filesystem path only when needed
    fsPath := m.ComputeFilesystemPath(logicalPath)
    if err := os.Chdir(fsPath); err != nil {
        return fmt.Errorf("cannot navigate to %s: %w", logicalPath, err)
    }
    
    m.currentPath = logicalPath
    return nil
}
```

## Implementation Steps

### 1. Create New Types
```go
// internal/tree/types.go

type TreeNode struct {
    Name     string            `json:"name"`
    Type     NodeType          `json:"type"`
    Children []string          `json:"children"`
    
    // Repository fields (only for type="repo")
    URL      string            `json:"url,omitempty"`
    Lazy     bool              `json:"lazy,omitempty"`
    State    RepoState         `json:"state,omitempty"`
}

type TreeState struct {
    CurrentPath string                     `json:"current_path"`
    Nodes       map[string]*TreeNode       `json:"nodes"`
}

type NodeType string
const (
    NodeTypeRoot NodeType = "root"
    NodeTypeRepo NodeType = "repo"
)

type RepoState string
const (
    RepoStateMissing  RepoState = "missing"
    RepoStateCloned   RepoState = "cloned"
    RepoStateModified RepoState = "modified"
)
```

### 2. Replace Manager Implementation
```go
// internal/tree/manager.go

type Manager struct {
    workspacePath string
    state         *TreeState
    gitCmd        GitCommand
}

func (m *Manager) ComputeFilesystemPath(logicalPath string) string {
    if logicalPath == "/" {
        return filepath.Join(m.workspacePath, "repos")
    }
    
    parts := strings.Split(strings.TrimPrefix(logicalPath, "/"), "/")
    fsPath := filepath.Join(m.workspacePath, "repos")
    
    for i, part := range parts {
        fsPath = filepath.Join(fsPath, part)
        if i < len(parts)-1 {
            fsPath = filepath.Join(fsPath, "repos")
        }
    }
    
    return fsPath
}

func (m *Manager) AddRepo(parentPath, name, url string, lazy bool) error {
    // Update tree state
    parent := m.state.Nodes[parentPath]
    if parent == nil {
        return fmt.Errorf("parent node not found: %s", parentPath)
    }
    
    childPath := path.Join(parentPath, name)
    m.state.Nodes[childPath] = &TreeNode{
        Name:     name,
        Type:     NodeTypeRepo,
        URL:      url,
        Lazy:     lazy,
        State:    RepoStateMissing,
        Children: []string{},
    }
    
    parent.Children = append(parent.Children, name)
    
    // Clone if not lazy
    if !lazy {
        fsPath := m.ComputeFilesystemPath(childPath)
        if err := m.cloneToPath(url, fsPath); err != nil {
            return err
        }
        m.state.Nodes[childPath].State = RepoStateCloned
    }
    
    return m.saveState()
}

func (m *Manager) UseNode(logicalPath string) error {
    node := m.state.Nodes[logicalPath]
    if node == nil {
        return fmt.Errorf("node not found: %s", logicalPath)
    }
    
    fsPath := m.ComputeFilesystemPath(logicalPath)
    
    // Auto-clone if lazy
    if node.Type == NodeTypeRepo && node.State == RepoStateMissing {
        if err := m.cloneToPath(node.URL, fsPath); err != nil {
            return err
        }
        node.State = RepoStateCloned
        m.saveState()
    }
    
    if err := os.Chdir(fsPath); err != nil {
        return err
    }
    
    m.state.CurrentPath = logicalPath
    return m.saveState()
}

func (m *Manager) cloneToPath(url, fsPath string) error {
    // Create parent directory
    if err := os.MkdirAll(filepath.Dir(fsPath), 0755); err != nil {
        return err
    }
    
    return m.gitCmd.Clone(url, fsPath)
}
```

### 3. Update Display Functions
```go
// internal/tree/display.go

func (m *Manager) DisplayTree() string {
    return m.displayNode("/", "", true, 0)
}

func (m *Manager) displayNode(logicalPath, prefix string, isLast bool, depth int) string {
    node := m.state.Nodes[logicalPath]
    if node == nil {
        return ""
    }
    
    var sb strings.Builder
    
    // Display current node
    symbol := "├─"
    if isLast {
        symbol = "└─"
    }
    
    icon := "📁"
    if node.Type == NodeTypeRepo && node.State == RepoStateMissing {
        icon = "💤"
    }
    
    sb.WriteString(fmt.Sprintf("%s%s %s %s\n", prefix, symbol, icon, node.Name))
    
    // Display children
    childPrefix := prefix
    if isLast {
        childPrefix += "    "
    } else {
        childPrefix += "│   "
    }
    
    for i, childName := range node.Children {
        childPath := path.Join(logicalPath, childName)
        isLastChild := i == len(node.Children)-1
        sb.WriteString(m.displayNode(childPath, childPrefix, isLastChild, depth+1))
    }
    
    return sb.String()
}
```

### 4. State Persistence
```go
// internal/tree/state.go

func (m *Manager) saveState() error {
    data, err := json.MarshalIndent(m.state, "", "  ")
    if err != nil {
        return err
    }
    
    statePath := filepath.Join(m.workspacePath, ".repo-claude-tree.json")
    return os.WriteFile(statePath, data, 0644)
}

func (m *Manager) loadState() error {
    statePath := filepath.Join(m.workspacePath, ".repo-claude-tree.json")
    data, err := os.ReadFile(statePath)
    if err != nil {
        if os.IsNotExist(err) {
            // Initialize new state
            m.state = &TreeState{
                CurrentPath: "/",
                Nodes: map[string]*TreeNode{
                    "/": {
                        Name:     "root",
                        Type:     NodeTypeRoot,
                        Children: []string{},
                    },
                },
            }
            return nil
        }
        return err
    }
    
    return json.Unmarshal(data, &m.state)
}
```

### Files to Delete/Refactor Completely

1. **internal/tree/types.go** - Replace Node struct completely
2. **internal/tree/manager.go** - Rewrite core logic
3. **internal/tree/state.go** - New state management
4. **internal/tree/display.go** - Simplified display logic
5. **internal/tree/manager_config.go** - Remove or simplify
6. **internal/config/config_v3_tree.go** - Simplify config structure

## Testing Strategy

### Unit Tests
```go
func TestComputeFilesystemPath(t *testing.T) {
    tests := []struct {
        logical  string
        expected string
    }{
        {"/", "workspace/repos"},
        {"/level1", "workspace/repos/level1"},
        {"/level1/level2", "workspace/repos/level1/repos/level2"},
        {"/a/b/c", "workspace/repos/a/repos/b/repos/c"},
    }
    // ...
}
```

### Integration Test
```go
func TestTreeNavigation(t *testing.T) {
    // Create workspace
    // Add repos at different levels
    // Navigate and verify filesystem paths
    // Verify state contains no filesystem paths
}
```

## Success Criteria

- [ ] State file contains only logical paths and repo metadata
- [ ] Navigation works correctly at all depths
- [ ] Filesystem follows `repos/node/repos/child/` pattern
- [ ] No absolute paths anywhere in the codebase
- [ ] Tree display shows correct hierarchy
- [ ] All operations (add, remove, clone) work correctly