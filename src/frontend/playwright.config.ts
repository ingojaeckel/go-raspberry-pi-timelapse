import { defineConfig } from '@playwright/test';

// Smoke test of the real Go binary serving the embedded frontend build
// (`npm run test:e2e` builds the frontend first). See e2e/serve.sh.
const port = Number(process.env.E2E_PORT ?? 18080);

export default defineConfig({
  testDir: 'e2e',
  timeout: 30_000,
  retries: process.env.CI ? 1 : 0,
  use: { baseURL: `http://127.0.0.1:${port}`, browserName: 'chromium' },
  webServer: {
    command: 'sh e2e/serve.sh',
    url: `http://127.0.0.1:${port}/version`,
    timeout: 120_000,
    reuseExistingServer: false,
    env: { E2E_PORT: String(port) },
  },
});
