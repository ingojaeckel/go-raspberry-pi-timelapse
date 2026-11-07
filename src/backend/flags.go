package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

type options struct {
	listen                 string
	logToFile              bool
	storageFolder          string
	secondsBetweenCaptures int
	captureRetries         int
	minFreeDiskMB          int
	pprof                  bool
	version                bool
	allowedHosts           string
	selfCheck              bool
	selfCheckKeep          bool
	selfCheckMinBrightness int
}

// defaultSelfCheckMinBrightness is the lowest mean brightness (0..255) the
// self-check accepts: lower is a black photo (lens cap, no light, no signal).
const defaultSelfCheckMinBrightness = 10

func parseFlags(args []string) (options, error) {
	var o options
	fs := flag.NewFlagSet("go-raspberry-pi-timelapse", flag.ContinueOnError)
	o.register(fs)
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	return o, o.validate()
}

func (o *options) register(fs *flag.FlagSet) {
	fs.BoolVar(&o.version, "version", false, "Print version and exit.")
	fs.StringVar(&o.listen, "port", conf.DefaultListenAddress, "HTTP address to listen on.")
	fs.BoolVar(&o.logToFile, "logToFile", conf.DefaultLogToFile, "Log to "+conf.LogFile+" (rotated at 10 MB) instead of stdout. Recommended for long term operation.")
	fs.StringVar(&o.storageFolder, "storageFolder", conf.DefaultStorageFolder, "Folder for storage of timelapse pictures.")
	fs.IntVar(&o.secondsBetweenCaptures, "secondsBetweenCaptures", conf.DefaultSecondsBetweenCaptures, "Seconds between captures used until settings are saved in the UI. Must divide an hour evenly, between 60 and 3600.")
	fs.IntVar(&o.captureRetries, "captureRetries", conf.DefaultCaptureRetries, "Additional attempts within the same slot when a capture fails.")
	fs.IntVar(&o.minFreeDiskMB, "minFreeDiskMB", conf.DefaultMinFreeDiskMB, "Free disk space to keep. The oldest photos are deleted before a capture to keep at least this much free.")
	fs.BoolVar(&o.pprof, "pprof", false, "Enable pprof profiling endpoints at /debug/pprof/")
	fs.BoolVar(&o.selfCheck, "selfcheck", false, "Take one test photo with the saved settings, check it and the Pi (camera detected and free, storage, power, temperature), print the results and exit: 0 when no check failed, 1 otherwise. Run it after installing or changing the hardware.")
	fs.BoolVar(&o.selfCheckKeep, "selfcheckKeep", false, "With -selfcheck, keep the test photo and print its path.")
	fs.IntVar(&o.selfCheckMinBrightness, "selfcheckMinBrightness", defaultSelfCheckMinBrightness, "With -selfcheck, the lowest mean brightness (0-255) of a usable photo. Lower it, or use 0, when testing in the dark.")
	fs.StringVar(&o.allowedHosts, "allowedHosts", "", "Comma-separated public host names the UI may be reached under. IP addresses, localhost, single-label names and .local/.lan/.home.arpa/.internal names are always allowed; anything else is rejected to block DNS rebinding.")
}

func (o options) validate() error {
	switch {
	case !conf.ValidInterval(o.secondsBetweenCaptures):
		return fmt.Errorf("-secondsBetweenCaptures must divide an hour evenly and be between %d and %d, got %d", conf.MinSecondsBetweenCaptures, conf.MaxSecondsBetweenCaptures, o.secondsBetweenCaptures)
	case o.captureRetries < 0:
		return errors.New("-captureRetries must not be negative")
	case o.minFreeDiskMB < 0:
		return errors.New("-minFreeDiskMB must not be negative")
	case o.selfCheckMinBrightness < 0 || o.selfCheckMinBrightness > 255:
		return errors.New("-selfcheckMinBrightness must be between 0 and 255")
	}
	return nil
}

// allowedHostList splits -allowedHosts into host names.
func (o options) allowedHostList() []string {
	var hosts []string
	for h := range strings.SplitSeq(o.allowedHosts, ",") {
		if h = strings.TrimSpace(h); h != "" {
			hosts = append(hosts, h)
		}
	}
	return hosts
}
