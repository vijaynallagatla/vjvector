#!/bin/bash

# Test script for VJVector OpenAPI endpoints
set -e

echo "🧪 Testing VJVector OpenAPI Endpoints"
echo "====================================="

# Start the API server in background
echo "🚀 Starting API server..."
./api > api.log 2>&1 &
API_PID=$!

# Wait for server to start
echo "⏳ Waiting for server to start..."
sleep 3

# Test health endpoint
echo -e "\n1️⃣ Testing Health Endpoint:"
curl -s http://localhost:8080/health | jq .

# Test OpenAPI specification
echo -e "\n2️⃣ Testing OpenAPI Specification:"
curl -s http://localhost:8080/openapi.yaml | head -10

# Test API documentation
echo -e "\n3️⃣ Testing API Documentation:"
curl -s http://localhost:8080/docs | head -5

# Test v1 endpoints
echo -e "\n4️⃣ Testing v1 Indexes Endpoint:"
curl -s http://localhost:8080/v1/indexes | jq .

# Create a test index
echo -e "\n5️⃣ Creating Test Index:"
curl -s -X POST http://localhost:8080/v1/indexes \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test_index",
    "type": "hnsw",
    "dimension": 128,
    "max_elements": 1000,
    "m": 16,
    "ef_construction": 200,
    "ef_search": 100,
    "max_layers": 16,
    "distance_metric": "cosine",
    "normalize": true
  }' | jq .

# List indexes
echo -e "\n6️⃣ Listing Indexes:"
curl -s http://localhost:8080/v1/indexes | jq .

# Cleanup
echo -e "\n🛑 Stopping API server..."
kill $API_PID 2>/dev/null || true
wait $API_PID 2>/dev/null || true

echo -e "\n✅ OpenAPI test completed successfully!"
echo "📚 API Documentation: http://localhost:8080/docs"
echo "🔍 OpenAPI Spec: http://localhost:8080/openapi.yaml"
