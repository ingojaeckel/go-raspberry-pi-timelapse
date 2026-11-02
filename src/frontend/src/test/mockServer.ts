import { vi } from 'vitest';
import type { Monitoring, Photo, Settings } from '../api';

export const settings: Settings = {
  SecondsBetweenCaptures: 1800,
  OffsetWithinHour: 900,
  PhotoResolutionWidth: 2186,
  PhotoResolutionHeight: 1642,
  PreviewResolutionWidth: 640,
  PreviewResolutionHeight: 480,
  RotateBy: 0,
  ResolutionSetting: 1,
  Quality: 100,
  DebugEnabled: false,
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
};

type Handler = (url: URL, init?: RequestInit) => Response | Promise<Response>;

export const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });

/** Stubs fetch with a tiny fake of the Go backend. Override routes per test. */
export function mockServer(overrides: Record<string, Handler> = {}) {
  const routes: Record<string, Handler> = {
    'GET /configuration': () => json(settings),
    'POST /configuration': (_u, init) => json(JSON.parse(String(init?.body))),
    'GET /photos': () => json({ Photos: photos }),
    'GET /monitoring': () => json(monitoring),
    'GET /logs': () => json({ Logs: '2026/11/02 10:00:00 started\n2026/11/02 10:15:00 Error during capture: boom\n' }),
    'GET /version': () => new Response('Hello from arm on linux [version:abc]'),
    'GET /file/delete': () => json(1),
    'GET /admin/restart': () => new Response(''),
    'GET /admin/shutdown': () => new Response(''),
    'GET /admin/clear': () => new Response(''),
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
