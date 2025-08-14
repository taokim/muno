package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/yourusername/repo-claude/internal/config"
)

// setupCoordination creates coordination files
func (m *Manager) setupCoordination() error {
	// Create shared memory file
	sharedMemoryPath := filepath.Join(m.WorkspacePath, "shared-memory.md")
	if _, err := os.Stat(sharedMemoryPath); os.IsNotExist(err) {
		content := `# Shared Agent Memory

## Current Tasks
- No active tasks

## Coordination Notes
- Agents will update this file with their progress
- Use this for cross-repository coordination
- All repositories managed by Repo tool

## Repo Commands Available
- ` + "`repo status`" + ` - Show status of all projects
- ` + "`repo sync`" + ` - Sync all projects
- ` + "`repo forall -c '<command>'`" + ` - Run command in all projects
- ` + "`repo list`" + ` - List all projects

## Decisions
- Document architectural decisions here
`
		if err := os.WriteFile(sharedMemoryPath, []byte(content), 0644); err != nil {
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
	repoPath := filepath.Join(m.WorkspacePath, project.Name)
	
	// Create repo directory if it doesn't exist
	if err := os.MkdirAll(repoPath, 0755); err != nil {
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
			otherPath, _ := filepath.Rel(repoPath, filepath.Join(m.WorkspacePath, p.Name))
			agent := p.Agent
			if agent == "" {
				agent = "no agent"
			}
			otherProjects = append(otherProjects, fmt.Sprintf("- **%s** (%s): @%s - %s", 
				p.Name, p.Groups, otherPath, agent))
		}
	}
	
	content := fmt.Sprintf(`# %s - %s

## Agent Information
- **Repository**: %s
- **Project Groups**: %s
- **Specialization**: %s
- **Model**: %s

## Repo Tool Integration
- **This workspace uses Repo tool for multi-repository management**
- **All work happens on main branch (trunk-based development)**
- **Workspace root**: @%s

## Coordination
- **Shared Memory**: @%s

## Repo Commands You Can Use
- %s - Show status of all projects
- %s - Sync all projects from remotes
- %s - Run git status in all projects
- %s - Pull latest changes in all projects
- %s - List all projects in workspace

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

## Example Repo Tool Usage
%s
# See status of all projects
repo status

# Sync all projects
repo sync

# Run a command in all projects
repo forall -c 'git log --oneline -5'

# See workspace structure
repo list
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
		"`repo status`",
		"`repo sync`",
		"`repo forall -c 'git status'`",
		"`repo forall -c 'git pull'`",
		"`repo list`",
		strings.Join(otherProjects, "\n"),
		"`repo sync`",
		project.Name,
		"`repo forall`",
		"```bash",
		"```",
	)
	
	return os.WriteFile(claudeMDPath, []byte(content), 0644)
}