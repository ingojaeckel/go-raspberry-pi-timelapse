package timelapse

import "github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"

// PhotoOptions are the options for scheduled photos.
func PhotoOptions(s conf.Settings) Options {
	return Options{Width: s.PhotoResolutionWidth, Height: s.PhotoResolutionHeight, Quality: s.Quality, Rotate: s.RotateBy == 180}
}

// PreviewOptions are the options for preview captures.
func PreviewOptions(s conf.Settings) Options {
	return Options{Width: s.PreviewResolutionWidth, Height: s.PreviewResolutionHeight, Quality: s.Quality, Rotate: s.RotateBy == 180}
}
