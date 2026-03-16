package valid

import "github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"

type strictValidator struct{}

func (s strictValidator) Validate(settings conf.Settings) error {
	if settings.Quality < conf.MinQuality || settings.Quality > conf.MaxQuality {
		return errQualityOutOfBounds
	}
	if settings.SecondsBetweenCaptures < conf.MinSecondsBetweenCaptures {
		return errSecondsBetweenCapturesOutOfBounds
	}
	if settings.OffsetWithinHour != -1 {
		if settings.OffsetWithinHour < conf.MinOffsetWithinHour || settings.OffsetWithinHour > conf.MaxOffsetWithinHour {
			return errOffsetWithinHourOutOfBounds
		}
	}
	return nil
}
