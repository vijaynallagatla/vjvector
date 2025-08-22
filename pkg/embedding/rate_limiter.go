package embedding

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for embedding providers
type RateLimiter struct {
	config   RateLimitConfig
	limiters map[ProviderType]*rate.Limiter
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		config:   config,
		limiters: make(map[ProviderType]*rate.Limiter),
	}

	return rl
}

// Allow checks if a request is allowed based on rate limits
func (rl *RateLimiter) Allow(providerType ProviderType, texts []string) error {
	if !rl.config.Enabled {
		return nil
	}

	rl.mu.RLock()
	limiter, exists := rl.limiters[providerType]
	rl.mu.RUnlock()

	if !exists {
		// Create default limiter for this provider
		limiter = rl.createDefaultLimiter(providerType)
		rl.mu.Lock()
		rl.limiters[providerType] = limiter
		rl.mu.Unlock()
	}

	// Calculate tokens (rough estimate: 1 token ≈ 4 characters)
	totalTokens := 0
	for _, text := range texts {
		totalTokens += len(text) / 4
		if totalTokens < 1 {
			totalTokens = 1
		}
	}

	// Check if we can allow the request
	if !limiter.AllowN(time.Now(), totalTokens) {
		return fmt.Errorf("rate limit exceeded for provider %s", providerType)
	}

	return nil
}

// createDefaultLimiter creates a default rate limiter for a provider
func (rl *RateLimiter) createDefaultLimiter(providerType ProviderType) *rate.Limiter {
	var requestsPerSecond float64
	var tokensPerSecond float64

	switch providerType {
	case ProviderTypeOpenAI:
		requestsPerSecond = float64(rl.config.RequestsPerMinute) / 60.0
		tokensPerSecond = float64(rl.config.TokensPerMinute) / 60.0
	case ProviderTypeLocal:
		requestsPerSecond = 100.0 // High limit for local models
		tokensPerSecond = 10000.0
	case ProviderTypeSentenceTransformers:
		requestsPerSecond = 50.0
		tokensPerSecond = 5000.0
	default:
		requestsPerSecond = 10.0
		tokensPerSecond = 1000.0
	}

	// Use the more restrictive limit
	limit := requestsPerSecond
	if tokensPerSecond < requestsPerSecond {
		limit = tokensPerSecond
	}

	// Create rate limiter with burst capability
	return rate.NewLimiter(rate.Limit(limit), int(rl.config.BurstSize))
}

// SetProviderLimits sets custom rate limits for a specific provider
func (rl *RateLimiter) SetProviderLimits(providerType ProviderType, requestsPerSecond, tokensPerSecond float64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Use the more restrictive limit
	limit := requestsPerSecond
	if tokensPerSecond < requestsPerSecond {
		limit = tokensPerSecond
	}

	limiter := rate.NewLimiter(rate.Limit(limit), int(rl.config.BurstSize))
	rl.limiters[providerType] = limiter
}

// GetProviderLimits returns current rate limits for a provider
func (rl *RateLimiter) GetProviderLimits(providerType ProviderType) (requestsPerSecond, tokensPerSecond float64) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limiter, exists := rl.limiters[providerType]
	if !exists {
		return 0, 0
	}

	limit := float64(limiter.Limit())
	return limit, limit
}

// Reset resets all rate limiters
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limiters = make(map[ProviderType]*rate.Limiter)
}

// Wait waits for rate limit to allow the request
func (rl *RateLimiter) Wait(ctx context.Context, providerType ProviderType, texts []string) error {
	if !rl.config.Enabled {
		return nil
	}

	rl.mu.RLock()
	limiter, exists := rl.limiters[providerType]
	rl.mu.RUnlock()

	if !exists {
		limiter = rl.createDefaultLimiter(providerType)
		rl.mu.Lock()
		rl.limiters[providerType] = limiter
		rl.mu.Unlock()
	}

	// Calculate tokens
	totalTokens := 0
	for _, text := range texts {
		totalTokens += len(text) / 4
		if totalTokens < 1 {
			totalTokens = 1
		}
	}

	// Wait for rate limit
	return limiter.WaitN(ctx, totalTokens)
}
