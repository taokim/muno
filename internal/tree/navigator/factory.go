package navigator

import (
	"fmt"
	"time"

	"github.com/taokim/muno/internal/config"
	"github.com/taokim/muno/internal/interfaces"
)

// NavigatorType specifies the type of navigator to create
type NavigatorType string

const (
	// TypeFilesystem creates a filesystem-based navigator (default)
	TypeFilesystem NavigatorType = "filesystem"
	
	// TypeCached creates a cached navigator wrapping filesystem
	TypeCached NavigatorType = "cached"
	
	// TypeInMemory creates an in-memory navigator (for testing)
	TypeInMemory NavigatorType = "inmemory"
)

// Factory creates TreeNavigator instances based on configuration
type Factory struct {
	workspace string
	config    *config.ConfigTree
	gitCmd    interfaces.GitInterface
}

// NewFactory creates a new navigator factory
func NewFactory(workspace string, cfg *config.ConfigTree, gitCmd interfaces.GitInterface) *Factory {
	return &Factory{
		workspace: workspace,
		config:    cfg,
		gitCmd:    gitCmd,
	}
}

// Create creates a navigator based on the specified type and options
func (f *Factory) Create(navType NavigatorType, opts *NavigatorOptions) (TreeNavigator, error) {
	if opts == nil {
		opts = f.defaultOptions()
	}

	switch navType {
	case TypeFilesystem:
		return f.createFilesystemNavigator(opts)
	
	case TypeCached:
		return f.createCachedNavigator(opts)
	
	case TypeInMemory:
		return f.createInMemoryNavigator(opts)
	
	default:
		return nil, fmt.Errorf("unknown navigator type: %s", navType)
	}
}

// CreateFromConfig creates a navigator based on configuration
func (f *Factory) CreateFromConfig() (TreeNavigator, error) {
	opts := f.parseConfigOptions()
	
	// Determine navigator type from config
	navType := TypeFilesystem // Default
	
	if f.config != nil {
		// Check for navigator configuration in the config tree
		// Note: ConfigTree doesn't have Metadata field currently
		// This would need to be added to support navigator config
		if false {
			navConfig := map[string]interface{}{}
			if typeStr, ok := navConfig["type"].(string); ok {
				navType = NavigatorType(typeStr)
			}
			
			// Override cache settings if specified
			if cacheConfig, ok := navConfig["cache"].(map[string]interface{}); ok {
				if enabled, ok := cacheConfig["enabled"].(bool); ok && enabled {
					navType = TypeCached
					opts.CacheEnabled = true
				}
				if ttlStr, ok := cacheConfig["ttl"].(string); ok {
					if ttl, err := time.ParseDuration(ttlStr); err == nil {
						opts.CacheTTL = ttl
					}
				}
				if maxSize, ok := cacheConfig["max_size"].(int); ok {
					opts.MaxCacheSize = maxSize
				}
			}
		}
	}

	return f.Create(navType, opts)
}

// CreateDefault creates the default navigator (filesystem)
func (f *Factory) CreateDefault() (TreeNavigator, error) {
	return f.Create(TypeFilesystem, f.defaultOptions())
}

// CreateWithCache creates a cached navigator with specified TTL
func (f *Factory) CreateWithCache(ttl time.Duration) (TreeNavigator, error) {
	opts := f.defaultOptions()
	opts.CacheEnabled = true
	opts.CacheTTL = ttl
	return f.Create(TypeCached, opts)
}

// CreateForTesting creates an in-memory navigator for testing
func (f *Factory) CreateForTesting() (TreeNavigator, error) {
	return f.Create(TypeInMemory, nil)
}

// Private helper methods

func (f *Factory) createFilesystemNavigator(opts *NavigatorOptions) (TreeNavigator, error) {
	workspace := opts.WorkspacePath
	if workspace == "" {
		workspace = f.workspace
	}

	return NewFilesystemNavigator(workspace, f.config, f.gitCmd)
}

func (f *Factory) createCachedNavigator(opts *NavigatorOptions) (TreeNavigator, error) {
	// Create base filesystem navigator
	base, err := f.createFilesystemNavigator(opts)
	if err != nil {
		return nil, fmt.Errorf("creating base navigator: %w", err)
	}

	// Wrap with cache
	ttl := opts.CacheTTL
	if ttl == 0 {
		ttl = 30 * time.Second // Default TTL
	}

	cached := NewCachedNavigator(base, ttl)
	
	if opts.MaxCacheSize > 0 {
		cached.WithMaxSize(opts.MaxCacheSize)
	}

	// Start cleanup routine if refresh interval is set
	if opts.RefreshInterval > 0 {
		cached.StartCleanupRoutine(opts.RefreshInterval)
	}

	return cached, nil
}

func (f *Factory) createInMemoryNavigator(opts *NavigatorOptions) (TreeNavigator, error) {
	return NewInMemoryNavigator(), nil
}

func (f *Factory) defaultOptions() *NavigatorOptions {
	return &NavigatorOptions{
		WorkspacePath:   f.workspace,
		ConfigPath:      "", // Will use default search
		CacheEnabled:    false,
		CacheTTL:        30 * time.Second,
		MaxCacheSize:    1000,
		LazyLoadTimeout: 5 * time.Minute,
		RefreshInterval: 5 * time.Minute,
	}
}

func (f *Factory) parseConfigOptions() *NavigatorOptions {
	opts := f.defaultOptions()

	if f.config == nil {
		return opts
	}

	// Parse navigator configuration from config metadata
	// Note: ConfigTree doesn't have Metadata field currently
	if false {
		navConfig := map[string]interface{}{}
		// Parse cache settings
		if cacheConfig, ok := navConfig["cache"].(map[string]interface{}); ok {
			if enabled, ok := cacheConfig["enabled"].(bool); ok {
				opts.CacheEnabled = enabled
			}
			if ttlStr, ok := cacheConfig["ttl"].(string); ok {
				if ttl, err := time.ParseDuration(ttlStr); err == nil {
					opts.CacheTTL = ttl
				}
			}
			if maxSize, ok := cacheConfig["max_size"].(int); ok {
				opts.MaxCacheSize = maxSize
			}
		}

		// Parse timeout settings
		if timeoutStr, ok := navConfig["lazy_load_timeout"].(string); ok {
			if timeout, err := time.ParseDuration(timeoutStr); err == nil {
				opts.LazyLoadTimeout = timeout
			}
		}

		// Parse refresh interval
		if intervalStr, ok := navConfig["refresh_interval"].(string); ok {
			if interval, err := time.ParseDuration(intervalStr); err == nil {
				opts.RefreshInterval = interval
			}
		}
	}

	return opts
}

// Quick creation functions for common use cases

// NewFilesystem creates a filesystem navigator directly
func NewFilesystem(workspace string, cfg *config.ConfigTree, gitCmd interfaces.GitInterface) (TreeNavigator, error) {
	return NewFilesystemNavigator(workspace, cfg, gitCmd)
}

// NewCached creates a cached navigator with default settings
func NewCached(workspace string, cfg *config.ConfigTree, gitCmd interfaces.GitInterface) (TreeNavigator, error) {
	base, err := NewFilesystemNavigator(workspace, cfg, gitCmd)
	if err != nil {
		return nil, err
	}
	return NewCachedNavigator(base, 30*time.Second), nil
}

// NewInMemory creates an in-memory navigator for testing
func NewInMemory() TreeNavigator {
	return NewInMemoryNavigator()
}