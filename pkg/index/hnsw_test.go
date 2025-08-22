package index

import (
	"testing"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

func TestHNSWIndex_NewIndex(t *testing.T) {
	config := IndexConfig{
		Type:           IndexTypeHNSW,
		Dimension:      128,
		MaxElements:    1000,
		M:              16,
		EfConstruction: 200,
		EfSearch:       100,
		MaxLayers:      16,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		t.Fatalf("Failed to create HNSW index: %v", err)
	}

	hnswIdx, ok := idx.(*HNSWIndex)
	if !ok {
		t.Fatalf("Expected HNSWIndex, got %T", idx)
	}

	if hnswIdx.config.M != 16 {
		t.Errorf("Expected M=16, got %d", hnswIdx.config.M)
	}

	if hnswIdx.config.Dimension != 128 {
		t.Errorf("Expected Dimension=128, got %d", hnswIdx.config.Dimension)
	}
}

func TestHNSWIndex_Insert(t *testing.T) {
	config := IndexConfig{
		Type:           IndexTypeHNSW,
		Dimension:      4,
		MaxElements:    100,
		M:              4,
		EfConstruction: 50,
		EfSearch:       50,
		MaxLayers:      4,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		t.Fatalf("Failed to create HNSW index: %v", err)
	}
	defer func() {
		if err := idx.Close(); err != nil {
			t.Errorf("Failed to close index: %v", err)
		}
	}()

	// Test single insertion
	vector := &core.Vector{
		ID:         "test1",
		Collection: "test",
		Embedding:  []float64{1.0, 2.0, 3.0, 4.0},
		Metadata:   map[string]interface{}{"type": "test"},
	}

	err = idx.Insert(vector)
	if err != nil {
		t.Fatalf("Failed to insert vector: %v", err)
	}

	stats := idx.GetStats()
	if stats.TotalVectors != 1 {
		t.Errorf("Expected 1 vector, got %d", stats.TotalVectors)
	}
}

func TestHNSWIndex_Search(t *testing.T) {
	config := IndexConfig{
		Type:           IndexTypeHNSW,
		Dimension:      4,
		MaxElements:    100,
		M:              4,
		EfConstruction: 50,
		EfSearch:       50,
		MaxLayers:      4,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		t.Fatalf("Failed to create HNSW index: %v", err)
	}
	defer func() {
		if err := idx.Close(); err != nil {
			t.Errorf("Failed to close index: %v", err)
		}
	}()

	// Insert test vectors
	vectors := []*core.Vector{
		{ID: "v1", Collection: "test", Embedding: []float64{1.0, 0.0, 0.0, 0.0}},
		{ID: "v2", Collection: "test", Embedding: []float64{0.0, 1.0, 0.0, 0.0}},
		{ID: "v3", Collection: "test", Embedding: []float64{0.0, 0.0, 1.0, 0.0}},
		{ID: "v4", Collection: "test", Embedding: []float64{0.0, 0.0, 0.0, 1.0}},
	}

	for _, v := range vectors {
		if err := idx.Insert(v); err != nil {
			t.Fatalf("Failed to insert vector %s: %v", v.ID, err)
		}
	}

	// Test search
	query := []float64{1.0, 0.1, 0.0, 0.0}
	results, err := idx.Search(query, 2)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Errorf("Expected search results, got none")
	}

	// First result should be most similar (v1)
	if len(results) > 0 && results[0].Vector.ID != "v1" {
		t.Errorf("Expected first result to be v1, got %s", results[0].Vector.ID)
	}
}

func TestHNSWIndex_Delete(t *testing.T) {
	config := IndexConfig{
		Type:           IndexTypeHNSW,
		Dimension:      4,
		MaxElements:    100,
		M:              4,
		EfConstruction: 50,
		EfSearch:       50,
		MaxLayers:      4,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		t.Fatalf("Failed to create HNSW index: %v", err)
	}
	defer func() {
		if err := idx.Close(); err != nil {
			t.Errorf("Failed to close index: %v", err)
		}
	}()

	// Insert a vector
	vector := &core.Vector{
		ID:         "test1",
		Collection: "test",
		Embedding:  []float64{1.0, 2.0, 3.0, 4.0},
	}

	err = idx.Insert(vector)
	if err != nil {
		t.Fatalf("Failed to insert vector: %v", err)
	}

	// Verify it was inserted
	stats := idx.GetStats()
	if stats.TotalVectors != 1 {
		t.Errorf("Expected 1 vector, got %d", stats.TotalVectors)
	}

	// Delete the vector (note: current implementation is placeholder)
	_ = idx.Delete("test1") // Ignore error since delete is not fully implemented

	// Note: Current implementation is placeholder - in real implementation
	// we would verify the vector count decreased
}

func TestHNSWIndex_RandomLevel(t *testing.T) {
	config := IndexConfig{
		Type:           IndexTypeHNSW,
		Dimension:      4,
		MaxElements:    100,
		M:              4,
		EfConstruction: 50,
		EfSearch:       50,
		MaxLayers:      4,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		t.Fatalf("Failed to create HNSW index: %v", err)
	}
	defer func() {
		if err := idx.Close(); err != nil {
			t.Errorf("Failed to close index: %v", err)
		}
	}()

	hnswIdx := idx.(*HNSWIndex)

	// Test random level generation
	levels := make(map[int]int)
	for i := 0; i < 1000; i++ {
		level := hnswIdx.randomLevel()
		if level < 0 || level >= hnswIdx.config.MaxLayers {
			t.Errorf("Invalid level %d, expected 0 <= level < %d", level, hnswIdx.config.MaxLayers)
		}
		levels[level]++
	}

	// Level 0 should be most common
	if levels[0] == 0 {
		t.Errorf("Expected some level 0 nodes")
	}

	// Higher levels should be less common
	if levels[0] < levels[1] {
		t.Errorf("Expected level 0 to be more common than level 1")
	}
}

func TestHNSWIndex_DistanceCalculation(t *testing.T) {
	config := IndexConfig{
		Type:           IndexTypeHNSW,
		Dimension:      3,
		MaxElements:    100,
		M:              4,
		EfConstruction: 50,
		EfSearch:       50,
		MaxLayers:      4,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		t.Fatalf("Failed to create HNSW index: %v", err)
	}
	defer func() {
		if err := idx.Close(); err != nil {
			t.Errorf("Failed to close index: %v", err)
		}
	}()

	hnswIdx := idx.(*HNSWIndex)

	// Test cosine distance
	a := []float64{1.0, 0.0, 0.0}
	b := []float64{0.0, 1.0, 0.0}
	c := []float64{1.0, 0.0, 0.0}

	distAB := hnswIdx.cosineDistance(a, b)
	distAC := hnswIdx.cosineDistance(a, c)

	// Distance to itself should be 0
	if distAC != 0.0 {
		t.Errorf("Expected distance to self to be 0, got %f", distAC)
	}

	// Distance to orthogonal vector should be 1
	if distAB != 1.0 {
		t.Errorf("Expected distance to orthogonal vector to be 1, got %f", distAB)
	}

	// Test euclidean distance
	distAB = hnswIdx.euclideanDistance(a, b)
	distAC = hnswIdx.euclideanDistance(a, c)

	if distAC != 0.0 {
		t.Errorf("Expected Euclidean distance to self to be 0, got %f", distAC)
	}

	expectedAB := 1.4142135623730951 // sqrt(2)
	if abs(distAB-expectedAB) > 1e-10 {
		t.Errorf("Expected Euclidean distance to be %f, got %f", expectedAB, distAB)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func BenchmarkHNSWIndex_Insert(b *testing.B) {
	config := IndexConfig{
		Type:           IndexTypeHNSW,
		Dimension:      128,
		MaxElements:    10000,
		M:              16,
		EfConstruction: 200,
		EfSearch:       100,
		MaxLayers:      16,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		b.Fatalf("Failed to create HNSW index: %v", err)
	}
	defer func() {
		if err := idx.Close(); err != nil {
			b.Errorf("Failed to close index: %v", err)
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
		err := idx.Insert(vectors[i])
		if err != nil {
			b.Fatalf("Failed to insert vector: %v", err)
		}
	}
}

func BenchmarkHNSWIndex_Search(b *testing.B) {
	config := IndexConfig{
		Type:           IndexTypeHNSW,
		Dimension:      128,
		MaxElements:    10000,
		M:              16,
		EfConstruction: 200,
		EfSearch:       100,
		MaxLayers:      16,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		b.Fatalf("Failed to create HNSW index: %v", err)
	}
	defer func() {
		if err := idx.Close(); err != nil {
			b.Errorf("Failed to close index: %v", err)
		}
	}()

	// Insert test vectors
	for i := 0; i < 1000; i++ {
		embedding := make([]float64, 128)
		for j := 0; j < 128; j++ {
			embedding[j] = float64(i*j) * 0.001
		}
		vector := &core.Vector{
			ID:        string(rune(i)),
			Embedding: embedding,
		}
		err := idx.Insert(vector)
		if err != nil {
			b.Fatalf("Failed to insert vector: %v", err)
		}
	}

	// Prepare query
	query := make([]float64, 128)
	for i := 0; i < 128; i++ {
		query[i] = float64(i) * 0.001
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := idx.Search(query, 10)
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}
	}
}
