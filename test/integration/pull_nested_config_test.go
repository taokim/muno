package integration_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	
	"github.com/taokim/muno/internal/config"
)

// TestPullNestedConfigRecursive verifies that pull --recursive works
// properly with nested config nodes (config nodes referencing other config nodes)
func TestPullNestedConfigRecursive(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Create workspace
	workspace, err := os.MkdirTemp("", "muno-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workspace)

	// Set up a complex hierarchy with nested config references:
	// root
	// ├── team-frontend (config reference to frontend/muno.yaml)
	// │   ├── web-app (git repo)
	// │   └── mobile-app (git repo)  
	// └── team-backend (config reference to backend/muno.yaml)
	//     ├── api-service (git repo)
	//     └── infra (config reference to backend/infra/muno.yaml)
	//         ├── monitoring (git repo)
	//         └── logging (git repo)

	// Create external config directories
	frontendDir := filepath.Join(filepath.Dir(workspace), "frontend")
	backendDir := filepath.Join(filepath.Dir(workspace), "backend")
	infraDir := filepath.Join(backendDir, "infra")
	
	if err := os.MkdirAll(frontendDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(infraDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test repos
	webRepo := createTestGitRepo(t, "web-app")
	mobileRepo := createTestGitRepo(t, "mobile-app")
	apiRepo := createTestGitRepo(t, "api-service")
	monitoringRepo := createTestGitRepo(t, "monitoring")
	loggingRepo := createTestGitRepo(t, "logging")

	// Create frontend config (mark repos as eager to ensure they get cloned)
	frontendConfig := fmt.Sprintf(`workspace:
  name: team-frontend
nodes:
  - name: web-app
    url: %s
    fetch: eager
  - name: mobile-app
    url: %s
    fetch: eager
`, webRepo, mobileRepo)
	writeFile(t, filepath.Join(frontendDir, "muno.yaml"), frontendConfig)

	// Create infrastructure config (nested config - mark as eager)
	infraConfig := fmt.Sprintf(`workspace:
  name: infra
nodes:
  - name: monitoring
    url: %s
    fetch: eager
  - name: logging
    url: %s
    fetch: eager
`, monitoringRepo, loggingRepo)
	writeFile(t, filepath.Join(infraDir, "muno.yaml"), infraConfig)

	// Create backend config with nested config reference
	backendConfig := fmt.Sprintf(`workspace:
  name: team-backend
nodes:
  - name: api-service
    url: %s
    fetch: eager
  - name: infra
    file: ./infra/muno.yaml
`, apiRepo)
	writeFile(t, filepath.Join(backendDir, "muno.yaml"), backendConfig)

	// Create main workspace config
	mainConfig := `workspace:
  name: test-workspace
nodes:
  - name: team-frontend
    file: ../frontend/muno.yaml
  - name: team-backend
    file: ../backend/muno.yaml
`
	writeFile(t, filepath.Join(workspace, "muno.yaml"), mainConfig)

	// Build muno binary
	munoPath := filepath.Join(workspace, "muno")
	buildCmd := exec.Command("go", "build", "-o", munoPath, "../../cmd/muno")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build muno: %v\n%s", err, output)
	}

	// First, check tree structure before clone
	treeCmd := exec.Command(munoPath, "tree")
	treeCmd.Dir = workspace
	treeOutput, _ := treeCmd.CombinedOutput()
	t.Logf("Tree before clone:\n%s", treeOutput)
	
	// First, clone with --recursive to set up the repositories
	cloneCmd := exec.Command(munoPath, "clone", "--recursive")
	cloneCmd.Dir = workspace
	output, err := cloneCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to clone repositories: %v\nOutput: %s", err, output)
	}
	t.Logf("Clone output:\n%s", output)
	
	// Check what was actually cloned
	if err := filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && filepath.Base(path) == ".git" {
			t.Logf("Found git repo at: %s", filepath.Dir(path))
		}
		return nil
	}); err != nil {
		t.Logf("Error walking workspace: %v", err)
	}

	// Make changes in all repositories
	// Use the actual nodes directory from config
	nodesDir := config.GetDefaultNodesDir()
	repos := []string{
		filepath.Join(nodesDir, "team-frontend", "web-app"),
		filepath.Join(nodesDir, "team-frontend", "mobile-app"),
		filepath.Join(nodesDir, "team-backend", "api-service"),
		filepath.Join(nodesDir, "team-backend", "infra", ".nodes", "monitoring"),
		filepath.Join(nodesDir, "team-backend", "infra", ".nodes", "logging"),
	}

	for _, repo := range repos {
		// Make a change in the original repo
		origRepoName := filepath.Base(repo)
		var origRepo string
		switch origRepoName {
		case "web-app":
			origRepo = webRepo
		case "mobile-app":
			origRepo = mobileRepo
		case "api-service":
			origRepo = apiRepo
		case "monitoring":
			origRepo = monitoringRepo
		case "logging":
			origRepo = loggingRepo
		}
		
		// Add a new file to the original repo
		testFile := filepath.Join(origRepo, "update.txt")
		writeFile(t, testFile, "updated content")
		
		// Commit in original repo
		gitAdd := exec.Command("git", "add", ".")
		gitAdd.Dir = origRepo
		if output, err := gitAdd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to add changes in %s: %v\n%s", origRepo, err, output)
		}
		
		gitCommit := exec.Command("git", "commit", "-m", "Update")
		gitCommit.Dir = origRepo
		if output, err := gitCommit.CombinedOutput(); err != nil {
			t.Fatalf("Failed to commit in %s: %v\n%s", origRepo, err, output)
		}
	}

	// Now test pull --recursive
	pullCmd := exec.Command(munoPath, "pull", "--recursive")
	pullCmd.Dir = workspace
	output, err = pullCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Pull failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	
	// Check that all repos were pulled, including those in nested config nodes
	expectedRepos := []string{
		"web-app",
		"mobile-app", 
		"api-service",
		"monitoring",  // From nested infra config
		"logging",     // From nested infra config
	}

	for _, repo := range expectedRepos {
		if !strings.Contains(outputStr, fmt.Sprintf("Pulling: %s", repo)) {
			t.Errorf("Expected to pull %s but it wasn't in output:\n%s", repo, outputStr)
		}
	}

	// Verify files were actually pulled
	for _, repo := range repos {
		updateFile := filepath.Join(workspace, repo, "update.txt")
		if _, err := os.Stat(updateFile); os.IsNotExist(err) {
			t.Errorf("Update file not found in %s after pull", repo)
		}
	}
}


// Helper functions
func createTestGitRepo(t *testing.T, name string) string {
	repoDir, err := os.MkdirTemp("", fmt.Sprintf("test-repo-%s-*", name))
	if err != nil {
		t.Fatal(err)
	}

	// Init repo
	gitInit := exec.Command("git", "init")
	gitInit.Dir = repoDir
	if output, err := gitInit.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init repo %s: %v\n%s", name, err, output)
	}

	// Configure git
	gitConfig := exec.Command("git", "config", "user.email", "test@example.com")
	gitConfig.Dir = repoDir
	gitConfig.Run()
	
	gitConfig = exec.Command("git", "config", "user.name", "Test User")
	gitConfig.Dir = repoDir
	gitConfig.Run()

	// Create initial file
	testFile := filepath.Join(repoDir, "README.md")
	writeFile(t, testFile, fmt.Sprintf("# %s", name))

	// Add and commit
	gitAdd := exec.Command("git", "add", ".")
	gitAdd.Dir = repoDir
	if output, err := gitAdd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to add files in %s: %v\n%s", name, err, output)
	}

	gitCommit := exec.Command("git", "commit", "-m", "Initial commit")
	gitCommit.Dir = repoDir
	if output, err := gitCommit.CombinedOutput(); err != nil {
		t.Fatalf("Failed to commit in %s: %v\n%s", name, err, output)
	}

	return repoDir
}

func writeFile(t *testing.T, path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func copyDir(t *testing.T, src, dst string) {
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}
	
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		
		dstPath := filepath.Join(dst, relPath)
		
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()
		
		_, err = io.Copy(dstFile, srcFile)
		return err
	})
	
	if err != nil {
		t.Fatal(err)
	}
}