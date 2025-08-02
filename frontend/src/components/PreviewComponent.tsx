import React, { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import { BaseUrl } from '../conf/config';
import { DetectionResponse, SettingsResponse, ObjectDetail } from '../models/response';

export default function PhotosComponent() {
  const [detection, setDetection] = useState<DetectionResponse | null>(null);
  const [settings, setSettings] = useState<SettingsResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [imageLoaded, setImageLoaded] = useState(false);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const imgRef = useRef<HTMLImageElement>(null);

  // Color mapping for different object categories
  const categoryColors: { [key: string]: string } = {
    'human': '#FF0000',      // Red
    'animal': '#00FF00',     // Green  
    'vehicle': '#0000FF',    // Blue
    'machinery': '#FF8000',  // Orange
    'bird': '#FF00FF',       // Magenta
    'cat': '#00FFFF',        // Cyan
    'dog': '#FFFF00',        // Yellow
    'car': '#800080',        // Purple
    'truck': '#FFA500',      // Dark Orange
    'person': '#FF4500',     // Red Orange
    'default': '#FFFFFF'     // White
  };

  useEffect(() => {
    // Load settings to check if object detection is enabled
    axios.get<SettingsResponse>(BaseUrl + "/configuration")
      .then(resp => {
        setSettings(resp.data);
        if (resp.data.ObjectDetectionEnabled) {
          loadDetectionResults();
        }
      })
      .catch(err => console.log("Error loading settings:", err));
  }, []);

  useEffect(() => {
    // Draw bounding boxes when detection data or image changes
    if (imageLoaded && detection?.Details) {
      drawBoundingBoxes();
    }
  }, [imageLoaded, detection]);

  const loadDetectionResults = () => {
    setLoading(true);
    axios.get<DetectionResponse>(BaseUrl + "/detection")
      .then(resp => {
        setDetection(resp.data);
        setLoading(false);
      })
      .catch(err => {
        console.log("Error loading detection results:", err);
        setLoading(false);
      });
  };

  const handleRefreshDetection = () => {
    if (settings?.ObjectDetectionEnabled) {
      loadDetectionResults();
    }
  };

  const handleImageLoad = () => {
    setImageLoaded(true);
  };

  const getColorForCategory = (detail: ObjectDetail): string => {
    // First try to match specific class
    if (categoryColors[detail.class]) {
      return categoryColors[detail.class];
    }
    // Then try to match category
    if (categoryColors[detail.category]) {
      return categoryColors[detail.category];
    }
    // Default color
    return categoryColors.default;
  };

  const drawBoundingBoxes = () => {
    const canvas = canvasRef.current;
    const img = imgRef.current;
    
    if (!canvas || !img || !detection?.Details) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Set canvas size to match image
    canvas.width = img.naturalWidth;
    canvas.height = img.naturalHeight;

    // Clear canvas
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    // Draw bounding boxes
    detection.Details.forEach((detail, index) => {
      if (detail.bbox) {
        const color = getColorForCategory(detail);
        
        // Set line properties
        ctx.strokeStyle = color;
        ctx.lineWidth = 3;
        ctx.font = '16px Arial';
        ctx.fillStyle = color;

        // Draw bounding box
        ctx.strokeRect(detail.bbox.x, detail.bbox.y, detail.bbox.width, detail.bbox.height);

        // Draw label background
        const label = `${detail.class} (${(detail.confidence * 100).toFixed(1)}%)`;
        const textMetrics = ctx.measureText(label);
        const textHeight = 20;
        
        ctx.fillStyle = color;
        ctx.globalAlpha = 0.7;
        ctx.fillRect(detail.bbox.x, detail.bbox.y - textHeight, textMetrics.width + 8, textHeight);
        
        // Draw label text
        ctx.globalAlpha = 1.0;
        ctx.fillStyle = '#FFFFFF';
        ctx.fillText(label, detail.bbox.x + 4, detail.bbox.y - 5);
      }
    });
  };

  return (
    <React.Fragment>
      <div style={{ position: 'relative', display: 'inline-block' }}>
        <img 
          ref={imgRef}
          src={BaseUrl+"/capture"} 
          alt="preview" 
          onLoad={handleImageLoad}
          style={{ display: 'block' }}
        />
        {settings?.ObjectDetectionEnabled && imageLoaded && detection?.Details && (
          <canvas
            ref={canvasRef}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              pointerEvents: 'none',
              width: '100%',
              height: '100%'
            }}
          />
        )}
      </div>
      
      {settings?.ObjectDetectionEnabled && (
        <div style={{ margin: '10px 0', padding: '10px', border: '1px solid #ccc', borderRadius: '5px' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <strong>Object Detection Results:</strong>
            <button onClick={handleRefreshDetection} disabled={loading}>
              {loading ? 'Loading...' : 'Refresh'}
            </button>
          </div>
          {detection ? (
            <div style={{ marginTop: '10px' }}>
              <p><strong>Summary:</strong> {detection.Summary}</p>
              {detection.Objects && detection.Objects.length > 0 && (
                <p><strong>Detected objects:</strong> {detection.Objects.join(', ')}</p>
              )}
              <div style={{ display: 'flex', gap: '20px', marginTop: '8px', fontSize: '0.9em', color: '#666' }}>
                <span><strong>Detection time:</strong> {detection.LatencyMs}ms</span>
                <span><strong>Confidence:</strong> {(detection.OverallConfidence * 100).toFixed(1)}%</span>
              </div>
              
              {/* Display color legend */}
              {detection.Details && detection.Details.length > 0 && (
                <div style={{ marginTop: '10px' }}>
                  <strong>Detection Legend:</strong>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '10px', marginTop: '5px' }}>
                    {Array.from(new Set(detection.Details.map(d => d.category))).map(category => (
                      <div key={category} style={{ display: 'flex', alignItems: 'center', fontSize: '0.9em' }}>
                        <div 
                          style={{ 
                            width: '12px', 
                            height: '12px', 
                            backgroundColor: categoryColors[category] || categoryColors.default,
                            marginRight: '5px',
                            border: '1px solid #333'
                          }}
                        />
                        {category}
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ) : (
            <p>No detection results available. Click "Refresh" to analyze the most recent photo.</p>
          )}
        </div>
      )}

      <div>
          Tips for fine-tuning the camera position, focus, etc:
          <ul>
              <li>For a low-latency viewfinder, consider attaching a 2-3 inch TFT to your Pi. The small TFT can mirror the image normally shown on a HDMI-connected monitor. This can be used to show a live-feed of the Pi camera on the TFT. For more details checkout the tutorials <a href="https://learn.adafruit.com/adafruit-pitft-28-inch-resistive-touchscreen-display-raspberry-pi/easy-install-2">[1]</a>, <a href="https://learn.adafruit.com/running-opengl-based-games-and-emulators-on-adafruit-pitft-displays/tuning-performance">[2]</a>, <a href="https://willhaley.com/blog/power-off-raspberry-pi-adafruit-tft-screen-shutdown/">[3]</a>. Note this comes with additional power use and increases the setup complexity.</li>
              <li>For a medium-latency viewfinder, consider having the Pi stream its video signal via Wifi to another device on the same network. As an example, this can be achieved via <a href="https://www.videolan.org/">VLC</a>. On the receiving device (e.g. laptop) let VLC listen for the incoming video stream (e.g. on port :1234): <pre>vlc udp://@:1234 :demux=h264</pre>Run the following on the Pi to start streaming: <pre>raspivid -t 60000 -o udp://192.168.0.123:1234</pre>This is fairly resource intensive and requires a fairly good Wifi connection. However, it works ok even on a Pi Zero W.</li>
          </ul>
    </div>
    </React.Fragment>
  );
}
