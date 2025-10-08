# Notification Examples

This directory contains working examples demonstrating the real-time notification mechanisms in the C++ object detection application.

## Examples Included

### 1. notification_demo.sh
**Description:** Interactive demo script showing all notification mechanisms

**Usage:**
```bash
bash notification_demo.sh
```

**Features:**
- Shows configuration for all 4 notification types
- Provides use case examples
- Includes testing tips
- Copy-paste ready commands

### 2. sse_dashboard.html
**Description:** Real-time web dashboard using Server-Sent Events

**Usage:**
```bash
# Start object detection with SSE
./object_detection --enable-notifications --enable-sse --sse-port 8081

# Open sse_dashboard.html in your browser
# Or use a simple web server:
python3 -m http.server 8000
# Then navigate to: http://localhost:8000/sse_dashboard.html
```

**Features:**
- Real-time object detection notifications
- Live image display with bounding boxes
- System status monitoring (FPS, processing time, etc.)
- Notification history
- Modern, responsive UI
- Auto-reconnect on connection loss

**Screenshot:**
- Latest detection image with bounding boxes
- Live FPS and performance metrics
- Scrollable notification history
- Connection status indicator

### 3. webhook_receiver.py
**Description:** Python HTTP server to receive webhook notifications

**Usage:**
```bash
# Terminal 1: Start webhook receiver
python3 webhook_receiver.py 9000

# Terminal 2: Start object detection
./object_detection \
  --enable-notifications \
  --enable-webhook \
  --webhook-url http://localhost:9000/webhook
```

**Features:**
- Receives and parses webhook notifications
- Displays detection details in console
- Extracts and saves images to `webhook_images/` directory
- Shows system status and top detected objects
- Simple and extensible Python code

**Output Example:**
```
============================================================
New Notification Received at 2025-01-12 14:30:45
============================================================
Event: new_object_detected
Timestamp: 2025-01-12 14:30:45

Detected Object:
  Type: person
  Position: (320.5, 240.8)
  Confidence: 92.3%

System Status:
  FPS: 5.0
  Processing Time: 150.5ms
  Total Objects Detected: 42
  Total Images Saved: 15
  GPU Enabled: True
  Burst Mode: True

All Detections in Frame (2):
  - person (92.3% confidence)
  - car (85.1% confidence)

Top Detected Objects:
  - person: 25 times
  - car: 12 times

Image saved: webhook_images/2025-01-12_14-30-45_person.jpg
```

## Quick Start

### Test Stdio Notifications
```bash
./object_detection --enable-notifications --enable-stdio-notification 2>&1 | grep "NEW OBJECT"
```

### Test File Notifications
```bash
# Terminal 1
./object_detection \
  --enable-notifications \
  --enable-file-notification \
  --notification-file-path /tmp/notifications.json

# Terminal 2
tail -f /tmp/notifications.json | jq .
```

### Test SSE Notifications
```bash
# Terminal 1
./object_detection --enable-notifications --enable-sse --sse-port 8081

# Terminal 2 (or browser)
curl http://localhost:8081
# Or open sse_dashboard.html in browser
```

### Test Webhook Notifications
```bash
# Terminal 1
python3 webhook_receiver.py 9000

# Terminal 2
./object_detection \
  --enable-notifications \
  --enable-webhook \
  --webhook-url http://localhost:9000/webhook
```

### Test All Mechanisms Together
```bash
# Terminal 1: Webhook receiver
python3 webhook_receiver.py 9000

# Terminal 2: Object detection with all notifications
./object_detection \
  --enable-notifications \
  --enable-webhook --webhook-url http://localhost:9000/webhook \
  --enable-sse --sse-port 8081 \
  --enable-file-notification --notification-file-path /tmp/notifications.json \
  --enable-stdio-notification

# Browser: Open sse_dashboard.html
# Terminal 3: Monitor file
tail -f /tmp/notifications.json | jq .
```

## Integration Examples

### Home Assistant
```bash
./object_detection \
  --enable-notifications \
  --enable-webhook \
  --webhook-url http://homeassistant.local:8123/api/webhook/object_detection
```

### Discord Bot
Modify `webhook_receiver.py` to forward notifications to Discord using webhooks.

### Slack Integration
Use webhook URL from Slack incoming webhook configuration.

### Node-RED
Connect to SSE endpoint from Node-RED using `http in` node configured for SSE.

### Log Aggregation (ELK Stack)
```bash
./object_detection \
  --enable-notifications \
  --enable-file-notification \
  --notification-file-path /var/log/detections.json

# Filebeat configuration to ship to Elasticsearch
```

## Customization

### Modify SSE Dashboard
Edit `sse_dashboard.html`:
- Change color scheme by modifying CSS
- Add custom charts or visualizations
- Filter notifications by object type
- Adjust refresh rate and history size

### Extend Webhook Receiver
Edit `webhook_receiver.py`:
- Forward to other services (Discord, Slack, etc.)
- Add database storage
- Implement filtering logic
- Add email/SMS notifications

### Create Custom Clients

**JavaScript (Browser):**
```javascript
const eventSource = new EventSource('http://localhost:8081');
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('New object:', data.object.type);
};
```

**Python (SSE Client):**
```python
import sseclient
import requests

response = requests.get('http://localhost:8081', stream=True)
client = sseclient.SSEClient(response)

for event in client.events():
    data = json.loads(event.data)
    print(f"Detected: {data['object']['type']}")
```

**cURL (Testing):**
```bash
curl http://localhost:8081
```

**Node.js:**
```javascript
const EventSource = require('eventsource');
const es = new EventSource('http://localhost:8081');

es.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Detection:', data.object.type);
};
```

## Troubleshooting

### SSE Connection Issues
- Check firewall: `sudo ufw allow 8081`
- Verify port not in use: `netstat -tulpn | grep 8081`
- Test with curl: `curl http://localhost:8081`

### Webhook Not Receiving
- Check URL is accessible: `curl -X POST -d '{}' http://webhook-url`
- Verify webhook receiver is running
- Check application logs for errors

### File Permissions
- Ensure write permissions: `chmod 666 /tmp/notifications.json`
- Verify directory exists: `mkdir -p /var/log`

### Image Not Displayed
- Check base64 encoding is valid
- Verify JPEG quality setting
- Ensure sufficient memory for encoding

## See Also

- [NOTIFICATION_FEATURE.md](../NOTIFICATION_FEATURE.md) - Complete feature documentation
- [NOTIFICATION_IMPLEMENTATION_SUMMARY.md](../NOTIFICATION_IMPLEMENTATION_SUMMARY.md) - Implementation details
- [README.md](../README.md) - Main application documentation
