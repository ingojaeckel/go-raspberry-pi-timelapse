import React, { useState } from 'react';
import { BaseUrl } from '../conf/config';
import { Alert, Box, Typography } from '@mui/material';

export default function PhotosComponent() {
  const [imageError, setImageError] = useState<boolean>(false);
  
  const handleImageError = () => {
    setImageError(true);
  };

  return (
    <React.Fragment>
      {imageError && (
        <Alert severity="error" sx={{ mb: 2 }}>
          Failed to load camera preview. This may indicate an issue with the camera setup or connection. 
          Please verify that the camera is properly connected and configured.
        </Alert>
      )}
      <Box sx={{ mb: 2 }}>
        <img 
          src={BaseUrl+"/capture"} 
          alt="preview" 
          onError={handleImageError}
          style={{ maxWidth: '100%', height: 'auto' }}
        />
      </Box>
      <div>
          <Typography variant="h6" gutterBottom>
            Tips for fine-tuning the camera position, focus, etc:
          </Typography>
          <ul>
              <li>For a low-latency viewfinder, consider attaching a 2-3 inch TFT to your Pi. The small TFT can mirror the image normally shown on a HDMI-connected monitor. This can be used to show a live-feed of the Pi camera on the TFT. For more details checkout the tutorials <a href="https://learn.adafruit.com/adafruit-pitft-28-inch-resistive-touchscreen-display-raspberry-pi/easy-install-2">[1]</a>, <a href="https://learn.adafruit.com/running-opengl-based-games-and-emulators-on-adafruit-pitft-displays/tuning-performance">[2]</a>, <a href="https://willhaley.com/blog/power-off-raspberry-pi-adafruit-tft-screen-shutdown/">[3]</a>. Note this comes with additional power use and increases the setup complexity.</li>
              <li>For a medium-latency viewfinder, consider having the Pi stream its video signal via Wifi to another device on the same network. As an example, this can be achieved via <a href="https://www.videolan.org/">VLC</a>. On the receiving device (e.g. laptop) let VLC listen for the incoming video stream (e.g. on port :1234): <pre>vlc udp://@:1234 :demux=h264</pre>Run the following on the Pi to start streaming: <pre>raspivid -t 60000 -o udp://192.168.0.123:1234</pre>This is fairly resource intensive and requires a fairly good Wifi connection. However, it works ok even on a Pi Zero W.</li>
          </ul>
    </div>
    </React.Fragment>
  );
}
