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

	MaxFileSizeBytes = 100485760 // 100 MB
	TempFilesFolder  = "/tmp"

	// SettingsFile and LogFile are relative to the working directory
	// (WorkingDirectory in timelapse.service).
	SettingsFile   = "timelapse-settings.db"
	LogFile        = "timelapse.log"
	LogFileMaxSize = 10 << 20 // bytes before the log is rotated to LogFile.1

	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentTypeJSON    = "application/json"
)
