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
#   (mAP = mean Average Precision: 0-100%, higher = better detection accuracy)
# Best for: Real-time webcam processing
```

**Option 2: YOLOv5m (Balanced)**
```bash
curl -L -o detector.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5m.onnx
# Inference: ~120ms on 2018 MacBook Pro CPU
# Accuracy: 45.4 mAP on COCO
#   (mAP = mean Average Precision: 0-100%, higher = better detection accuracy)
# Best for: Balance between speed and accuracy
```

**Option 3: YOLOv5l (High accuracy)**
```bash
curl -L -o detector.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5l.onnx
# Inference: ~180ms on 2018 MacBook Pro CPU
# Accuracy: 49.0 mAP on COCO
#   (mAP = mean Average Precision: 0-100%, higher = better detection accuracy)
# Best for: Offline video processing where accuracy matters
```

**Option 4: YOLOv5x (Maximum accuracy, slower)**
```bash
curl -L -o detector.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5x.onnx
# Inference: ~330ms on 2018 MacBook Pro CPU, ~150ms with OpenCL on AMD GPU
# Accuracy: 50.7 mAP on COCO
#   (mAP = mean Average Precision: 0-100%, higher = better detection accuracy)
# Best for: Batch processing, archival analysis, or when maximum accuracy is required
# Note: Slower but provides best object detection quality
```

**What is mAP?**
- **mAP** (mean Average Precision) measures how accurately a model detects and localizes objects
- Scale: 0-100% (higher is better)
- Combines two metrics:
  - **Precision**: % of detections that are correct (not false positives)
  - **Recall**: % of actual objects that are detected (not missed)
- Example: 50.7 mAP means the model correctly detects and localizes objects with 50.7% average precision across all object classes

### Hardware Requirements for 24 FPS Real-time Processing

To achieve 24 fps (frames per second), each frame must be processed in ≤41.7ms (1000ms / 24fps).

**CPU Options (24+ fps):**
- **Intel Core i9-13900K** (24 cores) - YOLOv5s: ~15-20ms → **50-67 fps**
- **AMD Ryzen 9 7950X** (16 cores) - YOLOv5s: ~18-22ms → **45-55 fps**
- **Intel Core i7-12700K** (12 cores) - YOLOv5s: ~22-28ms → **36-45 fps**
- **Apple M2 Pro** (12 cores) - YOLOv5s: ~20-25ms → **40-50 fps**
- **AMD Ryzen 7 5800X** (8 cores) - YOLOv5s: ~28-35ms → **28-36 fps**

**GPU Options (24+ fps with YOLOv5s/m/l):**

*NVIDIA GPUs (CUDA - requires TensorRT, not OpenCL):*
- **RTX 4090** - YOLOv5s: ~2-3ms, YOLOv5x: ~8-10ms → **100-500 fps** / **100-125 fps**
- **RTX 4080** - YOLOv5s: ~3-4ms, YOLOv5x: ~12-15ms → **250-333 fps** / **67-83 fps**
- **RTX 4070** - YOLOv5s: ~4-5ms, YOLOv5x: ~15-18ms → **200-250 fps** / **55-67 fps**
- **RTX 3090** - YOLOv5s: ~4-5ms, YOLOv5x: ~16-20ms → **200-250 fps** / **50-62 fps**
- **RTX 3080** - YOLOv5s: ~5-6ms, YOLOv5x: ~20-25ms → **167-200 fps** / **40-50 fps**
- **RTX 3070** - YOLOv5s: ~6-7ms, YOLOv5x: ~25-30ms → **143-167 fps** / **33-40 fps**
- **GTX 1660 Ti** - YOLOv5s: ~10-13ms, YOLOv5x: ~45-55ms → **77-100 fps** / **18-22 fps**
- **GTX 1080 Ti** - YOLOv5s: ~8-10ms, YOLOv5x: ~35-40ms → **100-125 fps** / **25-28 fps**

*AMD GPUs (OpenCL - supported by this application):*
- **Radeon Pro 560X** (Mobile/MacBook Pro) - YOLOv5s: ~25-35ms, YOLOv5x: ~110-140ms → **28-40 fps** / **7-9 fps**
- **RX 7900 XTX** - YOLOv5s: ~8-12ms, YOLOv5x: ~40-50ms → **83-125 fps** / **20-25 fps**
- **RX 6900 XT** - YOLOv5s: ~10-15ms, YOLOv5x: ~50-60ms → **67-100 fps** / **17-20 fps**
- **RX 6800 XT** - YOLOv5s: ~12-16ms, YOLOv5x: ~55-65ms → **62-83 fps** / **15-18 fps**
- **RX 6700 XT** - YOLOv5s: ~15-20ms, YOLOv5x: ~65-75ms → **50-67 fps** / **13-15 fps**
- **RX 5700 XT** - YOLOv5s: ~18-25ms, YOLOv5x: ~80-100ms → **40-55 fps** / **10-12 fps**

*Intel GPUs (OpenCL):*
- **Arc A770** - YOLOv5s: ~12-18ms, YOLOv5x: ~60-80ms → **55-83 fps** / **12-17 fps**
- **Arc A750** - YOLOv5s: ~15-22ms, YOLOv5x: ~75-95ms → **45-67 fps** / **10-13 fps**

*NVIDIA Jetson (Embedded AI Platform):*
- **Jetson AGX Orin 64GB** (275 TOPS INT8, 15-60W TDP)
  - YOLOv5s: ~8-12ms, YOLOv5m: ~18-25ms, YOLOv5l: ~35-45ms, YOLOv5x: ~55-70ms → **83-125 fps** / **40-55 fps** / **22-28 fps** / **14-18 fps**
  - Power: 15W idle, 25-40W typical inference, 60W max
  - **Ideal for edge deployment**: Low power, high performance, fanless operation possible
  
- **Jetson AGX Orin 32GB** (200 TOPS INT8, 15-50W TDP)
  - YOLOv5s: ~10-15ms, YOLOv5m: ~22-30ms, YOLOv5l: ~42-55ms, YOLOv5x: ~65-85ms → **67-100 fps** / **33-45 fps** / **18-24 fps** / **12-15 fps**
  - Power: 15W idle, 25-35W typical inference, 50W max
  
- **Jetson Orin NX 16GB** (100 TOPS INT8, 10-25W TDP)
  - YOLOv5s: ~15-22ms, YOLOv5m: ~35-48ms, YOLOv5l: ~70-90ms → **45-67 fps** / **21-28 fps** / **11-14 fps**
  - YOLOv5x: Too slow for real-time (~120-150ms → 6-8 fps)
  - Power: 10W idle, 15-20W typical inference, 25W max
  
- **Jetson Orin Nano 8GB** (40 TOPS INT8, 5-15W TDP)
  - YOLOv5s: ~25-35ms, YOLOv5m: ~60-80ms → **28-40 fps** / **12-17 fps**
  - YOLOv5l/x: Not recommended for real-time
  - Power: 5W idle, 10-12W typical inference, 15W max
  - **Best for ultra-low-power deployment**

**Jetson vs 2018 MacBook Pro CPU Performance:**
- 2018 MBP Intel CPU: YOLOv5s ~65ms, YOLOv5m ~120ms, YOLOv5l ~180ms, YOLOv5x ~330ms
- **Jetson Orin Nano** (lowest Jetson): 2-3x faster than 2018 MBP CPU, **1/4 the power**
- **Jetson AGX Orin 64GB**: 5-8x faster than 2018 MBP CPU, **similar power to laptop**

**Models That Run at ~20 fps on Jetson but Not on 2018 MBP CPU:**
- **YOLOv5l** (49.0 mAP): Jetson AGX Orin 64GB: 22-28 fps | 2018 MBP: 5.5 fps ❌
- **YOLOv5x** (50.7 mAP): Jetson AGX Orin 64GB: 14-18 fps | 2018 MBP: 3 fps ❌
- **YOLOv5x**: Jetson Orin NX 16GB: Not viable | 2018 MBP: 3 fps ❌

**Jetson Deployment Advantages:**
1. **Power Efficiency**: 10-40W vs 45-87W for laptops
2. **Fanless Options**: Passive cooling possible at lower power modes
3. **Compact Form Factor**: Carrier boards fit in small enclosures
4. **Extended Temperature Range**: Industrial-grade reliability (-25°C to 80°C)
5. **GPIO/CSI Support**: Direct camera interface, no USB bottleneck
6. **Always-On Operation**: Designed for 24/7 edge deployment

**Recommended Configurations for 24 fps:**

1. **Budget (CPU only):**
   - Intel Core i5-12600K or AMD Ryzen 5 5600X
   - Use YOLOv5s model (~30-40ms → **25-33 fps**)

2. **Mid-range (CPU + OpenCL GPU):**
   - Intel Core i5-13600K + AMD RX 6700 XT
   - Use YOLOv5s with OpenCL (~15ms → **67 fps**)
   - Use YOLOv5m with OpenCL (~25ms → **40 fps**)

3. **High-end (Maximum accuracy at 24 fps):**
   - Any modern CPU + NVIDIA RTX 4070 or AMD RX 7900 XTX
   - Use YOLOv5x with GPU (~15-50ms → **20-67 fps**)

4. **Embedded/Edge (Specialized):**
   - NVIDIA Jetson AGX Orin - YOLOv5s: ~15-25ms → **40-67 fps**
   - NVIDIA Jetson Xavier NX - YOLOv5s: ~30-40ms → **25-33 fps**

**Notes:**
- Performance varies with image resolution (benchmarks assume 640×640 input)
- OpenCL performance is typically 2-3x slower than CUDA on NVIDIA GPUs
- For 24 fps with maximum accuracy (YOLOv5x), high-end GPUs are required
- CPU-only 24 fps requires modern high-core-count processors with YOLOv5s/nano
- Multi-stream processing can achieve higher total throughput on GPUs

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

## Transformer-Based Architectures for Higher Accuracy

Transformer-based object detectors can achieve significantly higher mAP scores than CNN-based models like YOLO, with more diverse label support and better generalization.

### DETR (DEtection TRansformer) Family

**DETR (Facebook/Meta AI)**
- **mAP**: 42.0 on COCO (base), 43.5 (ResNet-101 backbone)
- **Architecture**: Vision transformer with set prediction approach
- **Labels**: 80 COCO classes (same as YOLO), but better at small objects
- **Advantages**: Superior detection quality, fewer false positives, better at occlusions
- **Disadvantages**: ~10x slower than YOLOv5 for similar hardware

**Hardware for 1 fps (1000ms per frame) with DETR:**
- **CPU**: Intel i9-12900K or AMD Ryzen 9 5950X (inference: ~800-1200ms)
- **GPU (OpenCL)**: AMD RX 6700 XT or better (inference: ~600-900ms)
- **GPU (CUDA)**: NVIDIA RTX 3060 or better (inference: ~80-120ms with TensorRT)
- **Recommended**: NVIDIA RTX 3070 with TensorRT optimization (~60-90ms → **11-16 fps**)

**Deformable DETR**
- **mAP**: 46.9 on COCO (improved backbone)
- **Architecture**: Deformable attention for faster convergence
- **Hardware for 1 fps**: Similar to DETR but 20-30% faster
- **NVIDIA RTX 3070**: ~40-60ms → **16-25 fps**

### DINO (DETR with Improved deNoising anchOr boxes)

**DINO (2022)**
- **mAP**: 49.0 on COCO (Swin-L backbone), **63.2 mAP** (with larger backbone)
- **Architecture**: State-of-the-art transformer detector
- **Labels**: 80 COCO classes + easily fine-tunable for custom classes
- **Advantages**: Best-in-class detection quality, excellent zero-shot transfer

**Hardware for 1 fps with DINO:**
- **CPU**: Not practical (>5000ms per frame)
- **GPU (OpenCL)**: AMD RX 7900 XTX (inference: ~800-1200ms)
- **GPU (CUDA)**: 
  - NVIDIA RTX 3080: ~150-200ms → **5-6 fps**
  - NVIDIA RTX 4070: ~100-130ms → **7-10 fps**
  - NVIDIA RTX 4090: ~60-80ms → **12-16 fps**

### ViTDet (Vision Transformer Detector)

**ViTDet (Meta AI, 2022)**
- **mAP**: 55.9 on COCO (ViT-H backbone)
- **Architecture**: Plain vision transformer with simple detection head
- **Labels**: 80 COCO classes, excellent for transfer learning
- **Advantages**: Scales very well with larger models, excellent fine-tuning

**Hardware for 1 fps:**
- **NVIDIA RTX 3080**: ~400-600ms → **1.6-2.5 fps**
- **NVIDIA RTX 4080**: ~200-300ms → **3-5 fps**
- **NVIDIA RTX 4090**: ~120-180ms → **5-8 fps**

### Grounding DINO (Open-Set Detection)

**Grounding DINO (2023)**
- **mAP**: 52.5 on COCO (zero-shot), 56.9 (fine-tuned)
- **Architecture**: Combines vision-language transformers for text-prompted detection
- **Labels**: **Open vocabulary** - can detect any object described in text!
- **Advantages**: 
  - Detects objects not in training set via text prompts
  - Perfect for environmental monitoring: "tree", "shed", "construction crane"
  - No retraining needed for new object types
- **Disadvantages**: Computationally expensive

**Hardware for 1 fps:**
- **NVIDIA RTX 3080**: ~800-1200ms
- **NVIDIA RTX 4070**: ~400-600ms → **1.6-2.5 fps**
- **NVIDIA RTX 4090**: ~250-350ms → **2.8-4 fps**

**Example use case for environmental monitoring:**
```python
# Detect with custom text prompts
prompts = ["oak tree", "pine tree", "metal shed", "wooden barn", 
           "excavator", "construction crane", "pickup truck"]
# Model detects these without retraining!
```

### Florence-2 (Microsoft, 2023)

**Florence-2**
- **mAP**: Not directly comparable (unified vision model)
- **Architecture**: Unified vision-language model (detection + captioning + segmentation)
- **Labels**: Open vocabulary + generates natural language descriptions
- **Advantages**: Can describe scenes in natural language, detect+caption simultaneously

**Hardware for 1 fps:**
- **NVIDIA RTX 4070**: ~600-900ms → **1-1.6 fps**
- **NVIDIA RTX 4090**: ~300-450ms → **2-3 fps**

### YOLO-World (2024) - Best Practical Option

**YOLO-World**
- **mAP**: 35.4 (zero-shot), 52.0 (fine-tuned) on COCO
- **Architecture**: YOLO + vision-language model for open vocabulary
- **Labels**: **Open vocabulary** - detects via text prompts
- **Advantages**: 
  - 10-20x faster than Grounding DINO
  - Real-time capable with modern GPUs
  - Easy to use with YOLO deployment pipelines
  - Can detect: "residential house", "agricultural shed", "deciduous tree", "coniferous tree"

**Hardware for 1 fps:**
- **CPU**: Intel i9-13900K (~300-450ms → **2-3 fps**)
- **GPU (OpenCL)**: AMD RX 6700 XT (~120-180ms → **5-8 fps**)
- **GPU (CUDA)**:
  - NVIDIA RTX 3070: ~40-60ms → **16-25 fps**
  - NVIDIA RTX 4070: ~25-35ms → **28-40 fps**
  - NVIDIA RTX 4090: ~15-20ms → **50-67 fps**
- **Jetson (CUDA/TensorRT)**:
  - Jetson AGX Orin 64GB: ~45-65ms → **15-22 fps** ⚡ (25-40W)
  - Jetson AGX Orin 32GB: ~55-75ms → **13-18 fps** (20-35W)
  - Jetson Orin NX 16GB: ~100-140ms → **7-10 fps** (15-20W)

### Transformer Models on Jetson Orin That Outperform 2018 MBP

The Jetson AGX Orin can run higher-accuracy transformer models at ~20 fps that would be too slow on a 2018 MacBook Pro CPU:

**1. YOLOv5l (49.0 mAP) - Viable on Jetson, Not on MBP:**
- Jetson AGX Orin 64GB: **22-28 fps** at 25-35W ✅
- 2018 MBP Intel CPU: **5.5 fps** at 45-87W ❌
- **3-5x faster on Jetson, 50% less power**

**2. YOLOv5x (50.7 mAP) - Viable on Jetson, Not on MBP:**
- Jetson AGX Orin 64GB: **14-18 fps** at 30-40W ✅
- 2018 MBP Intel CPU: **3 fps** at 45-87W ❌
- **4-6x faster on Jetson, 30% less power**

**3. YOLO-World-m (Open Vocabulary, 45-52 mAP) - Real-time on Jetson:**
- Jetson AGX Orin 64GB: **15-22 fps** at 25-40W ✅ **[Recommended for environmental monitoring]**
- 2018 MBP Intel CPU: **2-3 fps** at 45-87W ❌
- **7-10x faster on Jetson, 40% less power**
- **Key Advantage**: Detects custom objects via text ("oak tree", "metal shed") without retraining

**4. Deformable DETR (46.9 mAP) - Borderline viable on Jetson:**
- Jetson AGX Orin 64GB with TensorRT: **8-12 fps** at 35-45W ⚠️
- 2018 MBP Intel CPU: **0.8-1.2 fps** at 45-87W ❌
- **10x faster on Jetson, similar power**

**Why Jetson Excels for Environmental Edge Deployment:**
1. **Power Efficiency**: 15-40W vs 45-87W for laptops (2-3x better)
2. **24/7 Operation**: Designed for always-on deployment
3. **Fanless Options**: Jetson Orin Nano/NX can run passively cooled
4. **Direct Camera Interface**: CSI cameras bypass USB bottleneck
5. **Industrial Temperature Range**: -25°C to 80°C operation
6. **Compact**: Carrier boards fit in weatherproof enclosures
7. **TensorRT Optimization**: 2-4x speedup vs generic ONNX runtime

### Recommendations for Environmental Monitoring

**Best Overall (Accuracy + Speed + Flexibility): YOLO-World**
- **Why**: Open vocabulary (custom text prompts), near-real-time on good GPU, good mAP
- **Hardware**: NVIDIA RTX 3070 or better for real-time
- **Use case**: "detect oak trees, garden sheds, construction equipment" without retraining

**Maximum Accuracy: DINO**
- **Why**: 63.2 mAP (vs. 50.7 for YOLOv5x)
- **Hardware**: NVIDIA RTX 4080/4090 for 1+ fps
- **Use case**: Offline analysis where accuracy is critical

**Most Flexible: Grounding DINO**
- **Why**: True open vocabulary, detects anything described in text
- **Hardware**: NVIDIA RTX 4090 for usable speed (2-4 fps)
- **Use case**: Rapidly changing monitoring requirements without model retraining

**Practical CPU Option: Deformable DETR**
- **Why**: Better than YOLO accuracy (46.9 mAP) with manageable CPU load
- **Hardware**: Modern 16+ core CPU for 1 fps
- **Use case**: No GPU available but need better accuracy than YOLO

### Converting to ONNX

Most transformer models require additional steps for ONNX export:

```bash
# Example: Export Grounding DINO to ONNX
pip install transformers onnx onnxruntime

python -c "
from transformers import AutoModel
import torch

model = AutoModel.from_pretrained('IDEA-Research/grounding-dino-base')
dummy_input = torch.randn(1, 3, 800, 800)
torch.onnx.export(model, dummy_input, 'grounding_dino.onnx',
                  opset_version=14, 
                  input_names=['input'],
                  output_names=['output'])
"
```

**Note**: Transformer models often require custom post-processing code and may not be fully compatible with OpenCV DNN. Consider using ONNX Runtime directly for better transformer support.

## License Notes

- YOLOv5 models: AGPL-3.0 license
- DETR family: Apache 2.0 license
- Grounding DINO: Apache 2.0 license
- YOLO-World: GPL-3.0 license
- Custom models: Check respective licenses
- COCO dataset labels: CC BY 4.0

## References

- YOLOv5: https://github.com/ultralytics/yolov5
- DETR: https://github.com/facebookresearch/detr
- DINO: https://github.com/IDEA-Research/DINO
- Grounding DINO: https://github.com/IDEA-Research/GroundingDINO
- YOLO-World: https://github.com/AILab-CVC/YOLO-World
- ViTDet: https://github.com/facebookresearch/detectron2
- Florence-2: https://huggingface.co/microsoft/Florence-2
- ONNX: https://onnx.ai/
- COCO dataset: https://cocodataset.org/
