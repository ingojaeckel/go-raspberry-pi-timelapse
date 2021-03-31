package valid

import (
	"errors"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

var (
	errQualityOutOfBounds                = errors.New("Quality is out of bounds")
	errOffsetWithinHourOutOfBounds       = errors.New("Offset within hour is out of bounds")
	errSecondsBetweenCapturesOutOfBounds = errors.New("Seconds between captures is out of bounds")
)

type Validator interface {
	Validate(conf.Settings) error
}

func New() Validator {
	return strictValidator{}
}
