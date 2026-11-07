// Package conf holds the capture settings: their rules (Validate,
// Normalized) and their storage (Store).
package conf

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/sun"
)

const (
	DefaultSecondsBetweenCaptures = 1800 // 30 min
	DefaultOffsetWithinHour       = 900  // 15 min

	MinSecondsBetweenCaptures = 60
	MaxSecondsBetweenCaptures = 3600
	// OffsetDisabled turns off alignment to the hour: captures then run every
	// SecondsBetweenCaptures counted from start-up.
	OffsetDisabled = -1

	// DefaultPhotoWidth x DefaultPhotoHeight is the photo size until one is
	// chosen, unless the app detects the camera and picks one of its modes.
	DefaultPhotoWidth  = 2186
	DefaultPhotoHeight = 1642
	// MinPhotoSide and MaxPhotoSide bound the photo width and height; which
	// sizes the camera supports is checked by the API (timelapse.Sensor).
	MinPhotoSide = 64
	MaxPhotoSide = 16384
)

// ErrInvalid is wrapped by every validation error.
var ErrInvalid = errors.New("invalid settings")

// Settings is persisted as JSON and served to the frontend unchanged; the
// field names are part of the API (src/frontend/src/api.ts).
type Settings struct {
	SecondsBetweenCaptures int
	// OffsetWithinHour is the first capture in seconds past the hour,
	// 0 <= offset < SecondsBetweenCaptures, or OffsetDisabled.
	OffsetWithinHour      int
	PhotoResolutionWidth  int
	PhotoResolutionHeight int
	// RotateBy is 0 or 180.
	RotateBy     int
	Quality      int
	DebugEnabled bool
	// StopWhenDiskFull stops taking photos when the disk is full. By default
	// (false) the oldest photos are deleted to make room, like a ring buffer.
	StopWhenDiskFull bool
	// SkipNightPhotos pauses captures between sunset and sunrise at
	// Latitude/Longitude (degrees; north and east positive), computed
	// offline (package sun). Off by default: photos are taken around the
	// clock.
	SkipNightPhotos     bool
	Latitude, Longitude float64
}

func (s Settings) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

// Default returns the settings used when nothing has been saved yet.
func Default(secondsBetweenCaptures int) Settings {
	return Settings{
		SecondsBetweenCaptures: secondsBetweenCaptures,
		OffsetWithinHour:       DefaultOffsetWithinHour % secondsBetweenCaptures,
		PhotoResolutionWidth:   DefaultPhotoWidth,
		PhotoResolutionHeight:  DefaultPhotoHeight,
		Quality:                100,
	}
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

// A rule is one requirement on the settings: what Validate reports when it
// is broken, and how Normalized repairs it.
type rule struct {
	ok      func(Settings) bool
	problem func(Settings) string
	repair  func(*Settings)
}

// rules are checked and repaired in order: the offset depends on the
// interval, skipping nights on the location.
var rules = []rule{{
	ok: func(s Settings) bool { return ValidInterval(s.SecondsBetweenCaptures) },
	problem: func(s Settings) string {
		return fmt.Sprintf("SecondsBetweenCaptures must be between %d and %d and divide an hour evenly, got %d",
			MinSecondsBetweenCaptures, MaxSecondsBetweenCaptures, s.SecondsBetweenCaptures)
	},
	// Older versions accepted any interval, treating anything under a minute
	// as a minute and anything over an hour as an hour.
	repair: func(s *Settings) { s.SecondsBetweenCaptures = roundUpInterval(s.SecondsBetweenCaptures) },
}, {
	ok: func(s Settings) bool { return validOffset(s.OffsetWithinHour, s.SecondsBetweenCaptures) },
	problem: func(s Settings) string {
		return fmt.Sprintf("OffsetWithinHour must be %d or between 0 and %d, got %d", OffsetDisabled, s.SecondsBetweenCaptures-1, s.OffsetWithinHour)
	},
	// An offset at or beyond the interval, reduced modulo the interval, gives
	// the same capture times every hour.
	repair: func(s *Settings) {
		if s.OffsetWithinHour < 0 {
			s.OffsetWithinHour = OffsetDisabled
		} else {
			s.OffsetWithinHour %= s.SecondsBetweenCaptures
		}
	},
}, {
	ok:      func(s Settings) bool { return s.Quality >= 1 && s.Quality <= 100 },
	problem: func(s Settings) string { return fmt.Sprintf("Quality must be between 1 and 100, got %d", s.Quality) },
	repair:  func(s *Settings) { s.Quality = min(max(s.Quality, 1), 100) },
}, {
	ok:      func(s Settings) bool { return s.RotateBy == 0 || s.RotateBy == 180 },
	problem: func(s Settings) string { return fmt.Sprintf("RotateBy must be 0 or 180, got %d", s.RotateBy) },
	repair:  func(s *Settings) { s.RotateBy = 0 }, // only 180 ever had an effect on the camera
}, {
	ok: func(s Settings) bool { return sun.ValidLocation(s.Latitude, s.Longitude) },
	problem: func(s Settings) string {
		return fmt.Sprintf("Latitude must be between -90 and 90 and Longitude between -180 and 180, got %v, %v", s.Latitude, s.Longitude)
	},
	repair: func(s *Settings) { s.Latitude, s.Longitude = 0, 0 },
}, {
	ok:      func(s Settings) bool { return !s.SkipNightPhotos || s.Latitude != 0 || s.Longitude != 0 },
	problem: func(Settings) string { return "SkipNightPhotos needs the camera's location (Latitude and Longitude)" },
	repair:  func(s *Settings) { s.SkipNightPhotos = false },
}, {
	ok: func(s Settings) bool { return validSide(s.PhotoResolutionWidth) && validSide(s.PhotoResolutionHeight) },
	problem: func(s Settings) string {
		return fmt.Sprintf("the photo size must be between %d and %d pixels on each side, got %dx%d",
			MinPhotoSide, MaxPhotoSide, s.PhotoResolutionWidth, s.PhotoResolutionHeight)
	},
	repair: func(s *Settings) {
		s.PhotoResolutionWidth, s.PhotoResolutionHeight = DefaultPhotoWidth, DefaultPhotoHeight
	},
}}

// Validate reports the first problem with s, wrapping ErrInvalid.
func (s Settings) Validate() error {
	for _, r := range rules {
		if !r.ok(s) {
			return fmt.Errorf("%w: %s", ErrInvalid, r.problem(s))
		}
	}
	return nil
}

// Normalized repairs whatever Validate would reject, e.g. settings saved by
// older versions, keeping the capture times the camera actually used where
// possible. The result always passes Validate.
func (s Settings) Normalized() Settings {
	for _, r := range rules {
		if !r.ok(s) {
			r.repair(&s)
		}
	}
	return s
}

// roundUpInterval returns the smallest valid interval >= seconds, or the
// longest one.
func roundUpInterval(seconds int) int {
	for v := max(seconds, MinSecondsBetweenCaptures); v < MaxSecondsBetweenCaptures; v++ {
		if ValidInterval(v) {
			return v
		}
	}
	return MaxSecondsBetweenCaptures
}

func validOffset(offset, interval int) bool {
	return offset == OffsetDisabled || (offset >= 0 && offset < interval)
}

func validSide(pixels int) bool {
	return pixels >= MinPhotoSide && pixels <= MaxPhotoSide
}
