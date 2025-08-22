// Package main demonstrates the Week 9-10 Performance Optimization implementation
package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	vectormath "github.com/vijaynallagatla/vjvector/pkg/math"
	"github.com/vijaynallagatla/vjvector/pkg/parallel"
)

func main() {
	fmt.Println("🚀 VJVector Performance Optimization Demo: SIMD & Batch Operations")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Test vector sizes
	sizes := []int{128, 512, 1536, 4096}
	iterations := 10000

	for _, size := range sizes {
		fmt.Printf("📊 Testing with %d-dimensional vectors (%d iterations)\n", size, iterations)
		demoSIMDPerformance(size, iterations)
		fmt.Println()
	}

	// Demo batch operations
	demoBatchOperations()

	// Demo parallel processing
	demoParallelProcessing()

	fmt.Println("✅ Performance Optimization Demo Complete!")
}

func demoSIMDPerformance(dimension, iterations int) {
	// Create test vectors
	a := make([]float64, dimension)
	b := make([]float64, dimension)

	for i := 0; i < dimension; i++ {
		a[i] = float64(i) * 0.1
		b[i] = float64(i) * 0.2
	}

	// Test scalar vs SIMD performance
	scalarMath := vectormath.NewVectorMath(false)
	simdMath := vectormath.NewVectorMath(true)

	// Dot Product Performance
	fmt.Printf("   🔢 Dot Product Performance:\n")

	// Scalar benchmark
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_ = scalarMath.DotProduct(a, b)
	}
	scalarTime := time.Since(start)

	// SIMD benchmark
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_ = simdMath.DotProduct(a, b)
	}
	simdTime := time.Since(start)

	speedup := float64(scalarTime) / float64(simdTime)
	fmt.Printf("      📈 Scalar: %v (%s per op)\n", scalarTime, time.Duration(int64(scalarTime)/int64(iterations)))
	fmt.Printf("      ⚡ SIMD:   %v (%s per op)\n", simdTime, time.Duration(int64(simdTime)/int64(iterations)))
	fmt.Printf("      🚀 Speedup: %.2fx\n", speedup)

	// Cosine Similarity Performance
	fmt.Printf("   📐 Cosine Similarity Performance:\n")

	// Scalar benchmark
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_ = scalarMath.CosineSimilarity(a, b)
	}
	scalarTime = time.Since(start)

	// SIMD benchmark
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_ = simdMath.CosineSimilarity(a, b)
	}
	simdTime = time.Since(start)

	speedup = float64(scalarTime) / float64(simdTime)
	fmt.Printf("      📈 Scalar: %v (%s per op)\n", scalarTime, time.Duration(int64(scalarTime)/int64(iterations)))
	fmt.Printf("      ⚡ SIMD:   %v (%s per op)\n", simdTime, time.Duration(int64(simdTime)/int64(iterations)))
	fmt.Printf("      🚀 Speedup: %.2fx\n", speedup)

	// Euclidean Distance Performance
	fmt.Printf("   📏 Euclidean Distance Performance:\n")

	// Scalar benchmark
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_ = scalarMath.EuclideanDistance(a, b)
	}
	scalarTime = time.Since(start)

	// SIMD benchmark
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_ = simdMath.EuclideanDistance(a, b)
	}
	simdTime = time.Since(start)

	speedup = float64(scalarTime) / float64(simdTime)
	fmt.Printf("      📈 Scalar: %v (%s per op)\n", scalarTime, time.Duration(int64(scalarTime)/int64(iterations)))
	fmt.Printf("      ⚡ SIMD:   %v (%s per op)\n", simdTime, time.Duration(int64(simdTime)/int64(iterations)))
	fmt.Printf("      🚀 Speedup: %.2fx\n", speedup)
}

func demoBatchOperations() {
	fmt.Println("📊 Batch Operations Performance Demo")
	fmt.Println(strings.Repeat("-", 50))

	// Create test data
	batchSize := 1000
	dimension := 1536

	vectorsA := make([][]float64, batchSize)
	vectorsB := make([][]float64, batchSize)

	for i := 0; i < batchSize; i++ {
		vectorsA[i] = make([]float64, dimension)
		vectorsB[i] = make([]float64, dimension)

		for j := 0; j < dimension; j++ {
			vectorsA[i][j] = float64(i*j) * 0.001
			vectorsB[i][j] = float64(i+j) * 0.001
		}
	}

	// Test batch operations
	batchOps := vectormath.NewBatchOperations(true)

	fmt.Printf("   🔢 Batch Dot Product (%d vectors):\n", batchSize)
	start := time.Now()
	results := batchOps.BatchDotProduct(vectorsA, vectorsB)
	elapsed := time.Since(start)

	fmt.Printf("      ✅ Processed %d dot products in %v\n", len(results), elapsed)
	fmt.Printf("      📈 Throughput: %.2f ops/ms\n", float64(batchSize)/float64(elapsed.Milliseconds()))

	fmt.Printf("   🔄 Batch Normalization (%d vectors):\n", batchSize)
	start = time.Now()
	normalized := batchOps.BatchNormalize(vectorsA)
	elapsed = time.Since(start)

	fmt.Printf("      ✅ Normalized %d vectors in %v\n", len(normalized), elapsed)
	fmt.Printf("      📈 Throughput: %.2f ops/ms\n", float64(batchSize)/float64(elapsed.Milliseconds()))
}

func demoParallelProcessing() {
	fmt.Println("📊 Parallel Processing Performance Demo")
	fmt.Println(strings.Repeat("-", 50))

	numCPUs := runtime.NumCPU()
	fmt.Printf("   💻 Available CPU cores: %d\n", numCPUs)

	// Create test data
	batchSize := 10000
	dimension := 1536

	vectorsA := make([][]float64, batchSize)
	vectorsB := make([][]float64, batchSize)

	for i := 0; i < batchSize; i++ {
		vectorsA[i] = make([]float64, dimension)
		vectorsB[i] = make([]float64, dimension)

		for j := 0; j < dimension; j++ {
			vectorsA[i][j] = float64(i*j) * 0.001
			vectorsB[i][j] = float64(i+j) * 0.001
		}
	}

	// Test sequential vs parallel performance
	batchOps := vectormath.NewBatchOperations(true)
	parallelProcessor := parallel.NewParallelBatchProcessor(numCPUs, true)

	fmt.Printf("   🔢 Dot Product Performance (%d operations):\n", batchSize)

	// Sequential benchmark
	start := time.Now()
	_ = batchOps.BatchDotProduct(vectorsA, vectorsB)
	sequentialTime := time.Since(start)

	// Parallel benchmark
	start = time.Now()
	_ = parallelProcessor.ProcessDotProductBatch(vectorsA, vectorsB)
	parallelTime := time.Since(start)

	speedup := float64(sequentialTime) / float64(parallelTime)
	fmt.Printf("      📈 Sequential: %v (%.2f ops/ms)\n", sequentialTime, float64(batchSize)/float64(sequentialTime.Milliseconds()))
	fmt.Printf("      ⚡ Parallel:   %v (%.2f ops/ms)\n", parallelTime, float64(batchSize)/float64(parallelTime.Milliseconds()))
	fmt.Printf("      🚀 Speedup: %.2fx\n", speedup)

	fmt.Printf("   🔄 Normalization Performance (%d operations):\n", batchSize)

	// Sequential benchmark
	start = time.Now()
	_ = batchOps.BatchNormalize(vectorsA)
	sequentialTime = time.Since(start)

	// Parallel benchmark
	start = time.Now()
	_ = parallelProcessor.ProcessNormalizeBatch(vectorsA)
	parallelTime = time.Since(start)

	speedup = float64(sequentialTime) / float64(parallelTime)
	fmt.Printf("      📈 Sequential: %v (%.2f ops/ms)\n", sequentialTime, float64(batchSize)/float64(sequentialTime.Milliseconds()))
	fmt.Printf("      ⚡ Parallel:   %v (%.2f ops/ms)\n", parallelTime, float64(batchSize)/float64(parallelTime.Milliseconds()))
	fmt.Printf("      🚀 Speedup: %.2fx\n", speedup)

	// Test parallel search
	query := make([]float64, dimension)
	for i := 0; i < dimension; i++ {
		query[i] = float64(i) * 0.001
	}

	fmt.Printf("   🔍 Vector Search Performance (query against %d vectors, k=10):\n", batchSize)

	start = time.Now()
	results := parallelProcessor.ProcessSearchBatch(query, vectorsA, 10)
	elapsed := time.Since(start)

	fmt.Printf("      ✅ Found %d results in %v\n", len(results), elapsed)
	fmt.Printf("      📈 Search throughput: %.2f vectors/ms\n", float64(batchSize)/float64(elapsed.Milliseconds()))

	if len(results) > 0 {
		fmt.Printf("      🎯 Top result similarity: %.4f\n", results[0].Similarity)
	}
}
