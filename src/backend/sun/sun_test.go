package sun

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// cases is src/frontend/src/sun-cases.json, shared with sun.test.ts and
// generated with an independent library (astral).
type cases struct {
	ToleranceSeconds int64
	Days             []struct {
		Place           string
		Lat, Lon        float64
		Date            string
		Sunrise, Sunset int64
	}
	Polar []struct {
		Place    string
		Lat, Lon float64
		Unix     int64
		Up       bool
	}
}

func load(t *testing.T) cases {
	t.Helper()
	raw, err := os.ReadFile("../../frontend/src/sun-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var c cases
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatal(err)
	}
	if len(c.Days) == 0 || len(c.Polar) == 0 {
		t.Fatal("no cases")
	}
	return c
}

func TestUpMatchesSunriseAndSunset(t *testing.T) {
	c := load(t)
	tol := time.Duration(c.ToleranceSeconds) * time.Second
	for _, d := range c.Days {
		rise, set := time.Unix(d.Sunrise, 0), time.Unix(d.Sunset, 0)
		for _, check := range []struct {
			at   time.Time
			want bool
		}{
			{rise.Add(-tol), false},
			{rise.Add(tol), true},
			{set.Add(-tol), true},
			{set.Add(tol), false},
		} {
			if got := Up(check.at, d.Lat, d.Lon); got != check.want {
				t.Errorf("%s %s: Up(%s) = %v", d.Place, d.Date, check.at.UTC().Format(time.TimeOnly), got)
			}
		}
	}
}

func TestNextSunriseAndSunset(t *testing.T) {
	c := load(t)
	tol := time.Duration(c.ToleranceSeconds) * time.Second
	near := func(got, want time.Time) bool { return got.Sub(want).Abs() <= tol }
	for _, d := range c.Days {
		rise, set := time.Unix(d.Sunrise, 0), time.Unix(d.Sunset, 0)
		if got, ok := NextSunrise(rise.Add(-3*time.Hour), d.Lat, d.Lon); !ok || !near(got, rise) {
			t.Errorf("%s %s: NextSunrise = %s, want %s", d.Place, d.Date, got.UTC(), rise.UTC())
		}
		// In Reykjavik at midsummer the day's sunset is just after midnight,
		// before its sunrise: search from before each event.
		if got, ok := NextSunset(set.Add(-3*time.Hour), d.Lat, d.Lon); !ok || !near(got, set) {
			t.Errorf("%s %s: NextSunset = %s, want %s", d.Place, d.Date, got.UTC(), set.UTC())
		}
	}
}

func TestPolarDayAndNight(t *testing.T) {
	for _, p := range load(t).Polar {
		at := time.Unix(p.Unix, 0)
		if got := Up(at, p.Lat, p.Lon); got != p.Up {
			t.Errorf("%s %s: Up = %v, want %v", p.Place, at.UTC(), got, p.Up)
		}
	}
}

func TestNextSunriseAfterThePolarNight(t *testing.T) {
	// Tromsø: the sun stays below the horizon from late November to mid
	// January.
	from := time.Date(2026, 12, 21, 12, 0, 0, 0, time.UTC)
	got, ok := NextSunrise(from, 69.6492, 18.9553)
	if !ok || got.Before(time.Date(2027, 1, 10, 0, 0, 0, 0, time.UTC)) || got.After(time.Date(2027, 1, 20, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("got %s %v", got, ok)
	}
}

func TestNoSunriseAtThePole(t *testing.T) {
	// North pole: the sun sets at the September equinox and rises at the
	// March one, five months later, still within the search limit.
	from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	got, ok := NextSunrise(from, 90, 0)
	if !ok || got.Month() != time.March {
		t.Fatalf("got %s %v", got, ok)
	}
	if _, ok := NextSunset(time.Date(2026, 6, 21, 0, 0, 0, 0, time.UTC), 90, 0); !ok {
		t.Fatal("the polar day ends in September")
	}
}

func TestValidLocation(t *testing.T) {
	for _, tc := range []struct {
		lat, lon float64
		ok       bool
	}{
		{51.5, -0.13, true}, {-90, 180, true}, {90.1, 0, false}, {0, -180.5, false},
	} {
		if got := ValidLocation(tc.lat, tc.lon); got != tc.ok {
			t.Errorf("%v,%v: %v", tc.lat, tc.lon, got)
		}
	}
}
