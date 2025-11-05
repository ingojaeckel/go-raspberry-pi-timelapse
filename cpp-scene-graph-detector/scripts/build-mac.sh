#!/bin/bash
set -e

echo "Building C++ Scene Graph Detector for macOS (Intel)..."

# Create build directory
mkdir -p build
cd build

# Configure for macOS
cmake .. -DCMAKE_BUILD_TYPE=Release

# Build
make -j$(sysctl -n hw.ncpu)

echo ""
echo "Build completed successfully!"
echo "Binary: build/cpp-scene-graph-detector"
echo ""
echo "To build with OpenCL support:"
echo "  cmake .. -DCMAKE_BUILD_TYPE=Release -DSCENE_GRAPH_WITH_OPENCL=ON"
echo "  make -j$(sysctl -n hw.ncpu)"
