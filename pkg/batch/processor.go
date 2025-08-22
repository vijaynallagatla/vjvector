package batch

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"log/slog"

	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/embedding"
	"github.com/vijaynallagatla/vjvector/pkg/rag"
)

// processor implements BatchProcessor
type processor struct {
	config             BatchConfig
	embeddingProcessor BatchEmbeddingProcessor
	vectorProcessor    BatchVectorProcessor
	ragEngine          interface {
		ProcessBatch(context.Context, []*rag.Query) ([]*rag.QueryResponse, error)
	}
	statistics       processorStatistics
	progressCallback BatchProgressCallback
	logger           *slog.Logger
	mu               sync.RWMutex
}

// processorStatistics tracks overall batch processor statistics
type processorStatistics struct {
	totalBatches        int64
	totalItems          int64
	totalProcessingTime time.Duration
	averageThroughput   float64
	averageLatency      time.Duration
	successRate         float64
	cacheHitRate        float64
	memoryUsage         int64
	activeBatches       int
	mu                  sync.Mutex
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(config BatchConfig, embeddingService embedding.Service, ragEngine rag.Engine) BatchProcessor {
	// Apply default configurations
	if config.EmbeddingConfig.DefaultBatchSize <= 0 {
		config.EmbeddingConfig.DefaultBatchSize = 100
	}
	if config.VectorConfig.DefaultBatchSize <= 0 {
		config.VectorConfig.DefaultBatchSize = 1000
	}
	if config.VectorConfig.WorkerCount <= 0 {
		config.VectorConfig.WorkerCount = runtime.NumCPU()
	}

	processor := &processor{
		config:             config,
		embeddingProcessor: NewBatchEmbeddingProcessor(embeddingService, config.EmbeddingConfig),
		vectorProcessor:    NewBatchVectorProcessor(config.VectorConfig),
		ragEngine:          ragEngine,
		logger:             slog.Default(),
	}

	return processor
}

// ProcessBatchEmbeddings processes a batch of texts for embedding generation
func (p *processor) ProcessBatchEmbeddings(ctx context.Context, req *BatchEmbeddingRequest) (*BatchEmbeddingResponse, error) {
	p.incrementActiveBatches()
	defer p.decrementActiveBatches()

	startTime := time.Now()

	// Set up progress tracking if callback is configured
	if p.progressCallback != nil {
		// Create a context for progress updates
		progressCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		go p.trackEmbeddingProgress(progressCtx, req, startTime)
	}

	// Process the batch
	response, err := p.embeddingProcessor.GenerateBatchEmbeddings(ctx, req)
	if err != nil {
		return nil, err
	}

	// Update global statistics
	p.updateEmbeddingStatistics(response)

	p.logger.Info("Batch embedding processing completed",
		"texts", len(req.Texts),
		"processing_time", response.ProcessingTime,
		"throughput", response.Statistics.Throughput,
		"cache_hits", response.CacheHits,
		"errors", len(response.Errors))

	return response, nil
}

// ProcessBatchVectors processes a batch of vector operations
func (p *processor) ProcessBatchVectors(ctx context.Context, req *BatchVectorRequest) (*BatchVectorResponse, error) {
	p.incrementActiveBatches()
	defer p.decrementActiveBatches()

	startTime := time.Now()

	// Set up progress tracking if callback is configured
	if p.progressCallback != nil {
		// Create a context for progress updates
		progressCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		go p.trackVectorProgress(progressCtx, req, startTime)
	}

	// Process the batch
	response, err := p.vectorProcessor.ProcessVectorBatch(ctx, req)
	if err != nil {
		return nil, err
	}

	// Update global statistics
	p.updateVectorStatistics(response)

	p.logger.Info("Batch vector processing completed",
		"operation", req.Operation,
		"vectors", len(req.Vectors),
		"processing_time", response.ProcessingTime,
		"throughput", response.Statistics.Throughput,
		"processed", response.ProcessedCount,
		"errors", response.ErrorCount)

	return response, nil
}

// ProcessBatchRAG processes a batch of RAG operations
func (p *processor) ProcessBatchRAG(ctx context.Context, req *BatchRAGRequest) (*BatchRAGResponse, error) {
	p.incrementActiveBatches()
	defer p.decrementActiveBatches()

	startTime := time.Now()

	// Set up progress tracking if callback is configured
	if p.progressCallback != nil {
		// Create a context for progress updates
		progressCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		go p.trackRAGProgress(progressCtx, req, startTime)
	}

	// Convert batch request to RAG engine queries
	ragQueries := make([]*rag.Query, len(req.Queries))
	for i, query := range req.Queries {
		ragQueries[i] = &rag.Query{
			Text:    query,
			Context: req.Context,
		}
	}

	// Process queries using RAG engine
	responses, err := p.ragEngine.ProcessBatch(ctx, ragQueries)
	if err != nil {
		return nil, fmt.Errorf("RAG batch processing failed: %w", err)
	}

	// Debug logging
	p.logger.Info("RAG batch responses received",
		"count", len(responses),
		"first_response_results", len(responses[0].Results))

	processingTime := time.Since(startTime)

	// Convert RAG responses to batch responses
	batchResults := make([]RAGQueryResult, len(responses))
	for i, response := range responses {
		// Convert RAG results to SearchResult format
		searchResults := make([]SearchResult, len(response.Results))
		for j, ragResult := range response.Results {
			searchResults[j] = SearchResult{
				Vector: &core.Vector{
					ID:         ragResult.Vector.ID,
					Collection: ragResult.Vector.Collection,
					Embedding:  ragResult.Vector.Embedding,
					Metadata:   ragResult.Vector.Metadata,
				},
				Score:      ragResult.Score,
				Rank:       j + 1,
				Similarity: ragResult.Score,
				Context:    response.Query.Context,
				Metadata:   ragResult.Vector.Metadata,
			}
		}

		batchResults[i] = RAGQueryResult{
			Query:               response.Query.Text,
			OriginalQuery:       response.Query.Text,
			ExpandedQueries:     response.QueryExpansion,
			Results:             searchResults,
			RerankedResults:     searchResults, // For now, use same results
			ContextEnhancements: []string{},    // Will be populated when context enhancement is implemented
			ProcessingTime:      response.ProcessingTime,
			Confidence:          0.85, // Extract from metadata if available
			Metadata:            response.Metadata,
		}
	}

	// Create batch response
	batchResponse := &BatchRAGResponse{
		Operation:      req.Operation,
		Results:        batchResults,
		ProcessingTime: processingTime,
		ProcessedCount: len(responses),
		ErrorCount:     0,
		Errors:         []BatchError{},
		Statistics: BatchStatistics{
			StartTime:      startTime,
			EndTime:        time.Now(),
			TotalItems:     len(req.Queries),
			ProcessedItems: len(responses),
			FailedItems:    0,
			Throughput:     float64(len(responses)) / processingTime.Seconds(),
			AverageLatency: processingTime / time.Duration(len(responses)),
		},
		RAGMetrics: RAGBatchMetrics{
			QueryExpansionCount:     len(responses), // Count queries that had expansion
			RerankingCount:          len(responses), // Count queries that had reranking
			ContextEnhancementCount: len(responses), // Count queries that had context enhancement
			AverageExpansionRatio:   1.0,            // Mock value
			AverageRerankingTime:    processingTime / time.Duration(len(responses)),
			CacheHitRate:            0.2,  // Mock value
			AccuracyImprovement:     0.15, // Mock value
		},
	}

	p.logger.Info("Batch RAG processing completed",
		"operation", req.Operation,
		"queries", len(req.Queries),
		"processing_time", processingTime)

	return batchResponse, nil
}

// GetOptimalBatchSize returns the optimal batch size for the given operation
func (p *processor) GetOptimalBatchSize(operation interface{}, totalItems int) int {
	// Try to handle as BatchOperation first
	if batchOp, ok := operation.(BatchOperation); ok {
		switch batchOp {
		case BatchOperationInsert, BatchOperationUpdate, BatchOperationDelete,
			BatchOperationSearch, BatchOperationSimilarity, BatchOperationNormalize, BatchOperationDistance:
			return p.vectorProcessor.GetOptimalBatchSize(batchOp, totalItems)
		}
	}

	// Try to handle as BatchRAGOperation
	if _, ok := operation.(BatchRAGOperation); ok {
		// For RAG operations, we'll need to implement RAG processor
		// For now, return a reasonable default
		return 50
	}

	// Assume it's an embedding operation if not a vector operation
	// This could be enhanced to detect embedding operations more precisely
	return p.embeddingProcessor.GetOptimalBatchSize(embedding.ProviderTypeOpenAI, totalItems)
}

// GetStatistics returns current batch processing statistics
func (p *processor) GetStatistics() BatchProcessorStatistics {
	p.statistics.mu.Lock()
	defer p.statistics.mu.Unlock()

	// Calculate current memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return BatchProcessorStatistics{
		TotalBatches:        p.statistics.totalBatches,
		TotalItems:          p.statistics.totalItems,
		TotalProcessingTime: p.statistics.totalProcessingTime,
		AverageThroughput:   p.statistics.averageThroughput,
		AverageLatency:      p.statistics.averageLatency,
		SuccessRate:         p.statistics.successRate,
		CacheHitRate:        p.statistics.cacheHitRate,
		MemoryUsage:         int64(memStats.Alloc),
		ActiveBatches:       p.statistics.activeBatches,
	}
}

// SetProgressCallback sets a callback for progress updates
func (p *processor) SetProgressCallback(callback BatchProgressCallback) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.progressCallback = callback
}

// Close closes the batch processor and cleans up resources
func (p *processor) Close() error {
	p.logger.Info("Closing batch processor")

	// Wait for active batches to complete
	for {
		stats := p.GetStatistics()
		if stats.ActiveBatches == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	p.logger.Info("Batch processor closed successfully")
	return nil
}

// trackEmbeddingProgress tracks progress for embedding operations
func (p *processor) trackEmbeddingProgress(ctx context.Context, req *BatchEmbeddingRequest, startTime time.Time) {
	if p.progressCallback == nil {
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	totalTexts := len(req.Texts)
	_ = p.embeddingProcessor.GetOptimalBatchSize(req.Provider, totalTexts)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			// Estimate progress based on elapsed time and expected processing time
			estimatedTotal := p.embeddingProcessor.EstimateProcessingTime(req)
			progress := int(float64(totalTexts) * elapsed.Seconds() / estimatedTotal.Seconds())
			if progress > totalTexts {
				progress = totalTexts
			}

			p.progressCallback(progress, totalTexts, elapsed)
		}
	}
}

// trackVectorProgress tracks progress for vector operations
func (p *processor) trackVectorProgress(ctx context.Context, req *BatchVectorRequest, startTime time.Time) {
	if p.progressCallback == nil {
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	totalVectors := len(req.Vectors)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			// Estimate progress based on elapsed time and expected processing time
			estimatedTotal := p.vectorProcessor.EstimateProcessingTime(req)
			progress := int(float64(totalVectors) * elapsed.Seconds() / estimatedTotal.Seconds())
			if progress > totalVectors {
				progress = totalVectors
			}

			p.progressCallback(progress, totalVectors, elapsed)
		}
	}
}

// trackRAGProgress tracks progress for RAG operations
func (p *processor) trackRAGProgress(ctx context.Context, req *BatchRAGRequest, startTime time.Time) {
	if p.progressCallback == nil {
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	totalQueries := len(req.Queries)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			// For RAG operations, we'll use a simple time-based progress estimate
			// In a real implementation, this would be more sophisticated
			estimatedTotal := 30 * time.Second // Default estimate for RAG operations
			progress := int(float64(totalQueries) * elapsed.Seconds() / estimatedTotal.Seconds())
			if progress > totalQueries {
				progress = totalQueries
			}

			p.progressCallback(progress, totalQueries, elapsed)
		}
	}
}

// incrementActiveBatches increments the active batch counter
func (p *processor) incrementActiveBatches() {
	p.statistics.mu.Lock()
	defer p.statistics.mu.Unlock()
	p.statistics.activeBatches++
}

// decrementActiveBatches decrements the active batch counter
func (p *processor) decrementActiveBatches() {
	p.statistics.mu.Lock()
	defer p.statistics.mu.Unlock()
	p.statistics.activeBatches--
}

// updateEmbeddingStatistics updates statistics after embedding processing
func (p *processor) updateEmbeddingStatistics(response *BatchEmbeddingResponse) {
	p.statistics.mu.Lock()
	defer p.statistics.mu.Unlock()

	p.statistics.totalBatches++
	p.statistics.totalItems += int64(response.Statistics.TotalItems)
	p.statistics.totalProcessingTime += response.ProcessingTime

	// Update averages
	if p.statistics.totalBatches > 0 {
		p.statistics.averageThroughput = float64(p.statistics.totalItems) / p.statistics.totalProcessingTime.Seconds()
		p.statistics.averageLatency = p.statistics.totalProcessingTime / time.Duration(p.statistics.totalItems)
	}

	// Update success rate
	successfulItems := int64(response.Statistics.ProcessedItems)
	p.statistics.successRate = float64(successfulItems) / float64(response.Statistics.TotalItems)

	// Update cache hit rate
	totalCacheRequests := int64(response.CacheHits + response.CacheMisses)
	if totalCacheRequests > 0 {
		p.statistics.cacheHitRate = float64(response.CacheHits) / float64(totalCacheRequests)
	}
}

// updateVectorStatistics updates statistics after vector processing
func (p *processor) updateVectorStatistics(response *BatchVectorResponse) {
	p.statistics.mu.Lock()
	defer p.statistics.mu.Unlock()

	p.statistics.totalBatches++
	p.statistics.totalItems += int64(response.Statistics.TotalItems)
	p.statistics.totalProcessingTime += response.ProcessingTime

	// Update averages
	if p.statistics.totalBatches > 0 {
		p.statistics.averageThroughput = float64(p.statistics.totalItems) / p.statistics.totalProcessingTime.Seconds()
		p.statistics.averageLatency = p.statistics.totalProcessingTime / time.Duration(p.statistics.totalItems)
	}

	// Update success rate
	successfulItems := int64(response.ProcessedCount)
	p.statistics.successRate = float64(successfulItems) / float64(response.Statistics.TotalItems)
}

// GetEmbeddingProcessorCapabilities returns embedding processor capabilities
func (p *processor) GetEmbeddingProcessorCapabilities() map[embedding.ProviderType]ProviderCapabilities {
	return p.embeddingProcessor.GetProviderCapabilities()
}

// GetVectorProcessorCapabilities returns vector processor capabilities
func (p *processor) GetVectorProcessorCapabilities() map[BatchOperation]OperationCapabilities {
	return p.vectorProcessor.GetOperationCapabilities()
}

// EstimateEmbeddingProcessingTime estimates embedding processing time
func (p *processor) EstimateEmbeddingProcessingTime(req *BatchEmbeddingRequest) time.Duration {
	return p.embeddingProcessor.EstimateProcessingTime(req)
}

// EstimateVectorProcessingTime estimates vector processing time
func (p *processor) EstimateVectorProcessingTime(req *BatchVectorRequest) time.Duration {
	return p.vectorProcessor.EstimateProcessingTime(req)
}

// GetDefaultConfig returns a default batch configuration
func GetDefaultConfig() BatchConfig {
	return BatchConfig{
		EmbeddingConfig: EmbeddingBatchConfig{
			DefaultBatchSize:   100,
			MaxBatchSize:       1000,
			MaxConcurrentBatch: 10,
			DefaultTimeout:     30 * time.Second,
			EnableCache:        true,
			EnableRetry:        true,
			RetryAttempts:      3,
			ProviderSettings: map[embedding.ProviderType]ProviderBatchConfig{
				embedding.ProviderTypeOpenAI: {
					BatchSize:          100,
					MaxConcurrentBatch: 10,
					Timeout:            30 * time.Second,
					RetryAttempts:      3,
					RateLimitRPM:       3000,
					RateLimitTPM:       1000000,
				},
				embedding.ProviderTypeLocal: {
					BatchSize:          50,
					MaxConcurrentBatch: 5,
					Timeout:            20 * time.Second,
					RetryAttempts:      1,
					RateLimitRPM:       0,
					RateLimitTPM:       0,
				},
			},
		},
		VectorConfig: VectorBatchConfig{
			DefaultBatchSize:   1000,
			MaxBatchSize:       10000,
			MaxConcurrentBatch: runtime.NumCPU(),
			DefaultTimeout:     60 * time.Second,
			EnableSIMD:         true,
			EnableParallel:     true,
			WorkerCount:        runtime.NumCPU(),
			OperationSettings: map[BatchOperation]OperationBatchConfig{
				BatchOperationInsert: {
					BatchSize:          1000,
					MaxConcurrentBatch: runtime.NumCPU(),
					Timeout:            30 * time.Second,
					WorkerCount:        runtime.NumCPU(),
					MemoryLimit:        1024 * 1024 * 1024, // 1GB
				},
				BatchOperationSearch: {
					BatchSize:          5000,
					MaxConcurrentBatch: runtime.NumCPU(),
					Timeout:            60 * time.Second,
					WorkerCount:        runtime.NumCPU(),
					MemoryLimit:        2048 * 1024 * 1024, // 2GB
				},
			},
		},
		PerformanceConfig: PerformanceBatchConfig{
			EnableMemoryPool:   true,
			MemoryPoolSize:     1024 * 1024 * 1024, // 1GB
			EnableProfiling:    false,
			ProfilingInterval:  10 * time.Second,
			GCOptimization:     true,
			CPUAffinityEnabled: false,
		},
		MonitoringConfig: MonitoringBatchConfig{
			EnableMetrics:       true,
			MetricsInterval:     30 * time.Second,
			EnableProgressLogs:  true,
			LogInterval:         10 * time.Second,
			EnableHealthCheck:   true,
			HealthCheckInterval: 60 * time.Second,
		},
	}
}
