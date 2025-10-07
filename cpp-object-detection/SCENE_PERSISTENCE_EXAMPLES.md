# Scene Persistence - Usage Examples

This document provides practical examples of using the scene persistence feature.

## Basic Usage

### Enable Scene Persistence

Start the application with scene persistence enabled:

```bash
./object_detection --enable-scene-persistence
```

This creates a `scenes.db` SQLite database in the current directory.

### Custom Database Location

Specify a custom path for the scene database:

```bash
./object_detection --enable-scene-persistence --scene-db-path /data/surveillance/scenes.db
```

### Combined with Other Features

Scene persistence works seamlessly with other features:

```bash
# High accuracy mode with scene tracking
./object_detection \
  --model-type yolov5l \
  --enable-scene-persistence \
  --scene-db-path /data/scenes.db \
  --verbose

# Energy-efficient mode with scene tracking
./object_detection \
  --analysis-rate-limit 0.5 \
  --enable-scene-persistence \
  --stationary-timeout 120

# Full-featured surveillance setup
./object_detection \
  --enable-scene-persistence \
  --enable-streaming \
  --enable-google-sheets \
  --google-sheets-id "YOUR_SHEET_ID" \
  --google-sheets-api-key "YOUR_API_KEY" \
  --enable-burst-mode \
  --show-preview
```

## Expected Output

### First Scene Detected

When the application first identifies a scene (after 1 minute of stationary objects):

```
[INFO] On 2024-01-15 14:30:00 PT, Scene persistence enabled - scenes will be stored in: scenes.db
[INFO] On 2024-01-15 14:31:00 PT, New scene identified: id=1 (2x cars, 1x person)
```

### Return to Known Scene

When the camera returns to a previously seen scene:

```
[INFO] On 2024-01-15 15:45:00 PT, Recognized return to earlier scene: id=1 (2x cars, 1x person)
```

### Different Scene

When a completely different scene is encountered:

```
[INFO] On 2024-01-15 16:15:00 PT, New scene identified: id=2 (3x people, 1x bicycle)
```

### Similar Scene (New)

When a scene is similar but not quite matching (below 75% threshold):

```
[INFO] On 2024-01-15 17:00:00 PT, New scene identified: id=3 (2x cars)
```

## Scenarios

### Scenario 1: Office Monitoring

**Setup:**
```bash
./object_detection \
  --enable-scene-persistence \
  --scene-db-path /var/surveillance/office_scenes.db \
  --stationary-timeout 300 \
  --analysis-rate-limit 0.5
```

**Timeline:**
- 09:00: Office empty → No scene recorded yet (no stationary objects)
- 09:30: Person arrives, sits at desk → After 1 min, scene ID 1 created: "1x person"
- 12:00: Person leaves for lunch → Scene ends (no stationary objects)
- 13:00: Person returns, sits at desk → After 1 min, recognizes scene ID 1
- 14:00: Second person joins → After 1 min, new scene ID 2: "2x people"
- 17:00: Everyone leaves → Scene ends

**Database Contents:**
- Scene 1: "1x person" (matched multiple times)
- Scene 2: "2x people" (matched once)

### Scenario 2: Parking Lot Surveillance

**Setup:**
```bash
./object_detection \
  --enable-scene-persistence \
  --model-type yolov5l \
  --enable-streaming \
  --stationary-timeout 120
```

**Timeline:**
- 08:00: 2 cars parked → Scene ID 1: "2x cars"
- 09:00: 1 more car arrives → Scene ID 2: "3x cars"
- 10:00: 1 car leaves → Scene ID 3: "2x cars" (different positions than ID 1)
- 11:00: Another car arrives → Matches Scene ID 2: "3x cars"

**Notes:**
- Scene 1 and Scene 3 both have "2x cars" but weren't matched
- This is because the cars are in different positions (spatial distribution differs)
- Scene 2 was successfully matched when the same parking configuration returned

### Scenario 3: Home Security

**Setup:**
```bash
./object_detection \
  --enable-scene-persistence \
  --enable-burst-mode \
  --enable-brightness-filter \
  --show-preview
```

**Common Scenes:**
- Scene 1: "1x person, 1x cat" (owner at computer, cat nearby)
- Scene 2: "1x cat" (cat alone in room)
- Scene 3: "2x people" (owner + guest)
- Scene 4: Empty room (no stationary objects, not recorded)

**Benefits:**
- Quickly identify unusual scenes (e.g., unexpected person)
- Track daily patterns (how often each scene occurs)
- Distinguish between familiar and unfamiliar configurations

## Querying the Database

You can query the SQLite database directly to analyze scene patterns:

### View All Scenes

```bash
sqlite3 scenes.db "SELECT id, datetime(created_at, 'unixepoch') as created, description FROM scenes;"
```

**Output:**
```
1|2024-01-15 14:31:00|2x cars, 1x person
2|2024-01-15 16:15:00|3x people, 1x bicycle
3|2024-01-15 17:00:00|2x cars
```

### Count Scene Occurrences

Track how many times the application has returned to each scene by monitoring logs or implementing custom analytics.

### Scene Objects Detail

```bash
sqlite3 scenes.db "SELECT object_type, position_x, position_y FROM scene_objects WHERE scene_id = 1;"
```

**Output:**
```
car|150.5|200.3
car|450.2|210.8
person|320.0|180.5
```

### Spatial Relationships

```bash
sqlite3 scenes.db "SELECT distance, angle FROM object_relationships WHERE scene_id = 1;"
```

**Output:**
```
300.5|0.034
250.3|1.234
...
```

## Troubleshooting

### Scene Never Identified

**Problem:** Application runs but never logs "New scene identified"

**Possible Causes:**
1. No stationary objects (all objects moving continuously)
2. Stationary timeout too short (objects don't stay stationary long enough)
3. Objects not detected confidently enough

**Solutions:**
```bash
# Increase stationary timeout
./object_detection --enable-scene-persistence --stationary-timeout 300

# Lower confidence threshold
./object_detection --enable-scene-persistence --min-confidence 0.3

# Enable verbose logging to see what's happening
./object_detection --enable-scene-persistence --verbose
```

### Scenes Always New (Never Matching)

**Problem:** Every scene is marked as new, even when returning to same location

**Possible Causes:**
1. Camera position changing slightly each time
2. Lighting changes affecting color extraction
3. Object detection inconsistencies

**Notes:**
- This is expected behavior if the scene actually differs by more than 25%
- Fuzzy matching already accounts for slight variations
- Consider if scenes are actually different (different object counts, positions)

**Adjusting Match Sensitivity:**

If scenes should match but don't, you can lower the similarity threshold in `include/scene_manager.hpp`:

```cpp
// Current default (line 93)
static constexpr float MATCH_THRESHOLD = 0.75f;  // 75% similarity required

// More lenient matching (allows more variation)
static constexpr float MATCH_THRESHOLD = 0.65f;  // 65% similarity required
```

After changing, rebuild: `cd build && make clean && make`

See "Adjusting Scene Matching Sensitivity" in SCENE_PERSISTENCE_FEATURE.md for detailed guidance.

### Database Growing Too Large

**Problem:** Database file getting too big

**Solutions:**
```bash
# Check database size
ls -lh scenes.db

# Count scenes
sqlite3 scenes.db "SELECT COUNT(*) FROM scenes;"

# Delete old scenes (older than 30 days)
sqlite3 scenes.db "DELETE FROM scenes WHERE created_at < strftime('%s', 'now', '-30 days');"
sqlite3 scenes.db "VACUUM;"
```

## Performance Impact

- **CPU**: Scene analysis adds ~5-15ms every 60 seconds (negligible)
- **Memory**: <1MB for typical usage (scenes loaded on-demand)
- **Disk I/O**: One transaction per scene (infrequent)
- **Overall**: No measurable impact on detection performance

## Best Practices

1. **Choose Appropriate Stationary Timeout**
   - Too short: Objects not stationary long enough for analysis
   - Too long: Scenes take too long to be identified
   - Recommended: 120-300 seconds

2. **Use High-Accuracy Model for Better Matching**
   - YOLOv5l or YOLOv8m recommended
   - Better object localization = more accurate scene fingerprints

3. **Stable Camera Position**
   - Scene persistence works best with fixed camera angle
   - Small movements okay, but major repositioning creates new scenes

4. **Regular Database Maintenance**
   - Archive old scenes periodically if needed
   - Keep database size manageable with VACUUM

5. **Monitor Logs**
   - Use `--verbose` during initial setup
   - Check that scenes are being identified as expected
   - Verify matching behavior with known scenes
