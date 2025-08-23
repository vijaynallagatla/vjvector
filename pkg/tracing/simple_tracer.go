package tracing

import (
	"context"
	"log/slog"
	"time"
)

// TraceConfig represents tracing configuration
type TraceConfig struct {
	ServiceName    string `json:"service_name"`
	ServiceVersion string `json:"service_version"`
	Environment    string `json:"environment"`
	Enabled        bool   `json:"enabled"`
}

// DefaultConfig returns the default tracing configuration
func DefaultConfig() *TraceConfig {
	return &TraceConfig{
		ServiceName:    "vjvector",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		Enabled:        true,
	}
}

// SimpleTracer provides basic tracing capabilities
type SimpleTracer struct {
	config *TraceConfig
	logger *slog.Logger
}

// NewSimpleTracer creates a new simple tracer
func NewSimpleTracer(config *TraceConfig, logger *slog.Logger) (*SimpleTracer, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if !config.Enabled {
		logger.Info("Tracing is disabled")
	}

	return &SimpleTracer{
		config: config,
		logger: logger,
	}, nil
}

// TraceHTTPRequest traces an HTTP request
func (st *SimpleTracer) TraceHTTPRequest(ctx context.Context, method, path string, statusCode int, latency time.Duration) {
	if !st.config.Enabled {
		return
	}

	st.logger.Info("HTTP Request Trace",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"latency_ms", latency.Milliseconds(),
		"timestamp", time.Now(),
	)
}

// TraceDatabaseOperation traces a database operation
func (st *SimpleTracer) TraceDatabaseOperation(ctx context.Context, operation, table string, latency time.Duration, rowsAffected int, err error) {
	if !st.config.Enabled {
		return
	}

	args := []any{
		"operation", operation,
		"table", table,
		"latency_ms", latency.Milliseconds(),
		"rows_affected", rowsAffected,
		"timestamp", time.Now(),
	}

	if err != nil {
		args = append(args, "error", err.Error())
		st.logger.Error("Database Operation Trace", args...)
	} else {
		st.logger.Info("Database Operation Trace", args...)
	}
}

// TraceVectorOperation traces a vector operation
func (st *SimpleTracer) TraceVectorOperation(ctx context.Context, operation, collection string, vectorCount int, latency time.Duration, err error) {
	if !st.config.Enabled {
		return
	}

	args := []any{
		"operation", operation,
		"collection", collection,
		"vector_count", vectorCount,
		"latency_ms", latency.Milliseconds(),
		"timestamp", time.Now(),
	}

	if err != nil {
		args = append(args, "error", err.Error())
		st.logger.Error("Vector Operation Trace", args...)
	} else {
		st.logger.Info("Vector Operation Trace", args...)
	}
}

// TraceRAGOperation traces a RAG operation
func (st *SimpleTracer) TraceRAGOperation(ctx context.Context, operation, query string, contextHits int, latency time.Duration, err error) {
	if !st.config.Enabled {
		return
	}

	args := []any{
		"operation", operation,
		"query", query,
		"context_hits", contextHits,
		"latency_ms", latency.Milliseconds(),
		"timestamp", time.Now(),
	}

	if err != nil {
		args = append(args, "error", err.Error())
		st.logger.Error("RAG Operation Trace", args...)
	} else {
		st.logger.Info("RAG Operation Trace", args...)
	}
}

// TraceClusterOperation traces a cluster operation
func (st *SimpleTracer) TraceClusterOperation(ctx context.Context, operation, nodeID string, details map[string]interface{}) {
	if !st.config.Enabled {
		return
	}

	args := []any{
		"operation", operation,
		"node_id", nodeID,
		"timestamp", time.Now(),
	}

	// Add additional details
	for k, v := range details {
		args = append(args, k, v)
	}

	st.logger.Info("Cluster Operation Trace", args...)
}

// TracePerformance traces performance metrics
func (st *SimpleTracer) TracePerformance(ctx context.Context, metric string, value float64, unit string, tags map[string]string) {
	if !st.config.Enabled {
		return
	}

	args := []any{
		"metric", metric,
		"value", value,
		"unit", unit,
		"timestamp", time.Now(),
	}

	// Add tags
	for k, v := range tags {
		args = append(args, k, v)
	}

	st.logger.Info("Performance Metric Trace", args...)
}

// Close closes the tracer and performs cleanup
func (st *SimpleTracer) Close() error {
	if !st.config.Enabled {
		return nil
	}

	st.logger.Info("Closing simple tracer")
	return nil
}
