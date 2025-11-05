# Scene Graph Detector Models

This directory contains ONNX models for object detection and relation prediction.

## Required Models

### Object Detector (Required)

Download a YOLOv5 or similar ONNX model for object detection:

**Option 1: YOLOv5s (Fast, recommended)**
```bash
curl -L -o detector.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx
# SHA256: (will vary by version)
```

**Option 2: YOLOv5m (Balanced)**
```bash
curl -L -o detector.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5m.onnx
```

**Option 3: YOLOv5l (High accuracy)**
```bash
curl -L -o detector.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5l.onnx
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
