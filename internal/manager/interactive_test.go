package manager

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taokim/repo-claude/internal/config"
)

func TestPrompt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple input",
			input:    "test input\n",
			expected: "test input",
		},
		{
			name:     "Input with spaces",
			input:    "  test input  \n",
			expected: "test input",
		},
		{
			name:     "Empty input",
			input:    "\n",
			expected: "",
		},
		{
			name:     "Input with tabs",
			input:    "\ttest\tinput\t\n",
			expected: "test\tinput",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			result := prompt(reader, "Test prompt: ")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInteractiveConfig(t *testing.T) {
	t.Run("AllDefaults", func(t *testing.T) {
		// Simulate pressing Enter for all prompts (use defaults)
		input := strings.Repeat("\n", 20) // Plenty of newlines
		
		// Replace stdin temporarily
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r
		defer func() {
			w.Close()
			os.Stdin = oldStdin
		}()
		
		go func() {
			w.WriteString(input)
			w.Close()
		}()
		
		cfg := config.DefaultConfig("test")
		mgr := &Manager{
			Config: cfg,
		}
		
		err := mgr.interactiveConfig()
		assert.NoError(t, err)
		
		// Config should remain mostly unchanged
		assert.Equal(t, "https://github.com/yourorg/", cfg.Workspace.Manifest.RemoteFetch)
		assert.Len(t, cfg.Workspace.Manifest.Projects, 4)
	})
	
	t.Run("CustomValues", func(t *testing.T) {
		// Simulate custom input
		inputs := []string{
			"https://github.com/myorg/\n",  // Remote fetch
			"api\n",                         // Project 1 name
			"backend,api\n",                 // Project 1 groups
			"api-agent\n",                   // Project 1 agent
			"\n",                            // Project 2 name (default)
			"\n",                            // Project 2 groups (default)
			"\n",                            // Project 2 agent (default)
			"\n",                            // Project 3 name (default)
			"\n",                            // Project 3 groups (default)
			"\n",                            // Project 3 agent (default)
			"\n",                            // Project 4 name (default)
			"\n",                            // Project 4 groups (default)
			"\n",                            // Project 4 agent (default)
			"y\n",                           // Add another project?
			"new-service\n",                 // New project name
			"service,new\n",                 // New project groups
			"service-agent\n",               // New project agent
			"n\n",                           // Add another project?
			"API service specialist\n",      // api-agent specialization
			"\n",                            // api-agent model (default)
			"\n",                            // api-agent auto-start (default)
			"Service handler\n",             // service-agent specialization
			"claude-opus-4\n",               // service-agent model
			"n\n",                           // service-agent auto-start
		}
		
		input := strings.Join(inputs, "")
		
		// Create a pipe to simulate stdin
		r, w, _ := os.Pipe()
		oldStdin := os.Stdin
		os.Stdin = r
		defer func() {
			w.Close()
			os.Stdin = oldStdin
		}()
		
		go func() {
			w.WriteString(input)
			w.Close()
		}()
		
		cfg := config.DefaultConfig("test")
		mgr := &Manager{
			Config: cfg,
		}
		
		err := mgr.interactiveConfig()
		assert.NoError(t, err)
		
		// Check customizations were applied
		assert.Equal(t, "https://github.com/myorg/", cfg.Workspace.Manifest.RemoteFetch)
		assert.Equal(t, "api", cfg.Workspace.Manifest.Projects[0].Name)
		assert.Equal(t, "backend,api", cfg.Workspace.Manifest.Projects[0].Groups)
		assert.Equal(t, "api-agent", cfg.Workspace.Manifest.Projects[0].Agent)
		
		// Check new project was added
		assert.Len(t, cfg.Workspace.Manifest.Projects, 5)
		lastProject := cfg.Workspace.Manifest.Projects[4]
		assert.Equal(t, "new-service", lastProject.Name)
		assert.Equal(t, "service,new", lastProject.Groups)
		assert.Equal(t, "service-agent", lastProject.Agent)
		
		// Check agents were configured
		apiAgent, exists := cfg.Agents["api-agent"]
		assert.True(t, exists)
		assert.Equal(t, "API service specialist", apiAgent.Specialization)
		assert.Equal(t, "claude-sonnet-4", apiAgent.Model)
		assert.True(t, apiAgent.AutoStart)
		
		serviceAgent, exists := cfg.Agents["service-agent"]
		assert.True(t, exists)
		assert.Equal(t, "Service handler", serviceAgent.Specialization)
		assert.Equal(t, "claude-opus-4", serviceAgent.Model)
		assert.False(t, serviceAgent.AutoStart)
	})
}