# System Monitoring and Debug Logging

This package provides comprehensive system monitoring and debug logging capabilities for long-term, remote deployments of the timelapse system, especially for solar-powered or battery-backed installations.

## Features

### Automatic Debug Information Recording

The monitoring system automatically tracks and logs:

1. **System Uptime and Runtime**
   - Tracks total runtime since startup
   - Records max daily runtime (useful for battery capacity assessment)
   - Automatically resets daily runtime tracking at midnight

2. **Temperature Monitoring**
   - CPU temperature from `/sys/class/thermal/thermal_zone0/temp`
   - GPU temperature via `vcgencmd measure_temp`
   - Helps identify thermal issues in enclosed deployments

3. **Disk Usage Tracking**
   - Monitors free disk space over time
   - Critical for identifying storage exhaustion issues

4. **System Clock State**
   - Records system time with timezone information
   - Helps identify RTC (Real-Time Clock) drift or NTP sync issues

5. **Startup/Shutdown Events**
   - Logs system startup with timestamp
   - Records shutdown events with total runtime
   - Helps identify unexpected restarts or power loss

## Usage

The monitoring system is automatically initialized and integrated when the application starts:

```go
monitor, err := monitoring.New()
if err != nil {
    log.Printf("Warning: Failed to initialize system monitor: %s\n", err.Error())
} else {
    defer monitor.Close()
}
```

### Periodic Checks

The monitoring system performs periodic checks every 5 minutes by default. It's integrated into the timelapse capture loop:

```go
if t.Monitor != nil {
    t.Monitor.PeriodicCheck()
}
```

## Debug Log File

All monitoring information is stored in `system-debug.log` with two types of entries:

### Event Logs

Events like startup and shutdown are logged in a human-readable format:

```
[STARTUP] 2025-10-06T20:27:03Z - System started
[SHUTDOWN] 2025-10-06T23:45:12Z - System shutdown after 11789 seconds runtime
[DAILY_RUNTIME] 2025-10-07T00:00:05Z - Max runtime for previous day: 11789 seconds
```

### System Statistics

System statistics are logged in JSON format for easy parsing:

```json
[STATS] {"timestamp":"2025-10-06T20:32:03Z","uptime":"20:32:03 up 1:05, 1 user, load average: 0.15, 0.10, 0.05","cpu_temperature":"45123","gpu_temperature":"temp=45.0'C","free_disk_space":"Filesystem      Size  Used Avail Use% Mounted on\n/dev/root        15G  4.2G   10G  30% /","system_clock":"2025-10-06 20:32:03 UTC","runtime_seconds":300}
```

## Monitoring Interval

The default monitoring interval is 5 minutes, defined by the `MonitoringInterval` constant in `monitor.go`. This can be adjusted based on deployment needs:

- For battery-powered systems: Consider longer intervals (10-15 minutes) to reduce power consumption
- For critical monitoring: Use shorter intervals (1-2 minutes) for more frequent checks

## Analyzing Debug Logs

### Extract Startup/Shutdown Events

```bash
grep -E '\[STARTUP\]|\[SHUTDOWN\]' system-debug.log
```

### Parse System Statistics

```bash
grep '\[STATS\]' system-debug.log | sed 's/\[STATS\] //' | jq '.'
```

### Temperature Trends

```bash
grep '\[STATS\]' system-debug.log | sed 's/\[STATS\] //' | jq -r '[.timestamp, .cpu_temperature] | @csv'
```

### Daily Runtime Analysis

```bash
grep '\[DAILY_RUNTIME\]' system-debug.log
```

## Troubleshooting Use Cases

### 1. Identifying Unexpected Restarts

Look for startup events without corresponding shutdown events:

```bash
grep '\[STARTUP\]' system-debug.log
```

If there are more startup events than shutdown events, the system experienced unexpected restarts (power loss, crashes, etc.).

### 2. Temperature Issues

Extract CPU temperatures to identify thermal problems:

```bash
grep '\[STATS\]' system-debug.log | sed 's/\[STATS\] //' | jq -r '.cpu_temperature' | \
    awk '{sum+=$1; count++} END {print "Average CPU temp (raw): " sum/count}'
```

Note: CPU temperature is in millidegrees Celsius (e.g., 45123 = 45.123°C)

### 3. Battery Capacity Assessment

Review max daily runtime to determine if battery capacity is sufficient:

```bash
grep '\[DAILY_RUNTIME\]' system-debug.log
```

If max runtime is decreasing over time, it may indicate battery degradation.

### 4. Disk Space Exhaustion

Track disk usage trends:

```bash
grep '\[STATS\]' system-debug.log | sed 's/\[STATS\] //' | jq -r '[.timestamp, .free_disk_space] | @csv'
```

### 5. System Clock Drift

Compare recorded system clock with expected values:

```bash
grep '\[STATS\]' system-debug.log | sed 's/\[STATS\] //' | jq -r '[.timestamp, .system_clock] | @csv'
```

Significant differences may indicate RTC battery issues or NTP sync problems.

## Integration with Main Application

The monitoring system is designed to be non-intrusive and fail-safe:

- If monitoring initialization fails, the application continues without monitoring
- Monitoring runs asynchronously and doesn't block timelapse operations
- Failed monitoring commands (e.g., missing `vcgencmd`) don't crash the application
- Debug logs are separate from main application logs

## Configuration

Current configuration constants in `monitor.go`:

```go
const (
    DebugLogFile = "system-debug.log"
    MonitoringInterval = 5 * time.Minute
)
```

These can be adjusted based on specific deployment requirements.
