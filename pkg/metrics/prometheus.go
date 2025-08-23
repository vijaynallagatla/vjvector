package metrics

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics implements Prometheus metrics collection for VJVector
type PrometheusMetrics struct {
	mu sync.RWMutex

	// Cluster metrics
	clusterNodesTotal    prometheus.Gauge
	clusterNodesActive   prometheus.Gauge
	clusterShardsTotal   prometheus.Gauge
	clusterReplicasTotal prometheus.Gauge

	// Request metrics
	requestsTotal  prometheus.Counter
	requestsActive prometheus.Gauge
	requestLatency prometheus.Histogram
	requestErrors  prometheus.Counter

	// Vector operations
	vectorsTotal    prometheus.Gauge
	vectorsInserted prometheus.Counter
	vectorsDeleted  prometheus.Counter
	vectorsSearched prometheus.Counter

	// Storage metrics
	storageBytesUsed  prometheus.Gauge
	storageBytesTotal prometheus.Gauge
	storageOperations prometheus.Counter
	storageLatency    prometheus.Histogram

	// RAG metrics
	ragQueriesTotal    prometheus.Counter
	ragQueriesLatency  prometheus.Histogram
	ragEmbeddingsTotal prometheus.Counter
	ragContextHits     prometheus.Counter

	// Node metrics
	nodeCPUUsage    prometheus.Gauge
	nodeMemoryUsage prometheus.Gauge
	nodeDiskUsage   prometheus.Gauge

	// Network metrics
	networkBytesReceived prometheus.Counter
	networkBytesSent     prometheus.Counter
	networkLatency       prometheus.Histogram

	logger *slog.Logger
}

// NewPrometheusMetrics creates a new Prometheus metrics collector
func NewPrometheusMetrics(logger *slog.Logger) *PrometheusMetrics {
	metrics := &PrometheusMetrics{
		logger: logger,
	}

	// Initialize Prometheus metrics
	metrics.initializeMetrics()

	return metrics
}

// initializeMetrics initializes all Prometheus metrics
func (pm *PrometheusMetrics) initializeMetrics() {
	// Cluster metrics
	pm.clusterNodesTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vjvector_cluster_nodes_total",
		Help: "Total number of nodes in the cluster",
	})

	pm.clusterNodesActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vjvector_cluster_nodes_active",
		Help: "Number of active nodes in the cluster",
	})

	pm.clusterShardsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vjvector_cluster_shards_total",
		Help: "Total number of shards in the cluster",
	})

	pm.clusterReplicasTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vjvector_cluster_replicas_total",
		Help: "Total number of replicas in the cluster",
	})

	// Request metrics
	pm.requestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vjvector_requests_total",
		Help: "Total number of requests processed",
	})

	pm.requestsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vjvector_requests_active",
		Help: "Number of currently active requests",
	})

	pm.requestLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "vjvector_request_latency_seconds",
		Help:    "Request latency in seconds",
		Buckets: prometheus.DefBuckets,
	})

	pm.requestErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vjvector_request_errors_total",
		Help: "Total number of request errors",
	})

	// Vector operations
	pm.vectorsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vjvector_vectors_total",
		Help: "Total number of vectors stored",
	})

	pm.vectorsInserted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vjvector_vectors_inserted_total",
		Help: "Total number of vectors inserted",
	})

	pm.vectorsDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vjvector_vectors_deleted_total",
		Help: "Total number of vectors deleted",
	})

	pm.vectorsSearched = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vjvector_vectors_searched_total",
		Help: "Total number of vector searches performed",
	})

	// Storage metrics
	pm.storageBytesUsed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vjvector_storage_bytes_used",
		Help: "Storage space used in bytes",
	})

	pm.storageBytesTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vjvector_storage_bytes_total",
		Help: "Total storage space available in bytes",
	})

	pm.storageOperations = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vjvector_storage_operations_total",
		Help: "Total number of storage operations",
	})

	pm.storageLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "vjvector_storage_latency_seconds",
		Help:    "Storage operation latency in seconds",
		Buckets: prometheus.DefBuckets,
	})

	// RAG metrics
	pm.ragQueriesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vjvector_rag_queries_total",
		Help: "Total number of RAG queries processed",
	})

	pm.ragQueriesLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "vjvector_rag_queries_latency_seconds",
		Help:    "RAG query latency in seconds",
		Buckets: prometheus.DefBuckets,
	})

	pm.ragEmbeddingsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vjvector_rag_embeddings_total",
		Help: "Total number of embeddings generated",
	})

	pm.ragContextHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vjvector_rag_context_hits_total",
		Help: "Total number of context hits in RAG queries",
	})

	// Node metrics
	pm.nodeCPUUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vjvector_node_cpu_usage_percent",
		Help: "CPU usage percentage",
	})

	pm.nodeMemoryUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vjvector_node_memory_usage_percent",
		Help: "Memory usage percentage",
	})

	pm.nodeDiskUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vjvector_node_disk_usage_percent",
		Help: "Disk usage percentage",
	})

	// Network metrics
	pm.networkBytesReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vjvector_network_bytes_received_total",
		Help: "Total bytes received over network",
	})

	pm.networkBytesSent = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vjvector_network_bytes_sent_total",
		Help: "Total bytes sent over network",
	})

	pm.networkLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "vjvector_network_latency_seconds",
		Help:    "Network latency in seconds",
		Buckets: prometheus.DefBuckets,
	})

	pm.logger.Info("Prometheus metrics initialized successfully")
}

// UpdateClusterMetrics updates cluster-related metrics
func (pm *PrometheusMetrics) UpdateClusterMetrics(nodesTotal, nodesActive, shardsTotal, replicasTotal int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.clusterNodesTotal.Set(float64(nodesTotal))
	pm.clusterNodesActive.Set(float64(nodesActive))
	pm.clusterShardsTotal.Set(float64(shardsTotal))
	pm.clusterReplicasTotal.Set(float64(replicasTotal))
}

// RecordRequest records a new request
func (pm *PrometheusMetrics) RecordRequest(latency time.Duration, success bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.requestsTotal.Inc()
	pm.requestLatency.Observe(latency.Seconds())

	if !success {
		pm.requestErrors.Inc()
	}
}

// SetActiveRequests sets the number of currently active requests
func (pm *PrometheusMetrics) SetActiveRequests(count int) {
	pm.requestsActive.Set(float64(count))
}

// RecordVectorOperation records vector-related operations
func (pm *PrometheusMetrics) RecordVectorOperation(operation string, count int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	switch operation {
	case "insert":
		pm.vectorsInserted.Add(float64(count))
		pm.vectorsTotal.Add(float64(count))
	case "delete":
		pm.vectorsDeleted.Add(float64(count))
		pm.vectorsTotal.Sub(float64(count))
	case "search":
		pm.vectorsSearched.Add(float64(count))
	}
}

// SetVectorCount sets the total number of vectors
func (pm *PrometheusMetrics) SetVectorCount(count int) {
	pm.vectorsTotal.Set(float64(count))
}

// RecordStorageOperation records storage-related operations
func (pm *PrometheusMetrics) RecordStorageOperation(latency time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.storageOperations.Inc()
	pm.storageLatency.Observe(latency.Seconds())
}

// SetStorageUsage sets storage usage metrics
func (pm *PrometheusMetrics) SetStorageUsage(used, total int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.storageBytesUsed.Set(float64(used))
	pm.storageBytesTotal.Set(float64(total))
}

// RecordRAGQuery records RAG-related operations
func (pm *PrometheusMetrics) RecordRAGQuery(latency time.Duration, contextHits int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.ragQueriesTotal.Inc()
	pm.ragQueriesLatency.Observe(latency.Seconds())
	pm.ragContextHits.Add(float64(contextHits))
}

// RecordEmbedding records embedding generation
func (pm *PrometheusMetrics) RecordEmbedding() {
	pm.ragEmbeddingsTotal.Inc()
}

// UpdateNodeMetrics updates node resource usage metrics
func (pm *PrometheusMetrics) UpdateNodeMetrics(cpuPercent, memoryPercent, diskPercent float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.nodeCPUUsage.Set(cpuPercent)
	pm.nodeMemoryUsage.Set(memoryPercent)
	pm.nodeDiskUsage.Set(diskPercent)
}

// RecordNetworkActivity records network activity
func (pm *PrometheusMetrics) RecordNetworkActivity(bytesReceived, bytesSent int64, latency time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.networkBytesReceived.Add(float64(bytesReceived))
	pm.networkBytesSent.Add(float64(bytesSent))
	pm.networkLatency.Observe(latency.Seconds())
}

// GetMetrics returns the current metrics in Prometheus format
func (pm *PrometheusMetrics) GetMetrics() ([]byte, error) {
	// Prometheus automatically handles metrics collection
	// This method can be used for custom metrics export if needed
	return []byte("# VJVector Prometheus metrics are automatically exposed"), nil
}

// StartMetricsCollection starts periodic metrics collection
func (pm *PrometheusMetrics) StartMetricsCollection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	pm.logger.Info("Starting metrics collection", "interval", interval)

	for {
		select {
		case <-ctx.Done():
			pm.logger.Info("Stopping metrics collection")
			return
		case <-ticker.C:
			pm.collectSystemMetrics()
		}
	}
}

// collectSystemMetrics collects system-level metrics
func (pm *PrometheusMetrics) collectSystemMetrics() {
	// TODO: Implement system metrics collection
	// This would include CPU, memory, disk usage, etc.
	pm.logger.Debug("Collecting system metrics")
}

// Close cleans up the metrics collector
func (pm *PrometheusMetrics) Close() error {
	pm.logger.Info("Closing Prometheus metrics collector")
	return nil
}
