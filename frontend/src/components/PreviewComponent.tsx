import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { BaseUrl } from '../conf/config';
import { DetectionResponse, SettingsResponse } from '../models/response';

export default function PhotosComponent() {
  const [detection, setDetection] = useState<DetectionResponse | null>(null);
  const [settings, setSettings] = useState<SettingsResponse | null>(null);
  const [loading, setLoading] = useState(false);

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

  return (
    <React.Fragment>
      <div><img src={BaseUrl+"/capture"} alt="preview" /></div>
      
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
