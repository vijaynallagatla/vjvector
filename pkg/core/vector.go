// Package core provides the fundamental data structures and interfaces for the VJVector database.
// It defines vectors, collections, and search operations that form the backbone of the system.
package core

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// Vector represents a vector in the database
type Vector struct {
	ID         string                 `json:"id"`
	Collection string                 `json:"collection"`
	Embedding  []float64              `json:"embedding"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Text       string                 `json:"text,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Dimension  int                    `json:"dimension"`
	Magnitude  float64                `json:"magnitude"`
	Normalized bool                   `json:"normalized"`
}

// NewVector creates a new vector with the given parameters
func NewVector(collection string, embedding []float64, text string, metadata map[string]interface{}) *Vector {
	now := time.Now()
	magnitude := calculateMagnitude(embedding)

	return &Vector{
		ID:         uuid.New().String(),
		Collection: collection,
		Embedding:  embedding,
		Metadata:   metadata,
		Text:       text,
		CreatedAt:  now,
		UpdatedAt:  now,
		Dimension:  len(embedding),
		Magnitude:  magnitude,
		Normalized: false,
	}
}

// Normalize normalizes the vector to unit length
func (v *Vector) Normalize() {
	if v.Normalized {
		return
	}

	if v.Magnitude == 0 {
		return
	}

	for i := range v.Embedding {
		v.Embedding[i] /= v.Magnitude
	}

	v.Magnitude = 1.0
	v.Normalized = true
	v.UpdatedAt = time.Now()
}

// Similarity calculates the cosine similarity with another vector
func (v *Vector) Similarity(other *Vector) (float64, error) {
	if v.Dimension != other.Dimension {
		return 0, fmt.Errorf("dimension mismatch: %d != %d", v.Dimension, other.Dimension)
	}

	dotProduct := 0.0
	for i := 0; i < v.Dimension; i++ {
		dotProduct += v.Embedding[i] * other.Embedding[i]
	}

	// Normalize by magnitudes if vectors aren't already normalized
	if !v.Normalized || !other.Normalized {
		return dotProduct / (v.Magnitude * other.Magnitude), nil
	}

	return dotProduct, nil
}

// Distance calculates the Euclidean distance to another vector
func (v *Vector) Distance(other *Vector) (float64, error) {
	if v.Dimension != other.Dimension {
		return 0, fmt.Errorf("dimension mismatch: %d != %d", v.Dimension, other.Dimension)
	}

	sum := 0.0
	for i := 0; i < v.Dimension; i++ {
		diff := v.Embedding[i] - other.Embedding[i]
		sum += diff * diff
	}

	return math.Sqrt(sum), nil
}

// calculateMagnitude calculates the magnitude (length) of a vector
func calculateMagnitude(embedding []float64) float64 {
	sum := 0.0
	for _, val := range embedding {
		sum += val * val
	}
	return math.Sqrt(sum)
}

// VectorSearchResult represents a search result with similarity score
type VectorSearchResult struct {
	Vector   *Vector `json:"vector"`
	Score    float64 `json:"score"`
	Distance float64 `json:"distance"`
}

// SearchQuery represents a search query
type SearchQuery struct {
	QueryVector []float64              `json:"query_vector"`
	Collection  string                 `json:"collection"`
	Limit       int                    `json:"limit"`
	Threshold   float64                `json:"threshold"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Collection represents a collection of vectors
type Collection struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Dimension   int                    `json:"dimension"`
	Count       int64                  `json:"count"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	IndexType   string                 `json:"index_type"`
}

// NewCollection creates a new collection
func NewCollection(name, description string, dimension int, indexType string) *Collection {
	now := time.Now()
	return &Collection{
		Name:        name,
		Description: description,
		Dimension:   dimension,
		Count:       0,
		CreatedAt:   now,
		UpdatedAt:   now,
		IndexType:   indexType,
		Metadata:    make(map[string]interface{}),
	}
}

// VectorRepository defines the interface for vector storage operations
type VectorRepository interface {
	Create(vector *Vector) error
	Get(id string) (*Vector, error)
	Update(vector *Vector) error
	Delete(id string) error
	Search(query *SearchQuery) ([]*VectorSearchResult, error)
	GetByCollection(collection string, limit, offset int) ([]*Vector, error)
	Count(collection string) (int64, error)
}

// CollectionRepository defines the interface for collection operations
type CollectionRepository interface {
	Create(collection *Collection) error
	Get(name string) (*Collection, error)
	Update(collection *Collection) error
	Delete(name string) error
	List() ([]*Collection, error)
}

// EmbeddingService defines the interface for text embedding generation
type EmbeddingService interface {
	EmbedText(text string) ([]float64, error)
	EmbedBatch(texts []string) ([][]float64, error)
	GetDimension() int
	GetModelName() string
}
