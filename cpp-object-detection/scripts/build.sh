#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building C++ Object Detection Application${NC}"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Create build directory
BUILD_DIR="$PROJECT_ROOT/build"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Check for required dependencies
echo -e "${YELLOW}Checking dependencies...${NC}"

# Check for OpenCV
if ! pkg-config --exists opencv4; then
    echo -e "${RED}Error: OpenCV 4 not found. Please install OpenCV development libraries.${NC}"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo -e "${YELLOW}macOS: brew install opencv${NC}"
    else
        echo -e "${YELLOW}Ubuntu/Debian: sudo apt-get install libopencv-dev${NC}"
        echo -e "${YELLOW}CentOS/RHEL: sudo yum install opencv-devel${NC}"
    fi
    exit 1
fi

echo -e "${GREEN}OpenCV found: $(pkg-config --modversion opencv4)${NC}"

# Check for CMake
if ! command -v cmake &> /dev/null; then
    echo -e "${RED}Error: CMake not found. Please install CMake.${NC}"
    exit 1
fi

echo -e "${GREEN}CMake found: $(cmake --version | head -n1)${NC}"

# Configure build
echo -e "${YELLOW}Configuring build...${NC}"
cmake .. -DCMAKE_BUILD_TYPE=Release

# Build
echo -e "${YELLOW}Building...${NC}"
if [[ "$OSTYPE" == "darwin"* ]]; then
    make -j$(sysctl -n hw.ncpu)
else
    make -j$(nproc)
fi

echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}Executable: $BUILD_DIR/object_detection${NC}"

# Check if model files exist
MODEL_DIR="$PROJECT_ROOT/models"
if [ ! -f "$MODEL_DIR/yolov5s.onnx" ]; then
    echo -e "${YELLOW}Warning: No YOLO model found at $MODEL_DIR/yolov5s.onnx${NC}"
    echo -e "${YELLOW}Download a model to use the application:${NC}"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo -e "${YELLOW}  curl -L -o $MODEL_DIR/yolov5s.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx${NC}"
    else
        echo -e "${YELLOW}  wget -O $MODEL_DIR/yolov5s.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx${NC}"
    fi
fi

echo -e "${GREEN}Usage: $BUILD_DIR/object_detection --help${NC}"