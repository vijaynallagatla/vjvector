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
	Texts         []string               `json:"texts"`
	Model         string                 `json:"model"`
	Provider      embedding.ProviderType `json:"provider"`
	BatchSize     int                    `json:"batch_size"`
	MaxConcurrent int                    `json:"max_concurrent"`
	Timeout       time.Duration          `json:"timeout"`
	Options       map[string]interface{} `json:"options,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	EnableCache   bool                   `json:"enable_cache"`
	Priority      BatchPriority          `json:"priority"`
}

// BatchEmbeddingResponse represents the response from batch embedding generation
type BatchEmbeddingResponse struct {
	Embeddings     [][]float64            `json:"embeddings"`
	Model          string                 `json:"model"`
	Provider       embedding.ProviderType `json:"provider"`
	ProcessingTime time.Duration          `json:"processing_time"`
	TotalTokens    int                    `json:"total_tokens"`
	CacheHits      int                    `json:"cache_hits"`
	CacheMisses    int                    `json:"cache_misses"`
	Errors         []BatchError           `json:"errors,omitempty"`
	Statistics     BatchStatistics        `json:"statistics"`
}

// BatchVectorRequest represents a batch request for vector operations
type BatchVectorRequest struct {
	Operation     BatchOperation         `json:"operation"`
	Vectors       []*core.Vector         `json:"vectors"`
	QueryVector   []float64              `json:"query_vector,omitempty"`
	Collection    string                 `json:"collection,omitempty"`
	BatchSize     int                    `json:"batch_size"`
	MaxConcurrent int                    `json:"max_concurrent"`
	Timeout       time.Duration          `json:"timeout"`
	Options       map[string]interface{} `json:"options,omitempty"`
	Priority      BatchPriority          `json:"priority"`
}

// BatchRAGRequest represents a batch request for RAG operations
type BatchRAGRequest struct {
	Operation     BatchRAGOperation      `json:"operation"`
	Queries       []string               `json:"queries"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Collection    string                 `json:"collection,omitempty"`
	BatchSize     int                    `json:"batch_size"`
	MaxConcurrent int                    `json:"max_concurrent"`
	Timeout       time.Duration          `json:"timeout"`
	Options       map[string]interface{} `json:"options,omitempty"`
	Priority      BatchPriority          `json:"priority"`
	RAGConfig     RAGBatchConfig         `json:"rag_config,omitempty"`
}

// BatchRAGResponse represents the response from batch RAG operations
type BatchRAGResponse struct {
	Operation      BatchRAGOperation `json:"operation"`
	Results        []RAGQueryResult  `json:"results"`
	ProcessingTime time.Duration     `json:"processing_time"`
	ProcessedCount int               `json:"processed_count"`
	ErrorCount     int               `json:"error_count"`
	Errors         []BatchError      `json:"errors,omitempty"`
	Statistics     BatchStatistics   `json:"statistics"`
	RAGMetrics     RAGBatchMetrics   `json:"rag_metrics,omitempty"`
}

// BatchRAGOperation represents the type of batch RAG operation
type BatchRAGOperation string

const (
	BatchRAGOperationQueryExpansion   BatchRAGOperation = "query_expansion"
	BatchRAGOperationResultReranking  BatchRAGOperation = "result_reranking"
	BatchRAGOperationContextRetrieval BatchRAGOperation = "context_retrieval"
	BatchRAGOperationEndToEndRAG      BatchRAGOperation = "end_to_end_rag"
	BatchRAGOperationBatchSearch      BatchRAGOperation = "batch_search"
	BatchRAGOperationBatchRerank      BatchRAGOperation = "batch_rerank"
)

// RAGBatchConfig represents configuration for RAG batch operations
type RAGBatchConfig struct {
	EnableQueryExpansion   bool                 `json:"enable_query_expansion"`
	EnableResultReranking  bool                 `json:"enable_result_reranking"`
	EnableContextAwareness bool                 `json:"enable_context_awareness"`
	QueryExpansionConfig   QueryExpansionConfig `json:"query_expansion_config,omitempty"`
	RerankingConfig        RerankingConfig      `json:"reranking_config,omitempty"`
	ContextConfig          ContextConfig        `json:"context_config,omitempty"`
	SearchConfig           SearchConfig         `json:"search_config,omitempty"`
}

// QueryExpansionConfig represents configuration for query expansion
type QueryExpansionConfig struct {
	Strategies          []string            `json:"strategies"`
	MaxExpansions       int                 `json:"max_expansions"`
	SimilarityThreshold float64             `json:"similarity_threshold"`
	DomainSynonyms      map[string][]string `json:"domain_synonyms,omitempty"`
	CustomPatterns      []string            `json:"custom_patterns,omitempty"`
}

// RerankingConfig represents configuration for result reranking
type RerankingConfig struct {
	Strategies     []string           `json:"strategies"`
	Weights        map[string]float64 `json:"weights"`
	MaxResults     int                `json:"max_results"`
	SemanticWeight float64            `json:"semantic_weight"`
	ContextWeight  float64            `json:"context_weight"`
	HybridWeight   float64            `json:"hybrid_weight"`
}

// ContextConfig represents configuration for context-aware retrieval
type ContextConfig struct {
	UserContext         bool    `json:"user_context"`
	DomainContext       bool    `json:"domain_context"`
	TemporalContext     bool    `json:"temporal_context"`
	LocationContext     bool    `json:"location_context"`
	ContextDecay        float64 `json:"context_decay"`
	ConfidenceThreshold float64 `json:"confidence_threshold"`
}

// SearchConfig represents configuration for search operations
type SearchConfig struct {
	SearchType       string                 `json:"search_type"`
	IndexType        string                 `json:"index_type"`
	SimilarityMetric string                 `json:"similarity_metric"`
	MaxResults       int                    `json:"max_results"`
	Threshold        float64                `json:"threshold"`
	EnableFilters    bool                   `json:"enable_filters"`
	Filters          map[string]interface{} `json:"filters,omitempty"`
}

// RAGQueryResult represents a single RAG query result
type RAGQueryResult struct {
	Query               string                 `json:"query"`
	OriginalQuery       string                 `json:"original_query,omitempty"`
	ExpandedQueries     []string               `json:"expanded_queries,omitempty"`
	Results             []SearchResult         `json:"results"`
	RerankedResults     []SearchResult         `json:"reranked_results,omitempty"`
	ContextEnhancements []string               `json:"context_enhancements,omitempty"`
	ProcessingTime      time.Duration          `json:"processing_time"`
	Confidence          float64                `json:"confidence"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// SearchResult represents a search result with ranking
type SearchResult struct {
	Vector     *core.Vector           `json:"vector"`
	Score      float64                `json:"score"`
	Rank       int                    `json:"rank"`
	Similarity float64                `json:"similarity"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RAGBatchMetrics represents metrics specific to RAG operations
type RAGBatchMetrics struct {
	QueryExpansionCount     int           `json:"query_expansion_count"`
	RerankingCount          int           `json:"reranking_count"`
	ContextEnhancementCount int           `json:"context_enhancement_count"`
	AverageExpansionRatio   float64       `json:"average_expansion_ratio"`
	AverageRerankingTime    time.Duration `json:"average_reranking_time"`
	CacheHitRate            float64       `json:"cache_hit_rate"`
	AccuracyImprovement     float64       `json:"accuracy_improvement"`
}

// RAGOperationCapabilities represents the capabilities for RAG operations
type RAGOperationCapabilities struct {
	MaxBatchSize           int           `json:"max_batch_size"`
	OptimalBatchSize       int           `json:"optimal_batch_size"`
	MaxConcurrentBatch     int           `json:"max_concurrent_batch"`
	EstimatedLatency       time.Duration `json:"estimated_latency"`
	SupportsQueryExpansion bool          `json:"supports_query_expansion"`
	SupportsReranking      bool          `json:"supports_reranking"`
	SupportsContextAware   bool          `json:"supports_context_aware"`
	MemoryRequirement      int64         `json:"memory_requirement"`
	AccuracyImprovement    float64       `json:"accuracy_improvement"`
}

// RAGProcessorStatistics represents RAG processor statistics
type RAGProcessorStatistics struct {
	TotalRAGBatches         int64         `json:"total_rag_batches"`
	TotalQueries            int64         `json:"total_queries"`
	TotalProcessingTime     time.Duration `json:"total_processing_time"`
	QueryExpansionCount     int64         `json:"query_expansion_count"`
	RerankingCount          int64         `json:"reranking_count"`
	ContextEnhancementCount int64         `json:"context_enhancement_count"`
	AverageAccuracy         float64       `json:"average_accuracy"`
	CacheHitRate            float64       `json:"cache_hit_rate"`
	LastUpdated             time.Time     `json:"last_updated"`
}

// BatchVectorResponse represents the response from batch vector operations
type BatchVectorResponse struct {
	Operation      BatchOperation  `json:"operation"`
	Results        interface{}     `json:"results"`
	ProcessingTime time.Duration   `json:"processing_time"`
	ProcessedCount int             `json:"processed_count"`
	ErrorCount     int             `json:"error_count"`
	Errors         []BatchError    `json:"errors,omitempty"`
	Statistics     BatchStatistics `json:"statistics"`
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
	BatchPriorityLow BatchPriority = iota
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
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	TotalItems     int           `json:"total_items"`
	ProcessedItems int           `json:"processed_items"`
	FailedItems    int           `json:"failed_items"`
	Throughput     float64       `json:"throughput"` // items per second
	AverageLatency time.Duration `json:"average_latency"`
	MemoryUsage    int64         `json:"memory_usage"`
	CPUUsage       float64       `json:"cpu_usage"`
}

// BatchProgressCallback represents a callback function for batch progress updates
type BatchProgressCallback func(processed, total int, elapsed time.Duration)

// BatchProcessor defines the interface for batch processing operations
type BatchProcessor interface {
	// ProcessBatchEmbeddings processes a batch of texts for embedding generation
	ProcessBatchEmbeddings(ctx context.Context, req *BatchEmbeddingRequest) (*BatchEmbeddingResponse, error)

	// ProcessBatchVectors processes a batch of vector operations
	ProcessBatchVectors(ctx context.Context, req *BatchVectorRequest) (*BatchVectorResponse, error)

	// ProcessBatchRAG processes a batch of RAG operations
	ProcessBatchRAG(ctx context.Context, req *BatchRAGRequest) (*BatchRAGResponse, error)

	// GetOptimalBatchSize returns the optimal batch size for the given operation
	GetOptimalBatchSize(operation interface{}, totalItems int) int

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

// BatchRAGProcessor defines the interface for batch RAG operations
type BatchRAGProcessor interface {
	// ProcessRAGBatch processes a batch of RAG operations
	ProcessRAGBatch(ctx context.Context, req *BatchRAGRequest) (*BatchRAGResponse, error)

	// GetOptimalBatchSize returns the optimal batch size for RAG operations
	GetOptimalBatchSize(operation BatchRAGOperation, totalQueries int) int

	// EstimateProcessingTime estimates the processing time for a batch RAG operation
	EstimateProcessingTime(req *BatchRAGRequest) time.Duration

	// GetOperationCapabilities returns the capabilities for different RAG operations
	GetOperationCapabilities() map[BatchRAGOperation]RAGOperationCapabilities

	// GetRAGStatistics returns RAG-specific statistics and metrics
	GetRAGStatistics() RAGProcessorStatistics
}

// ProviderCapabilities represents the capabilities of an embedding provider for batch processing
type ProviderCapabilities struct {
	MaxBatchSize       int           `json:"max_batch_size"`
	OptimalBatchSize   int           `json:"optimal_batch_size"`
	MaxConcurrentBatch int           `json:"max_concurrent_batch"`
	EstimatedLatency   time.Duration `json:"estimated_latency"`
	SupportsCaching    bool          `json:"supports_caching"`
	SupportsRetry      bool          `json:"supports_retry"`
	RateLimitRPM       int           `json:"rate_limit_rpm"`
	RateLimitTPM       int           `json:"rate_limit_tpm"`
}

// OperationCapabilities represents the capabilities for vector operations
type OperationCapabilities struct {
	MaxBatchSize       int           `json:"max_batch_size"`
	OptimalBatchSize   int           `json:"optimal_batch_size"`
	MaxConcurrentBatch int           `json:"max_concurrent_batch"`
	EstimatedLatency   time.Duration `json:"estimated_latency"`
	SupportsParallel   bool          `json:"supports_parallel"`
	SupportsSIMD       bool          `json:"supports_simd"`
	MemoryRequirement  int64         `json:"memory_requirement"`
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

	// RAG operation configuration
	RAGConfig RAGBatchConfig `json:"rag_config"`

	// Performance configuration
	PerformanceConfig PerformanceBatchConfig `json:"performance_config"`

	// Monitoring configuration
	MonitoringConfig MonitoringBatchConfig `json:"monitoring_config"`
}

// EmbeddingBatchConfig represents configuration for batch embedding operations
type EmbeddingBatchConfig struct {
	DefaultBatchSize   int                                            `json:"default_batch_size"`
	MaxBatchSize       int                                            `json:"max_batch_size"`
	MaxConcurrentBatch int                                            `json:"max_concurrent_batch"`
	DefaultTimeout     time.Duration                                  `json:"default_timeout"`
	EnableCache        bool                                           `json:"enable_cache"`
	EnableRetry        bool                                           `json:"enable_retry"`
	RetryAttempts      int                                            `json:"retry_attempts"`
	ProviderSettings   map[embedding.ProviderType]ProviderBatchConfig `json:"provider_settings"`
}

// VectorBatchConfig represents configuration for batch vector operations
type VectorBatchConfig struct {
	DefaultBatchSize   int                                     `json:"default_batch_size"`
	MaxBatchSize       int                                     `json:"max_batch_size"`
	MaxConcurrentBatch int                                     `json:"max_concurrent_batch"`
	DefaultTimeout     time.Duration                           `json:"default_timeout"`
	EnableSIMD         bool                                    `json:"enable_simd"`
	EnableParallel     bool                                    `json:"enable_parallel"`
	WorkerCount        int                                     `json:"worker_count"`
	OperationSettings  map[BatchOperation]OperationBatchConfig `json:"operation_settings"`
}

// PerformanceBatchConfig represents configuration for batch performance optimization
type PerformanceBatchConfig struct {
	EnableMemoryPool   bool          `json:"enable_memory_pool"`
	MemoryPoolSize     int64         `json:"memory_pool_size"`
	EnableProfiling    bool          `json:"enable_profiling"`
	ProfilingInterval  time.Duration `json:"profiling_interval"`
	GCOptimization     bool          `json:"gc_optimization"`
	CPUAffinityEnabled bool          `json:"cpu_affinity_enabled"`
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
	BatchSize          int           `json:"batch_size"`
	MaxConcurrentBatch int           `json:"max_concurrent_batch"`
	Timeout            time.Duration `json:"timeout"`
	RetryAttempts      int           `json:"retry_attempts"`
	RateLimitRPM       int           `json:"rate_limit_rpm"`
	RateLimitTPM       int           `json:"rate_limit_tpm"`
}

// OperationBatchConfig represents batch configuration for a specific operation
type OperationBatchConfig struct {
	BatchSize          int           `json:"batch_size"`
	MaxConcurrentBatch int           `json:"max_concurrent_batch"`
	Timeout            time.Duration `json:"timeout"`
	WorkerCount        int           `json:"worker_count"`
	MemoryLimit        int64         `json:"memory_limit"`
}
