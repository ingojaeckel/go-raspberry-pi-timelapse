# Viewfinder Debug Information - Architecture

## Data Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Main Processing Loop                        │
│                         (application.cpp)                            │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
                            ├─> Capture frame from webcam
                            │
                            ├─> Process frame (detect objects)
                            │
                            ├─> Gather statistics from components:
                            │   │
                            │   ├─> PerformanceMonitor::getCurrentFPS()
                            │   ├─> PerformanceMonitor::getAverageProcessingTime()
                            │   ├─> ObjectDetector::getTotalObjectsDetected()
                            │   ├─> ObjectDetector::getTopDetectedObjects(10)
                            │   ├─> ParallelFrameProcessor::getTotalImagesSaved()
                            │   └─> ApplicationContext::start_time (for uptime)
                            │
                            └─> ViewfinderWindow::showFrameWithStats()
                                │
                                ├─> Draw bounding boxes
                                │
                                ├─> if (show_debug_info_) {
                                │       drawDebugInfo()
                                │       │
                                │       ├─> Calculate uptime
                                │       ├─> Format statistics
                                │       ├─> Draw semi-transparent background
                                │       └─> Render text overlay
                                │   }
                                │
                                └─> Display frame in window
```

## Statistics Tracking

### ObjectDetector

```cpp
class ObjectDetector {
    // Statistics tracking
    int total_objects_detected_;           // Incremented when new object enters frame
    std::map<std::string, int> object_type_counts_;  // Count per object type
    
public:
    int getTotalObjectsDetected() const;
    std::vector<std::pair<std::string, int>> getTopDetectedObjects(int top_n) const;
};
```

**Update Logic:**
- When a new object is detected (not an existing tracked object):
  - `total_objects_detected_++`
  - `object_type_counts_[detection.class_name]++`

### ParallelFrameProcessor

```cpp
class ParallelFrameProcessor {
    int total_images_saved_;  // Incremented when photo is saved to disk
    
public:
    int getTotalImagesSaved() const;
};
```

**Update Logic:**
- After successfully saving detection photo:
  - `total_images_saved_++`

### ApplicationContext

```cpp
struct ApplicationContext {
    std::chrono::steady_clock::time_point start_time;  // Set during initialization
    int detection_width;   // Calculated from camera width * scale factor
    int detection_height;  // Calculated from camera height * scale factor
};
```

**Update Logic:**
- On initialization:
  - `start_time = std::chrono::steady_clock::now()`
  - `detection_width = frame_width * detection_scale_factor`
  - `detection_height = frame_height * detection_scale_factor`

## Debug Overlay Layout

```
┌────────────────────────────────────────────────────────┐
│ FPS: 15                                                │
│ Avg proc: 45 ms                                        │
│ Objects: 127                                           │
│ Images: 12                                             │
│ Uptime: 00:15:32                                       │
│ Camera 0: 1280x720                                     │
│ Detection: 640x360                                     │
│ --- Top Objects ---                                    │
│ person: 85                                             │
│ cat: 24                                                │
│ dog: 12                                                │
│ car: 4                                                 │
│ bicycle: 2                                             │
│                                                        │
│                                                        │
│                  [Video feed with                      │
│                   bounding boxes]                      │
│                                                        │
│                                                        │
│                                                        │
└────────────────────────────────────────────────────────┘
```

**Rendering Details:**
- Font: `cv::FONT_HERSHEY_SIMPLEX`
- Font scale: `0.4` (small)
- Font thickness: `1`
- Text color: White (255, 255, 255)
- Background: Semi-transparent black (60% opacity)
- Line spacing: 15 pixels
- Padding: 5 pixels

## Toggle Mechanism

The debug info can be toggled on/off with the SPACE key:

```cpp
bool ViewfinderWindow::shouldClose() {
    int key = cv::waitKey(1);
    
    // Toggle debug info with space key
    if (key == ' ' || key == 32) {
        show_debug_info_ = !show_debug_info_;
        logger_->info(show_debug_info_ ? "Debug info enabled" : "Debug info disabled");
    }
    
    return (key == 'q' || key == 27);  // q or ESC to close
}
```

**Default State:** ON (debug info shown by default)

## Performance Impact

The debug overlay has minimal performance impact:

1. **Statistics Collection:** O(1) operations (simple counters and getters)
2. **Top Objects Sorting:** O(n log n) where n is number of unique object types (typically < 20)
3. **Text Rendering:** Small overhead from OpenCV putText (~1-2ms on typical hardware)
4. **Total overhead:** < 5ms per frame (negligible compared to detection time of 30-100ms)

## Testing

Unit tests verify:
- Statistics tracking (getTotalObjectsDetected, getTotalImagesSaved, getTopDetectedObjects)
- Display method doesn't crash with various input combinations
- Toggle mechanism works correctly

Integration would require:
- GUI environment with X11/display support
- Actual camera device
- Object detection model files
