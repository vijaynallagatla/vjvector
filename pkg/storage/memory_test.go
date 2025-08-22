package storage

import (
	"testing"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

func TestMemoryStorage_NewStorage(t *testing.T) {
	config := StorageConfig{
		Type:        StorageTypeMemory,
		DataPath:    "/tmp/test",
		MaxFileSize: 1024 * 1024 * 1024, // 1GB
		PageSize:    4096,
		BatchSize:   100,
	}

	factory := &DefaultStorageFactory{}
	storage, err := factory.CreateStorage(config)
	if err != nil {
		t.Fatalf("Failed to create memory storage: %v", err)
	}

	memStorage, ok := storage.(*MemoryStorage)
	if !ok {
		t.Fatalf("Expected MemoryStorage, got %T", storage)
	}

	if memStorage.config.Type != StorageTypeMemory {
		t.Errorf("Expected StorageTypeMemory, got %s", memStorage.config.Type)
	}
}

func TestMemoryStorage_Write(t *testing.T) {
	config := StorageConfig{
		Type:        StorageTypeMemory,
		DataPath:    "/tmp/test",
		MaxFileSize: 1024 * 1024 * 1024, // 1GB
		PageSize:    4096,
		BatchSize:   100,
	}

	factory := &DefaultStorageFactory{}
	storage, err := factory.CreateStorage(config)
	if err != nil {
		t.Fatalf("Failed to create memory storage: %v", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			t.Errorf("Failed to close storage: %v", err)
		}
	}()

	// Test single vector write
	vectors := []*core.Vector{
		{
			ID:         "test1",
			Collection: "test",
			Embedding:  []float64{1.0, 2.0, 3.0, 4.0},
			Metadata:   map[string]interface{}{"type": "test"},
		},
	}

	err = storage.Write(vectors)
	if err != nil {
		t.Fatalf("Failed to write vector: %v", err)
	}

	stats := storage.GetStats()
	if stats.TotalVectors != 1 {
		t.Errorf("Expected 1 vector, got %d", stats.TotalVectors)
	}
}

func TestMemoryStorage_Read(t *testing.T) {
	config := StorageConfig{
		Type:        StorageTypeMemory,
		DataPath:    "/tmp/test",
		MaxFileSize: 1024 * 1024 * 1024, // 1GB
		PageSize:    4096,
		BatchSize:   100,
	}

	factory := &DefaultStorageFactory{}
	storage, err := factory.CreateStorage(config)
	if err != nil {
		t.Fatalf("Failed to create memory storage: %v", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			t.Errorf("Failed to close storage: %v", err)
		}
	}()

	// Write test vectors
	vectors := []*core.Vector{
		{ID: "v1", Collection: "test", Embedding: []float64{1.0, 2.0, 3.0, 4.0}},
		{ID: "v2", Collection: "test", Embedding: []float64{5.0, 6.0, 7.0, 8.0}},
	}

	err = storage.Write(vectors)
	if err != nil {
		t.Fatalf("Failed to write vectors: %v", err)
	}

	// Test reading existing vectors
	readVectors, err := storage.Read([]string{"v1", "v2"})
	if err != nil {
		t.Fatalf("Failed to read vectors: %v", err)
	}

	if len(readVectors) != 2 {
		t.Errorf("Expected 2 vectors, got %d", len(readVectors))
	}

	// Test reading non-existent vector
	readVectors, err = storage.Read([]string{"nonexistent"})
	if err != nil {
		t.Fatalf("Failed to read vectors: %v", err)
	}

	if len(readVectors) != 0 {
		t.Errorf("Expected 0 vectors for non-existent ID, got %d", len(readVectors))
	}
}

func TestMemoryStorage_Delete(t *testing.T) {
	config := StorageConfig{
		Type:        StorageTypeMemory,
		DataPath:    "/tmp/test",
		MaxFileSize: 1024 * 1024 * 1024, // 1GB
		PageSize:    4096,
		BatchSize:   100,
	}

	factory := &DefaultStorageFactory{}
	storage, err := factory.CreateStorage(config)
	if err != nil {
		t.Fatalf("Failed to create memory storage: %v", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			t.Errorf("Failed to close storage: %v", err)
		}
	}()

	// Write test vectors
	vectors := []*core.Vector{
		{ID: "v1", Collection: "test", Embedding: []float64{1.0, 2.0, 3.0, 4.0}},
		{ID: "v2", Collection: "test", Embedding: []float64{5.0, 6.0, 7.0, 8.0}},
	}

	err = storage.Write(vectors)
	if err != nil {
		t.Fatalf("Failed to write vectors: %v", err)
	}

	// Verify both vectors exist
	stats := storage.GetStats()
	if stats.TotalVectors != 2 {
		t.Errorf("Expected 2 vectors, got %d", stats.TotalVectors)
	}

	// Delete one vector
	err = storage.Delete([]string{"v1"})
	if err != nil {
		t.Fatalf("Failed to delete vector: %v", err)
	}

	// Verify vector count decreased
	stats = storage.GetStats()
	if stats.TotalVectors != 1 {
		t.Errorf("Expected 1 vector after delete, got %d", stats.TotalVectors)
	}

	// Verify the correct vector was deleted
	readVectors, err := storage.Read([]string{"v1", "v2"})
	if err != nil {
		t.Fatalf("Failed to read vectors: %v", err)
	}

	if len(readVectors) != 1 {
		t.Errorf("Expected 1 vector after delete, got %d", len(readVectors))
	}

	if len(readVectors) > 0 && readVectors[0].ID != "v2" {
		t.Errorf("Expected remaining vector to be v2, got %s", readVectors[0].ID)
	}
}

func TestMemoryStorage_Compact(t *testing.T) {
	config := StorageConfig{
		Type:        StorageTypeMemory,
		DataPath:    "/tmp/test",
		MaxFileSize: 1024 * 1024 * 1024, // 1GB
		PageSize:    4096,
		BatchSize:   100,
	}

	factory := &DefaultStorageFactory{}
	storage, err := factory.CreateStorage(config)
	if err != nil {
		t.Fatalf("Failed to create memory storage: %v", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			t.Errorf("Failed to close storage: %v", err)
		}
	}()

	// Test compaction (should not fail)
	err = storage.Compact()
	if err != nil {
		t.Errorf("Compact failed: %v", err)
	}
}

func BenchmarkMemoryStorage_Write(b *testing.B) {
	config := StorageConfig{
		Type:        StorageTypeMemory,
		DataPath:    "/tmp/test",
		MaxFileSize: 1024 * 1024 * 1024, // 1GB
		PageSize:    4096,
		BatchSize:   100,
	}

	factory := &DefaultStorageFactory{}
	storage, err := factory.CreateStorage(config)
	if err != nil {
		b.Fatalf("Failed to create memory storage: %v", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			b.Errorf("Failed to close storage: %v", err)
		}
	}()

	// Pre-generate vectors
	vectors := make([]*core.Vector, b.N)
	for i := 0; i < b.N; i++ {
		embedding := make([]float64, 128)
		for j := 0; j < 128; j++ {
			embedding[j] = float64(i*j) * 0.001
		}
		vectors[i] = &core.Vector{
			ID:        string(rune(i)),
			Embedding: embedding,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := storage.Write([]*core.Vector{vectors[i]})
		if err != nil {
			b.Fatalf("Failed to write vector: %v", err)
		}
	}
}

func BenchmarkMemoryStorage_Read(b *testing.B) {
	config := StorageConfig{
		Type:        StorageTypeMemory,
		DataPath:    "/tmp/test",
		MaxFileSize: 1024 * 1024 * 1024, // 1GB
		PageSize:    4096,
		BatchSize:   100,
	}

	factory := &DefaultStorageFactory{}
	storage, err := factory.CreateStorage(config)
	if err != nil {
		b.Fatalf("Failed to create memory storage: %v", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			b.Errorf("Failed to close storage: %v", err)
		}
	}()

	// Pre-populate storage
	vectors := make([]*core.Vector, 1000)
	for i := 0; i < 1000; i++ {
		embedding := make([]float64, 128)
		for j := 0; j < 128; j++ {
			embedding[j] = float64(i*j) * 0.001
		}
		vectors[i] = &core.Vector{
			ID:        string(rune(i)),
			Embedding: embedding,
		}
	}

	err = storage.Write(vectors)
	if err != nil {
		b.Fatalf("Failed to populate storage: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := string(rune(i % 1000))
		_, err := storage.Read([]string{id})
		if err != nil {
			b.Fatalf("Failed to read vector: %v", err)
		}
	}
}
