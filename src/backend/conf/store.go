package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	bolt "go.etcd.io/bbolt"
	berrors "go.etcd.io/bbolt/errors"
)

var (
	bucketName  = []byte("settings")
	settingsKey = []byte("s")
)

// Store owns the settings: it keeps the database open for the lifetime of the
// process, serves the current value from memory and notifies subscribers
// whenever it changes. It is safe for concurrent use.
type Store struct {
	db      *bolt.DB
	mu      sync.Mutex // serializes Update
	current atomic.Pointer[Settings]

	subsMu sync.Mutex
	subs   []chan struct{}
}

// OpenStore opens (or creates) the settings database at path. A missing or
// unreadable value is replaced by defaults. A corrupt database file, e.g.
// after a power cut, is moved aside to path.corrupt-<unix time> and replaced
// so the camera keeps working.
func OpenStore(path string, defaults Settings) (*Store, error) {
	db, err := openOrReplaceDB(path)
	if err != nil {
		return nil, fmt.Errorf("open settings: %w", err)
	}
	s := &Store{db: db}
	current, err := s.initialSettings(defaults)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("save settings: %w", err)
	}
	s.current.Store(&current)
	return s, nil
}

// openOrReplaceDB opens the database. A file whose content bbolt cannot
// read is moved aside and replaced by a new one. A file that cannot be
// opened, locked or examined (permissions, another process holds it) is left
// alone and the error returned: its settings may be fine.
func openOrReplaceDB(path string) (*bolt.DB, error) {
	db, err := openDB(path, false)
	if err == nil || !isCorrupt(err) {
		return db, err
	}
	aside := fmt.Sprintf("%s.corrupt-%d", path, time.Now().Unix())
	slog.Error("settings database is unreadable, starting with defaults", "path", path, "movedTo", aside, "err", err)
	if rerr := os.Rename(path, aside); rerr != nil {
		return nil, fmt.Errorf("%w (and moving it aside failed: %v)", err, rerr)
	}
	return openDB(path, false)
}

// initialSettings loads the saved settings, or defaults if there are none,
// and saves them back if they had to be repaired.
func (s *Store) initialSettings(defaults Settings) (Settings, error) {
	loaded, err := s.load()
	if err != nil {
		slog.Warn("no usable saved settings, using defaults", "err", err)
		normalized := defaults.Normalized()
		return normalized, s.put(normalized)
	}
	normalized := loaded.Normalized()
	if normalized == loaded {
		return loaded, nil
	}
	slog.Warn("saved settings were invalid and have been adjusted", "from", loaded, "to", normalized)
	return normalized, s.put(normalized)
}

// isCorrupt reports whether an error from bolt.Open is about the file's
// content, as opposed to an operating system error or a lock timeout.
func isCorrupt(err error) bool {
	var pathErr *fs.PathError
	var errno syscall.Errno
	return !errors.Is(err, berrors.ErrTimeout) && !errors.As(err, &pathErr) && !errors.As(err, &errno)
}

// openDB waits at most a second for another process's lock on the file.
func openDB(path string, readOnly bool) (*bolt.DB, error) {
	return bolt.Open(path, 0o600, &bolt.Options{Timeout: time.Second, ReadOnly: readOnly})
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// Get returns the current settings.
func (s *Store) Get() Settings { return *s.current.Load() }

// Subscribe returns a channel that receives a value after each successful
// Update. Signals coalesce: a subscriber that is busy sees one pending signal
// and should then call Get for the latest settings.
func (s *Store) Subscribe() <-chan struct{} {
	ch := make(chan struct{}, 1)
	s.subsMu.Lock()
	s.subs = append(s.subs, ch)
	s.subsMu.Unlock()
	return ch
}

// Update applies change to a copy of the current settings, validates and
// persists the result and notifies subscribers. Validation errors wrap
// ErrInvalid and leave the settings as they were.
func (s *Store) Update(change func(*Settings)) (Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.Get()
	change(&next)
	if err := next.Validate(); err != nil {
		return Settings{}, err
	}
	if err := s.put(next); err != nil {
		return Settings{}, fmt.Errorf("save settings: %w", err)
	}
	s.current.Store(&next)
	slog.Info("settings updated", "settings", next)
	s.notify()
	return next, nil
}

// notify signals every subscriber without blocking.
func (s *Store) notify() {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	for _, ch := range s.subs {
		select {
		case ch <- struct{}{}:
		default: // a signal is already pending
		}
	}
}

func (s *Store) put(v Settings) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		return bucket.Put(settingsKey, b)
	})
}

func (s *Store) load() (v Settings, err error) {
	defer func() {
		// bbolt panics on some forms of page corruption.
		if r := recover(); r != nil {
			err = fmt.Errorf("reading settings: %v", r)
		}
	}()
	err = s.db.View(func(tx *bolt.Tx) error {
		var raw []byte
		if bucket := tx.Bucket(bucketName); bucket != nil {
			raw = bucket.Get(settingsKey)
		}
		if raw == nil {
			return errors.New("settings not found")
		}
		return json.Unmarshal(raw, &v)
	})
	return v, err
}

// ReadSettings reads the settings saved at path without creating, repairing
// or locking the file for writing, for tools that run next to the app. It
// fails with os.ErrNotExist when nothing was saved yet, and with a lock
// timeout while the app has the file open.
func ReadSettings(path string) (Settings, error) {
	if _, err := os.Stat(path); err != nil {
		return Settings{}, err
	}
	db, err := openDB(path, true)
	if err != nil {
		return Settings{}, err
	}
	defer db.Close()
	v, err := (&Store{db: db}).load()
	if err != nil {
		return Settings{}, err
	}
	return v.Normalized(), nil
}
