package embedding

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockProvider for benchmarking
type MockProvider struct {
	providerType ProviderType
	name         string
}

func (m *MockProvider) Type() ProviderType { return m.providerType }
func (m *MockProvider) Name() string       { return m.name }
func (m *MockProvider) GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	// Return mock embeddings for benchmarking
	embeddings := make([][]float64, len(req.Texts))
	for i := range embeddings {
		embeddings[i] = make([]float64, 1536) // OpenAI ada-002 dimensions
		for j := range embeddings[i] {
			embeddings[i][j] = float64(i+j) / 1000.0
		}
	}

	return &EmbeddingResponse{
		Embeddings: embeddings,
		Model:      req.Model,
		Provider:   m.providerType,
		Usage: UsageStats{
			TotalTokens: len(req.Texts) * 100, // Mock token count
			TotalCost:   float64(len(req.Texts)) * 0.0001,
			Provider:    m.name,
		},
		CacheHit:       false,
		ProcessingTime: 50 * time.Millisecond, // Mock processing time
	}, nil
}
func (m *MockProvider) GetModels(ctx context.Context) ([]Model, error) {
	return []Model{}, nil
}
func (m *MockProvider) GetCapabilities() Capabilities {
	return Capabilities{}
}
func (m *MockProvider) HealthCheck(ctx context.Context) error { return nil }
func (m *MockProvider) Close() error                          { return nil }

// Benchmark tests for the embedding service
func BenchmarkEmbeddingService_GenerateEmbeddings(b *testing.B) {
	// Create test configuration
	config := &Config{
		DefaultProvider: ProviderTypeOpenAI,
		Timeout:         30 * time.Second,
		MaxBatchSize:    100,
		EnableFallback:  true,
		Cache: CacheConfig{
			Enabled:   true,
			Type:      "memory",
			TTL:       5 * time.Minute,
			MaxSize:   1000,
			MaxMemory: 100 * 1024 * 1024, // 100MB
		},
		RateLimiting: RateLimitConfig{
			Enabled: false, // Disable for benchmarks
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
	if err != nil {
		b.Fatalf("Failed to create service: %v", err)
	}
	defer func() {
		if err := service.Close(); err != nil {
			b.Fatalf("Failed to close service: %v", err)
		}
	}()

	// Create mock provider
	mockProvider := &MockProvider{
		providerType: ProviderTypeOpenAI,
		name:         "MockOpenAI",
	}

	// Register provider
	if err := service.RegisterProvider(mockProvider); err != nil {
		b.Fatalf("Failed to register provider: %v", err)
	}

	// Test data
	testTexts := []string{
		"This is a test sentence for benchmarking.",
		"Another test sentence with different content.",
		"A third sentence to test batch processing.",
		"Testing the performance of embedding generation.",
		"Final test sentence for comprehensive benchmarking.",
	}

	ctx := context.Background()

	// Benchmark single text embedding
	b.Run("SingleText", func(b *testing.B) {
		req := &EmbeddingRequest{
			Texts:    []string{testTexts[0]},
			Model:    "text-embedding-ada-002",
			Provider: ProviderTypeOpenAI,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := service.GenerateEmbeddings(ctx, req)
			if err != nil {
				b.Fatalf("Service failed: %v", err)
			}
		}
	})

	// Benchmark batch text embedding
	b.Run("BatchTexts", func(b *testing.B) {
		req := &EmbeddingRequest{
			Texts:     testTexts,
			Model:     "text-embedding-ada-002",
			Provider:  ProviderTypeOpenAI,
			BatchSize: 5,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := service.GenerateEmbeddings(ctx, req)
			if err != nil {
				b.Fatalf("Service failed: %v", err)
			}
		}
	})

	// Benchmark with caching
	b.Run("WithCaching", func(b *testing.B) {
		req := &EmbeddingRequest{
			Texts:    []string{testTexts[0]},
			Model:    "text-embedding-ada-002",
			Provider: ProviderTypeOpenAI,
			CacheKey: "benchmark_cache_key",
		}

		// First call to populate cache
		_, err := service.GenerateEmbeddings(ctx, req)
		if err != nil {
			b.Fatalf("Service failed: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := service.GenerateEmbeddings(ctx, req)
			if err != nil {
				b.Fatalf("Service failed: %v", err)
			}
		}
	})
}

// Benchmark cache performance
func BenchmarkEmbeddingCache(b *testing.B) {
	config := CacheConfig{
		Enabled:   true,
		Type:      "memory",
		TTL:       5 * time.Minute,
		MaxSize:   1000,
		MaxMemory: 100 * 1024 * 1024, // 100MB
	}

	cache, err := NewCache(config)
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}
	defer func() {
		if err := cache.Close(); err != nil {
			b.Fatal("failed to close  cache: %w", err)
		}

		// Test data
		testEmbeddings := [][]float64{
			{0.1, 0.2, 0.3, 0.4, 0.5},
			{0.6, 0.7, 0.8, 0.9, 1.0},
			{0.2, 0.3, 0.4, 0.5, 0.6},
		}

		// Benchmark cache operations
		b.Run("Set", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("benchmark_key_%d", i)
				err := cache.Set(key, testEmbeddings, 5*time.Minute)
				if err != nil {
					b.Fatalf("Cache set failed: %v", err)
				}
			}
		})

		b.Run("Get", func(b *testing.B) {
			// Pre-populate cache
			for i := 0; i < 100; i++ {
				key := fmt.Sprintf("benchmark_key_%d", i)
				err := cache.Set(key, testEmbeddings, 5*time.Minute)
				if err != nil {
					b.Fatalf("Cache set failed: %v", err)
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("benchmark_key_%d", i%100)
				_, hit := cache.Get(key)
				if !hit {
					b.Fatalf("Cache miss for key: %s", key)
				}
			}
		})

		b.Run("GetMiss", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("miss_key_%d", i)
				_, hit := cache.Get(key)
				if hit {
					b.Fatalf("Unexpected cache hit for key: %s", key)
				}
			}
		})
	}()
}

// Benchmark rate limiter performance
func BenchmarkRateLimiter(b *testing.B) {
	config := RateLimitConfig{
		Enabled: false, // Disable for benchmarks
	}

	limiter := NewRateLimiter(config)

	testTexts := []string{
		"This is a test sentence for rate limiting benchmarks.",
		"Another test sentence to measure performance.",
		"A third sentence for comprehensive testing.",
	}

	b.Run("Allow", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := limiter.Allow(ProviderTypeOpenAI, testTexts)
			if err != nil {
				b.Skipf("Rate limit exceeded: %v", err)
			}
		}
	})

	b.Run("Wait", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := limiter.Wait(ctx, ProviderTypeOpenAI, testTexts)
			if err != nil {
				b.Skipf("Rate limit wait failed: %v", err)
			}
		}
	})
}

// Benchmark retry manager performance
func BenchmarkRetryManager(b *testing.B) {
	config := RetryConfig{
		Enabled:       true,
		MaxRetries:    3,
		InitialDelay:  1 * time.Millisecond, // Use small delays for benchmarks
		MaxDelay:      10 * time.Millisecond,
		BackoffFactor: 2.0,
	}

	manager := NewRetryManager(config)

	// Benchmark successful operation
	b.Run("Success", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := manager.Do(func() error {
				return nil // Always succeed
			})
			if err != nil {
				b.Fatalf("Retry manager failed: %v", err)
			}
		}
	})

	// Benchmark retryable operation
	b.Run("RetryableError", func(b *testing.B) {
		attempts := 0
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			attempts = 0
			err := manager.Do(func() error {
				attempts++
				if attempts < 3 {
					return fmt.Errorf("retryable error")
				}
				return nil
			})
			if err != nil {
				b.Fatalf("Retry manager failed: %v", err)
			}
		}
	})
}

// Benchmark service statistics
func BenchmarkServiceStats(b *testing.B) {
	config := &Config{
		DefaultProvider: ProviderTypeOpenAI,
		Timeout:         30 * time.Second,
		MaxBatchSize:    100,
		EnableFallback:  true,
	}

	service, err := NewService(config)
	if err != nil {
		b.Fatalf("Failed to create service: %v", err)
	}
	defer func() {
		if err := service.Close(); err != nil {
			b.Fatalf("failed to close service: %v", err)
		}
	}()

	// Pre-populate with mock provider for realistic benchmarking
	mockProvider := &MockProvider{
		providerType: ProviderTypeOpenAI,
		name:         "MockOpenAI",
	}

	if err := service.RegisterProvider(mockProvider); err != nil {
		b.Fatal("failed to register provider: %w", err)
	}

	b.Run("GetProviderStats", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = service.GetProviderStats()
		}
	})

	b.Run("ListProviders", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = service.ListProviders()
		}
	})

	b.Run("HealthCheck", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = service.HealthCheck(ctx)
		}
	})
}
