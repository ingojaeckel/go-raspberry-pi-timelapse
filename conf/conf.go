package conf

const (
	DefaultListenAddress          = ":8080"
	DefaultLogToFile              = false
	DefaultSkipPhotosAtNight      = false
	DefaultSecondsBetweenCaptures = 1800      // 30min
	DefaultOffsetWithinHour       = 900       // 15min
	MaxFileSizeBytes              = 100485760 // 100 MB
	HeaderContentType             = "Content-Type"
	HeaderContentDisposition      = "Content-Disposition"
	HeaderContentTypeJSON         = "application/json"
	DefaultStorageFolder          = "timelapse-pictures"
	TempFilesFolder               = "/tmp"
	LogFile                       = "timelapse.log"
)

var (
	LogToFile              = DefaultLogToFile
	ListenAddress          = DefaultListenAddress
	StorageFolder          = DefaultStorageFolder
	SecondsBetweenCaptures = DefaultSecondsBetweenCaptures
)

// TODO merge values coming from different sources: config, settings, CLI

// OverrideDefaultConfig Override default config values which were provided.
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
