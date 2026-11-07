package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

const (
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
	OffsetWithinHour        int
	PhotoResolutionWidth    int
	PhotoResolutionHeight   int
	PreviewResolutionWidth  int
	PreviewResolutionHeight int
	// RotateBy is 0 or 180.
	RotateBy     int
	Quality      int
	DebugEnabled bool
	// StopWhenDiskFull stops taking photos when the disk is full. By default
	// (false) the oldest photos are deleted to make room, like a ring buffer.
	StopWhenDiskFull bool
}

func (s Settings) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

// Default returns the settings used when nothing has been saved yet.
func Default(secondsBetweenCaptures int) Settings {
	return Settings{
		SecondsBetweenCaptures:  secondsBetweenCaptures,
		OffsetWithinHour:        DefaultOffsetWithinHour % secondsBetweenCaptures,
		PhotoResolutionWidth:    DefaultPhotoWidth,
		PhotoResolutionHeight:   DefaultPhotoHeight,
		PreviewResolutionWidth:  640,
		PreviewResolutionHeight: 480,
		Quality:                 100,
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

// Validate reports the first problem with s, wrapping ErrInvalid.
func (s Settings) Validate() error {
	invalid := func(format string, args ...any) error {
		return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
	}
	if !ValidInterval(s.SecondsBetweenCaptures) {
		return invalid("SecondsBetweenCaptures must be between %d and %d and divide an hour evenly, got %d",
			MinSecondsBetweenCaptures, MaxSecondsBetweenCaptures, s.SecondsBetweenCaptures)
	}
	if !validOffset(s.OffsetWithinHour, s.SecondsBetweenCaptures) {
		return invalid("OffsetWithinHour must be %d or between 0 and %d, got %d",
			OffsetDisabled, s.SecondsBetweenCaptures-1, s.OffsetWithinHour)
	}
	if s.Quality < 1 || s.Quality > 100 {
		return invalid("Quality must be between 1 and 100, got %d", s.Quality)
	}
	if s.RotateBy != 0 && s.RotateBy != 180 {
		return invalid("RotateBy must be 0 or 180, got %d", s.RotateBy)
	}
	if !validSide(s.PhotoResolutionWidth) || !validSide(s.PhotoResolutionHeight) {
		return invalid("the photo size must be between %d and %d pixels on each side, got %dx%d",
			MinPhotoSide, MaxPhotoSide, s.PhotoResolutionWidth, s.PhotoResolutionHeight)
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
	s.SecondsBetweenCaptures = roundUpInterval(s.SecondsBetweenCaptures)
	if s.OffsetWithinHour < 0 {
		s.OffsetWithinHour = OffsetDisabled
	} else {
		s.OffsetWithinHour %= s.SecondsBetweenCaptures
	}
	s.Quality = min(max(s.Quality, 1), 100)
	if s.RotateBy != 180 {
		s.RotateBy = 0 // only 180 ever had an effect on the camera
	}
	if !validSide(s.PhotoResolutionWidth) || !validSide(s.PhotoResolutionHeight) {
		s.PhotoResolutionWidth, s.PhotoResolutionHeight = DefaultPhotoWidth, DefaultPhotoHeight
	}
	return s
}

// roundUpInterval returns the smallest valid interval >= seconds, or the
// longest one.
func roundUpInterval(seconds int) int {
	valid := ValidIntervals()
	if i := slices.IndexFunc(valid, func(v int) bool { return v >= seconds }); i >= 0 {
		return valid[i]
	}
	return MaxSecondsBetweenCaptures
}

func validOffset(offset, interval int) bool {
	return offset == OffsetDisabled || (offset >= 0 && offset < interval)
}

func validSide(pixels int) bool {
	return pixels >= MinPhotoSide && pixels <= MaxPhotoSide
}
