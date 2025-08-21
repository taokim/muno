package manager

import (
	"testing"

	"github.com/taokim/repo-claude/internal/config"
)

func TestResolveScopeRepos(t *testing.T) {
	m := &Manager{
		Config: &config.Config{
			Workspace: config.WorkspaceConfig{
				Manifest: config.Manifest{
					Projects: []config.Project{
						{Name: "backend-api"},
						{Name: "backend-db"},
						{Name: "frontend-web"},
						{Name: "frontend-mobile"},
						{Name: "shared-lib"},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		patterns []string
		expected []string
	}{
		{
			name:     "Exact match",
			patterns: []string{"backend-api", "frontend-web"},
			expected: []string{"backend-api", "frontend-web"},
		},
		{
			name:     "Wildcard match",
			patterns: []string{"backend-*"},
			expected: []string{"backend-api", "backend-db"},
		},
		{
			name:     "Multiple prefix wildcards",
			patterns: []string{"frontend-*", "shared-*"},
			expected: []string{"frontend-mobile", "frontend-web", "shared-lib"},
		},
		{
			name:     "All repos",
			patterns: []string{"*"},
			expected: []string{"backend-api", "backend-db", "frontend-mobile", "frontend-web", "shared-lib"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.resolveScopeRepos(tt.patterns)
			
			// Sort both slices for comparison
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d repos, got %d", len(tt.expected), len(result))
				return
			}
			
			// Check all expected repos are present
			for _, exp := range tt.expected {
				found := false
				for _, res := range result {
					if res == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected repo %s not found in result", exp)
				}
			}
		})
	}
}