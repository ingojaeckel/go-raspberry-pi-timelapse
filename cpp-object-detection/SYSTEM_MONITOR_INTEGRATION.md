# System Monitor Integration in Viewfinder & Network Stream

## Overview

This document describes the integration of SystemMonitor metrics (disk usage percentage and CPU temperature) into the debug information overlay displayed in both the viewfinder and network stream.

## Motivation

While the debug overlay previously showed performance metrics, detection statistics, and camera information, it lacked critical system health metrics. Users monitoring the application remotely or locally needed visibility into:
- **Disk usage**: To ensure sufficient storage for captured images
- **CPU temperature**: To detect potential thermal throttling or cooling issues

This integration provides real-time system health monitoring directly in the video feed.

## Implementation Details

### Changes Made

#### 1. Header Files Updated

**viewfinder_window.hpp:**
- Added `disk_usage_percent` and `cpu_temp_celsius` parameters to `showFrameWithStats()`
- Added `disk_usage_percent` and `cpu_temp_celsius` parameters to `drawDebugInfo()`
- Both parameters have default values of `-1.0` (indicates unavailable)

**network_streamer.hpp:**
- Added `disk_usage_percent` and `cpu_temp_celsius` parameters to `updateFrameWithStats()`
- Added `disk_usage_percent` and `cpu_temp_celsius` parameters to `drawDebugInfo()`
- Both parameters have default values of `-1.0` (indicates unavailable)

#### 2. Implementation Files Updated

**viewfinder_window.cpp:**
- Updated `showFrameWithStats()` to accept and forward system monitor metrics
- Updated `drawDebugInfo()` to display disk usage and CPU temperature
- Metrics are only displayed if their values are >= 0.0 (available)
- Format: `Disk: XX.X%` and `CPU: XX.X°C`

**network_streamer.cpp:**
- Updated `updateFrameWithStats()` to accept and forward system monitor metrics
- Updated `drawDebugInfo()` to display disk usage and CPU temperature
- Identical display logic to viewfinder for consistency

**application.cpp:**
- Modified viewfinder update section to retrieve system monitor metrics
- Modified network streamer update section to retrieve system monitor metrics
- Uses null-safe checks: only queries if `ctx.system_monitor` exists
- Passes metrics to both `showFrameWithStats()` and `updateFrameWithStats()`

#### 3. Test Files Updated

**test_viewfinder_window.cpp:**
- Updated `ShowFrameWithStatsWithoutInitialization` test
- Added sample disk usage (75.5%) and CPU temp (62.3°C) values

**test_network_streamer.cpp:**
- Updated `UpdateFrameWithStats` test
- Added sample disk usage (75.5%) and CPU temp (62.3°C) values

### Display Format

The system monitor metrics appear after the GPU and Burst mode status lines:

```
FPS: 15.0
Avg proc: 45 ms
Objects: 10
Images: 3
Uptime: 01:23:45
Camera 0: 640x480
Detection: 320x240
GPU: ON
Burst: OFF
Disk: 75.5%        <-- NEW
CPU: 62.3°C        <-- NEW
--- Top Objects ---
person: 5
cat: 3
```

### Conditional Display

- **Disk usage**: Only shown if `disk_usage_percent >= 0.0`
- **CPU temperature**: Only shown if `cpu_temp_celsius >= 0.0`

This allows the system to gracefully handle:
- Missing `/sys/class/thermal/thermal_zone0/temp` (CPU temp unavailable)
- Permission issues accessing disk statistics
- Platforms where these metrics aren't available

## Architecture Integration

The system monitor is already initialized in `application.cpp` and performs periodic checks. This integration simply exposes those metrics in the UI:

```
SystemMonitor (existing)
    ↓
    getDiskUsagePercent()
    getCPUTemperature()
    ↓
Application (gathers metrics)
    ↓
ViewfinderWindow / NetworkStreamer (displays metrics)
```

## Benefits

1. **Real-time monitoring**: Users can see system health without checking logs
2. **Early warning**: Disk space issues visible before running out
3. **Thermal awareness**: CPU temperature helps identify cooling problems
4. **Remote visibility**: Network stream shows metrics for headless operation
5. **Consistency**: Same metrics in both viewfinder and network stream

## Testing

### Unit Tests
- ✅ `ViewfinderWindowTest.ShowFrameWithStatsWithoutInitialization` - Updated with sample metrics
- ✅ `NetworkStreamerTest.UpdateFrameWithStats` - Updated with sample metrics

### Manual Testing Required
1. Run with viewfinder: `./object_detection --show-preview`
   - Verify disk usage and CPU temp appear in overlay
   - Check values match system reality (`df -h` and temperature sensors)
2. Run with network stream: `./object_detection --enable-streaming`
   - Access stream at `http://IP:8080/stream`
   - Verify metrics appear in network stream
3. Run on system without thermal sensors
   - Verify CPU temp line is omitted gracefully
4. Run with full disk (or mock high usage)
   - Verify disk usage displays correctly

## Files Modified

### Headers (include/)
- `viewfinder_window.hpp` - Added system monitor parameters
- `network_streamer.hpp` - Added system monitor parameters

### Source (src/)
- `viewfinder_window.cpp` - Display logic for system metrics
- `network_streamer.cpp` - Display logic for system metrics  
- `application.cpp` - Retrieval and passing of system metrics

### Tests (tests/)
- `test_viewfinder_window.cpp` - Updated test with new parameters
- `test_network_streamer.cpp` - Updated test with new parameters

### Documentation
- `NETWORK_STREAM_STATS_INTEGRATION.md` - Updated statistics list
- `FEATURE_README.md` - Added SystemMonitor to architecture
- `SYSTEM_MONITOR_INTEGRATION.md` - This document

## Future Enhancements

Potential additions to system monitoring:
- Memory usage percentage
- Available disk space in MB/GB (in addition to percentage)
- Network bandwidth usage
- GPU temperature (if available)
- Process CPU usage percentage

## Compatibility

- **Backward compatible**: All new parameters have default values
- **Graceful degradation**: Missing metrics simply don't appear
- **Cross-platform**: Works on Linux systems with thermal sensors
- **No breaking changes**: Existing code continues to work
