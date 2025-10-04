# Smart Photo Storage Implementation Summary

## Issue
**cpp-object-detection: only write image to disk when new types of objects have been detected**

Photos were being continually stored every 10 seconds even when detecting the same stationary object (e.g., a parked car), leading to redundant storage.

## Requirements
1. Only save photos when:
   - New instances of the same object type appear (2nd car enters)
   - A new type of object is detected (person enters in addition to car)
2. Save photo immediately when changes detected (bypass 10s delay)
3. Maintain 10s interval for stationary objects (no changes)

## Implementation

### Changes Made

#### 1. ObjectDetector API Enhancement (`include/object_detector.hpp`)
```cpp
// Added public methods to expose tracking state
const std::vector<ObjectTracker>& getTrackedObjects() const;
void updateTracking(const std::vector<Detection>& detections);
```

**Purpose**: Allow ParallelFrameProcessor to access tracking information to make intelligent photo storage decisions.

#### 2. ParallelFrameProcessor State Tracking (`include/parallel_frame_processor.hpp`)
```cpp
// Added state tracking for last saved photo
std::map<std::string, int> last_saved_object_counts_;
```

**Purpose**: Track what objects and how many of each type were present when the last photo was saved.

#### 3. Smart Photo Storage Logic (`src/parallel_frame_processor.cpp`)

**New Decision Algorithm:**
```cpp
// Count current objects by type
std::map<std::string, int> current_object_counts;

// Check for new types (e.g., person enters when only car was present)
bool has_new_types = false;
for (const auto& [type, count] : current_object_counts) {
    if (last_saved_object_counts_.find(type) == last_saved_object_counts_.end()) {
        has_new_types = true;
        break;
    }
}

// Check for new instances (e.g., 2nd car enters)
bool has_new_objects = false;
for (const auto& [type, count] : current_object_counts) {
    auto it = last_saved_object_counts_.find(type);
    if (it != last_saved_object_counts_.end() && count > it->second) {
        has_new_objects = true;
        break;
    }
}

// Also check object tracker for newly entered objects
const auto& tracked = detector->getTrackedObjects();
for (const auto& obj : tracked) {
    if (obj.is_new && obj.frames_since_detection == 0) {
        has_new_objects = true;
        break;
    }
}

// Save immediately if changes detected, otherwise use 10s interval
bool should_save_immediately = has_new_types || has_new_objects;
bool enough_time_passed = elapsed.count() >= PHOTO_INTERVAL_SECONDS;

if (should_save_immediately || enough_time_passed) {
    // Save photo and update state
    last_saved_object_counts_ = current_object_counts;
    last_photo_time_ = now;
}
```

#### 4. Object Tracking Integration
```cpp
// Update tracking before making photo storage decision
if (!target_detections.empty()) {
    detector_->updateTracking(target_detections);
}

// Then check for changes and decide whether to save
if (!target_detections.empty()) {
    saveDetectionPhoto(frame, target_detections, detector_);
}
```

### Behavior Examples

#### Scenario 1: New Object Type
```
Time: 0s  → Car detected → Photo saved immediately (new type)
Time: 3s  → Car present  → No photo (same objects, < 10s)
Time: 7s  → Person enters → Photo saved immediately (new type: person)
Time: 9s  → Car + Person → No photo (same objects, < 10s)
Time: 17s → Car + Person → Photo saved (10s passed since last photo)
```

#### Scenario 2: New Instance
```
Time: 0s  → 1 car → Photo saved immediately (new type)
Time: 5s  → 1 car → No photo (same count, < 10s)
Time: 8s  → 2 cars → Photo saved immediately (new instance!)
Time: 12s → 2 cars → No photo (same count, < 10s)
Time: 18s → 2 cars → Photo saved (10s passed since last photo)
```

#### Scenario 3: Stationary Object
```
Time: 0s  → Car → Photo saved immediately (new type)
Time: 5s  → Car → No photo (< 10s)
Time: 10s → Car → Photo saved (10s interval)
Time: 15s → Car → No photo (< 10s)
Time: 20s → Car → Photo saved (10s interval)
```

## Testing

Added comprehensive test suite (`tests/test_photo_storage_logic.cpp`):
- Object tracking API tests
- New type detection tests
- New instance detection tests  
- Multiple object tracking tests
- Object state verification tests

**Test count: 7 test cases**

## Documentation

Updated `PHOTO_STORAGE_FEATURE.md` with:
- Smart photo storage behavior description
- Example scenarios with timelines
- Technical implementation details
- Updated processing flow diagram
- Example log output showing immediate saves

## Code Quality

- **Thread-safe**: All photo storage logic protected by mutex
- **Minimal changes**: Only modified necessary files
- **Backward compatible**: Maintains existing 10s interval behavior
- **Well-documented**: Comments explain decision logic
- **Testable**: New methods exposed for testing

## Files Changed

1. `cpp-object-detection/include/object_detector.hpp` - Added tracking access methods
2. `cpp-object-detection/include/parallel_frame_processor.hpp` - Added state tracking
3. `cpp-object-detection/src/parallel_frame_processor.cpp` - Implemented smart logic
4. `cpp-object-detection/tests/test_photo_storage_logic.cpp` - New test suite
5. `cpp-object-detection/tests/CMakeLists.txt` - Added new test
6. `cpp-object-detection/PHOTO_STORAGE_FEATURE.md` - Updated documentation

## Impact

**Before:**
- Photos saved every 10s when objects detected
- Redundant photos of stationary objects
- Delayed response to scene changes

**After:**
- Immediate photo when new object type enters
- Immediate photo when new instance detected
- 10s interval maintained for stationary scenes
- Reduced redundant storage
- Faster response to important changes
