// Package profiling provides memory and performance profiling utilities for VJVector
package profiling

import (
	"fmt"
	"runtime"
	"time"
)

// MemoryStats holds memory usage statistics
type MemoryStats struct {
	AllocMB      float64 `json:"alloc_mb"`
	TotalAllocMB float64 `json:"total_alloc_mb"`
	SysMB        float64 `json:"sys_mb"`
	NumGC        uint32  `json:"num_gc"`
	GCCPUPercent float64 `json:"gc_cpu_percent"`
}

// MemoryProfiler provides memory usage monitoring and optimization
type MemoryProfiler struct {
	startTime   time.Time
	startStats  runtime.MemStats
	snapshots   []MemorySnapshot
	gcOptimized bool
}

// MemorySnapshot represents a point-in-time memory usage snapshot
type MemorySnapshot struct {
	Timestamp time.Time   `json:"timestamp"`
	Stats     MemoryStats `json:"stats"`
	Label     string      `json:"label"`
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler() *MemoryProfiler {
	mp := &MemoryProfiler{
		startTime: time.Now(),
		snapshots: make([]MemorySnapshot, 0),
	}

	// Record initial memory state
	runtime.ReadMemStats(&mp.startStats)
	mp.TakeSnapshot("initial")

	return mp
}

// TakeSnapshot records current memory usage
func (mp *MemoryProfiler) TakeSnapshot(label string) MemorySnapshot {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	stats := MemoryStats{
		AllocMB:      float64(ms.Alloc) / 1024 / 1024,
		TotalAllocMB: float64(ms.TotalAlloc) / 1024 / 1024,
		SysMB:        float64(ms.Sys) / 1024 / 1024,
		NumGC:        ms.NumGC,
		GCCPUPercent: ms.GCCPUFraction * 100,
	}

	snapshot := MemorySnapshot{
		Timestamp: time.Now(),
		Stats:     stats,
		Label:     label,
	}

	mp.snapshots = append(mp.snapshots, snapshot)
	return snapshot
}

// GetCurrentStats returns current memory statistics
func (mp *MemoryProfiler) GetCurrentStats() MemoryStats {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return MemoryStats{
		AllocMB:      float64(ms.Alloc) / 1024 / 1024,
		TotalAllocMB: float64(ms.TotalAlloc) / 1024 / 1024,
		SysMB:        float64(ms.Sys) / 1024 / 1024,
		NumGC:        ms.NumGC,
		GCCPUPercent: ms.GCCPUFraction * 100,
	}
}

// PrintStats prints formatted memory statistics
func (mp *MemoryProfiler) PrintStats() {
	current := mp.GetCurrentStats()
	fmt.Printf("📊 Memory Usage Statistics\n")
	fmt.Printf("   💾 Current Alloc: %.2f MB\n", current.AllocMB)
	fmt.Printf("   📈 Total Alloc: %.2f MB\n", current.TotalAllocMB)
	fmt.Printf("   🖥️  System Memory: %.2f MB\n", current.SysMB)
	fmt.Printf("   🗑️  GC Runs: %d\n", current.NumGC)
	fmt.Printf("   ⚡ GC CPU: %.2f%%\n", current.GCCPUPercent)
}

// PrintComparison prints memory usage comparison between snapshots
func (mp *MemoryProfiler) PrintComparison() {
	if len(mp.snapshots) < 2 {
		fmt.Println("Need at least 2 snapshots for comparison")
		return
	}

	first := mp.snapshots[0]
	last := mp.snapshots[len(mp.snapshots)-1]

	fmt.Printf("📊 Memory Usage Comparison (%s -> %s)\n", first.Label, last.Label)
	fmt.Printf("   💾 Alloc Change: %.2f MB -> %.2f MB (%.2f MB)\n",
		first.Stats.AllocMB, last.Stats.AllocMB,
		last.Stats.AllocMB-first.Stats.AllocMB)
	fmt.Printf("   📈 Total Alloc: %.2f MB -> %.2f MB\n",
		first.Stats.TotalAllocMB, last.Stats.TotalAllocMB)
	fmt.Printf("   🗑️  GC Runs: %d -> %d (+%d)\n",
		first.Stats.NumGC, last.Stats.NumGC,
		last.Stats.NumGC-first.Stats.NumGC)
}

// OptimizeGC applies garbage collection optimizations
func (mp *MemoryProfiler) OptimizeGC() {
	if mp.gcOptimized {
		return
	}

	// Force garbage collection
	runtime.GC()

	// Set GC target percentage (default is 100, lower values = more frequent GC)
	// For vector databases, we might want more frequent GC to keep memory usage stable
	oldGOGC := runtime.GOMAXPROCS(0)
	_ = oldGOGC // Avoid unused variable warning

	// Note: GOGC is controlled by environment variable, not runtime
	// In a real implementation, we would provide configuration options

	mp.gcOptimized = true
	mp.TakeSnapshot("gc_optimized")
}

// ForceGC triggers garbage collection and records the impact
func (mp *MemoryProfiler) ForceGC() MemorySnapshot {
	before := mp.TakeSnapshot("before_gc")
	runtime.GC()
	after := mp.TakeSnapshot("after_gc")

	fmt.Printf("🗑️  Forced GC: %.2f MB -> %.2f MB (freed %.2f MB)\n",
		before.Stats.AllocMB, after.Stats.AllocMB,
		before.Stats.AllocMB-after.Stats.AllocMB)

	return after
}

// GetSnapshots returns all recorded snapshots
func (mp *MemoryProfiler) GetSnapshots() []MemorySnapshot {
	return mp.snapshots
}

// VectorPool provides object pooling for vector operations
type VectorPool struct {
	floatSlicePools map[int]*FloatSlicePool
	maxDimension    int
}

// FloatSlicePool pools float64 slices of a specific size
type FloatSlicePool struct {
	dimension int
	pool      chan []float64
	size      int
}

// NewVectorPool creates a new vector pool
func NewVectorPool(maxDimension int, poolSize int) *VectorPool {
	return &VectorPool{
		floatSlicePools: make(map[int]*FloatSlicePool),
		maxDimension:    maxDimension,
	}
}

// GetFloatSlice gets a float64 slice from the pool
func (vp *VectorPool) GetFloatSlice(dimension int) []float64 {
	if dimension > vp.maxDimension {
		// Return new slice for oversized requests
		return make([]float64, dimension)
	}

	pool := vp.getOrCreatePool(dimension)

	select {
	case slice := <-pool.pool:
		// Clear the slice
		for i := range slice {
			slice[i] = 0
		}
		return slice
	default:
		// Pool is empty, create new slice
		return make([]float64, dimension)
	}
}

// PutFloatSlice returns a float64 slice to the pool
func (vp *VectorPool) PutFloatSlice(slice []float64) {
	dimension := len(slice)
	if dimension > vp.maxDimension {
		// Don't pool oversized slices
		return
	}

	pool := vp.getOrCreatePool(dimension)

	select {
	case pool.pool <- slice:
		// Successfully returned to pool
	default:
		// Pool is full, discard the slice
	}
}

func (vp *VectorPool) getOrCreatePool(dimension int) *FloatSlicePool {
	pool, exists := vp.floatSlicePools[dimension]
	if !exists {
		pool = &FloatSlicePool{
			dimension: dimension,
			pool:      make(chan []float64, 100), // Pool size of 100
			size:      100,
		}
		vp.floatSlicePools[dimension] = pool
	}
	return pool
}

// GetStats returns pool statistics
func (vp *VectorPool) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["max_dimension"] = vp.maxDimension
	stats["num_pools"] = len(vp.floatSlicePools)

	poolStats := make(map[string]interface{})
	for dim, pool := range vp.floatSlicePools {
		poolStats[fmt.Sprintf("dim_%d", dim)] = map[string]interface{}{
			"dimension":   pool.dimension,
			"pool_size":   pool.size,
			"available":   len(pool.pool),
			"utilization": float64(pool.size-len(pool.pool)) / float64(pool.size),
		}
	}
	stats["pools"] = poolStats

	return stats
}

// PrintPoolStats prints formatted pool statistics
func (vp *VectorPool) PrintPoolStats() {
	fmt.Printf("🏊 Vector Pool Statistics\n")
	fmt.Printf("   📐 Max Dimension: %d\n", vp.maxDimension)
	fmt.Printf("   🎱 Number of Pools: %d\n", len(vp.floatSlicePools))

	for dim, pool := range vp.floatSlicePools {
		utilization := float64(pool.size-len(pool.pool)) / float64(pool.size) * 100
		fmt.Printf("   📊 Dim %d: %d/%d used (%.1f%% utilization)\n",
			dim, pool.size-len(pool.pool), pool.size, utilization)
	}
}
