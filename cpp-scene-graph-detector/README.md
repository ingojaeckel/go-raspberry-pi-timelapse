# C++ Scene Graph Detector

A C++ component that detects objects in images/videos and produces **scene graphs** (nodes = objects, edges = spatial relations). Built to mirror the structure and conventions of `cpp-object-detection`.

## What is a Scene Graph?

A scene graph represents the semantic structure of a visual scene:
- **Nodes**: Detected objects with bounding boxes, labels, and confidence scores
- **Edges**: Spatial or semantic relations between objects (e.g., "left_of", "contains", "overlaps")

Scene graphs enable higher-level visual understanding beyond simple object detection.

## Features

- **Object Detection**: YOLO-family models via OpenCV DNN (ONNX format)
- **Relation Prediction**: 
  - Geometric inference (spatial predicates: left_of, right_of, overlaps, contains, etc.)
  - Optional ONNX relation model support
- **Multiple Backends**: CPU (default), OpenCL (optional)
- **Multiple Platforms**: macOS (Intel, 2018 MBP), Linux AMD64, ARM64
- **Output Formats**: JSON, Graphviz DOT, ASCII summary
- **Video Support**: Process video frames at configurable FPS
- **Visualization**: Overlay bounding boxes on images

## Supported Spatial Predicates

- `left_of`, `right_of` - Horizontal positioning
- `in_front_of`, `behind` - Depth ordering (model-based)
- `overlaps`, `intersects` - Geometric overlap
- `contains` - Containment relation
- `between` - Intermediate positioning
- `on`, `under` - Vertical positioning
- `next_to` - Proximity

## Building

### Prerequisites

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y cmake build-essential libopencv-dev pkg-config
```

**macOS (Intel):**
```bash
brew update
brew install cmake opencv pkg-config
```

**For testing (optional):**
```bash
# Ubuntu/Debian
sudo apt-get install -y libgtest-dev

# macOS
brew install googletest
```

### Build Instructions

```bash
cd cpp-scene-graph-detector

# Create build directory
mkdir -p build && cd build

# Configure
cmake .. -DCMAKE_BUILD_TYPE=Release

# Build
make -j$(nproc)  # Linux
make -j$(sysctl -n hw.ncpu)  # macOS
```

### Build with OpenCL Support

```bash
cmake .. -DCMAKE_BUILD_TYPE=Release -DSCENE_GRAPH_WITH_OPENCL=ON
make -j$(nproc)
```

### Cross-Platform Builds

**Using build scripts:**
```bash
# Auto-detect platform
./scripts/build.sh

# macOS specific
./scripts/build-mac.sh
```

## Usage

### Basic Usage

```bash
cpp-scene-graph-detector \
  --input path/to/image.jpg \
  --model.detector assets/models/yolov5s.onnx \
  --labels assets/coco.names \
  --out.json output/scene_graph.json
```

### Advanced Usage

```bash
cpp-scene-graph-detector \
  --input path/to/video.mp4 \
  --model.detector assets/models/yolov5s.onnx \
  --model.relations assets/models/relations.onnx \
  --labels assets/coco.names \
  --backend opencl \
  --threshold.obj 0.3 \
  --threshold.rel 0.5 \
  --max-objects 128 \
  --video-fps 2 \
  --out.json output/scene_graph.json \
  --out.dot output/scene_graph.dot \
  --visualize output/overlay.jpg \
  --verbose
```

### Command-Line Options

```
Required:
  --input PATH                   Input image or video file
  --model.detector PATH          Path to object detector ONNX model
  --labels PATH                  Path to class labels file

Optional:
  --model.relations PATH         Path to relations ONNX model
  --backend [cpu|opencl|auto]    Inference backend (default: cpu)
  --threshold.obj FLOAT          Object confidence threshold (default: 0.25)
  --threshold.rel FLOAT          Relation confidence threshold (default: 0.5)
  --max-objects INT              Maximum objects to detect (default: 128)
  --out.json PATH                Output JSON file path
  --out.dot PATH                 Output Graphviz DOT file path
  --visualize PATH               Output visualization image path
  --video-fps INT                Frames per second to process (default: 1)
  -v, --verbose                  Enable verbose logging
  -h, --help                     Show help message
```

## Models and Assets

Models should be placed in `assets/models/`. See `assets/models/README.md` for download instructions.

**Required:**
- Object detector: YOLO ONNX model (e.g., `yolov5s.onnx`)
- Labels file: Class names (e.g., `coco.names`)

**Optional:**
- Relations model: ONNX model for predicate prediction

### Downloading Models

```bash
cd assets/models

# YOLOv5s (lightweight, fast)
wget https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx

# COCO labels
wget https://raw.githubusercontent.com/ultralytics/yolov5/master/data/coco.yaml
# Extract class names to coco.names manually or use provided file
```

## Output Format

### JSON Schema

See [docs/SCENE_GRAPH_FORMAT.md](docs/SCENE_GRAPH_FORMAT.md) for complete schema and examples.

**Structure:**
```json
{
  "meta": {
    "num_objects": 3,
    "num_relations": 2
  },
  "objects": [
    {
      "id": 0,
      "label": "person",
      "class_id": 0,
      "score": 0.85,
      "bbox": { "x": 320, "y": 240, "width": 120, "height": 200 }
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

### Graphviz DOT

Generate visual graphs with:
```bash
dot -Tpng output/scene_graph.dot -o output/scene_graph.png
```

## Testing

### Run Unit Tests

```bash
cd build
./tests/scene_graph_tests
```

### Run with Sample Data

```bash
# Process sample image
cpp-scene-graph-detector \
  --input samples/example.jpg \
  --model.detector models/yolov5s.onnx \
  --labels models/coco.names \
  --out.json samples/example_output.json \
  --verbose
```

## Performance

### Target Performance

| Platform | Hardware | Expected FPS @ 720p |
|----------|----------|---------------------|
| macOS Intel | 2018 MBP (i7) | ~1 FPS (CPU) |
| Linux AMD64 | Modern CPU | ~2-5 FPS (CPU) |
| Linux AMD64 | AMD GPU (OpenCL) | ~5-10 FPS |

### Optimization Tips

1. **Lower object threshold** for faster inference (fewer objects)
2. **Reduce max-objects** to limit processing overhead
3. **Use OpenCL backend** if GPU available
4. **Downsample video** before processing for better FPS

## GPU Acceleration

### OpenCL Support

**macOS (Intel):**
```bash
cmake .. -DSCENE_GRAPH_WITH_OPENCL=ON
make
cpp-scene-graph-detector --backend opencl ...
```

**Linux (AMD GPU):**
```bash
# Install OpenCL drivers first
sudo apt-get install ocl-icd-opencl-dev

cmake .. -DSCENE_GRAPH_WITH_OPENCL=ON
make
cpp-scene-graph-detector --backend opencl ...
```

**Note:** Runtime gracefully falls back to CPU if OpenCL is unavailable.

## Architecture

```
cpp-scene-graph-detector/
├── include/scene_graph/
│   ├── Graph.h          # Scene graph data structures
│   ├── Detector.h       # Object detection interface
│   ├── RelPredictor.h   # Relation prediction
│   ├── Runner.h         # Pipeline orchestrator
│   ├── config_manager.h # Configuration handling
│   └── logger.h         # Logging utilities
├── src/
│   ├── graph.cpp
│   ├── detector.cpp
│   ├── rel_predictor.cpp
│   ├── runner.cpp
│   ├── config_manager.cpp
│   ├── logger.cpp
│   └── main.cpp
├── tests/               # Unit tests
├── assets/
│   ├── models/          # ONNX models
│   └── labels.txt       # Class labels
├── samples/             # Sample images and outputs
├── docs/                # Documentation
└── scripts/             # Build scripts
```

## Troubleshooting

**"Failed to load ONNX model"**
- Verify model file exists and is valid ONNX format
- Check OpenCV DNN is compiled with ONNX support

**"OpenCL requested but not available"**
- Install OpenCL drivers
- Rebuild with `-DSCENE_GRAPH_WITH_OPENCL=ON`
- Use `--backend cpu` as fallback

**Low FPS**
- Lower `--threshold.obj` to reduce detections
- Use smaller YOLO model (yolov5n vs yolov5s)
- Process fewer video frames (`--video-fps 1`)

## License

This project follows the same license as the parent repository.

## Credits

- **OpenCV** for computer vision and DNN inference
- **YOLO** models for object detection
- Inspired by the structure of `cpp-object-detection`
