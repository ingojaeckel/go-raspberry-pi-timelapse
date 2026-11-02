// A handful of inline icons; drawing them here avoids shipping an icon library to a Pi Zero.
const PATHS = {
  photos: 'M4 6h16v12H4zM4 15l4.5-4.5 3.5 3.5 2.5-2.5L20 17M15.5 9.5h.01',
  camera: 'M4 8h3l2-2.5h6L17 8h3v11H4zM12 16.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7z',
  status: 'M4 18a8 8 0 1 1 16 0M12 18l4-5M12 18h.01',
  settings: 'M5 7h9M18 7h1M5 17h3M12 17h7M14 4.5v5M8 14.5v5',
  logs: 'M6 4h9l3 3v13H6zM9 10h6M9 13.5h6M9 17h4',
  download: 'M12 4v11M7.5 10.5 12 15l4.5-4.5M5 19h14',
  trash: 'M5 7h14M10 7V4.5h4V7M7 7l1 13h8l1-13M10.5 11v5.5M13.5 11v5.5',
  refresh: 'M19 12a7 7 0 1 1-2.05-4.95M19 4.5V8h-3.5',
  close: 'M6 6l12 12M18 6 6 18',
  prev: 'M14.5 6 8.5 12l6 6',
  next: 'M9.5 6l6 6-6 6',
  power: 'M12 3.5v8M7.05 6.5a7 7 0 1 0 9.9 0',
  restart: 'M5 12a7 7 0 1 0 2.05-4.95M5 4.5V8h3.5',
  external: 'M14 5h5v5M19 5l-8 8M17 14v5H5V7h5',
} as const;

export type IconName = keyof typeof PATHS;

export function Icon({ name, size = 20 }: { name: IconName; size?: number }) {
  return (
    <svg
      className="icon"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.8}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      <path d={PATHS[name]} />
    </svg>
  );
}
