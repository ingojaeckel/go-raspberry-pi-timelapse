import type { Settings } from './api';
import { daylightFraction, noLocation, polarNight, validLocation } from './sun';

type Schedule = Pick<Settings, 'SecondsBetweenCaptures' | 'OffsetWithinHour'>;

const MIN_INTERVAL = 60;
const MAX_INTERVAL = 3600;

/**
 * An interval is valid when it is between a minute and an hour and divides the
 * hour evenly, so every hour has the same capture times. Mirrors
 * conf.ValidInterval in src/backend/conf/settings.go.
 */
export function isValidInterval(seconds: number): boolean {
  return Number.isInteger(seconds) && seconds >= MIN_INTERVAL && seconds <= MAX_INTERVAL && 3600 % seconds === 0;
}

/** Every valid interval in ascending order. */
export const VALID_INTERVALS: readonly number[] = Array.from(
  { length: MAX_INTERVAL - MIN_INTERVAL + 1 },
  (_, i) => MIN_INTERVAL + i,
).filter(isValidInterval);

/** The intervals Settings offers: whole minutes (the camera accepts any valid interval). */
export const OFFERED_INTERVALS: readonly number[] = VALID_INTERVALS.filter((seconds) => seconds % 60 === 0);

/** An interval as "1 min 12 s" or "1 hour": one unit throughout, so a list of them reads in order. */
export function formatInterval(seconds: number): string {
  if (seconds % 3600 === 0) return seconds === 3600 ? '1 hour' : `${seconds / 3600} hours`;
  const rest = seconds % 60;
  return `${Math.floor(seconds / 60)} min${rest ? ` ${rest} s` : ''}`;
}

export function isAligned(s: Schedule): boolean {
  return s.OffsetWithinHour >= 0;
}

function usable(s: Schedule): boolean {
  return isValidInterval(s.SecondsBetweenCaptures) && s.OffsetWithinHour < s.SecondsBetweenCaptures;
}

/** Capture times as seconds past the hour (0..3599), sorted. Null when not aligned or invalid. */
export function captureSlots(s: Schedule): number[] | null {
  if (!isAligned(s) || !usable(s)) return null;
  const slots: number[] = [];
  for (let t = s.OffsetWithinHour; t < 3600; t += s.SecondsBetweenCaptures) slots.push(t);
  return slots;
}

/** Photos per day at the current schedule; 0 when the schedule is invalid. */
export function capturesPerDay(s: Schedule): number {
  if (!isValidInterval(s.SecondsBetweenCaptures)) return 0;
  return 86400 / s.SecondsBetweenCaptures;
}

type NightSchedule = Schedule & Partial<Pick<Settings, 'SkipNightPhotos' | 'Latitude' | 'Longitude'>>;

/** Whether the camera skips nights with these settings: asked to, a valid location, and not the polar night. */
export function skipsNights(s: NightSchedule, now: Date): boolean {
  const { SkipNightPhotos, Latitude: lat = 0, Longitude: lon = 0 } = s;
  return !!SkipNightPhotos && validLocation(lat, lon) && !noLocation(lat, lon) && !polarNight(now, lat, lon);
}

/** Photos a day: when nights are skipped, only the share of the coming 24 hours with the sun up. */
export function photosPerDay(s: NightSchedule, now: Date): number {
  const all = capturesPerDay(s);
  if (!skipsNights(s, now)) return all;
  return Math.round(all * daylightFraction(now, s.Latitude ?? 0, s.Longitude ?? 0));
}
