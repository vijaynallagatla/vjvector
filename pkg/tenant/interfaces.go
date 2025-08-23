package tenant

import (
	"context"
	"time"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Domain      string            `json:"domain,omitempty"`
	Status      TenantStatus      `json:"status"`
	Plan        string            `json:"plan"` // "basic", "pro", "enterprise"
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Settings    *TenantSettings   `json:"settings"`
	Quotas      *TenantQuotas     `json:"quotas"`
	Usage       *TenantUsage      `json:"usage"`
}

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusPending   TenantStatus = "pending"
)

// TenantSettings represents configurable settings for a tenant
type TenantSettings struct {
	// API Configuration
	APIRateLimit       int           `json:"api_rate_limit"` // requests per minute
	MaxConcurrentUsers int           `json:"max_concurrent_users"`
	SessionTimeout     time.Duration `json:"session_timeout"`

	// Security Settings
	IPWhitelist    []string `json:"ip_whitelist"`
	RequireMFA     bool     `json:"require_mfa"`
	PasswordPolicy string   `json:"password_policy"` // "basic", "strong", "enterprise"

	// Feature Flags
	EnableAdvancedRAG  bool `json:"enable_advanced_rag"`
	EnableCustomModels bool `json:"enable_custom_models"`
	EnableAnalytics    bool `json:"enable_analytics"`

	// Integration Settings
	WebhookURLs    []string `json:"webhook_urls"`
	OAuthProviders []string `json:"oauth_providers"`
	LDAPEnabled    bool     `json:"ldap_enabled"`

	// Compliance Settings
	DataRetentionDays int  `json:"data_retention_days"`
	AuditLogging      bool `json:"audit_logging"`
	GDPRCompliance    bool `json:"gdpr_compliance"`
}

// TenantQuotas represents resource quotas for a tenant
type TenantQuotas struct {
	// Storage Quotas
	MaxCollections int   `json:"max_collections"`
	MaxVectors     int64 `json:"max_vectors"`
	MaxStorageGB   int64 `json:"max_storage_gb"`

	// Compute Quotas
	MaxCPU            int `json:"max_cpu"` // CPU cores
	MaxMemoryGB       int `json:"max_memory_gb"`
	MaxConcurrentJobs int `json:"max_concurrent_jobs"`

	// API Quotas
	MaxAPICallsPerDay  int64 `json:"max_api_calls_per_day"`
	MaxAPICallsPerHour int64 `json:"max_api_calls_per_hour"`
	MaxAPICallsPerMin  int64 `json:"max_api_calls_per_min"`

	// User Quotas
	MaxUsers   int `json:"max_users"`
	MaxAPIKeys int `json:"max_api_keys"`
}

// TenantUsage represents current resource usage for a tenant
type TenantUsage struct {
	// Storage Usage
	CollectionsCount int   `json:"collections_count"`
	VectorsCount     int64 `json:"vectors_count"`
	StorageUsedGB    int64 `json:"storage_used_gb"`

	// Compute Usage
	CPUUsage      float64 `json:"cpu_usage"` // percentage
	MemoryUsageGB float64 `json:"memory_usage_gb"`
	ActiveJobs    int     `json:"active_jobs"`

	// API Usage
	APICallsToday    int64 `json:"api_calls_today"`
	APICallsThisHour int64 `json:"api_calls_this_hour"`
	APICallsThisMin  int64 `json:"api_calls_this_min"`

	// User Usage
	ActiveUsers   int `json:"active_users"`
	ActiveAPIKeys int `json:"active_api_keys"`

	// Last Updated
	LastUpdated time.Time `json:"last_updated"`
}

// TenantContext represents the tenant context for a request
type TenantContext struct {
	TenantID   string            `json:"tenant_id"`
	TenantName string            `json:"tenant_name"`
	UserID     string            `json:"user_id,omitempty"`
	APIKey     string            `json:"api_key,omitempty"`
	IPAddress  string            `json:"ip_address"`
	UserAgent  string            `json:"user_agent"`
	RequestID  string            `json:"request_id"`
	Timestamp  time.Time         `json:"timestamp"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// TenantService defines the main tenant management service interface
type TenantService interface {
	// Tenant Management
	CreateTenant(ctx context.Context, tenant *Tenant) error
	GetTenant(ctx context.Context, tenantID string) (*Tenant, error)
	GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error)
	UpdateTenant(ctx context.Context, tenant *Tenant) error
	DeleteTenant(ctx context.Context, tenantID string) error
	ListTenants(ctx context.Context, limit, offset int) ([]*Tenant, error)

	// Tenant Lifecycle
	ActivateTenant(ctx context.Context, tenantID string) error
	SuspendTenant(ctx context.Context, tenantID string, reason string) error
	DeactivateTenant(ctx context.Context, tenantID string) error

	// Settings Management
	UpdateTenantSettings(ctx context.Context, tenantID string, settings *TenantSettings) error
	GetTenantSettings(ctx context.Context, tenantID string) (*TenantSettings, error)

	// Quota Management
	UpdateTenantQuotas(ctx context.Context, tenantID string, quotas *TenantQuotas) error
	GetTenantQuotas(ctx context.Context, tenantID string) (*TenantQuotas, error)
	CheckQuota(ctx context.Context, tenantID string, resource string, amount int64) (bool, error)

	// Usage Tracking
	UpdateTenantUsage(ctx context.Context, tenantID string, usage *TenantUsage) error
	GetTenantUsage(ctx context.Context, tenantID string) (*TenantUsage, error)
	ResetUsageCounters(ctx context.Context, tenantID string) error

	// Health Check
	HealthCheck(ctx context.Context) error
	Close() error
}

// TenantIsolationService defines the service for ensuring tenant data isolation
type TenantIsolationService interface {
	// Context Management
	CreateTenantContext(ctx context.Context, tenantID, userID, apiKey, ipAddress, userAgent, requestID string) (*TenantContext, error)
	ValidateTenantContext(ctx context.Context, tenantContext *TenantContext) error

	// Data Isolation
	IsolateCollection(ctx context.Context, tenantID, collectionID string) error
	IsolateVector(ctx context.Context, tenantID, vectorID string) error
	ValidateTenantAccess(ctx context.Context, tenantID, resourceType, resourceID string) error

	// Resource Scoping
	ScopeToTenant(ctx context.Context, tenantID string) context.Context
	GetTenantFromContext(ctx context.Context) (string, error)

	// Health Check
	HealthCheck(ctx context.Context) error
	Close() error
}

// TenantAnalyticsService defines the service for tenant analytics and reporting
type TenantAnalyticsService interface {
	// Usage Analytics
	TrackAPICall(ctx context.Context, tenantID, endpoint, method string, duration time.Duration, success bool) error
	TrackResourceUsage(ctx context.Context, tenantID, resourceType string, amount int64) error
	TrackUserActivity(ctx context.Context, tenantID, userID, action string) error

	// Reporting
	GetTenantReport(ctx context.Context, tenantID string, startDate, endDate time.Time) (*TenantReport, error)
	GetSystemReport(ctx context.Context, startDate, endDate time.Time) (*SystemReport, error)
	GetQuotaUtilizationReport(ctx context.Context, tenantID string) (*QuotaUtilizationReport, error)

	// Health Check
	HealthCheck(ctx context.Context) error
	Close() error
}

// TenantReport represents a comprehensive report for a tenant
type TenantReport struct {
	TenantID    string    `json:"tenant_id"`
	TenantName  string    `json:"tenant_name"`
	Period      string    `json:"period"`
	GeneratedAt time.Time `json:"generated_at"`

	// Usage Summary
	TotalAPICalls int64 `json:"total_api_calls"`
	TotalUsers    int   `json:"total_users"`
	TotalStorage  int64 `json:"total_storage_gb"`

	// Performance Metrics
	AvgResponseTime float64 `json:"avg_response_time_ms"`
	SuccessRate     float64 `json:"success_rate"`
	ErrorRate       float64 `json:"error_rate"`

	// Resource Utilization
	CPUUtilization     float64 `json:"cpu_utilization_percent"`
	MemoryUtilization  float64 `json:"memory_utilization_percent"`
	StorageUtilization float64 `json:"storage_utilization_percent"`

	// Top Endpoints
	TopEndpoints []EndpointUsage `json:"top_endpoints"`

	// User Activity
	UserActivity []UserActivity `json:"user_activity"`

	// Cost Analysis
	EstimatedCost float64            `json:"estimated_cost"`
	CostBreakdown map[string]float64 `json:"cost_breakdown"`
}

// SystemReport represents a system-wide report
type SystemReport struct {
	Period      string    `json:"period"`
	GeneratedAt time.Time `json:"generated_at"`

	// System Overview
	TotalTenants  int   `json:"total_tenants"`
	ActiveTenants int   `json:"active_tenants"`
	TotalUsers    int   `json:"total_users"`
	TotalVectors  int64 `json:"total_vectors"`

	// Performance Metrics
	SystemUptime    float64 `json:"system_uptime_percent"`
	AvgResponseTime float64 `json:"avg_response_time_ms"`
	TotalAPICalls   int64   `json:"total_api_calls"`

	// Resource Utilization
	TotalCPUUsage     float64 `json:"total_cpu_usage_percent"`
	TotalMemoryUsage  float64 `json:"total_memory_usage_percent"`
	TotalStorageUsage float64 `json:"total_storage_usage_percent"`

	// Top Tenants
	TopTenants []TenantUsage `json:"top_tenants"`
}

// QuotaUtilizationReport represents quota utilization for a tenant
type QuotaUtilizationReport struct {
	TenantID    string    `json:"tenant_id"`
	GeneratedAt time.Time `json:"generated_at"`

	// Storage Quotas
	CollectionsUtilization float64 `json:"collections_utilization_percent"`
	VectorsUtilization     float64 `json:"vectors_utilization_percent"`
	StorageUtilization     float64 `json:"storage_utilization_percent"`

	// Compute Quotas
	CPUUtilization    float64 `json:"cpu_utilization_percent"`
	MemoryUtilization float64 `json:"memory_utilization_percent"`
	JobsUtilization   float64 `json:"jobs_utilization_percent"`

	// API Quotas
	DailyAPICallsUtilization  float64 `json:"daily_api_calls_utilization_percent"`
	HourlyAPICallsUtilization float64 `json:"hourly_api_calls_utilization_percent"`
	MinuteAPICallsUtilization float64 `json:"minute_api_calls_utilization_percent"`

	// User Quotas
	UsersUtilization   float64 `json:"users_utilization_percent"`
	APIKeysUtilization float64 `json:"api_keys_utilization_percent"`

	// Recommendations
	Recommendations []string `json:"recommendations"`
}

// EndpointUsage represents usage statistics for an API endpoint
type EndpointUsage struct {
	Endpoint    string  `json:"endpoint"`
	Method      string  `json:"method"`
	CallCount   int64   `json:"call_count"`
	AvgDuration float64 `json:"avg_duration_ms"`
	SuccessRate float64 `json:"success_rate"`
	ErrorRate   float64 `json:"error_rate"`
}

// UserActivity represents user activity statistics
type UserActivity struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	LoginCount    int       `json:"login_count"`
	LastLogin     time.Time `json:"last_login"`
	APICallCount  int64     `json:"api_call_count"`
	ActiveMinutes int       `json:"active_minutes"`
}
