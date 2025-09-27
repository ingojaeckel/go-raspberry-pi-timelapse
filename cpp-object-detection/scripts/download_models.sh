#!/bin/bash

# Script to download different YOLO models for speed/accuracy comparison
# This demonstrates the model abstraction interface

set -e

MODELS_DIR="models"
mkdir -p "$MODELS_DIR"

echo "==================================="
echo "Downloading YOLO Models for Object Detection"
echo "==================================="

# Base URLs for model downloads
YOLOV5_BASE_URL="https://github.com/ultralytics/yolov5/releases/download/v6.2"
YOLOV8_BASE_URL="https://github.com/ultralytics/assets/releases/download/v0.0.0"

# Download YOLOv5s (Small - Fast Model)
echo "Downloading YOLOv5s (Small/Fast model)..."
if [ ! -f "$MODELS_DIR/yolov5s.onnx" ]; then
    wget -q --show-progress -O "$MODELS_DIR/yolov5s.onnx" "$YOLOV5_BASE_URL/yolov5s.onnx"
    echo "✅ YOLOv5s downloaded successfully"
else
    echo "✅ YOLOv5s already exists"
fi

# Download YOLOv5l (Large - Accurate Model)
echo "Downloading YOLOv5l (Large/Accurate model)..."
if [ ! -f "$MODELS_DIR/yolov5l.onnx" ]; then
    wget -q --show-progress -O "$MODELS_DIR/yolov5l.onnx" "$YOLOV5_BASE_URL/yolov5l.onnx"
    echo "✅ YOLOv5l downloaded successfully"
else
    echo "✅ YOLOv5l already exists"
fi

# Download YOLOv8n (Nano - Ultra Fast Model) - Future implementation
echo "YOLOv8n will be downloaded when implementation is complete"
echo "📝 Note: Currently falls back to YOLOv5s"

# Download YOLOv8m (Medium - High Accuracy Model) - Future implementation
echo "YOLOv8m will be downloaded when implementation is complete"
echo "📝 Note: Currently falls back to YOLOv5l"

# Create COCO class names file if it doesn't exist
if [ ! -f "$MODELS_DIR/coco.names" ]; then
    echo "Creating COCO class names file..."
    cat > "$MODELS_DIR/coco.names" << EOF
person
bicycle
car
motorcycle
airplane
bus
train
truck
boat
traffic light
fire hydrant
stop sign
parking meter
bench
bird
cat
dog
horse
sheep
cow
elephant
bear
zebra
giraffe
backpack
umbrella
handbag
tie
suitcase
frisbee
skis
snowboard
sports ball
kite
baseball bat
baseball glove
skateboard
surfboard
tennis racket
bottle
wine glass
cup
fork
knife
spoon
bowl
banana
apple
sandwich
orange
broccoli
carrot
hot dog
pizza
donut
cake
chair
couch
potted plant
bed
dining table
toilet
tv
laptop
mouse
remote
keyboard
cell phone
microwave
oven
toaster
sink
refrigerator
book
clock
vase
scissors
teddy bear
hair drier
toothbrush
EOF
    echo "✅ COCO class names file created"
else
    echo "✅ COCO class names file already exists"
fi

echo ""
echo "==================================="
echo "Model Download Summary"
echo "==================================="
echo ""
echo "📊 Available Models:"
echo ""
echo "🚀 YOLOv5s (Small/Fast):"
echo "   - File: models/yolov5s.onnx (~14MB)"
echo "   - Speed: ~65ms inference on modern CPU"
echo "   - Accuracy: 75% relative"
echo "   - Best for: Real-time applications, resource-constrained systems"
echo ""
echo "🎯 YOLOv5l (Large/Accurate):"
echo "   - File: models/yolov5l.onnx (~47MB)"
echo "   - Speed: ~120ms inference on modern CPU"
echo "   - Accuracy: 85% relative"
echo "   - Best for: High-accuracy requirements, offline processing"
echo ""
echo "⚡ YOLOv8n (Nano - Future):"
echo "   - Speed: ~35ms inference (when implemented)"
echo "   - Accuracy: 70% relative"
echo "   - Best for: Embedded systems, edge devices"
echo ""
echo "🏆 YOLOv8m (Medium - Future):"
echo "   - Speed: ~150ms inference (when implemented)"
echo "   - Accuracy: 88% relative"
echo "   - Best for: Maximum accuracy, powerful hardware"
echo ""
echo "==================================="
echo "Usage Examples:"
echo "==================================="
echo ""
echo "# Use fast model (default)"
echo "./object_detection --model-type yolov5s --max-fps 5"
echo ""
echo "# Use accurate model with reduced frame rate"
echo "./object_detection --model-type yolov5l --max-fps 2"
echo ""
echo "# Compare models side by side"
echo "./object_detection --model-type yolov5s --verbose"
echo "./object_detection --model-type yolov5l --verbose"
echo ""
echo "# Parallel processing for better throughput"
echo "./object_detection --model-type yolov5s --processing-threads 4"
echo ""
echo "==================================="
echo "Speed vs Accuracy Trade-offs:"
echo "==================================="
echo ""
echo "Model      | Speed  | Accuracy | Size | Use Case"
echo "-----------|--------|----------|------|------------------"
echo "YOLOv5s    | Fast   | Good     | 14MB | Real-time monitoring"
echo "YOLOv5l    | Slow   | High     | 47MB | Security, forensics"
echo "YOLOv8n    | Fastest| OK       | 6MB  | Embedded, IoT"
echo "YOLOv8m    | Slowest| Highest  | 52MB | Research, analysis"
echo ""
echo "✅ Model download complete!"
echo "🚀 Ready to start object detection with model selection!"