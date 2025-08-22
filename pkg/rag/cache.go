package rag

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// QueryCache implements a cache for RAG query results
type QueryCache struct {
	data    map[string]cacheEntry
	ttl     time.Duration
	maxSize int64
	mu      sync.RWMutex
	stats   CacheStats
}

// cacheEntry represents a cached query result
type cacheEntry struct {
	response  *QueryResponse
	expiresAt time.Time
	size      int64
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	Size        int64   `json:"size"`
	Capacity    int64   `json:"capacity"`
	HitRate     float64 `json:"hit_rate"`
	MemoryUsage int64   `json:"memory_usage"`
}

// NewQueryCache creates a new query cache
func NewQueryCache(ttl time.Duration, maxSize int64) *QueryCache {
	cache := &QueryCache{
		data:    make(map[string]cacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
		stats: CacheStats{
			Capacity: maxSize,
		},
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves cached query results
func (c *QueryCache) Get(query *Query) (*QueryResponse, bool) {
	key := c.generateKey(query)

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
	return entry.response, true
}

// Set stores query results in cache
func (c *QueryCache) Set(query *Query, response *QueryResponse) error {
	key := c.generateKey(query)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check capacity limits
	if c.maxSize > 0 && c.stats.Size >= c.maxSize {
		c.evictOldest()
	}

	// Calculate size (rough estimate)
	size := int64(1) // Minimum size for any entry

	entry := cacheEntry{
		response:  response,
		expiresAt: time.Now().Add(c.ttl),
		size:      size,
	}

	c.data[key] = entry
	c.stats.Size++
	c.stats.MemoryUsage += size

	return nil
}

// Delete removes cached query results
func (c *QueryCache) Delete(query *Query) error {
	key := c.generateKey(query)

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
func (c *QueryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]cacheEntry)
	c.stats.Size = 0
	c.stats.MemoryUsage = 0

	return nil
}

// Stats returns cache statistics
func (c *QueryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	if stats.Size > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	}

	return stats
}

// Close closes the cache
func (c *QueryCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = nil
	return nil
}

// generateKey generates a cache key for a query
func (c *QueryCache) generateKey(query *Query) string {
	// Create a hash of the query to use as cache key
	data, _ := json.Marshal(query)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// evictOldest removes the oldest entries to make room
func (c *QueryCache) evictOldest() {
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
func (c *QueryCache) cleanup() {
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
