#!/bin/bash
#
# Download YOLO model files for offline object detection
# This script downloads YOLOv3-tiny which is optimized for Raspberry Pi
#

set -e

MODELS_DIR="/opt/yolo"
TEMP_DIR="/tmp/yolo_download"

echo "Setting up YOLO models for offline object detection..."

# Create directories
sudo mkdir -p "$MODELS_DIR"
mkdir -p "$TEMP_DIR"

cd "$TEMP_DIR"

# Download YOLOv3-tiny (smaller, faster, suitable for RPi Zero)
echo "Downloading YOLOv3-tiny weights..."
wget -q --no-check-certificate https://pjreddie.com/media/files/yolov3-tiny.weights -O yolov3-tiny.weights

echo "Downloading YOLOv3-tiny config..."
wget -q --no-check-certificate https://raw.githubusercontent.com/pjreddie/darknet/master/cfg/yolov3-tiny.cfg -O yolov3-tiny.cfg

echo "Downloading COCO class names..."
wget -q --no-check-certificate https://raw.githubusercontent.com/pjreddie/darknet/master/data/coco.names -O coco.names

# Optional: Download full YOLOv3 if space permits (much larger but more accurate)
if [[ "${DOWNLOAD_FULL_YOLO}" == "1" ]]; then
    echo "Downloading full YOLOv3 weights (this may take several minutes)..."
    wget -q --no-check-certificate https://pjreddie.com/media/files/yolov3.weights -O yolov3.weights
    
    echo "Downloading full YOLOv3 config..."
    wget -q --no-check-certificate https://raw.githubusercontent.com/pjreddie/darknet/master/cfg/yolov3.cfg -O yolov3.cfg
fi

# Move files to final location
echo "Installing model files..."
sudo mv *.weights *.cfg *.names "$MODELS_DIR/"

# Set permissions
sudo chmod 644 "$MODELS_DIR"/*

# Cleanup
rm -rf "$TEMP_DIR"

echo "YOLO models installed successfully in $MODELS_DIR"
echo "Available models:"
ls -la "$MODELS_DIR"

# Test the installation
echo "Testing OpenCV installation..."
python3 -c "import cv2; print(f'OpenCV version: {cv2.__version__}')" || {
    echo "WARNING: OpenCV not properly installed"
    exit 1
}

echo "YOLO setup complete!"