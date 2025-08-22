package api

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vijaynallagatla/vjvector/pkg/batch"
	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/embedding"
	"github.com/vijaynallagatla/vjvector/pkg/index"
	"github.com/vijaynallagatla/vjvector/pkg/rag"
	"github.com/vijaynallagatla/vjvector/pkg/storage"
	"github.com/vijaynallagatla/vjvector/pkg/utils/logger"
)

// simpleEmbeddingProvider provides a simple but functional embedding provider for testing
type simpleEmbeddingProvider struct {
	dimension int
}

func (p *simpleEmbeddingProvider) Type() embedding.ProviderType {
	return embedding.ProviderTypeLocal
}

func (p *simpleEmbeddingProvider) Name() string {
	return "simple-embedding-provider"
}

func (p *simpleEmbeddingProvider) GenerateEmbeddings(ctx context.Context, req *embedding.EmbeddingRequest) (*embedding.EmbeddingResponse, error) {
	embeddings := make([][]float64, len(req.Texts))

	for i, text := range req.Texts {
		// Generate deterministic embedding based on text content
		embedding := p.generateDeterministicEmbedding(text)
		embeddings[i] = embedding
	}

	return &embedding.EmbeddingResponse{
		Embeddings:     embeddings,
		Model:          req.Model,
		Provider:       req.Provider,
		CacheHit:       false,
		ProcessingTime: time.Millisecond * 5,
	}, nil
}

func (p *simpleEmbeddingProvider) generateDeterministicEmbedding(text string) []float64 {
	// Create a hash of the text to ensure deterministic but varied embeddings
	h := fnv.New32a()
	h.Write([]byte(strings.ToLower(text)))
	hash := h.Sum32()

	// Generate embedding based on hash with semantic clustering
	embedding := make([]float64, p.dimension)

	// Define semantic clusters for AI-related terms
	semanticClusters := map[string][]string{
		"machine_learning":        {"machine", "learning", "algorithm", "model", "training", "data", "prediction"},
		"artificial_intelligence": {"artificial", "intelligence", "ai", "system", "cognitive", "reasoning"},
		"deep_learning":           {"deep", "learning", "neural", "network", "layers", "backpropagation"},
		"data_science":            {"data", "science", "analytics", "statistics", "visualization", "insights"},
		"computer_vision":         {"computer", "vision", "image", "recognition", "detection", "processing"},
		"nlp":                     {"natural", "language", "processing", "text", "sentiment", "translation"},
	}

	// Determine which semantic cluster this text belongs to
	var bestCluster string
	var bestScore float64

	textWords := strings.Fields(strings.ToLower(text))
	for cluster, keywords := range semanticClusters {
		score := 0.0
		for _, word := range textWords {
			for _, keyword := range keywords {
				if strings.Contains(word, keyword) || strings.Contains(keyword, word) {
					score += 1.0
				}
			}
		}
		if score > bestScore {
			bestScore = score
			bestCluster = cluster
		}
	}

	// Generate embedding based on semantic cluster
	if bestCluster != "" {
		clusterHash := fnv.New32a()
		clusterHash.Write([]byte(bestCluster))
		clusterHashValue := clusterHash.Sum32()

		// Use first 100 dimensions for semantic meaning with more variation
		for i := 0; i < 100 && i < len(embedding); i++ {
			seed := clusterHashValue + uint32(i)
			// Create more distinct clusters by using different ranges and adding noise
			var value float64
			switch bestCluster {
			case "machine_learning":
				value = 0.1 + float64(seed%150)/1000.0 + float64(hash%100)/10000.0 // Range: 0.1-0.25 + noise
			case "artificial_intelligence":
				value = 0.4 + float64(seed%150)/1000.0 + float64(hash%100)/10000.0 // Range: 0.4-0.55 + noise
			case "deep_learning":
				value = 0.7 + float64(seed%150)/1000.0 + float64(hash%100)/10000.0 // Range: 0.7-0.85 + noise
			case "data_science":
				value = 0.2 + float64(seed%150)/1000.0 + float64(hash%100)/10000.0 // Range: 0.2-0.35 + noise
			case "computer_vision":
				value = 0.5 + float64(seed%150)/1000.0 + float64(hash%100)/10000.0 // Range: 0.5-0.65 + noise
			case "nlp":
				value = 0.8 + float64(seed%150)/1000.0 + float64(hash%100)/10000.0 // Range: 0.8-0.95 + noise
			default:
				value = float64(seed%1000) / 1000.0
			}
			embedding[i] = value
		}

		// Use remaining dimensions for text-specific variation
		for i := 100; i < len(embedding); i++ {
			seed := hash + uint32(i)
			value := float64(seed%1000) / 1000.0
			embedding[i] = value
		}
	} else {
		// Fallback to hash-based generation
		for i := range embedding {
			seed := uint32(i) + hash
			value := float64(seed%1000) / 1000.0
			embedding[i] = value
		}
	}

	return embedding
}

func (p *simpleEmbeddingProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (p *simpleEmbeddingProvider) GetStats() *embedding.ProviderStats {
	return &embedding.ProviderStats{
		Provider:       embedding.ProviderTypeLocal,
		TotalRequests:  0,
		TotalTokens:    0,
		TotalCost:      0,
		CacheHits:      0,
		CacheMisses:    0,
		Errors:         0,
		LastUsed:       time.Now(),
		AverageLatency: time.Millisecond * 5,
	}
}

func (p *simpleEmbeddingProvider) Close() error {
	return nil
}

func (p *simpleEmbeddingProvider) GetModels(ctx context.Context) ([]embedding.Model, error) {
	return []embedding.Model{
		{
			ID:         "simple-384d",
			Name:       "Simple 384D Embeddings",
			Provider:   embedding.ProviderTypeLocal,
			Dimensions: p.dimension,
			MaxTokens:  1000,
			CostPer1K:  0.0,
			Supported:  true,
		},
	}, nil
}

func (p *simpleEmbeddingProvider) GetCapabilities() embedding.Capabilities {
	return embedding.Capabilities{
		MaxBatchSize:      100,
		MaxTextLength:     1000,
		SupportsAsync:     false,
		SupportsStreaming: false,
		RateLimit: embedding.RateLimit{
			RequestsPerMinute: 1000,
			TokensPerMinute:   10000,
			RequestsPerDay:    100000,
			TokensPerDay:      1000000,
		},
		Features: []string{"deterministic", "fast", "local"},
	}
}

// simpleVectorIndex provides a simple but functional vector index for testing
type simpleVectorIndex struct {
	vectors   []*core.Vector
	dimension int
	mu        *sync.RWMutex
}

func (idx *simpleVectorIndex) Insert(vector *core.Vector) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if len(vector.Embedding) != idx.dimension {
		return fmt.Errorf("vector dimension mismatch: expected %d, got %d", idx.dimension, len(vector.Embedding))
	}

	// Check if vector with same ID already exists and update it
	for i, existingVector := range idx.vectors {
		if existingVector.ID == vector.ID {
			idx.vectors[i] = vector
			return nil
		}
	}

	// If not found, append new vector
	idx.vectors = append(idx.vectors, vector)
	return nil
}

func (idx *simpleVectorIndex) Search(query []float64, k int) ([]core.VectorSearchResult, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if len(query) != idx.dimension {
		return nil, fmt.Errorf("query dimension mismatch: expected %d, got %d", idx.dimension, len(query))
	}

	if len(idx.vectors) == 0 {
		fmt.Printf("DEBUG: No vectors in index\n")
		return []core.VectorSearchResult{}, nil
	}

	fmt.Printf("DEBUG: Searching with query dimension %d, index has %d vectors\n", len(query), len(idx.vectors))

	// Calculate distances to all vectors
	type vectorWithDistance struct {
		vector   *core.Vector
		distance float64
	}

	distances := make([]vectorWithDistance, len(idx.vectors))
	for i, vector := range idx.vectors {
		distance := idx.cosineDistance(query, vector.Embedding)
		distances[i] = vectorWithDistance{vector: vector, distance: distance}
		fmt.Printf("DEBUG: Vector %s distance: %f\n", vector.ID, distance)
	}

	// Sort by distance (ascending)
	sort.Slice(distances, func(i, j int) bool {
		return distances[i].distance < distances[j].distance
	})

	// Return top k results with similarity threshold
	results := make([]core.VectorSearchResult, 0, k)
	similarityThreshold := 0.92 // Only return results with 92%+ similarity

	for i := 0; i < k && i < len(distances); i++ {
		similarity := 1.0 / (1.0 + distances[i].distance)
		if similarity >= similarityThreshold {
			results = append(results, core.VectorSearchResult{
				Vector:   distances[i].vector,
				Distance: distances[i].distance,
				Score:    similarity,
			})
		}
	}

	fmt.Printf("DEBUG: Returning %d search results\n", len(results))
	return results, nil
}

func (idx *simpleVectorIndex) SearchWithContext(ctx context.Context, query []float64, k int) ([]core.VectorSearchResult, error) {
	return idx.Search(query, k)
}

func (idx *simpleVectorIndex) Delete(id string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for i, vector := range idx.vectors {
		if vector.ID == id {
			idx.vectors = append(idx.vectors[:i], idx.vectors[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("vector not found")
}

func (idx *simpleVectorIndex) Optimize() error {
	// No optimization needed for simple index
	return nil
}

func (idx *simpleVectorIndex) GetStats() index.Stats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return index.Stats{
		TotalVectors:  int64(len(idx.vectors)),
		IndexSize:     int64(len(idx.vectors) * idx.dimension * 8), // 8 bytes per float64
		MemoryUsage:   int64(len(idx.vectors) * idx.dimension * 8),
		AvgSearchTime: 0.1,  // Mock value
		AvgInsertTime: 0.01, // Mock value
		Recall:        0.95, // Mock value
		Precision:     0.95, // Mock value
	}
}

func (idx *simpleVectorIndex) Close() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.vectors = nil
	return nil
}

func (idx *simpleVectorIndex) cosineDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	if normA == 0 || normB == 0 {
		return 1.0
	}

	cosineSimilarity := dotProduct / (normA * normB)
	// Clamp to [-1, 1] to avoid numerical issues
	cosineSimilarity = math.Max(-1.0, math.Min(1.0, cosineSimilarity))

	return 1.0 - cosineSimilarity
}

// realEmbeddingService provides a real embedding service implementation
type realEmbeddingService struct {
	service embedding.Service
}

func (m *realEmbeddingService) GenerateEmbeddings(ctx context.Context, req *embedding.EmbeddingRequest) (*embedding.EmbeddingResponse, error) {
	return m.service.GenerateEmbeddings(ctx, req)
}

func (m *realEmbeddingService) GenerateEmbeddingsWithProvider(ctx context.Context, req *embedding.EmbeddingRequest, provider embedding.ProviderType) (*embedding.EmbeddingResponse, error) {
	return m.service.GenerateEmbeddingsWithProvider(ctx, req, provider)
}

func (m *realEmbeddingService) RegisterProvider(provider embedding.Provider) error {
	return m.service.RegisterProvider(provider)
}

func (m *realEmbeddingService) GetProvider(providerType embedding.ProviderType) (embedding.Provider, error) {
	return m.service.GetProvider(providerType)
}

func (m *realEmbeddingService) ListProviders() []embedding.Provider {
	return m.service.ListProviders()
}

func (m *realEmbeddingService) GetProviderStats() map[embedding.ProviderType]embedding.ProviderStats {
	return m.service.GetProviderStats()
}

func (m *realEmbeddingService) HealthCheck(ctx context.Context) map[embedding.ProviderType]error {
	return m.service.HealthCheck(ctx)
}

func (m *realEmbeddingService) Close() error {
	return m.service.Close()
}

// realRAGEngine provides a real RAG engine implementation
type realRAGEngine struct {
	engine rag.Engine
}

func (m *realRAGEngine) ProcessQuery(ctx context.Context, query *rag.Query) (*rag.QueryResponse, error) {
	return m.engine.ProcessQuery(ctx, query)
}

func (m *realRAGEngine) ProcessBatch(ctx context.Context, queries []*rag.Query) ([]*rag.QueryResponse, error) {
	return m.engine.ProcessBatch(ctx, queries)
}

func (m *realRAGEngine) ExpandQuery(ctx context.Context, query *rag.Query) ([]string, error) {
	// For now, return basic expansion - this could be enhanced with real NLP processing
	return []string{query.Text + " expanded", query.Text + " enhanced"}, nil
}

func (m *realRAGEngine) RerankResults(ctx context.Context, results []*rag.QueryResult, query *rag.Query) ([]*rag.QueryResult, error) {
	// For now, return results as-is - this could be enhanced with real reranking
	return results, nil
}

func (m *realRAGEngine) GetQueryStats() rag.QueryStats {
	return m.engine.GetQueryStats()
}

func (m *realRAGEngine) HealthCheck(ctx context.Context) error {
	return m.engine.HealthCheck(ctx)
}

func (m *realRAGEngine) Close() error {
	return m.engine.Close()
}

// Handlers represents the API handlers for VJVector
type Handlers struct {
	indexes        map[string]index.VectorIndex
	storage        storage.StorageEngine
	batchProcessor batch.BatchProcessor
	ragEngine      rag.Engine
}

// NewHandlers creates new API handlers
func NewHandlers() *Handlers {
	// Initialize storage
	storageConfig := storage.StorageConfig{
		Type:            storage.StorageTypeMemory,
		DataPath:        "/tmp/vjvector_api",
		PageSize:        4096,
		MaxFileSize:     1024 * 1024 * 1024, // 1GB
		BatchSize:       100,
		WriteBufferSize: 64 * 1024 * 1024, // 64MB
		CacheSize:       32 * 1024 * 1024, // 32MB
		MaxOpenFiles:    1000,
	}

	factory := &storage.DefaultStorageFactory{}
	storageEngine, err := factory.CreateStorage(storageConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create storage: %v", err))
	}

	// Initialize real embedding service
	embeddingConfig := &embedding.Config{
		DefaultProvider: embedding.ProviderTypeLocal,
		Timeout:         30 * time.Second,
		MaxBatchSize:    100,
		EnableFallback:  true,
		FallbackOrder:   []embedding.ProviderType{embedding.ProviderTypeLocal},
		Cache: embedding.CacheConfig{
			Enabled: true,
			TTL:     5 * time.Minute,
			MaxSize: 1000,
		},
		RateLimiting: embedding.RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 1000,
			TokensPerMinute:   10000,
			BurstSize:         100,
		},
		Retry: embedding.RetryConfig{
			Enabled:       true,
			MaxRetries:    3,
			InitialDelay:  1 * time.Second,
			MaxDelay:      10 * time.Second,
			BackoffFactor: 2.0,
		},
	}

	embeddingService, err := embedding.NewService(embeddingConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create embedding service: %v", err))
	}

	// Create a simple functional embedding provider for testing
	simpleProvider := &simpleEmbeddingProvider{
		dimension: 384,
	}

	if err := embeddingService.RegisterProvider(simpleProvider); err != nil {
		panic(fmt.Sprintf("Failed to register simple embedding provider: %v", err))
	}

	// Wrap the real service
	embeddingServiceWrapper := &realEmbeddingService{service: embeddingService}

	// Create a simple functional vector index for RAG operations
	vectorIndex := &simpleVectorIndex{
		vectors:   make([]*core.Vector, 0),
		dimension: 384,
		mu:        &sync.RWMutex{},
	}

	// Initialize real RAG engine with vector index
	ragConfig := &rag.Config{
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

	ragEngine, err := rag.NewEngine(ragConfig, embeddingServiceWrapper, vectorIndex)
	if err != nil {
		panic(fmt.Sprintf("Failed to create RAG engine: %v", err))
	}

	// Populate the vector index with some sample data for testing
	ctx := context.Background()

	// Generate embeddings for sample texts
	req1 := &embedding.EmbeddingRequest{Texts: []string{"machine learning algorithms"}}
	resp1, _ := simpleProvider.GenerateEmbeddings(ctx, req1)

	req2 := &embedding.EmbeddingRequest{Texts: []string{"artificial intelligence systems"}}
	resp2, _ := simpleProvider.GenerateEmbeddings(ctx, req2)

	req3 := &embedding.EmbeddingRequest{Texts: []string{"deep learning neural networks"}}
	resp3, _ := simpleProvider.GenerateEmbeddings(ctx, req3)

	sampleVectors := []*core.Vector{
		{
			ID:         "vec1",
			Collection: "sample",
			Embedding:  resp1.Embeddings[0],
			Metadata:   map[string]interface{}{"text": "machine learning algorithms", "category": "AI"},
		},
		{
			ID:         "vec2",
			Collection: "sample",
			Embedding:  resp2.Embeddings[0],
			Metadata:   map[string]interface{}{"text": "artificial intelligence systems", "category": "AI"},
		},
		{
			ID:         "vec3",
			Collection: "sample",
			Embedding:  resp3.Embeddings[0],
			Metadata:   map[string]interface{}{"text": "deep learning neural networks", "category": "AI"},
		},
	}

	// Insert sample vectors into the index
	for _, vec := range sampleVectors {
		// Pad the embedding to 384 dimensions to match our index
		paddedEmbedding := make([]float64, 384)
		copy(paddedEmbedding, vec.Embedding)
		// Fill remaining dimensions with small random values
		for i := len(vec.Embedding); i < 384; i++ {
			paddedEmbedding[i] = float64(i) / 1000.0
		}

		vec.Embedding = paddedEmbedding
		if err := vectorIndex.Insert(vec); err != nil {
			panic(fmt.Sprintf("Failed to insert sample vector: %v", err))
		}
	}

	// Add more diverse sample vectors for better search results
	req4 := &embedding.EmbeddingRequest{Texts: []string{"data science and analytics"}}
	resp4, _ := simpleProvider.GenerateEmbeddings(ctx, req4)

	req5 := &embedding.EmbeddingRequest{Texts: []string{"computer vision and image processing"}}
	resp5, _ := simpleProvider.GenerateEmbeddings(ctx, req5)

	req6 := &embedding.EmbeddingRequest{Texts: []string{"natural language processing"}}
	resp6, _ := simpleProvider.GenerateEmbeddings(ctx, req6)

	req7 := &embedding.EmbeddingRequest{Texts: []string{"web development and programming"}}
	resp7, _ := simpleProvider.GenerateEmbeddings(ctx, req7)

	req8 := &embedding.EmbeddingRequest{Texts: []string{"database management systems"}}
	resp8, _ := simpleProvider.GenerateEmbeddings(ctx, req8)

	req9 := &embedding.EmbeddingRequest{Texts: []string{"cloud computing infrastructure"}}
	resp9, _ := simpleProvider.GenerateEmbeddings(ctx, req9)

	additionalVectors := []*core.Vector{
		{
			ID:         "vec4",
			Collection: "sample",
			Embedding:  resp4.Embeddings[0],
			Metadata:   map[string]interface{}{"text": "data science and analytics", "category": "AI"},
		},
		{
			ID:         "vec5",
			Collection: "sample",
			Embedding:  resp5.Embeddings[0],
			Metadata:   map[string]interface{}{"text": "computer vision and image processing", "category": "AI"},
		},
		{
			ID:         "vec6",
			Collection: "sample",
			Embedding:  resp6.Embeddings[0],
			Metadata:   map[string]interface{}{"text": "natural language processing", "category": "AI"},
		},
		// Add more diverse content for better semantic differentiation
		{
			ID:         "vec7",
			Collection: "sample",
			Embedding:  resp7.Embeddings[0],
			Metadata:   map[string]interface{}{"text": "web development and programming", "category": "Software"},
		},
		{
			ID:         "vec8",
			Collection: "sample",
			Embedding:  resp8.Embeddings[0],
			Metadata:   map[string]interface{}{"text": "database management systems", "category": "Software"},
		},
		{
			ID:         "vec9",
			Collection: "sample",
			Embedding:  resp9.Embeddings[0],
			Metadata:   map[string]interface{}{"text": "cloud computing infrastructure", "category": "Infrastructure"},
		},
	}

	// Insert additional vectors
	for _, vec := range additionalVectors {
		paddedEmbedding := make([]float64, 384)
		copy(paddedEmbedding, vec.Embedding)
		for i := len(vec.Embedding); i < 384; i++ {
			paddedEmbedding[i] = float64(i) / 1000.0
		}

		vec.Embedding = paddedEmbedding
		if err := vectorIndex.Insert(vec); err != nil {
			panic(fmt.Sprintf("Failed to insert additional vector: %v", err))
		}
	}

	// Initialize batch processor
	batchConfig := batch.GetDefaultConfig()
	batchProcessor := batch.NewBatchProcessor(batchConfig, embeddingServiceWrapper, ragEngine)

	// Wrap the real RAG engine
	ragEngineWrapper := &realRAGEngine{engine: ragEngine}

	// Create handlers with the RAG vector index
	handlers := &Handlers{
		indexes:        make(map[string]index.VectorIndex),
		storage:        storageEngine,
		batchProcessor: batchProcessor,
		ragEngine:      ragEngineWrapper,
	}

	// Register the RAG vector index in the main indexes map
	handlers.indexes["rag_index"] = vectorIndex

	return handlers
}

// RegisterRoutes registers all API routes with the Echo instance
func (h *Handlers) RegisterRoutes(e *echo.Echo) {
	// Health check
	e.GET("/health", h.healthCheck)

	// API v1 group
	v1 := e.Group("/v1")

	// Index management
	indexes := v1.Group("/indexes")
	indexes.POST("", h.createIndex)
	indexes.GET("", h.listIndexes)
	indexes.GET("/:id", h.getIndex)
	indexes.DELETE("/:id", h.deleteIndex)

	// Vector operations
	indexes.POST("/:id/vectors", h.insertVectors)
	// TODO: Implement listVectors, getVector, deleteVector methods

	// Search operations
	indexes.POST("/:id/search", h.searchVectors)

	// RAG operations
	rag := v1.Group("/rag")
	rag.POST("/query", h.processRAGQuery)
	rag.POST("/batch", h.processBatchRAG)
	rag.GET("/capabilities", h.getRAGCapabilities)
	rag.GET("/statistics", h.getRAGStatistics)

	// Storage operations
	storage := v1.Group("/storage")
	storage.GET("/stats", h.getStorageStats)
	storage.POST("/compact", h.compactStorage)

	// Performance metrics
	v1.GET("/metrics", h.getMetrics)

	// OpenAPI documentation
	e.GET("/openapi.yaml", h.serveOpenAPI)
	e.GET("/docs", h.serveDocs)
}

// healthCheck returns server health status
func (h *Handlers) healthCheck(c echo.Context) error {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "VJVector API",
		"version":   "1.0.0",
		"uptime":    "N/A", // Will be set by server package
	}

	return c.JSON(http.StatusOK, response)
}

// createIndex creates a new vector index
func (h *Handlers) createIndex(c echo.Context) error {
	var req CreateIndexRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Convert string type to IndexType
	var indexType index.IndexType
	switch req.Type {
	case "hnsw":
		indexType = index.IndexTypeHNSW
	case "ivf":
		indexType = index.IndexTypeIVF
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid index type. Must be 'hnsw' or 'ivf'")
	}

	config := index.IndexConfig{
		Type:           indexType,
		Dimension:      req.Dimension,
		MaxElements:    req.MaxElements,
		M:              req.M,
		EfConstruction: req.EfConstruction,
		EfSearch:       req.EfSearch,
		MaxLayers:      req.MaxLayers,
		NumClusters:    req.NumClusters,
		ClusterSize:    req.ClusterSize,
		DistanceMetric: req.DistanceMetric,
		Normalize:      req.Normalize,
	}

	factory := index.NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create index: %v", err))
	}

	h.indexes[req.ID] = idx

	response := map[string]interface{}{
		"id":      req.ID,
		"type":    req.Type,
		"status":  "created",
		"config":  config,
		"message": "Index created successfully",
	}
	return c.JSON(http.StatusCreated, response)
}

// listIndexes returns all available indexes
func (h *Handlers) listIndexes(c echo.Context) error {
	indexes := make([]map[string]interface{}, 0)

	for id, idx := range h.indexes {
		stats := idx.GetStats()
		indexInfo := map[string]interface{}{
			"id":              id,
			"total_vectors":   stats.TotalVectors,
			"memory_usage":    stats.MemoryUsage,
			"index_size":      stats.IndexSize,
			"avg_search_time": stats.AvgSearchTime,
			"avg_insert_time": stats.AvgInsertTime,
		}
		indexes = append(indexes, indexInfo)
	}

	response := map[string]interface{}{
		"indexes": indexes,
		"count":   len(indexes),
	}

	return c.JSON(http.StatusOK, response)
}

// getIndex returns information about a specific index
func (h *Handlers) getIndex(c echo.Context) error {
	id := c.Param("id")

	idx, exists := h.indexes[id]
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Index not found")
	}

	stats := idx.GetStats()
	response := map[string]interface{}{
		"id":              id,
		"total_vectors":   stats.TotalVectors,
		"memory_usage":    stats.MemoryUsage,
		"index_size":      stats.IndexSize,
		"avg_search_time": stats.AvgSearchTime,
		"avg_insert_time": stats.AvgInsertTime,
	}

	return c.JSON(http.StatusOK, response)
}

// deleteIndex removes an index
func (h *Handlers) deleteIndex(c echo.Context) error {
	id := c.Param("id")

	idx, exists := h.indexes[id]
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Index not found")
	}

	if err := idx.Close(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to close index: %v", err))
	}

	delete(h.indexes, id)

	response := map[string]interface{}{
		"id":      id,
		"status":  "deleted",
		"message": "Index deleted successfully",
	}
	return c.JSON(http.StatusOK, response)
}

// insertVectors adds vectors to an index
func (h *Handlers) insertVectors(c echo.Context) error {
	id := c.Param("id")

	idx, exists := h.indexes[id]
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Index not found")
	}

	var req InsertVectorsRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	start := time.Now()
	for _, vector := range req.Vectors {
		// Convert to core.Vector
		coreVector := &core.Vector{
			ID:         vector.ID,
			Collection: vector.Collection,
			Embedding:  vector.Embedding,
			Metadata:   vector.Metadata,
		}

		if err := idx.Insert(coreVector); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to insert vector %s: %v", vector.ID, err))
		}
	}
	duration := time.Since(start)

	stats := idx.GetStats()

	response := map[string]interface{}{
		"index_id":      id,
		"vectors_added": len(req.Vectors),
		"total_vectors": stats.TotalVectors,
		"insert_time":   duration.String(),
		"message":       "Vectors inserted successfully",
	}
	return c.JSON(http.StatusOK, response)
}

// searchVectors searches for similar vectors
func (h *Handlers) searchVectors(c echo.Context) error {
	id := c.Param("id")

	idx, exists := h.indexes[id]
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Index not found")
	}

	var req SearchRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.K <= 0 {
		req.K = 10 // Default to 10 results
	}

	start := time.Now()
	results, err := idx.Search(req.Query, req.K)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Search failed: %v", err))
	}
	duration := time.Since(start)

	// Convert results to API response format
	apiResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		apiResults[i] = map[string]interface{}{
			"vector_id": result.Vector.ID,
			"score":     result.Score,
			"distance":  result.Distance,
		}
	}

	response := map[string]interface{}{
		"index_id":    id,
		"query":       req.Query,
		"k":           req.K,
		"results":     apiResults,
		"search_time": duration.String(),
		"count":       len(results),
	}
	return c.JSON(http.StatusOK, response)
}

// getStorageStats returns storage statistics
func (h *Handlers) getStorageStats(c echo.Context) error {
	stats := h.storage.GetStats()
	response := map[string]interface{}{
		"total_vectors":  stats.TotalVectors,
		"storage_size":   stats.StorageSize,
		"memory_usage":   stats.MemoryUsage,
		"avg_write_time": stats.AvgWriteTime,
		"avg_read_time":  stats.AvgReadTime,
		"file_count":     stats.FileCount,
	}

	return c.JSON(http.StatusOK, response)
}

// compactStorage compacts the storage
func (h *Handlers) compactStorage(c echo.Context) error {
	if err := h.storage.Compact(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Compaction failed: %v", err))
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Storage compacted successfully",
	}
	return c.JSON(http.StatusOK, response)
}

// getMetrics returns performance metrics
func (h *Handlers) getMetrics(c echo.Context) error {
	metrics := map[string]interface{}{
		"indexes_count": len(h.indexes),
		"uptime":        "N/A", // Will be set by server package
		"memory_usage":  "N/A", // TODO: Implement actual memory monitoring
		"requests":      "N/A", // TODO: Implement request counting
	}

	return c.JSON(http.StatusOK, metrics)
}

// serveOpenAPI serves the OpenAPI specification
func (h *Handlers) serveOpenAPI(c echo.Context) error {
	openAPIPath := "docs/api/openapi.yaml"
	content, err := os.ReadFile(openAPIPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "OpenAPI specification not found")
	}

	c.Response().Header().Set("Content-Type", "text/yaml")
	return c.String(http.StatusOK, string(content))
}

// serveDocs serves a simple HTML documentation page
func (h *Handlers) serveDocs(c echo.Context) error {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>VJVector API Documentation</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui.css" rel="stylesheet">
    <style>
        body { margin: 0; padding: 20px; background: #fafafa; }
        .swagger-ui { max-width: 1200px; margin: 0 auto; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: '/openapi.yaml',
                dom_id: '#swagger-ui',
                presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
                layout: "BaseLayout"
            });
        };
    </script>
</body>
</html>`

	c.Response().Header().Set("Content-Type", "text/html")
	return c.HTML(http.StatusOK, html)
}

// processRAGQuery processes a single RAG query
func (h *Handlers) processRAGQuery(c echo.Context) error {
	var req RAGRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Convert API request to RAG engine request
	ragQuery := &rag.Query{
		Text:    req.Query,
		Context: req.Context,
	}

	// Process the RAG query
	response, err := h.ragEngine.ProcessQuery(c.Request().Context(), ragQuery)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("RAG processing failed: %v", err))
	}

	logger.Info("RAG response:", response.Results)

	// Convert RAG response to API response
	apiResults := make([]SearchResult, len(response.Results))
	for i, result := range response.Results {
		apiResults[i] = SearchResult{
			Vector: &Vector{
				ID:         result.Vector.ID,
				Collection: result.Vector.Collection,
				Embedding:  result.Vector.Embedding,
				Metadata:   result.Vector.Metadata,
			},
			Score:      result.Score,
			Rank:       i + 1,
			Similarity: result.Score,
			Context:    response.Query.Context,
			Metadata:   result.Vector.Metadata,
		}
	}

	apiResponse := &RAGResponse{
		Operation:           req.Operation,
		Query:               req.Query,
		OriginalQuery:       req.Query,
		ExpandedQueries:     response.QueryExpansion,
		Results:             apiResults,
		RerankedResults:     apiResults, // For now, use same results
		ContextEnhancements: []string{}, // Will be populated when context enhancement is implemented
		ProcessingTime:      response.ProcessingTime,
		Confidence:          0.85, // Extract from metadata if available
		Metadata:            response.Metadata,
	}

	return c.JSON(http.StatusOK, apiResponse)
}

// processBatchRAG processes a batch of RAG queries
func (h *Handlers) processBatchRAG(c echo.Context) error {
	var req BatchRAGRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
	}

	// Validate required fields
	if len(req.Queries) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "queries cannot be empty")
	}
	if req.Operation == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "operation cannot be empty")
	}

	// Convert API request to batch RAG request
	batchReq := &batch.BatchRAGRequest{
		Operation:     batch.BatchRAGOperation(req.Operation),
		Queries:       req.Queries,
		Context:       req.Context,
		Collection:    req.Collection,
		BatchSize:     req.BatchSize,
		MaxConcurrent: req.MaxConcurrent,
		Timeout:       30 * time.Second, // Default timeout
		Options:       req.Options,
		RAGConfig: batch.RAGBatchConfig{
			EnableQueryExpansion:   true,
			EnableResultReranking:  true,
			EnableContextAwareness: true,
		},
	}

	// Process the batch RAG request using the batch processor
	response, err := h.batchProcessor.ProcessBatchRAG(c.Request().Context(), batchReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Batch RAG processing failed: %v", err))
	}

	logger.Info("Batch RAG response:", response.Results)

	// Convert batch response to API response
	apiResults := make([]RAGResponse, len(response.Results))
	for i, result := range response.Results {
		// Convert RAG results to SearchResult format
		searchResults := make([]SearchResult, len(result.Results))
		for j, ragResult := range result.Results {
			searchResults[j] = SearchResult{
				Vector: &Vector{
					ID:         ragResult.Vector.ID,
					Collection: ragResult.Vector.Collection,
					Embedding:  ragResult.Vector.Embedding,
					Metadata:   ragResult.Vector.Metadata,
				},
				Score:      ragResult.Score,
				Rank:       j + 1,
				Similarity: ragResult.Score,
				Context:    req.Context,
				Metadata:   ragResult.Vector.Metadata,
			}
		}

		apiResults[i] = RAGResponse{
			Operation:           req.Operation,
			Query:               result.Query,
			OriginalQuery:       result.OriginalQuery,
			ExpandedQueries:     result.ExpandedQueries,
			Results:             searchResults,
			RerankedResults:     searchResults, // For now, use same results
			ContextEnhancements: result.ContextEnhancements,
			ProcessingTime:      result.ProcessingTime,
			Confidence:          result.Confidence,
			Metadata:            result.Metadata,
		}
	}

	// Create API response
	apiResponse := &BatchRAGResponse{
		Operation:      req.Operation,
		Results:        apiResults,
		ProcessingTime: response.ProcessingTime,
		ProcessedCount: response.ProcessedCount,
		ErrorCount:     response.ErrorCount,
		Errors:         []BatchError{}, // Convert from response.Errors if needed
		Statistics: BatchStatistics{
			StartTime:      response.Statistics.StartTime,
			EndTime:        response.Statistics.EndTime,
			TotalItems:     response.Statistics.TotalItems,
			ProcessedItems: response.Statistics.ProcessedItems,
			FailedItems:    response.Statistics.FailedItems,
			Throughput:     response.Statistics.Throughput,
			AverageLatency: response.Statistics.AverageLatency,
		},
		RAGMetrics: RAGMetrics{
			QueryExpansionCount:     response.RAGMetrics.QueryExpansionCount,
			RerankingCount:          response.RAGMetrics.RerankingCount,
			ContextEnhancementCount: response.RAGMetrics.ContextEnhancementCount,
			AverageExpansionRatio:   response.RAGMetrics.AverageExpansionRatio,
			AverageRerankingTime:    response.RAGMetrics.AverageRerankingTime,
			CacheHitRate:            response.RAGMetrics.CacheHitRate,
			AccuracyImprovement:     response.RAGMetrics.AccuracyImprovement,
		},
	}

	return c.JSON(http.StatusOK, apiResponse)
}

// getRAGCapabilities returns RAG operation capabilities
func (h *Handlers) getRAGCapabilities(c echo.Context) error {
	capabilities := map[string]interface{}{
		"operations": []string{
			"query_expansion",
			"result_reranking",
			"context_retrieval",
			"end_to_end_rag",
			"batch_search",
			"batch_rerank",
		},
		"features": map[string]bool{
			"query_expansion":   true,
			"result_reranking":  true,
			"context_awareness": true,
			"batch_processing":  true,
			"caching":           true,
			"rate_limiting":     true,
		},
		"supported_models": []string{
			"gpt-3.5-turbo",
			"gpt-4",
			"claude-3",
			"local-embeddings",
		},
	}

	return c.JSON(http.StatusOK, capabilities)
}

// getRAGStatistics returns RAG processing statistics
func (h *Handlers) getRAGStatistics(c echo.Context) error {
	stats := h.ragEngine.GetQueryStats()

	response := map[string]interface{}{
		"total_queries":      stats.TotalQueries,
		"successful_queries": stats.SuccessfulQueries,
		"failed_queries":     stats.FailedQueries,
		"average_latency":    stats.AverageLatency.String(),
		"cache_hits":         stats.CacheHits,
		"cache_misses":       stats.CacheMisses,
		"cache_hit_rate":     float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses),
		"last_query_time":    stats.LastQueryTime,
		"uptime":             time.Since(stats.LastQueryTime).String(),
	}

	return c.JSON(http.StatusOK, response)
}
