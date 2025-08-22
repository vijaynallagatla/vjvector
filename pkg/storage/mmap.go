package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

// MMapStorage provides memory-mapped file storage for vectors
type MMapStorage struct {
	config   StorageConfig
	filePath string
	mmapFile *MMapFile // Real memory-mapped file
	mutex    sync.RWMutex

	// Statistics
	stats     StorageStats
	startTime time.Time
}

// NewMMapStorage creates a new memory-mapped file storage engine
func NewMMapStorage(config StorageConfig) (StorageEngine, error) {
	// Create real memory-mapped file
	mmapFile, err := NewMMapFile(config.DataPath, config.PageSize, config.Compression)
	if err != nil {
		return nil, fmt.Errorf("failed to create mmap file: %w", err)
	}

	storage := &MMapStorage{
		config:    config,
		filePath:  config.DataPath,
		mmapFile:  mmapFile,
		startTime: time.Now(),
	}

	return storage, nil
}

// Write stores multiple vectors to memory-mapped file storage
func (m *MMapStorage) Write(vectors []*core.Vector) error {
	return m.WriteWithContext(context.Background(), vectors)
}

// WriteWithContext stores multiple vectors with context support
func (m *MMapStorage) WriteWithContext(_ context.Context, vectors []*core.Vector) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	start := time.Now()

	// Write vectors to memory-mapped file
	for _, vector := range vectors {
		if err := m.mmapFile.Write(vector); err != nil {
			return fmt.Errorf("failed to write vector %s: %w", vector.ID, err)
		}
	}

	// Update statistics
	stats := m.mmapFile.GetStats()
	m.stats.TotalVectors = stats.TotalVectors
	m.stats.AvgWriteTime = float64(time.Since(start).Microseconds()) / float64(len(vectors))

	return nil
}

// Read retrieves vectors by their IDs
func (m *MMapStorage) Read(ids []string) ([]*core.Vector, error) {
	return m.ReadWithContext(context.Background(), ids)
}

// ReadWithContext retrieves vectors with context support
func (m *MMapStorage) ReadWithContext(_ context.Context, ids []string) ([]*core.Vector, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	start := time.Now()

	// Read vectors from memory-mapped file
	vectors := make([]*core.Vector, 0, len(ids))
	for _, id := range ids {
		vector, err := m.mmapFile.Read(id)
		if err != nil {
			// Skip vectors that can't be read
			continue
		}
		vectors = append(vectors, vector)
	}

	// Update statistics
	m.stats.AvgReadTime = float64(time.Since(start).Microseconds()) / float64(len(ids))

	return vectors, nil
}

// Delete removes vectors by their IDs
func (m *MMapStorage) Delete(ids []string) error {
	return m.DeleteWithContext(context.Background(), ids)
}

// DeleteWithContext removes vectors with context support
func (m *MMapStorage) DeleteWithContext(_ context.Context, ids []string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	start := time.Now()

	// Delete vectors from memory-mapped file
	for _, id := range ids {
		if err := m.mmapFile.Delete(id); err != nil {
			// Continue with other deletions even if one fails
			continue
		}
	}

	// Update statistics
	stats := m.mmapFile.GetStats()
	m.stats.TotalVectors = stats.TotalVectors
	m.stats.AvgDeleteTime = float64(time.Since(start).Microseconds()) / float64(len(ids))

	return nil
}

// Compact performs storage optimization and cleanup
func (m *MMapStorage) Compact() error {
	// TODO: Implement mmap compaction
	// This is a placeholder - actual implementation will be added in Week 7-8
	return nil
}

// GetStats returns storage performance and usage statistics
func (m *MMapStorage) GetStats() StorageStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := m.stats
	stats.StorageSize = m.mmapFile.GetStats().StorageSize
	stats.MemoryUsage = m.mmapFile.GetStats().MemoryUsage
	stats.PageSize = m.config.PageSize
	stats.FileCount = 1 // Single mmap file

	return stats
}

// Close performs cleanup and resource management
func (m *MMapStorage) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Close the memory-mapped file
	if m.mmapFile != nil {
		if err := m.mmapFile.Close(); err != nil {
			return fmt.Errorf("failed to close mmap file: %w", err)
		}
	}

	return nil
}
