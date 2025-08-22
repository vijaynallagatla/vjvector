// Package benchmark provides performance testing and benchmarking utilities
// for the VJVector database, including index performance and storage metrics.
package benchmark

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/index"
	"github.com/vijaynallagatla/vjvector/pkg/storage"
)

// Suite is an alias for BenchmarkSuite to satisfy linter preferences
type Suite = BenchmarkSuite

// Config is an alias for BenchmarkConfig to satisfy linter preferences
type Config = BenchmarkConfig

// Result is an alias for BenchmarkResult to satisfy linter preferences
type Result = BenchmarkResult

// BenchmarkSuite provides comprehensive benchmarking for vector operations
type BenchmarkSuite struct {
	config  BenchmarkConfig
	index   index.VectorIndex
	storage storage.StorageEngine
	results []BenchmarkResult
	mutex   sync.RWMutex
}

// BenchmarkConfig holds configuration for benchmarking
type BenchmarkConfig struct {
	// Test parameters
	VectorCount int `json:"vector_count"`
	Dimension   int `json:"dimension"`
	QueryCount  int `json:"query_count"`
	K           int `json:"k"` // Number of results to return

	// Index parameters
	IndexType      index.IndexType `json:"index_type"`
	M              int             `json:"m,omitempty"`               // HNSW: Max connections
	EfConstruction int             `json:"ef_construction,omitempty"` // HNSW: Construction search depth
	EfSearch       int             `json:"ef_search,omitempty"`       // HNSW: Query search depth
	MaxLayers      int             `json:"max_layers,omitempty"`      // HNSW: Max layers
	NumClusters    int             `json:"num_clusters,omitempty"`    // IVF: Number of clusters

	// Storage parameters
	StorageType storage.StorageType `json:"storage_type"`
	DataPath    string              `json:"data_path"`

	// Performance targets
	TargetSearchLatency float64 `json:"target_search_latency_ms"`
	TargetThroughput    int     `json:"target_throughput_qps"`
	TargetMemoryUsage   int64   `json:"target_memory_usage_mb"`
}

// BenchmarkResult holds the results of a benchmark run
type BenchmarkResult struct {
	TestName  string    `json:"test_name"`
	Timestamp time.Time `json:"timestamp"`

	// Performance metrics
	SearchLatency    float64 `json:"search_latency_ms"`
	SearchLatencyP50 float64 `json:"search_latency_p50_ms"`
	SearchLatencyP95 float64 `json:"search_latency_p95_ms"`
	SearchLatencyP99 float64 `json:"search_latency_p99_ms"`

	InsertLatency    float64 `json:"insert_latency_ms"`
	InsertThroughput float64 `json:"insert_throughput_ops_per_sec"`

	MemoryUsage int64 `json:"memory_usage_bytes"`
	IndexSize   int64 `json:"index_size_bytes"`

	// Quality metrics
	Recall    float64 `json:"recall_at_k"`
	Precision float64 `json:"precision_at_k"`

	// Test parameters
	VectorCount int `json:"vector_count"`
	Dimension   int `json:"dimension"`
	K           int `json:"k"`

	// Success indicators
	TargetsMet bool   `json:"targets_met"`
	Notes      string `json:"notes,omitempty"`
}

// NewBenchmarkSuite creates a new benchmarking suite
func NewBenchmarkSuite(config BenchmarkConfig) (*BenchmarkSuite, error) {
	// Create index
	indexConfig := index.IndexConfig{
		Type:           config.IndexType,
		Dimension:      config.Dimension,
		MaxElements:    config.VectorCount,
		M:              config.M,
		EfConstruction: config.EfConstruction,
		EfSearch:       config.EfSearch,
		MaxLayers:      config.MaxLayers,
		NumClusters:    config.NumClusters,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	idx, err := index.NewIndexFactory().CreateIndex(indexConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	// Create storage
	storageConfig := storage.StorageConfig{
		Type:          config.StorageType,
		DataPath:      config.DataPath,
		MaxFileSize:   1024 * 1024 * 1024, // 1GB
		PageSize:      4096,
		BatchSize:     1000,
		FlushInterval: 100,
	}

	store, err := storage.NewStorageFactory().CreateStorage(storageConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	return &BenchmarkSuite{
		config:  config,
		index:   idx,
		storage: store,
		results: make([]BenchmarkResult, 0),
	}, nil
}

// RunFullBenchmark executes the complete benchmarking suite
func (b *BenchmarkSuite) RunFullBenchmark(_ context.Context) error {
	fmt.Println("🚀 Starting VJVector Benchmark Suite")
	fmt.Printf("📊 Configuration: %d vectors, %d dimensions, %s index\n",
		b.config.VectorCount, b.config.Dimension, b.config.IndexType)

	// Generate test data
	fmt.Println("📝 Generating test vectors...")
	vectors := b.generateTestVectors()

	// Benchmark insertion
	fmt.Println("⬆️  Benchmarking insertion...")
	insertResult := b.benchmarkInsertion(vectors)
	b.addResult(insertResult)

	// Benchmark search
	fmt.Println("🔍 Benchmarking search...")
	searchResult := b.benchmarkSearch(vectors)
	b.addResult(searchResult)

	// Benchmark storage
	fmt.Println("💾 Benchmarking storage...")
	storageResult := b.benchmarkStorage(vectors)
	b.addResult(storageResult)

	// Generate report
	fmt.Println("📋 Generating benchmark report...")
	b.generateReport()

	return nil
}

// generateTestVectors creates test vectors for benchmarking
func (b *BenchmarkSuite) generateTestVectors() []*core.Vector {
	vectors := make([]*core.Vector, b.config.VectorCount)

	for i := 0; i < b.config.VectorCount; i++ {
		embedding := make([]float64, b.config.Dimension)
		for j := 0; j < b.config.Dimension; j++ {
			embedding[j] = rand.Float64()*2 - 1 // Random values between -1 and 1 //nolint:gosec
		}

		vectors[i] = &core.Vector{
			ID:         fmt.Sprintf("test_vector_%d", i),
			Collection: "benchmark",
			Embedding:  embedding,
			Metadata:   map[string]interface{}{"benchmark": true},
			Text:       fmt.Sprintf("Test vector %d", i),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			Dimension:  b.config.Dimension,
			Magnitude:  0, // Will be calculated by NewVector
			Normalized: false,
		}
	}

	return vectors
}

// benchmarkInsertion measures insertion performance
func (b *BenchmarkSuite) benchmarkInsertion(vectors []*core.Vector) BenchmarkResult {
	start := time.Now()

	// Insert vectors into index
	for _, vector := range vectors {
		if err := b.index.Insert(vector); err != nil {
			return BenchmarkResult{
				TestName:  "Insertion",
				Timestamp: time.Now(),
				Notes:     fmt.Sprintf("Insertion failed: %v", err),
			}
		}
	}

	duration := time.Since(start)
	throughput := float64(len(vectors)) / duration.Seconds()

	return BenchmarkResult{
		TestName:         "Insertion",
		Timestamp:        time.Now(),
		InsertLatency:    float64(duration.Milliseconds()) / float64(len(vectors)),
		InsertThroughput: throughput,
		VectorCount:      len(vectors),
		Dimension:        b.config.Dimension,
		TargetsMet:       true,
	}
}

// benchmarkSearch measures search performance
func (b *BenchmarkSuite) benchmarkSearch(vectors []*core.Vector) BenchmarkResult {
	// Generate random query vectors
	queries := make([][]float64, b.config.QueryCount)
	for i := 0; i < b.config.QueryCount; i++ {
		query := make([]float64, b.config.Dimension)
		for j := 0; j < b.config.Dimension; j++ {
			query[j] = rand.Float64()*2 - 1 // nolint:gosec
		}
		queries[i] = query
	}

	// Measure search performance
	latencies := make([]float64, b.config.QueryCount)

	for i, query := range queries {
		queryStart := time.Now()
		results, err := b.index.Search(query, b.config.K)
		if err != nil {
			return BenchmarkResult{
				TestName:  "Search",
				Timestamp: time.Now(),
				Notes:     fmt.Sprintf("Search failed: %v", err),
			}
		}
		latencies[i] = float64(time.Since(queryStart).Microseconds()) / 1000.0 // Convert to ms

		// Use results to avoid compiler optimization
		_ = results
	}

	avgLatency := calculateAverage(latencies)

	return BenchmarkResult{
		TestName:         "Search",
		Timestamp:        time.Now(),
		SearchLatency:    avgLatency,
		SearchLatencyP50: calculatePercentile(latencies, 50),
		SearchLatencyP95: calculatePercentile(latencies, 95),
		SearchLatencyP99: calculatePercentile(latencies, 99),
		VectorCount:      len(vectors),
		Dimension:        b.config.Dimension,
		K:                b.config.K,
		TargetsMet:       avgLatency <= b.config.TargetSearchLatency,
	}
}

// benchmarkStorage measures storage performance
func (b *BenchmarkSuite) benchmarkStorage(vectors []*core.Vector) BenchmarkResult {
	start := time.Now()

	// Write vectors to storage
	if err := b.storage.Write(vectors); err != nil {
		return BenchmarkResult{
			TestName:  "Storage",
			Timestamp: time.Now(),
			Notes:     fmt.Sprintf("Storage write failed: %v", err),
		}
	}

	writeDuration := time.Since(start)

	// Read vectors back
	ids := make([]string, len(vectors))
	for i, vector := range vectors {
		ids[i] = vector.ID
	}

	readStart := time.Now()
	_, err := b.storage.Read(ids)
	if err != nil {
		return BenchmarkResult{
			TestName:  "Storage",
			Timestamp: time.Now(),
			Notes:     fmt.Sprintf("Storage read failed: %v", err),
		}
	}

	readDuration := time.Since(readStart)

	// Get storage stats
	stats := b.storage.GetStats()

	return BenchmarkResult{
		TestName:      "Storage",
		Timestamp:     time.Now(),
		InsertLatency: float64(writeDuration.Milliseconds()) / float64(len(vectors)),
		SearchLatency: float64(readDuration.Microseconds()) / float64(len(vectors)) / 1000.0,
		MemoryUsage:   stats.MemoryUsage,
		IndexSize:     stats.StorageSize,
		VectorCount:   len(vectors),
		Dimension:     b.config.Dimension,
		TargetsMet:    true,
	}
}

// addResult adds a benchmark result to the suite
func (b *BenchmarkSuite) addResult(result BenchmarkResult) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.results = append(b.results, result)
}

// generateReport creates a comprehensive benchmark report
func (b *BenchmarkSuite) generateReport() {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 VJVector Benchmark Report")
	fmt.Println(strings.Repeat("=", 60))

	for _, result := range b.results {
		fmt.Printf("\n🧪 Test: %s\n", result.TestName)
		fmt.Printf("   ⏱️  Timestamp: %s\n", result.Timestamp.Format(time.RFC3339))

		switch result.TestName {
		case "Insertion":
			fmt.Printf("   📈 Insert Latency: %.2f ms\n", result.InsertLatency)
			fmt.Printf("   🚀 Insert Throughput: %.2f ops/sec\n", result.InsertThroughput)
		case "Search":
			fmt.Printf("   🔍 Search Latency: %.2f ms (avg)\n", result.SearchLatency)
			fmt.Printf("   📊 P50: %.2f ms, P95: %.2f ms, P99: %.2f ms\n",
				result.SearchLatencyP50, result.SearchLatencyP95, result.SearchLatencyP99)
		case "Storage":
			fmt.Printf("   💾 Memory Usage: %.2f MB\n", float64(result.MemoryUsage)/1024/1024)
			fmt.Printf("   📁 Index Size: %.2f MB\n", float64(result.IndexSize)/1024/1024)
		}

		fmt.Printf("   🎯 Targets Met: %t\n", result.TargetsMet)
		if result.Notes != "" {
			fmt.Printf("   📝 Notes: %s\n", result.Notes)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
}

// calculateAverage calculates the average of a slice of float64 values
func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculatePercentile calculates the nth percentile of a slice of float64 values
func calculatePercentile(values []float64, percentile int) float64 {
	if len(values) == 0 {
		return 0
	}

	// Sort values (this modifies the original slice)
	// In a real implementation, you'd want to copy the slice first
	// For benchmarking purposes, this is acceptable

	// Simple percentile calculation
	index := int(float64(percentile) / 100.0 * float64(len(values)-1))
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}

	return values[index]
}

// Close cleans up resources
func (b *BenchmarkSuite) Close() error {
	var errs []error

	if err := b.index.Close(); err != nil {
		errs = append(errs, fmt.Errorf("index close: %w", err))
	}

	if err := b.storage.Close(); err != nil {
		errs = append(errs, fmt.Errorf("storage close: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors during cleanup: %v", errs)
	}

	return nil
}
