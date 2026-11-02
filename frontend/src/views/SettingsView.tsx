import { useEffect, useState, type FormEvent } from 'react';
import { api, RESOLUTIONS, type AdminCommand, type Settings } from '../api';
import { HourDial } from '../components/HourDial';
import { Icon } from '../components/Icon';
import { ConfirmDialog, Notice, PageHeader, Spinner } from '../components/ui';
import { formatPastHour } from '../format';
import { useNow, useResource } from '../hooks';
import { capturesPerDay, captureSlots } from '../schedule';

type Unit = 'sec' | 'min' | 'h';
const UNIT_SECONDS: Record<Unit, number> = { sec: 1, min: 60, h: 3600 };

function bestUnit(seconds: number): Unit {
  if (seconds % 3600 === 0) return 'h';
  if (seconds % 60 === 0) return 'min';
  return 'sec';
}

/** Mirrors conf/valid/strict_validator.go so problems show before saving. */
export function validate(s: Settings): Partial<Record<keyof Settings, string>> {
  const errors: Partial<Record<keyof Settings, string>> = {};
  if (!Number.isInteger(s.SecondsBetweenCaptures) || s.SecondsBetweenCaptures < 1) {
    errors.SecondsBetweenCaptures = 'Use a whole number of seconds, 1 or more.';
  }
  if (s.OffsetWithinHour !== -1 && (!Number.isInteger(s.OffsetWithinHour) || s.OffsetWithinHour < 0 || s.OffsetWithinHour > 3599)) {
    errors.OffsetWithinHour = 'Pick a minute from 0 to 59.';
  }
  if (!Number.isInteger(s.Quality) || s.Quality < 1 || s.Quality > 100) {
    errors.Quality = 'Use a value from 1 to 100.';
  }
  return errors;
}

const ADMIN: Record<AdminCommand, { title: string; body: string; confirm: string; done: string }> = {
  restart: {
    title: 'Restart the camera?',
    body: "The camera will be unreachable for about a minute. If this device drops off the camera's WiFi, reconnect to it afterwards.",
    confirm: 'Restart',
    done: 'Restarting. Reload this page in about a minute.',
  },
  shutdown: {
    title: 'Shut down the camera?',
    body: 'No more photos will be taken until it is powered on again. You will need physical access to turn it back on.',
    confirm: 'Shut down',
    done: 'Shutting down. Wait until the green light on the Pi stays off before unplugging it, so the SD card is not damaged.',
  },
  clear: {
    title: 'Delete every photo?',
    body: "All photos on the camera's SD card will be removed. This can't be undone. Download an archive from Photos first if you want to keep them.",
    confirm: 'Delete all photos',
    done: 'All photos deleted.',
  },
};

interface Props {
  settings: Settings | undefined;
  error: string | null;
  onSaved: (s: Settings) => void;
}

export function SettingsView({ settings, error, onSaved }: Props) {
  const [draft, setDraft] = useState<Settings | undefined>(settings);
  const [unit, setUnit] = useState<Unit>('min');
  const [saving, setSaving] = useState(false);
  const [flash, setFlash] = useState<{ tone: 'ok' | 'error' | 'warn'; text: string } | null>(null);
  const [pending, setPending] = useState<AdminCommand | null>(null);
  const [adminBusy, setAdminBusy] = useState(false);
  const now = useNow();

  useEffect(() => {
    if (settings && !draft) {
      setDraft(settings);
      setUnit(bestUnit(settings.SecondsBetweenCaptures));
    }
  }, [settings, draft]);

  if (!draft) {
    return (
      <>
        <PageHeader title="Settings" />
        {error ? <Notice tone="error">{error}</Notice> : <Spinner label="Loading settings…" />}
      </>
    );
  }

  const errors = validate(draft);
  const invalid = Object.keys(errors).length > 0;
  const dirty = settings !== undefined && JSON.stringify(settings) !== JSON.stringify(draft);
  const set = <K extends keyof Settings>(k: K, v: Settings[K]) => setDraft({ ...draft, [k]: v });
  const aligned = draft.OffsetWithinHour !== -1;
  const intervalValue = draft.SecondsBetweenCaptures / UNIT_SECONDS[unit];
  const slots = captureSlots(draft);

  const save = async (e: FormEvent) => {
    e.preventDefault();
    if (invalid || !dirty) return;
    setSaving(true);
    setFlash(null);
    try {
      const saved = await api.saveSettings(draft);
      onSaved(saved);
      setDraft(saved);
      const resolution = RESOLUTIONS.find((r) => r.setting === draft.ResolutionSetting);
      const resolutionApplied =
        !resolution || (saved.PhotoResolutionWidth === resolution.width && saved.PhotoResolutionHeight === resolution.height);
      setFlash(
        resolutionApplied
          ? { tone: 'ok', text: 'Settings saved. They apply from the next capture.' }
          : {
              tone: 'warn',
              text: `Settings saved, but the camera kept ${saved.PhotoResolutionWidth}×${saved.PhotoResolutionHeight} instead of ${resolution.width}×${resolution.height}.`,
            },
      );
    } catch (err) {
      setFlash({ tone: 'error', text: `Settings were not saved. ${(err as Error).message}` });
    } finally {
      setSaving(false);
    }
  };

  const runAdmin = async () => {
    if (!pending) return;
    const cmd = pending;
    setAdminBusy(true);
    try {
      await api.admin(cmd);
      setFlash({ tone: cmd === 'clear' ? 'ok' : 'warn', text: ADMIN[cmd].done });
    } catch (err) {
      setFlash({ tone: 'error', text: `${ADMIN[cmd].confirm} failed. ${(err as Error).message}` });
    } finally {
      setAdminBusy(false);
      setPending(null);
    }
  };

  const rotations = [0, 180, ...(draft.RotateBy !== 0 && draft.RotateBy !== 180 ? [draft.RotateBy] : [])];

  return (
    <>
      <PageHeader title="Settings" />

      {flash && (
        <Notice tone={flash.tone} onDismiss={() => setFlash(null)}>
          {flash.text}
        </Notice>
      )}

      <form className="settings" onSubmit={save} noValidate>
        <fieldset className="panel schedule">
          <legend>Schedule</legend>
          <div className="schedule-grid">
            <div className="fields">
              <div className="field">
                <label htmlFor="interval">Take a photo every</label>
                <div className="inline">
                  <input
                    id="interval"
                    type="number"
                    inputMode="decimal"
                    min={unit === 'sec' ? 1 : 0}
                    step={unit === 'sec' ? 1 : 'any'}
                    value={Number.isFinite(intervalValue) ? +intervalValue.toFixed(4) : ''}
                    aria-invalid={!!errors.SecondsBetweenCaptures}
                    aria-describedby="interval-help"
                    onChange={(e) => set('SecondsBetweenCaptures', Math.round(e.target.valueAsNumber * UNIT_SECONDS[unit]))}
                  />
                  <select aria-label="Unit" value={unit} onChange={(e) => setUnit(e.target.value as Unit)}>
                    <option value="sec">seconds</option>
                    <option value="min">minutes</option>
                    <option value="h">hours</option>
                  </select>
                </div>
                <p id="interval-help" className={errors.SecondsBetweenCaptures ? 'help error' : 'help'}>
                  {errors.SecondsBetweenCaptures ??
                    `About ${capturesPerDay(draft).toLocaleString()} photos a day.`}
                </p>
              </div>

              <div className="field">
                <label className="check">
                  <input
                    type="checkbox"
                    checked={aligned}
                    onChange={(e) =>
                      set('OffsetWithinHour', e.target.checked ? (settings && settings.OffsetWithinHour >= 0 ? settings.OffsetWithinHour : 0) : -1)
                    }
                  />
                  Line captures up with the clock
                </label>
                <p className="help">
                  {aligned
                    ? 'Photos are taken at the same minutes every hour, which makes days easy to compare.'
                    : 'Photos are taken at even intervals, counted from when the camera starts.'}
                </p>
              </div>

              {aligned && (
                <div className="field">
                  <label htmlFor="offset">First capture at minute</label>
                  <div className="inline">
                    <input
                      id="offset"
                      type="number"
                      inputMode="numeric"
                      min={0}
                      max={59}
                      step="any"
                      value={+(draft.OffsetWithinHour / 60).toFixed(2)}
                      aria-invalid={!!errors.OffsetWithinHour}
                      aria-describedby="offset-help"
                      onChange={(e) => set('OffsetWithinHour', Math.round(e.target.valueAsNumber * 60))}
                    />
                    <span className="affix">past the hour</span>
                  </div>
                  <p id="offset-help" className={errors.OffsetWithinHour ? 'help error' : 'help'}>
                    {errors.OffsetWithinHour ??
                      (slots && slots.length <= 6
                        ? `Captures at ${slots.map(formatPastHour).join(', ')}.`
                        : `${slots?.length ?? 0} captures an hour.`)}
                  </p>
                </div>
              )}
            </div>
            <HourDial
              schedule={draft}
              now={now}
              label={dirty ? 'Unsaved schedule' : undefined}
            />
          </div>
        </fieldset>

        <fieldset className="panel">
          <legend>Photo</legend>

          <div className="field">
            <span className="label" id="res-label">Resolution</span>
            <div className="segmented" role="radiogroup" aria-labelledby="res-label">
              {RESOLUTIONS.map((r) => (
                <label key={r.setting}>
                  <input
                    type="radio"
                    name="resolution"
                    checked={draft.ResolutionSetting === r.setting}
                    onChange={() => set('ResolutionSetting', r.setting)}
                  />
                  <span>
                    <strong>{r.label}</strong>
                    {r.width}×{r.height}
                  </span>
                </label>
              ))}
            </div>
            <p className="help">
              Photos are currently saved at {settings?.PhotoResolutionWidth}×{settings?.PhotoResolutionHeight}. Lower
              resolutions fit more photos on the card.
            </p>
          </div>

          <div className="field">
            <label htmlFor="quality">JPEG quality</label>
            <div className="inline">
              <input
                id="quality-range"
                type="range"
                min={1}
                max={100}
                value={Number.isFinite(draft.Quality) ? draft.Quality : 1}
                aria-label="JPEG quality slider"
                onChange={(e) => set('Quality', e.target.valueAsNumber)}
              />
              <input
                id="quality"
                className="narrow"
                type="number"
                inputMode="numeric"
                min={1}
                max={100}
                value={Number.isFinite(draft.Quality) ? draft.Quality : ''}
                aria-invalid={!!errors.Quality}
                aria-describedby="quality-help"
                onChange={(e) => set('Quality', e.target.valueAsNumber)}
              />
            </div>
            <p id="quality-help" className={errors.Quality ? 'help error' : 'help'}>
              {errors.Quality ?? 'Lower values make smaller files. Few people can tell 85 from 100.'}
            </p>
          </div>

          <div className="field">
            <span className="label" id="rot-label">Orientation</span>
            <div className="segmented" role="radiogroup" aria-labelledby="rot-label">
              {rotations.map((deg) => (
                <label key={deg}>
                  <input type="radio" name="rotation" checked={draft.RotateBy === deg} onChange={() => set('RotateBy', deg)} />
                  <span>
                    <strong>{deg === 0 ? 'Upright' : deg === 180 ? 'Upside down' : `${deg}°`}</strong>
                    {deg === 0 ? 'No rotation' : deg === 180 ? 'Rotate 180°' : 'Not supported by the camera'}
                  </span>
                </label>
              ))}
            </div>
            <p className="help">Choose upside down if the camera is mounted hanging from its cable.</p>
          </div>

          <div className="field">
            <label className="check">
              <input type="checkbox" checked={draft.DebugEnabled} onChange={(e) => set('DebugEnabled', e.target.checked)} />
              Detailed logging
            </label>
            <p className="help">Writes a line every second while waiting for a capture. Useful for troubleshooting only.</p>
          </div>
        </fieldset>

        <div className="save-bar" data-dirty={dirty}>
          <span>{dirty ? (invalid ? 'Fix the highlighted fields to save.' : 'You have unsaved changes.') : 'All changes saved.'}</span>
          <div>
            <button
              type="button"
              className="btn ghost"
              disabled={!dirty || saving}
              onClick={() => settings && setDraft(settings)}
            >
              Discard
            </button>
            <button type="submit" className="btn primary" disabled={!dirty || invalid || saving}>
              {saving ? 'Saving…' : 'Save changes'}
            </button>
          </div>
        </div>
      </form>

      <section className="panel system">
        <h2>Camera</h2>
        <div className="system-actions">
          <button type="button" className="btn" onClick={() => setPending('restart')}>
            <Icon name="restart" /> Restart
          </button>
          <button type="button" className="btn" onClick={() => setPending('shutdown')}>
            <Icon name="power" /> Shut down
          </button>
          <button type="button" className="btn danger-ghost" onClick={() => setPending('clear')}>
            <Icon name="trash" /> Delete all photos
          </button>
        </div>
        <Version />
      </section>

      <ConfirmDialog
        open={pending !== null}
        title={pending ? ADMIN[pending].title : ''}
        confirmLabel={pending ? ADMIN[pending].confirm : ''}
        danger
        busy={adminBusy}
        onConfirm={runAdmin}
        onCancel={() => setPending(null)}
      >
        {pending ? ADMIN[pending].body : null}
      </ConfirmDialog>
    </>
  );
}

function Version() {
  const server = useResource((signal) => api.version(signal));
  const b = __BUILD_INFO__;
  return (
    <p className="version">
      Interface{' '}
      {b.sha ? (
        <>
          <a href={`https://github.com/ingojaeckel/go-raspberry-pi-timelapse/commit/${b.sha}`}>{b.shaShort || b.sha.slice(0, 7)}</a>
          {b.time && `, built ${b.time}`}
        </>
      ) : (
        'development build'
      )}
      {server.data && (
        <>
          <br />
          Server: {server.data.replace(/^Hello from /, '')}
        </>
      )}
    </p>
  );
}

