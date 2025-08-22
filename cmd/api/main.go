package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/index"
	"github.com/vijaynallagatla/vjvector/pkg/storage"
)

// APIServer represents the VJVector REST API server
type APIServer struct {
	echo      *echo.Echo
	indexes   map[string]index.VectorIndex
	storage   storage.StorageEngine
	logger    *slog.Logger
	startTime time.Time
}

// NewAPIServer creates a new API server instance
func NewAPIServer() *APIServer {
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

	server := &APIServer{
		echo:    e,
		indexes: make(map[string]index.VectorIndex),
		logger:  logger,
	}

	// Initialize storage
	storageConfig := storage.StorageConfig{
		Type:            storage.StorageTypeMemory,
		DataPath:        "/tmp/vjvector_api",
		PageSize:        4096,
		MaxFileSize:     1024 * 1024 * 1024, // 1GB
		BatchSize:       100,
		WriteBufferSize: 64 * 1024 * 1024, // 64MB
		CacheSize:       32 * 1024 * 1024, // 32MB
		MaxOpenFiles:    1000,
	}

	factory := &storage.DefaultStorageFactory{}
	storageEngine, err := factory.CreateStorage(storageConfig)
	if err != nil {
		logger.Error("Failed to create storage", "error", err)
		os.Exit(1)
	}
	server.storage = storageEngine

	// Setup routes
	server.setupRoutes()

	return server
}

// setupRoutes configures all API endpoints
func (s *APIServer) setupRoutes() {
	// Health check
	s.echo.GET("/health", s.healthCheck)

	// API v1 group
	v1 := s.echo.Group("/v1")

	// Index management
	indexes := v1.Group("/indexes")
	indexes.POST("", s.createIndex)
	indexes.GET("", s.listIndexes)
	indexes.GET("/:id", s.getIndex)
	indexes.DELETE("/:id", s.deleteIndex)

	// Vector operations
	indexes.POST("/:id/vectors", s.insertVectors)
	// TODO: Implement listVectors, getVector, deleteVector methods

	// Search operations
	indexes.POST("/:id/search", s.searchVectors)

	// Storage operations
	storage := v1.Group("/storage")
	storage.GET("/stats", s.getStorageStats)
	storage.POST("/compact", s.compactStorage)

	// Performance metrics
	v1.GET("/metrics", s.getMetrics)

	// OpenAPI documentation
	s.echo.GET("/openapi.yaml", s.serveOpenAPI)
	s.echo.GET("/docs", s.serveDocs)
}

// healthCheck returns server health status
func (s *APIServer) healthCheck(c echo.Context) error {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "VJVector API",
		"version":   "1.0.0",
		"uptime":    time.Since(s.startTime).String(),
	}

	s.logger.Info("Health check requested", "client_ip", c.RealIP())
	return c.JSON(http.StatusOK, response)
}

// createIndex creates a new vector index
func (s *APIServer) createIndex(c echo.Context) error {
	var req struct {
		ID             string `json:"id"`
		Type           string `json:"type"`
		Dimension      int    `json:"dimension"`
		MaxElements    int    `json:"max_elements"`
		M              int    `json:"m,omitempty"`
		EfConstruction int    `json:"ef_construction,omitempty"`
		EfSearch       int    `json:"ef_search,omitempty"`
		MaxLayers      int    `json:"max_layers,omitempty"`
		NumClusters    int    `json:"num_clusters,omitempty"`
		DistanceMetric string `json:"distance_metric"`
		Normalize      bool   `json:"normalize"`
	}

	if err := c.Bind(&req); err != nil {
		s.logger.Warn("Invalid request body", "error", err, "client_ip", c.RealIP())
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Convert string type to IndexType
	var indexType index.IndexType
	switch req.Type {
	case "hnsw":
		indexType = index.IndexTypeHNSW
	case "ivf":
		indexType = index.IndexTypeIVF
	default:
		s.logger.Warn("Invalid index type", "type", req.Type, "client_ip", c.RealIP())
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid index type. Must be 'hnsw' or 'ivf'")
	}

	config := index.IndexConfig{
		Type:           indexType,
		Dimension:      req.Dimension,
		MaxElements:    req.MaxElements,
		M:              req.M,
		EfConstruction: req.EfConstruction,
		EfSearch:       req.EfSearch,
		MaxLayers:      req.MaxLayers,
		NumClusters:    req.NumClusters,
		DistanceMetric: req.DistanceMetric,
		Normalize:      req.Normalize,
	}

	factory := index.NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		s.logger.Error("Failed to create index", "error", err, "config", config, "client_ip", c.RealIP())
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create index: %v", err))
	}

	s.indexes[req.ID] = idx

	s.logger.Info("Index created successfully", "id", req.ID, "type", req.Type, "dimension", req.Dimension, "client_ip", c.RealIP())

	response := map[string]interface{}{
		"id":      req.ID,
		"type":    req.Type,
		"status":  "created",
		"config":  config,
		"message": "Index created successfully",
	}
	return c.JSON(http.StatusCreated, response)
}

// listIndexes returns all available indexes
func (s *APIServer) listIndexes(c echo.Context) error {
	indexes := make([]map[string]interface{}, 0)

	for id, idx := range s.indexes {
		stats := idx.GetStats()
		indexInfo := map[string]interface{}{
			"id":              id,
			"total_vectors":   stats.TotalVectors,
			"memory_usage":    stats.MemoryUsage,
			"index_size":      stats.IndexSize,
			"avg_search_time": stats.AvgSearchTime,
			"avg_insert_time": stats.AvgInsertTime,
		}
		indexes = append(indexes, indexInfo)
	}

	response := map[string]interface{}{
		"indexes": indexes,
		"count":   len(indexes),
	}

	s.logger.Info("Indexes listed", "count", len(indexes), "client_ip", c.RealIP())
	return c.JSON(http.StatusOK, response)
}

// getIndex returns information about a specific index
func (s *APIServer) getIndex(c echo.Context) error {
	id := c.Param("id")

	idx, exists := s.indexes[id]
	if !exists {
		s.logger.Warn("Index not found", "id", id, "client_ip", c.RealIP())
		return echo.NewHTTPError(http.StatusNotFound, "Index not found")
	}

	stats := idx.GetStats()
	response := map[string]interface{}{
		"id":              id,
		"total_vectors":   stats.TotalVectors,
		"memory_usage":    stats.MemoryUsage,
		"index_size":      stats.IndexSize,
		"avg_search_time": stats.AvgSearchTime,
		"avg_insert_time": stats.AvgInsertTime,
	}

	s.logger.Info("Index info retrieved", "id", id, "client_ip", c.RealIP())
	return c.JSON(http.StatusOK, response)
}

// deleteIndex removes an index
func (s *APIServer) deleteIndex(c echo.Context) error {
	id := c.Param("id")

	idx, exists := s.indexes[id]
	if !exists {
		s.logger.Warn("Index not found for deletion", "id", id, "client_ip", c.RealIP())
		return echo.NewHTTPError(http.StatusNotFound, "Index not found")
	}

	if err := idx.Close(); err != nil {
		s.logger.Error("Failed to close index", "error", err, "id", id, "client_ip", c.RealIP())
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to close index: %v", err))
	}

	delete(s.indexes, id)

	s.logger.Info("Index deleted successfully", "id", id, "client_ip", c.RealIP())

	response := map[string]interface{}{
		"id":      id,
		"status":  "deleted",
		"message": "Index deleted successfully",
	}
	return c.JSON(http.StatusOK, response)
}

// insertVectors adds vectors to an index
func (s *APIServer) insertVectors(c echo.Context) error {
	id := c.Param("id")

	idx, exists := s.indexes[id]
	if !exists {
		s.logger.Warn("Index not found for vector insertion", "id", id, "client_ip", c.RealIP())
		return echo.NewHTTPError(http.StatusNotFound, "Index not found")
	}

	var req struct {
		Vectors []*core.Vector `json:"vectors"`
	}

	if err := c.Bind(&req); err != nil {
		s.logger.Warn("Invalid request body for vector insertion", "error", err, "index_id", id, "client_ip", c.RealIP())
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	start := time.Now()
	for _, vector := range req.Vectors {
		if err := idx.Insert(vector); err != nil {
			s.logger.Error("Failed to insert vector", "error", err, "vector_id", vector.ID, "index_id", id, "client_ip", c.RealIP())
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to insert vector %s: %v", vector.ID, err))
		}
	}
	duration := time.Since(start)

	stats := idx.GetStats()

	s.logger.Info("Vectors inserted successfully",
		"index_id", id,
		"vectors_added", len(req.Vectors),
		"insert_time", duration,
		"total_vectors", stats.TotalVectors,
		"client_ip", c.RealIP())

	response := map[string]interface{}{
		"index_id":      id,
		"vectors_added": len(req.Vectors),
		"total_vectors": stats.TotalVectors,
		"insert_time":   duration.String(),
		"message":       "Vectors inserted successfully",
	}
	return c.JSON(http.StatusOK, response)
}

// searchVectors searches for similar vectors
func (s *APIServer) searchVectors(c echo.Context) error {
	id := c.Param("id")

	idx, exists := s.indexes[id]
	if !exists {
		s.logger.Warn("Index not found for search", "id", id, "client_ip", c.RealIP())
		return echo.NewHTTPError(http.StatusNotFound, "Index not found")
	}

	var req struct {
		Query []float64 `json:"query"`
		K     int       `json:"k"`
	}

	if err := c.Bind(&req); err != nil {
		s.logger.Warn("Invalid request body for search", "error", err, "index_id", id, "client_ip", c.RealIP())
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.K <= 0 {
		req.K = 10 // Default to 10 results
	}

	start := time.Now()
	results, err := idx.Search(req.Query, req.K)
	if err != nil {
		s.logger.Error("Search failed", "error", err, "index_id", id, "client_ip", c.RealIP())
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Search failed: %v", err))
	}
	duration := time.Since(start)

	// Convert results to API response format
	apiResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		apiResults[i] = map[string]interface{}{
			"vector_id": result.Vector.ID,
			"score":     result.Score,
			"distance":  result.Distance,
		}
	}

	s.logger.Info("Search completed successfully",
		"index_id", id,
		"query_dimension", len(req.Query),
		"k", req.K,
		"results_found", len(results),
		"search_time", duration,
		"client_ip", c.RealIP())

	response := map[string]interface{}{
		"index_id":    id,
		"query":       req.Query,
		"k":           req.K,
		"results":     apiResults,
		"search_time": duration.String(),
		"count":       len(results),
	}
	return c.JSON(http.StatusOK, response)
}

// getStorageStats returns storage statistics
func (s *APIServer) getStorageStats(c echo.Context) error {
	stats := s.storage.GetStats()
	response := map[string]interface{}{
		"total_vectors":  stats.TotalVectors,
		"storage_size":   stats.StorageSize,
		"memory_usage":   stats.MemoryUsage,
		"avg_write_time": stats.AvgWriteTime,
		"avg_read_time":  stats.AvgReadTime,
		"file_count":     stats.FileCount,
	}

	s.logger.Info("Storage stats retrieved", "client_ip", c.RealIP())
	return c.JSON(http.StatusOK, response)
}

// compactStorage compacts the storage
func (s *APIServer) compactStorage(c echo.Context) error {
	if err := s.storage.Compact(); err != nil {
		s.logger.Error("Storage compaction failed", "error", err, "client_ip", c.RealIP())
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Compaction failed: %v", err))
	}

	s.logger.Info("Storage compaction completed successfully", "client_ip", c.RealIP())

	response := map[string]interface{}{
		"status":  "success",
		"message": "Storage compacted successfully",
	}
	return c.JSON(http.StatusOK, response)
}

// getMetrics returns performance metrics
func (s *APIServer) getMetrics(c echo.Context) error {
	metrics := map[string]interface{}{
		"indexes_count": len(s.indexes),
		"uptime":        time.Since(s.startTime).String(),
		"memory_usage":  "N/A", // TODO: Implement actual memory monitoring
		"requests":      "N/A", // TODO: Implement request counting
	}

	s.logger.Info("Metrics retrieved", "client_ip", c.RealIP())
	return c.JSON(http.StatusOK, metrics)
}

// serveOpenAPI serves the OpenAPI specification
func (s *APIServer) serveOpenAPI(c echo.Context) error {
	openAPIPath := "docs/api/openapi.yaml"
	content, err := os.ReadFile(openAPIPath)
	if err != nil {
		s.logger.Error("Failed to read OpenAPI spec", "error", err, "path", openAPIPath)
		return echo.NewHTTPError(http.StatusNotFound, "OpenAPI specification not found")
	}

	c.Response().Header().Set("Content-Type", "text/yaml")
	return c.String(http.StatusOK, string(content))
}

// serveDocs serves a simple HTML documentation page
func (s *APIServer) serveDocs(c echo.Context) error {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>VJVector API Documentation</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui.css" rel="stylesheet">
    <style>
        body { margin: 0; padding: 20px; background: #fafafa; }
        .swagger-ui { max-width: 1200px; margin: 0 auto; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: '/openapi.yaml',
                dom_id: '#swagger-ui',
                presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
                layout: "BaseLayout"
            });
        };
    </script>
</body>
</html>`

	c.Response().Header().Set("Content-Type", "text/html")
	return c.HTML(http.StatusOK, html)
}

// Start starts the API server
func (s *APIServer) Start(addr string) error {
	s.startTime = time.Now()

	s.logger.Info("Starting VJVector API Server", "address", addr)

	// Start server in a goroutine
	go func() {
		if err := s.echo.Start(addr); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", "error", err)
		}
	}()

	s.logger.Info("🚀 VJVector API Server started successfully", "address", addr)
	s.logger.Info("📚 API Documentation available at", "docs_url", fmt.Sprintf("%s/docs", addr))
	s.logger.Info("🔍 OpenAPI spec available at", "openapi_url", fmt.Sprintf("%s/openapi.yaml", addr))

	return nil
}

// Shutdown gracefully shuts down the server
func (s *APIServer) Shutdown(ctx context.Context) error {
	s.logger.Info("🛑 Shutting down VJVector API Server...")

	if err := s.echo.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown error", "error", err)
		return err
	}

	s.logger.Info("✅ VJVector API Server stopped gracefully")
	return nil
}

func main() {
	// Get port from environment or use default
	port := os.Getenv("VJVECTOR_PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	// Create and start server
	server := NewAPIServer()
	if err := server.Start(addr); err != nil {
		server.logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		server.logger.Error("Server shutdown error", "error", err)
	}
}
