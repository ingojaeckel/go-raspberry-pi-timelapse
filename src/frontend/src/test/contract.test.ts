import { describe, expect, it } from 'vitest';
import shapes from './api-shapes.json';
import { mockServer } from './mockServer';

// api-shapes.json is written by the Go test TestAPIShapesMatchTheFrontendFixtures
// from the real handlers. The fake server may leave out fields (e.g. optional
// errors), but every field it sends must exist in Go with the same JSON type.

type Shape = string | Shape[] | { [key: string]: Shape };

/** Lists where `value` differs from `shape`: unknown fields and wrong types. */
function mismatches(value: unknown, shape: Shape, path = ''): string[] {
  if (Array.isArray(shape)) {
    if (!Array.isArray(value)) return [`${path}: ${typeName(value)}, Go sends an array`];
    const [itemShape] = shape;
    if (itemShape === undefined) return [];
    return value.flatMap((item, i) => mismatches(item, itemShape, `${path}[${i}]`));
  }
  if (typeof shape === 'object') {
    if (typeName(value) !== 'object') return [`${path}: ${typeName(value)}, Go sends an object`];
    return Object.entries(value as object).flatMap(([key, field]) =>
      shape[key] === undefined ? [`${path}.${key}: Go has no such field`] : mismatches(field, shape[key], `${path}.${key}`),
    );
  }
  return typeName(value) === shape ? [] : [`${path}: ${typeName(value)}, Go sends ${shape}`];
}

function typeName(value: unknown): string {
  if (value === null) return 'null';
  if (Array.isArray(value)) return 'array';
  return typeof value;
}

describe('mismatches', () => {
  it('reports renamed fields, wrong types and arrays of the wrong kind', () => {
    const shape: Shape = { A: 'number', L: [{ B: 'string' }] };
    expect(mismatches({ A: 1, L: [{ B: 'x' }] }, shape)).toEqual([]);
    expect(mismatches({ a: 1 }, shape)).toEqual(['.a: Go has no such field']);
    expect(mismatches({ A: '1', L: [{ B: 2 }] }, shape)).toEqual(['.A: string, Go sends number', '.L[0].B: number, Go sends string']);
    expect(mismatches({ L: {} }, shape)).toEqual(['.L: object, Go sends an array']);
  });
});

describe('the fake server answers like the Go handlers', () => {
  it.each(Object.entries(shapes as Record<string, Shape>))('%s', async (route, shape) => {
    const [method, path] = route.split(' ') as [string, string];
    const body = method === 'POST' ? JSON.stringify({ UnixMillis: 1_800_000_000_000 }) : undefined;
    const response = await mockServer()(path, { method, body });
    expect(response.status).toBe(200);
    expect(mismatches(await response.json(), shape)).toEqual([]);
  });
});
