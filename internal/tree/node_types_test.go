package tree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taokim/muno/internal/config"
)

func TestGetNodeKind(t *testing.T) {
	tests := []struct {
		name     string
		node     *config.NodeDefinition
		expected NodeKind
	}{
		{
			name: "node with URL only",
			node: &config.NodeDefinition{
				Name: "repo",
				URL:  "https://github.com/test/repo.git",
			},
			expected: NodeKindRepo,
		},
		{
			name: "node with config only",
			node: &config.NodeDefinition{
				Name:   "parent",
				Config: "muno.yaml",
			},
			expected: NodeKindConfigRef,
		},
		{
			name: "node with both URL and config",
			node: &config.NodeDefinition{
				Name:   "hybrid",
				URL:    "https://github.com/test/hybrid.git",
				Config: "muno.yaml",
			},
			expected: NodeKindInvalid,
		},
		{
			name: "node with neither URL nor config",
			node: &config.NodeDefinition{
				Name: "empty",
			},
			expected: NodeKindInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNodeKind(tt.node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveConfigPath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		basePath string
		node     *config.NodeDefinition
		expected string
	}{
		{
			name:     "node with relative config",
			basePath: tmpDir,
			node: &config.NodeDefinition{
				Name:   "test",
				Config: "custom/config.yaml",
			},
			expected: filepath.Join(tmpDir, "test", "custom", "config.yaml"),
		},
		{
			name:     "node with absolute config",
			basePath: tmpDir,
			node: &config.NodeDefinition{
				Name:   "test",
				Config: "/absolute/path/config.yaml",
			},
			expected: "/absolute/path/config.yaml",
		},
		{
			name:     "node with no config",
			basePath: tmpDir,
			node: &config.NodeDefinition{
				Name: "test",
			},
			expected: "",
		},
		{
			name:     "node with simple config filename",
			basePath: "/workspace",
			node: &config.NodeDefinition{
				Name:   "myrepo",
				Config: "muno.yaml",
			},
			expected: filepath.Join("/workspace", "myrepo", "muno.yaml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveConfigPath(tt.basePath, tt.node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsMetaRepo(t *testing.T) {
	tests := []struct {
		name     string
		repoName string
		expected bool
	}{
		{
			name:     "monorepo suffix",
			repoName: "backend-monorepo",
			expected: true,
		},
		{
			name:     "muno suffix",
			repoName: "platform-muno",
			expected: true,
		},
		{
			name:     "metarepo suffix",
			repoName: "services-metarepo",
			expected: true,
		},
		{
			name:     "platform suffix",
			repoName: "enterprise-platform",
			expected: true,
		},
		{
			name:     "workspace suffix",
			repoName: "dev-workspace",
			expected: true,
		},
		{
			name:     "root-repo suffix",
			repoName: "company-root-repo",
			expected: true,
		},
		{
			name:     "regular repo",
			repoName: "payment-service",
			expected: false,
		},
		{
			name:     "case insensitive",
			repoName: "BACKEND-MONOREPO",
			expected: true,
		},
		{
			name:     "contains but not suffix",
			repoName: "monorepo-backend",
			expected: false,
		},
		{
			name:     "empty name",
			repoName: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMetaRepo(tt.repoName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEffectiveLazy(t *testing.T) {
	tests := []struct {
		name         string
		node         *config.NodeDefinition
		expectedLazy bool
	}{
		{
			name: "node with explicit lazy true",
			node: &config.NodeDefinition{
				Name: "test",
				Lazy: true,
			},
			expectedLazy: true,
		},
		{
			name: "node with explicit lazy false",
			node: &config.NodeDefinition{
				Name: "test",
				Lazy: false,
			},
			expectedLazy: true, // Default is true unless meta-repo
		},
		{
			name: "meta-repo is eager by default",
			node: &config.NodeDefinition{
				Name: "backend-monorepo",
				Lazy: false,
			},
			expectedLazy: false,
		},
		{
			name: "meta-repo can be lazy if explicitly set",
			node: &config.NodeDefinition{
				Name: "backend-monorepo",
				Lazy: true,
			},
			expectedLazy: true,
		},
		{
			name: "regular repo defaults to lazy",
			node: &config.NodeDefinition{
				Name: "payment-service",
				Lazy: false,
			},
			expectedLazy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEffectiveLazy(tt.node)
			assert.Equal(t, tt.expectedLazy, result)
		})
	}
}

func TestAutoDiscoverConfig(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		repoPath     string
		setup        func()
		expectedPath string
		expectedOk   bool
	}{
		{
			name:     "discovers muno.yaml",
			repoPath: filepath.Join(tmpDir, "nodes", "test"),
			setup: func() {
				configPath := filepath.Join(tmpDir, "nodes", "test", "muno.yaml")
				os.MkdirAll(filepath.Dir(configPath), 0755)
				os.WriteFile(configPath, []byte("test"), 0644)
			},
			expectedPath: filepath.Join(tmpDir, "nodes", "test", "muno.yaml"),
			expectedOk:   true,
		},
		{
			name:     "discovers .muno.yaml",
			repoPath: filepath.Join(tmpDir, "nodes", "test2"),
			setup: func() {
				configPath := filepath.Join(tmpDir, "nodes", "test2", ".muno.yaml")
				os.MkdirAll(filepath.Dir(configPath), 0755)
				os.WriteFile(configPath, []byte("test"), 0644)
			},
			expectedPath: filepath.Join(tmpDir, "nodes", "test2", ".muno.yaml"),
			expectedOk:   true,
		},
		{
			name:     "discovers muno.yml",
			repoPath: filepath.Join(tmpDir, "nodes", "test3"),
			setup: func() {
				configPath := filepath.Join(tmpDir, "nodes", "test3", "muno.yml")
				os.MkdirAll(filepath.Dir(configPath), 0755)
				os.WriteFile(configPath, []byte("test"), 0644)
			},
			expectedPath: filepath.Join(tmpDir, "nodes", "test3", "muno.yml"),
			expectedOk:   true,
		},
		{
			name:     "discovers .muno.yml",
			repoPath: filepath.Join(tmpDir, "nodes", "test4"),
			setup: func() {
				configPath := filepath.Join(tmpDir, "nodes", "test4", ".muno.yml")
				os.MkdirAll(filepath.Dir(configPath), 0755)
				os.WriteFile(configPath, []byte("test"), 0644)
			},
			expectedPath: filepath.Join(tmpDir, "nodes", "test4", ".muno.yml"),
			expectedOk:   true,
		},
		{
			name:     "no config file found",
			repoPath: filepath.Join(tmpDir, "nodes", "test5"),
			setup: func() {
				os.MkdirAll(filepath.Join(tmpDir, "nodes", "test5"), 0755)
			},
			expectedPath: "",
			expectedOk:   false,
		},
		{
			name:         "empty path",
			repoPath:     "",
			setup:        func() {},
			expectedPath: "",
			expectedOk:   false,
		},
		{
			name:     "prefers muno.yaml over other formats",
			repoPath: filepath.Join(tmpDir, "nodes", "test6"),
			setup: func() {
				dir := filepath.Join(tmpDir, "nodes", "test6")
				os.MkdirAll(dir, 0755)
				// Create multiple config files
				os.WriteFile(filepath.Join(dir, "muno.yaml"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(dir, ".muno.yaml"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(dir, "muno.yml"), []byte("test"), 0644)
			},
			expectedPath: filepath.Join(tmpDir, "nodes", "test6", "muno.yaml"),
			expectedOk:   true,
		},
		{
			name:     "non-existent directory",
			repoPath: filepath.Join(tmpDir, "nonexistent"),
			setup:    func() {},
			expectedPath: "",
			expectedOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			os.RemoveAll(filepath.Join(tmpDir, "nodes"))

			if tt.setup != nil {
				tt.setup()
			}

			path, ok := AutoDiscoverConfig(tt.repoPath)
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedPath, path)
		})
	}
}