package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/taokim/repo-claude/internal/config"
)

// setupCoordination creates coordination files
func (m *Manager) setupCoordination() error {
	// Create shared memory file
	if m.FileSystem == nil {
		m.FileSystem = RealFileSystem{}
	}
	sharedMemoryPath := filepath.Join(m.WorkspacePath, "shared-memory.md")
	if _, err := m.FileSystem.Stat(sharedMemoryPath); os.IsNotExist(err) {
		content := `# Shared Agent Memory

## Current Tasks
- No active tasks

## Coordination Notes
- Agents will update this file with their progress
- Use this for cross-repository coordination
- All repositories managed by repo-claude

## Commands Available
- ` + "`rc status`" + ` - Show status of all projects
- ` + "`rc sync`" + ` - Sync all projects
- ` + "`rc forall '<command>'`" + ` - Run command in all projects

## Decisions
- Document architectural decisions here
`
		if err := m.FileSystem.WriteFile(sharedMemoryPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("creating shared memory: %w", err)
		}
	}

	// Create CLAUDE.md files for each repository with agents
	for _, project := range m.Config.Workspace.Manifest.Projects {
		if project.Agent != "" {
			if err := m.createClaudeMD(project); err != nil {
				return fmt.Errorf("creating CLAUDE.md for %s: %w", project.Name, err)
			}
		}
	}

	return nil
}

// createClaudeMD creates a CLAUDE.md file for a repository
func (m *Manager) createClaudeMD(project config.Project) error {
	// Use project path if specified, otherwise use name
	path := project.Name
	if project.Path != "" {
		path = project.Path
	}
	repoPath := filepath.Join(m.WorkspacePath, path)
	
	if m.FileSystem == nil {
		m.FileSystem = RealFileSystem{}
	}
	// Create repo directory if it doesn't exist
	if err := m.FileSystem.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("creating repo directory: %w", err)
	}
	
	claudeMDPath := filepath.Join(repoPath, "CLAUDE.md")
	
	agentConfig, exists := m.Config.Agents[project.Agent]
	if !exists {
		return fmt.Errorf("agent %s not found in config", project.Agent)
	}
	
	// Calculate relative paths
	sharedMemoryRel, _ := filepath.Rel(repoPath, filepath.Join(m.WorkspacePath, "shared-memory.md"))
	workspaceRel, _ := filepath.Rel(repoPath, m.WorkspacePath)
	
	// Get other repositories for cross-awareness
	var otherProjects []string
	for _, p := range m.Config.Workspace.Manifest.Projects {
		if p.Name != project.Name {
			// Use project path if specified, otherwise use name
			otherPath := p.Name
			if p.Path != "" {
				otherPath = p.Path
			}
			otherPathRel, _ := filepath.Rel(repoPath, filepath.Join(m.WorkspacePath, otherPath))
			agent := p.Agent
			if agent == "" {
				agent = "no agent"
			}
			otherProjects = append(otherProjects, fmt.Sprintf("- **%s** (%s): @%s - %s", 
				p.Name, p.Groups, otherPathRel, agent))
		}
	}
	
	content := fmt.Sprintf(`# %s - %s

## Agent Information
- **Repository**: %s
- **Project Groups**: %s
- **Specialization**: %s
- **Model**: %s

## Multi-Repository Management
- **This workspace uses repo-claude for multi-repository management**
- **All work happens on main branch (trunk-based development)**
- **Workspace root**: @%s

## Coordination
- **Shared Memory**: @%s

## Commands You Can Use
- %s - Show status of all projects
- %s - Sync all projects from remotes
- %s - Run git status in all projects

## Cross-Repository Awareness
You have access to these related repositories:
%s

## Guidelines
1. Work directly on main branch (trunk-based development)
2. Make small, frequent commits
3. Update shared memory with your progress
4. Use %s to stay up to date with all projects
5. Consider impacts on other repositories
6. Focus on %s but be aware of cross-repo dependencies

## Workspace Commands
- Use relative paths to access other repositories
- Check shared memory before starting new work
- Use %s for workspace-wide operations
- Coordinate with other agents through shared memory

## Example Usage
%s
# See status of all projects
rc status

# Sync all projects
rc sync

# Run a command in all projects
rc forall 'git log --oneline -5'
%s
`,
		project.Agent,
		project.Name,
		project.Name,
		project.Groups,
		agentConfig.Specialization,
		agentConfig.Model,
		workspaceRel,
		sharedMemoryRel,
		"`rc status`",
		"`rc sync`",
		"`rc forall 'git status'`",
		strings.Join(otherProjects, "\n"),
		"`rc sync`",
		project.Name,
		"`rc forall`",
		"```bash",
		"```",
	)
	
	return m.FileSystem.WriteFile(claudeMDPath, []byte(content), 0644)
}