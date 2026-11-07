import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { App } from '../App';
import { LogsView } from './LogsView';
import { PreviewView } from './PreviewView';
import { StatusView } from './StatusView';
import { callsTo, json, mockServer, monitoring, settings } from '../test/mockServer';

describe('StatusView', () => {
  it('turns raw Pi readings into something readable', async () => {
    vi.useFakeTimers({ toFake: ['Date'] });
    vi.setSystemTime(new Date(2026, 10, 2, 11, 0, 0));
    mockServer();
    render(<StatusView settings={settings} />);
    expect(await screen.findByText('48.3')).toBeInTheDocument();
    expect(screen.getByText('51.0')).toBeInTheDocument();
    expect(screen.getAllByText('24 GB').length).toBeGreaterThan(0);
    expect(screen.getByText('3 days, 2:01')).toBeInTheDocument();
    expect(screen.getByRole('meter')).toHaveAttribute('value', '15');
    // (24 GB free - 256 MB kept free) / ~2.33 MB average photo, at 48 photos a day
    expect(
      screen.getByText(/Room for about 10,422 more photos \(217 days at the current schedule\), then the oldest photos are deleted/),
    ).toBeInTheDocument();
    expect(screen.getByText(/Runs at 600 MHz \(powersave, 600–1000 MHz\)/)).toBeInTheDocument();
    expect(screen.getByText('No under-voltage or throttling since the last restart.')).toBeInTheDocument();
    expect(screen.queryByText(/clock/i)).not.toBeInTheDocument();
  });

  it('counts only daylight photos in the runway when nights are skipped, like Settings', async () => {
    vi.useFakeTimers({ toFake: ['Date'] });
    vi.setSystemTime(new Date('2026-11-02T11:00:00Z'));
    mockServer();
    render(<StatusView settings={{ ...settings, SkipNightPhotos: true, Latitude: 51.5074, Longitude: -0.1278 }} />);
    const text = await screen.findByText(/Room for about 10,422 more photos/);
    const days = Number(text.textContent!.match(/\((\d+) days at the current schedule, nights skipped\)/)?.[1]);
    expect(days).toBeGreaterThan(217 * 2); // about 9 of 24 hours of daylight in London in November
    vi.useRealTimers();
  });

  it('warns when the camera clock is wrong', async () => {
    mockServer({ 'GET /monitoring': () => json({ ...monitoring, UnixMillis: 0 }) });
    render(<StatusView settings={settings} />);
    expect(await screen.findByText(/clock thinks it is 1970/)).toBeInTheDocument();
  });

  it("sets the camera's clock to this device's time", async () => {
    const f = mockServer({ 'GET /monitoring': () => json({ ...monitoring, UnixMillis: Date.now() - 3 * 86400_000 }) });
    render(<StatusView settings={settings} />);
    const user = userEvent.setup();
    expect(await screen.findByText(/behind this device/)).toBeInTheDocument();
    const before = Date.now();
    await user.click(screen.getByRole('button', { name: "Set the camera's clock to this device's time" }));
    expect(await screen.findByText(/Camera clock set\./)).toBeInTheDocument();
    const sent = JSON.parse(String(callsTo(f, 'POST', '/admin/time')[0]![1]?.body)).UnixMillis;
    expect(sent).toBeGreaterThanOrEqual(before);
    expect(sent).toBeLessThanOrEqual(Date.now());
  });

  it('says when the clock module could not be updated', async () => {
    mockServer({
      'GET /monitoring': () => json({ ...monitoring, UnixMillis: 0 }),
      'POST /admin/time': () => json({ UnixMillis: Date.now(), RTC: 'not updated: exit status 1' }),
    });
    render(<StatusView settings={settings} />);
    const user = userEvent.setup();
    await user.click(await screen.findByRole('button', { name: "Set the camera's clock to this device's time" }));
    expect(await screen.findByText(/clock module was not updated: exit status 1/)).toBeInTheDocument();
  });

  it('reports a failure to set the clock', async () => {
    mockServer({
      'GET /monitoring': () => json({ ...monitoring, UnixMillis: 0 }),
      'POST /admin/time': () => json('setting the clock: operation not permitted', 500),
    });
    render(<StatusView settings={settings} />);
    const user = userEvent.setup();
    await user.click(await screen.findByRole('button', { name: "Set the camera's clock to this device's time" }));
    expect(await screen.findByText(/The clock was not set. setting the clock: operation not permitted/)).toBeInTheDocument();
  });

  it('warns about under-voltage', async () => {
    mockServer({
      'GET /monitoring': () =>
        json({ ...monitoring, Power: { Now: ['under-voltage', 'throttled'], SinceBoot: ['under-voltage', 'throttled'] } }),
    });
    render(<StatusView settings={settings} />);
    const power = (await screen.findByRole('heading', { name: 'Power' })).parentElement!;
    expect(power).toHaveAttribute('data-tone', 'hot');
    expect(within(power).getByText('Under-voltage, throttled')).toBeInTheDocument();
    expect(within(power).getByText(/Check the power supply/)).toBeInTheDocument();
  });

  it('says when the card is full and the camera stopped', async () => {
    mockServer({
      'GET /monitoring': () =>
        json({
          ...monitoring,
          Storage: { ...monitoring.Storage, FreeBytes: 100 * 1024 ** 2, StopWhenDiskFull: true },
        }),
    });
    render(<StatusView settings={settings} />);
    expect(await screen.findByText(/The card is full and the camera has stopped taking photos/)).toBeInTheDocument();
    expect(screen.getByText(/Full: the camera stops taking photos/)).toBeInTheDocument();
  });

  it('counts photos deleted to make room', async () => {
    mockServer({ 'GET /monitoring': () => json({ ...monitoring, Storage: { ...monitoring.Storage, DeletedOldest: 1234 } }) });
    render(<StatusView settings={settings} />);
    expect(await screen.findByText(/1,234 old photos deleted to make room/)).toBeInTheDocument();
  });

  it('says when a sensor is missing instead of printing the error', async () => {
    mockServer({
      'GET /monitoring': () =>
        json({
          ...monitoring,
          GpuTemperature: 'fork/exec /opt/vc/bin/vcgencmd: no such file or directory',
          Power: { Now: [], SinceBoot: [], Error: 'fork/exec /opt/vc/bin/vcgencmd: no such file or directory' },
          CPU: { ...monitoring.CPU, Error: 'open scaling_governor: no such file or directory' },
        }),
    });
    render(<StatusView settings={settings} />);
    expect(await screen.findAllByText('Not available')).toHaveLength(2);
    expect(screen.queryByText(/MHz/)).not.toBeInTheDocument();
  });
});

describe('LogsView', () => {
  it('shows log lines and filters to problems', async () => {
    mockServer();
    render(<LogsView />);
    const user = userEvent.setup();
    expect(await screen.findByText(/started/)).toBeInTheDocument();
    await user.click(screen.getByRole('checkbox', { name: 'Problems only' }));
    expect(screen.queryByText(/started/)).not.toBeInTheDocument();
    expect(screen.getByText(/Error during capture/)).toBeInTheDocument();
  });

  it('explains why logs may be missing', async () => {
    mockServer({ 'GET /logs': () => json('open timelapse.log: no such file or directory', 500) });
    render(<LogsView />);
    expect(await screen.findByText(/only keeps a log file when started with/)).toBeInTheDocument();
  });
});

describe('PreviewView', () => {
  it('requests a fresh capture each time', async () => {
    render(<PreviewView settings={settings} />);
    const user = userEvent.setup();
    const img = screen.getByAltText('Camera preview');
    const first = img.getAttribute('src');
    expect(first).toMatch(/^\/capture\?t=\d+$/);
    expect(screen.getByRole('button', { name: 'Take preview' })).toBeDisabled();
    img.dispatchEvent(new Event('load'));
    await waitFor(() => expect(screen.getByRole('button', { name: 'Take preview' })).toBeEnabled());
    await user.click(screen.getByRole('button', { name: 'Take preview' }));
    expect(screen.getByAltText('Camera preview').getAttribute('src')).not.toBe(first);
  });
});

describe('App', () => {
  it('navigates between sections with the hash', async () => {
    mockServer();
    render(<App />);
    const user = userEvent.setup();
    expect(await screen.findByRole('heading', { name: 'Photos' })).toBeInTheDocument();
    await user.click(screen.getByRole('link', { name: 'Settings' }));
    expect(await screen.findByRole('heading', { name: 'Settings' })).toBeInTheDocument();
    expect(window.location.hash).toBe('#/settings');
    expect(screen.getByRole('link', { name: 'Settings' })).toHaveAttribute('aria-current', 'page');
  });

  it('shows the countdown to the next capture', async () => {
    mockServer();
    render(<App />);
    expect(await screen.findByText('Next shot')).toBeInTheDocument();
  });

  it("counts down to the camera's planned next shot, whatever this device's clock says", async () => {
    // The camera's clock is an hour behind this device; its next shot is 72 s after its now.
    const cameraNow = Date.now() - 3600_000;
    mockServer({ 'GET /schedule': () => json({ UnixMillis: cameraNow, NextMillis: cameraNow + 72_000, Night: false, ZoneOffsetSeconds: 0 }) });
    render(<App />);
    const pill = (await screen.findByText('Next shot')).closest('.next-pill')!;
    await waitFor(() => expect(pill.querySelector('strong')?.textContent).toMatch(/^1:1\d$/));
  });

  it('shows when photos resume after a skipped night, not a countdown', async () => {
    const cameraNow = Date.now();
    mockServer({ 'GET /schedule': () => json({ UnixMillis: cameraNow, NextMillis: cameraNow + 11 * 3600_000, Night: true, ZoneOffsetSeconds: 0 }) });
    render(<App />);
    const pill = (await screen.findByText('Next shot')).closest('.next-pill')!;
    await waitFor(() => expect(pill.querySelector('strong')?.textContent).toMatch(/^at \d\d:\d\d$/));
    expect(pill).toHaveAttribute('title', expect.stringMatching(/^Nights are skipped: photos resume at \d\d:\d\d$/));
  });

  it('counts down for captures every N minutes from start-up too', async () => {
    const cameraNow = Date.now();
    mockServer({
      'GET /configuration': () => json({ ...settings, OffsetWithinHour: -1 }),
      'GET /schedule': () => json({ UnixMillis: cameraNow, NextMillis: cameraNow + 300_000, Night: false, ZoneOffsetSeconds: 0 }),
    });
    render(<App />);
    const pill = (await screen.findByText('Next shot')).closest('.next-pill')!;
    await waitFor(() => expect(pill.querySelector('strong')?.textContent).toMatch(/^[45]:\d\d$/));
  });

  it('asks the camera again after saving settings', async () => {
    const f = mockServer();
    render(<App />);
    await screen.findByText('Next shot');
    const before = callsTo(f, 'GET', '/schedule').length;
    const user = userEvent.setup();
    await user.click(screen.getByRole('link', { name: 'Settings' }));
    const quality = await screen.findByLabelText('JPEG quality');
    await user.clear(quality);
    await user.type(quality, '80');
    await user.click(screen.getByRole('button', { name: 'Save changes' }));
    await waitFor(() => expect(callsTo(f, 'POST', '/configuration')).toHaveLength(1));
    await waitFor(() => expect(callsTo(f, 'GET', '/schedule').length).toBeGreaterThan(before));
  });
});
