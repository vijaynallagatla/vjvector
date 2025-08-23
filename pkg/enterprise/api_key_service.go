package enterprise

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// APIKey represents an API key for a tenant
type APIKey struct {
	ID                    string            `json:"id"`
	TenantID              string            `json:"tenant_id"`
	Name                  string            `json:"name"`
	Description           string            `json:"description"`
	KeyHash               string            `json:"-"`          // Never expose the actual key
	KeyPrefix             string            `json:"key_prefix"` // First 8 characters for identification
	Permissions           []string          `json:"permissions"`
	Scopes                []string          `json:"scopes"` // Resource scopes (e.g., collections, vectors)
	Status                APIKeyStatus      `json:"status"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
	ExpiresAt             *time.Time        `json:"expires_at,omitempty"`
	LastUsedAt            *time.Time        `json:"last_used_at,omitempty"`
	UsageCount            int64             `json:"usage_count"`
	IPRestrictions        []string          `json:"ip_restrictions,omitempty"`
	UserAgentRestrictions []string          `json:"user_agent_restrictions,omitempty"`
	Metadata              map[string]string `json:"metadata,omitempty"`
}

// APIKeyStatus represents the status of an API key
type APIKeyStatus string

const (
	APIKeyStatusActive   APIKeyStatus = "active"
	APIKeyStatusInactive APIKeyStatus = "inactive"
	APIKeyStatusExpired  APIKeyStatus = "expired"
	APIKeyStatusRevoked  APIKeyStatus = "revoked"
)

// APIKeyPermission represents a permission for an API key
type APIKeyPermission struct {
	Resource   string   `json:"resource"`   // e.g., "collections", "vectors", "rag"
	Actions    []string `json:"actions"`    // e.g., ["read", "write", "delete"]
	Resources  []string `json:"resources"`  // Specific resource IDs (empty for all)
	Conditions []string `json:"conditions"` // Additional conditions
}

// APIKeyService defines the API key management service interface
type APIKeyService interface {
	// API Key Management
	CreateAPIKey(ctx context.Context, tenantID, name, description string, permissions []string, scopes []string, expiresAt *time.Time) (*APIKey, string, error)
	GetAPIKey(ctx context.Context, keyID string) (*APIKey, error)
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKey, error)
	UpdateAPIKey(ctx context.Context, key *APIKey) error
	DeleteAPIKey(ctx context.Context, keyID string) error
	ListAPIKeys(ctx context.Context, tenantID string, limit, offset int) ([]*APIKey, error)

	// API Key Validation
	ValidateAPIKey(ctx context.Context, key string, tenantID string) (*APIKey, error)
	CheckPermission(ctx context.Context, key *APIKey, resource, action, resourceID string) bool
	ValidateScope(ctx context.Context, key *APIKey, scope string) bool

	// API Key Lifecycle
	ActivateAPIKey(ctx context.Context, keyID string) error
	DeactivateAPIKey(ctx context.Context, keyID string) error
	RevokeAPIKey(ctx context.Context, keyID string, reason string) error
	RenewAPIKey(ctx context.Context, keyID string, newExpiry time.Time) error

	// Usage Tracking
	TrackAPIKeyUsage(ctx context.Context, keyID string, endpoint, method, ipAddress, userAgent string) error
	GetAPIKeyUsage(ctx context.Context, keyID string, startDate, endDate time.Time) (*APIKeyUsage, error)

	// Health Check
	HealthCheck(ctx context.Context) error
	Close() error
}

// APIKeyUsage represents usage statistics for an API key
type APIKeyUsage struct {
	KeyID       string    `json:"key_id"`
	TenantID    string    `json:"tenant_id"`
	Period      string    `json:"period"`
	GeneratedAt time.Time `json:"generated_at"`

	// Usage Summary
	TotalCalls       int64 `json:"total_calls"`
	UniqueIPs        int   `json:"unique_ips"`
	UniqueUserAgents int   `json:"unique_user_agents"`

	// Endpoint Usage
	EndpointUsage []EndpointUsage `json:"endpoint_usage"`

	// Performance Metrics
	AvgResponseTime float64 `json:"avg_response_time_ms"`
	SuccessRate     float64 `json:"success_rate"`
	ErrorRate       float64 `json:"error_rate"`

	// Security Metrics
	FailedAttempts     int                  `json:"failed_attempts"`
	SuspiciousActivity []SuspiciousActivity `json:"suspicious_activity"`
}

// EndpointUsage represents usage statistics for an endpoint
type EndpointUsage struct {
	Endpoint    string    `json:"endpoint"`
	Method      string    `json:"method"`
	CallCount   int64     `json:"call_count"`
	AvgDuration float64   `json:"avg_duration_ms"`
	SuccessRate float64   `json:"success_rate"`
	ErrorRate   float64   `json:"error_rate"`
	LastUsed    time.Time `json:"last_used"`
}

// SuspiciousActivity represents suspicious activity detected for an API key
type SuspiciousActivity struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"` // "unusual_ip", "unusual_user_agent", "rate_limit_exceeded"
	Description string    `json:"description"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	RiskScore   int       `json:"risk_score"`
	Action      string    `json:"action"` // "logged", "blocked", "alerted"
}

// DefaultAPIKeyService implements the API key management service
type DefaultAPIKeyService struct {
	// In a real implementation, this would use a database
	keys map[string]*APIKey
}

// NewDefaultAPIKeyService creates a new default API key service
func NewDefaultAPIKeyService() *DefaultAPIKeyService {
	return &DefaultAPIKeyService{
		keys: make(map[string]*APIKey),
	}
}

// CreateAPIKey creates a new API key for a tenant
func (s *DefaultAPIKeyService) CreateAPIKey(ctx context.Context, tenantID, name, description string, permissions []string, scopes []string, expiresAt *time.Time) (*APIKey, string, error) {
	// Generate a secure random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate random key: %w", err)
	}

	// Create the full key (in production, this would be stored securely)
	fullKey := fmt.Sprintf("vk_%s", hex.EncodeToString(keyBytes))

	// Hash the key for storage
	keyHash := s.hashKey(fullKey)

	// Create the API key record
	now := time.Now()
	apiKey := &APIKey{
		ID:          s.generateID(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		KeyHash:     keyHash,
		KeyPrefix:   fullKey[:8],
		Permissions: permissions,
		Scopes:      scopes,
		Status:      APIKeyStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   expiresAt,
		UsageCount:  0,
		Metadata:    make(map[string]string),
	}

	// Store the key
	s.keys[apiKey.ID] = apiKey

	return apiKey, fullKey, nil
}

// GetAPIKey retrieves an API key by ID
func (s *DefaultAPIKeyService) GetAPIKey(ctx context.Context, keyID string) (*APIKey, error) {
	key, exists := s.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("API key not found: %s", keyID)
	}
	return key, nil
}

// GetAPIKeyByHash retrieves an API key by its hash
func (s *DefaultAPIKeyService) GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKey, error) {
	for _, key := range s.keys {
		if key.KeyHash == keyHash {
			return key, nil
		}
	}
	return nil, fmt.Errorf("API key not found")
}

// UpdateAPIKey updates an existing API key
func (s *DefaultAPIKeyService) UpdateAPIKey(ctx context.Context, key *APIKey) error {
	if _, exists := s.keys[key.ID]; !exists {
		return fmt.Errorf("API key not found: %s", key.ID)
	}

	key.UpdatedAt = time.Now()
	s.keys[key.ID] = key
	return nil
}

// DeleteAPIKey deletes an API key
func (s *DefaultAPIKeyService) DeleteAPIKey(ctx context.Context, keyID string) error {
	if _, exists := s.keys[keyID]; !exists {
		return fmt.Errorf("API key not found: %s", keyID)
	}

	delete(s.keys, keyID)
	return nil
}

// ListAPIKeys lists API keys for a tenant
func (s *DefaultAPIKeyService) ListAPIKeys(ctx context.Context, tenantID string, limit, offset int) ([]*APIKey, error) {
	var tenantKeys []*APIKey

	for _, key := range s.keys {
		if key.TenantID == tenantID {
			tenantKeys = append(tenantKeys, key)
		}
	}

	// Simple pagination (in production, use database pagination)
	if offset >= len(tenantKeys) {
		return []*APIKey{}, nil
	}

	end := offset + limit
	if end > len(tenantKeys) {
		end = len(tenantKeys)
	}

	return tenantKeys[offset:end], nil
}

// ValidateAPIKey validates an API key and returns the key record
func (s *DefaultAPIKeyService) ValidateAPIKey(ctx context.Context, key string, tenantID string) (*APIKey, error) {
	// Hash the provided key
	keyHash := s.hashKey(key)

	// Find the key by hash
	apiKey, err := s.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	// Check if the key belongs to the specified tenant
	if apiKey.TenantID != tenantID {
		return nil, fmt.Errorf("API key does not belong to tenant")
	}

	// Check if the key is active
	if apiKey.Status != APIKeyStatusActive {
		return nil, fmt.Errorf("API key is not active")
	}

	// Check if the key has expired
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, fmt.Errorf("API key has expired")
	}

	return apiKey, nil
}

// CheckPermission checks if an API key has permission for a specific action
func (s *DefaultAPIKeyService) CheckPermission(ctx context.Context, key *APIKey, resource, action, resourceID string) bool {
	// Check if the key has the required permission
	requiredPermission := fmt.Sprintf("%s:%s", resource, action)

	for _, permission := range key.Permissions {
		if permission == requiredPermission || permission == "*" {
			return true
		}
	}

	return false
}

// ValidateScope checks if an API key has access to a specific scope
func (s *DefaultAPIKeyService) ValidateScope(ctx context.Context, key *APIKey, scope string) bool {
	// Check if the key has access to the scope
	for _, keyScope := range key.Scopes {
		if keyScope == scope || keyScope == "*" {
			return true
		}
	}

	return false
}

// ActivateAPIKey activates an API key
func (s *DefaultAPIKeyService) ActivateAPIKey(ctx context.Context, keyID string) error {
	key, err := s.GetAPIKey(ctx, keyID)
	if err != nil {
		return err
	}

	key.Status = APIKeyStatusActive
	key.UpdatedAt = time.Now()

	return s.UpdateAPIKey(ctx, key)
}

// DeactivateAPIKey deactivates an API key
func (s *DefaultAPIKeyService) DeactivateAPIKey(ctx context.Context, keyID string) error {
	key, err := s.GetAPIKey(ctx, keyID)
	if err != nil {
		return err
	}

	key.Status = APIKeyStatusInactive
	key.UpdatedAt = time.Now()

	return s.UpdateAPIKey(ctx, key)
}

// RevokeAPIKey revokes an API key
func (s *DefaultAPIKeyService) RevokeAPIKey(ctx context.Context, keyID string, reason string) error {
	key, err := s.GetAPIKey(ctx, keyID)
	if err != nil {
		return err
	}

	key.Status = APIKeyStatusRevoked
	key.UpdatedAt = time.Now()
	key.Metadata["revocation_reason"] = reason
	key.Metadata["revoked_at"] = time.Now().Format(time.RFC3339)

	return s.UpdateAPIKey(ctx, key)
}

// RenewAPIKey renews an API key with a new expiry date
func (s *DefaultAPIKeyService) RenewAPIKey(ctx context.Context, keyID string, newExpiry time.Time) error {
	key, err := s.GetAPIKey(ctx, keyID)
	if err != nil {
		return err
	}

	key.ExpiresAt = &newExpiry
	key.UpdatedAt = time.Now()

	return s.UpdateAPIKey(ctx, key)
}

// TrackAPIKeyUsage tracks API key usage for analytics
func (s *DefaultAPIKeyService) TrackAPIKeyUsage(ctx context.Context, keyID string, endpoint, method, ipAddress, userAgent string) error {
	key, err := s.GetAPIKey(ctx, keyID)
	if err != nil {
		return err
	}

	now := time.Now()
	key.LastUsedAt = &now
	key.UsageCount++
	key.UpdatedAt = now

	return s.UpdateAPIKey(ctx, key)
}

// GetAPIKeyUsage retrieves usage statistics for an API key
func (s *DefaultAPIKeyService) GetAPIKeyUsage(ctx context.Context, keyID string, startDate, endDate time.Time) (*APIKeyUsage, error) {
	// In a real implementation, this would query usage logs
	// For now, return a basic usage structure
	return &APIKeyUsage{
		KeyID:              keyID,
		GeneratedAt:        time.Now(),
		TotalCalls:         0,
		UniqueIPs:          0,
		UniqueUserAgents:   0,
		EndpointUsage:      []EndpointUsage{},
		AvgResponseTime:    0,
		SuccessRate:        0,
		ErrorRate:          0,
		FailedAttempts:     0,
		SuspiciousActivity: []SuspiciousActivity{},
	}, nil
}

// HealthCheck performs a health check on the service
func (s *DefaultAPIKeyService) HealthCheck(ctx context.Context) error {
	// Simple health check - verify we can access our data
	if s.keys == nil {
		return fmt.Errorf("API key service not initialized")
	}
	return nil
}

// Close performs cleanup operations
func (s *DefaultAPIKeyService) Close() error {
	// Clean up any resources if needed
	return nil
}

// Helper methods

// hashKey creates a SHA-256 hash of the API key
func (s *DefaultAPIKeyService) hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// generateID generates a unique ID for an API key
func (s *DefaultAPIKeyService) generateID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}
