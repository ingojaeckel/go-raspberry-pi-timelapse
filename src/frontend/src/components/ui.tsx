import { useEffect, useRef, type ReactNode } from 'react';
import { Icon } from './Icon';

export function Notice({
  tone = 'info',
  children,
  onDismiss,
}: {
  tone?: 'info' | 'error' | 'ok' | 'warn';
  children: ReactNode;
  onDismiss?: () => void;
}) {
  return (
    <div className={`notice ${tone}`} role={tone === 'error' ? 'alert' : 'status'}>
      <div className="notice-body">{children}</div>
      {onDismiss && (
        <button type="button" className="icon-btn" onClick={onDismiss} aria-label="Dismiss">
          <Icon name="close" />
        </button>
      )}
    </div>
  );
}

interface ConfirmProps {
  open: boolean;
  title: string;
  children: ReactNode;
  confirmLabel: string;
  danger?: boolean;
  busy?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

/** Native <dialog>: focus trapping, Esc and backdrop handled by the browser. */
export function ConfirmDialog({ open, title, children, confirmLabel, danger, busy, onConfirm, onCancel }: ConfirmProps) {
  const ref = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const d = ref.current;
    if (!d) return;
    if (open && !d.open) d.showModal();
    if (!open && d.open) d.close();
  }, [open]);

  return (
    <dialog ref={ref} className="confirm" aria-labelledby="confirm-title" onCancel={(e) => busy && e.preventDefault()}
      onClose={() => open && onCancel()}
    >
      <h2 id="confirm-title">{title}</h2>
      <div className="confirm-body">{children}</div>
      <div className="confirm-actions">
        <button type="button" className="btn" onClick={onCancel} disabled={busy}>
          Cancel
        </button>
        <button
          type="button"
          className={danger ? 'btn danger' : 'btn primary'}
          onClick={onConfirm}
          disabled={busy}
          autoFocus
        >
          {confirmLabel}
        </button>
      </div>
    </dialog>
  );
}

export function PageHeader({ title, children }: { title: string; children?: ReactNode }) {
  return (
    <header className="page-header">
      <h1>{title}</h1>
      {children && <div className="page-actions">{children}</div>}
    </header>
  );
}

export function Spinner({ label }: { label: string }) {
  return (
    <div className="spinner" role="status">
      <span className="spinner-ring" aria-hidden="true" />
      {label}
    </div>
  );
}
