package conf

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"
)

func TestValidIntervalsDivideAnHourEvenly(t *testing.T) {
	got := ValidIntervals()
	want := []int{60, 72, 75, 80, 90, 100, 120, 144, 150, 180, 200, 225, 240, 300, 360, 400, 450, 600, 720, 900, 1200, 1800, 3600}
	if !slices.Equal(got, want) {
		t.Fatalf("ValidIntervals() = %v, want %v", got, want)
	}
}

func valid() Settings {
	s := Default(1800)
	return s
}

func TestDefaultIsValid(t *testing.T) {
	if err := valid().Validate(); err != nil {
		t.Fatalf("default settings must be valid: %v", err)
	}
}

func TestValidateAcceptsDisabledOffset(t *testing.T) {
	s := valid()
	s.OffsetWithinHour = OffsetDisabled
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRejects(t *testing.T) {
	cases := map[string]func(*Settings){
		"interval below minute":      func(s *Settings) { s.SecondsBetweenCaptures = 1 },
		"interval zero":              func(s *Settings) { s.SecondsBetweenCaptures = 0 },
		"interval not dividing hour": func(s *Settings) { s.SecondsBetweenCaptures = 1000 },
		"interval above hour":        func(s *Settings) { s.SecondsBetweenCaptures = 7200 },
		"offset equal to interval":   func(s *Settings) { s.SecondsBetweenCaptures = 600; s.OffsetWithinHour = 600 },
		"offset beyond interval":     func(s *Settings) { s.SecondsBetweenCaptures = 600; s.OffsetWithinHour = 900 },
		"offset below -1":            func(s *Settings) { s.OffsetWithinHour = -2 },
		"quality zero":               func(s *Settings) { s.Quality = 0 },
		"quality above 100":          func(s *Settings) { s.Quality = 101 },
		"rotation not 0 or 180":      func(s *Settings) { s.RotateBy = 90 },
		"photo too narrow":           func(s *Settings) { s.PhotoResolutionWidth = 63 },
		"photo too tall":             func(s *Settings) { s.PhotoResolutionHeight = 16385 },
		"no photo size":              func(s *Settings) { s.PhotoResolutionWidth, s.PhotoResolutionHeight = 0, 0 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			s := valid()
			mutate(&s)
			err := s.Validate()
			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("Validate() = %v, want ErrInvalid", err)
			}
		})
	}
}

func TestNormalizedRepairsSettingsFromOlderVersions(t *testing.T) {
	cases := []struct {
		name                 string
		interval, offset     int
		wantInterval, wantOf int
	}{
		{"too short is raised to the minimum", 10, -1, 60, -1},
		{"non-divisor rounds up to the next valid interval", 1000, 100, 1200, 100},
		{"longer than an hour was hourly before, stays hourly", 7200, 600, 3600, 600},
		{"offset beyond interval keeps the same slots", 1800, 2700, 1800, 900},
		{"offset beyond the hour wraps", 3600, 3700, 3600, 100},
		{"valid settings are untouched", 900, 300, 900, 300},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := valid()
			s.SecondsBetweenCaptures, s.OffsetWithinHour = c.interval, c.offset
			n := s.Normalized()
			if n.SecondsBetweenCaptures != c.wantInterval || n.OffsetWithinHour != c.wantOf {
				t.Fatalf("got interval=%d offset=%d, want %d %d", n.SecondsBetweenCaptures, n.OffsetWithinHour, c.wantInterval, c.wantOf)
			}
			if err := n.Validate(); err != nil {
				t.Fatalf("normalized settings must be valid: %v", err)
			}
		})
	}
}

func TestNormalizedRepairsPhotoFields(t *testing.T) {
	s := valid()
	s.Quality, s.RotateBy, s.PhotoResolutionWidth = 0, 90, 0
	n := s.Normalized()
	if n.Quality != 1 || n.RotateBy != 0 {
		t.Fatalf("got quality=%d rotate=%d", n.Quality, n.RotateBy)
	}
	if n.PhotoResolutionWidth != DefaultPhotoWidth || n.PhotoResolutionHeight != DefaultPhotoHeight {
		t.Fatalf("a missing size must become the default, got %dx%d", n.PhotoResolutionWidth, n.PhotoResolutionHeight)
	}
}

// Versions before the camera's own resolutions were offered saved a
// ResolutionSetting index next to the size it selected; the size is kept.
func TestSettingsOfOlderVersionsKeepTheirPhotoSize(t *testing.T) {
	var s Settings
	old := `{"SecondsBetweenCaptures":1800,"OffsetWithinHour":900,"PhotoResolutionWidth":1640,"PhotoResolutionHeight":1232,"ResolutionSetting":2,"Quality":90}`
	if err := json.Unmarshal([]byte(old), &s); err != nil {
		t.Fatal(err)
	}
	if n := s.Normalized(); n.PhotoResolutionWidth != 1640 || n.PhotoResolutionHeight != 1232 {
		t.Fatalf("got %dx%d", n.PhotoResolutionWidth, n.PhotoResolutionHeight)
	}
}

func TestDefaultUsesGivenInterval(t *testing.T) {
	if got := Default(600).SecondsBetweenCaptures; got != 600 {
		t.Fatalf("Default(600) interval = %d", got)
	}
}

func TestSettingsStringIsJSON(t *testing.T) {
	want := `{"SecondsBetweenCaptures":0,"OffsetWithinHour":0,"PhotoResolutionWidth":0,"PhotoResolutionHeight":0,"PreviewResolutionWidth":0,"PreviewResolutionHeight":0,"RotateBy":0,"Quality":0,"DebugEnabled":false,"StopWhenDiskFull":false}`
	if got := (Settings{}).String(); got != want {
		t.Fatalf("got %s", got)
	}
}
