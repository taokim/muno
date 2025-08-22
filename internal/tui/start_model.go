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

// StartItem represents a selectable item (scope or repository)
type StartItem struct {
	Name         string
	Type         string // "scope" or "repo"
	ItemDesc     string // renamed to avoid conflict
	IsRunning    bool
	Selected     bool
	Repos        []string // For scopes, list of included repos
}

func (i StartItem) Title() string {
	status := "⚫"
	if i.IsRunning {
		status = "🟢"
	}
	selected := " "
	if i.Selected {
		selected = "✓"
	}
	return fmt.Sprintf("[%s] %s %s", selected, status, i.Name)
}

func (i StartItem) Description() string {
	if i.Type == "scope" && len(i.Repos) > 0 {
		return fmt.Sprintf("%s: %s", i.ItemDesc, strings.Join(i.Repos, ", "))
	}
	return i.ItemDesc
}

func (i StartItem) FilterValue() string { return i.Name }

// StartModel represents the interactive start UI model
type StartModel struct {
	list         list.Model
	items        []list.Item
	selected     map[string]bool
	config       *config.Config
	state        *config.State
	err          error
	quitting     bool
	launching    bool
	help         help.Model
	keys         keyMap
	showHelp     bool
	filterMode   string // "", "running", "stopped", "scopes", "repos"
}

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Space    key.Binding
	Enter    key.Binding
	Quit     key.Binding
	Help     key.Binding
	SelectAll key.Binding
	ClearAll key.Binding
	Filter   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Space, k.Enter, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Space, k.Enter},
		{k.SelectAll, k.ClearAll, k.Filter},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
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
		key.WithHelp("space", "select/deselect"),
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
	SelectAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "select all"),
	),
	ClearAll: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "clear all"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "cycle filters"),
	),
}

// NewStartModel creates a new interactive start model
func NewStartModel(cfg *config.Config, state *config.State) *StartModel {
	m := &StartModel{
		config:   cfg,
		state:    state,
		selected: make(map[string]bool),
		help:     help.New(),
		keys:     keys,
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
	l.Title = "🚀 Select Scopes/Repositories to Start"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false) // We'll use our own help
	l.DisableQuitKeybindings() // Use our own quit handling
	
	m.list = l
	m.items = items
	
	return m
}

// buildItems creates list items from config
func (m *StartModel) buildItems() []list.Item {
	var items []list.Item
	
	// Add scopes
	if len(m.config.Scopes) > 0 {
		for name, scope := range m.config.Scopes {
			repos := m.resolveRepos(scope.Repos)
			isRunning := false
			if m.state != nil && m.state.Scopes != nil {
				if status, exists := m.state.Scopes[name]; exists {
					isRunning = status.Status == "running"
				}
			}
			
			items = append(items, StartItem{
				Name:        name,
				Type:        "scope",
				ItemDesc:    scope.Description,
				IsRunning:   isRunning,
				Repos:       repos,
			})
		}
	} else if len(m.config.Agents) > 0 {
		// Legacy mode: show agents
		for name, agent := range m.config.Agents {
			isRunning := false
			if m.state != nil && m.state.Agents != nil {
				if status, exists := m.state.Agents[name]; exists {
					isRunning = status.Status == "running"
				}
			}
			
			// Find associated repos
			var repos []string
			for _, project := range m.config.Workspace.Manifest.Projects {
				if project.Agent == name {
					repos = append(repos, project.Name)
				}
			}
			
			items = append(items, StartItem{
				Name:        name,
				Type:        "agent",
				ItemDesc:    agent.Specialization,
				IsRunning:   isRunning,
				Repos:       repos,
			})
		}
	}
	
	// Add individual repositories that aren't in any scope
	reposInScopes := make(map[string]bool)
	for _, scope := range m.config.Scopes {
		for _, repo := range m.resolveRepos(scope.Repos) {
			reposInScopes[repo] = true
		}
	}
	
	for _, project := range m.config.Workspace.Manifest.Projects {
		if !reposInScopes[project.Name] {
			items = append(items, StartItem{
				Name:        project.Name,
				Type:        "repo",
				ItemDesc:    fmt.Sprintf("Repository (%s)", project.Groups),
				IsRunning:   false, // Individual repos don't have running state
			})
		}
	}
	
	return items
}

// resolveRepos expands repository patterns including wildcards
func (m *StartModel) resolveRepos(patterns []string) []string {
	var repos []string
	seen := make(map[string]bool)
	
	for _, pattern := range patterns {
		if strings.Contains(pattern, "*") {
			// Wildcard pattern
			for _, project := range m.config.Workspace.Manifest.Projects {
				if matchPattern(pattern, project.Name) && !seen[project.Name] {
					repos = append(repos, project.Name)
					seen[project.Name] = true
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
	if pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(name, pattern[1:len(pattern)-1])
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(name, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(name, pattern[:len(pattern)-1])
	}
	return pattern == name
}

// Init implements tea.Model
func (m *StartModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *StartModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if item, ok := m.list.SelectedItem().(StartItem); ok {
				itemKey := fmt.Sprintf("%s:%s", item.Type, item.Name)
				m.selected[itemKey] = !m.selected[itemKey]
				
				// Update the item in the list
				items := m.list.Items()
				for i, listItem := range items {
					if si, ok := listItem.(StartItem); ok && si.Name == item.Name && si.Type == item.Type {
						si.Selected = m.selected[itemKey]
						items[i] = si
						break
					}
				}
				m.list.SetItems(items)
			}
			
		case key.Matches(msg, m.keys.SelectAll):
			m.selectAll(true)
			
		case key.Matches(msg, m.keys.ClearAll):
			m.selectAll(false)
			
		case key.Matches(msg, m.keys.Filter):
			m.cycleFilter()
			
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
		}
	}
	
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m *StartModel) View() string {
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
func (m *StartModel) hasSelection() bool {
	for _, selected := range m.selected {
		if selected {
			return true
		}
	}
	return false
}

// selectAll selects or deselects all items
func (m *StartModel) selectAll(select_ bool) {
	items := m.list.Items()
	for i, item := range items {
		if si, ok := item.(StartItem); ok {
			itemKey := fmt.Sprintf("%s:%s", si.Type, si.Name)
			m.selected[itemKey] = select_
			si.Selected = select_
			items[i] = si
		}
	}
	m.list.SetItems(items)
}

// cycleFilter cycles through filter modes
func (m *StartModel) cycleFilter() {
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
func (m *StartModel) applyFilter() {
	var filtered []list.Item
	
	for _, item := range m.items {
		if si, ok := item.(StartItem); ok {
			include := true
			
			switch m.filterMode {
			case "running":
				include = si.IsRunning
			case "stopped":
				include = !si.IsRunning
			case "scopes":
				include = si.Type == "scope" || si.Type == "agent"
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
		m.list.Title = "🚀 Select Scopes/Repositories to Start"
	}
}

// getStatusLine returns the status line text
func (m *StartModel) getStatusLine() string {
	selected := 0
	for _, sel := range m.selected {
		if sel {
			selected++
		}
	}
	
	status := fmt.Sprintf("%d selected", selected)
	if m.filterMode != "" {
		status += fmt.Sprintf(" | Filter: %s", m.filterMode)
	}
	status += " | Press ? for help"
	
	return lipgloss.NewStyle().Faint(true).Render(status)
}

// getLaunchSummary returns a summary of what will be launched
func (m *StartModel) getLaunchSummary() string {
	var toStart []string
	
	for key, selected := range m.selected {
		if selected {
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				toStart = append(toStart, fmt.Sprintf("%s %s", parts[0], parts[1]))
			}
		}
	}
	
	if len(toStart) == 0 {
		return "\n👋 Nothing selected\n"
	}
	
	return fmt.Sprintf("\n🚀 Starting: %s\n", strings.Join(toStart, ", "))
}

// GetSelected returns the selected items
func (m *StartModel) GetSelected() []StartItem {
	var selected []StartItem
	
	for _, item := range m.items {
		if si, ok := item.(StartItem); ok {
			itemKey := fmt.Sprintf("%s:%s", si.Type, si.Name)
			if m.selected[itemKey] {
				selected = append(selected, si)
			}
		}
	}
	
	return selected
}

// IsLaunching returns true if the user chose to launch
func (m *StartModel) IsLaunching() bool {
	return m.launching
}