package api

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/index"
	"github.com/vijaynallagatla/vjvector/pkg/storage"
)

// Handlers represents the API handlers for VJVector
type Handlers struct {
	indexes map[string]index.VectorIndex
	storage storage.StorageEngine
}

// NewHandlers creates new API handlers
func NewHandlers() *Handlers {
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
		panic(fmt.Sprintf("Failed to create storage: %v", err))
	}

	return &Handlers{
		indexes: make(map[string]index.VectorIndex),
		storage: storageEngine,
	}
}

// RegisterRoutes registers all API routes with the Echo instance
func (h *Handlers) RegisterRoutes(e *echo.Echo) {
	// Health check
	e.GET("/health", h.healthCheck)

	// API v1 group
	v1 := e.Group("/v1")

	// Index management
	indexes := v1.Group("/indexes")
	indexes.POST("", h.createIndex)
	indexes.GET("", h.listIndexes)
	indexes.GET("/:id", h.getIndex)
	indexes.DELETE("/:id", h.deleteIndex)

	// Vector operations
	indexes.POST("/:id/vectors", h.insertVectors)
	// TODO: Implement listVectors, getVector, deleteVector methods

	// Search operations
	indexes.POST("/:id/search", h.searchVectors)

	// Storage operations
	storage := v1.Group("/storage")
	storage.GET("/stats", h.getStorageStats)
	storage.POST("/compact", h.compactStorage)

	// Performance metrics
	v1.GET("/metrics", h.getMetrics)

	// OpenAPI documentation
	e.GET("/openapi.yaml", h.serveOpenAPI)
	e.GET("/docs", h.serveDocs)
}

// healthCheck returns server health status
func (h *Handlers) healthCheck(c echo.Context) error {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "VJVector API",
		"version":   "1.0.0",
		"uptime":    "N/A", // Will be set by server package
	}

	return c.JSON(http.StatusOK, response)
}

// createIndex creates a new vector index
func (h *Handlers) createIndex(c echo.Context) error {
	var req CreateIndexRequest

	if err := c.Bind(&req); err != nil {
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
		ClusterSize:    req.ClusterSize,
		DistanceMetric: req.DistanceMetric,
		Normalize:      req.Normalize,
	}

	factory := index.NewIndexFactory()
	idx, err := factory.CreateIndex(config)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create index: %v", err))
	}

	h.indexes[req.ID] = idx

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
func (h *Handlers) listIndexes(c echo.Context) error {
	indexes := make([]map[string]interface{}, 0)

	for id, idx := range h.indexes {
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

	return c.JSON(http.StatusOK, response)
}

// getIndex returns information about a specific index
func (h *Handlers) getIndex(c echo.Context) error {
	id := c.Param("id")

	idx, exists := h.indexes[id]
	if !exists {
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

	return c.JSON(http.StatusOK, response)
}

// deleteIndex removes an index
func (h *Handlers) deleteIndex(c echo.Context) error {
	id := c.Param("id")

	idx, exists := h.indexes[id]
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Index not found")
	}

	if err := idx.Close(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to close index: %v", err))
	}

	delete(h.indexes, id)

	response := map[string]interface{}{
		"id":      id,
		"status":  "deleted",
		"message": "Index deleted successfully",
	}
	return c.JSON(http.StatusOK, response)
}

// insertVectors adds vectors to an index
func (h *Handlers) insertVectors(c echo.Context) error {
	id := c.Param("id")

	idx, exists := h.indexes[id]
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Index not found")
	}

	var req InsertVectorsRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	start := time.Now()
	for _, vector := range req.Vectors {
		// Convert to core.Vector
		coreVector := &core.Vector{
			ID:         vector.ID,
			Collection: vector.Collection,
			Embedding:  vector.Embedding,
			Metadata:   vector.Metadata,
		}

		if err := idx.Insert(coreVector); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to insert vector %s: %v", vector.ID, err))
		}
	}
	duration := time.Since(start)

	stats := idx.GetStats()

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
func (h *Handlers) searchVectors(c echo.Context) error {
	id := c.Param("id")

	idx, exists := h.indexes[id]
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "Index not found")
	}

	var req SearchRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.K <= 0 {
		req.K = 10 // Default to 10 results
	}

	start := time.Now()
	results, err := idx.Search(req.Query, req.K)
	if err != nil {
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
func (h *Handlers) getStorageStats(c echo.Context) error {
	stats := h.storage.GetStats()
	response := map[string]interface{}{
		"total_vectors":  stats.TotalVectors,
		"storage_size":   stats.StorageSize,
		"memory_usage":   stats.MemoryUsage,
		"avg_write_time": stats.AvgWriteTime,
		"avg_read_time":  stats.AvgReadTime,
		"file_count":     stats.FileCount,
	}

	return c.JSON(http.StatusOK, response)
}

// compactStorage compacts the storage
func (h *Handlers) compactStorage(c echo.Context) error {
	if err := h.storage.Compact(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Compaction failed: %v", err))
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Storage compacted successfully",
	}
	return c.JSON(http.StatusOK, response)
}

// getMetrics returns performance metrics
func (h *Handlers) getMetrics(c echo.Context) error {
	metrics := map[string]interface{}{
		"indexes_count": len(h.indexes),
		"uptime":        "N/A", // Will be set by server package
		"memory_usage":  "N/A", // TODO: Implement actual memory monitoring
		"requests":      "N/A", // TODO: Implement request counting
	}

	return c.JSON(http.StatusOK, metrics)
}

// serveOpenAPI serves the OpenAPI specification
func (h *Handlers) serveOpenAPI(c echo.Context) error {
	openAPIPath := "docs/api/openapi.yaml"
	content, err := os.ReadFile(openAPIPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "OpenAPI specification not found")
	}

	c.Response().Header().Set("Content-Type", "text/yaml")
	return c.String(http.StatusOK, string(content))
}

// serveDocs serves a simple HTML documentation page
func (h *Handlers) serveDocs(c echo.Context) error {
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
