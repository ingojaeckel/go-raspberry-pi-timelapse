#!/bin/bash
# Object Detection Installation Script for Raspberry Pi
# This script installs the dependencies needed for YOLO object detection via Python

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

echo "Step 1: Installing Python dependencies..."
echo "----------------------------------------"

# Update package list
apt-get update

# Install Python 3 and pip if not already installed
apt-get install -y python3 python3-pip

# Install OpenCV and NumPy for Python
echo "Installing OpenCV and NumPy (this may take a few minutes)..."
apt-get install -y python3-opencv python3-numpy

# Verify installation
python3 -c "import cv2; import numpy as np; print('OpenCV and NumPy installed successfully')" || {
    echo "Error: Failed to install Python dependencies"
    exit 1
}

echo "Python dependencies installed successfully"
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
echo "Step 4: Installing detection script..."
echo "----------------------------------------"

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
YOLO_SCRIPT="${SCRIPT_DIR}/yolo_detect.py"

if [ ! -f "$YOLO_SCRIPT" ]; then
    echo "Error: yolo_detect.py not found in $SCRIPT_DIR"
    echo "Please ensure this script is run from the scripts/ directory"
    exit 1
fi

# Install the detection script
cp "$YOLO_SCRIPT" /usr/local/bin/yolo_detect.py
chmod +x /usr/local/bin/yolo_detect.py
# Create symlink without extension for convenience
ln -sf /usr/local/bin/yolo_detect.py /usr/local/bin/yolo_detect

echo "Detection script installed to /usr/local/bin/yolo_detect.py"

echo ""
echo "=================================================="
echo "Installation Complete!"
echo "=================================================="
echo ""
echo "Summary:"
echo "  - Python dependencies: Installed"
echo "  - YOLO model: $MODEL_PATH"
echo "  - Detection script: /usr/local/bin/yolo_detect.py"
echo ""
echo "Next steps:"
echo "  1. Enable object detection in the web interface"
echo "  2. Restart the timelapse application"
echo "  3. Check logs for detection results"
echo ""
echo "To manually test detection on an image:"
echo "  /usr/local/bin/yolo_detect.py --image /path/to/image.jpg --model $MODEL_PATH --json"
echo ""
echo "For more information, see docs/OBJECT_DETECTION.md"
echo "=================================================="
