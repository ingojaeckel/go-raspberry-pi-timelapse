// Package health reads the Pi's power, temperature and CPU state, records what
// happened between visits (daily summary, how the previous run ended) and
// logs problems as they appear, for cameras that run unattended for months.
package health

import (
	"fmt"
	"strconv"
	"strings"
)

// Throttled is the bit field printed by `vcgencmd get_throttled`. The low
// bits are the current state, the same bits shifted by 16 whether it
// happened since boot. Under-voltage is the usual cause of SD card
// corruption and failed captures on battery or solar power.
type Throttled uint32

var throttledFlags = [...]string{
	0: "under-voltage",
	1: "frequency capped",
	2: "throttled",
	3: "soft temperature limit",
}

const sinceBootShift = 16

// ParseThrottled parses vcgencmd output such as "throttled=0x50005".
func ParseThrottled(out string) (Throttled, error) {
	v, ok := strings.CutPrefix(strings.TrimSpace(out), "throttled=")
	if !ok {
		return 0, fmt.Errorf("unexpected vcgencmd output %q", strings.TrimSpace(out))
	}
	n, err := strconv.ParseUint(v, 0, 32)
	if err != nil {
		return 0, fmt.Errorf("unexpected vcgencmd output %q: %w", v, err)
	}
	return Throttled(n), nil
}

// Now lists the problems present at the moment.
func (t Throttled) Now() []string { return t.flags(0) }

// SinceBoot lists the problems that occurred at any time since boot.
func (t Throttled) SinceBoot() []string { return t.flags(sinceBootShift) }

func (t Throttled) flags(shift uint) []string {
	names := []string{}
	for bit, name := range throttledFlags {
		if t&(1<<(shift+uint(bit))) != 0 {
			names = append(names, name)
		}
	}
	return names
}

func (t Throttled) String() string { return fmt.Sprintf("0x%x", uint32(t)) }
