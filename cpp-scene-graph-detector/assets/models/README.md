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

## Models for Environmental Observation

### Pre-trained Models with Environmental Classes

For residential, industrial, and agricultural monitoring, standard COCO models have significant limitations. Here are recommended alternatives:

#### Option A: YOLOv8 with Expanded Classes (Recommended)

**YOLOv8n-seg** - Ultralytics' latest model with instance segmentation:
```bash
# Download YOLOv8 nano (fast for real-time)
curl -L -o detector.onnx https://github.com/ultralytics/assets/releases/download/v0.0.0/yolov8n.onnx

# Or YOLOv8s (balanced)
curl -L -o detector.onnx https://github.com/ultralytics/assets/releases/download/v0.0.0/yolov8s.onnx
```

**Compatible Labels File**: Use `assets/labels/coco.txt` (80 COCO classes)
- Includes: person, bicycle, car, motorcycle, bus, train, truck, bird, cat, dog, horse, sheep, cow
- Missing: trees, houses, sheds, construction equipment (requires custom training)

#### Option B: Objects365-trained YOLOv5 (365 Classes)

**Download pre-trained Objects365 model:**
```bash
# YOLOv5s trained on Objects365 dataset
# Note: Requires conversion from PyTorch to ONNX
git clone https://github.com/ultralytics/yolov5
cd yolov5
pip install -r requirements.txt

# Download Objects365 weights
wget https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s-obj365.pt

# Export to ONNX
python export.py --weights yolov5s-obj365.pt --include onnx --imgsz 640
```

**Objects365 Classes Include:**
- Buildings: house, building, shed, tent
- Vegetation: tree, potted plant, flower
- Vehicles: car, truck, van, SUV, bus, motorcycle, bicycle
- Construction: crane, excavator, bulldozer
- Agricultural: tractor, combine harvester
- Animals: various livestock and wildlife
- Tools and equipment

**Labels File**: Create `assets/labels/objects365.txt` with 365 class names
(Full list available at: https://www.objects365.org/overview.html)

#### Option C: Using Provided Environmental Labels

For custom-trained models targeting environmental monitoring, use:
```bash
# Use the provided environmental labels file
--labels assets/labels/environmental.txt
```

This labels file includes 28 key classes for residential/agricultural monitoring:
- People and vehicles: person, car, truck, bus, bicycle, motorcycle, tractor
- Buildings: house, building, shed, barn
- Vegetation: tree, bush, potted plant
- Animals: bird, cat, dog, horse, sheep, cow
- Equipment: crane, excavator, bulldozer, tool, fence

**Note**: Using this labels file requires a model trained on these specific classes (see custom training below).

### Standard COCO Model Capabilities

The standard COCO-trained YOLO models (Options 1-4 above) include some relevant classes:

**Residential/Agricultural Classes in COCO:**
- **Vehicles**: car, truck, bus, motorcycle, bicycle, train, boat
- **Animals**: bird, cat, dog, horse, sheep, cow
- **Objects**: backpack, umbrella, handbag, suitcase, sports ball, kite, etc.
- **Plants**: potted plant
- **People**: person

**Limitations of COCO Models:**
COCO does not include:
- Trees, bushes, hedges
- Houses, sheds, barns, greenhouses  
- Construction equipment (cranes, bulldozers, excavators)
- Agricultural equipment (tractors, combines)
- Tools and implements
- Fences and outdoor structures

### Custom Trained Models for Environmental Observation

For comprehensive environmental monitoring including all needed classes, custom training is recommended:

1. **Collect & Label Data** (~500-1000 images minimum):
   - Capture images from your specific environment
   - Use labeling tools like [CVAT](https://www.cvat.ai/), [LabelImg](https://github.com/tzutalin/labelImg), or [Roboflow](https://roboflow.com/)
   - Label objects: house, shed, tree, bush, crane, etc.

2. **Prepare Dataset**:
   ```bash
   # Organize in YOLO format:
   dataset/
   ├── images/
   │   ├── train/
   │   └── val/
   └── labels/
       ├── train/
       └── val/
   ```

3. **Train with YOLOv5**:
   ```bash
   git clone https://github.com/ultralytics/yolov5
   cd yolov5
   pip install -r requirements.txt
   
   # Create data.yaml with your custom classes
   # Train (use pre-trained weights for transfer learning)
   python train.py --img 640 --batch 16 --epochs 100 \
     --data data.yaml --weights yolov5s.pt --cache
   
   # Export to ONNX
   python export.py --weights runs/train/exp/weights/best.pt \
     --include onnx --imgsz 640
   ```

4. **Create Custom Labels File**:
   ```bash
   cat > custom_environmental_labels.txt << EOF
   person
   car
   truck
   tree
   bush
   house
   shed
   crane
   tool
   bird
   dog
   cat
   EOF
   ```

### Option 2: Pre-trained Models with Extended Classes

**Objects365 Dataset** - Includes more environmental classes:
- Available at: https://www.objects365.org/
- 365 object categories including buildings, trees, tools
- Pre-trained YOLOv5 models available

**Download Objects365-trained YOLO:**
```bash
# Note: These are larger models trained on broader categories
# Check Ultralytics model zoo or Objects365 release page for ONNX versions
# May require conversion from PyTorch to ONNX
```

### Option 3: Use Existing Specialized Models

**iNaturalist** - For nature/wildlife:
- Good for trees, plants, animals
- https://github.com/visipedia/inat_comp

**PlantNet** - For plant species:
- Specialized in plant identification
- https://plantnet.org/

### Training Resources

**Transfer Learning (Recommended)**:
- Start with COCO-pretrained YOLOv5 weights
- Fine-tune on your custom environmental dataset (500-1000 images)
- Reduces training time from days to hours
- Better accuracy with limited data

**Data Augmentation**:
```yaml
# In data.yaml
# Augmentation settings for environmental data
hsv_h: 0.015  # Image HSV-Hue augmentation
hsv_s: 0.7    # Image HSV-Saturation augmentation
hsv_v: 0.4    # Image HSV-Value augmentation
degrees: 10.0 # Image rotation
translate: 0.1 # Image translation
scale: 0.5    # Image scaling
flipud: 0.0   # No vertical flip (keeps horizon consistent)
fliplr: 0.5   # Horizontal flip for symmetry
```

**Best Practices**:
- Include images from different times of day (lighting variations)
- Capture seasonal variations if monitoring long-term
- Include weather conditions (sunny, cloudy, rain)
- Vary camera angles and distances
- Balance class distribution (similar number of examples per class)

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
