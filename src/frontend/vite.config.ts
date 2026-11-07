/// <reference types="vitest/config" />
import { readdirSync, readFileSync, unlinkSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { gzipSync } from 'node:zlib';
import { defineConfig, loadEnv, type Plugin } from 'vite';
import react from '@vitejs/plugin-react';

// The Go server embeds the build output (src/backend/web, //go:embed dist) and
// serves it under /static/frontend/build/ (frontendPath in src/backend/main.go).
// //go:embed cannot reach outside its package, so the build is written straight
// into the backend. Both values below are part of that contract.
const BASE = '/static/frontend/build/';
const OUT_DIR = '../backend/web/dist';

/**
 * Replaces each text file of the build by name.gz (gzip -9): the Go server
 * embeds only those and sends them as they are to browsers (src/backend/web,
 * Handler). Fonts are compressed already and stay as they are.
 */
function gzipBuild(): Plugin {
  const outDir = fileURLToPath(new URL(OUT_DIR, import.meta.url));
  return {
    name: 'gzip-build',
    apply: 'build',
    closeBundle() {
      for (const file of readdirSync(outDir, { recursive: true, encoding: 'utf8' })) {
        if (!/\.(html|js|css|svg)$/.test(file)) continue;
        const path = join(outDir, file);
        writeFileSync(`${path}.gz`, gzipSync(readFileSync(path), { level: 9 }));
        unlinkSync(path);
      }
    },
  };
}

// Backend endpoints, proxied to a locally running Go server during `npm start`.
const API_PATHS = ['/photos', '/file', '/archive', '/capture', '/configuration', '/monitoring', '/logs', '/admin', '/version'];

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const backend = env.VITE_BACKEND ?? 'http://localhost:8080';

  return {
    plugins: [react(), gzipBuild()],
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
