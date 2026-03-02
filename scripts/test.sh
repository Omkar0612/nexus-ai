#!/bin/bash

# Nexus AI Test Runner
# Runs all tests with coverage and race detection

set -e

echo "🧪 Nexus AI Test Suite"
echo "====================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✓${NC} Go version: $(go version)"
echo ""

# Set test environment variables
export NEXUS_MESH_PORT=15353  # Non-privileged port for testing
export NEXUS_PREDICTIVE_CONFIDENCE=0.7
export NEXUS_LOG_LEVEL=error  # Reduce noise during tests

echo "📦 Installing dependencies..."
go mod download
go mod verify
echo -e "${GREEN}✓${NC} Dependencies verified"
echo ""

# Run go vet
echo "🔍 Running go vet..."
if go vet ./...; then
    echo -e "${GREEN}✓${NC} go vet passed"
else
    echo -e "${RED}❌ go vet failed${NC}"
    exit 1
fi
echo ""

# Run go fmt check
echo "📐 Checking code formatting..."
UNFORMATTED=$(gofmt -l .)
if [ -z "$UNFORMATTED" ]; then
    echo -e "${GREEN}✓${NC} Code is properly formatted"
else
    echo -e "${RED}❌ The following files need formatting:${NC}"
    echo "$UNFORMATTED"
    echo ""
    echo "Run: gofmt -w ."
    exit 1
fi
echo ""

# Run tests with race detector
echo "🏃 Running tests with race detector..."
if go test -race -timeout 30s ./internal/mesh ./internal/predictive ./internal/shadow ./internal/n8n 2>&1 | tee test_output.log; then
    echo -e "${GREEN}✓${NC} All tests passed"
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
echo ""

# Run tests with coverage
echo "📊 Generating coverage report..."
go test -coverprofile=coverage.out -covermode=atomic ./internal/mesh ./internal/predictive ./internal/shadow ./internal/n8n

if [ -f coverage.out ]; then
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo -e "${GREEN}✓${NC} Total coverage: $COVERAGE"
    
    # Generate HTML report
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}✓${NC} HTML coverage report: coverage.html"
    
    # Check if coverage meets minimum threshold
    COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
    if (( $(echo "$COVERAGE_NUM < 40" | bc -l) )); then
        echo -e "${YELLOW}⚠${NC}  Coverage is below 40% threshold"
    fi
fi
echo ""

# Build check
echo "🔨 Testing build..."
if go build -o /tmp/nexus-ai-test ./cmd/nexus 2>&1 | tee build_output.log; then
    echo -e "${GREEN}✓${NC} Build successful"
    rm /tmp/nexus-ai-test 2>/dev/null || true
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
echo ""

# Check for common issues
echo "🔍 Running static analysis..."

# Check for TODO comments
TODO_COUNT=$(grep -r "TODO" internal/ --include="*.go" | wc -l)
echo "  - Found $TODO_COUNT TODO comments"

# Check for panic calls
PANIC_COUNT=$(grep -r "panic(" internal/ --include="*.go" | grep -v "_test.go" | wc -l)
if [ $PANIC_COUNT -gt 0 ]; then
    echo -e "  ${YELLOW}⚠${NC}  Found $PANIC_COUNT panic() calls in non-test code"
else
    echo -e "  ${GREEN}✓${NC} No panic() calls in production code"
fi

# Check for fmt.Print usage
PRINT_COUNT=$(grep -r "fmt.Print" internal/ --include="*.go" | grep -v "_test.go" | wc -l)
if [ $PRINT_COUNT -gt 0 ]; then
    echo -e "  ${YELLOW}⚠${NC}  Found $PRINT_COUNT fmt.Print calls (use zerolog instead)"
else
    echo -e "  ${GREEN}✓${NC} No fmt.Print calls found"
fi

echo ""
echo "====================================="
echo -e "${GREEN}✅ All checks passed!${NC}"
echo "====================================="
echo ""
echo "Test artifacts:"
echo "  - test_output.log"
echo "  - coverage.out"
echo "  - coverage.html"
echo ""
