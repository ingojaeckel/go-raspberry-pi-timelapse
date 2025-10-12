#!/bin/bash
# Object Detection Installation Script for Raspberry Pi
# This script installs the dependencies needed for YOLO object detection using GoCV

set -e

echo "=================================================="
echo "YOLO Object Detection Setup for Timelapse Camera"
echo "=================================================="
echo ""

# Check if running on Linux
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    echo "Error: This script is designed for Linux systems (Raspberry Pi OS)"
    exit 1
fi

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    echo "Please run this script with sudo:"
    echo "  sudo bash $0"
    exit 1
fi

echo "Step 1: Installing OpenCV dependencies..."
echo "----------------------------------------"

# Update package list
apt-get update

# Install OpenCV development libraries
echo "Installing OpenCV (this may take a few minutes)..."
apt-get install -y libopencv-dev pkg-config

# Verify OpenCV installation
pkg-config --modversion opencv4 &>/dev/null || {
    echo "Error: Failed to install OpenCV"
    echo "Trying to install opencv (version 3)..."
    apt-get install -y libopencv-dev
    pkg-config --modversion opencv &>/dev/null || {
        echo "Error: Failed to install OpenCV"
        exit 1
    }
}

echo "OpenCV installed successfully"
echo ""
echo "Step 2: Setting up YOLO model directory..."
echo "----------------------------------------"

# Create model directory
MODEL_DIR="/usr/local/share/yolo"
mkdir -p "$MODEL_DIR"

echo "Model directory created: $MODEL_DIR"

echo ""
echo "Step 3: Downloading YOLO model..."
echo "----------------------------------------"

# Determine which model to download based on system
TOTAL_MEM=$(grep MemTotal /proc/meminfo | awk '{print $2}')
TOTAL_MEM_MB=$((TOTAL_MEM / 1024))

echo "Detected system memory: ${TOTAL_MEM_MB} MB"

if [ $TOTAL_MEM_MB -lt 1024 ]; then
    MODEL="yolov5n"
    echo "Downloading YOLOv5 Nano (smallest, fastest) - recommended for systems with <1GB RAM"
elif [ $TOTAL_MEM_MB -lt 2048 ]; then
    MODEL="yolov5n"
    echo "Downloading YOLOv5 Nano - recommended for systems with 1-2GB RAM"
elif [ $TOTAL_MEM_MB -lt 4096 ]; then
    MODEL="yolov5s"
    echo "Downloading YOLOv5 Small - recommended for systems with 2-4GB RAM"
else
    MODEL="yolov5s"
    echo "Downloading YOLOv5 Small (default) - recommended for systems with 4GB+ RAM"
fi

MODEL_URL="https://github.com/ultralytics/yolov5/releases/download/v7.0/${MODEL}.onnx"
MODEL_PATH="${MODEL_DIR}/${MODEL}.onnx"

if [ -f "$MODEL_PATH" ]; then
    echo "Model already exists at $MODEL_PATH"
    read -p "Do you want to re-download? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Skipping download"
    else
        wget -O "$MODEL_PATH" "$MODEL_URL" || {
            echo "Error: Failed to download model"
            exit 1
        }
    fi
else
    wget -O "$MODEL_PATH" "$MODEL_URL" || {
        echo "Error: Failed to download model"
        exit 1
    }
fi

# Create symlink to default model name if using non-standard model
if [ "$MODEL" != "yolov5s" ]; then
    ln -sf "$MODEL_PATH" "${MODEL_DIR}/yolov5s.onnx"
    echo "Created symlink: ${MODEL_DIR}/yolov5s.onnx -> $MODEL_PATH"
fi

echo "Model downloaded successfully"

echo ""
echo "Step 4: Installation complete!"
echo "----------------------------------------"

echo ""
echo "=================================================="
echo "Installation Complete!"
echo "=================================================="
echo ""
echo "Summary:"
echo "  - OpenCV: Installed"
echo "  - YOLO model: $MODEL_PATH"
echo ""
echo "Next steps:"
echo "  1. Rebuild the timelapse application (it will now include GoCV support)"
echo "  2. Enable object detection in the web interface"
echo "  3. Restart the timelapse application"
echo "  4. Check logs for detection results"
echo ""
echo "For more information, see docs/OBJECT_DETECTION.md"
echo "=================================================="
