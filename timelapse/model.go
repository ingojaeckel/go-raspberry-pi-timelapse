package timelapse

import (
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

type Timelapse struct {
	Folder              string
	Settings            conf.Settings
	ConfigUpdateChannel <-chan conf.Settings
}
