# CLAUDE.md File Structure and Behavior

## Overview

When `rc start` is executed on a scope, three levels of CLAUDE.md files are created/used to provide context to Claude Code:

1. **Root CLAUDE.md** - Project-level (under git control)
2. **Scope CLAUDE.md** - Scope-level (NOT under git control)
3. **Repository CLAUDE.md** - Repository-level (under git control in each repo)

## File Creation and Persistence

### Root Level CLAUDE.md
- **Location**: `<project-root>/CLAUDE.md`
- **Git Status**: ✅ Under git control (persisted)
- **Created**: During `rc init`
- **Template**: Hard-coded in `internal/manager/manager.go::createRootCLAUDE()`

### Scope Level CLAUDE.md
- **Location**: `workspaces/<scope-name>/CLAUDE.md`
- **Git Status**: ❌ NOT under git control (temporary)
- **Created**: During `rc scope create` or `rc start`
- **Template**: Hard-coded in `internal/scope/scope.go::createScopeCLAUDE()`
- **Includes .gitignore**: Automatically creates `.gitignore` to exclude:
  - `CLAUDE.md`
  - `shared-memory.md`
  - `.scope-meta.json`
  - `.gitignore`

### Repository Level CLAUDE.md
- **Location**: `workspaces/<scope-name>/<repo-name>/CLAUDE.md`
- **Git Status**: ✅ Under git control (in each repository)
- **Created**: During repository clone within scope
- **Template**: Hard-coded in `internal/scope/scope.go` (lines 83-130)

## Link Structure

The CLAUDE.md files properly link to each other:

### Repository CLAUDE.md Links:
```markdown
- [Root CLAUDE.md](../../CLAUDE.md) - Project-wide instructions
- [Scope CLAUDE.md](../CLAUDE.md) - Scope-specific context (not in git)
- [Shared Memory](../shared-memory.md) - Inter-scope coordination
```

### Scope CLAUDE.md Links:
```markdown
- [Root CLAUDE.md](<project-path>/CLAUDE.md) - Main project instructions
- [Shared Memory](./shared-memory.md) - Coordination with other scopes
```

## Content Templates

All CLAUDE.md templates are **hard-coded** in the Go source code, not loaded from external template files.

### Root CLAUDE.md Template Content
Located in: `internal/manager/manager.go::createRootCLAUDE()`

Key sections:
- Project structure explanation (three-level architecture)
- Available scopes reference
- Documentation structure rules
- Command reference

### Scope CLAUDE.md Template Content
Located in: `internal/scope/scope.go::createScopeCLAUDE()`

Key sections:
- Scope identification and details
- Three-level structure explanation
- Current working directory
- Repositories in scope
- Documentation guidelines (CRITICAL: where to store docs)
- Available commands specific to this scope

### Repository CLAUDE.md Template Content
Located in: `internal/scope/scope.go` (lines 83-130)

Key sections:
- Repository and scope identification
- Three-level structure
- Links to parent CLAUDE.md files
- Documentation location rules
- Other repositories in the same scope
- Working directory notes

## Documentation Guidelines Embedded in Templates

All CLAUDE.md files emphasize the **critical documentation location rules**:

1. **Cross-repository documentation**: MUST be stored in `<root>/docs/scopes/<scope-name>/`
   - Why: Scope directory is temporary and not under git
   - This ensures documentation persists in git

2. **Repository-specific docs**: Store in each repository's `docs/` folder
   - Standard repository documentation

3. **Global project docs**: Store in `<root>/docs/global/`
   - Project-wide documentation

## Why This Structure?

### Scope CLAUDE.md Not in Git
- **Temporary Nature**: Scopes are ephemeral workspaces
- **Isolation**: Each scope instance is independent
- **Clean Git History**: Prevents pollution of repositories with workspace files
- **Dynamic Content**: Contains runtime-specific information (paths, timestamps)

### Repository CLAUDE.md in Git
- **Context Preservation**: Provides AI context even when cloned elsewhere
- **Repository-Specific**: Contains information about the repository's role
- **Stable Content**: Links and structure remain consistent

## Implementation Details

### File Creation Flow
```
rc start <scope>
  ├── Load scope metadata
  ├── Create scope directory if needed
  ├── Call createScopeCLAUDE() → Creates scope CLAUDE.md + .gitignore
  ├── Clone repositories (if needed)
  │   └── For each repo: Creates repository CLAUDE.md
  └── Launch Claude Code session
```

### Template Customization

Currently, all templates are hard-coded. To modify templates, you must:

1. Edit the Go source code:
   - Root: `internal/manager/manager.go::createRootCLAUDE()`
   - Scope: `internal/scope/scope.go::createScopeCLAUDE()`
   - Repository: `internal/scope/scope.go` (clone function)

2. Rebuild the binary:
   ```bash
   make build
   ```

### Future Enhancement Possibilities

1. **External Templates**: Load from `.claude-templates/` directory
2. **Template Variables**: Support custom variables in templates
3. **Per-Scope Customization**: Allow scope-specific template overrides
4. **Markdown Includes**: Support including external markdown files

## Testing CLAUDE.md Creation

```bash
# Create a new scope
rc scope create test-scope --type ephemeral --repos "wms-core,shared-libs"

# Verify files created
ls workspaces/test-scope/
# Should show: CLAUDE.md, .gitignore, .scope-meta.json

# Check scope CLAUDE.md is not tracked by git
cd workspaces/test-scope/wms-core
git status
# CLAUDE.md should appear as untracked (repository level)
# ../CLAUDE.md should not appear (scope level, in .gitignore)

# Verify links work
cat workspaces/test-scope/wms-core/CLAUDE.md | grep "Root CLAUDE.md"
# Should show proper relative path to root
```

## Best Practices

1. **Don't Modify Scope CLAUDE.md**: It's regenerated on each scope start
2. **Document in Proper Location**: Use `docs/scopes/<name>/` for persistent documentation
3. **Keep Templates Informative**: Include clear guidance about documentation structure
4. **Use Relative Links**: Ensure links work regardless of absolute paths
5. **Version Control**: Only root and repository CLAUDE.md files should be in git