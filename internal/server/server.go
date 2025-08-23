// Package server provides HTTP server functionality for the VJVector database.
// It includes RESTful API endpoints for vector operations, collections, and health checks.
package server

import (
	"context"
	"net/http"
	"time"

	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vijaynallagatla/vjvector/pkg/logging"
	"github.com/vijaynallagatla/vjvector/pkg/metrics"
)

// Server represents a generic HTTP server with Echo
type Server struct {
	echo      *echo.Echo
	logger    *slog.Logger
	startTime time.Time

	// Observability components
	structuredLogger *logging.StructuredLogger
	metrics          *metrics.PrometheusMetrics
}

// NewServer creates a new server instance
func NewServer() (*Server, error) {
	// Initialize structured logger
	logConfig := logging.DefaultConfig()
	structuredLogger, err := logging.NewStructuredLogger(logConfig)
	if err != nil {
		return nil, err
	}

	// Initialize Prometheus metrics
	prometheusMetrics := metrics.NewPrometheusMetrics(structuredLogger.Logger())

	// Create Echo instance
	e := echo.New()

	// Configure Echo
	e.HideBanner = true
	e.HidePort = true

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(middleware.Gzip())

	// Add custom middleware for metrics and logging
	e.Use(middleware.RequestID())
	e.Use(middleware.Gzip())

	server := &Server{
		echo:             e,
		logger:           structuredLogger.Logger(),
		structuredLogger: structuredLogger,
		metrics:          prometheusMetrics,
	}

	return server, nil
}

// Echo returns the underlying Echo instance for route configuration
func (s *Server) Echo() *echo.Echo {
	return s.echo
}

// Logger returns the server logger
func (s *Server) Logger() *slog.Logger {
	return s.logger
}

// Start starts the server
func (s *Server) Start(addr string) error {
	s.startTime = time.Now()

	s.logger.Info("Starting HTTP Server", "address", addr)

	// Start metrics collection
	ctx := context.Background()
	go s.metrics.StartMetricsCollection(ctx, 30*time.Second)

	// Start server in a goroutine
	go func() {
		if err := s.echo.Start(addr); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", "error", err)
		}
	}()

	s.logger.Info("🚀 HTTP Server started successfully", "address", addr)
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("🛑 Shutting down HTTP Server...")

	if err := s.echo.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown error", "error", err)
		return err
	}

	s.logger.Info("✅ HTTP Server stopped gracefully")
	return nil
}

// StartTime returns when the server started
func (s *Server) StartTime() time.Time {
	return s.startTime
}

// Metrics returns the Prometheus metrics collector
func (s *Server) Metrics() *metrics.PrometheusMetrics {
	return s.metrics
}

// StructuredLogger returns the structured logger
func (s *Server) StructuredLogger() *logging.StructuredLogger {
	return s.structuredLogger
}

// RecordRequest records metrics for an HTTP request
func (s *Server) RecordRequest(method, path string, statusCode int, latency time.Duration) {
	if s.metrics != nil {
		success := statusCode < 400
		s.metrics.RecordRequest(latency, success)
	}
}

// RecordVectorOperation records metrics for vector operations
func (s *Server) RecordVectorOperation(operation string, count int) {
	if s.metrics != nil {
		s.metrics.RecordVectorOperation(operation, count)
	}
}

// RecordRAGQuery records metrics for RAG operations
func (s *Server) RecordRAGQuery(latency time.Duration, contextHits int) {
	if s.metrics != nil {
		s.metrics.RecordRAGQuery(latency, contextHits)
	}
}
