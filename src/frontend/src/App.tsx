import { api } from './api';
import { Icon, type IconName } from './components/Icon';
import { formatClock, formatCountdown } from './format';
import { useNow, useResource, useRoute, type Route } from './hooks';
import { nextCapture } from './schedule';
import { LogsView } from './views/LogsView';
import { PhotosView } from './views/PhotosView';
import { PreviewView } from './views/PreviewView';
import { SettingsView } from './views/SettingsView';
import { StatusView } from './views/StatusView';

const NAV: { route: Route; label: string; icon: IconName }[] = [
  { route: 'photos', label: 'Photos', icon: 'photos' },
  { route: 'preview', label: 'Preview', icon: 'camera' },
  { route: 'status', label: 'Status', icon: 'status' },
  { route: 'settings', label: 'Settings', icon: 'settings' },
  { route: 'logs', label: 'Logs', icon: 'logs' },
];

export function App() {
  const route = useRoute();
  const settings = useResource((signal) => api.settings(signal));
  const now = useNow();
  const next = settings.data ? nextCapture(settings.data, now) : null;
  const wait = next && Math.max(0, Math.round((next.at.getTime() - now.getTime()) / 1000));

  return (
    <div className="shell">
      <header className="topbar">
        <a className="brand" href="#/photos">
          <svg viewBox="0 0 32 32" width="26" height="26" aria-hidden="true">
            <circle cx="16" cy="16" r="14" className="brand-face" />
            <circle cx="16" cy="16" r="10" className="brand-ring" />
            <path d="M16 16V8" className="brand-hand" />
          </svg>
          Timelapse
        </a>
        {next && (
          <span
            className="next-pill"
            aria-live="off"
            title={next.night ? `Nights are skipped: photos resume at ${formatClock(next.at)}` : undefined}
          >
            Next shot <strong>{next.night ? `at ${formatClock(next.at)}` : wait === 0 ? 'now' : formatCountdown(wait!)}</strong>
          </span>
        )}
      </header>

      <nav className="nav" aria-label="Sections">
        {NAV.map((n) => (
          <a key={n.route} href={`#/${n.route}`} aria-current={route === n.route ? 'page' : undefined}>
            <Icon name={n.icon} size={22} />
            <span>{n.label}</span>
          </a>
        ))}
      </nav>

      <main className="main">
        {route === 'photos' && <PhotosView settings={settings.data} />}
        {route === 'preview' && <PreviewView settings={settings.data} />}
        {route === 'status' && <StatusView settings={settings.data} />}
        {route === 'settings' && <SettingsView settings={settings.data} error={settings.error} onSaved={settings.set} />}
        {route === 'logs' && <LogsView />}
      </main>
    </div>
  );
}
