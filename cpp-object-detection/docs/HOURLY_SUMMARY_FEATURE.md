# Hourly Object Detection Summary Feature

## Overview

The hourly summary feature provides periodic summaries of object detections printed to stdout. This feature helps monitor detection activity over time and identify patterns in object movements.

## Configuration

The summary interval is configurable via the `summary_interval_minutes` configuration option:

```cpp
// In config_manager.hpp
int summary_interval_minutes = 60;  // Default: 1 hour
```

To change the interval (e.g., to 30 minutes), modify the configuration or add command-line support for this parameter.

## How It Works

### Detection Recording

Every time an object is detected, the system records:
- Object type (e.g., "person", "car", "cat")
- Timestamp of detection
- Whether the object is stationary or dynamic

Stationary objects are identified by tracking movement between frames:
- If an object moves less than 5 pixels between frames, it's marked as stationary
- If an object moves more than 5 pixels, it's marked as dynamic

### Summary Generation

At the configured interval (default: every hour), the system prints a summary to stdout containing:

1. **Count Summary**: Total number of each object type detected
   - Proper pluralization (e.g., "people" not "persons")
   - Example: "1x car, 3x people, 2x cats were detected"

2. **Timeline**: Chronological list of detections with time-based grouping
   - Stationary objects are fused into time ranges (e.g., "from 0:00-0:10 a car was detected")
   - Dynamic objects close in time are grouped (e.g., "at 0:50, two people were detected")
   - Single detections shown with specific time (e.g., "at 0:10, a person was detected")

## Example Output

### Periodic Summary (Hourly)

```
========================================
Detection Summary: 00:00-01:00
========================================
1x car, 3x people, 2x animals were detected.

Timeline:
from 00:00-00:10 a car was detected
at 00:10, a person was detected
at 00:30, a cat was detected
at 00:31, a dog was detected
at 00:50, two people were detected
from 00:50-01:00 a car was detected
========================================
```

### Final Summary (Program Exit)

When the program exits, a final summary covering the entire program runtime is printed:

```
========================================
Final Detection Summary: 08:00-18:45
Program Runtime: 10h 45m 30s
========================================
12x cars, 45x people, 8x cats, 5x dogs were detected.

Timeline:
from 08:00-08:15 a car was detected
at 08:15, a person was detected
at 09:30, a cat was detected
...
from 18:30-18:45 a car was detected
========================================
```

This final summary uses the same summarization logic but spans the entire program execution time.

## API

### Recording Detections

```cpp
// Record a detection for the summary
logger->recordDetection(object_type, is_stationary);

// Examples:
logger->recordDetection("person", false);  // Dynamic person
logger->recordDetection("car", true);      // Stationary car
```

### Checking and Printing Summary

```cpp
// Check if interval has elapsed and print summary if needed
logger->checkAndPrintSummary(interval_minutes);

// Or manually trigger periodic summary
logger->printHourlySummary();

// Print final summary covering entire program runtime
logger->printFinalSummary();
```

### Integration in Main Loop

The feature is integrated into the main processing loop in `application.cpp`:

```cpp
// Periodic heartbeat logging
auto now = std::chrono::steady_clock::now();
if (now - ctx.last_heartbeat >= ctx.heartbeat_interval) {
    ctx.logger->logHeartbeat();
    ctx.perf_monitor->logPerformanceReport();
    ctx.last_heartbeat = now;
}

// Check and print hourly summary
ctx.logger->checkAndPrintSummary(ctx.config.summary_interval_minutes);
```

### Integration in Shutdown

The final summary is automatically printed during graceful shutdown in `performGracefulShutdown()`:

```cpp
// Print final summary covering entire program runtime
ctx.logger->printFinalSummary();

ctx.logger->info("Object Detection Application stopped");
```
auto now = std::chrono::steady_clock::now();
if (now - ctx.last_heartbeat >= ctx.heartbeat_interval) {
    ctx.logger->logHeartbeat();
    ctx.perf_monitor->logPerformanceReport();
    ctx.last_heartbeat = now;
}

// Check and print hourly summary
ctx.logger->checkAndPrintSummary(ctx.config.summary_interval_minutes);
```

## Timeline Fusion Logic

### Stationary Objects

Consecutive stationary detections of the same object type are fused into a single time range:

```
Detection: car (stationary) at 00:00
Detection: car (stationary) at 00:01
Detection: car (stationary) at 00:02
→ Output: "from 00:00-00:02 a car was detected"
```

### Dynamic Objects

Dynamic objects detected within 10 seconds of each other are grouped:

```
Detection: person at 00:50:00
Detection: person at 00:50:05
→ Output: "at 00:50, two people were detected"
```

### Mixed Scenarios

The system handles transitions between stationary and dynamic objects:

```
Detection: car (stationary) at 00:00
Detection: car (stationary) at 00:05
Detection: person (dynamic) at 00:10
Detection: car (stationary) at 00:50
Detection: car (stationary) at 00:55

→ Output:
from 00:00-00:05 a car was detected
at 00:10, a person was detected
from 00:50-00:55 a car was detected
```

## Thread Safety

All summary-related methods are thread-safe using mutex locks:
- `recordDetection()` uses `summary_mutex_` to protect event recording
- `printHourlySummary()` uses `summary_mutex_` to protect event access
- `checkAndPrintSummary()` safely checks elapsed time and triggers summary

## Testing

Tests are provided in `tests/test_hourly_summary.cpp`:

- `RecordDetections`: Verifies basic detection recording
- `StationaryObjectFusion`: Tests fusion of stationary object periods
- `CheckAndPrintSummaryByTime`: Validates interval-based triggering
- `MultipleObjectTypes`: Tests handling of various object types
- `EmptySummary`: Ensures graceful handling of no detections
- `ConsecutiveDynamicObjects`: Tests grouping of similar detections
- `FinalSummary`: Tests final summary covering all events with periodic summaries in between
- `FinalSummaryEmpty`: Tests final summary with no detections

Run tests with:
```bash
cd build
ctest -R HourlySummary
```

## Summary Types

### Periodic Summary (Hourly)

- Triggered at configurable intervals (default: 1 hour)
- Shows detections from the current period only
- Events are cleared after each periodic summary
- Continues throughout program runtime

### Final Summary (Program Exit)

- Triggered once when program exits
- Shows all detections from entire program runtime
- Includes program runtime duration
- Uses same timeline fusion logic
- Automatically printed during graceful shutdown

## Future Enhancements

Potential improvements to consider:

1. **Configurable Fusion Threshold**: Make the 10-second grouping window configurable
2. **JSON Output**: Option to output summaries in JSON format for programmatic processing
3. **Historical Summaries**: Save summaries to file for later analysis
4. **Visualization**: Generate charts or graphs of detection patterns
5. **Alerts**: Trigger alerts when unusual detection patterns occur
