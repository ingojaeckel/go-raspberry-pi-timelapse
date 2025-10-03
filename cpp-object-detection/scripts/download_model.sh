#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Downloading YOLO Model for Object Detection${NC}"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MODEL_DIR="$PROJECT_ROOT/models"

# Create models directory if it doesn't exist
mkdir -p "$MODEL_DIR"

# Check if model already exists
MODEL_FILE="$MODEL_DIR/yolov5s.onnx"
if [ -f "$MODEL_FILE" ]; then
    echo -e "${YELLOW}Model file already exists: $MODEL_FILE${NC}"
    echo -e "${YELLOW}File size: $(du -h "$MODEL_FILE" | cut -f1)${NC}"
    read -p "Do you want to re-download? (y/N): " confirm
    if [[ $confirm != [yY] ]]; then
        echo -e "${GREEN}Using existing model file${NC}"
        exit 0
    fi
fi

# Download YOLOv5s model
echo -e "${YELLOW}Downloading YOLOv5s ONNX model (~16MB)...${NC}"
MODEL_URL="https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx"

if command -v wget &> /dev/null; then
    wget -O "$MODEL_FILE" "$MODEL_URL"
elif command -v curl &> /dev/null; then
    curl -L -o "$MODEL_FILE" "$MODEL_URL"
else
    echo -e "${RED}Error: Neither wget nor curl is available. Please install one of them.${NC}"
    echo -e "${YELLOW}Manual download: $MODEL_URL${NC}"
    exit 1
fi

# Verify download
if [ -f "$MODEL_FILE" ]; then
    MODEL_SIZE=$(stat -c%s "$MODEL_FILE" 2>/dev/null || stat -f%z "$MODEL_FILE")
    if [ "$MODEL_SIZE" -gt 10000000 ]; then  # At least 10MB
        echo -e "${GREEN}Model downloaded successfully!${NC}"
        echo -e "${GREEN}File: $MODEL_FILE${NC}"
        echo -e "${GREEN}Size: $(du -h "$MODEL_FILE" | cut -f1)${NC}"
        
        # Test if file is a valid ONNX model (basic check)
        if command -v file &> /dev/null; then
            FILE_TYPE=$(file "$MODEL_FILE")
            echo -e "${YELLOW}File type: $FILE_TYPE${NC}"
        fi
        
        echo -e "${GREEN}You can now run the object detection application:${NC}"
        echo -e "${GREEN}  $PROJECT_ROOT/build/object_detection --help${NC}"
    else
        echo -e "${RED}Error: Downloaded file seems too small (${MODEL_SIZE} bytes)${NC}"
        echo -e "${RED}Please try downloading again or check your internet connection${NC}"
        rm -f "$MODEL_FILE"
        exit 1
    fi
else
    echo -e "${RED}Error: Failed to download model file${NC}"
    exit 1
fi

echo -e "${GREEN}Model setup complete!${NC}"