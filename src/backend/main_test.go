package main

import (
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"slices"
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
		selfCheckMinBrightness: defaultSelfCheckMinBrightness,
	}
	if o != want {
		t.Fatalf("got %+v, want %+v", o, want)
	}
}

func TestParseFlags(t *testing.T) {
	o, err := parseFlags([]string{"-port=:9090", "-logToFile=true", "-storageFolder=/p", "-secondsBetweenCaptures=600", "-captureRetries=0", "-minFreeDiskMB=10", "-pprof", "-allowedHosts=camera.example.org, cam.example.net", "-selfcheck", "-selfcheckKeep", "-selfcheckMinBrightness=0"})
	if err != nil {
		t.Fatal(err)
	}
	want := options{listen: ":9090", logToFile: true, storageFolder: "/p", secondsBetweenCaptures: 600, captureRetries: 0, minFreeDiskMB: 10, pprof: true, allowedHosts: "camera.example.org, cam.example.net", selfCheck: true, selfCheckKeep: true}
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
		{"-selfcheckMinBrightness=-1"},
		{"-selfcheckMinBrightness=256"},
		{"-unknown"},
	} {
		if _, err := parseFlags(args); err == nil {
			t.Errorf("%v: expected an error", args)
		}
	}
}

func TestAllowedHostsList(t *testing.T) {
	o := options{allowedHosts: " camera.example.org, ,cam.example.net "}
	if got := o.allowedHostList(); !slices.Equal(got, []string{"camera.example.org", "cam.example.net"}) {
		t.Fatalf("got %q", got)
	}
	if got := (options{}).allowedHostList(); len(got) != 0 {
		t.Fatalf("got %q", got)
	}
}

func TestVersionString(t *testing.T) {
	if got := versionString("abc123", "Mon Nov 2"); got != "abc123 built at Mon Nov 2" {
		t.Fatalf("got %q", got)
	}
	fallback := versionString("", "")
	if fallback == "" {
		t.Fatal("expected a fallback version")
	}
	// Both values are needed; one alone falls back too.
	for _, partial := range [][2]string{{"abc", ""}, {"", "Mon Nov 2"}} {
		if got := versionString(partial[0], partial[1]); got != fallback {
			t.Errorf("versionString(%q, %q) = %q, want the fallback %q", partial[0], partial[1], got, fallback)
		}
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
	// The scheduler's plan reaches the API (the page shows it as "Next shot").
	var sched struct{ NextMillis int64 }
	for deadline := time.Now().Add(15 * time.Second); sched.NextMillis == 0 && time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		_, body := get("/schedule")
		json.Unmarshal([]byte(body), &sched)
	}
	if next := time.UnixMilli(sched.NextMillis); time.Until(next) > 10*time.Minute || time.Until(next) < -time.Minute {
		t.Errorf("/schedule next capture %v, want within the 10 minute interval", next)
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

// photoBackend writes a JPEG of the requested size in one shade of grey with
// a gradient, so it has contrast.
type photoBackend struct{ shade uint8 }

func (b photoBackend) Capture(_ context.Context, path string, o timelapse.Options) error {
	img := image.NewGray(image.Rect(0, 0, o.Width/8, o.Height/8))
	for i := range img.Pix {
		img.Pix[i] = b.shade + uint8(i%3)*b.shade/4
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, nil)
}

func TestSelfCheck(t *testing.T) {
	t.Chdir(t.TempDir())
	opts, err := parseFlags([]string{"-selfcheck"})
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if code := selfCheck(t.Context(), opts, photoBackend{shade: 100}, &out); code != 0 {
		t.Fatalf("exit code %d:\n%s", code, out.String())
	}
	for _, want := range []string{"PASS  capture", "photo            273x205 JPEG, brightness", "nothing saved yet"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("lacks %q:\n%s", want, out.String())
		}
	}
	if _, err := os.Stat(conf.SettingsFile); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the self-check created %s", conf.SettingsFile)
	}

	out.Reset()
	if code := selfCheck(t.Context(), opts, photoBackend{shade: 1}, &out); code != 1 || !strings.Contains(out.String(), "too dark") {
		t.Fatalf("black photo: exit code %d:\n%s", code, out.String())
	}
}
