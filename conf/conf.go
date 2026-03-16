package conf

const (
	DefaultListenAddress          = ":8080"
	DefaultLogToFile              = false
	DefaultSecondsBetweenCaptures = 1800      // 30min
	DefaultOffsetWithinHour       = 900       // 15min
	MaxFileSizeBytes              = 100485760 // 100 MB
	HeaderContentType             = "Content-Type"
	HeaderContentDisposition      = "Content-Disposition"
	HeaderContentTypeJSON         = "application/json"
	DefaultStorageFolder          = "timelapse-pictures"
	TempFilesFolder               = "/tmp"
	LogFile                       = "timelapse.log"
	// Validation bounds
	MinSecondsBetweenCaptures = 10  // Allow sufficient exposure time
	MinQuality                = 1   // Quality must be at least 1
	MaxQuality                = 100 // Quality cannot exceed 100
	MinOffsetWithinHour       = 0   // Offset within hour minimum
	MaxOffsetWithinHour       = 3599 // Offset within hour maximum (59 minutes 59 seconds)
)

var (
	LogToFile              = DefaultLogToFile
	ListenAddress          = DefaultListenAddress
	StorageFolder          = DefaultStorageFolder
	SecondsBetweenCaptures = DefaultSecondsBetweenCaptures
)

// OverrideDefaultConfig Override default config values which were provided.
// Note: These global values are used for initial configuration only. 
// The Settings struct handles the actual runtime configuration with proper priority:
// CLI flags → persisted settings → defaults (see CONFIGURATION.md for details)
func OverrideDefaultConfig(listenAddressOverride *string, storageAddressOverride *string, logToFileOverride *bool, secondsBetweenCapturesOverride *int) {
	if logToFileOverride != nil {
		LogToFile = *logToFileOverride
	}
	if listenAddressOverride != nil {
		ListenAddress = *listenAddressOverride
	}
	if storageAddressOverride != nil {
		StorageFolder = *storageAddressOverride
	}
	if secondsBetweenCapturesOverride != nil {
		SecondsBetweenCaptures = *secondsBetweenCapturesOverride
	}
}
