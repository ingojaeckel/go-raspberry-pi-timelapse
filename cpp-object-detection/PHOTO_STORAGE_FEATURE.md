# Photo Storage with Bounding Boxes - Feature Documentation

## Overview

This feature automatically saves photos with bounding boxes when objects are detected by the cpp-object-detection application. Photos are saved with timestamped filenames and color-coded bounding boxes for easy review and improvement of the detection system.

## Requirements Implemented

### 1. Photo Recording on Detection ✅
- Photos are automatically saved when target objects enter the frame
- Supports detection of: person, car, truck, bus, motorcycle, bicycle, cat, dog

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
| Car, Truck, Bus | Yellow | (0, 255, 255) |
| Motorcycle, Bicycle | Magenta | (255, 0, 255) |
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

### 6. Rate Limiting (10 seconds) ✅
- Maximum of 1 photo saved every 10 seconds
- Prevents disk space exhaustion
- Thread-safe implementation using mutex protection

## Architecture

### Key Components

1. **ParallelFrameProcessor** - Main processing class
   - `saveDetectionPhoto()` - Saves annotated photos
   - `getColorForClass()` - Maps object types to colors
   - `generateFilename()` - Creates timestamped filenames
   - `processFrameInternal()` - Processes frames and triggers photo saving

2. **Configuration**
   - `--output-dir DIR` - Specify output directory (default: `detections`)
   - Automatically creates output directory if it doesn't exist

3. **Thread Safety**
   - Uses mutex (`photo_mutex_`) to protect photo saving
   - Prevents race conditions in multi-threaded mode
   - Safe time tracking for 10-second interval

### Processing Flow

```
Frame Capture → Object Detection → Target Filtering → Photo Storage (if rate limit passed)
                                                      ↓
                                  Draw Bounding Boxes + Labels
                                                      ↓
                                  Generate Timestamped Filename
                                                      ↓
                                  Save to Output Directory
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
[INFO] On Fri 04 Oct at 2:00:20PM PT, Saved detection photo: detections/2025-10-04 140020 person detected.jpg
[INFO] On Fri 04 Oct at 2:00:25PM PT, detected cat at coordinates: (320, 240) with confidence 87%
[INFO] On Fri 04 Oct at 2:00:30PM PT, detected person at coordinates: (650, 370) with confidence 94%
[INFO] On Fri 04 Oct at 2:00:30PM PT, Saved detection photo: detections/2025-10-04 140030 person detected.jpg
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

### Rate Limiting Implementation
```cpp
// Thread-safe 10-second interval check
std::lock_guard<std::mutex> lock(photo_mutex_);
auto now = std::chrono::steady_clock::now();
auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(now - last_photo_time_);

if (elapsed.count() < PHOTO_INTERVAL_SECONDS) {
    return;  // Skip this photo
}
last_photo_time_ = now;
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
