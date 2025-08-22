// Package server provides HTTP server functionality for the VJVector database.
// It includes RESTful API endpoints for vector operations, collections, and health checks.
package server

import (
	"context"
	"net/http"
	"os"
	"time"

	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server represents a generic HTTP server with Echo
type Server struct {
	echo      *echo.Echo
	logger    *slog.Logger
	startTime time.Time
}

// NewServer creates a new server instance
func NewServer() *Server {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))

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

	server := &Server{
		echo:   e,
		logger: logger,
	}

	return server
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
