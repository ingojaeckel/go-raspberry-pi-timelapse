#!/bin/bash

# Notification Mechanisms Demo Script
# This script demonstrates the different notification mechanisms available
# in the C++ object detection application

echo "======================================"
echo "Notification Mechanisms Demo"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}This demo shows how to use the notification system${NC}"
echo ""

# 1. Stdio Notification Demo
echo -e "${GREEN}1. Stdio Notification Demo${NC}"
echo "   Outputs notifications directly to stdout"
echo "   Command: ./object_detection --enable-notifications --enable-stdio-notification"
echo ""

# 2. File Notification Demo
echo -e "${GREEN}2. File-based Notification Demo${NC}"
echo "   Appends JSON notifications to a file"
echo "   Command: ./object_detection --enable-notifications --enable-file-notification --notification-file-path /tmp/notifications.json"
echo "   Monitor with: tail -f /tmp/notifications.json | jq ."
echo ""

# 3. SSE Server Demo
echo -e "${GREEN}3. Server-Sent Events (SSE) Demo${NC}"
echo "   Runs an SSE server that browsers can connect to"
echo "   Command: ./object_detection --enable-notifications --enable-sse --sse-port 8081"
echo "   Connect from browser: http://localhost:8081"
echo ""
echo "   JavaScript client example:"
echo "   -------------------------"
cat << 'EOF'
   const eventSource = new EventSource('http://localhost:8081');
   eventSource.onmessage = function(event) {
     const data = JSON.parse(event.data);
     console.log('New object detected:', data.object.type);
     // Display image: img.src = 'data:image/jpeg;base64,' + data.image;
   };
EOF
echo ""

# 4. Webhook Demo
echo -e "${GREEN}4. Webhook/Callback URL Demo${NC}"
echo "   Sends HTTP POST to a webhook URL"
echo "   Command: ./object_detection --enable-notifications --enable-webhook --webhook-url http://localhost:9000/webhook"
echo ""
echo "   Test webhook server (Python):"
echo "   -----------------------------"
cat << 'EOF'
   python3 -c "
   from http.server import HTTPServer, BaseHTTPRequestHandler
   import json
   
   class WebhookHandler(BaseHTTPRequestHandler):
       def do_POST(self):
           content_length = int(self.headers['Content-Length'])
           body = self.rfile.read(content_length)
           data = json.loads(body)
           print(f'Received notification: {data[\"object\"][\"type\"]} at ({data[\"object\"][\"x\"]}, {data[\"object\"][\"y\"]})')
           self.send_response(200)
           self.end_headers()
   
   HTTPServer(('localhost', 9000), WebhookHandler).serve_forever()
   "
EOF
echo ""

# 5. Multiple Channels Demo
echo -e "${GREEN}5. Multiple Notification Channels Demo${NC}"
echo "   Enable multiple notification mechanisms at once"
echo "   Command:"
echo "   ./object_detection \\"
echo "     --enable-notifications \\"
echo "     --enable-webhook --webhook-url http://localhost:9000/webhook \\"
echo "     --enable-sse --sse-port 8081 \\"
echo "     --enable-file-notification --notification-file-path /tmp/notifications.json \\"
echo "     --enable-stdio-notification"
echo ""

# Use Cases
echo -e "${YELLOW}Common Use Cases:${NC}"
echo ""
echo "📱 Home Automation:"
echo "   - Webhook → Home Assistant/SmartThings"
echo "   - Trigger lights, alarms, or recordings"
echo ""
echo "📊 Monitoring Dashboard:"
echo "   - SSE → Web dashboard with live updates"
echo "   - Real-time object tracking visualization"
echo ""
echo "🔔 Alert Systems:"
echo "   - Webhook → Slack/Discord/Telegram bot"
echo "   - Instant notifications on mobile devices"
echo ""
echo "📝 Log Analysis:"
echo "   - File → Log aggregation and analysis"
echo "   - Historical data processing"
echo ""
echo "🔗 Pipeline Integration:"
echo "   - Stdio → Unix pipes and stream processing"
echo "   - Docker container logging"
echo ""

# Testing Tips
echo -e "${YELLOW}Testing Tips:${NC}"
echo ""
echo "1. Test stdio output:"
echo "   ./object_detection --enable-notifications --enable-stdio-notification 2>&1 | grep 'NEW OBJECT'"
echo ""
echo "2. Monitor file notifications:"
echo "   tail -f /tmp/notifications.json | jq '.object.type'"
echo ""
echo "3. Test SSE with curl:"
echo "   curl http://localhost:8081"
echo ""
echo "4. Test webhook with netcat:"
echo "   nc -l -p 9000"
echo ""

echo "======================================"
echo "For more details, see:"
echo "  - cpp-object-detection/NOTIFICATION_FEATURE.md"
echo "  - cpp-object-detection/README.md"
echo "======================================"
