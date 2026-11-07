import { api, ApiError, OFFLINE_MESSAGE } from './api';
import { json, mockServer } from './test/mockServer';

/** The error api.settings() fails with when GET /configuration answers `response`. */
async function failure(response: () => Response) {
  mockServer({ 'GET /configuration': response });
  return api.settings().catch((e: unknown) => e as ApiError);
}

describe('API errors', () => {
  it.each([
    ['a JSON string', () => json('invalid settings: quality', 400), 'invalid settings: quality', 400],
    ['plain text', () => new Response('404 page not found\n', { status: 404 }), '404 page not found', 404],
    ['an empty body', () => new Response('', { status: 500 }), 'The camera answered with HTTP 500.', 500],
    ['a JSON object', () => json({ error: 'x' }, 500), '{"error":"x"}', 500],
  ])('reads the message from %s', async (_kind, response, message, status) => {
    const err = await failure(response);
    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({ message, status });
  });

  it('says the camera is unreachable when the request fails', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));
    await expect(api.settings()).rejects.toMatchObject({ message: OFFLINE_MESSAGE, status: 0 });
  });
});
