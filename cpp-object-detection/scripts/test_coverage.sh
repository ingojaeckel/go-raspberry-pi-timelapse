#!/bin/bash
set -e

echo "Running C++ Object Detection Tests with Code Coverage"

# Check dependencies
echo "Checking dependencies..."
if ! command -v lcov &> /dev/null; then
    echo "lcov not found. Installing..."
    sudo apt-get update && sudo apt-get install -y lcov gcovr
fi

if ! command -v gcovr &> /dev/null; then
    echo "gcovr not found. Installing..."
    sudo apt-get update && sudo apt-get install -y gcovr
fi

# Create build directory
BUILD_DIR="build-coverage"
if [ -d "$BUILD_DIR" ]; then
    echo "Removing existing coverage build directory..."
    rm -rf "$BUILD_DIR"
fi

mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

echo "Configuring build with coverage enabled..."
cmake .. -DCMAKE_BUILD_TYPE=Debug -DENABLE_COVERAGE=ON

echo "Building with coverage..."
make -j$(nproc)

echo "Running tests..."
if [ -f "tests/object_detection_tests" ]; then
    ./tests/object_detection_tests
elif [ -f "object_detection_tests" ]; then
    ./object_detection_tests
else
    echo "Error: Test executable not found!"
    exit 1
fi

echo "Generating coverage reports..."

# Generate lcov report
lcov --capture --directory . --output-file coverage.info
lcov --remove coverage.info '/usr/*' '*/gtest/*' '*/gmock/*' --output-file coverage_filtered.info

# Generate HTML report
genhtml coverage_filtered.info --output-directory coverage_html

# Generate console report
echo ""
echo "=== Coverage Summary ==="
lcov --summary coverage_filtered.info

# Generate detailed gcovr report
echo ""
echo "=== Detailed Coverage Report ==="
gcovr --exclude '/usr/.*' --exclude '.*/gtest/.*' --exclude '.*/gmock/.*' --print-summary

# Generate JSON reports for tooling
gcovr --json coverage.json --exclude '/usr/.*' --exclude '.*/gtest/.*' --exclude '.*/gmock/.*'
gcovr --json-summary coverage_summary.json --exclude '/usr/.*' --exclude '.*/gtest/.*' --exclude '.*/gmock/.*'

echo ""
echo "=== Coverage Reports Generated ==="
echo "HTML Report: file://$(pwd)/coverage_html/index.html"
echo "JSON Report: $(pwd)/coverage.json"
echo "JSON Summary: $(pwd)/coverage_summary.json"

# Extract coverage percentage for easy reading
COVERAGE_PERCENT=$(gcovr --exclude '/usr/.*' --exclude '.*/gtest/.*' --exclude '.*/gmock/.*' --print-summary | grep "TOTAL" | awk '{print $4}')
echo "Overall Coverage: $COVERAGE_PERCENT"

# Check if coverage meets minimum threshold
MIN_COVERAGE=70
COVERAGE_NUM=$(echo $COVERAGE_PERCENT | sed 's/%//')
if (( $(echo "$COVERAGE_NUM >= $MIN_COVERAGE" | bc -l) )); then
    echo "✅ Coverage meets minimum threshold of $MIN_COVERAGE%"
    exit 0
else
    echo "❌ Coverage ($COVERAGE_PERCENT) is below minimum threshold of $MIN_COVERAGE%"
    echo "Please add more tests to improve coverage."
    exit 1
fi