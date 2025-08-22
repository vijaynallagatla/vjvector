# VJVector - AI-First Vector Database

[![CI](https://github.com/vijaynallagatla/vjvector/actions/workflows/ci.yml/badge.svg)](https://github.com/vijaynallagatla/vjvector/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vijaynallagatla/vjvector)](https://goreportcard.com/report/github.com/vijaynallagatla/vjvector)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

VJVector is a high-performance vector database built from scratch with native AI embedding support, designed specifically for RAG (Retrieval-Augmented Generation) applications. Built in Go for performance and reliability.

## 🚀 Features

- **AI-First Design**: Built with AI embeddings in mind from the ground up
- **High Performance**: Optimized for fast vector similarity search
- **Multiple Index Types**: Support for HNSW, IVF, and other vector indexing algorithms
- **RESTful API**: Simple HTTP interface for easy integration
- **Collection Management**: Organize vectors into logical collections
- **Metadata Support**: Rich metadata storage and filtering
- **Scalable Architecture**: Modular design for easy extension
- **Production Ready**: Comprehensive testing, CI/CD, and security scanning

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Server   │    │  Vector Index   │    │  Storage Layer  │
│   (REST API)    │◄──►│   (HNSW/IVF)    │◄──►│   (File/DB)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Embedding      │    │   Core Vector   │    │   Collection    │
│   Service       │    │    Types        │    │   Management    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📦 Installation

### Prerequisites

- Go 1.25 or later
- Git

### From Source

```bash
git clone https://github.com/vijaynallagatla/vjvector.git
cd vjvector
go mod download
go build -o bin/vjvector ./cmd/vjvector
```

### Using Docker

```bash
docker build -t vjvector .
docker run -p 8080:8080 vjvector
```

## 🚀 Quick Start

### 1. Start the Server

```bash
./bin/vjvector serve
```

The server will start on `http://localhost:8080`

### 2. Create a Collection

```bash
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{
    "name": "documents",
    "description": "Document embeddings",
    "dimension": 1536,
    "index_type": "hnsw"
  }'
```

### 3. Add Vectors

```bash
curl -X POST http://localhost:8080/vectors \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "documents",
    "embedding": [0.1, 0.2, 0.3, ...],
    "text": "Sample document text",
    "metadata": {"source": "web", "category": "tech"}
  }'
```

### 4. Search Similar Vectors

```bash
curl -X POST http://localhost:8080/vectors/search \
  -H "Content-Type: application/json" \
  -d '{
    "query_vector": [0.1, 0.2, 0.3, ...],
    "collection": "documents",
    "limit": 10,
    "threshold": 0.8
  }'
```

## 🔧 Configuration

Create a `config.yaml` file:

```yaml
server:
  host: "localhost"
  port: 8080

database:
  data_dir: "./data"
  max_vectors: 1000000
  index_type: "hnsw"

embedding:
  model_name: "text-embedding-ada-002"
  dimension: 1536
  batch_size: 100

logging:
  level: "info"
  format: "json"
```

## 📚 API Reference

### Collections

- `POST /collections` - Create a collection
- `GET /collections` - List all collections
- `GET /collections/{name}` - Get collection details
- `DELETE /collections/{name}` - Delete a collection

### Vectors

- `POST /vectors` - Create a vector
- `GET /vectors/{id}` - Get a vector
- `PUT /vectors/{id}` - Update a vector
- `DELETE /vectors/{id}` - Delete a vector
- `POST /vectors/search` - Search similar vectors

### Embeddings

- `POST /embed` - Generate text embeddings

### Health

- `GET /health` - Health check endpoint

## 🧪 Development

### Running Tests

```bash
go test -v ./...
```

### Running with Coverage

```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting

```bash
golangci-lint run
```

### Building

```bash
go build -o bin/vjvector ./cmd/vjvector
```

## 🏗️ Project Structure

```
vjvector/
├── cmd/vjvector/          # Main application entry point
├── pkg/                   # Public packages
│   ├── core/             # Core vector types and interfaces
│   ├── embedding/        # Embedding service implementations
│   ├── storage/          # Storage layer implementations
│   ├── index/            # Vector indexing algorithms
│   ├── query/            # Query processing
│   ├── api/              # API utilities
│   ├── config/           # Configuration management
│   └── utils/            # Utility functions
├── internal/              # Internal packages
│   ├── server/           # HTTP server implementation
│   ├── handlers/         # HTTP request handlers
│   └── middleware/       # HTTP middleware
├── docs/                 # Documentation
├── examples/             # Example applications
├── scripts/              # Build and deployment scripts
├── .github/workflows/    # GitHub Actions CI/CD
├── Dockerfile            # Docker container definition
├── go.mod                # Go module definition
└── README.md             # This file
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with Go for performance and reliability
- Inspired by modern vector database architectures
- Designed for AI-first applications and RAG systems

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/vijaynallagatla/vjvector/issues)
- **Discussions**: [GitHub Discussions](https://github.com/vijaynallagatla/vjvector/discussions)
- **Wiki**: [Project Wiki](https://github.com/vijaynallagatla/vjvector/wiki)

## 🗺️ Roadmap

- [ ] HNSW index implementation
- [ ] IVF index implementation
- [ ] OpenAI embedding integration
- [ ] Sentence-transformers integration
- [ ] GraphQL API
- [ ] gRPC support
- [ ] Kubernetes deployment
- [ ] Monitoring and metrics
- [ ] Backup and recovery
- [ ] Multi-tenant support

---

**VJVector** - Empowering AI applications with high-performance vector search.
