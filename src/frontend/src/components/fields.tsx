import type { ReactNode } from 'react';

/** The text under a form field, red when the field is invalid. */
export function Help({ id, invalid, children }: { id?: string; invalid?: boolean; children: ReactNode }) {
  return <p id={id} className={invalid ? 'help error' : 'help'}>{children}</p>;
}

interface CheckProps {
  label: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
}

export function Check({ label, checked, onChange }: CheckProps) {
  return (
    <label className="check">
      <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} />
      {label}
    </label>
  );
}

/** A checkbox with help text below. */
export function CheckField({ help, ...check }: CheckProps & { help: ReactNode }) {
  return (
    <div className="field">
      <Check {...check} />
      <Help>{help}</Help>
    </div>
  );
}

/** One option of a ChoiceField: a bold title above a detail line. */
export interface Choice {
  key: string | number;
  title: ReactNode;
  detail: ReactNode;
  checked: boolean;
  onSelect: () => void;
}

interface ChoiceProps {
  id: string;
  label: string;
  name: string;
  choices: Choice[];
  help: ReactNode;
}

/** Radio buttons shown as side-by-side tiles, with help text below. */
export function ChoiceField({ id, label, name, choices, help }: ChoiceProps) {
  return (
    <div className="field">
      <span className="label" id={id}>{label}</span>
      <div className="segmented" role="radiogroup" aria-labelledby={id}>
        {choices.map((c) => (
          <label key={c.key}>
            <input type="radio" name={name} checked={c.checked} onChange={c.onSelect} />
            <span>
              <strong>{c.title}</strong>
              {c.detail}
            </span>
          </label>
        ))}
      </div>
      <Help>{help}</Help>
    </div>
  );
}
