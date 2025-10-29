package valid

import "github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"

type strictValidator struct{}

func (s strictValidator) Validate(settings conf.Settings) error {
	if settings.Quality < 1 || settings.Quality > 100 {
		return errQualityOutOfBounds
	}
	if settings.SecondsBetweenCaptures < 1 {
		return errSecondsBetweenCapturesOutOfBounds
	}
	if settings.OffsetWithinHour != -1 {
		if settings.OffsetWithinHour < 0 || settings.OffsetWithinHour > 3599 {
			return errOffsetWithinHourOutOfBounds
		}
	}
	return nil
}
