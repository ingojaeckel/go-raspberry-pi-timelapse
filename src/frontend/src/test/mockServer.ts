import { vi } from 'vitest';
import type { Camera, Monitoring, Photo, Settings } from '../api';

export const settings: Settings = {
  SecondsBetweenCaptures: 1800,
  OffsetWithinHour: 900,
  PhotoResolutionWidth: 1640,
  PhotoResolutionHeight: 1232,
  RotateBy: 0,
  Quality: 100,
  DebugEnabled: false,
  StopWhenDiskFull: false,
  SkipNightPhotos: false,
  Latitude: 0,
  Longitude: 0,
};

/** A Camera Module v2 (IMX219) and its modes. */
export const camera: Camera = {
  Model: 'imx219',
  Width: 3280,
  Height: 2464,
  Resolutions: [
    { Width: 3280, Height: 2464 },
    { Width: 1920, Height: 1080 },
    { Width: 1640, Height: 1232 },
    { Width: 640, Height: 480 },
  ],
};

export const photos: Photo[] = [
  { Name: '20261101-081500.jpg', ModTime: '2026-11-01 08:15:00 +0000 UTC', Size: '2 MB' },
  { Name: '20261101-084500.jpg', ModTime: '2026-11-01 08:45:00 +0000 UTC', Size: '2 MB' },
  { Name: '20261102-091500.jpg', ModTime: '2026-11-02 09:15:00 +0000 UTC', Size: '3 MB' },
];

export const monitoring: Monitoring = {
  Time: 'Mon Nov  2 10:00:00 UTC 2026\n',
  Uptime: ' 10:00:00 up 3 days,  2:01,  1 user,  load average: 0.08, 0.10, 0.09\n',
  CpuTemperature: '48312\n',
  GpuTemperature: "temp=51.0'C\n",
  FreeDiskSpace:
    'Filesystem      Size  Used Avail Use% Mounted on\n/dev/root        29G  4.1G   24G  15% /\n/dev/mmcblk0p1  253M   51M  202M  20% /boot/firmware\n',
  // Same as the browser's clock unless a test changes it (Date.now() at load).
  get UnixMillis() {
    return Date.now();
  },
  Power: { Now: [], SinceBoot: [] },
  CPU: { Governor: 'powersave', CurrentMHz: 600, MinMHz: 600, MaxMHz: 1000 },
  Storage: {
    FreeBytes: 24 * 1024 ** 3,
    ReserveBytes: 256 * 1024 ** 2,
    Photos: 3,
    PhotoBytes: 7 * 1024 ** 2,
    StopWhenDiskFull: false,
    DeletedOldest: 0,
    // 0.45 bytes per pixel at quality 100: about 0.9 MB for 1640x1232
    SampleBytesPerPixel: 0.45,
    SampleQuality: 100,
    SamplePhotos: 3,
  },
};

type Handler = (url: URL, init?: RequestInit) => Response | Promise<Response>;

export const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });

/** A GET /schedule response: the next shot `inMs` after the camera's clock `cameraNow`. */
export const schedule = (inMs: number, { cameraNow = Date.now(), night = false, zoneOffsetSeconds = 0 } = {}) =>
  json({ UnixMillis: cameraNow, NextMillis: cameraNow + inMs, Night: night, ZoneOffsetSeconds: zoneOffsetSeconds });

/** A route override for GET /monitoring with some fields changed. */
export const withMonitoring = (patch: Partial<Monitoring>) => ({ 'GET /monitoring': () => json({ ...monitoring, ...patch }) });

/** A route override for GET /monitoring with some storage fields changed. */
export const withStorage = (patch: Partial<Monitoring['Storage']>) => withMonitoring({ Storage: { ...monitoring.Storage, ...patch } });

export type Routes = Record<string, Handler>;

/** Stubs fetch with a tiny fake of the Go backend. Override routes per test. */
export function mockServer(overrides: Routes = {}) {
  const routes: Routes = {
    'GET /configuration': () => json(settings),
    'POST /configuration': (_u, init) => json(JSON.parse(String(init?.body))),
    'GET /photos': () => json({ Photos: photos }),
    'GET /monitoring': () => json(monitoring),
    'GET /camera': () => json(camera),
    'GET /schedule': () => schedule(600_000),
    'GET /logs': () => json({ Logs: '2026/11/02 10:00:00 started\n2026/11/02 10:15:00 Error during capture: boom\n' }),
    'GET /version': () => new Response('Hello from arm on linux [version:abc]'),
    'DELETE /file': () => json(1),
    'DELETE /photos': () => json(3),
    'POST /admin/restart': () => new Response(null, { status: 204 }),
    'POST /admin/shutdown': () => new Response(null, { status: 204 }),
    'POST /admin/time': (_u, init) => json({ UnixMillis: JSON.parse(String(init?.body)).UnixMillis, RTC: '' }),
    ...overrides,
  };
  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = new URL(String(input), 'http://camera.local');
    const key = `${init?.method ?? 'GET'} ${url.pathname}`;
    const handler = routes[key];
    if (!handler) return new Response('not found', { status: 404 });
    return handler(url, init);
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

export function callsTo(fetchMock: ReturnType<typeof mockServer>, method: string, path: string) {
  return fetchMock.mock.calls.filter(([input, init]) => {
    const url = new URL(String(input), 'http://camera.local');
    return (init?.method ?? 'GET') === method && url.pathname === path;
  });
}
