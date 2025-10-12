# Scripts Directory

This directory contains build, test, and utility scripts for the C++ Object Detection application.

## Build Scripts

- **build.sh** - Cross-platform build script (auto-detects OS)
- **build-mac.sh** - macOS-specific build script
- **build-rpi.sh** - Raspberry Pi ARM64 build script
- **build-linux-386.sh** - 32-bit Linux build script

## Test Scripts

- **test.sh** - Run unit tests
- **test-mac.sh** - macOS-specific test runner
- **test_coverage.sh** - Generate code coverage reports
- **manual_test.sh** - Manual integration testing guide
- **test-integration-*.sh** - Platform-specific integration tests

## Utility Scripts

- **download_model.sh** - Download pre-converted YOLO ONNX models
- **download_models.sh** - Download multiple YOLO model variants
- **convert_yolo_to_onnx.py** - Convert YOLO PyTorch models to ONNX format
- **verify_google_sheets.sh** - Verify Google Sheets integration

## YOLO Model Conversion

The `convert_yolo_to_onnx.py` script converts recent YOLO models (v9, v10, v11) from PyTorch (.pt) format to ONNX format.

### Quick Start

```bash
# Convert YOLOv11n model
python convert_yolo_to_onnx.py --version 11 --model-size n

# Convert YOLOv10s model
python convert_yolo_to_onnx.py --version 10 --model-size s
```

### Requirements

```bash
pip install ultralytics onnx
```

### For Complete Documentation

See [../docs/YOLO_MODEL_CONVERSION.md](../docs/YOLO_MODEL_CONVERSION.md) for:
- Detailed usage examples
- Model size comparison
- Storage recommendations
- GitHub Actions workflow usage
- Troubleshooting guide

## Usage Examples

### Build the Application

```bash
# Auto-detect platform and build
./build.sh

# macOS specific
./build-mac.sh

# Raspberry Pi
./build-rpi.sh
```

### Download YOLO Models

```bash
# Download YOLOv5s (recommended for getting started)
./download_model.sh

# Download multiple model variants
./download_models.sh
```

### Convert YOLO Models

```bash
# Convert latest YOLOv11 nano model
python convert_yolo_to_onnx.py --version 11 --model-size n --verbose

# Convert with custom output location
python convert_yolo_to_onnx.py --version 11 --model-size s --output-dir ../models

# Convert existing .pt file
python convert_yolo_to_onnx.py --pt-file path/to/yolo11n.pt --output yolo11n.onnx
```

### Run Tests

```bash
# Run all tests
./test.sh

# Run with coverage
./test_coverage.sh
```

## Script Maintenance

- All scripts use `set -e` for fail-fast behavior
- Scripts include colored output for better readability
- Cross-platform compatibility is maintained where possible
- Error messages include helpful troubleshooting information
