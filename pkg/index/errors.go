package index

import "errors"

// Index-related errors
var (
	ErrUnsupportedIndexType = errors.New("unsupported index type")
	ErrInvalidDimension     = errors.New("invalid dimension")
	ErrInvalidMaxElements   = errors.New("invalid max elements")
	ErrInvalidHNSWParameter = errors.New("invalid HNSW parameter")
	ErrInvalidIVFParameter  = errors.New("invalid IVF parameter")
	ErrIndexNotInitialized  = errors.New("index not initialized")
	ErrVectorNotFound       = errors.New("vector not found")
	ErrInvalidQuery         = errors.New("invalid query vector")
	ErrIndexFull            = errors.New("index is full")
)
