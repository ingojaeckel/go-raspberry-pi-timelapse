import type { ReactNode } from 'react';
import type { Camera, Settings, Size, Storage } from '../api';
import { ChoiceField, Help } from '../components/fields';
import {
  aspectRatio,
  estimateCapacity,
  estimatePhotoBytes,
  formatBytes,
  formatDays,
  formatMegapixels,
  formatSize,
  plural,
  type Capacity,
} from '../format';
import type { Resource } from '../hooks';
import { photosPerDay, skipsNights } from '../schedule';

const sameSize = (a: Size, b: Size) => a.Width === b.Width && a.Height === b.Height;

/** How a photo size relates to the camera: one of its own sizes, scaled down from one, or enlarged beyond the sensor. */
type Fit = 'native' | 'scaled' | 'too large';

function fitOf(size: Size, camera: Camera | null): Fit {
  if (!camera || camera.Resolutions?.some((r) => sameSize(r, size))) return 'native';
  return size.Width > camera.Width || size.Height > camera.Height ? 'too large' : 'scaled';
}

/** The camera, if it was detected and reported its sizes. */
function detectedCamera(camera: Camera | undefined): Camera | null {
  return camera && !camera.Error && camera.Resolutions?.length ? camera : null;
}

interface ResolutionProps {
  camera: Resource<Camera>;
  value: Size;
  saved: Size | undefined;
  onChange: (s: Size) => void;
}

/**
 * The connected camera's photo sizes (its sensor modes), so only sizes it
 * supports can be chosen. A saved size it lacks (e.g. after swapping the
 * camera) stays selectable until another one is chosen.
 */
export function ResolutionField({ camera, value, saved, onChange }: ResolutionProps) {
  const detected = detectedCamera(camera.data);
  const sizes = [...(detected?.Resolutions ?? [])];
  for (const extra of [saved, value]) {
    if (extra && !sizes.some((s) => sameSize(s, extra))) sizes.push(extra);
  }
  const choices = sizes.map((size) => {
    const fit = fitOf(size, detected);
    return {
      key: formatSize(size),
      title: formatMegapixels(size.Width, size.Height),
      detail: `${formatSize(size)} · ${fit === 'native' ? aspectRatio(size.Width, size.Height) : fit}`,
      checked: sameSize(value, size),
      onSelect: () => onChange(size),
    };
  });
  return <ChoiceField id="res-label" label="Resolution" name="resolution" choices={choices} help={resolutionHelp(camera, detected, value)} />;
}

function resolutionHelp(camera: Resource<Camera>, detected: Camera | null, value: Size): ReactNode {
  if (!detected) {
    if (camera.loading) return 'Detecting the camera…';
    return `Couldn't detect the camera (${camera.data?.Error ?? camera.error ?? 'unknown error'}), so only the current size is shown.`;
  }
  const sensor = formatSize(detected);
  const largest = formatSize(detected.Resolutions?.[0] ?? detected);
  const fit = fitOf(value, detected);
  return (
    <>
      Sizes of the {detected.Model} camera ({sensor} sensor). Lower resolutions fit more photos on the card.
      {fit === 'too large' && (
        <>
          {' '}
          <strong>{`${formatSize(value)} is larger than the camera can take (${sensor}), so photos are enlarged: bigger files, no extra detail. Choose ${largest}.`}</strong>
        </>
      )}
      {fit === 'scaled' && ` ${formatSize(value)} isn't one of the camera's own sizes, so photos are scaled to it.`}
    </>
  );
}

interface QualityProps {
  draft: Settings;
  error: string | undefined;
  storage: Storage | undefined;
  onChange: (quality: number) => void;
}

/** JPEG quality as a slider and a number, with the resulting photo size and card capacity. */
export function QualityField({ draft, error, storage, onChange }: QualityProps) {
  const quality = Number.isFinite(draft.Quality) ? draft.Quality : undefined;
  const range = { min: 1, max: 100, onChange: (e: { target: HTMLInputElement }) => onChange(e.target.valueAsNumber) };
  return (
    <div className="field">
      <label htmlFor="quality">JPEG quality</label>
      <div className="inline">
        <input id="quality-range" type="range" {...range} value={quality ?? 1} aria-label="JPEG quality slider" />
        <input
          id="quality"
          className="narrow"
          type="number"
          inputMode="numeric"
          {...range}
          value={quality ?? ''}
          aria-invalid={!!error}
          aria-describedby="quality-help"
        />
      </div>
      <Help id="quality-help" invalid={!!error}>
        {error ?? 'Lower values make smaller files. Few people can tell 85 from 100.'}
      </Help>
      {!error && <StorageEstimate draft={draft} storage={storage} />}
    </div>
  );
}

/** Photo size at the chosen size and quality, and how long the card lasts at the chosen schedule. */
function StorageEstimate({ draft, storage }: { draft: Settings; storage: Storage | undefined }) {
  const photo = estimatePhotoBytes(draft.PhotoResolutionWidth, draft.PhotoResolutionHeight, draft.Quality, storage);
  const perDay = photosPerDay(draft, new Date());
  const capacity = storage && estimateCapacity(storage, photo.bytes, perDay);
  const basis =
    photo.fromPhotos > 0 ? `estimated from the last ${plural(photo.fromPhotos, 'photo')}` : 'a rough estimate until the camera has taken photos';
  return (
    <p className="help estimate">
      About {formatBytes(photo.bytes)} per photo, {basis}.{capacity && ` ${capacityText(capacity, perDay, skipsNights(draft, new Date()))}`}
    </p>
  );
}

function capacityText(c: Capacity, perDay: number, nightsSkipped: boolean): string {
  const schedule =
    c.days === null ? '' : `: ${formatDays(c.days)} at ${perDay.toLocaleString()} photos a day${nightsSkipped ? ' (nights skipped)' : ''}`;
  const then =
    c.kept === null
      ? 'the camera stops taking photos.'
      : `the oldest photos are deleted, so the card keeps about the last ${c.kept.toLocaleString()} photos${
          c.keptDays === null ? '' : ` (${formatDays(c.keptDays)})`
        }.`;
  return `The card has room for about ${c.more.toLocaleString()} more${schedule}, then ${then}`;
}

const ROTATIONS: Record<number, [title: string, detail: string]> = {
  0: ['Upright', 'No rotation'],
  180: ['Upside down', 'Rotate 180°'],
};

/** Upright or upside down; another saved angle stays shown until changed. */
export function RotationField({ rotateBy, onChange }: { rotateBy: number; onChange: (degrees: number) => void }) {
  const angles = [0, 180, ...(rotateBy in ROTATIONS ? [] : [rotateBy])];
  const choices = angles.map((degrees) => {
    const [title, detail] = ROTATIONS[degrees] ?? [`${degrees}°`, 'Not supported by the camera'];
    return { key: degrees, title, detail, checked: rotateBy === degrees, onSelect: () => onChange(degrees) };
  });
  return (
    <ChoiceField id="rot-label" label="Orientation" name="rotation" choices={choices} help="Choose upside down if the camera is mounted hanging from its cable." />
  );
}
