# Deployment Guide

This guide covers deployment of the C++ Object Detection application to target Linux systems.

## Quick Start

### 1. Build the Application

```bash
cd cpp-object-detection
./scripts/build.sh
```

### 2. Download YOLO Model

```bash
./scripts/download_model.sh
```

### 3. Test the Application

```bash
# Test with help
./build/object_detection --help

# Test with mock camera (should fail gracefully)
./build/object_detection --camera-id 999 --verbose

# Run manual tests
./scripts/manual_test.sh
```

## System Requirements

### Hardware

- **CPU Architectures**: x86_64, 386 (32-bit)
- **Memory**: Minimum 2GB RAM (4GB+ recommended)
- **Storage**: 500MB for application and models
- **Camera**: USB Video Class (UVC) compatible webcam
  - Recommended: Logitech C920 or similar
  - 720p resolution support required

### Software

- **Operating System**: Linux (Ubuntu 18.04+, CentOS 7+, Debian 9+)
- **Dependencies**: 
  - OpenCV 4.x development libraries
  - CMake 3.16+
  - GCC 7+ or Clang 6+

### Network

- Internet connection for initial model download (~28MB)
- No ongoing network requirements for operation

## Installation

### Ubuntu/Debian

```bash
# Install dependencies
sudo apt-get update
sudo apt-get install -y cmake build-essential libopencv-dev pkg-config

# Clone repository
git clone https://github.com/ingojaeckel/go-raspberry-pi-timelapse.git
cd go-raspberry-pi-timelapse/cpp-object-detection

# Build
./scripts/build.sh

# Download model
./scripts/download_model.sh
```

### CentOS/RHEL

```bash
# Install dependencies
sudo yum install -y cmake gcc-c++ opencv-devel pkgconfig

# Build and setup (same as Ubuntu)
./scripts/build.sh
./scripts/download_model.sh
```

## Configuration

### Command Line Options

All configuration is done via command-line arguments:

```bash
./object_detection [OPTIONS]

Key Options:
  --max-fps N                Maximum processing rate (default: 5)
  --min-confidence N         Detection threshold (default: 0.5)
  --camera-id N             Camera device ID (default: 0)
  --log-file FILE           Log output location
  --heartbeat-interval N    Status logging interval in minutes
  --verbose                 Enable debug output
```

### Example Configurations

**Low-power deployment:**
```bash
./object_detection --max-fps 2 --min-confidence 0.8 --heartbeat-interval 15
```

**High-accuracy monitoring:**
```bash
./object_detection --max-fps 8 --min-confidence 0.3 --verbose
```

**Custom logging:**
```bash
./object_detection --log-file /var/log/security_detection.log --heartbeat-interval 30
```

## Production Deployment

### As Systemd Service

1. **Create service user:**
```bash
sudo useradd --system --no-create-home --shell /bin/false objdetect
sudo usermod -a -G video objdetect
```

2. **Install application:**
```bash
sudo mkdir -p /opt/object-detection
sudo cp build/object_detection /opt/object-detection/
sudo cp -r models /opt/object-detection/
sudo chown -R objdetect:objdetect /opt/object-detection
```

3. **Create systemd service:**
```bash
sudo tee /etc/systemd/system/object-detection.service << EOF
[Unit]
Description=Real-time Object Detection Service
After=network.target
Wants=network.target

[Service]
Type=simple
User=objdetect
Group=objdetect
WorkingDirectory=/opt/object-detection
ExecStart=/opt/object-detection/object_detection --log-file /var/log/object_detection.log --heartbeat-interval 10
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Resource limits
MemoryMax=1G
CPUQuota=80%

[Install]
WantedBy=multi-user.target
EOF
```

4. **Enable and start service:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable object-detection
sudo systemctl start object-detection

# Check status
sudo systemctl status object-detection
sudo journalctl -u object-detection -f
```

### SSH Deployment

For remote deployment and monitoring:

```bash
# Deploy to remote system
scp build/object_detection user@target:/usr/local/bin/
scp -r models user@target:/usr/local/share/object-detection/

# Start via SSH
ssh user@target 'nohup /usr/local/bin/object_detection --log-file /var/log/detection.log > /dev/null 2>&1 &'

# Monitor logs
ssh user@target 'tail -f /var/log/detection.log'

# Stop service
ssh user@target 'pkill object_detection'
```

## Monitoring and Maintenance

### Log Monitoring

The application produces structured logs:

```bash
# Monitor real-time logs
tail -f /var/log/object_detection.log

# Filter for detections only
grep "entered frame\|exited frame" /var/log/object_detection.log

# Monitor performance warnings
grep "Performance warning" /var/log/object_detection.log
```

### Log Rotation

Set up logrotate for the application logs:

```bash
sudo tee /etc/logrotate.d/object-detection << EOF
/var/log/object_detection.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 objdetect objdetect
    postrotate
        systemctl reload object-detection || true
    endscript
}
EOF
```

### Health Monitoring

Monitor application health:

```bash
# Check if process is running
pgrep -f object_detection

# Monitor memory usage
ps aux | grep object_detection

# Check recent heartbeats
tail -50 /var/log/object_detection.log | grep "heartbeat"

# Monitor detection rate
tail -100 /var/log/object_detection.log | grep "entered frame" | wc -l
```

## Troubleshooting

### Common Issues

**Application won't start:**
```bash
# Check dependencies
ldd build/object_detection

# Verify camera permissions
ls -l /dev/video*
groups $USER | grep video

# Test camera manually
v4l2-ctl --list-devices
```

**Low performance:**
```bash
# Check CPU usage
top -p $(pgrep object_detection)

# Monitor frame rates in logs
grep "Performance report" /var/log/object_detection.log

# Reduce processing load
./object_detection --max-fps 2 --min-confidence 0.7
```

**No objects detected:**
```bash
# Lower confidence threshold
./object_detection --min-confidence 0.3 --verbose

# Check model file
ls -la models/yolov5s.onnx
file models/yolov5s.onnx

# Verify camera view
ffmpeg -f v4l2 -i /dev/video0 -frames:v 1 test.jpg
```

### Performance Tuning

**For Intel Core i7 / AMD Ryzen systems:**
```bash
./object_detection --max-fps 10 --processing-threads 2 --enable-gpu
```

**For Intel Pentium M (32-bit):**
```bash
./object_detection --max-fps 1 --min-confidence 0.8 --frame-width 640 --frame-height 480
```

**For headless servers:**
```bash
./object_detection --headless --log-file /var/log/detection.log --heartbeat-interval 5
```

## Security Considerations

### File Permissions

```bash
# Secure the executable
chmod 755 /opt/object-detection/object_detection
chown root:objdetect /opt/object-detection/object_detection

# Secure log files
chmod 640 /var/log/object_detection.log
chown objdetect:adm /var/log/object_detection.log
```

### Network Security

- Application runs locally, no network ports opened
- Logs may contain sensitive detection data
- Consider encrypting log files for long-term storage

### Camera Access

- Limit camera access to dedicated service user
- Monitor camera usage to detect unauthorized access
- Consider physical camera controls for sensitive areas

## Scaling and Integration

### Multiple Cameras

Run separate instances for multiple cameras:

```bash
# Camera 0
./object_detection --camera-id 0 --log-file /var/log/camera0.log &

# Camera 1  
./object_detection --camera-id 1 --log-file /var/log/camera1.log &
```

### Log Integration

Integrate with syslog or log aggregation systems:

```bash
# Forward logs to syslog
./object_detection --log-file /dev/stdout | logger -t object-detection

# Integration with ELK stack, Splunk, etc.
# Configure log shipping from /var/log/object_detection.log
```

### Alerting

Set up alerts based on log patterns:

```bash
# Email alerts for specific detections
tail -f /var/log/object_detection.log | grep "person entered" | \
while read line; do
    echo "$line" | mail -s "Security Alert" admin@example.com
done
```

## Performance Benchmarks

### Typical Performance

| Platform | CPU | Expected FPS | Recommended Settings |
|----------|-----|--------------|---------------------|
| Intel Core i7-8700K | x86_64 | 8-15 fps | `--max-fps 10` |
| AMD Ryzen 5 3600 | x86_64 | 10-20 fps | `--max-fps 15` |
| Intel Pentium M 1.7GHz | 386 | 1-2 fps | `--max-fps 1 --frame-width 640` |
| Raspberry Pi 4 | ARM64 | 2-5 fps | `--max-fps 3` |

### Memory Usage

- **Base application**: ~50MB
- **With YOLO model**: ~150MB
- **Per frame processing**: ~10MB additional
- **Recommended system RAM**: 2GB minimum, 4GB preferred

## Support and Updates

### Getting Help

1. **Check logs**: Review application logs for error messages
2. **Run tests**: Use `./scripts/manual_test.sh` for diagnostics
3. **Check documentation**: Review README.md for detailed information
4. **GitHub Issues**: Report bugs and feature requests

### Updates

To update the application:

```bash
# Backup current installation
sudo systemctl stop object-detection
sudo cp /opt/object-detection/object_detection /opt/object-detection/object_detection.backup

# Update from source
git pull origin main
./scripts/build.sh

# Deploy update
sudo cp build/object_detection /opt/object-detection/
sudo systemctl start object-detection
```

### Model Updates

Update to newer YOLO models:

```bash
# Download newer model
wget -O models/yolov5m.onnx https://github.com/ultralytics/yolov5/releases/download/v6.2/yolov5m.onnx

# Update configuration
./object_detection --model-path models/yolov5m.onnx
```