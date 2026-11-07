import { useEffect, useState, type FormEvent, type ReactNode } from 'react';
import { api, type AdminCommand, type Camera, type Settings, type Size, type Storage } from '../api';
import { HourDial } from '../components/HourDial';
import { Icon } from '../components/Icon';
import { ConfirmDialog, Notice, PageHeader, Spinner } from '../components/ui';
import {
  aspectRatio,
  estimateCapacity,
  estimatePhotoBytes,
  formatBytes,
  formatDays,
  formatMegapixels,
  formatPastHour,
} from '../format';
import { useNow, useResource } from '../hooks';
import { capturesPerDay, captureSlots, formatInterval, isValidInterval, VALID_INTERVALS } from '../schedule';

function offsetHint(interval: number): string {
  const last = (interval - 1) / 60;
  return interval >= 120 ? `Pick a minute from 0 to ${Math.floor(last)}.` : `Pick a value from 0 to ${+last.toFixed(2)} minutes.`;
}

/** Mirrors Settings.Validate in src/backend/conf/settings.go so problems show before saving. */
export function validate(s: Settings): Partial<Record<keyof Settings, string>> {
  const errors: Partial<Record<keyof Settings, string>> = {};
  if (!isValidInterval(s.SecondsBetweenCaptures)) {
    errors.SecondsBetweenCaptures = 'Pick one of the listed intervals.';
  } else if (
    s.OffsetWithinHour !== -1 &&
    (!Number.isInteger(s.OffsetWithinHour) || s.OffsetWithinHour < 0 || s.OffsetWithinHour >= s.SecondsBetweenCaptures)
  ) {
    errors.OffsetWithinHour = offsetHint(s.SecondsBetweenCaptures);
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
  const [saving, setSaving] = useState(false);
  const [flash, setFlash] = useState<{ tone: 'ok' | 'error' | 'warn'; text: string } | null>(null);
  const [pending, setPending] = useState<AdminCommand | null>(null);
  const [adminBusy, setAdminBusy] = useState(false);
  const now = useNow();
  const camera = useResource((signal) => api.camera(signal));
  const monitoring = useResource((signal) => api.monitoring(signal));

  useEffect(() => {
    if (settings && !draft) setDraft(settings);
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
  const intervalSupported = isValidInterval(draft.SecondsBetweenCaptures);
  const slots = captureSlots(draft);
  // Changing the interval keeps the capture times lined up: offset modulo the
  // new interval is the earliest capture of the old schedule within the hour.
  const setInterval = (seconds: number) =>
    setDraft({
      ...draft,
      SecondsBetweenCaptures: seconds,
      OffsetWithinHour: aligned ? Math.max(0, draft.OffsetWithinHour) % seconds : -1,
    });

  const save = async (e: FormEvent) => {
    e.preventDefault();
    if (invalid || !dirty) return;
    setSaving(true);
    setFlash(null);
    try {
      const saved = await api.saveSettings(draft);
      onSaved(saved);
      setDraft(saved);
      const resolutionApplied =
        saved.PhotoResolutionWidth === draft.PhotoResolutionWidth && saved.PhotoResolutionHeight === draft.PhotoResolutionHeight;
      setFlash(
        resolutionApplied
          ? { tone: 'ok', text: 'Settings saved. They apply from the next capture.' }
          : {
              tone: 'warn',
              text: `Settings saved, but the camera kept ${saved.PhotoResolutionWidth}×${saved.PhotoResolutionHeight} instead of ${draft.PhotoResolutionWidth}×${draft.PhotoResolutionHeight}.`,
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
                <select
                  id="interval"
                  value={draft.SecondsBetweenCaptures}
                  aria-invalid={!!errors.SecondsBetweenCaptures}
                  aria-describedby="interval-help"
                  onChange={(e) => setInterval(Number(e.target.value))}
                >
                  {!intervalSupported && (
                    <option value={draft.SecondsBetweenCaptures} disabled>
                      {formatInterval(draft.SecondsBetweenCaptures)} (not supported)
                    </option>
                  )}
                  {VALID_INTERVALS.map((v) => (
                    <option key={v} value={v}>
                      {formatInterval(v)}
                    </option>
                  ))}
                </select>
                <p id="interval-help" className={errors.SecondsBetweenCaptures ? 'help error' : 'help'}>
                  {errors.SecondsBetweenCaptures ??
                    `${capturesPerDay(draft).toLocaleString()} photos a day. Intervals divide the hour evenly so every hour has the same capture times.`}
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
                      max={(draft.SecondsBetweenCaptures - 1) / 60}
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

          <ResolutionField
            camera={camera.data}
            loading={camera.loading}
            error={camera.error}
            value={{ Width: draft.PhotoResolutionWidth, Height: draft.PhotoResolutionHeight }}
            saved={settings && { Width: settings.PhotoResolutionWidth, Height: settings.PhotoResolutionHeight }}
            onChange={(r) => setDraft({ ...draft, PhotoResolutionWidth: r.Width, PhotoResolutionHeight: r.Height })}
          />

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
            {!errors.Quality && <StorageEstimate draft={draft} storage={monitoring.data?.Storage} />}
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
            <p className="help">Logs every request and scheduling step. Useful for troubleshooting only.</p>
          </div>
        </fieldset>

        <fieldset className="panel">
          <legend>Storage</legend>
          <div className="field">
            <span className="label" id="full-label">When the card is full</span>
            <div className="segmented" role="radiogroup" aria-labelledby="full-label">
              <label>
                <input
                  type="radio"
                  name="when-full"
                  checked={!draft.StopWhenDiskFull}
                  onChange={() => set('StopWhenDiskFull', false)}
                />
                <span>
                  <strong>Delete oldest</strong>
                  Keep taking photos
                </span>
              </label>
              <label>
                <input
                  type="radio"
                  name="when-full"
                  checked={draft.StopWhenDiskFull}
                  onChange={() => set('StopWhenDiskFull', true)}
                />
                <span>
                  <strong>Stop</strong>
                  Keep every photo
                </span>
              </label>
            </div>
            <p className="help">
              {draft.StopWhenDiskFull
                ? 'The camera stops taking photos when the card is full, until photos are deleted.'
                : 'The oldest photos are deleted to make room for new ones, so the card always holds the most recent ones.'}{' '}
              Status shows how long until the card is full.
            </p>
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

const sameSize = (a: Size, b: Size) => a.Width === b.Width && a.Height === b.Height;

interface ResolutionProps {
  camera: Camera | undefined;
  loading: boolean;
  error: string | null;
  value: Size;
  saved: Size | undefined;
  onChange: (s: Size) => void;
}

/**
 * The connected camera's photo sizes (its sensor modes), so only sizes it
 * supports can be chosen. A saved size it lacks (e.g. after swapping the
 * camera) stays selectable until another one is chosen.
 */
function ResolutionField({ camera, loading, error, value, saved, onChange }: ResolutionProps) {
  const detected = camera && !camera.Error && camera.Resolutions?.length ? camera : null;
  const supported = detected?.Resolutions ?? [];
  const options = [...supported];
  for (const extra of [saved, value]) {
    if (extra && !options.some((o) => sameSize(o, extra))) options.push(extra);
  }
  const isSupported = (s: Size) => !detected || supported.some((o) => sameSize(o, s));

  let help: ReactNode;
  if (detected) {
    help = (
      <>
        Sizes of the {detected.Model} camera ({detected.Width}×{detected.Height} sensor). Lower resolutions fit more photos
        on the card.
        {!isSupported(value) && (
          <>
            {' '}
            <strong>
              The camera does not support {value.Width}×{value.Height}: photos are scaled. Choose one of its sizes.
            </strong>
          </>
        )}
      </>
    );
  } else if (loading) {
    help = 'Detecting the camera…';
  } else {
    help = `Couldn't detect the camera (${camera?.Error ?? error ?? 'unknown error'}), so only the current size is shown.`;
  }

  return (
    <div className="field">
      <span className="label" id="res-label">
        Resolution
      </span>
      <div className="segmented" role="radiogroup" aria-labelledby="res-label">
        {options.map((r) => (
          <label key={`${r.Width}x${r.Height}`}>
            <input type="radio" name="resolution" checked={sameSize(value, r)} onChange={() => onChange(r)} />
            <span>
              <strong>{formatMegapixels(r.Width, r.Height)}</strong>
              {r.Width}×{r.Height} · {isSupported(r) ? aspectRatio(r.Width, r.Height) : 'not supported'}
            </span>
          </label>
        ))}
      </div>
      <p className="help">{help}</p>
    </div>
  );
}

/** Photo size at the chosen size and quality, and how long the card lasts at the chosen schedule. */
function StorageEstimate({ draft, storage }: { draft: Settings; storage: Storage | undefined }) {
  const photo = estimatePhotoBytes(draft.PhotoResolutionWidth, draft.PhotoResolutionHeight, draft.Quality, storage);
  const perDay = capturesPerDay(draft);
  const capacity = storage && estimateCapacity(storage, photo.bytes, perDay);
  const basis =
    photo.fromPhotos > 0
      ? `estimated from the last ${photo.fromPhotos.toLocaleString()} photo${photo.fromPhotos === 1 ? '' : 's'}`
      : 'a rough estimate until the camera has taken photos';
  const atSchedule = (days: number | null) => (days === null ? '' : `: ${formatDays(days)} at ${perDay.toLocaleString()} photos a day`);
  return (
    <p className="help estimate">
      About {formatBytes(photo.bytes)} per photo, {basis}.
      {capacity && (
        <>
          {' '}
          The card has room for about {capacity.more.toLocaleString()} more{atSchedule(capacity.days)}, then{' '}
          {capacity.kept === null
            ? 'the camera stops taking photos.'
            : `the oldest photos are deleted, so the card keeps about the last ${capacity.kept.toLocaleString()} photos${
                capacity.keptDays === null ? '' : ` (${formatDays(capacity.keptDays)})`
              }.`}
        </>
      )}
    </p>
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

