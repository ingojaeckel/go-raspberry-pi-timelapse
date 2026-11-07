// Port of src/backend/sun/sun.go: day or night from the time and location,
// offline (NOAA's solar position equations). Both are tested against
// sun-cases.json, so the UI predicts the nights the camera skips.

/** The sun's elevation at sunrise and sunset: its radius plus refraction below the horizon. */
const HORIZON = -0.833;
/** Search step for the next sunrise or sunset. */
const STEP = 10 * 60_000;

const rad = (d: number) => (d * Math.PI) / 180;
const deg = (r: number) => (r * 180) / Math.PI;
const sin = (d: number) => Math.sin(rad(d));
const cos = (d: number) => Math.cos(rad(d));
const mod = (a: number, n: number) => ((a % n) + n) % n;

/** Latitude -90..90 (north positive), longitude -180..180 (east positive). */
export function validLocation(lat: number, lon: number): boolean {
  return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180;
}

/** 0, 0 is what an unset location is saved as. */
export function noLocation(lat: number, lon: number): boolean {
  return lat === 0 && lon === 0;
}

/** The sun's angle above the horizon in degrees, without refraction. */
export function elevation(t: Date, lat: number, lon: number): number {
  const unix = t.getTime() / 1000;
  const jc = (unix / 86400 + 2440587.5 - 2451545) / 36525;
  const meanLong = mod(280.46646 + jc * (36000.76983 + jc * 0.0003032), 360);
  const meanAnom = 357.52911 + jc * (35999.05029 - 0.0001537 * jc);
  const ecc = 0.016708634 - jc * (0.000042037 + 0.0000001267 * jc);
  const center =
    sin(meanAnom) * (1.914602 - jc * (0.004817 + 0.000014 * jc)) +
    sin(2 * meanAnom) * (0.019993 - 0.000101 * jc) +
    sin(3 * meanAnom) * 0.000289;
  const omega = 125.04 - 1934.136 * jc;
  const appLong = meanLong + center - 0.00569 - 0.00478 * sin(omega);
  const meanObliq = 23 + (26 + (21.448 - jc * (46.815 + jc * (0.00059 - jc * 0.001813))) / 60) / 60;
  const obliq = meanObliq + 0.00256 * cos(omega);
  const decl = deg(Math.asin(sin(obliq) * sin(appLong)));
  const y = Math.tan(rad(obliq / 2)) ** 2;
  const eqTime =
    4 *
    deg(
      y * sin(2 * meanLong) -
        2 * ecc * sin(meanAnom) +
        4 * ecc * y * sin(meanAnom) * cos(2 * meanLong) -
        0.5 * y * y * sin(4 * meanLong) -
        1.25 * ecc * ecc * sin(2 * meanAnom),
    );
  const minutes = mod(unix, 86400) / 60;
  const hourAngle = mod(minutes + eqTime + 4 * lon, 1440) / 4 - 180;
  const cosZenith = sin(lat) * sin(decl) + cos(lat) * cos(decl) * cos(hourAngle);
  return 90 - deg(Math.acos(Math.max(-1, Math.min(1, cosZenith))));
}

export function sunUp(t: Date, lat: number, lon: number): boolean {
  return elevation(t, lat, lon) > HORIZON;
}

function nextChange(t: Date, lat: number, lon: number, up: boolean, limitDays: number): Date | null {
  let before = t.getTime();
  let wasUp = sunUp(t, lat, lon);
  for (let elapsed = 0; elapsed < limitDays * 86400_000; elapsed += STEP) {
    let after = before + STEP;
    const isUp = sunUp(new Date(after), lat, lon);
    if (isUp === up && wasUp !== up) {
      while (after - before > 1000) {
        const mid = before + (after - before) / 2;
        if (sunUp(new Date(mid), lat, lon) === up) after = mid;
        else before = mid;
      }
      return new Date(Math.floor(after / 1000) * 1000);
    }
    before = after;
    wasUp = isUp;
  }
  return null;
}

/** The first sunrise after t (after the coming sunset if the sun is up), or null within limitDays. */
export function nextSunrise(t: Date, lat: number, lon: number, limitDays = 366): Date | null {
  return nextChange(t, lat, lon, true, limitDays);
}

/** The first sunset after t, or null within limitDays. */
export function nextSunset(t: Date, lat: number, lon: number, limitDays = 366): Date | null {
  return nextChange(t, lat, lon, false, limitDays);
}

/**
 * Nights longer than this are the polar night, which the camera does not skip
 * (see longestNight in src/backend/timelapse/scheduler.go).
 */
export const LONGEST_NIGHT_DAYS = 1;

/** Whether it is the polar night at t: the sun is down and does not rise within LONGEST_NIGHT_DAYS. */
export function polarNight(t: Date, lat: number, lon: number): boolean {
  return !sunUp(t, lat, lon) && nextSunrise(t, lat, lon, LONGEST_NIGHT_DAYS) === null;
}

/** The share of the 24 hours from t with the sun up (0..1), for estimating photos a day. */
export function daylightFraction(t: Date, lat: number, lon: number): number {
  const samples = 24 * 6;
  let up = 0;
  for (let i = 0; i < samples; i++) {
    if (sunUp(new Date(t.getTime() + i * STEP), lat, lon)) up++;
  }
  return up / samples;
}
