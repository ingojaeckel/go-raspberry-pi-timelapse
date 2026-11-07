package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

// diskApp is an app whose photo folder pretends to be on a disk with free
// bytes available.
func diskApp(t *testing.T, minFreeMB int, free uint64) (*app, string) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.jpg"), nil, 0o600)
	store, err := conf.OpenStore(filepath.Join(t.TempDir(), "settings.json"), conf.Default(600))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	a := &app{
		opts:      options{storageFolder: dir, minFreeDiskMB: minFreeMB, captureRetries: 3},
		store:     store,
		photos:    files.NewIndex(dir),
		freeBytes: func(string) (uint64, error) { return free, nil },
		started:   time.Now(),
	}
	return a, dir
}

// captureLogs collects what is logged until the test ends.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func TestReclaimDiskSpaceKeepsPhotosWhenThereIsRoom(t *testing.T) {
	a, dir := diskApp(t, 1, 1<<20)
	a.reclaimDiskSpace()
	if _, err := os.Stat(filepath.Join(dir, "a.jpg")); err != nil {
		t.Fatal("photo deleted although there was enough space")
	}
}

func TestReclaimDiskSpaceDeletesOldestWhenFull(t *testing.T) {
	a, dir := diskApp(t, 1, 0) // never enough: delete everything, then give up
	a.reclaimDiskSpace()
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("left %v", entries)
	}
	if a.deleted.Load() != 1 || a.record.Summary().DeletedOldest != 1 {
		t.Fatalf("deleted %d, summary %+v", a.deleted.Load(), a.record.Summary())
	}
}

func TestStopWhenDiskFullKeepsThePhotosAndSkipsTheCapture(t *testing.T) {
	a, dir := diskApp(t, 1, 0)
	a.store.Update(func(s *conf.Settings) { s.StopWhenDiskFull = true })
	err := a.scheduler().BeforeCapture()
	if !errors.Is(err, errDiskFull) {
		t.Fatalf("got %v, want errDiskFull", err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("photos deleted although StopWhenDiskFull is set: %v", entries)
	}
	if a.deleted.Load() != 0 {
		t.Fatalf("deleted %d", a.deleted.Load())
	}
}

func TestStopWhenDiskFullCapturesWhileThereIsRoom(t *testing.T) {
	a, _ := diskApp(t, 1, 2<<20)
	a.store.Update(func(s *conf.Settings) { s.StopWhenDiskFull = true })
	if err := a.prepareDisk(); err != nil {
		t.Fatal(err)
	}
	a.freeBytes = func(string) (uint64, error) { return 0, errors.New("statfs failed") }
	if err := a.prepareDisk(); err != nil {
		t.Fatalf("an unknown free space must not stop the camera: %v", err)
	}
}

func TestSampleHealth(t *testing.T) {
	logs := captureLogs(t)
	a, _ := diskApp(t, 1, 5<<20)
	a.thermalFile = filepath.Join(t.TempDir(), "temp")
	os.WriteFile(a.thermalFile, []byte("61500\n"), 0o600)
	a.runCommand = func(context.Context, string, ...string) ([]byte, error) { return []byte("throttled=0x50005\n"), nil }
	a.sampleHealth(t.Context())
	if !strings.Contains(logs.String(), "power or temperature problem") {
		t.Fatalf("under-voltage not logged:\n%s", logs)
	}
	a.record.Captured(nil)
	a.record.Captured(errors.New("no camera"))
	a.logSummary()
	for _, want := range []string{"level=WARN msg=\"daily summary\"", "captured=1", "failed=1", "lastError=\"no camera\"", "minCelsius=61.5", "photos=1", "freeMB=5", "throttledSinceBoot=\"[under-voltage throttled]\""} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("summary lacks %s:\n%s", want, logs)
		}
	}
}

func TestDailySummaryOfAHealthyCamera(t *testing.T) {
	logs := captureLogs(t)
	a, _ := diskApp(t, 1, 5<<20) // no sensors, no commands
	a.sampleHealth(t.Context())
	a.record.Captured(nil)
	a.logSummary()
	if !strings.Contains(logs.String(), "level=INFO msg=\"daily summary\" captured=1 failed=0") || strings.Contains(logs.String(), "Celsius") {
		t.Fatalf("summary:\n%s", logs)
	}
}

func TestWatchHealthSummarizesEveryDay(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		logs := captureLogs(t)
		a, _ := diskApp(t, 1, 5<<20)
		samples := 0
		a.runCommand = func(context.Context, string, ...string) ([]byte, error) {
			samples++
			return []byte("throttled=0x0\n"), nil
		}
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan struct{})
		go func() { defer close(done); a.watchHealth(ctx) }()
		time.Sleep(2*summaryInterval + time.Minute)
		cancel()
		<-done
		if n := strings.Count(logs.String(), "daily summary"); n != 2 {
			t.Fatalf("%d summaries in two days:\n%s", n, logs)
		}
		if want := int(2*summaryInterval/healthSampleInterval) + 1; samples != want {
			t.Fatalf("%d samples, want %d", samples, want)
		}
	})
}

func TestMarkRunReportsHowThePreviousRunEnded(t *testing.T) {
	t.Chdir(t.TempDir())
	logs := captureLogs(t)
	a := &app{}
	a.markRun()
	if !strings.Contains(logs.String(), "previous run stopped cleanly") {
		t.Fatalf("first start:\n%s", logs)
	}
	// Killed: close never ran, the marker is still there.
	b := &app{}
	b.markRun()
	if !strings.Contains(logs.String(), "previous run did not stop cleanly") {
		t.Fatalf("after a kill:\n%s", logs)
	}
	b.close()
	if _, err := os.Stat(runMarkerFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("marker left after a clean stop: %v", err)
	}
	os.Mkdir(runMarkerFile, 0o700) // unreadable: warn, but start anyway
	if err := (&app{}).markRun(); err != nil || !strings.Contains(logs.String(), "cannot tell how the previous run ended") {
		t.Fatalf("err %v\n%s", err, logs)
	}
}

func TestSchedulerReclaimsDiskSpaceBeforeCaptures(t *testing.T) {
	a, dir := diskApp(t, 1, 0)
	s := a.scheduler()
	if s.Retries != 3 || s.Folder != dir {
		t.Fatalf("scheduler %+v", s)
	}
	if err := s.BeforeCapture(); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("BeforeCapture did not reclaim space: %v", entries)
	}
}

func TestHandlerServesPprofOnlyWhenEnabled(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		h := (&app{opts: options{pprof: enabled}}).handler()
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "http://localhost/debug/pprof/", nil))
		if got := rec.Code == http.StatusOK; got != enabled {
			t.Errorf("pprof enabled=%v: status %d", enabled, rec.Code)
		}
	}
}

func TestNewAppReleasesResourcesWhenAStepFails(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile("not-a-dir", nil, 0o600)
	_, err := newApp(options{storageFolder: "not-a-dir/photos", secondsBetweenCaptures: 600, listen: "127.0.0.1:0"}, noCamera{})
	if err == nil {
		t.Fatal("expected an error for a storage folder below a file")
	}
	// The settings database was closed again, so it can be opened right away.
	a, err := newApp(options{storageFolder: "photos", secondsBetweenCaptures: 600, listen: "127.0.0.1:0"}, noCamera{})
	if err != nil {
		t.Fatalf("settings still locked after the failed start: %v", err)
	}
	a.close()
}

func TestServeReturnsServeErrors(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ln.Close() // Serve fails immediately
	if err := serve(context.Background(), &http.Server{}, ln, nil); err == nil {
		t.Fatal("expected the Serve error")
	}
}

func TestSetupLoggingToFile(t *testing.T) {
	t.Chdir(t.TempDir())
	defer slog.SetDefault(slog.Default())
	closeLog, err := setupLogging(true, new(slog.LevelVar))
	if err != nil {
		t.Fatal(err)
	}
	slog.Info("hello file")
	closeLog()
	if b, _ := os.ReadFile("timelapse.log"); len(b) == 0 {
		t.Fatal("nothing logged to the file")
	}
}

func TestSetupLoggingFailsWhenTheFileCannotBeCreated(t *testing.T) {
	t.Chdir(t.TempDir())
	os.Mkdir("timelapse.log", 0o700) // a directory where the file should be
	if _, err := setupLogging(true, new(slog.LevelVar)); err == nil {
		t.Fatal("expected an error")
	}
}

func TestLogLevel(t *testing.T) {
	if logLevel(true) != slog.LevelDebug || logLevel(false) != slog.LevelInfo {
		t.Fatal("wrong levels")
	}
}

func TestVcsVersion(t *testing.T) {
	cases := []struct {
		settings []debug.BuildSetting
		want     string
	}{
		{nil, ""},
		{[]debug.BuildSetting{{Key: "vcs.revision", Value: "0123456789abcdef"}, {Key: "vcs.time", Value: "2026-11-02T10:00:00Z"}}, "0123456 built at 2026-11-02T10:00:00Z"},
		{[]debug.BuildSetting{{Key: "vcs.revision", Value: "0123456789abcdef"}, {Key: "vcs.modified", Value: "true"}, {Key: "vcs.time", Value: "t"}}, "0123456-dirty built at t"},
	}
	for _, c := range cases {
		if got := vcsVersion(c.settings); got != c.want {
			t.Errorf("vcsVersion(%v) = %q, want %q", c.settings, got, c.want)
		}
	}
}

func TestRunReturnsWhenServingFailsOnItsOwn(t *testing.T) {
	t.Chdir(t.TempDir())
	a, err := newApp(options{storageFolder: "photos", secondsBetweenCaptures: 600, listen: "127.0.0.1:0"}, noCamera{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.close()
	a.listener.Close() // Serve fails although nobody cancelled ctx

	done := make(chan error, 1)
	go func() { done <- a.run(context.Background(), nil) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected the Serve error, so main exits and systemd restarts the service")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("run hung after the HTTP server failed")
	}
}

// listingCamera is a camera whose rpicam-still lists an IMX290.
type listingCamera struct{ noCamera }

func (listingCamera) Sensors() *timelapse.Sensors {
	return timelapse.NewSensors(func(context.Context, string, ...string) ([]byte, error) {
		return []byte("0 : imx290 [1920x1080 12-bit RGGB] (/base/soc/i2c0mux/i2c@1/imx290@1a)\n" +
			"    Modes: 'SRGGB10_CSI2P' : 1280x720 [30.00 fps - (65535, 65535)/65535x65535 crop]\n" +
			"                             1920x1080 [30.00 fps - (65535, 65535)/65535x65535 crop]\n"), nil
	})
}

func TestFirstRunUsesTheCamerasDefaultSize(t *testing.T) {
	t.Chdir(t.TempDir())
	a, err := newApp(options{storageFolder: "photos", secondsBetweenCaptures: 600, listen: "127.0.0.1:0"}, listingCamera{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.close()
	if s := a.store.Get(); s.PhotoResolutionWidth != 1920 || s.PhotoResolutionHeight != 1080 {
		t.Fatalf("first run saved %dx%d", s.PhotoResolutionWidth, s.PhotoResolutionHeight)
	}
	sensor, err := a.api().Sensor(t.Context())
	if err != nil || sensor.Model != "imx290" {
		t.Fatalf("API camera: %+v %v", sensor, err)
	}
}

func TestSavedUnsupportedSizeIsKeptAndLogged(t *testing.T) {
	t.Chdir(t.TempDir())
	s, err := conf.OpenStore(settingsFile, conf.Default(600)) // 2186x1642
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })
	a, err := newApp(options{storageFolder: "photos", secondsBetweenCaptures: 600, listen: "127.0.0.1:0", logToFile: true}, listingCamera{})
	if err != nil {
		t.Fatal(err)
	}
	got := a.store.Get()
	a.close()
	if got.PhotoResolutionWidth != 2186 {
		t.Fatalf("saved size changed to %dx%d", got.PhotoResolutionWidth, got.PhotoResolutionHeight)
	}
	if logs, _ := os.ReadFile(logFile); !strings.Contains(string(logs), "not one of the camera's sizes") {
		t.Fatalf("logs:\n%s", logs)
	}
}

// brokenCamera is a camera whose rpicam-still cannot list cameras.
type brokenCamera struct{ noCamera }

func (brokenCamera) Sensors() *timelapse.Sensors {
	return timelapse.NewSensors(func(context.Context, string, ...string) ([]byte, error) {
		return nil, errors.New("no cameras available")
	})
}

func TestFailedCameraDetectionKeepsTheDefaultSizeAndWarns(t *testing.T) {
	t.Chdir(t.TempDir())
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })
	a, err := newApp(options{storageFolder: "photos", secondsBetweenCaptures: 600, listen: "127.0.0.1:0", logToFile: true}, brokenCamera{})
	if err != nil {
		t.Fatal(err)
	}
	got := a.store.Get()
	a.close()
	if want := conf.Default(600); got.PhotoResolutionWidth != want.PhotoResolutionWidth || got.PhotoResolutionHeight != want.PhotoResolutionHeight {
		t.Fatalf("first run saved %dx%d, want the default %dx%d", got.PhotoResolutionWidth, got.PhotoResolutionHeight, want.PhotoResolutionWidth, want.PhotoResolutionHeight)
	}
	if logs, _ := os.ReadFile(logFile); !strings.Contains(string(logs), "camera not detected") {
		t.Fatalf("logs:\n%s", logs)
	}
}

func TestWithoutCameraDetectionTheAPIHasNoSensor(t *testing.T) {
	t.Chdir(t.TempDir())
	a, err := newApp(options{storageFolder: "photos", secondsBetweenCaptures: 600, listen: "127.0.0.1:0"}, noCamera{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.close()
	if a.api().Sensor != nil {
		t.Fatal("Sensor set for a backend that cannot list cameras")
	}
}

// With StopWhenDiskFull, exactly -minFreeDiskMB free still captures; one
// byte less skips the slot.
func TestStopWhenDiskFullThreshold(t *testing.T) {
	for free, wantFull := range map[uint64]bool{8<<20 - 1: true, 8 << 20: false} {
		a, _ := diskApp(t, 8, free)
		a.store.Update(func(s *conf.Settings) { s.StopWhenDiskFull = true })
		if err := a.prepareDisk(); errors.Is(err, errDiskFull) != wantFull {
			t.Errorf("%d bytes free of 8 MB: got %v", free, err)
		}
	}
}
