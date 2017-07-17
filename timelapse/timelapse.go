package timelapse

import (
	"errors"
	"fmt"
	"github.com/loranbriggs/go-camera"
	"os"
	"time"
)

type Timelapse struct {
	Camera             camera.Camera
	Folder             string
	TimeBetweenCapture time.Duration
}

func New(folder string, timeBetweenCapture time.Duration) (*Timelapse, error) {
	_, err := os.Stat(folder)
	createFolder := err != nil && os.IsNotExist(err)

	if createFolder {
		if err := os.Mkdir(folder, 0700); err != nil {
			return nil, err
		}
	}
	// Assume folder exists
	c := camera.New(folder)
	if c == nil {
		return nil, errors.New("Failed to instantiate camera")
	}

	return &Timelapse{*c, folder, timeBetweenCapture}, nil
}

func (t Timelapse) CapturePeriodically() {
	go func() {
		s, err := t.Camera.Capture()
		if err != nil {
			fmt.Errorf("Error during capture: %s\n", err.Error())
		}

		fmt.Printf("captured picture in %s\n", s)
		time.Sleep(t.TimeBetweenCapture)
	}()
}
