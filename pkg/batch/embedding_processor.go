package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"log/slog"

	"github.com/vijaynallagatla/vjvector/pkg/embedding"
)

// embeddingProcessor implements BatchEmbeddingProcessor
type embeddingProcessor struct {
	embeddingService embedding.Service
	config           EmbeddingBatchConfig
	capabilities     map[embedding.ProviderType]ProviderCapabilities
	statistics       embeddingStatistics
	logger           *slog.Logger
	mu               sync.RWMutex
}

// embeddingStatistics tracks statistics for embedding batch processing
type embeddingStatistics struct {
	totalBatches        int64
	totalTexts          int64
	totalProcessingTime time.Duration
	totalTokens         int64
	cacheHits           int64
	cacheMisses         int64
	errors              int64
}

// NewBatchEmbeddingProcessor creates a new batch embedding processor
func NewBatchEmbeddingProcessor(embeddingService embedding.Service, config EmbeddingBatchConfig) BatchEmbeddingProcessor {
	if config.DefaultBatchSize <= 0 {
		config.DefaultBatchSize = 100
	}
	if config.MaxBatchSize <= 0 {
		config.MaxBatchSize = 1000
	}
	if config.MaxConcurrentBatch <= 0 {
		config.MaxConcurrentBatch = 10
	}
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 30 * time.Second
	}

	processor := &embeddingProcessor{
		embeddingService: embeddingService,
		config:           config,
		capabilities:     make(map[embedding.ProviderType]ProviderCapabilities),
		logger:           slog.Default(),
	}

	// Initialize provider capabilities
	processor.initializeProviderCapabilities()

	return processor
}

// GenerateBatchEmbeddings generates embeddings for a batch of texts
func (ep *embeddingProcessor) GenerateBatchEmbeddings(ctx context.Context, req *BatchEmbeddingRequest) (*BatchEmbeddingResponse, error) {
	startTime := time.Now()

	// Validate request
	if err := ep.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Apply default configuration
	ep.applyDefaults(req)

	// Initialize response
	response := &BatchEmbeddingResponse{
		Model:    req.Model,
		Provider: req.Provider,
		Statistics: BatchStatistics{
			StartTime:  startTime,
			TotalItems: len(req.Texts),
		},
	}

	// Determine optimal batch processing strategy
	batchSize := ep.GetOptimalBatchSize(req.Provider, len(req.Texts))
	if req.BatchSize > 0 && req.BatchSize < batchSize {
		batchSize = req.BatchSize
	}

	// Process texts in batches
	embeddings, errors, stats, err := ep.processBatches(ctx, req, batchSize)
	if err != nil {
		return nil, fmt.Errorf("batch processing failed: %w", err)
	}

	// Populate response
	response.Embeddings = embeddings
	response.ProcessingTime = time.Since(startTime)
	response.TotalTokens = int(stats.totalTokens)
	response.CacheHits = int(stats.cacheHits)
	response.CacheMisses = int(stats.cacheMisses)
	response.Errors = errors
	response.Statistics.EndTime = time.Now()
	response.Statistics.ProcessedItems = len(embeddings)
	response.Statistics.FailedItems = len(errors)
	response.Statistics.Throughput = float64(len(embeddings)) / response.ProcessingTime.Seconds()
	response.Statistics.AverageLatency = response.ProcessingTime / time.Duration(len(embeddings))

	// Update global statistics
	ep.updateStatistics(response)

	return response, nil
}

// GetOptimalBatchSize returns the optimal batch size for embedding generation
func (ep *embeddingProcessor) GetOptimalBatchSize(provider embedding.ProviderType, totalTexts int) int {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	capabilities, exists := ep.capabilities[provider]
	if !exists {
		return ep.config.DefaultBatchSize
	}

	// Use provider's optimal batch size, but consider total texts
	optimalSize := capabilities.OptimalBatchSize
	if totalTexts < optimalSize {
		return totalTexts
	}

	// For very large batches, use smaller chunks to enable better parallelization
	if totalTexts > 10000 {
		return min(optimalSize, totalTexts/ep.config.MaxConcurrentBatch)
	}

	return optimalSize
}

// EstimateProcessingTime estimates the processing time for a batch
func (ep *embeddingProcessor) EstimateProcessingTime(req *BatchEmbeddingRequest) time.Duration {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	capabilities, exists := ep.capabilities[req.Provider]
	if !exists {
		// Default estimation: 100 texts per second
		return time.Duration(len(req.Texts)) * 10 * time.Millisecond
	}

	// Calculate based on provider capabilities and batch configuration
	batchSize := ep.GetOptimalBatchSize(req.Provider, len(req.Texts))
	numBatches := (len(req.Texts) + batchSize - 1) / batchSize
	concurrentBatches := min(req.MaxConcurrent, ep.config.MaxConcurrentBatch)

	// Estimate time considering parallelization
	sequentialBatches := (numBatches + concurrentBatches - 1) / concurrentBatches
	estimatedTime := time.Duration(sequentialBatches) * capabilities.EstimatedLatency

	return estimatedTime
}

// GetProviderCapabilities returns the capabilities of embedding providers for batch processing
func (ep *embeddingProcessor) GetProviderCapabilities() map[embedding.ProviderType]ProviderCapabilities {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[embedding.ProviderType]ProviderCapabilities)
	for k, v := range ep.capabilities {
		result[k] = v
	}
	return result
}

// processBatches processes texts in optimized batches
func (ep *embeddingProcessor) processBatches(ctx context.Context, req *BatchEmbeddingRequest, batchSize int) ([][]float64, []BatchError, *processStats, error) {
	numTexts := len(req.Texts)
	numBatches := (numTexts + batchSize - 1) / batchSize
	maxConcurrent := min(min(req.MaxConcurrent, ep.config.MaxConcurrentBatch), numBatches)

	// Initialize result containers
	embeddings := make([][]float64, numTexts)
	var errors []BatchError
	var errorsMu sync.Mutex

	// Statistics tracking
	stats := &processStats{}

	// Create worker pool for parallel processing
	type batchJob struct {
		startIdx int
		endIdx   int
		batchIdx int
	}

	jobs := make(chan batchJob, numBatches)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < maxConcurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				ep.processBatch(ctx, req, job.startIdx, job.endIdx, job.batchIdx, embeddings, &errors, &errorsMu, stats)
			}
		}()
	}

	// Submit jobs
	go func() {
		defer close(jobs)
		for i := 0; i < numBatches; i++ {
			startIdx := i * batchSize
			endIdx := min(startIdx+batchSize, numTexts)
			jobs <- batchJob{
				startIdx: startIdx,
				endIdx:   endIdx,
				batchIdx: i,
			}
		}
	}()

	// Wait for completion
	wg.Wait()

	return embeddings, errors, stats, nil
}

// processBatch processes a single batch of texts
func (ep *embeddingProcessor) processBatch(ctx context.Context, req *BatchEmbeddingRequest, startIdx, endIdx, batchIdx int,
	embeddings [][]float64, errors *[]BatchError, errorsMu *sync.Mutex, stats *processStats) {

	batchTexts := req.Texts[startIdx:endIdx]

	// Create embedding request for this batch
	embeddingReq := &embedding.EmbeddingRequest{
		Texts:     batchTexts,
		Model:     req.Model,
		Provider:  req.Provider,
		Options:   req.Options,
		BatchSize: len(batchTexts),
		Timeout:   req.Timeout,
		Metadata:  req.Metadata,
	}

	// Add cache key if caching is enabled
	if req.EnableCache {
		embeddingReq.CacheKey = ep.generateCacheKey(batchTexts, req.Model, req.Provider)
	}

	// Generate embeddings for this batch
	response, err := ep.embeddingService.GenerateEmbeddings(ctx, embeddingReq)
	if err != nil {
		// Record errors for this batch
		errorsMu.Lock()
		for i := 0; i < len(batchTexts); i++ {
			*errors = append(*errors, BatchError{
				Index:   startIdx + i,
				Message: err.Error(),
				Code:    "EMBEDDING_GENERATION_FAILED",
			})
		}
		errorsMu.Unlock()

		ep.logger.Error("Batch embedding generation failed", "batch", batchIdx, "error", err)
		return
	}

	// Store embeddings in the result array
	for i, emb := range response.Embeddings {
		embeddings[startIdx+i] = emb
	}

	// Update statistics
	stats.mu.Lock()
	stats.totalTokens += int64(response.Usage.TotalTokens)
	if response.CacheHit {
		stats.cacheHits++
	} else {
		stats.cacheMisses++
	}
	stats.mu.Unlock()

	ep.logger.Debug("Batch processed successfully",
		"batch", batchIdx,
		"texts", len(batchTexts),
		"processing_time", response.ProcessingTime,
		"cache_hit", response.CacheHit)
}

// processStats tracks statistics during batch processing
type processStats struct {
	totalTokens int64
	cacheHits   int64
	cacheMisses int64
	mu          sync.Mutex
}

// validateRequest validates the batch embedding request
func (ep *embeddingProcessor) validateRequest(req *BatchEmbeddingRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if len(req.Texts) == 0 {
		return fmt.Errorf("texts cannot be empty")
	}
	if len(req.Texts) > ep.config.MaxBatchSize {
		return fmt.Errorf("batch size %d exceeds maximum %d", len(req.Texts), ep.config.MaxBatchSize)
	}
	if req.MaxConcurrent > ep.config.MaxConcurrentBatch {
		return fmt.Errorf("max concurrent %d exceeds limit %d", req.MaxConcurrent, ep.config.MaxConcurrentBatch)
	}
	return nil
}

// applyDefaults applies default configuration to the request
func (ep *embeddingProcessor) applyDefaults(req *BatchEmbeddingRequest) {
	if req.BatchSize <= 0 {
		req.BatchSize = ep.config.DefaultBatchSize
	}
	if req.MaxConcurrent <= 0 {
		req.MaxConcurrent = ep.config.MaxConcurrentBatch
	}
	if req.Timeout <= 0 {
		req.Timeout = ep.config.DefaultTimeout
	}
	if req.Provider == "" {
		req.Provider = embedding.ProviderTypeOpenAI
	}
}

// generateCacheKey generates a cache key for the batch
func (ep *embeddingProcessor) generateCacheKey(texts []string, model string, provider embedding.ProviderType) string {
	// Simple cache key generation - in production, you might want a more sophisticated approach
	return fmt.Sprintf("batch:%s:%s:%d", provider, model, len(texts))
}

// initializeProviderCapabilities initializes capabilities for known providers
func (ep *embeddingProcessor) initializeProviderCapabilities() {
	// OpenAI capabilities
	ep.capabilities[embedding.ProviderTypeOpenAI] = ProviderCapabilities{
		MaxBatchSize:       2048,
		OptimalBatchSize:   100,
		MaxConcurrentBatch: 10,
		EstimatedLatency:   200 * time.Millisecond,
		SupportsCaching:    true,
		SupportsRetry:      true,
		RateLimitRPM:       3000,
		RateLimitTPM:       1000000,
	}

	// Local/Sentence-Transformers capabilities
	ep.capabilities[embedding.ProviderTypeLocal] = ProviderCapabilities{
		MaxBatchSize:       1000,
		OptimalBatchSize:   50,
		MaxConcurrentBatch: 5,
		EstimatedLatency:   100 * time.Millisecond,
		SupportsCaching:    true,
		SupportsRetry:      false,
		RateLimitRPM:       0, // No rate limits for local processing
		RateLimitTPM:       0,
	}

	ep.capabilities[embedding.ProviderTypeSentenceTransformers] = ProviderCapabilities{
		MaxBatchSize:       500,
		OptimalBatchSize:   32,
		MaxConcurrentBatch: 3,
		EstimatedLatency:   150 * time.Millisecond,
		SupportsCaching:    true,
		SupportsRetry:      false,
		RateLimitRPM:       0,
		RateLimitTPM:       0,
	}

	// Custom provider capabilities (conservative defaults)
	ep.capabilities[embedding.ProviderTypeCustom] = ProviderCapabilities{
		MaxBatchSize:       100,
		OptimalBatchSize:   25,
		MaxConcurrentBatch: 2,
		EstimatedLatency:   500 * time.Millisecond,
		SupportsCaching:    false,
		SupportsRetry:      false,
		RateLimitRPM:       100,
		RateLimitTPM:       10000,
	}
}

// updateStatistics updates global statistics after batch processing
func (ep *embeddingProcessor) updateStatistics(response *BatchEmbeddingResponse) {
	ep.statistics.totalBatches++
	ep.statistics.totalTexts += int64(response.Statistics.TotalItems)
	ep.statistics.totalProcessingTime += response.ProcessingTime
	ep.statistics.totalTokens += int64(response.TotalTokens)
	ep.statistics.cacheHits += int64(response.CacheHits)
	ep.statistics.cacheMisses += int64(response.CacheMisses)
	ep.statistics.errors += int64(len(response.Errors))
}

// GetStatistics returns current embedding processor statistics
func (ep *embeddingProcessor) GetStatistics() embeddingStatistics {
	// Return a copy to avoid lock copying
	stats := embeddingStatistics{
		totalBatches:        ep.statistics.totalBatches,
		totalTexts:          ep.statistics.totalTexts,
		totalProcessingTime: ep.statistics.totalProcessingTime,
		totalTokens:         ep.statistics.totalTokens,
		cacheHits:           ep.statistics.cacheHits,
		cacheMisses:         ep.statistics.cacheMisses,
		errors:              ep.statistics.errors,
	}
	return stats
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
