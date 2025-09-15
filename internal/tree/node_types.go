package tree

import (
	"os"
	"path/filepath"
	"strings"
	"github.com/taokim/muno/internal/config"
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
	hasConfig := node.ConfigRef != ""
	
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
	if node.ConfigRef == "" {
		return ""
	}
	
	// If config path is absolute, use it directly
	if filepath.IsAbs(node.ConfigRef) {
		return node.ConfigRef
	}
	
	// Otherwise resolve relative to the node's location
	nodePath := filepath.Join(basePath, node.Name)
	return filepath.Join(nodePath, node.ConfigRef)
}

// IsMetaRepo checks if the repository name indicates it's a meta-repository
// Meta-repos are typically small and should be eagerly cloned to discover their children
func IsMetaRepo(repoName string) bool {
	name := strings.ToLower(repoName)
	
	for _, pattern := range config.GetEagerLoadPatterns() {
		if strings.HasSuffix(name, pattern) {
			return true
		}
	}
	return false
}

// GetEffectiveLazy determines the effective lazy setting for a node
// Default is true (lazy) unless it's a meta-repo or explicitly set to false
func GetEffectiveLazy(node *config.NodeDefinition) bool {
	// Check the fetch mode directly
	switch node.Fetch {
	case config.FetchEager:
		return false
	case config.FetchLazy:
		return true
	case config.FetchAuto, "":
		// Auto mode: meta-repos are eager, others are lazy
		if IsMetaRepo(node.Name) {
			return false
		}
		return true
	default:
		// Default: meta-repos are eager, others are lazy
		if IsMetaRepo(node.Name) {
			return false
		}
		return true
	}
}

// AutoDiscoverConfig checks if a cloned repository has its own muno.yaml
func AutoDiscoverConfig(repoPath string) (string, bool) {
	for _, configName := range config.GetConfigFileNames() {
		configPath := filepath.Join(repoPath, configName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, true
		}
	}
	
	return "", false
}