package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
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
	db, err := openDB(path)
	if err != nil && !errors.Is(err, berrors.ErrTimeout) {
		aside := fmt.Sprintf("%s.corrupt-%d", path, time.Now().Unix())
		slog.Error("settings database is unreadable, starting with defaults", "path", path, "movedTo", aside, "err", err)
		if rerr := os.Rename(path, aside); rerr != nil {
			return nil, fmt.Errorf("open settings: %w (and moving it aside failed: %v)", err, rerr)
		}
		db, err = openDB(path)
	}
	if err != nil {
		return nil, fmt.Errorf("open settings: %w", err)
	}

	s := &Store{db: db}
	loaded, err := s.load()
	if err != nil {
		slog.Warn("no usable saved settings, using defaults", "err", err)
		loaded = defaults
	}
	normalized := loaded.Normalized()
	if normalized != loaded || err != nil {
		if normalized != loaded && err == nil {
			slog.Warn("saved settings were invalid and have been adjusted", "from", loaded, "to", normalized)
		}
		if err := s.put(normalized); err != nil {
			db.Close()
			return nil, fmt.Errorf("save settings: %w", err)
		}
	}
	s.current.Store(&normalized)
	return s, nil
}

func openDB(path string) (*bolt.DB, error) {
	return bolt.Open(path, 0o600, &bolt.Options{Timeout: time.Second})
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

// Update applies change to a copy of the current settings, derives the photo
// size from ResolutionSetting, validates and persists the result and notifies
// subscribers. Validation errors wrap ErrInvalid and leave the settings as
// they were.
func (s *Store) Update(change func(*Settings)) (Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.Get()
	change(&next)
	if err := next.Validate(); err != nil {
		return Settings{}, err
	}
	next = next.withResolution()
	if err := s.put(next); err != nil {
		return Settings{}, fmt.Errorf("save settings: %w", err)
	}
	s.current.Store(&next)
	slog.Info("settings updated", "settings", next)

	s.subsMu.Lock()
	for _, ch := range s.subs {
		select {
		case ch <- struct{}{}:
		default: // a signal is already pending
		}
	}
	s.subsMu.Unlock()
	return next, nil
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
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return errors.New("settings not found")
		}
		raw := bucket.Get(settingsKey)
		if raw == nil {
			return errors.New("settings not found")
		}
		return json.Unmarshal(raw, &v)
	})
	return v, err
}
