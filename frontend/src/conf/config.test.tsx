import { BaseUrl } from './config';

describe('config', () => {
  const originalEnv = process.env.NODE_ENV;

  afterEach(() => {
    process.env.NODE_ENV = originalEnv;
  });

  test('BaseUrl is empty string in production', () => {
    process.env.NODE_ENV = 'production';
    // Need to re-import to get the new value
    jest.resetModules();
    const { BaseUrl } = require('./config');
    expect(BaseUrl).toBe('');
  });

  test('BaseUrl is localhost:8080 in development', () => {
    process.env.NODE_ENV = 'development';
    jest.resetModules();
    const { BaseUrl } = require('./config');
    expect(BaseUrl).toBe('http://localhost:8080');
  });

  test('BaseUrl is localhost:8080 in test environment', () => {
    process.env.NODE_ENV = 'test';
    jest.resetModules();
    const { BaseUrl } = require('./config');
    expect(BaseUrl).toBe('http://localhost:8080');
  });
});
