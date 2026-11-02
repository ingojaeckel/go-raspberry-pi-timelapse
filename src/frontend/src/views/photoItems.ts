import type { Photo } from '../api';
import { dayKey, parseGoTime } from '../format';

export interface PhotoItem {
  name: string;
  size: string;
  date: Date | null;
}

export interface Day {
  key: string;
  date: Date | null;
  photos: PhotoItem[];
}

/** Newest first, grouped by calendar day in this device's time zone. */
export function toItems(photos: Photo[]): PhotoItem[] {
  return photos
    .map((p) => ({ name: p.Name, size: p.Size, date: parseGoTime(p.ModTime) }))
    .sort((a, b) => (b.date?.getTime() ?? 0) - (a.date?.getTime() ?? 0) || b.name.localeCompare(a.name));
}

export function groupByDay(items: PhotoItem[]): Day[] {
  const days: Day[] = [];
  for (const item of items) {
    const key = item.date ? dayKey(item.date) : 'unknown';
    const last = days[days.length - 1];
    if (last && last.key === key) last.photos.push(item);
    else days.push({ key, date: item.date, photos: [item] });
  }
  return days;
}
