# Scene Persistence - Quick Reference

## Quick Start

```bash
# Enable scene persistence
./object_detection --enable-scene-persistence

# With custom database path
./object_detection --enable-scene-persistence --scene-db-path /path/to/scenes.db
```

## CLI Options

| Flag | Description | Default |
|------|-------------|---------|
| `--enable-scene-persistence` | Enable scene recognition and persistence | disabled |
| `--scene-db-path PATH` | Path to SQLite database | scenes.db |

## How It Works

1. **Observation Period**: System observes stationary objects for 60 seconds
2. **Analysis**: Extracts colors, positions, sizes, relationships
3. **Fingerprinting**: Creates unique fingerprint (object counts + spatial distribution + relationships)
4. **Matching**: Compares against existing scenes (75% similarity threshold)
5. **Logging**: Logs "new scene" or "recognized return to scene"

## Fuzzy Matching Weights

- **40%** - Object type counts (e.g., 2 cars, 1 person)
- **40%** - Spatial distribution (4x4 grid histogram)
- **20%** - Relationships (distances/angles between objects)

## Recommended Models

| Model | Accuracy | Speed | Use Case |
|-------|----------|-------|----------|
| YOLOv5s | 75% | ~65ms | Balanced (default) |
| YOLOv5l | 85% | ~120ms | **Recommended for scenes** |
| YOLOv8m | 88% | ~150ms | Maximum accuracy |

## Database Schema

### scenes
- `id`: Unique scene ID
- `created_at`: Unix timestamp
- `description`: "2x cars, 1x person"
- `object_counts`: Serialized counts
- `spatial_histogram`: Position grid

### scene_objects
- Object type, position (x, y)
- Color (RGB), size (width, height)
- Orientation

### object_relationships
- Distance and angle between object pairs

## Log Messages

```
[INFO] New scene identified: id=1 (2x cars, 1x person)
[INFO] Recognized return to earlier scene: id=1 (2x cars, 1x person)
```

## Performance

- **Analysis Frequency**: Every 60 seconds (when stationary objects present)
- **CPU Impact**: 5-15ms per analysis (negligible)
- **Memory**: <1MB for typical usage
- **Database Growth**: ~1-2KB per unique scene

## Common Combinations

```bash
# High accuracy scene tracking
./object_detection --model-type yolov5l --enable-scene-persistence

# Energy-efficient with scenes
./object_detection --analysis-rate-limit 0.5 --enable-scene-persistence

# Full surveillance setup
./object_detection \
  --enable-scene-persistence \
  --enable-streaming \
  --enable-burst-mode \
  --show-preview
```

## Query Database

```bash
# List all scenes
sqlite3 scenes.db "SELECT id, datetime(created_at, 'unixepoch'), description FROM scenes;"

# Scene details
sqlite3 scenes.db "SELECT * FROM scene_objects WHERE scene_id = 1;"
```

## Adjusting Match Sensitivity

**Location**: `include/scene_manager.hpp` line 93

**Default Threshold**: 0.75 (75% similarity required)

### Modify Threshold

```cpp
// More strict (85% similarity)
static constexpr float MATCH_THRESHOLD = 0.85f;

// More lenient (65% similarity)  
static constexpr float MATCH_THRESHOLD = 0.65f;
```

### Guidelines

| Threshold | Behavior | Use When |
|-----------|----------|----------|
| 0.85-0.90 | Very strict | Need precise differentiation |
| 0.75 | **Default** | Balanced, most use cases |
| 0.65-0.70 | Lenient | Camera position varies |

**After changing**: Rebuild with `make clean && make`

## Troubleshooting

| Issue | Solution |
|-------|----------|
| No scenes identified | Increase `--stationary-timeout`, lower `--min-confidence` |
| Scenes never match | Expected if scenes differ >25%, check camera stability |
| Database too large | Delete old scenes, run VACUUM |

## Code Integration Points

### SceneManager Methods

```cpp
// Initialize database
bool initialize();

// Check if analysis should run
bool shouldAnalyzeScene(const std::vector<ObjectTracker>& tracked_objects);

// Analyze and match scene
int analyzeAndMatchScene(const cv::Mat& frame, const std::vector<ObjectTracker>& tracked_objects);

// Get scene description
std::string getSceneDescription(int scene_id);
```

### Application Integration

```cpp
// In application_context.hpp
std::shared_ptr<SceneManager> scene_manager;

// In initializeComponents()
if (config.enable_scene_persistence) {
    scene_manager = std::make_shared<SceneManager>(logger, config.scene_db_path);
    scene_manager->initialize();
}

// In main processing loop
if (config.enable_scene_persistence && scene_manager) {
    if (scene_manager->shouldAnalyzeScene(tracked_objects)) {
        scene_manager->analyzeAndMatchScene(frame, tracked_objects);
    }
}
```

## Dependencies

```bash
# Ubuntu/Debian
sudo apt-get install libsqlite3-dev libopencv-dev

# macOS
brew install sqlite3 opencv
```

## Testing

```bash
# Run tests
cd build
cmake ..
make
./object_detection_tests --gtest_filter=SceneManager*
```

## Documentation

- **SCENE_PERSISTENCE_FEATURE.md** - Technical details and algorithm
- **SCENE_PERSISTENCE_EXAMPLES.md** - Usage scenarios and examples
- **README.md** - Feature overview and prerequisites

## Thread Safety

- Single-threaded design (called from main loop)
- SQLite handles internal locking
- All methods safe for sequential use

## Future Enhancements

- Configurable similarity threshold
- Scene transition detection
- Sub-part object analysis
- Machine learning embeddings
- Scene analytics/reporting
