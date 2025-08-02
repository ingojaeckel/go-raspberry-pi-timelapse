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
	ObjectDetectionEnabled:  false,
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
	ObjectDetectionEnabled  bool
}

func (s Settings) String() string {
	jsonStr, _ := json.Marshal(s)
	return string(jsonStr)
}
