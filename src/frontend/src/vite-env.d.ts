/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE?: string;
}

/** Injected by vite.config.ts from build.sh's environment. */
declare const __BUILD_INFO__: { sha: string; shaShort: string; time: string };
