package embedding

import (
	"context"
	"time"
)

// ProviderType represents the type of embedding provider
type ProviderType string

const (
	ProviderTypeOpenAI               ProviderType = "openai"
	ProviderTypeLocal                ProviderType = "local"
	ProviderTypeCustom               ProviderType = "custom"
	ProviderTypeSentenceTransformers ProviderType = "sentence-transformers"
)

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	Texts     []string               `json:"texts"`
	Model     string                 `json:"model"`
	Provider  ProviderType           `json:"provider"`
	Options   map[string]interface{} `json:"options,omitempty"`
	CacheKey  string                 `json:"cache_key,omitempty"`
	BatchSize int                    `json:"batch_size,omitempty"`
	Timeout   time.Duration          `json:"timeout,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EmbeddingResponse represents the response from embedding generation
type EmbeddingResponse struct {
	Embeddings     [][]float64   `json:"embeddings"`
	Model          string        `json:"model"`
	Provider       ProviderType  `json:"provider"`
	Usage          UsageStats    `json:"usage"`
	CacheHit       bool          `json:"cache_hit"`
	ProcessingTime time.Duration `json:"processing_time"`
	Error          error         `json:"error,omitempty"`
}

// UsageStats represents usage statistics for embedding generation
type UsageStats struct {
	TotalTokens      int     `json:"total_tokens"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalCost        float64 `json:"total_cost"`
	Provider         string  `json:"provider"`
}

// Provider represents an embedding service provider
type Provider interface {
	// Type returns the provider type
	Type() ProviderType

	// Name returns the provider name
	Name() string

	// GenerateEmbeddings generates embeddings for the given texts
	GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// GetModels returns available models for this provider
	GetModels(ctx context.Context) ([]Model, error)

	// GetCapabilities returns provider capabilities
	GetCapabilities() Capabilities

	// HealthCheck checks if the provider is healthy
	HealthCheck(ctx context.Context) error

	// Close closes the provider and releases resources
	Close() error
}

// Model represents an embedding model
type Model struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Provider   ProviderType           `json:"provider"`
	Dimensions int                    `json:"dimensions"`
	MaxTokens  int                    `json:"max_tokens"`
	CostPer1K  float64                `json:"cost_per_1k_tokens"`
	Supported  bool                   `json:"supported"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Capabilities represents what a provider can do
type Capabilities struct {
	MaxBatchSize      int       `json:"max_batch_size"`
	MaxTextLength     int       `json:"max_text_length"`
	SupportsAsync     bool      `json:"supports_async"`
	SupportsStreaming bool      `json:"supports_streaming"`
	RateLimit         RateLimit `json:"rate_limit"`
	Features          []string  `json:"features"`
}

// RateLimit represents rate limiting information
type RateLimit struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	TokensPerMinute   int `json:"tokens_per_minute"`
	RequestsPerDay    int `json:"requests_per_day"`
	TokensPerDay      int `json:"tokens_per_day"`
}

// Service represents the main embedding service
type Service interface {
	// GenerateEmbeddings generates embeddings using the best available provider
	GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// GenerateEmbeddingsWithProvider generates embeddings using a specific provider
	GenerateEmbeddingsWithProvider(ctx context.Context, req *EmbeddingRequest, provider ProviderType) (*EmbeddingResponse, error)

	// RegisterProvider registers a new embedding provider
	RegisterProvider(provider Provider) error

	// GetProvider returns a provider by type
	GetProvider(providerType ProviderType) (Provider, error)

	// ListProviders returns all available providers
	ListProviders() []Provider

	// GetProviderStats returns statistics for all providers
	GetProviderStats() map[ProviderType]ProviderStats

	// HealthCheck checks health of all providers
	HealthCheck(ctx context.Context) map[ProviderType]error

	// Close closes all providers
	Close() error
}

// ProviderStats represents statistics for a provider
type ProviderStats struct {
	Provider       ProviderType  `json:"provider"`
	TotalRequests  int64         `json:"total_requests"`
	TotalTokens    int64         `json:"total_tokens"`
	TotalCost      float64       `json:"total_cost"`
	CacheHits      int64         `json:"cache_hits"`
	CacheMisses    int64         `json:"cache_misses"`
	Errors         int64         `json:"errors"`
	LastUsed       time.Time     `json:"last_used"`
	AverageLatency time.Duration `json:"average_latency"`
}

// Cache represents an embedding cache
type Cache interface {
	// Get retrieves cached embeddings
	Get(key string) ([][]float64, bool)

	// Set stores embeddings in cache
	Set(key string, embeddings [][]float64, ttl time.Duration) error

	// Delete removes cached embeddings
	Delete(key string) error

	// Clear clears all cached data
	Clear() error

	// Stats returns cache statistics
	Stats() CacheStats

	// Close closes the cache
	Close() error
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	Size        int64   `json:"size"`
	Capacity    int64   `json:"capacity"`
	HitRate     float64 `json:"hit_rate"`
	MemoryUsage int64   `json:"memory_usage"`
}

// Config represents embedding service configuration
type Config struct {
	DefaultProvider ProviderType                    `json:"default_provider"`
	Providers       map[ProviderType]ProviderConfig `json:"providers"`
	Cache           CacheConfig                     `json:"cache"`
	RateLimiting    RateLimitConfig                 `json:"rate_limiting"`
	Retry           RetryConfig                     `json:"retry"`
	Timeout         time.Duration                   `json:"timeout"`
	MaxBatchSize    int                             `json:"max_batch_size"`
	EnableFallback  bool                            `json:"enable_fallback"`
	FallbackOrder   []ProviderType                  `json:"fallback_order"`
}

// ProviderConfig represents configuration for a specific provider
type ProviderConfig struct {
	APIKey     string                 `json:"api_key"`
	BaseURL    string                 `json:"base_url"`
	Timeout    time.Duration          `json:"timeout"`
	MaxRetries int                    `json:"max_retries"`
	RateLimit  RateLimit              `json:"rate_limit"`
	Options    map[string]interface{} `json:"options"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Enabled   bool          `json:"enabled"`
	Type      string        `json:"type"` // memory, redis, etc.
	TTL       time.Duration `json:"ttl"`
	MaxSize   int64         `json:"max_size"`
	MaxMemory int64         `json:"max_memory"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `json:"enabled"`
	RequestsPerMinute int  `json:"requests_per_minute"`
	TokensPerMinute   int  `json:"tokens_per_minute"`
	BurstSize         int  `json:"burst_size"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	Enabled       bool          `json:"enabled"`
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}
