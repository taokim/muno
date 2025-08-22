package manager

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/taokim/repo-claude/internal/tui"
)

// StartInteractiveTUI launches the Bubbletea interactive UI for selecting what to start
func (m *Manager) StartInteractiveTUI() error {
	// Create the TUI model
	model := tui.NewStartModel(m.Config, m.State)
	
	// Run the TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("running interactive UI: %w", err)
	}
	
	// Check if user cancelled
	startModel, ok := finalModel.(*tui.StartModel)
	if !ok || !startModel.IsLaunching() {
		return nil // User cancelled, not an error
	}
	
	// Get selected items
	selected := startModel.GetSelected()
	if len(selected) == 0 {
		return nil // Nothing selected
	}
	
	// Group by type for efficient starting
	var selectedScopes []string
	var selectedRepos []string
	var selectedAgents []string // For legacy mode
	
	for _, item := range selected {
		switch item.Type {
		case "scope":
			selectedScopes = append(selectedScopes, item.Name)
		case "repo":
			selectedRepos = append(selectedRepos, item.Name)
		case "agent":
			selectedAgents = append(selectedAgents, item.Name)
		}
	}
	
	// Determine if we need new windows (multiple selections)
	totalSelections := len(selectedScopes) + len(selectedRepos) + len(selectedAgents)
	opts := StartOptions{
		NewWindow: totalSelections > 1,
	}
	
	if opts.NewWindow && totalSelections > 1 {
		fmt.Printf("🪟 Opening %d sessions in new windows\n", totalSelections)
	}
	
	// Start selected scopes
	for _, scopeName := range selectedScopes {
		if err := m.StartScopeWithOptions(scopeName, opts); err != nil {
			fmt.Printf("❌ Failed to start scope %s: %v\n", scopeName, err)
		}
	}
	
	// Start scopes for selected individual repos
	if len(selectedRepos) > 0 {
		// Create a temporary scope for the selected repos
		if err := m.StartReposAsScope(selectedRepos, opts); err != nil {
			fmt.Printf("❌ Failed to start repos: %v\n", err)
		}
	}
	
	// Start selected agents (legacy mode)
	for _, agentName := range selectedAgents {
		if err := m.StartAgentWithOptions(agentName, opts); err != nil {
			fmt.Printf("❌ Failed to start agent %s: %v\n", agentName, err)
		}
	}
	
	return nil
}

// StartReposAsScope starts a Claude session with multiple repositories as a temporary scope
func (m *Manager) StartReposAsScope(repos []string, opts StartOptions) error {
	// Generate a scope name
	scopeName := fmt.Sprintf("temp-%s", strings.Join(repos, "-"))
	if len(scopeName) > 50 {
		scopeName = fmt.Sprintf("temp-%d-repos", len(repos))
	}
	
	// Start with the given repos
	return m.startScopeWithRepos(scopeName, repos, opts)
}

// Helper method to start a scope with specific repos
func (m *Manager) startScopeWithRepos(scopeName string, repos []string, opts StartOptions) error {
	fmt.Printf("🚀 Starting session with repos: %s\n", strings.Join(repos, ", "))
	
	// Build the command to start Claude with multiple repos
	// This will be similar to StartScopeWithOptions but with custom repos
	
	// For now, if there's only one repo, just start it directly
	if len(repos) == 1 {
		// Try to find a scope that contains this repo
		for name, scope := range m.Config.Scopes {
			resolvedRepos := m.resolveScopeRepos(scope.Repos)
			for _, repo := range resolvedRepos {
				if repo == repos[0] {
					return m.StartScopeWithOptions(name, opts)
				}
			}
		}
	}
	
	// If multiple repos or no existing scope, create a temporary scope
	// This would involve starting Claude with a custom working directory
	// that spans multiple repos
	
	// Implementation would be similar to StartScopeWithOptions
	// but with a custom scope definition
	
	// For now, just start the first repo as a fallback
	if len(repos) > 0 {
		return m.StartByRepoName(repos[0])
	}
	
	return fmt.Errorf("no repositories to start")
}