import shared from './sun-cases.json';
import { daylightFraction, nextSunrise, nextSunset, sunUp, validLocation } from './sun';

// sun-cases.json is shared with src/backend/sun/sun_test.go and generated
// with an independent library, so the UI and the camera agree on night.
const tol = shared.toleranceSeconds * 1000;

describe('sun', () => {
  it.each(shared.days)('is up between sunrise and sunset in $place on $date', (d) => {
    const rise = d.sunrise * 1000;
    const set = d.sunset * 1000;
    expect(sunUp(new Date(rise - tol), d.lat, d.lon)).toBe(false);
    expect(sunUp(new Date(rise + tol), d.lat, d.lon)).toBe(true);
    expect(sunUp(new Date(set - tol), d.lat, d.lon)).toBe(true);
    expect(sunUp(new Date(set + tol), d.lat, d.lon)).toBe(false);
  });

  it.each(shared.days)('finds the next sunrise and sunset in $place on $date', (d) => {
    const rise = nextSunrise(new Date(d.sunrise * 1000 - 3 * 3600_000), d.lat, d.lon);
    const set = nextSunset(new Date(d.sunset * 1000 - 3 * 3600_000), d.lat, d.lon);
    expect(Math.abs(rise!.getTime() - d.sunrise * 1000)).toBeLessThanOrEqual(tol);
    expect(Math.abs(set!.getTime() - d.sunset * 1000)).toBeLessThanOrEqual(tol);
  });

  it.each(shared.polar)('knows polar day and night in $place at $unix', (p) => {
    expect(sunUp(new Date(p.unix * 1000), p.lat, p.lon)).toBe(p.up);
  });

  it('has no sunrise within a day of the polar night', () => {
    expect(nextSunrise(new Date('2026-12-21T12:00:00Z'), 69.6492, 18.9553, 1)).toBeNull();
  });

  it('measures the share of the day with the sun up', () => {
    // London: about 16.6 h at midsummer, 7.8 h at midwinter
    expect(daylightFraction(new Date('2026-06-21T00:00:00Z'), 51.5074, -0.1278)).toBeCloseTo(16.6 / 24, 1);
    expect(daylightFraction(new Date('2026-12-21T00:00:00Z'), 51.5074, -0.1278)).toBeCloseTo(7.8 / 24, 1);
    expect(daylightFraction(new Date('2026-12-21T00:00:00Z'), 69.6492, 18.9553)).toBe(0);
  });

  it('validates locations like the backend', () => {
    expect(validLocation(51.5, -0.13)).toBe(true);
    expect(validLocation(-90, 180)).toBe(true);
    expect(validLocation(90.1, 0)).toBe(false);
    expect(validLocation(0, -180.5)).toBe(false);
    expect(validLocation(NaN, 0)).toBe(false);
  });
});
