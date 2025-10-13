package timelapse

import (
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/detection"
)

type Timelapse struct {
	Folder              string
	Settings            conf.Settings
	ConfigUpdateChannel <-chan conf.Settings
	Detector            detection.Detector
	DetectionStore      *detection.ResultStore
}
