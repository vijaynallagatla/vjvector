package embedding

import (
	"sync"
	"time"
)

// MemoryCache implements an in-memory cache for embeddings
type MemoryCache struct {
	data      map[string]cacheEntry
	capacity  int64
	maxMemory int64
	mu        sync.RWMutex
	stats     CacheStats
}

// cacheEntry represents a cached item
type cacheEntry struct {
	embeddings [][]float64
	expiresAt  time.Time
	size       int64
}

// NewCache creates a new cache based on configuration
func NewCache(config CacheConfig) (Cache, error) {
	if !config.Enabled {
		return &NoOpCache{}, nil
	}

	switch config.Type {
	case "memory":
		return NewMemoryCache(config), nil
	default:
		return NewMemoryCache(config), nil
	}
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(config CacheConfig) *MemoryCache {
	cache := &MemoryCache{
		data:      make(map[string]cacheEntry),
		capacity:  config.MaxSize,
		maxMemory: config.MaxMemory,
		stats: CacheStats{
			Capacity: config.MaxSize,
		},
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves cached embeddings
func (c *MemoryCache) Get(key string) ([][]float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.data, key)
		c.stats.Size--
		c.mu.Unlock()
		c.mu.RLock()
		c.stats.Misses++
		return nil, false
	}

	c.stats.Hits++
	return entry.embeddings, true
}

// Set stores embeddings in cache
func (c *MemoryCache) Set(key string, embeddings [][]float64, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate size (rough estimate)
	size := int64(len(embeddings) * len(embeddings[0]) * 8) // 8 bytes per float64

	// Check capacity limits
	if c.capacity > 0 && c.stats.Size >= c.capacity {
		c.evictOldest()
	}

	if c.maxMemory > 0 && c.stats.MemoryUsage+size > c.maxMemory {
		c.evictOldest()
	}

	entry := cacheEntry{
		embeddings: embeddings,
		expiresAt:  time.Now().Add(ttl),
		size:       size,
	}

	c.data[key] = entry
	c.stats.Size++
	c.stats.MemoryUsage += size

	return nil
}

// Delete removes cached embeddings
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.data[key]; exists {
		c.stats.MemoryUsage -= entry.size
		c.stats.Size--
		delete(c.data, key)
	}

	return nil
}

// Clear clears all cached data
func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]cacheEntry)
	c.stats.Size = 0
	c.stats.MemoryUsage = 0

	return nil
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	if stats.Size > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	}

	return stats
}

// Close closes the cache
func (c *MemoryCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = nil
	return nil
}

// evictOldest removes the oldest entries to make room
func (c *MemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.data {
		if oldestKey == "" || entry.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.expiresAt
		}
	}

	if oldestKey != "" {
		if entry, exists := c.data[oldestKey]; exists {
			c.stats.MemoryUsage -= entry.size
			c.stats.Size--
			delete(c.data, oldestKey)
		}
	}
}

// cleanup periodically removes expired entries
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.expiresAt) {
				c.stats.MemoryUsage -= entry.size
				c.stats.Size--
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// NoOpCache implements a no-operation cache
type NoOpCache struct{}

func (c *NoOpCache) Get(key string) ([][]float64, bool)                              { return nil, false }
func (c *NoOpCache) Set(key string, embeddings [][]float64, ttl time.Duration) error { return nil }
func (c *NoOpCache) Delete(key string) error                                         { return nil }
func (c *NoOpCache) Clear() error                                                    { return nil }
func (c *NoOpCache) Stats() CacheStats                                               { return CacheStats{} }
func (c *NoOpCache) Close() error                                                    { return nil }
