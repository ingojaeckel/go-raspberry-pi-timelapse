#!/bin/bash
# OpenAPI Validation and Code Generation Check
# This script validates the OpenAPI setup is working correctly

set -e

echo "🔍 OpenAPI Integration Validation Script"
echo "========================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check for required tools
echo "📋 Checking required tools..."
echo ""

check_command() {
    if command -v $1 &> /dev/null; then
        echo -e "${GREEN}✓${NC} $1 is installed"
        return 0
    else
        echo -e "${RED}✗${NC} $1 is not installed"
        return 1
    fi
}

TOOLS_OK=true

if ! check_command spectral; then
    echo "  Install: npm install -g @stoplight/spectral-cli"
    TOOLS_OK=false
fi

if ! check_command go; then
    echo "  Install: https://golang.org/doc/install"
    TOOLS_OK=false
fi

if ! check_command npm; then
    echo "  Install: https://nodejs.org/"
    TOOLS_OK=false
fi

echo ""

if [ "$TOOLS_OK" = false ]; then
    echo -e "${YELLOW}⚠${NC}  Some tools are missing. Install them before continuing."
    echo "   Or run: make install-tools"
    exit 1
fi

# Lint OpenAPI spec
echo "🔬 Linting OpenAPI specification..."
if spectral lint openapi.yaml > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} OpenAPI spec is valid"
else
    echo -e "${RED}✗${NC} OpenAPI spec has errors"
    spectral lint openapi.yaml
    exit 1
fi
echo ""

# Generate Go code
echo "🔨 Generating Go code..."
if go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -config codegen-config.yaml openapi.yaml > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Go code generated successfully"
    if [ -f "generated/server.gen.go" ]; then
        LINES=$(wc -l < generated/server.gen.go)
        echo "  Generated file: generated/server.gen.go ($LINES lines)"
    fi
else
    echo -e "${RED}✗${NC} Failed to generate Go code"
    exit 1
fi
echo ""

# Generate TypeScript code
echo "🔨 Generating TypeScript code..."
cd frontend
if npm run generate:api > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} TypeScript code generated successfully"
    if [ -f "src/generated/api.ts" ]; then
        LINES=$(wc -l < src/generated/api.ts)
        echo "  Generated file: frontend/src/generated/api.ts ($LINES lines)"
    fi
else
    echo -e "${RED}✗${NC} Failed to generate TypeScript code"
    cd ..
    exit 1
fi
cd ..
echo ""

# Run Go tests
echo "🧪 Running Go tests..."
if go test ./... > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} All Go tests passed"
else
    echo -e "${RED}✗${NC} Some Go tests failed"
    go test ./...
    exit 1
fi
echo ""

# Build Go binary
echo "🏗️  Building Go application..."
if go build -v . > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Go application built successfully"
    if [ -f "go-raspberry-pi-timelapse" ]; then
        SIZE=$(ls -lh go-raspberry-pi-timelapse | awk '{print $5}')
        echo "  Binary: go-raspberry-pi-timelapse ($SIZE)"
    fi
else
    echo -e "${RED}✗${NC} Failed to build Go application"
    exit 1
fi
echo ""

echo "========================================"
echo -e "${GREEN}✅ All checks passed!${NC}"
echo ""
echo "Your OpenAPI integration is working correctly."
echo ""
echo "Next steps:"
echo "  - Run 'make build' to build the full application"
echo "  - Run 'make test' to run all tests"
echo "  - See OPENAPI.md for more information"
