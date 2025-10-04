#!/bin/bash
set -e

# macOS Integration Test Script
# This script tests macOS-specific functionality for the cpp-object-detection project

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}macOS Integration Test for C++ Object Detection${NC}"

# Detect if we're actually running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo -e "${YELLOW}Warning: This script is designed for macOS but running on $OSTYPE${NC}"
    echo -e "${YELLOW}Will test macOS-specific code paths in cross-platform scripts${NC}"
fi

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${YELLOW}Testing macOS build script...${NC}"
if [ -f "$SCRIPT_DIR/build-mac.sh" ] && [ -x "$SCRIPT_DIR/build-mac.sh" ]; then
    echo -e "${GREEN}✓ build-mac.sh exists and is executable${NC}"
else
    echo -e "${RED}✗ build-mac.sh missing or not executable${NC}"
    exit 1
fi

echo -e "${YELLOW}Testing macOS test script...${NC}"
if [ -f "$SCRIPT_DIR/test-mac.sh" ] && [ -x "$SCRIPT_DIR/test-mac.sh" ]; then
    echo -e "${GREEN}✓ test-mac.sh exists and is executable${NC}"
else
    echo -e "${RED}✗ test-mac.sh missing or not executable${NC}"
    exit 1
fi

echo -e "${YELLOW}Testing cross-platform build script detection...${NC}"
# Test that the build script would use the right commands on macOS
if grep -q "sysctl -n hw.ncpu" "$SCRIPT_DIR/build.sh"; then
    echo -e "${GREEN}✓ build.sh contains macOS CPU detection${NC}"
else
    echo -e "${RED}✗ build.sh missing macOS CPU detection${NC}"
    exit 1
fi

if grep -q "darwin" "$SCRIPT_DIR/build.sh"; then
    echo -e "${GREEN}✓ build.sh contains macOS platform detection${NC}"
else
    echo -e "${RED}✗ build.sh missing macOS platform detection${NC}"
    exit 1
fi

echo -e "${YELLOW}Testing macOS scripts pkg-config dependency checks...${NC}"
if grep -q "command -v pkg-config" "$SCRIPT_DIR/build-mac.sh"; then
    echo -e "${GREEN}✓ build-mac.sh checks for pkg-config availability${NC}"
else
    echo -e "${RED}✗ build-mac.sh missing pkg-config check${NC}"
    exit 1
fi

if grep -q "brew install pkg-config" "$SCRIPT_DIR/build-mac.sh"; then
    echo -e "${GREEN}✓ build-mac.sh suggests pkg-config installation${NC}"
else
    echo -e "${RED}✗ build-mac.sh missing pkg-config installation suggestion${NC}"
    exit 1
fi

if grep -q "command -v pkg-config" "$SCRIPT_DIR/test-mac.sh"; then
    echo -e "${GREEN}✓ test-mac.sh checks for pkg-config availability${NC}"
else
    echo -e "${RED}✗ test-mac.sh missing pkg-config check${NC}"
    exit 1
fi

echo -e "${YELLOW}Testing CMakeLists.txt macOS support...${NC}"
if grep -q "if(APPLE)" "$PROJECT_ROOT/CMakeLists.txt"; then
    echo -e "${GREEN}✓ CMakeLists.txt contains macOS-specific configuration${NC}"
else
    echo -e "${RED}✗ CMakeLists.txt missing macOS configuration${NC}"
    exit 1
fi

if grep -q "find_package(OpenCV REQUIRED)" "$PROJECT_ROOT/CMakeLists.txt" && grep -A5 "if(APPLE)" "$PROJECT_ROOT/CMakeLists.txt" | grep -q "find_package(OpenCV REQUIRED)"; then
    echo -e "${GREEN}✓ CMakeLists.txt uses simple OpenCV find for macOS${NC}"
else
    echo -e "${RED}✗ CMakeLists.txt missing proper macOS OpenCV configuration${NC}"
    exit 1
fi

echo -e "${YELLOW}Testing README macOS documentation...${NC}"
if grep -q "macOS" "$PROJECT_ROOT/README.md" && grep -q "build-mac.sh" "$PROJECT_ROOT/README.md"; then
    echo -e "${GREEN}✓ README.md contains macOS build instructions${NC}"
else
    echo -e "${RED}✗ README.md missing macOS documentation${NC}"
    exit 1
fi

echo -e "${YELLOW}Testing model download URLs...${NC}"
if grep -q "v7.0" "$SCRIPT_DIR/build-mac.sh"; then
    echo -e "${GREEN}✓ macOS build script uses correct YOLO model URL${NC}"
else
    echo -e "${RED}✗ macOS build script has incorrect model URL${NC}"
    exit 1
fi

echo -e "${YELLOW}Testing CI workflow macOS support...${NC}"
if grep -q "macos-latest" "$PROJECT_ROOT/../.github/workflows/cpp-object-detection.yml"; then
    echo -e "${GREEN}✓ GitHub Actions workflow includes macOS runner${NC}"
else
    echo -e "${RED}✗ GitHub Actions workflow missing macOS support${NC}"
    exit 1
fi

if grep -q "brew install" "$PROJECT_ROOT/../.github/workflows/cpp-object-detection.yml"; then
    echo -e "${GREEN}✓ GitHub Actions workflow includes homebrew dependencies${NC}"
else
    echo -e "${RED}✗ GitHub Actions workflow missing homebrew setup${NC}"
    exit 1
fi

echo -e "${GREEN}All macOS integration tests passed! ✅${NC}"
echo -e "${YELLOW}To test on actual macOS hardware:${NC}"
echo -e "${YELLOW}1. Clone this repo on a Mac${NC}"
echo -e "${YELLOW}2. Run: brew install cmake opencv pkg-config googletest${NC}"
echo -e "${YELLOW}3. Run: ./scripts/build-mac.sh${NC}"
echo -e "${YELLOW}4. Run: ./scripts/test-mac.sh${NC}"