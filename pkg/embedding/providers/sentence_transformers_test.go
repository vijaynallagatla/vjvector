package providers

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vijaynallagatla/vjvector/pkg/embedding"
)

func TestNewSentenceTransformersProvider(t *testing.T) {
	tests := []struct {
		name    string
		config  SentenceTransformersConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: SentenceTransformersConfig{
				ModelPath: "./test-models",
				ModelName: "test-model",
				Device:    "cpu",
				MaxLength: 256,
				BatchSize: 16,
				RateLimit: 500,
			},
			wantErr: false,
		},
		{
			name:    "default config",
			config:  SentenceTransformersConfig{},
			wantErr: false,
		},
		{
			name: "custom device",
			config: SentenceTransformersConfig{
				Device: "cuda",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewSentenceTransformersProvider(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, provider)
			assert.Equal(t, embedding.ProviderTypeLocal, provider.Type())
			assert.Contains(t, provider.Name(), "sentence-transformers")
			assert.True(t, provider.isInitialized)
		})
	}
}

func TestSentenceTransformersProvider_Type(t *testing.T) {
	provider, err := NewSentenceTransformersProvider(SentenceTransformersConfig{})
	require.NoError(t, err)

	assert.Equal(t, embedding.ProviderTypeLocal, provider.Type())
}

func TestSentenceTransformersProvider_Name(t *testing.T) {
	config := SentenceTransformersConfig{ModelName: "test-model"}
	provider, err := NewSentenceTransformersProvider(config)
	require.NoError(t, err)

	assert.Equal(t, "sentence-transformers-test-model", provider.Name())
}

func TestSentenceTransformersProvider_GetModels(t *testing.T) {
	provider, err := NewSentenceTransformersProvider(SentenceTransformersConfig{})
	require.NoError(t, err)

	ctx := context.Background()
	models, err := provider.GetModels(ctx)
	require.NoError(t, err)
	require.Len(t, models, 1)

	model := models[0]
	assert.Equal(t, "all-MiniLM-L6-v2", model.ID)
	assert.Equal(t, "all-MiniLM-L6-v2", model.Name)
	assert.Equal(t, embedding.ProviderTypeSentenceTransformers, model.Provider)
	assert.Equal(t, 384, model.Dimensions)
	assert.Equal(t, 512, model.MaxTokens)
	assert.True(t, model.Supported)
}

func TestSentenceTransformersProvider_GetCapabilities(t *testing.T) {
	config := SentenceTransformersConfig{
		MaxLength: 256,
		BatchSize: 16,
	}
	provider, err := NewSentenceTransformersProvider(config)
	require.NoError(t, err)

	capabilities := provider.GetCapabilities()
	assert.Equal(t, 16, capabilities.MaxBatchSize)
	assert.Equal(t, 256, capabilities.MaxTextLength)
	assert.False(t, capabilities.SupportsAsync)
	assert.False(t, capabilities.SupportsStreaming)
	assert.Equal(t, 1000, capabilities.RateLimit.RequestsPerMinute)
	assert.Contains(t, capabilities.Features, "text-embeddings")
	assert.Contains(t, capabilities.Features, "batch-processing")
}

func TestSentenceTransformersProvider_GenerateEmbeddings(t *testing.T) {
	provider, err := NewSentenceTransformersProvider(SentenceTransformersConfig{})
	require.NoError(t, err)

	ctx := context.Background()
	request := &embedding.EmbeddingRequest{
		Texts: []string{"Hello world", "Test sentence"},
		Model: "all-MiniLM-L6-v2",
	}

	response, err := provider.GenerateEmbeddings(ctx, request)
	require.NoError(t, err)
	assert.NotNil(t, response)

	// Check response structure
	assert.Len(t, response.Embeddings, 2)
	assert.Equal(t, "all-MiniLM-L6-v2", response.Model)
	assert.Equal(t, 384, len(response.Embeddings[0])) // Dimension check
	assert.Equal(t, 384, len(response.Embeddings[1])) // Dimension check

	// Check usage stats
	assert.Greater(t, response.Usage.TotalTokens, 0)
	assert.Equal(t, response.Usage.TotalTokens, response.Usage.PromptTokens)
	assert.Equal(t, 0, response.Usage.CompletionTokens)
}

func TestSentenceTransformersProvider_GenerateEmbeddings_Validation(t *testing.T) {
	provider, err := NewSentenceTransformersProvider(SentenceTransformersConfig{})
	require.NoError(t, err)

	ctx := context.Background()

	// Test empty texts
	request := &embedding.EmbeddingRequest{
		Texts: []string{},
		Model: "test-model",
	}

	_, err = provider.GenerateEmbeddings(ctx, request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no texts provided")
}

func TestSentenceTransformersProvider_GenerateEmbeddings_BatchProcessing(t *testing.T) {
	config := SentenceTransformersConfig{
		BatchSize: 2,
	}
	provider, err := NewSentenceTransformersProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	texts := []string{"Text 1", "Text 2", "Text 3", "Text 4", "Text 5"}
	request := &embedding.EmbeddingRequest{
		Texts: texts,
		Model: "test-model",
	}

	response, err := provider.GenerateEmbeddings(ctx, request)
	require.NoError(t, err)
	assert.Len(t, response.Embeddings, 5)

	// Verify all embeddings have correct dimensions
	for i, embedding := range response.Embeddings {
		assert.Equal(t, 384, len(embedding), "Embedding %d has wrong dimension", i)
	}
}

func TestSentenceTransformersProvider_HealthCheck(t *testing.T) {
	provider, err := NewSentenceTransformersProvider(SentenceTransformersConfig{})
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.HealthCheck(ctx)
	assert.NoError(t, err)
}

func TestSentenceTransformersProvider_Close(t *testing.T) {
	provider, err := NewSentenceTransformersProvider(SentenceTransformersConfig{})
	require.NoError(t, err)

	// Verify provider is initialized
	assert.True(t, provider.isInitialized)

	// Close provider
	err = provider.Close()
	assert.NoError(t, err)

	// Verify provider is closed
	assert.False(t, provider.isInitialized)
}

func TestSentenceTransformersProvider_Stats(t *testing.T) {
	provider, err := NewSentenceTransformersProvider(SentenceTransformersConfig{})
	require.NoError(t, err)

	// Generate some embeddings to update stats
	ctx := context.Background()
	request := &embedding.EmbeddingRequest{
		Texts: []string{"Test text"},
		Model: "test-model",
	}

	_, err = provider.GenerateEmbeddings(ctx, request)
	require.NoError(t, err)

	// Check stats
	stats := provider.GetStats()
	assert.Equal(t, embedding.ProviderTypeLocal, stats.Provider)
	assert.Equal(t, int64(1), stats.TotalRequests)
	assert.Greater(t, stats.TotalTokens, int64(0))
	assert.Equal(t, int64(0), stats.Errors)
	assert.False(t, stats.LastUsed.IsZero())
}

func TestSentenceTransformersProvider_RateLimiting(t *testing.T) {
	config := SentenceTransformersConfig{
		RateLimit: 10, // 10 requests per minute
	}
	provider, err := NewSentenceTransformersProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	request := &embedding.EmbeddingRequest{
		Texts: []string{"Test text"},
		Model: "test-model",
	}

	// First few requests should succeed
	for i := 0; i < 5; i++ {
		_, err = provider.GenerateEmbeddings(ctx, request)
		assert.NoError(t, err)
	}

	// Verify rate limiter is configured
	stats := provider.GetStats()
	assert.Equal(t, embedding.ProviderTypeLocal, stats.Provider)
	assert.Equal(t, int64(5), stats.TotalRequests)
}

func TestSentenceTransformersProvider_MockEmbeddingGeneration(t *testing.T) {
	provider, err := NewSentenceTransformersProvider(SentenceTransformersConfig{})
	require.NoError(t, err)

	// Test that mock embeddings are deterministic for the same input
	text1 := "Hello world"
	text2 := "Hello world"
	text3 := "Different text"

	embedding1 := provider.generateMockEmbedding(text1, 0)
	embedding2 := provider.generateMockEmbedding(text2, 0)
	embedding3 := provider.generateMockEmbedding(text3, 0)

	// Same text should produce same embedding
	assert.Equal(t, embedding1, embedding2)

	// Different text should produce different embedding
	assert.NotEqual(t, embedding1, embedding3)

	// Check dimensions
	assert.Equal(t, 384, len(embedding1))
	assert.Equal(t, 384, len(embedding2))
	assert.Equal(t, 384, len(embedding3))

	// Check values are in expected range (0.0 to 1.0)
	for _, val := range embedding1 {
		assert.GreaterOrEqual(t, val, 0.0)
		assert.LessOrEqual(t, val, 1.0)
	}
}

func TestSentenceTransformersProvider_TokenCalculation(t *testing.T) {
	provider, err := NewSentenceTransformersProvider(SentenceTransformersConfig{})
	require.NoError(t, err)

	tests := []struct {
		texts       []string
		expectedMin int
		expectedMax int
	}{
		{
			texts:       []string{"Hello"},
			expectedMin: 1,
			expectedMax: 2,
		},
		{
			texts:       []string{"Hello world", "This is a test sentence"},
			expectedMin: 3,
			expectedMax: 8,
		},
		{
			texts:       []string{},
			expectedMin: 0,
			expectedMax: 0,
		},
	}

	for _, tt := range tests {
		tokens := provider.calculateTokens(tt.texts)
		assert.GreaterOrEqual(t, tokens, tt.expectedMin)
		assert.LessOrEqual(t, tokens, tt.expectedMax)
	}
}

func TestSentenceTransformersProvider_Concurrency(t *testing.T) {
	provider, err := NewSentenceTransformersProvider(SentenceTransformersConfig{})
	require.NoError(t, err)

	ctx := context.Background()
	request := &embedding.EmbeddingRequest{
		Texts: []string{"Test text"},
		Model: "test-model",
	}

	// Test concurrent access
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := provider.GenerateEmbeddings(ctx, request)
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		// Some requests might fail due to rate limiting, which is expected
		if err != nil {
			assert.Contains(t, err.Error(), "rate limit exceeded")
		}
	}

	// Verify stats are updated
	stats := provider.GetStats()
	assert.GreaterOrEqual(t, stats.TotalRequests, int64(1))
}

func BenchmarkSentenceTransformersProvider_GenerateEmbeddings(b *testing.B) {
	provider, err := NewSentenceTransformersProvider(SentenceTransformersConfig{})
	require.NoError(b, err)

	ctx := context.Background()
	request := &embedding.EmbeddingRequest{
		Texts: []string{"Benchmark test text for performance testing"},
		Model: "benchmark-model",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.GenerateEmbeddings(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSentenceTransformersProvider_BatchProcessing(b *testing.B) {
	config := SentenceTransformersConfig{
		BatchSize: 32,
	}
	provider, err := NewSentenceTransformersProvider(config)
	require.NoError(b, err)

	ctx := context.Background()
	texts := make([]string, 100)
	for i := range texts {
		texts[i] = fmt.Sprintf("Text %d for batch processing benchmark", i)
	}
	request := &embedding.EmbeddingRequest{
		Texts: texts,
		Model: "batch-model",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.GenerateEmbeddings(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}
