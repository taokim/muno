package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
	"gopkg.in/yaml.v3"
)

// ConfigAdapter wraps the existing config package to implement ConfigProvider
type ConfigAdapter struct {
	cache map[string]interface{}
}

// NewConfigAdapter creates a new config adapter
func NewConfigAdapter() interfaces.ConfigProvider {
	return &ConfigAdapter{
		cache: make(map[string]interface{}),
	}
}

// Load loads configuration from a file
func (c *ConfigAdapter) Load(path string) (interface{}, error) {
	// Check cache first
	if cached, ok := c.cache[path]; ok {
		return cached, nil
	}
	
	// Determine config type based on file extension
	ext := filepath.Ext(path)
	
	switch ext {
	case ".yaml", ".yml":
		// Load YAML config
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		
		// Try to load as ConfigTree (muno.yaml)
		if filepath.Base(path) == "muno.yaml" {
			var cfg config.ConfigTree
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("failed to unmarshal config: %w", err)
			}
			c.cache[path] = &cfg
			return &cfg, nil
		}
		
		// Generic YAML loading
		var cfg interface{}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
		}
		c.cache[path] = cfg
		return cfg, nil
		
	case ".json":
		// Load JSON config (if needed in future)
		return nil, fmt.Errorf("JSON config not yet implemented")
		
	default:
		return nil, fmt.Errorf("unsupported config format: %s", ext)
	}
}

// Save saves configuration to a file
func (c *ConfigAdapter) Save(path string, cfg interface{}) error {
	// Update cache
	c.cache[path] = cfg
	
	// Determine format based on extension
	ext := filepath.Ext(path)
	
	switch ext {
	case ".yaml", ".yml":
		// Handle ConfigTree specifically
		if configTree, ok := cfg.(*config.ConfigTree); ok {
			return configTree.Save(path)
		}
		
		// Generic YAML save
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		
		// Ensure directory exists
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		
		return os.WriteFile(path, data, 0644)
		
	case ".json":
		return fmt.Errorf("JSON config not yet implemented")
		
	default:
		return fmt.Errorf("unsupported config format: %s", ext)
	}
}

// Exists checks if a config file exists
func (c *ConfigAdapter) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Watch watches for configuration changes (placeholder for now)
func (c *ConfigAdapter) Watch(path string) (<-chan interfaces.ConfigEvent, error) {
	// This would use fsnotify or similar in a full implementation
	// For now, return a dummy channel
	ch := make(chan interfaces.ConfigEvent)
	close(ch)
	return ch, fmt.Errorf("config watching not yet implemented")
}