import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from 'react';
import { errorMessage } from './api';

export interface Resource<T> {
  data: T | undefined;
  error: string | null;
  loading: boolean;
  reload: () => void;
  set: (value: T) => void;
}

/** Loads data on mount and on `reload()`. Stale responses are dropped. */
export function useResource<T>(load: (signal: AbortSignal) => Promise<T>): Resource<T> {
  const [data, setData] = useState<T>();
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [tick, setTick] = useState(0);
  const loadRef = useRef(load);
  loadRef.current = load;

  useEffect(() => {
    const ctrl = new AbortController();
    setLoading(true);
    loadRef
      .current(ctrl.signal)
      .then((d) => {
        setData(d);
        setError(null);
      })
      .catch((e: unknown) => {
        if (ctrl.signal.aborted) return;
        setError(errorMessage(e));
      })
      .finally(() => {
        if (!ctrl.signal.aborted) setLoading(false);
      });
    return () => ctrl.abort();
  }, [tick]);

  const reload = useCallback(() => setTick((t) => t + 1), []);
  const set = useCallback((v: T) => setData(v), []);
  return { data, error, loading, reload, set };
}

/** Whether an async action is running, and a function that runs one and tracks it. */
export function useBusy(): [boolean, (action: () => Promise<void>) => Promise<void>] {
  const [busy, setBusy] = useState(false);
  const run = async (action: () => Promise<void>) => {
    setBusy(true);
    try {
      await action();
    } finally {
      setBusy(false);
    }
  };
  return [busy, run];
}

/** A ref for a native <dialog>, opened as a modal while `open` and closed otherwise. */
export function useModal(open: boolean) {
  const ref = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const dialog = ref.current;
    if (open && dialog && !dialog.open) dialog.showModal();
    if (!open && dialog?.open) dialog.close();
  }, [open]);
  return ref;
}

/** Current time, updated every `ms`. */
export function useNow(ms = 1000): Date {
  const [now, setNow] = useState(() => new Date());
  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), ms);
    return () => clearInterval(id);
  }, [ms]);
  return now;
}

/** Runs `fn` every `ms` while the page is visible. */
export function useInterval(fn: () => void, ms: number | null) {
  const ref = useRef(fn);
  ref.current = fn;
  useEffect(() => {
    if (ms === null) return;
    const id = setInterval(() => {
      if (document.visibilityState === 'visible') ref.current();
    }, ms);
    return () => clearInterval(id);
  }, [ms]);
}

export const ROUTES = ['photos', 'preview', 'status', 'settings', 'logs'] as const;
export type Route = (typeof ROUTES)[number];

function subscribeHash(cb: () => void) {
  window.addEventListener('hashchange', cb);
  return () => window.removeEventListener('hashchange', cb);
}

function readRoute(): Route {
  const h = window.location.hash.replace(/^#\/?/, '');
  return (ROUTES as readonly string[]).includes(h) ? (h as Route) : 'photos';
}

/** Hash-based routing keeps the Go server untouched: everything stays on index.html. */
export function useRoute(): Route {
  return useSyncExternalStore(subscribeHash, readRoute, () => 'photos');
}
