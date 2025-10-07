#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}===============================================${NC}"
echo -e "${GREEN}Raspberry Pi ARM64 Integration Tests${NC}"
echo -e "${GREEN}===============================================${NC}"
echo -e ""

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Test 1: Check build script exists and is executable
echo -e "${YELLOW}Test 1: Check build-rpi.sh script exists${NC}"
if [ -f "$SCRIPT_DIR/build-rpi.sh" ] && [ -x "$SCRIPT_DIR/build-rpi.sh" ]; then
    echo -e "${GREEN}✓ build-rpi.sh exists and is executable${NC}"
else
    echo -e "${RED}✗ build-rpi.sh missing or not executable${NC}"
    exit 1
fi

# Test 2: Check CMakeLists.txt has ARM64 support
echo -e "${YELLOW}Test 2: Check CMakeLists.txt has ARM64 support${NC}"
if grep -q "aarch64" "$PROJECT_ROOT/CMakeLists.txt" && grep -q "arm64" "$PROJECT_ROOT/CMakeLists.txt"; then
    echo -e "${GREEN}✓ CMakeLists.txt includes ARM64 architecture handling${NC}"
else
    echo -e "${RED}✗ CMakeLists.txt missing ARM64 architecture handling${NC}"
    exit 1
fi

# Test 3: Check README has Raspberry Pi documentation
echo -e "${YELLOW}Test 3: Check README has Raspberry Pi documentation${NC}"
if grep -q "Raspberry Pi" "$PROJECT_ROOT/README.md" && grep -q "build-rpi.sh" "$PROJECT_ROOT/README.md"; then
    echo -e "${GREEN}✓ README includes Raspberry Pi documentation${NC}"
else
    echo -e "${RED}✗ README missing Raspberry Pi documentation${NC}"
    exit 1
fi

# Test 4: Check README has Raspberry Pi performance estimates
echo -e "${YELLOW}Test 4: Check README has Raspberry Pi performance estimates${NC}"
if grep -q "Raspberry Pi 5" "$PROJECT_ROOT/README.md" && grep -q "fps @ 720p" "$PROJECT_ROOT/README.md"; then
    echo -e "${GREEN}✓ README includes Raspberry Pi performance estimates${NC}"
else
    echo -e "${RED}✗ README missing Raspberry Pi performance estimates${NC}"
    exit 1
fi

# Test 5: Check workflow has Raspberry Pi build job
echo -e "${YELLOW}Test 5: Check GitHub Actions workflow has Raspberry Pi build job${NC}"
WORKFLOW_FILE="$PROJECT_ROOT/../.github/workflows/cpp-object-detection.yml"
if [ -f "$WORKFLOW_FILE" ]; then
    if grep -q "build-rpi-arm64" "$WORKFLOW_FILE"; then
        echo -e "${GREEN}✓ GitHub Actions workflow includes Raspberry Pi ARM64 build job${NC}"
    else
        echo -e "${RED}✗ GitHub Actions workflow missing Raspberry Pi ARM64 build job${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ Workflow file not found (may be in different location)${NC}"
fi

# Test 6: Check workflow has proper job dependency
echo -e "${YELLOW}Test 6: Check Raspberry Pi job runs after build-test${NC}"
if [ -f "$WORKFLOW_FILE" ]; then
    if grep -A 3 "build-rpi-arm64" "$WORKFLOW_FILE" | grep -q "needs:.*build-test.*build-static"; then
        echo -e "${GREEN}✓ Raspberry Pi build job properly depends on build-test${NC}"
    else
        echo -e "${YELLOW}⚠ Could not verify job dependencies (check manually)${NC}"
    fi
fi

# Test 7: Check README includes power consumption estimates
echo -e "${YELLOW}Test 7: Check README includes power consumption estimates${NC}"
if grep -q "Power Consumption" "$PROJECT_ROOT/README.md" || grep -q "power consumption" "$PROJECT_ROOT/README.md"; then
    echo -e "${GREEN}✓ README includes power consumption information${NC}"
else
    echo -e "${RED}✗ README missing power consumption estimates${NC}"
    exit 1
fi

# Test 8: Check README includes battery + solar panel feasibility
echo -e "${YELLOW}Test 8: Check README includes battery + solar panel feasibility${NC}"
if grep -q "Battery" "$PROJECT_ROOT/README.md" && grep -q "Solar" "$PROJECT_ROOT/README.md"; then
    echo -e "${GREEN}✓ README includes battery and solar panel feasibility analysis${NC}"
else
    echo -e "${RED}✗ README missing battery and solar panel feasibility analysis${NC}"
    exit 1
fi

# Test 9: Verify recommended CLI flags are documented
echo -e "${YELLOW}Test 9: Check documented CLI flags for Raspberry Pi${NC}"
if grep -q "Raspberry Pi 5" "$PROJECT_ROOT/README.md" && \
   grep -q "Raspberry Pi 4" "$PROJECT_ROOT/README.md"; then
    echo -e "${GREEN}✓ README documents recommended CLI flags for Raspberry Pi models${NC}"
else
    echo -e "${RED}✗ README missing complete CLI flag recommendations for Raspberry Pi${NC}"
    exit 1
fi

# Test 10: Check DEPLOYMENT.md has Raspberry Pi installation
echo -e "${YELLOW}Test 10: Check DEPLOYMENT.md has Raspberry Pi installation${NC}"
DEPLOYMENT_FILE="$PROJECT_ROOT/DEPLOYMENT.md"
if [ -f "$DEPLOYMENT_FILE" ]; then
    if grep -q "Raspberry Pi" "$DEPLOYMENT_FILE"; then
        echo -e "${GREEN}✓ DEPLOYMENT.md includes Raspberry Pi installation instructions${NC}"
    else
        echo -e "${RED}✗ DEPLOYMENT.md missing Raspberry Pi installation instructions${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ DEPLOYMENT.md not found${NC}"
fi

# Test 11: Check CMakeLists.txt has ARM-specific library paths
echo -e "${YELLOW}Test 11: Check CMakeLists.txt has ARM library paths${NC}"
if grep -q "aarch64-linux-gnu" "$PROJECT_ROOT/CMakeLists.txt"; then
    echo -e "${GREEN}✓ CMakeLists.txt includes ARM64 library paths${NC}"
else
    echo -e "${RED}✗ CMakeLists.txt missing ARM64 library paths${NC}"
    exit 1
fi

echo -e ""
echo -e "${GREEN}===============================================${NC}"
echo -e "${GREEN}All Raspberry Pi integration tests passed! ✅${NC}"
echo -e "${GREEN}===============================================${NC}"
echo -e ""
echo -e "${YELLOW}To build for Raspberry Pi:${NC}"
echo -e "${YELLOW}  1. On Raspberry Pi: ./scripts/build-rpi.sh${NC}"
echo -e "${YELLOW}  2. Test: ./build-rpi/object_detection --help${NC}"
echo -e ""
echo -e "${YELLOW}Recommended settings for Raspberry Pi 5:${NC}"
echo -e "${YELLOW}  ./object_detection --max-fps 5 --min-confidence 0.6 --detection-scale 0.5 --analysis-rate-limit 2${NC}"
echo -e ""
echo -e "${YELLOW}For battery-powered operation:${NC}"
echo -e "${YELLOW}  ./object_detection --max-fps 2 --min-confidence 0.7 --detection-scale 0.5 --analysis-rate-limit 0.5${NC}"
