package logging

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogConfig represents logging configuration
type LogConfig struct {
	Level      LogLevel `json:"level"`
	Format     string   `json:"format"` // "json" or "text"
	Output     string   `json:"output"` // "stdout", "stderr", or file path
	AddSource  bool     `json:"add_source"`
	AddTime    bool     `json:"add_time"`
	MaxSize    int      `json:"max_size"`    // Maximum log file size in MB
	MaxBackups int      `json:"max_backups"` // Maximum number of backup files
	MaxAge     int      `json:"max_age"`     // Maximum age of log files in days
}

// DefaultConfig returns the default logging configuration
func DefaultConfig() *LogConfig {
	return &LogConfig{
		Level:      LogLevelInfo,
		Format:     "json",
		Output:     "stdout",
		AddSource:  true,
		AddTime:    true,
		MaxSize:    100, // 100 MB
		MaxBackups: 3,   // Keep 3 backup files
		MaxAge:     30,  // Keep logs for 30 days
	}
}

// StructuredLogger provides structured logging capabilities
type StructuredLogger struct {
	logger *slog.Logger
	config *LogConfig
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(config *LogConfig) (*StructuredLogger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     parseLogLevel(config.Level),
		AddSource: config.AddSource,
	}

	// Create handler based on format
	var handler slog.Handler
	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return &StructuredLogger{
		logger: logger,
		config: config,
	}, nil
}

// parseLogLevel converts LogLevel to slog.Level
func parseLogLevel(level LogLevel) slog.Level {
	switch level {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelInfo:
		return slog.LevelInfo
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Logger returns the underlying slog.Logger
func (sl *StructuredLogger) Logger() *slog.Logger {
	return sl.logger
}

// Debug logs a debug message
func (sl *StructuredLogger) Debug(msg string, args ...any) {
	sl.logger.Debug(msg, args...)
}

// Info logs an info message
func (sl *StructuredLogger) Info(msg string, args ...any) {
	sl.logger.Info(msg, args...)
}

// Warn logs a warning message
func (sl *StructuredLogger) Warn(msg string, args ...any) {
	sl.logger.Warn(msg, args...)
}

// Error logs an error message
func (sl *StructuredLogger) Error(msg string, args ...any) {
	sl.logger.Error(msg, args...)
}

// WithContext creates a logger with context
func (sl *StructuredLogger) WithContext(ctx context.Context) *slog.Logger {
	// slog.Logger doesn't have WithContext, return the logger as is
	// Context can be passed to individual log calls
	return sl.logger
}

// WithFields creates a logger with additional fields
func (sl *StructuredLogger) WithFields(fields map[string]any) *slog.Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return sl.logger.With(args...)
}

// LogRequest logs an HTTP request
func (sl *StructuredLogger) LogRequest(ctx context.Context, method, path string, statusCode int, latency time.Duration, userID string) {
	sl.logger.Info("HTTP Request",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"latency_ms", latency.Milliseconds(),
		"user_id", userID,
		"timestamp", time.Now(),
	)
}

// LogDatabase logs a database operation
func (sl *StructuredLogger) LogDatabase(ctx context.Context, operation, table string, latency time.Duration, rowsAffected int, error error) {
	args := []any{
		"operation", operation,
		"table", table,
		"latency_ms", latency.Milliseconds(),
		"rows_affected", rowsAffected,
		"timestamp", time.Now(),
	}

	if error != nil {
		args = append(args, "error", error.Error())
		sl.logger.Error("Database Operation", args...)
	} else {
		sl.logger.Info("Database Operation", args...)
	}
}

// LogVectorOperation logs a vector operation
func (sl *StructuredLogger) LogVectorOperation(ctx context.Context, operation, collection string, vectorCount int, latency time.Duration, error error) {
	args := []any{
		"operation", operation,
		"collection", collection,
		"vector_count", vectorCount,
		"latency_ms", latency.Milliseconds(),
		"timestamp", time.Now(),
	}

	if error != nil {
		args = append(args, "error", error.Error())
		sl.logger.Error("Vector Operation", args...)
	} else {
		sl.logger.Info("Vector Operation", args...)
	}
}

// LogRAGOperation logs a RAG operation
func (sl *StructuredLogger) LogRAGOperation(ctx context.Context, operation, query string, contextHits int, latency time.Duration, error error) {
	args := []any{
		"operation", operation,
		"query", query,
		"context_hits", contextHits,
		"latency_ms", latency.Milliseconds(),
		"timestamp", time.Now(),
	}

	if error != nil {
		args = append(args, "error", error.Error())
		sl.logger.Error("RAG Operation", args...)
	} else {
		sl.logger.Info("RAG Operation", args...)
	}
}

// LogClusterEvent logs a cluster event
func (sl *StructuredLogger) LogClusterEvent(ctx context.Context, eventType, nodeID string, details map[string]any) {
	args := []any{
		"event_type", eventType,
		"node_id", nodeID,
		"timestamp", time.Now(),
	}

	// Add additional details
	for k, v := range details {
		args = append(args, k, v)
	}

	sl.logger.Info("Cluster Event", args...)
}

// LogPerformance logs performance metrics
func (sl *StructuredLogger) LogPerformance(ctx context.Context, metric string, value float64, unit string, tags map[string]string) {
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

	sl.logger.Info("Performance Metric", args...)
}

// LogSecurity logs security-related events
func (sl *StructuredLogger) LogSecurity(ctx context.Context, eventType, userID, resource string, success bool, details map[string]any) {
	args := []any{
		"event_type", eventType,
		"user_id", userID,
		"resource", resource,
		"success", success,
		"timestamp", time.Now(),
	}

	// Add additional details
	for k, v := range details {
		args = append(args, k, v)
	}

	if success {
		sl.logger.Info("Security Event", args...)
	} else {
		sl.logger.Warn("Security Event", args...)
	}
}

// Close closes the logger and performs cleanup
func (sl *StructuredLogger) Close() error {
	// TODO: Implement cleanup logic for file handlers, etc.
	return nil
}
