# ADR-001: Simplify Git Management by Removing Android Repo Tool Dependency

Date: 2024-08-15

## Status

Accepted

## Context

Repo-Claude was initially built using the Android Repo tool, a repository management system created by Google for managing many Git repositories (used for Android and ChromiumOS development). While Repo is a powerful and battle-tested tool, it introduced significant complexity for our use case:

1. **Complex initialization flow**: Users had to understand manifest repositories, XML generation, and the relationship between `repo-claude.yaml` and Repo's manifest format
2. **Python dependency**: Repo downloads ~30MB of Python scripts per workspace
3. **Unnecessary features**: Repo's advanced features (branch management, code review upload, cherry-picking) aren't needed for AI agent orchestration
4. **Confusing UX**: Users questioned why they needed to create manifests when they already had `repo-claude.yaml`
5. **Extra abstraction layer**: The translation between our config and Repo's manifest added complexity without clear benefit
6. **Poor AI context flow**: AI agents need global documentation at the workspace root to understand the system context. Repo's manifest-only repositories don't support this pattern well, as they're designed purely for repository structure, not documentation

## Decision

We decided to remove the Android Repo tool dependency and implement direct Git operations using Go's `os/exec` to call native git commands.

The new implementation:
- Uses a single configuration file (`repo-claude.yaml`)
- Performs direct `git clone` and `git pull` operations
- Maintains the same user-facing commands
- Provides parallel operations for performance
- Keeps the multi-agent orchestration features intact

## Consequences

### Positive
- **Dramatically simpler user experience**: No manifest confusion, just one config file
- **Faster initialization**: No Python runtime download, no `.repo` metadata creation
- **Easier to understand**: Direct git operations are familiar to all developers
- **Better error messages**: Direct git output instead of Repo's abstraction
- **Smaller footprint**: No `.repo` directory with duplicated git objects
- **Easier maintenance**: Less code, fewer edge cases to handle

### Negative
- **Loss of some advanced features**: No longer have `repo upload`, `repo cherry-pick`, etc. (but these weren't used)
- **Custom implementation**: We maintain our own git orchestration code instead of using Google's
- **No manifest includes**: Can't use Repo's manifest include features (but we didn't need them)

### Neutral
- **Performance**: Similar performance - both approaches use parallel git operations
- **Reliability**: Git commands are stable, but we lose Repo's battle-tested error handling
- **Compatibility**: Can no longer interoperate with existing Repo workspaces

## Implementation

The new architecture consists of:

1. **GitManager** (`internal/git/git.go`): Handles all git operations
   - Parallel clone/sync operations
   - Status checking across repositories
   - ForAll command execution

2. **Simplified Manager** (`internal/manager/manager.go`): 
   - No manifest generation
   - Direct configuration loading
   - Cleaner initialization flow

3. **Commands remain the same**:
   - `init`: Initialize or resume workspace
   - `sync`: Pull latest changes
   - `status`: Show repository and agent status
   - `forall`: Run commands across repositories
   - `start/stop`: Manage agents

## Lessons Learned

While the Android Repo tool is excellent for its intended use case (managing massive projects like Android), it was overkill for repo-claude. The lesson is to carefully evaluate whether powerful tools are actually needed before adopting them. Sometimes a simpler, purpose-built solution is better than a general-purpose tool.

## References

- [Android Repo Tool](https://gerrit.googlesource.com/git-repo/)
- [Original repo-claude design](https://github.com/yourusername/repo-claude)
- [Git commands documentation](https://git-scm.com/docs)