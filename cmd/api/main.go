package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/index"
	"github.com/vijaynallagatla/vjvector/pkg/storage"
)

// APIServer represents the VJVector REST API server
type APIServer struct {
	router     *mux.Router
	indexes    map[string]index.VectorIndex
	storage    storage.StorageEngine
	httpServer *http.Server
}

// NewAPIServer creates a new API server instance
func NewAPIServer() *APIServer {
	router := mux.NewRouter()
	server := &APIServer{
		router:  router,
		indexes: make(map[string]index.VectorIndex),
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
		log.Fatalf("Failed to create storage: %v", err)
	}
	server.storage = storageEngine

	// Setup routes
	server.setupRoutes()

	return server
}

// setupRoutes configures all API endpoints
func (s *APIServer) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")

	// Index management
	s.router.HandleFunc("/indexes", s.createIndex).Methods("POST")
	s.router.HandleFunc("/indexes", s.listIndexes).Methods("GET")
	s.router.HandleFunc("/indexes/{id}", s.getIndex).Methods("GET")
	s.router.HandleFunc("/indexes/{id}", s.deleteIndex).Methods("DELETE")

	// Vector operations
	s.router.HandleFunc("/indexes/{id}/vectors", s.insertVectors).Methods("POST")
	// TODO: Implement listVectors, getVector, deleteVector methods

	// Search operations
	s.router.HandleFunc("/indexes/{id}/search", s.searchVectors).Methods("POST")

	// Storage operations
	s.router.HandleFunc("/storage/stats", s.getStorageStats).Methods("GET")
	s.router.HandleFunc("/storage/compact", s.compactStorage).Methods("POST")

	// Performance metrics
	s.router.HandleFunc("/metrics", s.getMetrics).Methods("GET")
}

// healthCheck returns server health status
func (s *APIServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "VJVector API",
		"version":   "1.0.0",
	}
	writeJSON(w, http.StatusOK, response)
}

// createIndex creates a new vector index
func (s *APIServer) createIndex(w http.ResponseWriter, r *http.Request) {
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

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert string type to IndexType
	var indexType index.IndexType
	switch req.Type {
	case "hnsw":
		indexType = index.IndexTypeHNSW
	case "ivf":
		indexType = index.IndexTypeIVF
	default:
		writeError(w, http.StatusBadRequest, "Invalid index type. Must be 'hnsw' or 'ivf'")
		return
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
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create index: %v", err))
		return
	}

	s.indexes[req.ID] = idx

	response := map[string]interface{}{
		"id":      req.ID,
		"type":    req.Type,
		"status":  "created",
		"config":  config,
		"message": "Index created successfully",
	}
	writeJSON(w, http.StatusCreated, response)
}

// listIndexes returns all available indexes
func (s *APIServer) listIndexes(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, http.StatusOK, response)
}

// getIndex returns information about a specific index
func (s *APIServer) getIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	idx, exists := s.indexes[id]
	if !exists {
		writeError(w, http.StatusNotFound, "Index not found")
		return
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
	writeJSON(w, http.StatusOK, response)
}

// deleteIndex removes an index
func (s *APIServer) deleteIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	idx, exists := s.indexes[id]
	if !exists {
		writeError(w, http.StatusNotFound, "Index not found")
		return
	}

	if err := idx.Close(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to close index: %v", err))
		return
	}

	delete(s.indexes, id)

	response := map[string]interface{}{
		"id":      id,
		"status":  "deleted",
		"message": "Index deleted successfully",
	}
	writeJSON(w, http.StatusOK, response)
}

// insertVectors adds vectors to an index
func (s *APIServer) insertVectors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	idx, exists := s.indexes[id]
	if !exists {
		writeError(w, http.StatusNotFound, "Index not found")
		return
	}

	var req struct {
		Vectors []*core.Vector `json:"vectors"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	start := time.Now()
	for _, vector := range req.Vectors {
		if err := idx.Insert(vector); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to insert vector %s: %v", vector.ID, err))
			return
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
	writeJSON(w, http.StatusOK, response)
}

// searchVectors searches for similar vectors
func (s *APIServer) searchVectors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	idx, exists := s.indexes[id]
	if !exists {
		writeError(w, http.StatusNotFound, "Index not found")
		return
	}

	var req struct {
		Query []float64 `json:"query"`
		K     int       `json:"k"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.K <= 0 {
		req.K = 10 // Default to 10 results
	}

	start := time.Now()
	results, err := idx.Search(req.Query, req.K)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Search failed: %v", err))
		return
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
	writeJSON(w, http.StatusOK, response)
}

// getStorageStats returns storage statistics
func (s *APIServer) getStorageStats(w http.ResponseWriter, r *http.Request) {
	stats := s.storage.GetStats()
	response := map[string]interface{}{
		"total_vectors":  stats.TotalVectors,
		"storage_size":   stats.StorageSize,
		"memory_usage":   stats.MemoryUsage,
		"avg_write_time": stats.AvgWriteTime,
		"avg_read_time":  stats.AvgReadTime,
		"file_count":     stats.FileCount,
	}
	writeJSON(w, http.StatusOK, response)
}

// compactStorage compacts the storage
func (s *APIServer) compactStorage(w http.ResponseWriter, r *http.Request) {
	if err := s.storage.Compact(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Compaction failed: %v", err))
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Storage compacted successfully",
	}
	writeJSON(w, http.StatusOK, response)
}

// getMetrics returns performance metrics
func (s *APIServer) getMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"indexes_count": len(s.indexes),
		"uptime":        time.Since(time.Now()).String(),
		"memory_usage":  "N/A", // TODO: Implement actual memory monitoring
		"requests":      "N/A", // TODO: Implement request counting
	}
	writeJSON(w, http.StatusOK, metrics)
}

// Helper functions
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	response := map[string]interface{}{
		"error":   message,
		"status":  status,
		"success": false,
	}
	writeJSON(w, status, response)
}

// Start starts the API server
func (s *APIServer) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Graceful shutdown
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Printf("🚀 VJVector API Server started on %s", addr)
	log.Printf("📚 API Documentation available at %s/health", addr)

	return nil
}

// Shutdown gracefully shuts down the server
func (s *APIServer) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
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
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down VJVector API Server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("✅ VJVector API Server stopped gracefully")
}
