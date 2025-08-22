#!/bin/bash

# VJVector RAG Optimization Demo Script
# Week 19-20: Query Expansion, Result Reranking, and Context-Aware Retrieval

set -e

echo "🚀 VJVector RAG Optimization Demo - Week 19-20"
echo "================================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f "pkg/rag/query_expansion.go" ]; then
    print_error "Please run this script from the VJVector project root directory"
    exit 1
fi

print_status "Building VJVector CLI..."
go build -o bin/vjvector ./cmd/cli

if [ $? -eq 0 ]; then
    print_success "VJVector built successfully"
else
    print_error "Failed to build VJVector"
    exit 1
fi

echo ""
print_status "Running RAG Optimization Tests..."
echo ""

# Run complete RAG test suite
print_status "Running Complete RAG Test Suite..."
go test ./pkg/rag/... -v

echo ""
print_status "Creating RAG Optimization Demo Program..."

# Create a demo Go program
cat > /tmp/rag_optimization_demo.go << 'EOF'
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vijaynallagatla/vjvector/pkg/core"
	"github.com/vijaynallagatla/vjvector/pkg/rag"
)

func main() {
	fmt.Println("🚀 VJVector RAG Optimization Demo")
	fmt.Println("==================================")
	fmt.Println("")

	// Demo 1: Query Expansion
	demoQueryExpansion()

	// Demo 2: Result Reranking
	demoResultReranking()

	// Demo 3: Context-Aware Retrieval
	demoContextAwareRetrieval()

	fmt.Println("\n✅ All RAG optimization demos completed successfully!")
}

func demoQueryExpansion() {
	fmt.Println("📚 Demo 1: Query Expansion")
	fmt.Println("----------------------------")

	// Create query expansion manager
	expansionManager := rag.NewQueryExpansionManager(nil)

	// Test queries
	queries := []*rag.Query{
		{
			Text: "how to build a fast car",
			Context: map[string]interface{}{
				"domain": "automotive",
			},
		},
		{
			Text: "what is the best programming language",
			Context: map[string]interface{}{
				"domain": "technical",
			},
		},
		{
			Text: "compare Python vs Go",
			Context: map[string]interface{}{
				"domain": "programming",
			},
		},
	}

	for i, query := range queries {
		fmt.Printf("\nQuery %d: %s\n", i+1, query.Text)
		
		expansions, err := expansionManager.ExpandQuery(context.Background(), query, nil)
		if err != nil {
			log.Printf("Error expanding query: %v", err)
			continue
		}

		fmt.Printf("Expansions: %v\n", expansions)
	}
}

func demoResultReranking() {
	fmt.Println("\n🔄 Demo 2: Result Reranking")
	fmt.Println("----------------------------")

	// Create reranking manager
	rerankingManager := rag.NewResultRerankingManager(nil)

	// Create mock results
	results := []*rag.QueryResult{
		{
			Vector: &core.Vector{ID: "vec1"},
			Score:  0.8,
			Context: map[string]interface{}{
				"domain": "technical",
				"type":   "tutorial",
			},
		},
		{
			Vector: &core.Vector{ID: "vec2"},
			Score:  0.6,
			Context: map[string]interface{}{
				"domain": "general",
				"type":   "article",
			},
		},
		{
			Vector: &core.Vector{ID: "vec3"},
			Score:  0.9,
			Context: map[string]interface{}{
				"domain": "technical",
				"type":   "documentation",
			},
		},
	}

	query := &rag.Query{
		Text: "programming tutorial",
		Context: map[string]interface{}{
			"domain": "technical",
			"user_id": "user123",
		},
	}

	fmt.Printf("Original Results: %d\n", len(results))
	for i, result := range results {
		fmt.Printf("  Result %d: Score=%.2f, Domain=%s\n", 
			i+1, result.Score, result.Context["domain"])
	}

	// Rerank results
	reranked, err := rerankingManager.RerankResults(context.Background(), results, query, nil)
	if err != nil {
		log.Printf("Error reranking results: %v", err)
		return
	}

	fmt.Printf("\nReranked Results: %d\n", len(reranked))
	for i, result := range reranked {
		fmt.Printf("  Result %d: Score=%.2f, Domain=%s\n", 
			i+1, result.Score, result.Context["domain"])
	}
}

func demoContextAwareRetrieval() {
	fmt.Println("\n🎯 Demo 3: Context-Aware Retrieval")
	fmt.Println("-----------------------------------")

	// Create context-aware retrieval manager
	contextManager := rag.NewContextAwareRetrievalManager(nil)

	// Test queries with different contexts
	queries := []*rag.Query{
		{
			Text: "restaurant recommendations",
			Context: map[string]interface{}{
				"location": "San Francisco",
				"time_context": "dinner",
				"user_id": "user123",
			},
		},
		{
			Text: "health checkup",
			Context: map[string]interface{}{
				"domain": "medical",
				"location": "New York",
				"time_context": "morning",
			},
		},
		{
			Text: "programming tutorial",
			Context: map[string]interface{}{
				"domain": "technical",
				"user_id": "user456",
				"time_context": "afternoon",
			},
		},
	}

	for i, query := range queries {
		fmt.Printf("\nQuery %d: %s\n", i+1, query.Text)
		fmt.Printf("Context: %+v\n", query.Context)
		
		enhanced, err := contextManager.ProcessContextAwareQuery(context.Background(), query, nil)
		if err != nil {
			log.Printf("Error processing context-aware query: %v", err)
			continue
		}

		fmt.Printf("Enhancements: %v\n", enhanced.Enhancements)
		fmt.Printf("Detected Domain: %v\n", enhanced.Context["detected_domain"])
		fmt.Printf("User Location: %v\n", enhanced.Context["user_location"])
		fmt.Printf("Confidence: %.2f\n", enhanced.Confidence)
	}
}

EOF

print_status "Compiling demo program..."
go build -o bin/rag_optimization_demo /tmp/rag_optimization_demo.go

if [ $? -eq 0 ]; then
    print_success "Demo program compiled successfully"
else
    print_error "Failed to compile demo program"
    exit 1
fi

print_status "Running demo program..."
./bin/rag_optimization_demo

if [ $? -eq 0 ]; then
    print_success "Demo program ran successfully"
else
    print_error "Demo program failed to run"
    exit 1
fi

print_status "Cleaning up temporary files..."
rm /tmp/rag_optimization_demo.go

echo ""
print_success "VJVector RAG Optimization Demo - Week 19-20 completed successfully!"
