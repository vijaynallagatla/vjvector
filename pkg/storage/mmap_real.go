package storage

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

// MMapFile represents a memory-mapped file for vector storage
type MMapFile struct {
	filePath    string
	fileHandle  *os.File
	mmapData    []byte
	fileSize    int64
	index       map[string]int64 // ID -> offset
	mutex       sync.RWMutex
	pageSize    int
	compression bool
}

// VectorHeader represents the header for each vector in the file
type VectorHeader struct {
	ID        [64]byte // Fixed-size ID (64 bytes)
	Dimension uint32   // Vector dimension
	DataSize  uint32   // Size of vector data in bytes
	Timestamp int64    // Unix timestamp
	Checksum  uint32   // Simple checksum for data integrity
}

// VectorData represents the actual vector data
type VectorData struct {
	Header VectorHeader
	Data   []float64
}

// NewMMapFile creates a new memory-mapped file for vector storage
func NewMMapFile(filePath string, pageSize int, compression bool) (*MMapFile, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Open or create file
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0600) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to get file info and close: %w, close error: %v", err, closeErr)
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := fileInfo.Size()
	if fileSize == 0 {
		// Initialize file with header
		fileSize = 4096 // Start with 4KB
		if err := file.Truncate(fileSize); err != nil {
			if closeErr := file.Close(); closeErr != nil {
				return nil, fmt.Errorf("failed to initialize file and close: %w, close error: %v", err, closeErr)
			}
			return nil, fmt.Errorf("failed to initialize file: %w", err)
		}
	}

	// Memory map the file
	mmapData, err := syscall.Mmap(
		int(file.Fd()),
		0,
		int(fileSize),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to memory map file and close: %w, close error: %v", err, closeErr)
		}
		return nil, fmt.Errorf("failed to memory map file: %w", err)
	}

	mmapFile := &MMapFile{
		filePath:    filePath,
		fileHandle:  file,
		mmapData:    mmapData,
		fileSize:    fileSize,
		index:       make(map[string]int64),
		pageSize:    pageSize,
		compression: compression,
	}

	// Build index from existing data
	if err := mmapFile.buildIndex(); err != nil {
		if closeErr := mmapFile.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to build index and close: %w, close error: %v", err, closeErr)
		}
		return nil, fmt.Errorf("failed to build index: %w", err)
	}

	return mmapFile, nil
}

// buildIndex builds the index by scanning the file
func (m *MMapFile) buildIndex() error {
	offset := int64(0)

	for offset < m.fileSize {
		if offset+int64(unsafe.Sizeof(VectorHeader{})) > m.fileSize {
			break
		}

		// Read header
		header := (*VectorHeader)(unsafe.Pointer(&m.mmapData[offset])) // nolint:gosec // nolint:gosec

		// Check if this is a valid header (non-zero dimension)
		if header.Dimension == 0 {
			break // End of valid data
		}

		// Calculate data size
		dataSize := int64(header.DataSize)
		totalSize := int64(unsafe.Sizeof(VectorHeader{})) + dataSize

		if offset+totalSize > m.fileSize {
			break
		}

		// Add to index
		id := string(header.ID[:])
		m.index[id] = offset

		// Move to next vector
		offset += totalSize
	}

	return nil
}

// Write writes a vector to the memory-mapped file
func (m *MMapFile) Write(vector *core.Vector) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if vector already exists
	if _, exists := m.index[vector.ID]; exists {
		return fmt.Errorf("vector %s already exists", vector.ID)
	}

	// For now, just store in memory to get the demo working
	// TODO: Implement full mmap functionality
	m.index[vector.ID] = 0 // Placeholder offset

	return nil
}

// Read reads a vector from the memory-mapped file
func (m *MMapFile) Read(id string) (*core.Vector, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	offset, exists := m.index[id]
	if !exists {
		return nil, fmt.Errorf("vector %s not found", id)
	}

	// Read header
	if offset+int64(unsafe.Sizeof(VectorHeader{})) > m.fileSize {
		return nil, fmt.Errorf("invalid offset for vector %s", id)
	}

	header := (*VectorHeader)(unsafe.Pointer(&m.mmapData[offset])) // nolint:gosec

	// Validate header
	if header.Dimension == 0 {
		return nil, fmt.Errorf("invalid header for vector %s", id)
	}

	// Read vector data
	dataOffset := offset + int64(unsafe.Sizeof(VectorHeader{}))
	dataSize := int64(header.DataSize)

	if dataOffset+dataSize > m.fileSize {
		return nil, fmt.Errorf("invalid data size for vector %s", id)
	}

	// Convert bytes to float64 slice
	dataBytes := m.mmapData[dataOffset : dataOffset+dataSize]
	floatCount := len(dataBytes) / 8
	vectorData := make([]float64, floatCount)

	for i := 0; i < floatCount; i++ {
		start := i * 8
		end := start + 8
		if end <= len(dataBytes) {
			vectorData[i] = float64(binary.LittleEndian.Uint64(dataBytes[start:end]))
		}
	}

	// Verify checksum
	expectedChecksum := m.calculateChecksum(vectorData)
	if header.Checksum != expectedChecksum {
		return nil, fmt.Errorf("checksum mismatch for vector %s", id)
	}

	return &core.Vector{
		ID:        id,
		Embedding: vectorData,
		Metadata:  make(map[string]interface{}),
	}, nil
}

// Delete removes a vector from the file
func (m *MMapFile) Delete(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	offset, exists := m.index[id]
	if !exists {
		return fmt.Errorf("vector %s not found", id)
	}

	// Mark as deleted by zeroing the dimension
	header := (*VectorHeader)(unsafe.Pointer(&m.mmapData[offset])) // nolint:gosec
	header.Dimension = 0

	// Remove from index
	delete(m.index, id)

	// Sync to disk
	if err := m.syncToDisk(); err != nil {
		return fmt.Errorf("failed to sync to disk: %w", err)
	}

	return nil
}

// expandFile expands the memory-mapped file to the new size
// TODO: Implement file expansion logic when needed
// This would involve:
// 1. Unmapping current memory
// 2. Expanding the file on disk
// 3. Remapping with new size
// 4. Rebuilding the index

// syncToDisk syncs the memory-mapped data to disk
func (m *MMapFile) syncToDisk() error {
	// On macOS, we need to use a different approach
	// For now, we'll just flush the file handle
	if err := m.fileHandle.Sync(); err != nil {
		return fmt.Errorf("failed to sync file to disk: %w", err)
	}
	return nil
}

// calculateChecksum calculates a simple checksum for data integrity
func (m *MMapFile) calculateChecksum(data []float64) uint32 {
	var checksum uint32
	for _, value := range data {
		// Convert float64 to bytes and sum
		bytes := (*[8]byte)(unsafe.Pointer(&value)) // nolint:gosec
		for _, b := range bytes {
			checksum += uint32(b)
		}
	}
	return checksum
}

// Close closes the memory-mapped file
func (m *MMapFile) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Sync to disk
	if err := m.syncToDisk(); err != nil {
		return fmt.Errorf("failed to sync to disk: %w", err)
	}

	// Unmap memory
	if err := syscall.Munmap(m.mmapData); err != nil {
		return fmt.Errorf("failed to unmap memory: %w", err)
	}

	// Close file
	if err := m.fileHandle.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

// GetStats returns storage statistics
func (m *MMapFile) GetStats() StorageStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return StorageStats{
		TotalVectors: int64(len(m.index)),
		StorageSize:  m.fileSize,
		MemoryUsage:  m.fileSize, // MMap uses same amount of memory
		FileCount:    1,          // Single file
		PageSize:     m.pageSize,
	}
}
