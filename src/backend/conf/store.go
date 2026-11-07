package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ErrInUse means another Store, i.e. another running app, has the settings
// open.
var ErrInUse = errors.New("the settings are in use by another running app")

// errDamaged means the settings file is not valid JSON.
var errDamaged = errors.New("the settings file is damaged")

// Store owns the settings: a JSON file, replaced atomically on every update,
// so that after a power cut it holds either the old or the new settings. It
// serves the current value from memory and notifies subscribers whenever it
// changes. While open it holds a lock (path.lock), so only one app at a time
// runs the camera. It is safe for concurrent use.
type Store struct {
	path    string
	lock    *os.File
	current atomic.Pointer[Settings]

	mu   sync.Mutex // serializes Update, guards subs
	subs []chan struct{}
}

// OpenStore opens the settings file at path, creating it with defaults if
// nothing was saved yet. Invalid saved settings are repaired (Normalized). A
// damaged file is moved aside to path.corrupt-<unix time> and replaced by the
// defaults so the camera keeps working; a file that cannot be read (e.g.
// permissions) is left alone and the error returned, as its settings may be
// fine.
func OpenStore(path string, defaults Settings) (*Store, error) {
	lock, err := lockFile(path)
	if err != nil {
		return nil, fmt.Errorf("open settings: %w", err)
	}
	s := &Store{path: path, lock: lock}
	current, err := s.initialSettings(defaults)
	if err != nil {
		lock.Close()
		return nil, fmt.Errorf("open settings: %w", err)
	}
	s.current.Store(&current)
	return s, nil
}

// initialSettings loads the saved settings, or the defaults if there are
// none, and saves them back if they had to be repaired.
func (s *Store) initialSettings(defaults Settings) (Settings, error) {
	saved, err := readFile(s.path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		slog.Info("no settings saved yet, using the defaults")
		return s.save(defaults.Normalized())
	case errors.Is(err, errDamaged):
		if err := moveAside(s.path, err); err != nil {
			return Settings{}, err
		}
		return s.save(defaults.Normalized())
	case err != nil:
		return Settings{}, err
	}
	if repaired := saved.Normalized(); repaired != saved {
		slog.Warn("saved settings were invalid and have been adjusted", "from", saved, "to", repaired)
		return s.save(repaired)
	}
	return saved, nil
}

func (s *Store) save(v Settings) (Settings, error) {
	return v, writeSettings(s.path, v)
}

// Close releases the lock; the settings are saved already.
func (s *Store) Close() error { return s.lock.Close() }

// Get returns the current settings.
func (s *Store) Get() Settings { return *s.current.Load() }

// Subscribe returns a channel that receives a value after each successful
// Update. Signals coalesce: a subscriber that is busy sees one pending signal
// and should then call Get for the latest settings.
func (s *Store) Subscribe() <-chan struct{} {
	ch := make(chan struct{}, 1)
	s.mu.Lock()
	s.subs = append(s.subs, ch)
	s.mu.Unlock()
	return ch
}

// Update applies change to a copy of the current settings, validates and
// saves the result and notifies subscribers. Validation errors wrap
// ErrInvalid and leave the settings as they were.
func (s *Store) Update(change func(*Settings)) (Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.Get()
	change(&next)
	if err := next.Validate(); err != nil {
		return Settings{}, err
	}
	if err := writeSettings(s.path, next); err != nil {
		return Settings{}, fmt.Errorf("save settings: %w", err)
	}
	s.current.Store(&next)
	slog.Info("settings updated", "settings", next)
	for _, ch := range s.subs {
		select {
		case ch <- struct{}{}:
		default: // a signal is already pending
		}
	}
	return next, nil
}

// ReadSettings reads the settings saved at path, repaired like OpenStore
// would, without creating, changing or locking anything: for tools that run
// next to the app. It fails with os.ErrNotExist when nothing was saved yet.
func ReadSettings(path string) (Settings, error) {
	v, err := readFile(path)
	if err != nil {
		return Settings{}, err
	}
	return v.Normalized(), nil
}

// InUse reports whether a Store has the settings at path open, e.g. the
// running app.
func InUse(path string) bool {
	f, err := os.OpenFile(path+".lock", os.O_RDWR, 0)
	if err != nil {
		return false // never opened
	}
	defer f.Close() // releases the lock if tryLock got it
	return errors.Is(tryLock(f), ErrInUse)
}

func readFile(path string) (Settings, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, err
	}
	var v Settings
	if err := json.Unmarshal(b, &v); err != nil {
		return Settings{}, fmt.Errorf("%w: %v", errDamaged, err)
	}
	return v, nil
}

func moveAside(path string, why error) error {
	aside := fmt.Sprintf("%s.corrupt-%d", path, time.Now().Unix())
	slog.Error("settings file is damaged, starting with the defaults", "path", path, "movedTo", aside, "err", why)
	if err := os.Rename(path, aside); err != nil {
		return fmt.Errorf("%w (and moving it aside failed: %v)", why, err)
	}
	return nil
}

// writeSettings saves v as indented JSON, readable and editable by hand.
func writeSettings(path string, v Settings) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(path, append(b, '\n'))
}

// writeFileAtomic replaces path with data: written to a temporary file in
// the same folder, synced, renamed over path, and the folder synced, so that
// after a power cut path is either the old or the new version.
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name()) // fails harmlessly once renamed
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err == nil {
		err = syncDir(dir)
	}
	return err
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

// lockFile opens path.lock and locks it, or fails with ErrInUse.
func lockFile(path string) (*os.File, error) {
	f, err := os.OpenFile(path+".lock", os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	if err := tryLock(f); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

// tryLock takes an exclusive lock on f without waiting; closing f releases
// it, as does the process ending, however it ends.
func tryLock(f *os.File) error {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return ErrInUse
	}
	return err
}
