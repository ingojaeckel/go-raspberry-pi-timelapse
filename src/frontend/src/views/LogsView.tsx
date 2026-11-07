import { useEffect, useMemo, useRef, useState } from 'react';
import { api } from '../api';
import { Icon } from '../components/Icon';
import { Notice, PageHeader, RefreshButton, Spinner } from '../components/ui';
import { useResource } from '../hooks';

const TAIL = 400;
const PROBLEM = /\b(error|fail(ed|ure)?|panic|fatal|unable|cannot)\b/i;

export function LogsView() {
  const logs = useResource((signal) => api.logs(signal));
  const [query, setQuery] = useState('');
  const [problemsOnly, setProblemsOnly] = useState(false);
  const [all, setAll] = useState(false);
  const bottom = useRef<HTMLDivElement>(null);

  const lines = useMemo(() => (logs.data ?? '').split('\n').filter((l) => l.trim() !== ''), [logs.data]);
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return lines.filter((l) => (!problemsOnly || PROBLEM.test(l)) && (!q || l.toLowerCase().includes(q)));
  }, [lines, query, problemsOnly]);
  const shown = all ? filtered : filtered.slice(-TAIL);

  useEffect(() => {
    bottom.current?.scrollIntoView?.({ block: 'end' });
  }, [logs.data]);

  const saveUrl = useMemo(
    () => (logs.data ? URL.createObjectURL(new Blob([logs.data], { type: 'text/plain' })) : null),
    [logs.data],
  );
  useEffect(() => () => void (saveUrl && URL.revokeObjectURL(saveUrl)), [saveUrl]);

  return (
    <>
      <PageHeader title="Logs">
        {saveUrl && (
          <a className="btn ghost" href={saveUrl} download="timelapse.log">
            <Icon name="download" /> Save
          </a>
        )}
        <RefreshButton resource={logs} />
      </PageHeader>

      {logs.error && (
        <Notice tone="error">
          Couldn't read the log file: {logs.error}. The server only keeps a log file when started with{' '}
          <code>-logToFile</code>.
        </Notice>
      )}
      {logs.loading && !logs.data && <Spinner label="Loading logs…" />}

      {logs.data !== undefined && (
        <>
          <div className="log-filters">
            <input
              type="search"
              placeholder="Filter lines"
              aria-label="Filter log lines"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
            <label className="check">
              <input type="checkbox" checked={problemsOnly} onChange={(e) => setProblemsOnly(e.target.checked)} />
              Problems only
            </label>
          </div>
          <p className="muted small">
            {filtered.length > shown.length ? (
              <>
                Showing the last {shown.length.toLocaleString()} of {filtered.length.toLocaleString()} lines.{' '}
                <button type="button" className="link" onClick={() => setAll(true)}>
                  Show all
                </button>
              </>
            ) : (
              `${shown.length.toLocaleString()} lines`
            )}
          </p>
          <div className="log" tabIndex={0} aria-label="Log output">
            {shown.length === 0 ? (
              <p className="muted">{lines.length === 0 ? 'The log is empty.' : 'No lines match.'}</p>
            ) : (
              <ol>
                {shown.map((l, i) => (
                  <li key={i} data-problem={PROBLEM.test(l) || undefined}>
                    {l}
                  </li>
                ))}
              </ol>
            )}
            <div ref={bottom} />
          </div>
        </>
      )}
    </>
  );
}
