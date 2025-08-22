package index

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

// HNSWIndex implements the Hierarchical Navigable Small World algorithm
// for approximate nearest neighbor search
type HNSWIndex struct {
	config     IndexConfig
	vectors    []*core.Vector
	layers     [][]*Node
	entryPoint *Node
	mutex      sync.RWMutex

	// Statistics
	stats     IndexStats
	startTime time.Time
}

// Node represents a node in the HNSW graph
type Node struct {
	ID      string    `json:"id"`
	Vector  []float64 `json:"vector"`
	Level   int       `json:"level"`
	Friends [][]int   `json:"friends"` // Friends at each level
}

// NewHNSWIndex creates a new HNSW index with the given configuration
func NewHNSWIndex(config IndexConfig) (VectorIndex, error) {
	if err := validateHNSWConfig(config); err != nil {
		return nil, err
	}

	index := &HNSWIndex{
		config:    config,
		vectors:   make([]*core.Vector, 0, config.MaxElements),
		layers:    make([][]*Node, config.MaxLayers),
		startTime: time.Now(),
	}

	// Initialize layers
	for i := range index.layers {
		index.layers[i] = make([]*Node, 0)
	}

	return index, nil
}

// Insert adds a vector to the HNSW index
func (h *HNSWIndex) Insert(vector *core.Vector) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if len(h.vectors) >= h.config.MaxElements {
		return ErrIndexFull
	}

	// Validate vector dimension
	if len(vector.Embedding) != h.config.Dimension {
		return ErrInvalidDimension
	}

	// Add vector to storage
	h.vectors = append(h.vectors, vector)

	// Create node
	node := &Node{
		ID:      vector.ID,
		Vector:  vector.Embedding,
		Level:   h.randomLevel(),
		Friends: make([][]int, h.config.MaxLayers),
	}

	// Add node to appropriate layers
	for level := 0; level <= node.Level; level++ {
		h.layers[level] = append(h.layers[level], node)
	}

	// Update entry point if this is the first node
	if h.entryPoint == nil {
		h.entryPoint = node
	} else {
		// TODO: Implement HNSW insertion algorithm
		// This is a placeholder - actual implementation will be added in Week 3-4
	}

	// Update statistics
	h.stats.TotalVectors++

	return nil
}

// Search finds the k most similar vectors to the query vector
func (h *HNSWIndex) Search(query []float64, k int) ([]core.VectorSearchResult, error) {
	return h.SearchWithContext(context.Background(), query, k)
}

// SearchWithContext finds the k most similar vectors with context support
func (h *HNSWIndex) SearchWithContext(ctx context.Context, query []float64, k int) ([]core.VectorSearchResult, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.entryPoint == nil {
		return nil, ErrIndexNotInitialized
	}

	if len(query) != h.config.Dimension {
		return nil, ErrInvalidDimension
	}

	if k <= 0 {
		return nil, ErrInvalidQuery
	}

	// TODO: Implement HNSW search algorithm
	// This is a placeholder - actual implementation will be added in Week 3-4

	// For now, return empty results
	results := make([]core.VectorSearchResult, 0)

	return results, nil
}

// Delete removes a vector from the index by ID
func (h *HNSWIndex) Delete(id string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// TODO: Implement HNSW deletion algorithm
	// This is a placeholder - actual implementation will be added in Week 3-4

	return ErrVectorNotFound
}

// Optimize performs index optimization and maintenance
func (h *HNSWIndex) Optimize() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// TODO: Implement HNSW optimization
	// This is a placeholder - actual implementation will be added in Week 3-4

	return nil
}

// GetStats returns index performance and structure statistics
func (h *HNSWIndex) GetStats() IndexStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	stats := h.stats
	stats.NumLayers = len(h.layers)
	stats.MaxConnections = h.config.M

	// Calculate memory usage (rough estimate)
	stats.MemoryUsage = int64(len(h.vectors) * h.config.Dimension * 8) // 8 bytes per float64

	return stats
}

// Close performs cleanup and resource management
func (h *HNSWIndex) Close() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Clear data structures
	h.vectors = nil
	h.layers = nil
	h.entryPoint = nil

	return nil
}

// randomLevel generates a random level for a new node
// Uses the geometric distribution as described in the HNSW paper
func (h *HNSWIndex) randomLevel() int {
	level := 0
	for level < h.config.MaxLayers-1 && rand.Float64() < 0.5 {
		level++
	}
	return level
}

// validateHNSWConfig validates HNSW-specific configuration
func validateHNSWConfig(config IndexConfig) error {
	if config.M <= 0 {
		return ErrInvalidHNSWParameter
	}
	if config.EfConstruction <= 0 {
		return ErrInvalidHNSWParameter
	}
	if config.EfSearch <= 0 {
		return ErrInvalidHNSWParameter
	}
	if config.MaxLayers <= 0 {
		return ErrInvalidHNSWParameter
	}
	return nil
}
