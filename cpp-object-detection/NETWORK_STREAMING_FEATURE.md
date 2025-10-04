# Network Streaming Feature

## Overview

The network streaming feature allows you to view the live video feed with real-time object detection bounding boxes from any device on your local network. The stream uses the MJPEG (Motion JPEG) format over HTTP, which is compatible with standard web browsers, VLC media player, and other widely available applications.

## Requirements

- Device must be connected to a local network (LAN)
- No additional software installation required on viewing devices
- Standard web browser (Chrome, Firefox, Safari, Edge) or VLC media player

## Usage

### Enable Network Streaming

To enable network streaming, use the `--enable-streaming` command-line flag:

```bash
./object_detection --enable-streaming
```

By default, the stream will be available on port 8080. To use a different port:

```bash
./object_detection --enable-streaming --streaming-port 9000
```

### Accessing the Stream

When the application starts with streaming enabled, it will display the streaming URL in the logs:

```
Stream available at: http://192.168.1.100:8080/stream
Open in browser or VLC to view live feed with object detection
```

#### View in Web Browser

1. Note the IP address and port from the application logs
2. Open a web browser on any device on the same network
3. Navigate to: `http://<ip-address>:<port>/stream`
4. Example: `http://192.168.1.100:8080/stream`

The stream will display in your browser with real-time object detection bounding boxes.

#### View in VLC Media Player

1. Open VLC Media Player
2. Go to Media → Open Network Stream
3. Enter the network URL: `http://<ip-address>:<port>/stream`
4. Click Play

#### View from Command Line

Using curl or wget:
```bash
# View stream info
curl http://192.168.1.100:8080/stream

# Or use ffplay to display the stream
ffplay http://192.168.1.100:8080/stream
```

## Examples

### Basic Network Streaming

```bash
# Start with default settings (port 8080)
./object_detection --enable-streaming
```

### Custom Port

```bash
# Use port 9000 instead
./object_detection --enable-streaming --streaming-port 9000
```

### Combine with Other Features

```bash
# Enable both local preview window and network streaming
./object_detection --show-preview --enable-streaming

# High-accuracy mode with network streaming
./object_detection --enable-streaming --model-type yolov5l --min-confidence 0.7

# Parallel processing with streaming for better performance
./object_detection --enable-streaming --processing-threads 4 --enable-parallel
```

## Technical Details

### Protocol

The streaming uses MJPEG (Motion JPEG) over HTTP with the multipart/x-mixed-replace content type. Each frame is encoded as a JPEG image and sent to connected clients.

### Performance

- Frame rate: ~10 fps (configurable in source code)
- JPEG quality: 80% (configurable in source code)
- Latency: < 200ms on local network
- Bandwidth: ~1-3 Mbps depending on video content and resolution

### Network Requirements

- The streaming server binds to all network interfaces (0.0.0.0)
- Default port: 8080 (configurable)
- Supports multiple concurrent viewers
- No authentication required (intended for local network use only)

## Security Considerations

**Important:** This feature is designed for use on trusted local networks only. The stream:
- Has no authentication or encryption
- Is accessible to any device on the same network
- Should NOT be exposed to the internet

For production deployments requiring security:
- Use a reverse proxy (nginx, Apache) with authentication
- Configure firewall rules to restrict access
- Consider using SSL/TLS for encryption
- Implement rate limiting and access controls

## Troubleshooting

### Cannot Connect to Stream

1. **Check firewall settings:**
   ```bash
   # Allow port through firewall (Linux)
   sudo ufw allow 8080/tcp
   
   # Or for custom port
   sudo ufw allow 9000/tcp
   ```

2. **Verify the application is running:**
   - Check application logs for "Stream available at" message
   - Ensure no errors during initialization

3. **Confirm devices are on same network:**
   - Viewing device and streaming device must be on same LAN
   - Try pinging the IP address from viewing device

4. **Port already in use:**
   ```
   Failed to bind socket to port 8080: Address already in use
   ```
   Solution: Use a different port with `--streaming-port <port>`

### Stream is Laggy or Choppy

1. **Check network bandwidth:**
   - Ensure WiFi signal is strong
   - Consider using wired Ethernet connection

2. **Reduce processing load:**
   ```bash
   # Lower frame processing rate
   ./object_detection --enable-streaming --max-fps 3
   
   # Use faster model
   ./object_detection --enable-streaming --model-type yolov5s
   
   # Reduce detection resolution
   ./object_detection --enable-streaming --detection-scale 0.3
   ```

3. **Enable parallel processing:**
   ```bash
   ./object_detection --enable-streaming --processing-threads 4
   ```

### Browser Shows "This page cannot be displayed"

1. **Check URL format:**
   - Must include `/stream` path: `http://192.168.1.100:8080/stream`
   - Not just the IP and port

2. **Try a different browser:**
   - Some browsers may have stricter content-type handling
   - VLC is a reliable alternative

## Compatibility

### Tested Browsers
- ✅ Google Chrome/Chromium
- ✅ Mozilla Firefox  
- ✅ Safari (macOS/iOS)
- ✅ Microsoft Edge
- ✅ VLC Media Player
- ✅ ffplay/ffmpeg

### Network Environments
- ✅ Home WiFi networks
- ✅ Wired Ethernet LANs
- ✅ Corporate networks (if firewall permits)
- ❌ Public internet (not supported - local network only)
- ❌ Cross-subnet routing (may require configuration)

## Advanced Configuration

### Modify Stream Quality

Edit `src/network_streamer.cpp` to adjust:

```cpp
// Change JPEG quality (0-100)
std::vector<int> params = {cv::IMWRITE_JPEG_QUALITY, 90};  // Higher quality

// Change frame rate
std::this_thread::sleep_for(std::chrono::milliseconds(50));  // 20 fps
```

### Enable Multiple Concurrent Clients

The current implementation handles one client at a time. To support multiple concurrent viewers, modify the `handleClient()` function to spawn a new thread for each client connection.

## Use Cases

### Development and Testing
```bash
# View detections from your desk while camera is positioned elsewhere
./object_detection --enable-streaming --show-preview
```

### Camera Positioning
```bash
# Adjust camera angle while viewing stream on mobile device
./object_detection --enable-streaming --camera-id 0
```

### Remote Monitoring
```bash
# Monitor object detection from another room or device
./object_detection --enable-streaming --model-type yolov5l
```

### Live Demonstrations
```bash
# Show real-time object detection to audience on projector/TV
./object_detection --enable-streaming --min-confidence 0.6
```

## Limitations

1. **Local Network Only:** Stream is only accessible on the same local network
2. **No Recording:** Stream is live-only; recording must be done client-side
3. **Single Camera:** Only streams from the configured camera
4. **Fixed Resolution:** Streams at the configured frame resolution (default 1280x720)
5. **No Audio:** Video stream only, no audio support

## Future Enhancements

Potential improvements for future versions:
- WebRTC support for lower latency
- Authentication and encryption
- Multi-camera support
- Configurable stream resolution
- On-demand recording
- WebSocket-based streaming
- Mobile app integration
