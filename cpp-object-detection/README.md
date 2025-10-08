# C++ Object Detection Application

A standalone C++ executable for real-time object detection from webcam data at 720p resolution. This application features a pluggable model architecture that allows engineers to select different detection models based on their speed/accuracy requirements.

**⚡ Optimized for Long-Term Operation**: This application is designed to run continuously for days, weeks, or months without manual intervention. See [LONG_TERM_OPERATION.md](docs/LONG_TERM_OPERATION.md) for details on memory management, camera resilience, and resource monitoring.

## Features

- **Real-time object detection** from USB webcam input
- **720p resolution support** with configurable frame rates
- **Headless operation** - no X11 required
- **Object tracking and permanence** - Distinguishes new objects from moving objects
- **Position-based tracking** for people, vehicles, and small animals (cat/dog/fox)
- **🆕 Stationary object detection** - Automatically stops taking photos of stationary objects after configurable timeout
- **Real-time viewfinder** - optional on-screen preview with detection bounding boxes and performance statistics (--show-preview)
  - **Debug overlay** with performance metrics, detection counts, uptime, and top detected objects
  - **Toggle overlay** with SPACE key for minimal screen coverage
- **Network streaming** - MJPEG HTTP streaming to view feed on any device on local network (--enable-streaming)
- **🆕 Google Sheets Integration** - Optional cloud logging of detection events to Google Sheets (--enable-google-sheets)
- **🆕 Real-time Notifications** - Get instant alerts when new objects are detected via:
  - **Webhook/Callback URL** - HTTP POST to custom endpoints
  - **Server-Sent Events (SSE)** - Browser-compatible push notifications
  - **File-based** - JSON append to file for simple integrations
  - **Stdio** - Output to stdout for pipeline workflows
- **Confidence-based filtering** with configurable thresholds
- **Performance monitoring** with automatic warnings for low frame rates
- **Structured logging** with timestamps and detailed position tracking
- **Configurable parameters** via command-line interface
- **Static linking** for standalone deployment
- **🆕 Pluggable Model Architecture** - Choose between multiple detection models
- **🆕 Speed vs Accuracy Trade-offs** - Select optimal model for your use case
- **🆕 Parallel Processing** - Multi-threaded frame processing support
- **🆕 CPU Rate Limiting** - Energy-efficient analysis with configurable sleep intervals
- **🆕 GPU Acceleration** - Optional CUDA (Linux) or OpenCL (macOS) backend support for 2-4x faster inference
- **🆕 Burst Mode** - Automatically max out FPS when new objects enter the scene
- **🆕 Detection Scale Factor** - In-memory image downscaling for 4x faster processing
- **🆕 Long-Term Operation Optimizations**:
  - Bounded data structures to prevent memory leaks
  - Automatic camera reconnection and keep-alive
  - System resource monitoring (disk space, CPU temperature)
  - Warnings and logging for low disk space
  - Overflow protection for long-running counters

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

For complete details on movement detection improvements, see [MOVEMENT_DETECTION_IMPROVEMENTS.md](docs/MOVEMENT_DETECTION_IMPROVEMENTS.md).

## Stationary Object Detection

The application includes intelligent stationary object detection to avoid filling disk space with redundant photos of objects that aren't moving.

### How It Works

1. **Movement Tracking**: The system tracks object movement by analyzing position history over the last 10 frames
2. **Stationary Threshold**: Objects with average movement ≤ 10 pixels are considered stationary
3. **Timeout Period**: After objects have been stationary for a configurable period (default: 120 seconds / 2 minutes), photo capture stops
4. **Automatic Resume**: If objects start moving again, photo capture automatically resumes

### Configuration

```bash
# Set stationary timeout to 5 minutes (300 seconds)
./object_detection --stationary-timeout 300

# Set to 30 seconds for frequent updates
./object_detection --stationary-timeout 30

# Default is 120 seconds (2 minutes)
./object_detection
```

### Behavior Examples

**Scenario 1: Stationary Car**
```
Time: 0s   - Car detected → Photo saved (new object)
Time: 10s  - Same car, no movement → Photo saved (10s interval)
Time: 20s  - Same car, no movement → Photo saved (10s interval)
...
Time: 120s - Same car, no movement → Photo saved (10s interval)
Time: 130s - Same car, no movement → Photo SKIPPED (stationary timeout reached)
Time: 140s - Same car, no movement → Photo SKIPPED
...
```

**Scenario 2: Object Starts Moving Again**
```
Time: 0s   - Person detected → Photo saved
Time: 120s - Person stationary → Photo saved (last before timeout)
Time: 130s - Person stationary → Photo SKIPPED (timeout reached)
Time: 150s - Person moves → Photo saved immediately (movement detected)
Time: 160s - Person still moving → Photo saved (10s interval)
```

**Scenario 3: Multiple Objects with Different States**
```
Time: 0s   - Car detected → Photo saved
Time: 120s - Car stationary, person enters → Photo saved (new object)
Time: 130s - Both stationary for < 120s → Photo saved (10s interval)
Time: 250s - Both stationary for > 120s → Photo SKIPPED
```

### Benefits

- **Disk Space Savings**: Avoid hundreds of redundant photos of parked cars or static objects
- **Meaningful Photos**: Only capture photos when something interesting is happening
- **Configurable**: Adjust timeout based on your monitoring needs
- **Smart Resume**: Automatically starts capturing again when movement is detected

## Model Selection

The application supports multiple detection models with different speed/accuracy characteristics:

| Model    | Speed      | Accuracy | Size | Use Case |
|----------|------------|----------|------|----------|
| YOLOv5s  | Fast (~65ms) | 75%    | 14MB | Real-time monitoring |
| YOLOv5l  | Slow (~120ms)| 85%    | 47MB | High-accuracy security |
| YOLOv8n  | Fastest (~35ms) | 70% | 6MB  | Embedded systems |
| YOLOv8m  | Slowest (~150ms) | 88% | 52MB | Maximum accuracy |

> **📊 Exploring Alternative Models?** See [ALTERNATIVE_MODELS_ANALYSIS.md](docs/ALTERNATIVE_MODELS_ANALYSIS.md) for an in-depth analysis of 3 additional model architectures (EfficientDet, Faster R-CNN, DETR) optimized for outdoor scenes, fine-grained classification, and handling occlusions.

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

> **📖 For comprehensive architecture documentation including sequence diagrams, state machines, and detailed component interactions, see [ARCHITECTURE.md](docs/ARCHITECTURE.md)**

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
     - See [MOVEMENT_DETECTION_IMPROVEMENTS.md](docs/MOVEMENT_DETECTION_IMPROVEMENTS.md) for details
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
# Linux x86_64
./scripts/build.sh

# macOS (Intel)
./scripts/build-mac.sh

# Raspberry Pi (ARM64)
./scripts/build-rpi.sh

# 32-bit Linux (x86)
./scripts/build-linux-386.sh
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

**For ARM64 (Raspberry Pi):**
```bash
cmake .. -DCMAKE_BUILD_TYPE=Release
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

**Raspberry Pi:**
- Supports ARM64 architecture (Raspberry Pi 4/5)
- Requires 64-bit Raspberry Pi OS (Bookworm or later)
- Static linking of libgcc/libstdc++ for portability
- Optimized for Cortex-A72/A76 CPU architecture
- Use `ldd ./object_detection` to check dependencies
- Recommended: 4GB+ RAM for optimal performance

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

For detailed information about network streaming, see [NETWORK_STREAMING_FEATURE.md](docs/NETWORK_STREAMING_FEATURE.md).

### Google Sheets Integration

The application supports optional **Google Sheets integration** for cloud-based logging of detection events:

```bash
# Enable Google Sheets logging
./object_detection \
  --enable-google-sheets \
  --google-sheets-id "YOUR_SPREADSHEET_ID" \
  --google-sheets-api-key "YOUR_API_KEY"
```

When enabled, detection events (object entries and movements) are automatically logged to a Google Sheet with:
- Timestamp (ISO 8601 format)
- Object type (person, cat, car, etc.)
- Event type (entry or movement)
- Coordinates and movement distance
- Confidence scores and movement details

**Example Sheet Output:**
| Timestamp | Object Type | Event Type | X | Y | Distance | Description |
|-----------|-------------|------------|---|---|----------|-------------|
| 2024-10-05T14:30:15.123 | person | entry | 320.5 | 240.8 | | Confidence: 87% |
| 2024-10-05T14:30:16.456 | person | movement | 325.2 | 245.3 | 6.8 | From (320,240) to (325,245) |

For detailed setup instructions, security considerations, and troubleshooting, see [GOOGLE_SHEETS_FEATURE.md](docs/GOOGLE_SHEETS_FEATURE.md).

### Real-time Notifications

The application supports **real-time notifications** when new objects are detected, with multiple delivery mechanisms:

```bash
# Webhook notifications (HTTP POST to custom endpoint)
./object_detection \
  --enable-notifications \
  --enable-webhook \
  --webhook-url http://example.com/webhook

# Server-Sent Events (browser-compatible push notifications)
./object_detection \
  --enable-notifications \
  --enable-sse \
  --sse-port 8081

# File-based notifications (append JSON to file)
./object_detection \
  --enable-notifications \
  --enable-file-notification \
  --notification-file-path /var/log/notifications.json

# Stdio notifications (output to stdout for pipelines)
./object_detection \
  --enable-notifications \
  --enable-stdio-notification

# Enable multiple notification channels
./object_detection \
  --enable-notifications \
  --enable-webhook --webhook-url http://example.com/webhook \
  --enable-sse --sse-port 8081 \
  --enable-stdio-notification
```

**Notification Content:**
Each notification includes:
- Detected object type, position, and confidence
- Photo with bounding boxes (base64-encoded JPEG)
- All current detections in frame
- System status (FPS, processing time, total objects, etc.)
- Top detected objects and counts

**Use Cases:**
- **Webhook**: Integration with automation platforms (Zapier, IFTTT, n8n), custom servers, alerting systems
- **SSE**: Web dashboards, real-time browser notifications, mobile apps
- **File**: Log monitoring, file watchers, offline processing, simple integrations
- **Stdio**: Unix pipelines, Docker logging, systemd journal integration

For detailed documentation, examples, and integration guides, see [NOTIFICATION_FEATURE.md](NOTIFICATION_FEATURE.md).

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

### Burst Mode

The **burst mode feature** intelligently adjusts frame rate based on scene activity:

```bash
# Enable burst mode with 1 FPS baseline
./object_detection --enable-burst-mode --analysis-rate-limit 1
```

**How it works:**
- **Normal operation**: Rate limiting applies (e.g., 1 FPS for energy efficiency)
- **New object detected**: Automatically maxes out FPS by removing sleep intervals
- **Objects become stationary**: Returns to normal rate limiting
- **State transitions logged**: Track when burst mode activates/deactivates

**Example scenario:**
```
Idle (no objects)           → 1 FPS (energy efficient)
Person enters scene         → Burst mode ON → ~5-30 FPS (capture detail)
Person stands still         → Burst mode OFF → 1 FPS (back to baseline)
```

See [BURST_MODE_FEATURE.md](docs/BURST_MODE_FEATURE.md) for detailed documentation.

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

**🆕 Burst mode for event-driven capture:**
```bash
./object_detection --enable-burst-mode --analysis-rate-limit 1 --min-confidence 0.6
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

**🆕 GPU-accelerated inference:**
```bash
# Enable GPU acceleration (Linux CUDA or macOS OpenCL)
./object_detection --enable-gpu

# GPU with high-accuracy model for 2-4x speedup
./object_detection --enable-gpu --model-type yolov5l

# GPU with maximum quality settings
./object_detection --enable-gpu --model-type yolov5l --detection-scale 1.0 --max-fps 15
```

See [GPU_ACCELERATION.md](docs/GPU_ACCELERATION.md) for detailed performance benchmarks and platform-specific information.

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

## Hardware Requirements and Platform Support

### Supported Platforms

| Platform | CPU Architecture | Expected Performance | Status |
|----------|-----------------|---------------------|--------|
| **Linux x86_64** | Intel Core i7, AMD Ryzen | 8-15 fps @ 720p | ✅ Fully Supported |
| **Linux 386** | Intel Pentium M | 1-3 fps @ 720p | ✅ Fully Supported |
| **macOS x86_64** | Intel-based Macs | 8-15 fps @ 720p | ✅ Fully Supported |
| **Raspberry Pi 5** | ARM64 (Cortex-A76) | 3-8 fps @ 720p | ✅ Fully Supported |
| **Raspberry Pi 4** | ARM64 (Cortex-A72) | 2-5 fps @ 720p | ✅ Fully Supported |
| **Headless** | Any supported arch | Same as base platform | ✅ Fully Supported |

### Raspberry Pi Support

**Supported Models:**
- **Raspberry Pi 5** (8GB recommended, 4GB minimum): Best performance for real-time detection
- **Raspberry Pi 4** (4GB minimum, 8GB recommended): Good performance for most use cases
- **Raspberry Pi 400**: Same as Raspberry Pi 4

**Operating System:**
- Raspberry Pi OS (64-bit) - Bookworm or later recommended
- Ubuntu Server 22.04 LTS (ARM64)
- Debian 12 (ARM64)

**Building for Raspberry Pi:**
```bash
# Install dependencies
sudo apt-get update
sudo apt-get install -y cmake build-essential libopencv-dev libcurl4-openssl-dev pkg-config

# Build
cd cpp-object-detection
./scripts/build-rpi.sh
```

**Performance Characteristics:**

| Model | RAM | Expected FPS @ 720p | Power Consumption |
|-------|-----|---------------------|-------------------|
| **Raspberry Pi 5** (8GB) | 8GB | 5-8 fps | 5-8W (idle), 8-12W (active) |
| **Raspberry Pi 4** (8GB) | 8GB | 3-5 fps | 3-5W (idle), 6-8W (active) |
| **Raspberry Pi 4** (4GB) | 4GB | 2-4 fps | 3-5W (idle), 6-8W (active) |

**Recommended Settings for Raspberry Pi:**

*Standard Operation (Raspberry Pi 5):*
```bash
./object_detection \
  --max-fps 5 \
  --min-confidence 0.6 \
  --detection-scale 0.5 \
  --analysis-rate-limit 2
```

*Battery-Powered Mode (optimized for low power):*
```bash
./object_detection \
  --max-fps 2 \
  --min-confidence 0.7 \
  --detection-scale 0.5 \
  --analysis-rate-limit 0.5 \
  --heartbeat-interval 15
```

*Raspberry Pi 4 (4GB):*
```bash
./object_detection \
  --max-fps 3 \
  --min-confidence 0.65 \
  --detection-scale 0.5 \
  --frame-width 960 \
  --frame-height 540
```

**Power Consumption Analysis:**

*Raspberry Pi 5 (8GB):*
- **Idle**: ~5W
- **Active Detection (3 fps)**: ~9W
- **Active Detection (5 fps)**: ~11W
- **Peak Load**: ~12W

*Raspberry Pi 4 (8GB):*
- **Idle**: ~3.5W
- **Active Detection (2 fps)**: ~6.5W
- **Active Detection (4 fps)**: ~7.5W
- **Peak Load**: ~8W

**Battery + Solar Panel Feasibility:**

*20,000mAh Battery Bank (74Wh @ 5V):*
- Raspberry Pi 5 @ 3 fps: ~8 hours runtime
- Raspberry Pi 4 @ 2 fps: ~11 hours runtime

*50W Solar Panel (optimal conditions):*
- Can sustain Raspberry Pi 5 @ 3 fps with 2-3 hours sun/day
- Can sustain Raspberry Pi 4 @ 2 fps with 1-2 hours sun/day
- Requires MPPT charge controller and 12V battery system

*Recommended Solar Setup for 24/7 Operation:*
- **Solar Panel**: 100W (12V)
- **Battery**: 100Ah LiFePO4 (12V, 1280Wh)
- **Charge Controller**: 10A MPPT
- **Buck Converter**: 12V → 5V/3A for Raspberry Pi
- **Expected Runtime**: 48-72 hours without sun (Raspberry Pi 4 @ 2 fps)
- **Daily Sun Required**: 2-4 hours for sustained operation

*Cost-Effective 24/7 Solar Setup:*
- Total cost: ~$250-350 USD
- Components: 100W panel ($60), 100Ah battery ($150), MPPT controller ($40), buck converter ($20), cables/mounting ($30-80)
- Breakeven vs. grid power: ~2-3 years in typical residential settings

**Thermal Considerations:**
- Passive cooling sufficient for continuous 2-3 fps operation
- Active cooling (small fan) recommended for 5+ fps on Raspberry Pi 5
- CPU temperature monitoring built into system_monitor.cpp
- Automatic thermal throttling at 80°C (Raspberry Pi firmware)

**Real-World Frame Rate Estimates:**

Based on typical workloads:
- **Raspberry Pi 5**: 3-5 fps sustained, 6-8 fps burst (with proper cooling)
- **Raspberry Pi 4**: 2-3 fps sustained, 4-5 fps burst
- Detection scale factor 0.5 provides best balance of speed vs. accuracy
- GPU acceleration not available (requires CUDA or OpenCL)

**CI/CD Integration:**
- GitHub Actions workflow includes ARM64 build job
- Cross-compilation supported from x86_64 hosts
- Static binary distribution for easy deployment

### 32-bit Linux (386) Specific Notes

**Building for 32-bit Linux:**
```bash
# Install 32-bit build tools
sudo apt-get install -y gcc-multilib g++-multilib

# For 32-bit OpenCV (required for building)
sudo dpkg --add-architecture i386
sudo apt-get update
sudo apt-get install -y libopencv-dev:i386

# Build 32-bit executable
cd cpp-object-detection
./scripts/build-linux-386.sh
```

> **Note:** 32-bit OpenCV libraries (`libopencv-dev:i386`) may not be available in all Ubuntu versions. If you encounter installation issues, you'll need to build on a native 32-bit system or use an older Ubuntu version that provides 32-bit packages.

**Hardware Constraints:**

For older hardware like **Intel Pentium M with 1.5GB RAM**, use these recommended settings:

```bash
# Minimal resource usage configuration
./object_detection \
  --max-fps 1 \
  --min-confidence 0.8 \
  --frame-width 640 \
  --frame-height 480 \
  --analysis-rate-limit 0.5 \
  --heartbeat-interval 15

# Alternative: Ultra-low memory mode (320p)
./object_detection \
  --max-fps 1 \
  --min-confidence 0.85 \
  --frame-width 320 \
  --frame-height 240 \
  --analysis-rate-limit 0.33 \
  --detection-scale 1.0
```

**Key Recommendations for 32-bit / Low-Memory Systems:**
- ✅ Use `--max-fps 1` to limit processing overhead
- ✅ Set `--min-confidence 0.8` or higher to reduce false positives
- ✅ Use `--frame-width 640 --frame-height 480` (VGA) instead of 720p
- ✅ Set `--analysis-rate-limit 0.5` for one analysis every 2 seconds
- ✅ Increase `--heartbeat-interval` to reduce log I/O
- ✅ Use `--detection-scale 0.5` to process downscaled frames (2x speedup)
- ⚠️ Avoid `--enable-gpu` on systems without GPU support
- ⚠️ Avoid `--show-preview` on headless systems

**Memory Usage on 32-bit Systems:**
- Base application: ~50MB
- YOLO model (loaded): ~100MB
- Frame buffers: ~10-20MB
- Total: ~150-170MB (comfortably fits in 1.5GB RAM)

**Dependencies for 32-bit Linux:**
```bash
# On 64-bit system building for 32-bit
sudo dpkg --add-architecture i386
sudo apt-get update
sudo apt-get install -y gcc-multilib g++-multilib libopencv-dev:i386

# On native 32-bit system
sudo apt-get install -y cmake build-essential libopencv-dev pkg-config
```

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