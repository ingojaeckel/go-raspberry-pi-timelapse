package conf

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	MinSecondsBetweenCaptures = 60
	MaxSecondsBetweenCaptures = 3600
	// OffsetDisabled turns off alignment to the hour: captures then run every
	// SecondsBetweenCaptures counted from start-up.
	OffsetDisabled = -1
)

// ErrInvalid is wrapped by every validation error.
var ErrInvalid = errors.New("invalid settings")

// Settings is persisted as JSON and served to the frontend unchanged; the
// field names are part of the API (src/frontend/src/api.ts).
type Settings struct {
	SecondsBetweenCaptures int
	// OffsetWithinHour is the first capture in seconds past the hour,
	// 0 <= offset < SecondsBetweenCaptures, or OffsetDisabled.
	OffsetWithinHour        int
	PhotoResolutionWidth    int
	PhotoResolutionHeight   int
	PreviewResolutionWidth  int
	PreviewResolutionHeight int
	// RotateBy is 0 or 180.
	RotateBy int
	// ResolutionSetting selects PhotoResolutionWidth/Height, see resolutions.
	ResolutionSetting int
	Quality           int
	DebugEnabled      bool
}

func (s Settings) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

// resolutions maps ResolutionSetting to the photo size (Camera Module v2).
var resolutions = [...][2]int{
	0: {3280, 2464}, // 8 MP
	1: {2186, 1642}, // 3.5 MP
	2: {1640, 1232}, // 2 MP
}

// Default returns the settings used when nothing has been saved yet.
func Default(secondsBetweenCaptures int) Settings {
	return Settings{
		SecondsBetweenCaptures:  secondsBetweenCaptures,
		OffsetWithinHour:        DefaultOffsetWithinHour % secondsBetweenCaptures,
		PreviewResolutionWidth:  640,
		PreviewResolutionHeight: 480,
		ResolutionSetting:       1,
		Quality:                 100,
	}.withResolution()
}

// ValidInterval reports whether seconds is an allowed capture interval: at
// least a minute, at most an hour, and dividing the hour evenly so that every
// hour has the same capture times.
func ValidInterval(seconds int) bool {
	return seconds >= MinSecondsBetweenCaptures && seconds <= MaxSecondsBetweenCaptures && 3600%seconds == 0
}

// ValidIntervals lists every allowed interval in ascending order.
func ValidIntervals() []int {
	var out []int
	for s := MinSecondsBetweenCaptures; s <= MaxSecondsBetweenCaptures; s++ {
		if ValidInterval(s) {
			out = append(out, s)
		}
	}
	return out
}

// Validate reports the first problem with s, wrapping ErrInvalid.
func (s Settings) Validate() error {
	invalid := func(format string, args ...any) error {
		return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
	}
	if !ValidInterval(s.SecondsBetweenCaptures) {
		return invalid("SecondsBetweenCaptures must be between %d and %d and divide an hour evenly, got %d",
			MinSecondsBetweenCaptures, MaxSecondsBetweenCaptures, s.SecondsBetweenCaptures)
	}
	if s.OffsetWithinHour != OffsetDisabled && (s.OffsetWithinHour < 0 || s.OffsetWithinHour >= s.SecondsBetweenCaptures) {
		return invalid("OffsetWithinHour must be %d or between 0 and %d, got %d",
			OffsetDisabled, s.SecondsBetweenCaptures-1, s.OffsetWithinHour)
	}
	if s.Quality < 1 || s.Quality > 100 {
		return invalid("Quality must be between 1 and 100, got %d", s.Quality)
	}
	if s.RotateBy != 0 && s.RotateBy != 180 {
		return invalid("RotateBy must be 0 or 180, got %d", s.RotateBy)
	}
	if s.ResolutionSetting < 0 || s.ResolutionSetting >= len(resolutions) {
		return invalid("ResolutionSetting must be between 0 and %d, got %d", len(resolutions)-1, s.ResolutionSetting)
	}
	return nil
}

// Normalized repairs settings saved by older versions, which accepted any
// interval, into the closest valid ones while keeping the capture times the
// camera actually used:
//   - intervals round up to the next valid one (anything under a minute was
//     already treated as a minute, anything over an hour as an hour);
//   - an offset at or beyond the interval is reduced modulo the interval,
//     which yields the same capture times every hour.
func (s Settings) Normalized() Settings {
	if !ValidInterval(s.SecondsBetweenCaptures) {
		next := MaxSecondsBetweenCaptures
		for _, v := range ValidIntervals() {
			if v >= s.SecondsBetweenCaptures {
				next = v
				break
			}
		}
		s.SecondsBetweenCaptures = next
	}
	if s.OffsetWithinHour < 0 {
		s.OffsetWithinHour = OffsetDisabled
	} else {
		s.OffsetWithinHour %= s.SecondsBetweenCaptures
	}
	s.Quality = min(max(s.Quality, 1), 100)
	if s.RotateBy != 180 {
		s.RotateBy = 0 // only 180 ever had an effect on the camera
	}
	if s.ResolutionSetting < 0 || s.ResolutionSetting >= len(resolutions) {
		s.ResolutionSetting = 0 // unknown values always meant 3280x2464
	}
	return s.withResolution()
}

// withResolution sets the photo size from ResolutionSetting, which must be valid.
func (s Settings) withResolution() Settings {
	r := resolutions[s.ResolutionSetting]
	s.PhotoResolutionWidth, s.PhotoResolutionHeight = r[0], r[1]
	return s
}
