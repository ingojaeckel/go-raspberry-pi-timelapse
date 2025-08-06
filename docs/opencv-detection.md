# OpenCV Object Detection Integration

This document describes the high-accuracy object detection system integrated into the Raspberry Pi timelapse camera.

## Overview

The enhanced object detection system provides OpenCV-level capabilities for identifying animals, humans, machinery, and vehicles in timelapse photos. The system is designed to run offline on Raspberry Pi Zero hardware with processing times up to 5 minutes for maximum accuracy.

## Features

### Advanced Object Recognition
- **Human Detection**: HOG (Histogram of Oriented Gradients) detection with skin tone analysis
- **Animal Detection**: Texture analysis, earth tone patterns, and organic shape recognition
- **Machinery Detection**: Edge analysis, metal surface detection, and geometric patterns for tractors, tools, equipment
- **Vehicle Detection**: Contour analysis, rectangular shapes, and artificial color detection
- **Day/Night Analysis**: Enhanced brightness and contrast analysis

### Multi-Algorithm Analysis
The system uses multiple detection approaches for maximum accuracy:
1. **YOLO Object Detection**: Pre-trained deep learning model for high accuracy
2. **OpenCV Feature Detection**: HOG, contour analysis, edge detection
3. **Enhanced Color Analysis**: Advanced color space analysis for object classification
4. **Texture Analysis**: Local Binary Patterns and gradient analysis
5. **Fallback Detection**: Pixel-level enhanced analysis when OpenCV unavailable

## Installation

### Automatic Setup
The enhanced setup script automatically installs all dependencies:

```bash
wget -O - https://raw.githubusercontent.com/ingojaeckel/go-raspberry-pi-timelapse/master/setup.sh | bash
```

### Manual Installation
1. Install OpenCV and dependencies:
```bash
sudo apt-get install python3-opencv python3-numpy python3-pip
pip3 install --user opencv-python numpy
```

2. Download YOLO models:
```bash
cd go-raspberry-pi-timelapse/
./detection/download_models.sh
```

## Configuration

### Settings
The system adds new configuration options:

```json
{
  "ObjectDetectionEnabled": true,
  "UseOpenCVDetection": true,
  "DetectionTimeout": 300
}
```

- `ObjectDetectionEnabled`: Enable/disable object detection (default: false)
- `UseOpenCVDetection`: Prefer OpenCV detection when available (default: true)
- `DetectionTimeout`: Maximum detection time in seconds (default: 300 = 5 minutes)

### Detection Modes

1. **OpenCV Mode** (Preferred): Uses YOLO + OpenCV algorithms
2. **Enhanced Mode** (Fallback): Uses advanced pixel analysis
3. **Backward Compatible**: Existing API maintained

## Technical Implementation

### Python OpenCV Script
Location: `detection/opencv_detector.py`

Features:
- YOLO v3-tiny optimized for Raspberry Pi
- HOG human detection
- Contour-based vehicle detection
- Texture analysis for animals
- Edge analysis for machinery
- Graceful degradation when models unavailable

### Go Integration
The Go code automatically:
1. Attempts OpenCV detection first (if enabled)
2. Falls back to enhanced detection on failure
3. Respects timeout configuration
4. Maintains backward compatibility

### API Integration

```go
// New configurable API
config := &detection.DetectionConfig{
    UseOpenCV: true,
    Timeout:   5 * time.Minute,
}
result, err := detection.AnalyzePhotoWithConfig(photoPath, config)

// Backward compatible API
result, err := detection.AnalyzePhoto(photoPath)
```

### REST API
The `/detection` endpoint automatically uses current configuration settings.

## Performance

### Raspberry Pi Zero Optimization
- YOLO-tiny model for faster processing
- Configurable timeout (up to 5 minutes)
- Memory-efficient algorithms
- CPU-only processing (no GPU required)

### Expected Performance
- OpenCV detection: 30 seconds - 5 minutes depending on complexity
- Enhanced detection: 1-10 seconds
- Graceful fallback ensures reliability

## Object Categories

### Detection Classes
- **Animals**: cat, dog, bird, horse, sheep, cow, elephant, bear, zebra, giraffe
- **Humans**: person (all ethnicities and clothing)
- **Vehicles**: car, motorcycle, bus, truck, bicycle, airplane, boat
- **Machinery**: tractors, construction equipment, industrial machinery

### Confidence Scoring
Each detection includes confidence scores:
```json
{
  "details": [
    {"class": "person", "confidence": 0.85, "category": "human"},
    {"class": "animal", "confidence": 0.72, "category": "animal"},
    {"class": "machinery", "confidence": 0.68, "category": "machinery"}
  ]
}
```

## Offline Operation

The system is designed for complete offline operation:
- All models downloaded during setup
- No internet connectivity required during detection
- Works in remote locations without network access
- All processing done locally on device

## Troubleshooting

### OpenCV Not Available
If OpenCV is not installed, the system automatically falls back to enhanced detection without affecting functionality.

### Model Files Missing
If YOLO model files are missing, the system uses OpenCV feature detection algorithms as fallback.

### Memory Issues
On memory-constrained devices, the system automatically adjusts processing parameters and uses smaller models.

### Timeout Issues
Detection timeout can be configured up to 5 minutes. Longer timeouts provide more accuracy but slower response.

## Examples

### Basic Detection
```bash
python3 detection/opencv_detector.py photo.jpg --output-json
```

### Integration with Timelapse
Object detection runs automatically after each photo capture when `ObjectDetectionEnabled` is true.

### Manual API Testing
```go
result, err := detection.AnalyzePhoto("/path/to/photo.jpg")
fmt.Printf("Summary: %s\n", result.Summary)
```

## Architecture

```
Photo Capture
     ↓
Detection Enabled? → No → Skip Detection
     ↓ Yes
OpenCV Available? → No → Enhanced Detection
     ↓ Yes               ↓
YOLO Detection          Pixel Analysis
     ↓                   ↓
OpenCV Features    →     Result
     ↓
   Result
```

This architecture ensures maximum accuracy when possible while maintaining reliability across all hardware configurations.