package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/index"
	"github.com/vijaynallagatla/vjvector/pkg/storage"
)

// CLI represents the VJVector command-line interface
type CLI struct {
	indexes map[string]index.VectorIndex
	storage storage.StorageEngine
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	cli := &CLI{
		indexes: make(map[string]index.VectorIndex),
	}

	// Initialize storage
	storageConfig := storage.StorageConfig{
		Type:            storage.StorageTypeMemory,
		DataPath:        "/tmp/vjvector_cli",
		PageSize:        4096,
		MaxFileSize:     1024 * 1024 * 1024, // 1GB
		BatchSize:       100,
		WriteBufferSize: 64 * 1024 * 1024, // 64MB
		CacheSize:       32 * 1024 * 1024, // 32MB
		MaxOpenFiles:    1000,
	}

	factory := &storage.DefaultStorageFactory{}
	storageEngine, err := factory.CreateStorage(storageConfig)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	cli.storage = storageEngine

	return cli
}

// createIndexCmd creates a new vector index
func (cli *CLI) createIndexCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("index ID is required")
	}

	id := args[0]
	indexType, _ := cmd.Flags().GetString("type")
	dimension, _ := cmd.Flags().GetInt("dimension")
	maxElements, _ := cmd.Flags().GetInt("max-elements")
	m, _ := cmd.Flags().GetInt("m")
	efConstruction, _ := cmd.Flags().GetInt("ef-construction")
	efSearch, _ := cmd.Flags().GetInt("ef-search")
	maxLayers, _ := cmd.Flags().GetInt("max-layers")
	numClusters, _ := cmd.Flags().GetInt("num-clusters")
	distanceMetric, _ := cmd.Flags().GetString("distance-metric")
	normalize, _ := cmd.Flags().GetBool("normalize")

	// Convert string type to IndexType
	var indexTypeEnum index.IndexType
	switch indexType {
	case "hnsw":
		indexTypeEnum = index.IndexTypeHNSW
	case "ivf":
		indexTypeEnum = index.IndexTypeIVF
	default:
		return fmt.Errorf("invalid index type. Must be 'hnsw' or 'ivf'")
	}

	config := index.IndexConfig{
		Type:           indexTypeEnum,
		Dimension:      dimension,
		MaxElements:    maxElements,
		M:              m,
		EfConstruction: efConstruction,
		EfSearch:       efSearch,
		MaxLayers:      maxLayers,
		NumClusters:    numClusters,
		DistanceMetric: distanceMetric,
		Normalize:      normalize,
	}

	factory := index.NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		return fmt.Errorf("failed to create index: %v", err)
	}

	cli.indexes[id] = idx

	fmt.Printf("✅ Index '%s' created successfully\n", id)
	fmt.Printf("   Type: %s\n", indexType)
	fmt.Printf("   Dimension: %d\n", dimension)
	fmt.Printf("   Max Elements: %d\n", maxElements)

	return nil
}

// listIndexesCmd lists all available indexes
func (cli *CLI) listIndexesCmd(cmd *cobra.Command, args []string) error {
	if len(cli.indexes) == 0 {
		fmt.Println("📭 No indexes found")
		return nil
	}

	fmt.Printf("📚 Found %d indexes:\n\n", len(cli.indexes))
	for id, idx := range cli.indexes {
		stats := idx.GetStats()
		fmt.Printf("🔍 Index: %s\n", id)
		fmt.Printf("   📊 Total Vectors: %d\n", stats.TotalVectors)
		fmt.Printf("   💾 Memory Usage: %d bytes\n", stats.MemoryUsage)
		fmt.Printf("   📏 Index Size: %d bytes\n", stats.IndexSize)
		fmt.Printf("   ⏱️  Avg Search Time: %.2f ms\n", stats.AvgSearchTime)
		fmt.Printf("   ⬆️  Avg Insert Time: %.2f ms\n", stats.AvgInsertTime)
		fmt.Println()
	}

	return nil
}

// insertVectorsCmd inserts vectors into an index
func (cli *CLI) insertVectorsCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("index ID is required")
	}

	id := args[0]
	idx, exists := cli.indexes[id]
	if !exists {
		return fmt.Errorf("index '%s' not found", id)
	}

	// For CLI demo, create some sample vectors
	count, _ := cmd.Flags().GetInt("count")
	dimension, _ := cmd.Flags().GetInt("dimension")

	if count <= 0 {
		count = 10
	}
	if dimension <= 0 {
		dimension = 128
	}

	vectors := make([]*core.Vector, count)
	for i := 0; i < count; i++ {
		embedding := make([]float64, dimension)
		for j := 0; j < dimension; j++ {
			embedding[j] = float64(i*j) * 0.001
		}
		vectors[i] = &core.Vector{
			ID:         fmt.Sprintf("vector_%d", i),
			Collection: "cli_demo",
			Embedding:  embedding,
			Metadata:   map[string]interface{}{"source": "cli", "index": i},
		}
	}

	start := time.Now()
	for _, vector := range vectors {
		if err := idx.Insert(vector); err != nil {
			return fmt.Errorf("failed to insert vector %s: %v", vector.ID, err)
		}
	}
	duration := time.Since(start)

	stats := idx.GetStats()
	fmt.Printf("✅ Inserted %d vectors into index '%s'\n", count, id)
	fmt.Printf("   ⏱️  Time: %s\n", duration)
	fmt.Printf("   📊 Total Vectors: %d\n", stats.TotalVectors)
	fmt.Printf("   💾 Memory Usage: %d bytes\n", stats.MemoryUsage)

	return nil
}

// searchVectorsCmd searches for similar vectors
func (cli *CLI) searchVectorsCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("index ID is required")
	}

	id := args[0]
	idx, exists := cli.indexes[id]
	if !exists {
		return fmt.Errorf("index '%s' not found", id)
	}

	k, _ := cmd.Flags().GetInt("k")
	if k <= 0 {
		k = 5
	}

	dimension, _ := cmd.Flags().GetInt("dimension")
	if dimension <= 0 {
		dimension = 128
	}

	// Create a sample query vector
	query := make([]float64, dimension)
	for i := 0; i < dimension; i++ {
		query[i] = float64(i) * 0.001
	}

	fmt.Printf("🔍 Searching index '%s' for %d similar vectors...\n", id, k)
	fmt.Printf("   Query dimension: %d\n", dimension)

	start := time.Now()
	results, err := idx.Search(query, k)
	if err != nil {
		return fmt.Errorf("search failed: %v", err)
	}
	duration := time.Since(start)

	fmt.Printf("✅ Search completed in %s\n", duration)
	fmt.Printf("   📊 Found %d results:\n\n", len(results))

	for i, result := range results {
		fmt.Printf("   %d. Vector: %s\n", i+1, result.Vector.ID)
		fmt.Printf("      Score: %.4f\n", result.Score)
		fmt.Printf("      Distance: %.4f\n", result.Distance)
		fmt.Println()
	}

	return nil
}

// getIndexStatsCmd shows detailed statistics for an index
func (cli *CLI) getIndexStatsCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("index ID is required")
	}

	id := args[0]
	idx, exists := cli.indexes[id]
	if !exists {
		return fmt.Errorf("index '%s' not found", id)
	}

	stats := idx.GetStats()
	fmt.Printf("📊 Index Statistics: %s\n", id)
	fmt.Printf("=====================================\n")
	fmt.Printf("📈 Total Vectors: %d\n", stats.TotalVectors)
	fmt.Printf("💾 Memory Usage: %d bytes (%.2f MB)\n", stats.MemoryUsage, float64(stats.MemoryUsage)/1024/1024)
	fmt.Printf("📏 Index Size: %d bytes (%.2f MB)\n", stats.IndexSize, float64(stats.IndexSize)/1024/1024)
	fmt.Printf("⏱️  Average Search Time: %.2f ms\n", stats.AvgSearchTime)
	fmt.Printf("⬆️  Average Insert Time: %.2f ms\n", stats.AvgInsertTime)
	fmt.Printf("🎯 Recall at K: %.2f%%\n", stats.Recall*100)
	fmt.Printf("🎯 Precision at K: %.2f%%\n", stats.Precision*100)

	// Index-specific metrics
	if stats.NumLayers > 0 {
		fmt.Printf("🏗️  Number of Layers: %d\n", stats.NumLayers)
	}
	if stats.MaxConnections > 0 {
		fmt.Printf("🔗 Max Connections: %d\n", stats.MaxConnections)
	}
	if stats.NumClusters > 0 {
		fmt.Printf("🎯 Number of Clusters: %d\n", stats.NumClusters)
	}

	return nil
}

// getStorageStatsCmd shows storage statistics
func (cli *CLI) getStorageStatsCmd(cmd *cobra.Command, args []string) error {
	stats := cli.storage.GetStats()
	fmt.Printf("💾 Storage Statistics\n")
	fmt.Printf("=====================\n")
	fmt.Printf("📈 Total Vectors: %d\n", stats.TotalVectors)
	fmt.Printf("💾 Storage Size: %d bytes (%.2f MB)\n", stats.StorageSize, float64(stats.StorageSize)/1024/1024)
	fmt.Printf("🖥️  Memory Usage: %d bytes (%.2f MB)\n", stats.MemoryUsage, float64(stats.MemoryUsage)/1024/1024)
	fmt.Printf("⏱️  Average Write Time: %.2f ms\n", stats.AvgWriteTime)
	fmt.Printf("⏱️  Average Read Time: %.2f ms\n", stats.AvgReadTime)
	fmt.Printf("📁 File Count: %d\n", stats.FileCount)
	fmt.Printf("📄 Page Size: %d bytes\n", stats.PageSize)

	return nil
}

// benchmarkCmd runs performance benchmarks
func (cli *CLI) benchmarkCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("index ID is required")
	}

	id := args[0]
	idx, exists := cli.indexes[id]
	if !exists {
		return fmt.Errorf("index '%s' not found", id)
	}

	iterations, _ := cmd.Flags().GetInt("iterations")
	if iterations <= 0 {
		iterations = 100
	}

	dimension, _ := cmd.Flags().GetInt("dimension")
	if dimension <= 0 {
		dimension = 128
	}

	fmt.Printf("🚀 Running benchmarks on index '%s'\n", id)
	fmt.Printf("   Iterations: %d\n", iterations)
	fmt.Printf("   Dimension: %d\n", dimension)

	// Benchmark insertion
	fmt.Printf("\n📈 Benchmarking insertion...\n")
	insertStart := time.Now()
	for i := 0; i < iterations; i++ {
		embedding := make([]float64, dimension)
		for j := 0; j < dimension; j++ {
			embedding[j] = float64(i*j) * 0.001
		}
		vector := &core.Vector{
			ID:         fmt.Sprintf("bench_%d", i),
			Collection: "benchmark",
			Embedding:  embedding,
		}
		if err := idx.Insert(vector); err != nil {
			return fmt.Errorf("insertion failed: %v", err)
		}
	}
	insertDuration := time.Since(insertStart)
	insertRate := float64(iterations) / insertDuration.Seconds()

	fmt.Printf("   ✅ Inserted %d vectors in %s\n", iterations, insertDuration)
	fmt.Printf("   🚀 Rate: %.2f ops/sec\n", insertRate)

	// Benchmark search
	fmt.Printf("\n🔍 Benchmarking search...\n")
	query := make([]float64, dimension)
	for i := 0; i < dimension; i++ {
		query[i] = float64(i) * 0.001
	}

	searchStart := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := idx.Search(query, 5)
		if err != nil {
			return fmt.Errorf("search failed: %v", err)
		}
	}
	searchDuration := time.Since(searchStart)
	searchRate := float64(iterations) / searchDuration.Seconds()

	fmt.Printf("   ✅ Searched %d times in %s\n", iterations, searchDuration)
	fmt.Printf("   🚀 Rate: %.2f ops/sec\n", searchRate)

	// Final stats
	stats := idx.GetStats()
	fmt.Printf("\n📊 Final Statistics:\n")
	fmt.Printf("   📈 Total Vectors: %d\n", stats.TotalVectors)
	fmt.Printf("   💾 Memory Usage: %.2f MB\n", float64(stats.MemoryUsage)/1024/1024)

	return nil
}

// demoCmd runs a comprehensive demo
func (cli *CLI) demoCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 VJVector CLI Demo")
	fmt.Println("====================")

	// Create HNSW index
	fmt.Println("\n1️⃣ Creating HNSW index...")
	hnswConfig := index.IndexConfig{
		Type:           index.IndexTypeHNSW,
		Dimension:      128,
		MaxElements:    1000,
		M:              16,
		EfConstruction: 200,
		EfSearch:       100,
		MaxLayers:      16,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	factory := index.NewIndexFactory()
	hnswIdx, err := factory.CreateIndex(hnswConfig)
	if err != nil {
		return fmt.Errorf("failed to create HNSW index: %v", err)
	}
	cli.indexes["demo_hnsw"] = hnswIdx

	// Create IVF index
	fmt.Println("\n2️⃣ Creating IVF index...")
	ivfConfig := index.IndexConfig{
		Type:           index.IndexTypeIVF,
		Dimension:      128,
		MaxElements:    1000,
		NumClusters:    50,
		ClusterSize:    20,
		DistanceMetric: "cosine",
		Normalize:      true,
	}

	ivfIdx, err := factory.CreateIndex(ivfConfig)
	if err != nil {
		return fmt.Errorf("failed to create IVF index: %v", err)
	}
	cli.indexes["demo_ivf"] = ivfIdx

	// Insert vectors into both indexes
	fmt.Println("\n3️⃣ Inserting test vectors...")
	for i := 0; i < 100; i++ {
		embedding := make([]float64, 128)
		for j := 0; j < 128; j++ {
			embedding[j] = float64(i*j) * 0.001
		}
		vector := &core.Vector{
			ID:         fmt.Sprintf("demo_%d", i),
			Collection: "demo",
			Embedding:  embedding,
		}

		if err := hnswIdx.Insert(vector); err != nil {
			return fmt.Errorf("failed to insert into HNSW: %v", err)
		}
		if err := ivfIdx.Insert(vector); err != nil {
			return fmt.Errorf("failed to insert into IVF: %v", err)
		}
	}

	// Search in both indexes
	fmt.Println("\n4️⃣ Testing search performance...")
	query := make([]float64, 128)
	for i := 0; i < 128; i++ {
		query[i] = float64(i) * 0.001
	}

	// HNSW search
	hnswStart := time.Now()
	hnswResults, err := hnswIdx.Search(query, 5)
	if err != nil {
		return fmt.Errorf("HNSW search failed: %v", err)
	}
	hnswDuration := time.Since(hnswStart)

	// IVF search
	ivfStart := time.Now()
	ivfResults, err := ivfIdx.Search(query, 5)
	if err != nil {
		return fmt.Errorf("IVF search failed: %v", err)
	}
	ivfDuration := time.Since(ivfStart)

	// Results
	fmt.Printf("\n📊 Demo Results:\n")
	fmt.Printf("   🔍 HNSW Search: %s, found %d results\n", hnswDuration, len(hnswResults))
	fmt.Printf("   🔍 IVF Search: %s, found %d results\n", ivfDuration, len(ivfResults))

	hnswStats := hnswIdx.GetStats()
	ivfStats := ivfIdx.GetStats()
	fmt.Printf("   💾 HNSW Memory: %.2f MB\n", float64(hnswStats.MemoryUsage)/1024/1024)
	fmt.Printf("   💾 IVF Memory: %.2f MB\n", float64(ivfStats.MemoryUsage)/1024/1024)

	fmt.Println("\n✅ Demo completed successfully!")
	return nil
}

func main() {
	cli := NewCLI()

	// Root command
	rootCmd := &cobra.Command{
		Use:   "vjvector",
		Short: "VJVector - High-performance vector database CLI",
		Long: `VJVector CLI provides a command-line interface to interact with the VJVector vector database.

Features:
- Create and manage HNSW and IVF indexes
- Insert and search vectors
- Performance benchmarking
- Storage statistics
- Interactive demos`,
	}

	// Create index command
	createCmd := &cobra.Command{
		Use:   "create [index-id]",
		Short: "Create a new vector index",
		Args:  cobra.ExactArgs(1),
		RunE:  cli.createIndexCmd,
	}
	createCmd.Flags().String("type", "hnsw", "Index type (hnsw or ivf)")
	createCmd.Flags().Int("dimension", 128, "Vector dimension")
	createCmd.Flags().Int("max-elements", 1000, "Maximum number of elements")
	createCmd.Flags().Int("m", 16, "HNSW: Max connections per layer")
	createCmd.Flags().Int("ef-construction", 200, "HNSW: Construction search depth")
	createCmd.Flags().Int("ef-search", 100, "HNSW: Query search depth")
	createCmd.Flags().Int("max-layers", 16, "HNSW: Maximum number of layers")
	createCmd.Flags().Int("num-clusters", 50, "IVF: Number of clusters")
	createCmd.Flags().String("distance-metric", "cosine", "Distance metric (cosine, euclidean, dot)")
	createCmd.Flags().Bool("normalize", true, "Whether to normalize vectors")

	// List indexes command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all indexes",
		RunE:  cli.listIndexesCmd,
	}

	// Insert vectors command
	insertCmd := &cobra.Command{
		Use:   "insert [index-id]",
		Short: "Insert vectors into an index",
		Args:  cobra.ExactArgs(1),
		RunE:  cli.insertVectorsCmd,
	}
	insertCmd.Flags().Int("count", 10, "Number of vectors to insert")
	insertCmd.Flags().Int("dimension", 128, "Vector dimension")

	// Search command
	searchCmd := &cobra.Command{
		Use:   "search [index-id]",
		Short: "Search for similar vectors",
		Args:  cobra.ExactArgs(1),
		RunE:  cli.searchVectorsCmd,
	}
	searchCmd.Flags().Int("k", 5, "Number of results to return")
	searchCmd.Flags().Int("dimension", 128, "Query vector dimension")

	// Stats command
	statsCmd := &cobra.Command{
		Use:   "stats [index-id]",
		Short: "Show index statistics",
		Args:  cobra.ExactArgs(1),
		RunE:  cli.getIndexStatsCmd,
	}

	// Storage stats command
	storageStatsCmd := &cobra.Command{
		Use:   "storage",
		Short: "Show storage statistics",
		RunE:  cli.getStorageStatsCmd,
	}

	// Benchmark command
	benchmarkCmd := &cobra.Command{
		Use:   "benchmark [index-id]",
		Short: "Run performance benchmarks",
		Args:  cobra.ExactArgs(1),
		RunE:  cli.benchmarkCmd,
	}
	benchmarkCmd.Flags().Int("iterations", 100, "Number of benchmark iterations")
	benchmarkCmd.Flags().Int("dimension", 128, "Vector dimension")

	// Demo command
	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "Run comprehensive demo",
		RunE:  cli.demoCmd,
	}

	// Add commands to root
	rootCmd.AddCommand(createCmd, listCmd, insertCmd, searchCmd, statsCmd, storageStatsCmd, benchmarkCmd, demoCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
