package main

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

// notifySocket listens like systemd does and returns what it receives.
func notifySocket(t *testing.T, name string) <-chan string {
	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: name, Net: "unixgram"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	got := make(chan string, 10)
	go func() {
		buf := make([]byte, 256)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			got <- string(buf[:n])
		}
	}()
	return got
}

func receive(t *testing.T, got <-chan string, want string) {
	t.Helper()
	select {
	case msg := <-got:
		if msg != want {
			t.Fatalf("systemd received %q, want %q", msg, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("systemd received nothing, want %q", want)
	}
}

func TestSdNotify(t *testing.T) {
	t.Setenv("NOTIFY_SOCKET", "")
	if err := sdNotify("READY=1"); !errors.Is(err, errNoNotifySocket) {
		t.Fatalf("without systemd: %v", err)
	}
	// Socket paths are limited to about 100 bytes, shorter than t.TempDir().
	dir, err := os.MkdirTemp("", "sd")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	path := filepath.Join(dir, "notify")
	got := notifySocket(t, path)
	t.Setenv("NOTIFY_SOCKET", path)
	if err := sdNotify("WATCHDOG=1"); err != nil {
		t.Fatal(err)
	}
	receive(t, got, "WATCHDOG=1")

	t.Setenv("NOTIFY_SOCKET", filepath.Join(dir, "missing"))
	if err := sdNotify("READY=1"); err == nil {
		t.Fatal("expected an error for a missing socket")
	}
}

func TestSdNotifyAbstractSocket(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("abstract unix sockets are Linux only")
	}
	name := "timelapse-test-" + strconv.Itoa(os.Getpid())
	got := notifySocket(t, "\x00"+name)
	t.Setenv("NOTIFY_SOCKET", "@"+name)
	if err := sdNotify("READY=1"); err != nil {
		t.Fatal(err)
	}
	receive(t, got, "READY=1")
}

func TestWatchdogInterval(t *testing.T) {
	for _, c := range []struct {
		usec, pid string
		want      time.Duration
		ok        bool
	}{
		{"", "", 0, false},
		{"garbage", "", 0, false},
		{"0", "", 0, false},
		{"120000000", "", time.Minute, true},
		{"120000000", strconv.Itoa(os.Getpid()), time.Minute, true},
		{"120000000", "1", 0, false}, // meant for another process
	} {
		t.Setenv("WATCHDOG_USEC", c.usec)
		t.Setenv("WATCHDOG_PID", c.pid)
		if got, ok := watchdogInterval(); got != c.want || ok != c.ok {
			t.Errorf("WATCHDOG_USEC=%q WATCHDOG_PID=%q: %v %v, want %v %v", c.usec, c.pid, got, ok, c.want, c.ok)
		}
	}
}

// Started by systemd with WatchdogSec, the app wires the watchdog to the
// scheduler's heartbeat and pings systemd.
func TestStartWatchdogUnderSystemd(t *testing.T) {
	quietLogs(t)
	dir, err := os.MkdirTemp("", "sd") // socket paths must be short
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	got := notifySocket(t, filepath.Join(dir, "notify"))
	t.Setenv("NOTIFY_SOCKET", filepath.Join(dir, "notify"))
	t.Setenv("WATCHDOG_USEC", "200000") // ping every 100 ms
	t.Setenv("WATCHDOG_PID", "")

	sched := &timelapse.Scheduler{}
	ctx, cancel := context.WithCancel(t.Context())
	var wg sync.WaitGroup
	(&app{opts: options{captureRetries: 2}}).startWatchdog(ctx, &wg, sched)
	defer func() { cancel(); wg.Wait() }()
	if sched.Heartbeat == nil {
		t.Fatal("the scheduler's heartbeat is not wired to the watchdog")
	}
	sched.Heartbeat()
	receive(t, got, "WATCHDOG=1")
}

// The watchdog pings systemd while the scheduler keeps beating, stops when
// it stalls (so systemd restarts the camera), and resumes when it recovers.
func TestWatchdogPingsOnlyWhileTheSchedulerIsAlive(t *testing.T) {
	quietLogs(t)
	synctest.Test(t, func(t *testing.T) {
		var mu sync.Mutex
		pings := 0
		w := newWatchdog(5 * time.Minute)
		w.notify = func(state string) error {
			mu.Lock()
			defer mu.Unlock()
			if state == "WATCHDOG=1" {
				pings++
			}
			return errors.New("ignored")
		}
		count := func() int { mu.Lock(); defer mu.Unlock(); n := pings; pings = 0; return n }
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan struct{})
		go func() { defer close(done); w.run(ctx, time.Minute) }()

		for range 10 { // healthy: beats every minute
			time.Sleep(time.Minute)
			w.beat()
		}
		synctest.Wait()
		if n := count(); n != 10 {
			t.Fatalf("%d pings while healthy, want 10", n)
		}
		time.Sleep(30 * time.Minute) // stalled
		synctest.Wait()
		if n := count(); n != 5 {
			t.Fatalf("%d pings in 30 minutes without heartbeat, want 5 (until maxStall)", n)
		}
		w.beat()
		time.Sleep(time.Minute + time.Second)
		synctest.Wait()
		if n := count(); n != 1 {
			t.Fatalf("%d pings after recovering, want 1", n)
		}
		cancel()
		<-done
	})
}

func TestMaxStallCoversAllAttemptsOfACapture(t *testing.T) {
	a := &app{opts: options{captureRetries: 2}}
	if got := a.maxStall(); got < 3*captureTimeout || got > 15*time.Minute {
		t.Fatalf("maxStall = %v", got)
	}
}

// Setting the wall clock must not look like a stall (or hide one): the age of
// the last heartbeat is measured with the monotonic clock, which only
// time.Now() readings carry ("m=" in their String form).
func TestWatchdogIgnoresWallClockChanges(t *testing.T) {
	w := newWatchdog(time.Minute)
	if !strings.Contains(w.lastBeat().String(), "m=") {
		t.Fatalf("last heartbeat %v has no monotonic clock reading", w.lastBeat())
	}
	if !w.healthy(time.Now()) {
		t.Fatal("unhealthy right after a heartbeat")
	}
}
