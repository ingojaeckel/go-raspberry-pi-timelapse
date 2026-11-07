import { useEffect, useState } from 'react';
import type { Settings } from '../api';
import { Check, Help } from '../components/fields';
import { formatClock } from '../format';
import { PLACES, placeAt, type Place } from '../places';
import { nextSunrise, nextSunset, noLocation, polarNight } from '../sun';

export type SettingsErrors = Partial<Record<keyof Settings, string>>;

interface Props {
  draft: Settings;
  errors: SettingsErrors;
  now: Date;
  onChange: (patch: Partial<Settings>) => void;
}

/** "Skip night photos" and the camera's location, which sunrise and sunset are computed from (offline). */
export function NightFields({ draft, errors, now, onChange }: Props) {
  const place = placeAt(draft.Latitude, draft.Longitude);
  // Coordinates the list doesn't have are shown in the coordinate fields.
  const [exact, setExact] = useState(!place && !noLocation(draft.Latitude, draft.Longitude));
  const skip = draft.SkipNightPhotos;
  const locationError = [errors.Latitude, errors.Longitude].filter(Boolean).join(' ');
  const choosePlace = (chosen: Place | undefined) => {
    setExact(false);
    onChange({ Latitude: chosen?.lat ?? 0, Longitude: chosen?.lon ?? 0 });
  };
  return (
    <div className="field">
      <Check label="Skip night photos" checked={skip} onChange={(on) => onChange({ SkipNightPhotos: on })} />
      {skip && (
        <LocationSelect place={place} exact={exact} invalid={!exact && !!locationError} onExact={() => setExact(true)} onPlace={choosePlace} />
      )}
      {skip && exact && <CoordinateFields draft={draft} errors={errors} onChange={onChange} />}
      <Help id="night-help" invalid={skip && !!locationError}>
        {nightHelp(draft, locationError, exact ? undefined : place, now)}
        {skip && exact && ' Location in decimal degrees: north and east positive, e.g. from a map app.'}
      </Help>
    </div>
  );
}

/** What skipping nights means at this location: the next sunrise and sunset, or what is wrong. */
function nightHelp(s: Settings, locationError: string, place: Place | undefined, now: Date): string {
  if (!s.SkipNightPhotos) return 'Photos are taken around the clock.';
  if (locationError) return locationError;
  if (polarNight(now, s.Latitude, s.Longitude)) {
    return 'The sun does not rise here within a day (polar night): photos are taken around the clock until it does.';
  }
  // The next sunrise and sunset, not today's: what happens next matters.
  const rise = nextSunrise(now, s.Latitude, s.Longitude);
  const set = nextSunset(now, s.Latitude, s.Longitude);
  if (!rise || !set) {
    return 'The sun does not set here for now: photos are taken around the clock.';
  }
  const middle = place ? `Uses the middle of ${place.name.replace(/^.* – /, '')}: elsewhere in it the times differ by a few minutes. ` : '';
  return `${middle}Here the sun rises at ${formatClock(rise)} and sets at ${formatClock(set)} (next times, in this device's time zone). No photos are taken in between. Computed on the camera, without internet.`;
}

interface LocationProps {
  place: Place | undefined;
  exact: boolean;
  invalid: boolean;
  onExact: () => void;
  onPlace: (place: Place | undefined) => void;
}

function LocationSelect({ place, exact, invalid, onExact, onPlace }: LocationProps) {
  const select = (value: string) => (value === 'exact' ? onExact() : onPlace(PLACES.find((p) => p.name === value)));
  return (
    <div className="coordinate location">
      <label htmlFor="location">Location</label>
      <select
        id="location"
        value={exact ? 'exact' : (place?.name ?? '')}
        aria-invalid={invalid}
        aria-describedby="night-help"
        onChange={(e) => select(e.target.value)}
      >
        {!place && !exact && <option value="">Choose a country…</option>}
        {PLACES.map((p) => (
          <option key={p.name} value={p.name}>
            {p.name}
          </option>
        ))}
        <option value="exact">Exact coordinates…</option>
      </select>
    </div>
  );
}

function CoordinateFields({ draft, errors, onChange }: Omit<Props, 'now'>) {
  const field = (name: 'Latitude' | 'Longitude', hint: string) => (
    <Coordinate name={name} hint={hint} value={draft[name]} invalid={!!errors[name]} onChange={(v) => onChange({ [name]: v })} />
  );
  return (
    <div className="inline coordinates">
      {field('Latitude', 'e.g. 51.5074')}
      {field('Longitude', 'e.g. -0.1278')}
    </div>
  );
}

type CoordinateProps = { name: string; hint: string; value: number; invalid: boolean; onChange: (v: number) => void };

/** Shows a coordinate as typed ("-", "-0.") while saving the number; empty means 0 (no location). */
function Coordinate({ name, hint, value, invalid, onChange }: CoordinateProps) {
  const [text, setText] = useState(coordinateText(value));
  // Follow changes from outside (Discard, saved settings).
  useEffect(() => {
    setText((typed) => (Object.is(parseCoordinate(typed), value) ? typed : coordinateText(value)));
  }, [value]);
  const id = name.toLowerCase();
  return (
    <div className="coordinate">
      <label htmlFor={id}>{name}</label>
      <input
        id={id}
        type="text"
        inputMode="decimal"
        placeholder={hint}
        value={text}
        aria-invalid={invalid}
        aria-describedby="night-help"
        onChange={(e) => {
          setText(e.target.value);
          onChange(parseCoordinate(e.target.value));
        }}
      />
    </div>
  );
}

const coordinateText = (value: number) => (value === 0 ? '' : String(value));

/** "" is 0, a number is itself, anything else (e.g. "-" while typing) NaN. */
function parseCoordinate(text: string): number {
  const t = text.trim().replace(',', '.');
  if (t === '') return 0;
  return /^[-+]?(\d+\.?\d*|\.\d+)$/.test(t) ? Number(t) : NaN;
}
