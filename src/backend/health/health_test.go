package health

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestParseThrottled(t *testing.T) {
	// Under-voltage now, plus under-voltage and throttling since boot.
	th, err := ParseThrottled("throttled=0x50005\n")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := th.Now(), []string{"under-voltage", "throttled"}; !slices.Equal(got, want) {
		t.Errorf("Now = %v, want %v", got, want)
	}
	if got, want := th.SinceBoot(), []string{"under-voltage", "throttled"}; !slices.Equal(got, want) {
		t.Errorf("SinceBoot = %v, want %v", got, want)
	}
	if th.String() != "0x50005" {
		t.Errorf("String = %s", th)
	}
	ok, _ := ParseThrottled("throttled=0x0")
	if ok.Now() == nil || len(ok.Now()) != 0 || len(ok.SinceBoot()) != 0 {
		t.Errorf("0x0: Now %v SinceBoot %v, want empty, not nil", ok.Now(), ok.SinceBoot())
	}
	for _, bad := range []string{"", "temp=40.0'C", "throttled=zz"} {
		if _, err := ParseThrottled(bad); err == nil {
			t.Errorf("ParseThrottled(%q) succeeded", bad)
		}
	}
}

func TestReadThrottled(t *testing.T) {
	run := func(_ context.Context, name string, args ...string) ([]byte, error) {
		if !strings.HasSuffix(name, "vcgencmd") || !slices.Equal(args, []string{"get_throttled"}) {
			return nil, errors.New("unexpected command")
		}
		return []byte("throttled=0x8\n"), nil
	}
	th, err := ReadThrottled(t.Context(), run)
	if err != nil || !slices.Equal(th.Now(), []string{"soft temperature limit"}) {
		t.Fatalf("got %v, %v", th.Now(), err)
	}
	failing := func(context.Context, string, ...string) ([]byte, error) { return nil, errors.New("no vcgencmd") }
	if _, err := ReadThrottled(t.Context(), failing); err == nil {
		t.Fatal("no error")
	}
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReadCPUFreq(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "scaling_governor", "powersave\n")
	write(t, dir, "scaling_cur_freq", "600000\n")
	write(t, dir, "scaling_min_freq", "600000\n")
	write(t, dir, "scaling_max_freq", "1000000\n")
	c, err := ReadCPUFreq(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := (CPUFreq{Governor: "powersave", CurrentMHz: 600, MinMHz: 600, MaxMHz: 1000}); c != want {
		t.Fatalf("got %+v, want %+v", c, want)
	}
	write(t, dir, "scaling_max_freq", "fast\n")
	if _, err := ReadCPUFreq(dir); err == nil {
		t.Fatal("no error for a garbled frequency")
	}
	if _, err := ReadCPUFreq(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("no error without cpufreq")
	}
}

func TestReadCelsius(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "temp", "48312\n")
	if c, err := ReadCelsius(filepath.Join(dir, "temp")); err != nil || c != 48.312 {
		t.Fatalf("got %v, %v", c, err)
	}
	write(t, dir, "temp", "hot")
	if _, err := ReadCelsius(filepath.Join(dir, "temp")); err == nil {
		t.Fatal("no error for garbage")
	}
	if _, err := ReadCelsius(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("no error for a missing file")
	}
}

func TestStartRun(t *testing.T) {
	dir := t.TempDir()
	marker, boot := filepath.Join(dir, "running"), filepath.Join(dir, "boot_id")
	write(t, dir, "boot_id", "boot-1\n")

	start := func(want PreviousRun) func() error {
		t.Helper()
		prev, done, err := StartRun(marker, boot)
		if err != nil {
			t.Fatal(err)
		}
		if prev != want {
			t.Fatalf("previous run: %v, want %v", prev, want)
		}
		return done
	}
	done := start(StoppedCleanly) // first start
	if err := done(); err != nil {
		t.Fatal(err)
	}
	start(StoppedCleanly) // stopped cleanly, then killed:
	start(AppDied)        // same boot
	write(t, dir, "boot_id", "boot-2\n")
	start(PiRestarted)
	// Without boot IDs every unclean end is AppDied.
	if prev, _, _ := StartRun(marker, filepath.Join(dir, "none")); prev != AppDied {
		t.Fatalf("without boot id: %v", prev)
	}
	for _, p := range []PreviousRun{StoppedCleanly, AppDied, PiRestarted} {
		if p.String() == "" {
			t.Errorf("%d has no description", p)
		}
	}
	if _, _, err := StartRun(filepath.Join(dir, "no", "such", "dir"), boot); err == nil {
		t.Fatal("no error for an unwritable marker")
	}
	if _, _, err := StartRun(dir, boot); err == nil {
		t.Fatal("no error for an unreadable marker")
	}
}

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func TestRecorderSummary(t *testing.T) {
	var r Recorder
	r.Captured(nil)
	r.Captured(nil)
	r.Captured(errors.New("camera not found"))
	r.Deleted(3)
	r.Deleted(2)
	r.Sample(Sample{Celsius: new(51.5)})
	r.Sample(Sample{Celsius: new(43.0)})
	r.Sample(Sample{}) // no readings
	th := Throttled(0x10000)
	r.Sample(Sample{Throttled: &th})

	s := r.Summary()
	if s.Captured != 2 || s.Failed != 1 || s.LastError != "camera not found" || s.DeletedOldest != 5 {
		t.Fatalf("summary %+v", s)
	}
	if *s.MinCelsius != 43 || *s.MaxCelsius != 51.5 {
		t.Fatalf("temperatures %v..%v", *s.MinCelsius, *s.MaxCelsius)
	}
	next := r.Summary()
	if next.Captured != 0 || next.Failed != 0 || next.DeletedOldest != 0 || next.MinCelsius != nil || next.LastError != "" {
		t.Fatalf("not reset: %+v", next)
	}
	if !slices.Equal(next.ThrottledSinceBoot, []string{"under-voltage"}) {
		t.Fatalf("since-boot flags must carry over, got %v", next.ThrottledSinceBoot)
	}
}

func TestRecorderLogsPowerProblemsWhenTheyChange(t *testing.T) {
	logs := captureLogs(t)
	var r Recorder
	sample := func(v Throttled) { r.Sample(Sample{Throttled: &v}) }

	sample(0) // healthy: nothing to say
	if logs.Len() != 0 {
		t.Fatalf("logged %q for a healthy Pi", logs)
	}
	sample(0x50005) // under-voltage now
	sample(0x50005) // unchanged: logged once
	if n := strings.Count(logs.String(), "power or temperature problem"); n != 1 {
		t.Fatalf("logged %d problems, want 1:\n%s", n, logs)
	}
	sample(0x50000) // over
	if !strings.Contains(logs.String(), "back to normal") {
		t.Fatalf("recovery not logged:\n%s", logs)
	}

	logs.Reset()
	var fresh Recorder // e.g. after a restart: problems earlier in this boot
	v := Throttled(0x10000)
	fresh.Sample(Sample{Throttled: &v})
	if !strings.Contains(logs.String(), "problem since boot") {
		t.Fatalf("earlier problem not logged:\n%s", logs)
	}
}

func TestVcgencmdPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	if got := VcgencmdPath(); got != "/opt/vc/bin/vcgencmd" {
		t.Errorf("fallback: %s", got)
	}
	os.WriteFile(filepath.Join(dir, "vcgencmd"), []byte("#!/bin/sh\n"), 0o755)
	if got := VcgencmdPath(); got != filepath.Join(dir, "vcgencmd") {
		t.Errorf("found: %s", got)
	}
}
