package api

// Vector represents a vector in the API layer
type Vector struct {
	ID         string                 `json:"id"`
	Collection string                 `json:"collection,omitempty"`
	Embedding  []float64              `json:"embedding"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// CreateIndexRequest represents the request to create a new index
type CreateIndexRequest struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	Dimension      int    `json:"dimension"`
	MaxElements    int    `json:"max_elements"`
	M              int    `json:"m,omitempty"`
	EfConstruction int    `json:"ef_construction,omitempty"`
	EfSearch       int    `json:"ef_search,omitempty"`
	MaxLayers      int    `json:"max_layers,omitempty"`
	NumClusters    int    `json:"num_clusters,omitempty"`
	ClusterSize    int    `json:"cluster_size,omitempty"`
	DistanceMetric string `json:"distance_metric"`
	Normalize      bool   `json:"normalize"`
}

// InsertVectorsRequest represents the request to insert vectors
type InsertVectorsRequest struct {
	Vectors []*Vector `json:"vectors"`
}

// SearchRequest represents the request to search for similar vectors
type SearchRequest struct {
	Query []float64 `json:"query"`
	K     int       `json:"k"`
}
