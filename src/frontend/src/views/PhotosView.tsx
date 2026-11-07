import { useMemo, useState } from 'react';
import { api, errorMessage, urls, type Settings } from '../api';
import { HourDial } from '../components/HourDial';
import { Icon } from '../components/Icon';
import { PhotoViewer } from '../components/PhotoViewer';
import { ConfirmDialog, MessageNotice, Notice, PageHeader, RefreshButton, Spinner, type Message } from '../components/ui';
import { formatClock, formatDay, formatDuration, formatPastHour, formatPastHours, formatRelative, plural } from '../format';
import { useBusy, useInterval, useNow, useResource } from '../hooks';
import type { NextShot } from '../nextShot';
import { captureSlots } from '../schedule';
import { groupByDay, toItems, type Day, type PhotoItem } from './photoItems';

/** Selected photo names, with a way to add or remove several at once. */
function useSelection() {
  const [selected, setSelected] = useState<ReadonlySet<string>>(new Set());
  const toggle = (names: string[], on: boolean) =>
    setSelected((prev) => {
      const next = new Set(prev);
      for (const n of names) {
        if (on) next.add(n);
        else next.delete(n);
      }
      return next;
    });
  return { selected, toggle, clear: () => setSelected(new Set()) };
}

export function PhotosView({ settings, next = null }: { settings: Settings | undefined; next?: NextShot | null }) {
  const photos = useResource((signal) => api.photos(signal));
  const items = useMemo(() => toItems(photos.data ?? []), [photos.data]);
  const days = useMemo(() => groupByDay(items), [items]);
  const indexByName = useMemo(() => new Map(items.map((p, i) => [p.name, i])), [items]);
  const { selected, toggle, clear } = useSelection();
  const [viewing, setViewing] = useState<number | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [deleting, runDelete] = useBusy();
  const [message, setMessage] = useState<Message | null>(null);
  const now = useNow();

  // Pick up new captures while the page is open.
  useInterval(photos.reload, 60_000);

  const names = [...selected].filter((n) => indexByName.has(n));

  const deleteSelected = async () => {
    try {
      await api.deletePhotos(names);
      setMessage({ tone: 'ok', text: `Deleted ${plural(names.length, 'photo')}.` });
      clear();
      photos.reload();
    } catch (e) {
      setMessage({ tone: 'error', text: `Couldn't delete photos. ${errorMessage(e)}` });
    }
  };
  const confirmed = async () => {
    await runDelete(deleteSelected);
    setConfirmDelete(false);
  };

  return (
    <>
      <PageHeader title="Photos">
        <RefreshButton resource={photos} />
      </PageHeader>

      {/* No image here: photos load only when opened (the Preview page shows the camera's view). */}
      <section className="hero" aria-label="Latest capture">
        <div className="hero-side">
          {settings ? <HourDial schedule={settings} now={now} next={next} /> : <div className="dial placeholder" />}
          <Facts settings={settings} items={items} days={days} now={now} />
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
      <MessageNotice message={message} onDismiss={() => setMessage(null)} />

      {items.length > 0 && <Toolbar names={names} onDelete={() => setConfirmDelete(true)} onClear={clear} />}

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
        onConfirm={confirmed}
        onCancel={() => setConfirmDelete(false)}
      >
        They will be removed from the camera's SD card. Download them first if you want to keep a copy.
      </ConfirmDialog>
    </>
  );
}

interface FactsProps {
  settings: Settings | undefined;
  items: PhotoItem[];
  days: Day[];
  now: Date;
}

/** Last capture, schedule and photo count next to the dial. */
function Facts({ settings, items, days, now }: FactsProps) {
  const latest = items[0];
  const today = days[0] && latest?.date && formatDay(latest.date) === 'Today' ? days[0].photos.length : 0;
  return (
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
  );
}

/** Downloads of the selection (or everything), and deleting or clearing the selection. */
function Toolbar({ names, onDelete, onClear }: { names: string[]; onDelete: () => void; onClear: () => void }) {
  const any = names.length > 0;
  return (
    <div className="toolbar" data-active={any}>
      <span className="toolbar-count">{any ? `${plural(names.length, 'photo')} selected` : 'Download everything'}</span>
      <div className="toolbar-actions">
        {(['zip', 'tar'] as const).map((format) => (
          <a key={format} className="btn" href={urls.archive(format, names)} download>
            <Icon name="download" /> {format.toUpperCase()}
          </a>
        ))}
        {any && (
          <>
            <button type="button" className="btn danger-ghost" onClick={onDelete}>
              <Icon name="trash" /> Delete
            </button>
            <button type="button" className="btn ghost" onClick={onClear}>
              Clear
            </button>
          </>
        )}
      </div>
    </div>
  );
}

function describeSchedule(s: Settings): string {
  const daytime = s.SkipNightPhotos ? ', daytime only' : '';
  const every = `Every ${formatDuration(s.SecondsBetweenCaptures)}`;
  const slots = captureSlots(s);
  if (!slots) return every + daytime;
  if (slots.length <= 4) return `${every} at ${formatPastHours(slots)}${daytime}`;
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
            <PhotoRow key={p.name} photo={p} selected={selected.has(p.name)} onToggle={onToggle} onOpen={onOpen} />
          ))}
        </ul>
      )}
    </details>
  );
}

interface RowProps {
  photo: PhotoItem;
  selected: boolean;
  onToggle: (names: string[], on: boolean) => void;
  onOpen: (name: string) => void;
}

function PhotoRow({ photo, selected, onToggle, onOpen }: RowProps) {
  const { name } = photo;
  return (
    <li data-selected={selected}>
      <input type="checkbox" aria-label={`Select ${name}`} checked={selected} onChange={(e) => onToggle([name], e.target.checked)} />
      <button type="button" className="photo-open" onClick={() => onOpen(name)}>
        <span className="photo-time">{photo.date ? formatClock(photo.date) : '—'}</span>
        <span className="photo-name">{name}</span>
      </button>
      <span className="photo-size">{photo.size}</span>
      <a className="icon-btn" href={urls.photo(name)} download={name} aria-label={`Download ${name}`}>
        <Icon name="download" />
      </a>
    </li>
  );
}
