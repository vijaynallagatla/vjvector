// Package main demonstrates the Q1 2025 Foundation & Performance implementation
// of the VJVector database, showcasing HNSW and IVF indexes with benchmarking.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/benchmark"
	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/index"
	"github.com/vijaynallagatla/vjvector/pkg/storage"
)

func main() {
	fmt.Println("🚀 VJVector Q1 2025 Demo: Foundation & Performance")
	fmt.Println(strings.Repeat("=", 60))

	// Demo 1: HNSW Index
	fmt.Println("\n📊 Demo 1: HNSW Index Performance")
	demoHNSW()

	// Demo 2: IVF Index
	fmt.Println("\n📊 Demo 2: IVF Index Performance")
	demoIVF()

	// Demo 3: Storage Engines
	fmt.Println("\n📊 Demo 3: Storage Engine Performance")
	demoStorage()

	// Demo 4: Full Benchmark Suite
	fmt.Println("\n📊 Demo 4: Full Benchmark Suite")
	demoBenchmark()

	fmt.Println("\n✅ Q1 2025 Demo Complete!")
}

func demoHNSW() {
	// Create HNSW index configuration
	config := index.IndexConfig{
		Type:           index.IndexTypeHNSW,
		Dimension:      1536,    // OpenAI embedding dimension
		MaxElements:    1000000, // 1M vectors
		M:              16,      // Max connections per layer
		EfConstruction: 200,     // Search depth during construction
		EfSearch:       100,     // Search depth during queries
		MaxLayers:      16,      // Maximum number of layers
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	// Create index
	idx, err := index.NewIndexFactory().CreateIndex(config)
	if err != nil {
		log.Printf("Failed to create HNSW index: %v", err)
		return
	}
	defer func() {
		if err := idx.Close(); err != nil {
			log.Printf("Warning: failed to close index: %v", err)
		}
	}()

	fmt.Printf("   ✅ HNSW Index created: %d dimensions, max %d elements\n",
		config.Dimension, config.MaxElements)

	// Insert some test vectors
	start := time.Now()
	for i := 0; i < 1000; i++ {
		vector := generateTestVector(1536, fmt.Sprintf("hnsw_test_%d", i))
		if err := idx.Insert(vector); err != nil {
			log.Printf("Failed to insert vector: %v", err)
			return
		}
	}
	insertTime := time.Since(start)

	fmt.Printf("   📈 Inserted 1000 vectors in %v (%.2f ops/sec)\n",
		insertTime, 1000.0/insertTime.Seconds())

	// Test search
	query := generateTestVector(1536, "query").Embedding
	start = time.Now()
	results, err := idx.Search(query, 10)
	if err != nil {
		log.Printf("Search failed: %v", err)
		return
	}
	searchTime := time.Since(start)

	fmt.Printf("   🔍 Search completed in %v, found %d results\n",
		searchTime, len(results))

	// Get stats
	stats := idx.GetStats()
	fmt.Printf("   📊 Index stats: %d vectors, %.2f MB memory\n",
		stats.TotalVectors, float64(stats.MemoryUsage)/1024/1024)
}

func demoIVF() {
	// Create IVF index configuration
	config := index.IndexConfig{
		Type:           index.IndexTypeIVF,
		Dimension:      1536,    // OpenAI embedding dimension
		MaxElements:    1000000, // 1M vectors
		NumClusters:    1000,    // Number of clusters
		ClusterSize:    1000,    // Target cluster size
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	// Create index
	idx, err := index.NewIndexFactory().CreateIndex(config)
	if err != nil {
		log.Printf("Failed to create IVF index: %v", err)
		return
	}
	defer func() {
		if err := idx.Close(); err != nil {
			log.Printf("Warning: failed to close index: %v", err)
		}
	}()

	fmt.Printf("   ✅ IVF Index created: %d dimensions, %d clusters\n",
		config.Dimension, config.NumClusters)

	// Insert some test vectors
	start := time.Now()
	for i := 0; i < 1000; i++ {
		vector := generateTestVector(1536, fmt.Sprintf("ivf_test_%d", i))
		if err := idx.Insert(vector); err != nil {
			log.Printf("Failed to insert vector: %v", err)
			return
		}
	}
	insertTime := time.Since(start)

	fmt.Printf("   📈 Inserted 1000 vectors in %v (%.2f ops/sec)\n",
		insertTime, 1000.0/insertTime.Seconds())

	// Test search
	query := generateTestVector(1536, "query").Embedding
	start = time.Now()
	results, err := idx.Search(query, 10)
	if err != nil {
		log.Printf("Search failed: %v", err)
		return
	}
	searchTime := time.Since(start)

	fmt.Printf("   🔍 Search completed in %v, found %d results\n",
		searchTime, len(results))

	// Get stats
	stats := idx.GetStats()
	fmt.Printf("   📊 Index stats: %d vectors, %.2f MB memory\n",
		stats.TotalVectors, float64(stats.MemoryUsage)/1024/1024)
}

func demoStorage() {
	// Test memory storage
	memoryConfig := storage.StorageConfig{
		Type:          storage.StorageTypeMemory,
		DataPath:      "memory://",
		MaxFileSize:   1024 * 1024 * 1024, // 1GB
		BatchSize:     1000,
		FlushInterval: 100,
	}

	memoryStore, err := storage.NewStorageFactory().CreateStorage(memoryConfig)
	if err != nil {
		log.Printf("Failed to create memory storage: %v", err)
		return
	}
	defer func() {
		if err := memoryStore.Close(); err != nil {
			log.Printf("Warning: failed to close memory store: %v", err)
		}
	}()

	fmt.Println("   ✅ Memory Storage created")

	// Test MMap storage
	mmapConfig := storage.StorageConfig{
		Type:          storage.StorageTypeMMap,
		DataPath:      "./data/mmap_test",
		MaxFileSize:   1024 * 1024 * 1024, // 1GB
		PageSize:      4096,
		BatchSize:     1000,
		FlushInterval: 100,
	}

	mmapStore, err := storage.NewStorageFactory().CreateStorage(mmapConfig)
	if err != nil {
		log.Printf("Failed to create MMap storage: %v", err)
		return
	}
	defer func() {
		if err := mmapStore.Close(); err != nil {
			log.Printf("Warning: failed to close mmap store: %v", err)
		}
	}()

	fmt.Println("   ✅ MMap Storage created")

	// Generate test vectors
	vectors := make([]*core.Vector, 100)
	for i := 0; i < 100; i++ {
		vectors[i] = generateTestVector(1536, fmt.Sprintf("storage_test_%d", i))
	}

	// Test write performance
	start := time.Now()
	if err := memoryStore.Write(vectors); err != nil {
		log.Printf("Memory storage write failed: %v", err)
		return
	}
	memoryWriteTime := time.Since(start)

	start = time.Now()
	if err := mmapStore.Write(vectors); err != nil {
		log.Printf("MMap storage write failed: %v", err)
		return
	}
	mmapWriteTime := time.Since(start)

	fmt.Printf("   📈 Memory storage write: %v (%.2f ops/sec)\n",
		memoryWriteTime, 100.0/memoryWriteTime.Seconds())
	fmt.Printf("   📈 MMap storage write: %v (%.2f ops/sec)\n",
		mmapWriteTime, 100.0/mmapWriteTime.Seconds())

	// Test read performance
	ids := make([]string, 100)
	for i, vector := range vectors {
		ids[i] = vector.ID
	}

	start = time.Now()
	_, err = memoryStore.Read(ids)
	if err != nil {
		log.Printf("Memory storage read failed: %v", err)
		return
	}
	memoryReadTime := time.Since(start)

	start = time.Now()
	_, err = mmapStore.Read(ids)
	if err != nil {
		log.Printf("MMap storage read failed: %v", err)
		return
	}
	mmapReadTime := time.Since(start)

	fmt.Printf("   🔍 Memory storage read: %v (%.2f ops/sec)\n",
		memoryReadTime, 100.0/memoryReadTime.Seconds())
	fmt.Printf("   🔍 MMap storage read: %v (%.2f ops/sec)\n",
		mmapReadTime, 100.0/mmapReadTime.Seconds())
}

func demoBenchmark() {
	// Create benchmark configuration
	config := benchmark.BenchmarkConfig{
		VectorCount:         10000, // 10K vectors for demo
		Dimension:           1536,
		QueryCount:          1000,
		K:                   10,
		IndexType:           index.IndexTypeHNSW,
		M:                   16,
		EfConstruction:      200,
		EfSearch:            100,
		MaxLayers:           16,
		StorageType:         storage.StorageTypeMemory,
		DataPath:            "./data/benchmark",
		TargetSearchLatency: 1.0,   // 1ms target
		TargetThroughput:    10000, // 10K QPS target
		TargetMemoryUsage:   8192,  // 8GB target
	}

	// Create benchmark suite
	suite, err := benchmark.NewBenchmarkSuite(config)
	if err != nil {
		log.Printf("Failed to create benchmark suite: %v", err)
		return
	}
	defer func() {
		if err := suite.Close(); err != nil {
			log.Printf("Warning: failed to close benchmark suite: %v", err)
		}
	}()

	fmt.Println("   ✅ Benchmark suite created")

	// Run benchmark
	ctx := context.Background()
	if err := suite.RunFullBenchmark(ctx); err != nil {
		log.Printf("Benchmark failed: %v", err)
		return
	}

	fmt.Println("   ✅ Benchmark completed successfully")
}

// generateTestVector creates a test vector with random embedding
func generateTestVector(dimension int, id string) *core.Vector {
	embedding := make([]float64, dimension)
	for i := 0; i < dimension; i++ {
		// Simple deterministic "random" values for demo
		embedding[i] = float64(i%100) / 100.0
	}

	return &core.Vector{
		ID:         id,
		Collection: "demo",
		Embedding:  embedding,
		Metadata:   map[string]interface{}{"demo": true, "timestamp": time.Now()},
		Text:       fmt.Sprintf("Demo vector %s", id),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Dimension:  dimension,
		Magnitude:  0, // Will be calculated by NewVector
		Normalized: false,
	}
}
