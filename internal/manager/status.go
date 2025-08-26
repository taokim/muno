//go:build legacy
// +build legacy

package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

// ShowStatus displays the current workspace status
func (m *Manager) ShowStatus() error {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println(" REPO-CLAUDE STATUS")
	fmt.Println(strings.Repeat("=", 80))
	
	// Workspace info
	fmt.Printf(" Workspace: %s\n", m.Config.Workspace.Name)
	fmt.Printf(" Project Path: %s\n", m.ProjectPath)
	fmt.Printf(" Workspace Path: %s\n", m.WorkspacePath)
	fmt.Println(strings.Repeat("-", 80))
	
	// Scopes
	if len(m.Config.Scopes) > 0 {
		fmt.Println("\n📦 Scopes:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "  NAME\tREPOS\tDESCRIPTION")
		fmt.Fprintln(w, "  ----\t-----\t-----------")
		
		for name, scope := range m.Config.Scopes {
			repos := m.resolveScopeRepos(scope.Repos)
			reposStr := strings.Join(repos, ", ")
			if len(reposStr) > 40 {
				reposStr = reposStr[:37] + "..."
			}
			
			desc := scope.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
			
			fmt.Fprintf(w, "  %s\t%s\t%s\n", name, reposStr, desc)
		}
		w.Flush()
	}
	
	// Repositories
	fmt.Println("\n📂 Repositories:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "  REPO\tPATH\tSTATUS")
	fmt.Fprintln(w, "  ----\t----\t------")
	
	for _, project := range m.Config.Workspace.Manifest.Projects {
		path := project.Name
		if project.Path != "" {
			path = project.Path
		}
		
		repoPath := filepath.Join(m.WorkspacePath, path)
		status := "not cloned"
		
		// Check if repository exists
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
			status = "✓ cloned"
		}
		
		fmt.Fprintf(w, "  %s\t%s\t%s\n", project.Name, path, status)
	}
	w.Flush()
	
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("\n💡 Tips:")
	fmt.Println("  rc start           # Start scopes interactively")
	fmt.Println("  rc start <scope>   # Start a specific scope")
	fmt.Println("  rc list            # List available scopes")
	fmt.Println("  rc pull            # Pull all repositories")
	
	return nil
}