package enterprise

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimitConfig represents configuration for rate limiting
type RateLimitConfig struct {
	// Global limits
	GlobalRequestsPerSecond int `json:"global_requests_per_second"`
	GlobalBurstSize         int `json:"global_burst_size"`

	// Per-tenant limits
	DefaultTenantRequestsPerSecond int `json:"default_tenant_requests_per_second"`
	DefaultTenantBurstSize         int `json:"default_tenant_burst_size"`

	// Per-endpoint limits
	EndpointLimits map[string]EndpointLimit `json:"endpoint_limits"`

	// IP-based limits
	IPRequestsPerSecond int `json:"ip_requests_per_second"`
	IPBurstSize         int `json:"ip_burst_size"`

	// Cleanup settings
	CleanupInterval time.Duration `json:"cleanup_interval"`
	MaxEntries      int           `json:"max_entries"`
}

// EndpointLimit represents rate limits for a specific endpoint
type EndpointLimit struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	Window            time.Duration `json:"window"`
}

// DefaultRateLimitConfig returns the default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		GlobalRequestsPerSecond:        1000,
		GlobalBurstSize:                100,
		DefaultTenantRequestsPerSecond: 100,
		DefaultTenantBurstSize:         10,
		IPRequestsPerSecond:            50,
		IPBurstSize:                    5,
		CleanupInterval:                5 * time.Minute,
		MaxEntries:                     10000,
		EndpointLimits: map[string]EndpointLimit{
			"POST:/collections": {
				RequestsPerSecond: 10,
				BurstSize:         2,
				Window:            1 * time.Second,
			},
			"POST:/vectors": {
				RequestsPerSecond: 100,
				BurstSize:         20,
				Window:            1 * time.Second,
			},
			"GET:/search": {
				RequestsPerSecond: 200,
				BurstSize:         50,
				Window:            1 * time.Second,
			},
		},
	}
}

// RateLimiter defines the rate limiting service interface
type RateLimiter interface {
	// Rate Limiting
	Allow(ctx context.Context, key string, limit int, window time.Duration) bool
	AllowTenant(ctx context.Context, tenantID string, endpoint string) bool
	AllowIP(ctx context.Context, ipAddress string) bool
	AllowGlobal(ctx context.Context) bool

	// Configuration
	SetTenantLimit(ctx context.Context, tenantID string, requestsPerSecond, burstSize int) error
	SetEndpointLimit(ctx context.Context, endpoint string, limit EndpointLimit) error
	GetTenantLimit(ctx context.Context, tenantID string) (int, int, error)
	GetEndpointLimit(ctx context.Context, endpoint string) (EndpointLimit, error)

	// Monitoring
	GetRemaining(ctx context.Context, key string) (int, error)
	GetTenantUsage(ctx context.Context, tenantID string) (*TenantRateUsage, error)
	GetIPUsage(ctx context.Context, ipAddress string) (*IPRateUsage, error)

	// Health Check
	HealthCheck(ctx context.Context) error
	Close() error
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	Tokens     int           `json:"tokens"`
	Capacity   int           `json:"capacity"`
	LastRefill time.Time     `json:"last_refill"`
	RefillRate float64       `json:"refill_rate"` // tokens per second
	Window     time.Duration `json:"window"`
}

// TenantRateUsage represents rate limiting usage for a tenant
type TenantRateUsage struct {
	TenantID       string         `json:"tenant_id"`
	CurrentUsage   map[string]int `json:"current_usage"`   // endpoint -> current requests
	RemainingQuota map[string]int `json:"remaining_quota"` // endpoint -> remaining requests
	LastReset      time.Time      `json:"last_reset"`
	NextReset      time.Time      `json:"next_reset"`
}

// IPRateUsage represents rate limiting usage for an IP address
type IPRateUsage struct {
	IPAddress      string    `json:"ip_address"`
	CurrentUsage   int       `json:"current_usage"`
	RemainingQuota int       `json:"remaining_quota"`
	LastReset      time.Time `json:"last_reset"`
	NextReset      time.Time `json:"next_reset"`
}

// DefaultRateLimiter implements the rate limiting service
type DefaultRateLimiter struct {
	config          *RateLimitConfig
	tenantBuckets   map[string]*TokenBucket
	ipBuckets       map[string]*TokenBucket
	globalBucket    *TokenBucket
	endpointBuckets map[string]*TokenBucket
	mu              sync.RWMutex
	cleanupTicker   *time.Ticker
	done            chan bool
}

// NewDefaultRateLimiter creates a new default rate limiter
func NewDefaultRateLimiter(config *RateLimitConfig) *DefaultRateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	limiter := &DefaultRateLimiter{
		config:          config,
		tenantBuckets:   make(map[string]*TokenBucket),
		ipBuckets:       make(map[string]*TokenBucket),
		endpointBuckets: make(map[string]*TokenBucket),
		done:            make(chan bool),
	}

	// Initialize global bucket
	limiter.globalBucket = &TokenBucket{
		Tokens:     config.GlobalBurstSize,
		Capacity:   config.GlobalBurstSize,
		LastRefill: time.Now(),
		RefillRate: float64(config.GlobalRequestsPerSecond),
		Window:     1 * time.Second,
	}

	// Initialize endpoint buckets
	for endpoint, limit := range config.EndpointLimits {
		limiter.endpointBuckets[endpoint] = &TokenBucket{
			Tokens:     limit.BurstSize,
			Capacity:   limit.BurstSize,
			LastRefill: time.Now(),
			RefillRate: float64(limit.RequestsPerSecond),
			Window:     limit.Window,
		}
	}

	// Start cleanup routine
	limiter.startCleanup()

	return limiter
}

// Allow checks if a request is allowed for a generic key
func (r *DefaultRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	bucket := r.getOrCreateBucket(key, limit, window)
	return r.consumeToken(bucket)
}

// AllowTenant checks if a request is allowed for a tenant
func (r *DefaultRateLimiter) AllowTenant(ctx context.Context, tenantID string, endpoint string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check global limit first
	if !r.consumeToken(r.globalBucket) {
		return false
	}

	// Check endpoint-specific limit
	if endpointLimit, exists := r.config.EndpointLimits[endpoint]; exists {
		endpointKey := fmt.Sprintf("endpoint:%s", endpoint)
		bucket, exists := r.endpointBuckets[endpointKey]
		if !exists {
			bucket = &TokenBucket{
				Tokens:     endpointLimit.BurstSize,
				Capacity:   endpointLimit.BurstSize,
				LastRefill: time.Now(),
				RefillRate: float64(endpointLimit.RequestsPerSecond),
				Window:     endpointLimit.Window,
			}
			r.endpointBuckets[endpointKey] = bucket
		}
		if !r.consumeToken(bucket) {
			return false
		}
	}

	// Check tenant limit
	tenantKey := fmt.Sprintf("tenant:%s", tenantID)
	bucket, exists := r.tenantBuckets[tenantKey]
	if !exists {
		bucket = &TokenBucket{
			Tokens:     r.config.DefaultTenantBurstSize,
			Capacity:   r.config.DefaultTenantBurstSize,
			LastRefill: time.Now(),
			RefillRate: float64(r.config.DefaultTenantRequestsPerSecond),
			Window:     1 * time.Second,
		}
		r.tenantBuckets[tenantKey] = bucket
	}

	return r.consumeToken(bucket)
}

// AllowIP checks if a request is allowed for an IP address
func (r *DefaultRateLimiter) AllowIP(ctx context.Context, ipAddress string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	ipKey := fmt.Sprintf("ip:%s", ipAddress)
	bucket, exists := r.ipBuckets[ipKey]
	if !exists {
		bucket = &TokenBucket{
			Tokens:     r.config.IPBurstSize,
			Capacity:   r.config.IPBurstSize,
			LastRefill: time.Now(),
			RefillRate: float64(r.config.IPRequestsPerSecond),
			Window:     1 * time.Second,
		}
		r.ipBuckets[ipKey] = bucket
	}

	return r.consumeToken(bucket)
}

// AllowGlobal checks if a request is allowed globally
func (r *DefaultRateLimiter) AllowGlobal(ctx context.Context) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.consumeToken(r.globalBucket)
}

// SetTenantLimit sets rate limits for a specific tenant
func (r *DefaultRateLimiter) SetTenantLimit(ctx context.Context, tenantID string, requestsPerSecond, burstSize int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tenantKey := fmt.Sprintf("tenant:%s", tenantID)
	bucket := &TokenBucket{
		Tokens:     burstSize,
		Capacity:   burstSize,
		LastRefill: time.Now(),
		RefillRate: float64(requestsPerSecond),
		Window:     1 * time.Second,
	}

	r.tenantBuckets[tenantKey] = bucket
	return nil
}

// SetEndpointLimit sets rate limits for a specific endpoint
func (r *DefaultRateLimiter) SetEndpointLimit(ctx context.Context, endpoint string, limit EndpointLimit) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	endpointKey := fmt.Sprintf("endpoint:%s", endpoint)
	bucket := &TokenBucket{
		Tokens:     limit.BurstSize,
		Capacity:   limit.BurstSize,
		LastRefill: time.Now(),
		RefillRate: float64(limit.RequestsPerSecond),
		Window:     limit.Window,
	}

	r.endpointBuckets[endpointKey] = bucket
	r.config.EndpointLimits[endpoint] = limit
	return nil
}

// GetTenantLimit gets the current rate limits for a tenant
func (r *DefaultRateLimiter) GetTenantLimit(ctx context.Context, tenantID string) (int, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantKey := fmt.Sprintf("tenant:%s", tenantID)
	bucket, exists := r.tenantBuckets[tenantKey]
	if !exists {
		return r.config.DefaultTenantRequestsPerSecond, r.config.DefaultTenantBurstSize, nil
	}

	return int(bucket.RefillRate), bucket.Capacity, nil
}

// GetEndpointLimit gets the current rate limits for an endpoint
func (r *DefaultRateLimiter) GetEndpointLimit(ctx context.Context, endpoint string) (EndpointLimit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	limit, exists := r.config.EndpointLimits[endpoint]
	if !exists {
		return EndpointLimit{}, fmt.Errorf("endpoint limit not found: %s", endpoint)
	}

	return limit, nil
}

// GetRemaining gets the remaining tokens for a key
func (r *DefaultRateLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try to find the bucket in different categories
	if bucket, exists := r.tenantBuckets[key]; exists {
		r.refillBucket(bucket)
		return bucket.Tokens, nil
	}
	if bucket, exists := r.ipBuckets[key]; exists {
		r.refillBucket(bucket)
		return bucket.Tokens, nil
	}
	if bucket, exists := r.endpointBuckets[key]; exists {
		r.refillBucket(bucket)
		return bucket.Tokens, nil
	}

	return 0, fmt.Errorf("key not found: %s", key)
}

// GetTenantUsage gets the current usage for a tenant
func (r *DefaultRateLimiter) GetTenantUsage(ctx context.Context, tenantID string) (*TenantRateUsage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantKey := fmt.Sprintf("tenant:%s", tenantID)
	bucket, exists := r.tenantBuckets[tenantKey]
	if !exists {
		// Return default limits
		return &TenantRateUsage{
			TenantID:       tenantID,
			CurrentUsage:   make(map[string]int),
			RemainingQuota: make(map[string]int),
			LastReset:      time.Now(),
			NextReset:      time.Now().Add(1 * time.Second),
		}, nil
	}

	r.refillBucket(bucket)
	usage := &TenantRateUsage{
		TenantID:       tenantID,
		CurrentUsage:   make(map[string]int),
		RemainingQuota: make(map[string]int),
		LastReset:      bucket.LastRefill,
		NextReset:      bucket.LastRefill.Add(1 * time.Second),
	}

	// Calculate remaining quota
	usage.RemainingQuota["tenant"] = bucket.Tokens

	return usage, nil
}

// GetIPUsage gets the current usage for an IP address
func (r *DefaultRateLimiter) GetIPUsage(ctx context.Context, ipAddress string) (*IPRateUsage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ipKey := fmt.Sprintf("ip:%s", ipAddress)
	bucket, exists := r.ipBuckets[ipKey]
	if !exists {
		// Return default limits
		return &IPRateUsage{
			IPAddress:      ipAddress,
			CurrentUsage:   0,
			RemainingQuota: r.config.IPBurstSize,
			LastReset:      time.Now(),
			NextReset:      time.Now().Add(1 * time.Second),
		}, nil
	}

	r.refillBucket(bucket)
	usage := &IPRateUsage{
		IPAddress:      ipAddress,
		CurrentUsage:   r.config.IPBurstSize - bucket.Tokens,
		RemainingQuota: bucket.Tokens,
		LastReset:      bucket.LastRefill,
		NextReset:      bucket.LastRefill.Add(1 * time.Second),
	}

	return usage, nil
}

// HealthCheck performs a health check on the service
func (r *DefaultRateLimiter) HealthCheck(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.tenantBuckets == nil || r.ipBuckets == nil || r.globalBucket == nil {
		return fmt.Errorf("rate limiter not properly initialized")
	}

	return nil
}

// Close performs cleanup operations
func (r *DefaultRateLimiter) Close() error {
	if r.cleanupTicker != nil {
		r.cleanupTicker.Stop()
	}
	close(r.done)
	return nil
}

// Helper methods

// getOrCreateBucket gets an existing bucket or creates a new one
func (r *DefaultRateLimiter) getOrCreateBucket(key string, limit int, window time.Duration) *TokenBucket {
	bucket, exists := r.tenantBuckets[key]
	if !exists {
		bucket = &TokenBucket{
			Tokens:     limit,
			Capacity:   limit,
			LastRefill: time.Now(),
			RefillRate: float64(limit),
			Window:     window,
		}
		r.tenantBuckets[key] = bucket
	}
	return bucket
}

// consumeToken consumes a token from a bucket
func (r *DefaultRateLimiter) consumeToken(bucket *TokenBucket) bool {
	r.refillBucket(bucket)

	if bucket.Tokens > 0 {
		bucket.Tokens--
		return true
	}
	return false
}

// refillBucket refills tokens in a bucket based on time elapsed
func (r *DefaultRateLimiter) refillBucket(bucket *TokenBucket) {
	now := time.Now()
	elapsed := now.Sub(bucket.LastRefill).Seconds()

	tokensToAdd := int(elapsed * bucket.RefillRate)
	if tokensToAdd > 0 {
		bucket.Tokens = min(bucket.Capacity, bucket.Tokens+tokensToAdd)
		bucket.LastRefill = now
	}
}

// startCleanup starts the cleanup routine for expired entries
func (r *DefaultRateLimiter) startCleanup() {
	r.cleanupTicker = time.NewTicker(r.config.CleanupInterval)

	go func() {
		for {
			select {
			case <-r.cleanupTicker.C:
				r.cleanup()
			case <-r.done:
				return
			}
		}
	}()
}

// cleanup removes old entries to prevent memory leaks
func (r *DefaultRateLimiter) cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clean up old tenant buckets (older than 1 hour)
	cutoff := time.Now().Add(-1 * time.Hour)
	for key, bucket := range r.tenantBuckets {
		if bucket.LastRefill.Before(cutoff) {
			delete(r.tenantBuckets, key)
		}
	}

	// Clean up old IP buckets (older than 1 hour)
	for key, bucket := range r.ipBuckets {
		if bucket.LastRefill.Before(cutoff) {
			delete(r.ipBuckets, key)
		}
	}

	// Limit total entries
	if len(r.tenantBuckets) > r.config.MaxEntries {
		// Remove oldest entries
		// In a real implementation, use a more sophisticated cleanup strategy
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
