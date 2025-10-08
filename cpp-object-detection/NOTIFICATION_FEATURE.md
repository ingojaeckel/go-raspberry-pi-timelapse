# Real-time Notification Mechanisms Feature

## Overview

The C++ object detection application now supports real-time notifications when new objects are detected in the scene. This feature allows users to be instantly alerted when a new type of object appears, enabling automated workflows, monitoring, and integrations.

## Supported Notification Mechanisms

### 1. Webhook / Callback URL (HTTP POST)

Send HTTP POST requests to a configured webhook URL whenever a new object is detected.

**Use Cases:**
- Integration with automation platforms (Zapier, IFTTT, n8n)
- Custom server-side processing
- Cloud logging and analytics
- Alerting systems (PagerDuty, Slack, Discord)

**Configuration:**
```bash
./object_detection \
  --enable-notifications \
  --enable-webhook \
  --webhook-url http://example.com/webhook
```

**Payload Format:**
```json
{
  "event": "new_object_detected",
  "timestamp": "2025-01-12 14:30:45",
  "object": {
    "type": "person",
    "x": 320.5,
    "y": 240.8,
    "confidence": 0.92
  },
  "all_detections": [
    {
      "class": "person",
      "confidence": 0.92,
      "bbox": {"x": 100, "y": 150, "width": 80, "height": 120}
    }
  ],
  "status": {
    "fps": 5.0,
    "avg_processing_time_ms": 150.5,
    "total_objects_detected": 42,
    "total_images_saved": 15,
    "brightness_filter_active": false,
    "gpu_enabled": true,
    "burst_mode_enabled": true
  },
  "top_objects": [
    {"type": "person", "count": 25},
    {"type": "car", "count": 12}
  ],
  "image": "base64_encoded_jpeg_with_bounding_boxes"
}
```

### 2. Server-Sent Events (SSE) / HTTP Push

Run an SSE server that browsers and applications can connect to for real-time push notifications.

**Use Cases:**
- Web dashboards and monitoring UIs
- Real-time browser notifications
- Mobile app integrations
- Low-latency event streaming

**Configuration:**
```bash
./object_detection \
  --enable-notifications \
  --enable-sse \
  --sse-port 8081
```

**Client Connection:**
```javascript
// JavaScript example
const eventSource = new EventSource('http://localhost:8081');

eventSource.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('New object detected:', data.object.type);
  
  // Display image with bounding boxes
  const img = document.getElementById('detection-image');
  img.src = 'data:image/jpeg;base64,' + data.image;
};

eventSource.onerror = function(error) {
  console.error('SSE error:', error);
};
```

```python
# Python example
import sseclient
import requests
import json

url = 'http://localhost:8081'
response = requests.get(url, stream=True)
client = sseclient.SSEClient(response)

for event in client.events():
    data = json.loads(event.data)
    print(f"New {data['object']['type']} detected at ({data['object']['x']}, {data['object']['y']})")
```

### 3. File-based Notifications

Append notification data to a JSON file for simple integration without network dependencies.

**Use Cases:**
- Log file monitoring with tools like `tail -f`
- File watchers and schedulers (cron, systemd timers)
- Simple integrations without network requirements
- Offline processing and batching

**Configuration:**
```bash
./object_detection \
  --enable-notifications \
  --enable-file-notification \
  --notification-file-path /var/log/object_notifications.json
```

**Monitoring:**
```bash
# Watch notifications in real-time
tail -f /var/log/object_notifications.json | jq .

# Process new notifications with a script
while inotifywait -e modify /var/log/object_notifications.json; do
  tail -n 1 /var/log/object_notifications.json | jq . | ./process_notification.sh
done
```

### 4. Stdio Notifications

Output notifications to standard output (stdout) for pipeline integration.

**Use Cases:**
- Unix pipe workflows
- Docker container logging (stdout → Docker logs)
- Integration with systemd journal
- Simple debugging and testing

**Configuration:**
```bash
./object_detection \
  --enable-notifications \
  --enable-stdio-notification
```

**Output Format:**
```
=== NEW OBJECT NOTIFICATION ===
{"event":"new_object_detected","timestamp":"2025-01-12 14:30:45",...}
===============================
```

**Pipeline Examples:**
```bash
# Filter and process specific object types
./object_detection --enable-notifications --enable-stdio-notification | \
  grep "person" | jq .object.type

# Send to remote logging server
./object_detection --enable-notifications --enable-stdio-notification 2>&1 | \
  nc logging-server.local 514

# Docker logging
docker logs -f object-detection-container | grep "NEW OBJECT NOTIFICATION"
```

## Notification Content

### Photo with Bounding Boxes

Each notification includes a base64-encoded JPEG image showing:
- The current camera frame
- Bounding boxes around all detected objects
- Object labels with confidence scores
- Same visual representation as the HTTP stream and viewfinder

### Status Information

All notifications include the same status information available in the HTTP stream:
- Current FPS and processing time
- Total objects detected
- Total images saved
- Top detected objects (type and count)
- Active features (brightness filter, GPU, burst mode)

## Multiple Notification Channels

You can enable multiple notification mechanisms simultaneously:

```bash
./object_detection \
  --enable-notifications \
  --enable-webhook --webhook-url http://example.com/webhook \
  --enable-sse --sse-port 8081 \
  --enable-file-notification --notification-file-path /tmp/notifications.json \
  --enable-stdio-notification
```

When multiple channels are enabled, notifications are sent to all of them in parallel.

## Performance Considerations

### Impact on Processing Speed

- **Webhook**: Minimal impact (~5-10ms per notification) due to asynchronous HTTP requests
- **SSE**: Minimal impact when clients are connected (~2-5ms per notification)
- **File**: Very low impact (~1-2ms per notification)
- **Stdio**: Negligible impact (<1ms per notification)

### Network Bandwidth

- Each notification with a 640x360 JPEG image is approximately 30-50 KB
- Base64 encoding increases payload size by ~33%
- For high-frequency detection scenarios, consider:
  - Reducing image resolution with `--detection-scale`
  - Using burst mode to reduce notifications for stationary objects
  - Implementing notification throttling in your webhook endpoint

### Memory Usage

- SSE server: ~100 KB per connected client
- Webhook: Temporary buffer for HTTP request (~50 KB)
- File/Stdio: Minimal memory usage (<10 KB)

## Security Considerations

### Current Implementation (Local Network Only)

The notification system is designed for **local network use only**:
- No authentication or encryption by default
- SSE server binds to all interfaces (0.0.0.0)
- Webhook URLs should be on trusted networks
- File-based notifications have standard file permissions

### Best Practices

1. **Use a reverse proxy** (nginx, Caddy) for SSL/TLS encryption
2. **Implement authentication** in your webhook endpoint
3. **Restrict network access** using firewall rules
4. **Validate webhook URLs** before deployment
5. **Monitor file permissions** for file-based notifications

## Integration Examples

### Slack Notification

```bash
# Using webhook with Slack incoming webhook
./object_detection \
  --enable-notifications \
  --enable-webhook \
  --webhook-url https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

Create a middleware endpoint that transforms the notification to Slack format.

### Home Assistant Integration

```yaml
# configuration.yaml
sensor:
  - platform: rest
    name: object_detection
    resource: http://localhost:8081
    method: GET
    value_template: "{{ value_json.object.type }}"
    json_attributes:
      - timestamp
      - object
      - status
```

### Discord Bot

```python
import discord
from discord.ext import commands
import requests
import json
import base64
from io import BytesIO

# Connect to SSE stream
def listen_for_detections():
    url = 'http://localhost:8081'
    response = requests.get(url, stream=True)
    client = sseclient.SSEClient(response)
    
    for event in client.events():
        data = json.loads(event.data)
        
        # Decode image
        image_data = base64.b64decode(data['image'])
        
        # Send to Discord
        channel = bot.get_channel(CHANNEL_ID)
        await channel.send(
            f"🚨 New {data['object']['type']} detected!",
            file=discord.File(BytesIO(image_data), 'detection.jpg')
        )
```

### Automation with n8n

1. Create webhook trigger node listening on the configured webhook URL
2. Add image processing node to decode base64 image
3. Add notification nodes (email, SMS, push notification)
4. Add conditional logic to filter object types

## Troubleshooting

### Webhook Not Receiving Notifications

1. Check webhook URL is accessible: `curl -X POST -d '{}' http://your-webhook-url`
2. Verify `--enable-notifications` and `--enable-webhook` are set
3. Check application logs for webhook errors
4. Test with a simple HTTP server: `python -m http.server 8000`

### SSE Connection Issues

1. Verify SSE port is not blocked by firewall
2. Check no other service is using the port: `netstat -tulpn | grep 8081`
3. Test with curl: `curl http://localhost:8081`
4. Ensure browser supports EventSource API

### File Notifications Not Written

1. Check write permissions on notification file path
2. Verify parent directory exists
3. Monitor disk space: `df -h`
4. Check application logs for file I/O errors

### Stdio Not Showing Notifications

1. Ensure `--enable-stdio-notification` is set
2. Check stdout is not redirected elsewhere
3. Verify notifications are actually being triggered (check logs)

## Future Enhancements

Potential improvements for future releases:

- **Authentication**: Add API key or token-based auth for webhooks and SSE
- **Encryption**: SSL/TLS support for secure communication
- **Message queuing**: Integration with RabbitMQ, Kafka, or MQTT
- **Cloud platforms**: Direct integration with AWS SNS, Google Cloud Pub/Sub, Azure Event Hub
- **Filtering**: Configure which object types trigger notifications
- **Rate limiting**: Prevent notification flooding
- **Retry logic**: Automatic retry for failed webhook deliveries
- **Notification templates**: Customizable message formats

## See Also

- [Network Streaming Feature](NETWORK_STREAMING_FEATURE.md) - MJPEG HTTP streaming
- [Google Sheets Integration](GOOGLE_SHEETS_FEATURE.md) - Cloud logging
- [Burst Mode](BURST_MODE_FEATURE.md) - High-speed capture for new objects
- [Photo Storage](PHOTO_STORAGE_FEATURE.md) - Automated photo capture with bounding boxes
