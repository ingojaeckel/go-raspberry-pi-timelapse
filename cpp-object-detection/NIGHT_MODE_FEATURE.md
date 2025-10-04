# Night Mode Detection and Preprocessing - Feature Documentation

## Overview

The night mode feature automatically detects when photos are taken at night or in low-light conditions, and applies intelligent preprocessing to improve object detection accuracy. When night mode is detected, the system saves both the original image and an enhanced version, making it easier to review detections even in challenging lighting conditions.

## Key Features

1. **Automatic Night Detection** - Combines time-of-day analysis with brightness level measurement
2. **CLAHE Preprocessing** - Contrast Limited Adaptive Histogram Equalization for night frames
3. **Dual Photo Storage** - Saves both original and enhanced versions during night mode
4. **Improved Detection** - Enhanced frames are used for object detection to improve confidence at night

## Night Mode Detection Logic

The system determines night mode using two criteria:

### 1. Time-Based Detection
- Night time is defined as 8:00 PM (20:00) to 6:00 AM (6:00)
- Uses local system time to determine current hour

### 2. Brightness-Based Detection
- Calculates average brightness of the frame (0-255 scale)
- Night mode is triggered when brightness < 50 (approximately 20% of max)
- Works even during daytime in very dark environments

### Combined Logic
Night mode is activated when **either**:
- It's night time (20:00-6:00), **OR**
- The frame brightness is below threshold (< 50/255)

## Preprocessing Algorithm

When night mode is detected, frames undergo CLAHE (Contrast Limited Adaptive Histogram Equalization) processing:

### CLAHE Parameters
- **Clip Limit**: 2.0 - Prevents over-amplification of noise
- **Grid Size**: 8x8 - Local histogram equalization for adaptive enhancement
- **Color Space**: LAB - Processes only the lightness channel, preserving color information

### Why CLAHE?
- **Better than standard histogram equalization** - Prevents over-amplification in already bright regions
- **Preserves local contrast** - Adapts to different areas of the image independently
- **Noise-aware** - Clip limit prevents excessive noise amplification
- **Color-preserving** - Works on lightness channel only, maintaining natural colors

## Photo Storage Behavior

### Day Mode (Normal Operation)
```
Frame → Object Detection → Save Photo with Bounding Boxes
```

### Night Mode (Enhanced Operation)
```
Frame → Preprocess (CLAHE) → Object Detection → Save TWO Photos:
  1. Original frame with bounding boxes
  2. Enhanced frame with bounding boxes (marked "night-enhanced")
```

## File Naming Convention

### Regular Detection Photos
```
2025-10-04 183042 person detected.jpg
```

### Night-Enhanced Detection Photos
```
2025-10-04 203042 person detected.jpg                  (original)
2025-10-04 203042 person detected night-enhanced.jpg   (enhanced)
```

The "night-enhanced" suffix is inserted before the `.jpg` extension to clearly identify preprocessed versions.

## Implementation Details

### Key Methods in ParallelFrameProcessor

#### `isNightTime() const`
- Returns `true` if current hour is between 20:00-6:00
- Uses system local time via `localtime_r()`

#### `calculateBrightness(const cv::Mat& frame) const`
- Converts frame to grayscale
- Calculates mean brightness value (0-255)
- Returns average brightness across entire frame

#### `isNightMode(const cv::Mat& frame) const`
- Combines time-based and brightness-based detection
- Returns `true` if either criterion is met
- Logs detection decision for debugging

#### `preprocessForNight(const cv::Mat& frame) const`
- Converts BGR → LAB color space
- Applies CLAHE to L (lightness) channel
- Converts back to BGR
- Returns enhanced frame

### Processing Flow

```cpp
// In processFrameInternal()
bool night_mode = isNightMode(frame);

// Preprocess for detection if night mode
cv::Mat detection_frame = frame;
if (night_mode) {
    detection_frame = preprocessForNight(frame);
}

// Perform detection on enhanced frame
result.detections = detector_->detectObjects(detection_frame);

// Save both versions if detections found
if (!target_detections.empty()) {
    saveDetectionPhoto(frame, target_detections, detector_);
    // ^^ This method now saves both original and enhanced in night mode
}
```

## Benefits

### 1. Better Night Detection
- Preprocessing improves visibility of objects in low light
- CLAHE enhancement increases detection confidence
- Reduces false negatives at night

### 2. Visual Review
- Original photos preserved for reference
- Enhanced photos show what the detector "saw"
- Easy to verify detection accuracy

### 3. Debugging & Improvement
- Compare original vs. enhanced to tune preprocessing
- Identify scenarios where enhancement helps most
- Track detection performance across different lighting conditions

### 4. Automatic Operation
- No manual intervention required
- Adapts to time of day and actual lighting
- Works in both scheduled nighttime and unexpected dark conditions

## Performance Considerations

### Computational Cost
- CLAHE preprocessing adds ~20-50ms per frame
- Only applied when night mode is detected
- Minimal impact on day-time operation

### Storage Impact
- Night mode doubles photo storage (2x files)
- Balanced by improved detection quality
- Storage only increases during night/dark periods (~8 hours/day)

### Memory Usage
- Temporary frame copies for preprocessing
- LAB color space conversion buffers
- All allocations are local to preprocessing function

## Testing

The feature includes comprehensive unit tests:
- Brightness calculation on various frame types
- CLAHE preprocessing validation
- Night mode detection logic
- Frame processing in different lighting conditions
- Multiple frame processing sequences

Run night mode tests:
```bash
cd build/tests
./object_detection_tests --gtest_filter="NightModeTest.*"
```

## Configuration

Currently uses hardcoded values optimized for typical use cases:

- **Night time hours**: 20:00-6:00 (8 PM - 6 AM)
- **Brightness threshold**: 50/255 (~20%)
- **CLAHE clip limit**: 2.0
- **CLAHE grid size**: 8x8

Future enhancement: Make these configurable via command-line arguments.

## Example Log Output

### Night Mode Detected
```
[DEBUG] Night mode detected - time: yes, brightness: 35
[INFO] Applied night mode preprocessing for detection
[INFO] detected cat at coordinates: (320, 240) with confidence 78%
[INFO] Saved detection photo: detections/2025-10-04 203042 cat detected.jpg
[INFO] Saved night-enhanced detection photo: detections/2025-10-04 203042 cat detected night-enhanced.jpg
```

### Day Mode
```
[INFO] detected cat at coordinates: (320, 240) with confidence 85%
[INFO] Saved detection photo: detections/2025-10-04 143042 cat detected.jpg
```

## Integration with Existing Features

### Works With
- ✅ Smart photo storage (10-second rate limiting)
- ✅ Object tracking and permanence
- ✅ Parallel frame processing
- ✅ Multi-threaded operation
- ✅ All detection models (YOLOv5s, YOLOv5l, YOLOv8n, YOLOv8m)

### Thread Safety
- Night mode detection is stateless (no shared state)
- Photo saving uses existing mutex protection
- Preprocessing creates local copies, no race conditions

## Future Enhancements

Potential improvements:
1. **Configurable thresholds** - Command-line args for time ranges and brightness
2. **Adaptive preprocessing** - Adjust CLAHE parameters based on brightness level
3. **Multi-level enhancement** - Different strategies for different darkness levels
4. **Statistics tracking** - Count night vs. day detections, avg brightness
5. **ISO/exposure detection** - Use EXIF data if available from camera
6. **Dawn/dusk handling** - Smooth transitions at twilight hours

## Security Considerations

- No new external dependencies
- Preprocessing uses standard OpenCV functions
- File naming prevents path traversal (no user input in filenames)
- Memory allocations are bounded by frame size
- No network or file I/O beyond existing photo storage

## Backward Compatibility

- Fully backward compatible with existing code
- Day-time operation unchanged
- Existing photo storage features preserved
- No changes to command-line interface
- No database or configuration file changes required
