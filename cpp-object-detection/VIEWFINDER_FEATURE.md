# Real-Time Viewfinder Feature

## Overview

The real-time viewfinder provides instant visual feedback during development by displaying detected objects with bounding boxes in an on-screen preview window. This feature accelerates development iteration by eliminating the delay of reviewing detection files on disk.

## Usage

Enable the viewfinder with the `--show-preview` command-line flag:

```bash
# Basic usage with viewfinder
./object_detection --show-preview

# Development mode with higher frame rate
./object_detection --show-preview --max-fps 10

# Combined with verbose logging
./object_detection --show-preview --verbose --min-confidence 0.7
```

## Features

- **Real-time display**: Shows live camera feed with detection overlays
- **Bounding boxes**: Colored boxes around detected objects
- **Confidence labels**: Displays object type and confidence percentage
- **Color-coded classes**: Different colors for different object types:
  - Green: Person
  - Red: Cat
  - Blue: Dog
  - Yellow: Vehicles (car, truck, bus)
  - Magenta: Motorcycles and bicycles
  - White: Other detected objects

- **Interactive control**: Press 'q' or ESC to close the viewfinder and stop the application

## Default Behavior

The viewfinder is **disabled by default** to support headless operation on systems without X11/display support. Enable it explicitly with `--show-preview` when needed.

## Technical Details

### Implementation

The viewfinder is implemented through the `ViewfinderWindow` class which:
1. Creates an OpenCV window when enabled
2. Receives frames and detection results from the processing pipeline
3. Draws bounding boxes with labels on the frames
4. Displays the annotated frames in real-time
5. Monitors for user input to close the window

### Integration with Processing Pipeline

```
┌────────────┐    ┌──────────────┐    ┌─────────────────┐
│  Webcam    │───>│    Frame     │───>│   Detection     │
│  Capture   │    │  Processing  │    │   Results       │
└────────────┘    └──────────────┘    └─────────────────┘
                                               │
                                               │ (if --show-preview)
                                               ▼
                                      ┌─────────────────┐
                                      │   Viewfinder    │
                                      │   Window        │
                                      │  (Live Display) │
                                      └─────────────────┘
```

### Performance Considerations

- The viewfinder adds minimal overhead to the processing pipeline
- Frame display uses non-blocking operations (1ms wait)
- Detection photo saving (every 10 seconds) is independent of viewfinder display
- Both viewfinder display and photo saving use the same bounding box drawing logic

## Use Cases

### Development and Testing

```bash
# Quick iteration during model tuning
./object_detection --show-preview --model-type yolov5l --max-fps 10
```

### Camera Positioning

```bash
# Verify camera field of view
./object_detection --show-preview --camera-id 0
```

### Confidence Threshold Adjustment

```bash
# See impact of different confidence levels in real-time
./object_detection --show-preview --min-confidence 0.3
```

## Requirements

- OpenCV compiled with GUI support (highgui module)
- X11 display server (Linux) or native window system (macOS)
- Not available in truly headless environments without virtual display

## Compatibility

- ✅ Linux with X11
- ✅ macOS with native window system
- ✅ Remote desktop/VNC sessions
- ❌ Headless servers without virtual display
- ❌ Docker containers without X11 forwarding

## Example Output

When running with `--show-preview`, you'll see:

1. **Console log output**:
```
[INFO] On Sat 04 Oct at 02:24:44PM PT, Initializing viewfinder window: Object Detection - Live Preview
[INFO] On Sat 04 Oct at 02:24:44PM PT, Viewfinder window initialized successfully
[INFO] On Sat 04 Oct at 02:24:44PM PT, Real-time viewfinder enabled - press 'q' or ESC to stop
```

2. **On-screen window**: Live camera feed with:
   - Colored bounding boxes around detected objects
   - Labels showing "object_type XX%" (e.g., "person 92%")
   - Smooth real-time updates at configured frame rate

3. **Detection photos**: Still saved to disk every 10 seconds (independent of viewfinder)

## Troubleshooting

### "Failed to initialize viewfinder window"

**Cause**: No display available or OpenCV not compiled with GUI support

**Solutions**:
- Ensure X11 is running (`echo $DISPLAY` should show a value)
- Install OpenCV with highgui support
- Use `--no-headless` flag is NOT required (viewfinder uses separate flag)

### Window doesn't appear

**Cause**: Display server not accessible

**Solutions**:
- Check `DISPLAY` environment variable is set
- Verify X11 forwarding if using SSH: `ssh -X user@host`
- Try setting `DISPLAY=:0` before running

### Performance degradation

**Cause**: Display rendering overhead on slow systems

**Solutions**:
- Reduce frame rate: `--max-fps 3`
- Use faster model: `--model-type yolov5s` or `yolov8n`
- Disable viewfinder for production: remove `--show-preview` flag
