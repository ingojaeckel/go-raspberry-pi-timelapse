#!/bin/bash
set -e

# Build script for cpp-scene-graph-detector
# Auto-detects platform and builds accordingly

echo "Building cpp-scene-graph-detector..."

# Detect platform
if [[ "$OSTYPE" == "darwin"* ]]; then
    PLATFORM="macos"
    NPROC=$(sysctl -n hw.ncpu)
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    PLATFORM="linux"
    NPROC=$(nproc)
else
    echo "Unsupported platform: $OSTYPE"
    exit 1
fi

echo "Platform: $PLATFORM"
echo "Processors: $NPROC"

# Create build directory
mkdir -p build
cd build

# Configure
echo "Configuring with CMake..."
cmake .. -DCMAKE_BUILD_TYPE=Release

# Build
echo "Building..."
make -j${NPROC}

echo "Build complete!"
echo "Binary: build/cpp-scene-graph-detector"

# Check if binary exists
if [ -f "cpp-scene-graph-detector" ]; then
    echo "✓ Build successful"
    ./cpp-scene-graph-detector --help || true
else
    echo "✗ Build failed - binary not found"
    exit 1
fi
