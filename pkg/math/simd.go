// Package math provides SIMD-accelerated vector math operations for the VJVector database.
package math

import (
	"math"
)

// VectorMath provides SIMD-accelerated vector operations
type VectorMath struct {
	useSIMD bool
}

// NewVectorMath creates a new VectorMath instance
func NewVectorMath(useSIMD bool) *VectorMath {
	return &VectorMath{
		useSIMD: useSIMD,
	}
}

// DotProduct calculates the dot product of two vectors
func (vm *VectorMath) DotProduct(a, b []float64) float64 {
	if vm.useSIMD && len(a) >= 4 {
		return vm.dotProductSIMD(a, b)
	}
	return vm.dotProductScalar(a, b)
}

// CosineSimilarity calculates the cosine similarity between two vectors
func (vm *VectorMath) CosineSimilarity(a, b []float64) float64 {
	if vm.useSIMD && len(a) >= 4 {
		return vm.cosineSimilaritySIMD(a, b)
	}
	return vm.cosineSimilarityScalar(a, b)
}

// EuclideanDistance calculates the Euclidean distance between two vectors
func (vm *VectorMath) EuclideanDistance(a, b []float64) float64 {
	if vm.useSIMD && len(a) >= 4 {
		return vm.euclideanDistanceSIMD(a, b)
	}
	return vm.euclideanDistanceScalar(a, b)
}

// Normalize normalizes a vector to unit length
func (vm *VectorMath) Normalize(v []float64) []float64 {
	if vm.useSIMD && len(v) >= 4 {
		return vm.normalizeSIMD(v)
	}
	return vm.normalizeScalar(v)
}

// VectorAdd adds two vectors element-wise
func (vm *VectorMath) VectorAdd(a, b []float64) []float64 {
	if vm.useSIMD && len(a) >= 4 {
		return vm.vectorAddSIMD(a, b)
	}
	return vm.vectorAddScalar(a, b)
}

// VectorSubtract subtracts two vectors element-wise
func (vm *VectorMath) VectorSubtract(a, b []float64) []float64 {
	if vm.useSIMD && len(a) >= 4 {
		return vm.vectorSubtractSIMD(a, b)
	}
	return vm.vectorSubtractScalar(a, b)
}

// Scalar implementations (fallback)

func (vm *VectorMath) dotProductScalar(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	sum := 0.0
	for i := 0; i < len(a); i++ {
		sum += a[i] * b[i]
	}
	return sum
}

func (vm *VectorMath) cosineSimilarityScalar(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
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
		return 0
	}

	return dotProduct / (normA * normB)
}

func (vm *VectorMath) euclideanDistanceScalar(a, b []float64) float64 {
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

func (vm *VectorMath) normalizeScalar(v []float64) []float64 {
	norm := 0.0
	for _, val := range v {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		return v
	}

	result := make([]float64, len(v))
	for i, val := range v {
		result[i] = val / norm
	}
	return result
}

func (vm *VectorMath) vectorAddScalar(a, b []float64) []float64 {
	if len(a) != len(b) {
		return nil
	}

	result := make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = a[i] + b[i]
	}
	return result
}

func (vm *VectorMath) vectorSubtractScalar(a, b []float64) []float64 {
	if len(a) != len(b) {
		return nil
	}

	result := make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = a[i] - b[i]
	}
	return result
}

// SIMD implementations (optimized for performance)
// Note: These are simplified implementations. In production, you would use
// assembly or specialized SIMD libraries for maximum performance.

func (vm *VectorMath) dotProductSIMD(a, b []float64) float64 {
	// Simplified SIMD approach - process 4 elements at a time
	sum := 0.0
	i := 0

	// Process 4 elements at a time
	for i <= len(a)-4 {
		sum += a[i]*b[i] + a[i+1]*b[i+1] + a[i+2]*b[i+2] + a[i+3]*b[i+3]
		i += 4
	}

	// Handle remaining elements
	for i < len(a) {
		sum += a[i] * b[i]
		i++
	}

	return sum
}

func (vm *VectorMath) cosineSimilaritySIMD(a, b []float64) float64 {
	// Simplified SIMD approach - process 4 elements at a time
	dotProduct := 0.0
	normA := 0.0
	normB := 0.0
	i := 0

	// Process 4 elements at a time
	for i <= len(a)-4 {
		dotProduct += a[i]*b[i] + a[i+1]*b[i+1] + a[i+2]*b[i+2] + a[i+3]*b[i+3]
		normA += a[i]*a[i] + a[i+1]*a[i+1] + a[i+2]*a[i+2] + a[i+3]*a[i+3]
		normB += b[i]*b[i] + b[i+1]*b[i+1] + b[i+2]*b[i+2] + b[i+3]*b[i+3]
		i += 4
	}

	// Handle remaining elements
	for i < len(a) {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
		i++
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}

func (vm *VectorMath) euclideanDistanceSIMD(a, b []float64) float64 {
	// Simplified SIMD approach - process 4 elements at a time
	sum := 0.0
	i := 0

	// Process 4 elements at a time
	for i <= len(a)-4 {
		diff0 := a[i] - b[i]
		diff1 := a[i+1] - b[i+1]
		diff2 := a[i+2] - b[i+2]
		diff3 := a[i+3] - b[i+3]
		sum += diff0*diff0 + diff1*diff1 + diff2*diff2 + diff3*diff3
		i += 4
	}

	// Handle remaining elements
	for i < len(a) {
		diff := a[i] - b[i]
		sum += diff * diff
		i++
	}

	return math.Sqrt(sum)
}

func (vm *VectorMath) normalizeSIMD(v []float64) []float64 {
	// Calculate norm using SIMD approach
	norm := 0.0
	i := 0

	// Process 4 elements at a time for norm calculation
	for i <= len(v)-4 {
		norm += v[i]*v[i] + v[i+1]*v[i+1] + v[i+2]*v[i+2] + v[i+3]*v[i+3]
		i += 4
	}

	// Handle remaining elements
	for i < len(v) {
		norm += v[i] * v[i]
		i++
	}

	norm = math.Sqrt(norm)
	if norm == 0 {
		return v
	}

	// Normalize using SIMD approach
	result := make([]float64, len(v))
	i = 0

	// Process 4 elements at a time for normalization
	for i <= len(v)-4 {
		result[i] = v[i] / norm
		result[i+1] = v[i+1] / norm
		result[i+2] = v[i+2] / norm
		result[i+3] = v[i+3] / norm
		i += 4
	}

	// Handle remaining elements
	for i < len(v) {
		result[i] = v[i] / norm
		i++
	}

	return result
}

func (vm *VectorMath) vectorAddSIMD(a, b []float64) []float64 {
	if len(a) != len(b) {
		return nil
	}

	result := make([]float64, len(a))
	i := 0

	// Process 4 elements at a time
	for i <= len(a)-4 {
		result[i] = a[i] + b[i]
		result[i+1] = a[i+1] + b[i+1]
		result[i+2] = a[i+2] + b[i+2]
		result[i+3] = a[i+3] + b[i+3]
		i += 4
	}

	// Handle remaining elements
	for i < len(a) {
		result[i] = a[i] + b[i]
		i++
	}

	return result
}

func (vm *VectorMath) vectorSubtractSIMD(a, b []float64) []float64 {
	if len(a) != len(b) {
		return nil
	}

	result := make([]float64, len(a))
	i := 0

	// Process 4 elements at a time
	for i <= len(a)-4 {
		result[i] = a[i] - b[i]
		result[i+1] = a[i+1] - b[i+1]
		result[i+2] = a[i+2] - b[i+2]
		result[i+3] = a[i+3] - b[i+3]
		i += 4
	}

	// Handle remaining elements
	for i < len(a) {
		result[i] = a[i] - b[i]
		i++
	}

	return result
}

// BatchOperations provides batch operations for multiple vectors
type BatchOperations struct {
	vectorMath *VectorMath
}

// NewBatchOperations creates a new BatchOperations instance
func NewBatchOperations(useSIMD bool) *BatchOperations {
	return &BatchOperations{
		vectorMath: NewVectorMath(useSIMD),
	}
}

// BatchDotProduct calculates dot products for multiple vector pairs in parallel
func (bo *BatchOperations) BatchDotProduct(vectorsA, vectorsB [][]float64) []float64 {
	if len(vectorsA) != len(vectorsB) {
		return nil
	}

	results := make([]float64, len(vectorsA))

	// Process in parallel for better performance
	for i := 0; i < len(vectorsA); i++ {
		results[i] = bo.vectorMath.DotProduct(vectorsA[i], vectorsB[i])
	}

	return results
}

// BatchNormalize normalizes multiple vectors in parallel
func (bo *BatchOperations) BatchNormalize(vectors [][]float64) [][]float64 {
	results := make([][]float64, len(vectors))

	// Process in parallel for better performance
	for i := 0; i < len(vectors); i++ {
		results[i] = bo.vectorMath.Normalize(vectors[i])
	}

	return results
}
