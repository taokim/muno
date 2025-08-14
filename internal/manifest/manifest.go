package manifest

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yourusername/repo-claude/internal/config"
)

// Manifest represents the Repo tool manifest structure
type Manifest struct {
	XMLName xml.Name `xml:"manifest"`
	Remotes []Remote `xml:"remote"`
	Default Default  `xml:"default"`
	Projects []Project `xml:"project"`
}

// Remote represents a git remote
type Remote struct {
	Name  string `xml:"name,attr"`
	Fetch string `xml:"fetch,attr"`
}

// Default represents default settings
type Default struct {
	Remote   string `xml:"remote,attr"`
	Revision string `xml:"revision,attr"`
	SyncJ    string `xml:"sync-j,attr"`
}

// Project represents a repository project
type Project struct {
	Name   string `xml:"name,attr"`
	Path   string `xml:"path,attr"`
	Groups string `xml:"groups,attr,omitempty"`
}

// Generate creates a manifest XML from configuration
func Generate(cfg *config.Config) (string, error) {
	m := Manifest{
		Remotes: []Remote{
			{
				Name:  cfg.Workspace.Manifest.RemoteName,
				Fetch: cfg.Workspace.Manifest.RemoteFetch,
			},
		},
		Default: Default{
			Remote:   cfg.Workspace.Manifest.RemoteName,
			Revision: cfg.Workspace.Manifest.DefaultRevision,
			SyncJ:    "4",
		},
	}

	// Add projects
	for _, p := range cfg.Workspace.Manifest.Projects {
		m.Projects = append(m.Projects, Project{
			Name:   p.Name,
			Path:   p.Name,
			Groups: p.Groups,
		})
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling manifest: %w", err)
	}

	return xml.Header + string(output), nil
}

// CreateManifestRepo creates a local git repository for the manifest
func CreateManifestRepo(workspacePath string, manifestXML string) error {
	manifestDir := filepath.Join(workspacePath, ".manifest-repo")
	
	// Create directory
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("creating manifest directory: %w", err)
	}

	// Initialize git repo if not exists
	gitDir := filepath.Join(manifestDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		cmd := exec.Command("git", "init")
		cmd.Dir = manifestDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("initializing git repo: %w", err)
		}

		// Configure git user if not set
		checkCmd := exec.Command("git", "config", "user.email")
		checkCmd.Dir = manifestDir
		if err := checkCmd.Run(); err != nil {
			// Set default user
			cmds := [][]string{
				{"git", "config", "user.email", "repo-claude@example.com"},
				{"git", "config", "user.name", "Repo-Claude"},
			}
			for _, args := range cmds {
				cmd := exec.Command(args[0], args[1:]...)
				cmd.Dir = manifestDir
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("configuring git: %w", err)
				}
			}
		}
	}

	// Write manifest file
	manifestFile := filepath.Join(manifestDir, "default.xml")
	if err := os.WriteFile(manifestFile, []byte(manifestXML), 0644); err != nil {
		return fmt.Errorf("writing manifest file: %w", err)
	}

	// Ensure we're on main branch
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchCmd.Dir = manifestDir
	branchOutput, _ := branchCmd.Output()
	if strings.TrimSpace(string(branchOutput)) == "" {
		// No branch yet, create main branch
		checkoutCmd := exec.Command("git", "checkout", "-b", "main")
		checkoutCmd.Dir = manifestDir
		checkoutCmd.Run()
	}

	// Commit manifest
	cmds := [][]string{
		{"git", "add", "default.xml"},
		{"git", "commit", "-m", "Update manifest"},
	}
	
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = manifestDir
		// Ignore error on commit if no changes
		cmd.Run()
	}

	return nil
}