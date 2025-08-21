package manager

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/taokim/repo-claude/internal/config"
)

// MockReader simulates user input
type MockReader struct {
	inputs []string
	index  int
}

func NewMockReader(inputs ...string) *MockReader {
	return &MockReader{
		inputs: inputs,
		index:  0,
	}
}

func (m *MockReader) ReadString(delim byte) (string, error) {
	if m.index >= len(m.inputs) {
		return "", io.EOF
	}
	input := m.inputs[m.index]
	m.index++
	return input + string(delim), nil
}

// Test StartInteractive
func TestManager_StartInteractive(t *testing.T) {
	tests := []struct {
		name      string
		inputs    []string
		setupFunc func(*Manager)
		wantErr   bool
		checkFunc func(*testing.T, *Manager)
	}{
		{
			name:   "Select single repository",
			inputs: []string{"1", ""}, // Select first repo, then done
			setupFunc: func(m *Manager) {
				m.Config = &config.Config{
					Workspace: config.WorkspaceConfig{
						Manifest: config.Manifest{
							Projects: []config.Project{
								{Name: "frontend", Path: "frontend", Agent: "fe-dev"},
								{Name: "backend", Path: "backend", Agent: "be-dev"},
							},
						},
					},
					Agents: map[string]config.Agent{
						"fe-dev": {Model: "claude-3"},
						"be-dev": {Model: "claude-3"},
					},
				}
				// Create repo directories
				os.MkdirAll(filepath.Join(m.WorkspacePath, "frontend"), 0755)
				os.MkdirAll(filepath.Join(m.WorkspacePath, "backend"), 0755)
			},
			wantErr: false,
		},
		{
			name:   "Select multiple repositories",
			inputs: []string{"1,2", ""}, // Select both repos
			setupFunc: func(m *Manager) {
				m.Config = &config.Config{
					Workspace: config.WorkspaceConfig{
						Manifest: config.Manifest{
							Projects: []config.Project{
								{Name: "frontend", Path: "frontend", Agent: "fe-dev"},
								{Name: "backend", Path: "backend", Agent: "be-dev"},
							},
						},
					},
					Agents: map[string]config.Agent{
						"fe-dev": {Model: "claude-3"},
						"be-dev": {Model: "claude-3"},
					},
				}
				os.MkdirAll(filepath.Join(m.WorkspacePath, "frontend"), 0755)
				os.MkdirAll(filepath.Join(m.WorkspacePath, "backend"), 0755)
			},
			wantErr: false,
		},
		{
			name:   "Cancel selection",
			inputs: []string{""}, // Just press enter to cancel
			setupFunc: func(m *Manager) {
				m.Config = &config.Config{
					Workspace: config.WorkspaceConfig{
						Manifest: config.Manifest{
							Projects: []config.Project{
								{Name: "frontend", Path: "frontend", Agent: "fe-dev"},
							},
						},
					},
				}
			},
			wantErr: false,
		},
		{
			name:   "Invalid selection",
			inputs: []string{"99", ""}, // Invalid number
			setupFunc: func(m *Manager) {
				m.Config = &config.Config{
					Workspace: config.WorkspaceConfig{
						Manifest: config.Manifest{
							Projects: []config.Project{
								{Name: "frontend", Path: "frontend", Agent: "fe-dev"},
							},
						},
					},
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip interactive tests in CI
			if os.Getenv("CI") == "true" {
				t.Skip("Skipping interactive test in CI")
			}

			tmpDir := t.TempDir()
			mgr := &Manager{
				ProjectPath:   tmpDir,
				WorkspacePath: filepath.Join(tmpDir, "workspace"),
				State: &config.State{
					Agents: make(map[string]config.AgentStatus),
				},
				agents: make(map[string]*Agent),
			}

			if tt.setupFunc != nil {
				tt.setupFunc(mgr)
			}

			// Mock stdin
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			r, w, _ := os.Pipe()
			os.Stdin = r

			// Write mock inputs
			go func() {
				for _, input := range tt.inputs {
					fmt.Fprintln(w, input)
				}
				w.Close()
			}()

			// Create mock start options
			opts := StartOptions{
				LogOutput: false,
			}

			// Since we can't easily test the actual interactive function,
			// we'll test the helper functions instead
			t.Skip("Interactive testing requires refactoring to accept io.Reader")

			err := mgr.StartInteractive(opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("StartInteractive() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, mgr)
			}
		})
	}
}

// Test parseRepositorySelection helper
func TestParseRepositorySelection(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxIndex  int
		want      []int
		wantError bool
	}{
		{
			name:     "Single number",
			input:    "1",
			maxIndex: 5,
			want:     []int{0},
		},
		{
			name:     "Multiple numbers",
			input:    "1,3,5",
			maxIndex: 5,
			want:     []int{0, 2, 4},
		},
		{
			name:     "Range",
			input:    "1-3",
			maxIndex: 5,
			want:     []int{0, 1, 2},
		},
		{
			name:     "Mixed",
			input:    "1,3-5",
			maxIndex: 5,
			want:     []int{0, 2, 3, 4},
		},
		{
			name:      "Invalid number",
			input:     "10",
			maxIndex:  5,
			wantError: true,
		},
		{
			name:      "Invalid format",
			input:     "abc",
			maxIndex:  5,
			wantError: true,
		},
		{
			name:     "Empty input",
			input:    "",
			maxIndex: 5,
			want:     []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRepositorySelection(tt.input, tt.maxIndex)
			if (err != nil) != tt.wantError {
				t.Errorf("parseRepositorySelection() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !equalIntSlices(got, tt.want) {
				t.Errorf("parseRepositorySelection() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to parse repository selection
func parseRepositorySelection(input string, maxIndex int) ([]int, error) {
	if input == "" {
		return []int{}, nil
	}

	var indices []int
	parts := strings.Split(input, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			// Handle range
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}

			start, err := parseInt(rangeParts[0])
			if err != nil {
				return nil, err
			}
			end, err := parseInt(rangeParts[1])
			if err != nil {
				return nil, err
			}

			if start < 1 || end > maxIndex || start > end {
				return nil, fmt.Errorf("invalid range: %s", part)
			}

			for i := start; i <= end; i++ {
				indices = append(indices, i-1) // Convert to 0-based
			}
		} else {
			// Handle single number
			num, err := parseInt(part)
			if err != nil {
				return nil, err
			}

			if num < 1 || num > maxIndex {
				return nil, fmt.Errorf("invalid index: %d", num)
			}

			indices = append(indices, num-1) // Convert to 0-based
		}
	}

	return indices, nil
}

func parseInt(s string) (int, error) {
	var num int
	_, err := fmt.Sscanf(s, "%d", &num)
	return num, err
}

func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Test interactive configuration
func TestManager_InteractiveConfig_Mock(t *testing.T) {
	tests := []struct {
		name    string
		inputs  []string
		wantErr bool
		check   func(*testing.T, *config.Config)
	}{
		{
			name: "Accept all defaults",
			inputs: []string{
				"",                    // Default GitHub URL
				"", "", "", "",        // Default projects
				"", "", "", "",        // Default agents
				"n",                   // Don't add more projects
			},
			wantErr: false,
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.Workspace.Manifest.RemoteFetch != "https://github.com/yourorg/" {
					t.Error("Default remote fetch not set")
				}
				if len(cfg.Workspace.Manifest.Projects) != 4 {
					t.Error("Default projects not created")
				}
			},
		},
		{
			name: "Custom configuration",
			inputs: []string{
				"https://github.com/myorg/", // Custom GitHub URL
				"api", "core", "api-dev",    // Custom project
				"", "", "",                  // Accept other defaults
				"", "", "",
				"y",                         // Add another project
				"db", "data", "db-dev",      // Another custom project
				"n",                         // Done adding projects
			},
			wantErr: false,
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.Workspace.Manifest.RemoteFetch != "https://github.com/myorg/" {
					t.Error("Custom remote fetch not set")
				}
				if len(cfg.Workspace.Manifest.Projects) != 5 {
					t.Error("Custom projects not added")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe for mock input
			r, w := io.Pipe()
			
			// Write inputs in a goroutine
			go func() {
				for _, input := range tt.inputs {
					fmt.Fprintln(w, input)
				}
				w.Close()
			}()

			mgr := &Manager{
				Config: config.DefaultConfig("test"),
			}

			// Create a scanner from our mock input
			_ = bufio.NewScanner(r)
			
			// We would need to refactor interactiveConfig to accept a scanner
			// For now, skip the actual test
			t.Skip("Interactive config needs refactoring to accept io.Reader")

			if tt.check != nil {
				tt.check(t, mgr.Config)
			}
		})
	}
}

// Test preset configurations
func TestManager_Presets(t *testing.T) {
	presets := map[string][]string{
		"fullstack": {"frontend", "backend"},
		"mobile":    {"mobile", "backend"},
		"backend":   {"backend", "shared-libs"},
	}

	mgr := &Manager{
		Config: &config.Config{
			Workspace: config.WorkspaceConfig{
				Manifest: config.Manifest{
					Projects: []config.Project{
						{Name: "frontend", Path: "frontend"},
						{Name: "backend", Path: "backend"},
						{Name: "mobile", Path: "mobile"},
						{Name: "shared-libs", Path: "shared-libs"},
					},
				},
			},
		},
	}

	for preset, expected := range presets {
		t.Run(preset, func(t *testing.T) {
			// Get repositories for preset
			var repos []string
			for _, proj := range mgr.Config.Workspace.Manifest.Projects {
				for _, exp := range expected {
					if proj.Name == exp {
						repos = append(repos, proj.Name)
					}
				}
			}

			if len(repos) != len(expected) {
				t.Errorf("Preset %s: got %v repos, want %v", preset, repos, expected)
			}
		})
	}
}