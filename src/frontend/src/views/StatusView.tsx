import { useMemo, useState, type ReactNode } from 'react';
import { api, type Monitoring, type MonitoringOutput, type Settings, type Storage } from '../api';
import { Icon } from '../components/Icon';
import { Notice, PageHeader, Spinner } from '../components/ui';
import {
  checkClock,
  estimateRunway,
  formatBytes,
  formatDays,
  formatDuration,
  parseCpuTemp,
  parseDf,
  parseGpuTemp,
  parseSize,
  parseUptime,
  primaryDisk,
} from '../format';
import { useInterval, useResource } from '../hooks';
import { photosPerDay } from '../schedule';

interface Snapshot {
  monitoring: Monitoring;
  fetchedAt: Date;
}

async function loadSnapshot(signal: AbortSignal): Promise<Snapshot> {
  return { monitoring: await api.monitoring(signal), fetchedAt: new Date() };
}

// The Pi firmware starts throttling at 80 °C.
const tempTone = (t: number) => (t >= 80 ? 'hot' : t >= 70 ? 'warm' : 'ok');

/** Command output shown under "Raw readings". */
const RAW_KEYS: (keyof MonitoringOutput)[] = ['Time', 'Uptime', 'CpuTemperature', 'GpuTemperature', 'FreeDiskSpace'];

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
      clock: checkClock(m.UnixMillis, snap.data.fetchedAt),
    };
  }, [snap.data]);

  const full = view ? isFull(view.m.Storage) : false;

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
          <ClockWarning clock={view.clock} onSet={snap.reload} />
          {full && view.m.Storage.StopWhenDiskFull && (
            <Notice tone="error">
              The card is full and the camera has stopped taking photos. Download and delete photos, or let the camera
              delete the oldest ones (Settings, "When the card is full").
            </Notice>
          )}

          <section className="stat-grid">
            <Stat label="Storage left" tone={full || (view.disk && view.disk.usedPercent >= 90) ? 'hot' : 'ok'}>
              {view.disk ? (
                <>
                  <span className="stat-value">{humanDf(view.disk.avail)}</span>
                  <meter min={0} max={100} low={75} high={90} optimum={0} value={view.disk.usedPercent}>
                    {view.disk.usedPercent}%
                  </meter>
                  <span className="stat-note">
                    {view.disk.usedPercent}% of {humanDf(view.disk.size)} used
                    <Runway storage={view.m.Storage} settings={settings} />
                  </span>
                </>
              ) : (
                <span className="stat-value muted">Unknown</span>
              )}
            </Stat>

            <PowerStat power={view.m.Power} />

            <Stat label="Processor" tone={view.cpu !== null ? tempTone(view.cpu) : 'ok'}>
              <Temp value={view.cpu} />
              <span className="stat-note">
                {view.cpu !== null && view.cpu >= 80 ? 'Throttling, consider shade or a heatsink' : 'CPU temperature'}
                <CpuClock cpu={view.m.CPU} />
              </span>
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
              {RAW_KEYS.map((k) => (
                <div key={k}>
                  <dt>{k}</dt>
                  <dd>
                    <pre>{view.m[k].trim() || '(empty)'}</pre>
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

const sentence = (flags: string[]) => {
  const s = flags.join(', ');
  return s.charAt(0).toUpperCase() + s.slice(1);
};

/** Under-voltage is the usual cause of SD card corruption and failed captures on battery or solar power. */
function PowerStat({ power }: { power: Monitoring['Power'] }) {
  if (power.Error) {
    return (
      <Stat label="Power" tone="ok">
        <span className="stat-value muted">Not available</span>
        <span className="stat-note">Reported by the Pi firmware (vcgencmd).</span>
      </Stat>
    );
  }
  const tone = power.Now.length ? 'hot' : power.SinceBoot.length ? 'warm' : 'ok';
  return (
    <Stat label="Power" tone={tone}>
      <span className="stat-value small">{power.Now.length ? sentence(power.Now) : 'OK'}</span>
      <span className="stat-note">
        {power.SinceBoot.length
          ? `Since the last restart: ${power.SinceBoot.join(', ')}. Check the power supply and cable.`
          : 'No under-voltage or throttling since the last restart.'}
      </span>
    </Stat>
  );
}

function CpuClock({ cpu }: { cpu: Monitoring['CPU'] }) {
  if (cpu.Error) return null;
  return (
    <>
      . Runs at {cpu.CurrentMHz} MHz ({cpu.Governor}, {cpu.MinMHz}–{cpu.MaxMHz} MHz)
    </>
  );
}

function isFull(s: Storage): boolean {
  return !s.Error && s.FreeBytes < s.ReserveBytes;
}

function Runway({ storage, settings }: { storage: Storage; settings: Settings | undefined }) {
  const est = estimateRunway(storage, settings ? photosPerDay(settings, new Date()) : 0);
  const then = storage.StopWhenDiskFull ? 'the camera stops taking photos' : 'the oldest photos are deleted to make room';
  const deleted =
    storage.DeletedOldest > 0 ? (
      <> {storage.DeletedOldest.toLocaleString()} old photos deleted to make room since the camera started.</>
    ) : null;
  if (!est) return deleted;
  if (est.photos === 0) {
    return (
      <>
        . Full: {then}.{deleted}
      </>
    );
  }
  return (
    <>
      . Room for about {est.photos.toLocaleString()} more photos
      {est.days !== null && (
        <>
          {' '}
          ({formatDays(est.days)} at the current schedule{settings?.SkipNightPhotos ? ', nights skipped' : ''})
        </>
      )}, then {then}.{deleted}
    </>
  );
}

/** "396Gi" → "396 GB" */
function humanDf(raw: string): string {
  const b = parseSize(raw);
  return b === null ? raw : formatBytes(b);
}

function ClockWarning({ clock, onSet }: { clock: ReturnType<typeof checkClock>; onSet: () => void }) {
  const [busy, setBusy] = useState(false);
  const [result, setResult] = useState<{ tone: 'ok' | 'warn' | 'error'; text: string } | null>(null);

  const setClock = async () => {
    setBusy(true);
    try {
      const r = await api.setTime(Date.now());
      setResult(
        r.RTC === '' || r.RTC === 'updated'
          ? { tone: 'ok', text: `Camera clock set${r.RTC === 'updated' ? ', and saved to the clock module' : ''}.` }
          : { tone: 'warn', text: `Camera clock set, but the clock module was ${r.RTC}.` },
      );
      onSet();
    } catch (err) {
      setResult({ tone: 'error', text: `The clock was not set. ${(err as Error).message}` });
    } finally {
      setBusy(false);
    }
  };

  const off = clock.wrongYear !== null || Math.abs(clock.driftSeconds) > 60;
  if (!off) {
    return result ? <Notice tone={result.tone}>{result.text}</Notice> : null;
  }
  return (
    <Notice tone="warn">
      {clock.wrongYear !== null ? (
        <>The camera's clock thinks it is {clock.wrongYear}. Photos will carry the wrong date until its clock is set. </>
      ) : (
        <>
          The camera's clock is {formatDuration(Math.abs(clock.driftSeconds))}{' '}
          {clock.driftSeconds > 0 ? 'ahead of' : 'behind'} this device, so photos carry the wrong time.{' '}
        </>
      )}
      Without a battery-backed clock module the Pi loses time at every power cut (see "Real-time clock" in the
      project's <code>docs/installer.md</code>).
      <br />
      <button type="button" className="btn" onClick={setClock} disabled={busy}>
        <Icon name="refresh" /> {busy ? 'Setting the clock…' : "Set the camera's clock to this device's time"}
      </button>
      {result && <> {result.text}</>}
    </Notice>
  );
}
