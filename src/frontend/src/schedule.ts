import type { Settings } from './api';

type Schedule = Pick<Settings, 'SecondsBetweenCaptures' | 'OffsetWithinHour'>;

export function isAligned(s: Schedule): boolean {
  return s.OffsetWithinHour >= 0;
}

/**
 * Seconds until the next capture. A direct port of
 * Timelapse.getSecondsToFirstCapture in timelapse/timelapse.go so the UI
 * predicts exactly what the camera will do. Returns null when captures are
 * not aligned to the hour (they then run every N seconds since start-up).
 */
export function secondsToNextCapture(s: Schedule, now: Date): number | null {
  if (!isAligned(s) || s.SecondsBetweenCaptures < 1) return null;
  const interval = s.SecondsBetweenCaptures;
  const offset = s.OffsetWithinHour;
  const picturesPerHour = Math.floor(3600 / interval);
  const into = now.getMinutes() * 60 + now.getSeconds();

  for (let i = 0; i < picturesPerHour; i++) {
    if (i === 0 && into >= 0 && into <= offset) return offset - into;
    const lower = offset + (i - 1) * interval;
    const upper = offset + i * interval;
    if (lower <= into && into <= upper) return upper - into;
  }
  return 3600 - into + offset;
}

/** Capture times as seconds past the hour (0..3599), sorted. Null when not aligned. */
export function captureSlots(s: Schedule): number[] | null {
  if (!isAligned(s) || s.SecondsBetweenCaptures < 1) return null;
  const perHour = Math.floor(3600 / s.SecondsBetweenCaptures);
  if (perHour === 0) return [s.OffsetWithinHour % 3600];
  const slots = new Set<number>();
  for (let i = 0; i < perHour; i++) {
    slots.add((s.OffsetWithinHour + i * s.SecondsBetweenCaptures) % 3600);
  }
  return [...slots].sort((a, b) => a - b);
}

/** Photos per day at the current schedule. */
export function capturesPerDay(s: Schedule): number {
  if (s.SecondsBetweenCaptures < 1) return 0;
  if (!isAligned(s)) return Math.floor(86400 / s.SecondsBetweenCaptures);
  return (captureSlots(s)?.length ?? 0) * 24;
}
