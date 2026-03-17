# Sample Images and Outputs

This directory contains sample images and their corresponding scene graph outputs.

## Included Samples

### sample_scene.json
Golden output for a sample scene containing:
- Two houses
- One tree (between the houses)
- One bush
- Spatial relations: left_of, between

## Running Samples

To process your own images:

```bash
# Download a YOLO model first
cd cpp-scene-graph-detector
curl -L -o assets/models/yolov5s.onnx \
  https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx

# Process an image
./build/cpp-scene-graph-detector \
  --input path/to/your/image.jpg \
  --model.detector assets/models/yolov5s.onnx \
  --labels assets/labels/coco.txt \
  --out.json output.json \
  --visualize output_viz.jpg
```

## Sample Output Format

See [../docs/SCENE_GRAPH_FORMAT.md](../docs/SCENE_GRAPH_FORMAT.md) for the full specification.
