import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { App } from '../App';
import { formatClock } from '../format';
import { LogsView } from './LogsView';
import { PreviewView } from './PreviewView';
import { StatusView } from './StatusView';
import { callsTo, json, mockServer, monitoring, schedule, settings, withMonitoring, withStorage, type Routes } from '../test/mockServer';

function renderStatus(routes: Routes = {}) {
  const f = mockServer(routes);
  render(<StatusView settings={settings} />);
  return { f, user: userEvent.setup() };
}

const clockButton = () => screen.findByRole('button', { name: "Set the camera's clock to this device's time" });

/** The top bar's "Next shot" pill, once it shows. */
const nextShotPill = async () => (await screen.findByText('Next shot')).closest('.next-pill')!;

describe('StatusView', () => {
  it('turns raw Pi readings into something readable', async () => {
    vi.useFakeTimers({ toFake: ['Date'] });
    vi.setSystemTime(new Date(2026, 10, 2, 11, 0, 0));
    renderStatus();
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
    renderStatus(withMonitoring({ UnixMillis: 0 }));
    expect(await screen.findByText(/clock thinks it is 1970/)).toBeInTheDocument();
  });

  it("sets the camera's clock to this device's time", async () => {
    const { f, user } = renderStatus(withMonitoring({ UnixMillis: Date.now() - 3 * 86400_000 }));
    expect(await screen.findByText(/behind this device/)).toBeInTheDocument();
    const before = Date.now();
    await user.click(await clockButton());
    expect(await screen.findByText(/Camera clock set\./)).toBeInTheDocument();
    const sent = JSON.parse(String(callsTo(f, 'POST', '/admin/time')[0]![1]?.body)).UnixMillis;
    expect(sent).toBeGreaterThanOrEqual(before);
    expect(sent).toBeLessThanOrEqual(Date.now());
  });

  it('says when the clock module could not be updated', async () => {
    const { user } = renderStatus({
      ...withMonitoring({ UnixMillis: 0 }),
      'POST /admin/time': () => json({ UnixMillis: Date.now(), RTC: 'not updated: exit status 1' }),
    });
    await user.click(await clockButton());
    expect(await screen.findByText(/clock module was not updated: exit status 1/)).toBeInTheDocument();
  });

  it('reports a failure to set the clock', async () => {
    const { user } = renderStatus({
      ...withMonitoring({ UnixMillis: 0 }),
      'POST /admin/time': () => json('setting the clock: operation not permitted', 500),
    });
    await user.click(await clockButton());
    expect(await screen.findByText(/The clock was not set. setting the clock: operation not permitted/)).toBeInTheDocument();
  });

  it('shows why the readings could not be loaded', async () => {
    renderStatus({ 'GET /monitoring': () => json('vcgencmd: timeout', 500) });
    expect(await screen.findByText('vcgencmd: timeout')).toBeInTheDocument();
  });

  it('accepts a camera clock up to a minute off', async () => {
    renderStatus(withMonitoring({ UnixMillis: Date.now() - 50_000 }));
    expect(await screen.findByText('48.3')).toBeInTheDocument();
    expect(screen.queryByText(/behind this device/)).not.toBeInTheDocument();
  });

  it('remembers under-voltage since the last restart when the power is fine now', async () => {
    renderStatus(withMonitoring({ Power: { Now: [], SinceBoot: ['under-voltage'] } }));
    const power = (await screen.findByRole('heading', { name: 'Power' })).parentElement!;
    expect(power).toHaveAttribute('data-tone', 'warm');
    expect(within(power).getByText('OK')).toBeInTheDocument();
    expect(within(power).getByText('Since the last restart: under-voltage. Check the power supply and cable.')).toBeInTheDocument();
  });

  it('warns about under-voltage', async () => {
    renderStatus(withMonitoring({ Power: { Now: ['under-voltage', 'throttled'], SinceBoot: ['under-voltage', 'throttled'] } }));
    const power = (await screen.findByRole('heading', { name: 'Power' })).parentElement!;
    expect(power).toHaveAttribute('data-tone', 'hot');
    expect(within(power).getByText('Under-voltage, throttled')).toBeInTheDocument();
    expect(within(power).getByText(/Check the power supply/)).toBeInTheDocument();
  });

  it('says when the card is full and the camera stopped', async () => {
    renderStatus(withStorage({ FreeBytes: 100 * 1024 ** 2, StopWhenDiskFull: true }));
    expect(await screen.findByText(/The card is full and the camera has stopped taking photos/)).toBeInTheDocument();
    expect(screen.getByText(/Full: the camera stops taking photos/)).toBeInTheDocument();
  });

  it('counts photos deleted to make room', async () => {
    renderStatus(withStorage({ DeletedOldest: 1234 }));
    expect(await screen.findByText(/1,234 old photos deleted to make room/)).toBeInTheDocument();
  });

  it('says when a sensor is missing instead of printing the error', async () => {
    const missing = 'fork/exec /opt/vc/bin/vcgencmd: no such file or directory';
    renderStatus(
      withMonitoring({
        GpuTemperature: missing,
        Power: { Now: [], SinceBoot: [], Error: missing },
        CPU: { ...monitoring.CPU, Error: 'open scaling_governor: no such file or directory' },
      }),
    );
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

  it('filters lines by the search text', async () => {
    mockServer();
    render(<LogsView />);
    const user = userEvent.setup();
    await user.type(await screen.findByRole('searchbox', { name: 'Filter log lines' }), 'CAPTURE');
    expect(screen.queryByText(/started/)).not.toBeInTheDocument();
    expect(screen.getByText(/Error during capture/)).toBeInTheDocument();
    await user.type(screen.getByRole('searchbox'), ' nothing like this');
    expect(screen.getByText('No lines match.')).toBeInTheDocument();
  });

  it('explains why logs may be missing', async () => {
    mockServer({ 'GET /logs': () => json('open timelapse.log: no such file or directory', 500) });
    render(<LogsView />);
    expect(await screen.findByText(/only keeps a log file when started with/)).toBeInTheDocument();
  });
});

describe('PreviewView', () => {
  it('requests a fresh capture each time', async () => {
    render(<PreviewView />);
    const user = userEvent.setup();
    expect(screen.getByText(/Taken at 640×480\./)).toBeInTheDocument();
    const img = screen.getByAltText('Camera preview');
    const first = img.getAttribute('src');
    expect(first).toMatch(/^\/capture\?t=\d+$/);
    expect(screen.getByRole('button', { name: 'Take preview' })).toBeDisabled();
    img.dispatchEvent(new Event('load'));
    await waitFor(() => expect(screen.getByRole('button', { name: 'Take preview' })).toBeEnabled());
    await user.click(screen.getByRole('button', { name: 'Take preview' }));
    expect(screen.getByAltText('Camera preview').getAttribute('src')).not.toBe(first);
  });

  it('says when the camera could not take a preview, and allows a retry', async () => {
    render(<PreviewView />);
    fireEvent.error(screen.getByAltText('Camera preview'));
    expect(await screen.findByText(/The camera couldn't take a preview/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Take preview' })).toBeEnabled();
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

  it("counts down to the camera's planned next shot, whatever this device's clock says", async () => {
    vi.useFakeTimers({ toFake: ['Date'] });
    vi.setSystemTime(new Date('2026-11-02T10:00:00Z'));
    // The camera's clock is an hour behind this device; its next shot is 72 s after its now.
    mockServer({ 'GET /schedule': () => schedule(72_000, { cameraNow: Date.now() - 3600_000 }) });
    render(<App />);
    const pill = await nextShotPill();
    await waitFor(() => expect(pill.querySelector('strong')).toHaveTextContent(/^1:12$/));
  });

  it('shows when photos resume after a skipped night, not a countdown', async () => {
    vi.useFakeTimers({ toFake: ['Date'] });
    vi.setSystemTime(new Date('2026-11-02T20:00:00Z'));
    mockServer({ 'GET /schedule': () => schedule(11 * 3600_000, { night: true }) });
    render(<App />);
    const pill = await nextShotPill();
    const resume = `at ${formatClock(new Date('2026-11-03T07:00:00Z'))}`;
    await waitFor(() => expect(pill.querySelector('strong')).toHaveTextContent(resume));
    expect(pill).toHaveAttribute('title', `Nights are skipped: photos resume ${resume}`);
  });

  it('shows why the settings could not be loaded', async () => {
    mockServer({ 'GET /configuration': () => json('reading settings: database not open', 500) });
    window.location.hash = '#/settings';
    render(<App />);
    expect(await screen.findByText('reading settings: database not open')).toBeInTheDocument();
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
