import type { Settings } from '../api';
import { formatClock, formatCountdown, formatDuration, formatPastHour } from '../format';
import { captureSlots, secondsToNextCapture } from '../schedule';
import { nextSunrise, sunUp } from '../sun';

type Schedule = Pick<Settings, 'SecondsBetweenCaptures' | 'OffsetWithinHour'> &
  Partial<Pick<Settings, 'SkipNightPhotos' | 'Latitude' | 'Longitude'>>;

/** With nights skipped, the sunrise the next capture waits for if it falls at night (like the camera's scheduler). */
function resumesAt(schedule: Schedule, now: Date, wait: number | null): Date | null {
  const { SkipNightPhotos, Latitude: lat = 0, Longitude: lon = 0 } = schedule;
  if (!SkipNightPhotos) return null;
  const next = new Date(now.getTime() + (wait ?? 0) * 1000);
  return sunUp(next, lat, lon) ? null : nextSunrise(next, lat, lon);
}

interface Props {
  schedule: Schedule;
  now: Date;
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

/** A clock face for one hour: amber marks show when the camera fires. */
export function HourDial({ schedule, now, label }: Props) {
  const slots = captureSlots(schedule);
  const wait = secondsToNextCapture(schedule, now);
  const resume = resumesAt(schedule, now, wait);
  const into = now.getMinutes() * 60 + now.getSeconds();
  const dense = slots !== null && slots.length > 90;
  const [hx, hy] = polar(into, R - 16);

  const description =
    slots === null
      ? `Captures every ${formatDuration(schedule.SecondsBetweenCaptures)}, counted from when the camera started.`
      : `Captures at ${slots.length > 6 ? `${slots.length} times` : slots.map(formatPastHour).join(', ')} past each hour.` +
        (resume && !label
          ? ` Night: photos resume at ${formatClock(resume)}.`
          : wait !== null && !label
            ? ` Next capture in ${formatCountdown(wait)}.`
            : '');

  return (
    <figure className="dial">
      <svg viewBox={`0 0 ${SIZE} ${SIZE}`} role="img" aria-label={description}>
        <circle className="dial-face" cx={C} cy={C} r={R + 12} />
        {Array.from({ length: 60 }, (_, m) => {
          const major = m % 5 === 0;
          const [x1, y1] = polar(m * 60, R + 8);
          const [x2, y2] = polar(m * 60, major ? R + 1 : R + 5);
          return <line key={m} className={major ? 'dial-tick major' : 'dial-tick'} x1={x1} y1={y1} x2={x2} y2={y2} />;
        })}
        {dense && <circle className="dial-band" cx={C} cy={C} r={R - 5} />}
        {!dense &&
          slots?.map((s) => {
            const [x, y] = polar(s, R - 5);
            return <circle key={s} className="dial-slot" cx={x} cy={y} r={slots.length > 24 ? 2.2 : 4.5} />;
          })}
        {slots === null && <circle className="dial-band free" cx={C} cy={C} r={R - 5} />}
        <line className="dial-hand" x1={C} y1={C} x2={hx} y2={hy} />
        <circle className="dial-hub" cx={C} cy={C} r={3.5} />
      </svg>
      <figcaption>
        {label ? (
          <span className="dial-caption">{label}</span>
        ) : resume ? (
          <>
            <span className="dial-count small">Night</span>
            <span className="dial-caption">photos resume at {formatClock(resume)}</span>
          </>
        ) : wait !== null ? (
          <>
            <span className="dial-count">{wait === 0 ? 'Now' : formatCountdown(wait)}</span>
            <span className="dial-caption">until next capture</span>
          </>
        ) : (
          <>
            <span className="dial-count small">Every {formatDuration(schedule.SecondsBetweenCaptures)}</span>
            <span className="dial-caption">since start-up</span>
          </>
        )}
      </figcaption>
    </figure>
  );
}
