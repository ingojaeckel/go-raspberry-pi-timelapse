import { act, renderHook } from '@testing-library/react';
import { useNextShot } from './nextShot';
import { callsTo, json, mockServer, settings } from './test/mockServer';

const start = new Date('2026-11-02T10:00:00Z').getTime();

/** A camera whose next shot stays at `at` (it has not planned the one after yet). */
function cameraPlanning(at: number) {
  return mockServer({
    'GET /schedule': () => json({ UnixMillis: Date.now(), NextMillis: at, Night: false, ZoneOffsetSeconds: 0 }),
  });
}

/** Renders the hook at `start` and lets the first request finish. */
async function renderAtStart() {
  const hook = renderHook(({ now }) => useNextShot(settings, now), { initialProps: { now: new Date(start) } });
  await act(() => vi.advanceTimersByTimeAsync(0));
  return hook;
}

/** Moves the clock on and shows the hook the new time, as App's useNow does. */
async function advance(hook: Awaited<ReturnType<typeof renderAtStart>>, ms: number) {
  await act(() => vi.advanceTimersByTimeAsync(ms));
  hook.rerender({ now: new Date(Date.now()) });
  await act(() => vi.advanceTimersByTimeAsync(0));
}

describe('useNextShot', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(start);
  });

  it('asks the camera again every minute', async () => {
    const f = cameraPlanning(start + 30 * 60_000);
    const hook = await renderAtStart();
    expect(hook.result.current?.at.getTime()).toBe(start + 30 * 60_000);
    await advance(hook, 59_000);
    expect(callsTo(f, 'GET', '/schedule')).toHaveLength(1);
    await advance(hook, 1_000);
    expect(callsTo(f, 'GET', '/schedule')).toHaveLength(2);
  });

  it('asks every few seconds once the planned shot is overdue', async () => {
    const f = cameraPlanning(start + 3_000);
    const hook = await renderAtStart();
    await advance(hook, 9_000); // 6 s past the planned shot
    const overdue = callsTo(f, 'GET', '/schedule').length;
    await advance(hook, 5_000);
    expect(callsTo(f, 'GET', '/schedule').length).toBe(overdue + 1);
  });
});
