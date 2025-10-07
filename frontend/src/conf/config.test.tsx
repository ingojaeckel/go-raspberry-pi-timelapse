import { BaseUrl } from './config';

describe('config', () => {
  it('exports BaseUrl', () => {
    expect(BaseUrl).toBeDefined();
  });

  it('BaseUrl is a string', () => {
    expect(typeof BaseUrl).toBe('string');
  });

  it('BaseUrl is either empty or localhost URL', () => {
    expect(BaseUrl === '' || BaseUrl === 'http://localhost:8080').toBe(true);
  });

  it('BaseUrl matches expected pattern based on environment', () => {
    if (process.env.NODE_ENV === 'production') {
      expect(BaseUrl).toBe('');
    } else {
      expect(BaseUrl).toBe('http://localhost:8080');
    }
  });
});
