package timelapse

import (
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

// Monitor interface for system monitoring
type Monitor interface {
	PeriodicCheck()
}

type Timelapse struct {
	Folder              string
	Settings            conf.Settings
	ConfigUpdateChannel <-chan conf.Settings
	Monitor             Monitor
}
