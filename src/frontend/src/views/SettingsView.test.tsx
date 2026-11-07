import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SettingsView, validate } from './SettingsView';
import { callsTo, json, mockServer, settings } from '../test/mockServer';
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
    await user.click(screen.getByRole('radio', { name: /2 MP/ }));
    await user.click(screen.getByRole('radio', { name: /Upside down/ }));
    await user.click(screen.getByRole('button', { name: 'Save changes' }));

    await waitFor(() => expect(onSaved).toHaveBeenCalled());
    expect(posted(f)).toEqual({
      ...settings,
      SecondsBetweenCaptures: 600,
      OffsetWithinHour: 180,
      ResolutionSetting: 2,
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
    await user.click(screen.getByRole('radio', { name: /2 MP/ }));
    await user.click(screen.getByRole('button', { name: 'Save changes' }));
    expect(await screen.findByText('Settings were not saved. Quality is out of bounds')).toBeInTheDocument();
  });

  it('warns when the camera did not apply the chosen resolution', async () => {
    // Safety net: if the server ever keeps a different size than requested, say so.
    mockServer({ 'POST /configuration': (_u, init) => json({ ...JSON.parse(String(init?.body)) }) });
    renderSettings();
    const user = userEvent.setup();
    await user.click(screen.getByRole('radio', { name: /8 MP/ }));
    await user.click(screen.getByRole('button', { name: 'Save changes' }));
    expect(await screen.findByText(/camera kept 2186×1642 instead of 3280×2464/)).toBeInTheDocument();
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
