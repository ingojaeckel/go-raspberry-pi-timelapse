package valid

import (
	"testing"

	"github.com/facebookgo/ensure"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

func TestNew(t *testing.T) {
	ensure.Nil(t, New().Validate(conf.Settings{Quality: 100, SecondsBetweenCaptures: 60, OffsetWithinHour: 900}))
	ensure.Nil(t, New().Validate(conf.Settings{Quality: 100, SecondsBetweenCaptures: 60, OffsetWithinHour: -1}))
}

func TestOutOfBounds(t *testing.T) {
	ensure.NotNil(t, New().Validate(conf.Settings{Quality: 0}))
	ensure.NotNil(t, New().Validate(conf.Settings{Quality: 100, SecondsBetweenCaptures: 0}))
	ensure.NotNil(t, New().Validate(conf.Settings{Quality: 100, SecondsBetweenCaptures: 60, OffsetWithinHour: -2}))
}
