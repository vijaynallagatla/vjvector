// Package main demonstrates basic usage of the VJVector database core functionality.
// This example shows how to create vectors, collections, and perform similarity operations.
package main

import (
	"fmt"
	"log"

	"github.com/vijaynallagatla/vjvector/pkg/core"
)

func main() {
	fmt.Println("VJVector Basic Usage Example")
	fmt.Println("============================")

	// Create a collection
	collection := core.NewCollection("documents", "Sample document collection", 1536, "hnsw")
	fmt.Printf("Created collection: %s\n", collection.Name)

	// Create sample vectors (simulating OpenAI embeddings)
	vectors := []*core.Vector{
		core.NewVector("documents", []float64{0.1, 0.2, 0.3, 0.4, 0.5}, "Document about AI", map[string]interface{}{
			"source": "web",
			"topic":  "artificial intelligence",
		}),
		core.NewVector("documents", []float64{0.2, 0.3, 0.4, 0.5, 0.6}, "Document about machine learning", map[string]interface{}{
			"source": "web",
			"topic":  "machine learning",
		}),
		core.NewVector("documents", []float64{0.3, 0.4, 0.5, 0.6, 0.7}, "Document about neural networks", map[string]interface{}{
			"source": "web",
			"topic":  "neural networks",
		}),
	}

	// Display vector information
	for i, vector := range vectors {
		fmt.Printf("\nVector %d:\n", i+1)
		fmt.Printf("  ID: %s\n", vector.ID)
		fmt.Printf("  Text: %s\n", vector.Text)
		fmt.Printf("  Dimension: %d\n", vector.Dimension)
		fmt.Printf("  Magnitude: %.4f\n", vector.Magnitude)
		fmt.Printf("  Metadata: %v\n", vector.Metadata)
	}

	// Demonstrate vector operations
	fmt.Println("\nVector Operations:")
	fmt.Println("==================")

	// Calculate similarity between first two vectors
	similarity, err := vectors[0].Similarity(vectors[1])
	if err != nil {
		log.Printf("Error calculating similarity: %v", err)
	} else {
		fmt.Printf("Similarity between vector 1 and 2: %.4f\n", similarity)
	}

	// Calculate distance between first two vectors
	distance, err := vectors[0].Distance(vectors[1])
	if err != nil {
		log.Printf("Error calculating distance: %v", err)
	} else {
		fmt.Printf("Distance between vector 1 and 2: %.4f\n", distance)
	}

	// Normalize a vector
	fmt.Println("\nNormalizing vector 1...")
	originalMagnitude := vectors[0].Magnitude
	vectors[0].Normalize()
	fmt.Printf("Original magnitude: %.4f, Normalized magnitude: %.4f\n",
		originalMagnitude, vectors[0].Magnitude)

	// Demonstrate search query structure
	searchQuery := &core.SearchQuery{
		QueryVector: []float64{0.15, 0.25, 0.35, 0.45, 0.55},
		Collection:  "documents",
		Limit:       5,
		Threshold:   0.8,
		Metadata: map[string]interface{}{
			"topic": "AI",
		},
	}

	fmt.Printf("\nSearch Query:\n")
	fmt.Printf("  Collection: %s\n", searchQuery.Collection)
	fmt.Printf("  Limit: %d\n", searchQuery.Limit)
	fmt.Printf("  Threshold: %.2f\n", searchQuery.Threshold)
	fmt.Printf("  Metadata filter: %v\n", searchQuery.Metadata)

	fmt.Println("\nExample completed successfully!")
	fmt.Println("This demonstrates the basic vector operations available in VJVector.")
	fmt.Println("In a real application, you would:")
	fmt.Println("  1. Connect to the VJVector server")
	fmt.Println("  2. Use the REST API to manage collections and vectors")
	fmt.Println("  3. Implement proper error handling and logging")
	fmt.Println("  4. Add authentication and security measures")
}
