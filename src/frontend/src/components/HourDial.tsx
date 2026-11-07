import type { Settings } from '../api';
import { formatClock, formatCountdown, formatDuration, formatPastHour } from '../format';
import type { NextShot } from '../nextShot';
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

/** A clock face for one hour: amber marks show when the camera fires. */
export function HourDial({ schedule, now, next, label }: Props) {
  const slots = captureSlots(schedule);
  const shot = label ? null : (next ?? null);
  const wait = shot && Math.max(0, Math.round((shot.at.getTime() - now.getTime()) / 1000));
  // Minutes past the hour on the camera's clock, which the capture marks refer to.
  const into = shot
    ? (((Math.floor(now.getTime() / 1000) + shot.zoneOffsetSeconds) % 3600) + 3600) % 3600
    : now.getMinutes() * 60 + now.getSeconds();
  const dense = slots !== null && slots.length > 90;
  const [hx, hy] = polar(into, R - 16);

  const description =
    (slots === null
      ? `Captures every ${formatDuration(schedule.SecondsBetweenCaptures)}, counted from when the camera started.`
      : `Captures at ${slots.length > 6 ? `${slots.length} times` : slots.map(formatPastHour).join(', ')} past each hour.`) +
    (shot?.night ? ` Night: photos resume at ${formatClock(shot.at)}.` : wait !== null ? ` Next capture in ${formatCountdown(wait)}.` : '');

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
        ) : shot?.night ? (
          <>
            <span className="dial-count small">Night</span>
            <span className="dial-caption">photos resume at {formatClock(shot.at)}</span>
          </>
        ) : wait !== null ? (
          <>
            <span className="dial-count">{wait === 0 ? 'Now' : formatCountdown(wait)}</span>
            <span className="dial-caption">until next capture</span>
          </>
        ) : (
          <>
            <span className="dial-count small">Every {formatDuration(schedule.SecondsBetweenCaptures)}</span>
            <span className="dial-caption">{slots === null ? 'since start-up' : 'at the marked minutes'}</span>
          </>
        )}
      </figcaption>
    </figure>
  );
}
