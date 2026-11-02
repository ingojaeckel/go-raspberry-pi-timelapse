import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { App } from '../App';
import { LogsView } from './LogsView';
import { PreviewView } from './PreviewView';
import { StatusView } from './StatusView';
import { json, mockServer, monitoring, settings } from '../test/mockServer';

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
    // 24 GB / ~2.33 MB average photo, at 48 photos a day
    expect(screen.getByText(/Room for about 10,532 more photos \(219 days/)).toBeInTheDocument();
  });

  it('warns when the camera clock is wrong', async () => {
    mockServer({ 'GET /monitoring': () => json({ ...monitoring, Time: 'Thu Jan  1 00:00:00 UTC 1970' }) });
    render(<StatusView settings={settings} />);
    expect(await screen.findByText(/clock thinks it is 1970/)).toBeInTheDocument();
  });

  it('says when a sensor is missing instead of printing the error', async () => {
    mockServer({
      'GET /monitoring': () =>
        json({ ...monitoring, GpuTemperature: 'fork/exec /opt/vc/bin/vcgencmd: no such file or directory' }),
    });
    render(<StatusView settings={settings} />);
    expect(await screen.findByText('Not available')).toBeInTheDocument();
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
});
