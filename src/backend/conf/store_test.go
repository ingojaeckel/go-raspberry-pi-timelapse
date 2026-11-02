package conf

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func openTestStore(t *testing.T, path string) *Store {
	t.Helper()
	s, err := OpenStore(path, Default(1800))
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestStoreCreatesDefaultsWhenMissing(t *testing.T) {
	s := openTestStore(t, filepath.Join(t.TempDir(), "settings.db"))
	if got := s.Get(); got != Default(1800) {
		t.Fatalf("Get() = %s, want defaults", got)
	}
}

func TestStoreUsesGivenDefaultsOnFirstRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.db")
	s, err := OpenStore(path, Default(600))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if got := s.Get().SecondsBetweenCaptures; got != 600 {
		t.Fatalf("interval = %d, want the -secondsBetweenCaptures default 600", got)
	}
}

func TestStorePersistsUpdatesAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.db")
	s, err := OpenStore(path, Default(1800))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(func(c *Settings) { c.Quality = 85; c.RotateBy = 180 }); err != nil {
		t.Fatal(err)
	}
	s.Close()

	again := openTestStore(t, path)
	if got := again.Get(); got.Quality != 85 || got.RotateBy != 180 {
		t.Fatalf("reloaded %s", got)
	}
}

func TestStoreRejectsInvalidUpdateAndKeepsOldSettings(t *testing.T) {
	s := openTestStore(t, filepath.Join(t.TempDir(), "settings.db"))
	before := s.Get()
	_, err := s.Update(func(c *Settings) { c.SecondsBetweenCaptures = 1 })
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
	if s.Get() != before {
		t.Fatalf("settings changed after rejected update: %s", s.Get())
	}
}

func TestStoreUpdateDerivesPhotoResolution(t *testing.T) {
	s := openTestStore(t, filepath.Join(t.TempDir(), "settings.db"))
	got, err := s.Update(func(c *Settings) { c.ResolutionSetting = 2 })
	if err != nil {
		t.Fatal(err)
	}
	if got.PhotoResolutionWidth != 1640 || got.PhotoResolutionHeight != 1232 {
		t.Fatalf("got %dx%d", got.PhotoResolutionWidth, got.PhotoResolutionHeight)
	}
}

func TestStoreNotifiesEverySubscriberOfEveryUpdate(t *testing.T) {
	s := openTestStore(t, filepath.Join(t.TempDir(), "settings.db"))
	a, b := s.Subscribe(), s.Subscribe()
	// Two quick updates must never leave a subscriber with stale settings:
	// the signal coalesces, Get() always returns the latest value.
	s.Update(func(c *Settings) { c.Quality = 50 })
	s.Update(func(c *Settings) { c.Quality = 60 })
	for _, ch := range []<-chan struct{}{a, b} {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("subscriber was not notified")
		}
	}
	if s.Get().Quality != 60 {
		t.Fatalf("Get() = %d, want latest update", s.Get().Quality)
	}
}

func TestStoreNormalizesInvalidStoredSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.db")
	s := openTestStore(t, path)
	// Simulate a value written by an older version that accepted anything.
	old := Default(1800)
	old.SecondsBetweenCaptures, old.OffsetWithinHour = 10, 900
	if err := s.put(old); err != nil {
		t.Fatal(err)
	}
	s.Close()

	again := openTestStore(t, path)
	if got := again.Get(); got.SecondsBetweenCaptures != 60 || got.OffsetWithinHour != 0 {
		t.Fatalf("got interval=%d offset=%d", got.SecondsBetweenCaptures, got.OffsetWithinHour)
	}
}

func TestStoreRecoversFromCorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.db")
	if err := os.WriteFile(path, []byte(strings.Repeat("garbage", 1000)), 0o600); err != nil {
		t.Fatal(err)
	}
	s := openTestStore(t, path)
	if s.Get() != Default(1800) {
		t.Fatalf("expected defaults after recovering, got %s", s.Get())
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "settings.db.corrupt-*"))
	if len(matches) != 1 {
		t.Fatalf("corrupt file should be kept aside for inspection, found %v", matches)
	}
}
