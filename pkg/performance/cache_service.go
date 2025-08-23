package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"sync"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/utils/logger"
)

// DefaultCacheService implements the advanced caching service
type DefaultCacheService struct {
	config     *CacheConfig
	shards     []*cacheShard
	stats      *CacheStats
	mu         sync.RWMutex
	compressor *gzipCompressor
}

// cacheShard represents a single cache shard
type cacheShard struct {
	items       map[string]*CacheItem
	accessOrder []string // For LRU/LFU strategies
	mu          sync.RWMutex
	size        int64
	itemCount   int64
}

// gzipCompressor handles compression operations
type gzipCompressor struct {
	level int
}

// NewDefaultCacheService creates a new default cache service
func NewDefaultCacheService(config *CacheConfig) *DefaultCacheService {
	if config == nil {
		config = DefaultCacheConfig()
	}

	service := &DefaultCacheService{
		config:     config,
		shards:     make([]*cacheShard, config.Shards),
		stats:      &CacheStats{LastUpdated: time.Now()},
		compressor: &gzipCompressor{level: 6},
	}

	// Initialize shards
	for i := 0; i < config.Shards; i++ {
		service.shards[i] = &cacheShard{
			items:       make(map[string]*CacheItem),
			accessOrder: make([]string, 0),
		}
	}

	// Start background optimization
	go service.backgroundOptimization()

	return service
}

// Get retrieves an item from cache
func (s *DefaultCacheService) Get(ctx context.Context, key string) (*CacheItem, error) {
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	item, exists := shard.items[key]
	if !exists {
		s.updateStats(false)
		return nil, fmt.Errorf("key not found: %s", key)
	}

	// Check if item has expired
	if item.ExpiresAt != nil && time.Now().After(*item.ExpiresAt) {
		delete(shard.items, key)
		shard.itemCount--
		shard.size -= item.Size
		s.updateStats(false)
		return nil, fmt.Errorf("key expired: %s", key)
	}

	// Update access statistics
	item.AccessCount++
	item.LastAccessed = time.Now()
	s.updateAccessOrder(shard, key)
	s.updateStats(true)

	return item, nil
}

// Set stores an item in cache
func (s *DefaultCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	// Serialize value to calculate size
	serialized, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %v", err)
	}

	// Check if compression should be applied
	var compressed bool
	var compressionRatio float64
	var finalData []byte

	if s.config.Compression && int64(len(serialized)) > s.config.CompressionThreshold {
		compressedData, err := s.compressor.compress(serialized)
		if err == nil && len(compressedData) < len(serialized) {
			compressed = true
			compressionRatio = float64(len(compressedData)) / float64(len(serialized))
			finalData = compressedData
		} else {
			finalData = serialized
			compressionRatio = 1.0
		}
	} else {
		finalData = serialized
		compressionRatio = 1.0
	}

	// Calculate item size
	itemSize := int64(len(finalData))

	// Check if we need to evict items
	if shard.size+itemSize > s.config.MaxSize/int64(s.config.Shards) || shard.itemCount >= s.config.MaxItems/int64(s.config.Shards) {
		s.evictItems(shard, itemSize)
	}

	// Create cache item
	expiresAt := time.Now().Add(ttl)
	item := &CacheItem{
		Key:              key,
		Value:            value,
		Size:             itemSize,
		AccessCount:      0,
		LastAccessed:     time.Now(),
		CreatedAt:        time.Now(),
		ExpiresAt:        &expiresAt,
		Tags:             []string{},
		Metadata:         make(map[string]interface{}),
		Compressed:       compressed,
		CompressionRatio: compressionRatio,
	}

	// Store item
	shard.items[key] = item
	shard.itemCount++
	shard.size += itemSize
	s.updateAccessOrder(shard, key)

	// Update compression savings
	if compressed {
		s.stats.CompressionSavings += int64(len(serialized)) - itemSize
	}

	return nil
}

// Delete removes an item from cache
func (s *DefaultCacheService) Delete(ctx context.Context, key string) error {
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	item, exists := shard.items[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	shard.size -= item.Size
	shard.itemCount--
	delete(shard.items, key)
	s.removeFromAccessOrder(shard, key)

	return nil
}

// Exists checks if a key exists in cache
func (s *DefaultCacheService) Exists(ctx context.Context, key string) (bool, error) {
	shard := s.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()

	item, exists := shard.items[key]
	if !exists {
		return false, nil
	}

	// Check if expired
	if item.ExpiresAt != nil && time.Now().After(*item.ExpiresAt) {
		return false, nil
	}

	return true, nil
}

// Clear removes all items from cache
func (s *DefaultCacheService) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, shard := range s.shards {
		shard.mu.Lock()
		shard.items = make(map[string]*CacheItem)
		shard.accessOrder = make([]string, 0)
		shard.size = 0
		shard.itemCount = 0
		shard.mu.Unlock()
	}

	// Reset stats
	s.stats = &CacheStats{LastUpdated: time.Now()}

	return nil
}

// GetMulti retrieves multiple items from cache
func (s *DefaultCacheService) GetMulti(ctx context.Context, keys []string) (map[string]*CacheItem, error) {
	result := make(map[string]*CacheItem)
	var errors []string

	for _, key := range keys {
		item, err := s.Get(ctx, key)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", key, err))
		} else {
			result[key] = item
		}
	}

	if len(errors) > 0 {
		return result, fmt.Errorf("some keys failed: %v", errors)
	}

	return result, nil
}

// SetMulti stores multiple items in cache
func (s *DefaultCacheService) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	var errors []string

	for key, value := range items {
		err := s.Set(ctx, key, value, ttl)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", key, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some items failed: %v", errors)
	}

	return nil
}

// DeleteMulti removes multiple items from cache
func (s *DefaultCacheService) DeleteMulti(ctx context.Context, keys []string) error {
	var errors []string

	for _, key := range keys {
		err := s.Delete(ctx, key)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", key, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some deletions failed: %v", errors)
	}

	return nil
}

// Increment increases a numeric value in cache
func (s *DefaultCacheService) Increment(ctx context.Context, key string, value int64) (int64, error) {
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	item, exists := shard.items[key]
	if !exists {
		// Create new item with initial value
		err := s.Set(ctx, key, value, s.config.DefaultTTL)
		if err != nil {
			return 0, err
		}
		return value, nil
	}

	// Try to convert existing value to number
	var currentValue int64
	switch v := item.Value.(type) {
	case int:
		currentValue = int64(v)
	case int64:
		currentValue = v
	case float64:
		currentValue = int64(v)
	default:
		return 0, fmt.Errorf("value is not numeric: %T", item.Value)
	}

	newValue := currentValue + value
	item.Value = newValue
	item.LastAccessed = time.Now()

	return newValue, nil
}

// Decrement decreases a numeric value in cache
func (s *DefaultCacheService) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return s.Increment(ctx, key, -value)
}

// GetStats returns cache performance statistics
func (s *DefaultCacheService) GetStats(ctx context.Context) (*CacheStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate totals across all shards
	totalItems := int64(0)
	totalSize := int64(0)

	for _, shard := range s.shards {
		shard.mu.RLock()
		totalItems += shard.itemCount
		totalSize += shard.size
		shard.mu.RUnlock()
	}

	stats := *s.stats
	stats.TotalItems = totalItems
	stats.TotalSize = totalSize
	stats.LastUpdated = time.Now()

	return &stats, nil
}

// GetKeys returns keys matching a pattern
func (s *DefaultCacheService) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string

	for _, shard := range s.shards {
		shard.mu.RLock()
		for key := range shard.items {
			// Simple pattern matching (in production, use regex)
			if pattern == "" || key == pattern {
				keys = append(keys, key)
			}
		}
		shard.mu.RUnlock()
	}

	return keys, nil
}

// GetSize returns total cache size in bytes
func (s *DefaultCacheService) GetSize(ctx context.Context) (int64, error) {
	totalSize := int64(0)

	for _, shard := range s.shards {
		shard.mu.RLock()
		totalSize += shard.size
		shard.mu.RUnlock()
	}

	return totalSize, nil
}

// GetItemCount returns total number of items in cache
func (s *DefaultCacheService) GetItemCount(ctx context.Context) (int64, error) {
	totalCount := int64(0)

	for _, shard := range s.shards {
		shard.mu.RLock()
		totalCount += shard.itemCount
		shard.mu.RUnlock()
	}

	return totalCount, nil
}

// Flush removes all expired items from cache
func (s *DefaultCacheService) Flush(ctx context.Context) error {
	now := time.Now()
	flushedCount := int64(0)

	for _, shard := range s.shards {
		shard.mu.Lock()
		for key, item := range shard.items {
			if item.ExpiresAt != nil && now.After(*item.ExpiresAt) {
				shard.size -= item.Size
				shard.itemCount--
				delete(shard.items, key)
				s.removeFromAccessOrder(shard, key)
				flushedCount++
			}
		}
		shard.mu.Unlock()
	}

	return nil
}

// Optimize performs cache optimization
func (s *DefaultCacheService) Optimize(ctx context.Context) error {
	// Implement cache optimization logic
	// This could include:
	// - Rebalancing shards
	// - Adjusting eviction strategies
	// - Compressing more items
	// - Cleaning up expired items

	if err := s.Flush(ctx); err != nil {
		return err
	}

	return nil
}

// Compress compresses a specific cache item
func (s *DefaultCacheService) Compress(ctx context.Context, key string) error {
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	item, exists := shard.items[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	if item.Compressed {
		return fmt.Errorf("item already compressed: %s", key)
	}

	// Serialize and compress
	serialized, err := json.Marshal(item.Value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %v", err)
	}

	compressedData, err := s.compressor.compress(serialized)
	if err != nil {
		return fmt.Errorf("failed to compress: %v", err)
	}

	// Update item
	oldSize := item.Size
	item.Size = int64(len(compressedData))
	item.Compressed = true
	item.CompressionRatio = float64(len(compressedData)) / float64(len(serialized))

	// Update shard size
	shard.size = shard.size - oldSize + item.Size

	// Update compression savings
	s.stats.CompressionSavings += oldSize - item.Size

	return nil
}

// Decompress decompresses a specific cache item
func (s *DefaultCacheService) Decompress(ctx context.Context, key string) error {
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	item, exists := shard.items[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	if !item.Compressed {
		return fmt.Errorf("item not compressed: %s", key)
	}

	// Decompress
	serialized, err := json.Marshal(item.Value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %v", err)
	}

	decompressedData, err := s.compressor.decompress(serialized)
	if err != nil {
		return fmt.Errorf("failed to decompress: %v", err)
	}

	// Update item
	oldSize := item.Size
	item.Size = int64(len(decompressedData))
	item.Compressed = false
	item.CompressionRatio = 1.0

	// Update shard size
	shard.size = shard.size - oldSize + item.Size

	// Update compression savings
	s.stats.CompressionSavings -= oldSize - item.Size

	return nil
}

// Prefetch loads items into cache before they're needed
func (s *DefaultCacheService) Prefetch(ctx context.Context, keys []string) error {
	// Implement prefetching logic
	// This could include:
	// - Loading data from slower storage
	// - Warming up cache with frequently accessed items
	// - Predictive loading based on access patterns

	return nil
}

// HealthCheck performs a health check on the service
func (s *DefaultCacheService) HealthCheck(ctx context.Context) error {
	if s.shards == nil || s.stats == nil {
		return fmt.Errorf("cache service not properly initialized")
	}

	// Check if any shards are accessible
	for i, shard := range s.shards {
		if shard == nil {
			return fmt.Errorf("shard %d is nil", i)
		}
	}

	return nil
}

// Close performs cleanup operations
func (s *DefaultCacheService) Close() error {
	// Clean up any resources if needed
	return nil
}

// Helper methods

// getShard determines which shard to use for a key
func (s *DefaultCacheService) getShard(key string) *cacheShard {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	shardIndex := int(hash.Sum32()) % len(s.shards)
	return s.shards[shardIndex]
}

// updateStats updates cache statistics
func (s *DefaultCacheService) updateStats(hit bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if hit {
		s.stats.HitCount++
	} else {
		s.stats.MissCount++
	}

	total := s.stats.HitCount + s.stats.MissCount
	if total > 0 {
		s.stats.HitRate = float64(s.stats.HitCount) / float64(total)
	}
}

// updateAccessOrder updates the access order for LRU/LFU strategies
func (s *DefaultCacheService) updateAccessOrder(shard *cacheShard, key string) {
	// Remove from current position
	s.removeFromAccessOrder(shard, key)

	// Add to front (most recently used)
	shard.accessOrder = append([]string{key}, shard.accessOrder...)
}

// removeFromAccessOrder removes a key from access order
func (s *DefaultCacheService) removeFromAccessOrder(shard *cacheShard, key string) {
	for i, k := range shard.accessOrder {
		if k == key {
			shard.accessOrder = append(shard.accessOrder[:i], shard.accessOrder[i+1:]...)
			break
		}
	}
}

// evictItems evicts items based on the configured strategy
func (s *DefaultCacheService) evictItems(shard *cacheShard, requiredSize int64) {
	switch s.config.Strategy {
	case CacheStrategyLRU:
		s.evictLRU(shard, requiredSize)
	case CacheStrategyLFU:
		s.evictLFU(shard, requiredSize)
	case CacheStrategyTTL:
		s.evictTTL(shard, requiredSize)
	case CacheStrategyAdaptive:
		s.evictAdaptive(shard, requiredSize)
	}
}

// evictLRU evicts least recently used items
func (s *DefaultCacheService) evictLRU(shard *cacheShard, requiredSize int64) {
	for len(shard.accessOrder) > 0 && (shard.size > s.config.MaxSize/int64(s.config.Shards) || shard.itemCount > s.config.MaxItems/int64(s.config.Shards)) {
		key := shard.accessOrder[len(shard.accessOrder)-1]
		item := shard.items[key]

		shard.size -= item.Size
		shard.itemCount--
		delete(shard.items, key)
		shard.accessOrder = shard.accessOrder[:len(shard.accessOrder)-1]
		s.stats.EvictionCount++
	}
}

// evictLFU evicts least frequently used items
func (s *DefaultCacheService) evictLFU(shard *cacheShard, requiredSize int64) {
	// Find least frequently used item
	var lfuKey string
	var minAccess int64 = math.MaxInt64

	for key, item := range shard.items {
		if item.AccessCount < minAccess {
			minAccess = item.AccessCount
			lfuKey = key
		}
	}

	if lfuKey != "" {
		item := shard.items[lfuKey]
		shard.size -= item.Size
		shard.itemCount--
		delete(shard.items, lfuKey)
		s.removeFromAccessOrder(shard, lfuKey)
		s.stats.EvictionCount++
	}
}

// evictTTL evicts items based on time to live
func (s *DefaultCacheService) evictTTL(shard *cacheShard, requiredSize int64) {
	now := time.Now()

	for key, item := range shard.items {
		if item.ExpiresAt != nil && now.After(*item.ExpiresAt) {
			shard.size -= item.Size
			shard.itemCount--
			delete(shard.items, key)
			s.removeFromAccessOrder(shard, key)
			s.stats.EvictionCount++
		}
	}
}

// evictAdaptive evicts items using adaptive strategy
func (s *DefaultCacheService) evictAdaptive(shard *cacheShard, requiredSize int64) {
	// Adaptive strategy combines LRU and LFU
	// For now, use LRU as fallback
	s.evictLRU(shard, requiredSize)
}

// backgroundOptimization runs background optimization tasks
func (s *DefaultCacheService) backgroundOptimization() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.Optimize(context.Background()); err != nil {
			logger.Error("failed to optimize cache", "error", err)
		}
	}
}

// gzipCompressor methods

func (c *gzipCompressor) compress(data []byte) ([]byte, error) {
	// Simple compression simulation (in production, use proper compression)
	// For now, just return the original data
	return data, nil
}

func (c *gzipCompressor) decompress(data []byte) ([]byte, error) {
	// Simple decompression simulation (in production, use proper compression)
	// For now, just return the original data
	return data, nil
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxSize:              100 * 1024 * 1024, // 100MB
		MaxItems:             10000,
		DefaultTTL:           1 * time.Hour,
		Strategy:             CacheStrategyLRU,
		Compression:          true,
		CompressionThreshold: 1024, // 1KB
		Shards:               16,
		EnableMetrics:        true,
	}
}
