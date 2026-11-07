import { useEffect, useState } from 'react';
import { urls } from '../api';
import { Icon } from '../components/Icon';
import { Notice, PageHeader } from '../components/ui';

/** The size of preview photos: timelapse.PreviewWidth x PreviewHeight on the camera. */
const PREVIEW_SIZE = '640×480';

export function PreviewView() {
  const [shot, setShot] = useState(() => Date.now());
  const [state, setState] = useState<'loading' | 'ready' | 'error'>('loading');
  const [elapsed, setElapsed] = useState(0);

  useEffect(() => {
    if (state !== 'loading') return;
    const started = Date.now();
    setElapsed(0);
    const id = setInterval(() => setElapsed(Math.round((Date.now() - started) / 1000)), 500);
    return () => clearInterval(id);
  }, [state, shot]);

  const retake = () => {
    setState('loading');
    setShot((prev) => Math.max(Date.now(), prev + 1));
  };

  return (
    <>
      <PageHeader title="Preview">
        <button type="button" className="btn primary" onClick={retake} disabled={state === 'loading'}>
          <Icon name="camera" /> Take preview
        </button>
      </PageHeader>

      <p className="lede">
        A quick, low-resolution shot to check framing and focus. It is not saved with your timelapse. Taken at{' '}
        {PREVIEW_SIZE}.
      </p>

      {state === 'error' && (
        <Notice tone="error">
          The camera couldn't take a preview. Check that the ribbon cable is seated firmly at both ends, then try
          again. If a scheduled capture was running at the same moment, simply retry.
        </Notice>
      )}

      <div className="preview-frame" data-state={state}>
        {state === 'loading' && (
          <div className="preview-wait" role="status">
            <span className="spinner-ring" aria-hidden="true" />
            Capturing{elapsed > 1 ? `… ${elapsed} s` : '…'}
          </div>
        )}
        <img
          key={shot}
          src={urls.preview(shot)}
          alt="Camera preview"
          onLoad={() => setState('ready')}
          onError={() => setState('error')}
        />
      </div>

      <details className="tips">
        <summary>Getting a live viewfinder</summary>
        <div className="tips-body">
          <p>Previews take a few seconds each. For framing in real time you have two options.</p>
          <h3>A small screen on the Pi</h3>
          <p>
            A 2–3 inch TFT can mirror the Pi's display output and show a live camera feed. It adds power draw and
            setup work. See Adafruit's guides for the{' '}
            <a href="https://learn.adafruit.com/adafruit-pitft-28-inch-resistive-touchscreen-display-raspberry-pi/easy-install-2">
              PiTFT install
            </a>{' '}
            and{' '}
            <a href="https://learn.adafruit.com/running-opengl-based-games-and-emulators-on-adafruit-pitft-displays/tuning-performance">
              performance tuning
            </a>
            , and <a href="https://willhaley.com/blog/power-off-raspberry-pi-adafruit-tft-screen-shutdown/">turning the screen off</a>.
          </p>
          <h3>Stream to a laptop over WiFi</h3>
          <p>
            The stream holds the camera, so scheduled captures are skipped while it runs. On the laptop, have{' '}
            <a href="https://www.videolan.org/">VLC</a> listen for the stream:
          </p>
          <pre>vlc udp://@:1234 :demux=h264</pre>
          <p>Then start streaming on the Pi, using your laptop's IP address:</p>
          <pre>rpicam-vid -t 60000 --inline -o udp://192.168.0.123:1234</pre>
          <p>This works even on a Pi Zero W, given a reasonable WiFi signal.</p>
          <p className="muted">These links open once your device is back on a network with internet access.</p>
        </div>
      </details>
    </>
  );
}
