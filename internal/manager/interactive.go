package manager

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/yourusername/repo-claude/internal/config"
)

// interactiveConfig handles interactive configuration setup
func (m *Manager) interactiveConfig() error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("\n🔧 Configuration Setup")
	fmt.Println("Press Enter to use defaults, or type new values:")
	
	// Configure manifest settings
	manifest := &m.Config.Workspace.Manifest
	fmt.Println("\n📦 Manifest Configuration:")
	
	remoteFetch := prompt(reader, fmt.Sprintf("  GitHub organization URL [%s]: ", manifest.RemoteFetch))
	if remoteFetch != "" {
		manifest.RemoteFetch = remoteFetch
	}
	
	// Configure projects
	fmt.Println("\n📁 Project Configuration:")
	projects := []config.Project{}
	
	for i, project := range manifest.Projects {
		fmt.Printf("\nProject %d:\n", i+1)
		
		name := prompt(reader, fmt.Sprintf("  Name [%s]: ", project.Name))
		if name == "" {
			name = project.Name
		}
		
		groups := prompt(reader, fmt.Sprintf("  Groups [%s]: ", project.Groups))
		if groups == "" {
			groups = project.Groups
		}
		
		agent := project.Agent
		if agent != "" {
			newAgent := prompt(reader, fmt.Sprintf("  Agent [%s]: ", agent))
			if newAgent != "" {
				agent = newAgent
			}
		} else {
			agent = prompt(reader, "  Agent (optional): ")
		}
		
		projects = append(projects, config.Project{
			Name:   name,
			Groups: groups,
			Agent:  agent,
		})
	}
	
	// Ask if they want to add more projects
	for {
		addMore := prompt(reader, "\nAdd another project? (y/N): ")
		if strings.ToLower(addMore) != "y" && strings.ToLower(addMore) != "yes" {
			break
		}
		
		name := prompt(reader, "  Name: ")
		if name == "" {
			continue
		}
		
		groups := prompt(reader, "  Groups: ")
		if groups == "" {
			groups = "default"
		}
		
		agent := prompt(reader, "  Agent (optional): ")
		
		projects = append(projects, config.Project{
			Name:   name,
			Groups: groups,
			Agent:  agent,
		})
	}
	
	manifest.Projects = projects
	
	// Configure agents
	fmt.Println("\n🤖 Agent Configuration:")
	agents := make(map[string]config.Agent)
	
	for _, project := range projects {
		if project.Agent != "" && agents[project.Agent].Model == "" {
			fmt.Printf("\nAgent: %s\n", project.Agent)
			
			specialization := prompt(reader, "  Specialization: ")
			
			model := prompt(reader, "  Model [claude-sonnet-4]: ")
			if model == "" {
				model = "claude-sonnet-4"
			}
			
			autoStartStr := prompt(reader, "  Auto-start? [Y/n]: ")
			autoStart := strings.ToLower(autoStartStr) != "n" && strings.ToLower(autoStartStr) != "no"
			
			agents[project.Agent] = config.Agent{
				Model:          model,
				Specialization: specialization,
				AutoStart:      autoStart,
				Dependencies:   []string{},
			}
		}
	}
	
	m.Config.Agents = agents
	return nil
}

// prompt reads user input with a prompt message
func prompt(reader *bufio.Reader, message string) string {
	fmt.Print(message)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}