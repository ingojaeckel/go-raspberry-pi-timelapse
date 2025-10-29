package timelapse

import (
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/go-app/conf/settings"
)

type Timelapse struct {
	Folder              string
	Settings            settings.Settings
	ConfigUpdateChannel <-chan settings.Settings
}
