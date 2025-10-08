# Viewfinder Debug Information Feature

## Overview

The viewfinder now displays comprehensive debug information overlay to help monitor performance and track detection statistics in real-time during development and debugging.

## Features

### Performance Metrics
- **Current FPS**: Real-time frames per second
- **Average Processing Time**: Average time (in milliseconds) to process each frame

### Detection Statistics
- **Total Objects Detected**: Cumulative count of all objects detected since startup
- **Total Images Saved**: Number of detection photos saved to disk
- **Top 10 Objects**: Most frequently detected object types with occurrence counts

### System Information
- **Uptime**: Application runtime in HH:MM:SS format
- **Camera Information**: Camera ID, name (if available), and resolution (e.g., "Camera 0: 1280x720")
- **Detection Resolution**: Resolution used for object detection (may differ from camera resolution if scaling is enabled)

## Usage

### Enabling the Viewfinder

The debug information is shown by default when the viewfinder is enabled:

```bash
./object_detection --show-preview
```

### Keyboard Controls

- **SPACE**: Toggle debug information display on/off
  - Default: **ON** (debug info shown)
  - Pressing SPACE hides the overlay for unobstructed view
  - Pressing SPACE again shows the overlay
  
- **q** or **ESC**: Close viewfinder and stop application

## Display Format

The debug information is displayed in the top-left corner with:
- **Small font** (0.4 scale) to minimize screen coverage
- **Semi-transparent black background** (60% opacity) for readability
- **White text** for high contrast
- **Compact layout** with 15-pixel line spacing

### Example Display

```
FPS: 15
Avg proc: 45 ms
Objects: 127
Images: 12
Uptime: 00:15:32
Camera 0: 1280x720
Detection: 640x360
--- Top Objects ---
person: 85
cat: 24
dog: 12
car: 4
bicycle: 2
```

## Technical Implementation

### Data Sources

The debug information is gathered from multiple components:

1. **PerformanceMonitor**
   - `getCurrentFPS()`: Current frames per second
   - `getAverageProcessingTime()`: Average processing time in milliseconds

2. **ObjectDetector**
   - `getTotalObjectsDetected()`: Total objects detected
   - `getTopDetectedObjects(10)`: Top 10 most frequently detected objects

3. **ParallelFrameProcessor**
   - `getTotalImagesSaved()`: Total detection photos saved

4. **ApplicationContext**
   - `start_time`: Application start timestamp for uptime calculation
   - Camera and detection resolution information

### Integration Points

The feature integrates seamlessly with existing code:

```cpp
// In application.cpp - main processing loop
if (ctx.config.show_preview && ctx.viewfinder) {
    ctx.viewfinder->showFrameWithStats(
        ctx.frame, 
        result.detections,
        ctx.perf_monitor->getCurrentFPS(),
        ctx.perf_monitor->getAverageProcessingTime(),
        ctx.detector->getTotalObjectsDetected(),
        ctx.frame_processor->getTotalImagesSaved(),
        ctx.start_time,
        ctx.detector->getTopDetectedObjects(10),
        ctx.config.frame_width,
        ctx.config.frame_height,
        ctx.config.camera_id,
        camera_name,
        ctx.detection_width,
        ctx.detection_height
    );
}
```

## Benefits

### Development Efficiency
- **Real-time feedback**: Instant visibility into performance and detection accuracy
- **Quick debugging**: Identify performance bottlenecks immediately
- **Statistics tracking**: Monitor detection patterns without reviewing log files

### Production Monitoring
- **Performance validation**: Verify FPS and processing time meet requirements
- **Detection verification**: Confirm expected objects are being detected
- **Uptime tracking**: Monitor application stability

### Minimal Impact
- **Small font**: Uses only ~200x300 pixels in top-left corner
- **Toggleable**: Can be hidden with SPACE key when full view is needed
- **No performance overhead**: Statistics already tracked for logging

## Future Enhancements

Potential improvements for future versions:

- Configurable overlay position (top-left, top-right, bottom-left, bottom-right)
- Customizable number of top objects to display
- Additional metrics (memory usage, disk space, network status)
- Graphical FPS/processing time charts
- Color-coded performance indicators (green/yellow/red based on thresholds)
