package performance

import (
	"context"
	"time"
)

// CacheStrategy represents different caching strategies
type CacheStrategy string

const (
	CacheStrategyLRU      CacheStrategy = "lru"      // Least Recently Used
	CacheStrategyLFU      CacheStrategy = "lfu"      // Least Frequently Used
	CacheStrategyTTL      CacheStrategy = "ttl"      // Time To Live
	CacheStrategyAdaptive CacheStrategy = "adaptive" // Adaptive based on usage patterns
)

// CacheLevel represents different cache levels
type CacheLevel string

const (
	CacheLevelL1  CacheLevel = "l1"  // Memory cache (fastest)
	CacheLevelL2  CacheLevel = "l2"  // Disk cache (medium)
	CacheLevelL3  CacheLevel = "l3"  // Distributed cache (slower but shared)
	CacheLevelCDN CacheLevel = "cdn" // CDN cache (slowest but global)
)

// CacheItem represents a cached item
type CacheItem struct {
	Key              string                 `json:"key"`
	Value            interface{}            `json:"value"`
	Size             int64                  `json:"size"`
	AccessCount      int64                  `json:"access_count"`
	LastAccessed     time.Time              `json:"last_accessed"`
	CreatedAt        time.Time              `json:"created_at"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	Tags             []string               `json:"tags"`
	Metadata         map[string]interface{} `json:"metadata"`
	Compressed       bool                   `json:"compressed"`
	CompressionRatio float64                `json:"compression_ratio"`
}

// CacheStats represents cache performance statistics
type CacheStats struct {
	TotalItems         int64     `json:"total_items"`
	TotalSize          int64     `json:"total_size"`
	HitCount           int64     `json:"hit_count"`
	MissCount          int64     `json:"miss_count"`
	HitRate            float64   `json:"hit_rate"`
	EvictionCount      int64     `json:"eviction_count"`
	CompressionSavings int64     `json:"compression_savings"`
	AverageLatency     float64   `json:"average_latency"`
	LastUpdated        time.Time `json:"last_updated"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	MaxSize              int64         `json:"max_size"`              // Maximum cache size in bytes
	MaxItems             int64         `json:"max_items"`             // Maximum number of items
	DefaultTTL           time.Duration `json:"default_ttl"`           // Default time to live
	Strategy             CacheStrategy `json:"strategy"`              // Eviction strategy
	Compression          bool          `json:"compression"`           // Enable compression
	CompressionThreshold int64         `json:"compression_threshold"` // Size threshold for compression
	Shards               int           `json:"shards"`                // Number of cache shards
	EnableMetrics        bool          `json:"enable_metrics"`        // Enable performance metrics
}

// CDNConfig represents CDN configuration
type CDNConfig struct {
	Provider      string            `json:"provider"`       // CDN provider (Cloudflare, AWS CloudFront, etc.)
	Endpoint      string            `json:"endpoint"`       // CDN endpoint URL
	APIKey        string            `json:"api_key"`        // CDN API key
	SecretKey     string            `json:"secret_key"`     // CDN secret key
	Region        string            `json:"region"`         // CDN region
	CacheTTL      time.Duration     `json:"cache_ttl"`      // CDN cache TTL
	Compression   bool              `json:"compression"`    // Enable CDN compression
	SSL           bool              `json:"ssl"`            // Enable SSL
	CustomHeaders map[string]string `json:"custom_headers"` // Custom CDN headers
}

// LoadTestConfig represents load testing configuration
type LoadTestConfig struct {
	Duration       time.Duration  `json:"duration"`        // Test duration
	Concurrency    int            `json:"concurrency"`     // Number of concurrent users
	RampUpTime     time.Duration  `json:"ramp_up_time"`    // Time to ramp up to full load
	RampDownTime   time.Duration  `json:"ramp_down_time"`  // Time to ramp down from full load
	RequestRate    int            `json:"request_rate"`    // Requests per second
	TargetLatency  time.Duration  `json:"target_latency"`  // Target response time
	ErrorThreshold float64        `json:"error_threshold"` // Maximum error rate allowed
	TestScenarios  []TestScenario `json:"test_scenarios"`  // Different test scenarios
}

// TestScenario represents a specific test scenario
type TestScenario struct {
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Weight         float64           `json:"weight"` // Weight in the test mix
	Endpoint       string            `json:"endpoint"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	Body           interface{}       `json:"body"`
	ExpectedStatus int               `json:"expected_status"`
}

// LoadTestResult represents load test results
type LoadTestResult struct {
	TestID             string            `json:"test_id"`
	Config             *LoadTestConfig   `json:"config"`
	StartTime          time.Time         `json:"start_time"`
	EndTime            time.Time         `json:"end_time"`
	Duration           time.Duration     `json:"duration"`
	TotalRequests      int64             `json:"total_requests"`
	SuccessfulRequests int64             `json:"successful_requests"`
	FailedRequests     int64             `json:"failed_requests"`
	ErrorRate          float64           `json:"error_rate"`
	AverageLatency     time.Duration     `json:"average_latency"`
	P95Latency         time.Duration     `json:"p95_latency"`
	P99Latency         time.Duration     `json:"p99_latency"`
	Throughput         float64           `json:"throughput"` // Requests per second
	ResourceUsage      *ResourceUsage    `json:"resource_usage"`
	ScenarioResults    []*ScenarioResult `json:"scenario_results"`
	Recommendations    []string          `json:"recommendations"`
}

// ResourceUsage represents system resource usage during testing
type ResourceUsage struct {
	CPUUsage       float64 `json:"cpu_usage"`        // CPU usage percentage
	MemoryUsage    float64 `json:"memory_usage"`     // Memory usage percentage
	DiskIO         float64 `json:"disk_io"`          // Disk I/O operations per second
	NetworkIO      float64 `json:"network_io"`       // Network I/O bytes per second
	GPUMemoryUsage float64 `json:"gpu_memory_usage"` // GPU memory usage percentage
	GPUUtilization float64 `json:"gpu_utilization"`  // GPU utilization percentage
}

// ScenarioResult represents results for a specific test scenario
type ScenarioResult struct {
	ScenarioName       string        `json:"scenario_name"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	ErrorRate          float64       `json:"error_rate"`
	AverageLatency     time.Duration `json:"average_latency"`
	P95Latency         time.Duration `json:"p95_latency"`
	P99Latency         time.Duration `json:"p99_latency"`
	Throughput         float64       `json:"throughput"`
}

// CacheService defines the interface for advanced caching operations
type CacheService interface {
	// Basic cache operations
	Get(ctx context.Context, key string) (*CacheItem, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error

	// Advanced cache operations
	GetMulti(ctx context.Context, keys []string) (map[string]*CacheItem, error)
	SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	DeleteMulti(ctx context.Context, keys []string) error
	Increment(ctx context.Context, key string, value int64) (int64, error)
	Decrement(ctx context.Context, key string, value int64) (int64, error)

	// Cache management
	GetStats(ctx context.Context) (*CacheStats, error)
	GetKeys(ctx context.Context, pattern string) ([]string, error)
	GetSize(ctx context.Context) (int64, error)
	GetItemCount(ctx context.Context) (int64, error)
	Flush(ctx context.Context) error

	// Cache optimization
	Optimize(ctx context.Context) error
	Compress(ctx context.Context, key string) error
	Decompress(ctx context.Context, key string) error
	Prefetch(ctx context.Context, keys []string) error

	// Health check
	HealthCheck(ctx context.Context) error
	Close() error
}

// CDNService defines the interface for CDN operations
type CDNService interface {
	// CDN operations
	Purge(ctx context.Context, urls []string) error
	PurgeAll(ctx context.Context) error
	GetStatus(ctx context.Context, url string) (*CDNStatus, error)
	GetAnalytics(ctx context.Context, timeRange time.Duration) (*CDNAnalytics, error)

	// CDN configuration
	UpdateConfig(ctx context.Context, config *CDNConfig) error
	GetConfig(ctx context.Context) (*CDNConfig, error)
	TestConnection(ctx context.Context) error

	// Health check
	HealthCheck(ctx context.Context) error
	Close() error
}

// CDNStatus represents the status of a CDN URL
type CDNStatus struct {
	URL          string        `json:"url"`
	Status       string        `json:"status"`
	LastPurged   time.Time     `json:"last_purged"`
	CacheHit     bool          `json:"cache_hit"`
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code"`
}

// CDNAnalytics represents CDN analytics data
type CDNAnalytics struct {
	TimeRange      time.Duration `json:"time_range"`
	TotalRequests  int64         `json:"total_requests"`
	CacheHits      int64         `json:"cache_hits"`
	CacheMisses    int64         `json:"cache_misses"`
	HitRate        float64       `json:"hit_rate"`
	Bandwidth      int64         `json:"bandwidth"`
	ErrorRate      float64       `json:"error_rate"`
	AverageLatency time.Duration `json:"average_latency"`
}

// LoadTestService defines the interface for load testing operations
type LoadTestService interface {
	// Load testing operations
	RunTest(ctx context.Context, config *LoadTestConfig) (*LoadTestResult, error)
	StopTest(ctx context.Context, testID string) error
	GetTestStatus(ctx context.Context, testID string) (*TestStatus, error)
	GetTestResults(ctx context.Context, testID string) (*LoadTestResult, error)

	// Test management
	ListTests(ctx context.Context) ([]*TestInfo, error)
	DeleteTest(ctx context.Context, testID string) error
	ExportResults(ctx context.Context, testID string, format string) ([]byte, error)

	// Health check
	HealthCheck(ctx context.Context) error
	Close() error
}

// TestStatus represents the current status of a load test
type TestStatus struct {
	TestID            string        `json:"test_id"`
	Status            string        `json:"status"`   // running, completed, failed, stopped
	Progress          float64       `json:"progress"` // 0.0 to 1.0
	StartTime         time.Time     `json:"start_time"`
	EstimatedEnd      time.Time     `json:"estimated_end"`
	EndTime           *time.Time    `json:"end_time,omitempty"`
	CurrentLoad       int           `json:"current_load"` // Current concurrent users
	CurrentLatency    time.Duration `json:"current_latency"`
	CurrentThroughput float64       `json:"current_throughput"`
	ErrorCount        int64         `json:"error_count"`
}

// TestInfo represents basic information about a load test
type TestInfo struct {
	TestID         string        `json:"test_id"`
	Name           string        `json:"name"`
	Status         string        `json:"status"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        *time.Time    `json:"end_time,omitempty"`
	Duration       time.Duration `json:"duration"`
	TotalRequests  int64         `json:"total_requests"`
	ErrorRate      float64       `json:"error_rate"`
	AverageLatency time.Duration `json:"average_latency"`
}
