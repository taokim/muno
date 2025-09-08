package navigator

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// CacheEntry holds cached data with expiration
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// CachedNavigator wraps any TreeNavigator with caching capabilities.
// It provides improved performance by caching node and status information.
type CachedNavigator struct {
	base     TreeNavigator
	cache    map[string]*CacheEntry
	mu       sync.RWMutex
	ttl      time.Duration
	maxSize  int
}

// NewCachedNavigator creates a new caching wrapper around a base navigator
func NewCachedNavigator(base TreeNavigator, ttl time.Duration) *CachedNavigator {
	return &CachedNavigator{
		base:    base,
		cache:   make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: 1000, // Default max size
	}
}

// WithMaxSize sets the maximum cache size
func (c *CachedNavigator) WithMaxSize(size int) *CachedNavigator {
	c.maxSize = size
	return c
}

// GetCurrentPath returns the current position in the tree
func (c *CachedNavigator) GetCurrentPath() (string, error) {
	// Current path should not be cached as it's stateful
	return c.base.GetCurrentPath()
}

// Navigate changes the current position to the specified path
func (c *CachedNavigator) Navigate(path string) error {
	// Navigation changes state, so clear relevant caches
	err := c.base.Navigate(path)
	if err == nil {
		c.invalidatePath(path)
	}
	return err
}

// GetNode retrieves a single node by its path with caching
func (c *CachedNavigator) GetNode(path string) (*Node, error) {
	cacheKey := fmt.Sprintf("node:%s", path)
	
	// Check cache
	if cached := c.getFromCache(cacheKey); cached != nil {
		if node, ok := cached.(*Node); ok {
			return node, nil
		}
	}

	// Get from base navigator
	node, err := c.base.GetNode(path)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if node != nil {
		c.putInCache(cacheKey, node)
	}

	return node, nil
}

// ListChildren returns all direct children of a node with caching
func (c *CachedNavigator) ListChildren(path string) ([]*Node, error) {
	cacheKey := fmt.Sprintf("children:%s", path)
	
	// Check cache
	if cached := c.getFromCache(cacheKey); cached != nil {
		if children, ok := cached.([]*Node); ok {
			return children, nil
		}
	}

	// Get from base navigator
	children, err := c.base.ListChildren(path)
	if err != nil {
		return nil, err
	}

	// Cache the result
	c.putInCache(cacheKey, children)

	return children, nil
}

// GetTree returns a tree view with caching
func (c *CachedNavigator) GetTree(path string, depth int) (*TreeView, error) {
	cacheKey := fmt.Sprintf("tree:%s:%d", path, depth)
	
	// Check cache
	if cached := c.getFromCache(cacheKey); cached != nil {
		if tree, ok := cached.(*TreeView); ok {
			return tree, nil
		}
	}

	// Get from base navigator
	tree, err := c.base.GetTree(path, depth)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if tree != nil {
		c.putInCache(cacheKey, tree)
	}

	return tree, nil
}

// GetNodeStatus returns the current status of a node with caching
func (c *CachedNavigator) GetNodeStatus(path string) (*NodeStatus, error) {
	cacheKey := fmt.Sprintf("status:%s", path)
	
	// Check cache
	if cached := c.getFromCache(cacheKey); cached != nil {
		if status, ok := cached.(*NodeStatus); ok {
			return status, nil
		}
	}

	// Get from base navigator
	status, err := c.base.GetNodeStatus(path)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if status != nil {
		c.putInCache(cacheKey, status)
	}

	return status, nil
}

// RefreshStatus forces a status refresh and clears cache
func (c *CachedNavigator) RefreshStatus(path string) error {
	// Clear cache for this path and its status
	c.invalidatePath(path)
	c.invalidateStatus(path)
	
	// Refresh in base navigator
	return c.base.RefreshStatus(path)
}

// IsLazy checks if a node is configured for lazy loading
func (c *CachedNavigator) IsLazy(path string) (bool, error) {
	// This is typically configuration-based and stable, so cache it
	cacheKey := fmt.Sprintf("lazy:%s", path)
	
	// Check cache
	if cached := c.getFromCache(cacheKey); cached != nil {
		if lazy, ok := cached.(bool); ok {
			return lazy, nil
		}
	}

	// Get from base navigator
	lazy, err := c.base.IsLazy(path)
	if err != nil {
		return false, err
	}

	// Cache the result
	c.putInCache(cacheKey, lazy)

	return lazy, nil
}

// TriggerLazyLoad initiates loading of a lazy node
func (c *CachedNavigator) TriggerLazyLoad(path string) error {
	// Lazy loading changes state, so clear caches
	err := c.base.TriggerLazyLoad(path)
	if err == nil {
		c.invalidatePath(path)
		c.invalidateStatus(path)
		c.invalidateChildren(path)
	}
	return err
}

// ClearCache removes all cached entries
func (c *CachedNavigator) ClearCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*CacheEntry)
}

// ClearExpired removes expired cache entries
func (c *CachedNavigator) ClearExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.cache {
		if entry.IsExpired() {
			delete(c.cache, key)
		}
	}
}

// GetCacheStats returns cache statistics
func (c *CachedNavigator) GetCacheStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{
		Size:    len(c.cache),
		MaxSize: c.maxSize,
	}

	// Count expired entries
	for _, entry := range c.cache {
		if entry.IsExpired() {
			stats.Expired++
		}
	}

	return stats
}

// CacheStats holds cache statistics
type CacheStats struct {
	Size    int
	MaxSize int
	Expired int
	Hits    int64
	Misses  int64
}

// Helper methods for cache management

func (c *CachedNavigator) getFromCache(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, exists := c.cache[key]; exists {
		if !entry.IsExpired() {
			return entry.Data
		}
	}
	return nil
}

func (c *CachedNavigator) putInCache(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check cache size limit
	if len(c.cache) >= c.maxSize {
		// Simple eviction: remove expired entries first
		for k, entry := range c.cache {
			if entry.IsExpired() {
				delete(c.cache, k)
				if len(c.cache) < c.maxSize {
					break
				}
			}
		}

		// If still over limit, remove oldest entries
		if len(c.cache) >= c.maxSize {
			// Find and remove the oldest entry
			var oldestKey string
			var oldestTime time.Time
			for k, entry := range c.cache {
				if oldestKey == "" || entry.ExpiresAt.Before(oldestTime) {
					oldestKey = k
					oldestTime = entry.ExpiresAt
				}
			}
			if oldestKey != "" {
				delete(c.cache, oldestKey)
			}
		}
	}

	c.cache[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *CachedNavigator) invalidatePath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove node cache
	delete(c.cache, fmt.Sprintf("node:%s", path))
	
	// Remove tree caches that include this path
	for key := range c.cache {
		if len(key) > 5 && key[:5] == "tree:" {
			// Tree caches start with "tree:"
			delete(c.cache, key)
		}
	}
}

func (c *CachedNavigator) invalidateStatus(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, fmt.Sprintf("status:%s", path))
}

func (c *CachedNavigator) invalidateChildren(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, fmt.Sprintf("children:%s", path))
	
	// Also invalidate parent's children cache
	if path != "/" && path != "" {
		parentPath := path[:strings.LastIndex(path, "/")]
		if parentPath == "" {
			parentPath = "/"
		}
		delete(c.cache, fmt.Sprintf("children:%s", parentPath))
	}
}

// StartCleanupRoutine starts a background routine to clean expired cache entries
func (c *CachedNavigator) StartCleanupRoutine(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			c.ClearExpired()
		}
	}()
}