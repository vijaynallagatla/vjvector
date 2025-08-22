package batch

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"log/slog"

	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/parallel"
)

// vectorProcessor implements BatchVectorProcessor
type vectorProcessor struct {
	config         VectorBatchConfig
	capabilities   map[BatchOperation]OperationCapabilities
	statistics     vectorStatistics
	workerPool     *parallel.WorkerPool
	batchProcessor *parallel.ParallelBatchProcessor
	logger         *slog.Logger
	mu             sync.RWMutex
}

// vectorStatistics tracks statistics for vector batch processing
type vectorStatistics struct {
	totalBatches        int64
	totalVectors        int64
	totalProcessingTime time.Duration
	totalOperations     map[BatchOperation]int64
	errors              int64
	mu                  sync.Mutex
}

// NewBatchVectorProcessor creates a new batch vector processor
func NewBatchVectorProcessor(config VectorBatchConfig) BatchVectorProcessor {
	if config.DefaultBatchSize <= 0 {
		config.DefaultBatchSize = 1000
	}
	if config.MaxBatchSize <= 0 {
		config.MaxBatchSize = 10000
	}
	if config.MaxConcurrentBatch <= 0 {
		config.MaxConcurrentBatch = runtime.NumCPU()
	}
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 60 * time.Second
	}
	if config.WorkerCount <= 0 {
		config.WorkerCount = runtime.NumCPU()
	}

	processor := &vectorProcessor{
		config:         config,
		capabilities:   make(map[BatchOperation]OperationCapabilities),
		statistics:     vectorStatistics{totalOperations: make(map[BatchOperation]int64)},
		workerPool:     parallel.NewWorkerPool(config.WorkerCount, config.EnableSIMD),
		batchProcessor: parallel.NewParallelBatchProcessor(config.WorkerCount, config.EnableSIMD),
		logger:         slog.Default(),
	}

	// Initialize operation capabilities
	processor.initializeOperationCapabilities()

	return processor
}

// ProcessVectorBatch processes a batch of vector operations
func (vp *vectorProcessor) ProcessVectorBatch(ctx context.Context, req *BatchVectorRequest) (*BatchVectorResponse, error) {
	startTime := time.Now()

	// Validate request
	if err := vp.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Apply default configuration
	vp.applyDefaults(req)

	// Initialize response
	response := &BatchVectorResponse{
		Operation: req.Operation,
		Statistics: BatchStatistics{
			StartTime:  startTime,
			TotalItems: len(req.Vectors),
		},
	}

	// Process based on operation type
	results, errors, err := vp.processOperation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("operation processing failed: %w", err)
	}

	// Populate response
	response.Results = results
	response.ProcessingTime = time.Since(startTime)
	response.ProcessedCount = len(req.Vectors) - len(errors)
	response.ErrorCount = len(errors)
	response.Errors = errors
	response.Statistics.EndTime = time.Now()
	response.Statistics.ProcessedItems = response.ProcessedCount
	response.Statistics.FailedItems = response.ErrorCount
	if response.ProcessingTime.Seconds() > 0 {
		response.Statistics.Throughput = float64(response.ProcessedCount) / response.ProcessingTime.Seconds()
	}
	if response.ProcessedCount > 0 {
		response.Statistics.AverageLatency = response.ProcessingTime / time.Duration(response.ProcessedCount)
	}

	// Update global statistics
	vp.updateStatistics(req.Operation, response)

	return response, nil
}

// GetOptimalBatchSize returns the optimal batch size for vector operations
func (vp *vectorProcessor) GetOptimalBatchSize(operation BatchOperation, totalVectors int) int {
	vp.mu.RLock()
	defer vp.mu.RUnlock()

	capabilities, exists := vp.capabilities[operation]
	if !exists {
		return vp.config.DefaultBatchSize
	}

	// Use operation's optimal batch size, but consider total vectors
	optimalSize := capabilities.OptimalBatchSize
	if totalVectors < optimalSize {
		return totalVectors
	}

	// For very large batches, use smaller chunks to enable better parallelization
	if totalVectors > 50000 {
		return min(optimalSize, totalVectors/vp.config.MaxConcurrentBatch)
	}

	return optimalSize
}

// EstimateProcessingTime estimates the processing time for a batch vector operation
func (vp *vectorProcessor) EstimateProcessingTime(req *BatchVectorRequest) time.Duration {
	vp.mu.RLock()
	defer vp.mu.RUnlock()

	capabilities, exists := vp.capabilities[req.Operation]
	if !exists {
		// Default estimation based on operation complexity
		return vp.getDefaultEstimate(req.Operation, len(req.Vectors))
	}

	// Calculate based on operation capabilities and batch configuration
	batchSize := vp.GetOptimalBatchSize(req.Operation, len(req.Vectors))
	numBatches := (len(req.Vectors) + batchSize - 1) / batchSize
	concurrentBatches := min(req.MaxConcurrent, vp.config.MaxConcurrentBatch)

	// Estimate time considering parallelization
	sequentialBatches := (numBatches + concurrentBatches - 1) / concurrentBatches
	estimatedTime := time.Duration(sequentialBatches) * capabilities.EstimatedLatency

	return estimatedTime
}

// GetOperationCapabilities returns the capabilities for different vector operations
func (vp *vectorProcessor) GetOperationCapabilities() map[BatchOperation]OperationCapabilities {
	vp.mu.RLock()
	defer vp.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[BatchOperation]OperationCapabilities)
	for k, v := range vp.capabilities {
		result[k] = v
	}
	return result
}

// processOperation processes the batch operation based on its type
func (vp *vectorProcessor) processOperation(ctx context.Context, req *BatchVectorRequest) (interface{}, []BatchError, error) {
	switch req.Operation {
	case BatchOperationInsert:
		return vp.processInsertBatch(ctx, req)
	case BatchOperationUpdate:
		return vp.processUpdateBatch(ctx, req)
	case BatchOperationDelete:
		return vp.processDeleteBatch(ctx, req)
	case BatchOperationSearch:
		return vp.processSearchBatch(ctx, req)
	case BatchOperationSimilarity:
		return vp.processSimilarityBatch(ctx, req)
	case BatchOperationNormalize:
		return vp.processNormalizeBatch(ctx, req)
	case BatchOperationDistance:
		return vp.processDistanceBatch(ctx, req)
	default:
		return nil, nil, fmt.Errorf("unsupported operation: %s", req.Operation)
	}
}

// processInsertBatch processes a batch of vector insertions
func (vp *vectorProcessor) processInsertBatch(ctx context.Context, req *BatchVectorRequest) (interface{}, []BatchError, error) {
	results := make([]string, len(req.Vectors))
	var errors []BatchError

	batchSize := vp.GetOptimalBatchSize(req.Operation, len(req.Vectors))
	numBatches := (len(req.Vectors) + batchSize - 1) / batchSize

	var wg sync.WaitGroup
	var errorsMu sync.Mutex

	// Process in parallel batches
	for i := 0; i < numBatches; i++ {
		startIdx := i * batchSize
		endIdx := min(startIdx+batchSize, len(req.Vectors))
		
		wg.Add(1)
		go func(batchIdx, start, end int) {
			defer wg.Done()
			
			for j := start; j < end; j++ {
				// Simulate vector insertion (in real implementation, this would interact with storage)
				vector := req.Vectors[j]
				if vector == nil {
					errorsMu.Lock()
					errors = append(errors, BatchError{
						Index:   j,
						Message: "nil vector",
						Code:    "NIL_VECTOR",
					})
					errorsMu.Unlock()
					continue
				}
				
				// Validate vector
				if len(vector.Embedding) == 0 {
					errorsMu.Lock()
					errors = append(errors, BatchError{
						Index:   j,
						Message: "empty embedding",
						Code:    "EMPTY_EMBEDDING",
					})
					errorsMu.Unlock()
					continue
				}
				
				results[j] = vector.ID
			}
		}(i, startIdx, endIdx)
	}

	wg.Wait()
	return results, errors, nil
}

// processUpdateBatch processes a batch of vector updates
func (vp *vectorProcessor) processUpdateBatch(ctx context.Context, req *BatchVectorRequest) (interface{}, []BatchError, error) {
	results := make([]bool, len(req.Vectors))
	var errors []BatchError

	// Similar pattern to insert but for updates
	for i, vector := range req.Vectors {
		if vector == nil {
			errors = append(errors, BatchError{
				Index:   i,
				Message: "nil vector",
				Code:    "NIL_VECTOR",
			})
			continue
		}
		
		// Simulate update operation
		results[i] = true
	}

	return results, errors, nil
}

// processDeleteBatch processes a batch of vector deletions
func (vp *vectorProcessor) processDeleteBatch(ctx context.Context, req *BatchVectorRequest) (interface{}, []BatchError, error) {
	results := make([]bool, len(req.Vectors))
	var errors []BatchError

	// Process deletions
	for i, vector := range req.Vectors {
		if vector == nil || vector.ID == "" {
			errors = append(errors, BatchError{
				Index:   i,
				Message: "invalid vector ID",
				Code:    "INVALID_VECTOR_ID",
			})
			continue
		}
		
		// Simulate delete operation
		results[i] = true
	}

	return results, errors, nil
}

// processSearchBatch processes a batch of vector searches
func (vp *vectorProcessor) processSearchBatch(ctx context.Context, req *BatchVectorRequest) (interface{}, []BatchError, error) {
	if len(req.QueryVector) == 0 {
		return nil, nil, fmt.Errorf("query vector is required for search operation")
	}

	// Extract embeddings for parallel processing
	embeddings := make([][]float64, len(req.Vectors))
	for i, vector := range req.Vectors {
		if vector != nil {
			embeddings[i] = vector.Embedding
		}
	}

	// Use parallel vector search
	k := 10 // Default top-k
	if options, ok := req.Options["k"].(int); ok {
		k = options
	}

	searchResults := vp.batchProcessor.ProcessSearchBatch(req.QueryVector, embeddings, k)
	
	// Convert to our result format
	results := make([]core.VectorSearchResult, len(searchResults))
	for i, result := range searchResults {
		if result.Index < len(req.Vectors) && req.Vectors[result.Index] != nil {
			results[i] = core.VectorSearchResult{
				Vector:   req.Vectors[result.Index],
				Score:    result.Similarity,
				Distance: 1.0 - result.Similarity, // Convert similarity to distance
			}
		}
	}

	return results, []BatchError{}, nil
}

// processSimilarityBatch processes a batch of similarity calculations
func (vp *vectorProcessor) processSimilarityBatch(ctx context.Context, req *BatchVectorRequest) (interface{}, []BatchError, error) {
	if len(req.QueryVector) == 0 {
		return nil, nil, fmt.Errorf("query vector is required for similarity operation")
	}

	// Extract embeddings
	embeddings := make([][]float64, len(req.Vectors))
	var errors []BatchError

	for i, vector := range req.Vectors {
		if vector == nil || len(vector.Embedding) == 0 {
			errors = append(errors, BatchError{
				Index:   i,
				Message: "invalid vector embedding",
				Code:    "INVALID_EMBEDDING",
			})
			continue
		}
		embeddings[i] = vector.Embedding
	}

	// Calculate similarities in parallel
	queryVectors := make([][]float64, len(embeddings))
	for i := range queryVectors {
		queryVectors[i] = req.QueryVector
	}

	similarities := vp.workerPool.ParallelCosineSimilarity(queryVectors, embeddings)
	
	return similarities, errors, nil
}

// processNormalizeBatch processes a batch of vector normalizations
func (vp *vectorProcessor) processNormalizeBatch(ctx context.Context, req *BatchVectorRequest) (interface{}, []BatchError, error) {
	// Extract embeddings
	embeddings := make([][]float64, len(req.Vectors))
	var errors []BatchError

	for i, vector := range req.Vectors {
		if vector == nil || len(vector.Embedding) == 0 {
			errors = append(errors, BatchError{
				Index:   i,
				Message: "invalid vector embedding",
				Code:    "INVALID_EMBEDDING",
			})
			continue
		}
		embeddings[i] = vector.Embedding
	}

	// Normalize in parallel
	normalizedEmbeddings := vp.batchProcessor.ProcessNormalizeBatch(embeddings)
	
	// Update original vectors
	results := make([]*core.Vector, len(req.Vectors))
	for i, vector := range req.Vectors {
		if vector != nil && i < len(normalizedEmbeddings) {
			// Create a copy and update
			updated := *vector
			updated.Embedding = normalizedEmbeddings[i]
			updated.Normalized = true
			updated.Magnitude = 1.0
			updated.UpdatedAt = time.Now()
			results[i] = &updated
		}
	}

	return results, errors, nil
}

// processDistanceBatch processes a batch of distance calculations
func (vp *vectorProcessor) processDistanceBatch(ctx context.Context, req *BatchVectorRequest) (interface{}, []BatchError, error) {
	if len(req.QueryVector) == 0 {
		return nil, nil, fmt.Errorf("query vector is required for distance operation")
	}

	// Extract embeddings
	embeddings := make([][]float64, len(req.Vectors))
	var errors []BatchError

	for i, vector := range req.Vectors {
		if vector == nil || len(vector.Embedding) == 0 {
			errors = append(errors, BatchError{
				Index:   i,
				Message: "invalid vector embedding",
				Code:    "INVALID_EMBEDDING",
			})
			continue
		}
		embeddings[i] = vector.Embedding
	}

	// Calculate distances in parallel
	queryVectors := make([][]float64, len(embeddings))
	for i := range queryVectors {
		queryVectors[i] = req.QueryVector
	}

	distances := vp.workerPool.ParallelEuclideanDistance(queryVectors, embeddings)
	
	return distances, errors, nil
}

// validateRequest validates the batch vector request
func (vp *vectorProcessor) validateRequest(req *BatchVectorRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if len(req.Vectors) == 0 {
		return fmt.Errorf("vectors cannot be empty")
	}
	if len(req.Vectors) > vp.config.MaxBatchSize {
		return fmt.Errorf("batch size %d exceeds maximum %d", len(req.Vectors), vp.config.MaxBatchSize)
	}
	if req.MaxConcurrent > vp.config.MaxConcurrentBatch {
		return fmt.Errorf("max concurrent %d exceeds limit %d", req.MaxConcurrent, vp.config.MaxConcurrentBatch)
	}
	return nil
}

// applyDefaults applies default configuration to the request
func (vp *vectorProcessor) applyDefaults(req *BatchVectorRequest) {
	if req.BatchSize <= 0 {
		req.BatchSize = vp.config.DefaultBatchSize
	}
	if req.MaxConcurrent <= 0 {
		req.MaxConcurrent = vp.config.MaxConcurrentBatch
	}
	if req.Timeout <= 0 {
		req.Timeout = vp.config.DefaultTimeout
	}
}

// initializeOperationCapabilities initializes capabilities for known operations
func (vp *vectorProcessor) initializeOperationCapabilities() {
	// Insert operation capabilities
	vp.capabilities[BatchOperationInsert] = OperationCapabilities{
		MaxBatchSize:       10000,
		OptimalBatchSize:   1000,
		MaxConcurrentBatch: vp.config.WorkerCount,
		EstimatedLatency:   10 * time.Microsecond,
		SupportsParallel:   true,
		SupportsSIMD:       false,
		MemoryRequirement:  1024, // bytes per vector
	}

	// Update operation capabilities
	vp.capabilities[BatchOperationUpdate] = OperationCapabilities{
		MaxBatchSize:       10000,
		OptimalBatchSize:   1000,
		MaxConcurrentBatch: vp.config.WorkerCount,
		EstimatedLatency:   15 * time.Microsecond,
		SupportsParallel:   true,
		SupportsSIMD:       false,
		MemoryRequirement:  1024,
	}

	// Delete operation capabilities
	vp.capabilities[BatchOperationDelete] = OperationCapabilities{
		MaxBatchSize:       10000,
		OptimalBatchSize:   2000,
		MaxConcurrentBatch: vp.config.WorkerCount,
		EstimatedLatency:   5 * time.Microsecond,
		SupportsParallel:   true,
		SupportsSIMD:       false,
		MemoryRequirement:  128,
	}

	// Search operation capabilities
	vp.capabilities[BatchOperationSearch] = OperationCapabilities{
		MaxBatchSize:       100000,
		OptimalBatchSize:   5000,
		MaxConcurrentBatch: vp.config.WorkerCount,
		EstimatedLatency:   100 * time.Microsecond,
		SupportsParallel:   true,
		SupportsSIMD:       vp.config.EnableSIMD,
		MemoryRequirement:  2048,
	}

	// Similarity operation capabilities
	vp.capabilities[BatchOperationSimilarity] = OperationCapabilities{
		MaxBatchSize:       100000,
		OptimalBatchSize:   10000,
		MaxConcurrentBatch: vp.config.WorkerCount,
		EstimatedLatency:   50 * time.Microsecond,
		SupportsParallel:   true,
		SupportsSIMD:       vp.config.EnableSIMD,
		MemoryRequirement:  1536,
	}

	// Normalize operation capabilities
	vp.capabilities[BatchOperationNormalize] = OperationCapabilities{
		MaxBatchSize:       100000,
		OptimalBatchSize:   10000,
		MaxConcurrentBatch: vp.config.WorkerCount,
		EstimatedLatency:   20 * time.Microsecond,
		SupportsParallel:   true,
		SupportsSIMD:       vp.config.EnableSIMD,
		MemoryRequirement:  1024,
	}

	// Distance operation capabilities
	vp.capabilities[BatchOperationDistance] = OperationCapabilities{
		MaxBatchSize:       100000,
		OptimalBatchSize:   10000,
		MaxConcurrentBatch: vp.config.WorkerCount,
		EstimatedLatency:   60 * time.Microsecond,
		SupportsParallel:   true,
		SupportsSIMD:       vp.config.EnableSIMD,
		MemoryRequirement:  1536,
	}
}

// getDefaultEstimate provides default time estimates for operations
func (vp *vectorProcessor) getDefaultEstimate(operation BatchOperation, numVectors int) time.Duration {
	baseTime := map[BatchOperation]time.Duration{
		BatchOperationInsert:     10 * time.Microsecond,
		BatchOperationUpdate:     15 * time.Microsecond,
		BatchOperationDelete:     5 * time.Microsecond,
		BatchOperationSearch:     100 * time.Microsecond,
		BatchOperationSimilarity: 50 * time.Microsecond,
		BatchOperationNormalize:  20 * time.Microsecond,
		BatchOperationDistance:   60 * time.Microsecond,
	}

	base, exists := baseTime[operation]
	if !exists {
		base = 50 * time.Microsecond
	}

	// Scale by number of vectors with parallelization factor
	parallelizationFactor := float64(vp.config.WorkerCount)
	estimatedTime := time.Duration(float64(numVectors) * float64(base) / parallelizationFactor)
	
	return estimatedTime
}

// updateStatistics updates global statistics after batch processing
func (vp *vectorProcessor) updateStatistics(operation BatchOperation, response *BatchVectorResponse) {
	vp.statistics.mu.Lock()
	defer vp.statistics.mu.Unlock()

	vp.statistics.totalBatches++
	vp.statistics.totalVectors += int64(response.Statistics.TotalItems)
	vp.statistics.totalProcessingTime += response.ProcessingTime
	vp.statistics.totalOperations[operation]++
	vp.statistics.errors += int64(response.ErrorCount)
}

// GetStatistics returns current vector processor statistics
func (vp *vectorProcessor) GetStatistics() vectorStatistics {
	vp.statistics.mu.Lock()
	defer vp.statistics.mu.Unlock()
	
	// Return a copy to avoid lock copying
	stats := vectorStatistics{
		totalBatches:        vp.statistics.totalBatches,
		totalVectors:        vp.statistics.totalVectors,
		totalProcessingTime: vp.statistics.totalProcessingTime,
		totalOperations:     make(map[BatchOperation]int64),
		errors:              vp.statistics.errors,
	}
	for k, v := range vp.statistics.totalOperations {
		stats.totalOperations[k] = v
	}
	return stats
}
