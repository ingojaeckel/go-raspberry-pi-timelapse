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
	// BeforeCapture, if set, runs before every attempt (e.g. to free disk space).
	BeforeCapture func()
	// Location defines the hours captures are aligned to; nil means time.Local.
	Location *time.Location
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
	started := time.Now()
	// anchor is the last capture slot, or start-up: unaligned captures are
	// anchor + interval. lastSlot prevents capturing an aligned slot twice.
	anchor, lastSlot := started, started.Add(-time.Second)

	for {
		set := s.Settings.Get()
		next := s.next(set, time.Now(), anchor, lastSlot)
		if wait := time.Until(next); wait > 0 {
			slog.Debug("waiting for next capture", "at", next, "in", wait.Round(time.Second))
			timer := time.NewTimer(min(wait, maxSleep))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-changed:
				timer.Stop()
				slog.Info("schedule changed", "secondsBetweenCaptures", s.Settings.Get().SecondsBetweenCaptures, "offsetWithinHour", s.Settings.Get().OffsetWithinHour)
			case <-timer.C:
			}
			continue // re-evaluate with the latest settings and clock
		}
		anchor, lastSlot = next, next
		s.captureWithRetries(ctx)
		if ctx.Err() != nil {
			return
		}
	}
}

func (s *Scheduler) next(set conf.Settings, now, anchor, lastSlot time.Time) time.Time {
	interval := time.Duration(set.SecondsBetweenCaptures) * time.Second
	if set.OffsetWithinHour == conf.OffsetDisabled {
		next := anchor.Add(interval)
		if limit := now.Add(interval); next.After(limit) {
			next = limit // the clock moved backwards: wait at most one interval
		}
		return next
	}
	from := now.Add(-lateTolerance)
	if !from.After(lastSlot) && lastSlot.Sub(now) < interval {
		from = lastSlot.Add(time.Second)
	}
	return NextAligned(from.In(s.location()), set.SecondsBetweenCaptures, set.OffsetWithinHour)
}

func (s *Scheduler) location() *time.Location {
	if s.Location != nil {
		return s.Location
	}
	return time.Local
}

func (s *Scheduler) captureWithRetries(ctx context.Context) {
	for attempt := range s.Retries + 1 {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(s.RetryDelay):
			}
		}
		if s.BeforeCapture != nil {
			s.BeforeCapture()
		}
		set := s.Settings.Get()
		started := time.Now()
		path, err := s.Camera.Capture(ctx, s.Folder, Options{
			Width:   set.PhotoResolutionWidth,
			Height:  set.PhotoResolutionHeight,
			Quality: set.Quality,
			Rotate:  set.RotateBy == 180,
		})
		if err == nil {
			slog.Info("photo stored", "path", path, "took", time.Since(started).Round(time.Millisecond))
			return
		}
		slog.Error("capture failed", "attempt", attempt+1, "of", s.Retries+1, "err", err)
		if ctx.Err() != nil {
			return
		}
	}
}
