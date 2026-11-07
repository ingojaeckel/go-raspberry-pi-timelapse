import type { Settings } from '../api';
import { formatClock, formatCountdown, formatDuration, formatPastHours } from '../format';
import { secondsUntil, type NextShot } from '../nextShot';
import { captureSlots } from '../schedule';

type Schedule = Pick<Settings, 'SecondsBetweenCaptures' | 'OffsetWithinHour'>;

interface Props {
  schedule: Schedule;
  now: Date;
  /** The camera's planned next capture (GET /schedule); null or absent shows no countdown. */
  next?: NextShot | null;
  /** Hide the countdown, e.g. when previewing unsaved settings. */
  label?: string;
}

const SIZE = 200;
const C = SIZE / 2;
const R = 82;

function polar(secondsIntoHour: number, radius: number): [number, number] {
  const a = (secondsIntoHour / 3600) * Math.PI * 2 - Math.PI / 2;
  return [C + Math.cos(a) * radius, C + Math.sin(a) * radius];
}

/** Seconds past the hour on the camera's clock (which the capture marks refer to), or on this device's without a plan. */
function secondsIntoHour(now: Date, shot: NextShot | null): number {
  if (!shot) return now.getMinutes() * 60 + now.getSeconds();
  return (((Math.floor(now.getTime() / 1000) + shot.zoneOffsetSeconds) % 3600) + 3600) % 3600;
}

/** A clock face for one hour: amber marks show when the camera fires. */
export function HourDial({ schedule, now, next, label }: Props) {
  const slots = captureSlots(schedule);
  const shot = label ? null : (next ?? null);
  const wait = shot && secondsUntil(shot, now);
  const [hx, hy] = polar(secondsIntoHour(now, shot), R - 16);
  return (
    <figure className="dial">
      <svg viewBox={`0 0 ${SIZE} ${SIZE}`} role="img" aria-label={describe(schedule, slots, shot, wait)}>
        <circle className="dial-face" cx={C} cy={C} r={R + 12} />
        <MinuteTicks />
        <SlotMarks slots={slots} />
        <line className="dial-hand" x1={C} y1={C} x2={hx} y2={hy} />
        <circle className="dial-hub" cx={C} cy={C} r={3.5} />
      </svg>
      <figcaption>
        <Caption schedule={schedule} slots={slots} shot={shot} wait={wait} label={label} />
      </figcaption>
    </figure>
  );
}

function describe(schedule: Schedule, slots: number[] | null, shot: NextShot | null, wait: number | null): string {
  const captures =
    slots === null
      ? `Captures every ${formatDuration(schedule.SecondsBetweenCaptures)}, counted from when the camera started.`
      : `Captures at ${slots.length > 6 ? `${slots.length} times` : formatPastHours(slots)} past each hour.`;
  if (shot?.night) return `${captures} Night: photos resume at ${formatClock(shot.at)}.`;
  return wait !== null ? `${captures} Next capture in ${formatCountdown(wait)}.` : captures;
}

function MinuteTicks() {
  return Array.from({ length: 60 }, (_, m) => {
    const major = m % 5 === 0;
    const [x1, y1] = polar(m * 60, R + 8);
    const [x2, y2] = polar(m * 60, major ? R + 1 : R + 5);
    return <line key={m} className={major ? 'dial-tick major' : 'dial-tick'} x1={x1} y1={y1} x2={x2} y2={y2} />;
  });
}

/** A dot per capture; a band when they are too dense to tell apart or not tied to the hour (null). */
function SlotMarks({ slots }: { slots: number[] | null }) {
  if (slots === null) return <circle className="dial-band free" cx={C} cy={C} r={R - 5} />;
  if (slots.length > 90) return <circle className="dial-band" cx={C} cy={C} r={R - 5} />;
  return slots.map((s) => {
    const [x, y] = polar(s, R - 5);
    return <circle key={s} className="dial-slot" cx={x} cy={y} r={slots.length > 24 ? 2.2 : 4.5} />;
  });
}

interface CaptionProps {
  schedule: Schedule;
  slots: number[] | null;
  shot: NextShot | null;
  wait: number | null;
  label: string | undefined;
}

function Caption({ schedule, slots, shot, wait, label }: CaptionProps) {
  if (label) return <span className="dial-caption">{label}</span>;
  const [count, caption, small] = shot?.night
    ? ['Night', `photos resume at ${formatClock(shot.at)}`, true]
    : wait !== null
      ? [wait === 0 ? 'Now' : formatCountdown(wait), 'until next capture', false]
      : [`Every ${formatDuration(schedule.SecondsBetweenCaptures)}`, slots === null ? 'since start-up' : 'at the marked minutes', true];
  return (
    <>
      <span className={small ? 'dial-count small' : 'dial-count'}>{count}</span>
      <span className="dial-caption">{caption}</span>
    </>
  );
}
