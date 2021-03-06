package conf

const (
	DefaultListenAddress     = ":8080"
	DefaultLogToFile         = false
	MaxFileSizeBytes         = 100485760 // 100 MB
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentTypeJSON    = "application/json"
	DefaultStorageFolder     = "timelapse-pictures"
	TempFilesFolder          = "/tmp"
	LogFile                  = "timelapse.log"
)

var (
	LogToFile     = DefaultLogToFile
	ListenAddress = DefaultListenAddress
	StorageFolder = DefaultStorageFolder
)

// Update Updates configuration depending on the state of CLI flags provided.
func Update(listenAddress *string, storageAddress *string, logToFile *bool) {
	if logToFile != nil {
		LogToFile = *logToFile
	}
	if listenAddress != nil {
		ListenAddress = *listenAddress
	}
	if storageAddress != nil {
		StorageFolder = *storageAddress
	}
}
