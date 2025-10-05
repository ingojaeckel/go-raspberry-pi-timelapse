# Photo Storage with Bounding Boxes - Feature Documentation

## Overview

This feature automatically saves photos with bounding boxes when objects are detected by the cpp-object-detection application. Photos are saved with timestamped filenames and color-coded bounding boxes for easy review and improvement of the detection system.

## Requirements Implemented

### 1. Photo Recording on Detection ✅
- Photos are automatically saved when target objects enter the frame
- Supports detection of: person, car, truck, bus, motorcycle, bicycle, cat, dog, bird, bear, chair, book

### 2. Timestamped Filenames ✅
**Format:** `YYYY-MM-DD HHMMSS [objects] detected.jpg`

**Examples:**
- `2025-10-04 010000 person detected.jpg`
- `2025-10-04 143522 person cat detected.jpg`
- `2025-10-04 201530 car truck detected.jpg`

### 3. Bounding Boxes ✅
Each detected object has a rectangle drawn around it with:
- Colored border (2px thick)
- Label showing class name and confidence percentage
- Label has colored background matching the border color

### 4. Color Mapping by Object Type ✅

| Object Type | Color | RGB (BGR format) |
|------------|-------|------------------|
| Person | Green | (0, 255, 0) |
| Cat | Red | (0, 0, 255) |
| Dog | Blue | (255, 0, 0) |
| Bird | Cyan | (255, 255, 0) |
| Bear | Dark Cyan/Teal | (0, 128, 128) |
| Car, Truck, Bus | Yellow | (0, 255, 255) |
| Motorcycle, Bicycle | Magenta | (255, 0, 255) |
| Chair | Purple | (128, 0, 128) |
| Book | Orange | (255, 128, 0) |
| Other objects | White | (255, 255, 255) |

### 5. Center Coordinates Logging ✅
**Log format:**
```
detected [class] at coordinates: (x, y) with confidence N%
```

**Example logs:**
```
detected person at coordinates: (640, 360) with confidence 92%
detected cat at coordinates: (320, 240) with confidence 87%
Saved detection photo: detections/2025-10-04 143022 person cat detected.jpg
```

### 6. Smart Photo Storage ✅
Photos are saved intelligently based on detected changes:

**Immediate Save (bypasses 10s limit):**
- When a new object type enters the frame (e.g., person detected when only car was present)
- When a new instance of existing type appears (e.g., 2nd car enters when 1 car was already present)

**Rate-Limited Save (10 second interval):**
- When same stationary objects remain in frame with no changes
- Prevents disk space exhaustion from redundant photos

**🆕 Stationary Object Detection (configurable timeout, default: 120s):**
- When objects have been stationary for longer than timeout period, photo capture stops
- Stationary detection based on average movement over last 10 frames (threshold: ≤10 pixels)
- Automatically resumes photo capture when objects start moving again
- Configurable via `--stationary-timeout N` parameter

**Implementation:**
- Tracks object counts by type from last saved photo
- Compares current detections with previous state
- Uses object tracking to detect new entries
- 🆕 Analyzes position history to determine if objects are stationary
- 🆕 Tracks how long objects have been stationary
- Thread-safe implementation using mutex protection

**Benefits:**
- Immediate response to meaningful changes in scene
- Reduced storage of redundant photos
- Better capture of dynamic events
- 🆕 Massive disk space savings for long-term stationary objects (e.g., parked cars)
- 🆕 Focus on capturing meaningful movement and changes

## Architecture

### Key Components

1. **ParallelFrameProcessor** - Main processing class
   - `saveDetectionPhoto()` - Saves annotated photos
   - `getColorForClass()` - Maps object types to colors
   - `generateFilename()` - Creates timestamped filenames
   - `processFrameInternal()` - Processes frames and triggers photo saving
   - 🆕 Checks stationary status before saving photos

2. **ObjectDetector** - Object tracking and stationary detection
   - `updateStationaryStatus()` - Analyzes movement to determine if object is stationary
   - `isStationaryPastTimeout()` - Checks if stationary object has exceeded timeout period
   - Tracks position history for movement analysis

3. **Configuration**
   - `--output-dir DIR` - Specify output directory (default: `detections`)
   - 🆕 `--stationary-timeout N` - Seconds before stopping photos of stationary objects (default: 120)
   - Automatically creates output directory if it doesn't exist

4. **Thread Safety**
   - Uses mutex (`photo_mutex_`) to protect photo saving
   - Prevents race conditions in multi-threaded mode
   - Safe time tracking for 10-second interval

### Processing Flow

```
Frame Capture → Object Detection → Target Filtering → Update Object Tracking
                                                      ↓
                                  Check for New Objects/Types
                                                      ↓
                      Photo Storage Decision (immediate OR rate-limited)
                                                      ↓
                                  Draw Bounding Boxes + Labels
                                                      ↓
                                  Generate Timestamped Filename
                                                      ↓
                                  Save to Output Directory
                                                      ↓
                                  Update Saved Object Counts
                                                      ↓
                                  Log Success/Failure
```

## Usage Examples

### Basic Usage
```bash
# Uses default output directory "detections"
./object_detection

# Photos will be saved to: ./detections/
```

### Custom Output Directory
```bash
# Specify custom directory
./object_detection --output-dir /path/to/photos

# Photos will be saved to: /path/to/photos/
```

### Production Deployment
```bash
# Security monitoring with high accuracy
./object_detection \
  --model-type yolov5l \
  --min-confidence 0.7 \
  --output-dir /var/security/detections \
  --verbose

# Photos saved to: /var/security/detections/
```

### Testing/Development
```bash
# Fast detection with verbose logging
./object_detection \
  --model-type yolov5s \
  --max-fps 5 \
  --output-dir /tmp/test-detections \
  --verbose

# Photos saved to: /tmp/test-detections/
```

## File System Layout

```
detections/                          # Default output directory
├── 2025-10-04 010000 person detected.jpg
├── 2025-10-04 010015 car detected.jpg
├── 2025-10-04 010030 person cat detected.jpg
├── 2025-10-04 010045 dog detected.jpg
└── ...

/custom/path/                        # Custom output directory (if specified)
├── 2025-10-04 120000 person detected.jpg
├── 2025-10-04 120015 cat detected.jpg
└── ...
```

## Sample Output

### Example Log Output
```
[INFO] On Fri 04 Oct at 2:00:15PM PT, Detection photos will be saved to: detections
[INFO] On Fri 04 Oct at 2:00:16PM PT, Created output directory: detections
[INFO] On Fri 04 Oct at 2:00:20PM PT, detected person at coordinates: (640, 360) with confidence 92%
[INFO] On Fri 04 Oct at 2:00:20PM PT, New object type detected: person
[INFO] On Fri 04 Oct at 2:00:20PM PT, Saving photo immediately due to new objects/types detected
[INFO] On Fri 04 Oct at 2:00:20PM PT, Saved detection photo: detections/2025-10-04 140020 person detected.jpg
[INFO] On Fri 04 Oct at 2:00:25PM PT, detected cat at coordinates: (320, 240) with confidence 87%
[INFO] On Fri 04 Oct at 2:00:25PM PT, New object type detected: cat
[INFO] On Fri 04 Oct at 2:00:25PM PT, Saving photo immediately due to new objects/types detected
[INFO] On Fri 04 Oct at 2:00:25PM PT, Saved detection photo: detections/2025-10-04 140025 person cat detected.jpg
[INFO] On Fri 04 Oct at 2:00:30PM PT, detected person at coordinates: (650, 370) with confidence 94%
[INFO] On Fri 04 Oct at 2:00:30PM PT, detected cat at coordinates: (325, 245) with confidence 89%
# No photo saved - same objects, within 10s interval
[INFO] On Fri 04 Oct at 2:00:35PM PT, detected person at coordinates: (655, 375) with confidence 93%
[INFO] On Fri 04 Oct at 2:00:35PM PT, detected cat at coordinates: (330, 250) with confidence 88%
[INFO] On Fri 04 Oct at 2:00:35PM PT, Saved detection photo: detections/2025-10-04 140035 person cat detected.jpg
# Photo saved after 10s with stationary objects
[INFO] On Fri 04 Oct at 2:00:40PM PT, detected person at coordinates: (660, 380) with confidence 91%
[INFO] On Fri 04 Oct at 2:00:40PM PT, detected cat at coordinates: (335, 255) with confidence 87%
[INFO] On Fri 04 Oct at 2:00:40PM PT, detected car at coordinates: (400, 300) with confidence 85%
[INFO] On Fri 04 Oct at 2:00:40PM PT, New object type detected: car
[INFO] On Fri 04 Oct at 2:00:40PM PT, Saving photo immediately due to new objects/types detected
[INFO] On Fri 04 Oct at 2:00:40PM PT, Saved detection photo: detections/2025-10-04 140040 person cat car detected.jpg
```

### Smart Photo Storage Scenarios

**Scenario 1: New Object Type Enters**
```
Time: 0s  - Car detected → Photo saved immediately (new type)
Time: 3s  - Car still present → No photo (same objects, < 10s)
Time: 7s  - Person enters → Photo saved immediately (new type: person)
Time: 9s  - Car + Person → No photo (same objects, < 10s)
Time: 17s - Car + Person → Photo saved (10s passed since last photo)
```

**Scenario 2: New Instance of Same Type**
```
Time: 0s  - 1 car detected → Photo saved immediately (new type)
Time: 5s  - 1 car still present → No photo (same count, < 10s)
Time: 8s  - 2 cars detected → Photo saved immediately (new instance)
Time: 12s - 2 cars still present → No photo (same count, < 10s)
Time: 18s - 2 cars still present → Photo saved (10s passed since last photo)
```

**Scenario 3: Stationary Objects**
```
Time: 0s  - Car detected → Photo saved immediately (new type)
Time: 5s  - Same car → No photo (< 10s)
Time: 10s - Same car → Photo saved (10s interval)
Time: 15s - Same car → No photo (< 10s)
Time: 20s - Same car → Photo saved (10s interval)
```

### Photo Contents
Each saved photo includes:
- Original camera frame
- Colored bounding boxes around detected objects
- Labels with object type and confidence
- Multiple objects of different types shown in their respective colors

Example visualization:
```
┌─────────────────────────────────────┐
│                                     │
│    ┌─Green Box─────────┐           │
│    │ person 92%        │           │
│    │                   │           │
│    │                   │           │
│    └───────────────────┘           │
│                                     │
│           ┌─Red Box────┐           │
│           │ cat 87%    │           │
│           │            │           │
│           └────────────┘           │
│                                     │
└─────────────────────────────────────┘
```

## Technical Details

### Smart Photo Storage Implementation
```cpp
// Count current object types
std::map<std::string, int> current_object_counts;
for (const auto& detection : detections) {
    current_object_counts[detection.class_name]++;
}

// Check for new types
bool has_new_types = false;
for (const auto& [type, count] : current_object_counts) {
    if (last_saved_object_counts_.find(type) == last_saved_object_counts_.end()) {
        has_new_types = true;
        break;
    }
}

// Check for new instances of existing types
bool has_new_objects = false;
for (const auto& [type, count] : current_object_counts) {
    auto it = last_saved_object_counts_.find(type);
    if (it != last_saved_object_counts_.end() && count > it->second) {
        has_new_objects = true;
        break;
    }
}

// Save photo if new changes detected OR 10s interval passed
bool should_save_immediately = has_new_types || has_new_objects;
bool enough_time_passed = elapsed.count() >= PHOTO_INTERVAL_SECONDS;

if (should_save_immediately || enough_time_passed) {
    // Save photo and update tracking state
    last_saved_object_counts_ = current_object_counts;
    last_photo_time_ = now;
}
```

### Object Tracking Integration
```cpp
// Update tracking before photo decision
detector_->updateTracking(target_detections);

// Check tracker for newly entered objects
const auto& tracked = detector->getTrackedObjects();
for (const auto& obj : tracked) {
    if (obj.is_new && obj.frames_since_detection == 0) {
        has_new_objects = true;  // Trigger immediate save
    }
}
```

### Bounding Box Drawing
```cpp
// Draw rectangle with class-specific color
cv::rectangle(annotated_frame, detection.bbox, color, 2);

// Draw label background and text
cv::rectangle(annotated_frame, text_background_rect, color, cv::FILLED);
cv::putText(annotated_frame, label, text_origin, 
           cv::FONT_HERSHEY_SIMPLEX, 0.5, cv::Scalar(255, 255, 255), 1);
```

### Filename Generation
```cpp
// Format: "YYYY-MM-DD HHMMSS [objects] detected.jpg"
std::ostringstream timestamp;
timestamp << std::put_time(&tm_now, "%Y-%m-%d %H%M%S");

std::ostringstream object_str;
for (const auto& obj_type : object_types) {
    object_str << obj_type << " ";
}
object_str << "detected";

filename = timestamp.str() + " " + object_str.str() + ".jpg";
```

## Benefits

1. **Easy Review** - Timestamped photos make it easy to correlate with logs
2. **Visual Debugging** - Bounding boxes show exactly what was detected and where
3. **Model Improvement** - Photos help identify false positives/negatives
4. **Color Coding** - Instant visual identification of object types
5. **Disk Space Management** - 10-second rate limit prevents storage overflow
6. **Production Ready** - Thread-safe, robust error handling

## Security Considerations

- CodeQL analysis passed with 0 vulnerabilities
- Thread-safe photo saving prevents race conditions
- Rate limiting prevents disk exhaustion attacks
- Directory creation with secure permissions (0755)
- Error handling for file I/O operations

## Future Enhancements (Optional)

Potential improvements that could be added:
- Configurable rate limit interval
- Photo retention policy (auto-delete old photos)
- Compression settings for JPEG quality
- Multiple detection zones with different colors
- Video recording option in addition to photos
- Cloud upload integration (S3, etc.)
