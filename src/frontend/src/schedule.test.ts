import {
  VALID_INTERVALS,
  captureSlots,
  capturesPerDay,
  formatInterval,
  isValidInterval,
} from './schedule';
import shared from './schedule-cases.json';

// schedule-cases.json is shared with src/backend/timelapse/schedule_test.go, so
// the UI offers exactly the intervals the camera accepts.
const halfHourly = { SecondsBetweenCaptures: 1800, OffsetWithinHour: 900 };

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
