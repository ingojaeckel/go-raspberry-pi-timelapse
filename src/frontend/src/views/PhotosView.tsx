import { useMemo, useState } from 'react';
import { api, urls, type Settings } from '../api';
import { HourDial } from '../components/HourDial';
import { Icon } from '../components/Icon';
import { PhotoViewer } from '../components/PhotoViewer';
import { ConfirmDialog, Notice, PageHeader, Spinner } from '../components/ui';
import { formatClock, formatDay, formatDuration, formatPastHour, formatRelative } from '../format';
import { useInterval, useNow, useResource } from '../hooks';
import type { NextShot } from '../nextShot';
import { captureSlots } from '../schedule';
import { groupByDay, toItems, type Day } from './photoItems';

const plural = (n: number, word: string) => `${n.toLocaleString()} ${word}${n === 1 ? '' : 's'}`;

export function PhotosView({ settings, next = null }: { settings: Settings | undefined; next?: NextShot | null }) {
  const photos = useResource((signal) => api.photos(signal));
  const items = useMemo(() => toItems(photos.data ?? []), [photos.data]);
  const days = useMemo(() => groupByDay(items), [items]);
  const indexByName = useMemo(() => new Map(items.map((p, i) => [p.name, i])), [items]);
  const [selected, setSelected] = useState<ReadonlySet<string>>(new Set());
  const [viewing, setViewing] = useState<number | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [flash, setFlash] = useState<{ tone: 'ok' | 'error'; text: string } | null>(null);
  const now = useNow();

  // Pick up new captures while the page is open.
  useInterval(photos.reload, 60_000);

  const latest = items[0];
  const today = days[0] && latest?.date && formatDay(latest.date) === 'Today' ? days[0].photos.length : 0;
  const names = [...selected].filter((n) => indexByName.has(n));

  const toggle = (list: string[], on: boolean) =>
    setSelected((prev) => {
      const next = new Set(prev);
      for (const n of list) {
        if (on) next.add(n);
        else next.delete(n);
      }
      return next;
    });

  const doDelete = async () => {
    setDeleting(true);
    try {
      await api.deletePhotos(names);
      setFlash({ tone: 'ok', text: `Deleted ${plural(names.length, 'photo')}.` });
      setSelected(new Set());
      photos.reload();
    } catch (e) {
      setFlash({ tone: 'error', text: `Couldn't delete photos. ${(e as Error).message}` });
    } finally {
      setDeleting(false);
      setConfirmDelete(false);
    }
  };

  return (
    <>
      <PageHeader title="Photos">
        <button type="button" className="btn ghost" onClick={photos.reload} disabled={photos.loading}>
          <Icon name="refresh" /> Refresh
        </button>
      </PageHeader>

      {/* No image here: photos load only when opened (the Preview page shows the camera's view). */}
      <section className="hero" aria-label="Latest capture">
        <div className="hero-side">
          {settings ? <HourDial schedule={settings} now={now} next={next} /> : <div className="dial placeholder" />}
          <dl className="facts">
            <div>
              <dt>Last capture</dt>
              <dd>{latest?.date ? `${formatClock(latest.date)}, ${formatRelative(latest.date, now)}` : 'None yet'}</dd>
            </div>
            <div>
              <dt>Schedule</dt>
              <dd>{settings ? describeSchedule(settings) : '…'}</dd>
            </div>
            <div>
              <dt>On the card</dt>
              <dd>
                {plural(items.length, 'photo')}
                {today > 0 && `, ${today.toLocaleString()} today`}
              </dd>
            </div>
          </dl>
        </div>
      </section>

      {photos.error && (
        <Notice tone="error">
          {photos.error}{' '}
          <button type="button" className="link" onClick={photos.reload}>
            Try again
          </button>
        </Notice>
      )}
      {flash && (
        <Notice tone={flash.tone} onDismiss={() => setFlash(null)}>
          {flash.text}
        </Notice>
      )}

      {items.length > 0 && (
        <div className="toolbar" data-active={names.length > 0}>
          <span className="toolbar-count">
            {names.length > 0 ? `${plural(names.length, 'photo')} selected` : 'Download everything'}
          </span>
          <div className="toolbar-actions">
            <a className="btn" href={urls.archive('zip', names)} download>
              <Icon name="download" /> ZIP
            </a>
            <a className="btn" href={urls.archive('tar', names)} download>
              <Icon name="download" /> TAR
            </a>
            {names.length > 0 && (
              <>
                <button type="button" className="btn danger-ghost" onClick={() => setConfirmDelete(true)}>
                  <Icon name="trash" /> Delete
                </button>
                <button type="button" className="btn ghost" onClick={() => setSelected(new Set())}>
                  Clear
                </button>
              </>
            )}
          </div>
        </div>
      )}

      {photos.loading && !photos.data && <Spinner label="Loading photos…" />}

      {photos.data && items.length === 0 && !photos.error && (
        <p className="empty">
          The camera hasn't saved any photos yet. Check the <a href="#/preview">preview</a> to make sure it's pointing
          where you want.
        </p>
      )}

      <div className="days">
        {days.map((day, i) => (
          <DayGroup
            key={day.key}
            day={day}
            defaultOpen={i === 0}
            selected={selected}
            onToggle={toggle}
            onOpen={(name) => setViewing(indexByName.get(name) ?? null)}
          />
        ))}
      </div>

      <PhotoViewer photos={items} index={viewing} onIndex={setViewing} />

      <ConfirmDialog
        open={confirmDelete}
        title={`Delete ${plural(names.length, 'photo')}?`}
        confirmLabel="Delete"
        danger
        busy={deleting}
        onConfirm={doDelete}
        onCancel={() => setConfirmDelete(false)}
      >
        They will be removed from the camera's SD card. Download them first if you want to keep a copy.
      </ConfirmDialog>
    </>
  );
}

function describeSchedule(s: Settings): string {
  const daytime = s.SkipNightPhotos ? ', daytime only' : '';
  const every = `Every ${formatDuration(s.SecondsBetweenCaptures)}`;
  const slots = captureSlots(s);
  if (!slots) return every + daytime;
  if (slots.length <= 4) return `${every} at ${slots.map(formatPastHour).join(', ')}${daytime}`;
  return `${every}, from ${formatPastHour(s.OffsetWithinHour % 3600)}${daytime}`;
}

interface DayProps {
  day: Day;
  defaultOpen: boolean;
  selected: ReadonlySet<string>;
  onToggle: (names: string[], on: boolean) => void;
  onOpen: (name: string) => void;
}

function DayGroup({ day, defaultOpen, selected, onToggle, onOpen }: DayProps) {
  const [open, setOpen] = useState(defaultOpen);
  const names = day.photos.map((p) => p.name);
  const count = names.filter((n) => selected.has(n)).length;
  const label = day.date ? formatDay(day.date) : 'Unknown date';
  const first = day.photos[day.photos.length - 1]?.date;
  const last = day.photos[0]?.date;

  return (
    <details className="day" open={open} onToggle={(e) => setOpen(e.currentTarget.open)}>
      <summary>
        <input
          type="checkbox"
          aria-label={`Select all photos from ${label}`}
          checked={count === names.length}
          ref={(el) => {
            if (el) el.indeterminate = count > 0 && count < names.length;
          }}
          onClick={(e) => e.stopPropagation()}
          onChange={(e) => onToggle(names, e.target.checked)}
        />
        <span className="day-label">{label}</span>
        <span className="day-meta">
          {plural(day.photos.length, 'photo')}
          {first && last && day.photos.length > 1 && (
            <span className="day-span">
              {' '}
              {formatClock(first)}–{formatClock(last)}
            </span>
          )}
        </span>
        <Icon name="next" />
      </summary>
      {open && (
        <ul className="photo-list">
          {day.photos.map((p) => (
            <li key={p.name} data-selected={selected.has(p.name)}>
              <input
                type="checkbox"
                aria-label={`Select ${p.name}`}
                checked={selected.has(p.name)}
                onChange={(e) => onToggle([p.name], e.target.checked)}
              />
              <button type="button" className="photo-open" onClick={() => onOpen(p.name)}>
                <span className="photo-time">{p.date ? formatClock(p.date) : '—'}</span>
                <span className="photo-name">{p.name}</span>
              </button>
              <span className="photo-size">{p.size}</span>
              <a className="icon-btn" href={urls.photo(p.name)} download={p.name} aria-label={`Download ${p.name}`}>
                <Icon name="download" />
              </a>
            </li>
          ))}
        </ul>
      )}
    </details>
  );
}
