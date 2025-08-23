package node

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/cluster"
	"github.com/vijaynallagatla/vjvector/pkg/embedding"
	"github.com/vijaynallagatla/vjvector/pkg/index"
	"github.com/vijaynallagatla/vjvector/pkg/rag"
	"github.com/vijaynallagatla/vjvector/pkg/storage"
)

// VJVectorNode represents a production VJVector node
type VJVectorNode struct {
	mu sync.RWMutex

	// Node identity
	nodeID    string
	version   string
	startTime time.Time

	// Core services
	embeddingService embedding.Service
	ragEngine        rag.Engine
	vectorIndex      index.VectorIndex
	storage          storage.StorageEngine

	// Clustering
	cluster cluster.Cluster
	// clusterConfig *cluster.Config

	// Node state
	state  cluster.NodeState
	role   cluster.NodeRole
	health *NodeHealth

	// Configuration
	config *NodeConfig

	// Monitoring and metrics
	metrics       *NodeMetrics
	healthChecker *HealthChecker

	// Context management
	ctx    context.Context
	cancel context.CancelFunc

	// Logging
	logger *slog.Logger
}

// NodeConfig holds node configuration
type NodeConfig struct {
	// Node configuration
	NodeID  string           `json:"node_id"`
	Address string           `json:"address"`
	Port    int              `json:"port"`
	Role    cluster.NodeRole `json:"role"`

	// Service configuration
	EmbeddingConfig *embedding.Config      `json:"embedding_config"`
	RAGConfig       *rag.Config            `json:"rag_config"`
	IndexConfig     *index.IndexConfig     `json:"index_config"`
	StorageConfig   *storage.StorageConfig `json:"storage_config"`

	// Clustering configuration
	ClusterConfig *cluster.Config `json:"cluster_config"`

	// Performance configuration
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`
	RequestTimeout        time.Duration `json:"request_timeout"`
	HealthCheckInterval   time.Duration `json:"health_check_interval"`

	// Security configuration
	EnableAuth     bool     `json:"enable_auth"`
	JWTSecret      string   `json:"jwt_secret"`
	APIKeyRequired bool     `json:"api_key_required"`
	AllowedOrigins []string `json:"allowed_origins"`
}

// NodeHealth represents the health status of the node
type NodeHealth struct {
	Status    string            `json:"status"`
	LastCheck time.Time         `json:"last_check"`
	Uptime    time.Duration     `json:"uptime"`
	Services  map[string]string `json:"services"`
	Resources *ResourceUsage    `json:"resources"`
	Errors    []string          `json:"errors,omitempty"`
	Warnings  []string          `json:"warnings,omitempty"`
}

// ResourceUsage represents resource usage information
type ResourceUsage struct {
	CPUUsage    float64    `json:"cpu_usage"`
	MemoryUsage float64    `json:"memory_usage"`
	DiskUsage   float64    `json:"disk_usage"`
	NetworkIO   *NetworkIO `json:"network_io"`
}

// NetworkIO represents network I/O statistics
type NetworkIO struct {
	BytesReceived   int64 `json:"bytes_received"`
	BytesSent       int64 `json:"bytes_sent"`
	PacketsReceived int64 `json:"packets_received"`
	PacketsSent     int64 `json:"packets_sent"`
	Errors          int64 `json:"errors"`
}

// NodeMetrics represents node performance metrics
type NodeMetrics struct {
	// Request metrics
	TotalRequests      int64         `json:"total_requests"`
	ActiveRequests     int64         `json:"active_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`

	// Service metrics
	EmbeddingRequests int64 `json:"embedding_requests"`
	RAGRequests       int64 `json:"rag_requests"`
	VectorOperations  int64 `json:"vector_operations"`

	// Resource metrics
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`

	// Timestamps
	LastUpdated time.Time `json:"last_updated"`
	StartTime   time.Time `json:"start_time"`
}

// HealthChecker checks the health of node components
type HealthChecker struct {
	node     *VJVectorNode
	interval time.Duration
	ticker   *time.Ticker
	logger   *slog.Logger
}

// NewVJVectorNode creates a new VJVector node
func NewVJVectorNode(config *NodeConfig, logger *slog.Logger) (*VJVectorNode, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if logger == nil {
		logger = slog.Default()
	}

	// Set defaults
	if config.MaxConcurrentRequests == 0 {
		config.MaxConcurrentRequests = 1000
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 30 * time.Second
	}
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 30 * time.Second
	}

	node := &VJVectorNode{
		nodeID:    config.NodeID,
		version:   "1.0.0", // TODO: Get from build info
		startTime: time.Now(),
		state:     cluster.NodeStateStarting,
		role:      config.Role,
		config:    config,
		logger:    logger,
		health: &NodeHealth{
			Status:    "starting",
			LastCheck: time.Now(),
			Services:  make(map[string]string),
			Resources: &ResourceUsage{},
		},
		metrics: &NodeMetrics{
			StartTime:   time.Now(),
			LastUpdated: time.Now(),
		},
	}

	// Initialize health checker
	node.healthChecker = NewHealthChecker(node, config.HealthCheckInterval, logger)

	return node, nil
}

// Start starts the VJVector node
func (n *VJVectorNode) Start(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.state != cluster.NodeStateStarting {
		return fmt.Errorf("node is not in starting state: %s", n.state)
	}

	n.logger.Info("Starting VJVector node",
		"node_id", n.nodeID,
		"address", n.config.Address,
		"port", n.config.Port,
		"role", n.config.Role)

	// Initialize core services
	if err := n.initializeServices(ctx); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	// Initialize clustering if configured
	if n.config.ClusterConfig != nil {
		if err := n.initializeClustering(ctx); err != nil {
			n.logger.Warn("Failed to initialize clustering, continuing without clustering", "error", err)
		}
	}

	// Create context for background operations
	n.ctx, n.cancel = context.WithCancel(context.Background())

	// Start background operations
	go n.startHealthChecking()
	go n.startMetricsCollection()

	// Update state
	n.state = cluster.NodeStateRunning
	n.health.Status = "healthy"

	n.logger.Info("VJVector node started successfully",
		"node_id", n.nodeID,
		"state", n.state)

	return nil
}

// Stop stops the VJVector node
func (n *VJVectorNode) Stop(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.state == cluster.NodeStateStopped {
		return nil
	}

	n.logger.Info("Stopping VJVector node", "node_id", n.nodeID)

	// Update state
	n.state = cluster.NodeStateStopping
	n.health.Status = "stopping"

	// Cancel background operations
	if n.cancel != nil {
		n.cancel()
	}

	// Stop clustering
	if n.cluster != nil {
		if err := n.cluster.Stop(ctx); err != nil {
			n.logger.Error("Failed to stop cluster", "error", err)
		}
	}

	// Close storage
	if n.storage != nil {
		if err := n.storage.Close(); err != nil {
			n.logger.Error("Failed to close storage", "error", err)
		}
	}

	// Update state
	n.state = cluster.NodeStateStopped
	n.health.Status = "stopped"

	n.logger.Info("VJVector node stopped", "node_id", n.nodeID)

	return nil
}

// GetNodeInfo returns information about the node
func (n *VJVectorNode) GetNodeInfo() *cluster.NodeInfo {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return &cluster.NodeInfo{
		ID:        n.nodeID,
		Address:   n.config.Address,
		Port:      n.config.Port,
		Role:      n.role,
		State:     n.state,
		Version:   n.version,
		StartTime: n.startTime,
		LastSeen:  time.Now(),
		Metadata: map[string]interface{}{
			"uptime": time.Since(n.startTime).String(),
			"health": n.health.Status,
		},
	}
}

// GetHealth returns the health status of the node
func (n *VJVectorNode) GetHealth() *NodeHealth {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Return a copy to avoid race conditions
	health := *n.health
	health.Uptime = time.Since(n.startTime)
	return &health
}

// GetMetrics returns the performance metrics of the node
func (n *VJVectorNode) GetMetrics() *NodeMetrics {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Return a copy to avoid race conditions
	metrics := *n.metrics
	return &metrics
}

// IsHealthy returns true if the node is healthy
func (n *VJVectorNode) IsHealthy() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.health.Status == "healthy"
}

// IsMaster returns true if this node is the master
func (n *VJVectorNode) IsMaster() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.role == cluster.NodeRoleMaster
}

// initializeServices initializes core services
func (n *VJVectorNode) initializeServices(ctx context.Context) error {
	n.logger.Info("Initializing core services", "node_id", n.nodeID)

	// Initialize embedding service
	if n.config.EmbeddingConfig != nil {
		embeddingService, err := embedding.NewService(n.config.EmbeddingConfig)
		if err != nil {
			return fmt.Errorf("failed to create embedding service: %w", err)
		}
		n.embeddingService = embeddingService
		n.health.Services["embedding"] = "healthy"
	}

	// Initialize RAG engine (requires embedding service and vector index)
	if n.config.RAGConfig != nil && n.embeddingService != nil && n.vectorIndex != nil {
		ragEngine, err := rag.NewEngine(n.config.RAGConfig, n.embeddingService, n.vectorIndex)
		if err != nil {
			return fmt.Errorf("failed to create RAG engine: %w", err)
		}
		n.ragEngine = ragEngine
		n.health.Services["rag"] = "healthy"
	}

	// Initialize vector index
	if n.config.IndexConfig != nil {
		vectorIndex, err := index.NewIndexFactory().CreateIndex(*n.config.IndexConfig)
		if err != nil {
			return fmt.Errorf("failed to create vector index: %w", err)
		}
		n.vectorIndex = vectorIndex
		n.health.Services["vector_index"] = "healthy"
	}

	// Initialize storage
	if n.config.StorageConfig != nil {
		storage, err := storage.NewStorageFactory().CreateStorage(*n.config.StorageConfig)
		if err != nil {
			return fmt.Errorf("failed to create storage: %w", err)
		}
		n.storage = storage
		n.health.Services["storage"] = "healthy"
	}

	n.logger.Info("Core services initialized successfully", "node_id", n.nodeID)
	return nil
}

// initializeClustering initializes clustering
func (n *VJVectorNode) initializeClustering(ctx context.Context) error {
	n.logger.Info("Initializing clustering", "node_id", n.nodeID)

	// TODO: Implement proper etcd client initialization
	// For now, skip clustering to avoid dependency issues
	n.logger.Info("Clustering temporarily disabled - etcd client not configured", "node_id", n.nodeID)
	return nil
}

// startHealthChecking starts the health checking process
func (n *VJVectorNode) startHealthChecking() {
	n.healthChecker.Start()
}

// startMetricsCollection starts the metrics collection process
func (n *VJVectorNode) startMetricsCollection() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.collectMetrics()
		}
	}
}

// collectMetrics collects current metrics
func (n *VJVectorNode) collectMetrics() {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Update timestamps
	n.metrics.LastUpdated = time.Now()

	// TODO: Collect actual resource usage metrics
	// This would involve:
	// 1. CPU usage from runtime/debug
	// 2. Memory usage from runtime/debug
	// 3. Disk usage from syscall
	// 4. Network I/O from netstat

	// Update health status based on metrics
	if n.metrics.FailedRequests > 0 && float64(n.metrics.FailedRequests)/float64(n.metrics.TotalRequests) > 0.1 {
		n.health.Status = "degraded"
		n.health.Warnings = append(n.health.Warnings, "High failure rate detected")
	} else {
		n.health.Status = "healthy"
		n.health.Warnings = nil
	}

	n.health.LastCheck = time.Now()
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(node *VJVectorNode, interval time.Duration, logger *slog.Logger) *HealthChecker {
	return &HealthChecker{
		node:     node,
		interval: interval,
		logger:   logger,
	}
}

// Start starts the health checker
func (h *HealthChecker) Start() {
	h.ticker = time.NewTicker(h.interval)
	defer h.ticker.Stop()

	for {
		select {
		case <-h.node.ctx.Done():
			return
		case <-h.ticker.C:
			h.checkHealth()
		}
	}
}

// checkHealth performs health checks
func (h *HealthChecker) checkHealth() {
	// TODO: Implement comprehensive health checks
	// This would involve:
	// 1. Service health checks
	// 2. Resource usage checks
	// 3. Dependency health checks
	// 4. Performance threshold checks

	h.logger.Debug("Health check completed", "node_id", h.node.nodeID)
}
