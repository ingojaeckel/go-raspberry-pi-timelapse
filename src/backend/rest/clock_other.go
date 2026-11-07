//go:build !linux

package rest

import (
	"errors"
	"time"
)

// SetSystemClock is only implemented on Linux, the Pi.
func SetSystemClock(time.Time) error { return errors.ErrUnsupported }
