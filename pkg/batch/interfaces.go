// Package batch provides efficient batch processing capabilities for VJVector operations.
// It includes batch embedding generation, vector operations, and performance optimizations.
package batch

import (
	"context"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/embedding"
)

// BatchEmbeddingRequest represents a batch request for embedding generation
type BatchEmbeddingRequest struct {
	Texts         []string                `json:"texts"`
	Model         string                  `json:"model"`
	Provider      embedding.ProviderType  `json:"provider"`
	BatchSize     int                     `json:"batch_size"`
	MaxConcurrent int                     `json:"max_concurrent"`
	Timeout       time.Duration           `json:"timeout"`
	Options       map[string]interface{}  `json:"options,omitempty"`
	Metadata      map[string]interface{}  `json:"metadata,omitempty"`
	EnableCache   bool                    `json:"enable_cache"`
	Priority      BatchPriority           `json:"priority"`
}

// BatchEmbeddingResponse represents the response from batch embedding generation
type BatchEmbeddingResponse struct {
	Embeddings      [][]float64            `json:"embeddings"`
	Model           string                 `json:"model"`
	Provider        embedding.ProviderType `json:"provider"`
	ProcessingTime  time.Duration          `json:"processing_time"`
	TotalTokens     int                    `json:"total_tokens"`
	CacheHits       int                    `json:"cache_hits"`
	CacheMisses     int                    `json:"cache_misses"`
	Errors          []BatchError           `json:"errors,omitempty"`
	Statistics      BatchStatistics        `json:"statistics"`
}

// BatchVectorRequest represents a batch request for vector operations
type BatchVectorRequest struct {
	Operation      BatchOperation          `json:"operation"`
	Vectors        []*core.Vector          `json:"vectors"`
	QueryVector    []float64               `json:"query_vector,omitempty"`
	Collection     string                  `json:"collection,omitempty"`
	BatchSize      int                     `json:"batch_size"`
	MaxConcurrent  int                     `json:"max_concurrent"`
	Timeout        time.Duration           `json:"timeout"`
	Options        map[string]interface{}  `json:"options,omitempty"`
	Priority       BatchPriority           `json:"priority"`
}

// BatchVectorResponse represents the response from batch vector operations
type BatchVectorResponse struct {
	Operation       BatchOperation                 `json:"operation"`
	Results         interface{}                    `json:"results"`
	ProcessingTime  time.Duration                  `json:"processing_time"`
	ProcessedCount  int                            `json:"processed_count"`
	ErrorCount      int                            `json:"error_count"`
	Errors          []BatchError                   `json:"errors,omitempty"`
	Statistics      BatchStatistics                `json:"statistics"`
}

// BatchOperation represents the type of batch operation
type BatchOperation string

const (
	BatchOperationInsert     BatchOperation = "insert"
	BatchOperationUpdate     BatchOperation = "update"
	BatchOperationDelete     BatchOperation = "delete"
	BatchOperationSearch     BatchOperation = "search"
	BatchOperationSimilarity BatchOperation = "similarity"
	BatchOperationNormalize  BatchOperation = "normalize"
	BatchOperationDistance   BatchOperation = "distance"
)

// BatchPriority represents the priority of batch operations
type BatchPriority int

const (
	BatchPriorityLow    BatchPriority = iota
	BatchPriorityNormal
	BatchPriorityHigh
	BatchPriorityCritical
)

// BatchError represents an error that occurred during batch processing
type BatchError struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// BatchStatistics represents statistics for batch operations
type BatchStatistics struct {
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	TotalItems      int           `json:"total_items"`
	ProcessedItems  int           `json:"processed_items"`
	FailedItems     int           `json:"failed_items"`
	Throughput      float64       `json:"throughput"` // items per second
	AverageLatency  time.Duration `json:"average_latency"`
	MemoryUsage     int64         `json:"memory_usage"`
	CPUUsage        float64       `json:"cpu_usage"`
}

// BatchProgressCallback represents a callback function for batch progress updates
type BatchProgressCallback func(processed, total int, elapsed time.Duration)

// BatchProcessor defines the interface for batch processing operations
type BatchProcessor interface {
	// ProcessBatchEmbeddings processes a batch of texts for embedding generation
	ProcessBatchEmbeddings(ctx context.Context, req *BatchEmbeddingRequest) (*BatchEmbeddingResponse, error)

	// ProcessBatchVectors processes a batch of vector operations
	ProcessBatchVectors(ctx context.Context, req *BatchVectorRequest) (*BatchVectorResponse, error)

	// GetOptimalBatchSize returns the optimal batch size for the given operation
	GetOptimalBatchSize(operation BatchOperation, totalItems int) int

	// GetStatistics returns current batch processing statistics
	GetStatistics() BatchProcessorStatistics

	// SetProgressCallback sets a callback for progress updates
	SetProgressCallback(callback BatchProgressCallback)

	// Close closes the batch processor and cleans up resources
	Close() error
}

// BatchEmbeddingProcessor defines the interface for batch embedding generation
type BatchEmbeddingProcessor interface {
	// GenerateBatchEmbeddings generates embeddings for a batch of texts
	GenerateBatchEmbeddings(ctx context.Context, req *BatchEmbeddingRequest) (*BatchEmbeddingResponse, error)

	// GetOptimalBatchSize returns the optimal batch size for embedding generation
	GetOptimalBatchSize(provider embedding.ProviderType, totalTexts int) int

	// EstimateProcessingTime estimates the processing time for a batch
	EstimateProcessingTime(req *BatchEmbeddingRequest) time.Duration

	// GetProviderCapabilities returns the capabilities of embedding providers for batch processing
	GetProviderCapabilities() map[embedding.ProviderType]ProviderCapabilities
}

// BatchVectorProcessor defines the interface for batch vector operations
type BatchVectorProcessor interface {
	// ProcessVectorBatch processes a batch of vector operations
	ProcessVectorBatch(ctx context.Context, req *BatchVectorRequest) (*BatchVectorResponse, error)

	// GetOptimalBatchSize returns the optimal batch size for vector operations
	GetOptimalBatchSize(operation BatchOperation, totalVectors int) int

	// EstimateProcessingTime estimates the processing time for a batch vector operation
	EstimateProcessingTime(req *BatchVectorRequest) time.Duration

	// GetOperationCapabilities returns the capabilities for different vector operations
	GetOperationCapabilities() map[BatchOperation]OperationCapabilities
}

// ProviderCapabilities represents the capabilities of an embedding provider for batch processing
type ProviderCapabilities struct {
	MaxBatchSize        int           `json:"max_batch_size"`
	OptimalBatchSize    int           `json:"optimal_batch_size"`
	MaxConcurrentBatch  int           `json:"max_concurrent_batch"`
	EstimatedLatency    time.Duration `json:"estimated_latency"`
	SupportsCaching     bool          `json:"supports_caching"`
	SupportsRetry       bool          `json:"supports_retry"`
	RateLimitRPM        int           `json:"rate_limit_rpm"`
	RateLimitTPM        int           `json:"rate_limit_tpm"`
}

// OperationCapabilities represents the capabilities for vector operations
type OperationCapabilities struct {
	MaxBatchSize        int           `json:"max_batch_size"`
	OptimalBatchSize    int           `json:"optimal_batch_size"`
	MaxConcurrentBatch  int           `json:"max_concurrent_batch"`
	EstimatedLatency    time.Duration `json:"estimated_latency"`
	SupportsParallel    bool          `json:"supports_parallel"`
	SupportsSIMD        bool          `json:"supports_simd"`
	MemoryRequirement   int64         `json:"memory_requirement"`
}

// BatchProcessorStatistics represents overall batch processor statistics
type BatchProcessorStatistics struct {
	TotalBatches        int64         `json:"total_batches"`
	TotalItems          int64         `json:"total_items"`
	TotalProcessingTime time.Duration `json:"total_processing_time"`
	AverageThroughput   float64       `json:"average_throughput"`
	AverageLatency      time.Duration `json:"average_latency"`
	SuccessRate         float64       `json:"success_rate"`
	CacheHitRate        float64       `json:"cache_hit_rate"`
	MemoryUsage         int64         `json:"memory_usage"`
	ActiveBatches       int           `json:"active_batches"`
}

// BatchConfig represents configuration for batch processing
type BatchConfig struct {
	// Embedding configuration
	EmbeddingConfig EmbeddingBatchConfig `json:"embedding_config"`

	// Vector operation configuration
	VectorConfig VectorBatchConfig `json:"vector_config"`

	// Performance configuration
	PerformanceConfig PerformanceBatchConfig `json:"performance_config"`

	// Monitoring configuration
	MonitoringConfig MonitoringBatchConfig `json:"monitoring_config"`
}

// EmbeddingBatchConfig represents configuration for batch embedding operations
type EmbeddingBatchConfig struct {
	DefaultBatchSize    int                                             `json:"default_batch_size"`
	MaxBatchSize        int                                             `json:"max_batch_size"`
	MaxConcurrentBatch  int                                             `json:"max_concurrent_batch"`
	DefaultTimeout      time.Duration                                   `json:"default_timeout"`
	EnableCache         bool                                            `json:"enable_cache"`
	EnableRetry         bool                                            `json:"enable_retry"`
	RetryAttempts       int                                             `json:"retry_attempts"`
	ProviderSettings    map[embedding.ProviderType]ProviderBatchConfig  `json:"provider_settings"`
}

// VectorBatchConfig represents configuration for batch vector operations
type VectorBatchConfig struct {
	DefaultBatchSize    int                                      `json:"default_batch_size"`
	MaxBatchSize        int                                      `json:"max_batch_size"`
	MaxConcurrentBatch  int                                      `json:"max_concurrent_batch"`
	DefaultTimeout      time.Duration                            `json:"default_timeout"`
	EnableSIMD          bool                                     `json:"enable_simd"`
	EnableParallel      bool                                     `json:"enable_parallel"`
	WorkerCount         int                                      `json:"worker_count"`
	OperationSettings   map[BatchOperation]OperationBatchConfig  `json:"operation_settings"`
}

// PerformanceBatchConfig represents configuration for batch performance optimization
type PerformanceBatchConfig struct {
	EnableMemoryPool    bool          `json:"enable_memory_pool"`
	MemoryPoolSize      int64         `json:"memory_pool_size"`
	EnableProfiling     bool          `json:"enable_profiling"`
	ProfilingInterval   time.Duration `json:"profiling_interval"`
	GCOptimization      bool          `json:"gc_optimization"`
	CPUAffinityEnabled  bool          `json:"cpu_affinity_enabled"`
}

// MonitoringBatchConfig represents configuration for batch monitoring
type MonitoringBatchConfig struct {
	EnableMetrics       bool          `json:"enable_metrics"`
	MetricsInterval     time.Duration `json:"metrics_interval"`
	EnableProgressLogs  bool          `json:"enable_progress_logs"`
	LogInterval         time.Duration `json:"log_interval"`
	EnableHealthCheck   bool          `json:"enable_health_check"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

// ProviderBatchConfig represents batch configuration for a specific provider
type ProviderBatchConfig struct {
	BatchSize           int           `json:"batch_size"`
	MaxConcurrentBatch  int           `json:"max_concurrent_batch"`
	Timeout             time.Duration `json:"timeout"`
	RetryAttempts       int           `json:"retry_attempts"`
	RateLimitRPM        int           `json:"rate_limit_rpm"`
	RateLimitTPM        int           `json:"rate_limit_tpm"`
}

// OperationBatchConfig represents batch configuration for a specific operation
type OperationBatchConfig struct {
	BatchSize           int           `json:"batch_size"`
	MaxConcurrentBatch  int           `json:"max_concurrent_batch"`
	Timeout             time.Duration `json:"timeout"`
	WorkerCount         int           `json:"worker_count"`
	MemoryLimit         int64         `json:"memory_limit"`
}
