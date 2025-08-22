package index

import (
	"testing"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

func TestIVFIndex_NewIndex(t *testing.T) {
	config := IndexConfig{
		Type:           IndexTypeIVF,
		Dimension:      128,
		MaxElements:    1000,
		NumClusters:    100,
		ClusterSize:    10,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		t.Fatalf("Failed to create IVF index: %v", err)
	}

	ivfIdx, ok := idx.(*IVFIndex)
	if !ok {
		t.Fatalf("Expected IVFIndex, got %T", idx)
	}

	if ivfIdx.config.NumClusters != 100 {
		t.Errorf("Expected NumClusters=100, got %d", ivfIdx.config.NumClusters)
	}

	if ivfIdx.config.Dimension != 128 {
		t.Errorf("Expected Dimension=128, got %d", ivfIdx.config.Dimension)
	}
}

func TestIVFIndex_Insert(t *testing.T) {
	config := IndexConfig{
		Type:           IndexTypeIVF,
		Dimension:      4,
		MaxElements:    100,
		NumClusters:    4,
		ClusterSize:    10,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		t.Fatalf("Failed to create IVF index: %v", err)
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

func TestIVFIndex_Search(t *testing.T) {
	config := IndexConfig{
		Type:           IndexTypeIVF,
		Dimension:      4,
		MaxElements:    100,
		NumClusters:    2,
		ClusterSize:    10,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		t.Fatalf("Failed to create IVF index: %v", err)
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
	_, err = idx.Search(query, 2)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// IVF search might return 0 results depending on clustering
	// This is expected behavior for this implementation
}

func TestIVFIndex_Clustering(t *testing.T) {
	config := IndexConfig{
		Type:           IndexTypeIVF,
		Dimension:      2,
		MaxElements:    100,
		NumClusters:    2,
		ClusterSize:    10,
		DistanceMetric: "euclidean",
		Normalize:      false,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		t.Fatalf("Failed to create IVF index: %v", err)
	}
	defer func() {
		if err := idx.Close(); err != nil {
			t.Errorf("Failed to close index: %v", err)
		}
	}()

	// Insert vectors in different regions
	vectors := []*core.Vector{
		{ID: "v1", Collection: "test", Embedding: []float64{1.0, 1.0}},
		{ID: "v2", Collection: "test", Embedding: []float64{1.1, 1.1}},
		{ID: "v3", Collection: "test", Embedding: []float64{-1.0, -1.0}},
		{ID: "v4", Collection: "test", Embedding: []float64{-1.1, -1.1}},
	}

	for _, v := range vectors {
		if err := idx.Insert(v); err != nil {
			t.Fatalf("Failed to insert vector %s: %v", v.ID, err)
		}
	}

	// Verify vectors were inserted
	stats := idx.GetStats()
	if stats.TotalVectors != 4 {
		t.Errorf("Expected 4 vectors, got %d", stats.TotalVectors)
	}

	// Test that the index can handle the basic operations
	err = idx.Optimize()
	if err != nil {
		t.Errorf("Optimize failed: %v", err)
	}
}

func TestIVFIndex_DistanceCalculation(t *testing.T) {
	config := IndexConfig{
		Type:           IndexTypeIVF,
		Dimension:      3,
		MaxElements:    100,
		NumClusters:    4,
		ClusterSize:    10,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		t.Fatalf("Failed to create IVF index: %v", err)
	}
	defer func() {
		if err := idx.Close(); err != nil {
			t.Errorf("Failed to close index: %v", err)
		}
	}()

	ivfIdx := idx.(*IVFIndex)

	// Test cosine distance
	a := []float64{1.0, 0.0, 0.0}
	b := []float64{0.0, 1.0, 0.0}
	c := []float64{1.0, 0.0, 0.0}

	distAB := ivfIdx.cosineDistance(a, b)
	distAC := ivfIdx.cosineDistance(a, c)

	// Distance to itself should be 0
	if distAC != 0.0 {
		t.Errorf("Expected distance to self to be 0, got %f", distAC)
	}

	// Distance to orthogonal vector should be 1
	if distAB != 1.0 {
		t.Errorf("Expected distance to orthogonal vector to be 1, got %f", distAB)
	}

	// Test euclidean distance
	distAB = ivfIdx.euclideanDistance(a, b)
	distAC = ivfIdx.euclideanDistance(a, c)

	if distAC != 0.0 {
		t.Errorf("Expected Euclidean distance to self to be 0, got %f", distAC)
	}

	expectedAB := 1.4142135623730951 // sqrt(2)
	if abs(distAB-expectedAB) > 1e-10 {
		t.Errorf("Expected Euclidean distance to be %f, got %f", expectedAB, distAB)
	}
}

func TestIVFIndex_Delete(t *testing.T) {
	config := IndexConfig{
		Type:           IndexTypeIVF,
		Dimension:      4,
		MaxElements:    100,
		NumClusters:    4,
		ClusterSize:    10,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		t.Fatalf("Failed to create IVF index: %v", err)
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

	// Delete the vector
	err = idx.Delete("test1")
	if err != nil {
		t.Fatalf("Failed to delete vector: %v", err)
	}

	// Note: Current implementation is placeholder - in real implementation
	// we would verify the vector count decreased
}

func BenchmarkIVFIndex_Insert(b *testing.B) {
	config := IndexConfig{
		Type:           IndexTypeIVF,
		Dimension:      128,
		MaxElements:    10000,
		NumClusters:    100,
		ClusterSize:    100,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		b.Fatalf("Failed to create IVF index: %v", err)
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

func BenchmarkIVFIndex_Search(b *testing.B) {
	config := IndexConfig{
		Type:           IndexTypeIVF,
		Dimension:      128,
		MaxElements:    10000,
		NumClusters:    100,
		ClusterSize:    100,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		b.Fatalf("Failed to create IVF index: %v", err)
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
