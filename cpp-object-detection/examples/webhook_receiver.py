#!/usr/bin/env python3
"""
Webhook Notification Receiver Example

This script demonstrates how to receive webhook notifications from the
C++ object detection application. It starts a simple HTTP server that
listens for POST requests containing detection notifications.

Usage:
    python3 webhook_receiver.py [port]

Example:
    # Start webhook receiver on port 9000
    python3 webhook_receiver.py 9000
    
    # In another terminal, start object detection with webhook
    ./object_detection \\
        --enable-notifications \\
        --enable-webhook \\
        --webhook-url http://localhost:9000/webhook
"""

import sys
import json
import base64
from datetime import datetime
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path

class WebhookHandler(BaseHTTPRequestHandler):
    """Handler for incoming webhook notifications"""
    
    def do_POST(self):
        """Handle POST requests from the object detection application"""
        
        # Read the request body
        content_length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(content_length)
        
        try:
            # Parse JSON payload
            data = json.loads(body.decode('utf-8'))
            
            # Extract notification details
            event = data.get('event', 'unknown')
            timestamp = data.get('timestamp', 'unknown')
            obj = data.get('object', {})
            status = data.get('status', {})
            all_detections = data.get('all_detections', [])
            top_objects = data.get('top_objects', [])
            
            # Print notification details
            print(f"\n{'='*60}")
            print(f"New Notification Received at {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
            print(f"{'='*60}")
            print(f"Event: {event}")
            print(f"Timestamp: {timestamp}")
            print(f"\nDetected Object:")
            print(f"  Type: {obj.get('type', 'unknown')}")
            print(f"  Position: ({obj.get('x', 0):.1f}, {obj.get('y', 0):.1f})")
            print(f"  Confidence: {obj.get('confidence', 0)*100:.1f}%")
            
            print(f"\nSystem Status:")
            print(f"  FPS: {status.get('fps', 0):.1f}")
            print(f"  Processing Time: {status.get('avg_processing_time_ms', 0):.1f}ms")
            print(f"  Total Objects Detected: {status.get('total_objects_detected', 0)}")
            print(f"  Total Images Saved: {status.get('total_images_saved', 0)}")
            print(f"  GPU Enabled: {status.get('gpu_enabled', False)}")
            print(f"  Burst Mode: {status.get('burst_mode_enabled', False)}")
            
            if all_detections:
                print(f"\nAll Detections in Frame ({len(all_detections)}):")
                for det in all_detections:
                    print(f"  - {det.get('class', 'unknown')} "
                          f"({det.get('confidence', 0)*100:.1f}% confidence)")
            
            if top_objects:
                print(f"\nTop Detected Objects:")
                for top in top_objects[:5]:  # Show top 5
                    print(f"  - {top.get('type', 'unknown')}: {top.get('count', 0)} times")
            
            # Save image if included
            if 'image' in data and data['image']:
                self.save_image(data['image'], obj.get('type', 'unknown'), timestamp)
            
            # Send success response
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            response = {'status': 'success', 'message': 'Notification received'}
            self.wfile.write(json.dumps(response).encode('utf-8'))
            
        except Exception as e:
            print(f"Error processing notification: {e}")
            self.send_response(500)
            self.end_headers()
    
    def save_image(self, base64_image, object_type, timestamp):
        """Save the base64-encoded image to disk"""
        try:
            # Create output directory
            output_dir = Path('webhook_images')
            output_dir.mkdir(exist_ok=True)
            
            # Generate filename
            safe_timestamp = timestamp.replace(':', '-').replace(' ', '_')
            filename = f"{safe_timestamp}_{object_type}.jpg"
            filepath = output_dir / filename
            
            # Decode and save image
            image_data = base64.b64decode(base64_image)
            with open(filepath, 'wb') as f:
                f.write(image_data)
            
            print(f"\nImage saved: {filepath}")
            
        except Exception as e:
            print(f"Error saving image: {e}")
    
    def log_message(self, format, *args):
        """Override to suppress default HTTP logging"""
        pass

def main():
    """Main function to start the webhook server"""
    
    # Get port from command line or use default
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 9000
    
    # Create and start server
    server_address = ('', port)
    httpd = HTTPServer(server_address, WebhookHandler)
    
    print(f"Webhook Notification Receiver")
    print(f"{'='*60}")
    print(f"Listening on: http://localhost:{port}")
    print(f"Webhook URL: http://localhost:{port}/webhook")
    print(f"\nTo use with object detection, run:")
    print(f"  ./object_detection \\")
    print(f"    --enable-notifications \\")
    print(f"    --enable-webhook \\")
    print(f"    --webhook-url http://localhost:{port}/webhook")
    print(f"\nImages will be saved to: ./webhook_images/")
    print(f"{'='*60}")
    print(f"\nWaiting for notifications... (Press Ctrl+C to stop)")
    
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\n\nShutting down webhook receiver...")
        httpd.shutdown()

if __name__ == '__main__':
    main()
