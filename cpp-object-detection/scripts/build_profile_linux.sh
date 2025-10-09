#!/bin/bash
# Build script for Linux with profiling enabled
# Builds the application with debug symbols and optimizations for accurate profiling

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_DIR/build-profile"

echo "Building C++ Object Detection Application (Linux Profile Build)"
echo "================================================================"

# Check if we're on Linux
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    echo "Error: This script is for Linux only. Use build_profile_mac.sh for macOS."
    exit 1
fi

# Check dependencies
echo "Checking dependencies..."

# Check for cmake
if ! command -v cmake &> /dev/null; then
    echo "Error: cmake not found. Install with:"
    echo "  Ubuntu/Debian: sudo apt-get install cmake"
    echo "  CentOS/RHEL: sudo yum install cmake"
    exit 1
fi

# Check for g++
if ! command -v g++ &> /dev/null; then
    echo "Error: g++ not found. Install with:"
    echo "  Ubuntu/Debian: sudo apt-get install build-essential"
    echo "  CentOS/RHEL: sudo yum install gcc-c++"
    exit 1
fi

# Check for OpenCV
if ! pkg-config --exists opencv4 2> /dev/null && ! pkg-config --exists opencv 2> /dev/null; then
    echo "Error: OpenCV not found. Install with:"
    echo "  Ubuntu/Debian: sudo apt-get install libopencv-dev"
    echo "  CentOS/RHEL: sudo yum install opencv-devel"
    exit 1
fi

echo "Dependencies OK"
echo ""

# Create build directory
echo "Creating build directory: $BUILD_DIR"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Configure with CMake (RelWithDebInfo = optimizations + debug symbols)
echo "Configuring with CMake (RelWithDebInfo + profiling flags)..."
cmake .. \
    -DCMAKE_BUILD_TYPE=RelWithDebInfo \
    -DCMAKE_CXX_FLAGS="-fno-omit-frame-pointer -g" \
    -DCMAKE_EXPORT_COMPILE_COMMANDS=ON

# Build
NPROC=$(nproc 2>/dev/null || echo 2)
echo "Building with $NPROC cores..."
make -j$NPROC

echo ""
echo "✅ Build complete!"
echo "Binary location: $BUILD_DIR/object_detection"
echo ""
echo "Profile Build Configuration:"
echo "  - Build Type: RelWithDebInfo (optimized with debug symbols)"
echo "  - Frame Pointers: Enabled (for accurate call stacks)"
echo "  - Platform: Linux (64-bit Intel/AMD)"
echo ""
echo "Next steps:"
echo "  1. Run CPU profiling:    ./scripts/profile_cpu_linux.sh"
echo "  2. Run memory profiling: ./scripts/profile_memory_linux.sh"
echo "  3. See PROFILING.md for detailed profiling guide"
echo ""
