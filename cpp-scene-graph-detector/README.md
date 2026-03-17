# C++ Scene Graph Detector

A standalone C++ application for real-time webcam monitoring with object detection and spatial relation inference.

## Overview

This application provides continuous environmental observation by analyzing webcam feeds in real-time. It combines object detection (using YOLO models via OpenCV DNN) with spatial relation prediction to create scene graphs that describe "what is where" in the current view.

**Scene Graph**: A structured representation where:
- **Nodes** = Detected objects (person, car, tree, house, etc.)
- **Edges** = Spatial relationships (left_of, on, contains, etc.)

## Features

- ✅ **Real-time Webcam Monitoring** with live preview
- ✅ **Object Detection** via ONNX models (YOLOv5, etc.)
- ✅ **Spatial Relation Prediction** (11 geometric predicates)
- ✅ **Bounding Box Overlay** with labels and confidence scores
- ✅ **Scene Description Overlay** showing detected relationships
- ✅ **Automatic Photo Capture** when scene changes (new objects detected)
- ✅ **Console Logging** of detected objects and relations
- ✅ **Multi-backend** support (CPU, OpenCL)
- ✅ **Cross-platform** (macOS Intel, Linux AMD64)

## Quick Start

### Prerequisites

**Ubuntu 22.04+ (AMD64 or ARM64):**
```bash
sudo apt-get update
sudo apt-get install -y cmake build-essential libopencv-dev pkg-config libgtest-dev
```

**macOS (Intel or Apple Silicon):**
```bash
brew install cmake opencv pkg-config googletest
```

**NVIDIA Jetson (Ubuntu 20.04/22.04 ARM64):**
```bash
# JetPack already includes OpenCV with CUDA support
sudo apt-get update
sudo apt-get install -y cmake build-essential pkg-config libgtest-dev

# Verify CUDA is available
nvcc --version
```

### Build

```bash
cd cpp-scene-graph-detector
mkdir build && cd build
cmake ..
make -j$(nproc)
```

### Download Model

```bash
# Download YOLOv5s (fast, recommended for real-time)
cd cpp-scene-graph-detector
curl -L -o assets/models/yolov5s.onnx \
  https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx

# OR download YOLOv5x (slower but more accurate)
# Better for offline processing or high-accuracy requirements
curl -L -o assets/models/yolov5x.onnx \
  https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5x.onnx
```

**Model Selection Guide:**
- **YOLOv5s**: ~65ms inference, 37.4 mAP - Best for real-time on 2018 MacBook Pro
- **YOLOv5m**: ~120ms inference, 45.4 mAP - Balanced accuracy/speed
- **YOLOv5l**: ~180ms inference, 49.0 mAP - High accuracy
- **YOLOv5x**: ~330ms inference, 50.7 mAP - **Maximum accuracy, slower** (recommended for offline batch processing)

**For Environmental Monitoring:**
- COCO models (above) detect: person, car, truck, bicycle, animals, but **not** trees, houses, sheds
- See [assets/models/README.md](assets/models/README.md) for:
  - Objects365 models (365 classes including trees, buildings, construction equipment)
  - Custom training guide for environmental-specific models
  - Pre-configured `assets/labels/environmental.txt` for custom models

### Run

```bash
# Start real-time webcam detection with preview
./build/cpp-scene-graph-detector \
  --camera-id 0 \
  --model.detector assets/models/yolov5s.onnx \
  --labels assets/labels/coco.txt \
  --output-dir captures/

# With verbose timing information
./build/cpp-scene-graph-detector \
  --camera-id 0 \
  --model.detector assets/models/yolov5s.onnx \
  --labels assets/labels/coco.txt \
  --output-dir captures/ \
  --verbose

# With OpenCL GPU acceleration (if available)
./build/cpp-scene-graph-detector \
  --camera-id 0 \
  --model.detector assets/models/yolov5s.onnx \
  --labels assets/labels/coco.txt \
  --backend opencl \
  --verbose
```

The application will:
- Show a live preview window with bounding boxes and scene description
- Display analysis time (in milliseconds) in the top-right corner of the preview
- Log detected objects and spatial relations to console
- With `--verbose`: Show detailed timing for each frame analysis
- With `backend=opencl` on macOS: Show warning if OpenCL fails and falls back to CPU
- Automatically save photos to `captures/` when the scene changes (new object types detected)
- Press ESC or 'q' to quit

## Usage

### Command Line Options

```
cpp-scene-graph-detector [OPTIONS]

Required:
  --model.detector PATH     Path to ONNX object detector model
  --labels PATH             Path to class labels file

Input:
  --camera-id N             Webcam device ID (default: 0)

Configuration:
  --backend TYPE            cpu|opencl|auto (default: cpu)
  --threshold.obj N         Object confidence threshold (default: 0.25)
  --threshold.rel N         Relation confidence threshold (default: 0.5)
  --max-objects N           Max objects to detect (default: 128)
  --output-dir PATH         Directory for scene change photos (default: output/)
  --fps N                   Processing FPS (default: 1)
```

### Examples

**Basic webcam monitoring (fast, real-time):**
```bash
./cpp-scene-graph-detector \
  --camera-id 0 \
  --model.detector models/yolov5s.onnx \
  --labels assets/labels/coco.txt
```
*Uses YOLOv5s for fast real-time detection at ~1 FPS on 2018 MacBook Pro CPU.*

**High-accuracy environmental monitoring:**
```bash
./cpp-scene-graph-detector \
  --camera-id 0 \
  --model.detector models/yolov5x.onnx \
  --labels assets/labels/coco.txt \
  --threshold.obj 0.5 \
  --output-dir environmental_captures/
```
*Uses YOLOv5x for maximum accuracy. Slower (~0.3 FPS) but better for detailed environmental observation.*

**GPU-accelerated webcam:**
```bash
./cpp-scene-graph-detector \
  --camera-id 0 \
  --model.detector models/yolov5s.onnx \
  --labels assets/labels/coco.txt \
  --backend opencl
```
*Leverages OpenCL for GPU acceleration (2-3x faster on compatible hardware).*

## Scene Graph Output

When a scene change is detected, the application:

1. **Logs to console**: Displays detected objects and spatial relations
2. **Saves a photo**: Captures the current frame with bounding box and scene description overlays to the output directory

Example console output:
```
Scene changed - saved outputs
Objects detected: 3
  - house (confidence: 0.92)
  - tree (confidence: 0.87)
  - car (confidence: 0.88)
Relations found: 2
  - tree left_of house
  - car right_of house
```

The saved photo includes:
- Green bounding boxes around detected objects
- Labels with class name and confidence score
- Scene description overlay at the bottom showing object count and key relationships

## Spatial Predicates

The system automatically infers the following spatial relationships:

| Predicate | Description | Example |
|-----------|-------------|---------|
| `left_of` | Object A is to the left of B | car left_of person |
| `right_of` | Object A is to the right of B | tree right_of house |
| `on` | Object A is on top of B | book on table |
| `under` | Object A is below B | dog under table |
| `overlaps` | Objects A and B overlap | person overlaps car |
| `contains` | Object A contains B | bowl contains fruit |
| `next_to` | Objects are adjacent | chair next_to table |

## Building and Testing

### Build with tests

```bash
mkdir build && cd build
cmake .. -DSCENE_GRAPH_BUILD_TESTS=ON
make -j$(nproc)
make test
```

### Platform-specific builds

**macOS (Intel):**
```bash
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(sysctl -n hw.ncpu)
```

**Linux (AMD64):**
```bash
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)
```

**NVIDIA Jetson (ARM64 with CUDA):**
```bash
# Build with CUDA-optimized OpenCV (included in JetPack)
cmake .. -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_CUDA_ARCHITECTURES=87  # Orin = 87, Xavier = 72
make -j$(nproc)

# For TensorRT optimization (best performance):
cmake .. -DCMAKE_BUILD_TYPE=Release \
  -DSCENE_GRAPH_WITH_TENSORRT=ON \
  -DCMAKE_CUDA_ARCHITECTURES=87
make -j$(nproc)
```

**With OpenCL (optional):**
```bash
cmake .. -DSCENE_GRAPH_WITH_OPENCL=ON
make -j$(nproc)
```

## Jetson Deployment Guide

### Why Jetson for Environmental Monitoring?

NVIDIA Jetson platforms are ideal for edge deployment of scene graph detection in environmental monitoring scenarios:

**Key Advantages:**
1. **Power Efficient**: 10-40W vs 45-87W for laptops (2-3x better)
2. **24/7 Operation**: Designed for always-on continuous deployment
3. **Fanless Options**: Jetson Orin Nano/NX can run passively cooled
4. **Industrial Grade**: -25°C to 80°C operating temperature range
5. **Direct Camera Interface**: CSI cameras bypass USB bottleneck
6. **Compact**: Fits in weatherproof enclosures for outdoor deployment
7. **High Performance**: 3-8x faster than 2018 MacBook Pro CPU

**Performance Comparison (YOLOv5s):**
- 2018 MacBook Pro CPU: ~65ms (15 FPS, 45-87W)
- Jetson Orin Nano: ~25-35ms (28-40 FPS, 10-15W) - **2x faster, 1/4 power**
- Jetson AGX Orin 64GB: ~8-12ms (83-125 FPS, 25-40W) - **5x faster, 50% power**

### Models That Work Well on Jetson (vs 2018 MBP)

**High-Accuracy Models at ~20 fps (Viable on Jetson, Not on MBP CPU):**

1. **YOLOv5l** (49.0 mAP):
   - Jetson AGX Orin 64GB: **22-28 fps** at 25-35W ✅
   - 2018 MBP CPU: **5.5 fps** at 45-87W ❌

2. **YOLOv5x** (50.7 mAP):
   - Jetson AGX Orin 64GB: **14-18 fps** at 30-40W ✅
   - 2018 MBP CPU: **3 fps** at 45-87W ❌

3. **YOLO-World** (Open Vocabulary, 45-52 mAP):
   - Jetson AGX Orin 64GB: **15-22 fps** at 25-40W ✅ **[Recommended]**
   - 2018 MBP CPU: **2-3 fps** at 45-87W ❌
   - **Detects custom objects via text prompts** ("oak tree", "metal shed") without retraining

See [assets/models/README.md](assets/models/README.md) for detailed Jetson performance specs.

### Jetson Setup

**1. Install JetPack SDK** (includes CUDA, cuDNN, TensorRT, OpenCV):
```bash
# JetPack 5.1.2 or newer (Ubuntu 20.04)
# Or JetPack 6.0+ (Ubuntu 22.04)
# Follow NVIDIA setup guide: https://developer.nvidia.com/jetpack
```

**2. Verify CUDA and OpenCV:**
```bash
nvcc --version  # Should show CUDA 11.4+ (JP 5.x) or 12.x (JP 6.x)
python3 -c "import cv2; print(cv2.getBuildInformation())" | grep -i cuda
# Should show "CUDA: YES"
```

**3. Build cpp-scene-graph-detector:**
```bash
cd cpp-scene-graph-detector
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_CUDA_ARCHITECTURES=87
make -j$(nproc)
```

**4. Download Model:**
```bash
curl -L -o assets/models/yolov5s.onnx \
  https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx
```

**5. Run with CSI Camera:**
```bash
# For CSI camera (Raspberry Pi Camera Module v2 or similar)
./build/cpp-scene-graph-detector \
  --camera-id 0 \
  --model.detector assets/models/yolov5s.onnx \
  --labels assets/labels/coco.txt \
  --backend auto \
  --verbose

# For USB webcam
./build/cpp-scene-graph-detector \
  --camera-id 1 \
  --model.detector assets/models/yolov5s.onnx \
  --labels assets/labels/coco.txt \
  --backend auto
```

### Power Management on Jetson

**Set Power Mode:**
```bash
# Maximum performance (Jetson AGX Orin)
sudo nvpmodel -m 0  # MAXN mode (60W)

# Balanced (30W)
sudo nvpmodel -m 2

# Low power (15W)
sudo nvpmodel -m 4

# Check current mode
sudo nvpmodel -q
```

**Monitor Power Consumption:**
```bash
# Install jetson-stats
sudo pip3 install jetson-stats

# Monitor in real-time
jtop
```

### Cross-Compilation (Optional)

To build on x86_64 for Jetson ARM64:

```bash
# Install cross-compiler
sudo apt-get install g++-aarch64-linux-gnu

# Cross-compile
cmake .. -DCMAKE_SYSTEM_NAME=Linux \
  -DCMAKE_SYSTEM_PROCESSOR=aarch64 \
  -DCMAKE_C_COMPILER=aarch64-linux-gnu-gcc \
  -DCMAKE_CXX_COMPILER=aarch64-linux-gnu-g++ \
  -DCMAKE_FIND_ROOT_PATH=/usr/aarch64-linux-gnu

make -j$(nproc)
```

**Note**: Cross-compilation requires ARM64 OpenCV libraries. Native compilation on Jetson is recommended.

## GPU/OpenCL Notes

### Linux
- Requires OpenCL runtime and ICD loader
- Install AMD/Intel OpenCL drivers for GPU support
- Falls back to CPU if OpenCL not available

### macOS
- OpenCL supported on Intel Macs
- Automatically uses integrated/discrete GPU
- No additional drivers needed

## Performance

Expected performance on 1280x720 images:

### Real-time Models (YOLOv5s)
| Platform | Backend | FPS | Model | mAP |
|----------|---------|-----|-------|-----|
| 2018 MacBook Pro (Intel) | CPU | ~1 FPS | YOLOv5s | 37.4 |
| 2018 MacBook Pro (Intel) | OpenCL | ~2-3 FPS | YOLOv5s | 37.4 |
| Ubuntu 22.04 AMD64 (8-core) | CPU | ~3-5 FPS | YOLOv5s | 37.4 |
| Ubuntu 22.04 AMD64 (AMD GPU) | OpenCL | ~8-12 FPS | YOLOv5s | 37.4 |

### High-Accuracy Models (for offline processing)
| Platform | Backend | FPS | Model | mAP | Use Case |
|----------|---------|-----|-------|-----|----------|
| 2018 MacBook Pro (Intel) | CPU | ~0.3 FPS | YOLOv5x | 50.7 | Maximum accuracy, batch processing |
| 2018 MacBook Pro (Intel) | OpenCL | ~0.6 FPS | YOLOv5x | 50.7 | High-quality offline analysis |
| Ubuntu 22.04 AMD64 (8-core) | CPU | ~1 FPS | YOLOv5x | 50.7 | Archival video processing |
| Ubuntu 22.04 AMD64 (AMD GPU) | OpenCL | ~3-7 FPS | YOLOv5x | 50.7 | GPU-accelerated high accuracy |

**Model Selection by Use Case:**
- **Real-time webcam**: YOLOv5s (fast, good for live monitoring)
- **Offline video analysis**: YOLOv5l or YOLOv5x (better accuracy, acceptable speed for batch processing)
- **Critical accuracy needs**: YOLOv5x (maximum quality, use with GPU or for low-FPS batch work)
- **Balanced**: YOLOv5m (middle ground between speed and accuracy)

## Project Structure

```
cpp-scene-graph-detector/
├── CMakeLists.txt           # Build configuration
├── README.md                # This file
├── include/
│   └── scene_graph/
│       ├── Graph.h          # Scene graph data structures
│       ├── Detector.h       # Object detector interface
│       ├── RelPredictor.h   # Relation predictor interface
│       └── Runner.h         # Pipeline orchestrator
├── src/
│   ├── main.cpp             # CLI entry point
│   ├── Graph.cpp            # Scene graph implementation
│   ├── Detector.cpp         # Object detection
│   ├── RelPredictor.cpp     # Relation prediction
│   └── Runner.cpp           # Pipeline runner
├── tests/                   # Unit tests (GTest)
├── assets/
│   ├── models/              # ONNX models (download separately)
│   └── labels/              # Class label files
├── docs/                    # Documentation
├── samples/                 # Sample images and outputs
└── scripts/                 # Build and test scripts
```

## Documentation

- [Scene Graph Format](docs/SCENE_GRAPH_FORMAT.md) - JSON schema and examples
- [Model README](assets/models/README.md) - Download instructions for ONNX models

## Dependencies

- **OpenCV** ≥ 4.8 (with DNN module)
- **CMake** ≥ 3.16
- **C++17** compiler (GCC, Clang, MSVC)
- **Google Test** (optional, for testing)

## License

See repository root for license information.

## Acknowledgments

- YOLOv5: https://github.com/ultralytics/yolov5
- OpenCV DNN: https://docs.opencv.org/4.x/d2/d58/tutorial_table_of_content_dnn.html
- COCO Dataset: https://cocodataset.org/
