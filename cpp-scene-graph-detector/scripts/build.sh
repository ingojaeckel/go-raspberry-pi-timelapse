#!/bin/bash
set -e

# Detect OS and build accordingly
OS="$(uname -s)"

echo "Building C++ Scene Graph Detector..."
echo "OS detected: $OS"

# Create build directory
mkdir -p build
cd build

# Configure based on OS
if [[ "$OS" == "Darwin" ]]; then
    echo "Configuring for macOS..."
    cmake .. -DCMAKE_BUILD_TYPE=Release
    make -j$(sysctl -n hw.ncpu)
elif [[ "$OS" == "Linux" ]]; then
    echo "Configuring for Linux..."
    cmake .. -DCMAKE_BUILD_TYPE=Release
    make -j$(nproc)
else
    echo "Unsupported OS: $OS"
    exit 1
fi

echo ""
echo "Build completed successfully!"
echo "Binary: build/cpp-scene-graph-detector"
echo ""
echo "Next steps:"
echo "1. Download YOLO model: cd assets/models && wget https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx"
echo "2. Run: ./build/cpp-scene-graph-detector --help"
