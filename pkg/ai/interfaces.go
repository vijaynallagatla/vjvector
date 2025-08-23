package ai

import (
	"context"
	"time"
)

// AIModel represents an AI model in the system
type AIModel struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Type         ModelType              `json:"type"`
	Status       ModelStatus            `json:"status"`
	Framework    string                 `json:"framework"` // TensorFlow, PyTorch, ONNX, etc.
	Architecture string                 `json:"architecture"`
	InputFormat  string                 `json:"input_format"`
	OutputFormat string                 `json:"output_format"`
	ModelSize    int64                  `json:"model_size_bytes"`
	Parameters   int64                  `json:"parameters_count"`
	Accuracy     float64                `json:"accuracy"`
	Latency      float64                `json:"latency_ms"`
	Throughput   float64                `json:"throughput_rps"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	DeployedAt   *time.Time             `json:"deployed_at,omitempty"`
	Performance  *ModelPerformance      `json:"performance,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
}

// ModelType represents the type of AI model
type ModelType string

const (
	ModelTypeEmbedding      ModelType = "embedding"
	ModelTypeGeneration     ModelType = "generation"
	ModelTypeClassification ModelType = "classification"
	ModelTypeReranking      ModelType = "reranking"
	ModelTypeTranslation    ModelType = "translation"
	ModelTypeSummarization  ModelType = "summarization"
)

// ModelStatus represents the status of an AI model
type ModelStatus string

const (
	ModelStatusDraft      ModelStatus = "draft"
	ModelStatusTraining   ModelStatus = "training"
	ModelStatusTesting    ModelStatus = "testing"
	ModelStatusDeployed   ModelStatus = "deployed"
	ModelStatusDeprecated ModelStatus = "deprecated"
	ModelStatusFailed     ModelStatus = "failed"
)

// ModelPerformance represents performance metrics for an AI model
type ModelPerformance struct {
	ModelID   string    `json:"model_id"`
	Timestamp time.Time `json:"timestamp"`

	// Accuracy Metrics
	Accuracy  float64 `json:"accuracy"`
	Precision float64 `json:"precision"`
	Recall    float64 `json:"recall"`
	F1Score   float64 `json:"f1_score"`

	// Performance Metrics
	Latency        float64 `json:"latency_ms"`
	Throughput     float64 `json:"throughput_rps"`
	MemoryUsage    float64 `json:"memory_usage_mb"`
	GPUUtilization float64 `json:"gpu_utilization_percent"`

	// Usage Metrics
	RequestCount int64   `json:"request_count"`
	ErrorCount   int64   `json:"error_count"`
	SuccessRate  float64 `json:"success_rate"`

	// Custom Metrics
	CustomMetrics map[string]float64 `json:"custom_metrics,omitempty"`
}

// ModelRegistry represents the AI model registry
type ModelRegistry struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Models      []*AIModel             `json:"models"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ModelDeployment represents a model deployment
type ModelDeployment struct {
	ID          string                 `json:"id"`
	ModelID     string                 `json:"model_id"`
	Environment string                 `json:"environment"` // dev, staging, prod
	Status      DeploymentStatus       `json:"status"`
	Replicas    int                    `json:"replicas"`
	Resources   *DeploymentResources   `json:"resources"`
	Config      map[string]interface{} `json:"config,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Health      *DeploymentHealth      `json:"health,omitempty"`
}

// DeploymentStatus represents the status of a model deployment
type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusDeploying DeploymentStatus = "deploying"
	DeploymentStatusRunning   DeploymentStatus = "running"
	DeploymentStatusScaling   DeploymentStatus = "scaling"
	DeploymentStatusFailed    DeploymentStatus = "failed"
	DeploymentStatusStopped   DeploymentStatus = "stopped"
)

// DeploymentResources represents resource requirements for a deployment
type DeploymentResources struct {
	CPU     float64 `json:"cpu_cores"`
	Memory  float64 `json:"memory_gb"`
	GPU     int     `json:"gpu_count"`
	GPUType string  `json:"gpu_type,omitempty"`
	Storage float64 `json:"storage_gb"`
	Network string  `json:"network_bandwidth,omitempty"`
}

// DeploymentHealth represents the health status of a deployment
type DeploymentHealth struct {
	Status        string                 `json:"status"`
	ReadyReplicas int                    `json:"ready_replicas"`
	TotalReplicas int                    `json:"total_replicas"`
	LastCheck     time.Time              `json:"last_check"`
	Issues        []HealthIssue          `json:"issues,omitempty"`
	Metrics       map[string]interface{} `json:"metrics,omitempty"`
}

// HealthIssue represents a health issue with a deployment
type HealthIssue struct {
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Resolved  bool      `json:"resolved"`
}

// AdvancedRAGRequest represents an advanced RAG request
type AdvancedRAGRequest struct {
	ID        string                 `json:"id"`
	Query     string                 `json:"query"`
	Context   string                 `json:"context,omitempty"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	Options   *RAGOptions            `json:"options,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	TenantID  string                 `json:"tenant_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Priority  int                    `json:"priority"`
}

// RAGOptions represents options for RAG processing
type RAGOptions struct {
	MaxResults          int                    `json:"max_results"`
	SimilarityThreshold float64                `json:"similarity_threshold"`
	RerankingEnabled    bool                   `json:"reranking_enabled"`
	QueryExpansion      bool                   `json:"query_expansion"`
	ContextWindow       int                    `json:"context_window"`
	Language            string                 `json:"language,omitempty"`
	CustomOptions       map[string]interface{} `json:"custom_options,omitempty"`
}

// AdvancedRAGResponse represents an advanced RAG response
type AdvancedRAGResponse struct {
	ID             string       `json:"id"`
	RequestID      string       `json:"request_id"`
	Results        []*RAGResult `json:"results"`
	Metadata       *RAGMetadata `json:"metadata,omitempty"`
	ProcessingTime float64      `json:"processing_time_ms"`
	ModelUsed      string       `json:"model_used"`
	Timestamp      time.Time    `json:"timestamp"`
}

// RAGResult represents a single RAG result
type RAGResult struct {
	ID            string                 `json:"id"`
	Content       string                 `json:"content"`
	Source        string                 `json:"source"`
	Similarity    float64                `json:"similarity_score"`
	Rank          int                    `json:"rank"`
	Confidence    float64                `json:"confidence"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Embedding     []float64              `json:"embedding,omitempty"`
	RerankedScore float64                `json:"reranked_score,omitempty"`
}

// RAGMetadata represents metadata about RAG processing
type RAGMetadata struct {
	TotalResults    int                `json:"total_results"`
	RetrievedCount  int                `json:"retrieved_count"`
	RerankedCount   int                `json:"reranked_count"`
	QueryExpanded   bool               `json:"query_expanded"`
	ExpandedQuery   string             `json:"expanded_query,omitempty"`
	ProcessingSteps []string           `json:"processing_steps"`
	ModelVersions   map[string]string  `json:"model_versions"`
	Performance     map[string]float64 `json:"performance,omitempty"`
}

// QueryExpansion represents query expansion results
type QueryExpansion struct {
	OriginalQuery   string                 `json:"original_query"`
	ExpandedQueries []string               `json:"expanded_queries"`
	Confidence      []float64              `json:"confidence"`
	Method          string                 `json:"method"`
	Context         string                 `json:"context,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// RerankingRequest represents a reranking request
type RerankingRequest struct {
	ID         string                 `json:"id"`
	Query      string                 `json:"query"`
	Candidates []*RAGResult           `json:"candidates"`
	ModelID    string                 `json:"model_id"`
	Options    map[string]interface{} `json:"options,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// RerankingResponse represents a reranking response
type RerankingResponse struct {
	ID              string       `json:"id"`
	RequestID       string       `json:"request_id"`
	RerankedResults []*RAGResult `json:"reranked_results"`
	ModelUsed       string       `json:"model_used"`
	ProcessingTime  float64      `json:"processing_time_ms"`
	Timestamp       time.Time    `json:"timestamp"`
}

// AITrafficSplit represents traffic splitting for A/B testing
type AITrafficSplit struct {
	ID         string               `json:"id"`
	Experiment string               `json:"experiment"`
	ModelA     string               `json:"model_a"`
	ModelB     string               `json:"model_b"`
	SplitRatio float64              `json:"split_ratio"` // 0.0 to 1.0
	Status     TrafficSplitStatus   `json:"status"`
	StartDate  time.Time            `json:"start_date"`
	EndDate    *time.Time           `json:"end_date,omitempty"`
	Metrics    *TrafficSplitMetrics `json:"metrics,omitempty"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
}

// TrafficSplitStatus represents the status of traffic splitting
type TrafficSplitStatus string

const (
	TrafficSplitStatusActive    TrafficSplitStatus = "active"
	TrafficSplitStatusPaused    TrafficSplitStatus = "paused"
	TrafficSplitStatusCompleted TrafficSplitStatus = "completed"
	TrafficSplitStatusFailed    TrafficSplitStatus = "failed"
)

// TrafficSplitMetrics represents metrics for traffic splitting
type TrafficSplitMetrics struct {
	ModelARequests          int64              `json:"model_a_requests"`
	ModelBRequests          int64              `json:"model_b_requests"`
	ModelAPerformance       map[string]float64 `json:"model_a_performance"`
	ModelBPerformance       map[string]float64 `json:"model_b_performance"`
	StatisticalSignificance float64            `json:"statistical_significance"`
	Winner                  string             `json:"winner,omitempty"`
	Confidence              float64            `json:"confidence"`
	LastUpdated             time.Time          `json:"last_updated"`
}

// AIModelService defines the AI model management service interface
type AIModelService interface {
	// Model Management
	CreateModel(ctx context.Context, model *AIModel) error
	GetModel(ctx context.Context, modelID string) (*AIModel, error)
	UpdateModel(ctx context.Context, model *AIModel) error
	DeleteModel(ctx context.Context, modelID string) error
	ListModels(ctx context.Context, filter *ModelFilter) ([]*AIModel, error)

	// Model Deployment
	DeployModel(ctx context.Context, modelID string, config *DeploymentConfig) (*ModelDeployment, error)
	UndeployModel(ctx context.Context, deploymentID string) error
	GetDeployment(ctx context.Context, deploymentID string) (*ModelDeployment, error)
	ListDeployments(ctx context.Context, filter *DeploymentFilter) ([]*ModelDeployment, error)

	// Model Performance
	GetModelPerformance(ctx context.Context, modelID string) (*ModelPerformance, error)
	UpdateModelPerformance(ctx context.Context, performance *ModelPerformance) error

	// Health Check
	HealthCheck(ctx context.Context) error
	Close() error
}

// AdvancedRAGService defines the advanced RAG service interface
type AdvancedRAGService interface {
	// RAG Processing
	ProcessRAG(ctx context.Context, request *AdvancedRAGRequest) (*AdvancedRAGResponse, error)
	ExpandQuery(ctx context.Context, query string, context string) (*QueryExpansion, error)
	RerankResults(ctx context.Context, request *RerankingRequest) (*RerankingResponse, error)

	// RAG Configuration
	UpdateRAGConfig(ctx context.Context, config *RAGConfig) error
	GetRAGConfig(ctx context.Context) (*RAGConfig, error)

	// Health Check
	HealthCheck(ctx context.Context) error
	Close() error
}

// AIOrchestratorService defines the AI orchestration service interface
type AIOrchestratorService interface {
	// Model Orchestration
	RouteRequest(ctx context.Context, request *AIRequest) (*AIModel, error)
	LoadBalance(ctx context.Context, request *AIRequest) (*AIModel, error)
	AutoScale(ctx context.Context, modelID string) error

	// Traffic Management
	CreateTrafficSplit(ctx context.Context, split *AITrafficSplit) error
	UpdateTrafficSplit(ctx context.Context, split *AITrafficSplit) error
	GetTrafficSplit(ctx context.Context, splitID string) (*AITrafficSplit, error)

	// Performance Monitoring
	GetSystemMetrics(ctx context.Context) (*SystemMetrics, error)
	GetModelMetrics(ctx context.Context, modelID string) (*ModelMetrics, error)

	// Health Check
	HealthCheck(ctx context.Context) error
	Close() error
}

// AIRequest represents a generic AI request
type AIRequest struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      interface{}            `json:"data"`
	Priority  int                    `json:"priority"`
	UserID    string                 `json:"user_id,omitempty"`
	TenantID  string                 `json:"tenant_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// DeploymentConfig represents configuration for model deployment
type DeploymentConfig struct {
	Environment string                 `json:"environment"`
	Replicas    int                    `json:"replicas"`
	Resources   *DeploymentResources   `json:"resources"`
	Config      map[string]interface{} `json:"config,omitempty"`
	HealthCheck *HealthCheckConfig     `json:"health_check,omitempty"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Endpoint         string        `json:"endpoint"`
	Interval         time.Duration `json:"interval"`
	Timeout          time.Duration `json:"timeout"`
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
}

// ModelFilter represents filters for model queries
type ModelFilter struct {
	Type          ModelType   `json:"type,omitempty"`
	Status        ModelStatus `json:"status,omitempty"`
	Framework     string      `json:"framework,omitempty"`
	Tags          []string    `json:"tags,omitempty"`
	CreatedAfter  *time.Time  `json:"created_after,omitempty"`
	CreatedBefore *time.Time  `json:"created_before,omitempty"`
	Limit         int         `json:"limit,omitempty"`
	Offset        int         `json:"offset,omitempty"`
}

// DeploymentFilter represents filters for deployment queries
type DeploymentFilter struct {
	ModelID       string           `json:"model_id,omitempty"`
	Environment   string           `json:"environment,omitempty"`
	Status        DeploymentStatus `json:"status,omitempty"`
	CreatedAfter  *time.Time       `json:"created_after,omitempty"`
	CreatedBefore *time.Time       `json:"created_before,omitempty"`
	Limit         int              `json:"limit,omitempty"`
	Offset        int              `json:"offset,omitempty"`
}

// RAGConfig represents configuration for RAG processing
type RAGConfig struct {
	DefaultMaxResults          int                    `json:"default_max_results"`
	DefaultSimilarityThreshold float64                `json:"default_similarity_threshold"`
	DefaultRerankingEnabled    bool                   `json:"default_reranking_enabled"`
	DefaultQueryExpansion      bool                   `json:"default_query_expansion"`
	DefaultContextWindow       int                    `json:"default_context_window"`
	SupportedLanguages         []string               `json:"supported_languages"`
	ModelMappings              map[string]string      `json:"model_mappings"`
	CustomConfig               map[string]interface{} `json:"custom_config,omitempty"`
}

// SystemMetrics represents system-wide AI metrics
type SystemMetrics struct {
	Timestamp time.Time `json:"timestamp"`

	// System Overview
	TotalModels        int `json:"total_models"`
	ActiveModels       int `json:"active_models"`
	TotalDeployments   int `json:"total_deployments"`
	RunningDeployments int `json:"running_deployments"`

	// Performance Metrics
	TotalRequests   int64   `json:"total_requests"`
	ActiveRequests  int     `json:"active_requests"`
	AvgResponseTime float64 `json:"avg_response_time_ms"`
	ErrorRate       float64 `json:"error_rate"`

	// Resource Metrics
	TotalCPUUsage    float64 `json:"total_cpu_usage_percent"`
	TotalMemoryUsage float64 `json:"total_memory_usage_percent"`
	TotalGPUUsage    float64 `json:"total_gpu_usage_percent"`

	// Traffic Metrics
	TrafficSplits []*TrafficSplitMetrics `json:"traffic_splits"`
}

// ModelMetrics represents metrics for a specific model
type ModelMetrics struct {
	ModelID   string    `json:"model_id"`
	Timestamp time.Time `json:"timestamp"`

	// Request Metrics
	RequestCount    int64   `json:"request_count"`
	ActiveRequests  int     `json:"active_requests"`
	AvgResponseTime float64 `json:"avg_response_time_ms"`
	ErrorCount      int64   `json:"error_count"`
	SuccessRate     float64 `json:"success_rate"`

	// Performance Metrics
	Throughput     float64 `json:"throughput_rps"`
	Latency        float64 `json:"latency_ms"`
	MemoryUsage    float64 `json:"memory_usage_mb"`
	GPUUtilization float64 `json:"gpu_utilization_percent"`

	// Quality Metrics
	Accuracy  float64 `json:"accuracy"`
	Precision float64 `json:"precision"`
	Recall    float64 `json:"recall"`
	F1Score   float64 `json:"f1_score"`
}
