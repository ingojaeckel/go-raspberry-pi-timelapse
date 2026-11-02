package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
	"time"
)

func TestReclaimDiskSpaceKeepsPhotosWhenThereIsRoom(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.jpg"), nil, 0o600)
	reclaimDiskSpace(dir, 0)
	if _, err := os.Stat(filepath.Join(dir, "a.jpg")); err != nil {
		t.Fatal("photo deleted although there was enough space")
	}
}

func TestReclaimDiskSpaceDeletesOldestWhenFull(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.jpg"), nil, 0o600)
	reclaimDiskSpace(dir, 1<<62) // more than any disk: delete everything, then give up
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("left %v", entries)
	}
}

func TestSchedulerReclaimsDiskSpaceBeforeCaptures(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.jpg"), nil, 0o600)
	a := &app{opts: options{storageFolder: dir, minFreeDiskMB: 1 << 40, captureRetries: 3}}
	s := a.scheduler()
	if s.Retries != 3 || s.Folder != dir {
		t.Fatalf("scheduler %+v", s)
	}
	s.BeforeCapture()
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

func TestVersionStringNeedsCommitAndBuildTime(t *testing.T) {
	if got := versionString("abc", ""); got == "abc built at " {
		t.Fatalf("got %q", got)
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
