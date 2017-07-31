package timelapse

import (
	"errors"
	"fmt"
	"github.com/loranbriggs/go-camera"
	"os"
	"time"
)

type Timelapse struct {
	Camera                camera.Camera
	Folder                string
	SecondsBetweenCapture int64
}

func New(folder string, secondsBetweenCapture int64) (*Timelapse, error) {
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

	return &Timelapse{*c, folder, secondsBetweenCapture}, nil
}

func (t Timelapse) CapturePeriodically() {
	go func() {
		for {
			beforeCapture := time.Now()
			s, err := t.Camera.Capture()
			if err != nil {
				fmt.Errorf("Error during capture: %s\n", err.Error())
			}

			fmt.Printf("captured picture in %s\n", s)
			timeToCaptureSeconds := time.Now().Unix() - beforeCapture.Unix()

			sleepTime := t.SecondsBetweenCapture - timeToCaptureSeconds
			fmt.Printf("capture took %d seconds, will sleep for %d seconds\n", timeToCaptureSeconds, sleepTime)
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}()
}
