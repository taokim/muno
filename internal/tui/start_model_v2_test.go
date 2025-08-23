package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/repo-claude/internal/config"
)

func TestNewStartModelV2(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Manifest: config.Manifest{
				Projects: []config.Project{
					{Name: "frontend", Groups: "ui"},
					{Name: "backend", Groups: "api"},
					{Name: "database", Groups: "data"},
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
			"data": {
				Repos:       []string{"database"},
				Description: "Database management",
				Model:       "claude-3",
				AutoStart:   false,
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
	
	model := NewStartModelV2(cfg, state)
	assert.NotNil(t, model)
	assert.NotNil(t, model.config)
	assert.NotNil(t, model.state)
	assert.NotNil(t, model.selectedRepos)
	assert.Equal(t, "", model.selectedScope)
	assert.Equal(t, ModeNone, model.selectionMode)
	
	// Check that items were built correctly
	assert.Greater(t, len(model.items), 0)
	
	// Check that we have both scopes and ALL repos
	scopeCount := 0
	repoCount := 0
	for _, item := range model.items {
		if si, ok := item.(StartItemV2); ok {
			if si.Type == "scope" {
				scopeCount++
			} else if si.Type == "repo" {
				repoCount++
			}
		}
	}
	
	assert.Equal(t, 2, scopeCount, "Should have 2 scopes")
	assert.Equal(t, 3, repoCount, "Should have ALL 3 repos listed")
}

func TestStartItemV2_Display(t *testing.T) {
	tests := []struct {
		name     string
		item     StartItemV2
		wantTitle string
		wantDesc  string
	}{
		{
			name: "Unselected scope (radio button)",
			item: StartItemV2{
				Name:      "backend",
				Type:      "scope",
				ItemDesc:  "Backend services",
				IsRunning: false,
				Selected:  false,
				Repos:     []string{"auth", "api", "db"},
			},
			wantTitle: "○ ⚫ [SCOPE] backend",
			wantDesc:  "Backend services | Includes: auth, api, db",
		},
		{
			name: "Selected scope (radio button)",
			item: StartItemV2{
				Name:      "backend",
				Type:      "scope",
				ItemDesc:  "Backend services",
				IsRunning: true,
				Selected:  true,
				Repos:     []string{"auth", "api", "db"},
			},
			wantTitle: "● 🟢 [SCOPE] backend",
			wantDesc:  "Backend services | Includes: auth, api, db",
		},
		{
			name: "Unselected repo (checkbox)",
			item: StartItemV2{
				Name:            "frontend",
				Type:            "repo",
				ItemDesc:        "Repository (ui)",
				IsRunning:       false,
				Selected:        false,
				ScopeContaining: "fullstack",
			},
			wantTitle: "☐ ⚫ [REPO]  frontend",
			wantDesc:  "Repository (ui) | In scope: fullstack",
		},
		{
			name: "Selected repo (checkbox)",
			item: StartItemV2{
				Name:      "frontend",
				Type:      "repo",
				ItemDesc:  "Repository (ui)",
				IsRunning: false,
				Selected:  true,
			},
			wantTitle: "✓ ⚫ [REPO]  frontend",
			wantDesc:  "Repository (ui)",
		},
		{
			name: "Scope with many repos (truncated)",
			item: StartItemV2{
				Name:      "large",
				Type:      "scope",
				ItemDesc:  "Large scope",
				IsRunning: false,
				Selected:  false,
				Repos:     []string{"r1", "r2", "r3", "r4", "r5"},
			},
			wantTitle: "○ ⚫ [SCOPE] large",
			wantDesc:  "Large scope | Includes: r1, r2, r3, ... +2 more",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantTitle, tt.item.Title())
			assert.Equal(t, tt.wantDesc, tt.item.Description())
		})
	}
}

func TestToggleSelection_ExclusiveBehavior(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Manifest: config.Manifest{
				Projects: []config.Project{
					{Name: "frontend", Groups: "ui"},
					{Name: "backend", Groups: "api"},
					{Name: "database", Groups: "data"},
				},
			},
		},
		Scopes: map[string]config.Scope{
			"fullstack": {
				Repos:       []string{"frontend", "backend"},
				Description: "Full-stack development",
			},
			"data": {
				Repos:       []string{"database"},
				Description: "Database management",
			},
		},
	}
	
	model := NewStartModelV2(cfg, nil)
	
	// Test 1: Select a scope (radio button behavior)
	scopeItem := StartItemV2{Name: "fullstack", Type: "scope"}
	model.toggleSelection(scopeItem)
	
	assert.Equal(t, "fullstack", model.selectedScope, "Scope should be selected")
	assert.Equal(t, ModeScope, model.selectionMode, "Should be in scope mode")
	assert.Empty(t, model.selectedRepos, "No repos should be selected")
	
	// Test 2: Select another scope (should deselect first)
	scopeItem2 := StartItemV2{Name: "data", Type: "scope"}
	model.toggleSelection(scopeItem2)
	
	assert.Equal(t, "data", model.selectedScope, "New scope should be selected")
	assert.Equal(t, ModeScope, model.selectionMode, "Should still be in scope mode")
	
	// Test 3: Deselect the scope
	model.toggleSelection(scopeItem2)
	
	assert.Equal(t, "", model.selectedScope, "No scope should be selected")
	assert.Equal(t, ModeNone, model.selectionMode, "Should be in none mode")
	
	// Test 4: Select repos (checkbox behavior)
	repoItem1 := StartItemV2{Name: "frontend", Type: "repo"}
	model.toggleSelection(repoItem1)
	
	assert.True(t, model.selectedRepos["frontend"], "Frontend repo should be selected")
	assert.Equal(t, ModeRepo, model.selectionMode, "Should be in repo mode")
	
	repoItem2 := StartItemV2{Name: "backend", Type: "repo"}
	model.toggleSelection(repoItem2)
	
	assert.True(t, model.selectedRepos["frontend"], "Frontend should still be selected")
	assert.True(t, model.selectedRepos["backend"], "Backend should also be selected")
	assert.Equal(t, ModeRepo, model.selectionMode, "Should still be in repo mode")
	
	// Test 5: Select a scope while repos are selected (should clear repos)
	model.toggleSelection(scopeItem)
	
	assert.Equal(t, "fullstack", model.selectedScope, "Scope should be selected")
	assert.Equal(t, ModeScope, model.selectionMode, "Should be in scope mode")
	assert.Empty(t, model.selectedRepos, "All repos should be deselected")
	
	// Test 6: Select a repo while scope is selected (should clear scope)
	model.toggleSelection(repoItem1)
	
	assert.Equal(t, "", model.selectedScope, "Scope should be deselected")
	assert.True(t, model.selectedRepos["frontend"], "Frontend repo should be selected")
	assert.Equal(t, ModeRepo, model.selectionMode, "Should be in repo mode")
}

func TestSwitchMode(t *testing.T) {
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
				Repos: []string{"frontend", "backend"},
			},
		},
	}
	
	model := NewStartModelV2(cfg, nil)
	
	// Start with scope selected
	model.selectedScope = "fullstack"
	model.selectionMode = ModeScope
	
	// Switch to repo mode
	model.switchMode()
	
	assert.Equal(t, "", model.selectedScope, "Scope should be cleared")
	assert.Equal(t, ModeRepo, model.selectionMode, "Should be in repo mode")
	
	// Select some repos
	model.selectedRepos["frontend"] = true
	model.selectedRepos["backend"] = true
	
	// Switch back to scope mode
	model.switchMode()
	
	assert.Empty(t, model.selectedRepos, "Repos should be cleared")
	assert.Equal(t, ModeScope, model.selectionMode, "Should be in scope mode")
}

func TestBuildItems_ShowsAllRepos(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Manifest: config.Manifest{
				Projects: []config.Project{
					{Name: "frontend", Groups: "ui"},
					{Name: "backend", Groups: "api"},
					{Name: "database", Groups: "data"},
					{Name: "cache", Groups: "infra"},
				},
			},
		},
		Scopes: map[string]config.Scope{
			"app": {
				Repos:       []string{"frontend", "backend"},
				Description: "Application scope",
			},
			"infra": {
				Repos:       []string{"database", "cache"},
				Description: "Infrastructure scope",
			},
		},
	}
	
	model := NewStartModelV2(cfg, nil)
	items := model.buildItems()
	
	// Count items by type
	scopes := make(map[string]bool)
	repos := make(map[string]bool)
	
	for _, item := range items {
		if si, ok := item.(StartItemV2); ok {
			if si.Type == "scope" {
				scopes[si.Name] = true
			} else if si.Type == "repo" {
				repos[si.Name] = true
			}
		}
	}
	
	assert.Equal(t, 2, len(scopes), "Should have 2 scopes")
	assert.Equal(t, 4, len(repos), "Should have ALL 4 repos")
	
	// Check that ALL repos are present
	assert.True(t, repos["frontend"], "frontend should be listed")
	assert.True(t, repos["backend"], "backend should be listed")
	assert.True(t, repos["database"], "database should be listed")
	assert.True(t, repos["cache"], "cache should be listed")
}

func TestRepoScopeMapping(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Manifest: config.Manifest{
				Projects: []config.Project{
					{Name: "frontend", Groups: "ui"},
					{Name: "backend", Groups: "api"},
					{Name: "standalone", Groups: "misc"},
				},
			},
		},
		Scopes: map[string]config.Scope{
			"fullstack": {
				Repos: []string{"frontend", "backend"},
			},
		},
	}
	
	model := NewStartModelV2(cfg, nil)
	items := model.buildItems()
	
	// Check that repos show which scope contains them
	for _, item := range items {
		if si, ok := item.(StartItemV2); ok {
			if si.Type == "repo" {
				switch si.Name {
				case "frontend", "backend":
					assert.Equal(t, "fullstack", si.ScopeContaining, 
						"%s should show it's in fullstack scope", si.Name)
				case "standalone":
					assert.Equal(t, "", si.ScopeContaining, 
						"standalone should not be in any scope")
				}
			}
		}
	}
}

func TestGetSelected(t *testing.T) {
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
				Repos: []string{"frontend", "backend"},
			},
		},
	}
	
	model := NewStartModelV2(cfg, nil)
	
	// Test scope selection
	model.selectedScope = "fullstack"
	model.selectionMode = ModeScope
	
	selected := model.GetSelected()
	require.Len(t, selected, 1, "Should have 1 selected item")
	assert.Equal(t, "fullstack", selected[0].Name)
	assert.Equal(t, "scope", selected[0].Type)
	
	// Test repo selection
	model.clearSelection()
	model.selectedRepos["frontend"] = true
	model.selectedRepos["backend"] = true
	model.selectionMode = ModeRepo
	
	selected = model.GetSelected()
	assert.Len(t, selected, 2, "Should have 2 selected items")
	
	selectedNames := make(map[string]bool)
	for _, item := range selected {
		selectedNames[item.Name] = true
		assert.Equal(t, "repo", item.Type)
	}
	assert.True(t, selectedNames["frontend"])
	assert.True(t, selectedNames["backend"])
}

func TestGetSelectionDetails(t *testing.T) {
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
				Repos: []string{"frontend", "backend"},
			},
		},
	}
	
	model := NewStartModelV2(cfg, nil)
	
	// Test scope mode
	model.selectedScope = "fullstack"
	model.selectionMode = ModeScope
	
	assert.Equal(t, ModeScope, model.GetSelectionMode())
	assert.Equal(t, "fullstack", model.GetSelectedScope())
	assert.Empty(t, model.GetSelectedRepos())
	
	// Test repo mode
	model.clearSelection()
	model.selectedRepos["frontend"] = true
	model.selectedRepos["backend"] = true
	model.selectionMode = ModeRepo
	
	assert.Equal(t, ModeRepo, model.GetSelectionMode())
	assert.Equal(t, "", model.GetSelectedScope())
	
	repos := model.GetSelectedRepos()
	assert.Len(t, repos, 2)
	repoMap := make(map[string]bool)
	for _, r := range repos {
		repoMap[r] = true
	}
	assert.True(t, repoMap["frontend"])
	assert.True(t, repoMap["backend"])
}

func TestLaunchSummary(t *testing.T) {
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
				Repos: []string{"frontend", "backend"},
			},
		},
	}
	
	model := NewStartModelV2(cfg, nil)
	model.launching = true
	model.quitting = true
	
	// Test scope launch summary
	model.selectedScope = "fullstack"
	model.selectionMode = ModeScope
	
	summary := model.getLaunchSummary()
	assert.Contains(t, summary, "Starting scope: fullstack")
	
	// Test single repo launch summary
	model.clearSelection()
	model.selectedRepos["frontend"] = true
	model.selectionMode = ModeRepo
	
	summary = model.getLaunchSummary()
	assert.Contains(t, summary, "Starting repository: frontend")
	
	// Test multiple repos launch summary
	model.selectedRepos["backend"] = true
	
	summary = model.getLaunchSummary()
	assert.Contains(t, summary, "Starting repositories:")
	assert.Contains(t, summary, "frontend")
	assert.Contains(t, summary, "backend")
	
	// Test no selection
	model.clearSelection()
	
	summary = model.getLaunchSummary()
	assert.Contains(t, summary, "Nothing selected")
}