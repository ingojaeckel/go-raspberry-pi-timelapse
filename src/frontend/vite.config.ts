/// <reference types="vitest/config" />
import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

// The Go server embeds the build output (src/backend/web, //go:embed dist) and
// serves it under /static/frontend/build/ (frontendPath in src/backend/main.go).
// //go:embed cannot reach outside its package, so the build is written straight
// into the backend. Both values below are part of that contract.
const BASE = '/static/frontend/build/';
const OUT_DIR = '../backend/web/dist';

// Backend endpoints, proxied to a locally running Go server during `npm start`.
const API_PATHS = ['/photos', '/file', '/archive', '/capture', '/configuration', '/monitoring', '/logs', '/admin', '/version'];

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const backend = env.VITE_BACKEND ?? 'http://localhost:8080';

  return {
    plugins: [react()],
    base: BASE,
    build: {
      outDir: OUT_DIR,
      emptyOutDir: true,
      // Source maps would be embedded into the Pi binary; keep it lean.
      sourcemap: false,
      target: 'es2022',
      assetsInlineLimit: 0,
    },
    define: {
      // Populated by build.sh. Names kept for compatibility with existing scripts.
      __BUILD_INFO__: JSON.stringify({
        sha: env.REACT_APP_GIT_SHA ?? '',
        shaShort: env.REACT_APP_GIT_SHA_ABBREV ?? '',
        time: env.REACT_APP_COMMIT_TIME ?? '',
      }),
    },
    server: {
      proxy: Object.fromEntries(API_PATHS.map((p) => [p, { target: backend, changeOrigin: true }])),
    },
    test: {
      environment: 'jsdom',
      globals: true,
      setupFiles: './src/test/setup.ts',
      // e2e/ holds the Playwright smoke test (npm run test:e2e).
      include: ['src/**/*.test.{ts,tsx}'],
    },
  };
});
