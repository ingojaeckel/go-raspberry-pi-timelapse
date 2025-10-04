# C++ Object Detection Application

A standalone C++ executable for real-time object detection from webcam data at 720p resolution. This application features a pluggable model architecture that allows engineers to select different detection models based on their speed/accuracy requirements.

## Features

- **Real-time object detection** from USB webcam input
- **720p resolution support** with configurable frame rates
- **Headless operation** - no X11 required
- **Object tracking and permanence** - Distinguishes new objects from moving objects
- **Position-based tracking** for people, vehicles, and small animals (cat/dog/fox)
- **Real-time viewfinder** - optional on-screen preview with detection bounding boxes and performance statistics (--show-preview)
  - **Debug overlay** with performance metrics, detection counts, uptime, and top detected objects
  - **Toggle overlay** with SPACE key for minimal screen coverage
- **Network streaming** - MJPEG HTTP streaming to view feed on any device on local network (--enable-streaming)
- **Confidence-based filtering** with configurable thresholds
- **Performance monitoring** with automatic warnings for low frame rates
- **Structured logging** with timestamps and detailed position tracking
- **Configurable parameters** via command-line interface
- **Static linking** for standalone deployment
- **🆕 Pluggable Model Architecture** - Choose between multiple detection models
- **🆕 Speed vs Accuracy Trade-offs** - Select optimal model for your use case
- **🆕 Parallel Processing** - Multi-threaded frame processing support
- **🆕 CPU Rate Limiting** - Energy-efficient analysis with configurable sleep intervals
- **🆕 Detection Scale Factor** - In-memory image downscaling for 4x faster processing

## Object Tracking and Permanence Model

The application implements an enhanced object tracking system that maintains object identity across frames with position history:

### How It Works

1. **Position-Based Tracking**: Objects are tracked using their center (x, y) coordinates and object type (e.g., "cat", "person")
2. **🆕 Position History**: Maintains up to 10 recent positions for each tracked object to analyze movement patterns
3. **🆕 Closest-Match Algorithm**: Finds the best matching object when multiple candidates exist within threshold
4. **Movement Detection**: When an object of the same type is detected in a subsequent frame:
   - If distance from previous position < 100 pixels → Same object (has moved)
   - If distance from previous position > 100 pixels → Different object (new entry)
5. **🆕 Movement Pattern Analysis**: 
   - Calculates average step size from position history
   - Tracks overall displacement and path length
   - Provides detailed movement statistics in debug mode
6. **Smart Logging**: 
   - New objects: `"new cat entered frame at (320, 240)"`
   - Moved objects: `"cat seen earlier moved from (320, 240) -> (325, 245)"`
   - 🆕 Debug mode: Detailed distance calculations and movement patterns

### Configuration

The tracking behavior is controlled by these parameters in the code:

- **MAX_MOVEMENT_DISTANCE**: 100 pixels - Maximum distance an object can move between frames and still be considered the same object
- **Movement threshold**: 5 pixels - Minimum movement to log (avoids noise from detection jitter)
- **Tracking timeout**: 30 frames - Objects not seen for 30 frames are removed from tracking
- **🆕 MAX_POSITION_HISTORY**: 10 positions - Number of recent positions to track for movement analysis

### Example Scenario

```
Frame 1: Detect "cat" at (100, 100) → Log: "new cat entered frame at (100, 100)"
Frame 2: Detect "cat" at (105, 102) → Log: "cat seen earlier moved from (100, 100) -> (105, 102)"
        Debug: "Movement pattern: 2 positions tracked, avg step: 5.4 px"
Frame 3: Detect "cat" at (300, 100) → Log: "new cat entered frame at (300, 100)" (too far, likely different cat)
```

**🆕 Enhanced Debug Output** (with `--verbose` flag):
```
[DEBUG] Processing detection: cat at (105, 102)
[DEBUG]   Distance to existing cat at (100, 100): 5.4 pixels
[DEBUG]   Matched to existing cat (distance: 5.4 pixels)
[DEBUG] Movement analysis: 2 positions in history, average step: 5.4 pixels
[DEBUG] Logging movement: cat moved 5.4 pixels [avg step: 5.4 px, overall path: 5.4 px]
```

For complete details on movement detection improvements, see [MOVEMENT_DETECTION_IMPROVEMENTS.md](MOVEMENT_DETECTION_IMPROVEMENTS.md).

## Model Selection

The application supports multiple detection models with different speed/accuracy characteristics:

| Model    | Speed      | Accuracy | Size | Use Case |
|----------|------------|----------|------|----------|
| YOLOv5s  | Fast (~65ms) | 75%    | 14MB | Real-time monitoring |
| YOLOv5l  | Slow (~120ms)| 85%    | 47MB | High-accuracy security |
| YOLOv8n  | Fastest (~35ms) | 70% | 6MB  | Embedded systems |
| YOLOv8m  | Slowest (~150ms) | 88% | 52MB | Maximum accuracy |

### Model Selection Examples

```bash
# Fast real-time detection (default)
./object_detection --model-type yolov5s --max-fps 5

# High accuracy with reduced frame rate
./object_detection --model-type yolov5l --max-fps 2

# Ultra-fast for embedded systems
./object_detection --model-type yolov8n --max-fps 8

# Maximum accuracy for security applications
./object_detection --model-type yolov8m --max-fps 1
```

## Architecture

> **📖 For comprehensive architecture documentation including sequence diagrams, state machines, and detailed component interactions, see [ARCHITECTURE.md](ARCHITECTURE.md)**

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

### 🆕 Pluggable Model Architecture

The application now features a modular detection model system:

```
                    DetectionModelFactory
                            │
              ┌─────────────┼─────────────┐
              │             │             │
    ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
    │IDetectionModel  │ │IDetectionModel  │ │IDetectionModel  │
    │  (Interface)    │ │  (Interface)    │ │  (Interface)    │
    └─────────────────┘ └─────────────────┘ └─────────────────┘
              │             │             │
    ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
    │   YOLOv5s       │ │   YOLOv5l       │ │  YOLOv8n/m      │
    │  Fast Model     │ │ Accurate Model  │ │ Future Models   │
    │  ~65ms, 75%     │ │ ~120ms, 85%     │ │   Various       │
    └─────────────────┘ └─────────────────┘ └─────────────────┘
```

**Key Benefits:**
- 🔧 **Easy Extension**: Add new models by implementing `IDetectionModel`
- 🎯 **Runtime Selection**: Choose models via `--model-type` parameter
- 📊 **Performance Tracking**: Built-in metrics for speed/accuracy comparison
- 🔄 **Hot Swapping**: Switch models during runtime for testing
- 🏗️ **Clean Architecture**: Decoupled detection logic from application flow

### Model Interface Definition

```cpp
class IDetectionModel {
public:
    virtual bool initialize(const std::string& model_path, ...) = 0;
    virtual std::vector<Detection> detect(const cv::Mat& frame) = 0;
    virtual ModelMetrics getMetrics() const = 0;
    virtual std::string getModelName() const = 0;
    // ... other interface methods
};
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
   - Object detection orchestrator using pluggable models
   - Target class filtering (person, vehicles, animals, furniture, books)
   - **🆕 Enhanced object tracking and permanence model**:
     - Tracks objects frame-to-frame based on position and type
     - Maintains position history (up to 10 recent positions) for movement pattern analysis
     - Distinguishes between new objects entering frame vs. tracked objects moving
     - Uses configurable distance threshold (100 pixels) for movement detection
     - Improved closest-match algorithm for accurate object association
     - Comprehensive debug logging for distance calculations and movement patterns
     - Logs "new [object] entered frame at (x, y)" for new detections
     - Logs "[object] moved from (x1, y1) -> (x2, y2)" with movement statistics
     - See [MOVEMENT_DETECTION_IMPROVEMENTS.md](MOVEMENT_DETECTION_IMPROVEMENTS.md) for details
   - Model switching and performance monitoring

5. **🆕 Detection Model Interface (`detection_model_interface.hpp`)**
   - Abstract interface for pluggable detection models
   - Standardized detection API and metrics
   - Factory pattern for model creation

6. **🆕 YOLO Model Implementations (`yolo_v5_model.hpp/cpp`)**
   - YOLOv5s: Fast model optimized for real-time detection
   - YOLOv5l: High-accuracy model for precision applications
   - Extensible framework for future model types

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

**macOS (Intel-based):**
```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install dependencies
brew update
brew install cmake opencv pkg-config
```

**For testing (optional):**
```bash
# Ubuntu/Debian
sudo apt-get install -y libgtest-dev lcov gcovr

# CentOS/RHEL  
sudo yum install -y gtest-devel lcov gcovr

# macOS
brew install googletest lcov gcovr
```

### Build Instructions

1. **Clone and navigate to the project:**
```bash
cd cpp-object-detection
```

2. **Build the application:**

**Cross-platform (auto-detects OS):**
```bash
./scripts/build.sh
```

**Platform-specific scripts:**
```bash
# Linux
./scripts/build.sh

# macOS (Intel)
./scripts/build-mac.sh
```

3. **Download a YOLO model (required for object detection):**
```bash
# Use curl on macOS or wget on Linux
curl -L -o models/yolov5s.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx
# OR
wget -O models/yolov5s.onnx https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5s.onnx
```

### Cross-Platform Builds

**For x86_64 (Linux/macOS):**
```bash
cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_SYSTEM_PROCESSOR=x86_64
make -j$(nproc)
```

**For 386 (32-bit Linux only):**
```bash
cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_SYSTEM_PROCESSOR=i386 -DCMAKE_C_FLAGS="-m32" -DCMAKE_CXX_FLAGS="-m32"
make -j$(nproc)
```

### Platform-Specific Notes

**macOS:**
- Supports Intel-based Macs (x86_64)
- Uses dynamic linking (static linking not supported the same way as Linux)
- OpenCV installed via Homebrew
- Use `otool -L ./object_detection` to check dependencies

**Linux:**
- Supports both x86_64 and 386 architectures
- Static linking available for standalone deployment
- Use `ldd ./object_detection` to check dependencies
- Headless operation supported (no X11 required)

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
  --analysis-rate-limit N        Maximum images to analyze per second (default: 1.0)
  --min-fps-warning N            FPS threshold for performance warnings (default: 1)
  --log-file FILE                Log file path (default: object_detection.log)
  --heartbeat-interval N         Heartbeat log interval in minutes (default: 10)
  --camera-id N                  Camera device ID (default: 0)
  --frame-width N                Frame width in pixels (default: 1280)
  --frame-height N               Frame height in pixels (default: 720)
  --model-path FILE              Path to ONNX model file
  --detection-scale N            Scale factor for detection (0.0-1.0, default: 0.5)
  --processing-threads N         Number of processing threads (default: 1)
  --enable-gpu                   Enable GPU acceleration if available
  --no-headless                  Disable headless mode (show GUI windows)
  --show-preview                 Show real-time viewfinder with detection bounding boxes
  --enable-streaming             Enable MJPEG HTTP streaming over network (default: disabled)
  --streaming-port N             Port for HTTP streaming server (default: 8080)
```

### Network Streaming

Stream live video with object detection bounding boxes to any device on your local network. Compatible with web browsers, VLC, and other standard video players.

```bash
# Enable network streaming on default port 8080
./object_detection --enable-streaming

# Use custom port
./object_detection --enable-streaming --streaming-port 9000

# Combine with other features
./object_detection --enable-streaming --show-preview --model-type yolov5l
```

**Accessing the stream:**
- The application will display the streaming URL when it starts
- Open the URL in any web browser: `http://<ip-address>:8080/stream`
- Or open in VLC: Media → Open Network Stream → enter the URL

**Example URLs:**
- `http://192.168.1.100:8080/stream`
- `http://10.0.0.5:9000/stream`

For detailed information about network streaming, see [NETWORK_STREAMING_FEATURE.md](NETWORK_STREAMING_FEATURE.md).

### CPU Rate Limiting

The application includes an **analysis rate limiting feature** to reduce CPU usage and energy consumption:

```bash
# Default: analyze 1 image per second (low CPU usage)
./object_detection

# Analyze 10 images per second (higher CPU usage)
./object_detection --analysis-rate-limit 10

# Analyze 0.5 images per second (minimal CPU usage, every 2 seconds)
./object_detection --analysis-rate-limit 0.5
```

**How it works:**
- After each image analysis, the application calculates sleep time: `sleep_time = (1000ms / rate_limit) - processing_time`
- Example: If analysis takes 50ms and rate limit is 1/sec → sleep 950ms
- This allows the CPU to idle between analyses, reducing energy consumption by up to 95%

### Examples

**Energy-efficient monitoring (low CPU):**
```bash
./object_detection --analysis-rate-limit 1 --min-confidence 0.7
```

**Low-power deployment:**
```bash
./object_detection --max-fps 1 --min-confidence 0.8 --heartbeat-interval 5 --analysis-rate-limit 0.5
```

**High-accuracy monitoring:**
```bash
./object_detection --max-fps 10 --min-confidence 0.3 --verbose
```

**🆕 Performance-optimized setup:**
```bash
./object_detection --detection-scale 0.25 --max-fps 10 --min-confidence 0.5
```

**Custom logging:**
```bash
./object_detection --log-file /var/log/security_detection.log --heartbeat-interval 30
```

**Development mode with real-time preview:**
```bash
./object_detection --show-preview --max-fps 10
```

**Network streaming for remote viewing:**
```bash
# Stream to devices on same network
./object_detection --enable-streaming

# Custom streaming port
./object_detection --enable-streaming --streaming-port 9000

# Streaming with high-accuracy model
./object_detection --enable-streaming --model-type yolov5l --min-confidence 0.7
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

**Cross-platform (auto-detects OS):**
```bash
# Run all tests
./scripts/test.sh
```

**Platform-specific scripts:**
```bash
# Linux
./scripts/test.sh

# macOS (Intel)  
./scripts/test-mac.sh

# Unit tests only
cd build && make object_detection_tests && ./tests/object_detection_tests
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
5. **🆕 Adjust detection scale factor** for significant performance improvements

### 🆕 Detection Scale Factor (Performance Optimization)

The application now supports in-memory image downscaling during object detection to dramatically improve performance while maintaining full-resolution storage of detection images.

#### How It Works

- Frames are downscaled **only during object detection** (in-memory)
- Detection images saved to disk remain at **full resolution**
- Bounding boxes are automatically scaled back to original frame dimensions
- Default scale factor: **0.5** (50% reduction = 75% fewer pixels)

#### Performance Impact

| Resolution | Pixels | Detection Time* | Scale Factor |
|------------|--------|----------------|--------------|
| 1280x720 (original) | 921,600 | ~240ms | 1.0 |
| 640x360 (default) | 230,400 | ~60ms | 0.5 |
| 960x540 | 518,400 | ~130ms | 0.75 |
| 320x180 | 57,600 | ~25ms | 0.25 |

*Reference: Intel Core i7 workstation with YOLOv5s model

#### Usage Examples

```bash
# Default: 50% scaling (recommended)
./object_detection

# Aggressive scaling for maximum performance
./object_detection --detection-scale 0.25

# Conservative scaling for better accuracy
./object_detection --detection-scale 0.75

# No scaling (full resolution, slower)
./object_detection --detection-scale 1.0
```

#### Trade-offs

- **Lower values** (e.g., 0.25): Faster processing, may miss small objects
- **Higher values** (e.g., 0.75): Better accuracy, slower processing
- **Default (0.5)**: Optimal balance for most use cases


## Logging Format

### Object Detection Events

The logger now distinguishes between new objects entering the frame and tracked objects moving:

**New Object Entry:**
```
[INFO] On Tue 23 Sep at 1:00:15PM PT, new person entered frame at (320, 240) (85% confidence)
[INFO] On Tue 23 Sep at 1:00:45PM PT, new car entered frame at (450, 180) (92% confidence)
```

**Object Movement:**
```
[INFO] On Tue 23 Sep at 1:00:16PM PT, person seen earlier moved from (320, 240) -> (325, 242) (87% confidence)
[INFO] On Tue 23 Sep at 1:00:47PM PT, car seen earlier moved from (450, 180) -> (470, 185) (91% confidence)
```

**Legacy Format (still supported):**
```
[INFO] On Tue 23 Sep at 1:00:15PM PT, person entered frame (85% confidence)
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
    return {"person", "car", "truck", "bus", "motorcycle", "bicycle", "cat", "dog", "bird", "bear", "chair", "book"};
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