package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initTestRepo initializes a test git repository
func initTestRepo(t *testing.T, path string) {
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}
	
	// Set git config
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = path
	cmd.Run()
	
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = path
	cmd.Run()
	
	// Create initial commit
	testFile := filepath.Join(path, "README.md")
	os.WriteFile(testFile, []byte("# Test"), 0644)
	
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = path
	cmd.Run()
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = path
	cmd.Run()
}

func TestDefaultExecutorOptions(t *testing.T) {
	opts := DefaultExecutorOptions()
	
	if !opts.IncludeRoot {
		t.Error("DefaultExecutorOptions should include root by default")
	}
	if !opts.Parallel {
		t.Error("DefaultExecutorOptions should be parallel by default")
	}
	if opts.MaxParallel != 4 {
		t.Errorf("DefaultExecutorOptions MaxParallel = %d, want 4", opts.MaxParallel)
	}
	if opts.FailFast {
		t.Error("DefaultExecutorOptions should not fail fast by default")
	}
}

func TestBuildTargetPaths(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	
	// Create root .git
	os.MkdirAll(filepath.Join(tmpDir, ".git"), 0755)
	
	// Create workspace repos
	repo1Dir := filepath.Join(workspaceDir, "repo1")
	repo2Dir := filepath.Join(workspaceDir, "repo2")
	os.MkdirAll(filepath.Join(repo1Dir, ".git"), 0755)
	os.MkdirAll(filepath.Join(repo2Dir, ".git"), 0755)
	
	mgr := &Manager{
		WorkspacePath: workspaceDir,
		Repositories: []Repository{
			{Name: "repo1", Path: "repo1"},
			{Name: "repo2", Path: "repo2"},
			{Name: "repo3", Path: "repo3"}, // Not cloned
		},
	}
	
	tests := []struct {
		name        string
		opts        ExecutorOptions
		wantTargets int
		wantRoot    bool
	}{
		{
			name: "include root",
			opts: ExecutorOptions{
				IncludeRoot: true,
				ExcludeRoot: false,
			},
			wantTargets: 3, // root + repo1 + repo2
			wantRoot:    true,
		},
		{
			name: "exclude root",
			opts: ExecutorOptions{
				IncludeRoot: true,
				ExcludeRoot: true,
			},
			wantTargets: 2, // repo1 + repo2
			wantRoot:    false,
		},
		{
			name: "no root option",
			opts: ExecutorOptions{
				IncludeRoot: false,
				ExcludeRoot: false,
			},
			wantTargets: 2, // repo1 + repo2
			wantRoot:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targets := mgr.buildTargetPaths(tt.opts)
			
			if len(targets) != tt.wantTargets {
				t.Errorf("buildTargetPaths() returned %d targets, want %d", len(targets), tt.wantTargets)
			}
			
			hasRoot := false
			for _, target := range targets {
				if target.Name == "root" {
					hasRoot = true
					break
				}
			}
			
			if hasRoot != tt.wantRoot {
				t.Errorf("buildTargetPaths() hasRoot = %v, want %v", hasRoot, tt.wantRoot)
			}
		})
	}
}

func TestExecuteCommand(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a git repo
	repoDir := filepath.Join(tmpDir, "test-repo")
	os.MkdirAll(repoDir, 0755)
	
	// Initialize git repo
	initTestRepo(t, repoDir)
	
	mgr := &Manager{
		WorkspacePath: tmpDir,
	}
	
	target := targetPath{
		Name: "test-repo",
		Path: repoDir,
	}
	
	// Test successful command
	result := mgr.executeCommand("status", []string{"-s"}, target, ExecutorOptions{Quiet: true})
	
	if !result.Success {
		t.Errorf("executeCommand failed: %v", result.Error)
	}
	if result.RepoName != "test-repo" {
		t.Errorf("executeCommand RepoName = %s, want test-repo", result.RepoName)
	}
	if result.RepoPath != repoDir {
		t.Errorf("executeCommand RepoPath = %s, want %s", result.RepoPath, repoDir)
	}
	
	// Test failed command
	result = mgr.executeCommand("invalid-command", []string{}, target, ExecutorOptions{Quiet: true})
	
	if result.Success {
		t.Error("executeCommand should have failed for invalid command")
	}
	if result.Error == nil {
		t.Error("executeCommand should have returned an error for invalid command")
	}
}

func TestExecuteInRepos_NoRepos(t *testing.T) {
	mgr := &Manager{
		WorkspacePath: t.TempDir(),
		Repositories:  []Repository{},
	}
	
	_, err := mgr.ExecuteInRepos("status", []string{}, DefaultExecutorOptions())
	if err == nil {
		t.Error("ExecuteInRepos should error when no repositories")
	}
}

func TestExecuteInRepos_Parallel(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	
	// Create multiple repos
	for i := 1; i <= 3; i++ {
		repoDir := filepath.Join(workspaceDir, fmt.Sprintf("repo%d", i))
		os.MkdirAll(repoDir, 0755)
		initTestRepo(t, repoDir)
	}
	
	mgr := &Manager{
		WorkspacePath: workspaceDir,
		Repositories: []Repository{
			{Name: "repo1", Path: "repo1"},
			{Name: "repo2", Path: "repo2"},
			{Name: "repo3", Path: "repo3"},
		},
	}
	
	opts := ExecutorOptions{
		Parallel:    true,
		MaxParallel: 2,
		Quiet:       true,
	}
	
	results, err := mgr.ExecuteInRepos("status", []string{"-s"}, opts)
	if err != nil {
		t.Fatalf("ExecuteInRepos failed: %v", err)
	}
	
	if len(results) != 3 {
		t.Errorf("ExecuteInRepos returned %d results, want 3", len(results))
	}
	
	for _, result := range results {
		if !result.Success {
			t.Errorf("Command failed for %s: %v", result.RepoName, result.Error)
		}
	}
}

func TestExecuteInRepos_Sequential(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	
	// Create a single repo
	repoDir := filepath.Join(workspaceDir, "repo1")
	os.MkdirAll(repoDir, 0755)
	initTestRepo(t, repoDir)
	
	mgr := &Manager{
		WorkspacePath: workspaceDir,
		Repositories: []Repository{
			{Name: "repo1", Path: "repo1"},
		},
	}
	
	opts := ExecutorOptions{
		Parallel: false,
		Quiet:    true,
	}
	
	results, err := mgr.ExecuteInRepos("status", []string{"-s"}, opts)
	if err != nil {
		t.Fatalf("ExecuteInRepos failed: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("ExecuteInRepos returned %d results, want 1", len(results))
	}
	
	if !results[0].Success {
		t.Errorf("Command failed: %v", results[0].Error)
	}
}

func TestCommit(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	
	// Create a repo with changes
	repoDir := filepath.Join(workspaceDir, "repo1")
	os.MkdirAll(repoDir, 0755)
	initTestRepo(t, repoDir)
	
	// Add a new file
	testFile := filepath.Join(repoDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)
	
	mgr := &Manager{
		WorkspacePath: workspaceDir,
		Repositories: []Repository{
			{Name: "repo1", Path: "repo1"},
		},
	}
	
	opts := ExecutorOptions{
		Quiet: true,
	}
	
	// Test commit with empty message
	_, err := mgr.Commit("", opts)
	if err == nil {
		t.Error("Commit should fail with empty message")
	}
	
	// Test commit with message
	results, err := mgr.Commit("Test commit", opts)
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}
	
	if len(results) > 0 && !results[0].Success {
		t.Errorf("Commit command failed: %v", results[0].Error)
	}
}

func TestForAllWithOptions(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	
	// Create a repo
	repoDir := filepath.Join(workspaceDir, "repo1")
	os.MkdirAll(repoDir, 0755)
	initTestRepo(t, repoDir)
	
	mgr := &Manager{
		WorkspacePath: workspaceDir,
		Repositories: []Repository{
			{Name: "repo1", Path: "repo1"},
		},
	}
	
	// Test git command
	err := mgr.ForAllWithOptions("git", []string{"status", "-s"}, DefaultExecutorOptions())
	if err != nil {
		t.Errorf("ForAllWithOptions failed for git command: %v", err)
	}
	
	// Test non-git command (should use old implementation)
	err = mgr.ForAllWithOptions("echo", []string{"test"}, DefaultExecutorOptions())
	if err != nil {
		t.Errorf("ForAllWithOptions failed for non-git command: %v", err)
	}
	
	// Test git command with no args
	err = mgr.ForAllWithOptions("git", []string{}, DefaultExecutorOptions())
	if err == nil {
		t.Error("ForAllWithOptions should fail with no git command")
	}
}