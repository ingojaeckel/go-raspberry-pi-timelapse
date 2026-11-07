package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// watchdog lets systemd restart the camera when it still runs but no longer
// takes photos (WatchdogSec= in timelapse.service): it pings systemd only
// while the scheduler loop keeps waking up, which it does at least once a
// minute and after every capture.
//
// Times are measured with the monotonic clock: setting the wall clock (from
// the web UI, an RTC or NTP) must neither look like a stall nor hide one.
type watchdog struct {
	start    time.Time    // carries the monotonic reading the others are relative to
	last     atomic.Int64 // time.Duration since start of the last heartbeat
	maxStall time.Duration
	notify   func(state string) error
}

func newWatchdog(maxStall time.Duration) *watchdog {
	w := &watchdog{start: time.Now(), maxStall: maxStall, notify: sdNotify}
	w.beat()
	return w
}

// beat records that the scheduler is alive. It is the scheduler's Heartbeat.
func (w *watchdog) beat() { w.last.Store(int64(time.Since(w.start))) }

// lastBeat is when the scheduler was last alive.
func (w *watchdog) lastBeat() time.Time { return w.start.Add(time.Duration(w.last.Load())) }

// healthy reports whether the last heartbeat is at most maxStall before now,
// a time.Now() reading (with its monotonic clock).
func (w *watchdog) healthy(now time.Time) bool {
	return now.Sub(w.lastBeat()) <= w.maxStall
}

// run pings systemd every interval while healthy, until ctx ends.
func (w *watchdog) run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	reported := false
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if w.healthy(now) {
				w.ping()
				reported = false
			} else if !reported {
				slog.Error("scheduler stalled, letting systemd restart the camera", "lastHeartbeat", w.lastBeat())
				reported = true
			}
		}
	}
}

func (w *watchdog) ping() {
	if err := w.notify("WATCHDOG=1"); err != nil {
		slog.Warn("systemd watchdog", "err", err)
	}
}

// watchdogInterval is how often to ping systemd: half the WatchdogSec it
// configured. ok is false when systemd does not watch this process.
func watchdogInterval() (interval time.Duration, ok bool) {
	if pid := os.Getenv("WATCHDOG_PID"); pid != "" && pid != strconv.Itoa(os.Getpid()) {
		return 0, false
	}
	usec, err := strconv.ParseInt(os.Getenv("WATCHDOG_USEC"), 10, 64)
	if err != nil || usec <= 0 {
		return 0, false
	}
	return time.Duration(usec) * time.Microsecond / 2, true
}

var errNoNotifySocket = errors.New("not started by systemd (no NOTIFY_SOCKET)")

// sdNotify sends a state such as "READY=1" to systemd's notification
// socket (see sd_notify(3)).
func sdNotify(state string) error {
	path := os.Getenv("NOTIFY_SOCKET")
	if path == "" {
		return errNoNotifySocket
	}
	if strings.HasPrefix(path, "@") {
		path = "\x00" + path[1:] // abstract socket
	}
	conn, err := net.DialUnix("unixgram", nil, &net.UnixAddr{Name: path, Net: "unixgram"})
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(state))
	return err
}
