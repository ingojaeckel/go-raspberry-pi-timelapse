package conf

import (
	"bytes"
	"errors"
	"log/slog"
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

func settingsPath(t *testing.T) string {
	return filepath.Join(t.TempDir(), "settings.json")
}

// captureLogs records slog output for the rest of the test.
func captureLogs(t *testing.T) *bytes.Buffer {
	var logs bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &logs
}

func movedAside(path string) []string {
	matches, _ := filepath.Glob(path + ".corrupt-*")
	return matches
}

func TestStoreSavesTheDefaultsOnTheFirstRun(t *testing.T) {
	logs := captureLogs(t)
	path := settingsPath(t)
	if got := openTestStore(t, path).Get(); got != Default(1800) {
		t.Fatalf("Get() = %s, want defaults", got)
	}
	if saved, err := ReadSettings(path); err != nil || saved != Default(1800) {
		t.Fatalf("saved %s, %v", saved, err)
	}
	if !strings.Contains(logs.String(), `level=INFO msg="no settings saved yet, using the defaults"`) {
		t.Fatalf("a first run is not a problem, logged:\n%s", logs)
	}
}

func TestStoreUsesGivenDefaultsOnFirstRun(t *testing.T) {
	s, err := OpenStore(settingsPath(t), Default(600))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if got := s.Get().SecondsBetweenCaptures; got != 600 {
		t.Fatalf("interval = %d, want the -secondsBetweenCaptures default 600", got)
	}
}

func TestStorePersistsUpdatesAcrossReopen(t *testing.T) {
	path := settingsPath(t)
	s, err := OpenStore(path, Default(1800))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(func(c *Settings) { c.Quality = 85; c.RotateBy = 180 }); err != nil {
		t.Fatal(err)
	}
	s.Close()

	if got := openTestStore(t, path).Get(); got.Quality != 85 || got.RotateBy != 180 {
		t.Fatalf("reloaded %s", got)
	}
}

func TestStoreWritesReadableJSONAndNoTemporaryFiles(t *testing.T) {
	path := settingsPath(t)
	s := openTestStore(t, path)
	if _, err := s.Update(func(c *Settings) { c.Quality = 85 }); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), "\n  \"Quality\": 85,\n") {
		t.Fatalf("saved %s", b)
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	for _, e := range entries {
		if name := e.Name(); name != "settings.json" && name != "settings.json.lock" {
			t.Errorf("left %s behind", name)
		}
	}
}

func TestStoreRejectsInvalidUpdateAndKeepsOldSettings(t *testing.T) {
	s := openTestStore(t, settingsPath(t))
	before := s.Get()
	_, err := s.Update(func(c *Settings) { c.SecondsBetweenCaptures = 1 })
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
	if s.Get() != before {
		t.Fatalf("settings changed after rejected update: %s", s.Get())
	}
}

func TestStoreUpdateKeepsThePhotoSize(t *testing.T) {
	s := openTestStore(t, settingsPath(t))
	got, err := s.Update(func(c *Settings) { c.PhotoResolutionWidth, c.PhotoResolutionHeight = 1640, 1232 })
	if err != nil {
		t.Fatal(err)
	}
	if got.PhotoResolutionWidth != 1640 || got.PhotoResolutionHeight != 1232 {
		t.Fatalf("got %dx%d", got.PhotoResolutionWidth, got.PhotoResolutionHeight)
	}
}

func TestStoreNotifiesEverySubscriberOfEveryUpdate(t *testing.T) {
	s := openTestStore(t, settingsPath(t))
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

func TestStoreRepairsInvalidSavedSettings(t *testing.T) {
	path := settingsPath(t)
	// Written by an older version that accepted anything.
	os.WriteFile(path, []byte(`{"SecondsBetweenCaptures":10,"OffsetWithinHour":900,"Quality":90,"PhotoResolutionWidth":1640,"PhotoResolutionHeight":1232}`), 0o600)
	if got := openTestStore(t, path).Get(); got.SecondsBetweenCaptures != 60 || got.OffsetWithinHour != 0 || got.Quality != 90 {
		t.Fatalf("got %s", got)
	}
	if saved, _ := ReadSettings(path); saved.SecondsBetweenCaptures != 60 {
		t.Fatalf("the repair was not saved: %s", saved)
	}
}

func TestStoreReplacesADamagedFile(t *testing.T) {
	for name, content := range map[string]string{"garbage": strings.Repeat("garbage", 1000), "empty": "", "cut off": `{"Quality":`} {
		t.Run(name, func(t *testing.T) {
			path := settingsPath(t)
			os.WriteFile(path, []byte(content), 0o600)
			if got := openTestStore(t, path).Get(); got != Default(1800) {
				t.Fatalf("expected defaults, got %s", got)
			}
			if aside := movedAside(path); len(aside) != 1 {
				t.Fatalf("the damaged file should be kept aside for inspection, found %v", aside)
			}
		})
	}
}

func TestStoreUpdateReportsStorageErrors(t *testing.T) {
	path := settingsPath(t)
	s := openTestStore(t, path)
	before := s.Get()
	os.RemoveAll(filepath.Dir(path)) // the SD card went away
	_, err := s.Update(func(c *Settings) { c.Quality = 50 })
	if err == nil || errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v, want a storage error", err)
	}
	if s.Get() != before {
		t.Fatal("settings changed although they were not saved")
	}
}

func TestOpenStoreFailsWhileAnotherStoreHasItOpen(t *testing.T) {
	path := settingsPath(t)
	openTestStore(t, path) // the running app
	if _, err := OpenStore(path, Default(1800)); !errors.Is(err, ErrInUse) {
		t.Fatalf("err = %v, want ErrInUse: two apps must not run the camera at once", err)
	}
}

func TestInUseWhileAStoreHasTheFileOpen(t *testing.T) {
	path := settingsPath(t)
	if InUse(path) {
		t.Fatal("in use before anything opened it")
	}
	s := openTestStore(t, path)
	if !InUse(path) {
		t.Fatal("not in use while the store is open")
	}
	s.Close()
	if InUse(path) {
		t.Fatal("still in use after Close")
	}
}

func TestOpenStoreFailsWhenADamagedFileCannotBeMovedAside(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	path := settingsPath(t)
	dir := filepath.Dir(path)
	os.WriteFile(path, []byte("garbage"), 0o600)
	os.WriteFile(path+".lock", nil, 0o600)
	os.Chmod(dir, 0o500)
	defer os.Chmod(dir, 0o700)
	if _, err := OpenStore(path, Default(1800)); err == nil {
		t.Fatal("expected an error")
	}
}

func TestOpenStoreKeepsASettingsFileItCannotRead(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	path := settingsPath(t)
	s := openTestStore(t, path)
	if _, err := s.Update(func(c *Settings) { c.Quality = 85 }); err != nil {
		t.Fatal(err)
	}
	s.Close()

	os.Chmod(path, 0o000) // e.g. wrong owner after copying it around: intact, just not readable
	if _, err := OpenStore(path, Default(1800)); err == nil {
		t.Fatal("expected an error so the service fails and the problem gets noticed")
	}
	if aside := movedAside(path); len(aside) != 0 {
		t.Fatalf("an intact settings file was moved aside: %v", aside)
	}
	os.Chmod(path, 0o600)
	if got := openTestStore(t, path).Get().Quality; got != 85 {
		t.Fatalf("saved settings lost: quality %d", got)
	}
}

func TestReadSettingsReadsWhileTheAppHasTheFileOpen(t *testing.T) {
	path := settingsPath(t)
	s := openTestStore(t, path) // stays open, like the running app
	if _, err := s.Update(func(v *Settings) { v.Quality = 42 }); err != nil {
		t.Fatal(err)
	}
	got, err := ReadSettings(path)
	if err != nil || got != s.Get() {
		t.Fatalf("got %s, %v, want %s", got, err, s.Get())
	}
}

func TestReadSettingsDoesNotCreateAMissingFile(t *testing.T) {
	path := settingsPath(t)
	if _, err := ReadSettings(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err %v, want ErrNotExist", err)
	}
	if entries, _ := os.ReadDir(filepath.Dir(path)); len(entries) != 0 {
		t.Fatalf("ReadSettings created %v", entries)
	}
}
