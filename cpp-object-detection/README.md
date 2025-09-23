# C++ Object Detection Application

A standalone C++ executable for real-time object detection from webcam data at 720p resolution. This application is designed to run headless on Linux systems and detect people, vehicles, and small animals entering and exiting the camera frame.

## Features

- **Real-time object detection** from USB webcam input
- **720p resolution support** with configurable frame rates
- **Headless operation** - no X11 required
- **Object tracking** for people, vehicles, and small animals (cat/dog/fox)
- **Confidence-based filtering** with configurable thresholds
- **Performance monitoring** with automatic warnings for low frame rates
- **Structured logging** with timestamps and object enter/exit events
- **Configurable parameters** via command-line interface
- **Static linking** for standalone deployment

## Architecture

### System Design

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   USB Webcam    │────│  Webcam          │────│   Frame         │
│   (Logitech     │    │  Interface       │    │   Buffer        │
│    C920, etc.)  │    └──────────────────┘    └─────────────────┘
└─────────────────┘                                      │
                                                         │
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Log Files     │────│   Logger         │    │   Main          │
│   (timestamped  │    │   System         │    │   Processing    │
│    events)      │    └──────────────────┘    │   Loop          │
└─────────────────┘                            └─────────────────┘
                                                         │
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Performance    │────│  Performance     │────│   Object        │
│  Metrics &      │    │  Monitor         │    │   Detector      │
│  Warnings       │    └──────────────────┘    │   (YOLO)        │
└─────────────────┘                            └─────────────────┘
                                                         │
                                               ┌─────────────────┐
                                               │   Object        │
                                               │   Tracker       │
                                               │   (Enter/Exit)  │
                                               └─────────────────┘
```

### Component Overview

1. **Main Application (`main.cpp`)**
   - Entry point and main processing loop
   - Signal handling for graceful shutdown
   - Frame rate control and timing

2. **Config Manager (`config_manager.hpp/cpp`)**
   - Command-line argument parsing
   - Configuration validation
   - Default value management

3. **Webcam Interface (`webcam_interface.hpp/cpp`)**
   - USB camera initialization and control
   - Frame capture and buffering
   - Camera capability detection

4. **Object Detector (`object_detector.hpp/cpp`)**
   - YOLO-based object detection
   - Target class filtering (person, vehicles, animals)
   - Object tracking for enter/exit detection

5. **Logger (`logger.hpp/cpp`)**
   - Structured logging with timestamps
   - Object detection event logging
   - Performance and heartbeat logging

6. **Performance Monitor (`performance_monitor.hpp/cpp`)**
   - Frame rate calculation
   - Processing time tracking
   - Performance warning system

## Building

### Prerequisites

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y cmake build-essential libopencv-dev pkg-config
```

**CentOS/RHEL:**
```bash
sudo yum install -y cmake gcc-c++ opencv-devel pkgconfig
```

### Build Instructions

1. **Clone and navigate to the project:**
```bash
cd cpp-object-detection
```

2. **Build the application:**
```bash
./scripts/build.sh
```

3. **Download a YOLO model (required for object detection):**
```bash
wget -O models/yolov5s.onnx https://github.com/ultralytics/yolov5/releases/download/v6.2/yolov5s.onnx
```

### Cross-Platform Builds

**For x86_64:**
```bash
cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_SYSTEM_PROCESSOR=x86_64
make -j$(nproc)
```

**For 386 (32-bit):**
```bash
cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_SYSTEM_PROCESSOR=i386 -DCMAKE_C_FLAGS="-m32" -DCMAKE_CXX_FLAGS="-m32"
make -j$(nproc)
```

## Usage

### Basic Usage

```bash
# Default settings (5 fps max, 50% confidence)
./object_detection

# Custom settings
./object_detection --max-fps 3 --min-confidence 0.7 --verbose

# Specific camera and resolution
./object_detection --camera-id 1 --frame-width 640 --frame-height 480
```

### Command Line Options

```
Usage: object_detection [OPTIONS]

Real-time object detection from webcam data (720p)
Detects people, vehicles, and small animals (cat/dog/fox)

OPTIONS:
  -h, --help                     Show help message
  -v, --verbose                  Enable verbose logging
  --max-fps N                    Maximum frames per second to process (default: 5)
  --min-confidence N             Minimum confidence threshold (0.0-1.0, default: 0.5)
  --min-fps-warning N            FPS threshold for performance warnings (default: 1)
  --log-file FILE                Log file path (default: object_detection.log)
  --heartbeat-interval N         Heartbeat log interval in minutes (default: 10)
  --camera-id N                  Camera device ID (default: 0)
  --frame-width N                Frame width in pixels (default: 1280)
  --frame-height N               Frame height in pixels (default: 720)
  --model-path FILE              Path to ONNX model file
  --processing-threads N         Number of processing threads (default: 1)
  --enable-gpu                   Enable GPU acceleration if available
  --no-headless                  Disable headless mode (show GUI windows)
```

### Examples

**Low-power deployment:**
```bash
./object_detection --max-fps 1 --min-confidence 0.8 --heartbeat-interval 5
```

**High-accuracy monitoring:**
```bash
./object_detection --max-fps 10 --min-confidence 0.3 --verbose
```

**Custom logging:**
```bash
./object_detection --log-file /var/log/security_detection.log --heartbeat-interval 30
```

## Webcam Setup

### Supported Cameras

- **Recommended:** Logitech C920 (excellent 720p support)
- **Compatible:** Any USB Video Class (UVC) camera
- **Requirements:** 720p resolution support, USB 2.0 or higher

### Connecting Different Webcams

1. **Connect USB webcam to system**
2. **Verify camera detection:**
   ```bash
   lsusb | grep -i camera
   v4l2-ctl --list-devices
   ```

3. **Test camera functionality:**
   ```bash
   # List available cameras
   ls /dev/video*
   
   # Test camera capture
   ffmpeg -f v4l2 -i /dev/video0 -frames:v 1 test.jpg
   ```

4. **Configure camera ID in application:**
   ```bash
   ./object_detection --camera-id 0  # Usually /dev/video0
   ```

### Troubleshooting Camera Issues

**Camera not detected:**
```bash
# Check permissions
sudo usermod -a -G video $USER
# Logout and login again
```

**Low frame rate:**
```bash
# Check camera capabilities
v4l2-ctl --device=/dev/video0 --list-formats-ext
```

**Permission denied:**
```bash
# Add user to video group
sudo usermod -a -G video $USER
```

## Testing

### Automated Tests

```bash
# Run all tests
./scripts/test.sh

# Unit tests only
cd build && make object_detection_tests && ./object_detection_tests
```

### Manual E2E Testing

```bash
# Interactive testing script
./scripts/manual_test.sh
```

### Test Scenarios

1. **Mock Camera Test** - Tests graceful failure handling
2. **Real Webcam Test** - Tests actual camera integration
3. **Performance Test** - Tests different frame rates
4. **Configuration Test** - Tests parameter validation
5. **Logging Test** - Tests log file creation and formatting

## Performance Characteristics

### Target Platforms

| Platform | CPU | Expected Performance |
|----------|-----|---------------------|
| **x86_64** | Intel Core i7 | 5-15 fps @ 720p |
| **x86_64** | AMD Ryzen 5 3600 | 8-20 fps @ 720p |
| **386** | Intel Pentium M | 1-3 fps @ 720p |

### Performance Monitoring

The application automatically monitors and logs:
- **Frames per second** processing rate
- **Average processing time** per frame
- **Performance warnings** when FPS drops below threshold
- **Resource utilization** and bottlenecks

### Optimization Tips

1. **Reduce frame rate** for lower-power systems
2. **Increase confidence threshold** to reduce false positives
3. **Use smaller resolution** if 720p is too demanding
4. **Enable GPU acceleration** if CUDA is available

## Logging Format

### Object Detection Events
```
[INFO] On Tue 23 Sep at 1:00:15PM PT, person entered frame (85% confidence)
[INFO] On Tue 23 Sep at 1:00:45PM PT, car entered frame (92% confidence)
```

### Performance Logs
```
[INFO] On Tue 23 Sep at 1:00:00PM PT, Performance report: FPS: 4.2, processed 42/50 frames (84%)
[WARNING] On Tue 23 Sep at 1:05:00PM PT, Performance warning: processing rate 0.8 fps is below threshold of 1.0 fps
```

### Heartbeat Logs
```
[INFO] On Tue 23 Sep at 1:10:00PM PT, Detection system operational - heartbeat
```

## Deployment

### Standalone Executable

The application is statically linked and can be deployed as a single executable:

```bash
# Copy executable and model files
cp build/object_detection /usr/local/bin/
cp -r models /usr/local/share/object_detection/
```

### Systemd Service

Create `/etc/systemd/system/object-detection.service`:
```ini
[Unit]
Description=Object Detection Service
After=network.target

[Service]
Type=simple
User=detection
ExecStart=/usr/local/bin/object_detection --log-file /var/log/object_detection.log
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### SSH Deployment

The application runs completely headless and can be managed via SSH:
```bash
# Start via SSH
ssh user@target-system '/usr/local/bin/object_detection --verbose &'

# Monitor logs via SSH
ssh user@target-system 'tail -f /var/log/object_detection.log'
```

## Development

### Adding New Object Classes

1. **Update target classes in `object_detector.cpp`:**
```cpp
std::vector<std::string> ObjectDetector::getTargetClasses() {
    return {"person", "car", "truck", "bus", "motorcycle", "bicycle", "cat", "dog", "bird"};
}
```

2. **Ensure model supports the classes** (COCO dataset includes most common objects)

### Extending Detection Logic

1. **Modify `processFrame()` method** for custom detection logic
2. **Update tracking algorithm** in `updateTrackedObjects()`
3. **Add custom logging** in `logObjectEvents()`

### Performance Tuning

1. **Adjust model input size** in `object_detector.hpp`
2. **Optimize post-processing** in `postProcess()` method
3. **Add threading** for parallel processing

## Troubleshooting

### Common Issues

**"Failed to initialize webcam interface"**
- Check camera connection and permissions
- Try different camera ID (0, 1, 2...)
- Verify camera supports requested resolution

**"Failed to load detection model"**
- Download YOLO model file
- Check model path and file permissions
- Verify ONNX model compatibility

**Low performance/frame rate**
- Reduce max-fps setting
- Lower camera resolution
- Increase confidence threshold
- Check CPU usage and system resources

**No objects detected**
- Lower confidence threshold
- Check camera view and lighting
- Verify model file is correct
- Enable verbose logging for debugging

### Debug Mode

```bash
# Enable verbose logging and debugging
./object_detection --verbose --log-file debug.log --max-fps 1

# Monitor debug output
tail -f debug.log
```

## License and Credits

This application uses:
- **OpenCV** for computer vision and camera interface
- **YOLO** models for object detection (requires separate download)
- **C++17** standard library

Designed for deployment on Linux systems with USB webcam support.