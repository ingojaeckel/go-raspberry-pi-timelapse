import { useEffect, useState, type FormEvent } from 'react';
import { api, errorMessage, type AdminCommand, type Camera, type Settings, type Storage } from '../api';
import { CheckField, ChoiceField, Help } from '../components/fields';
import { HourDial } from '../components/HourDial';
import { Icon, type IconName } from '../components/Icon';
import { ConfirmDialog, MessageNotice, Notice, PageHeader, Spinner, type Message } from '../components/ui';
import { formatPastHours, formatSize } from '../format';
import { useBusy, useNow, useResource, type Resource } from '../hooks';
import { capturesPerDay, captureSlots, formatInterval, isValidInterval, OFFERED_INTERVALS, photosPerDay, skipsNights } from '../schedule';
import { noLocation } from '../sun';
import { NightFields, type SettingsErrors } from './NightFields';
import { QualityField, ResolutionField, RotationField } from './PhotoFields';

function offsetHint(interval: number): string {
  const last = (interval - 1) / 60;
  return interval >= 120 ? `Pick a minute from 0 to ${Math.floor(last)}.` : `Pick a value from 0 to ${+last.toFixed(2)} minutes.`;
}

/** Mirrors Settings.Validate in src/backend/conf/settings.go so problems show before saving. */
export function validate(s: Settings): SettingsErrors {
  const errors: SettingsErrors = {};
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
  if (!(s.Latitude >= -90 && s.Latitude <= 90)) {
    errors.Latitude = 'Latitude goes from -90 (south) to 90 (north).';
  }
  if (!(s.Longitude >= -180 && s.Longitude <= 180)) {
    errors.Longitude = 'Longitude goes from -180 (west) to 180 (east).';
  }
  if (s.SkipNightPhotos && !errors.Latitude && !errors.Longitude && noLocation(s.Latitude, s.Longitude)) {
    errors.Latitude = "Choose the camera's country, or enter its coordinates.";
  }
  return errors;
}

interface AdminAction {
  title: string;
  body: string;
  /** The label of the button that asks, and of the one that confirms. */
  confirm: string;
  done: string;
  icon: IconName;
  button: string;
}

const ADMIN: Record<AdminCommand, AdminAction> = {
  restart: {
    title: 'Restart the camera?',
    body: "The camera will be unreachable for about a minute. If this device drops off the camera's WiFi, reconnect to it afterwards.",
    confirm: 'Restart',
    done: 'Restarting. Reload this page in about a minute.',
    icon: 'restart',
    button: 'btn',
  },
  shutdown: {
    title: 'Shut down the camera?',
    body: 'No more photos will be taken until it is powered on again. You will need physical access to turn it back on.',
    confirm: 'Shut down',
    done: 'Shutting down. Wait until the green light on the Pi stays off before unplugging it, so the SD card is not damaged.',
    icon: 'power',
    button: 'btn',
  },
  clear: {
    title: 'Delete every photo?',
    body: "All photos on the camera's SD card will be removed. This can't be undone. Download an archive from Photos first if you want to keep them.",
    confirm: 'Delete all photos',
    done: 'All photos deleted.',
    icon: 'trash',
    button: 'btn danger-ghost',
  },
};

interface Props {
  settings: Settings | undefined;
  error: string | null;
  onSaved: (s: Settings) => void;
}

export function SettingsView({ settings, error, onSaved }: Props) {
  const [draft, setDraft] = useState<Settings | undefined>(settings);
  const [message, setMessage] = useState<Message | null>(null);
  const [saving, runSave] = useBusy();
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
  const fields = { draft, errors, onChange: (patch: Partial<Settings>) => setDraft({ ...draft, ...patch }) };

  const saveDraft = async () => {
    setMessage(null);
    try {
      const saved = await api.saveSettings(draft);
      onSaved(saved);
      setDraft(saved);
      setMessage(savedMessage(draft, saved));
    } catch (err) {
      setMessage({ tone: 'error', text: `Settings were not saved. ${errorMessage(err)}` });
    }
  };
  const save = (e: FormEvent) => {
    e.preventDefault();
    if (!invalid && dirty) void runSave(saveDraft);
  };

  return (
    <>
      <PageHeader title="Settings" />
      <MessageNotice message={message} onDismiss={() => setMessage(null)} />
      <form className="settings" onSubmit={save} noValidate>
        <ScheduleSection {...fields} saved={settings} now={now} dirty={dirty} />
        <PhotoSection {...fields} saved={settings} camera={camera} storage={monitoring.data?.Storage} />
        <StorageSection {...fields} />
        <SaveBar dirty={dirty} invalid={invalid} saving={saving} onDiscard={() => settings && setDraft(settings)} />
      </form>
      <CameraActions onResult={setMessage} />
    </>
  );
}

/** Saved, or saved but the camera kept another photo size than the one chosen. */
function savedMessage(draft: Settings, saved: Settings): Message {
  const [chosen, kept] = [photoSize(draft), photoSize(saved)].map(formatSize);
  if (chosen === kept) return { tone: 'ok', text: 'Settings saved. They apply from the next capture.' };
  return { tone: 'warn', text: `Settings saved, but the camera kept ${kept} instead of ${chosen}.` };
}

const photoSize = (s: Settings) => ({ Width: s.PhotoResolutionWidth, Height: s.PhotoResolutionHeight });

interface FieldsProps {
  draft: Settings;
  errors: SettingsErrors;
  onChange: (patch: Partial<Settings>) => void;
}

function ScheduleSection({ saved, now, dirty, ...fields }: FieldsProps & { saved: Settings | undefined; now: Date; dirty: boolean }) {
  const { draft, onChange } = fields;
  const aligned = draft.OffsetWithinHour !== -1;
  // Turning alignment back on restores the saved offset.
  const alignedOffset = saved && saved.OffsetWithinHour >= 0 ? saved.OffsetWithinHour : 0;
  return (
    <fieldset className="panel schedule">
      <legend>Schedule</legend>
      <div className="schedule-grid">
        <div className="fields">
          <IntervalField {...fields} now={now} />
          <CheckField
            label="Line captures up with the clock"
            checked={aligned}
            onChange={(on) => onChange({ OffsetWithinHour: on ? alignedOffset : -1 })}
            help={
              aligned
                ? 'Photos are taken at the same minutes every hour, which makes days easy to compare.'
                : 'Photos are taken at even intervals, counted from when the camera starts.'
            }
          />
          <NightFields {...fields} now={now} />
          {aligned && <OffsetField {...fields} />}
        </div>
        <HourDial schedule={draft} now={now} label={dirty ? 'Unsaved schedule' : undefined} />
      </div>
    </fieldset>
  );
}

/**
 * Changing the interval keeps the capture times lined up: offset modulo the
 * new interval is the earliest capture of the old schedule within the hour.
 */
function withInterval(s: Settings, seconds: number): Partial<Settings> {
  const aligned = s.OffsetWithinHour !== -1;
  return { SecondsBetweenCaptures: seconds, OffsetWithinHour: aligned ? Math.max(0, s.OffsetWithinHour) % seconds : -1 };
}

function IntervalField({ draft, errors, now, onChange }: FieldsProps & { now: Date }) {
  const seconds = draft.SecondsBetweenCaptures;
  const error = errors.SecondsBetweenCaptures;
  return (
    <div className="field">
      <label htmlFor="interval">Take a photo every</label>
      <select
        id="interval"
        value={seconds}
        aria-invalid={!!error}
        aria-describedby="interval-help"
        onChange={(e) => onChange(withInterval(draft, Number(e.target.value)))}
      >
        {!isValidInterval(seconds) && (
          <option value={seconds} disabled>
            {formatInterval(seconds)} (not supported)
          </option>
        )}
        {intervalChoices(seconds).map((v) => (
          <option key={v} value={v}>
            {formatInterval(v)}
          </option>
        ))}
      </select>
      <Help id="interval-help" invalid={!!error}>
        {error ?? `${photosADay(draft, now)}. Intervals divide the hour evenly so every hour has the same capture times.`}
      </Help>
    </div>
  );
}

/** The offered intervals, plus the current one if it is valid but not offered (a saved interval of seconds). */
function intervalChoices(current: number): number[] {
  if (!isValidInterval(current) || OFFERED_INTERVALS.includes(current)) return [...OFFERED_INTERVALS];
  return [...OFFERED_INTERVALS, current].sort((a, b) => a - b);
}

function photosADay(s: Settings, now: Date): string {
  if (!skipsNights(s, now)) return `${capturesPerDay(s).toLocaleString()} photos a day`;
  return `About ${photosPerDay(s, now).toLocaleString()} photos a day with nights skipped`;
}

function OffsetField({ draft, errors, onChange }: FieldsProps) {
  const error = errors.OffsetWithinHour;
  const slots = captureSlots(draft);
  return (
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
          aria-invalid={!!error}
          aria-describedby="offset-help"
          onChange={(e) => onChange({ OffsetWithinHour: Math.round(e.target.valueAsNumber * 60) })}
        />
        <span className="affix">past the hour</span>
      </div>
      <Help id="offset-help" invalid={!!error}>
        {error ?? (slots && slots.length <= 6 ? `Captures at ${formatPastHours(slots)}.` : `${slots?.length ?? 0} captures an hour.`)}
      </Help>
    </div>
  );
}

interface PhotoProps extends FieldsProps {
  saved: Settings | undefined;
  camera: Resource<Camera>;
  storage: Storage | undefined;
}

function PhotoSection({ draft, errors, onChange, saved, camera, storage }: PhotoProps) {
  return (
    <fieldset className="panel">
      <legend>Photo</legend>
      <ResolutionField
        camera={camera}
        value={photoSize(draft)}
        saved={saved && photoSize(saved)}
        onChange={(size) => onChange({ PhotoResolutionWidth: size.Width, PhotoResolutionHeight: size.Height })}
      />
      <QualityField draft={draft} error={errors.Quality} storage={storage} onChange={(q) => onChange({ Quality: q })} />
      <RotationField rotateBy={draft.RotateBy} onChange={(degrees) => onChange({ RotateBy: degrees })} />
      <CheckField
        label="Detailed logging"
        checked={draft.DebugEnabled}
        onChange={(on) => onChange({ DebugEnabled: on })}
        help="Logs every request and scheduling step. Useful for troubleshooting only."
      />
    </fieldset>
  );
}

function StorageSection({ draft, onChange }: FieldsProps) {
  const stop = draft.StopWhenDiskFull;
  const choice = (key: string, title: string, detail: string, value: boolean) => ({
    key,
    title,
    detail,
    checked: stop === value,
    onSelect: () => onChange({ StopWhenDiskFull: value }),
  });
  const outcome = stop
    ? 'The camera stops taking photos when the card is full, until photos are deleted.'
    : 'The oldest photos are deleted to make room for new ones, so the card always holds the most recent ones.';
  return (
    <fieldset className="panel">
      <legend>Storage</legend>
      <ChoiceField
        id="full-label"
        label="When the card is full"
        name="when-full"
        choices={[choice('delete', 'Delete oldest', 'Keep taking photos', false), choice('stop', 'Stop', 'Keep every photo', true)]}
        help={`${outcome} Status shows how long until the card is full.`}
      />
    </fieldset>
  );
}

function SaveBar({ dirty, invalid, saving, onDiscard }: { dirty: boolean; invalid: boolean; saving: boolean; onDiscard: () => void }) {
  const status = !dirty ? 'All changes saved.' : invalid ? 'Fix the highlighted fields to save.' : 'You have unsaved changes.';
  return (
    <div className="save-bar" data-dirty={dirty}>
      <span>{status}</span>
      <div>
        <button type="button" className="btn ghost" disabled={!dirty || saving} onClick={onDiscard}>
          Discard
        </button>
        <button type="submit" className="btn primary" disabled={!dirty || invalid || saving}>
          {saving ? 'Saving…' : 'Save changes'}
        </button>
      </div>
    </div>
  );
}

async function runAdmin(command: AdminCommand, report: (m: Message) => void) {
  try {
    await api.admin(command);
    report({ tone: command === 'clear' ? 'ok' : 'warn', text: ADMIN[command].done });
  } catch (err) {
    report({ tone: 'error', text: `${ADMIN[command].confirm} failed. ${errorMessage(err)}` });
  }
}

/** Restart, shut down and delete all photos, each after confirming. */
function CameraActions({ onResult }: { onResult: (m: Message) => void }) {
  const [pending, setPending] = useState<AdminCommand | null>(null);
  const [busy, runBusy] = useBusy();
  const action = pending && ADMIN[pending];
  const confirm = async () => {
    if (!pending) return;
    await runBusy(() => runAdmin(pending, onResult));
    setPending(null);
  };
  return (
    <>
      <section className="panel system">
        <h2>Camera</h2>
        <div className="system-actions">
          {(Object.keys(ADMIN) as AdminCommand[]).map((command) => (
            <button key={command} type="button" className={ADMIN[command].button} onClick={() => setPending(command)}>
              <Icon name={ADMIN[command].icon} /> {ADMIN[command].confirm}
            </button>
          ))}
        </div>
        <Version />
      </section>
      <ConfirmDialog
        open={action !== null}
        title={action?.title ?? ''}
        confirmLabel={action?.confirm ?? ''}
        danger
        busy={busy}
        onConfirm={confirm}
        onCancel={() => setPending(null)}
      >
        {action?.body ?? null}
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

