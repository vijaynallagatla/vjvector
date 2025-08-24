#!/bin/bash

# Test script for VJVector API documentation endpoints
# This script verifies that the /docs and /openapi.yaml endpoints are working correctly

set -e

echo "🧪 Testing VJVector API Documentation Endpoints"
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Base URL
BASE_URL="http://localhost:8080"

# Function to test an endpoint
test_endpoint() {
    local endpoint=$1
    local expected_status=$2
    local description=$3
    
    echo -n "Testing $description... "
    
    if response=$(curl -s -w "%{http_code}" "$BASE_URL$endpoint" 2>/dev/null); then
        status_code="${response: -3}"
        content="${response%???}"
        
        if [ "$status_code" = "$expected_status" ]; then
            echo -e "${GREEN}✅ PASS${NC}"
            return 0
        else
            echo -e "${RED}❌ FAIL${NC} (Expected: $expected_status, Got: $status_code)"
            return 1
        fi
    else
        echo -e "${RED}❌ FAIL${NC} (Request failed)"
        return 1
    fi
}

# Function to test content
test_content() {
    local endpoint=$1
    local expected_content=$2
    local description=$3
    
    echo -n "Testing $description... "
    
    if content=$(curl -s "$BASE_URL$endpoint" 2>/dev/null); then
        if echo "$content" | grep -q "$expected_content"; then
            echo -e "${GREEN}✅ PASS${NC}"
            return 0
        else
            echo -e "${RED}❌ FAIL${NC} (Expected content not found)"
            return 1
        fi
    else
        echo -e "${RED}❌ FAIL${NC} (Request failed)"
        return 1
    fi
}

# Check if server is running
echo "Checking if server is running..."
if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ Server is not running at $BASE_URL${NC}"
    echo "Please start the server first with: go run ./cmd/api serve --config ./config.cluster.yaml"
    exit 1
fi
echo -e "${GREEN}✅ Server is running${NC}"
echo ""

# Test endpoints
echo "Testing API Endpoints:"
echo "----------------------"

# Test health endpoint
test_endpoint "/health" "200" "Health endpoint"

# Test OpenAPI endpoint
test_endpoint "/openapi.yaml" "200" "OpenAPI specification endpoint"

# Test docs endpoint
test_endpoint "/docs" "200" "Documentation page endpoint"

echo ""

# Test content
echo "Testing Content:"
echo "----------------"

# Test OpenAPI content
test_content "/openapi.yaml" "openapi: 3.0.3" "OpenAPI spec contains version"

# Test docs content
test_content "/docs" "swagger-ui" "Documentation page contains Swagger UI"

# Test CORS headers
echo ""
echo "Testing CORS Headers:"
echo "---------------------"

echo -n "Testing CORS preflight for OpenAPI... "
if cors_response=$(curl -s -H "Origin: http://localhost:8080" -H "Access-Control-Request-Method: GET" -X OPTIONS "$BASE_URL/openapi.yaml" -w "%{http_code}" 2>/dev/null); then
    status_code="${cors_response: -3}"
    if [ "$status_code" = "204" ]; then
        echo -e "${GREEN}✅ PASS${NC}"
    else
        echo -e "${RED}❌ FAIL${NC} (Expected: 204, Got: $status_code)"
    fi
else
    echo -e "${RED}❌ FAIL${NC} (Request failed)"
fi

echo -n "Testing CORS headers for OpenAPI... "
if cors_headers=$(curl -s -H "Origin: http://localhost:8080" "$BASE_URL/openapi.yaml" -I 2>/dev/null | grep -i "access-control-allow-origin"); then
    echo -e "${GREEN}✅ PASS${NC}"
else
    echo -e "${RED}❌ FAIL${NC} (CORS headers not found)"
fi

echo ""
echo "🎯 Documentation Test Summary:"
echo "=============================="

# Test browser simulation
echo -n "Testing browser-like request... "
if browser_response=$(curl -s -H "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36" -H "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8" "$BASE_URL/docs" 2>/dev/null); then
    if echo "$browser_response" | grep -q "swagger-ui" && echo "$browser_response" | grep -q "http://localhost:8080/openapi.yaml"; then
        echo -e "${GREEN}✅ PASS${NC}"
    else
        echo -e "${RED}❌ FAIL${NC} (Browser simulation failed)"
    fi
else
    echo -e "${RED}❌ FAIL${NC} (Request failed)"
fi

echo ""
echo -e "${GREEN}🎉 Documentation testing completed!${NC}"
echo ""
echo "To view the documentation in your browser:"
echo "1. Open: $BASE_URL/docs"
echo "2. Check the browser console for any JavaScript errors"
echo "3. Verify that the OpenAPI spec loads correctly"
echo ""
echo "If you encounter issues:"
echo "- Check the browser console for error messages"
echo "- Verify the server is running and accessible"
echo "- Check that the OpenAPI spec file exists at docs/api/openapi.yaml"
