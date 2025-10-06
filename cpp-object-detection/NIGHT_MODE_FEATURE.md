# Night Mode Detection and Preprocessing - Feature Documentation

## Overview

The night mode feature automatically detects when photos are taken at night or in extremely low-light conditions, and applies intelligent preprocessing to improve object detection accuracy. When night mode is detected, the system saves both the original image and an enhanced version, while also displaying the enhanced version in the viewfinder and network stream for better visual feedback.

## Key Features

1. **Automatic Night Detection** - Combines time-of-day analysis with brightness level measurement
2. **Enhanced CLAHE Preprocessing** - Stronger Contrast Limited Adaptive Histogram Equalization for extremely dark frames
3. **Dual Photo Storage** - Saves both original and enhanced versions during night mode
4. **Real-time Visual Feedback** - Displays preprocessed frames in viewfinder and network stream with "NIGHT MODE" indicator
5. **Improved Detection** - Enhanced frames are used for object detection to improve confidence at night

## Night Mode Detection Logic

The system determines night mode using a strict dual-criteria approach:

### 1. Time-Based Detection
- Night time is defined as 8:00 PM (20:00) to 6:00 AM (6:00)
- Uses local system time to determine current hour

### 2. Brightness-Based Detection
- Calculates average brightness of the frame (0-255 scale)
- Night mode is triggered when brightness < 15 (approximately 6% of max)
- Ensures only extremely dark images are processed

### Combined Logic
Night mode is activated when **both**:
- It's night time (20:00-6:00), **AND**
- The frame brightness is below threshold (< 15/255)

This conservative approach prevents degradation of object detection on images with poor but usable visibility.

## Preprocessing Algorithm

When night mode is detected, frames undergo enhanced CLAHE (Contrast Limited Adaptive Histogram Equalization) processing:

### Enhanced CLAHE Parameters
- **Clip Limit**: 20.0 - Stronger enhancement for extremely dark images
- **Grid Size**: 8x8 - Local histogram equalization for adaptive enhancement
- **Color Space**: LAB - Processes only the lightness channel, preserving color information
- **Additional Boost**: 1.5x contrast multiplier + 30 brightness offset

### Why Enhanced CLAHE?
- **Significantly better for nearly black images** - Produces ~4x brighter images (13/255 → 50/255)
- **Preserves local contrast** - Adapts to different areas of the image independently
- **Noise-aware** - Clip limit prevents excessive noise amplification
- **Color-preserving** - Works on lightness channel only, maintaining natural colors
- **Detection-optimized** - Makes objects visible enough for effective detection

## Photo Storage Behavior

### Day Mode (Normal Operation)
```
Frame → Object Detection → Save Photo with Bounding Boxes
```

### Night Mode (Enhanced Operation)
```
Frame → Enhanced CLAHE Preprocessing → Object Detection → Save TWO Photos:
  1. Original frame with bounding boxes
  2. Enhanced frame with bounding boxes (marked "night-enhanced")
```

## Real-Time Visual Feedback

When night mode is active:
- **Viewfinder Display**: Shows the CLAHE-preprocessed frame instead of the original
- **Network Stream**: Broadcasts the CLAHE-preprocessed frame instead of the original
- **Night Mode Indicator**: A yellow "NIGHT MODE" label appears in the top-right corner of both displays
- **Visual Consistency**: Users see exactly what the object detector analyzed, making it easier to debug and verify detection accuracy

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
- Returns `true` if both criteria are met (AND logic)
- Logs detection decision for debugging

#### `preprocessForNight(const cv::Mat& frame) const`
- Converts BGR → LAB color space
- Applies enhanced CLAHE to L (lightness) channel with clip limit 20.0
- Applies additional brightness boost (1.5x contrast + 30 offset)
- Converts back to BGR
- Returns significantly enhanced frame

### Processing Flow

```cpp
// In processFrameInternal()
bool night_mode = isNightMode(frame);
result.night_mode_active = night_mode;

// Preprocess for detection if night mode
cv::Mat detection_frame = frame;
if (night_mode) {
    detection_frame = preprocessForNight(frame);
    result.preprocessed_frame = detection_frame.clone();
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

### 1. Significantly Better Night Detection
- Enhanced preprocessing produces ~4x brighter images (13/255 → 50/255) for effective object detection
- Reduces false negatives in extremely dark conditions
- Improves detection confidence scores at night

### 2. Conservative Triggering
- Only activates for nearly black images (time AND extreme darkness)
- Avoids degradation in poor but usable visibility
- Day-time operation completely unaffected

### 3. Visual Review
- Original photos preserved for reference
- Enhanced photos show what the detector "saw"
- Easy to verify detection accuracy
- Side-by-side comparison available

### 4. Real-Time Feedback
- Viewfinder and network stream show what the detector analyzed
- Yellow "NIGHT MODE" indicator makes night mode status obvious
- Helps with debugging and verification during operation

### 5. Automatic Operation
- No manual intervention required
- Adapts to time of day and actual lighting
- Works in both scheduled nighttime and unexpected dark conditions

## Performance Considerations

### Computational Cost
- Enhanced CLAHE preprocessing adds ~20-50ms per frame
- Only applied when night mode is detected
- Minimal impact on day-time operation
- Detection runs on brightened frame for better accuracy

### Storage Impact
- Night mode doubles photo storage (2x files)
- Balanced by improved detection quality
- Storage only increases during night/dark periods (~8 hours/day typically)

### Memory Usage
- Temporary frame copies for preprocessing
- LAB color space conversion buffers
- All allocations are local to preprocessing function
- Preprocessed frame stored in FrameResult for display

## Testing

The feature includes comprehensive unit tests covering:
- Brightness calculation on various frame types
- Enhanced CLAHE preprocessing validation
- Night mode detection logic
- Frame processing in different lighting conditions
- Multiple frame processing sequences
- Empty frame handling

Run night mode tests:
```bash
cd build/tests
./object_detection_tests --gtest_filter="NightModeTest.*"
```

## Integration with Existing Features

### Works With
- ✅ Smart photo storage (10-second rate limiting)
- ✅ Object tracking and permanence
- ✅ Stationary object detection
- ✅ Parallel frame processing
- ✅ Multi-threaded operation
- ✅ Brightness filter for high-brightness conditions (mutually exclusive)
- ✅ All detection models (YOLOv5s, YOLOv5l, YOLOv8n, YOLOv8m)

### Thread Safety
- Night mode detection is stateless (no shared state)
- Photo saving uses existing mutex protection
- Preprocessing creates local copies, no race conditions
- Preprocessed frame stored in result structure

## Example Log Output

### Night Mode Detected (extremely dark image)
```
[DEBUG] Night mode detected - time: yes, brightness: 12
[DEBUG] Applied night mode preprocessing for detection
[INFO] detected cat at coordinates: (320, 240) with confidence 78%
[INFO] Saved detection photo: detections/2025-10-04 203042 cat detected.jpg
[INFO] Saved night-enhanced detection photo: detections/2025-10-04 203042 cat detected night-enhanced.jpg
```

### Day Mode or Usable Visibility
```
[INFO] detected cat at coordinates: (320, 240) with confidence 85%
[INFO] Saved detection photo: detections/2025-10-04 143042 cat detected.jpg
```

## Future Enhancements

Potential improvements that could be added:
1. **Configurable thresholds** - Command-line args for time ranges and brightness
2. **Adaptive preprocessing** - Adjust CLAHE parameters based on darkness level
3. **Multi-level enhancement** - Different strategies for different darkness levels
4. **Statistics tracking** - Count night vs. day detections, avg brightness
5. **ISO/exposure detection** - Use EXIF data if available from camera
6. **Dawn/dusk handling** - Smooth transitions at twilight hours
7. **IR camera integration** - Hardware-based low-light capture (requires Pi NoIR or similar)

## Security Considerations

- No new external dependencies
- Preprocessing uses standard OpenCV functions
- File naming prevents path traversal (no user input in filenames)
- Memory allocations are bounded by frame size
- No network or file I/O beyond existing photo storage

## Backward Compatibility

- ✅ Fully backward compatible with existing code
- ✅ Day-time operation unchanged
- ✅ Existing photo storage features preserved
- ✅ No changes to command-line interface
- ✅ No database or configuration file changes required
- ✅ Thread-safe implementation
- ✅ No API changes
