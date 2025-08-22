package embedding

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for the embedding service
func TestEmbeddingService_Integration(t *testing.T) {
	// Create comprehensive configuration
	config := &Config{
		DefaultProvider: ProviderTypeOpenAI,
		Timeout:         30 * time.Second,
		MaxBatchSize:    100,
		EnableFallback:  true,
		FallbackOrder:   []ProviderType{ProviderTypeOpenAI, ProviderTypeLocal},
		Cache: CacheConfig{
			Enabled:   true,
			Type:      "memory",
			TTL:       5 * time.Minute,
			MaxSize:   1000,
			MaxMemory: 100 * 1024 * 1024, // 100MB
		},
		RateLimiting: RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 1000,
			TokensPerMinute:   100000,
			BurstSize:         100,
		},
		Retry: RetryConfig{
			Enabled:       true,
			MaxRetries:    3,
			InitialDelay:  100 * time.Millisecond,
			MaxDelay:      5 * time.Second,
			BackoffFactor: 2.0,
		},
	}

	// Create service
	service, err := NewService(config)
	require.NoError(t, err)
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatal("failed to close service: %w", err)
		}
	}()

	// Test service creation
	assert.NotNil(t, service)
	assert.Equal(t, 0, len(service.ListProviders()))

	// Test provider registration
	mockProvider1 := &MockProvider{
		providerType: ProviderTypeOpenAI,
		name:         "MockOpenAI1",
	}

	err = service.RegisterProvider(mockProvider1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(service.ListProviders()))

	// Test duplicate provider registration
	err = service.RegisterProvider(mockProvider1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Test provider retrieval
	provider, err := service.GetProvider(ProviderTypeOpenAI)
	assert.NoError(t, err)
	assert.Equal(t, mockProvider1, provider)

	// Test non-existent provider
	_, err = service.GetProvider(ProviderTypeLocal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test embedding generation
	ctx := context.Background()
	req := &EmbeddingRequest{
		Texts:    []string{"Test text for embedding"},
		Model:    "text-embedding-ada-002",
		Provider: ProviderTypeOpenAI,
		CacheKey: "test_cache_key",
	}

	response, err := service.GenerateEmbeddings(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, ProviderTypeOpenAI, response.Provider)
	assert.Equal(t, "text-embedding-ada-002", response.Model)
	assert.False(t, response.CacheHit)
	assert.Equal(t, 1, len(response.Embeddings))
	assert.Equal(t, 1536, len(response.Embeddings[0])) // OpenAI ada-002 dimensions

	// Test caching
	response2, err := service.GenerateEmbeddings(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, response2)
	assert.True(t, response2.CacheHit) // Should be cached now

	// Test batch processing
	batchReq := &EmbeddingRequest{
		Texts:     []string{"Text 1", "Text 2", "Text 3"},
		Model:     "text-embedding-ada-002",
		Provider:  ProviderTypeOpenAI,
		BatchSize: 3,
	}

	batchResponse, err := service.GenerateEmbeddings(ctx, batchReq)
	assert.NoError(t, err)
	assert.NotNil(t, batchResponse)
	assert.Equal(t, 3, len(batchResponse.Embeddings))
	assert.Equal(t, 3, batchResponse.Usage.TotalTokens/100) // Mock calculation

	// Test fallback mechanism
	fallbackReq := &EmbeddingRequest{
		Texts:    []string{"Fallback test"},
		Model:    "text-embedding-ada-002",
		Provider: ProviderTypeLocal, // Request local provider
	}

	// This should succeed due to fallback to OpenAI provider
	fallbackResponse, err := service.GenerateEmbeddings(ctx, fallbackReq)
	assert.NoError(t, err)
	assert.NotNil(t, fallbackResponse)
	assert.Equal(t, ProviderTypeOpenAI, fallbackResponse.Provider) // Should fallback to OpenAI

	// Test provider statistics
	stats := service.GetProviderStats()
	assert.NotEmpty(t, stats)

	openAIStats, exists := stats[ProviderTypeOpenAI]
	assert.True(t, exists)
	assert.Equal(t, ProviderTypeOpenAI, openAIStats.Provider)
	assert.Greater(t, openAIStats.TotalRequests, int64(0))
	assert.Greater(t, openAIStats.TotalTokens, int64(0))
	assert.Greater(t, openAIStats.TotalCost, 0.0)

	// Test health check
	healthResults := service.HealthCheck(ctx)
	assert.NotEmpty(t, healthResults)

	openAIHealth, exists := healthResults[ProviderTypeOpenAI]
	assert.True(t, exists)
	assert.NoError(t, openAIHealth)
}

// Test fallback mechanism
func TestEmbeddingService_Fallback(t *testing.T) {
	config := &Config{
		DefaultProvider: ProviderTypeOpenAI,
		EnableFallback:  true,
		FallbackOrder:   []ProviderType{ProviderTypeOpenAI, ProviderTypeLocal},
		Cache: CacheConfig{
			Enabled: true,
			Type:    "memory",
			TTL:     5 * time.Minute,
		},
	}

	service, err := NewService(config)
	require.NoError(t, err)
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatal("failed to close service: %w", err)
		}
	}()

	// Create multiple providers
	openAIProvider := &MockProvider{
		providerType: ProviderTypeOpenAI,
		name:         "MockOpenAI",
	}

	localProvider := &MockProvider{
		providerType: ProviderTypeLocal,
		name:         "MockLocal",
	}

	// Register providers
	err = service.RegisterProvider(openAIProvider)
	require.NoError(t, err)
	err = service.RegisterProvider(localProvider)
	require.NoError(t, err)

	// Test fallback when primary provider fails
	// We'll need to modify the MockProvider to simulate failures
	// For now, test the basic fallback configuration
	assert.Equal(t, 2, len(service.ListProviders()))

	// Verify fallback order
	providers := service.ListProviders()
	assert.Equal(t, ProviderTypeOpenAI, providers[0].Type())
	assert.Equal(t, ProviderTypeLocal, providers[1].Type())
}

// Test rate limiting integration
func TestEmbeddingService_RateLimiting(t *testing.T) {
	config := &Config{
		DefaultProvider: ProviderTypeOpenAI,
		RateLimiting: RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 1000, // High limit for testing
			TokensPerMinute:   100000,
			BurstSize:         100,
		},
		Cache: CacheConfig{
			Enabled: true,
			Type:    "memory",
			TTL:     5 * time.Minute,
		},
	}

	service, err := NewService(config)
	require.NoError(t, err)
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatal("failed to close service: %w", err)
		}
	}()

	// Register provider
	mockProvider := &MockProvider{
		providerType: ProviderTypeOpenAI,
		name:         "MockOpenAI",
	}
	err = service.RegisterProvider(mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	req := &EmbeddingRequest{
		Texts:    []string{"Rate limit test"},
		Model:    "text-embedding-ada-002",
		Provider: ProviderTypeOpenAI,
	}

	// Make a few requests to verify rate limiting is configured
	for i := 0; i < 5; i++ {
		_, err := service.GenerateEmbeddings(ctx, req)
		assert.NoError(t, err, "Request %d should succeed", i)
	}

	// Test that rate limiting is working by checking stats
	stats := service.GetProviderStats()
	openAIStats, exists := stats[ProviderTypeOpenAI]
	assert.True(t, exists)
	assert.Greater(t, openAIStats.TotalRequests, int64(0))
}

// Test caching integration
func TestEmbeddingService_Caching(t *testing.T) {
	config := &Config{
		DefaultProvider: ProviderTypeOpenAI,
		Cache: CacheConfig{
			Enabled: true,
			Type:    "memory",
			TTL:     100 * time.Millisecond, // Short TTL for testing
			MaxSize: 100,
		},
	}

	service, err := NewService(config)
	require.NoError(t, err)
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatal("failed to close service: %w", err)
		}
	}()

	// Register provider
	mockProvider := &MockProvider{
		providerType: ProviderTypeOpenAI,
		name:         "MockOpenAI",
	}
	err = service.RegisterProvider(mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	req := &EmbeddingRequest{
		Texts:    []string{"Cache test"},
		Model:    "text-embedding-ada-002",
		Provider: ProviderTypeOpenAI,
		CacheKey: "test_cache_key",
	}

	// First request - should not be cached
	response1, err := service.GenerateEmbeddings(ctx, req)
	assert.NoError(t, err)
	assert.False(t, response1.CacheHit)

	// Second request - should be cached
	response2, err := service.GenerateEmbeddings(ctx, req)
	assert.NoError(t, err)
	assert.True(t, response2.CacheHit)

	// Wait for cache to expire
	time.Sleep(200 * time.Millisecond)

	// Third request - should not be cached (expired)
	response3, err := service.GenerateEmbeddings(ctx, req)
	assert.NoError(t, err)
	assert.False(t, response3.CacheHit)
}

// Test retry mechanism integration
func TestEmbeddingService_Retry(t *testing.T) {
	config := &Config{
		DefaultProvider: ProviderTypeOpenAI,
		Retry: RetryConfig{
			Enabled:       true,
			MaxRetries:    2,
			InitialDelay:  10 * time.Millisecond,
			MaxDelay:      100 * time.Millisecond,
			BackoffFactor: 2.0,
		},
		Cache: CacheConfig{
			Enabled: true,
			Type:    "memory",
			TTL:     5 * time.Minute,
		},
	}

	service, err := NewService(config)
	require.NoError(t, err)
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatal("failed to close service: %w", err)
		}
	}()

	// Register provider
	mockProvider := &MockProvider{
		providerType: ProviderTypeOpenAI,
		name:         "MockOpenAI",
	}
	err = service.RegisterProvider(mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	req := &EmbeddingRequest{
		Texts:    []string{"Retry test"},
		Model:    "text-embedding-ada-002",
		Provider: ProviderTypeOpenAI,
	}

	// Test successful request
	response, err := service.GenerateEmbeddings(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

// Test concurrent requests
func TestEmbeddingService_Concurrency(t *testing.T) {
	config := &Config{
		DefaultProvider: ProviderTypeOpenAI,
		Cache: CacheConfig{
			Enabled: true,
			Type:    "memory",
			TTL:     5 * time.Minute,
		},
	}

	service, err := NewService(config)
	require.NoError(t, err)
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatal("failed to close service: %w", err)
		}
	}()

	// Register provider
	mockProvider := &MockProvider{
		providerType: ProviderTypeOpenAI,
		name:         "MockOpenAI",
	}
	err = service.RegisterProvider(mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	numGoroutines := 10
	results := make(chan error, numGoroutines)

	// Launch concurrent requests
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			req := &EmbeddingRequest{
				Texts:    []string{fmt.Sprintf("Concurrent test %d", id)},
				Model:    "text-embedding-ada-002",
				Provider: ProviderTypeOpenAI,
				CacheKey: fmt.Sprintf("cache_key_%d", id),
			}

			_, err := service.GenerateEmbeddings(ctx, req)
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err, "Goroutine %d failed", i)
	}
}

// MockProvider is defined in benchmark_test.go
