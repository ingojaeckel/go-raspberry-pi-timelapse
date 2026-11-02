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
  it('shows the interval in minutes and the offset as a minute past the hour', () => {
    mockServer();
    renderSettings();
    expect(screen.getByLabelText('Take a photo every')).toHaveValue(30);
    expect(screen.getByRole('combobox', { name: 'Unit' })).toHaveValue('min');
    // 900 seconds is minute 15, not "900 minutes" as the old UI showed
    expect(screen.getByLabelText('First capture at minute')).toHaveValue(15);
    expect(screen.getByText('Captures at :15, :45.')).toBeInTheDocument();
  });

  it('saves the full settings object in the units the backend expects', async () => {
    const f = mockServer();
    const onSaved = renderSettings();
    const user = userEvent.setup();

    await user.clear(screen.getByLabelText('Take a photo every'));
    await user.type(screen.getByLabelText('Take a photo every'), '10');
    await user.clear(screen.getByLabelText('First capture at minute'));
    await user.type(screen.getByLabelText('First capture at minute'), '5');
    await user.click(screen.getByRole('radio', { name: /2 MP/ }));
    await user.click(screen.getByRole('radio', { name: /Upside down/ }));
    await user.click(screen.getByRole('button', { name: 'Save changes' }));

    await waitFor(() => expect(onSaved).toHaveBeenCalled());
    expect(posted(f)).toEqual({
      ...settings,
      SecondsBetweenCaptures: 600,
      OffsetWithinHour: 300,
      ResolutionSetting: 2,
      RotateBy: 180,
    });
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
    // rest.updatePartialConfiguration ignores ResolutionSetting 0 and keeps the previous size.
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
    expect(callsTo(f, 'GET', '/admin/shutdown')).toHaveLength(0);
    expect(screen.getByRole('heading', { name: 'Shut down the camera?' })).toBeInTheDocument();
    const dialog = screen.getByRole('dialog');
    await user.click(within(dialog).getByRole('button', { name: 'Shut down' }));
    await waitFor(() => expect(callsTo(f, 'GET', '/admin/shutdown')).toHaveLength(1));
    expect(await screen.findByText(/Wait until the green light/)).toBeInTheDocument();
  });

  it('exposes delete-all through /admin/clear behind a confirmation', async () => {
    const f = mockServer();
    renderSettings();
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: 'Delete all photos' }));
    await user.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(callsTo(f, 'GET', '/admin/clear')).toHaveLength(0);
  });
});

describe('validate', () => {
  it('matches the backend validator', () => {
    expect(validate(settings)).toEqual({});
    expect(validate({ ...settings, OffsetWithinHour: -1 })).toEqual({});
    expect(Object.keys(validate({ ...settings, OffsetWithinHour: 3600, SecondsBetweenCaptures: 0, Quality: 101 }))).toEqual([
      'SecondsBetweenCaptures',
      'OffsetWithinHour',
      'Quality',
    ]);
  });
});

