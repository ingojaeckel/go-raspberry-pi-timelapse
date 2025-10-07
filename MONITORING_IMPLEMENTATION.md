# Debug Information Tracking Implementation Summary

## Overview

This implementation adds comprehensive system monitoring and debug logging capabilities to address the issue requirements for long-term, remote, solar-powered deployments.

## Issue Requirements Addressed

### ✅ 1. Review when the Pi turned on/off/why
- **Implementation**: Automatic logging of STARTUP and SHUTDOWN events
- **Location**: `system-debug.log`
- **Format**: `[STARTUP] 2025-10-06T20:29:28Z - System started`
- **Details**: Each startup logs a timestamp, and each shutdown logs total runtime

### ✅ 2. Review temperature changes over time
- **Implementation**: Periodic logging of CPU and GPU temperatures
- **Frequency**: Every 5 minutes (configurable)
- **Sources**:
  - CPU: `/sys/class/thermal/thermal_zone0/temp`
  - GPU: `/opt/vc/bin/vcgencmd measure_temp`
- **Format**: JSON in stats entries

### ✅ 3. Review disk usage over time
- **Implementation**: Periodic logging of free disk space
- **Frequency**: Every 5 minutes (configurable)
- **Source**: `df -h` command
- **Format**: JSON in stats entries

### ✅ 4. Review state of system clock over time
- **Implementation**: Periodic logging of system time with timezone
- **Frequency**: Every 5 minutes (configurable)
- **Source**: `date +%Y-%m-%d %H:%M:%S %Z`
- **Format**: JSON in stats entries
- **Use case**: Detect RTC drift or NTP sync issues

### ✅ 5. Review max runtime in a day
- **Implementation**: Daily runtime tracking with automatic reset at midnight
- **Feature**: Logs maximum runtime achieved each day
- **Format**: `[DAILY_RUNTIME] 2025-10-07T00:00:05Z - Max runtime for previous day: 11789 seconds`
- **Use case**: Assess battery capacity and identify insufficient power supply

## Implementation Details

### New Files Created

1. **`monitoring/monitor.go`** (150 lines)
   - Core monitoring logic
   - System stats collection
   - Event logging
   - Daily runtime tracking

2. **`monitoring/monitor_test.go`** (100 lines)
   - Comprehensive test coverage
   - Tests for initialization, runtime tracking, periodic checks, and stats collection

3. **`monitoring/README.md`** (200+ lines)
   - Detailed usage documentation
   - Troubleshooting guide
   - Example queries for log analysis

### Modified Files

1. **`main.go`**
   - Added monitoring initialization
   - Graceful shutdown handling with signal capture
   - Integration with timelapse

2. **`timelapse/model.go`**
   - Added Monitor interface
   - Support for optional monitoring integration

3. **`timelapse/timelapse.go`**
   - Integrated periodic monitoring checks in capture loop
   - Both offset and non-offset modes

4. **`README.md`**
   - Added "System Monitoring for Remote Deployments" section
   - Link to detailed monitoring documentation

5. **`.github/workflows/go.yml`** and **`.github/workflows/go-arm.yml`**
   - Added monitoring package to workflow paths

## Architecture

### Monitoring Flow

```
Application Start
    ↓
Initialize Monitor
    ↓
Log STARTUP event
    ↓
Capture Loop
    ↓
PeriodicCheck() called each iteration
    ↓
Check if 5 minutes elapsed
    ↓
If yes: Collect & log system stats
    ↓
Check if day changed
    ↓
If yes: Log max daily runtime & reset
    ↓
Application Shutdown
    ↓
Log SHUTDOWN event with total runtime
```

### Log File Structure

The `system-debug.log` file contains two types of entries:

1. **Event Logs** (human-readable)
   ```
   [STARTUP] 2025-10-06T20:29:28Z - System started
   [SHUTDOWN] 2025-10-06T20:29:30Z - System shutdown after 2 seconds runtime
   [DAILY_RUNTIME] 2025-10-07T00:00:05Z - Max runtime for previous day: 11789 seconds
   ```

2. **System Statistics** (JSON format)
   ```json
   [STATS] {"timestamp":"2025-10-06T20:32:03Z","uptime":"20:32:03 up 1:05, 1 user, load average: 0.15, 0.10, 0.05","cpu_temperature":"45123","gpu_temperature":"temp=45.0'C","free_disk_space":"Filesystem      Size  Used Avail Use% Mounted on\n/dev/root        15G  4.2G   10G  30% /","system_clock":"2025-10-06 20:32:03 UTC","runtime_seconds":300}
   ```

## Configuration

Current defaults (can be modified in `monitoring/monitor.go`):

```go
const (
    DebugLogFile = "system-debug.log"
    MonitoringInterval = 5 * time.Minute
)
```

## Testing

All tests pass successfully:

```bash
$ go test ./monitoring/... -v
=== RUN   TestNew
--- PASS: TestNew (0.00s)
=== RUN   TestGetRuntimeSeconds
--- PASS: TestGetRuntimeSeconds (0.10s)
=== RUN   TestPeriodicCheckIntervalRespected
--- PASS: TestPeriodicCheckIntervalRespected (0.00s)
=== RUN   TestDailyRuntimeTracking
--- PASS: TestDailyRuntimeTracking (0.00s)
=== RUN   TestCollectStats
--- PASS: TestCollectStats (0.04s)
PASS
```

## Security

CodeQL analysis completed with **0 vulnerabilities** found.

## Usage Examples

### Analyzing Debug Logs

#### Check for unexpected restarts
```bash
grep '[STARTUP]' system-debug.log
```

#### Extract temperature trends
```bash
grep '\[STATS\]' system-debug.log | sed 's/\[STATS\] //' | jq -r '[.timestamp, .cpu_temperature] | @csv'
```

#### Review battery capacity
```bash
grep '\[DAILY_RUNTIME\]' system-debug.log
```

#### Parse all statistics
```bash
grep '\[STATS\]' system-debug.log | sed 's/\[STATS\] //' | jq '.'
```

## Benefits

1. **Remote Troubleshooting**: Debug issues without physical access to the device
2. **Predictive Maintenance**: Identify trends before they become critical
3. **Battery Assessment**: Understand power consumption patterns
4. **Thermal Management**: Monitor temperature trends in different conditions
5. **Storage Planning**: Track disk usage to prevent exhaustion
6. **Clock Accuracy**: Verify RTC and NTP are functioning correctly

## Future Enhancements

Potential improvements to consider:

1. **Configurable Monitoring Interval**: Make interval a command-line parameter
2. **Alert Thresholds**: Trigger warnings when metrics exceed thresholds
3. **Log Rotation**: Automatic rotation of debug logs to prevent disk exhaustion
4. **Web Dashboard**: Display monitoring metrics in the web interface
5. **Email Notifications**: Send alerts for critical events
6. **Historical Analysis**: Generate charts and graphs from log data

## Backward Compatibility

This implementation is fully backward compatible:

- Monitoring is optional (application continues if initialization fails)
- No changes to existing APIs or data structures
- Debug logs are separate from main application logs
- Can be disabled by not initializing the monitor

## Documentation

Complete documentation is available in:
- `monitoring/README.md` - Detailed monitoring guide
- `README.md` - Updated with monitoring overview
- Code comments - Inline documentation for all functions

## Verification

- ✅ All tests pass
- ✅ Code builds successfully
- ✅ No formatting issues (gofmt)
- ✅ No code quality issues (go vet)
- ✅ No security vulnerabilities (CodeQL)
- ✅ CI workflows updated
- ✅ Documentation complete
