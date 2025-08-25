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
	model := tui.NewStartModel(m.Config, nil) // No state tracking anymore
	
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
	
	// Legacy agent support removed - only scopes are supported now
	
	return nil
}

// StartReposAsScope starts a Claude session with multiple repositories as a temporary scope
func (m *Manager) StartReposAsScope(repos []string, opts StartOptions) error {
	// Generate a scope name based on repos
	scopeName := "backend" // Default to backend if multiple repos match
	
	// Check if all repos belong to an existing scope
	for name, scope := range m.Config.Scopes {
		resolvedRepos := m.resolveScopeRepos(scope.Repos)
		allMatch := true
		for _, repo := range repos {
			found := false
			for _, resolved := range resolvedRepos {
				if resolved == repo {
					found = true
					break
				}
			}
			if !found {
				allMatch = false
				break
			}
		}
		if allMatch && len(resolvedRepos) >= len(repos) {
			scopeName = name
			break
		}
	}
	
	// Start with the given repos
	return m.startScopeWithRepos(scopeName, repos, opts)
}

// Helper method to start a scope with specific repos
func (m *Manager) startScopeWithRepos(scopeName string, repos []string, opts StartOptions) error {
	fmt.Printf("🚀 Starting session with repos: %s\n", strings.Join(repos, ", "))
	
	// If there's an existing scope with these exact repos, use it
	if scopeConfig, exists := m.Config.Scopes[scopeName]; exists {
		resolvedRepos := m.resolveScopeRepos(scopeConfig.Repos)
		
		// Check if resolved repos match our target repos
		if len(resolvedRepos) == len(repos) {
			match := true
			for _, repo := range repos {
				found := false
				for _, resolved := range resolvedRepos {
					if resolved == repo {
						found = true
						break
					}
				}
				if !found {
					match = false
					break
				}
			}
			if match {
				// Use the existing scope configuration
				fmt.Printf("🚀 Starting scope %s with %d repositories (current terminal)\n", scopeName, len(repos))
				return m.StartScopeWithOptions(scopeName, opts)
			}
		} else {
			// Still use the scope even if it has more repos
			fmt.Printf("🚀 Starting scope %s with %d repositories (current terminal)\n", scopeName, len(resolvedRepos))
			return m.StartScopeWithOptions(scopeName, opts)
		}
	}
	
	// If no matching scope, create a temporary scope-like session
	// This should never happen in normal flow, but handle it gracefully
	if len(repos) > 0 {
		// Fall back to starting individual repo
		return m.StartRepoAsSingleScope(repos[0], opts)
	}
	
	return fmt.Errorf("no repositories to start")
}