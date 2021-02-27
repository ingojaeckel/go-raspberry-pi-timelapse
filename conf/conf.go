package conf

const (
	ListenAddress            = ":8080"
	MaxFileSizeBytes         = 100485760 // 100 MB
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentTypeJSON    = "application/json"
	StorageFolder            = "timelapse-pictures"
	TempFilesFolder          = "/tmp"
	LogFile                  = "timelapse.log"
	LogToFile                = false
)
