package valid

import (
	"testing"

	"github.com/facebookgo/ensure"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

func TestNew(t *testing.T) {
	ensure.Nil(t, New().Validate(conf.Settings{Quality: 100, SecondsBetweenCaptures: 60, OffsetWithinHour: 900}))
	ensure.Nil(t, New().Validate(conf.Settings{Quality: 100, SecondsBetweenCaptures: 60, OffsetWithinHour: -1}))
	// Test with minimum valid values
	ensure.Nil(t, New().Validate(conf.Settings{Quality: conf.MinQuality, SecondsBetweenCaptures: conf.MinSecondsBetweenCaptures, OffsetWithinHour: 0}))
}

func TestOutOfBounds(t *testing.T) {
	ensure.NotNil(t, New().Validate(conf.Settings{Quality: 0}))
	ensure.NotNil(t, New().Validate(conf.Settings{Quality: 100, SecondsBetweenCaptures: 0}))
	ensure.NotNil(t, New().Validate(conf.Settings{Quality: 100, SecondsBetweenCaptures: 60, OffsetWithinHour: -2}))
	// Test with values below minimum
	ensure.NotNil(t, New().Validate(conf.Settings{Quality: 100, SecondsBetweenCaptures: conf.MinSecondsBetweenCaptures - 1, OffsetWithinHour: 0}))
	ensure.NotNil(t, New().Validate(conf.Settings{Quality: conf.MinQuality - 1, SecondsBetweenCaptures: conf.MinSecondsBetweenCaptures, OffsetWithinHour: 0}))
}
