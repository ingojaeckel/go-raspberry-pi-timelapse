# YOLO Model Conversion Guide

This guide explains how to convert recent YOLO models (YOLOv9, YOLOv10, YOLOv11) from PyTorch (.pt) format to ONNX format for use with the C++ object detection application.

## Overview

The `cpp-object-detection` application uses ONNX models for object detection. While pre-converted ONNX models are available for older YOLO versions (v5, v8), newer versions (v9, v10, v11) may need to be converted from PyTorch format.

## Quick Start

### Prerequisites

Install required Python packages:

```bash
pip install ultralytics onnx
```

### Convert a YOLO Model

The simplest way to convert a model:

```bash
cd cpp-object-detection
python scripts/convert_yolo_to_onnx.py --version 11 --model-size n
```

This will:
1. Download YOLOv11n model (if not already cached)
2. Convert it to ONNX format
3. Save it to `models/yolo11n.onnx`

## Usage Examples

### Convert Different YOLO Versions

```bash
# Convert YOLOv11 nano model (fastest, smallest)
python scripts/convert_yolo_to_onnx.py --version 11 --model-size n

# Convert YOLOv10 small model
python scripts/convert_yolo_to_onnx.py --version 10 --model-size s

# Convert YOLOv9 medium model
python scripts/convert_yolo_to_onnx.py --version 9 --model-size m
```

### Model Size Options

- **n** (nano): Fastest, smallest, lower accuracy (~6MB)
- **s** (small): Good balance for real-time use (~14MB)
- **m** (medium): Better accuracy, moderate speed (~52MB)
- **l** (large): High accuracy, slower (~94MB)
- **x** (xlarge): Best accuracy, slowest (~136MB)

### Custom Output Location

```bash
# Save to custom directory
python scripts/convert_yolo_to_onnx.py --version 11 --model-size n --output-dir ./custom_models

# Specify exact output filename
python scripts/convert_yolo_to_onnx.py --version 11 --model-size n --output yolo11n_custom.onnx
```

### Convert Existing .pt File

If you already have a PyTorch model file:

```bash
python scripts/convert_yolo_to_onnx.py --pt-file path/to/yolo11n.pt --output yolo11n.onnx
```

### Advanced Options

```bash
# Custom image size (default: 640)
python scripts/convert_yolo_to_onnx.py --version 11 --model-size n --imgsz 1280

# Custom ONNX opset version (default: 12)
python scripts/convert_yolo_to_onnx.py --version 11 --model-size n --opset 13

# Verbose output for debugging
python scripts/convert_yolo_to_onnx.py --version 11 --model-size n --verbose
```

## Using Converted Models

Once converted, use the ONNX model with the object detection application:

```bash
./build/object_detection --model-path models/yolo11n.onnx --camera-id 0
```

## Model Storage Recommendations

### Should ONNX Files Be Added to Git Repository?

**We recommend NOT adding ONNX files to the Git repository** for the following reasons:

1. **File Size**: ONNX models range from 6MB to 136MB+ per file
2. **Binary Format**: Git doesn't handle binary files efficiently
3. **Repository Bloat**: Large files make cloning and pulling slower
4. **Version Control**: Model files rarely need version tracking like code

### Alternative Approaches

#### Option 1: Download Pre-converted Models (Recommended)

For YOLO v5 and v8, download pre-converted ONNX models directly:

```bash
# YOLOv5s (14MB, widely used)
wget -O models/yolov5s.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx

# YOLOv8n (6MB, fastest)
wget -O models/yolov8n.onnx https://github.com/ultralytics/assets/releases/download/v0.0.0/yolov8n.onnx
```

#### Option 2: Convert on Demand

Users can convert models as needed using the conversion script:

```bash
python scripts/convert_yolo_to_onnx.py --version 11 --model-size n
```

The Ultralytics library caches downloaded .pt files in `~/.cache/ultralytics/`, so subsequent conversions are faster.

#### Option 3: Use CI/CD Artifacts

The GitHub Actions workflow creates ONNX artifacts that can be downloaded from successful workflow runs:

1. Go to **Actions** tab in GitHub
2. Select **YOLO Model Conversion Validation** workflow
3. Choose a workflow run
4. Download artifacts from the run

#### Option 4: Git LFS (Advanced)

For teams that need version-controlled models, consider Git LFS:

```bash
# Install Git LFS
git lfs install

# Track ONNX files
git lfs track "*.onnx"

# Add and commit
git add .gitattributes
git add models/*.onnx
git commit -m "Add ONNX models with LFS"
```

**Note**: Git LFS has storage limits on GitHub (1GB free, bandwidth charges apply).

## GitHub Actions Workflow

### Validate Model Conversion

A GitHub Actions workflow is available to validate model conversion on multiple platforms.

#### Trigger Manually

1. Go to **Actions** tab in GitHub
2. Select **YOLO Model Conversion Validation**
3. Click **Run workflow**
4. Choose YOLO version and model size
5. Click **Run workflow**

The workflow will:
- Convert the specified model on Linux (AMD64) and macOS (x86_64)
- Validate the ONNX model structure
- Test inference with ONNX Runtime
- Upload converted models as artifacts

#### View Results

After the workflow completes:
- Check the **Summary** for conversion status
- Download ONNX artifacts from the workflow run
- Review logs for any issues

### Automatic Validation

The workflow also runs automatically when:
- The conversion script is modified
- The workflow file is modified

This ensures the conversion process stays functional.

## Model Comparison

| YOLO Version | Model Size | File Size | Speed | Accuracy | Use Case |
|--------------|------------|-----------|-------|----------|----------|
| YOLOv11n | Nano | ~6MB | Fastest | Good | Embedded systems, real-time |
| YOLOv11s | Small | ~14MB | Fast | Better | General purpose |
| YOLOv11m | Medium | ~52MB | Moderate | High | Accuracy-focused |
| YOLOv10n | Nano | ~6MB | Fastest | Good | Alternative to v11n |
| YOLOv9s | Small | ~15MB | Fast | Better | Proven reliability |
| YOLOv5s | Small | ~14MB | Fast | Good | Widely used, stable |

## Troubleshooting

### Import Error: ultralytics not found

```bash
pip install ultralytics
```

### ONNX Validation Failed

Ensure you have the latest ultralytics version:

```bash
pip install --upgrade ultralytics onnx
```

### Model Download Fails

Check your internet connection and retry. The Ultralytics library downloads models from GitHub releases.

### Out of Memory During Conversion

Try converting a smaller model size (n or s) instead of larger sizes (l or x).

### Platform-Specific Issues

#### macOS

If you encounter SSL errors:
```bash
pip install --upgrade certifi
```

#### Linux

Ensure you have sufficient disk space in `~/.cache/ultralytics/`

## Technical Details

### Conversion Process

The conversion script uses the official Ultralytics library:

1. **Model Loading**: Uses `YOLO()` to load/download the .pt file
2. **Export**: Calls `model.export(format='onnx')` with optimizations
3. **Validation**: Verifies the ONNX model structure with `onnx.checker`
4. **Output**: Saves to specified location with naming convention

### ONNX Export Settings

Default export settings:
- **Format**: ONNX
- **Image size**: 640x640
- **ONNX opset**: 12 (compatible with most runtimes)
- **Simplify**: True (uses onnx-simplifier for optimization)
- **Dynamic**: False (static input size for better performance)

These settings are optimized for inference with OpenCV's DNN module in the C++ application.

## Additional Resources

- [Ultralytics Documentation](https://docs.ultralytics.com/)
- [YOLO11 Model Guide](https://docs.ultralytics.com/models/yolo11/)
- [ONNX Export Guide](https://docs.ultralytics.com/modes/export/)
- [OpenCV DNN Module](https://docs.opencv.org/4.x/d2/d58/tutorial_table_of_content_dnn.html)

## Support

For issues or questions:
1. Check existing GitHub issues
2. Review [cpp-object-detection/README.md](../README.md)
3. Open a new issue with:
   - YOLO version and model size
   - Error messages or logs
   - Platform (Linux/macOS)
   - Python and library versions
