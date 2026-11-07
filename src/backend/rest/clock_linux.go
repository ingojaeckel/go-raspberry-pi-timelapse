package rest

import (
	"time"

	"golang.org/x/sys/unix"
)

// SetSystemClock sets the system clock. It needs CAP_SYS_TIME, which
// timelapse.service grants (AmbientCapabilities).
func SetSystemClock(t time.Time) error {
	ts := unix.NsecToTimespec(t.UnixNano())
	return unix.ClockSettime(unix.CLOCK_REALTIME, &ts)
}
