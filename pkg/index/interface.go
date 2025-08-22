// Package index provides vector indexing implementations for efficient similarity search.
// It includes HNSW and IVF algorithms for approximate nearest neighbor search.
package index

import (
	"context"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

// VectorIndex defines the interface for vector indexing operations.
// Implementations include HNSW and IVF algorithms.
type VectorIndex interface {
	// Insert adds a vector to the index
	Insert(vector *core.Vector) error

	// Search finds the k most similar vectors to the query vector
	Search(query []float64, k int) ([]core.VectorSearchResult, error)

	// SearchWithContext finds the k most similar vectors with context support
	SearchWithContext(ctx context.Context, query []float64, k int) ([]core.VectorSearchResult, error)

	// Delete removes a vector from the index by ID
	Delete(id string) error

	// Optimize performs index optimization and maintenance
	Optimize() error

	// GetStats returns index performance and structure statistics
	GetStats() IndexStats

	// Close performs cleanup and resource management
	Close() error
}

// IndexStats provides performance and structure information about an index
type IndexStats struct {
	// Basic statistics
	TotalVectors int64 `json:"total_vectors"`
	IndexSize    int64 `json:"index_size_bytes"`
	MemoryUsage  int64 `json:"memory_usage_bytes"`

	// Performance metrics
	AvgSearchTime float64 `json:"avg_search_time_ms"`
	AvgInsertTime float64 `json:"avg_insert_time_ms"`

	// Quality metrics
	Recall    float64 `json:"recall_at_k"`
	Precision float64 `json:"precision_at_k"`

	// Index-specific metrics
	NumLayers      int `json:"num_layers,omitempty"`      // HNSW specific
	NumClusters    int `json:"num_clusters,omitempty"`    // IVF specific
	MaxConnections int `json:"max_connections,omitempty"` // HNSW specific
}

// IndexType represents the type of vector index
type IndexType string

const (
	IndexTypeHNSW IndexType = "hnsw"
	IndexTypeIVF  IndexType = "ivf"
)

// IndexConfig holds configuration parameters for index creation
type IndexConfig struct {
	Type        IndexType `json:"type"`
	Dimension   int       `json:"dimension"`
	MaxElements int       `json:"max_elements"`

	// HNSW specific parameters
	M              int `json:"m,omitempty"`               // Max connections per layer
	EfConstruction int `json:"ef_construction,omitempty"` // Search depth during construction
	EfSearch       int `json:"ef_search,omitempty"`       // Search depth during queries
	MaxLayers      int `json:"max_layers,omitempty"`      // Maximum number of layers

	// IVF specific parameters
	NumClusters int `json:"num_clusters,omitempty"` // Number of clusters
	ClusterSize int `json:"cluster_size,omitempty"` // Target cluster size

	// General parameters
	DistanceMetric string `json:"distance_metric"` // "cosine", "euclidean", "dot"
	Normalize      bool   `json:"normalize"`       // Whether to normalize vectors
}

// IndexFactory creates new index instances based on configuration
type IndexFactory interface {
	// CreateIndex creates a new index with the given configuration
	CreateIndex(config IndexConfig) (VectorIndex, error)

	// ValidateConfig validates the configuration parameters
	ValidateConfig(config IndexConfig) error
}

// DefaultIndexFactory provides the standard implementation of IndexFactory
type DefaultIndexFactory struct{}

// NewIndexFactory creates a new default index factory
func NewIndexFactory() IndexFactory {
	return &DefaultIndexFactory{}
}

// CreateIndex creates a new index based on the configuration
func (f *DefaultIndexFactory) CreateIndex(config IndexConfig) (VectorIndex, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, err
	}

	switch config.Type {
	case IndexTypeHNSW:
		return NewHNSWIndex(config)
	case IndexTypeIVF:
		return NewIVFIndex(config)
	default:
		return nil, ErrUnsupportedIndexType
	}
}

// ValidateConfig validates the configuration parameters
func (f *DefaultIndexFactory) ValidateConfig(config IndexConfig) error {
	if config.Dimension <= 0 {
		return ErrInvalidDimension
	}

	if config.MaxElements <= 0 {
		return ErrInvalidMaxElements
	}

	switch config.Type {
	case IndexTypeHNSW:
		return f.validateHNSWConfig(config)
	case IndexTypeIVF:
		return f.validateIVFConfig(config)
	default:
		return ErrUnsupportedIndexType
	}
}

// validateHNSWConfig validates HNSW-specific configuration
func (f *DefaultIndexFactory) validateHNSWConfig(config IndexConfig) error {
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

// validateIVFConfig validates IVF-specific configuration
func (f *DefaultIndexFactory) validateIVFConfig(config IndexConfig) error {
	if config.NumClusters <= 0 {
		return ErrInvalidIVFParameter
	}
	if config.ClusterSize <= 0 {
		return ErrInvalidIVFParameter
	}
	return nil
}
