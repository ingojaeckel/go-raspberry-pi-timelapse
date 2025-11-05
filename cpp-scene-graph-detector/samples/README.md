# Sample Data for Scene Graph Detector

This directory contains sample images and expected outputs for testing the scene graph detector.

## Files

### Sample Image (placeholder)

Since we cannot include actual images in the repository without proper licensing, you should add your own test images here.

**Recommended test images:**
- `example1.jpg` - Simple scene with 2-3 objects
- `example2.jpg` - Complex outdoor scene
- `example3.jpg` - Indoor scene with furniture

### Expected Outputs

Golden outputs for comparison:
- `example1_output.json` - Expected scene graph for example1.jpg
- `example1_output.dot` - Expected DOT file for example1.jpg

## Running Tests

```bash
# Process sample image (you need to provide your own test image)
cd cpp-scene-graph-detector/build

./cpp-scene-graph-detector \
  --input ../samples/your_image.jpg \
  --model.detector ../assets/models/yolov5s.onnx \
  --labels ../assets/models/coco.names \
  --out.json ../samples/output.json \
  --out.dot ../samples/output.dot \
  --visualize ../samples/output_vis.jpg \
  --verbose
```

## Adding Your Own Samples

1. Add test images to this directory
2. Run the detector to generate outputs
3. Manually verify the outputs are correct
4. Save as golden outputs for regression testing
