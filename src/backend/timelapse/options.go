package timelapse

import "github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"

// PhotoOptions are the options for scheduled photos.
func PhotoOptions(s conf.Settings) Options {
	return Options{Width: s.PhotoResolutionWidth, Height: s.PhotoResolutionHeight, Quality: s.Quality, Rotate: s.RotateBy == 180}
}

// PreviewWidth x PreviewHeight is the size of preview captures: enough for a
// browser window, quick to take and send.
const (
	PreviewWidth  = 640
	PreviewHeight = 480
)

// PreviewOptions are the options for preview captures.
func PreviewOptions(s conf.Settings) Options {
	return Options{Width: PreviewWidth, Height: PreviewHeight, Quality: s.Quality, Rotate: s.RotateBy == 180}
}
