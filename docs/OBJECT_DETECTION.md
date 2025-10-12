# Object Detection with YOLO

This document describes the integrated YOLO object detection feature for the Go Raspberry Pi Timelapse application.

## Overview

The timelapse application now supports optional object detection on captured photos using YOLO (You Only Look Once) models. When enabled, each captured photo is analyzed to detect and identify objects, animals, and people.

## Features

- **Toggle on/off**: Enable or disable object detection via the web interface
- **Automatic detection**: Runs automatically after each photo capture
- **Summary logging**: Logs a human-readable summary of detected objects
- **Frontend display**: Shows detection results in the camera preview interface
- **Model flexibility**: Supports different YOLO models (YOLOv5s by default)
- **Graceful fallback**: Works without YOLO if not installed (mock mode)

## Architecture

The object detection integration consists of:

1. **Go Detection Package** (`detection/`):
   - `detector.go`: Core detection logic using GoCV
   - `store.go`: In-memory storage for detection results
   - YOLO detector implementation with native OpenCV integration
   - Mock detector for testing

2. **GoCV Integration**:
   - Uses `gocv.io/x/gocv` for native Go bindings to OpenCV
   - Loads YOLO models in ONNX format
   - Performs inference directly without IPC overhead
   - Efficient memory management

3. **Integration Points**:
   - Configuration: `ObjectDetectionEnabled` flag in Settings
   - Timelapse: Detection runs after `camera.Capture()`
   - REST API: `/detection/latest` endpoint for results
   - Frontend: Toggle in Setup, display in Preview

## Setup

### Prerequisites

For object detection with YOLO, you need:

1. **OpenCV Development Libraries**:
   ```bash
   sudo apt-get update
   sudo apt-get install -y libopencv-dev pkg-config
   ```

2. **YOLO Model File**:
   Download a YOLOv5 model in ONNX format:
   ```bash
   # Create model directory
   sudo mkdir -p /usr/local/share/yolo
   
   # Download YOLOv5s (small, fast)
   wget https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx \
        -O /usr/local/share/yolo/yolov5s.onnx
   ```

3. **Rebuild the Application**:
   After installing OpenCV, rebuild the timelapse application to include GoCV support:
   ```bash
   cd /path/to/go-raspberry-pi-timelapse
   go build
   ```

### Installation on Raspberry Pi

**Automated Installation (Recommended):**

```bash
# Run the installation script
cd /path/to/go-raspberry-pi-timelapse
sudo bash scripts/install_object_detection.sh
```

The script will:
- Install OpenCV development libraries
- Download the appropriate YOLO model for your system
- Verify the installation

After installation, rebuild the application:
```bash
go build
```

**Manual Installation:**

```bash
# Install OpenCV dependencies
sudo apt-get update
sudo apt-get install -y libopencv-dev pkg-config

# Create model directory
sudo mkdir -p /usr/local/share/yolo

# Download YOLOv5s model
cd /usr/local/share/yolo
sudo wget https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx

# Rebuild the application with GoCV support
cd /path/to/go-raspberry-pi-timelapse
go build
```

## Model Selection

The application supports different YOLO models with varying speed/accuracy tradeoffs:

| Model | Size | Speed | Accuracy | Recommended For |
|-------|------|-------|----------|-----------------|
| YOLOv5n | ~4 MB | Fastest | Good | Pi Zero W, battery operation |
| YOLOv5s | ~14 MB | Fast | Better | Pi 4, default choice |
| YOLOv5m | ~40 MB | Medium | High | Pi 5, high accuracy needs |
| YOLOv5l | ~90 MB | Slow | Highest | Desktop, maximum accuracy |

### Changing Models

To use a different model, download it and update the model path:

```bash
# Download YOLOv5n (nano - smallest, fastest)
wget https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5n.onnx \
     -O /usr/local/share/yolo/yolov5n.onnx
```

Then modify `detection/detector.go` to point to the new model.

## Optimization for Raspberry Pi 5

For Raspberry Pi 5 (4GB RAM):

1. **Model Choice**: Use YOLOv5s or YOLOv5m
   - YOLOv5s: ~100-200ms inference time
   - YOLOv5m: ~200-400ms inference time

2. **CPU Optimization**:
   - Detection runs only once per capture (not real-time)
   - Minimal power impact since captures are infrequent (minutes apart)
   - No GPU acceleration needed for this use case

3. **Memory Usage**:
   - YOLOv5s: ~50-100 MB RAM
   - Model loaded once and kept in memory
   - Safe for 4GB RAM systems

4. **Power Considerations**:
   - For solar/battery operation: Use YOLOv5n
   - For AC power: Use YOLOv5s or YOLOv5m
   - Detection can be toggled off during low-light hours

## Usage

### Enabling Object Detection

1. Open the web interface at `http://192.168.50.1:8080/`
2. Navigate to the "Setup" tab
3. Click "Edit" to enter edit mode
4. Toggle "Object Detection (YOLO)" to "Enabled"
5. Click "Save"

### Viewing Detection Results

1. Navigate to the "Preview" tab
2. Below the camera preview image, you'll see detection results like:
   - "No objects detected"
   - "The photo includes: one person"
   - "It's day time. The photo includes: two birds"

### Checking Logs

Detection results are also logged to the application log:

```bash
# View logs (if logging to file)
tail -f /var/log/timelapse.log

# Example log entries:
2025-10-12 15:45:23 Photo stored in '/storage/photos/2025-10-12_15-45-23.jpg'
2025-10-12 15:45:24 Object detection result: It's day time. The photo includes: one bird
```

## API Reference

### GET /detection/latest

Returns the most recent detection result.

**Response (200 OK):**
```json
{
  "detections": [
    {
      "class_name": "bird",
      "confidence": 0.87,
      "x": 245.3,
      "y": 189.7,
      "width": 56.2,
      "height": 42.8
    }
  ],
  "image_path": "/storage/photos/2025-10-12_15-45-23.jpg",
  "summary": "It's day time. The photo includes: one bird"
}
```

**Response (404 Not Found):**
```json
"No detection results available"
```

### POST /configuration

Update configuration including object detection setting.

**Request Body:**
```json
{
  "ObjectDetectionEnabled": true,
  ...other settings...
}
```

## Supported Objects

The YOLO model can detect 80 different object classes from the COCO dataset:

**Animals**: bird, cat, dog, horse, sheep, cow, elephant, bear, zebra, giraffe
**People**: person
**Vehicles**: car, motorcycle, airplane, bus, train, truck, boat, bicycle
**And 60+ more common objects**

## Troubleshooting

### "YOLO detector not available"

**Symptom**: Logs show "YOLO detector not available, using mock detector"

**Solution**:
1. Verify OpenCV is installed: `pkg-config --modversion opencv4`
2. Check model file exists: `ls -l /usr/local/share/yolo/yolov5s.onnx`
3. Rebuild the application: `go build`
4. Check for GoCV build errors in the build output

### "No objects detected" on every photo

**Possible causes**:
1. Confidence threshold too high (default 0.5)
2. Poor lighting conditions
3. Objects too small or far away
4. Model not suitable for scene type

**Solutions**:
1. Try a larger model (YOLOv5m instead of YOLOv5s)
2. Ensure adequate lighting
3. Adjust camera position to get objects closer

### High CPU usage

**Solution**: 
- Increase `SecondsBetweenCaptures` to reduce detection frequency
- Use a smaller model (YOLOv5n)
- Disable detection during periods when not needed

### Memory issues

**Solution**:
- Use YOLOv5n or YOLOv5s (smaller models)
- Ensure no other heavy processes running
- Consider disabling detection on devices with <2GB RAM

## Performance Impact

### Raspberry Pi 5 (4GB)

With YOLOv5s model:
- Detection time: ~150ms per image
- RAM usage: ~80 MB
- CPU usage: Spike during detection, negligible between captures
- Power impact: Minimal (captures are minutes apart)

### Raspberry Pi 4 (4GB)

With YOLOv5s model:
- Detection time: ~300-400ms per image
- RAM usage: ~80 MB
- CPU usage: Brief spike per capture

### Raspberry Pi Zero W

**Not recommended** for YOLO detection due to:
- Limited RAM (512 MB)
- Slow CPU (single core)
- Detection would take 5-10+ seconds

Use mock detector mode or disable feature on Zero W.

## Future Enhancements

Potential improvements planned:

- [ ] Model selection via web UI
- [ ] Confidence threshold configuration
- [ ] Day/night detection using image analysis
- [ ] Detection history and statistics
- [ ] Bounding box visualization on images
- [ ] Email/webhook notifications for specific objects
- [ ] Integration with cpp-object-detection for real-time tracking

## Related Documentation

- [cpp-object-detection/README.md](../cpp-object-detection/README.md) - Standalone C++ detection app
- [YOLO Official Docs](https://github.com/ultralytics/yolov5) - YOLOv5 documentation
- Main README sections on model comparison

## License

The detection integration uses:
- YOLO models: AGPL-3.0 (Ultralytics)
- OpenCV: Apache 2.0
- This integration code: Same license as main project
