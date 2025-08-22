package index

import (
	"context"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

// HNSWIndex implements the Hierarchical Navigable Small World algorithm
// for approximate nearest neighbor search
type HNSWIndex struct {
	config     IndexConfig
	vectors    []*core.Vector
	layers     [][]*Node
	entryPoint *Node
	mutex      sync.RWMutex

	// Statistics
	stats     IndexStats
	startTime time.Time
}

// Node represents a node in the HNSW graph
type Node struct {
	ID      string    `json:"id"`
	Vector  []float64 `json:"vector"`
	Level   int       `json:"level"`
	Friends [][]int   `json:"friends"` // Friends at each level
}

// NewHNSWIndex creates a new HNSW index with the given configuration
func NewHNSWIndex(config IndexConfig) (VectorIndex, error) {
	if err := validateHNSWConfig(config); err != nil {
		return nil, err
	}

	index := &HNSWIndex{
		config:    config,
		vectors:   make([]*core.Vector, 0, config.MaxElements),
		layers:    make([][]*Node, config.MaxLayers),
		startTime: time.Now(),
	}

	// Initialize layers
	for i := range index.layers {
		index.layers[i] = make([]*Node, 0)
	}

	return index, nil
}

// Insert adds a vector to the HNSW index
func (h *HNSWIndex) Insert(vector *core.Vector) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if len(h.vectors) >= h.config.MaxElements {
		return ErrIndexFull
	}

	// Validate vector dimension
	if len(vector.Embedding) != h.config.Dimension {
		return ErrInvalidDimension
	}

	// Add vector to storage
	h.vectors = append(h.vectors, vector)

	// Create node
	node := &Node{
		ID:      vector.ID,
		Vector:  vector.Embedding,
		Level:   h.randomLevel(),
		Friends: make([][]int, h.config.MaxLayers),
	}

	// Add node to appropriate layers
	for level := 0; level <= node.Level; level++ {
		h.layers[level] = append(h.layers[level], node)
	}

	// Update entry point if this is the first node
	if h.entryPoint == nil {
		h.entryPoint = node
	} else {
		// Use the actual HNSW insertion algorithm
		if err := h.insertHNSW(vector); err != nil {
			return err
		}
	}

	// Update statistics
	h.stats.TotalVectors++

	return nil
}

// Search finds the k most similar vectors to the query vector
func (h *HNSWIndex) Search(query []float64, k int) ([]core.VectorSearchResult, error) {
	return h.SearchWithContext(context.Background(), query, k)
}

// SearchWithContext finds the k most similar vectors with context support
func (h *HNSWIndex) SearchWithContext(_ context.Context, query []float64, k int) ([]core.VectorSearchResult, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.entryPoint == nil {
		return nil, ErrIndexNotInitialized
	}

	if len(query) != h.config.Dimension {
		return nil, ErrInvalidDimension
	}

	if k <= 0 {
		return nil, ErrInvalidQuery
	}

	// Use the actual HNSW search algorithm
	return h.searchHNSW(query, k)
}

// Delete removes a vector from the index by ID
func (h *HNSWIndex) Delete(_ string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// TODO: Implement HNSW deletion algorithm
	// This is a placeholder - actual implementation will be added in Week 3-4

	return ErrVectorNotFound
}

// Optimize performs index optimization and maintenance
func (h *HNSWIndex) Optimize() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// TODO: Implement HNSW optimization
	// This is a placeholder - actual implementation will be added in Week 3-4

	return nil
}

// GetStats returns index performance and structure statistics
func (h *HNSWIndex) GetStats() IndexStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	stats := h.stats
	stats.NumLayers = len(h.layers)
	stats.MaxConnections = h.config.M

	// Calculate memory usage (rough estimate)
	stats.MemoryUsage = int64(len(h.vectors) * h.config.Dimension * 8) // 8 bytes per float64

	return stats
}

// Close performs cleanup and resource management
func (h *HNSWIndex) Close() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Clear data structures
	h.vectors = nil
	h.layers = nil
	h.entryPoint = nil

	return nil
}

// randomLevel generates a random level for a new node
// Uses the geometric distribution as described in the HNSW paper
func (h *HNSWIndex) randomLevel() int {
	level := 0
	for level < h.config.MaxLayers-1 && rand.Float64() < 0.5 {
		level++
	}
	return level
}

// validateHNSWConfig validates HNSW-specific configuration
func validateHNSWConfig(config IndexConfig) error {
	if config.M <= 0 {
		return ErrInvalidHNSWParameter
	}
	if config.EfConstruction <= 0 {
		return ErrInvalidHNSWParameter
	}
	if config.EfSearch <= 0 {
		return ErrInvalidHNSWParameter
	}
	if config.MaxLayers <= 0 {
		return ErrInvalidHNSWParameter
	}
	return nil
}

// findNodeIndexInLayer finds the index of a node in a specific layer
func (h *HNSWIndex) findNodeIndexInLayer(node *Node, level int) int {
	for i, layerNode := range h.layers[level] {
		if layerNode.ID == node.ID {
			return i
		}
	}
	return -1
}

// calculateDistance calculates the distance between two vectors
func (h *HNSWIndex) calculateDistance(a, b []float64) float64 {
	switch h.config.DistanceMetric {
	case "cosine":
		return h.cosineDistance(a, b)
	case "euclidean":
		return h.euclideanDistance(a, b)
	case "dot":
		return h.dotDistance(a, b)
	default:
		return h.cosineDistance(a, b) // Default to cosine
	}
}

// cosineDistance calculates cosine distance (1 - cosine similarity)
func (h *HNSWIndex) cosineDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
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
func (h *HNSWIndex) euclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}

	sum := 0.0
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

// dotDistance calculates dot product distance (negative dot product)
func (h *HNSWIndex) dotDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}

	dotProduct := 0.0
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
	}

	return -dotProduct // Negative because we want to minimize distance
}

// SearchResult represents a search result with distance and node
type SearchResult struct {
	Node     *Node
	Distance float64
}

// searchHNSW performs the main HNSW search algorithm
func (h *HNSWIndex) searchHNSW(query []float64, k int) ([]core.VectorSearchResult, error) {
	if h.entryPoint == nil {
		return nil, ErrIndexNotInitialized
	}

	// Start from the top layer
	currentLevel := h.entryPoint.Level
	currentNode := h.entryPoint
	currentDistance := h.calculateDistance(query, currentNode.Vector)

	// Find the best entry point by going down layers
	for level := currentLevel; level > 0; level-- {
		levelNodes := h.layers[level]
		if len(levelNodes) == 0 {
			continue
		}

		// Search in current level for better entry point
		candidates := h.searchLayer(query, []*Node{currentNode}, h.config.EfSearch, level)
		if len(candidates) > 0 && candidates[0].Distance < currentDistance {
			currentNode = candidates[0].Node
			currentDistance = candidates[0].Distance
		}
	}

	// Search in the bottom layer (level 0) with full efSearch
	results := h.searchLayer(query, []*Node{currentNode}, h.config.EfSearch, 0)

	// Convert to VectorSearchResult format
	vectorResults := make([]core.VectorSearchResult, 0, k)
	for i, result := range results {
		if i >= k {
			break
		}

		// Find the corresponding vector
		var vector *core.Vector
		for _, v := range h.vectors {
			if v.ID == result.Node.ID {
				vector = v
				break
			}
		}

		if vector != nil {
			vectorResults = append(vectorResults, core.VectorSearchResult{
				Vector:   vector,
				Distance: result.Distance,
				Score:    1.0 / (1.0 + result.Distance), // Convert distance to similarity score
			})
		}
	}

	return vectorResults, nil
}

// searchLayer searches for nearest neighbors in a specific layer
func (h *HNSWIndex) searchLayer(query []float64, entryPoints []*Node, ef int, level int) []*SearchResult {
	if len(entryPoints) == 0 {
		return nil
	}

	// Initialize candidates and visited sets
	candidates := make([]*SearchResult, 0, ef)
	visited := make(map[string]bool)

	// Add entry points to candidates
	for _, entry := range entryPoints {
		distance := h.calculateDistance(query, entry.Vector)
		candidates = append(candidates, &SearchResult{
			Node:     entry,
			Distance: distance,
		})
		visited[entry.ID] = true
	}

	// Sort candidates by distance
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Distance < candidates[j].Distance
	})

	// Keep only top ef candidates
	if len(candidates) > ef {
		candidates = candidates[:ef]
	}

	// Search through candidates
	for i := 0; i < len(candidates); i++ {
		current := candidates[i]

		// Explore friends of current node
		for _, friendIndex := range current.Node.Friends[level] {
			if friendIndex >= len(h.layers[level]) {
				continue
			}

			friend := h.layers[level][friendIndex]
			if friend == nil || visited[friend.ID] {
				continue
			}

			visited[friend.ID] = true
			distance := h.calculateDistance(query, friend.Vector)

			// Add to candidates if it's better than worst candidate
			if len(candidates) < ef || distance < candidates[len(candidates)-1].Distance {
				candidates = append(candidates, &SearchResult{
					Node:     friend,
					Distance: distance,
				})

				// Sort and keep top ef
				sort.Slice(candidates, func(i, j int) bool {
					return candidates[i].Distance < candidates[j].Distance
				})

				if len(candidates) > ef {
					candidates = candidates[:ef]
				}
			}
		}
	}

	return candidates
}

// insertHNSW performs the main HNSW insertion algorithm
func (h *HNSWIndex) insertHNSW(vector *core.Vector) error {
	// Generate random level for the new node
	level := h.randomLevel()

	// Create the new node
	newNode := &Node{
		ID:      vector.ID,
		Vector:  vector.Embedding,
		Level:   level,
		Friends: make([][]int, h.config.MaxLayers),
	}

	// Initialize friends arrays
	for i := range newNode.Friends {
		newNode.Friends[i] = make([]int, 0, h.config.M)
	}

	// Add node to appropriate layers
	for l := 0; l <= level; l++ {
		h.layers[l] = append(h.layers[l], newNode)
	}

	// If this is the first node, set it as entry point
	if h.entryPoint == nil {
		h.entryPoint = newNode
		return nil
	}

	// Find the best entry point for insertion
	entryPoint := h.findBestEntryPoint(vector.Embedding, level)

	// Insert connections at each level
	for l := 0; l <= level; l++ {
		h.insertConnectionsAtLevel(newNode, entryPoint, l)
	}

	return nil
}

// findBestEntryPoint finds the best entry point for insertion
func (h *HNSWIndex) findBestEntryPoint(query []float64, targetLevel int) *Node {
	currentNode := h.entryPoint
	currentDistance := h.calculateDistance(query, currentNode.Vector)

	// Start from the top layer and go down
	for level := h.entryPoint.Level; level > targetLevel; level-- {
		// Search in current level for better entry point
		candidates := h.searchLayer(query, []*Node{currentNode}, h.config.EfConstruction, level)
		if len(candidates) > 0 && candidates[0].Distance < currentDistance {
			currentNode = candidates[0].Node
			currentDistance = candidates[0].Distance
		}
	}

	return currentNode
}

// insertConnectionsAtLevel inserts connections for a new node at a specific level
func (h *HNSWIndex) insertConnectionsAtLevel(newNode *Node, entryPoint *Node, level int) {
	// Find candidates for connections at this level
	candidates := h.searchLayer(newNode.Vector, []*Node{entryPoint}, h.config.EfConstruction, level)

	// Select top M candidates for connections
	connections := h.selectConnections(candidates, h.config.M)

	// Add bidirectional connections
	for _, candidate := range connections {
		// Add candidate to newNode's friends (store index in layer)
		candidateIndex := h.findNodeIndexInLayer(candidate.Node, level)
		if candidateIndex >= 0 {
			newNode.Friends[level] = append(newNode.Friends[level], candidateIndex)
		}

		// Add newNode to candidate's friends (if there's space)
		newNodeIndex := h.findNodeIndexInLayer(newNode, level)
		if newNodeIndex >= 0 {
			h.addFriendToNode(candidate.Node, newNodeIndex, level)
		}
	}
}

// selectConnections selects the best connections from candidates
func (h *HNSWIndex) selectConnections(candidates []*SearchResult, maxConnections int) []*SearchResult {
	if len(candidates) <= maxConnections {
		return candidates
	}

	// Sort by distance and take top M
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Distance < candidates[j].Distance
	})

	return candidates[:maxConnections]
}

// addFriendToNode adds a friend to a node's friends list at a specific level
func (h *HNSWIndex) addFriendToNode(node *Node, friendIndex int, level int) {
	// Check if we already have this friend
	for _, existingIndex := range node.Friends[level] {
		if existingIndex == friendIndex {
			return // Already friends
		}
	}

	// Add new friend
	node.Friends[level] = append(node.Friends[level], friendIndex)

	// If we exceed M connections, remove the worst one
	if len(node.Friends[level]) > h.config.M {
		h.pruneConnections(node, level)
	}
}

// pruneConnections removes the worst connection to maintain M connections
func (h *HNSWIndex) pruneConnections(node *Node, level int) {
	if len(node.Friends[level]) <= h.config.M {
		return
	}

	// Find the worst connection (furthest from node)
	worstIndex := 0
	worstDistance := 0.0

	for i, friendIndex := range node.Friends[level] {
		// Find the friend node
		var friend *Node
		if friendIndex < len(h.layers[level]) {
			friend = h.layers[level][friendIndex]
		}

		if friend != nil {
			distance := h.calculateDistance(node.Vector, friend.Vector)
			if distance > worstDistance {
				worstDistance = distance
				worstIndex = i
			}
		}
	}

	// Remove the worst connection
	node.Friends[level] = append(node.Friends[level][:worstIndex], node.Friends[level][worstIndex+1:]...)
}
