package tenant

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultTenantManager implements the tenant management service
type DefaultTenantManager struct {
	tenants map[string]*Tenant
	mu      sync.RWMutex
}

// NewDefaultTenantManager creates a new default tenant manager
func NewDefaultTenantManager() *DefaultTenantManager {
	return &DefaultTenantManager{
		tenants: make(map[string]*Tenant),
	}
}

// CreateTenant creates a new tenant
func (m *DefaultTenantManager) CreateTenant(ctx context.Context, tenant *Tenant) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tenant.ID == "" {
		tenant.ID = m.generateTenantID()
	}

	// Set default values
	now := time.Now()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now
	tenant.Status = TenantStatusActive

	// Set default settings if not provided
	if tenant.Settings == nil {
		tenant.Settings = &TenantSettings{
			APIRateLimit:       100,
			MaxConcurrentUsers: 10,
			SessionTimeout:     30 * time.Minute,
			RequireMFA:         false,
			PasswordPolicy:     "basic",
			EnableAdvancedRAG:  false,
			EnableCustomModels: false,
			EnableAnalytics:    true,
			DataRetentionDays:  365,
			AuditLogging:       true,
			GDPRCompliance:     false,
		}
	}

	// Set default quotas if not provided
	if tenant.Quotas == nil {
		tenant.Quotas = &TenantQuotas{
			MaxCollections:     10,
			MaxVectors:         100000,
			MaxStorageGB:       100,
			MaxCPU:             2,
			MaxMemoryGB:        4,
			MaxConcurrentJobs:  5,
			MaxAPICallsPerDay:  100000,
			MaxAPICallsPerHour: 10000,
			MaxAPICallsPerMin:  1000,
			MaxUsers:           10,
			MaxAPIKeys:         5,
		}
	}

	// Set default usage if not provided
	if tenant.Usage == nil {
		tenant.Usage = &TenantUsage{
			CollectionsCount: 0,
			VectorsCount:     0,
			StorageUsedGB:    0,
			CPUUsage:         0,
			MemoryUsageGB:    0,
			ActiveJobs:       0,
			APICallsToday:    0,
			APICallsThisHour: 0,
			APICallsThisMin:  0,
			ActiveUsers:      0,
			ActiveAPIKeys:    0,
			LastUpdated:      now,
		}
	}

	// Validate tenant
	if err := m.validateTenant(tenant); err != nil {
		return fmt.Errorf("invalid tenant: %w", err)
	}

	// Check if tenant already exists
	if _, exists := m.tenants[tenant.ID]; exists {
		return fmt.Errorf("tenant already exists: %s", tenant.ID)
	}

	// Store tenant
	m.tenants[tenant.ID] = tenant

	return nil
}

// GetTenant retrieves a tenant by ID
func (m *DefaultTenantManager) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}

	return tenant, nil
}

// GetTenantByDomain retrieves a tenant by domain
func (m *DefaultTenantManager) GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, tenant := range m.tenants {
		if tenant.Domain == domain {
			return tenant, nil
		}
	}

	return nil, fmt.Errorf("tenant not found for domain: %s", domain)
}

// UpdateTenant updates an existing tenant
func (m *DefaultTenantManager) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tenants[tenant.ID]; !exists {
		return fmt.Errorf("tenant not found: %s", tenant.ID)
	}

	tenant.UpdatedAt = time.Now()

	// Validate tenant
	if err := m.validateTenant(tenant); err != nil {
		return fmt.Errorf("invalid tenant: %w", err)
	}

	m.tenants[tenant.ID] = tenant
	return nil
}

// DeleteTenant deletes a tenant
func (m *DefaultTenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tenants[tenantID]; !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	delete(m.tenants, tenantID)
	return nil
}

// ListTenants lists tenants with pagination
func (m *DefaultTenantManager) ListTenants(ctx context.Context, limit, offset int) ([]*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tenantList []*Tenant
	for _, tenant := range m.tenants {
		tenantList = append(tenantList, tenant)
	}

	// Simple pagination (in production, use database pagination)
	if offset >= len(tenantList) {
		return []*Tenant{}, nil
	}

	end := offset + limit
	if end > len(tenantList) {
		end = len(tenantList)
	}

	return tenantList[offset:end], nil
}

// ActivateTenant activates a tenant
func (m *DefaultTenantManager) ActivateTenant(ctx context.Context, tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	tenant.Status = TenantStatusActive
	tenant.UpdatedAt = time.Now()

	return nil
}

// SuspendTenant suspends a tenant
func (m *DefaultTenantManager) SuspendTenant(ctx context.Context, tenantID string, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	tenant.Status = TenantStatusSuspended
	tenant.UpdatedAt = time.Now()
	if tenant.Metadata == nil {
		tenant.Metadata = make(map[string]string)
	}
	tenant.Metadata["suspension_reason"] = reason
	tenant.Metadata["suspended_at"] = time.Now().Format(time.RFC3339)

	return nil
}

// DeactivateTenant deactivates a tenant
func (m *DefaultTenantManager) DeactivateTenant(ctx context.Context, tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	tenant.Status = TenantStatusInactive
	tenant.UpdatedAt = time.Now()

	return nil
}

// UpdateTenantSettings updates tenant settings
func (m *DefaultTenantManager) UpdateTenantSettings(ctx context.Context, tenantID string, settings *TenantSettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	tenant.Settings = settings
	tenant.UpdatedAt = time.Now()

	return nil
}

// GetTenantSettings gets tenant settings
func (m *DefaultTenantManager) GetTenantSettings(ctx context.Context, tenantID string) (*TenantSettings, error) {
	tenant, err := m.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return tenant.Settings, nil
}

// UpdateTenantQuotas updates tenant quotas
func (m *DefaultTenantManager) UpdateTenantQuotas(ctx context.Context, tenantID string, quotas *TenantQuotas) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	tenant.Quotas = quotas
	tenant.UpdatedAt = time.Now()

	return nil
}

// GetTenantQuotas gets tenant quotas
func (m *DefaultTenantManager) GetTenantQuotas(ctx context.Context, tenantID string) (*TenantQuotas, error) {
	tenant, err := m.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return tenant.Quotas, nil
}

// CheckQuota checks if a tenant has quota for a resource
func (m *DefaultTenantManager) CheckQuota(ctx context.Context, tenantID string, resource string, amount int64) (bool, error) {
	tenant, err := m.GetTenant(ctx, tenantID)
	if err != nil {
		return false, err
	}

	switch resource {
	case "collections":
		return tenant.Usage.CollectionsCount+int(amount) <= tenant.Quotas.MaxCollections, nil
	case "vectors":
		return tenant.Usage.VectorsCount+amount <= tenant.Quotas.MaxVectors, nil
	case "storage":
		return tenant.Usage.StorageUsedGB+amount <= tenant.Quotas.MaxStorageGB, nil
	case "users":
		return tenant.Usage.ActiveUsers+int(amount) <= tenant.Quotas.MaxUsers, nil
	case "api_keys":
		return tenant.Usage.ActiveAPIKeys+int(amount) <= tenant.Quotas.MaxAPIKeys, nil
	default:
		return false, fmt.Errorf("unknown resource: %s", resource)
	}
}

// UpdateTenantUsage updates tenant usage
func (m *DefaultTenantManager) UpdateTenantUsage(ctx context.Context, tenantID string, usage *TenantUsage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	usage.LastUpdated = time.Now()
	tenant.Usage = usage
	tenant.UpdatedAt = time.Now()

	return nil
}

// GetTenantUsage gets tenant usage
func (m *DefaultTenantManager) GetTenantUsage(ctx context.Context, tenantID string) (*TenantUsage, error) {
	tenant, err := m.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return tenant.Usage, nil
}

// ResetUsageCounters resets tenant usage counters
func (m *DefaultTenantManager) ResetUsageCounters(ctx context.Context, tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	now := time.Now()
	tenant.Usage.APICallsToday = 0
	tenant.Usage.APICallsThisHour = 0
	tenant.Usage.APICallsThisMin = 0
	tenant.Usage.LastUpdated = now
	tenant.UpdatedAt = now

	return nil
}

// HealthCheck performs a health check on the service
func (m *DefaultTenantManager) HealthCheck(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tenants == nil {
		return fmt.Errorf("tenant manager not initialized")
	}

	return nil
}

// Close performs cleanup operations
func (m *DefaultTenantManager) Close() error {
	// Clean up any resources if needed
	return nil
}

// Helper methods

// validateTenant validates tenant data
func (m *DefaultTenantManager) validateTenant(tenant *Tenant) error {
	if tenant.Name == "" {
		return fmt.Errorf("tenant name is required")
	}

	if tenant.Settings != nil {
		if tenant.Settings.APIRateLimit <= 0 {
			return fmt.Errorf("API rate limit must be positive")
		}
		if tenant.Settings.MaxConcurrentUsers <= 0 {
			return fmt.Errorf("max concurrent users must be positive")
		}
	}

	if tenant.Quotas != nil {
		if tenant.Quotas.MaxCollections <= 0 {
			return fmt.Errorf("max collections must be positive")
		}
		if tenant.Quotas.MaxVectors <= 0 {
			return fmt.Errorf("max vectors must be positive")
		}
		if tenant.Quotas.MaxStorageGB <= 0 {
			return fmt.Errorf("max storage must be positive")
		}
	}

	return nil
}

// generateTenantID generates a unique tenant ID
func (m *DefaultTenantManager) generateTenantID() string {
	// Simple ID generation (in production, use UUID or similar)
	return fmt.Sprintf("tenant_%d", time.Now().UnixNano())
}
