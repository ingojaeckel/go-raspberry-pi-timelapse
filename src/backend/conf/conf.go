// Package conf holds the capture settings, their validation and persistence,
// and the process-wide defaults that can be overridden by command line flags.
package conf

const (
	DefaultListenAddress          = ":8080"
	DefaultLogToFile              = false
	DefaultSecondsBetweenCaptures = 1800 // 30min
	DefaultOffsetWithinHour       = 900  // 15min
	DefaultStorageFolder          = "timelapse-pictures"
	DefaultCaptureRetries         = 2
	DefaultMinFreeDiskMB          = 256

	MaxFileSizeBytes = 100 << 20 // largest file served as a download
	TempFilesFolder  = "/tmp"

	// SettingsFile, LogFile and RunMarkerFile are relative to the working
	// directory (WorkingDirectory in timelapse.service).
	SettingsFile = "timelapse-settings.db"
	LogFile      = "timelapse.log"
	// RunMarkerFile exists while the app runs; finding it at start-up means
	// the previous run did not stop cleanly (see health.StartRun).
	RunMarkerFile  = "timelapse.running"
	LogFileMaxSize = 10 << 20 // bytes before the log is rotated to LogFile.1

	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentTypeJSON    = "application/json"
)
