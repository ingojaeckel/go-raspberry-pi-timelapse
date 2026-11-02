package timelapse

import (
	"log"
	"os"
	"time"

	"github.com/ingojaeckel/go-raspberry-pi-timelapse/conf"
)

func New(folder string, initialSettings conf.Settings, configUpdatedChan <-chan conf.Settings) (*Timelapse, error) {
	_, err := os.Stat(folder)
	createFolder := err != nil && os.IsNotExist(err)

	if createFolder {
		if err := os.Mkdir(folder, 0700); err != nil {
			return nil, err
		}
	}
	// Assume folder exists

	return &Timelapse{
		Folder:              folder,
		Settings:            initialSettings,
		ConfigUpdateChannel: configUpdatedChan,
	}, nil
}

func (t *Timelapse) CapturePeriodically() {
	go func() {
		for {
			t.waitForCapture()

			beforeCapture := time.Now()

			camera, err := NewCamera(t.Folder, t.Settings.PhotoResolutionWidth, t.Settings.PhotoResolutionHeight, t.Settings.RotateBy == 180, t.Settings.Quality)
			if err != nil {
				log.Printf("Error instantiating camera: %s\n", err)
				time.Sleep(time.Duration(t.Settings.SecondsBetweenCaptures) * time.Second)
				continue
			}

			photoPath, err := camera.Capture()
			if err != nil {
				log.Printf("Error during capture: %s\n", err.Error())
				time.Sleep(1 * time.Second)
				continue
			}

			log.Printf("Photo stored in '%s'\n", photoPath)
			timeToCaptureSeconds := time.Now().Unix() - beforeCapture.Unix()
			log.Printf("Capture took %d seconds\n", timeToCaptureSeconds)
		}
	}()
}

func (t *Timelapse) waitForCapture() {
	offsetDisabled := t.Settings.OffsetWithinHour < 0

	if offsetDisabled {
		log.Printf("Offset is disabled. Sleeping for %d seconds.\n", t.Settings.SecondsBetweenCaptures)
		t.sleepWithConfigUpdates(time.Duration(t.Settings.SecondsBetweenCaptures) * time.Second)
		return
	}

	secondsUntilFirstCapture := t.getSecondsToFirstCapture(time.Now())
	sleepDuration := time.Duration(secondsUntilFirstCapture) * time.Second
	nextCaptureAt := time.Now().Add(sleepDuration)

	log.Printf("Will take the next picture in %d seconds at %v.\n", secondsUntilFirstCapture, nextCaptureAt)

	for {
		secondsUntilFirstCapture := t.getSecondsToFirstCapture(time.Now())
		if secondsUntilFirstCapture == 0 {
			return
		}
		if t.Settings.DebugEnabled {
			log.Printf("Sleeping for 1 second. Seconds left: %d. Time: %s.\n", secondsUntilFirstCapture, time.Now())
		}
		select {
		case newConfig := <-t.ConfigUpdateChannel:
			log.Printf("Received new configuration: %s\n", newConfig)
			t.Settings = newConfig
		case <-time.After(1 * time.Second):
		}
	}
}

// sleepWithConfigUpdates sleeps for the given duration while still processing config updates.
func (t *Timelapse) sleepWithConfigUpdates(d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	for {
		select {
		case newConfig := <-t.ConfigUpdateChannel:
			log.Printf("Received new configuration: %s\n", newConfig)
			t.Settings = newConfig
		case <-timer.C:
			return
		}
	}
}

func (t Timelapse) getSecondsToFirstCapture(currentTime time.Time) int {
	picturesPerHour := 3600 / t.Settings.SecondsBetweenCaptures
	secondsIntoCurrentHour := currentTime.Minute()*60 + currentTime.Second()

	for i := 0; i < int(picturesPerHour); i++ {
		if i == 0 {
			if 0 <= secondsIntoCurrentHour && secondsIntoCurrentHour <= t.Settings.OffsetWithinHour {
				return t.Settings.OffsetWithinHour - secondsIntoCurrentHour
			}
		}

		lowerBoundary := t.Settings.OffsetWithinHour + (i-1)*t.Settings.SecondsBetweenCaptures
		upperBoundary := t.Settings.OffsetWithinHour + (i)*t.Settings.SecondsBetweenCaptures

		if lowerBoundary <= secondsIntoCurrentHour && secondsIntoCurrentHour <= upperBoundary {
			return upperBoundary - secondsIntoCurrentHour
		}
	}

	return 3600 - secondsIntoCurrentHour + t.Settings.OffsetWithinHour
}
