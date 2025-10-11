# Scene Persistence Feature

## Overview

The Scene Persistence feature identifies and tracks stationary object configurations (scenes) in the camera view, storing them in a SQLite database for later recognition. This enables the application to recognize when the camera returns to a previously seen scene, even with slight variations in camera angle or object positions.

## Problem Statement

In object detection applications, it's useful to identify distinct "scenes" - stable configurations of objects in the camera's field of view. For example:
- A desk with a laptop and coffee mug
- A parking lot with several cars
- A room with furniture in specific positions

By persisting these scenes, the application can:
- Recognize when the camera returns to a known location/view
- Track how often different scenes are visited
- Detect when the camera has moved to a new, unseen location

## How It Works

### Scene Analysis Process

1. **Observation Period** (default: 60 seconds)
   - The system observes stationary objects in the frame
   - Collects information about object types, positions, and properties
   - Only stationary objects (not moving) are considered part of the scene

2. **Feature Extraction**
   - **Object Properties**: Type (person, car, etc.), position, orientation
   - **Dominant Color**: Average color of the object region
   - **Spatial Relationships**: Distances and angles between object pairs
   - **Bounding Boxes**: Size and location of each object

3. **Scene Representation**
   - A scene is defined as a tuple of:
     - List of stationary object types
     - Positions of objects relative to frame
     - Spatial relationships (distances/angles between objects)
     - Basic properties (orientation, color)

4. **Scene Matching**
   - When a new scene is analyzed, it's compared against stored scenes
   - **Fuzzy Matching** accounts for:
     - Slight camera angle variations
     - Small position changes
     - Minor object count differences
   - Matching based on:
     - Object type similarity (70% by default)
     - Spatial relationship similarity (30% by default)
     - Relative distances between objects
     - Angular relationships

5. **Database Persistence**
   - New scenes are stored in SQLite with:
     - Unique ID
     - Creation timestamp
     - Description (e.g., "2x person, 1x car arranged in frame")
     - Object details and spatial relationships

## Usage

### Basic Usage

Enable scene persistence with the `--enable-scene-persistence` flag:

```bash
./object_detection --enable-scene-persistence
```

This will:
- Create a `scenes.db` SQLite database in the current directory
- Observe scenes for 60 seconds before analysis
- Use default matching thresholds

### Custom Configuration

```bash
./object_detection \
  --enable-scene-persistence \
  --scene-db-path /path/to/scenes.db \
  --scene-observation-seconds 30 \
  --scene-match-threshold 0.8 \
  --scene-position-tolerance 0.2
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `--enable-scene-persistence` | disabled | Enable scene persistence feature |
| `--scene-db-path` | `scenes.db` | Path to SQLite database file |
| `--scene-observation-seconds` | `60` | Observation time before analyzing scene |
| `--scene-match-threshold` | `0.7` | Minimum similarity score (0-1) to match scenes |
| `--scene-position-tolerance` | `0.15` | Position tolerance as fraction of frame size |

### Example Output

When a new scene is identified:
```
[2025-10-11 15:30:00] INFO: New scene was identified: id=1 - 2x person, 1x car arranged in frame
```

When returning to a known scene:
```
[2025-10-11 16:45:00] INFO: Recognised return to earlier scene: id=1 (match score: 0.85)
```

## Configuration Guide

### Adjusting Match Accuracy

The `--scene-match-threshold` parameter controls how similar scenes must be to match:

- **0.5 - 0.6**: Very loose matching - matches scenes even with significant differences
- **0.7 - 0.8**: Balanced (default) - good for typical use cases with some camera movement
- **0.9 - 1.0**: Strict matching - requires very similar scenes, sensitive to small changes

### Position Tolerance

The `--scene-position-tolerance` parameter controls how much objects can move:

- **0.1**: Strict - objects must be very close to original positions (10% of frame)
- **0.15**: Balanced (default) - allows moderate position variation
- **0.25**: Loose - allows significant position changes

### Observation Time

Shorter observation times (10-30 seconds) enable faster scene detection but may be less stable. Longer times (60-120 seconds) ensure objects are truly stationary.

## Model Recommendations

### For Scene Analysis

The scene persistence feature works with any YOLO model, but different models offer trade-offs:

#### Recommended: YOLOv5s (Default)
- **Speed**: ~65ms per frame
- **Accuracy**: 75%
- **Best for**: Real-time scene analysis with good balance
- **Command**: `--model-type yolov5s`

#### High Accuracy: YOLOv5l
- **Speed**: ~120ms per frame
- **Accuracy**: 85%
- **Best for**: Precise object identification and positioning
- **Command**: `--model-type yolov5l`

#### Embedded Systems: YOLOv8n
- **Speed**: ~35ms per frame
- **Accuracy**: 70%
- **Best for**: Resource-constrained devices (Raspberry Pi)
- **Command**: `--model-type yolov8n`

### Model Selection Guidelines

1. **For Color and Orientation Analysis**: Use YOLOv5l or YOLOv8m for better feature extraction
2. **For Speed**: Use YOLOv5s or YOLOv8n with `--detection-scale 0.5` for faster processing
3. **For Embedded Devices**: YOLOv8n with reduced frame size (640x480)

## Database Schema

### Tables

#### `scenes`
- `id`: Primary key
- `created_at`: Timestamp when scene was first identified
- `description`: Human-readable scene description
- `object_count`: Number of objects in scene
- `object_types`: Summary of object types (e.g., "2x person, 1x car")

#### `scene_objects`
- `id`: Primary key
- `scene_id`: Foreign key to scenes table
- `object_type`: Object type (person, car, etc.)
- `position_x`, `position_y`: Object center position
- `orientation`: Orientation angle in degrees
- `color_r`, `color_g`, `color_b`: Dominant color (RGB)
- `bbox_x`, `bbox_y`, `bbox_width`, `bbox_height`: Bounding box

#### `spatial_relationships`
- `id`: Primary key
- `scene_id`: Foreign key to scenes table
- `object1_idx`, `object2_idx`: Indices of object pair
- `distance`: Euclidean distance between objects
- `angle`: Angle between objects in degrees

## Integration with Other Features

### Compatible Features

- ✅ **Google Sheets**: Scene IDs can be logged to sheets
- ✅ **Notifications**: Can notify on new scene identification
- ✅ **Burst Mode**: Works independently
- ✅ **Brightness Filter**: Helps with consistent color analysis
- ✅ **Stationary Detection**: Uses existing stationary object tracking

### Example: Scene Persistence + Notifications

```bash
./object_detection \
  --enable-scene-persistence \
  --enable-notifications \
  --enable-webhook \
  --webhook-url http://example.com/scene-webhook
```

## Performance Considerations

### Resource Usage

- **Memory**: Minimal (scenes stored in database, not RAM)
- **CPU**: Scene analysis happens once per observation period
- **Disk**: Database grows with new scenes (~1KB per scene)
- **I/O**: Minimal (1 write per scene, occasional reads for matching)

### Recommendations

1. **Observation Period**: Balance between responsiveness and stability
   - Shorter (30s): More responsive, may detect transient configurations
   - Longer (120s): More stable, only captures truly stationary scenes

2. **Database Maintenance**: Periodically backup or clean old scenes
   ```bash
   # Backup
   cp scenes.db scenes_backup_$(date +%Y%m%d).db
   ```

3. **Performance Impact**: Negligible - analysis happens in background after observation period

## Troubleshooting

### Scene Not Being Created

**Problem**: No scenes are being identified

**Solutions**:
1. Ensure objects are stationary for the full observation period
2. Check that `--enable-scene-persistence` flag is set
3. Verify database path is writable
4. Check logs for stationary object detections

### Too Many False Matches

**Problem**: Different scenes are being matched as the same

**Solutions**:
1. Increase `--scene-match-threshold` (try 0.8 or 0.9)
2. Decrease `--scene-position-tolerance` (try 0.1)
3. Increase observation time for more stable scene capture

### Too Few Matches

**Problem**: Same scene is being identified as new

**Solutions**:
1. Decrease `--scene-match-threshold` (try 0.6 or 0.5)
2. Increase `--scene-position-tolerance` (try 0.2 or 0.25)
3. Check if lighting/color changes are affecting matching

## Future Enhancements

Potential improvements for future versions:

- **Semantic Grouping**: Group similar scenes (e.g., "office scenes", "outdoor scenes")
- **Time-based Analysis**: Track which scenes occur at what times
- **Scene Transitions**: Detect and log transitions between scenes
- **Machine Learning**: Use ML to improve scene matching
- **Sub-part Detection**: Identify and match object sub-components
- **Advanced Color Analysis**: Use color histograms for better matching

## References

- Object tracking: `object_detector.cpp` - `ObjectTracker` struct
- Stationary detection: `HOURLY_SUMMARY_FEATURE.md`
- Configuration: `config_manager.hpp`
