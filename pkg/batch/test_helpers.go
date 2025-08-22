package batch

import (
	"context"
	"fmt"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/embedding"
)

// mockEmbeddingService provides a mock implementation of the embedding service
type mockEmbeddingService struct{}

func (m *mockEmbeddingService) GenerateEmbeddings(ctx context.Context, req *embedding.EmbeddingRequest) (*embedding.EmbeddingResponse, error) {
	// Generate mock embeddings
	embeddings := make([][]float64, len(req.Texts))
	for i := range req.Texts {
		// Create a simple embedding based on text length
		embedding := make([]float64, 128)
		for j := range embedding {
			embedding[j] = float64(len(req.Texts[i])%10) / 10.0
		}
		embeddings[i] = embedding
	}

	return &embedding.EmbeddingResponse{
		Embeddings: embeddings,
		Model:      req.Model,
		Provider:   req.Provider,
		Usage: embedding.UsageStats{
			TotalTokens: len(req.Texts) * 10,
		},
		ProcessingTime: 10 * time.Millisecond,
	}, nil
}

func (m *mockEmbeddingService) GenerateEmbeddingsWithProvider(ctx context.Context, req *embedding.EmbeddingRequest, provider embedding.ProviderType) (*embedding.EmbeddingResponse, error) {
	return m.GenerateEmbeddings(ctx, req)
}

func (m *mockEmbeddingService) RegisterProvider(provider embedding.Provider) error {
	return nil
}

func (m *mockEmbeddingService) GetProvider(providerType embedding.ProviderType) (embedding.Provider, error) {
	return nil, nil
}

func (m *mockEmbeddingService) ListProviders() []embedding.Provider {
	return nil
}

func (m *mockEmbeddingService) GetProviderStats() map[embedding.ProviderType]embedding.ProviderStats {
	return nil
}

func (m *mockEmbeddingService) HealthCheck(ctx context.Context) map[embedding.ProviderType]error {
	return nil
}

func (m *mockEmbeddingService) Close() error {
	return nil
}

// Helper functions for tests and benchmarks

func generateTestTexts(count int) []string {
	texts := make([]string, count)
	for i := 0; i < count; i++ {
		texts[i] = fmt.Sprintf("test text %d with some content", i)
	}
	return texts
}

func generateTestVectors(count int) []*core.Vector {
	vectors := make([]*core.Vector, count)
	for i := 0; i < count; i++ {
		vectors[i] = &core.Vector{
			ID:         fmt.Sprintf("vector-%d", i),
			Collection: "test-collection",
			Embedding:  generateTestEmbedding(128),
			Metadata:   map[string]interface{}{"index": i},
			Text:       fmt.Sprintf("vector text %d", i),
			Dimension:  128,
			Magnitude:  1.0,
			Normalized: false,
		}
	}
	return vectors
}

func generateTestEmbedding(dimension int) []float64 {
	embedding := make([]float64, dimension)
	for i := 0; i < dimension; i++ {
		embedding[i] = float64(i%10) / 10.0
	}
	return embedding
}
