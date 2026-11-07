// Package timelapse takes photos on a schedule.
package timelapse

import (
	"context"
	"log/slog"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

const (
	// lateTolerance is how late a capture may start and still count for its
	// slot, e.g. when a timer fires slightly late on a busy Pi.
	lateTolerance = 5 * time.Second
	// maxSleep bounds each wait so that wall clock adjustments (NTP, RTC at
	// boot) are noticed within a minute.
	maxSleep = time.Minute
)

// SettingsSource provides the current settings and change notifications.
// *conf.Store implements it.
type SettingsSource interface {
	Get() conf.Settings
	Subscribe() <-chan struct{}
}

// Capturer takes a photo into a folder. *Camera implements it.
type Capturer interface {
	Capture(ctx context.Context, dir string, o Options) (string, error)
}

// Scheduler captures photos into Folder according to the current settings.
type Scheduler struct {
	Settings SettingsSource
	Camera   Capturer
	Folder   string
	// Retries is the number of additional attempts within the same slot
	// after a failed capture, RetryDelay apart.
	Retries    int
	RetryDelay time.Duration
	// BeforeCapture, if set, runs before every attempt (e.g. to free disk
	// space). An error skips the slot without retrying (e.g. the disk is full
	// and photos must not be deleted).
	BeforeCapture func() error
	// Captured, if set, runs after every slot with nil or the last error.
	Captured func(err error)
	// Location defines the hours captures are aligned to; nil means time.Local.
	Location *time.Location
	// Heartbeat, if set, runs every time the loop wakes up: at least every
	// maxSleep while waiting, and after every capture (with its retries).
	Heartbeat func()
}

// Run captures photos until ctx is cancelled. Settings changes take effect
// immediately, also while waiting for the next capture.
//
// With OffsetWithinHour >= 0 photos are taken at fixed times every hour (see
// NextAligned). With conf.OffsetDisabled they are taken every
// SecondsBetweenCaptures counted from start-up; the time a capture takes does
// not shift later captures.
func (s *Scheduler) Run(ctx context.Context) {
	changed := s.Settings.Subscribe()
	// anchor is the last capture slot, or start-up: unaligned captures are
	// anchor + interval. lastSlot prevents capturing an aligned slot twice.
	anchor := time.Now()
	lastSlot := anchor.Add(-time.Second)

	for ctx.Err() == nil {
		if s.Heartbeat != nil {
			s.Heartbeat()
		}
		next := s.next(s.Settings.Get(), time.Now(), anchor, lastSlot)
		if wait := time.Until(next); wait > 0 {
			s.sleep(ctx, changed, next, wait) // then re-evaluate settings and clock
			continue
		}
		anchor, lastSlot = next, next
		err := s.captureWithRetries(ctx)
		if s.Captured != nil && ctx.Err() == nil {
			s.Captured(err)
		}
	}
}

// sleep waits until next, but at most maxSleep and only until the settings
// change or ctx is cancelled.
func (s *Scheduler) sleep(ctx context.Context, changed <-chan struct{}, next time.Time, wait time.Duration) {
	slog.Debug("waiting for next capture", "at", next, "in", wait.Round(time.Second))
	timer := time.NewTimer(min(wait, maxSleep))
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-changed:
		set := s.Settings.Get()
		slog.Info("schedule changed", "secondsBetweenCaptures", set.SecondsBetweenCaptures, "offsetWithinHour", set.OffsetWithinHour)
	case <-timer.C:
	}
}

func (s *Scheduler) next(set conf.Settings, now, anchor, lastSlot time.Time) time.Time {
	interval := time.Duration(set.SecondsBetweenCaptures) * time.Second
	if set.OffsetWithinHour == conf.OffsetDisabled {
		return nextUnaligned(now, anchor, interval)
	}
	from := alignedSearchStart(now, lastSlot, interval)
	return NextAligned(from.In(s.location()), set.SecondsBetweenCaptures, set.OffsetWithinHour)
}

// nextUnaligned is one interval after anchor. If the clock moved backwards,
// it waits at most one interval from now. If that time has long passed (the
// clock was set forward, or a capture with retries took longer than the
// interval), it is now: one capture, not one for every missed interval.
func nextUnaligned(now, anchor time.Time, interval time.Duration) time.Time {
	next := anchor.Add(interval)
	if limit := now.Add(interval); next.After(limit) {
		return limit
	}
	if now.Sub(next) > lateTolerance {
		return now
	}
	return next
}

// alignedSearchStart is where to look for the next aligned slot: slightly in
// the past to tolerate late wake-ups, but after the slot captured last
// (unless the clock moved backwards by more than an interval).
func alignedSearchStart(now, lastSlot time.Time, interval time.Duration) time.Time {
	from := now.Add(-lateTolerance)
	if !from.After(lastSlot) && lastSlot.Sub(now) < interval {
		return lastSlot.Add(time.Second)
	}
	return from
}

func (s *Scheduler) location() *time.Location {
	if s.Location != nil {
		return s.Location
	}
	return time.Local
}

// captureWithRetries tries up to 1+Retries times, RetryDelay apart, and
// returns the last error.
func (s *Scheduler) captureWithRetries(ctx context.Context) (err error) {
	for attempt := range s.Retries + 1 {
		if attempt > 0 && !sleepCtx(ctx, s.RetryDelay) {
			return ctx.Err()
		}
		if s.BeforeCapture != nil {
			if err := s.BeforeCapture(); err != nil {
				slog.Error("skipping capture", "err", err)
				return err
			}
		}
		if err = s.captureOnce(ctx); err == nil {
			return nil
		}
		slog.Error("capture failed", "attempt", attempt+1, "of", s.Retries+1, "err", err)
		if ctx.Err() != nil {
			return err
		}
	}
	return err
}

func (s *Scheduler) captureOnce(ctx context.Context) error {
	started := time.Now()
	path, err := s.Camera.Capture(ctx, s.Folder, PhotoOptions(s.Settings.Get()))
	if err == nil {
		slog.Info("photo stored", "path", path, "took", time.Since(started).Round(time.Millisecond))
	}
	return err
}

// sleepCtx sleeps for d and returns false if ctx is cancelled first.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
