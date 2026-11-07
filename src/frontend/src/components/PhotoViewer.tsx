import { useEffect, useState } from 'react';
import { urls } from '../api';
import { formatClock, formatDay } from '../format';
import { useModal } from '../hooks';
import type { PhotoItem } from '../views/photoItems';
import { Icon } from './Icon';

interface Props {
  photos: PhotoItem[];
  index: number | null;
  onIndex: (i: number | null) => void;
}

/** Full-screen viewer. Photos are only fetched when opened: full-size frames are several MB each. */
export function PhotoViewer({ photos, index, onIndex }: Props) {
  const [loaded, setLoaded] = useState(false);
  const photo = index !== null ? photos[index] : undefined;
  const ref = useModal(!!photo);

  useEffect(() => setLoaded(false), [photo?.name]);

  const go = (delta: number) => {
    if (index === null) return;
    const next = index + delta;
    if (next >= 0 && next < photos.length) onIndex(next);
  };

  return (
    <dialog
      ref={ref}
      className="viewer"
      aria-label={photo ? `Photo ${photo.name}` : 'Photo'}
      onClose={() => onIndex(null)}
      onKeyDown={(e) => {
        // Newest first: "previous" in time is further down the list.
        if (e.key === 'ArrowLeft') go(1);
        if (e.key === 'ArrowRight') go(-1);
      }}
    >
      {photo && (
        <>
          <div className="viewer-bar">
            <div className="viewer-title">
              <strong>{photo.date ? `${formatDay(photo.date)}, ${formatClock(photo.date)}` : photo.name}</strong>
              <span>
                {photo.name} ({photo.size})
              </span>
            </div>
            <a className="icon-btn" href={urls.photo(photo.name)} download={photo.name} aria-label="Download photo">
              <Icon name="download" />
            </a>
            <button type="button" className="icon-btn" onClick={() => onIndex(null)} aria-label="Close">
              <Icon name="close" />
            </button>
          </div>
          <div className="viewer-stage">
            {!loaded && <div className="viewer-loading">Loading full-size photo…</div>}
            <img
              key={photo.name}
              src={urls.photo(photo.name)}
              alt={`Timelapse frame ${photo.name}`}
              onLoad={() => setLoaded(true)}
              data-loaded={loaded}
            />
          </div>
          <div className="viewer-nav">
            <button type="button" className="btn" onClick={() => go(1)} disabled={index === photos.length - 1}>
              <Icon name="prev" /> Older
            </button>
            <span className="viewer-pos">
              {(index ?? 0) + 1} of {photos.length}
            </span>
            <button type="button" className="btn" onClick={() => go(-1)} disabled={index === 0}>
              Newer <Icon name="next" />
            </button>
          </div>
        </>
      )}
    </dialog>
  );
}
