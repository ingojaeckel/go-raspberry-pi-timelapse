import type { Settings } from './api';

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

export function formatInterval(seconds: number): string {
  if (seconds % 3600 === 0) return seconds === 3600 ? '1 hour' : `${seconds / 3600} hours`;
  if (seconds % 60 === 0) return seconds === 60 ? '1 minute' : `${seconds / 60} minutes`;
  return `${seconds} seconds`;
}

export function isAligned(s: Schedule): boolean {
  return s.OffsetWithinHour >= 0;
}

function usable(s: Schedule): boolean {
  return isValidInterval(s.SecondsBetweenCaptures) && s.OffsetWithinHour < s.SecondsBetweenCaptures;
}

/**
 * Seconds until the next capture. A port of timelapse.NextAligned in
 * src/backend/timelapse/schedule.go; both are tested against
 * schedule-cases.json. Returns null when captures are not aligned to the hour
 * (they then run every N seconds since start-up) or the schedule is invalid.
 */
export function secondsToNextCapture(s: Schedule, now: Date): number | null {
  if (!isAligned(s) || !usable(s)) return null;
  const interval = s.SecondsBetweenCaptures;
  const sinceFirst = now.getMinutes() * 60 + now.getSeconds() - s.OffsetWithinHour;
  // offset < interval, so this is never more than one interval before the first slot.
  const k = Math.max(0, Math.ceil(sinceFirst / interval));
  return k * interval - sinceFirst;
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
