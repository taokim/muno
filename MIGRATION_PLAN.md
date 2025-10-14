# MUNO Migration Plan: From Stateful to Stateless Architecture

## Overview
This document outlines the complete migration from MUNO's current stateful architecture (using `use` command with internal state) to a stateless architecture (using shell `mcd` function with filesystem as state).

## Current Architecture (Stateful)
- **State Management**: `.muno-state.json` tracks current position
- **Navigation**: `muno use <path>` updates internal state
- **Tool Execution**: `muno claude`, `muno gemini` launch from state position
- **Operations**: Commands use state to determine target

## Target Architecture (Stateless)
- **State Management**: Terminal's pwd IS the state
- **Navigation**: `mcd <path>` changes actual directory
- **Tool Execution**: Run tools directly (`claude`, `gemini`)
- **Operations**: Commands use pwd to determine target

## Implementation Phases

### ✅ Phase 1: Add New Commands (Completed)
**Status**: DONE - Non-breaking additions alongside existing functionality

1. **`path` command** - Virtual-to-physical path resolution
   - `muno path [target]` - Resolve path
   - `--ensure` flag - Auto-clone lazy repos
   - `--relative` flag - Show tree position

2. **`shell-init` command** - Shell integration setup
   - `muno shell-init` - Generate shell function
   - `--cmd-name` - Custom function name
   - `--check` - Test availability
   - `--install` - Auto-install to shell config

### Phase 2: Update Commands for Stateless Operation
**Goal**: Make all commands work with pwd instead of state

#### Commands to Update:
1. **`status`** 
   - Current: Uses state from `.muno-state.json`
   - Target: Use pwd to determine current location
   - Changes: Update `StatusNode()` to accept pwd context

2. **`list`**
   - Current: Lists children of state position
   - Target: List children of pwd location
   - Changes: Resolve pwd to tree position first

3. **`tree`**
   - Current: Highlights state position
   - Target: Highlight pwd position
   - Changes: Compute tree position from pwd

4. **`pull` / `push` / `commit`**
   - Current: Operate on state position
   - Target: Operate on pwd location
   - Changes: Use pwd for default target

5. **`add` / `remove`**
   - Current: Add to state position
   - Target: Add to pwd location
   - Changes: Resolve pwd as parent path

6. **`clone`**
   - Current: Clone at state position
   - Target: Clone at pwd location
   - Changes: Use pwd for target resolution

### Phase 3: Remove State Dependencies
**Goal**: Eliminate all state management code

1. **Remove from Manager**:
   - Delete `UseNode()` method
   - Delete `UseNodeWithClone()` method
   - Remove state loading from `LoadFromCurrentDir()`
   - Delete state saving logic

2. **Remove from TreeManager**:
   - Delete `GetState()` method
   - Delete `SaveState()` method
   - Remove state fields from struct

3. **Clean up files**:
   - Delete state file handling
   - Remove `.muno-state.json` generation
   - Update config to remove state fields

### Phase 4: Remove Deprecated Commands
**Goal**: Remove commands that no longer make sense

1. **Commands to remove**:
   - `use` - Replaced by shell `mcd`
   - `current` - Just use `pwd`
   - `agent` - Run tools directly
   - `claude` - Run claude directly
   - `gemini` - Run gemini directly
   - `cursor` - Run cursor directly

2. **Update help text**:
   - Remove references to removed commands
   - Update examples to use new workflow

### Phase 5: Testing & Validation
**Goal**: Ensure everything works in stateless mode

1. **Unit Tests**:
   - Remove state-dependent tests
   - Add pwd-based operation tests
   - Test path resolution logic
   - Test shell integration

2. **Integration Tests**:
   - Test navigation workflow
   - Test lazy cloning via mcd
   - Test all git operations
   - Test tree display

3. **Regression Tests**:
   - Update regression test suite
   - Remove state-based scenarios
   - Add stateless workflows

### Phase 6: Documentation Updates
**Goal**: Guide users through the new workflow

1. **README.md**:
   - New installation instructions (include shell-init)
   - Updated command examples
   - Migration guide section

2. **CLAUDE.md**:
   - Update architecture description
   - Remove state management section
   - Document new workflow

3. **Migration Guide**:
   - Step-by-step migration for users
   - Command mapping (old → new)
   - Common scenarios

## Migration Path for Users

### Installation
```bash
# 1. Update to new version
go install github.com/taokim/muno/cmd/muno@latest

# 2. Install shell integration
muno shell-init --install

# 3. Reload shell
source ~/.bashrc  # or ~/.zshrc
```

### Workflow Changes
```bash
# Old workflow (stateful)
muno use team-backend/payment-service
muno status
muno pull
muno claude

# New workflow (stateless)
mcd team-backend/payment-service  # Navigate with auto-clone
muno status                        # Works on current directory
muno pull                          # Works on current directory
claude                             # Run directly
```

### Command Mapping
| Old Command | New Approach |
|------------|--------------|
| `muno use <path>` | `mcd <path>` |
| `muno current` | `pwd` |
| `muno claude` | `claude` |
| `muno gemini` | `gemini` |
| `muno agent <name>` | Run tool directly |

## Benefits of Migration

1. **Simplicity**: ~60% less code, no state management complexity
2. **Unix Philosophy**: pwd-based operations, standard workflow
3. **Tool Freedom**: Run any tool directly without wrappers
4. **No State Issues**: Can't have state/filesystem mismatches
5. **Better Integration**: Works with pushd/popd, shell history
6. **Easier Testing**: No state to mock or manage

## Risks & Mitigation

1. **Breaking Change**: Existing users must adapt
   - Mitigation: Clear migration guide, keep `use` with deprecation warning temporarily

2. **Shell Integration Required**: Users must install shell function
   - Mitigation: Simple one-command installation, multiple shell support

3. **Loss of Features**: No navigation history in MUNO
   - Mitigation: Shell already has history, not a real loss

## Rollout Strategy

1. **Version 0.9.0**: Add new commands (path, shell-init) alongside existing
2. **Version 0.9.5**: Deprecation warnings on old commands
3. **Version 1.0.0**: Remove old commands, fully stateless

## Completion Criteria

- [ ] All commands work with pwd instead of state
- [ ] State management code completely removed
- [ ] Shell integration works on bash/zsh/fish
- [ ] Documentation fully updated
- [ ] Migration guide published
- [ ] Tests updated and passing