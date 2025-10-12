# YOLO Model Conversion - Implementation Summary

This document summarizes the implementation of YOLO model conversion from PyTorch (.pt) to ONNX format for the cpp-object-detection application.

## What Was Created

### 1. Conversion Script (`scripts/convert_yolo_to_onnx.py`)

A comprehensive Python script that converts YOLO models (v9, v10, v11) from PyTorch to ONNX format.

**Features:**
- Supports YOLO versions 9, 10, and 11
- Supports all model sizes: n (nano), s (small), m (medium), l (large), x (xlarge)
- Automatic model download via Ultralytics library
- ONNX validation and inference testing
- Customizable output location and parameters
- Detailed error messages and help text

**Usage Examples:**
```bash
# Convert YOLOv11n (recommended for getting started)
python scripts/convert_yolo_to_onnx.py --version 11 --model-size n

# Convert YOLOv10s with custom output
python scripts/convert_yolo_to_onnx.py --version 10 --model-size s --output-dir ./custom_models

# Convert existing .pt file
python scripts/convert_yolo_to_onnx.py --pt-file yolo11n.pt --output yolo11n.onnx
```

### 2. GitHub Actions Workflow (`.github/workflows/yolo-model-conversion.yml`)

A parameterized workflow that validates model conversion on multiple platforms.

**Features:**
- Manual trigger with selectable YOLO version and model size
- Automatic trigger on script/workflow changes
- Tests on Linux (AMD64) and macOS (x86_64)
- Validates ONNX model structure
- Tests inference with ONNX Runtime
- Creates downloadable artifacts
- Generates workflow summary

**How to Use:**
1. Go to Actions tab in GitHub
2. Select "YOLO Model Conversion Validation"
3. Click "Run workflow"
4. Choose YOLO version and model size
5. Download converted model from workflow artifacts

### 3. Documentation

#### YOLO_MODEL_CONVERSION.md
Comprehensive guide covering:
- Quick start guide
- Detailed usage examples
- Model size comparison
- Storage recommendations
- GitHub Actions workflow usage
- Troubleshooting guide
- Technical details

#### Updated README.md
Added references to model conversion in:
- Build instructions section
- Model selection section

#### scripts/README.md
New documentation for the scripts directory including:
- Script overview
- Usage examples
- Quick reference

### 4. Test Script (`scripts/test_conversion_script.sh`)

Automated tests for the conversion script:
- Script existence check
- Help output validation
- Argument validation
- Error handling verification
- Valid argument combinations

## Design Decisions

### Why Not Add ONNX Files to Git Repository?

**Recommendation: Do NOT add ONNX files to the repository**

Reasons:
1. **File Size**: Models range from 6MB to 136MB+ each
2. **Binary Format**: Git doesn't handle binary files efficiently
3. **Repository Bloat**: Large files slow down cloning and pulling
4. **Version Control**: Models rarely need version tracking like code

### Alternative Approaches

1. **Download Pre-converted Models** (for YOLO v5/v8)
   ```bash
   wget -O models/yolov5s.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx
   ```

2. **Convert on Demand** (for YOLO v9/v10/v11)
   ```bash
   python scripts/convert_yolo_to_onnx.py --version 11 --model-size n
   ```

3. **Use CI/CD Artifacts**
   - Download from GitHub Actions workflow runs

4. **Git LFS** (advanced, for teams needing version control)
   - Has storage limits and costs

### Technology Choices

**Ultralytics Library:**
- Official YOLO implementation
- Handles model download automatically
- Built-in ONNX export with optimization
- Well-maintained and documented

**ONNX Format:**
- Platform-independent
- Works with OpenCV's DNN module
- Optimized for inference
- Widely supported

**Python 3.11:**
- Modern Python version
- Good library compatibility
- Available in GitHub Actions

## Validation

### Local Testing
```bash
# Run conversion script tests
cd cpp-object-detection
./scripts/test_conversion_script.sh
```

All 8 tests pass:
- ✓ Script exists
- ✓ Help output works
- ✓ Argument validation
- ✓ Invalid version rejection
- ✓ Invalid model size rejection
- ✓ Dependency check
- ✓ .pt file validation (when dependencies installed)
- ✓ Valid argument parsing

### GitHub Actions Testing
The workflow validates:
- ✓ Model conversion on Linux (AMD64)
- ✓ Model conversion on macOS (x86_64)
- ✓ ONNX model structure
- ✓ Inference with ONNX Runtime
- ✓ File size validation
- ✓ Model metadata extraction

## Model Comparison

| YOLO Version | Model Size | File Size | Speed | Accuracy | Use Case |
|--------------|------------|-----------|-------|----------|----------|
| YOLOv11n | Nano | ~6MB | Fastest | Good | Embedded systems, real-time |
| YOLOv11s | Small | ~14MB | Fast | Better | General purpose |
| YOLOv11m | Medium | ~52MB | Moderate | High | Accuracy-focused |
| YOLOv10n | Nano | ~6MB | Fastest | Good | Alternative to v11n |
| YOLOv9s | Small | ~15MB | Fast | Better | Proven reliability |

## Usage in cpp-object-detection

After converting a model:

```bash
# Build the application
./scripts/build.sh

# Run with converted model
./build/object_detection --model-path models/yolo11n.onnx --camera-id 0
```

## Maintenance

### Updating the Conversion Script

When updating the script:
1. Test locally with `test_conversion_script.sh`
2. Commit changes
3. GitHub Actions automatically validates on push
4. Review workflow results

### Adding New YOLO Versions

To add support for future YOLO versions (e.g., v12):
1. Update `--version` choices in `convert_yolo_to_onnx.py`
2. Update workflow inputs in `yolo-model-conversion.yml`
3. Update documentation in `YOLO_MODEL_CONVERSION.md`
4. Test with new version

## Resources

- [Conversion Script](scripts/convert_yolo_to_onnx.py)
- [Detailed Documentation](YOLO_MODEL_CONVERSION.md)
- [GitHub Actions Workflow](../.github/workflows/yolo-model-conversion.yml)
- [Test Script](scripts/test_conversion_script.sh)
- [Ultralytics Documentation](https://docs.ultralytics.com/)
- [ONNX Documentation](https://onnx.ai/)

## Support

For issues or questions:
1. Check [YOLO_MODEL_CONVERSION.md](YOLO_MODEL_CONVERSION.md) troubleshooting section
2. Review existing GitHub issues
3. Open a new issue with:
   - YOLO version and model size
   - Error messages or logs
   - Platform (Linux/macOS)
   - Python and library versions
