import { act, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SettingsView, validate } from './SettingsView';
import { callsTo, json, mockServer, settings, withStorage, type Routes } from '../test/mockServer';
import type { Settings } from '../api';

/** Renders Settings and lets the camera and storage requests finish. */
async function setup(s: Settings = settings, routes: Routes = {}) {
  const f = mockServer(routes);
  const onSaved = vi.fn();
  render(<SettingsView settings={s} error={null} onSaved={onSaved} />);
  await act(() => new Promise((resolve) => setTimeout(resolve, 0)));
  return { f, onSaved, user: userEvent.setup() };
}

const saveButton = () => screen.getByRole('button', { name: 'Save changes' });

async function retype(user: ReturnType<typeof userEvent.setup>, label: string, text: string) {
  await user.clear(screen.getByLabelText(label));
  await user.type(screen.getByLabelText(label), text);
}

const posted = (f: ReturnType<typeof mockServer>) =>
  JSON.parse(String(callsTo(f, 'POST', '/configuration')[0]?.[1]?.body)) as Settings;

describe('SettingsView', () => {
  it('saves what happens when the card is full', async () => {
    const { f, onSaved, user } = await setup();
    expect(screen.getByRole('radio', { name: /Delete oldest/ })).toBeChecked();
    await user.click(screen.getByRole('radio', { name: /Stop/ }));
    expect(screen.getByText(/stops taking photos when the card is full/)).toBeInTheDocument();
    await user.click(saveButton());
    await waitFor(() => expect(onSaved).toHaveBeenCalled());
    expect(posted(f)).toEqual({ ...settings, StopWhenDiskFull: true });
  });

  it('shows the interval and the offset as a minute past the hour', async () => {
    await setup();
    expect(screen.getByRole('combobox', { name: 'Take a photo every' })).toHaveDisplayValue('30 minutes');
    // 900 seconds is minute 15, not "900 minutes" as the old UI showed
    expect(screen.getByLabelText('First capture at minute')).toHaveValue(15);
    expect(screen.getByText('Captures at :15, :45.')).toBeInTheDocument();
  });

  it('only offers intervals that divide an hour evenly', async () => {
    await setup();
    const options = within(screen.getByRole('combobox', { name: 'Take a photo every' }))
      .getAllByRole('option')
      .map((o) => Number((o as HTMLOptionElement).value));
    expect(options).toEqual([60, 72, 75, 80, 90, 100, 120, 144, 150, 180, 200, 225, 240, 300, 360, 400, 450, 600, 720, 900, 1200, 1800, 3600]);
  });

  it('keeps the offset within the new interval when the interval gets shorter', async () => {
    const { user } = await setup();
    // minute 15 of every 30 minutes = :15 and :45; every 10 minutes from :05 keeps :15 and :45
    await user.selectOptions(screen.getByRole('combobox', { name: 'Take a photo every' }), '600');
    expect(screen.getByLabelText('First capture at minute')).toHaveValue(5);
  });

  it('saves the full settings object in the units the backend expects', async () => {
    const { f, onSaved, user } = await setup();

    await user.selectOptions(screen.getByRole('combobox', { name: 'Take a photo every' }), '600');
    await retype(user, 'First capture at minute', '3');
    await user.click(await screen.findByRole('radio', { name: /1920×1080/ }));
    await user.click(screen.getByRole('radio', { name: /Upside down/ }));
    await user.click(saveButton());

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
    const { user } = await setup();
    await retype(user, 'First capture at minute', '30');
    expect(screen.getByText('Pick a minute from 0 to 29.')).toBeInTheDocument();
    expect(saveButton()).toBeDisabled();
  });

  it('lets saved settings with an interval the camera now rejects be fixed', async () => {
    await setup({ ...settings, SecondsBetweenCaptures: 1000, OffsetWithinHour: 0 });
    expect(screen.getByText('Pick one of the listed intervals.')).toBeInTheDocument();
  });

  it('turns hour alignment off by sending -1', async () => {
    const { f, user } = await setup();
    await user.click(screen.getByRole('checkbox', { name: 'Line captures up with the clock' }));
    expect(screen.queryByLabelText('First capture at minute')).not.toBeInTheDocument();
    await user.click(saveButton());
    await waitFor(() => expect(callsTo(f, 'POST', '/configuration')).toHaveLength(1));
    expect(posted(f).OffsetWithinHour).toBe(-1);
  });

  it('blocks saving invalid values', async () => {
    const { user } = await setup();
    await retype(user, 'JPEG quality', '0');
    expect(screen.getByText('Use a value from 1 to 100.')).toBeInTheDocument();
    expect(saveButton()).toBeDisabled();
  });

  it('shows the server error when saving fails', async () => {
    const { user } = await setup(settings, { 'POST /configuration': () => json('Quality is out of bounds', 400) });
    await user.click(await screen.findByRole('radio', { name: /640×480/ }));
    await user.click(saveButton());
    expect(await screen.findByText('Settings were not saved. Quality is out of bounds')).toBeInTheDocument();
  });

  it('warns when the camera did not apply the chosen resolution', async () => {
    // Safety net: if the server ever keeps a different size than requested, say so.
    const { user } = await setup(settings, { 'POST /configuration': (_u, init) => json({ ...JSON.parse(String(init?.body)), PhotoResolutionWidth: 1640, PhotoResolutionHeight: 1232 }) });
    await user.click(await screen.findByRole('radio', { name: /3280×2464/ }));
    await user.click(saveButton());
    expect(await screen.findByText(/camera kept 1640×1232 instead of 3280×2464/)).toBeInTheDocument();
  });

  it("offers only the connected camera's resolutions", async () => {
    await setup();
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

  it('keeps a saved size that is not one of the camera sizes, and says it is scaled', async () => {
    await setup({ ...settings, PhotoResolutionWidth: 2186, PhotoResolutionHeight: 1642 });
    const saved = await screen.findByRole('radio', { name: /2186×1642/ });
    expect(saved).toBeChecked();
    expect(saved.closest('label')).toHaveTextContent('2186×1642 · scaled');
    expect(screen.getByText(/2186×1642 isn't one of the camera's own sizes, so photos are scaled to it\./)).toBeInTheDocument();
  });

  it('warns that a saved size larger than the sensor enlarges photos', async () => {
    const imx290 = { Model: 'imx290', Width: 1920, Height: 1080, Resolutions: [{ Width: 1920, Height: 1080 }, { Width: 1280, Height: 720 }] };
    await setup({ ...settings, PhotoResolutionWidth: 3280, PhotoResolutionHeight: 2464 }, { 'GET /camera': () => json(imx290) });
    const saved = await screen.findByRole('radio', { name: /3280×2464/ });
    expect(saved.closest('label')).toHaveTextContent('3280×2464 · too large');
    expect(
      screen.getByText(
        '3280×2464 is larger than the camera can take (1920×1080), so photos are enlarged: bigger files, no extra detail. Choose 1920×1080.',
      ),
    ).toBeInTheDocument();
  });

  it('offers the saved size when the camera is not detected', async () => {
    await setup(settings, { 'GET /camera': () => json({ Model: '', Width: 0, Height: 0, Resolutions: null, Error: 'no camera detected' }) });
    expect(await screen.findByText(/couldn't detect the camera \(no camera detected\)/i)).toBeInTheDocument();
    const radios = within(screen.getByRole('radiogroup', { name: 'Resolution' })).getAllByRole('radio');
    expect(radios).toHaveLength(1);
    expect(radios[0]).toBeChecked();
  });

  it('estimates the photo size and how long the card lasts as quality changes', async () => {
    const { user } = await setup();
    // 1640x1232 at quality 100 with 0.45 bytes per pixel: 0.9 MB, 23.75 GB usable, 48 photos a day
    const estimate = await screen.findByText(/per photo/);
    expect(estimate).toHaveTextContent(/About 888 KB per photo, estimated from the last 3 photos\./);
    expect(estimate).toHaveTextContent(/room for about 28,0\d\d more: \d+ days at 48 photos a day, then the oldest photos are deleted/);

    await retype(user, 'JPEG quality', '50');
    expect(screen.getByText(/per photo/)).toHaveTextContent(/About 115 KB per photo/);
  });

  it('says the estimate is rough before the first photo', async () => {
    await setup(settings, withStorage({ Photos: 0, PhotoBytes: 0, SamplePhotos: 0, SampleBytesPerPixel: 0, SampleQuality: 0 }));
    expect(await screen.findByText(/rough estimate until the camera has taken photos/)).toBeInTheDocument();
  });

  it.each([
    ['Restart', 'Restart the camera?', 'POST', '/admin/restart', /Restarting\. Reload this page/],
    ['Shut down', 'Shut down the camera?', 'POST', '/admin/shutdown', /Wait until the green light/],
    ['Delete all photos', 'Delete every photo?', 'DELETE', '/photos', /All photos deleted\./],
  ])('%s asks first, then calls the camera (never with GET)', async (button, question, method, path, done) => {
    const { f, user } = await setup();
    await user.click(screen.getByRole('button', { name: button }));
    expect(screen.getByRole('heading', { name: question })).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Cancel' }));
    await user.click(screen.getByRole('button', { name: button }));
    expect(callsTo(f, method, path)).toHaveLength(0);
    await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: button }));
    await waitFor(() => expect(callsTo(f, method, path)).toHaveLength(1));
    expect(await screen.findByText(done)).toBeInTheDocument();
  });

  it('says when deleting all photos failed', async () => {
    const { user } = await setup(settings, { 'DELETE /photos': () => json('remove 20261101-081500.jpg: read-only file system', 500) });
    await user.click(screen.getByRole('button', { name: 'Delete all photos' }));
    await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Delete all photos' }));
    expect(await screen.findByText('Delete all photos failed. remove 20261101-081500.jpg: read-only file system')).toBeInTheDocument();
  });
});

describe('SettingsView night photos', () => {
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ['Date'] });
    vi.setSystemTime(new Date('2026-11-02T11:00:00Z'));
  });

  it('takes photos around the clock unless nights are skipped', async () => {
    const { f, onSaved, user } = await setup();
    const skip = screen.getByRole('checkbox', { name: 'Skip night photos' });
    expect(skip).not.toBeChecked();
    expect(screen.queryByLabelText('Latitude')).not.toBeInTheDocument();

    await user.click(skip);
    expect(screen.getByText(/Choose the camera's country, or enter its coordinates\./)).toBeInTheDocument();
    expect(saveButton()).toBeDisabled();

    await user.selectOptions(screen.getByLabelText('Location'), 'Exact coordinates…');
    await user.type(screen.getByLabelText('Latitude'), '51.5074');
    await user.type(screen.getByLabelText('Longitude'), '-0.1278');

    await user.click(saveButton());
    await waitFor(() => expect(onSaved).toHaveBeenCalled());
    expect(posted(f)).toEqual({ ...settings, SkipNightPhotos: true, Latitude: 51.5074, Longitude: -0.1278 });
  });

  it('takes the location from a country', async () => {
    const { f, onSaved, user } = await setup();
    await user.click(screen.getByRole('checkbox', { name: 'Skip night photos' }));
    await user.selectOptions(screen.getByLabelText('Location'), 'Ireland');
    expect(screen.queryByLabelText('Latitude')).not.toBeInTheDocument();
    expect(screen.getByText(/the sun rises at \d\d:\d\d and sets at \d\d:\d\d/)).toBeInTheDocument();
    expect(screen.getByText(/the middle of Ireland/)).toBeInTheDocument();
    await user.click(saveButton());
    await waitFor(() => expect(onSaved).toHaveBeenCalled());
    expect(posted(f)).toEqual({ ...settings, SkipNightPhotos: true, Latitude: 53.41, Longitude: -8.24 });
  });

  it('shows a saved location as the country it was picked as', async () => {
    await setup({ ...settings, SkipNightPhotos: true, Latitude: 53.41, Longitude: -8.24 });
    expect(screen.getByLabelText('Location')).toHaveDisplayValue('Ireland');
    expect(screen.queryByLabelText('Latitude')).not.toBeInTheDocument();
  });

  it('shows other saved coordinates in the coordinate fields', async () => {
    await setup({ ...settings, SkipNightPhotos: true, Latitude: 51.5074, Longitude: -0.1278 });
    expect(screen.getByLabelText('Location')).toHaveDisplayValue('Exact coordinates…');
    expect(screen.getByLabelText('Latitude')).toHaveValue('51.5074');
    expect(screen.getByLabelText('Longitude')).toHaveValue('-0.1278');
  });

  it('offers cities instead of the middle of large countries', async () => {
    await setup({ ...settings, SkipNightPhotos: true });
    const options = within(screen.getByLabelText('Location')).getAllByRole('option').map((o) => o.textContent);
    expect(options).toContain('United States – New York');
    expect(options).toContain('United States – Los Angeles');
    expect(options).not.toContain('United States');
    expect(options).toContain('Ireland');
  });

  it('starts the coordinate fields at the chosen country', async () => {
    const { user } = await setup({ ...settings, SkipNightPhotos: true, Latitude: 53.41, Longitude: -8.24 });
    await user.selectOptions(screen.getByLabelText('Location'), 'Exact coordinates…');
    expect(screen.getByLabelText('Latitude')).toHaveValue('53.41');
  });

  it('rejects a location that is not on Earth', async () => {
    const { user } = await setup({ ...settings, SkipNightPhotos: true, Latitude: 51.5, Longitude: 0 });
    await retype(user, 'Latitude', '95');
    expect(screen.getByText(/Latitude goes from -90 \(south\) to 90 \(north\)\./)).toBeInTheDocument();
  });

  it('counts only daytime photos in the photos a day and the storage estimate', async () => {
    await setup({ ...settings, SkipNightPhotos: true, Latitude: 51.5074, Longitude: -0.1278 });
    // About 9.5 of the coming 24 hours are daylight in London in early November: 19 of 48 photos.
    expect(screen.getByText(/About 19 photos a day with nights skipped/)).toBeInTheDocument();
    expect(await screen.findByText(/per photo/)).toHaveTextContent(/at 19 photos a day \(nights skipped\)/);
  });

  it('switches from exact coordinates back to a country', async () => {
    const { f, user } = await setup({ ...settings, SkipNightPhotos: true, Latitude: 51.5074, Longitude: -0.1278 });
    await user.selectOptions(screen.getByLabelText('Location'), 'Ireland');
    expect(screen.queryByLabelText('Latitude')).not.toBeInTheDocument();
    await user.click(saveButton());
    await waitFor(() => expect(callsTo(f, 'POST', '/configuration')).toHaveLength(1));
    expect(posted(f)).toMatchObject({ Latitude: 53.41, Longitude: -8.24 });
  });
});

describe('validate', () => {
  it('matches the backend validator', async () => {
    expect(validate(settings)).toEqual({});
    expect(validate({ ...settings, OffsetWithinHour: -1 })).toEqual({});
    expect(Object.keys(validate({ ...settings, OffsetWithinHour: 1800, Quality: 101 }))).toEqual(['OffsetWithinHour', 'Quality']);
    expect(Object.keys(validate({ ...settings, SecondsBetweenCaptures: 1000 }))).toEqual(['SecondsBetweenCaptures']);
    expect(Object.keys(validate({ ...settings, SecondsBetweenCaptures: 7200 }))).toEqual(['SecondsBetweenCaptures']);
    expect(Object.keys(validate({ ...settings, SkipNightPhotos: true }))).toEqual(['Latitude']);
    expect(Object.keys(validate({ ...settings, Latitude: 91, Longitude: 181 }))).toEqual(['Latitude', 'Longitude']);
    expect(validate({ ...settings, SkipNightPhotos: true, Latitude: 51.5, Longitude: -0.1 })).toEqual({});
  });
});
