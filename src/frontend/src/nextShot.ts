import { useEffect, useRef } from 'react';
import { api, type CameraSchedule } from './api';
import { useInterval, useResource } from './hooks';

/** The camera's next capture, on this device's clock. */
export interface NextShot {
  at: Date;
  /** The slot was moved past the night (SkipNightPhotos). */
  night: boolean;
  /** The camera's time zone offset from UTC, for minutes past its hour. */
  zoneOffsetSeconds: number;
}

/**
 * Converts the camera's plan to this device's clock: the countdown stays
 * right even if the two clocks differ. Null before the camera planned a capture.
 */
export function toNextShot(s: CameraSchedule, receivedAt: number): NextShot | null {
  if (!s.NextMillis) return null;
  return { at: new Date(receivedAt + s.NextMillis - s.UnixMillis), night: s.Night, zoneOffsetSeconds: s.ZoneOffsetSeconds };
}

/** Whole seconds until the shot, 0 once it is due. */
export function secondsUntil(shot: NextShot, now: Date): number {
  return Math.max(0, Math.round((shot.at.getTime() - now.getTime()) / 1000));
}

/** Asked again this long after the planned time, until the camera has planned the next one. */
const PASSED_RETRY_MS = 5000;
const REFRESH_MS = 60_000;

/**
 * The camera's next capture (GET /schedule). Asks again when `settings`
 * change, every minute, and every few seconds once the planned time has passed.
 */
export function useNextShot(settings: unknown, now: Date): NextShot | null {
  const r = useResource((signal) => api.schedule(signal).then((s) => toNextShot(s, Date.now())));
  const { reload } = r;
  const seen = useRef(settings);
  useEffect(() => {
    if (seen.current === settings) return;
    if (seen.current !== undefined) reload(); // not when the settings first load
    seen.current = settings;
  }, [settings, reload]);
  const passed = !!r.data && now.getTime() > r.data.at.getTime() + PASSED_RETRY_MS;
  useInterval(reload, passed ? PASSED_RETRY_MS : REFRESH_MS);
  return r.data ?? null;
}
