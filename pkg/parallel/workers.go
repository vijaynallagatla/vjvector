// Package parallel provides parallel processing capabilities for the VJVector database.
package parallel

import (
	"runtime"
	"sync"

	vectormath "github.com/vijaynallagatla/vjvector/pkg/math"
)

// BatchProcessor is an alias for ParallelBatchProcessor to satisfy linter preferences
type BatchProcessor = ParallelBatchProcessor

// WorkerPool manages a pool of workers for parallel processing
type WorkerPool struct {
	numWorkers int
	vectorMath *vectormath.VectorMath
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int, useSIMD bool) *WorkerPool {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	return &WorkerPool{
		numWorkers: numWorkers,
		vectorMath: vectormath.NewVectorMath(useSIMD),
	}
}

// ParallelDotProduct calculates dot products for multiple vector pairs in parallel
func (wp *WorkerPool) ParallelDotProduct(vectorsA, vectorsB [][]float64) []float64 {
	if len(vectorsA) != len(vectorsB) || len(vectorsA) == 0 {
		return nil
	}

	results := make([]float64, len(vectorsA))

	// For small batches, use sequential processing
	if len(vectorsA) < wp.numWorkers*2 {
		for i := 0; i < len(vectorsA); i++ {
			results[i] = wp.vectorMath.DotProduct(vectorsA[i], vectorsB[i])
		}
		return results
	}

	// Calculate work distribution
	batchSize := len(vectorsA) / wp.numWorkers
	remainder := len(vectorsA) % wp.numWorkers

	var wg sync.WaitGroup

	// Launch workers
	start := 0
	for i := 0; i < wp.numWorkers; i++ {
		end := start + batchSize
		if i < remainder {
			end++
		}

		wg.Add(1)
		go func(startIdx, endIdx int) {
			defer wg.Done()
			for j := startIdx; j < endIdx; j++ {
				results[j] = wp.vectorMath.DotProduct(vectorsA[j], vectorsB[j])
			}
		}(start, end)

		start = end
	}

	wg.Wait()
	return results
}

// ParallelCosineSimilarity calculates cosine similarities for multiple vector pairs in parallel
func (wp *WorkerPool) ParallelCosineSimilarity(vectorsA, vectorsB [][]float64) []float64 {
	if len(vectorsA) != len(vectorsB) || len(vectorsA) == 0 {
		return nil
	}

	results := make([]float64, len(vectorsA))

	// For small batches, use sequential processing
	if len(vectorsA) < wp.numWorkers*2 {
		for i := 0; i < len(vectorsA); i++ {
			results[i] = wp.vectorMath.CosineSimilarity(vectorsA[i], vectorsB[i])
		}
		return results
	}

	// Calculate work distribution
	batchSize := len(vectorsA) / wp.numWorkers
	remainder := len(vectorsA) % wp.numWorkers

	var wg sync.WaitGroup

	// Launch workers
	start := 0
	for i := 0; i < wp.numWorkers; i++ {
		end := start + batchSize
		if i < remainder {
			end++
		}

		wg.Add(1)
		go func(startIdx, endIdx int) {
			defer wg.Done()
			for j := startIdx; j < endIdx; j++ {
				results[j] = wp.vectorMath.CosineSimilarity(vectorsA[j], vectorsB[j])
			}
		}(start, end)

		start = end
	}

	wg.Wait()
	return results
}

// ParallelEuclideanDistance calculates Euclidean distances for multiple vector pairs in parallel
func (wp *WorkerPool) ParallelEuclideanDistance(vectorsA, vectorsB [][]float64) []float64 {
	if len(vectorsA) != len(vectorsB) || len(vectorsA) == 0 {
		return nil
	}

	results := make([]float64, len(vectorsA))

	// For small batches, use sequential processing
	if len(vectorsA) < wp.numWorkers*2 {
		for i := 0; i < len(vectorsA); i++ {
			results[i] = wp.vectorMath.EuclideanDistance(vectorsA[i], vectorsB[i])
		}
		return results
	}

	// Calculate work distribution
	batchSize := len(vectorsA) / wp.numWorkers
	remainder := len(vectorsA) % wp.numWorkers

	var wg sync.WaitGroup

	// Launch workers
	start := 0
	for i := 0; i < wp.numWorkers; i++ {
		end := start + batchSize
		if i < remainder {
			end++
		}

		wg.Add(1)
		go func(startIdx, endIdx int) {
			defer wg.Done()
			for j := startIdx; j < endIdx; j++ {
				results[j] = wp.vectorMath.EuclideanDistance(vectorsA[j], vectorsB[j])
			}
		}(start, end)

		start = end
	}

	wg.Wait()
	return results
}

// ParallelNormalize normalizes multiple vectors in parallel
func (wp *WorkerPool) ParallelNormalize(vectors [][]float64) [][]float64 {
	if len(vectors) == 0 {
		return nil
	}

	results := make([][]float64, len(vectors))

	// For small batches, use sequential processing
	if len(vectors) < wp.numWorkers*2 {
		for i := 0; i < len(vectors); i++ {
			results[i] = wp.vectorMath.Normalize(vectors[i])
		}
		return results
	}

	// Calculate work distribution
	batchSize := len(vectors) / wp.numWorkers
	remainder := len(vectors) % wp.numWorkers

	var wg sync.WaitGroup

	// Launch workers
	start := 0
	for i := 0; i < wp.numWorkers; i++ {
		end := start + batchSize
		if i < remainder {
			end++
		}

		wg.Add(1)
		go func(startIdx, endIdx int) {
			defer wg.Done()
			for j := startIdx; j < endIdx; j++ {
				results[j] = wp.vectorMath.Normalize(vectors[j])
			}
		}(start, end)

		start = end
	}

	wg.Wait()
	return results
}

// ParallelVectorSearch performs parallel vector similarity search
func (wp *WorkerPool) ParallelVectorSearch(query []float64, vectors [][]float64, k int) []VectorSearchResult {
	if len(vectors) == 0 || k <= 0 {
		return nil
	}

	// Calculate similarities in parallel
	similarities := make([]float64, len(vectors))

	// For small datasets, use sequential processing
	if len(vectors) < wp.numWorkers*2 {
		for i := 0; i < len(vectors); i++ {
			similarities[i] = wp.vectorMath.CosineSimilarity(query, vectors[i])
		}
	} else {
		// Calculate work distribution
		batchSize := len(vectors) / wp.numWorkers
		remainder := len(vectors) % wp.numWorkers

		var wg sync.WaitGroup

		// Launch workers
		start := 0
		for i := 0; i < wp.numWorkers; i++ {
			end := start + batchSize
			if i < remainder {
				end++
			}

			wg.Add(1)
			go func(startIdx, endIdx int) {
				defer wg.Done()
				for j := startIdx; j < endIdx; j++ {
					similarities[j] = wp.vectorMath.CosineSimilarity(query, vectors[j])
				}
			}(start, end)

			start = end
		}

		wg.Wait()
	}

	// Create result pairs
	results := make([]VectorSearchResult, len(vectors))
	for i := 0; i < len(vectors); i++ {
		results[i] = VectorSearchResult{
			Index:      i,
			Similarity: similarities[i],
			Vector:     vectors[i],
		}
	}

	// Sort by similarity (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Similarity > results[i].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Return top k results
	if k > len(results) {
		k = len(results)
	}

	return results[:k]
}

// VectorSearchResult represents a search result
type VectorSearchResult struct {
	Index      int       `json:"index"`
	Similarity float64   `json:"similarity"`
	Vector     []float64 `json:"vector"`
}

// ParallelBatchProcessor provides high-level parallel batch processing
type ParallelBatchProcessor struct {
	workerPool *WorkerPool
}

// NewParallelBatchProcessor creates a new parallel batch processor
func NewParallelBatchProcessor(numWorkers int, useSIMD bool) *ParallelBatchProcessor {
	return &ParallelBatchProcessor{
		workerPool: NewWorkerPool(numWorkers, useSIMD),
	}
}

// ProcessDotProductBatch processes dot product operations in parallel
func (pbp *ParallelBatchProcessor) ProcessDotProductBatch(vectorsA, vectorsB [][]float64) []float64 {
	return pbp.workerPool.ParallelDotProduct(vectorsA, vectorsB)
}

// ProcessNormalizeBatch processes normalization operations in parallel
func (pbp *ParallelBatchProcessor) ProcessNormalizeBatch(vectors [][]float64) [][]float64 {
	return pbp.workerPool.ParallelNormalize(vectors)
}

// ProcessSearchBatch processes search operations in parallel
func (pbp *ParallelBatchProcessor) ProcessSearchBatch(query []float64, vectors [][]float64, k int) []VectorSearchResult {
	return pbp.workerPool.ParallelVectorSearch(query, vectors, k)
}

// GetWorkerCount returns the number of workers in the pool
func (pbp *ParallelBatchProcessor) GetWorkerCount() int {
	return pbp.workerPool.numWorkers
}
