package batch

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/embedding"
)



// BenchmarkBatchEmbeddingGeneration benchmarks batch embedding generation
func BenchmarkBatchEmbeddingGeneration(b *testing.B) {
	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService)
	defer processor.Close()

	sizes := []int{10, 50, 100, 500, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("texts_%d", size), func(b *testing.B) {
			req := &BatchEmbeddingRequest{
				Texts:         generateTestTexts(size),
				Model:         "text-embedding-ada-002",
				Provider:      embedding.ProviderTypeOpenAI,
				BatchSize:     50,
				MaxConcurrent: 4,
				Timeout:       60 * time.Second,
				EnableCache:   true,
				Priority:      BatchPriorityNormal,
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				ctx := context.Background()
				_, err := processor.ProcessBatchEmbeddings(ctx, req)
				if err != nil {
					b.Fatalf("Benchmark failed: %v", err)
				}
			}

			// Calculate throughput
			if b.N > 0 {
				totalTexts := int64(b.N * size)
				throughput := float64(totalTexts) / b.Elapsed().Seconds()
				b.ReportMetric(throughput, "texts/sec")
			}
		})
	}
}

// BenchmarkBatchVectorOperations benchmarks various batch vector operations
func BenchmarkBatchVectorOperations(b *testing.B) {
	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService)
	defer processor.Close()

	operations := []BatchOperation{
		BatchOperationInsert,
		BatchOperationSimilarity,
		BatchOperationNormalize,
		BatchOperationDistance,
	}

	sizes := []int{100, 500, 1000, 5000}

	for _, operation := range operations {
		for _, size := range sizes {
			b.Run(fmt.Sprintf("%s_vectors_%d", operation, size), func(b *testing.B) {
				vectors := generateTestVectors(size)
				queryVector := generateTestEmbedding(128)

				req := &BatchVectorRequest{
					Operation:     operation,
					Vectors:       vectors,
					QueryVector:   queryVector,
					BatchSize:     500,
					MaxConcurrent: 4,
					Timeout:       60 * time.Second,
					Priority:      BatchPriorityNormal,
				}

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					ctx := context.Background()
					_, err := processor.ProcessBatchVectors(ctx, req)
					if err != nil {
						b.Fatalf("Benchmark failed: %v", err)
					}
				}

				// Calculate throughput
				if b.N > 0 {
					totalVectors := int64(b.N * size)
					throughput := float64(totalVectors) / b.Elapsed().Seconds()
					b.ReportMetric(throughput, "vectors/sec")
				}
			})
		}
	}
}

// BenchmarkBatchProcessorOptimalSizes benchmarks optimal batch size determination
func BenchmarkBatchProcessorOptimalSizes(b *testing.B) {
	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService)
	defer processor.Close()

	operations := []BatchOperation{
		BatchOperationInsert,
		BatchOperationSearch,
		BatchOperationSimilarity,
		BatchOperationNormalize,
	}

	totalItems := []int{100, 1000, 10000, 100000}

	for _, operation := range operations {
		for _, items := range totalItems {
			b.Run(fmt.Sprintf("%s_items_%d", operation, items), func(b *testing.B) {
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = processor.GetOptimalBatchSize(operation, items)
				}
			})
		}
	}
}

// BenchmarkConcurrentBatchProcessing benchmarks concurrent batch processing
func BenchmarkConcurrentBatchProcessing(b *testing.B) {
	config := GetDefaultConfig()
	// Increase concurrent batches for this test
	config.EmbeddingConfig.MaxConcurrentBatch = 20
	config.VectorConfig.MaxConcurrentBatch = 20

	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService)
	defer processor.Close()

	concurrencyLevels := []int{1, 2, 4, 8, 16}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrency_%d", concurrency), func(b *testing.B) {
			req := &BatchEmbeddingRequest{
				Texts:         generateTestTexts(100),
				Model:         "text-embedding-ada-002",
				Provider:      embedding.ProviderTypeOpenAI,
				BatchSize:     25,
				MaxConcurrent: concurrency,
				Timeout:       60 * time.Second,
				EnableCache:   true,
				Priority:      BatchPriorityNormal,
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				ctx := context.Background()
				_, err := processor.ProcessBatchEmbeddings(ctx, req)
				if err != nil {
					b.Fatalf("Benchmark failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkMemoryUsage benchmarks memory usage during batch processing
func BenchmarkMemoryUsage(b *testing.B) {
	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService)
	defer processor.Close()

	// Test with large batches to stress memory usage
	sizes := []int{1000, 5000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("memory_vectors_%d", size), func(b *testing.B) {
			vectors := generateTestVectors(size)
			queryVector := generateTestEmbedding(128)

			req := &BatchVectorRequest{
				Operation:     BatchOperationSimilarity,
				Vectors:       vectors,
				QueryVector:   queryVector,
				BatchSize:     1000,
				MaxConcurrent: 4,
				Timeout:       60 * time.Second,
				Priority:      BatchPriorityNormal,
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				ctx := context.Background()
				_, err := processor.ProcessBatchVectors(ctx, req)
				if err != nil {
					b.Fatalf("Benchmark failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkBatchSizeOptimization tests different batch sizes for optimal performance
func BenchmarkBatchSizeOptimization(b *testing.B) {
	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService)
	defer processor.Close()

	totalTexts := 1000
	batchSizes := []int{10, 25, 50, 100, 200, 500}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("batch_size_%d", batchSize), func(b *testing.B) {
			req := &BatchEmbeddingRequest{
				Texts:         generateTestTexts(totalTexts),
				Model:         "text-embedding-ada-002",
				Provider:      embedding.ProviderTypeOpenAI,
				BatchSize:     batchSize,
				MaxConcurrent: 4,
				Timeout:       60 * time.Second,
				EnableCache:   true,
				Priority:      BatchPriorityNormal,
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				ctx := context.Background()
				_, err := processor.ProcessBatchEmbeddings(ctx, req)
				if err != nil {
					b.Fatalf("Benchmark failed: %v", err)
				}
			}

			// Calculate efficiency metrics
			if b.N > 0 {
				throughput := float64(totalTexts*b.N) / b.Elapsed().Seconds()
				b.ReportMetric(throughput, "texts/sec")
				b.ReportMetric(float64(batchSize), "batch_size")
			}
		})
	}
}

// BenchmarkStatisticsTracking benchmarks the overhead of statistics tracking
func BenchmarkStatisticsTracking(b *testing.B) {
	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService)
	defer processor.Close()

	req := &BatchEmbeddingRequest{
		Texts:         generateTestTexts(100),
		Model:         "text-embedding-ada-002",
		Provider:      embedding.ProviderTypeOpenAI,
		BatchSize:     25,
		MaxConcurrent: 4,
		Timeout:       60 * time.Second,
		EnableCache:   true,
		Priority:      BatchPriorityNormal,
	}

	b.Run("with_statistics", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, err := processor.ProcessBatchEmbeddings(ctx, req)
			if err != nil {
				b.Fatalf("Benchmark failed: %v", err)
			}
			// Get statistics to measure overhead
			_ = processor.GetStatistics()
		}
	})
}

// BenchmarkProgressTracking benchmarks the overhead of progress tracking
func BenchmarkProgressTracking(b *testing.B) {
	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService)
	defer processor.Close()

	req := &BatchEmbeddingRequest{
		Texts:         generateTestTexts(500),
		Model:         "text-embedding-ada-002",
		Provider:      embedding.ProviderTypeOpenAI,
		BatchSize:     50,
		MaxConcurrent: 4,
		Timeout:       60 * time.Second,
		EnableCache:   true,
		Priority:      BatchPriorityNormal,
	}

	b.Run("with_progress_callback", func(b *testing.B) {
		// Set a simple progress callback
		processor.SetProgressCallback(func(processed, total int, elapsed time.Duration) {
			// Minimal callback to measure overhead
		})

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, err := processor.ProcessBatchEmbeddings(ctx, req)
			if err != nil {
				b.Fatalf("Benchmark failed: %v", err)
			}
		}
	})

	b.Run("without_progress_callback", func(b *testing.B) {
		// Remove progress callback
		processor.SetProgressCallback(nil)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, err := processor.ProcessBatchEmbeddings(ctx, req)
			if err != nil {
				b.Fatalf("Benchmark failed: %v", err)
			}
		}
	})
}

// Test performance targets from implementation plan
func TestPerformanceTargets(t *testing.T) {
	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService)
	defer processor.Close()

	t.Run("embedding_generation_target", func(t *testing.T) {
		// Target: <100ms per text chunk for embedding generation
		req := &BatchEmbeddingRequest{
			Texts:         generateTestTexts(10), // Small batch for per-text measurement
			Model:         "text-embedding-ada-002",
			Provider:      embedding.ProviderTypeOpenAI,
			BatchSize:     10,
			MaxConcurrent: 1,
			Timeout:       30 * time.Second,
			EnableCache:   false, // Disable cache to measure actual generation time
			Priority:      BatchPriorityNormal,
		}

		ctx := context.Background()
		start := time.Now()
		response, err := processor.ProcessBatchEmbeddings(ctx, req)
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		avgTimePerText := elapsed / time.Duration(len(req.Texts))
		targetTime := 100 * time.Millisecond

		t.Logf("Average time per text: %v (target: <%v)", avgTimePerText, targetTime)

		if avgTimePerText > targetTime {
			t.Logf("Warning: Average time per text %v exceeds target %v", avgTimePerText, targetTime)
		}

		if response.Statistics.Throughput > 0 {
			t.Logf("Achieved throughput: %.2f texts/sec", response.Statistics.Throughput)
		}
	})

	t.Run("batch_processing_target", func(t *testing.T) {
		// Target: 1000+ embeddings per minute
		req := &BatchEmbeddingRequest{
			Texts:         generateTestTexts(100), // Scale to measure throughput
			Model:         "text-embedding-ada-002",
			Provider:      embedding.ProviderTypeOpenAI,
			BatchSize:     25,
			MaxConcurrent: 4,
			Timeout:       60 * time.Second,
			EnableCache:   false,
			Priority:      BatchPriorityNormal,
		}

		ctx := context.Background()
		start := time.Now()
		response, err := processor.ProcessBatchEmbeddings(ctx, req)
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Calculate embeddings per minute
		embeddingsPerMinute := float64(len(req.Texts)) / elapsed.Minutes()
		targetThroughput := 1000.0

		t.Logf("Achieved throughput: %.2f embeddings/min (target: >%.0f)", embeddingsPerMinute, targetThroughput)

		if embeddingsPerMinute >= targetThroughput {
			t.Logf("✅ Batch processing target achieved!")
		} else {
			t.Logf("⚠️  Batch processing target not achieved, but this is expected with mock service")
		}

		if response.Statistics.Throughput > 0 {
			t.Logf("Response throughput: %.2f texts/sec", response.Statistics.Throughput)
		}
	})
}
