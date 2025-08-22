#!/bin/bash

# Performance Optimization Script for VJVector AI Integration
# Week 15-16: OpenAI Integration & Performance Optimization

set -e

echo "🚀 VJVector Performance Optimization"
echo "===================================="
echo "Week 15-16: OpenAI Integration & Performance Tuning"
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

# Check if Go is available
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

print_status "Go version: $(go version)"

# Create performance test directory
PERF_DIR="/tmp/vjvector_performance"
mkdir -p "$PERF_DIR"

print_status "Performance test directory: $PERF_DIR"

# Step 1: Run Unit Tests
echo ""
print_status "Step 1: Running Unit Tests"
echo "---------------------------------"
if go test -v ./pkg/embedding/...; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

# Step 2: Run Integration Tests
echo ""
print_status "Step 2: Running Integration Tests"
echo "----------------------------------------"
if go test -v ./pkg/embedding/... -run TestEmbeddingService_Integration; then
    print_success "Integration tests passed"
else
    print_error "Integration tests failed"
    exit 1
fi

# Step 3: Run Performance Benchmarks
echo ""
print_status "Step 3: Running Performance Benchmarks"
echo "---------------------------------------------"
print_status "Benchmarking embedding service performance..."

# Run benchmarks with different configurations
echo "Running benchmarks..."
go test -bench=. -benchmem ./pkg/embedding/... > "$PERF_DIR/benchmarks.txt" 2>&1

if [ $? -eq 0 ]; then
    print_success "Benchmarks completed successfully"
    print_status "Benchmark results saved to: $PERF_DIR/benchmarks.txt"
    
    # Display key benchmark results
    echo ""
    print_status "Key Benchmark Results:"
    echo "---------------------------"
    grep -E "^(Benchmark|op/s|ns/op|B/op|allocs/op)" "$PERF_DIR/benchmarks.txt" | head -20
else
    print_error "Benchmarks failed"
    exit 1
fi

# Step 4: Performance Analysis
echo ""
print_status "Step 4: Performance Analysis"
echo "-----------------------------------"

# Analyze cache performance
print_status "Analyzing cache performance..."
go test -bench=BenchmarkEmbeddingCache -benchmem ./pkg/embedding/... > "$PERF_DIR/cache_benchmarks.txt" 2>&1

# Analyze rate limiter performance
print_status "Analyzing rate limiter performance..."
go test -bench=BenchmarkRateLimiter -benchmem ./pkg/embedding/... > "$PERF_DIR/rate_limiter_benchmarks.txt" 2>&1

# Analyze retry manager performance
print_status "Analyzing retry manager performance..."
go test -bench=BenchmarkRetryManager -benchmem ./pkg/embedding/... > "$PERF_DIR/retry_benchmarks.txt" 2>&1

# Step 5: Memory Profiling
echo ""
print_status "Step 5: Memory Profiling"
echo "-------------------------------"

print_status "Running memory profiling tests..."
go test -bench=BenchmarkEmbeddingService_GenerateEmbeddings -memprofile="$PERF_DIR/memory.prof" ./pkg/embedding/... > "$PERF_DIR/memory_benchmark.txt" 2>&1

if [ $? -eq 0 ]; then
    print_success "Memory profiling completed"
    print_status "Memory profile saved to: $PERF_DIR/memory.prof"
else
    print_warning "Memory profiling failed (this is optional)"
fi

# Step 6: CPU Profiling
echo ""
print_status "Step 6: CPU Profiling"
echo "-----------------------------"

print_status "Running CPU profiling tests..."
go test -bench=BenchmarkEmbeddingService_GenerateEmbeddings -cpuprofile="$PERF_DIR/cpu.prof" ./pkg/embedding/... > "$PERF_DIR/cpu_benchmark.txt" 2>&1

if [ $? -eq 0 ]; then
    print_success "CPU profiling completed"
    print_status "CPU profile saved to: $PERF_DIR/cpu.prof"
else
    print_warning "CPU profiling failed (this is optional)"
fi

# Step 7: Performance Metrics Summary
echo ""
print_status "Step 7: Performance Metrics Summary"
echo "------------------------------------------"

# Calculate performance metrics
print_status "Calculating performance metrics..."

# Extract key metrics from benchmark results
if [ -f "$PERF_DIR/benchmarks.txt" ]; then
    echo "📊 Performance Summary:"
    echo "======================="
    
    # Extract embedding generation performance
    EMBEDDING_PERF=$(grep "BenchmarkEmbeddingService_GenerateEmbeddings/SingleText" "$PERF_DIR/benchmarks.txt" | tail -1)
    if [ ! -z "$EMBEDDING_PERF" ]; then
        echo "🔹 Single Text Embedding: $EMBEDDING_PERF"
    fi
    
    # Extract batch processing performance
    BATCH_PERF=$(grep "BenchmarkEmbeddingService_GenerateEmbeddings/BatchTexts" "$PERF_DIR/benchmarks.txt" | tail -1)
    if [ ! -z "$BATCH_PERF" ]; then
        echo "🔹 Batch Text Processing: $BATCH_PERF"
    fi
    
    # Extract caching performance
    CACHE_PERF=$(grep "BenchmarkEmbeddingService_GenerateEmbeddings/WithCaching" "$PERF_DIR/benchmarks.txt" | tail -1)
    if [ ! -z "$CACHE_PERF" ]; then
        echo "🔹 Cached Embeddings: $CACHE_PERF"
    fi
    
    # Extract cache operations performance
    CACHE_SET=$(grep "BenchmarkEmbeddingCache/Set" "$PERF_DIR/benchmarks.txt" | tail -1)
    if [ ! -z "$CACHE_SET" ]; then
        echo "🔹 Cache Set Operations: $CACHE_SET"
    fi
    
    CACHE_GET=$(grep "BenchmarkEmbeddingCache/Get" "$PERF_DIR/benchmarks.txt" | tail -1)
    if [ ! -z "$CACHE_GET" ]; then
        echo "🔹 Cache Get Operations: $CACHE_GET"
    fi
    
    # Extract rate limiting performance
    RATE_LIMIT=$(grep "BenchmarkRateLimiter/Allow" "$PERF_DIR/benchmarks.txt" | tail -1)
    if [ ! -z "$RATE_LIMIT" ]; then
        echo "🔹 Rate Limiting: $RATE_LIMIT"
    fi
    
    # Extract retry mechanism performance
    RETRY=$(grep "BenchmarkRetryManager/Success" "$PERF_DIR/benchmarks.txt" | tail -1)
    if [ ! -z "$RETRY" ]; then
        echo "🔹 Retry Mechanism: $RETRY"
    fi
fi

# Step 8: Performance Recommendations
echo ""
print_status "Step 8: Performance Recommendations"
echo "------------------------------------------"

echo "🎯 Performance Optimization Recommendations:"
echo "==========================================="

# Check if benchmarks indicate performance issues
if [ -f "$PERF_DIR/benchmarks.txt" ]; then
    # Check for high memory allocation
    HIGH_MEM=$(grep "B/op" "$PERF_DIR/benchmarks.txt" | grep -E "[0-9]{5,}" | head -5)
    if [ ! -z "$HIGH_MEM" ]; then
        echo "⚠️  High Memory Allocation Detected:"
        echo "$HIGH_MEM"
        echo "💡 Consider: Memory pooling, object reuse, reducing allocations"
    fi
    
    # Check for high allocation count
    HIGH_ALLOCS=$(grep "allocs/op" "$PERF_DIR/benchmarks.txt" | grep -E "[0-9]{3,}" | head -5)
    if [ ! -z "$HIGH_ALLOCS" ]; then
        echo "⚠️  High Allocation Count Detected:"
        echo "$HIGH_ALLOCS"
        echo "💡 Consider: Reducing object creation, using sync.Pool"
    fi
    
    # Check for slow operations
    SLOW_OPS=$(grep "ns/op" "$PERF_DIR/benchmarks.txt" | grep -E "[0-9]{7,}" | head -5)
    if [ ! -z "$SLOW_OPS" ]; then
        echo "⚠️  Slow Operations Detected:"
        echo "$SLOW_OPS"
        echo "💡 Consider: Algorithm optimization, parallelization, caching"
    fi
fi

echo ""
echo "🔧 General Optimization Strategies:"
echo "1. Enable caching for repeated embeddings"
echo "2. Use batch processing for multiple texts"
echo "3. Implement connection pooling for API calls"
echo "4. Use async processing where possible"
echo "5. Monitor and tune rate limiting parameters"
echo "6. Implement circuit breaker patterns"
echo "7. Use memory-mapped files for large datasets"
echo "8. Implement background garbage collection"

# Step 9: Generate Performance Report
echo ""
print_status "Step 9: Generating Performance Report"
echo "---------------------------------------------"

REPORT_FILE="$PERF_DIR/performance_report.md"
cat > "$REPORT_FILE" << EOF
# VJVector Performance Report
## Week 15-16: OpenAI Integration & Performance Optimization

### Test Environment
- **Go Version**: $(go version)
- **Date**: $(date)
- **Test Directory**: $PERF_DIR

### Performance Metrics

#### Embedding Service Performance
$(grep -E "BenchmarkEmbeddingService" "$PERF_DIR/benchmarks.txt" | head -10)

#### Cache Performance
$(grep -E "BenchmarkEmbeddingCache" "$PERF_DIR/benchmarks.txt" | head -10)

#### Rate Limiting Performance
$(grep -E "BenchmarkRateLimiter" "$PERF_DIR/benchmarks.txt" | head -10)

#### Retry Mechanism Performance
$(grep -E "BenchmarkRetryManager" "$PERF_DIR/benchmarks.txt" | head -10)

### Performance Analysis
- **Memory Usage**: Check memory.prof for detailed analysis
- **CPU Usage**: Check cpu.prof for detailed analysis
- **Benchmark Results**: Full results in benchmarks.txt

### Recommendations
1. Enable caching for repeated embeddings
2. Use batch processing for multiple texts
3. Monitor rate limiting performance
4. Optimize retry mechanisms
5. Profile memory usage for large datasets

### Next Steps
1. Implement performance monitoring
2. Add performance regression tests
3. Optimize based on profiling results
4. Benchmark against production workloads
EOF

print_success "Performance report generated: $REPORT_FILE"

# Step 10: Cleanup and Summary
echo ""
print_status "Step 10: Cleanup and Summary"
echo "-----------------------------------"

print_success "Performance optimization completed successfully!"
echo ""
print_status "Generated Files:"
echo "- 📊 Benchmarks: $PERF_DIR/benchmarks.txt"
echo "- 🧠 Memory Profile: $PERF_DIR/memory.prof"
echo "- ⚡ CPU Profile: $PERF_DIR/cpu.prof"
echo "- 📋 Performance Report: $REPORT_FILE"
echo "- 🔍 Cache Benchmarks: $PERF_DIR/cache_benchmarks.txt"
echo "- 🚦 Rate Limiter Benchmarks: $PERF_DIR/rate_limiter_benchmarks.txt"
echo "- 🔄 Retry Benchmarks: $PERF_DIR/retry_benchmarks.txt"

echo ""
print_status "Next Steps for Week 15-16:"
echo "1. Analyze profiling data for bottlenecks"
echo "2. Implement performance optimizations"
echo "3. Run regression tests"
echo "4. Benchmark with real OpenAI API"
echo "5. Document performance characteristics"

echo ""
print_status "VJVector Performance Optimization is ready for production! 🚀"
