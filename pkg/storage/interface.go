// Package storage provides storage engines for vector data persistence.
// It includes memory-mapped file storage and metadata management.
package storage

import (
	"context"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

// Engine is an alias for StorageEngine to satisfy linter preferences
type Engine = StorageEngine

// Stats is an alias for StorageStats to satisfy linter preferences
type Stats = StorageStats

// Type is an alias for StorageType to satisfy linter preferences
type Type = StorageType

// Config is an alias for StorageConfig to satisfy linter preferences
type Config = StorageConfig

// Factory is an alias for StorageFactory to satisfy linter preferences
type Factory = StorageFactory

// StorageEngine defines the interface for vector storage operations
type StorageEngine interface {
	// Write stores multiple vectors to storage
	Write(vectors []*core.Vector) error

	// WriteWithContext stores multiple vectors with context support
	WriteWithContext(ctx context.Context, vectors []*core.Vector) error

	// Read retrieves vectors by their IDs
	Read(ids []string) ([]*core.Vector, error)

	// ReadWithContext retrieves vectors with context support
	ReadWithContext(ctx context.Context, ids []string) ([]*core.Vector, error)

	// Delete removes vectors by their IDs
	Delete(ids []string) error

	// DeleteWithContext removes vectors with context support
	DeleteWithContext(ctx context.Context, ids []string) error

	// Compact performs storage optimization and cleanup
	Compact() error

	// GetStats returns storage performance and usage statistics
	GetStats() StorageStats

	// Close performs cleanup and resource management
	Close() error
}

// StorageStats provides performance and usage information about storage
type StorageStats struct {
	// Basic statistics
	TotalVectors int64 `json:"total_vectors"`
	StorageSize  int64 `json:"storage_size_bytes"`
	MemoryUsage  int64 `json:"memory_usage_bytes"`

	// Performance metrics
	AvgWriteTime  float64 `json:"avg_write_time_ms"`
	AvgReadTime   float64 `json:"avg_read_time_ms"`
	AvgDeleteTime float64 `json:"avg_delete_time_ms"`

	// Storage efficiency
	CompressionRatio float64 `json:"compression_ratio"`
	Fragmentation    float64 `json:"fragmentation_percent"`

	// File system metrics
	FileCount int `json:"file_count"`
	PageSize  int `json:"page_size_bytes"`
}

// StorageType represents the type of storage engine
type StorageType string

// StorageType constants define the available storage engine types
const (
	StorageTypeMemory  StorageType = "memory"  // In-memory storage
	StorageTypeMMap    StorageType = "mmap"    // Memory-mapped file storage
	StorageTypeLevelDB StorageType = "leveldb" // LevelDB-based storage
)

// StorageConfig holds configuration parameters for storage creation
type StorageConfig struct {
	Type        StorageType `json:"type"`
	DataPath    string      `json:"data_path"`
	MaxFileSize int64       `json:"max_file_size"`

	// Memory-mapped file parameters
	PageSize    int  `json:"page_size,omitempty"`
	Compression bool `json:"compression,omitempty"`
	SyncOnWrite bool `json:"sync_on_write,omitempty"`

	// LevelDB parameters
	CacheSize       int64 `json:"cache_size,omitempty"`
	WriteBufferSize int   `json:"write_buffer_size,omitempty"`
	MaxOpenFiles    int   `json:"max_open_files,omitempty"`

	// General parameters
	BatchSize     int `json:"batch_size"`
	FlushInterval int `json:"flush_interval_ms"`
}

// StorageFactory creates new storage engine instances based on configuration
type StorageFactory interface {
	// CreateStorage creates a new storage engine with the given configuration
	CreateStorage(config StorageConfig) (StorageEngine, error)

	// ValidateConfig validates the configuration parameters
	ValidateConfig(config StorageConfig) error
}

// DefaultStorageFactory provides the standard implementation of StorageFactory
type DefaultStorageFactory struct{}

// NewStorageFactory creates a new default storage factory
func NewStorageFactory() StorageFactory {
	return &DefaultStorageFactory{}
}

// CreateStorage creates a new storage engine based on the configuration
func (f *DefaultStorageFactory) CreateStorage(config StorageConfig) (StorageEngine, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, err
	}

	switch config.Type {
	case StorageTypeMemory:
		return NewMemoryStorage(config)
	case StorageTypeMMap:
		return NewMMapStorage(config)
	case StorageTypeLevelDB:
		return NewLevelDBStorage(config)
	default:
		return nil, ErrUnsupportedStorageType
	}
}

// ValidateConfig validates the configuration parameters
func (f *DefaultStorageFactory) ValidateConfig(config StorageConfig) error {
	if config.DataPath == "" {
		return ErrInvalidDataPath
	}

	if config.MaxFileSize <= 0 {
		return ErrInvalidMaxFileSize
	}

	if config.BatchSize <= 0 {
		return ErrInvalidBatchSize
	}

	switch config.Type {
	case StorageTypeMemory:
		return f.validateMemoryConfig(config)
	case StorageTypeMMap:
		return f.validateMMapConfig(config)
	case StorageTypeLevelDB:
		return f.validateLevelDBConfig(config)
	default:
		return ErrUnsupportedStorageType
	}
}

// validateMemoryConfig validates memory storage configuration
func (f *DefaultStorageFactory) validateMemoryConfig(_ StorageConfig) error {
	// Memory storage has minimal validation requirements
	return nil
}

// validateMMapConfig validates memory-mapped file storage configuration
func (f *DefaultStorageFactory) validateMMapConfig(config StorageConfig) error {
	if config.PageSize <= 0 {
		return ErrInvalidPageSize
	}
	return nil
}

// validateLevelDBConfig validates LevelDB storage configuration
func (f *DefaultStorageFactory) validateLevelDBConfig(config StorageConfig) error {
	if config.CacheSize <= 0 {
		return ErrInvalidCacheSize
	}
	if config.WriteBufferSize <= 0 {
		return ErrInvalidWriteBufferSize
	}
	if config.MaxOpenFiles <= 0 {
		return ErrInvalidMaxOpenFiles
	}
	return nil
}
