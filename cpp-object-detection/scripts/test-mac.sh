#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running C++ Object Detection Tests (macOS)${NC}"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"

# Build first if needed
if [ ! -d "$BUILD_DIR" ] || [ ! -f "$BUILD_DIR/object_detection" ]; then
    echo -e "${YELLOW}Build directory not found, building first...${NC}"
    "$SCRIPT_DIR/build-mac.sh"
fi

cd "$BUILD_DIR"

# Check for Google Test
echo -e "${YELLOW}Checking for Google Test...${NC}"
# First check if pkg-config is available
if ! command -v pkg-config &> /dev/null; then
    echo -e "${RED}Error: pkg-config not found. Please install pkg-config.${NC}"
    echo -e "${YELLOW}macOS: brew install pkg-config${NC}"
    exit 1
fi

if ! pkg-config --exists gtest; then
    echo -e "${RED}Warning: Google Test not found. Installing...${NC}"
    # Try to install gtest on macOS
    if command -v brew &> /dev/null; then
        brew install googletest
    else
        echo -e "${RED}Error: Please install Google Test manually with 'brew install googletest'${NC}"
        exit 1
    fi
fi

# Build tests if available
if [ -f "$PROJECT_ROOT/tests/CMakeLists.txt" ]; then
    echo -e "${YELLOW}Building tests...${NC}"
    make object_detection_tests

    echo -e "${YELLOW}Running unit tests...${NC}"
    if [ -f "./tests/object_detection_tests" ]; then
        ./tests/object_detection_tests
        echo -e "${GREEN}All unit tests passed!${NC}"
    else
        echo -e "${RED}Test executable not found at expected location${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}No test configuration found, skipping unit tests${NC}"
fi

# Integration test - check if executable runs with help
echo -e "${YELLOW}Running integration test...${NC}"
if ./object_detection --help > /dev/null 2>&1; then
    echo -e "${GREEN}Integration test passed - executable runs correctly${NC}"
else
    echo -e "${RED}Integration test failed - executable doesn't run${NC}"
    exit 1
fi

# Test basic argument parsing
echo -e "${YELLOW}Testing argument parsing...${NC}"
if ./object_detection --max-fps 3 --min-confidence 0.7 --verbose --help > /dev/null 2>&1; then
    echo -e "${GREEN}Argument parsing test passed${NC}"
else
    echo -e "${RED}Argument parsing test failed${NC}"
    exit 1
fi

echo -e "${GREEN}All tests completed successfully!${NC}"