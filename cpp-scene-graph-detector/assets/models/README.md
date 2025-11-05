# Scene Graph Detector Models

This directory contains ONNX models for object detection and relation prediction.

## Required Models

### Object Detector (Required)

Download a YOLOv5 or similar ONNX model for object detection:

**Option 1: YOLOv5s (Fast, recommended for real-time)**
```bash
curl -L -o detector.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx
# Inference: ~65ms on 2018 MacBook Pro CPU
# Accuracy: 37.4 mAP on COCO
# Best for: Real-time webcam processing
```

**Option 2: YOLOv5m (Balanced)**
```bash
curl -L -o detector.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5m.onnx
# Inference: ~120ms on 2018 MacBook Pro CPU
# Accuracy: 45.4 mAP on COCO
# Best for: Balance between speed and accuracy
```

**Option 3: YOLOv5l (High accuracy)**
```bash
curl -L -o detector.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5l.onnx
# Inference: ~180ms on 2018 MacBook Pro CPU
# Accuracy: 49.0 mAP on COCO
# Best for: Offline video processing where accuracy matters
```

**Option 4: YOLOv5x (Maximum accuracy, slower)**
```bash
curl -L -o detector.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5x.onnx
# Inference: ~330ms on 2018 MacBook Pro CPU, ~150ms with OpenCL on AMD GPU
# Accuracy: 50.7 mAP on COCO
# Best for: Batch processing, archival analysis, or when maximum accuracy is required
# Note: Slower but provides best object detection quality
```

### Relation Predictor (Optional)

The relation predictor can work in two modes:

1. **Geometric Mode (Default)**: Uses bounding box geometry to infer spatial relations
   - No model file needed
   - Works out of the box
   - Supports: left_of, right_of, on, under, overlaps, contains, next_to

2. **Learned Mode (Advanced)**: Uses a trained ONNX model
   - Requires custom-trained relation model
   - See [docs/TRAINING_RELATIONS.md](../docs/TRAINING_RELATIONS.md) for details

## Model Specifications

### Object Detector Requirements
- Format: ONNX (Open Neural Network Exchange)
- Input: [1, 3, 640, 640] (or 416, 512, etc.)
- Output: [1, 25200, 85] (YOLO format)
- Classes: COCO dataset (80 classes) or custom

### Relation Predictor Requirements (if using learned model)
- Format: ONNX
- Input: Concatenated features from object pairs + spatial features
- Output: Predicate class probabilities

## Verifying Models

Check model info using `netron` or `onnx` tools:

```bash
# Install netron
pip install netron

# Visualize model
netron detector.onnx
```

## License Notes

- YOLOv5 models: AGPL-3.0 license
- Custom models: Check respective licenses
- COCO dataset labels: CC BY 4.0

## References

- YOLOv5: https://github.com/ultralytics/yolov5
- ONNX: https://onnx.ai/
- COCO dataset: https://cocodataset.org/
