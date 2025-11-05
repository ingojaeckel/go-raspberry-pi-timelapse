# Scene Graph Detector Models

This directory contains ONNX models and label files for the scene graph detector.

## Required Files

### 1. Object Detector Model

**Recommended: YOLOv5s** (lightweight, good balance)
- **Size**: ~14 MB
- **Download**: 
  ```bash
  wget https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx
  ```
- **Checksum (SHA256)**: 
  ```
  # Verify with: sha256sum yolov5s.onnx
  ```

**Alternative: YOLOv5n** (ultra-lightweight)
- **Size**: ~4 MB
- **Download**:
  ```bash
  wget https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5n.onnx
  ```

**Alternative: YOLOv5l** (high accuracy)
- **Size**: ~47 MB
- **Download**:
  ```bash
  wget https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5l.onnx
  ```

### 2. Class Labels

**COCO Labels** (80 classes)
- File: `coco.names` (provided in this directory)
- Classes: person, bicycle, car, motorcycle, airplane, bus, train, truck, boat, etc.

Create `coco.names` with one class per line:
```
person
bicycle
car
motorcycle
...
```

### 3. Relation Predictor Model (Optional)

Currently, geometric relation inference is used by default. Support for ONNX-based relation models is available but requires custom-trained models.

**Training your own relation model:**
- Input: Pairs of object bounding boxes + features
- Output: Predicate classification (11 classes)
- Framework: PyTorch → ONNX export

## Quick Setup

```bash
cd cpp-scene-graph-detector/assets/models

# Download YOLOv5s
wget https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx

# COCO labels are already included as coco.names

# Verify
ls -lh yolov5s.onnx
cat coco.names | wc -l  # Should be 80 lines
```

## Model Specifications

### YOLOv5 Family

| Model | Size | Input | Classes | Use Case |
|-------|------|-------|---------|----------|
| YOLOv5n | 4 MB | 640×640 | 80 (COCO) | Embedded, real-time |
| YOLOv5s | 14 MB | 640×640 | 80 (COCO) | Balanced (recommended) |
| YOLOv5m | 40 MB | 640×640 | 80 (COCO) | Higher accuracy |
| YOLOv5l | 47 MB | 640×640 | 80 (COCO) | Maximum accuracy |

### Custom Models

To use custom YOLO models:
1. Train model in PyTorch
2. Export to ONNX format
3. Ensure input shape is 640×640
4. Update `coco.names` with your class labels

**ONNX Export Example (PyTorch):**
```python
import torch
from ultralytics import YOLO

# Load trained model
model = YOLO('your_model.pt')

# Export to ONNX
model.export(format='onnx', imgsz=640)
```

## Checksums

Verify downloaded models:

```bash
# YOLOv5s
sha256sum yolov5s.onnx
# Expected: (varies by version, check release notes)

# YOLOv5n
sha256sum yolov5n.onnx
# Expected: (varies by version)
```

## Troubleshooting

**"Failed to load ONNX model"**
- Verify file is not corrupted (re-download)
- Check OpenCV was compiled with ONNX support
- Ensure model is compatible ONNX opset (9-14)

**"Class ID out of range"**
- Ensure `coco.names` has correct number of classes
- Match model training classes with label file

**"Poor detection quality"**
- Try higher confidence threshold (`--threshold.obj 0.5`)
- Use larger model (yolov5s → yolov5l)
- Ensure input images are good quality

## License

YOLO models are released under GPL-3.0 license by Ultralytics.
Check individual model repositories for specific license terms.
