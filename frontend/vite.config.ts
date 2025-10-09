import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  
  return {
    plugins: [react()],
    base: mode === 'production' ? '/static/frontend/build/' : '/',
    build: {
      outDir: 'build',
      sourcemap: true,
    },
    define: {
      // Make environment variables available at build time
      'process.env.REACT_APP_GIT_SHA': JSON.stringify(env.REACT_APP_GIT_SHA || ''),
      'process.env.REACT_APP_GIT_SHA_ABBREV': JSON.stringify(env.REACT_APP_GIT_SHA_ABBREV || ''),
      'process.env.REACT_APP_COMMIT_TIME': JSON.stringify(env.REACT_APP_COMMIT_TIME || ''),
    },
    test: {
      globals: true,
      environment: 'jsdom',
      setupFiles: './src/setupTests.ts',
      exclude: ['e2e/**', 'node_modules/**'],
      css: {
        include: /.+/,
      },
      server: {
        deps: {
          inline: ['@mui/x-data-grid'],
        },
      },
    },
  };
});
