package conf

import "github.com/ingojaeckel/go-raspberry-pi-timelapse/timelapse"

const (
	Version                  = 1
	ListenAddress            = ":8080"
	MaxFileSizeBytes         = 100485760 // 100 MB
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentTypeJSON    = "application/json"
	StorageFolder            = "timelapse-pictures"
	TempFilesFolder          = "/tmp"
	LogFile                  = "timelapse.log"
)

// Default resolution: 3280x2464 (8 MP). 66%: 2186x1642 (3.5 MP), 50%: 1640x1232 (2 MP)
var PhotoResolution = timelapse.Resolution{2186, 1642}
var PreviewResolution = timelapse.Resolution{640, 480}
