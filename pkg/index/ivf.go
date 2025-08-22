package index

import (
	"context"
	"math"
	"sync"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

// IVFIndex implements the Inverted File Index algorithm
// for approximate nearest neighbor search using clustering
type IVFIndex struct {
	config     IndexConfig
	clusters   []*Cluster
	centroids  [][]float64
	assignment map[string]int // vector ID -> cluster ID
	mutex      sync.RWMutex

	// Statistics
	stats IndexStats
}

// Cluster represents a cluster in the IVF index
type Cluster struct {
	ID       int       `json:"id"`
	Centroid []float64 `json:"centroid"`
	Vectors  []string  `json:"vectors"` // Vector IDs in this cluster
	Size     int       `json:"size"`
}

// NewIVFIndex creates a new IVF index with the given configuration
func NewIVFIndex(config IndexConfig) (VectorIndex, error) {
	if err := validateIVFConfig(config); err != nil {
		return nil, err
	}

	index := &IVFIndex{
		config:     config,
		clusters:   make([]*Cluster, config.NumClusters),
		centroids:  make([][]float64, config.NumClusters),
		assignment: make(map[string]int),
	}

	// Initialize clusters
	for i := 0; i < config.NumClusters; i++ {
		index.clusters[i] = &Cluster{
			ID:      i,
			Vectors: make([]string, 0),
			Size:    0,
		}
	}

	return index, nil
}

// Insert adds a vector to the IVF index
func (i *IVFIndex) Insert(vector *core.Vector) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Validate vector dimension
	if len(vector.Embedding) != i.config.Dimension {
		return ErrInvalidDimension
	}

	// Use the actual IVF insertion algorithm
	if err := i.insertIVF(vector); err != nil {
		return err
	}

	// Update statistics
	i.stats.TotalVectors++

	return nil
}

// Search finds the k most similar vectors to the query vector
func (i *IVFIndex) Search(query []float64, k int) ([]core.VectorSearchResult, error) {
	return i.SearchWithContext(context.Background(), query, k)
}

// SearchWithContext finds the k most similar vectors with context support
func (i *IVFIndex) SearchWithContext(_ context.Context, query []float64, k int) ([]core.VectorSearchResult, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	if len(query) != i.config.Dimension {
		return nil, ErrInvalidDimension
	}

	if k <= 0 {
		return nil, ErrInvalidQuery
	}

	// Use the actual IVF search algorithm
	return i.searchIVF(query, k)
}

// Delete removes a vector from the index by ID
func (i *IVFIndex) Delete(id string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	clusterID, exists := i.assignment[id]
	if !exists {
		return ErrVectorNotFound
	}

	// Remove vector from cluster
	cluster := i.clusters[clusterID]
	for j, vectorID := range cluster.Vectors {
		if vectorID == id {
			// Remove from slice
			cluster.Vectors = append(cluster.Vectors[:j], cluster.Vectors[j+1:]...)
			cluster.Size--
			break
		}
	}

	// Remove from assignment
	delete(i.assignment, id)

	// Update statistics
	i.stats.TotalVectors--

	return nil
}

// Optimize performs index optimization and maintenance
func (i *IVFIndex) Optimize() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// TODO: Implement IVF optimization
	// This is a placeholder - actual implementation will be added in Week 5-6

	return nil
}

// GetStats returns index performance and structure statistics
func (i *IVFIndex) GetStats() IndexStats {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	stats := i.stats
	stats.NumClusters = len(i.clusters)

	// Calculate memory usage (rough estimate)
	stats.MemoryUsage = int64(len(i.assignment) * (i.config.Dimension*8 + 64)) // 8 bytes per float64 + overhead

	return stats
}

// Close performs cleanup and resource management
func (i *IVFIndex) Close() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Clear data structures
	i.clusters = nil
	i.centroids = nil
	i.assignment = nil

	return nil
}

// validateIVFConfig validates IVF-specific configuration
func validateIVFConfig(config IndexConfig) error {
	if config.NumClusters <= 0 {
		return ErrInvalidIVFParameter
	}
	if config.ClusterSize <= 0 {
		return ErrInvalidIVFParameter
	}
	return nil
}

// calculateDistance calculates the distance between two vectors
func (i *IVFIndex) calculateDistance(a, b []float64) float64 {
	switch i.config.DistanceMetric {
	case "cosine":
		return i.cosineDistance(a, b)
	case "euclidean":
		return i.euclideanDistance(a, b)
	case "dot":
		return i.dotDistance(a, b)
	default:
		return i.cosineDistance(a, b) // Default to cosine
	}
}

// cosineDistance calculates cosine distance (1 - cosine similarity)
func (i *IVFIndex) cosineDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for j := 0; j < len(a); j++ {
		dotProduct += a[j] * b[j]
		normA += a[j] * a[j]
		normB += b[j] * b[j]
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	if normA == 0 || normB == 0 {
		return 1.0
	}

	cosineSimilarity := dotProduct / (normA * normB)
	// Clamp to [-1, 1] to avoid numerical issues
	cosineSimilarity = math.Max(-1.0, math.Min(1.0, cosineSimilarity))

	return 1.0 - cosineSimilarity
}

// euclideanDistance calculates Euclidean distance
func (i *IVFIndex) euclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}

	sum := 0.0
	for j := 0; j < len(a); j++ {
		diff := a[j] - b[j]
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

// dotDistance calculates dot product distance (negative dot product)
func (i *IVFIndex) dotDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}

	dotProduct := 0.0
	for j := 0; j < len(a); j++ {
		dotProduct += a[j] * b[j]
	}

	dotProduct = math.Max(-1.0, math.Min(1.0, dotProduct))

	return -dotProduct // Negative because we want to minimize distance
}

// findNearestCluster finds the nearest cluster for a vector
func (i *IVFIndex) findNearestCluster(vector []float64) int {
	nearestCluster := 0
	minDistance := math.Inf(1)

	for clusterIndex, centroid := range i.centroids {
		distance := i.calculateDistance(vector, centroid)
		if distance < minDistance {
			minDistance = distance
			nearestCluster = clusterIndex
		}
	}

	return nearestCluster
}

// searchInCluster searches for similar vectors within a specific cluster
func (i *IVFIndex) searchInCluster(_ []float64, clusterID int, _ int) ([]core.VectorSearchResult, error) {
	cluster := i.clusters[clusterID]
	if cluster == nil || len(cluster.Vectors) == 0 {
		return nil, nil
	}

	// For now, return empty results since we don't have vector storage
	// In a real implementation, we would store vectors and search through them
	return nil, nil
}

// searchIVF performs the main IVF search algorithm
func (i *IVFIndex) searchIVF(query []float64, k int) ([]core.VectorSearchResult, error) {
	// Find the nearest cluster
	nearestCluster := i.findNearestCluster(query)

	// Search in the nearest cluster
	results, err := i.searchInCluster(query, nearestCluster, k)
	if err != nil {
		return nil, err
	}

	// For now, return empty results since we don't have full vector storage
	// In a real implementation, we would search through clusters
	return results, nil
}

// insertIVF performs the main IVF insertion algorithm
func (i *IVFIndex) insertIVF(vector *core.Vector) error {
	// Find the nearest cluster
	nearestCluster := i.findNearestCluster(vector.Embedding)

	// Add vector to the cluster
	i.clusters[nearestCluster].Vectors = append(i.clusters[nearestCluster].Vectors, vector.ID)
	i.clusters[nearestCluster].Size++

	// Store the assignment
	i.assignment[vector.ID] = nearestCluster

	// Update cluster centroid (simple moving average)
	cluster := i.clusters[nearestCluster]
	oldSize := cluster.Size - 1
	if oldSize > 0 {
		for cluster.Centroid == nil {
			// Initialize centroid if it doesn't exist
			cluster.Centroid = make([]float64, len(vector.Embedding))
		}
		for d := 0; d < len(vector.Embedding); d++ {
			cluster.Centroid[d] = (cluster.Centroid[d]*float64(oldSize) + vector.Embedding[d]) / float64(cluster.Size)
		}
	} else {
		// First vector in cluster, set centroid
		if cluster.Centroid == nil {
			cluster.Centroid = make([]float64, len(vector.Embedding))
		}
		copy(cluster.Centroid, vector.Embedding)
	}

	return nil
}
