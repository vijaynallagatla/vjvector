package integration

import (
	"context"
	"time"
)

// IntegrationType represents the type of external integration
type IntegrationType string

const (
	IntegrationTypeAPI          IntegrationType = "api"           // REST/GraphQL API integration
	IntegrationTypeDatabase     IntegrationType = "database"      // Database integration
	IntegrationTypeMessageQueue IntegrationType = "message_queue" // Message queue integration
	IntegrationTypeFileSystem   IntegrationType = "file_system"   // File system integration
	IntegrationTypeCloud        IntegrationType = "cloud"         // Cloud service integration
	IntegrationTypeMonitoring   IntegrationType = "monitoring"    // Monitoring system integration
	IntegrationTypeAnalytics    IntegrationType = "analytics"     // Analytics platform integration
)

// IntegrationStatus represents the status of an integration
type IntegrationStatus string

const (
	IntegrationStatusActive      IntegrationStatus = "active"      // Integration is active and working
	IntegrationStatusInactive    IntegrationStatus = "inactive"    // Integration is inactive
	IntegrationStatusError       IntegrationStatus = "error"       // Integration has errors
	IntegrationStatusPending     IntegrationStatus = "pending"     // Integration is being set up
	IntegrationStatusTesting     IntegrationStatus = "testing"     // Integration is being tested
	IntegrationStatusMaintenance IntegrationStatus = "maintenance" // Integration is under maintenance
)

// IntegrationConfig represents configuration for an external integration
type IntegrationConfig struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Type           IntegrationType        `json:"type"`
	Provider       string                 `json:"provider"`         // Integration provider (e.g., "AWS", "Google", "Azure")
	Endpoint       string                 `json:"endpoint"`         // Integration endpoint URL
	Credentials    map[string]interface{} `json:"credentials"`      // Integration credentials
	Settings       map[string]interface{} `json:"settings"`         // Integration-specific settings
	Headers        map[string]string      `json:"headers"`          // Custom headers for API calls
	Timeout        time.Duration          `json:"timeout"`          // Request timeout
	RetryCount     int                    `json:"retry_count"`      // Number of retry attempts
	RetryDelay     time.Duration          `json:"retry_delay"`      // Delay between retries
	RateLimit      int                    `json:"rate_limit"`       // Rate limit (requests per second)
	HealthCheckURL string                 `json:"health_check_url"` // Health check endpoint
	Enabled        bool                   `json:"enabled"`          // Whether integration is enabled
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// IntegrationHealth represents the health status of an integration
type IntegrationHealth struct {
	IntegrationID string                 `json:"integration_id"`
	Status        IntegrationStatus      `json:"status"`
	LastCheck     time.Time              `json:"last_check"`
	ResponseTime  time.Duration          `json:"response_time"`
	ErrorCount    int64                  `json:"error_count"`
	SuccessCount  int64                  `json:"success_count"`
	LastError     string                 `json:"last_error,omitempty"`
	ErrorRate     float64                `json:"error_rate"`
	Uptime        float64                `json:"uptime"`  // Percentage uptime
	Metrics       map[string]interface{} `json:"metrics"` // Integration-specific metrics
}

// IntegrationEvent represents an event from an external integration
type IntegrationEvent struct {
	ID            string                 `json:"id"`
	IntegrationID string                 `json:"integration_id"`
	EventType     string                 `json:"event_type"`
	EventData     map[string]interface{} `json:"event_data"`
	Timestamp     time.Time              `json:"timestamp"`
	Severity      string                 `json:"severity"`  // info, warning, error, critical
	Processed     bool                   `json:"processed"` // Whether event has been processed
}

// PluginType represents the type of plugin
type PluginType string

const (
	PluginTypeProcessor   PluginType = "processor"   // Data processing plugin
	PluginTypeConnector   PluginType = "connector"   // External system connector
	PluginTypeTransformer PluginType = "transformer" // Data transformation plugin
	PluginTypeValidator   PluginType = "validator"   // Data validation plugin
	PluginTypeAnalyzer    PluginType = "analyzer"    // Data analysis plugin
	PluginTypeRenderer    PluginType = "renderer"    // Data rendering plugin
	PluginTypeCustom      PluginType = "custom"      // Custom functionality plugin
)

// PluginStatus represents the status of a plugin
type PluginStatus string

const (
	PluginStatusActive    PluginStatus = "active"    // Plugin is active and working
	PluginStatusInactive  PluginStatus = "inactive"  // Plugin is inactive
	PluginStatusError     PluginStatus = "error"     // Plugin has errors
	PluginStatusLoading   PluginStatus = "loading"   // Plugin is being loaded
	PluginStatusUnloading PluginStatus = "unloading" // Plugin is being unloaded
	PluginStatusDisabled  PluginStatus = "disabled"  // Plugin is disabled
)

// Plugin represents a plugin in the system
type Plugin struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Type          PluginType             `json:"type"`
	Version       string                 `json:"version"`
	Author        string                 `json:"author"`
	Repository    string                 `json:"repository,omitempty"`
	License       string                 `json:"license,omitempty"`
	Status        PluginStatus           `json:"status"`
	Enabled       bool                   `json:"enabled"`
	Config        map[string]interface{} `json:"config"`
	EntryPoint    string                 `json:"entry_point"`  // Plugin entry point function
	API           map[string]interface{} `json:"api"`          // Plugin API endpoints
	Dependencies  []string               `json:"dependencies"` // Required dependencies
	Permissions   []string               `json:"permissions"`  // Required permissions
	ResourceUsage *PluginResourceUsage   `json:"resource_usage"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	LastLoaded    *time.Time             `json:"last_loaded,omitempty"`
}

// PluginResourceUsage represents resource usage of a plugin
type PluginResourceUsage struct {
	MemoryUsage   int64     `json:"memory_usage"`   // Memory usage in bytes
	CPUUsage      float64   `json:"cpu_usage"`      // CPU usage percentage
	DiskUsage     int64     `json:"disk_usage"`     // Disk usage in bytes
	NetworkIO     int64     `json:"network_io"`     // Network I/O in bytes
	ActiveThreads int       `json:"active_threads"` // Number of active threads
	LastUpdated   time.Time `json:"last_updated"`
}

// PluginExecution represents the execution of a plugin
type PluginExecution struct {
	ID          string                 `json:"id"`
	PluginID    string                 `json:"plugin_id"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output"`
	Status      string                 `json:"status"` // running, completed, failed, cancelled
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
	Logs        []string               `json:"logs"`
	Performance *PluginPerformance     `json:"performance,omitempty"`
}

// PluginPerformance represents performance metrics of a plugin execution
type PluginPerformance struct {
	MemoryPeak    int64         `json:"memory_peak"`    // Peak memory usage
	CPUPeak       float64       `json:"cpu_peak"`       // Peak CPU usage
	ExecutionTime time.Duration `json:"execution_time"` // Total execution time
	Throughput    float64       `json:"throughput"`     // Operations per second
}

// MarketplaceItem represents an item in the plugin marketplace
type MarketplaceItem struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Type          PluginType        `json:"type"`
	Category      string            `json:"category"`
	Tags          []string          `json:"tags"`
	Version       string            `json:"version"`
	Author        string            `json:"author"`
	Repository    string            `json:"repository"`
	License       string            `json:"license"`
	Price         *MarketplacePrice `json:"price,omitempty"`
	Rating        float64           `json:"rating"`         // Average rating (1-5)
	ReviewCount   int               `json:"review_count"`   // Number of reviews
	DownloadCount int64             `json:"download_count"` // Number of downloads
	Compatibility []string          `json:"compatibility"`  // Compatible versions
	Screenshots   []string          `json:"screenshots"`    // Screenshot URLs
	Documentation string            `json:"documentation"`  // Documentation URL
	Support       string            `json:"support"`        // Support contact
	LastUpdated   time.Time         `json:"last_updated"`
	PublishedAt   time.Time         `json:"published_at"`
	Status        string            `json:"status"`   // active, deprecated, beta
	Featured      bool              `json:"featured"` // Whether item is featured
}

// MarketplacePrice represents pricing information for a marketplace item
type MarketplacePrice struct {
	Amount   float64 `json:"amount"`           // Price amount
	Currency string  `json:"currency"`         // Price currency (e.g., "USD")
	Model    string  `json:"model"`            // Pricing model (e.g., "one-time", "subscription", "usage-based")
	Period   string  `json:"period,omitempty"` // Billing period for subscriptions
}

// MarketplaceReview represents a review for a marketplace item
type MarketplaceReview struct {
	ID        string     `json:"id"`
	ItemID    string     `json:"item_id"`
	UserID    string     `json:"user_id"`
	Rating    int        `json:"rating"` // Rating (1-5)
	Title     string     `json:"title"`
	Comment   string     `json:"comment"`
	Pros      []string   `json:"pros"`
	Cons      []string   `json:"cons"`
	Helpful   int        `json:"helpful"` // Number of helpful votes
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// IntegrationService defines the interface for external integrations
type IntegrationService interface {
	// Integration management
	CreateIntegration(ctx context.Context, config *IntegrationConfig) (*IntegrationConfig, error)
	GetIntegration(ctx context.Context, id string) (*IntegrationConfig, error)
	UpdateIntegration(ctx context.Context, id string, config *IntegrationConfig) (*IntegrationConfig, error)
	DeleteIntegration(ctx context.Context, id string) error
	ListIntegrations(ctx context.Context, filter *IntegrationFilter) ([]*IntegrationConfig, error)

	// Integration operations
	TestIntegration(ctx context.Context, id string) (*IntegrationHealth, error)
	EnableIntegration(ctx context.Context, id string) error
	DisableIntegration(ctx context.Context, id string) error
	GetIntegrationHealth(ctx context.Context, id string) (*IntegrationHealth, error)
	GetIntegrationEvents(ctx context.Context, id string, limit int) ([]*IntegrationEvent, error)

	// Integration execution
	ExecuteIntegration(ctx context.Context, id string, input map[string]interface{}) (map[string]interface{}, error)
	ScheduleIntegration(ctx context.Context, id string, schedule string, input map[string]interface{}) (string, error)
	CancelScheduledIntegration(ctx context.Context, scheduleID string) error

	// Health check
	HealthCheck(ctx context.Context) error
	Close() error
}

// IntegrationFilter represents filters for listing integrations
type IntegrationFilter struct {
	Type     *IntegrationType   `json:"type,omitempty"`
	Provider *string            `json:"provider,omitempty"`
	Status   *IntegrationStatus `json:"status,omitempty"`
	Enabled  *bool              `json:"enabled,omitempty"`
	Tags     []string           `json:"tags,omitempty"`
	Limit    int                `json:"limit,omitempty"`
	Offset   int                `json:"offset,omitempty"`
}

// PluginService defines the interface for plugin management
type PluginService interface {
	// Plugin management
	InstallPlugin(ctx context.Context, source string, config map[string]interface{}) (*Plugin, error)
	UninstallPlugin(ctx context.Context, id string) error
	EnablePlugin(ctx context.Context, id string) error
	DisablePlugin(ctx context.Context, id string) error
	UpdatePlugin(ctx context.Context, id string) (*Plugin, error)

	// Plugin information
	GetPlugin(ctx context.Context, id string) (*Plugin, error)
	ListPlugins(ctx context.Context, filter *PluginFilter) ([]*Plugin, error)
	GetPluginStatus(ctx context.Context, id string) (*Plugin, error)

	// Plugin execution
	ExecutePlugin(ctx context.Context, id string, input map[string]interface{}) (*PluginExecution, error)
	GetPluginExecution(ctx context.Context, executionID string) (*PluginExecution, error)
	ListPluginExecutions(ctx context.Context, pluginID string, limit int) ([]*PluginExecution, error)

	// Plugin configuration
	UpdatePluginConfig(ctx context.Context, id string, config map[string]interface{}) (*Plugin, error)
	ValidatePluginConfig(ctx context.Context, id string, config map[string]interface{}) error

	// Health check
	HealthCheck(ctx context.Context) error
	Close() error
}

// PluginFilter represents filters for listing plugins
type PluginFilter struct {
	Type    *PluginType   `json:"type,omitempty"`
	Status  *PluginStatus `json:"status,omitempty"`
	Enabled *bool         `json:"enabled,omitempty"`
	Author  *string       `json:"author,omitempty"`
	Tags    []string      `json:"tags,omitempty"`
	Limit   int           `json:"limit,omitempty"`
	Offset  int           `json:"offset,omitempty"`
}

// MarketplaceService defines the interface for the plugin marketplace
type MarketplaceService interface {
	// Marketplace browsing
	BrowseItems(ctx context.Context, filter *MarketplaceFilter) ([]*MarketplaceItem, error)
	GetItem(ctx context.Context, id string) (*MarketplaceItem, error)
	SearchItems(ctx context.Context, query string, filter *MarketplaceFilter) ([]*MarketplaceItem, error)
	GetFeaturedItems(ctx context.Context, limit int) ([]*MarketplaceItem, error)
	GetPopularItems(ctx context.Context, limit int) ([]*MarketplaceItem, error)

	// Item management
	PublishItem(ctx context.Context, item *MarketplaceItem) (*MarketplaceItem, error)
	UpdateItem(ctx context.Context, id string, item *MarketplaceItem) (*MarketplaceItem, error)
	DeprecateItem(ctx context.Context, id string) error
	DeleteItem(ctx context.Context, id string) error

	// Reviews and ratings
	AddReview(ctx context.Context, itemID string, review *MarketplaceReview) (*MarketplaceReview, error)
	GetReviews(ctx context.Context, itemID string, limit int) ([]*MarketplaceReview, error)
	UpdateReview(ctx context.Context, reviewID string, review *MarketplaceReview) (*MarketplaceReview, error)
	DeleteReview(ctx context.Context, reviewID string) error

	// Downloads and analytics
	DownloadItem(ctx context.Context, itemID string) ([]byte, error)
	GetItemAnalytics(ctx context.Context, itemID string) (map[string]interface{}, error)

	// Health check
	HealthCheck(ctx context.Context) error
	Close() error
}

// MarketplaceFilter represents filters for marketplace items
type MarketplaceFilter struct {
	Type       *PluginType `json:"type,omitempty"`
	Category   *string     `json:"category,omitempty"`
	Author     *string     `json:"author,omitempty"`
	Tags       []string    `json:"tags,omitempty"`
	MinRating  *float64    `json:"min_rating,omitempty"`
	PriceRange *PriceRange `json:"price_range,omitempty"`
	Status     *string     `json:"status,omitempty"`
	Featured   *bool       `json:"featured,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	Offset     int         `json:"offset,omitempty"`
}

// PriceRange represents a range of prices
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}
