package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/batch"
	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/embedding"
	"github.com/vijaynallagatla/vjvector/pkg/index"
	"github.com/vijaynallagatla/vjvector/pkg/metrics"
	"github.com/vijaynallagatla/vjvector/pkg/rag"
	"github.com/vijaynallagatla/vjvector/pkg/storage"
)

// simpleEmbeddingProvider moved to embedding_services.go

// simpleVectorIndex moved to vector_index_impl.go

// realEmbeddingService moved to embedding_services.go
// realRAGEngine moved to embedding_services.go

// Handlers represents the API handlers for VJVector
type Handlers struct {
	indexes        map[string]index.VectorIndex
	storage        storage.StorageEngine
	batchProcessor batch.BatchProcessor
	ragEngine      rag.Engine
	server         ServerInterface // Interface for accessing server metrics
}

// ServerInterface defines methods for accessing server functionality
type ServerInterface interface {
	Metrics() *metrics.PrometheusMetrics
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

// SetServer sets the server interface for accessing metrics
func (h *Handlers) SetServer(server ServerInterface) {
	h.server = server
}
