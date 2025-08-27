package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests that don't require actual git operations

func TestManager_Construction(t *testing.T) {
	repos := []Repository{
		{Name: "r1", Path: "r1", URL: "url1", Branch: "main"},
		{Name: "r2", Path: "path/r2", URL: "url2", Branch: "dev"},
	}
	
	mgr := NewManager("/workspace", repos)
	
	if mgr.WorkspacePath != "/workspace" {
		t.Errorf("WorkspacePath = %s, want /workspace", mgr.WorkspacePath)
	}
	
	if len(mgr.Repositories) != 2 {
		t.Errorf("len(Repositories) = %d, want 2", len(mgr.Repositories))
	}
}

func TestCloneRepo_ExistingRepo(t *testing.T) {
	tmpDir := t.TempDir()
	repo := Repository{Name: "test", Path: "test", URL: "url", Branch: "main"}
	
	// Create .git dir to simulate existing repo
	os.MkdirAll(filepath.Join(tmpDir, "test", ".git"), 0755)
	
	mgr := &Manager{WorkspacePath: tmpDir}
	err := mgr.cloneRepo(repo)
	
	if err != nil {
		t.Errorf("cloneRepo() error = %v, want nil for existing repo", err)
	}
}

func TestGetRepoStatus_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	repo := Repository{Name: "missing", Path: "missing"}
	
	mgr := &Manager{WorkspacePath: tmpDir}
	status := mgr.getRepoStatus(repo)
	
	if status.Error == nil {
		t.Error("Expected error for missing repo")
	}
	if !strings.Contains(status.Error.Error(), "not cloned") {
		t.Errorf("Error = %v, want 'not cloned'", status.Error)
	}
}

func TestStatus_Concurrent(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create multiple repos
	var repos []Repository
	for i := 0; i < 10; i++ {
		repos = append(repos, Repository{
			Name: fmt.Sprintf("repo%d", i),
			Path: fmt.Sprintf("repo%d", i),
		})
		
		// Create half as existing
		if i%2 == 0 {
			os.MkdirAll(filepath.Join(tmpDir, fmt.Sprintf("repo%d", i), ".git"), 0755)
		}
	}
	
	mgr := NewManager(tmpDir, repos)
	
	// Test concurrent status calls
	statuses, err := mgr.Status()
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	
	if len(statuses) != 10 {
		t.Errorf("len(statuses) = %d, want 10", len(statuses))
	}
	
	// Count existing vs non-existing
	var existCount, missingCount int
	for _, s := range statuses {
		if s.Error != nil {
			missingCount++
		} else {
			existCount++
		}
	}
	
	if existCount != 5 {
		t.Errorf("existCount = %d, want 5", existCount)
	}
	if missingCount != 5 {
		t.Errorf("missingCount = %d, want 5", missingCount)
	}
}

func TestClone_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create repos that will fail to clone
	repos := []Repository{
		{Name: "fail1", Path: "fail1", URL: "bad-url", Branch: "main"},
		{Name: "fail2", Path: "fail2", URL: "bad-url", Branch: "main"},
	}
	
	mgr := NewManager(tmpDir, repos)
	
	err := mgr.Clone()
	if err == nil {
		t.Error("Clone() should return error for bad URLs")
	}
	if !strings.Contains(err.Error(), "clone errors") {
		t.Errorf("Error should contain 'clone errors', got: %v", err)
	}
}

func TestSync_MixedRepos(t *testing.T) {
	tmpDir := t.TempDir()
	
	repos := []Repository{
		{Name: "existing", Path: "existing", URL: "url", Branch: "main"},
		{Name: "new", Path: "new", URL: "bad-url", Branch: "main"},
	}
	
	// Create one existing repo
	os.MkdirAll(filepath.Join(tmpDir, "existing", ".git"), 0755)
	
	mgr := NewManager(tmpDir, repos)
	
	// Sync will have errors but shouldn't panic
	err := mgr.Sync()
	if err == nil {
		t.Error("Sync() should return error when some repos fail")
	}
}

func TestForAll_HandlesErrors(t *testing.T) {
	tmpDir := t.TempDir()
	
	repos := []Repository{
		{Name: "exists", Path: "exists"},
		{Name: "missing", Path: "missing"},
	}
	
	// Create one repo
	os.MkdirAll(filepath.Join(tmpDir, "exists", ".git"), 0755)
	
	mgr := NewManager(tmpDir, repos)
	
	// ForAll should not error even if command fails
	err := mgr.ForAll("false") // command that always fails
	if err != nil {
		t.Errorf("ForAll() error = %v, want nil (errors are just printed)", err)
	}
}

func TestGetRepositories_Immutable(t *testing.T) {
	repos := []Repository{
		{Name: "r1", Path: "p1"},
		{Name: "r2", Path: "p2"},
	}
	
	mgr := NewManager("/ws", repos)
	
	got := mgr.GetRepositories()
	
	// Modify returned slice
	got[0].Name = "modified"
	
	// Original should be unchanged
	if mgr.Repositories[0].Name != "r1" {
		t.Error("GetRepositories() should return a copy")
	}
}

func TestCloneRepo_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	
	repo := Repository{
		Name:   "nested",
		Path:   "a/b/c/d/nested",
		URL:    "url",
		Branch: "main",
	}
	
	mgr := &Manager{WorkspacePath: tmpDir}
	
	// This will fail to clone but should create directories
	mgr.cloneRepo(repo)
	
	// Check parent directory exists
	parentPath := filepath.Join(tmpDir, "a/b/c/d")
	if _, err := os.Stat(parentPath); os.IsNotExist(err) {
		t.Error("Parent directories should be created")
	}
}

func TestSyncRepo_CloneIfMissing(t *testing.T) {
	tmpDir := t.TempDir()
	
	repo := Repository{
		Name:   "new",
		Path:   "new",
		URL:    "url",
		Branch: "main",
	}
	
	mgr := &Manager{WorkspacePath: tmpDir}
	
	// syncRepo should try to clone if repo doesn't exist
	err := mgr.syncRepo(repo)
	
	// Will fail but that's expected
	if err == nil {
		t.Error("Expected error from git clone in test environment")
	}
}

// Test error aggregation
func TestClone_ErrorAggregation(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Many repos that will fail
	var repos []Repository
	for i := 0; i < 5; i++ {
		repos = append(repos, Repository{
			Name:   fmt.Sprintf("repo%d", i),
			Path:   fmt.Sprintf("repo%d", i),
			URL:    "bad-url",
			Branch: "main",
		})
	}
	
	mgr := NewManager(tmpDir, repos)
	
	err := mgr.Clone()
	if err == nil {
		t.Fatal("Expected error from Clone()")
	}
	
	// Should have errors for all repos
	errStr := err.Error()
	for i := 0; i < 5; i++ {
		if !strings.Contains(errStr, fmt.Sprintf("repo%d", i)) {
			t.Errorf("Error should mention repo%d", i)
		}
	}
}

// Test repository with all fields
func TestRepository_AllFields(t *testing.T) {
	repo := Repository{
		Name:   "test",
		Path:   "custom/path",
		URL:    "https://github.com/example/test.git",
		Branch: "develop",
		Groups: []string{"backend", "core", "api"},
		Agent:  "backend-dev",
	}
	
	// Verify all fields
	if repo.Name != "test" {
		t.Errorf("Name = %s, want test", repo.Name)
	}
	if len(repo.Groups) != 3 {
		t.Errorf("len(Groups) = %d, want 3", len(repo.Groups))
	}
	if repo.Groups[1] != "core" {
		t.Errorf("Groups[1] = %s, want core", repo.Groups[1])
	}
}

// Test status with all fields
func TestStatus_AllFields(t *testing.T) {
	status := Status{
		Name:     "repo",
		Path:     "path/to/repo",
		Branch:   "feature/test",
		Clean:    false,
		Modified: []string{"file1.go", "file2.go", "README.md"},
		Ahead:    5,
		Behind:   2,
		Error:    fmt.Errorf("test error"),
	}
	
	// Verify all fields
	if status.Branch != "feature/test" {
		t.Errorf("Branch = %s, want feature/test", status.Branch)
	}
	if len(status.Modified) != 3 {
		t.Errorf("len(Modified) = %d, want 3", len(status.Modified))
	}
	if status.Ahead != 5 || status.Behind != 2 {
		t.Errorf("Ahead/Behind = %d/%d, want 5/2", status.Ahead, status.Behind)
	}
}

func TestCloneMissing(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create one existing repo and one missing
	existingPath := filepath.Join(tmpDir, "existing")
	err := os.MkdirAll(existingPath, 0755)
	require.NoError(t, err)
	
	repos := []Repository{
		{
			Name:   "existing",
			Path:   "existing",
			URL:    "https://example.com/existing.git",
			Branch: "main",
		},
		{
			Name:   "missing",
			Path:   "missing",
			URL:    "https://example.com/missing.git",
			Branch: "main",
		},
	}
	
	mgr := NewManager(tmpDir, repos)
	
	// CloneMissing should skip existing and try to clone missing
	err = mgr.CloneMissing()
	// Will fail for missing repo but that's expected
	assert.Error(t, err)
}

func TestPush(t *testing.T) {
	tmpDir := t.TempDir()
	
	repos := []Repository{
		{
			Name:   "repo1",
			Path:   "repo1",
			URL:    "https://example.com/repo1.git",
			Branch: "main",
		},
	}
	
	mgr := NewManager(tmpDir, repos)
	
	// Push will fail as repos don't exist
	_, err := mgr.Push(ExecutorOptions{})
	assert.Error(t, err)
}

func TestPushWithOptions(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a fake git repo
	repoPath := filepath.Join(tmpDir, "repo1")
	err := os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	err = cmd.Run()
	require.NoError(t, err)
	
	repos := []Repository{
		{
			Name:   "repo1",
			Path:   "repo1",
			URL:    "https://example.com/repo1.git",
			Branch: "main",
		},
	}
	
	mgr := NewManager(tmpDir, repos)
	
	// Push will fail (no remote) but tests the function
	_, err = mgr.PushWithOptions("origin", "main", ExecutorOptions{
		Parallel:  false,
	})
	assert.Error(t, err)
}

func TestPull(t *testing.T) {
	tmpDir := t.TempDir()
	
	repos := []Repository{
		{
			Name:   "repo1",
			Path:   "repo1",
			URL:    "https://example.com/repo1.git",
			Branch: "main",
		},
	}
	
	mgr := NewManager(tmpDir, repos)
	
	// Pull will fail as repos don't exist
	_, err := mgr.Pull(ExecutorOptions{})
	assert.Error(t, err)
}

func TestPullWithOptions(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a fake git repo
	repoPath := filepath.Join(tmpDir, "repo1")
	err := os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	err = cmd.Run()
	require.NoError(t, err)
	
	repos := []Repository{
		{
			Name:   "repo1",
			Path:   "repo1",
			URL:    "https://example.com/repo1.git",
			Branch: "main",
		},
	}
	
	mgr := NewManager(tmpDir, repos)
	
	// Pull will fail (no remote) but tests the function
	_, err = mgr.PullWithOptions("", "", false, ExecutorOptions{
		Parallel:  false,
	})
	assert.Error(t, err)
}

func TestFetch(t *testing.T) {
	tmpDir := t.TempDir()
	
	repos := []Repository{
		{
			Name:   "repo1",
			Path:   "repo1",
			URL:    "https://example.com/repo1.git",
			Branch: "main",
		},
	}
	
	mgr := NewManager(tmpDir, repos)
	
	// Fetch will fail as repos don't exist
	_, err := mgr.Fetch(ExecutorOptions{})
	assert.Error(t, err)
}

func TestFetchWithOptions(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a fake git repo
	repoPath := filepath.Join(tmpDir, "repo1")
	err := os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	err = cmd.Run()
	require.NoError(t, err)
	
	repos := []Repository{
		{
			Name:   "repo1",
			Path:   "repo1",
			URL:    "https://example.com/repo1.git",
			Branch: "main",
		},
	}
	
	mgr := NewManager(tmpDir, repos)
	
	// Fetch will fail (no remote) but tests the function
	_, err = mgr.FetchWithOptions("", true, false, ExecutorOptions{
		Parallel:  false,
	})
	assert.Error(t, err)
}