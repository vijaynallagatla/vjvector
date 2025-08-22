# Cursor AI Instructions for VJVector Repository

## 🎯 Project Overview

**VJVector** is an AI-first vector database built from scratch in Go, designed specifically for RAG (Retrieval-Augmented Generation) applications. This repository implements a high-performance vector database with native AI embedding support.

## 🏗️ Architecture Principles

### 1. AI-First Design
- **Primary Focus**: Optimize for AI embedding workflows and RAG applications
- **Vector Operations**: Prioritize vector similarity search and nearest neighbor algorithms
- **Metadata Support**: Rich metadata storage for AI context and filtering
- **Batch Processing**: Efficient batch operations for AI model outputs

### 2. Modular Architecture
- **Separation of Concerns**: Clear boundaries between core, storage, indexing, and API layers
- **Interface-Driven**: Use Go interfaces for loose coupling and testability
- **Plugin System**: Extensible design for different embedding models and index types

### 3. Performance-First
- **Memory Efficiency**: Optimize for large-scale vector operations
- **Indexing Algorithms**: Implement HNSW, IVF, and other high-performance vector indexes
- **Concurrent Operations**: Go's goroutines for parallel processing
- **Zero-Copy**: Minimize memory allocations in hot paths

## 📁 Repository Structure

```
vjvector/
├── cmd/vjvector/          # Main application entry point
├── pkg/                   # Public packages (importable by other projects)
│   ├── core/             # Core vector types, interfaces, and algorithms
│   ├── embedding/        # Embedding service implementations (OpenAI, local models)
│   ├── storage/          # Storage layer (file-based, database backends)
│   ├── index/            # Vector indexing algorithms (HNSW, IVF, etc.)
│   ├── query/            # Query processing and optimization
│   ├── api/              # API utilities and helpers
│   ├── config/           # Configuration management
│   └── utils/            # Utility functions and helpers
├── internal/              # Internal packages (not importable externally)
│   ├── server/           # HTTP server implementation
│   ├── handlers/         # HTTP request handlers
│   └── middleware/       # HTTP middleware (auth, logging, etc.)
├── docs/                 # Documentation and API specs
├── examples/             # Example applications and usage
├── scripts/              # Build, deployment, and utility scripts
└── .github/workflows/    # CI/CD pipelines
```

## 🔧 Development Guidelines

### 1. Go Best Practices
- **Go Modules**: Use Go 1.21+ modules with proper dependency management
- **Error Handling**: Use `fmt.Errorf` with `%w` for error wrapping
- **Context**: Use `context.Context` for cancellation and timeouts
- **Interfaces**: Define small, focused interfaces in the package that uses them
- **Testing**: Write comprehensive tests with table-driven tests for edge cases

### 2. Code Organization
- **Package Naming**: Use descriptive, single-word package names
- **File Organization**: Group related functionality in the same file
- **Exports**: Only export what's necessary for external consumption
- **Documentation**: Use Go doc comments for all exported types and functions

### 3. Performance Considerations
- **Memory Allocation**: Use object pools for frequently allocated objects
- **Goroutines**: Use worker pools for concurrent operations
- **Profiling**: Include benchmarks for critical paths
- **Metrics**: Add Prometheus metrics for monitoring

## 🧠 AI-Specific Implementation Guidelines

### 1. Vector Operations
```go
// Always validate vector dimensions
if len(queryVector) != expectedDimension {
    return fmt.Errorf("vector dimension mismatch: expected %d, got %d", expectedDimension, len(queryVector))
}

// Use efficient similarity calculations
func cosineSimilarity(a, b []float64) float64 {
    dotProduct := 0.0
    for i := range a {
        dotProduct += a[i] * b[i]
    }
    return dotProduct // Assuming normalized vectors
}
```

### 2. Embedding Services
```go
type EmbeddingService interface {
    EmbedText(text string) ([]float64, error)
    EmbedBatch(texts []string) ([][]float64, error)
    GetDimension() int
    GetModelName() string
}

// Implement rate limiting and retries for external APIs
// Cache embeddings when possible
// Support batch processing for efficiency
```

### 3. Indexing Strategies
```go
type VectorIndex interface {
    Insert(vector *core.Vector) error
    Search(query []float64, k int) ([]core.VectorSearchResult, error)
    Delete(id string) error
    Optimize() error
}

// HNSW (Hierarchical Navigable Small World) for approximate nearest neighbor
// IVF (Inverted File) for large-scale clustering
// Support for multiple index types per collection
```

## 🧪 Testing Strategy

### 1. Unit Tests
- **Core Algorithms**: Test vector operations with known mathematical results
- **Edge Cases**: Test with zero vectors, very large vectors, dimension mismatches
- **Mocking**: Use interfaces for testable code

### 2. Integration Tests
- **End-to-End**: Test complete workflows from embedding to search
- **Performance**: Benchmark critical operations
- **Concurrency**: Test with multiple goroutines

### 3. Test Data
```go
// Use realistic vector dimensions (1536 for OpenAI, 768 for BERT, etc.)
// Test with actual embedding outputs from popular models
// Include metadata filtering scenarios
```

## 🚀 Implementation Priorities

### Phase 1: Core Foundation ✅
- [x] Basic project structure
- [x] Core vector types and interfaces
- [x] Configuration management
- [x] HTTP server framework
- [x] CI/CD pipeline

### Phase 2: Storage & Indexing 🚧
- [ ] File-based storage implementation
- [ ] HNSW index implementation
- [ ] Basic vector CRUD operations
- [ ] Collection management

### Phase 3: AI Integration 🚧
- [ ] OpenAI embedding service
- [ ] Local embedding models (sentence-transformers)
- [ ] Batch embedding processing
- [ ] Embedding caching

### Phase 4: Advanced Features 📋
- [ ] Query optimization
- [ ] Advanced filtering and metadata search
- [ ] Performance monitoring
- [ ] Backup and recovery

## 🔍 Key Implementation Areas

### 1. Vector Indexing (HNSW)
```go
// Implement HNSW algorithm for approximate nearest neighbor search
// Focus on memory efficiency and search speed
// Support for configurable parameters (M, efConstruction, efSearch)
```

### 2. Storage Layer
```go
// File-based storage with memory mapping for large datasets
// Support for different storage backends (local files, S3, etc.)
// Efficient serialization of vectors and metadata
```

### 3. Query Processing
```go
// Vector similarity search with configurable metrics
// Metadata filtering and aggregation
// Support for complex queries (AND/OR operations)
```

## 📚 Useful Prompts for Cursor

### When Implementing New Features
```
"Implement [feature] following the AI-first design principles of this vector database. 
Consider performance implications and include comprehensive tests. 
Follow the established patterns in the codebase."
```

### When Debugging Issues
```
"Analyze this code for potential performance issues, especially around vector operations. 
Look for memory leaks, inefficient algorithms, or concurrency problems."
```

### When Adding Tests
```
"Write comprehensive tests for this [component] including edge cases, 
performance benchmarks, and integration scenarios. 
Use table-driven tests and proper mocking."
```

### When Optimizing Performance
```
"Optimize this [operation] for high-performance vector operations. 
Consider memory usage, CPU efficiency, and concurrent access patterns. 
Include benchmarks to measure improvements."
```

## 🎯 Success Metrics

- **Performance**: Sub-millisecond search times for 1M+ vectors
- **Scalability**: Support for 100M+ vectors with efficient memory usage
- **Reliability**: 99.9% uptime with proper error handling
- **Usability**: Simple API that AI developers can integrate quickly
- **Extensibility**: Easy to add new embedding models and index types

## 🚨 Common Pitfalls to Avoid

1. **Memory Leaks**: Don't hold references to large vectors unnecessarily
2. **Blocking Operations**: Use goroutines for I/O operations
3. **Hardcoded Dimensions**: Make vector dimensions configurable
4. **Inefficient Search**: Implement proper indexing before brute-force search
5. **Poor Error Handling**: Provide meaningful error messages for debugging

## 🔗 Related Technologies

- **Vector Similarity**: Cosine similarity, Euclidean distance, dot product
- **Indexing Algorithms**: HNSW, IVF, LSH, KD-trees
- **Embedding Models**: OpenAI, BERT, sentence-transformers
- **Storage**: LevelDB, RocksDB, S3, local filesystems
- **Monitoring**: Prometheus, Grafana, OpenTelemetry

---

**Remember**: This is an AI-first vector database. Every design decision should prioritize the needs of AI applications and RAG workflows. Performance, scalability, and ease of integration are paramount.
