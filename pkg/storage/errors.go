package storage

import "errors"

// Storage-related errors
var (
	ErrUnsupportedStorageType = errors.New("unsupported storage type")
	ErrInvalidDataPath        = errors.New("invalid data path")
	ErrInvalidMaxFileSize     = errors.New("invalid max file size")
	ErrInvalidBatchSize       = errors.New("invalid batch size")
	ErrInvalidPageSize        = errors.New("invalid page size")
	ErrInvalidCacheSize       = errors.New("invalid cache size")
	ErrInvalidWriteBufferSize = errors.New("invalid write buffer size")
	ErrInvalidMaxOpenFiles    = errors.New("invalid max open files")
	ErrStorageNotInitialized  = errors.New("storage not initialized")
	ErrVectorNotFound         = errors.New("vector not found")
	ErrWriteFailed            = errors.New("write operation failed")
	ErrReadFailed             = errors.New("read operation failed")
	ErrDeleteFailed           = errors.New("delete operation failed")
)
