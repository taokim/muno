package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	
	"github.com/taokim/muno/internal/config"
)

// TestSimpleConfigNode tests that basic config nodes are recognized and expanded
func TestSimpleConfigNode(t *testing.T) {
	// Create workspace
	workspace, err := os.MkdirTemp("", "muno-simple-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workspace)
	
	// Create external config directory  
	teamDir := filepath.Join(filepath.Dir(workspace), "team")
	if err := os.MkdirAll(teamDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(teamDir)
	
	// Create test repo
	repo1 := createSimpleTestRepo(t, "repo1")
	defer os.RemoveAll(repo1)
	
	// Create team config with one repo (marked as eager to ensure it gets cloned)
	teamConfig := fmt.Sprintf(`workspace:
  name: team
nodes:
  - name: repo1
    url: %s
    fetch: eager
`, repo1)
	writeSimpleFile(t, filepath.Join(teamDir, "muno.yaml"), teamConfig)
	
	// Create main config with config reference
	mainConfig := `workspace:
  name: test-workspace
nodes:
  - name: team
    file: ../team/muno.yaml
`
	writeSimpleFile(t, filepath.Join(workspace, "muno.yaml"), mainConfig)
	
	// Build muno binary
	munoPath := filepath.Join(workspace, "muno")
	buildCmd := exec.Command("go", "build", "-o", munoPath, "../../cmd/muno")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build muno: %v\n%s", err, output)
	}
	
	// Check tree structure
	treeCmd := exec.Command(munoPath, "tree")
	treeCmd.Dir = workspace
	output, err := treeCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Tree command failed: %v\n%s", err, output)
	}
	t.Logf("Tree output:\n%s", output)
	
	// The tree should show team node
	if !strings.Contains(string(output), "team") {
		t.Error("Tree doesn't show team node")
	}
	
	// Clone with --recursive
	cloneCmd := exec.Command(munoPath, "clone", "--recursive")
	cloneCmd.Dir = workspace
	output, err = cloneCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Clone failed: %v\n%s", err, output)
	}
	t.Logf("Clone output:\n%s", output)
	
	// Should have cloned repo1
	if !strings.Contains(string(output), "repo1") {
		t.Error("Clone didn't clone repo1 from config reference")
	}
	
	// Verify repo was actually cloned
	// Get the actual nodes directory from config
	nodesDir := config.GetDefaultNodesDir()
	possiblePaths := []string{
		filepath.Join(workspace, nodesDir, "team", "repo1", ".git"),
		filepath.Join(workspace, "repos", "team", "repo1", ".git"), // legacy fallback
		filepath.Join(workspace, "team", "repo1", ".git"),
	}
	
	found := false
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			t.Logf("Found repo1 at: %s", filepath.Dir(path))
			found = true
			break
		}
	}
	
	if !found {
		// List all directories to see where it went
		filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
			if err == nil && info.IsDir() && filepath.Base(path) == ".git" {
				t.Logf("Found .git directory at: %s", filepath.Dir(path))
			}
			return nil
		})
		t.Error("repo1 was not cloned to expected location")
	}
	
	// Verify that the config node has a symlinked muno.yaml
	teamConfigSymlink := filepath.Join(workspace, nodesDir, "team", "muno.yaml")
	if info, err := os.Lstat(teamConfigSymlink); err != nil {
		t.Errorf("Expected symlink at %s but not found: %v", teamConfigSymlink, err)
	} else if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("Expected %s to be a symlink but it's a regular file", teamConfigSymlink)
	} else {
		// Verify symlink points to the right place
		target, err := os.Readlink(teamConfigSymlink)
		if err != nil {
			t.Errorf("Failed to read symlink: %v", err)
		} else {
			t.Logf("Config symlink points to: %s", target)
			if !strings.Contains(target, "team/muno.yaml") {
				t.Errorf("Symlink points to unexpected location: %s", target)
			}
		}
	}
}

func createSimpleTestRepo(t *testing.T, name string) string {
	repoDir, err := os.MkdirTemp("", fmt.Sprintf("test-repo-%s-*", name))
	if err != nil {
		t.Fatal(err)
	}
	
	// Init repo
	gitInit := exec.Command("git", "init")
	gitInit.Dir = repoDir
	gitInit.Run()
	
	// Configure git
	exec.Command("git", "config", "user.email", "test@example.com").Dir = repoDir
	exec.Command("git", "config", "user.name", "Test User").Dir = repoDir
	
	// Create initial file
	writeSimpleFile(t, filepath.Join(repoDir, "README.md"), fmt.Sprintf("# %s", name))
	
	// Add and commit
	exec.Command("git", "add", ".").Dir = repoDir
	exec.Command("git", "commit", "-m", "Initial commit").Dir = repoDir
	
	return repoDir
}

func writeSimpleFile(t *testing.T, path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}