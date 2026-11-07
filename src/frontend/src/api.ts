// Typed client for the Go backend. Endpoints and payload shapes mirror
// rest/model.go and conf/model.go exactly; do not rename fields.

/**
 * Empty in production: the UI is served by the same Go server it talks to.
 * The server sends no CORS headers, so a different origin only works through
 * a proxy (`npm start` proxies to VITE_BACKEND).
 */
export const API_BASE: string = import.meta.env.VITE_API_BASE ?? '';

export interface Settings {
  /** Between 60 and 3600 and divides an hour evenly (see schedule.ts). */
  SecondsBetweenCaptures: number;
  /** Seconds into each hour of the first capture (below SecondsBetweenCaptures), or -1 to disable alignment. */
  OffsetWithinHour: number;
  PhotoResolutionWidth: number;
  PhotoResolutionHeight: number;
  PreviewResolutionWidth: number;
  PreviewResolutionHeight: number;
  /** 0 or 180. */
  RotateBy: number;
  /** 0: 3280x2464, 1: 2186x1642, 2: 1640x1232 (4:3; a 16:9 sensor is cropped to 4:3) */
  ResolutionSetting: number;
  Quality: number;
  DebugEnabled: boolean;
  /** When the card is full: true stops taking photos, false deletes the oldest ones. */
  StopWhenDiskFull: boolean;
}

/** Command output, or the error message if the command failed. */
export interface MonitoringOutput {
  Time: string;
  Uptime: string;
  CpuTemperature: string;
  GpuTemperature: string;
  FreeDiskSpace: string;
}

export interface Monitoring extends MonitoringOutput {
  /** The camera's clock. */
  UnixMillis: number;
  /** Under-voltage and throttling reported by the firmware. */
  Power: { Now: string[]; SinceBoot: string[]; Error?: string };
  CPU: { Governor: string; CurrentMHz: number; MinMHz: number; MaxMHz: number; Error?: string };
  Storage: Storage;
}

/** The photo folder's disk. */
export interface Storage {
  FreeBytes: number;
  /** Kept free: the disk counts as full below this. */
  ReserveBytes: number;
  Photos: number;
  PhotoBytes: number;
  StopWhenDiskFull: boolean;
  /** Photos deleted to make room since the camera started. */
  DeletedOldest: number;
  Error?: string;
}

export interface SetTimeResult {
  UnixMillis: number;
  /** Empty without a clock module, "updated", or why it was not. */
  RTC: string;
}

export interface Photo {
  Name: string;
  /** Go time.Time.String(), e.g. "2026-11-02 21:15:00.123 +0000 UTC" */
  ModTime: string;
  /** Human readable, e.g. "2 MB" */
  Size: string;
}

interface PhotosResponse {
  Photos: Photo[] | null;
}

interface LogResponse {
  Logs: string;
}

export type AdminCommand = 'restart' | 'shutdown' | 'clear';

export class ApiError extends Error {
  readonly status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

export const OFFLINE_MESSAGE =
  "Can't reach the camera. Check that this device is still connected to the camera's WiFi.";

async function request(path: string, init?: RequestInit): Promise<Response> {
  let res: Response;
  try {
    res = await fetch(API_BASE + path, init);
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') throw err;
    throw new ApiError(OFFLINE_MESSAGE, 0);
  }
  if (!res.ok) {
    throw new ApiError(await errorText(res), res.status);
  }
  return res;
}

/** The backend reports errors as a JSON-encoded string, or as plain text. */
async function errorText(res: Response): Promise<string> {
  const body = (await res.text().catch(() => '')).trim();
  if (!body) return `The camera answered with HTTP ${res.status}.`;
  try {
    const parsed: unknown = JSON.parse(body);
    if (typeof parsed === 'string') return parsed;
  } catch {
    /* not JSON */
  }
  return body;
}

async function json<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await request(path, init);
  return (await res.json()) as T;
}

function filesQuery(names: readonly string[]): string {
  const q = new URLSearchParams();
  for (const n of names) q.append('f', n);
  const s = q.toString();
  return s ? `?${s}` : '';
}

export const api = {
  async photos(signal?: AbortSignal): Promise<Photo[]> {
    const r = await json<PhotosResponse>('/photos', { signal });
    return r.Photos ?? [];
  },
  settings(signal?: AbortSignal): Promise<Settings> {
    return json<Settings>('/configuration', { signal });
  },
  saveSettings(s: Settings): Promise<Settings> {
    return json<Settings>('/configuration', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(s),
    });
  },
  monitoring(signal?: AbortSignal): Promise<Monitoring> {
    return json<Monitoring>('/monitoring', { signal });
  },
  /** Sets the camera's clock, normally to this device's. */
  setTime(unixMillis: number): Promise<SetTimeResult> {
    return json<SetTimeResult>('/admin/time', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ UnixMillis: unixMillis }),
    });
  },
  async logs(signal?: AbortSignal): Promise<string> {
    const r = await json<LogResponse>('/logs', { signal });
    return r.Logs ?? '';
  },
  async version(signal?: AbortSignal): Promise<string> {
    const res = await request('/version', { signal });
    return res.text();
  },
  async deletePhotos(names: readonly string[]): Promise<void> {
    if (names.length === 0) return;
    await request('/file' + filesQuery(names), { method: 'DELETE' });
  },
  async admin(command: AdminCommand): Promise<void> {
    if (command === 'clear') {
      await request('/photos', { method: 'DELETE' });
    } else {
      await request('/admin/' + command, { method: 'POST' });
    }
  },
};

export const urls = {
  photo: (name: string) => `${API_BASE}/file/${encodeURIComponent(name)}`,
  /** `key` busts the browser cache whenever the newest photo changes. */
  latestPhoto: (key: string) => `${API_BASE}/file/last?v=${encodeURIComponent(key)}`,
  preview: (key: number) => `${API_BASE}/capture?t=${key}`,
  archive: (format: 'zip' | 'tar', names: readonly string[] = []) =>
    `${API_BASE}/archive/${format}${filesQuery(names)}`,
};

export const RESOLUTIONS = [
  { setting: 0, width: 3280, height: 2464, label: '8 MP' },
  { setting: 1, width: 2186, height: 1642, label: '3.5 MP' },
  { setting: 2, width: 1640, height: 1232, label: '2 MP' },
] as const;
