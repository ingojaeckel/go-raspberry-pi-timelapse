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

**Ubuntu 22.04 (AMD64):**
```bash
sudo apt-get update
sudo apt-get install -y cmake build-essential libopencv-dev pkg-config libgtest-dev
```

**macOS (Intel):**
```bash
brew install cmake opencv pkg-config googletest
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

### Run

```bash
# Start real-time webcam detection with preview
./build/cpp-scene-graph-detector \
  --camera-id 0 \
  --model.detector assets/models/yolov5s.onnx \
  --labels assets/labels/coco.txt \
  --output-dir captures/
```

The application will:
- Show a live preview window with bounding boxes and scene description
- Log detected objects and spatial relations to console
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

**With OpenCL (optional):**
```bash
cmake .. -DSCENE_GRAPH_WITH_OPENCL=ON
make -j$(nproc)
```

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
