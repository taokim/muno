package tree

import (
	"os"
	"path/filepath"
	"strings"
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/constants"
)

// NodeKind represents the kind of node based on field presence
type NodeKind int

const (
	NodeKindRepo      NodeKind = iota  // URL only: git repository (may auto-discover config)
	NodeKindConfigRef                  // Config only: pure config delegation
	NodeKindInvalid                    // Both or neither: invalid configuration
)

// GetNodeKind determines the node type from its definition
func GetNodeKind(node *config.NodeDefinition) NodeKind {
	hasURL := node.URL != ""
	hasConfig := node.Config != ""
	
	switch {
	case hasURL && !hasConfig:
		return NodeKindRepo
	case hasConfig && !hasURL:
		return NodeKindConfigRef
	default:
		// Both or neither is invalid
		return NodeKindInvalid
	}
}

// ResolveConfigPath resolves the config path relative to the current location
func ResolveConfigPath(basePath string, node *config.NodeDefinition) string {
	if node.Config == "" {
		return ""
	}
	
	// If config path is absolute, use it directly
	if filepath.IsAbs(node.Config) {
		return node.Config
	}
	
	// Otherwise resolve relative to the node's location
	nodePath := filepath.Join(basePath, node.Name)
	return filepath.Join(nodePath, node.Config)
}

// IsMetaRepo checks if the repository name indicates it's a meta-repository
// Meta-repos are typically small and should be eagerly cloned to discover their children
func IsMetaRepo(repoName string) bool {
	name := strings.ToLower(repoName)
	
	for _, pattern := range constants.EagerLoadPatterns {
		if strings.HasSuffix(name, pattern) {
			return true
		}
	}
	return false
}

// GetEffectiveLazy determines the effective lazy setting for a node
// Default is true (lazy) unless it's a meta-repo or explicitly set to false
func GetEffectiveLazy(node *config.NodeDefinition) bool {
	// If explicitly set, use that value
	if node.Lazy {
		return true
	}
	
	// Meta-repos default to eager (lazy=false)
	if IsMetaRepo(node.Name) {
		return false
	}
	
	// Everything else defaults to lazy
	return true
}

// AutoDiscoverConfig checks if a cloned repository has its own muno.yaml
func AutoDiscoverConfig(repoPath string) (string, bool) {
	for _, configName := range constants.ConfigFileNames {
		configPath := filepath.Join(repoPath, configName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, true
		}
	}
	
	return "", false
}