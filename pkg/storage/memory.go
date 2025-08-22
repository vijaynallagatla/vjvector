package storage

import (
	"context"
	"sync"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

// MemoryStorage provides in-memory vector storage
type MemoryStorage struct {
	config  StorageConfig
	vectors map[string]*core.Vector
	mutex   sync.RWMutex

	// Statistics
	stats     StorageStats
	startTime time.Time
}

// NewMemoryStorage creates a new memory storage engine
func NewMemoryStorage(config StorageConfig) (StorageEngine, error) {
	storage := &MemoryStorage{
		config:    config,
		vectors:   make(map[string]*core.Vector),
		startTime: time.Now(),
	}

	return storage, nil
}

// Write stores multiple vectors to memory storage
func (m *MemoryStorage) Write(vectors []*core.Vector) error {
	return m.WriteWithContext(context.Background(), vectors)
}

// WriteWithContext stores multiple vectors with context support
func (m *MemoryStorage) WriteWithContext(ctx context.Context, vectors []*core.Vector) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	start := time.Now()

	for _, vector := range vectors {
		m.vectors[vector.ID] = vector
	}

	// Update statistics
	m.stats.TotalVectors = int64(len(m.vectors))
	m.stats.AvgWriteTime = float64(time.Since(start).Microseconds()) / float64(len(vectors))

	return nil
}

// Read retrieves vectors by their IDs
func (m *MemoryStorage) Read(ids []string) ([]*core.Vector, error) {
	return m.ReadWithContext(context.Background(), ids)
}

// ReadWithContext retrieves vectors with context support
func (m *MemoryStorage) ReadWithContext(ctx context.Context, ids []string) ([]*core.Vector, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	start := time.Now()

	vectors := make([]*core.Vector, 0, len(ids))
	for _, id := range ids {
		if vector, exists := m.vectors[id]; exists {
			vectors = append(vectors, vector)
		}
	}

	// Update statistics
	m.stats.AvgReadTime = float64(time.Since(start).Microseconds()) / float64(len(ids))

	return vectors, nil
}

// Delete removes vectors by their IDs
func (m *MemoryStorage) Delete(ids []string) error {
	return m.DeleteWithContext(context.Background(), ids)
}

// DeleteWithContext removes vectors with context support
func (m *MemoryStorage) DeleteWithContext(ctx context.Context, ids []string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	start := time.Now()

	for _, id := range ids {
		delete(m.vectors, id)
	}

	// Update statistics
	m.stats.TotalVectors = int64(len(m.vectors))
	m.stats.AvgDeleteTime = float64(time.Since(start).Microseconds()) / float64(len(ids))

	return nil
}

// Compact performs storage optimization and cleanup
func (m *MemoryStorage) Compact() error {
	// Memory storage doesn't need compaction
	return nil
}

// GetStats returns storage performance and usage statistics
func (m *MemoryStorage) GetStats() StorageStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := m.stats
	stats.StorageSize = int64(len(m.vectors) * 1024) // Rough estimate
	stats.MemoryUsage = int64(len(m.vectors) * 1024) // Rough estimate
	stats.FileCount = 0                              // Memory storage has no files

	return stats
}

// Close performs cleanup and resource management
func (m *MemoryStorage) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Clear data structures
	m.vectors = nil

	return nil
}
