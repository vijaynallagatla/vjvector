#!/bin/bash

# VJVector Test Suite Runner
# Runs comprehensive tests for the Q1 2025 implementation

set -e

echo "🧪 VJVector Comprehensive Test Suite"
echo "====================================="

# Set Go environment
export PATH="/usr/local/go/bin:$PATH"
export CGO_ENABLED=1

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

run_test() {
    local test_name="$1"
    local test_command="$2"
    
    echo -e "${BLUE}🔍 Running: ${test_name}${NC}"
    
    if eval "$test_command" > /tmp/test_output.log 2>&1; then
        echo -e "${GREEN}✅ PASSED: ${test_name}${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ FAILED: ${test_name}${NC}"
        echo "Error output:"
        cat /tmp/test_output.log
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

echo ""
echo "📋 Phase 1: Unit Tests"
echo "====================="

# Index tests
run_test "HNSW Index Tests" "go test ./pkg/index -v -run TestHNSW"
run_test "IVF Index Tests" "go test ./pkg/index -v -run TestIVF"

# Storage tests  
run_test "Memory Storage Tests" "go test ./pkg/storage -v -run TestMemoryStorage"

# Core tests (if any)
if ls pkg/core/*_test.go >/dev/null 2>&1; then
    run_test "Core Tests" "go test ./pkg/core -v"
fi

echo ""
echo "📋 Phase 2: Integration Tests"
echo "============================="

# Integration tests
if [ -f "tests/integration_test.go" ]; then
    run_test "End-to-End Workflow Tests" "go test ./tests -v -run TestEndToEndWorkflow"
    run_test "Storage Integration Tests" "go test ./tests -v -run TestStorageIntegration"
    run_test "Benchmark Integration Tests" "go test ./tests -v -run TestBenchmarkIntegration"
    run_test "Concurrent Operations Tests" "go test ./tests -v -run TestConcurrentOperations"
    run_test "Error Handling Tests" "go test ./tests -v -run TestErrorHandling"
    run_test "Memory Usage Tests" "go test ./tests -v -run TestMemoryUsage"
fi

echo ""
echo "📋 Phase 3: Performance Benchmarks"
echo "=================================="

# Benchmark tests
run_test "HNSW Benchmarks" "go test ./pkg/index -bench=BenchmarkHNSW -benchtime=100ms"
run_test "IVF Benchmarks" "go test ./pkg/index -bench=BenchmarkIVF -benchtime=100ms"
run_test "Storage Benchmarks" "go test ./pkg/storage -bench=BenchmarkMemoryStorage -benchtime=100ms"

echo ""
echo "📋 Phase 4: Build and Lint Tests"
echo "================================"

# Build tests
run_test "Full Project Build" "go build -v ./..."
run_test "Example Builds" "go build -v ./examples/..."
run_test "Command Build" "go build -v ./cmd/..."

# Lint tests
run_test "Code Linting" "./tools/run-lint.sh run"

echo ""
echo "📋 Phase 5: Demo Execution Tests"
echo "==============================="

# Demo tests
run_test "Q1 2025 Demo Execution" "timeout 30s go run ./examples/q1_2025_demo >/dev/null 2>&1"
run_test "Performance Demo Execution" "timeout 30s go run ./examples/performance_demo >/dev/null 2>&1"

echo ""
echo "📊 Test Results Summary"
echo "======================"
echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${PASSED_TESTS}${NC}"
echo -e "${RED}Failed: ${FAILED_TESTS}${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed! VJVector is ready for production.${NC}"
    exit 0
else
    echo -e "${RED}⚠️  Some tests failed. Please review the failures above.${NC}"
    exit 1
fi
