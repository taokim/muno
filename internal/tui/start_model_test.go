package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taokim/repo-claude/internal/config"
)

func TestNewStartModel(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Manifest: config.Manifest{
				Projects: []config.Project{
					{Name: "frontend", Groups: "ui"},
					{Name: "backend", Groups: "api"},
				},
			},
		},
		Scopes: map[string]config.Scope{
			"fullstack": {
				Repos:       []string{"frontend", "backend"},
				Description: "Full-stack development",
				Model:       "claude-3",
				AutoStart:   true,
			},
		},
	}
	
	state := &config.State{
		Scopes: map[string]config.ScopeStatus{
			"fullstack": {
				Name:   "fullstack",
				Status: "running",
				PID:    12345,
			},
		},
	}
	
	model := NewStartModel(cfg, state)
	assert.NotNil(t, model)
	assert.NotNil(t, model.config)
	assert.NotNil(t, model.state)
	assert.NotNil(t, model.selected)
	assert.NotNil(t, model.items)
	
	// Check that items were built correctly
	assert.Greater(t, len(model.items), 0)
}

func TestStartItem(t *testing.T) {
	item := StartItem{
		Name:      "backend",
		Type:      "scope",
		ItemDesc:  "Backend services",
		IsRunning: true,
		Selected:  true,
		Repos:     []string{"auth", "api", "db"},
	}
	
	// Test Title method
	title := item.Title()
	assert.Contains(t, title, "[✓]") // Selected
	assert.Contains(t, title, "🟢")  // Running
	assert.Contains(t, title, "backend")
	
	// Test Description method
	desc := item.Description()
	assert.Contains(t, desc, "Backend services")
	assert.Contains(t, desc, "auth, api, db")
	
	// Test FilterValue method
	assert.Equal(t, "backend", item.FilterValue())
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		name     string
		expected bool
	}{
		{"*", "anything", true},
		{"backend", "backend", true},
		{"backend", "frontend", false},
		{"back*", "backend", true},
		{"back*", "backoffice", true},
		{"back*", "frontend", false},
		{"*end", "backend", true},
		{"*end", "frontend", true},
		{"*end", "middle", false},
		{"*service*", "auth-service", true},
		{"*service*", "service-mesh", true},
		{"*service*", "backend", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.name, func(t *testing.T) {
			result := matchPattern(tt.pattern, tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveRepos(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Manifest: config.Manifest{
				Projects: []config.Project{
					{Name: "auth-service"},
					{Name: "order-service"},
					{Name: "payment-service"},
					{Name: "frontend"},
					{Name: "mobile"},
				},
			},
		},
	}
	
	model := &StartModel{
		config: cfg,
	}
	
	tests := []struct {
		patterns []string
		expected []string
	}{
		{
			patterns: []string{"frontend"},
			expected: []string{"frontend"},
		},
		{
			patterns: []string{"*-service"},
			expected: []string{"auth-service", "order-service", "payment-service"},
		},
		{
			patterns: []string{"frontend", "mobile"},
			expected: []string{"frontend", "mobile"},
		},
		{
			patterns: []string{"*"},
			expected: []string{"auth-service", "order-service", "payment-service", "frontend", "mobile"},
		},
	}
	
	for _, tt := range tests {
		repos := model.resolveRepos(tt.patterns)
		assert.ElementsMatch(t, tt.expected, repos)
	}
}