package benchmark

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/embedding"
	"github.com/vijaynallagatla/vjvector/pkg/index"
	"github.com/vijaynallagatla/vjvector/pkg/rag"
)

// AIIntegrationBenchmark benchmarks AI integration features
type AIIntegrationBenchmark struct {
	logger *slog.Logger
}

// AIIntegrationResult represents a single AI integration benchmark result
type AIIntegrationResult struct {
	Operation    string                 `json:"operation"`
	Provider     string                 `json:"provider"`
	Duration     time.Duration          `json:"duration"`
	Throughput   float64                `json:"throughput"`   // operations per second
	Latency      time.Duration          `json:"latency"`      // average latency
	MemoryUsage  int64                  `json:"memory_usage"` // bytes
	ErrorCount   int                    `json:"error_count"`
	SuccessCount int                    `json:"success_count"`
	SuccessRate  float64                `json:"success_rate"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AIIntegrationSuite represents a complete AI integration benchmark suite
type AIIntegrationSuite struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Results     []AIIntegrationResult `json:"results"`
	Summary     AIIntegrationSummary  `json:"summary"`
}

// AIIntegrationSummary provides overall AI integration benchmark statistics
type AIIntegrationSummary struct {
	TotalOperations   int           `json:"total_operations"`
	TotalDuration     time.Duration `json:"total_duration"`
	OverallThroughput float64       `json:"overall_throughput"`
	AverageLatency    time.Duration `json:"average_latency"`
	SuccessRate       float64       `json:"success_rate"`
	MemoryEfficiency  float64       `json:"memory_efficiency"` // operations per MB
}

// NewAIIntegrationBenchmark creates a new AI integration benchmark
func NewAIIntegrationBenchmark() *AIIntegrationBenchmark {
	return &AIIntegrationBenchmark{
		logger: slog.Default(),
	}
}

// RunEmbeddingBenchmark benchmarks embedding generation performance
func (b *AIIntegrationBenchmark) RunEmbeddingBenchmark(
	ctx context.Context,
	embeddingService embedding.Service,
	texts []string,
	iterations int,
) AIIntegrationResult {
	b.logger.Info("Starting embedding generation benchmark",
		"provider", "embedding-service",
		"texts", len(texts),
		"iterations", iterations)

	startTime := time.Now()
	var totalLatency time.Duration
	var errorCount int
	var successCount int

	// Warm up
	for i := 0; i < 5; i++ {
		req := &embedding.EmbeddingRequest{
			Texts: texts[:min(len(texts), 10)],
		}
		_, err := embeddingService.GenerateEmbeddings(ctx, req)
		if err != nil {
			b.logger.Warn("Warm-up embedding generation failed", "error", err)
		}
	}

	// Run benchmark
	for i := 0; i < iterations; i++ {
		req := &embedding.EmbeddingRequest{
			Texts: texts,
		}

		iterStart := time.Now()
		_, err := embeddingService.GenerateEmbeddings(ctx, req)
		iterDuration := time.Since(iterStart)

		if err != nil {
			errorCount++
			b.logger.Warn("Embedding generation failed", "iteration", i, "error", err)
		} else {
			successCount++
			totalLatency += iterDuration
		}
	}

	totalDuration := time.Since(startTime)
	successRate := float64(successCount) / float64(iterations)
	avgLatency := totalLatency / time.Duration(successCount)
	throughput := float64(successCount) / totalDuration.Seconds()

	return AIIntegrationResult{
		Operation:    "embedding_generation",
		Provider:     "embedding-service",
		Duration:     totalDuration,
		Throughput:   throughput,
		Latency:      avgLatency,
		ErrorCount:   errorCount,
		SuccessCount: successCount,
		SuccessRate:  successRate,
		Metadata: map[string]interface{}{
			"text_count": len(texts),
			"iterations": iterations,
		},
	}
}

// RunRAGBenchmark benchmarks RAG engine performance
func (b *AIIntegrationBenchmark) RunRAGBenchmark(
	ctx context.Context,
	ragEngine rag.Engine,
	queries []string,
	iterations int,
) AIIntegrationResult {
	b.logger.Info("Starting RAG engine benchmark",
		"queries", len(queries),
		"iterations", iterations)

	startTime := time.Now()
	var totalLatency time.Duration
	var errorCount int
	var successCount int

	// Warm up
	for i := 0; i < 5; i++ {
		query := &rag.Query{
			Text: queries[0],
		}
		_, err := ragEngine.ProcessQuery(ctx, query)
		if err != nil {
			b.logger.Warn("Warm-up RAG query failed", "error", err)
		}
	}

	// Run benchmark
	for i := 0; i < iterations; i++ {
		query := &rag.Query{
			Text: queries[i%len(queries)],
		}

		iterStart := time.Now()
		_, err := ragEngine.ProcessQuery(ctx, query)
		iterDuration := time.Since(iterStart)

		if err != nil {
			errorCount++
			b.logger.Warn("RAG query failed", "iteration", i, "error", err)
		} else {
			successCount++
			totalLatency += iterDuration
		}
	}

	totalDuration := time.Since(startTime)
	successRate := float64(successCount) / float64(iterations)
	avgLatency := totalLatency / time.Duration(successCount)
	throughput := float64(successCount) / totalDuration.Seconds()

	return AIIntegrationResult{
		Operation:    "rag_query_processing",
		Provider:     "rag-engine",
		Duration:     totalDuration,
		Throughput:   throughput,
		Latency:      avgLatency,
		ErrorCount:   errorCount,
		SuccessCount: successCount,
		SuccessRate:  successRate,
		Metadata: map[string]interface{}{
			"query_count": len(queries),
			"iterations":  iterations,
		},
	}
}

// RunVectorSearchBenchmark benchmarks vector search performance
func (b *AIIntegrationBenchmark) RunVectorSearchBenchmark(
	ctx context.Context,
	vectorIndex index.VectorIndex,
	queryVectors [][]float64,
	k int,
	iterations int,
) AIIntegrationResult {
	b.logger.Info("Starting vector search benchmark",
		"queries", len(queryVectors),
		"k", k,
		"iterations", iterations)

	startTime := time.Now()
	var totalLatency time.Duration
	var errorCount int
	var successCount int

	// Warm up
	for i := 0; i < 5; i++ {
		if i < len(queryVectors) {
			_, err := vectorIndex.Search(queryVectors[i], k)
			if err != nil {
				b.logger.Warn("Warm-up vector search failed", "error", err)
			}
		}
	}

	// Run benchmark
	for i := 0; i < iterations; i++ {
		queryVector := queryVectors[i%len(queryVectors)]

		iterStart := time.Now()
		_, err := vectorIndex.Search(queryVector, k)
		iterDuration := time.Since(iterStart)

		if err != nil {
			errorCount++
			b.logger.Warn("Vector search failed", "iteration", i, "error", err)
		} else {
			successCount++
			totalLatency += iterDuration
		}
	}

	totalDuration := time.Since(startTime)
	successRate := float64(successCount) / float64(iterations)
	avgLatency := totalLatency / time.Duration(successCount)
	throughput := float64(successCount) / totalDuration.Seconds()

	return AIIntegrationResult{
		Operation:    "vector_search",
		Provider:     "vector-index",
		Duration:     totalDuration,
		Throughput:   throughput,
		Latency:      avgLatency,
		ErrorCount:   errorCount,
		SuccessCount: successCount,
		SuccessRate:  successRate,
		Metadata: map[string]interface{}{
			"query_count": len(queryVectors),
			"k":           k,
			"iterations":  iterations,
		},
	}
}

// RunBatchRAGBenchmark benchmarks batch RAG processing
func (b *AIIntegrationBenchmark) RunBatchRAGBenchmark(
	ctx context.Context,
	ragEngine rag.Engine,
	batchQueries [][]string,
	iterations int,
) AIIntegrationResult {
	b.logger.Info("Starting batch RAG benchmark",
		"batches", len(batchQueries),
		"iterations", iterations)

	startTime := time.Now()
	var totalLatency time.Duration
	var errorCount int
	var successCount int

	// Warm up
	for i := 0; i < 5; i++ {
		if i < len(batchQueries) {
			queries := make([]*rag.Query, len(batchQueries[i]))
			for j, text := range batchQueries[i] {
				queries[j] = &rag.Query{Text: text}
			}
			_, err := ragEngine.ProcessBatch(ctx, queries)
			if err != nil {
				b.logger.Warn("Warm-up batch RAG failed", "error", err)
			}
		}
	}

	// Run benchmark
	for i := 0; i < iterations; i++ {
		batchIndex := i % len(batchQueries)
		queries := make([]*rag.Query, len(batchQueries[batchIndex]))
		for j, text := range batchQueries[batchIndex] {
			queries[j] = &rag.Query{Text: text}
		}

		iterStart := time.Now()
		_, err := ragEngine.ProcessBatch(ctx, queries)
		iterDuration := time.Since(iterStart)

		if err != nil {
			errorCount++
			b.logger.Warn("Batch RAG failed", "iteration", i, "error", err)
		} else {
			successCount++
			totalLatency += iterDuration
		}
	}

	totalDuration := time.Since(startTime)
	successRate := float64(successCount) / float64(iterations)
	avgLatency := totalLatency / time.Duration(successCount)
	throughput := float64(successCount) / totalDuration.Seconds()

	return AIIntegrationResult{
		Operation:    "batch_rag_processing",
		Provider:     "rag-engine",
		Duration:     totalDuration,
		Throughput:   throughput,
		Latency:      avgLatency,
		ErrorCount:   errorCount,
		SuccessCount: successCount,
		SuccessRate:  successRate,
		Metadata: map[string]interface{}{
			"batch_count": len(batchQueries),
			"iterations":  iterations,
		},
	}
}

// RunCompleteBenchmarkSuite runs all AI integration benchmarks
func (b *AIIntegrationBenchmark) RunCompleteBenchmarkSuite(
	ctx context.Context,
	embeddingService embedding.Service,
	ragEngine rag.Engine,
	vectorIndex index.VectorIndex,
) *AIIntegrationSuite {
	b.logger.Info("Starting complete AI integration benchmark suite")

	// Prepare test data
	testTexts := []string{
		"machine learning algorithms",
		"artificial intelligence systems",
		"deep learning neural networks",
		"data science and analytics",
		"computer vision and image processing",
		"natural language processing",
		"web development and programming",
		"database management systems",
		"cloud computing infrastructure",
		"distributed systems architecture",
	}

	testQueries := []string{
		"machine learning",
		"artificial intelligence",
		"deep learning",
		"data science",
		"computer vision",
	}

	// Generate test query vectors (mock for now)
	testQueryVectors := make([][]float64, len(testQueries))
	for i := range testQueries {
		testQueryVectors[i] = make([]float64, 384) // 384 dimensions
		for j := range testQueryVectors[i] {
			testQueryVectors[i][j] = float64(i+j) / 1000.0
		}
	}

	batchQueries := [][]string{
		testQueries[:3],
		testQueries[3:],
		testQueries,
	}

	// Run individual benchmarks
	results := []AIIntegrationResult{}

	// Embedding benchmark
	embeddingResult := b.RunEmbeddingBenchmark(ctx, embeddingService, testTexts, 100)
	results = append(results, embeddingResult)

	// RAG benchmark
	ragResult := b.RunRAGBenchmark(ctx, ragEngine, testQueries, 50)
	results = append(results, ragResult)

	// Vector search benchmark
	searchResult := b.RunVectorSearchBenchmark(ctx, vectorIndex, testQueryVectors, 5, 100)
	results = append(results, searchResult)

	// Batch RAG benchmark
	batchRAGResult := b.RunBatchRAGBenchmark(ctx, ragEngine, batchQueries, 30)
	results = append(results, batchRAGResult)

	// Calculate summary
	summary := b.calculateSummary(results)

	suite := &AIIntegrationSuite{
		Name:        "AI Integration Benchmark Suite",
		Description: "Comprehensive benchmarking of AI integration features",
		Results:     results,
		Summary:     summary,
	}

	b.logger.Info("AI integration benchmark suite completed",
		"total_operations", summary.TotalOperations,
		"overall_throughput", fmt.Sprintf("%.2f ops/sec", summary.OverallThroughput),
		"success_rate", fmt.Sprintf("%.2f%%", summary.SuccessRate*100))

	return suite
}

// calculateSummary calculates overall benchmark statistics
func (b *AIIntegrationBenchmark) calculateSummary(results []AIIntegrationResult) AIIntegrationSummary {
	var totalOps int
	var totalDuration time.Duration
	var totalLatency time.Duration
	var totalSuccess int
	var totalErrors int

	for _, result := range results {
		totalOps += result.SuccessCount + result.ErrorCount
		totalDuration += result.Duration
		totalLatency += result.Latency * time.Duration(result.SuccessCount)
		totalSuccess += result.SuccessCount
		totalErrors += result.ErrorCount
	}

	overallThroughput := float64(totalSuccess) / totalDuration.Seconds()
	avgLatency := totalLatency / time.Duration(totalSuccess)
	successRate := float64(totalSuccess) / float64(totalOps)

	return AIIntegrationSummary{
		TotalOperations:   totalOps,
		TotalDuration:     totalDuration,
		OverallThroughput: overallThroughput,
		AverageLatency:    avgLatency,
		SuccessRate:       successRate,
		MemoryEfficiency:  float64(totalSuccess) / 1024.0 / 1024.0, // ops per MB (simplified)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
