# Smart Photo Storage - Detailed Behavior Examples

## Overview
This document provides detailed examples of how the smart photo storage behaves in various scenarios.

## Core Logic

### When Photos Are Saved

**IMMEDIATE SAVE (bypasses 10s limit):**
- ✅ New object type enters frame
- ✅ Additional instance of existing type enters frame (count increases)

**RATE-LIMITED SAVE (10s interval):**
- ⏱️ Same objects remain (no changes)
- ⏱️ Object count decreases (car leaves)
- ⏱️ Object count stays same

## Detailed Scenarios

### Scenario 1: Single Stationary Car
```
Time  | Objects      | Count Change | Photo Saved? | Reason
------|-------------|--------------|--------------|------------------
0s    | 1 car       | 0→1 car      | ✅ YES       | New type (first detection)
5s    | 1 car       | 1→1 car      | ❌ NO        | Same count, < 10s
10s   | 1 car       | 1→1 car      | ✅ YES       | 10s interval
15s   | 1 car       | 1→1 car      | ❌ NO        | Same count, < 10s
20s   | 1 car       | 1→1 car      | ✅ YES       | 10s interval
```

### Scenario 2: Second Car Enters
```
Time  | Objects      | Count Change | Photo Saved? | Reason
------|-------------|--------------|--------------|------------------
0s    | 1 car       | 0→1 car      | ✅ YES       | New type
5s    | 1 car       | 1→1 car      | ❌ NO        | Same count, < 10s
8s    | 2 cars      | 1→2 cars     | ✅ YES       | NEW INSTANCE! (immediate)
12s   | 2 cars      | 2→2 cars     | ❌ NO        | Same count, < 10s
18s   | 2 cars      | 2→2 cars     | ✅ YES       | 10s interval
```

### Scenario 3: Person Enters (New Type)
```
Time  | Objects          | Count Change      | Photo Saved? | Reason
------|-----------------|-------------------|--------------|------------------
0s    | 1 car           | 0→1 car           | ✅ YES       | New type
5s    | 1 car           | 1→1 car           | ❌ NO        | Same count, < 10s
7s    | 1 car, 1 person | +1 person         | ✅ YES       | NEW TYPE! (immediate)
12s   | 1 car, 1 person | no change         | ❌ NO        | Same objects, < 10s
17s   | 1 car, 1 person | no change         | ✅ YES       | 10s interval
```

### Scenario 4: Car Leaves (Count Decreases)
```
Time  | Objects      | Count Change | Photo Saved? | Reason
------|-------------|--------------|--------------|------------------
0s    | 2 cars      | 0→2 cars     | ✅ YES       | New type
5s    | 2 cars      | 2→2 cars     | ❌ NO        | Same count, < 10s
8s    | 1 car       | 2→1 car      | ❌ NO        | Count decreased, < 10s
18s   | 1 car       | 1→1 car      | ✅ YES       | 10s interval
```

**Important:** When an object leaves (count decreases), we DON'T save immediately. We only save at the 10s interval.

### Scenario 5: Object Type Disappears Then Returns
```
Time  | Objects      | Count Change | Photo Saved? | Reason
------|-------------|--------------|--------------|------------------
0s    | 1 car       | 0→1 car      | ✅ YES       | New type
10s   | 1 car       | 1→1 car      | ✅ YES       | 10s interval
15s   | -           | 1→0 cars     | ❌ NO        | No objects detected
20s   | 1 car       | 0→1 car      | ✅ YES       | NEW TYPE! (returned after being gone)
```

### Scenario 6: Complex Multi-Object Scene
```
Time  | Objects                    | Changes           | Photo Saved? | Reason
------|---------------------------|-------------------|--------------|------------------
0s    | 1 car                     | +1 car            | ✅ YES       | New type
5s    | 1 car, 1 person           | +1 person         | ✅ YES       | NEW TYPE! (immediate)
10s   | 1 car, 1 person           | no change         | ❌ NO        | Same objects, < 10s
12s   | 2 cars, 1 person          | +1 car            | ✅ YES       | NEW INSTANCE! (immediate)
17s   | 2 cars, 1 person          | no change         | ❌ NO        | Same objects, < 10s
19s   | 2 cars, 2 persons         | +1 person         | ✅ YES       | NEW INSTANCE! (immediate)
24s   | 2 cars, 1 person          | -1 person         | ❌ NO        | Count decreased, < 10s
29s   | 2 cars, 1 person          | no change         | ✅ YES       | 10s interval
```

### Scenario 7: Rapid Changes
```
Time  | Objects          | Changes           | Photo Saved? | Reason
------|-----------------|-------------------|--------------|------------------
0s    | 1 person        | +1 person         | ✅ YES       | New type
1s    | 1 person, 1 cat | +1 cat            | ✅ YES       | NEW TYPE! (immediate)
2s    | 2 persons, 1 cat| +1 person         | ✅ YES       | NEW INSTANCE! (immediate)
3s    | 2 persons, 2 cats| +1 cat           | ✅ YES       | NEW INSTANCE! (immediate)
4s    | 3 persons, 2 cats| +1 person        | ✅ YES       | NEW INSTANCE! (immediate)
```

**Note:** Rapid changes all trigger immediate saves because each is a meaningful scene change.

## Log Output Examples

### Example 1: New Type Detected
```
[INFO] detected car at coordinates: (400, 300) with confidence 85%
[INFO] New object type detected: car
[INFO] Saving photo immediately due to new objects/types detected
[INFO] Saved detection photo: detections/2025-10-04 140020 car detected.jpg
```

### Example 2: New Instance Detected
```
[INFO] detected car at coordinates: (400, 300) with confidence 85%
[INFO] detected car at coordinates: (600, 320) with confidence 82%
[INFO] New instance of car detected (count: 1 -> 2)
[INFO] Saving photo immediately due to new objects/types detected
[INFO] Saved detection photo: detections/2025-10-04 140025 car detected.jpg
```

### Example 3: Stationary Object (10s Interval)
```
[INFO] detected car at coordinates: (400, 305) with confidence 86%
[INFO] Saved detection photo: detections/2025-10-04 140035 car detected.jpg
```

### Example 4: Tracker Detects New Entry
```
[INFO] detected person at coordinates: (320, 240) with confidence 92%
[INFO] Newly entered person detected by tracker
[INFO] Saving photo immediately due to new objects/types detected
[INFO] Saved detection photo: detections/2025-10-04 140040 person detected.jpg
```

## Benefits

### Storage Efficiency
**Before:** Photo every 10s with stationary car = 360 photos/hour
**After:** 
- 1st photo immediately (new type)
- Then 1 photo every 10s = 6 photos/minute
- **Same as before for stationary objects**

### Responsiveness
**Before:** Up to 10s delay when person enters frame with existing car
**After:** 
- Immediate photo when person enters
- Immediate photo when 2nd car enters
- **No delay for important changes**

### Example Storage Reduction (Active Scene)
```
Scenario: Car present, person enters at 5s, 2nd person at 8s

Before:
0s:  Photo (car)
10s: Photo (car + person)        ← Delayed response to person
20s: Photo (car + 2 persons)     ← Delayed response to 2nd person
Total: 3 photos, max 10s delay

After:
0s: Photo (car)                  ← New type
5s: Photo (car + person)         ← NEW TYPE (immediate!)
8s: Photo (car + 2 persons)      ← NEW INSTANCE (immediate!)
Total: 3 photos, 0s delay for changes
```

## Edge Cases Handled

1. ✅ **Empty initial state**: First detection always triggers save
2. ✅ **Object tracking**: Uses both count comparison AND tracker state
3. ✅ **Decreasing count**: Doesn't trigger immediate save
4. ✅ **Thread safety**: All operations protected by mutex
5. ✅ **Multiple types**: Handles mixed object types correctly
6. ✅ **Rapid changes**: Each meaningful change triggers save
