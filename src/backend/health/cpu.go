package health

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// CPUFreqDir is where Linux reports the clock of the first CPU (all cores
// of a Pi share it).
const CPUFreqDir = "/sys/devices/system/cpu/cpu0/cpufreq"

// CPUFreq is the CPU frequency scaling state. The image sets the powersave
// governor, which keeps the CPU at its lowest clock: 600 MHz on a Pi Zero 2 W,
// 700 MHz on a Pi Zero W.
type CPUFreq struct {
	Governor                   string
	CurrentMHz, MinMHz, MaxMHz int
}

// ReadCPUFreq reads the governor and clocks from dir, normally CPUFreqDir.
func ReadCPUFreq(dir string) (CPUFreq, error) {
	var c CPUFreq
	gov, err := os.ReadFile(filepath.Join(dir, "scaling_governor"))
	if err != nil {
		return c, err
	}
	c.Governor = strings.TrimSpace(string(gov))
	for _, f := range []struct {
		name string
		mhz  *int
	}{
		{"scaling_cur_freq", &c.CurrentMHz},
		{"scaling_min_freq", &c.MinMHz},
		{"scaling_max_freq", &c.MaxMHz},
	} {
		khz, err := readInt(filepath.Join(dir, f.name))
		if err != nil {
			return c, err
		}
		*f.mhz = khz / 1000
	}
	return c, nil
}

// ReadCelsius reads a thermal zone file, which holds millidegrees Celsius.
func ReadCelsius(path string) (float64, error) {
	milli, err := readInt(path)
	return float64(milli) / 1000, err
}

// readInt reads a file holding one integer, as the kernel's /sys files do.
func readInt(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return 0, fmt.Errorf("%s: %w", path, err)
	}
	return v, nil
}
