import { BaseUrl } from './config';

describe('config', () => {
  const originalEnv = process.env.NODE_ENV;

  afterEach(() => {
    process.env.NODE_ENV = originalEnv;
  });

  it('exports BaseUrl', () => {
    expect(BaseUrl).toBeDefined();
  });

  it('uses empty string for production', () => {
    process.env.NODE_ENV = 'production';
    // Need to re-import to get new value
    jest.resetModules();
    const { BaseUrl: prodUrl } = require('./config');
    expect(prodUrl).toBe('');
  });

  it('uses localhost:8080 for development', () => {
    process.env.NODE_ENV = 'development';
    jest.resetModules();
    const { BaseUrl: devUrl } = require('./config');
    expect(devUrl).toBe('http://localhost:8080');
  });

  it('uses localhost:8080 for non-production environments', () => {
    process.env.NODE_ENV = 'test';
    jest.resetModules();
    const { BaseUrl: testUrl } = require('./config');
    expect(testUrl).toBe('http://localhost:8080');
  });
});
