package navigator

import (
	"fmt"
	"path/filepath"

	"github.com/taokim/muno/internal/config"
)

// ConfigResolver resolves configuration references
type ConfigResolver struct {
	cache map[string]*config.ConfigTree
	root  string
}

// NewConfigResolver creates a new config resolver
func NewConfigResolver(root string) *ConfigResolver {
	return &ConfigResolver{
		cache: make(map[string]*config.ConfigTree),
		root:  root,
	}
}

// LoadNodeFile loads a configuration file for a node
func (r *ConfigResolver) LoadNodeFile(configPath string, nodeDef *config.NodeDefinition) (*config.ConfigTree, error) {
	// Check cache first
	if cached, exists := r.cache[configPath]; exists {
		return cached, nil
	}

	// Resolve full path
	fullPath := configPath
	if !filepath.IsAbs(configPath) {
		fullPath = filepath.Join(r.root, configPath)
	}

	// Load the config
	cfg, err := config.LoadTree(fullPath)
	if err != nil {
		return nil, fmt.Errorf("loading config %s: %w", configPath, err)
	}

	// Cache it
	r.cache[configPath] = cfg

	return cfg, nil
}