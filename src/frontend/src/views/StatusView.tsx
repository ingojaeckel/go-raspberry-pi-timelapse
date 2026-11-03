import { useMemo, type ReactNode } from 'react';
import { api, type Monitoring, type Settings } from '../api';
import { Icon } from '../components/Icon';
import { Notice, PageHeader, Spinner } from '../components/ui';
import {
  checkClock,
  formatBytes,
  formatDuration,
  parseCpuTemp,
  parseDf,
  parseGpuTemp,
  parseSize,
  parseUptime,
  primaryDisk,
} from '../format';
import { useInterval, useResource } from '../hooks';
import { capturesPerDay } from '../schedule';

interface Snapshot {
  monitoring: Monitoring;
  avgPhotoBytes: number | null;
  fetchedAt: Date;
}

async function loadSnapshot(signal: AbortSignal): Promise<Snapshot> {
  const [monitoring, photos] = await Promise.all([
    api.monitoring(signal),
    api.photos(signal).catch(() => []),
  ]);
  const sizes = photos.map((p) => parseSize(p.Size)).filter((n): n is number => n !== null && n > 0);
  return {
    monitoring,
    avgPhotoBytes: sizes.length ? sizes.reduce((a, b) => a + b, 0) / sizes.length : null,
    fetchedAt: new Date(),
  };
}

// The Pi firmware starts throttling at 80 °C.
const tempTone = (t: number) => (t >= 80 ? 'hot' : t >= 70 ? 'warm' : 'ok');

export function StatusView({ settings }: { settings: Settings | undefined }) {
  const snap = useResource(loadSnapshot);
  useInterval(snap.reload, 30_000);

  const view = useMemo(() => {
    if (!snap.data) return null;
    const m = snap.data.monitoring;
    const disks = parseDf(m.FreeDiskSpace);
    return {
      m,
      cpu: parseCpuTemp(m.CpuTemperature),
      gpu: parseGpuTemp(m.GpuTemperature),
      uptime: parseUptime(m.Uptime),
      disks,
      disk: primaryDisk(disks),
      clock: checkClock(m.Time, snap.data.fetchedAt),
    };
  }, [snap.data]);

  return (
    <>
      <PageHeader title="Status">
        <button type="button" className="btn ghost" onClick={snap.reload} disabled={snap.loading}>
          <Icon name="refresh" /> Refresh
        </button>
      </PageHeader>

      {snap.error && <Notice tone="error">{snap.error}</Notice>}
      {!view && snap.loading && <Spinner label="Asking the camera…" />}

      {view && (
        <>
          <ClockWarning clock={view.clock} />

          <section className="stat-grid">
            <Stat label="Storage left" tone={view.disk && view.disk.usedPercent >= 90 ? 'hot' : 'ok'}>
              {view.disk ? (
                <>
                  <span className="stat-value">{humanDf(view.disk.avail)}</span>
                  <meter min={0} max={100} low={75} high={90} optimum={0} value={view.disk.usedPercent}>
                    {view.disk.usedPercent}%
                  </meter>
                  <span className="stat-note">
                    {view.disk.usedPercent}% of {humanDf(view.disk.size)} used
                    <Runway avail={view.disk.avail} avg={snap.data?.avgPhotoBytes ?? null} settings={settings} />
                  </span>
                </>
              ) : (
                <span className="stat-value muted">Unknown</span>
              )}
            </Stat>

            <Stat label="Processor" tone={view.cpu !== null ? tempTone(view.cpu) : 'ok'}>
              <Temp value={view.cpu} />
              <span className="stat-note">{view.cpu !== null && view.cpu >= 80 ? 'Throttling, consider shade or a heatsink' : 'CPU temperature'}</span>
            </Stat>

            <Stat label="Graphics" tone={view.gpu !== null ? tempTone(view.gpu) : 'ok'}>
              <Temp value={view.gpu} />
              <span className="stat-note">GPU temperature</span>
            </Stat>

            <Stat label="Running for" tone="ok">
              <span className="stat-value small">{view.uptime.up ?? 'Unknown'}</span>
              {view.uptime.load && (
                <span className="stat-note">Load {view.uptime.load.map((l) => l.toFixed(2)).join(' / ')} (1, 5, 15 min)</span>
              )}
            </Stat>
          </section>

          {view.disks.length > 1 && (
            <section className="panel">
              <h2>Drives</h2>
              <table className="disks">
                <thead>
                  <tr>
                    <th scope="col">Mounted at</th>
                    <th scope="col">Free</th>
                    <th scope="col">Size</th>
                    <th scope="col">Used</th>
                  </tr>
                </thead>
                <tbody>
                  {view.disks.map((d) => (
                    <tr key={d.filesystem + d.mount}>
                      <td>{d.mount}</td>
                      <td>{humanDf(d.avail)}</td>
                      <td>{humanDf(d.size)}</td>
                      <td>{d.usedPercent}%</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </section>
          )}

          <details className="raw">
            <summary>Raw readings</summary>
            <dl>
              {(Object.entries(view.m) as [keyof Monitoring, string][]).map(([k, v]) => (
                <div key={k}>
                  <dt>{k}</dt>
                  <dd>
                    <pre>{v.trim() || '(empty)'}</pre>
                  </dd>
                </div>
              ))}
            </dl>
          </details>
        </>
      )}
    </>
  );
}

function Stat({ label, tone, children }: { label: string; tone: string; children: ReactNode }) {
  return (
    <div className="stat" data-tone={tone}>
      <h2>{label}</h2>
      {children}
    </div>
  );
}

function Temp({ value }: { value: number | null }) {
  return value === null ? (
    <span className="stat-value muted">Not available</span>
  ) : (
    <span className="stat-value">
      {value.toFixed(1)}
      <span className="unit"> °C</span>
    </span>
  );
}

function Runway({ avail, avg, settings }: { avail: string; avg: number | null; settings: Settings | undefined }) {
  const free = parseSize(avail);
  if (!free || !avg) return null;
  const photos = Math.floor(free / avg);
  const perDay = settings ? capturesPerDay(settings) : 0;
  const days = perDay > 0 ? photos / perDay : null;
  return (
    <>
      . Room for about {photos.toLocaleString()} more photos
      {days !== null && <> ({runway(days)} at the current schedule)</>}.
    </>
  );
}

function runway(days: number): string {
  if (days >= 730) return `over ${Math.floor(days / 365)} years`;
  if (days >= 2) return `${Math.floor(days).toLocaleString()} days`;
  return `${Math.max(1, Math.round(days * 24))} hours`;
}

/** "396Gi" → "396 GB" */
function humanDf(raw: string): string {
  const b = parseSize(raw);
  return b === null ? raw : formatBytes(b);
}

function ClockWarning({ clock }: { clock: ReturnType<typeof checkClock> }) {
  if (clock.wrongYear !== null) {
    return (
      <Notice tone="warn">
        The camera's clock thinks it is {clock.wrongYear}. Photos will carry the wrong date until its clock is set.
        A Pi Zero has no battery-backed clock and can't sync time while offline. See "Real-time clock" in the
        project's <code>docs/installer.md</code> for adding a clock module.
      </Notice>
    );
  }
  if (clock.driftSeconds !== null && Math.abs(clock.driftSeconds) > 60) {
    return (
      <Notice tone="warn">
        The camera's clock is {formatDuration(Math.abs(clock.driftSeconds))}{' '}
        {clock.driftSeconds > 0 ? 'ahead of' : 'behind'} this device, so capture times shown here may be off by that
        much.
      </Notice>
    );
  }
  return null;
}
