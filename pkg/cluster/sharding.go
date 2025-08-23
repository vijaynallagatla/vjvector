package cluster

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

// HashSharding implements hash-based sharding strategy
type HashSharding struct {
	mu         sync.RWMutex
	shardCount int
	shards     map[int]*ShardInfo
}

// ShardInfo represents information about a shard
type ShardInfo struct {
	ID          int      `json:"id"`
	Status      string   `json:"status"`
	NodeIDs     []string `json:"node_ids"`
	VectorCount int64    `json:"vector_count"`
	SizeBytes   int64    `json:"size_bytes"`
	LastUpdated int64    `json:"last_updated"`
}

// NewHashSharding creates a new hash-based sharding strategy
func NewHashSharding(shardCount int) *HashSharding {
	sharding := &HashSharding{
		shardCount: shardCount,
		shards:     make(map[int]*ShardInfo),
	}

	// Initialize shards
	for i := 0; i < shardCount; i++ {
		sharding.shards[i] = &ShardInfo{
			ID:          i,
			Status:      "active",
			NodeIDs:     make([]string, 0),
			VectorCount: 0,
			SizeBytes:   0,
			LastUpdated: time.Now().Unix(),
		}
	}

	return sharding
}

// GetShard returns the shard for a given key
func (h *HashSharding) GetShard(key string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Use MD5 hash for consistent shard assignment
	hash := md5.Sum([]byte(key))
	hashValue := binary.BigEndian.Uint32(hash[:4])

	return int(hashValue % uint32(h.shardCount))
}

// GetShardCount returns the total number of shards
func (h *HashSharding) GetShardCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.shardCount
}

// AddShard adds a new shard
func (h *HashSharding) AddShard() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	newShardID := h.shardCount
	h.shards[newShardID] = &ShardInfo{
		ID:          newShardID,
		Status:      "active",
		NodeIDs:     make([]string, 0),
		VectorCount: 0,
		SizeBytes:   0,
		LastUpdated: time.Now().Unix(),
	}

	h.shardCount++

	return nil
}

// RemoveShard removes a shard
func (h *HashSharding) RemoveShard(shardID int) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if shardID >= h.shardCount {
		return fmt.Errorf("invalid shard ID: %d", shardID)
	}

	if shardID == h.shardCount-1 {
		// Remove the last shard
		delete(h.shards, shardID)
		h.shardCount--
	} else {
		// Mark shard as inactive instead of removing
		if shard, exists := h.shards[shardID]; exists {
			shard.Status = "inactive"
			shard.LastUpdated = time.Now().Unix()
		}
	}

	return nil
}

// Rebalance rebalances data across shards
func (h *HashSharding) Rebalance() error {
	// TODO: Implement data rebalancing logic
	return nil
}

// GetShardInfo returns information about a specific shard
func (h *HashSharding) GetShardInfo(shardID int) (*ShardInfo, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if shardID >= h.shardCount {
		return nil, fmt.Errorf("invalid shard ID: %d", shardID)
	}

	shard, exists := h.shards[shardID]
	if !exists {
		return nil, fmt.Errorf("shard not found: %d", shardID)
	}

	return shard, nil
}

// UpdateShardInfo updates information for a specific shard
func (h *HashSharding) UpdateShardInfo(shardID int, info *ShardInfo) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if shardID >= h.shardCount {
		return fmt.Errorf("invalid shard ID: %d", shardID)
	}

	shard, exists := h.shards[shardID]
	if !exists {
		return fmt.Errorf("shard not found: %d", shardID)
	}

	// Update shard information
	shard.Status = info.Status
	shard.NodeIDs = info.NodeIDs
	shard.VectorCount = info.VectorCount
	shard.SizeBytes = info.SizeBytes
	shard.LastUpdated = info.LastUpdated

	return nil
}
