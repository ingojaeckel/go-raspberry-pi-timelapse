# Examples

This directory contains example code demonstrating the features of the cpp-object-detection application.

## Photo Storage Demo

**File:** `photo_storage_demo.cpp`

Demonstrates the photo storage with bounding boxes feature:
- Default and custom output directories
- Simulated detection scenarios
- Color mapping reference
- Feature summary

### To compile and run:

```bash
cd cpp-object-detection
mkdir -p build && cd build
cmake ..
make

# Note: This is a demonstration file showing API usage
# It requires the actual detector initialization which needs model files
# See ../PHOTO_STORAGE_FEATURE.md for complete documentation
```

## Usage in Production

For real-world usage, use the main application:

```bash
# Default settings
./object_detection

# With custom output directory
./object_detection --output-dir /path/to/photos

# With verbose logging
./object_detection --output-dir detections --verbose
```

See the main [README.md](../README.md) for complete usage instructions.
