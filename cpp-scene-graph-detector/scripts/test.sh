#!/bin/bash
set -e

echo "Running C++ Scene Graph Detector tests..."

cd build

# Run unit tests if they exist
if [ -f "tests/scene_graph_tests" ]; then
    echo "Running unit tests..."
    ./tests/scene_graph_tests
    echo "Unit tests passed!"
else
    echo "Unit tests not found. Build with -DSCENE_GRAPH_BUILD_TESTS=ON to enable tests."
fi

echo "All tests completed!"
