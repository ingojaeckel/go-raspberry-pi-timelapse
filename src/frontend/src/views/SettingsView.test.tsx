import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SettingsView, validate } from './SettingsView';
import { callsTo, json, mockServer, monitoring, settings } from '../test/mockServer';
import type { Settings } from '../api';

function renderSettings(s: Settings = settings, onSaved = vi.fn()) {
  render(<SettingsView settings={s} error={null} onSaved={onSaved} />);
  return onSaved;
}

const posted = (f: ReturnType<typeof mockServer>) =>
  JSON.parse(String(callsTo(f, 'POST', '/configuration')[0]?.[1]?.body)) as Settings;

describe('SettingsView', () => {
  it('saves what happens when the card is full', async () => {
    const f = mockServer();
    const onSaved = renderSettings();
    const user = userEvent.setup();
    expect(screen.getByRole('radio', { name: /Delete oldest/ })).toBeChecked();
    await user.click(screen.getByRole('radio', { name: /Stop/ }));
    expect(screen.getByText(/stops taking photos when the card is full/)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Save changes' }));
    await waitFor(() => expect(onSaved).toHaveBeenCalled());
    expect(posted(f)).toEqual({ ...settings, StopWhenDiskFull: true });
  });

  it('shows the interval and the offset as a minute past the hour', () => {
    mockServer();
    renderSettings();
    expect(screen.getByRole('combobox', { name: 'Take a photo every' })).toHaveDisplayValue('30 minutes');
    // 900 seconds is minute 15, not "900 minutes" as the old UI showed
    expect(screen.getByLabelText('First capture at minute')).toHaveValue(15);
    expect(screen.getByText('Captures at :15, :45.')).toBeInTheDocument();
  });

  it('only offers intervals that divide an hour evenly', () => {
    mockServer();
    renderSettings();
    const options = within(screen.getByRole('combobox', { name: 'Take a photo every' }))
      .getAllByRole('option')
      .map((o) => Number((o as HTMLOptionElement).value));
    expect(options).toEqual([60, 72, 75, 80, 90, 100, 120, 144, 150, 180, 200, 225, 240, 300, 360, 400, 450, 600, 720, 900, 1200, 1800, 3600]);
  });

  it('keeps the offset within the new interval when the interval gets shorter', async () => {
    mockServer();
    renderSettings();
    const user = userEvent.setup();
    // minute 15 of every 30 minutes = :15 and :45; every 10 minutes from :05 keeps :15 and :45
    await user.selectOptions(screen.getByRole('combobox', { name: 'Take a photo every' }), '600');
    expect(screen.getByLabelText('First capture at minute')).toHaveValue(5);
  });

  it('saves the full settings object in the units the backend expects', async () => {
    const f = mockServer();
    const onSaved = renderSettings();
    const user = userEvent.setup();

    await user.selectOptions(screen.getByRole('combobox', { name: 'Take a photo every' }), '600');
    await user.clear(screen.getByLabelText('First capture at minute'));
    await user.type(screen.getByLabelText('First capture at minute'), '3');
    await user.click(await screen.findByRole('radio', { name: /1920×1080/ }));
    await user.click(screen.getByRole('radio', { name: /Upside down/ }));
    await user.click(screen.getByRole('button', { name: 'Save changes' }));

    await waitFor(() => expect(onSaved).toHaveBeenCalled());
    expect(posted(f)).toEqual({
      ...settings,
      SecondsBetweenCaptures: 600,
      OffsetWithinHour: 180,
      PhotoResolutionWidth: 1920,
      PhotoResolutionHeight: 1080,
      RotateBy: 180,
    });
  });

  it('blocks an offset outside the interval', async () => {
    mockServer();
    renderSettings();
    const user = userEvent.setup();
    await user.clear(screen.getByLabelText('First capture at minute'));
    await user.type(screen.getByLabelText('First capture at minute'), '30');
    expect(screen.getByText('Pick a minute from 0 to 29.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Save changes' })).toBeDisabled();
  });

  it('lets saved settings with an interval the camera now rejects be fixed', () => {
    mockServer();
    renderSettings({ ...settings, SecondsBetweenCaptures: 1000, OffsetWithinHour: 0 });
    expect(screen.getByText('Pick one of the listed intervals.')).toBeInTheDocument();
  });

  it('turns hour alignment off by sending -1', async () => {
    const f = mockServer();
    renderSettings();
    const user = userEvent.setup();
    await user.click(screen.getByRole('checkbox', { name: 'Line captures up with the clock' }));
    expect(screen.queryByLabelText('First capture at minute')).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Save changes' }));
    await waitFor(() => expect(callsTo(f, 'POST', '/configuration')).toHaveLength(1));
    expect(posted(f).OffsetWithinHour).toBe(-1);
  });

  it('blocks saving invalid values', async () => {
    mockServer();
    renderSettings();
    const user = userEvent.setup();
    await user.clear(screen.getByLabelText('JPEG quality'));
    await user.type(screen.getByLabelText('JPEG quality'), '0');
    expect(screen.getByText('Use a value from 1 to 100.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Save changes' })).toBeDisabled();
  });

  it('shows the server error when saving fails', async () => {
    mockServer({ 'POST /configuration': () => json('Quality is out of bounds', 400) });
    renderSettings();
    const user = userEvent.setup();
    await user.click(await screen.findByRole('radio', { name: /640×480/ }));
    await user.click(screen.getByRole('button', { name: 'Save changes' }));
    expect(await screen.findByText('Settings were not saved. Quality is out of bounds')).toBeInTheDocument();
  });

  it('warns when the camera did not apply the chosen resolution', async () => {
    // Safety net: if the server ever keeps a different size than requested, say so.
    mockServer({ 'POST /configuration': (_u, init) => json({ ...JSON.parse(String(init?.body)), PhotoResolutionWidth: 1640, PhotoResolutionHeight: 1232 }) });
    renderSettings();
    const user = userEvent.setup();
    await user.click(await screen.findByRole('radio', { name: /3280×2464/ }));
    await user.click(screen.getByRole('button', { name: 'Save changes' }));
    expect(await screen.findByText(/camera kept 1640×1232 instead of 3280×2464/)).toBeInTheDocument();
  });

  it("offers only the connected camera's resolutions", async () => {
    mockServer();
    renderSettings();
    expect(await screen.findByText(/imx219 camera/)).toBeInTheDocument();
    const group = screen.getByRole('radiogroup', { name: 'Resolution' });
    expect(within(group).getAllByRole('radio').map((r) => r.closest('label')?.textContent)).toEqual([
      '8.1 MP3280×2464 · 4:3',
      '2.1 MP1920×1080 · 16:9',
      '2 MP1640×1232 · 4:3',
      '0.3 MP640×480 · 4:3',
    ]);
    expect(within(group).getByRole('radio', { name: /1640×1232/ })).toBeChecked();
  });

  it('keeps a saved size the camera does not support, and says so', async () => {
    mockServer();
    renderSettings({ ...settings, PhotoResolutionWidth: 2186, PhotoResolutionHeight: 1642 });
    const saved = await screen.findByRole('radio', { name: /2186×1642/ });
    expect(saved).toBeChecked();
    expect(saved.closest('label')).toHaveTextContent('not supported');
    expect(screen.getByText(/does not support 2186×1642/)).toBeInTheDocument();
  });

  it('offers the saved size when the camera is not detected', async () => {
    mockServer({ 'GET /camera': () => json({ Model: '', Width: 0, Height: 0, Resolutions: null, Error: 'no camera detected' }) });
    renderSettings();
    expect(await screen.findByText(/couldn't detect the camera \(no camera detected\)/i)).toBeInTheDocument();
    const radios = within(screen.getByRole('radiogroup', { name: 'Resolution' })).getAllByRole('radio');
    expect(radios).toHaveLength(1);
    expect(radios[0]).toBeChecked();
  });

  it('estimates the photo size and how long the card lasts as quality changes', async () => {
    mockServer();
    renderSettings();
    const user = userEvent.setup();
    // 1640x1232 at quality 100 with 0.45 bytes per pixel: 0.9 MB, 23.75 GB usable, 48 photos a day
    const estimate = await screen.findByText(/per photo/);
    expect(estimate).toHaveTextContent(/About 888 KB per photo, estimated from the last 3 photos\./);
    expect(estimate).toHaveTextContent(/room for about 28,0\d\d more: \d+ days at 48 photos a day, then the oldest photos are deleted/);

    await user.clear(screen.getByLabelText('JPEG quality'));
    await user.type(screen.getByLabelText('JPEG quality'), '50');
    expect(screen.getByText(/per photo/)).toHaveTextContent(/About 115 KB per photo/);
  });

  it('says the estimate is rough before the first photo', async () => {
    mockServer({
      'GET /monitoring': () =>
        json({ ...monitoring, Storage: { ...monitoring.Storage, Photos: 0, PhotoBytes: 0, SamplePhotos: 0, SampleBytesPerPixel: 0, SampleQuality: 0 } }),
    });
    renderSettings();
    expect(await screen.findByText(/rough estimate until the camera has taken photos/)).toBeInTheDocument();
  });

  it('asks before shutting down and only then calls the backend', async () => {
    const f = mockServer();
    renderSettings();
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: 'Shut down' }));
    expect(callsTo(f, 'POST', '/admin/shutdown')).toHaveLength(0);
    expect(screen.getByRole('heading', { name: 'Shut down the camera?' })).toBeInTheDocument();
    const dialog = screen.getByRole('dialog');
    await user.click(within(dialog).getByRole('button', { name: 'Shut down' }));
    await waitFor(() => expect(callsTo(f, 'POST', '/admin/shutdown')).toHaveLength(1));
    expect(await screen.findByText(/Wait until the green light/)).toBeInTheDocument();
  });

  it('deletes all photos with DELETE /photos only after confirmation', async () => {
    const f = mockServer();
    renderSettings();
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: 'Delete all photos' }));
    await user.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(callsTo(f, 'DELETE', '/photos')).toHaveLength(0);

    await user.click(screen.getByRole('button', { name: 'Delete all photos' }));
    await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Delete all photos' }));
    await waitFor(() => expect(callsTo(f, 'DELETE', '/photos')).toHaveLength(1));
    expect(await screen.findByText('All photos deleted.')).toBeInTheDocument();
  });

  it('never uses GET for actions that change state', async () => {
    const f = mockServer();
    renderSettings();
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: 'Restart' }));
    await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Restart' }));
    await waitFor(() => expect(callsTo(f, 'POST', '/admin/restart')).toHaveLength(1));
    expect(f.mock.calls.filter(([u]) => String(u).includes('/admin/')).every(([, init]) => init?.method === 'POST')).toBe(true);
  });
});

describe('validate', () => {
  it('matches the backend validator', () => {
    expect(validate(settings)).toEqual({});
    expect(validate({ ...settings, OffsetWithinHour: -1 })).toEqual({});
    expect(Object.keys(validate({ ...settings, OffsetWithinHour: 1800, Quality: 101 }))).toEqual(['OffsetWithinHour', 'Quality']);
    expect(Object.keys(validate({ ...settings, SecondsBetweenCaptures: 1000 }))).toEqual(['SecondsBetweenCaptures']);
    expect(Object.keys(validate({ ...settings, SecondsBetweenCaptures: 7200 }))).toEqual(['SecondsBetweenCaptures']);
  });
});
