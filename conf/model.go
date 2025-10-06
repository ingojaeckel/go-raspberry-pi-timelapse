package conf

import (
	"encoding/json"
	"errors"
)

const (
	settingsFile = "timelapse-settings.db"
	bucket       = "settings"
	settingsKey  = "s"
)

var missingBucketError = errors.New("missing bucket")
var settingsNotFound = errors.New("settings not found")

var initialConfiguration = Settings{
	DebugEnabled:           false,
	SecondsBetweenCaptures: SecondsBetweenCaptures,
	OffsetWithinHour:       DefaultOffsetWithinHour,
	// Default resolution: 3280x2464 (8 MP). 66%: 2186x1642 (3.5 MP), 50%: 1640x1232 (2 MP)
	PhotoResolutionWidth:    2186,
	PhotoResolutionHeight:   1642,
	PreviewResolutionWidth:  640,
	PreviewResolutionHeight: 480,
	RotateBy:                0,
	ResolutionSetting:       0,
	Quality:                 100,
}

type Settings struct {
	SecondsBetweenCaptures  int
	OffsetWithinHour        int
	PhotoResolutionWidth    int
	PhotoResolutionHeight   int
	PreviewResolutionWidth  int
	PreviewResolutionHeight int
	RotateBy                int
	ResolutionSetting       int
	Quality                 int
	DebugEnabled            bool
}

func (s Settings) String() string {
	jsonStr, _ := json.Marshal(s)
	return string(jsonStr)
}

// Sanitize ensures all configuration values are within acceptable bounds.
// This prevents broken configuration from being persisted or used.
func (s Settings) Sanitize() Settings {
	sanitized := s

	// Enforce minimum quality (cannot be 0 or negative)
	if sanitized.Quality < MinQuality {
		sanitized.Quality = MinQuality
	}
	// Enforce maximum quality
	if sanitized.Quality > MaxQuality {
		sanitized.Quality = MaxQuality
	}

	// Enforce minimum seconds between captures
	if sanitized.SecondsBetweenCaptures < MinSecondsBetweenCaptures {
		sanitized.SecondsBetweenCaptures = MinSecondsBetweenCaptures
	}

	// Enforce offset within hour bounds (-1 is allowed to disable, otherwise 0-3599)
	if sanitized.OffsetWithinHour != -1 {
		if sanitized.OffsetWithinHour < MinOffsetWithinHour {
			sanitized.OffsetWithinHour = MinOffsetWithinHour
		}
		if sanitized.OffsetWithinHour > MaxOffsetWithinHour {
			sanitized.OffsetWithinHour = MaxOffsetWithinHour
		}
	}

	return sanitized
}
