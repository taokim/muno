package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/taokim/repo-claude/internal/config"
)

// SelectionMode represents the current selection mode
type SelectionMode string

const (
	ModeNone  SelectionMode = "none"
	ModeScope SelectionMode = "scope"
	ModeRepo  SelectionMode = "repo"
)

// StartItemV2 represents a selectable item (scope or repository)
type StartItemV2 struct {
	Name         string
	Type         string // "scope" or "repo"
	ItemDesc     string
	IsRunning    bool
	Selected     bool
	Repos        []string // For scopes, list of included repos
	ScopeContaining string // For repos, which scope contains this repo (if any)
}

func (i StartItemV2) Title() string {
	status := "⚫"
	if i.IsRunning {
		status = "🟢"
	}
	
	// Different selection indicators based on type
	selected := " "
	if i.Selected {
		if i.Type == "scope" {
			selected = "●" // Radio button (filled circle)
		} else {
			selected = "✓" // Checkbox
		}
	} else {
		if i.Type == "scope" {
			selected = "○" // Radio button (empty circle)
		} else {
			selected = "☐" // Empty checkbox
		}
	}
	
	prefix := ""
	if i.Type == "scope" {
		prefix = "[SCOPE] "
	} else {
		prefix = "[REPO]  "
	}
	
	return fmt.Sprintf("%s %s %s%s", selected, status, prefix, i.Name)
}

func (i StartItemV2) Description() string {
	if i.Type == "scope" && len(i.Repos) > 0 {
		repoList := i.Repos
		if len(repoList) > 3 {
			repoList = append(repoList[:3], fmt.Sprintf("... +%d more", len(repoList)-3))
		}
		return fmt.Sprintf("%s | Includes: %s", i.ItemDesc, strings.Join(repoList, ", "))
	}
	if i.Type == "repo" && i.ScopeContaining != "" {
		return fmt.Sprintf("%s | In scope: %s", i.ItemDesc, i.ScopeContaining)
	}
	return i.ItemDesc
}

func (i StartItemV2) FilterValue() string { return i.Name }

// StartModelV2 represents the improved interactive start UI model
type StartModelV2 struct {
	list         list.Model
	items        []list.Item
	selectedScope string // Name of selected scope (radio button behavior)
	selectedRepos map[string]bool // Selected repos (checkbox behavior)
	selectionMode SelectionMode
	config       *config.Config
	state        *config.State
	err          error
	quitting     bool
	launching    bool
	help         help.Model
	keys         keyMapV2
	showHelp     bool
	filterMode   string // "", "running", "stopped", "scopes", "repos"
}

type keyMapV2 struct {
	Up       key.Binding
	Down     key.Binding
	Space    key.Binding
	Enter    key.Binding
	Quit     key.Binding
	Help     key.Binding
	ClearAll key.Binding
	Filter   key.Binding
	Tab      key.Binding // Switch between scope/repo mode
}

func (k keyMapV2) ShortHelp() []key.Binding {
	return []key.Binding{k.Space, k.Enter, k.Tab, k.Help, k.Quit}
}

func (k keyMapV2) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Space, k.Enter},
		{k.Tab, k.ClearAll, k.Filter},
		{k.Help, k.Quit},
	}
}

var keysV2 = keyMapV2{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "select"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "start selected"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q/esc", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	ClearAll: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clear selection"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "cycle filters"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch scope/repo mode"),
	),
}

// NewStartModelV2 creates a new improved interactive start model
func NewStartModelV2(cfg *config.Config, state *config.State) *StartModelV2 {
	m := &StartModelV2{
		config:        cfg,
		state:         state,
		selectedRepos: make(map[string]bool),
		selectionMode: ModeNone,
		help:          help.New(),
		keys:          keysV2,
	}

	// Build items list from config
	items := m.buildItems()
	
	// Create list with custom styles
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("170")).
		Foreground(lipgloss.Color("170")).
		Bold(true)
	
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedTitle.Copy().
		Foreground(lipgloss.Color("170")).
		Bold(false)

	l := list.New(items, delegate, 0, 0)
	l.Title = "🚀 Select Scope (○) OR Repositories (☐) to Start"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false) // We'll use our own help
	l.DisableQuitKeybindings() // Use our own quit handling
	
	m.list = l
	m.items = items
	
	return m
}

// buildItems creates list items from config - shows ALL repos and scopes
func (m *StartModelV2) buildItems() []list.Item {
	var items []list.Item
	
	// Track which repos belong to which scopes
	repoToScope := make(map[string]string)
	
	// Add scopes first
	if len(m.config.Scopes) > 0 {
		for name, scope := range m.config.Scopes {
			repos := m.resolveRepos(scope.Repos)
			isRunning := false
			if m.state != nil && m.state.Scopes != nil {
				if status, exists := m.state.Scopes[name]; exists {
					isRunning = status.Status == "running"
				}
			}
			
			// Track which repos belong to this scope
			for _, repo := range repos {
				repoToScope[repo] = name
			}
			
			items = append(items, StartItemV2{
				Name:        name,
				Type:        "scope",
				ItemDesc:    scope.Description,
				IsRunning:   isRunning,
				Repos:       repos,
			})
		}
	}
	
	// Add ALL repositories (not just those outside scopes)
	for name, repo := range m.config.Repositories {
		scopeContaining := repoToScope[name]
		groups := "general"
		if len(repo.Groups) > 0 {
			groups = strings.Join(repo.Groups, ", ")
		}
		items = append(items, StartItemV2{
			Name:            name,
			Type:            "repo",
			ItemDesc:        fmt.Sprintf("Repository (%s)", groups),
			IsRunning:       false, // Individual repos don't have running state
			ScopeContaining: scopeContaining,
		})
	}
	
	return items
}

// resolveRepos expands repository patterns including wildcards
func (m *StartModelV2) resolveRepos(patterns []string) []string {
	var repos []string
	seen := make(map[string]bool)
	
	for _, pattern := range patterns {
		if strings.Contains(pattern, "*") {
			// Wildcard pattern
			for repoName := range m.config.Repositories {
				if matchPattern(pattern, repoName) && !seen[repoName] {
					repos = append(repos, repoName)
					seen[repoName] = true
				}
			}
		} else {
			// Exact match
			if !seen[pattern] {
				repos = append(repos, pattern)
				seen[pattern] = true
			}
		}
	}
	
	return repos
}

// matchPattern checks if a name matches a wildcard pattern
func matchPattern(pattern, name string) bool {
	// Simple wildcard matching (just * at end for now)
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(name, prefix)
	}
	return pattern == name
}

// Init implements tea.Model
func (m *StartModelV2) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *StartModelV2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2) // Leave room for help
		return m, nil
		
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
			
		case key.Matches(msg, m.keys.Enter):
			if m.hasSelection() {
				m.launching = true
				return m, tea.Quit
			}
			
		case key.Matches(msg, m.keys.Space):
			if item, ok := m.list.SelectedItem().(StartItemV2); ok {
				m.toggleSelection(item)
				m.updateListItems()
			}
			
		case key.Matches(msg, m.keys.ClearAll):
			m.clearSelection()
			m.updateListItems()
			
		case key.Matches(msg, m.keys.Filter):
			m.cycleFilter()
			
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			
		case key.Matches(msg, m.keys.Tab):
			m.switchMode()
		}
	}
	
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// toggleSelection handles exclusive selection logic
func (m *StartModelV2) toggleSelection(item StartItemV2) {
	if item.Type == "scope" {
		// Scope selection (radio button behavior)
		if m.selectedScope == item.Name {
			// Deselect current scope
			m.selectedScope = ""
			m.selectionMode = ModeNone
		} else {
			// Select new scope (deselects all repos and previous scope)
			m.selectedScope = item.Name
			m.selectedRepos = make(map[string]bool) // Clear all repo selections
			m.selectionMode = ModeScope
		}
	} else if item.Type == "repo" {
		// Repo selection (checkbox behavior)
		if m.selectionMode == ModeScope {
			// Switching from scope to repos - clear scope selection
			m.selectedScope = ""
			m.selectionMode = ModeRepo
		}
		// Toggle repo selection
		m.selectedRepos[item.Name] = !m.selectedRepos[item.Name]
		
		// If no repos selected anymore, reset mode
		hasSelection := false
		for _, selected := range m.selectedRepos {
			if selected {
				hasSelection = true
				break
			}
		}
		if !hasSelection {
			m.selectionMode = ModeNone
		} else {
			m.selectionMode = ModeRepo
		}
	}
}

// clearSelection clears all selections
func (m *StartModelV2) clearSelection() {
	m.selectedScope = ""
	m.selectedRepos = make(map[string]bool)
	m.selectionMode = ModeNone
}

// switchMode switches between scope and repo selection modes
func (m *StartModelV2) switchMode() {
	if m.selectionMode == ModeScope {
		// Switch to repo mode
		m.selectedScope = ""
		m.selectionMode = ModeRepo
	} else if m.selectionMode == ModeRepo {
		// Switch to scope mode
		m.selectedRepos = make(map[string]bool)
		m.selectionMode = ModeScope
	}
	m.updateListItems()
}

// updateListItems updates the list items with current selection state
func (m *StartModelV2) updateListItems() {
	items := m.list.Items()
	for i, listItem := range items {
		if si, ok := listItem.(StartItemV2); ok {
			if si.Type == "scope" {
				si.Selected = (m.selectedScope == si.Name)
			} else if si.Type == "repo" {
				si.Selected = m.selectedRepos[si.Name]
			}
			items[i] = si
		}
	}
	m.list.SetItems(items)
}

// View implements tea.Model
func (m *StartModelV2) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n❌ Error: %v\n", m.err)
	}
	
	if m.quitting {
		if m.launching {
			return m.getLaunchSummary()
		}
		return "\n👋 Cancelled\n"
	}
	
	var s strings.Builder
	s.WriteString(m.list.View())
	
	if m.showHelp {
		s.WriteString("\n")
		s.WriteString(m.help.View(m.keys))
	} else {
		s.WriteString("\n")
		s.WriteString(m.getStatusLine())
	}
	
	return s.String()
}

// hasSelection checks if anything is selected
func (m *StartModelV2) hasSelection() bool {
	if m.selectedScope != "" {
		return true
	}
	for _, selected := range m.selectedRepos {
		if selected {
			return true
		}
	}
	return false
}

// cycleFilter cycles through filter modes
func (m *StartModelV2) cycleFilter() {
	filters := []string{"", "running", "stopped", "scopes", "repos"}
	currentIdx := 0
	for i, f := range filters {
		if f == m.filterMode {
			currentIdx = i
			break
		}
	}
	
	m.filterMode = filters[(currentIdx+1)%len(filters)]
	m.applyFilter()
}

// applyFilter applies the current filter to the list
func (m *StartModelV2) applyFilter() {
	var filtered []list.Item
	
	for _, item := range m.items {
		if si, ok := item.(StartItemV2); ok {
			include := true
			
			switch m.filterMode {
			case "running":
				include = si.IsRunning
			case "stopped":
				include = !si.IsRunning
			case "scopes":
				include = si.Type == "scope"
			case "repos":
				include = si.Type == "repo"
			}
			
			if include {
				filtered = append(filtered, item)
			}
		}
	}
	
	m.list.SetItems(filtered)
	if m.filterMode != "" {
		m.list.Title = fmt.Sprintf("🚀 Select to Start (Filter: %s)", m.filterMode)
	} else {
		m.list.Title = "🚀 Select Scope (○) OR Repositories (☐) to Start"
	}
}

// getStatusLine returns the status line text
func (m *StartModelV2) getStatusLine() string {
	status := ""
	
	if m.selectedScope != "" {
		status = fmt.Sprintf("Scope: %s", m.selectedScope)
	} else {
		repoCount := 0
		for _, selected := range m.selectedRepos {
			if selected {
				repoCount++
			}
		}
		if repoCount > 0 {
			status = fmt.Sprintf("Repos: %d selected", repoCount)
		} else {
			status = "No selection"
		}
	}
	
	if m.selectionMode != ModeNone {
		status += fmt.Sprintf(" | Mode: %s", m.selectionMode)
	}
	
	if m.filterMode != "" {
		status += fmt.Sprintf(" | Filter: %s", m.filterMode)
	}
	status += " | Press ? for help"
	
	return lipgloss.NewStyle().Faint(true).Render(status)
}

// getLaunchSummary returns a summary of what will be launched
func (m *StartModelV2) getLaunchSummary() string {
	if m.selectedScope != "" {
		return fmt.Sprintf("\n🚀 Starting scope: %s\n", m.selectedScope)
	}
	
	var repos []string
	for repo, selected := range m.selectedRepos {
		if selected {
			repos = append(repos, repo)
		}
	}
	
	if len(repos) == 0 {
		return "\n👋 Nothing selected\n"
	}
	
	if len(repos) == 1 {
		return fmt.Sprintf("\n🚀 Starting repository: %s\n", repos[0])
	}
	
	return fmt.Sprintf("\n🚀 Starting repositories: %s\n", strings.Join(repos, ", "))
}

// GetSelected returns the selected items
func (m *StartModelV2) GetSelected() []StartItemV2 {
	var selected []StartItemV2
	
	for _, item := range m.items {
		if si, ok := item.(StartItemV2); ok {
			if si.Type == "scope" && m.selectedScope == si.Name {
				selected = append(selected, si)
			} else if si.Type == "repo" && m.selectedRepos[si.Name] {
				selected = append(selected, si)
			}
		}
	}
	
	return selected
}

// IsLaunching returns true if the user chose to launch
func (m *StartModelV2) IsLaunching() bool {
	return m.launching
}

// GetSelectionMode returns the current selection mode
func (m *StartModelV2) GetSelectionMode() SelectionMode {
	return m.selectionMode
}

// GetSelectedScope returns the selected scope name (if any)
func (m *StartModelV2) GetSelectedScope() string {
	return m.selectedScope
}

// GetSelectedRepos returns the list of selected repository names
func (m *StartModelV2) GetSelectedRepos() []string {
	var repos []string
	for repo, selected := range m.selectedRepos {
		if selected {
			repos = append(repos, repo)
		}
	}
	return repos
}