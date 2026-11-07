package rest

import (
	"syscall"
	"time"
)

// SetSystemClock sets the system clock (to the microsecond). It needs
// CAP_SYS_TIME, which timelapse.service grants (AmbientCapabilities).
func SetSystemClock(t time.Time) error {
	tv := syscall.NsecToTimeval(t.UnixNano())
	return syscall.Settimeofday(&tv)
}
