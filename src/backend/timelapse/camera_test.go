package timelapse

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestCaptureCommandArgs(t *testing.T) {
	for _, c := range []struct {
		name      string
		got, want []string
	}{
		{"rpicam-still",
			rpicamStillArgs("foo/someFile.jpg", Options{Width: 200, Height: 100, Quality: 100}),
			[]string{"--width", "200", "--height", "100", "--quality", "100", "--output", "foo/someFile.jpg"}},
		{"rpicam-still rotated",
			rpicamStillArgs("x.jpg", Options{Width: 200, Height: 100, Quality: 90, Rotate: true}),
			[]string{"--width", "200", "--height", "100", "--quality", "90", "--vflip", "--hflip", "--output", "x.jpg"}},
		{"ffmpeg on Linux, rotated",
			ffmpegArgs("linux", "x.jpg", Options{Width: 640, Height: 480, Quality: 75, Rotate: true}),
			[]string{"-f", "v4l2", "-video_size", "640x480", "-i", "/dev/video0", "-frames:v", "1", "-q:v", "75", "-vf", "hflip,vflip", "-y", "x.jpg"}},
		{"ffmpeg on macOS",
			ffmpegArgs("darwin", "x.jpg", Options{Width: 640, Height: 480, Quality: 75}),
			[]string{"-f", "avfoundation", "-framerate", "30", "-i", "0", "-frames:v", "1", "-s", "640x480", "-y", "x.jpg"}},
	} {
		if !slices.Equal(c.got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestDetectBackendPrefersRpicamStillWhenInstalled(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	if _, ok := DetectBackend().(rpicamStill); ok {
		t.Fatal("rpicam-still is not installed, expected ffmpeg")
	}
	os.WriteFile(filepath.Join(dir, "rpicam-still"), []byte("#!/bin/sh\n"), 0o755)
	if _, ok := DetectBackend().(rpicamStill); !ok {
		t.Fatal("expected rpicam-still when it is on PATH, regardless of GOARCH")
	}
}

func TestFileName(t *testing.T) {
	if got := fileName(time.Date(1974, time.December, 19, 1, 2, 3, 4, time.UTC)); got != "19741219-010203.jpg" {
		t.Fatalf("got %s", got)
	}
}

// backendFunc adapts a function to Backend.
type backendFunc func(ctx context.Context, path string, o Options) error

func (f backendFunc) Capture(ctx context.Context, path string, o Options) error {
	return f(ctx, path, o)
}

func writeJPEG(_ context.Context, path string, _ Options) error {
	return os.WriteFile(path, []byte("jpeg"), 0o600)
}

func TestCaptureWritesToAHiddenFileAndRenames(t *testing.T) {
	dir := t.TempDir()
	var during string
	cam := NewCamera(backendFunc(func(ctx context.Context, path string, o Options) error {
		during = path
		return writeJPEG(ctx, path, o)
	}), time.Minute)

	path, err := cam.Capture(t.Context(), dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(during) != dir || !strings.HasPrefix(filepath.Base(during), ".") {
		t.Fatalf("camera should write to a hidden file in %s, wrote %s", dir, during)
	}
	if filepath.Dir(path) != dir || strings.HasPrefix(filepath.Base(path), ".") || !strings.HasSuffix(path, ".jpg") {
		t.Fatalf("unexpected final path %s", path)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected only the final photo, found %v", entries)
	}
}

func TestCaptureCreatesMissingFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "photos")
	if _, err := NewCamera(backendFunc(writeJPEG), time.Minute).Capture(t.Context(), dir, Options{}); err != nil {
		t.Fatal(err)
	}
}

func TestCaptureRemovesPartialFileOnFailure(t *testing.T) {
	dir := t.TempDir()
	cam := NewCamera(backendFunc(func(_ context.Context, path string, _ Options) error {
		os.WriteFile(path, []byte("half a jpeg"), 0o600)
		return errors.New("camera failed")
	}), time.Minute)
	if _, err := cam.Capture(t.Context(), dir, Options{}); err == nil {
		t.Fatal("expected an error")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("partial file left behind: %v", entries)
	}
}

func TestCaptureNeverOverwritesAnExistingPhoto(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		dir := t.TempDir()
		existing := filepath.Join(dir, fileName(time.Now()))
		os.WriteFile(existing, []byte("older photo"), 0o600)

		path, err := NewCamera(backendFunc(writeJPEG), time.Minute).Capture(t.Context(), dir, Options{})
		if err != nil {
			t.Fatal(err)
		}
		if path == existing {
			t.Fatal("overwrote a photo taken at the same wall clock second (e.g. after the clock moved back)")
		}
		if got, _ := os.ReadFile(existing); string(got) != "older photo" {
			t.Fatal("existing photo was modified")
		}
		if !strings.HasSuffix(path, "-1.jpg") {
			t.Fatalf("expected a numbered name, got %s", path)
		}
	})
}

func TestCaptureTimesOutAHangingCamera(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cam := NewCamera(backendFunc(func(ctx context.Context, _ string, _ Options) error {
			<-ctx.Done() // rpicam-still hangs
			return ctx.Err()
		}), 30*time.Second)
		start := time.Now()
		_, err := cam.Capture(t.Context(), t.TempDir(), Options{})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err = %v, want deadline exceeded", err)
		}
		if took := time.Since(start); took != 30*time.Second {
			t.Fatalf("gave up after %v", took)
		}
	})
}

func TestCaptureUsesTheCameraOneAtATime(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var active, maxActive atomic.Int32
		cam := NewCamera(backendFunc(func(ctx context.Context, path string, o Options) error {
			n := active.Add(1)
			defer active.Add(-1)
			if n > maxActive.Load() {
				maxActive.Store(n)
			}
			time.Sleep(5 * time.Second)
			return writeJPEG(ctx, path, o)
		}), time.Minute)

		var wg sync.WaitGroup
		dir := t.TempDir()
		for range 3 {
			wg.Go(func() {
				if _, err := cam.Capture(t.Context(), dir, Options{}); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
		if maxActive.Load() != 1 {
			t.Fatalf("%d captures ran at the same time", maxActive.Load())
		}
	})
}

func TestCaptureStopsWaitingForTheCameraWhenCancelled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		cam := NewCamera(backendFunc(func(context.Context, string, Options) error {
			<-release
			return nil
		}), time.Hour)
		go cam.Capture(t.Context(), t.TempDir(), Options{})
		synctest.Wait() // first capture holds the camera

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		if _, err := cam.Capture(ctx, t.TempDir(), Options{}); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err = %v", err)
		}
		close(release)
	})
}

func TestRunCommandKillsProcessOnTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := runCommand(ctx, "sleep", "10")
	if err == nil {
		t.Fatal("expected an error")
	}
	if took := time.Since(start); took > 5*time.Second {
		t.Fatalf("took %v", took)
	}
}

func TestRunCommandIncludesOutputInError(t *testing.T) {
	err := runCommand(t.Context(), "sh", "-c", "echo camera not found >&2; exit 1")
	if err == nil || !strings.Contains(err.Error(), "camera not found") {
		t.Fatalf("err = %v", err)
	}
}

func TestRemovePartialsDeletesOnlyUnfinishedPhotos(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{".partial-20260101-000000.jpg", "20260101-000100.jpg", ".hidden"} {
		os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o600)
	}
	os.Mkdir(filepath.Join(dir, ".partial-dir"), 0o700)
	removed, err := RemovePartials(dir)
	if err != nil || len(removed) != 1 || removed[0] != ".partial-20260101-000000.jpg" {
		t.Fatalf("removed %v, %v", removed, err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 3 {
		t.Fatalf("left %v", entries)
	}
	if _, err := RemovePartials(filepath.Join(dir, "nope")); err == nil {
		t.Fatal("expected an error for a missing folder")
	}
}

func TestCaptureFailsWhenTheBackendWroteNothing(t *testing.T) {
	c := NewCamera(backendFunc(func(context.Context, string, Options) error { return nil }), time.Second)
	if _, err := c.Capture(t.Context(), t.TempDir(), Options{}); err == nil {
		t.Fatal("expected an error when there is no photo to flush")
	}
}
