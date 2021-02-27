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
	ListenAddress = DefaultListenAddress
	LogToFile     = DefaultLogToFile
	StorageFolder = DefaultStorageFolder
)
