package security

import (
	"context"
	"time"
)

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Severity    SecuritySeverity       `json:"severity"`
	Source      string                 `json:"source"`
	UserID      string                 `json:"user_id,omitempty"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details,omitempty"`
	RiskScore   float64                `json:"risk_score"`
	Status      SecurityEventStatus    `json:"status"`
	Action      string                 `json:"action,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// SecuritySeverity represents the severity level of a security event
type SecuritySeverity string

const (
	SecuritySeverityLow      SecuritySeverity = "low"
	SecuritySeverityMedium   SecuritySeverity = "medium"
	SecuritySeverityHigh     SecuritySeverity = "high"
	SecuritySeverityCritical SecuritySeverity = "critical"
)

// SecurityEventStatus represents the status of a security event
type SecurityEventStatus string

const (
	SecurityEventStatusOpen          SecurityEventStatus = "open"
	SecurityEventStatusInvestigating SecurityEventStatus = "investigating"
	SecurityEventStatusResolved      SecurityEventStatus = "resolved"
	SecurityEventStatusFalsePositive SecurityEventStatus = "false_positive"
)

// Threat represents a detected threat
type Threat struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	ThreatType  string                 `json:"threat_type"`
	Severity    SecuritySeverity       `json:"severity"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target"`
	Description string                 `json:"description"`
	Indicators  []ThreatIndicator      `json:"indicators"`
	RiskScore   float64                `json:"risk_score"`
	Confidence  float64                `json:"confidence"`
	Status      ThreatStatus           `json:"status"`
	Mitigation  []string               `json:"mitigation,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ThreatStatus represents the status of a threat
type ThreatStatus string

const (
	ThreatStatusActive        ThreatStatus = "active"
	ThreatStatusMitigated     ThreatStatus = "mitigated"
	ThreatStatusResolved      ThreatStatus = "resolved"
	ThreatStatusFalsePositive ThreatStatus = "false_positive"
)

// ThreatIndicator represents an indicator of compromise
type ThreatIndicator struct {
	Type       string                 `json:"type"`
	Value      string                 `json:"value"`
	Confidence float64                `json:"confidence"`
	Source     string                 `json:"source"`
	FirstSeen  time.Time              `json:"first_seen"`
	LastSeen   time.Time              `json:"last_seen"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceRequirement represents a compliance requirement
type ComplianceRequirement struct {
	ID          string                `json:"id"`
	Framework   string                `json:"framework"` // GDPR, SOC2, HIPAA, PCI DSS
	Category    string                `json:"category"`
	Requirement string                `json:"requirement"`
	Description string                `json:"description"`
	Controls    []*ComplianceControl  `json:"controls"`
	Status      ComplianceStatus      `json:"status"`
	LastAudit   time.Time             `json:"last_audit"`
	NextAudit   time.Time             `json:"next_audit"`
	Evidence    []*ComplianceEvidence `json:"evidence,omitempty"`
}

// ComplianceStatus represents the status of compliance
type ComplianceStatus string

const (
	ComplianceStatusCompliant     ComplianceStatus = "compliant"
	ComplianceStatusNonCompliant  ComplianceStatus = "non_compliant"
	ComplianceStatusInProgress    ComplianceStatus = "in_progress"
	ComplianceStatusNotApplicable ComplianceStatus = "not_applicable"
)

// ComplianceControl represents a compliance control
type ComplianceControl struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Type           string           `json:"type"` // preventive, detective, corrective
	Status         ComplianceStatus `json:"status"`
	Implementation string           `json:"implementation"`
	Testing        string           `json:"testing"`
	Owner          string           `json:"owner"`
	LastUpdated    time.Time        `json:"last_updated"`
}

// ComplianceEvidence represents evidence of compliance
type ComplianceEvidence struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // log, report, test, audit
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Attachments []string               `json:"attachments,omitempty"`
}

// EncryptionPolicy represents an encryption policy
type EncryptionPolicy struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Algorithm   string                 `json:"algorithm"` // AES-256, ChaCha20-Poly1305
	KeySize     int                    `json:"key_size"`
	Mode        string                 `json:"mode"` // GCM, CBC, CTR
	KeyRotation int                    `json:"key_rotation_days"`
	DataTypes   []string               `json:"data_types"` // PII, PHI, financial, general
	Status      EncryptionPolicyStatus `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// EncryptionPolicyStatus represents the status of an encryption policy
type EncryptionPolicyStatus string

const (
	EncryptionPolicyStatusActive   EncryptionPolicyStatus = "active"
	EncryptionPolicyStatusInactive EncryptionPolicyStatus = "inactive"
	EncryptionPolicyStatusDraft    EncryptionPolicyStatus = "draft"
)

// EncryptionKey represents an encryption key
type EncryptionKey struct {
	ID         string              `json:"id"`
	TenantID   string              `json:"tenant_id"`
	KeyID      string              `json:"key_id"`
	Algorithm  string              `json:"algorithm"`
	KeySize    int                 `json:"key_size"`
	Status     EncryptionKeyStatus `json:"status"`
	CreatedAt  time.Time           `json:"created_at"`
	ExpiresAt  time.Time           `json:"expires_at"`
	LastUsed   time.Time           `json:"last_used"`
	UsageCount int64               `json:"usage_count"`
	Metadata   map[string]string   `json:"metadata,omitempty"`
}

// EncryptionKeyStatus represents the status of an encryption key
type EncryptionKeyStatus string

const (
	EncryptionKeyStatusActive      EncryptionKeyStatus = "active"
	EncryptionKeyStatusInactive    EncryptionKeyStatus = "inactive"
	EncryptionKeyStatusExpired     EncryptionKeyStatus = "expired"
	EncryptionKeyStatusCompromised EncryptionKeyStatus = "compromised"
)

// ThreatDetectionService defines the threat detection service interface
type ThreatDetectionService interface {
	// Threat Detection
	DetectThreat(ctx context.Context, event *SecurityEvent) (*Threat, error)
	AnalyzeBehavior(ctx context.Context, userID, tenantID string) (*Threat, error)
	CalculateRiskScore(ctx context.Context, event *SecurityEvent) (float64, error)

	// Threat Management
	GetThreat(ctx context.Context, threatID string) (*Threat, error)
	UpdateThreat(ctx context.Context, threat *Threat) error
	ListThreats(ctx context.Context, filter *ThreatFilter) ([]*Threat, error)

	// ML Models
	TrainModel(ctx context.Context, trainingData []*SecurityEvent) error
	UpdateModel(ctx context.Context, modelID string) error
	GetModelPerformance(ctx context.Context, modelID string) (*ModelPerformance, error)

	// Health Check
	HealthCheck(ctx context.Context) error
	Close() error
}

// ComplianceService defines the compliance service interface
type ComplianceService interface {
	// Compliance Management
	CheckCompliance(ctx context.Context, tenantID string, framework string) (*ComplianceReport, error)
	UpdateCompliance(ctx context.Context, requirement *ComplianceRequirement) error
	GetComplianceStatus(ctx context.Context, tenantID string) (*ComplianceStatus, error)

	// Compliance Reporting
	GenerateReport(ctx context.Context, tenantID string, framework string, startDate, endDate time.Time) (*ComplianceReport, error)
	GetComplianceEvidence(ctx context.Context, requirementID string) ([]ComplianceEvidence, error)

	// Framework Management
	GetFrameworks(ctx context.Context) ([]string, error)
	GetRequirements(ctx context.Context, framework string) ([]*ComplianceRequirement, error)

	// Health Check
	HealthCheck(ctx context.Context) error
	Close() error
}

// EncryptionService defines the encryption service interface
type EncryptionService interface {
	// Encryption Operations
	Encrypt(ctx context.Context, data []byte, policyID string) ([]byte, error)
	Decrypt(ctx context.Context, encryptedData []byte, keyID string) ([]byte, error)

	// Key Management
	GenerateKey(ctx context.Context, policyID string) (*EncryptionKey, error)
	RotateKey(ctx context.Context, keyID string) error
	GetKey(ctx context.Context, keyID string) (*EncryptionKey, error)

	// Policy Management
	CreatePolicy(ctx context.Context, policy *EncryptionPolicy) error
	UpdatePolicy(ctx context.Context, policy *EncryptionPolicy) error
	GetPolicy(ctx context.Context, policyID string) (*EncryptionPolicy, error)

	// Health Check
	HealthCheck(ctx context.Context) error
	Close() error
}

// SecurityAnalyticsService defines the security analytics service interface
type SecurityAnalyticsService interface {
	// Security Analytics
	AnalyzeSecurityEvents(ctx context.Context, filter *SecurityEventFilter) (*SecurityAnalytics, error)
	GetThreatIntelligence(ctx context.Context, threatType string) (*ThreatIntelligence, error)
	GenerateSecurityReport(ctx context.Context, tenantID string, startDate, endDate time.Time) (*SecurityReport, error)

	// Performance Metrics
	GetSecurityMetrics(ctx context.Context, tenantID string) (*SecurityMetrics, error)
	GetThreatMetrics(ctx context.Context, tenantID string) (*ThreatMetrics, error)

	// Health Check
	HealthCheck(ctx context.Context) error
	Close() error
}

// ThreatFilter represents filters for threat queries
type ThreatFilter struct {
	ThreatType string           `json:"threat_type,omitempty"`
	Severity   SecuritySeverity `json:"severity,omitempty"`
	Status     ThreatStatus     `json:"status,omitempty"`
	TenantID   string           `json:"tenant_id,omitempty"`
	StartDate  time.Time        `json:"start_date,omitempty"`
	EndDate    time.Time        `json:"end_date,omitempty"`
	RiskScore  float64          `json:"risk_score,omitempty"`
	Limit      int              `json:"limit,omitempty"`
	Offset     int              `json:"offset,omitempty"`
}

// SecurityEventFilter represents filters for security event queries
type SecurityEventFilter struct {
	EventType string           `json:"event_type,omitempty"`
	Severity  SecuritySeverity `json:"severity,omitempty"`
	Source    string           `json:"source,omitempty"`
	TenantID  string           `json:"tenant_id,omitempty"`
	UserID    string           `json:"user_id,omitempty"`
	StartDate time.Time        `json:"start_date,omitempty"`
	EndDate   time.Time        `json:"end_date,omitempty"`
	Limit     int              `json:"limit,omitempty"`
	Offset    int              `json:"offset,omitempty"`
}

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	TenantID    string    `json:"tenant_id"`
	Framework   string    `json:"framework"`
	GeneratedAt time.Time `json:"generated_at"`
	Period      string    `json:"period"`

	// Compliance Summary
	OverallStatus     ComplianceStatus `json:"overall_status"`
	CompliantCount    int              `json:"compliant_count"`
	NonCompliantCount int              `json:"non_compliant_count"`
	InProgressCount   int              `json:"in_progress_count"`

	// Requirements
	Requirements []*ComplianceRequirement `json:"requirements"`

	// Recommendations
	Recommendations []string `json:"recommendations"`

	// Evidence
	Evidence []*ComplianceEvidence `json:"evidence"`
}

// SecurityAnalytics represents security analytics data
type SecurityAnalytics struct {
	TenantID    string    `json:"tenant_id"`
	GeneratedAt time.Time `json:"generated_at"`
	Period      string    `json:"period"`

	// Event Analysis
	TotalEvents          int64            `json:"total_events"`
	EventTypes           map[string]int64 `json:"event_types"`
	SeverityDistribution map[string]int64 `json:"severity_distribution"`

	// Threat Analysis
	ThreatsDetected  int64            `json:"threats_detected"`
	ThreatTypes      map[string]int64 `json:"threat_types"`
	RiskDistribution map[string]int64 `json:"risk_distribution"`

	// Performance Metrics
	DetectionRate     float64 `json:"detection_rate"`
	FalsePositiveRate float64 `json:"false_positive_rate"`
	ResponseTime      float64 `json:"response_time_ms"`
}

// ThreatIntelligence represents threat intelligence data
type ThreatIntelligence struct {
	ThreatType string    `json:"threat_type"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Threat Information
	Description string            `json:"description"`
	Indicators  []ThreatIndicator `json:"indicators"`
	Mitigation  []string          `json:"mitigation"`

	// Intelligence Sources
	Sources    []string  `json:"sources"`
	Confidence float64   `json:"confidence"`
	LastSeen   time.Time `json:"last_seen"`
}

// SecurityReport represents a comprehensive security report
type SecurityReport struct {
	TenantID    string    `json:"tenant_id"`
	GeneratedAt time.Time `json:"generated_at"`
	Period      string    `json:"period"`

	// Security Overview
	SecurityScore float64 `json:"security_score"`
	ThreatLevel   string  `json:"threat_level"`
	RiskLevel     string  `json:"risk_level"`

	// Incident Summary
	Incidents []*SecurityEvent `json:"incidents"`
	Threats   []*Threat        `json:"threats"`

	// Recommendations
	Recommendations []string `json:"recommendations"`

	// Compliance Status
	Compliance map[string]ComplianceStatus `json:"compliance"`
}

// SecurityMetrics represents security performance metrics
type SecurityMetrics struct {
	TenantID    string    `json:"tenant_id"`
	GeneratedAt time.Time `json:"generated_at"`

	// Detection Metrics
	DetectionRate     float64 `json:"detection_rate"`
	FalsePositiveRate float64 `json:"false_positive_rate"`
	TruePositiveRate  float64 `json:"true_positive_rate"`

	// Response Metrics
	MeanTimeToDetect  float64 `json:"mean_time_to_detect_ms"`
	MeanTimeToRespond float64 `json:"mean_time_to_respond_ms"`
	MeanTimeToResolve float64 `json:"mean_time_to_resolve_ms"`

	// Threat Metrics
	ThreatsPerDay   float64 `json:"threats_per_day"`
	IncidentsPerDay float64 `json:"incidents_per_day"`
	RiskScore       float64 `json:"average_risk_score"`
}

// ThreatMetrics represents threat-specific metrics
type ThreatMetrics struct {
	TenantID    string    `json:"tenant_id"`
	GeneratedAt time.Time `json:"generated_at"`

	// Threat Distribution
	ThreatTypes          map[string]int64 `json:"threat_types"`
	SeverityDistribution map[string]int64 `json:"severity_distribution"`
	StatusDistribution   map[string]int64 `json:"status_distribution"`

	// Threat Trends
	ThreatTrend []ThreatTrendPoint `json:"threat_trend"`
	RiskTrend   []RiskTrendPoint   `json:"risk_trend"`
}

// ThreatTrendPoint represents a point in threat trend data
type ThreatTrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int64     `json:"count"`
	Type      string    `json:"type"`
}

// RiskTrendPoint represents a point in risk trend data
type RiskTrendPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	RiskScore  float64   `json:"risk_score"`
	ThreatType string    `json:"threat_type"`
}

// ModelPerformance represents ML model performance metrics
type ModelPerformance struct {
	ModelID   string    `json:"model_id"`
	UpdatedAt time.Time `json:"updated_at"`

	// Performance Metrics
	Accuracy  float64 `json:"accuracy"`
	Precision float64 `json:"precision"`
	Recall    float64 `json:"recall"`
	F1Score   float64 `json:"f1_score"`

	// Training Metrics
	TrainingDataSize int64     `json:"training_data_size"`
	LastTraining     time.Time `json:"last_training"`
	TrainingDuration float64   `json:"training_duration_seconds"`

	// Model Information
	Algorithm string `json:"algorithm"`
	Version   string `json:"version"`
	Status    string `json:"status"`
}
