package rag

import (
	"context"
	"fmt"
	"sync"
	"time"

	"log/slog"

	"github.com/vijaynallagatla/vjvector/pkg/embedding"
	"github.com/vijaynallagatla/vjvector/pkg/index"
)

// engine implements the RAG engine
type engine struct {
	config           *Config
	embeddingService embedding.Service
	vectorIndex      index.VectorIndex
	processors       []QueryProcessor
	expanders        []QueryExpander
	rerankers        []ResultReranker
	cache            *QueryCache
	stats            *QueryStats
	mu               sync.RWMutex
	logger           *slog.Logger
}

// NewEngine creates a new RAG engine
func NewEngine(config *Config, embeddingService embedding.Service, vectorIndex index.VectorIndex) (Engine, error) {
	if config == nil {
		config = &Config{
			EnableQueryExpansion: true,
			EnableReranking:      true,
			EnableContextAware:   true,
			MaxQueryLength:       1000,
			MaxExpansionTerms:    5,
			MaxConcurrentQueries: 10,
			QueryTimeout:         30 * time.Second,
			BatchSize:            100,
			EnableCache:          true,
			CacheTTL:             5 * time.Minute,
			MaxCacheSize:         1000,
		}
	}

	// Initialize cache
	var cache *QueryCache
	if config.EnableCache {
		cache = NewQueryCache(config.CacheTTL, config.MaxCacheSize)
	}

	// Initialize statistics
	stats := &QueryStats{
		LastQueryTime: time.Now(),
	}

	e := &engine{
		config:           config,
		embeddingService: embeddingService,
		vectorIndex:      vectorIndex,
		processors:       make([]QueryProcessor, 0),
		expanders:        make([]QueryExpander, 0),
		rerankers:        make([]ResultReranker, 0),
		cache:            cache,
		stats:            stats,
		logger:           slog.Default(),
	}

	// Register default components
	e.registerDefaultComponents()

	return e, nil
}

// ProcessQuery processes a RAG query
func (e *engine) ProcessQuery(ctx context.Context, query *Query) (*QueryResponse, error) {
	start := time.Now()
	e.mu.Lock()
	e.stats.TotalQueries++
	e.mu.Unlock()

	// Check cache first
	if e.cache != nil {
		if cached, hit := e.cache.Get(query); hit {
			e.mu.Lock()
			e.stats.CacheHits++
			e.mu.Unlock()
			return cached, nil
		}
	}

	// Apply timeout
	if e.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.config.QueryTimeout)
		defer cancel()
	}

	// Process query through processors
	processedQuery, err := e.processQuery(query)
	if err != nil {
		e.mu.Lock()
		e.stats.FailedQueries++
		e.mu.Unlock()
		return nil, fmt.Errorf("query processing failed: %w", err)
	}

	// Generate query embedding
	embeddingReq := &embedding.EmbeddingRequest{
		Texts:    []string{processedQuery.Text},
		Model:    "text-embedding-ada-002", // Default model
		Provider: embedding.ProviderTypeOpenAI,
	}

	embeddingResp, err := e.embeddingService.GenerateEmbeddings(ctx, embeddingReq)
	if err != nil {
		e.mu.Lock()
		e.stats.FailedQueries++
		e.mu.Unlock()
		return nil, fmt.Errorf("embedding generation failed: %w", err)
	}

	if len(embeddingResp.Embeddings) == 0 {
		e.mu.Lock()
		e.stats.FailedQueries++
		e.mu.Unlock()
		return nil, fmt.Errorf("no embeddings generated")
	}

	// Search vector index
	queryVector := embeddingResp.Embeddings[0]
	maxResults := processedQuery.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}

	searchResults, err := e.vectorIndex.Search(queryVector, maxResults)
	if err != nil {
		e.mu.Lock()
		e.stats.FailedQueries++
		e.mu.Unlock()
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Convert to RAG results
	ragResults := make([]*QueryResult, len(searchResults))
	for i, result := range searchResults {
		ragResults[i] = &QueryResult{
			Vector:    result.Vector,
			Score:     result.Score,
			Distance:  result.Distance,
			Relevance: 1.0 - result.Distance, // Convert distance to relevance
		}
	}

	// Apply reranking if enabled
	if e.config.EnableReranking && len(ragResults) > 1 {
		rerankedResults, err := e.RerankResults(ctx, ragResults, processedQuery)
		if err != nil {
			e.logger.Warn("Reranking failed, using original results", "error", err)
		} else {
			ragResults = rerankedResults
		}
	}

	// Create response
	response := &QueryResponse{
		Results:        ragResults,
		Query:          processedQuery,
		TotalResults:   len(ragResults),
		ProcessingTime: time.Since(start),
		Metadata:       make(map[string]interface{}),
	}

	// Add query expansion info if enabled
	if e.config.EnableQueryExpansion {
		expandedTerms, err := e.ExpandQuery(ctx, processedQuery)
		if err == nil {
			response.QueryExpansion = expandedTerms
		}
	}

	// Cache the result
	if e.cache != nil {
		err := e.cache.Set(processedQuery, response)
		if err != nil {
			e.logger.Warn("Failed to cache query result", "error", err)
		}

		e.mu.Lock()
		e.stats.CacheMisses++
		e.mu.Unlock()
	}

	// Update statistics
	e.mu.Lock()
	e.stats.SuccessfulQueries++
	e.stats.LastQueryTime = time.Now()
	e.stats.TotalLatency += response.ProcessingTime
	if e.stats.TotalQueries > 0 {
		e.stats.AverageLatency = e.stats.TotalLatency / time.Duration(e.stats.TotalQueries)
	}
	e.mu.Unlock()

	return response, nil
}

// ProcessBatch processes multiple queries in batch
func (e *engine) ProcessBatch(ctx context.Context, queries []*Query) ([]*QueryResponse, error) {
	if len(queries) == 0 {
		return []*QueryResponse{}, nil
	}

	// Limit concurrent queries
	semaphore := make(chan struct{}, e.config.MaxConcurrentQueries)
	responses := make([]*QueryResponse, len(queries))
	errors := make([]error, len(queries))

	var wg sync.WaitGroup
	for i, query := range queries {
		wg.Add(1)
		go func(index int, q *Query) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Process query
			response, err := e.ProcessQuery(ctx, q)
			responses[index] = response
			errors[index] = err
		}(i, query)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("batch processing failed: %w", err)
		}
	}

	return responses, nil
}

// ExpandQuery expands a query for better retrieval
func (e *engine) ExpandQuery(ctx context.Context, query *Query) ([]string, error) {
	var expandedTerms []string

	for _, expander := range e.expanders {
		terms, err := expander.Expand(ctx, query)
		if err != nil {
			e.logger.Warn("Query expansion failed", "expander", expander.Type(), "error", err)
			continue
		}
		expandedTerms = append(expandedTerms, terms...)
	}

	// Limit expansion terms
	if len(expandedTerms) > e.config.MaxExpansionTerms {
		expandedTerms = expandedTerms[:e.config.MaxExpansionTerms]
	}

	return expandedTerms, nil
}

// RerankResults reranks search results
func (e *engine) RerankResults(ctx context.Context, results []*QueryResult, query *Query) ([]*QueryResult, error) {
	if len(results) <= 1 {
		return results, nil
	}

	rerankedResults := results

	for _, reranker := range e.rerankers {
		reranked, err := reranker.Rerank(ctx, rerankedResults, query)
		if err != nil {
			e.logger.Warn("Result reranking failed", "reranker", reranker.Type(), "error", err)
			continue
		}
		rerankedResults = reranked
	}

	return rerankedResults, nil
}

// GetQueryStats returns query processing statistics
func (e *engine) GetQueryStats() QueryStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stats := *e.stats
	return stats
}

// HealthCheck checks if the engine is healthy
func (e *engine) HealthCheck(ctx context.Context) error {
	// Check embedding service
	if m := e.embeddingService.HealthCheck(ctx); m != nil {
		return fmt.Errorf("embedding service unhealthy: %s", m)
	}

	// Check vector index
	if e.vectorIndex == nil {
		return fmt.Errorf("vector index not available")
	}

	return nil
}

// Close closes the engine
func (e *engine) Close() error {
	if e.cache != nil {
		err := e.cache.Close()
		if err != nil {
			return fmt.Errorf("failed to close cache: %w", err)
		}
	}
	return nil
}

// processQuery processes a query through all registered processors
func (e *engine) processQuery(query *Query) (*Query, error) {
	processedQuery := query

	// Sort processors by priority
	e.mu.RLock()
	processors := make([]QueryProcessor, len(e.processors))
	copy(processors, e.processors)
	e.mu.RUnlock()

	// Process through each processor
	for _, processor := range processors {
		processed, err := processor.Process(context.Background(), processedQuery)
		if err != nil {
			return nil, fmt.Errorf("processor %s failed: %w", processor.Type(), err)
		}
		processedQuery = processed
	}

	return processedQuery, nil
}

// registerDefaultComponents registers default query processing components
func (e *engine) registerDefaultComponents() {
	// Register default processors
	e.processors = append(e.processors, &DefaultQueryProcessor{})

	// Register default expanders
	e.expanders = append(e.expanders, &DefaultQueryExpander{})

	// Register default rerankers
	e.rerankers = append(e.rerankers, &DefaultResultReranker{})
}

// DefaultQueryProcessor is a basic query processor
type DefaultQueryProcessor struct{}

func (p *DefaultQueryProcessor) Process(ctx context.Context, query *Query) (*Query, error) {
	// Basic text normalization and validation
	if len(query.Text) > 1000 {
		query.Text = query.Text[:1000]
	}
	return query, nil
}

func (p *DefaultQueryProcessor) Type() string  { return "default" }
func (p *DefaultQueryProcessor) Priority() int { return 100 }

// DefaultQueryExpander is a basic query expander
type DefaultQueryExpander struct{}

func (e *DefaultQueryExpander) Expand(ctx context.Context, query *Query) ([]string, error) {
	// Basic synonym expansion (placeholder)
	return []string{}, nil
}

func (e *DefaultQueryExpander) Type() string        { return "default" }
func (e *DefaultQueryExpander) Confidence() float64 { return 0.5 }

// DefaultResultReranker is a basic result reranker
type DefaultResultReranker struct{}

func (r *DefaultResultReranker) Rerank(ctx context.Context, results []*QueryResult, query *Query) ([]*QueryResult, error) {
	// Basic reranking by relevance score (already sorted by vector search)
	return results, nil
}

func (r *DefaultResultReranker) Type() string        { return "default" }
func (r *DefaultResultReranker) Confidence() float64 { return 0.5 }
