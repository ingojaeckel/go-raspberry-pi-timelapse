# Viewfinder Feature - Visual Summary

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                    BEFORE: Disk-Only Feedback                       │
│                                                                      │
│  Camera ──> Detection ──> Save Photo ─────────┐                     │
│                           (every 10s)          │                     │
│                                                ▼                     │
│                                           Disk Files                 │
│                                                │                     │
│                                                │ (delayed review)    │
│                                                ▼                     │
│                                         Developer View               │
│                                                                      │
│  Problem: Long feedback loop slows development iteration            │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│              AFTER: Real-Time + Disk Feedback (--show-preview)       │
│                                                                      │
│                         ┌─ Viewfinder Window ◄─┐                    │
│                         │  (real-time display)  │                    │
│                         │  ┌────────────────┐   │                    │
│                         │  │  Live Camera   │   │                    │
│  Camera ──> Detection ──┤  │  + Bounding    │   │ (instant)          │
│                         │  │    Boxes       │   │                    │
│                         │  └────────────────┘   │                    │
│                         │                       │                    │
│                         │  Press 'q' or ESC ────┘                    │
│                         │                                            │
│                         └─ Save Photo ─────────┐                     │
│                            (every 10s)          │                     │
│                                                 ▼                     │
│                                            Disk Files                 │
│                                                                      │
│  Solution: Instant visual feedback accelerates development           │
└─────────────────────────────────────────────────────────────────────┘
```

## User Experience Flow

### Without --show-preview (Default - Headless)
```
$ ./object_detection
[INFO] Object Detection Application Starting
[INFO] Webcam initialized
[INFO] Detection: person at (640, 360) confidence 92%
[INFO] Saved detection photo: detections/2025-10-04 140030 person detected.jpg

(No visual feedback - must check disk files later)
```

### With --show-preview (Development Mode)
```
$ ./object_detection --show-preview
[INFO] Object Detection Application Starting
[INFO] Webcam initialized
[INFO] Viewfinder window initialized successfully
[INFO] Real-time viewfinder enabled - press 'q' or ESC to stop

┌──────────────────────────────────────┐
│   Object Detection - Live Preview    │  ◄── New window appears
├──────────────────────────────────────┤
│                                      │
│    ┌──Green Box────────┐            │
│    │ person 92%        │            │  ◄── Instant visual feedback
│    │                   │            │
│    │                   │            │
│    └───────────────────┘            │
│                                      │
│           ┌─Red Box────┐            │
│           │ cat 87%    │            │
│           │            │            │
│           └────────────┘            │
│                                      │
└──────────────────────────────────────┘

[INFO] Detection: person at (640, 360) confidence 92%
[INFO] Detection: cat at (800, 500) confidence 87%
[INFO] Saved detection photo: detections/2025-10-04 140030 person cat detected.jpg

(Real-time visual + disk files for later review)
```

## Color-Coded Detection Classes

```
┌────────────────────────────────────────────────────────────┐
│                    Bounding Box Colors                      │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  🟢 Green      →  Person                                   │
│  🔴 Red        →  Cat                                      │
│  🔵 Blue       →  Dog                                      │
│  🟡 Yellow     →  Vehicles (car, truck, bus)              │
│  🟣 Magenta    →  Motorcycles, Bicycles                   │
│  ⚪ White      →  Other detected objects                   │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

## Command Line Examples

```bash
# 1. Basic development mode
./object_detection --show-preview

# 2. High frame rate for smooth preview
./object_detection --show-preview --max-fps 10

# 3. Development with verbose logging
./object_detection --show-preview --verbose --min-confidence 0.7

# 4. Testing different models with visual feedback
./object_detection --show-preview --model-type yolov5l

# 5. Camera positioning and field-of-view adjustment
./object_detection --show-preview --camera-id 1

# 6. Production mode (default - no preview)
./object_detection --max-fps 5 --min-confidence 0.5
```

## Feature Comparison

| Aspect                  | Without --show-preview | With --show-preview  |
|-------------------------|------------------------|----------------------|
| Display Mode            | Headless               | GUI Window           |
| Feedback Time           | 10+ seconds (delayed)  | Real-time (~instant) |
| Development Speed       | Slow iteration         | Fast iteration       |
| Resource Usage          | Minimal                | +Small display cost  |
| Requires X11/Display    | No                     | Yes                  |
| Photo Saving            | Yes (every 10s)        | Yes (every 10s)      |
| Logging                 | Yes                    | Yes                  |
| Default Behavior        | ✓ (Enabled)            | ✗ (Disabled)         |

## Interactive Controls

When viewfinder is active:

- **q** - Quit application and close viewfinder
- **ESC** - Quit application and close viewfinder
- **Ctrl+C** - Signal handler for graceful shutdown

## Technical Flow

```
Main Processing Loop (runMainProcessingLoop)
    │
    ├─> Capture Frame from Webcam
    │
    ├─> Submit Frame for Detection
    │
    ├─> Get Detection Results
    │   ├─> Log detections
    │   ├─> Save photo (every 10s)
    │   │
    │   └─> if (config.show_preview) ────┐
    │                                      │
    │   ┌──────────────────────────────────┘
    │   │
    │   ├─> viewfinder->showFrame(frame, detections)
    │   │       │
    │   │       ├─> Draw bounding boxes
    │   │       ├─> Add labels (class + confidence)
    │   │       └─> Display in window
    │   │
    │   └─> Check if user pressed 'q' or ESC
    │           │
    │           └─> if yes: Stop application
    │
    └─> Continue loop
```

## Requirements Checklist

✅ **Requirement 1**: Command line flag to enable real-time preview
   - Implemented: `--show-preview` flag
   - Default: Disabled (false)

✅ **Requirement 2**: Real-time viewfinder acting like camera preview
   - Implemented: ViewfinderWindow class
   - Displays: Live camera feed with detection overlays
   - Interactive: Press 'q' or ESC to close

✅ **Requirement 3**: Viewfinder disabled by default
   - Default: `show_preview = false`
   - Maintains: Headless operation compatibility
   - Optional: Only enabled with explicit flag

## Files Modified/Created

```
cpp-object-detection/
├── include/
│   ├── config_manager.hpp          (+ show_preview flag)
│   ├── application_context.hpp     (+ viewfinder member)
│   └── viewfinder_window.hpp       (NEW - 49 lines)
├── src/
│   ├── config_manager.cpp          (+ flag parsing)
│   ├── application.cpp             (+ viewfinder integration)
│   └── viewfinder_window.cpp       (NEW - 128 lines)
├── tests/
│   ├── test_config_manager.cpp     (+ show_preview tests)
│   └── test_viewfinder_window.cpp  (NEW - 64 lines)
├── README.md                       (+ feature documentation)
├── VIEWFINDER_FEATURE.md           (NEW - complete guide)
└── CMakeLists.txt                  (+ viewfinder_window.cpp)

Total: 11 files changed, 296 lines added
```
