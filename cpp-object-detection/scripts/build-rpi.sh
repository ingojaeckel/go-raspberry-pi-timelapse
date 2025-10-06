#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building C++ Object Detection Application for Raspberry Pi (ARM64)${NC}"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Create build directory
BUILD_DIR="$PROJECT_ROOT/build-rpi"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Check for required dependencies
echo -e "${YELLOW}Checking dependencies...${NC}"

# Check for CMake
if ! command -v cmake &> /dev/null; then
    echo -e "${RED}Error: CMake not found. Please install CMake.${NC}"
    exit 1
fi

echo -e "${GREEN}CMake found: $(cmake --version | head -n1)${NC}"

# Check for OpenCV libraries
echo -e "${YELLOW}Checking for OpenCV libraries...${NC}"
OPENCV_FOUND=false
for libdir in /usr/lib/aarch64-linux-gnu /usr/lib/arm-linux-gnueabihf /usr/lib; do
    if [ -d "$libdir" ]; then
        if find "$libdir" -name "libopencv*.so*" 2>/dev/null | grep -q .; then
            echo -e "${GREEN}OpenCV libraries found in $libdir${NC}"
            OPENCV_FOUND=true
            break
        fi
    fi
done

if [ "$OPENCV_FOUND" = false ]; then
    echo -e "${RED}Error: OpenCV libraries not found. Please install OpenCV development libraries:${NC}"
    echo -e "${YELLOW}sudo apt-get install -y libopencv-dev${NC}"
    exit 1
fi

echo -e "${GREEN}OpenCV libraries found${NC}"

# Check for libcurl
if ! dpkg -l | grep -q libcurl4-openssl-dev; then
    echo -e "${YELLOW}Warning: libcurl4-openssl-dev not found${NC}"
    echo -e "${YELLOW}Note: This build will likely fail without libcurl development libraries${NC}"
else
    echo -e "${GREEN}libcurl development libraries found${NC}"
fi

# Configure build for Raspberry Pi
echo -e "${YELLOW}Configuring Raspberry Pi build...${NC}"

cmake .. \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_EXE_LINKER_FLAGS="-static-libgcc -static-libstdc++"

if [ $? -ne 0 ]; then
    echo -e "${RED}CMake configuration failed${NC}"
    exit 1
fi

echo -e "${GREEN}Configuration successful${NC}"

# Build
echo -e "${YELLOW}Building application...${NC}"

# Get number of CPU cores
NUM_CORES=$(nproc 2>/dev/null || echo 4)
echo -e "${YELLOW}Building with $NUM_CORES cores${NC}"

make -j$NUM_CORES

if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed${NC}"
    exit 1
fi

echo -e "${GREEN}Build successful!${NC}"

# Verify binary
echo -e "${YELLOW}Verifying binary...${NC}"

if [ -f "object_detection" ]; then
    echo -e "${GREEN}Binary created successfully${NC}"
    file object_detection
    ls -lh object_detection
    
    # Show dependencies
    echo -e "${YELLOW}Binary dependencies:${NC}"
    ldd object_detection || echo "Note: ldd may not work for all configurations"
else
    echo -e "${RED}Error: object_detection binary not found after build${NC}"
    exit 1
fi

echo -e "${GREEN}===============================================${NC}"
echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}===============================================${NC}"
echo -e ""
echo -e "${YELLOW}Binary location: $BUILD_DIR/object_detection${NC}"
echo -e "${YELLOW}To test: cd $BUILD_DIR && ./object_detection --help${NC}"
echo -e ""
echo -e "${YELLOW}Recommended settings for Raspberry Pi 4/5:${NC}"
echo -e "${YELLOW}  ./object_detection --max-fps 5 --min-confidence 0.6 --detection-scale 0.5${NC}"
echo -e ""
echo -e "${YELLOW}For battery-powered operation:${NC}"
echo -e "${YELLOW}  ./object_detection --max-fps 2 --min-confidence 0.7 --analysis-rate-limit 0.5 --detection-scale 0.5${NC}"
