import { nextSunrise } from './sun';
import {
  VALID_INTERVALS,
  nextCapture,
  captureSlots,
  capturesPerDay,
  formatInterval,
  isValidInterval,
  secondsToNextCapture,
} from './schedule';
import shared from './schedule-cases.json';

// schedule-cases.json is shared with src/backend/timelapse/schedule_test.go, so
// the UI predicts exactly what the camera does.

const at = (min: number, sec: number) => new Date(2017, 0, 1, 1, min, sec);
const halfHourly = { SecondsBetweenCaptures: 1800, OffsetWithinHour: 900 };

describe('secondsToNextCapture', () => {
  it.each(shared.cases)(
    'interval $interval offset $offset at :$minute:$second waits $wait s like the Go scheduler',
    ({ interval, offset, minute, second, wait }) => {
      expect(secondsToNextCapture({ SecondsBetweenCaptures: interval, OffsetWithinHour: offset }, at(minute, second))).toBe(wait);
    },
  );

  it('is null when captures are not aligned to the hour', () => {
    expect(secondsToNextCapture({ SecondsBetweenCaptures: 60, OffsetWithinHour: -1 }, at(0, 0))).toBeNull();
  });

  it('is null for intervals the camera rejects', () => {
    expect(secondsToNextCapture({ SecondsBetweenCaptures: 7200, OffsetWithinHour: 600 }, at(20, 0))).toBeNull();
  });
});

describe('valid intervals', () => {
  it('match the Go validator', () => {
    expect(VALID_INTERVALS).toEqual(shared.validIntervals);
    for (const v of shared.validIntervals) expect(isValidInterval(v)).toBe(true);
    for (const v of shared.invalidIntervals) expect(isValidInterval(v)).toBe(false);
  });

  it('have readable labels', () => {
    expect(formatInterval(60)).toBe('1 minute');
    expect(formatInterval(72)).toBe('72 seconds');
    expect(formatInterval(90)).toBe('90 seconds');
    expect(formatInterval(600)).toBe('10 minutes');
    expect(formatInterval(3600)).toBe('1 hour');
  });
});

describe('captureSlots', () => {
  it('lists capture times past the hour', () => {
    expect(captureSlots(halfHourly)).toEqual([900, 2700]);
    expect(captureSlots({ SecondsBetweenCaptures: 900, OffsetWithinHour: 0 })).toEqual([0, 900, 1800, 2700]);
    expect(captureSlots({ SecondsBetweenCaptures: 3600, OffsetWithinHour: 600 })).toEqual([600]);
  });

  it('is null for intervals the camera rejects', () => {
    expect(captureSlots({ SecondsBetweenCaptures: 1000, OffsetWithinHour: 0 })).toBeNull();
  });

  it('counts photos per day', () => {
    expect(capturesPerDay(halfHourly)).toBe(48);
    expect(capturesPerDay({ SecondsBetweenCaptures: 600, OffsetWithinHour: -1 })).toBe(144);
    expect(capturesPerDay({ SecondsBetweenCaptures: 1000, OffsetWithinHour: -1 })).toBe(0);
  });
});

describe('nextCapture', () => {
  const london = { SecondsBetweenCaptures: 120, OffsetWithinHour: 0, SkipNightPhotos: true, Latitude: 51.5074, Longitude: -0.1278 };

  it('is the next slot during the day', () => {
    expect(nextCapture(london, new Date('2026-06-21T12:11:00Z'))).toEqual({
      at: new Date('2026-06-21T12:12:00Z'),
      night: false,
    });
  });

  it('is the first slot after sunrise at night, like the camera', () => {
    const now = new Date('2026-06-21T23:11:00Z');
    const rise = nextSunrise(now, london.Latitude, london.Longitude)!;
    const next = nextCapture(london, now)!;
    expect(next.night).toBe(true);
    expect(next.at.getTime()).toBeGreaterThanOrEqual(rise.getTime());
    expect(next.at.getTime() - rise.getTime()).toBeLessThan(120_000);
    expect(next.at.getTime() % 120_000).toBe(0);
  });

  it('is the next slot at night when night photos are taken', () => {
    expect(nextCapture({ ...london, SkipNightPhotos: false }, new Date('2026-06-21T23:11:00Z'))).toEqual({
      at: new Date('2026-06-21T23:12:00Z'),
      night: false,
    });
  });

  it('is null when captures are not aligned to the hour', () => {
    expect(nextCapture({ ...london, OffsetWithinHour: -1 }, new Date('2026-06-21T23:11:00Z'))).toBeNull();
  });
});
