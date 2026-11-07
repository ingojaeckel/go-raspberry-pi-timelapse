package main

import (
	"bufio"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"
)

// serviceFile is the systemd unit the image runs the app with.
const serviceFile = "../installer/custompios/modules/timelapsepi/filesystem/root/etc/systemd/system/timelapse.service"

// readUnit returns the key=value settings of a systemd unit file.
func readUnit(t *testing.T, path string) map[string]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	settings := map[string]string{}
	lines := bufio.NewScanner(f)
	for lines.Scan() {
		key, value, ok := strings.Cut(lines.Text(), "=")
		if ok && !strings.HasPrefix(key, "#") {
			settings[key] = value
		}
	}
	return settings
}

// unitDuration parses systemd time spans of the form used in the unit:
// "20" (seconds), "90s" or "2min".
func unitDuration(t *testing.T, span string) time.Duration {
	t.Helper()
	d, err := time.ParseDuration(strings.Replace(span, "min", "m", 1))
	if err != nil {
		d, err = time.ParseDuration(span + "s")
	}
	if err != nil {
		t.Fatalf("time span %q: %v", span, err)
	}
	return d
}

// A renamed or removed flag would make the image's app exit at every start.
func TestServiceStartsTheAppWithValidFlags(t *testing.T) {
	unit := readUnit(t, serviceFile)
	args := strings.Fields(unit["ExecStart"])
	if len(args) < 2 || args[0] != "/opt/timelapse/timelapse" {
		t.Fatalf("ExecStart = %q", unit["ExecStart"])
	}
	if _, err := parseFlags(args[1:]); err != nil {
		t.Fatalf("ExecStart flags: %v", err)
	}
}

// The watchdog only pings systemd while the scheduler wakes up within
// maxStall; an idle scheduler wakes every MaxSleep, which must count as alive.
func TestServiceWatchdogToleratesTheSchedulersPace(t *testing.T) {
	unit := readUnit(t, serviceFile)
	opts, err := parseFlags(strings.Fields(unit["ExecStart"])[1:])
	if err != nil {
		t.Fatal(err)
	}
	watchdogSec := unitDuration(t, unit["WatchdogSec"])
	if ping := watchdogSec / 2; ping < time.Second {
		t.Errorf("WatchdogSec=%v: pings every %v", watchdogSec, ping)
	}
	if stall := (&app{opts: opts}).maxStall(); timelapse.MaxSleep >= stall {
		t.Errorf("scheduler wakes every %v, but the watchdog calls a gap of %v a stall", timelapse.MaxSleep, stall)
	}
}

// systemd kills the app TimeoutStopSec after SIGTERM: it must have time to
// finish its HTTP responses and stop a running capture.
func TestServiceWaitsForACleanShutdown(t *testing.T) {
	stopTimeout := unitDuration(t, readUnit(t, serviceFile)["TimeoutStopSec"])
	if need := shutdownTimeout + timelapse.KillDelay; stopTimeout <= need {
		t.Fatalf("TimeoutStopSec=%v, the app needs up to %v to stop", stopTimeout, need)
	}
}
