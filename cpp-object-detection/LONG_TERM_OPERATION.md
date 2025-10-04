# Long-Term Operation Guide

This document describes the optimizations and best practices for running the C++ Object Detection application continuously for extended periods (days, weeks, or months).

## Overview

The application has been designed to run reliably for long periods without manual intervention. Key optimizations include:

- **Memory leak prevention** through bounded data structures
- **Camera resilience** with automatic reconnection
- **Resource monitoring** with warnings and logging
- **Overflow protection** for long-running counters

## Automatic Memory Management

### Bounded Object Tracking

The application limits the number of tracked objects to prevent unbounded memory growth:

- **MAX_TRACKED_OBJECTS**: 100 concurrent objects maximum
- **MAX_OBJECT_TYPE_ENTRIES**: 50 different object types tracked
- **Cleanup policy**: Oldest 20% removed when limit reached

These limits are appropriate for most monitoring scenarios where you're tracking people, vehicles, and animals entering/exiting a frame.

**Configuration** (in `object_detector.hpp`):
```cpp
static constexpr size_t MAX_TRACKED_OBJECTS = 100;
static constexpr int MAX_OBJECT_TYPE_ENTRIES = 50;
```

### Performance Counter Overflow Protection

Frame processing counters are automatically reset after 1,000,000 frames to prevent integer overflow:

- At 1 fps: ~11.5 days before reset
- At 5 fps: ~2.3 days before reset
- At 10 fps: ~1.15 days before reset

The reset preserves average processing time metrics while preventing overflow.

## Camera Health and Resilience

### USB Camera Standby Prevention

USB cameras may enter power-saving mode after periods of inactivity. The application prevents this through:

- **Keep-alive interval**: Every 30 seconds
- **Method**: Lightweight property query (`CAP_PROP_FPS`)
- **Impact**: Minimal CPU usage, prevents camera sleep

### Automatic Reconnection

The camera health monitoring system tracks consecutive capture failures:

- **Failure threshold**: 5 consecutive failures
- **Health check interval**: Every 60 seconds
- **Recovery action**: Automatic reconnection attempt
- **Reconnection delay**: 2 seconds between attempts

**Behavior**:
1. Application detects camera failures during frame capture
2. After 5 consecutive failures, health check triggers
3. Camera is released and reconnection attempted
4. If reconnection fails, application logs error and exits
5. System supervisor (systemd) can restart application automatically

### Camera Disconnect/Reconnect Scenarios

**Expected behavior**:
- Camera physically disconnected → Health check detects failure → Reconnection attempts
- Camera USB power-saving → Keep-alive prevents (no action needed)
- Camera driver crash → Health check detects → Reconnection attempts

## System Resource Monitoring

### Disk Space Management

The application monitors available disk space and logs warnings when space is low:

**Monitoring interval**: Every 5 minutes

**Thresholds**:
- **Warning**: 90% disk usage - logs warning message
- **Critical**: 95% disk usage or < 100 MB free - logs error message

**Actions taken**:
- Logs warning when disk usage exceeds 90%
- Logs error when disk usage exceeds 95% or free space < 100 MB
- No automatic file deletion - requires manual intervention

**Manual cleanup** (when needed):
```bash
# Find and remove photos older than 7 days
find detections/ -name "*.jpg" -mtime +7 -delete

# Keep only last 1000 photos
ls -t detections/*.jpg | tail -n +1001 | xargs rm --
```

### CPU Temperature Monitoring

The application monitors CPU temperature to detect overheating (Linux only):

**Temperature sources** (in order of preference):
1. `/sys/class/thermal/thermal_zone0/temp`
2. `/sys/devices/virtual/thermal/thermal_zone0/temp`

**Thresholds**:
- **Warning**: 75°C
- **Critical**: 85°C

**Actions**:
- Logs temperature warnings/errors
- No automatic throttling (relies on system thermal management)

**Note**: Temperature monitoring is unavailable on systems without thermal zones.

### System Statistics Logging

Every 5 minutes, the application logs:
```
System statistics: Disk: 45.2% used, 5234.5 MB free | CPU temp: 62.3°C
```

This helps track resource trends over time and identify issues before they become critical.

## Best Practices for Long-Term Operation

### 1. Use a Process Supervisor

Always run the application under a process supervisor like systemd to automatically restart on crashes:

```ini
# /etc/systemd/system/object-detection.service
[Unit]
Description=Object Detection Service
After=network.target

[Service]
Type=simple
User=detection
WorkingDirectory=/opt/object-detection
ExecStart=/opt/object-detection/build/object_detection \
    --log-file /var/log/object_detection.log \
    --heartbeat-interval 10 \
    --analysis-rate-limit 1
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable object-detection
sudo systemctl start object-detection
```

### 2. Configure Log Rotation

Prevent log files from consuming all disk space:

```
# /etc/logrotate.d/object-detection
/var/log/object_detection.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    copytruncate
}
```

### 3. Monitor Application Health

Check logs periodically:
```bash
# View recent logs
tail -f /var/log/object_detection.log

# Check for errors
grep -i error /var/log/object_detection.log

# Check for warnings
grep -i warning /var/log/object_detection.log

# With systemd
journalctl -u object-detection -f
```

### 4. Set Appropriate Rate Limits

For long-term operation, consider lower analysis rates to reduce CPU usage and wear:

```bash
# Analyze once per second (very light load)
./object_detection --analysis-rate-limit 1

# Analyze twice per second (balanced)
./object_detection --analysis-rate-limit 2

# Analyze every 2 seconds (minimal load)
./object_detection --analysis-rate-limit 0.5
```

### 5. Allocate Sufficient Disk Space

Calculate storage requirements:

- Image size: ~200 KB per detection photo (720p JPEG)
- Detection frequency: Depends on activity and photo interval (10 seconds minimum)
- Example: 100 detections/day × 200 KB = 20 MB/day = 600 MB/month

**Recommendations**:
- Development/testing: 1 GB minimum
- Light monitoring: 5 GB (handles ~8 months)
- Active monitoring: 20 GB (handles ~3 years)

### 6. Monitor System Resources

Use standard Linux tools to monitor the application:

```bash
# CPU and memory usage
top -p $(pgrep object_detection)

# Detailed memory info
cat /proc/$(pgrep object_detection)/status

# Disk usage for detections
du -sh detections/

# Camera device status
ls -l /dev/video*
v4l2-ctl --device=/dev/video0 --all
```

## Troubleshooting Long-Term Operation Issues

### Camera Stops Working After Days/Weeks

**Symptoms**: No frames captured, health check failures

**Solutions**:
1. Check USB connection and power
2. Verify camera appears in system: `lsusb | grep -i camera`
3. Check kernel logs: `dmesg | tail -50 | grep -i video`
4. Manually restart: `systemctl restart object-detection`
5. If persistent, check USB cable quality and port

### Disk Full Despite Warnings

**Symptoms**: Critical disk space warnings in logs

**Solutions**:
1. Monitor disk usage regularly with `df -h`
2. Set up alerts for disk space warnings in logs
3. Manually cleanup old photos periodically:
   ```bash
   find detections/ -name "*.jpg" -mtime +7 -delete
   ```
4. Consider external storage or network storage for detections
5. Implement log rotation for application logs

### Performance Degradation Over Time

**Symptoms**: Increasing processing time, lower FPS

**Solutions**:
1. Check CPU temperature: May indicate thermal throttling
2. Verify memory usage isn't growing: `top -p $(pgrep object_detection)`
3. Check for disk I/O bottlenecks: `iostat -x 1`
4. Restart application to reset all counters
5. Review log for memory/counter warnings

### Memory Usage Growing

**Symptoms**: Increasing RSS in `top`, system slowdown

**Solutions**:
1. Check if counters approaching limits (logged when limits hit)
2. Verify bounded data structure limits are effective
3. Check for leaked cv::Mat objects (should auto-clean)
4. Restart application if memory continues growing
5. Report issue if problem persists

### High CPU Temperature

**Symptoms**: Temperature warnings/errors in logs

**Solutions**:
1. Improve cooling (add fan, check airflow)
2. Reduce analysis rate: `--analysis-rate-limit 0.5`
3. Reduce detection scale: `--detection-scale 0.25`
4. Use faster YOLO model: `--model-type yolov8n`
5. Consider system thermal management settings

## Performance Tuning for Long-Term Operation

### Energy-Efficient Configuration

For Raspberry Pi or low-power systems:

```bash
./object_detection \
    --analysis-rate-limit 0.5 \
    --detection-scale 0.25 \
    --model-type yolov8n \
    --max-fps 2 \
    --heartbeat-interval 15
```

**Expected behavior**:
- CPU usage: 5-15% average
- Memory: ~100-200 MB
- Temperature: Minimal increase
- Battery/solar compatible

### High-Reliability Configuration

For security or critical monitoring:

```bash
./object_detection \
    --analysis-rate-limit 2 \
    --detection-scale 0.5 \
    --model-type yolov5l \
    --min-confidence 0.7 \
    --heartbeat-interval 5
```

**Expected behavior**:
- Higher accuracy detection
- More frequent logging
- Higher resource usage
- Suitable for mains power

### Balanced Configuration (Recommended)

For general-purpose long-term monitoring:

```bash
./object_detection \
    --analysis-rate-limit 1 \
    --detection-scale 0.5 \
    --model-type yolov5s \
    --min-confidence 0.5 \
    --heartbeat-interval 10
```

**Expected behavior**:
- Good detection accuracy
- Moderate resource usage
- Suitable for most scenarios

## Metrics and Monitoring

### Key Metrics to Track

1. **Frame processing rate**: Should match configured `--max-fps`
2. **Detection frequency**: Varies by scene activity
3. **Disk usage trend**: Monitor and manage manually
4. **CPU temperature**: Should remain stable
5. **Memory usage**: Should stabilize within first hour

### Log Analysis

Extract useful information from logs:

```bash
# Count detections per object type
grep "detected" object_detection.log | awk '{print $NF}' | sort | uniq -c

# Average CPU temperature over time
grep "CPU temp" object_detection.log | awk '{print $NF}' | sed 's/°C//' | \
    awk '{sum+=$1; n++} END {print "Average:", sum/n, "°C"}'

# Disk space warnings
grep -i "disk" object_detection.log | grep -i "warning\|critical"

# System uptime from heartbeat logs
grep "heartbeat" object_detection.log | tail -1
```

## Limits and Constraints

### System Limits

- **Concurrent tracked objects**: 100
- **Object type history**: 50 types
- **Frame counter**: 1,000,000 before reset
- **Resource checks**: Every 5 minutes

### Recommended Operational Limits

- **Continuous uptime**: Unlimited (tested up to months)
- **Detection photos**: Limited by disk space
- **Log file size**: Use log rotation (recommended 30 days)
- **Camera resolution**: Up to 1920x1080 (higher may require tuning)

## Support and Debugging

### Enable Debug Logging

For detailed troubleshooting:

```bash
./object_detection --verbose --log-file debug.log
```

This logs:
- Every frame capture attempt
- Detection processing details
- Resource check results
- Keep-alive operations
- Counter reset events

### Health Check Summary

Check application health:

```bash
# Is it running?
systemctl status object-detection

# Recent errors?
journalctl -u object-detection --since "1 hour ago" | grep -i error

# Resource usage?
ps aux | grep object_detection

# Disk space?
df -h $(pwd)/detections

# Camera connected?
v4l2-ctl --list-devices
```

## Future Improvements

Potential enhancements for even more robust long-term operation:

1. **Adaptive rate limiting** based on CPU temperature
2. **Quality-based photo storage** to reduce disk usage
3. **Network health monitoring** if streaming is enabled
4. **Automatic model switching** based on system load
5. **Automatic disk cleanup** with configurable policies (currently manual only)

## Conclusion

The application is designed for reliable long-term operation with minimal manual intervention. Key design principles:

- **Fail-safe**: Bounded data structures prevent unbounded growth
- **Self-healing**: Automatic camera reconnection
- **Observable**: Comprehensive logging and metrics
- **Efficient**: Configurable rate limiting and resource usage
- **Maintainable**: Clear limits and thresholds

For most use cases, the default configuration provides a good balance. Adjust settings based on your specific requirements for accuracy, resource usage, and reliability.
