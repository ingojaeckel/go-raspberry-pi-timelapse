package cli

import (
	"flag"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/go-app/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/go-app/conf/settings"
)

// Flags holds the state of command line flags.
type Flags struct {
	Version                *bool
	ListenAddress          *string
	LogToFile              *bool
	StorageFolder          *string
	SecondsBetweenCaptures *int
	EnablePprof            *bool
}

func ParseFlags() Flags {
	versionFlag := flag.Bool("version", false, "Print version and exit.")
	listenAddress := flag.String("port", conf.DefaultListenAddress, "HTTP port to listen on.")
	logToFile := flag.Bool("logToFile", conf.DefaultLogToFile, "Toggle to enable logging to a file on disk instead of stdout. Logging to a file is recommended for long term operation.")
	storageAddress := flag.String("storageFolder", conf.DefaultStorageFolder, "Folder for storage of timelapse pictures.")
	secondsBetweenCaptures := flag.Int("secondsBetweenCaptures", conf.DefaultSecondsBetweenCaptures, "Number of seconds between captures")
	enablePprof := flag.Bool("pprof", false, "Enable pprof profiling endpoints at /debug/pprof/")
	flag.Parse()

	return Flags{
		Version:                versionFlag,
		ListenAddress:          listenAddress,
		LogToFile:              logToFile,
		StorageFolder:          storageAddress,
		SecondsBetweenCaptures: secondsBetweenCaptures,
		EnablePprof:            enablePprof,
	}
}

// OverrideDefaultConfig Override default config values which were provided.
// Overlay the provided flag values on top of the given settings. Return the modified settings.
func OverrideDefaultConfig(s settings.Settings, f Flags) settings.Settings {
	if f.LogToFile != nil {
		s.LogToFile = *f.LogToFile
	}
	if f.ListenAddress != nil {
		s.ListenAddress = *f.ListenAddress
	}
	if f.StorageFolder != nil {
		s.StorageFolder = *f.StorageFolder
	}
	if f.SecondsBetweenCaptures != nil {
		s.SecondsBetweenCaptures = *f.SecondsBetweenCaptures
	}
	return s
}
