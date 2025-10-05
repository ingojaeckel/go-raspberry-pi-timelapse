# Burst Mode Feature

## Overview

Burst mode is an intelligent frame rate throttling feature that automatically adjusts the analysis rate based on scene activity. When enabled, the system temporarily maxes out the frame analysis rate when new objects enter the scene, then returns to normal rate limiting when the scene stabilizes.

## Problem Statement

The standard rate limiting feature (`--analysis-rate-limit`) effectively reduces CPU usage and energy consumption by adding sleep intervals between frame analyses. However, this can miss rapid action when interesting events occur (e.g., a person quickly walking through the scene). 

Burst mode solves this by:
- **Energy efficiency during quiet periods**: Normal rate limiting (e.g., 1 FPS) when scene is stable
- **Maximum temporal resolution during events**: Removing sleep delays when new objects appear
- **Automatic adaptation**: No manual intervention needed

## How It Works

### State Machine

Burst mode operates as a state machine with two states:

1. **INACTIVE**: Normal rate limiting applies, sleep intervals between frames
2. **ACTIVE**: Sleep intervals removed, frames analyzed as fast as possible (minimal 1ms delay only)

### State Transitions

**Activation (INACTIVE → ACTIVE)**

Burst mode activates when:
- A new object TYPE enters the scene that wasn't present in the previous frame
- OR a new INSTANCE of an object is detected (marked with `is_new` flag by object tracker)

**Deactivation (ACTIVE → INACTIVE)**  

Burst mode deactivates when:
- All tracked objects become stationary (based on movement detection)
- OR no objects are present in the scene
- OR only previously-known objects remain in the scene

### Example Scenario

```
// → Burst mode OFF at startup
...
// Frame N: {car (stationary)}
// All objects stationary and known
// → Burst mode: INACTIVE

// Frame N+1: {person, car (stationary)}
// "person" is a new object type!
// → Burst mode ACTIVATED
// Log: "Burst mode ACTIVATED - new object type detected"

// Frame N+2: {person, car (stationary)}
// Same object types as previous frame, person still moving
// → Burst mode: ACTIVE (continues)

// Frame N+3: {person (stationary), car (stationary)}
// All objects now stationary
// → Burst mode DEACTIVATED
// Log: "Burst mode DEACTIVATED - all objects stationary"

// Frame N+4: {car (stationary)}
// Person left scene
// → Burst mode: INACTIVE
```

## Implementation Details

### Configuration

**Command-line flag:**
```bash
--enable-burst-mode
```

**Default state:** Disabled (opt-in feature)

**Config struct field:**
```cpp
bool enable_burst_mode = false;  // Enable burst mode
```

### Application Context State

The application tracks burst mode state in the `ApplicationContext`:

```cpp
bool burst_mode_active = false;  // Current burst mode state
std::set<std::string> previous_object_types;  // Object types from previous frame
```

### Logic Flow

1. **After each frame is processed**, check if burst mode is enabled
2. **Get current object types** from the object detector's tracked objects
3. **Compare with previous frame's object types** to detect new entries
4. **Check object motion state** (new vs. stationary)
5. **Update burst mode state** and log transitions
6. **Apply appropriate rate limiting**:
   - Burst mode ACTIVE: 1ms minimal delay only
   - Burst mode INACTIVE: Normal rate limiting calculation

### Code Location

- **Configuration**: `include/config_manager.hpp`, `src/config_manager.cpp`
- **State tracking**: `include/application_context.hpp`
- **Logic implementation**: `src/application.cpp` in `runMainProcessingLoop()`
- **Unit tests**: `tests/test_config_manager.cpp`

## Usage Examples

### Basic Usage

Enable burst mode with default 1 FPS rate limiting:
```bash
./object_detection --enable-burst-mode
```

### Recommended Configuration

For energy-efficient monitoring with event capture:
```bash
./object_detection --enable-burst-mode --analysis-rate-limit 1 --min-confidence 0.6
```

### With Other Features

Combine with viewfinder and network streaming:
```bash
./object_detection \
  --enable-burst-mode \
  --analysis-rate-limit 1 \
  --show-preview \
  --enable-streaming \
  --streaming-port 8080
```

## Performance Impact

### Energy Consumption

- **Idle scene** (burst OFF): 1 FPS = ~99% idle time, minimal CPU usage
- **Active scene** (burst ON): Maximum FPS (~5-30 FPS depending on hardware)
- **Transition overhead**: Negligible (simple boolean checks)

### Detection Quality

- **No loss of detection accuracy**: Same confidence thresholds apply
- **Improved temporal resolution**: More frames during action = better tracking
- **Reduced missed events**: Fast-moving objects less likely to be missed

## Logging

Burst mode state transitions are logged at INFO level:

```
[INFO] Burst mode: ENABLED - will max out FPS when new objects enter the scene
[INFO] Burst mode ACTIVATED - new object type detected
[DEBUG] Burst mode active: skipping normal rate limiting (minimal 1ms delay)
[INFO] Burst mode DEACTIVATED - all objects stationary
[INFO] Burst mode DEACTIVATED - no objects detected
```

## Integration with Other Features

### Object Tracking

Burst mode relies on the object tracking system to:
- Detect new object instances (`is_new` flag)
- Track object motion state (`is_stationary` flag)
- Maintain object type information

### Stationary Object Detection

Works seamlessly with stationary timeout:
- Objects become stationary after minimal movement
- Burst mode deactivates when all objects stationary
- Photo storage may continue per stationary timeout rules

### Rate Limiting

Burst mode modifies the existing rate limiting behavior:
- When INACTIVE: Normal `--analysis-rate-limit` calculation applies
- When ACTIVE: Skip calculated sleep, use 1ms minimal delay only

## Testing

### Unit Tests

Location: `tests/test_config_manager.cpp`

Tests verify:
- Default state is disabled
- CLI flag properly enables burst mode
- Configuration validation passes

Run tests:
```bash
cd build
./tests/object_detection_tests --gtest_filter="ConfigManagerTest.BurstMode*"
```

### Manual Testing

1. Run with burst mode enabled and verbose logging:
```bash
./object_detection --enable-burst-mode --verbose
```

2. Observe logs for state transitions:
   - Watch for "ACTIVATED" when you enter the camera view
   - Watch for "DEACTIVATED" when you stand still or leave

3. Monitor FPS in viewfinder or streaming interface to see rate changes

## Future Enhancements

Potential improvements:
- Configurable burst duration timeout
- Burst intensity levels (partial FPS increase vs. full throttle)
- Machine learning-based scene change detection
- Integration with motion-based photo triggers

## Conclusion

Burst mode provides an intelligent balance between energy efficiency and event capture quality. It's particularly useful for:
- Security monitoring (capture intruders in detail)
- Wildlife observation (track animal movement precisely)
- Long-term deployments (reduce power while maintaining coverage)
- Raspberry Pi and embedded deployments (optimize limited resources)
