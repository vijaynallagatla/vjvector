package api

import "time"

// Vector represents a vector in the API layer
type Vector struct {
	ID         string                 `json:"id"`
	Collection string                 `json:"collection,omitempty"`
	Embedding  []float64              `json:"embedding"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// CreateIndexRequest represents the request to create a new index
type CreateIndexRequest struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	Dimension      int    `json:"dimension"`
	MaxElements    int    `json:"max_elements"`
	M              int    `json:"m,omitempty"`
	EfConstruction int    `json:"ef_construction,omitempty"`
	EfSearch       int    `json:"ef_search,omitempty"`
	MaxLayers      int    `json:"max_layers,omitempty"`
	NumClusters    int    `json:"num_clusters,omitempty"`
	ClusterSize    int    `json:"cluster_size,omitempty"`
	DistanceMetric string `json:"distance_metric"`
	Normalize      bool   `json:"normalize"`
}

// InsertVectorsRequest represents the request to insert vectors
type InsertVectorsRequest struct {
	Vectors []*Vector `json:"vectors"`
}

// SearchRequest represents the request to search for similar vectors
type SearchRequest struct {
	Query []float64 `json:"query"`
	K     int       `json:"k"`
}

// RAG Operations Types

// RAGOperation represents the type of RAG operation
type RAGOperation string

const (
	RAGOperationQueryExpansion   RAGOperation = "query_expansion"
	RAGOperationResultReranking  RAGOperation = "result_reranking"
	RAGOperationContextRetrieval RAGOperation = "context_retrieval"
	RAGOperationEndToEndRAG      RAGOperation = "end_to_end_rag"
	RAGOperationBatchSearch      RAGOperation = "batch_search"
	RAGOperationBatchRerank      RAGOperation = "batch_rerank"
)

// RAGRequest represents a RAG operation request
type RAGRequest struct {
	Operation  RAGOperation           `json:"operation"`
	Query      string                 `json:"query"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Collection string                 `json:"collection,omitempty"`
	Options    map[string]interface{} `json:"options,omitempty"`
	RAGConfig  RAGConfig              `json:"rag_config,omitempty"`
}

// RAGResponse represents a RAG operation response
type RAGResponse struct {
	Operation           RAGOperation           `json:"operation"`
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

// BatchRAGRequest represents a batch RAG operation request
type BatchRAGRequest struct {
	Operation     RAGOperation           `json:"operation"`
	Queries       []string               `json:"queries"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Collection    string                 `json:"collection,omitempty"`
	BatchSize     int                    `json:"batch_size"`
	MaxConcurrent int                    `json:"max_concurrent"`
	Timeout       string                 `json:"timeout"`
	Options       map[string]interface{} `json:"options,omitempty"`
	RAGConfig     RAGConfig              `json:"rag_config,omitempty"`
}

// BatchRAGResponse represents a batch RAG operation response
type BatchRAGResponse struct {
	Operation      RAGOperation    `json:"operation"`
	Results        []RAGResponse   `json:"results"`
	ProcessingTime time.Duration   `json:"processing_time"`
	ProcessedCount int             `json:"processed_count"`
	ErrorCount     int             `json:"error_count"`
	Errors         []BatchError    `json:"errors,omitempty"`
	Statistics     BatchStatistics `json:"statistics"`
	RAGMetrics     RAGMetrics      `json:"rag_metrics,omitempty"`
}

// RAGConfig represents configuration for RAG operations
type RAGConfig struct {
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

// SearchResult represents a search result with ranking
type SearchResult struct {
	Vector     *Vector                `json:"vector"`
	Score      float64                `json:"score"`
	Rank       int                    `json:"rank"`
	Similarity float64                `json:"similarity"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RAGMetrics represents metrics specific to RAG operations
type RAGMetrics struct {
	QueryExpansionCount     int           `json:"query_expansion_count"`
	RerankingCount          int           `json:"reranking_count"`
	ContextEnhancementCount int           `json:"context_enhancement_count"`
	AverageExpansionRatio   float64       `json:"average_expansion_ratio"`
	AverageRerankingTime    time.Duration `json:"average_reranking_time"`
	CacheHitRate            float64       `json:"cache_hit_rate"`
	AccuracyImprovement     float64       `json:"accuracy_improvement"`
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

// BatchError represents an error that occurred during batch processing
type BatchError struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
	Code    string `json:"code"`
}
