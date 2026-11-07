package timelapse

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

// fakeTool puts a shell script named name on PATH that records its
// arguments and writes "jpeg" to its last argument (the output file).
func fakeTool(t *testing.T, name string) (argsFile string) {
	t.Helper()
	dir := t.TempDir()
	argsFile = filepath.Join(dir, "args")
	script := "#!/bin/sh\necho \"$@\" > " + argsFile + "\nfor a; do last=$a; done\necho jpeg > \"$last\"\n"
	if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return argsFile
}

func TestRpicamStillBackendRunsTheCommand(t *testing.T) {
	args := fakeTool(t, "rpicam-still")
	out := filepath.Join(t.TempDir(), "x.jpg")
	if err := (rpicamStill{}).Capture(t.Context(), out, Options{Width: 2, Height: 1, Quality: 9}); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(args); strings.TrimSpace(string(got)) != "--width 2 --height 1 --quality 9 --output "+out {
		t.Fatalf("args %q", got)
	}
	if b, _ := os.ReadFile(out); string(b) != "jpeg\n" {
		t.Fatal("no photo written")
	}
}

func TestFfmpegBackendRunsTheCommand(t *testing.T) {
	args := fakeTool(t, "ffmpeg")
	out := filepath.Join(t.TempDir(), "x.jpg")
	if err := (ffmpeg{goos: "linux"}).Capture(t.Context(), out, Options{Width: 2, Height: 1, Quality: 9}); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(args); !strings.HasPrefix(string(got), "-f v4l2") {
		t.Fatalf("args %q", got)
	}
}

func TestCaptureWithDetectedBackend(t *testing.T) {
	fakeTool(t, "rpicam-still")
	path, err := NewCamera(DetectBackend(), time.Minute).Capture(t.Context(), t.TempDir(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); string(b) != "jpeg\n" {
		t.Fatal("photo not stored under its final name")
	}
}

func TestCaptureFailsWhenTheFolderCannotBeCreated(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	os.WriteFile(file, nil, 0o600)
	if _, err := NewCamera(backendFunc(writeJPEG), time.Minute).Capture(t.Context(), filepath.Join(file, "photos"), Options{}); err == nil {
		t.Fatal("expected an error")
	}
}

func TestFreePathReportsFileSystemErrors(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	os.WriteFile(file, nil, 0o600)
	if _, err := freePath(file, time.Now()); err == nil { // "file" is not a directory
		t.Fatal("expected an error")
	}
}

func TestCaptureFailsWhenRenameFails(t *testing.T) {
	dir := t.TempDir()
	cam := NewCamera(backendFunc(func(context.Context, string, Options) error {
		return nil // claims success without writing the partial file
	}), time.Minute)
	if _, err := cam.Capture(t.Context(), dir, Options{}); err == nil {
		t.Fatal("expected the rename error")
	}
}

func TestPhotoAndPreviewOptions(t *testing.T) {
	s := conf.Settings{PhotoResolutionWidth: 3, PhotoResolutionHeight: 2, PreviewResolutionWidth: 30, PreviewResolutionHeight: 20, Quality: 7, RotateBy: 180}
	if got := PhotoOptions(s); got != (Options{Width: 3, Height: 2, Quality: 7, Rotate: true}) {
		t.Errorf("photo %+v", got)
	}
	if got := PreviewOptions(s); got != (Options{Width: 30, Height: 20, Quality: 7, Rotate: true}) {
		t.Errorf("preview %+v", got)
	}
}

func TestSchedulerDefaultsToLocalTime(t *testing.T) {
	if (&Scheduler{}).location() != time.Local {
		t.Fatal("expected time.Local")
	}
}

func TestRetriesStopWhenCancelledDuringTheDelay(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cam := &fakeCamera{start: time.Now(), failures: 100}
		s := &Scheduler{Settings: &fakeSettings{}, Camera: cam, Retries: 5, RetryDelay: time.Minute}
		done := make(chan struct{})
		go func() {
			s.captureWithRetries(ctx)
			close(done)
		}()
		time.Sleep(time.Second) // first attempt failed, waiting for the retry
		cancel()
		<-done
		if n := len(cam.times()); n != 1 {
			t.Fatalf("%d attempts after cancelling", n)
		}
	})
}

// cancellingCamera fails because the capture was cancelled.
type cancellingCamera struct{ cancel context.CancelFunc }

func (c cancellingCamera) Capture(ctx context.Context, _ string, _ Options) (string, error) {
	c.cancel()
	return "", errors.New("killed")
}

func TestRetriesStopWhenTheCaptureWasCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	var attempts int
	s := &Scheduler{Settings: &fakeSettings{}, Retries: 5, BeforeCapture: func() error { attempts++; return nil }}
	s.Camera = cancellingCamera{cancel: cancel}
	s.captureWithRetries(ctx)
	if attempts != 1 {
		t.Fatalf("%d attempts, want 1", attempts)
	}
}

func TestBackendsNameTheirCommand(t *testing.T) {
	if got := fmt.Sprint(rpicamStill{}, " ", ffmpeg{goos: "linux"}); got != "rpicam-still ffmpeg" {
		t.Fatalf("got %q", got)
	}
}
