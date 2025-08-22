package rag

import (
	"context"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

// QueryType represents the type of RAG query
type QueryType string

const (
	QueryTypeSemantic   QueryType = "semantic"
	QueryTypeHybrid     QueryType = "hybrid"
	QueryTypeContextual QueryType = "contextual"
	QueryTypeMultiModal QueryType = "multimodal"
)

// Query represents a RAG query
type Query struct {
	Text            string                 `json:"text"`
	Type            QueryType              `json:"type"`
	Context         map[string]interface{} `json:"context,omitempty"`
	Filters         map[string]interface{} `json:"filters,omitempty"`
	MaxResults      int                    `json:"max_results,omitempty"`
	MinScore        float64                `json:"min_score,omitempty"`
	IncludeMetadata bool                   `json:"include_metadata,omitempty"`
	Options         map[string]interface{} `json:"options,omitempty"`
}

// QueryResult represents a single RAG query result
type QueryResult struct {
	Vector      *core.Vector           `json:"vector"`
	Score       float64                `json:"score"`
	Distance    float64                `json:"distance"`
	Relevance   float64                `json:"relevance"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Explanation string                 `json:"explanation,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// QueryResponse represents the complete RAG query response
type QueryResponse struct {
	Results        []*QueryResult         `json:"results"`
	Query          *Query                 `json:"query"`
	TotalResults   int                    `json:"total_results"`
	ProcessingTime time.Duration          `json:"processing_time"`
	QueryExpansion []string               `json:"query_expansion,omitempty"`
	RerankingInfo  map[string]interface{} `json:"reranking_info,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Engine represents the main RAG engine
type Engine interface {
	// ProcessQuery processes a RAG query
	ProcessQuery(ctx context.Context, query *Query) (*QueryResponse, error)

	// ProcessBatch processes multiple queries in batch
	ProcessBatch(ctx context.Context, queries []*Query) ([]*QueryResponse, error)

	// ExpandQuery expands a query for better retrieval
	ExpandQuery(ctx context.Context, query *Query) ([]string, error)

	// RerankResults reranks search results
	RerankResults(ctx context.Context, results []*QueryResult, query *Query) ([]*QueryResult, error)

	// GetQueryStats returns query processing statistics
	GetQueryStats() QueryStats

	// HealthCheck checks if the engine is healthy
	HealthCheck(ctx context.Context) error

	// Close closes the engine
	Close() error
}

// QueryProcessor represents a query processing component
type QueryProcessor interface {
	// Process processes a query
	Process(ctx context.Context, query *Query) (*Query, error)

	// Type returns the processor type
	Type() string

	// Priority returns the processing priority (lower = higher priority)
	Priority() int
}

// QueryExpander represents a query expansion component
type QueryExpander interface {
	// Expand expands a query
	Expand(ctx context.Context, query *Query) ([]string, error)

	// Type returns the expander type
	Type() string

	// Confidence returns confidence in the expansion
	Confidence() float64
}

// ResultReranker represents a result reranking component
type ResultReranker interface {
	// Rerank reranks search results
	Rerank(ctx context.Context, results []*QueryResult, query *Query) ([]*QueryResult, error)

	// Type returns the reranker type
	Type() string

	// Confidence returns confidence in the reranking
	Confidence() float64
}

// QueryStats represents query processing statistics
type QueryStats struct {
	TotalQueries      int64         `json:"total_queries"`
	SuccessfulQueries int64         `json:"successful_queries"`
	FailedQueries     int64         `json:"failed_queries"`
	AverageLatency    time.Duration `json:"average_latency"`
	TotalLatency      time.Duration `json:"total_latency"`
	CacheHits         int64         `json:"cache_hits"`
	CacheMisses       int64         `json:"cache_misses"`
	LastQueryTime     time.Time     `json:"last_query_time"`
}

// Config represents RAG engine configuration
type Config struct {
	// Query Processing
	EnableQueryExpansion bool `json:"enable_query_expansion"`
	EnableReranking      bool `json:"enable_reranking"`
	EnableContextAware   bool `json:"enable_context_aware"`
	MaxQueryLength       int  `json:"max_query_length"`
	MaxExpansionTerms    int  `json:"max_expansion_terms"`

	// Performance
	MaxConcurrentQueries int           `json:"max_concurrent_queries"`
	QueryTimeout         time.Duration `json:"query_timeout"`
	BatchSize            int           `json:"batch_size"`

	// Caching
	EnableCache  bool          `json:"enable_cache"`
	CacheTTL     time.Duration `json:"cache_ttl"`
	MaxCacheSize int64         `json:"max_cache_size"`

	// Providers
	EmbeddingProvider   string `json:"embedding_provider"`
	VectorIndexProvider string `json:"vector_index_provider"`

	// Advanced Features
	EnableHybridSearch   bool `json:"enable_hybrid_search"`
	EnableMultiModal     bool `json:"enable_multimodal"`
	EnableExplainability bool `json:"enable_explainability"`
}
