// Parsers for the raw strings the Pi reports, and display formatting.

import type { Storage } from './api';

/** Parses Go's time.Time.String() output, e.g. "2026-11-02 21:15:00.123456789 +0100 CET". */
export function parseGoTime(raw: string): Date | null {
  const m = /^(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2})(\.\d+)? ([+-]\d{2})(\d{2})/.exec(raw.trim());
  if (!m) return null;
  const frac = m[3] ? m[3].slice(0, 4) : '';
  const d = new Date(`${m[1]}T${m[2]}${frac}${m[4]}:${m[5]}`);
  return Number.isNaN(d.getTime()) ? null : d;
}

export function dayKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

export function formatDay(d: Date, now = new Date()): string {
  const today = dayKey(now);
  const y = new Date(now);
  y.setDate(now.getDate() - 1);
  const k = dayKey(d);
  if (k === today) return 'Today';
  if (k === dayKey(y)) return 'Yesterday';
  return d.toLocaleDateString(undefined, {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    year: d.getFullYear() === now.getFullYear() ? undefined : 'numeric',
  });
}

export function formatClock(d: Date): string {
  return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
}

/** "4:05", "1:02:03" */
export function formatCountdown(totalSeconds: number): string {
  const s = Math.max(0, Math.round(totalSeconds));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  const mm = h > 0 ? String(m).padStart(2, '0') : String(m);
  return `${h > 0 ? `${h}:` : ''}${mm}:${String(sec).padStart(2, '0')}`;
}

/** "45 s", "30 min", "1 h 30 min" */
export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds} s`;
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  const parts: string[] = [];
  if (h) parts.push(`${h} h`);
  if (m) parts.push(`${m} min`);
  if (s) parts.push(`${s} s`);
  return parts.join(' ');
}

export function formatRelative(from: Date, now = new Date()): string {
  const s = Math.round((now.getTime() - from.getTime()) / 1000);
  if (s < 45) return 'just now';
  if (s < 3600) return `${Math.round(s / 60)} min ago`;
  if (s < 86400) return `${Math.round(s / 3600)} h ago`;
  const d = Math.round(s / 86400);
  return `${d} day${d === 1 ? '' : 's'} ago`;
}

/** ":15", ":07:30" — a time past the hour. */
export function formatPastHour(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `:${String(m).padStart(2, '0')}${s ? `:${String(s).padStart(2, '0')}` : ''}`;
}

const UNIT: Record<string, number> = { B: 1, K: 1024, M: 1024 ** 2, G: 1024 ** 3, T: 1024 ** 4, P: 1024 ** 5 };

/** Parses "2 MB", "800 KB", "12 bytes" (backend) and "12G", "3.5M", "980Ki" (df -h). */
export function parseSize(raw: string): number | null {
  const m = /^\s*([\d.,]+)\s*(bytes?|[BKMGTP])?(i?B?)?\s*$/i.exec(raw);
  if (!m || !m[1]) return null;
  const n = Number.parseFloat(m[1].replace(',', '.'));
  if (Number.isNaN(n)) return null;
  const unit = (m[2] ?? 'B').toUpperCase().startsWith('BYTE') ? 'B' : (m[2] ?? 'B').toUpperCase();
  return n * (UNIT[unit] ?? 1);
}

export function formatBytes(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let v = bytes;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v >= 10 || i === 0 ? Math.round(v) : v.toFixed(1)} ${units[i]}`;
}

/** /sys/class/thermal/thermal_zone0/temp reports millidegrees, e.g. "48312". */
export function parseCpuTemp(raw: string): number | null {
  const t = raw.trim();
  if (!/^-?\d+$/.test(t)) return null;
  return Number(t) / 1000;
}

/** vcgencmd measure_temp reports "temp=48.3'C". */
export function parseGpuTemp(raw: string): number | null {
  const m = /temp=(-?[\d.]+)/.exec(raw);
  return m?.[1] ? Number(m[1]) : null;
}

export interface Uptime {
  up: string | null;
  load: [number, number, number] | null;
}

/** " 21:39:00 up 3 days,  2:01,  1 user,  load average: 0.08, 0.10, 0.09" */
export function parseUptime(raw: string): Uptime {
  const up = /\bup\s+(.+?),\s+\d+\s+users?/.exec(raw)?.[1] ?? /\bup\s+(.+?),\s+load/.exec(raw)?.[1] ?? null;
  const l = /load averages?:\s*([\d.]+),?\s+([\d.]+),?\s+([\d.]+)/.exec(raw);
  return {
    up: up ? up.replace(/\s+/g, ' ').trim() : null,
    load: l ? [Number(l[1]), Number(l[2]), Number(l[3])] : null,
  };
}

export interface Disk {
  filesystem: string;
  size: string;
  used: string;
  avail: string;
  usedPercent: number;
  mount: string;
}

/** Parses `df -h` output (GNU or BSD columns). Keeps only block devices. */
export function parseDf(raw: string): Disk[] {
  const disks: Disk[] = [];
  for (const line of raw.split('\n').slice(1)) {
    const tok = line.trim().split(/\s+/);
    if (tok.length < 6) continue;
    const pctIdx = tok.findIndex((t, i) => i >= 4 && /^\d+%$/.test(t));
    if (pctIdx < 0) continue;
    const lastPct = tok.reduce((acc, t, i) => (/^(\d+|-)%$/.test(t) ? i : acc), pctIdx);
    const [filesystem = '', size = '', used = '', avail = ''] = tok;
    disks.push({
      filesystem,
      size,
      used,
      avail,
      usedPercent: Number.parseInt(tok[pctIdx] ?? '0', 10),
      mount: tok.slice(lastPct + 1).join(' '),
    });
  }
  return disks.filter((d) => d.filesystem.startsWith('/') && d.mount !== '');
}

/** The disk that most likely holds the photos: "/" if present, otherwise the largest. */
export function primaryDisk(disks: Disk[]): Disk | null {
  return (
    disks.find((d) => d.mount === '/') ??
    [...disks].sort((a, b) => (parseSize(b.size) ?? 0) - (parseSize(a.size) ?? 0))[0] ??
    null
  );
}

export interface ClockCheck {
  /** Camera clock minus this device's clock. */
  driftSeconds: number;
  /** Year the camera believes it is, when it differs from this device's. */
  wrongYear: number | null;
}

/**
 * Compares the camera's clock with this device's. A Pi Zero has no
 * battery-backed clock, and offline it can't sync: after a power cut its
 * clock is behind by however long the power was out.
 */
export function checkClock(cameraMillis: number, now: Date): ClockCheck {
  const year = new Date(cameraMillis).getFullYear();
  return {
    driftSeconds: Math.round((cameraMillis - now.getTime()) / 1000),
    wrongYear: year !== now.getFullYear() ? year : null,
  };
}

export interface Runway {
  /** Photos that fit before the disk counts as full. */
  photos: number;
  /** Days until then at `perDay` photos a day, or null without a schedule. */
  days: number | null;
}

/** How long until the card is full, from the average size of the photos so far. */
export function estimateRunway(s: Storage, perDay: number): Runway | null {
  if (s.Error || s.Photos === 0 || s.PhotoBytes === 0) return null;
  const photos = Math.floor(Math.max(s.FreeBytes - s.ReserveBytes, 0) / (s.PhotoBytes / s.Photos));
  return { photos, days: perDay > 0 ? photos / perDay : null };
}

/**
 * JPEG size relative to quality 90, measured on Raspberry Pi camera photos
 * (libjpeg's quality scale, which rpicam-still uses). Sizes in between are
 * interpolated.
 */
const QUALITY_SIZE: readonly (readonly [number, number])[] = [
  [1, 0.12],
  [10, 0.17],
  [20, 0.24],
  [30, 0.29],
  [40, 0.34],
  [50, 0.39],
  [60, 0.43],
  [70, 0.52],
  [75, 0.57],
  [80, 0.65],
  [85, 0.78],
  [90, 1],
  [95, 1.5],
  [100, 3],
];

/** How much larger a photo is at `quality` than at quality 90. */
export function qualitySizeFactor(quality: number): number {
  if (!Number.isFinite(quality)) return 1;
  const q = Math.min(100, Math.max(1, quality));
  for (let i = 1; i < QUALITY_SIZE.length; i++) {
    const [q1, f1] = QUALITY_SIZE[i]!;
    if (q <= q1) {
      const [q0, f0] = QUALITY_SIZE[i - 1]!;
      return f0 + ((q - q0) / (q1 - q0)) * (f1 - f0);
    }
  }
  return QUALITY_SIZE[QUALITY_SIZE.length - 1]![1];
}

/** Bytes per pixel of a typical daylight photo at quality 90, until the camera has taken photos. */
const TYPICAL_BYTES_PER_PIXEL_Q90 = 0.2;

export interface PhotoSizeEstimate {
  bytes: number;
  /** The number of recent photos the estimate is based on; 0 for a typical photo. */
  fromPhotos: number;
}

/**
 * Estimates the size of one photo at a size and quality from the recent
 * photos' size per pixel (the scene matters more than anything else), or a
 * typical photo before there are any.
 */
export function estimatePhotoBytes(width: number, height: number, quality: number, s: Storage | undefined): PhotoSizeEstimate {
  const sampled = s && !s.Error && s.SamplePhotos > 0 && s.SampleBytesPerPixel > 0;
  const perPixelQ90 = sampled ? s.SampleBytesPerPixel / qualitySizeFactor(s.SampleQuality) : TYPICAL_BYTES_PER_PIXEL_Q90;
  return { bytes: width * height * perPixelQ90 * qualitySizeFactor(quality), fromPhotos: sampled ? s.SamplePhotos : 0 };
}

export interface Capacity {
  /** Photos that fit in the free space before the card counts as full. */
  more: number;
  /** Days until then at the schedule, or null without one. */
  days: number | null;
  /** Deleting the oldest photos: how many photos the card holds in all, and how many days that covers. Null when the camera stops instead. */
  kept: number | null;
  keptDays: number | null;
}

/** How many photos of `photoBytes` fit on the card, and for how long at `perDay` photos a day. */
export function estimateCapacity(s: Storage, photoBytes: number, perDay: number): Capacity | null {
  if (s.Error || !(photoBytes > 0)) return null;
  const usable = Math.max(s.FreeBytes - s.ReserveBytes, 0);
  const more = Math.floor(usable / photoBytes);
  const kept = s.StopWhenDiskFull ? null : Math.floor((usable + s.PhotoBytes) / photoBytes);
  const days = (n: number) => (perDay > 0 ? n / perDay : null);
  return { more, days: days(more), kept, keptDays: kept === null ? null : days(kept) };
}

/** "8.1 MP"; whole numbers from 10 MP. */
export function formatMegapixels(width: number, height: number): string {
  const mp = (width * height) / 1e6;
  return `${mp >= 10 ? Math.round(mp) : +mp.toFixed(1)} MP`;
}

/** "16:9", "4:3", or "1.88:1" for unusual shapes. */
export function aspectRatio(width: number, height: number): string {
  const ratio = width / height;
  for (const [w, h] of [[1, 1], [4, 3], [3, 2], [16, 9], [3, 4], [9, 16]] as const) {
    if (Math.abs(ratio - w / h) < 0.01) return `${w}:${h}`;
  }
  return `${+ratio.toFixed(2)}:1`;
}

/** A stretch of days for estimates: "584 days", "over 3 years", "5 hours". */
export function formatDays(days: number): string {
  if (days >= 730) return `over ${Math.floor(days / 365)} years`;
  if (days >= 2) return `${Math.floor(days).toLocaleString()} days`;
  return `${Math.max(1, Math.round(days * 24))} hours`;
}
