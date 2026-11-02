package timelapse

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

// scheduleCases is shared with the frontend (src/frontend/src/schedule.test.ts).
type scheduleCases struct {
	ValidIntervals   []int `json:"validIntervals"`
	InvalidIntervals []int `json:"invalidIntervals"`
	Cases            []struct {
		Interval, Offset, Minute, Second, Wait int
	} `json:"cases"`
}

func loadScheduleCases(t *testing.T) scheduleCases {
	t.Helper()
	raw, err := os.ReadFile("../../frontend/src/schedule-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var c scheduleCases
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNextAlignedMatchesSharedCases(t *testing.T) {
	for _, c := range loadScheduleCases(t).Cases {
		now := time.Date(2017, 1, 1, 1, c.Minute, c.Second, 0, time.UTC)
		got := NextAligned(now, c.Interval, c.Offset)
		if wait := int(got.Sub(now) / time.Second); wait != c.Wait {
			t.Errorf("interval=%d offset=%d at :%02d:%02d waits %ds, want %ds", c.Interval, c.Offset, c.Minute, c.Second, wait, c.Wait)
		}
	}
}

func TestValidIntervalsMatchSharedCases(t *testing.T) {
	c := loadScheduleCases(t)
	for _, v := range c.ValidIntervals {
		if !conf.ValidInterval(v) {
			t.Errorf("%d should be valid", v)
		}
	}
	for _, v := range c.InvalidIntervals {
		if conf.ValidInterval(v) {
			t.Errorf("%d should be invalid", v)
		}
	}
	if len(conf.ValidIntervals()) != len(c.ValidIntervals) {
		t.Errorf("conf.ValidIntervals() = %v, shared cases list %v", conf.ValidIntervals(), c.ValidIntervals)
	}
}

func TestNextAlignedUsesLocalHours(t *testing.T) {
	// India is UTC+5:30: "15 minutes past the hour" is local time, which is
	// 45 minutes past the UTC hour.
	ist := time.FixedZone("IST", 5*3600+1800)
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, ist)
	got := NextAligned(now, 3600, 900)
	if want := time.Date(2026, 1, 1, 10, 15, 0, 0, ist); !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestNextAlignedCountsTheCurrentSecond(t *testing.T) {
	// A slot is due for the whole second it starts in.
	now := time.Date(2026, 1, 1, 10, 15, 0, 500_000_000, time.UTC)
	if got, want := NextAligned(now, 1800, 900), time.Date(2026, 1, 1, 10, 15, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	now = time.Date(2026, 1, 1, 10, 15, 1, 0, time.UTC)
	if got, want := NextAligned(now, 1800, 900), time.Date(2026, 1, 1, 10, 45, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func BenchmarkNextAligned(b *testing.B) {
	now := time.Date(2017, time.December, 1, 8, 46, 1, 0, time.UTC)
	for b.Loop() {
		NextAligned(now, 1800, 900)
	}
}
