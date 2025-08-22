package embedding

import (
	"context"
	"fmt"
	"sync"
	"time"

	"log/slog"
)

// service implements the embedding service
type service struct {
	config        *Config
	providers     map[ProviderType]Provider
	providerStats map[ProviderType]*ProviderStats
	cache         Cache
	rateLimiter   *RateLimiter
	retryManager  *RetryManager
	mu            sync.RWMutex
	logger        *slog.Logger
}

// NewService creates a new embedding service
func NewService(config *Config) (Service, error) {
	if config == nil {
		config = &Config{
			DefaultProvider: ProviderTypeOpenAI,
			Timeout:         30 * time.Second,
			MaxBatchSize:    100,
			EnableFallback:  true,
			FallbackOrder:   []ProviderType{ProviderTypeOpenAI, ProviderTypeLocal},
		}
	}

	// Initialize cache
	var cache Cache
	if config.Cache.Enabled {
		var err error
		cache, err = NewCache(config.Cache)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache: %w", err)
		}
	} else {
		cache = &NoOpCache{}
	}

	// Initialize rate limiter
	rateLimiter := NewRateLimiter(config.RateLimiting)

	// Initialize retry manager
	retryManager := NewRetryManager(config.Retry)

	s := &service{
		config:        config,
		providers:     make(map[ProviderType]Provider),
		providerStats: make(map[ProviderType]*ProviderStats),
		cache:         cache,
		rateLimiter:   rateLimiter,
		retryManager:  retryManager,
		logger:        slog.Default(),
	}

	return s, nil
}

// GenerateEmbeddings generates embeddings using the best available provider
func (s *service) GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	// Check cache first
	if req.CacheKey != "" {
		if cached, hit := s.cache.Get(req.CacheKey); hit {
			s.logger.Debug("Cache hit for embeddings", "key", req.CacheKey)
			return &EmbeddingResponse{
				Embeddings:     cached,
				Model:          req.Model,
				Provider:       req.Provider,
				CacheHit:       true,
				ProcessingTime: 0,
			}, nil
		}
	}

	// Use specified provider or default
	providerType := req.Provider
	if providerType == "" {
		providerType = s.config.DefaultProvider
	}

	// Try to generate embeddings
	response, err := s.generateWithProvider(ctx, req, providerType)
	if err != nil && s.config.EnableFallback {
		// Try fallback providers
		for _, fallbackType := range s.config.FallbackOrder {
			if fallbackType == providerType {
				continue
			}
			s.logger.Info("Trying fallback provider", "from", providerType, "to", fallbackType)
			response, err = s.generateWithProvider(ctx, req, fallbackType)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("all providers failed: %w", err)
	}

	// Cache the result if cache key is provided
	if req.CacheKey != "" {
		if err := s.cache.Set(req.CacheKey, response.Embeddings, s.config.Cache.TTL); err != nil {
			s.logger.Warn("Failed to cache embeddings", "error", err)
		}
	}

	// Update statistics
	s.updateStats(providerType, response)

	return response, nil
}

// GenerateEmbeddingsWithProvider generates embeddings using a specific provider
func (s *service) GenerateEmbeddingsWithProvider(ctx context.Context, req *EmbeddingRequest, providerType ProviderType) (*EmbeddingResponse, error) {
	return s.generateWithProvider(ctx, req, providerType)
}

// generateWithProvider generates embeddings using a specific provider
func (s *service) generateWithProvider(ctx context.Context, req *EmbeddingRequest, providerType ProviderType) (*EmbeddingResponse, error) {
	s.mu.RLock()
	provider, exists := s.providers[providerType]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerType)
	}

	// Check rate limits
	if err := s.rateLimiter.Allow(providerType, req.Texts); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Apply timeout
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	// Generate embeddings with retry logic
	var response *EmbeddingResponse
	err := s.retryManager.Do(func() error {
		start := time.Now()
		var err error
		response, err = provider.GenerateEmbeddings(ctx, req)
		if err != nil {
			return err
		}
		response.ProcessingTime = time.Since(start)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("provider %s failed: %w", providerType, err)
	}

	response.Provider = providerType
	return response, nil
}

// RegisterProvider registers a new embedding provider
func (s *service) RegisterProvider(provider Provider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	providerType := provider.Type()
	if _, exists := s.providers[providerType]; exists {
		return fmt.Errorf("provider %s already registered", providerType)
	}

	s.providers[providerType] = provider
	s.providerStats[providerType] = &ProviderStats{
		Provider: providerType,
		LastUsed: time.Now(),
	}

	s.logger.Info("Registered embedding provider", "type", providerType, "name", provider.Name())
	return nil
}

// GetProvider returns a provider by type
func (s *service) GetProvider(providerType ProviderType) (Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider, exists := s.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerType)
	}

	return provider, nil
}

// ListProviders returns all available providers
func (s *service) ListProviders() []Provider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]Provider, 0, len(s.providers))
	for _, provider := range s.providers {
		providers = append(providers, provider)
	}

	return providers
}

// GetProviderStats returns statistics for all providers
func (s *service) GetProviderStats() map[ProviderType]ProviderStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[ProviderType]ProviderStats)
	for providerType, stat := range s.providerStats {
		stats[providerType] = *stat
	}

	return stats
}

// HealthCheck checks health of all providers
func (s *service) HealthCheck(ctx context.Context) map[ProviderType]error {
	s.mu.RLock()
	providers := make(map[ProviderType]Provider, len(s.providers))
	for k, v := range s.providers {
		providers[k] = v
	}
	s.mu.RUnlock()

	results := make(map[ProviderType]error)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for providerType, provider := range providers {
		wg.Add(1)
		go func(pt ProviderType, p Provider) {
			defer wg.Done()
			err := p.HealthCheck(ctx)
			mu.Lock()
			results[pt] = err
			mu.Unlock()
		}(providerType, provider)
	}

	wg.Wait()
	return results
}

// Close closes all providers
func (s *service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var lastErr error
	for _, provider := range s.providers {
		if err := provider.Close(); err != nil {
			lastErr = err
			s.logger.Error("Failed to close provider", "error", err)
		}
	}

	if err := s.cache.Close(); err != nil {
		lastErr = err
		s.logger.Error("Failed to close cache", "error", err)
	}

	return lastErr
}

// updateStats updates provider statistics
func (s *service) updateStats(providerType ProviderType, response *EmbeddingResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, exists := s.providerStats[providerType]
	if !exists {
		stats = &ProviderStats{
			Provider: providerType,
		}
		s.providerStats[providerType] = stats
	}

	stats.TotalRequests++
	stats.TotalTokens += int64(response.Usage.TotalTokens)
	stats.TotalCost += response.Usage.TotalCost
	stats.LastUsed = time.Now()

	if response.CacheHit {
		stats.CacheHits++
	} else {
		stats.CacheMisses++
	}

	if response.Error != nil {
		stats.Errors++
	}

	// Update average latency
	if stats.AverageLatency == 0 {
		stats.AverageLatency = response.ProcessingTime
	} else {
		stats.AverageLatency = (stats.AverageLatency + response.ProcessingTime) / 2
	}
}
