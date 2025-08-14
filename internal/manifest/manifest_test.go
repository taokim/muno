package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/repo-claude/internal/config"
)

func TestGenerateManifest(t *testing.T) {
	cfg := config.DefaultConfig("test-project")
	
	xml, err := Generate(cfg)
	require.NoError(t, err)
	
	// Check XML header
	assert.Contains(t, xml, "<?xml version=")
	
	// Check manifest structure
	assert.Contains(t, xml, "<manifest>")
	assert.Contains(t, xml, "</manifest>")
	
	// Check remote
	assert.Contains(t, xml, `<remote name="origin"`)
	assert.Contains(t, xml, `fetch="https://github.com/yourorg/"`)
	
	// Check default
	assert.Contains(t, xml, `<default remote="origin"`)
	assert.Contains(t, xml, `revision="main"`)
	assert.Contains(t, xml, `sync-j="4"`)
	
	// Check projects
	assert.Contains(t, xml, `<project name="backend"`)
	assert.Contains(t, xml, `path="backend"`)
	assert.Contains(t, xml, `groups="core,services"`)
}

func TestCreateManifestRepo(t *testing.T) {
	// Skip if git is not available
	if _, err := os.Stat("/usr/bin/git"); os.IsNotExist(err) {
		t.Skip("git not available")
	}
	
	tmpDir := t.TempDir()
	
	manifestXML := `<?xml version="1.0" encoding="UTF-8"?>
<manifest>
  <remote name="origin" fetch="https://github.com/test/"/>
  <default remote="origin" revision="main" sync-j="4"/>
  <project name="test-repo" path="test-repo"/>
</manifest>`
	
	err := CreateManifestRepo(tmpDir, manifestXML)
	require.NoError(t, err)
	
	// Check manifest repo was created
	manifestDir := filepath.Join(tmpDir, ".manifest-repo")
	assert.DirExists(t, manifestDir)
	
	// Check git repo was initialized
	gitDir := filepath.Join(manifestDir, ".git")
	assert.DirExists(t, gitDir)
	
	// Check manifest file was created
	manifestFile := filepath.Join(manifestDir, "default.xml")
	assert.FileExists(t, manifestFile)
	
	// Verify content
	content, err := os.ReadFile(manifestFile)
	require.NoError(t, err)
	assert.Equal(t, manifestXML, string(content))
}

func TestManifestValidation(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.Config
		shouldErr bool
	}{
		{
			name: "valid config",
			cfg:  config.DefaultConfig("test"),
			shouldErr: false,
		},
		{
			name: "empty projects",
			cfg: &config.Config{
				Workspace: config.WorkspaceConfig{
					Name: "test",
					Manifest: config.Manifest{
						RemoteName:      "origin",
						RemoteFetch:     "https://github.com/test/",
						DefaultRevision: "main",
						Projects:        []config.Project{},
					},
				},
			},
			shouldErr: false, // Empty projects is valid
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xml, err := Generate(tt.cfg)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, xml)
			}
		})
	}
}

func TestManifestEscaping(t *testing.T) {
	cfg := config.DefaultConfig("test")
	// Add project with special characters
	cfg.Workspace.Manifest.Projects = append(cfg.Workspace.Manifest.Projects, config.Project{
		Name:   "test-with-&-special",
		Groups: "test<>group",
	})
	
	xml, err := Generate(cfg)
	require.NoError(t, err)
	
	// Check that special characters are properly escaped
	assert.Contains(t, xml, "test-with-&amp;-special")
	assert.Contains(t, xml, "test&lt;&gt;group")
}