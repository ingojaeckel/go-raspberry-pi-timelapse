package health

import (
	"errors"
	"io/fs"
	"os"
	"strings"
)

// BootIDFile changes its content on every boot (Linux).
const BootIDFile = "/proc/sys/kernel/random/boot_id"

// PreviousRun is how the previous run of the app ended.
type PreviousRun int

const (
	// StoppedCleanly: the app was stopped (or this is its first start).
	StoppedCleanly PreviousRun = iota
	// AppDied: the app ended without stopping, but the Pi kept running:
	// a crash or a kill by the systemd watchdog. systemd restarted it.
	AppDied
	// PiRestarted: the Pi restarted while the app was running: a power cut,
	// the hardware watchdog or a kernel crash.
	PiRestarted
)

func (p PreviousRun) String() string {
	switch p {
	case AppDied:
		return "the app ended without stopping (crash or watchdog) and was restarted"
	case PiRestarted:
		return "the Pi restarted while the app was running (power cut, hardware watchdog or crash)"
	default:
		return "stopped cleanly"
	}
}

// StartRun reports how the previous run ended and marks this one as
// running: marker holds the boot ID while the app runs and is removed by
// done when it stops cleanly. bootIDFile is normally BootIDFile; without one
// (not Linux) every unclean end counts as AppDied.
func StartRun(marker, bootIDFile string) (previous PreviousRun, done func() error, err error) {
	boot := readBootID(bootIDFile)
	prev, err := os.ReadFile(marker)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		previous = StoppedCleanly
	case err != nil:
		return 0, nil, err
	case boot != "" && strings.TrimSpace(string(prev)) != boot:
		previous = PiRestarted
	default:
		previous = AppDied
	}
	if err := os.WriteFile(marker, []byte(boot+"\n"), 0o644); err != nil {
		return 0, nil, err
	}
	return previous, func() error { return os.Remove(marker) }, nil
}

func readBootID(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
