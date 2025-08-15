# backend-agent - backend

## Agent Information
- **Repository**: backend
- **Project Groups**: core,services
- **Specialization**: API development, database design, backend services
- **Model**: claude-sonnet-4

## Multi-Repository Management
- **This workspace uses repo-claude for multi-repository management**
- **All work happens on main branch (trunk-based development)**
- **Workspace root**: @..

## Coordination
- **Shared Memory**: @../shared-memory.md

## Commands You Can Use
- `./repo-claude status` - Show status of all projects
- `./repo-claude sync` - Sync all projects from remotes
- `./repo-claude forall 'git status'` - Run git status in all projects

## Cross-Repository Awareness
You have access to these related repositories:
- **frontend** (core,ui): @../frontend - frontend-agent
- **mobile** (mobile,ui): @../mobile - mobile-agent
- **shared-libs** (shared,core): @../shared-libs - no agent

## Guidelines
1. Work directly on main branch (trunk-based development)
2. Make small, frequent commits
3. Update shared memory with your progress
4. Use `./repo-claude sync` to stay up to date with all projects
5. Consider impacts on other repositories
6. Focus on backend but be aware of cross-repo dependencies

## Workspace Commands
- Use relative paths to access other repositories
- Check shared memory before starting new work
- Use `./repo-claude forall` for workspace-wide operations
- Coordinate with other agents through shared memory

## Example Usage
```bash
# See status of all projects
./repo-claude status

# Sync all projects
./repo-claude sync

# Run a command in all projects
./repo-claude forall 'git log --oneline -5'
```
