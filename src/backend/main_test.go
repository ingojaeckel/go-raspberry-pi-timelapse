package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

func TestParseFlagsDefaults(t *testing.T) {
	o, err := parseFlags(nil)
	if err != nil {
		t.Fatal(err)
	}
	want := options{
		listen:                 conf.DefaultListenAddress,
		storageFolder:          conf.DefaultStorageFolder,
		secondsBetweenCaptures: conf.DefaultSecondsBetweenCaptures,
		captureRetries:         conf.DefaultCaptureRetries,
		minFreeDiskMB:          conf.DefaultMinFreeDiskMB,
	}
	if o != want {
		t.Fatalf("got %+v, want %+v", o, want)
	}
}

func TestParseFlags(t *testing.T) {
	o, err := parseFlags([]string{"-port=:9090", "-logToFile=true", "-storageFolder=/p", "-secondsBetweenCaptures=600", "-captureRetries=0", "-minFreeDiskMB=10", "-pprof"})
	if err != nil {
		t.Fatal(err)
	}
	want := options{listen: ":9090", logToFile: true, storageFolder: "/p", secondsBetweenCaptures: 600, captureRetries: 0, minFreeDiskMB: 10, pprof: true}
	if o != want {
		t.Fatalf("got %+v, want %+v", o, want)
	}
}

func TestParseFlagsRejectsInvalidValues(t *testing.T) {
	for _, args := range [][]string{
		{"-secondsBetweenCaptures=1000"},
		{"-secondsBetweenCaptures=10"},
		{"-captureRetries=-1"},
		{"-minFreeDiskMB=-1"},
		{"-unknown"},
	} {
		if _, err := parseFlags(args); err == nil {
			t.Errorf("%v: expected an error", args)
		}
	}
}

func TestVersionString(t *testing.T) {
	if got := versionString("abc123", "Mon Nov 2"); got != "abc123 built at Mon Nov 2" {
		t.Fatalf("got %q", got)
	}
	if got := versionString("", ""); got == "" {
		t.Fatal("expected a fallback version")
	}
}

type noCamera struct{}

func (noCamera) Capture(context.Context, string, timelapse.Options) error { return nil }

func TestRunServesTheAppAndShutsDownCleanly(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx, cancel := context.WithCancel(t.Context())
	addr := make(chan string, 1)
	done := make(chan error, 1)
	opts := options{listen: "127.0.0.1:0", storageFolder: "photos", secondsBetweenCaptures: 600, captureRetries: 2, minFreeDiskMB: 1}
	go func() { done <- run(ctx, opts, noCamera{}, func(a string) { addr <- a }) }()

	var base string
	select {
	case a := <-addr:
		base = "http://" + a
	case err := <-done:
		t.Fatalf("run returned early: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("server did not start")
	}

	get := func(path string) (int, string) {
		res, err := http.Get(base + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		return res.StatusCode, string(b)
	}

	if code, body := get("/"); code != 200 || !strings.Contains(body, "<html") {
		t.Errorf("/ should redirect to the frontend, got %d", code)
	}
	if code, body := get("/configuration"); code != 200 || !strings.Contains(body, `"SecondsBetweenCaptures":600`) {
		t.Errorf("-secondsBetweenCaptures must be the first-run default: %d %s", code, body)
	}
	if _, err := os.Stat("photos"); err != nil {
		t.Errorf("storage folder not created: %v", err)
	}
	if _, err := os.Stat(conf.SettingsFile); err != nil {
		t.Errorf("settings not stored in the working directory: %v", err)
	}
	if code, _ := get("/debug/pprof/"); code == 200 {
		t.Error("pprof must be off unless requested")
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run: %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("did not shut down")
	}
}

func TestRunFailsWhenThePortIsTaken(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	addr := make(chan string, 1)
	go run(ctx, options{listen: "127.0.0.1:0", storageFolder: "a", secondsBetweenCaptures: 600}, noCamera{}, func(a string) { addr <- a })
	taken := <-addr

	t.Chdir(t.TempDir())
	err := run(ctx, options{listen: taken, storageFolder: "b", secondsBetweenCaptures: 600}, noCamera{}, nil)
	if err == nil {
		t.Fatal("expected an error so systemd restarts the service")
	}
}
