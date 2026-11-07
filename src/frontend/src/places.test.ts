import { PLACES, placeAt } from './places';
import { validLocation } from './sun';

describe('places', () => {
  it('are valid locations with unique names and coordinates', () => {
    for (const p of PLACES) expect(validLocation(p.lat, p.lon), p.name).toBe(true);
    expect(new Set(PLACES.map((p) => p.name)).size).toBe(PLACES.length);
    // placeAt shows a saved location as its place, so coordinates must be unique (and never 0, 0: no location).
    expect(new Set(PLACES.map((p) => `${p.lat},${p.lon}`)).size).toBe(PLACES.length);
    expect(placeAt(0, 0)).toBeUndefined();
  });

  it('are sorted by name', () => {
    const names = PLACES.map((p) => p.name);
    expect(names).toEqual([...names].sort((a, b) => a.localeCompare(b)));
  });

  it('find a place by its exact coordinates', () => {
    expect(placeAt(53.41, -8.24)?.name).toBe('Ireland');
    expect(placeAt(53.4, -8.24)).toBeUndefined();
  });
});
