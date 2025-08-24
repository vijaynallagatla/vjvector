package batch

import (
	"context"
	"testing"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/embedding"
)

func TestNewBatchProcessor(t *testing.T) {
	tests := []struct {
		name           string
		config         BatchConfig
		expectDefaults bool
	}{
		{
			name:           "default_config",
			config:         GetDefaultConfig(),
			expectDefaults: true,
		},
		{
			name: "custom_config",
			config: BatchConfig{
				EmbeddingConfig: EmbeddingBatchConfig{
					DefaultBatchSize: 50,
					MaxBatchSize:     500,
				},
				VectorConfig: VectorBatchConfig{
					DefaultBatchSize: 2000,
					WorkerCount:      4,
				},
			},
			expectDefaults: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEmbeddingService{}
			processor := NewBatchProcessor(tt.config, mockService, &mockRAGEngine{})

			if processor == nil {
				t.Fatal("Expected processor to be created, got nil")
			}

			// Test that processor can be closed
			err := processor.Close()
			if err != nil {
				t.Errorf("Expected no error closing processor, got %v", err)
			}
		})
	}
}

func TestProcessBatchEmbeddings(t *testing.T) {
	tests := []struct {
		name        string
		request     *BatchEmbeddingRequest
		expectError bool
	}{
		{
			name: "valid_small_batch",
			request: &BatchEmbeddingRequest{
				Texts:         []string{"hello", "world", "test"},
				Model:         "text-embedding-ada-002",
				Provider:      embedding.ProviderTypeOpenAI,
				BatchSize:     10,
				MaxConcurrent: 2,
				Timeout:       30 * time.Second,
				EnableCache:   true,
				Priority:      BatchPriorityNormal,
			},
			expectError: false,
		},
		{
			name: "valid_large_batch",
			request: &BatchEmbeddingRequest{
				Texts:         generateTestTexts(100),
				Model:         "text-embedding-ada-002",
				Provider:      embedding.ProviderTypeOpenAI,
				BatchSize:     25,
				MaxConcurrent: 4,
				Timeout:       60 * time.Second,
				EnableCache:   true,
				Priority:      BatchPriorityHigh,
			},
			expectError: false,
		},
		{
			name: "empty_texts",
			request: &BatchEmbeddingRequest{
				Texts:    []string{},
				Model:    "text-embedding-ada-002",
				Provider: embedding.ProviderTypeOpenAI,
			},
			expectError: true,
		},
	}

	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService, &mockRAGEngine{})
	defer func() {
		if err := processor.Close(); err != nil {
			t.Fatalf("Failed to close processor: %v", err)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			response, err := processor.ProcessBatchEmbeddings(ctx, tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if response == nil {
				t.Fatal("Expected response, got nil")
			}

			if len(response.Embeddings) != len(tt.request.Texts) {
				t.Errorf("Expected %d embeddings, got %d", len(tt.request.Texts), len(response.Embeddings))
			}

			if response.ProcessingTime <= 0 {
				t.Error("Expected positive processing time")
			}

			if response.TotalTokens <= 0 {
				t.Error("Expected positive token count")
			}
		})
	}
}

func TestProcessBatchVectors(t *testing.T) {
	tests := []struct {
		name        string
		request     *BatchVectorRequest
		expectError bool
	}{
		{
			name: "insert_operation",
			request: &BatchVectorRequest{
				Operation:     BatchOperationInsert,
				Vectors:       generateTestVectors(10),
				BatchSize:     5,
				MaxConcurrent: 2,
				Timeout:       30 * time.Second,
				Priority:      BatchPriorityNormal,
			},
			expectError: false,
		},
		// {
		// 	name: "similarity_operation",
		// 	request: &BatchVectorRequest{
		// 		Operation:     BatchOperationSimilarity,
		// 		Vectors:       generateTestVectors(50),
		// 		QueryVector:   generateTestEmbedding(128),
		// 		BatchSize:     25,
		// 		MaxConcurrent: 4,
		// 		Timeout:       60 * time.Second,
		// 		Priority:      BatchPriorityHigh,
		// 	},
		// 	expectError: false,
		// },
		{
			name: "normalize_operation",
			request: &BatchVectorRequest{
				Operation:     BatchOperationNormalize,
				Vectors:       generateTestVectors(20),
				BatchSize:     10,
				MaxConcurrent: 2,
				Timeout:       30 * time.Second,
				Priority:      BatchPriorityNormal,
			},
			expectError: false,
		},
		{
			name: "empty_vectors",
			request: &BatchVectorRequest{
				Operation: BatchOperationInsert,
				Vectors:   []*core.Vector{},
			},
			expectError: true,
		},
	}

	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService, &mockRAGEngine{})
	defer func() {
		if err := processor.Close(); err != nil {
			t.Fatalf("Failed to close processor: %v", err)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			response, err := processor.ProcessBatchVectors(ctx, tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if response == nil {
				t.Fatal("Expected response, got nil")
			}

			if response.Operation != tt.request.Operation {
				t.Errorf("Expected operation %s, got %s", tt.request.Operation, response.Operation)
			}

			if response.ProcessingTime <= 0 {
				t.Error("Expected positive processing time")
			}

			if response.ProcessedCount < 0 {
				t.Error("Expected non-negative processed count")
			}
		})
	}
}

func TestGetOptimalBatchSize(t *testing.T) {
	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService, &mockRAGEngine{})
	defer func() {
		if err := processor.Close(); err != nil {
			t.Fatalf("Failed to close processor: %v", err)
		}
	}()

	tests := []struct {
		name       string
		operation  BatchOperation
		totalItems int
		expectMin  int
		expectMax  int
	}{
		{
			name:       "insert_small",
			operation:  BatchOperationInsert,
			totalItems: 100,
			expectMin:  100,
			expectMax:  1000,
		},
		{
			name:       "search_large",
			operation:  BatchOperationSearch,
			totalItems: 10000,
			expectMin:  1000,
			expectMax:  10000,
		},
		{
			name:       "similarity_medium",
			operation:  BatchOperationSimilarity,
			totalItems: 1000,
			expectMin:  1000,
			expectMax:  10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchSize := processor.GetOptimalBatchSize(tt.operation, tt.totalItems)

			if batchSize < tt.expectMin {
				t.Errorf("Expected batch size >= %d, got %d", tt.expectMin, batchSize)
			}

			if batchSize > tt.expectMax {
				t.Errorf("Expected batch size <= %d, got %d", tt.expectMax, batchSize)
			}
		})
	}
}

func TestGetStatistics(t *testing.T) {
	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService, &mockRAGEngine{})
	defer func() {
		if err := processor.Close(); err != nil {
			t.Fatalf("Failed to close processor: %v", err)
		}
	}()

	// Get initial statistics
	initialStats := processor.GetStatistics()
	if initialStats.TotalBatches != 0 {
		t.Error("Expected initial total batches to be 0")
	}

	// Process a batch to generate statistics
	ctx := context.Background()
	req := &BatchEmbeddingRequest{
		Texts:    []string{"test1", "test2", "test3"},
		Model:    "test-model",
		Provider: embedding.ProviderTypeOpenAI,
	}

	_, err := processor.ProcessBatchEmbeddings(ctx, req)
	if err != nil {
		t.Fatalf("Unexpected error processing batch: %v", err)
	}

	// Get updated statistics
	stats := processor.GetStatistics()
	if stats.TotalBatches != 1 {
		t.Errorf("Expected total batches to be 1, got %d", stats.TotalBatches)
	}

	if stats.TotalItems != 3 {
		t.Errorf("Expected total items to be 3, got %d", stats.TotalItems)
	}

	if stats.AverageThroughput <= 0 {
		t.Error("Expected positive average throughput")
	}
}

func TestProgressCallback(t *testing.T) {
	config := GetDefaultConfig()
	mockService := &mockEmbeddingService{}
	processor := NewBatchProcessor(config, mockService, &mockRAGEngine{})
	defer func() {
		if err := processor.Close(); err != nil {
			t.Fatalf("Failed to close processor: %v", err)
		}
	}()

	progressCalled := false
	processor.SetProgressCallback(func(processed, total int, elapsed time.Duration) {
		progressCalled = true
		if processed < 0 || processed > total {
			t.Errorf("Invalid progress: processed=%d, total=%d", processed, total)
		}
		if elapsed < 0 {
			t.Errorf("Invalid elapsed time: %v", elapsed)
		}
	})

	// Test that the callback can be set and retrieved
	if progressCalled {
		t.Error("Progress callback should not be called before processing")
	}

	// Test removing callback
	processor.SetProgressCallback(nil)

	ctx := context.Background()
	req := &BatchEmbeddingRequest{
		Texts:    generateTestTexts(10), // Smaller batch for faster test
		Model:    "test-model",
		Provider: embedding.ProviderTypeOpenAI,
	}

	_, err := processor.ProcessBatchEmbeddings(ctx, req)
	if err != nil {
		t.Fatalf("Unexpected error processing batch: %v", err)
	}

	// Since callback was removed, it should not be called
	if progressCalled {
		t.Error("Progress callback should not be called after being removed")
	}
}
