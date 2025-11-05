# C++ Scene Graph Detector

A standalone C++ application for detecting objects and spatial relationships in images and video, producing structured scene graphs.

## Overview

This application combines object detection (using YOLO models via OpenCV DNN) with spatial relation prediction to create scene graphs that describe "what is where" in an image.

**Scene Graph**: A structured representation where:
- **Nodes** = Detected objects (person, car, tree, etc.)
- **Edges** = Spatial relationships (left_of, on, contains, etc.)

## Features

- ✅ **Object Detection** via ONNX models (YOLOv5, etc.)
- ✅ **Spatial Relation Prediction** (geometric or learned)
- ✅ **JSON Export** with standardized schema
- ✅ **Graphviz DOT Export** for visualization
- ✅ **Real-time Preview** with bounding boxes and scene description
- ✅ **Webcam Support** with automatic scene change detection
- ✅ **Video Processing** with frame sampling
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
# Download YOLOv5s (fast, recommended)
cd cpp-scene-graph-detector
curl -L -o assets/models/yolov5s.onnx \
  https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx
```

### Run

```bash
# Process single image
./build/cpp-scene-graph-detector \
  --input samples/scene.jpg \
  --model.detector assets/models/yolov5s.onnx \
  --labels assets/labels/coco.txt \
  --out.json output.json \
  --visualize output.jpg

# Process webcam with preview
./build/cpp-scene-graph-detector \
  --camera-id 0 \
  --model.detector assets/models/yolov5s.onnx \
  --labels assets/labels/coco.txt \
  --out.json output/ \
  --show-preview
```

## Usage

### Command Line Options

```
cpp-scene-graph-detector [OPTIONS]

Required:
  --model.detector PATH     Path to ONNX object detector model
  --labels PATH             Path to class labels file

Input (one required):
  --input PATH              Input image or video file
  --camera-id N             Webcam device ID (default: 0)

Output:
  --out.json PATH           Output JSON file/directory
  --out.dot PATH            Output Graphviz DOT file (optional)
  --visualize PATH          Save visualization image (optional)

Configuration:
  --backend TYPE            cpu|opencl|auto (default: cpu)
  --threshold.obj N         Object confidence threshold (default: 0.25)
  --threshold.rel N         Relation confidence threshold (default: 0.5)
  --max-objects N           Max objects to detect (default: 128)
  --fps N                   Video/webcam FPS (default: 1)
  --show-preview            Show real-time preview window
```

### Examples

**High-accuracy image processing:**
```bash
./cpp-scene-graph-detector \
  --input image.jpg \
  --model.detector models/yolov5l.onnx \
  --labels assets/labels/coco.txt \
  --threshold.obj 0.5 \
  --out.json scene.json \
  --out.dot scene.dot \
  --visualize scene_viz.jpg
```

**Real-time webcam with OpenCL:**
```bash
./cpp-scene-graph-detector \
  --camera-id 0 \
  --model.detector models/yolov5s.onnx \
  --labels assets/labels/coco.txt \
  --backend opencl \
  --show-preview \
  --out.json captures/
```

**Video processing:**
```bash
./cpp-scene-graph-detector \
  --input video.mp4 \
  --model.detector models/yolov5m.onnx \
  --labels assets/labels/coco.txt \
  --fps 2 \
  --out.json output/
```

## Scene Graph Format

The application outputs JSON with the following structure:

```json
{
  "meta": {
    "timestamp": "2024-01-15 14:30:00",
    "image_width": "1280",
    "image_height": "720"
  },
  "objects": [
    {
      "id": 0,
      "class_id": 0,
      "label": "person",
      "score": 0.95,
      "bbox": {
        "x": 0.5,
        "y": 0.5,
        "width": 0.2,
        "height": 0.3
      }
    }
  ],
  "relations": [
    {
      "subject_id": 0,
      "predicate": "left_of",
      "object_id": 1,
      "score": 1.0
    }
  ]
}
```

See [docs/SCENE_GRAPH_FORMAT.md](docs/SCENE_GRAPH_FORMAT.md) for full specification.

## Spatial Predicates

The system supports the following spatial relationships:

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

| Platform | Backend | FPS | Notes |
|----------|---------|-----|-------|
| 2018 MacBook Pro (Intel) | CPU | ~1 FPS | YOLOv5s |
| 2018 MacBook Pro (Intel) | OpenCL | ~2-3 FPS | YOLOv5s |
| Ubuntu 22.04 AMD64 (8-core) | CPU | ~3-5 FPS | YOLOv5s |
| Ubuntu 22.04 AMD64 (AMD GPU) | OpenCL | ~8-12 FPS | YOLOv5s |

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
