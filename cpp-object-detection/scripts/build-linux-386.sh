#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building C++ Object Detection Application for 32-bit Linux (386)${NC}"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Create build directory
BUILD_DIR="$PROJECT_ROOT/build-linux-386"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Check for required dependencies
echo -e "${YELLOW}Checking dependencies...${NC}"

# Check for multilib support
if ! dpkg -l | grep -q gcc-multilib; then
    echo -e "${RED}Error: gcc-multilib not found. Please install it:${NC}"
    echo -e "${YELLOW}sudo apt-get install -y gcc-multilib g++-multilib${NC}"
    exit 1
fi

echo -e "${GREEN}Multilib support found${NC}"

# Check for CMake
if ! command -v cmake &> /dev/null; then
    echo -e "${RED}Error: CMake not found. Please install CMake.${NC}"
    exit 1
fi

echo -e "${GREEN}CMake found: $(cmake --version | head -n1)${NC}"

# Check for 32-bit OpenCV libraries
echo -e "${YELLOW}Checking for 32-bit OpenCV libraries...${NC}"
if [ -d "/usr/lib/i386-linux-gnu" ]; then
    if find /usr/lib/i386-linux-gnu -name "libopencv*.so*" 2>/dev/null | grep -q .; then
        echo -e "${GREEN}32-bit OpenCV libraries found${NC}"
    else
        echo -e "${YELLOW}Warning: 32-bit OpenCV libraries not found in /usr/lib/i386-linux-gnu${NC}"
        echo -e "${YELLOW}You may need to install: libopencv-dev:i386${NC}"
        echo -e "${YELLOW}Note: This may require multiarch setup${NC}"
    fi
fi

# Configure build for 32-bit
echo -e "${YELLOW}Configuring 32-bit build...${NC}"
cmake .. \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_SYSTEM_PROCESSOR=i386 \
    -DCMAKE_C_FLAGS="-m32" \
    -DCMAKE_CXX_FLAGS="-m32" \
    -DCMAKE_EXE_LINKER_FLAGS="-m32 -static-libgcc -static-libstdc++"

# Build
echo -e "${YELLOW}Building...${NC}"
make -j$(nproc)

echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}Executable: $BUILD_DIR/object_detection${NC}"

# Verify it's 32-bit
if file "$BUILD_DIR/object_detection" | grep -q "32-bit"; then
    echo -e "${GREEN}✓ Verified: Binary is 32-bit${NC}"
else
    echo -e "${RED}✗ Warning: Binary may not be 32-bit${NC}"
    file "$BUILD_DIR/object_detection"
fi

# Check if model files exist
MODEL_DIR="$PROJECT_ROOT/models"
if [ ! -f "$MODEL_DIR/yolov5s.onnx" ]; then
    echo -e "${YELLOW}Warning: No YOLO model found at $MODEL_DIR/yolov5s.onnx${NC}"
    echo -e "${YELLOW}Download a model to use the application:${NC}"
    echo -e "${YELLOW}  wget -O $MODEL_DIR/yolov5s.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx${NC}"
fi

echo -e "${GREEN}Usage: $BUILD_DIR/object_detection --help${NC}"
echo -e "${YELLOW}Recommended settings for Pentium M (1.5GB RAM):${NC}"
echo -e "${YELLOW}  $BUILD_DIR/object_detection --max-fps 1 --min-confidence 0.8 --frame-width 640 --frame-height 480 --analysis-rate-limit 0.5${NC}"
