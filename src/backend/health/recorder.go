package health

import (
	"context"
	"log/slog"
	"os/exec"
	"slices"
	"sync"
	"time"
)

// Run executes a command and returns its output, like rest.RunCommand.
type Run func(ctx context.Context, name string, args ...string) ([]byte, error)

// VcgencmdPath finds vcgencmd: /usr/bin on current Raspberry Pi OS,
// /opt/vc/bin on older releases.
func VcgencmdPath() string {
	if p, err := exec.LookPath("vcgencmd"); err == nil {
		return p
	}
	return "/opt/vc/bin/vcgencmd"
}

// ReadThrottled asks the firmware for under-voltage and throttling, giving
// vcgencmd at most 10 seconds.
func ReadThrottled(ctx context.Context, run Run) (Throttled, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := run(ctx, VcgencmdPath(), "get_throttled")
	if err != nil {
		return 0, err
	}
	return ParseThrottled(string(out))
}

// Sample is one reading of the Pi's health. Missing readings are nil.
type Sample struct {
	Celsius   *float64
	Throttled *Throttled
}

// Summary is what happened since the previous summary.
type Summary struct {
	// Captured and Failed count capture slots.
	Captured, Failed int
	LastError        string
	// DeletedOldest counts photos deleted to free disk space.
	DeletedOldest int
	// MinCelsius and MaxCelsius are nil without temperature readings.
	MinCelsius, MaxCelsius *float64
	// ThrottledSinceBoot lists the power and temperature problems the
	// firmware reported since boot.
	ThrottledSinceBoot []string
}

// Recorder collects captures, deletions and health samples for the daily
// summary, and logs power or temperature problems when they appear or go
// away. It is safe for concurrent use.
type Recorder struct {
	mu      sync.Mutex
	sum     Summary
	last    Throttled
	sampled bool // last holds a reading
}

// Captured records the outcome of a capture slot (the scheduler's Captured).
func (r *Recorder) Captured(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err != nil {
		r.sum.Failed++
		r.sum.LastError = err.Error()
	} else {
		r.sum.Captured++
	}
}

// Deleted records photos deleted to free disk space.
func (r *Recorder) Deleted(n int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sum.DeletedOldest += n
}

// Sample records a reading.
func (r *Recorder) Sample(s Sample) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c := s.Celsius; c != nil {
		if r.sum.MinCelsius == nil || *c < *r.sum.MinCelsius {
			r.sum.MinCelsius = new(*c)
		}
		if r.sum.MaxCelsius == nil || *c > *r.sum.MaxCelsius {
			r.sum.MaxCelsius = new(*c)
		}
	}
	if t := s.Throttled; t != nil {
		if *t != r.last || !r.sampled && *t != 0 {
			logThrottled(*t, r.last)
		}
		r.last, r.sampled = *t, true
		r.sum.ThrottledSinceBoot = t.SinceBoot()
	}
}

func logThrottled(t, before Throttled) {
	switch {
	case len(t.Now()) > 0:
		slog.Warn("power or temperature problem", "now", t.Now(), "sinceBoot", t.SinceBoot(), "throttled", t)
	case len(before.Now()) > 0:
		slog.Info("power and temperature back to normal", "sinceBoot", t.SinceBoot(), "throttled", t)
	default:
		slog.Warn("power or temperature problem since boot", "sinceBoot", t.SinceBoot(), "throttled", t)
	}
}

// Summary returns what happened since the previous call and starts over.
// The throttling flags carry over: they are since boot.
func (r *Recorder) Summary() Summary {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.sum
	s.ThrottledSinceBoot = slices.Clone(s.ThrottledSinceBoot)
	r.sum = Summary{ThrottledSinceBoot: r.sum.ThrottledSinceBoot}
	return s
}
