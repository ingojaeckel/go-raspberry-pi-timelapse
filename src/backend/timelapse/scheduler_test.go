package timelapse

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/sun"
)

// All scheduler tests run in a synctest bubble: time starts at
// 2000-01-01 00:00:00 UTC and only advances when every goroutine is blocked,
// so hours of scheduling run instantly and deterministically.

type fakeSettings struct {
	mu   sync.Mutex
	s    conf.Settings
	subs []chan struct{}
}

func (f *fakeSettings) Get() conf.Settings {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.s
}

func (f *fakeSettings) Subscribe() <-chan struct{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	ch := make(chan struct{}, 1)
	f.subs = append(f.subs, ch)
	return ch
}

func (f *fakeSettings) set(interval, offset int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.s.SecondsBetweenCaptures, f.s.OffsetWithinHour = interval, offset
	for _, ch := range f.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// fakeCamera records when captures start (as offsets from the bubble's start)
// and fails the first `failures` attempts.
type fakeCamera struct {
	mu       sync.Mutex
	start    time.Time
	took     time.Duration
	failures int
	attempts []time.Duration
	opts     []Options
}

func (c *fakeCamera) Capture(ctx context.Context, dir string, o Options) (string, error) {
	c.mu.Lock()
	c.attempts = append(c.attempts, time.Since(c.start))
	c.opts = append(c.opts, o)
	fail := len(c.attempts) <= c.failures
	c.mu.Unlock()
	time.Sleep(c.took)
	if fail {
		return "", errors.New("camera busy")
	}
	return dir + "/photo.jpg", nil
}

func (c *fakeCamera) times() []time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]time.Duration(nil), c.attempts...)
}

type harness struct {
	settings *fakeSettings
	camera   *fakeCamera
	sched    *Scheduler
	cancel   context.CancelFunc
	done     chan struct{}
}

func start(t *testing.T, interval, offset int, configure ...func(*harness)) *harness {
	t.Helper()
	h := &harness{
		settings: &fakeSettings{s: conf.Settings{SecondsBetweenCaptures: interval, OffsetWithinHour: offset, Quality: 90, PhotoResolutionWidth: 800, PhotoResolutionHeight: 600}},
		camera:   &fakeCamera{start: time.Now(), took: 5 * time.Second},
		done:     make(chan struct{}),
	}
	h.sched = &Scheduler{
		Settings:   h.settings,
		Camera:     h.camera,
		Folder:     "photos",
		Retries:    2,
		RetryDelay: time.Second,
		Location:   time.UTC,
	}
	for _, c := range configure {
		c(h)
	}
	ctx, cancel := context.WithCancel(t.Context())
	h.cancel = cancel
	go func() {
		defer close(h.done)
		h.sched.Run(ctx)
	}()
	t.Cleanup(h.stop)
	return h
}

func (h *harness) stop() {
	h.cancel()
	<-h.done
}

// advance lets the scheduler run until d has passed since the bubble started.
func advance(h *harness, d time.Duration) {
	time.Sleep(d - time.Since(h.camera.start))
	synctest.Wait()
}

func secs(s ...int) []time.Duration {
	out := make([]time.Duration, len(s))
	for i, v := range s {
		out[i] = time.Duration(v) * time.Second
	}
	return out
}

func assertTimes(t *testing.T, got, want []time.Duration) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("captures at %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("captures at %v, want %v", got, want)
		}
	}
}

func TestAlignedCapturesHappenOnEverySlot(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 600, 0)
		advance(h, time.Hour+time.Second)
		h.stop()
		// No slot is skipped and nothing drifts, even though each capture takes 5s.
		assertTimes(t, h.camera.times(), secs(0, 600, 1200, 1800, 2400, 3000, 3600))
	})
}

func TestAlignedCapturesUseTheOffset(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 1800, 900)
		advance(h, time.Hour+time.Second)
		h.stop()
		assertTimes(t, h.camera.times(), secs(900, 2700))
	})
}

func TestAlignedConfigChangeAppliesWhileWaiting(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 3600, 0)
		advance(h, 10*time.Second)
		h.settings.set(60, 30)
		advance(h, 100*time.Second)
		h.stop()
		assertTimes(t, h.camera.times(), secs(0, 30, 90))
	})
}

func TestUnalignedCapturesDoNotDrift(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 600, conf.OffsetDisabled)
		advance(h, 1800*time.Second+time.Second)
		h.stop()
		// Counted from start-up; the 5s capture time does not push later captures back.
		assertTimes(t, h.camera.times(), secs(600, 1200, 1800))
	})
}

// A new unaligned interval applies at once, not after the old sleep, and
// captures stay counted from start-up.
func TestUnalignedConfigChange(t *testing.T) {
	for _, c := range []struct {
		name            string
		from, to        int           // interval before and after the change
		changeAt, until time.Duration // since start-up
		want            []time.Duration
	}{
		{"shorter, while waiting", 3600, 60, 30 * time.Second, 200 * time.Second, secs(60, 120, 180)},
		{"shorter, after its first slot passed", 3600, 72, 70 * time.Second, 150 * time.Second, secs(72, 144)},
		{"longer", 60, 300, 70 * time.Second, 400 * time.Second, secs(60, 360)},
	} {
		t.Run(c.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				h := start(t, c.from, conf.OffsetDisabled)
				advance(h, c.changeAt)
				h.settings.set(c.to, conf.OffsetDisabled)
				advance(h, c.until)
				h.stop()
				assertTimes(t, h.camera.times(), c.want)
			})
		})
	}
}

func TestSwitchingFromUnalignedToAligned(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 3600, conf.OffsetDisabled)
		advance(h, 10*time.Second)
		h.settings.set(600, 120)
		advance(h, 800*time.Second)
		h.stop()
		assertTimes(t, h.camera.times(), secs(120, 720))
	})
}

func TestFailedCaptureIsRetriedWithinTheSameSlot(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 1800, 0, func(h *harness) { h.camera.failures = 2 })
		advance(h, 60*time.Second)
		h.stop()
		// 2 retries after the first failure, 1s apart (+5s per attempt).
		assertTimes(t, h.camera.times(), secs(0, 6, 12))
	})
}

func TestRetriesAreLimited(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 1800, 0, func(h *harness) { h.camera.failures = 100 })
		advance(h, 1830*time.Second)
		h.stop()
		// 3 attempts in the first slot, then 3 in the next one.
		assertTimes(t, h.camera.times(), secs(0, 6, 12, 1800, 1806, 1812))
	})
}

func TestNoRetriesWhenConfiguredToZero(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 1800, 0, func(h *harness) {
			h.sched.Retries = 0
			h.camera.failures = 100
		})
		advance(h, 60*time.Second)
		h.stop()
		assertTimes(t, h.camera.times(), secs(0))
	})
}

func TestSlowCaptureDoesNotRepeatTheSameSlot(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 60, 0, func(h *harness) { h.camera.took = 90 * time.Second }) // longer than the interval
		advance(h, 200*time.Second)
		h.stop()
		// Slot 60 passed while capturing at 0; it is skipped, not captured late twice.
		assertTimes(t, h.camera.times(), secs(0, 120))
	})
}

func TestCaptureUsesLatestPhotoSettings(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 600, 0)
		advance(h, 10*time.Second)
		h.settings.mu.Lock()
		h.settings.s.RotateBy = 180
		h.settings.mu.Unlock()
		advance(h, 601*time.Second)
		h.stop()
		want := Options{Width: 800, Height: 600, Quality: 90, Rotate: true}
		if len(h.camera.opts) != 2 || h.camera.opts[1] != want {
			t.Fatalf("opts = %+v, want %+v", h.camera.opts, want)
		}
	})
}

func TestBeforeCaptureRunsBeforeEveryAttempt(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls int
		var mu sync.Mutex
		h := start(t, 1800, 0, func(h *harness) {
			h.sched.BeforeCapture = func() error {
				mu.Lock()
				calls++
				mu.Unlock()
				return nil
			}
			h.camera.failures = 1
		})
		advance(h, 30*time.Second)
		h.stop()
		mu.Lock()
		defer mu.Unlock()
		if calls != 2 {
			t.Fatalf("BeforeCapture ran %d times, want 2", calls)
		}
	})
}

func TestRunStopsWhenCancelled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 3600, 0)
		advance(h, 10*time.Second)
		h.stop() // returns only once Run has exited
	})
}

func TestNextToleratesASlightlyLateWakeUp(t *testing.T) {
	s := &Scheduler{Location: time.UTC}
	set := conf.Settings{SecondsBetweenCaptures: 600, OffsetWithinHour: 0}
	slot := time.Date(2026, 1, 1, 10, 10, 0, 0, time.UTC)
	before := slot.Add(-600 * time.Second)

	if got := s.next(set, slot.Add(2*time.Second), before, before); !got.Equal(slot) {
		t.Fatalf("a timer firing 2s late must still capture the slot, got %v", got)
	}
	if got := s.next(set, slot.Add(2*time.Second), slot, slot); !got.Equal(slot.Add(600 * time.Second)) {
		t.Fatalf("a captured slot must not be captured again, got %v", got)
	}
	if got := s.next(set, slot.Add(6*time.Second), before, before); !got.Equal(slot.Add(600 * time.Second)) {
		t.Fatalf("a slot missed by 6s (over the 5s tolerance) is skipped, got %v", got)
	}
}

func TestNextFollowsTheClockBackwards(t *testing.T) {
	s := &Scheduler{Location: time.UTC}
	now := time.Date(2026, 1, 1, 10, 0, 30, 0, time.UTC)
	future := now.Add(24 * time.Hour) // last capture happened before the clock was set back a day

	aligned := conf.Settings{SecondsBetweenCaptures: 600, OffsetWithinHour: 0}
	if got := s.next(aligned, now, future, future); !got.Equal(now.Add(570 * time.Second)) {
		t.Fatalf("aligned: got %v", got)
	}
	unaligned := conf.Settings{SecondsBetweenCaptures: 600, OffsetWithinHour: conf.OffsetDisabled}
	if got := s.next(unaligned, now, future, future); !got.Equal(now.Add(600 * time.Second)) {
		t.Fatalf("unaligned: got %v", got)
	}
}

func TestHeartbeatAtLeastEveryMaxSleep(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var beats []time.Duration
		began := time.Now()
		h := start(t, 3600, 0, func(h *harness) {
			h.sched.Heartbeat = func() { beats = append(beats, time.Since(began)) }
		})
		advance(h, 10*time.Minute)
		h.stop()
		for i := 1; i < len(beats); i++ {
			if gap := beats[i] - beats[i-1]; gap > MaxSleep+10*time.Second {
				t.Fatalf("no heartbeat for %v", gap)
			}
		}
		if len(beats) < 10 {
			t.Fatalf("only %d heartbeats in 10 minutes", len(beats))
		}
	})
}

func TestNextUnalignedAfterTheClockJumpsForward(t *testing.T) {
	s := &Scheduler{Location: time.UTC}
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	before := now.Add(-72 * time.Hour) // last capture, before the clock was set forward 3 days
	unaligned := conf.Settings{SecondsBetweenCaptures: 600, OffsetWithinHour: conf.OffsetDisabled}
	// One capture now, not one for each of the 432 missed intervals.
	if got := s.next(unaligned, now, before, before); !got.Equal(now) {
		t.Fatalf("got %v, want now", got)
	}
	if got := s.next(unaligned, now.Add(5*time.Second), now, now); !got.Equal(now.Add(600 * time.Second)) {
		t.Fatalf("after catching up: got %v", got)
	}
}

func TestBeforeCaptureErrorSkipsTheSlot(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var results []error
		h := start(t, 600, 0, func(h *harness) {
			h.sched.BeforeCapture = func() error { return errors.New("disk full") }
			h.sched.Captured = func(err error) { results = append(results, err) }
		})
		advance(h, 1201*time.Second)
		h.stop()
		assertTimes(t, h.camera.times(), nil) // no capture, no retries
		if len(results) != 3 || results[0] == nil {
			t.Fatalf("Captured got %v, want 3 errors (slots 0, 600, 1200)", results)
		}
	})
}

func TestCapturedReportsEverySlotOnce(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var results []error
		h := start(t, 600, 0, func(h *harness) {
			h.camera.failures = 3 // the first slot fails all three attempts
			h.sched.Captured = func(err error) { results = append(results, err) }
		})
		advance(h, 610*time.Second) // a capture takes 5s
		h.stop()
		if len(results) != 2 || results[0] == nil || results[1] != nil {
			t.Fatalf("Captured got %v, want [error, nil]", results)
		}
	})
}

// london skips night photos in London. The bubble starts at 2000-01-01
// 00:00 UTC: sunrise is about 08:06, sunset about 16:02.
func london(h *harness) {
	h.settings.s.SkipNightPhotos, h.settings.s.Latitude, h.settings.s.Longitude = true, 51.5074, -0.1278
}

func TestNightPhotosAreSkippedWhenAligned(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 3600, 0, london)
		advance(h, 24*time.Hour)
		h.stop()
		assertTimes(t, h.camera.times(), secs(9*3600, 10*3600, 11*3600, 12*3600, 13*3600, 14*3600, 15*3600, 16*3600))
	})
}

func TestNightPhotosAreSkippedWhenUnaligned(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start0 := time.Now()
		h := start(t, 1800, conf.OffsetDisabled, london)
		advance(h, 12*time.Hour)
		h.stop()
		rise, ok := sun.NextSunrise(start0, 51.5074, -0.1278)
		if !ok {
			t.Fatal("no sunrise")
		}
		got := h.camera.times()
		if len(got) == 0 || got[0] != rise.Sub(start0) {
			t.Fatalf("first capture at %v, want sunrise at %v", got, rise.Sub(start0))
		}
		// then every 30 minutes from sunrise until noon
		for i := 1; i < len(got); i++ {
			if got[i]-got[i-1] != 30*time.Minute {
				t.Fatalf("captures at %v", got)
			}
		}
		if len(got) != 8 { // about 08:06 to 12:00
			t.Fatalf("%d captures: %v", len(got), got)
		}
	})
}

func TestSkippingNightsAppliesWhileWaiting(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 600, 0)
		advance(h, 30*time.Minute+time.Second) // 00:00, 00:10, 00:20, 00:30
		h.settings.mu.Lock()
		london(h)
		h.settings.mu.Unlock()
		h.settings.set(600, 0) // notify
		advance(h, 6*time.Hour)
		h.stop()
		assertTimes(t, h.camera.times(), secs(0, 600, 1200, 1800))
	})
}

func TestNightPhotosAreTakenWhenNotSkipped(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 3600, 0)
		advance(h, 3*time.Hour+time.Second)
		h.stop()
		assertTimes(t, h.camera.times(), secs(0, 3600, 7200, 10800))
	})
}

func TestNextIsUnknownBeforeTheSchedulerRuns(t *testing.T) {
	if _, ok := (&Scheduler{}).Next(); ok {
		t.Fatal("a plan before Run")
	}
}

func TestNextIsThePlannedCapture(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 3600, 0)
		advance(h, 10*time.Minute)
		p, ok := h.sched.Next()
		if want := h.camera.start.Add(time.Hour); !ok || !p.At.Equal(want) || p.Night {
			t.Fatalf("Next() = %+v %v, want %v by day", p, ok, want)
		}
	})
}

func TestNextIsUnalignedFromStartUp(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 600, conf.OffsetDisabled)
		advance(h, time.Minute)
		if p, ok := h.sched.Next(); !ok || !p.At.Equal(h.camera.start.Add(10*time.Minute)) {
			t.Fatalf("Next() = %+v %v", p, ok)
		}
	})
}

func TestNextIsAfterSunriseAtNight(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 3600, 0, london) // midnight in London on 2000-01-01; sunrise about 08:06
		advance(h, 10*time.Minute)
		p, ok := h.sched.Next()
		if want := h.camera.start.Add(9 * time.Hour); !ok || !p.At.Equal(want) || !p.Night {
			t.Fatalf("Next() = %+v %v, want %v after the night", p, ok, want)
		}
	})
}

func (s *Scheduler) next(set conf.Settings, now, anchor, lastSlot time.Time) time.Time {
	return s.plan(set, now, anchor, lastSlot).At
}

// syncBuffer is a log destination safe for the scheduler's goroutine.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) count(s string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strings.Count(b.buf.String(), s)
}

func captureLogs(t *testing.T) *syncBuffer {
	logs := &syncBuffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(logs, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return logs
}

func TestEachNightIsLoggedOnce(t *testing.T) {
	logs := captureLogs(t)
	synctest.Test(t, func(t *testing.T) {
		h := start(t, 600, 0, london)
		advance(h, 40*time.Hour) // the night of Jan 1 and of Jan 1 to 2, until 16:00 on Jan 2
		h.stop()
	})
	if got := logs.count("night: no photos until sunrise"); got != 2 {
		t.Fatalf("logged %d nights, want 2", got)
	}
}

// In the polar night the scheduler sleeps until the sun rises in March,
// waking only every MaxSleep, and takes no photos.
func TestPolarNightWaitsForSpringWithoutSpinning(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		beats := 0
		h := start(t, 600, 0, func(h *harness) {
			h.settings.s.SkipNightPhotos, h.settings.s.Latitude, h.settings.s.Longitude = true, 89.9, 0
			h.sched.Heartbeat = func() { beats++ }
		})
		advance(h, 2*time.Hour)
		plan, _ := h.sched.Next()
		h.stop()
		assertTimes(t, h.camera.times(), nil)
		if want := int(2 * time.Hour / MaxSleep); beats > want+2 {
			t.Fatalf("%d wake-ups in 2 hours, want about %d", beats, want)
		}
		if !plan.Night || plan.At.Month() != time.March {
			t.Fatalf("next capture %+v, want after sunrise in March", plan)
		}
	})
}
