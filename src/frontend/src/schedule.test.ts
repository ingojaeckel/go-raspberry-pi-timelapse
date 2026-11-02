import { captureSlots, capturesPerDay, secondsToNextCapture } from './schedule';

const at = (min: number, sec: number) => new Date(2017, 0, 1, 1, min, sec);
const halfHourly = { SecondsBetweenCaptures: 1800, OffsetWithinHour: 900 };

describe('secondsToNextCapture', () => {
  // Same cases as TestGetInitialSleepTime / TestGetSecondsToFirstCapture in timelapse_test.go
  it.each([
    [9, 13, 47 + 5 * 60],
    [8, 0, 7 * 60],
    [16, 0, 29 * 60],
    [30, 59, 14 * 60 + 1],
    [46, 1, 29 * 60 - 1],
    [32, 1, 13 * 60 - 1],
    [16, 1, 29 * 60 - 1],
    [8, 1, 7 * 60 - 1],
  ])('at :%i:%i waits %i s like the Go scheduler', (m, s, expected) => {
    expect(secondsToNextCapture(halfHourly, at(m, s))).toBe(expected);
  });

  it('is null when captures are not aligned to the hour', () => {
    expect(secondsToNextCapture({ SecondsBetweenCaptures: 60, OffsetWithinHour: -1 }, at(0, 0))).toBeNull();
  });

  it('waits for the next hour when the interval is longer than an hour', () => {
    expect(secondsToNextCapture({ SecondsBetweenCaptures: 7200, OffsetWithinHour: 600 }, at(20, 0))).toBe(50 * 60);
  });
});

describe('captureSlots', () => {
  it('lists capture times past the hour', () => {
    expect(captureSlots(halfHourly)).toEqual([900, 2700]);
    expect(captureSlots({ SecondsBetweenCaptures: 900, OffsetWithinHour: 0 })).toEqual([0, 900, 1800, 2700]);
  });

  it('wraps slots past the end of the hour', () => {
    expect(captureSlots({ SecondsBetweenCaptures: 1200, OffsetWithinHour: 1800 })).toEqual([600, 1800, 3000]);
  });

  it('counts photos per day', () => {
    expect(capturesPerDay(halfHourly)).toBe(48);
    expect(capturesPerDay({ SecondsBetweenCaptures: 600, OffsetWithinHour: -1 })).toBe(144);
  });
});
